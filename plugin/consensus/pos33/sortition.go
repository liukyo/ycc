package pos33

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	"github.com/33cn/chain33/types"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/golang/protobuf/proto"
	pt "github.com/yccproject/ycc/plugin/dapp/pos33/types"
)

const diffValue = 1.0

var max = big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil)
var fmax = big.NewFloat(0).SetInt(max) // 2^^256

// 算法依据：
// 1. 通过签名，然后hash，得出的Hash值是在[0，max]的范围内均匀分布并且随机的, 那么Hash/max实在[1/max, 1]之间均匀分布的
// 2. 那么从N个选票中抽出M个选票，等价于计算N次Hash, 并且Hash/max < M/N

func calcuVrfHash(input proto.Message, priv crypto.PrivKey) ([]byte, []byte) {
	privKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), priv.Bytes())
	vrfPriv := &vrf.PrivateKey{PrivateKey: (*ecdsa.PrivateKey)(privKey)}
	in := types.Encode(input)
	vrfHash, vrfProof := vrfPriv.Evaluate(in)
	hash := vrfHash[:]
	return hash, vrfProof
}

func changeDiff(size, round int) int {
	return size
}

func (n *node) sort(seed []byte, height int64, round, step, allw int) []*pt.Pos33SortMsg {
	count := n.myCount()
	if allw < count {
		return nil
	}

	priv := n.getPriv()
	if priv == nil {
		return nil
	}
	input := &pt.VrfInput{Seed: seed, Height: height, Round: int32(round), Step: int32(step)}
	vrfHash, vrfProof := calcuVrfHash(input, priv)
	proof := &pt.HashProof{
		Input:    input,
		VrfHash:  vrfHash,
		VrfProof: vrfProof,
		Pubkey:   priv.PubKey().Bytes(),
	}

	diff := n.calcDiff(step, allw, round)
	var msgs []*pt.Pos33SortMsg
	var minHash []byte
	index := 0
	for i := 0; i < count; i++ {
		data := fmt.Sprintf("%x+%d", vrfHash, i)
		hash := hash2([]byte(data))

		// 转为big.Float计算，比较难度diff
		y := new(big.Int).SetBytes(hash)
		z := new(big.Float).SetInt(y)
		if new(big.Float).Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
			continue
		}

		if minHash == nil {
			minHash = hash
		}
		// minHash use string compare, define a rule for which one is min
		if string(minHash) > string(hash) {
			minHash = hash
			index = len(msgs)
		}
		// 符合，表示抽中了
		m := &pt.Pos33SortMsg{
			SortHash: &pt.SortHash{Hash: hash, Index: int64(i)},
			Proof:    proof,
		}
		msgs = append(msgs, m)
	}

	plog.Info("block sort", "height", height, "round", round, "step", step, "allw", allw, "mycount", count, "len", len(msgs))

	if len(msgs) == 0 {
		return nil
	}
	if step == 0 {
		return []*pt.Pos33SortMsg{msgs[index]}
	}
	sort.Sort(pt.Sorts(msgs))
	c := pt.Pos33VoterSize
	if len(msgs) > c {
		return msgs[:c]
	}
	return msgs
}

func vrfVerify(pub []byte, input []byte, proof []byte, hash []byte) error {
	pubKey, err := secp256k1.ParsePubKey(pub, secp256k1.S256())
	if err != nil {
		plog.Error("vrfVerify", "err", err)
		return pt.ErrVrfVerify
	}
	vrfPub := &vrf.PublicKey{PublicKey: (*ecdsa.PublicKey)(pubKey)}
	vrfHash, err := vrfPub.ProofToHash(input, proof)
	if err != nil {
		plog.Error("vrfVerify", "err", err)
		return pt.ErrVrfVerify
	}
	// plog.Debug("vrf verify", "ProofToHash", fmt.Sprintf("(%x, %x): %x", input, proof, vrfHash), "hash", hex.EncodeToString(hash))
	if !bytes.Equal(vrfHash[:], hash) {
		plog.Error("vrfVerify", "err", fmt.Errorf("invalid VRF hash"))
		return pt.ErrVrfVerify
	}
	return nil
}

var errDiff = errors.New("diff error")

func (n *node) queryDeposit(addr string) (*pt.Pos33DepositMsg, error) {
	resp, err := n.GetAPI().Query(pt.Pos33TicketX, "Pos33Deposit", &types.ReqAddr{Addr: addr})
	if err != nil {
		return nil, err
	}
	reply := resp.(*pt.Pos33DepositMsg)
	return reply, nil
}

// 本轮难度：委员会票数 / (总票数 * 在线率)
func (n *node) calcDiff(step, allw, round int) float64 {
	size := pt.Pos33VoterSize
	if step == 0 {
		size = pt.Pos33ProposerSize
	}

	onlineR := 1.
	n.lock.Lock()
	if len(n.nvsMap) >= calcuDiffBlockN {
		l := 0
		for _, n := range n.nvsMap {
			l += n
		}
		onlineR = float64(l) / float64(len(n.nvsMap)) / float64(pt.Pos33RewardVotes)
	}
	n.lock.Unlock()
	// onlineR -= 0.03 * float64(round)
	return float64(size) / float64(allw) / onlineR
}

func (n *node) verifySort(height int64, step, allw int, seed []byte, m *pt.Pos33SortMsg) error {
	if height <= pt.Pos33SortitionSize {
		return nil
	}
	if m == nil || m.Proof == nil || m.SortHash == nil || m.Proof.Input == nil {
		return fmt.Errorf("verifySort error: sort msg is nil")
	}
	round := m.Proof.Input.Round
	diff := n.calcDiff(step, allw, int(round))

	addr := address.PubKeyToAddr(m.Proof.Pubkey)
	d, err := n.queryDeposit(addr)
	if err != nil {
		return err
	}
	count := d.Count
	if d.CloseHeight >= height-pt.Pos33SortitionSize {
		count = d.PreCount
	}
	if count <= m.SortHash.Index {
		return fmt.Errorf("sort index %d > %d your count, height %d, close-height %d, precount %d", m.SortHash.Index, count, height, d.CloseHeight, d.PreCount)
	}

	input := &pt.VrfInput{Seed: seed, Height: height, Round: round, Step: int32(step)}
	in := types.Encode(input)
	err = vrfVerify(m.Proof.Pubkey, in, m.Proof.VrfProof, m.Proof.VrfHash)
	if err != nil {
		return err
	}
	data := fmt.Sprintf("%x+%d", m.Proof.VrfHash, m.SortHash.Index)
	hash := hash2([]byte(data))
	if string(hash) != string(m.SortHash.Hash) {
		return fmt.Errorf("sort hash error")
	}

	y := new(big.Int).SetBytes(hash)
	z := new(big.Float).SetInt(y)
	if new(big.Float).Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
		plog.Error("verifySort diff error", "height", height, "step", step, "round", round, "allw", allw)
		return errDiff
	}

	return nil
}

func hash2(data []byte) []byte {
	return crypto.Sha256(crypto.Sha256(data))
}

func (n *node) bp(height int64, round int) (string, []byte) {
	sortHeight := height - pt.Pos33SortitionSize
	seed, err := n.getSortSeed(sortHeight)
	if err != nil {
		plog.Error("bp error", "err", err)
		return "", nil
	}
	allw := n.allCount(sortHeight)
	pss := make(map[string]*pt.Pos33SortMsg)
	for _, s := range n.cps[height][round] {
		err := n.checkSort(s, seed, allw, 0)
		if err != nil {
			plog.Error("checkSort error", "err", err)
			continue
		}
		pss[string(s.SortHash.Hash)] = s
	}
	if len(pss) == 0 {
		return "", nil
	}

	// find min sortition hash, use string compare
	var min string
	var ss *pt.Pos33SortMsg
	for sh, s := range pss {
		// _, ok := n.blackList[string(s.Proof.Pubkey)]
		// if ok {
		// 	continue
		// }
		if min == "" {
			min = sh
			ss = s
		} else {
			if min > sh {
				min = sh
				ss = s
			}
		}
	}

	return min, ss.Proof.Pubkey
}
