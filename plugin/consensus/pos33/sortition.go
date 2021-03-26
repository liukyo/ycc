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

func (n *node) voterSort(seed []byte, height int64, round, num int, diff float64) []*pt.Pos33SortMsg {
	count := n.myCount()
	priv := n.getPriv()
	if priv == nil {
		return nil
	}

	diff *= float64(pt.Pos33VoterSize) / float64(pt.Pos33MakerSize)

	input := &pt.VrfInput{Seed: seed, Height: height, Round: int32(round), Step: int32(1)}
	vrfHash, vrfProof := calcuVrfHash(input, priv)
	proof := &pt.HashProof{
		Input:    input,
		Diff:     diff,
		VrfHash:  vrfHash,
		VrfProof: vrfProof,
		Pubkey:   priv.PubKey().Bytes(),
	}

	var msgs []*pt.Pos33SortMsg
	for i := 0; i < count; i++ {
		data := fmt.Sprintf("%x+%d+%d", vrfHash, i, num)
		hash := hash2([]byte(data))

		// 转为big.Float计算，比较难度diff
		y := new(big.Int).SetBytes(hash)
		z := new(big.Float).SetInt(y)
		if new(big.Float).Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
			continue
		}

		// 符合，表示抽中了
		m := &pt.Pos33SortMsg{
			SortHash: &pt.SortHash{Hash: hash, Index: int64(i), Num: int32(num)},
			Proof:    proof,
		}
		msgs = append(msgs, m)
	}

	if len(msgs) == 0 {
		return nil
	}
	if len(msgs) > pt.Pos33RewardVotes {
		sort.Sort(pt.Sorts(msgs))
		msgs = msgs[:pt.Pos33RewardVotes]
	}
	plog.Info("voter sort", "height", height, "round", round, "mycount", count, "diff", diff*1000000, "addr", address.PubKeyToAddr(proof.Pubkey)[:16])
	return msgs
}

func (n *node) makerSort(seed []byte, height int64, round int) []*pt.Pos33SortMsg {
	count := n.myCount()

	priv := n.getPriv()
	if priv == nil {
		return nil
	}

	diff := n.getDiff(height, round)
	input := &pt.VrfInput{Seed: seed, Height: height, Round: int32(round), Step: int32(0)}
	vrfHash, vrfProof := calcuVrfHash(input, priv)
	proof := &pt.HashProof{
		Input:    input,
		Diff:     diff,
		VrfHash:  vrfHash,
		VrfProof: vrfProof,
		Pubkey:   priv.PubKey().Bytes(),
	}

	var minSort *pt.Pos33SortMsg
	for j := 0; j < 3; j++ {
		for i := 0; i < count; i++ {
			data := fmt.Sprintf("%x+%d+%d", vrfHash, i, j)
			hash := hash2([]byte(data))

			// 转为big.Float计算，比较难度diff
			y := new(big.Int).SetBytes(hash)
			z := new(big.Float).SetInt(y)
			if new(big.Float).Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
				continue
			}

			// 符合，表示抽中了
			m := &pt.Pos33SortMsg{
				SortHash: &pt.SortHash{Hash: hash, Index: int64(i), Num: int32(j)},
				Proof:    proof,
			}
			if minSort == nil {
				minSort = m
			}
			// minHash use string compare, define a rule for which one is min
			if string(minSort.SortHash.Hash) > string(hash) {
				minSort = m
			}
		}
	}
	plog.Info("maker sort", "height", height, "round", round, "mycount", count, "diff", diff*1000000, "addr", address.PubKeyToAddr(proof.Pubkey)[:16], "sortHash", minSort != nil)
	return []*pt.Pos33SortMsg{minSort}
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

func (n *node) verifySort(height int64, step int, seed []byte, m *pt.Pos33SortMsg) error {
	if height <= pt.Pos33SortBlocks {
		return nil
	}
	if m == nil || m.Proof == nil || m.SortHash == nil || m.Proof.Input == nil {
		return fmt.Errorf("verifySort error: sort msg is nil")
	}

	addr := address.PubKeyToAddr(m.Proof.Pubkey)
	d, err := n.queryDeposit(addr)
	if err != nil {
		return err
	}
	count := d.Count
	if d.CloseHeight >= height-pt.Pos33SortBlocks {
		count = d.PreCount
	}
	if count <= m.SortHash.Index {
		return fmt.Errorf("sort index %d > %d your count, height %d, close-height %d, precount %d", m.SortHash.Index, count, height, d.CloseHeight, d.PreCount)
	}

	round := m.Proof.Input.Round
	input := &pt.VrfInput{Seed: seed, Height: height, Round: round, Step: int32(step)}
	in := types.Encode(input)
	err = vrfVerify(m.Proof.Pubkey, in, m.Proof.VrfProof, m.Proof.VrfHash)
	if err != nil {
		plog.Info("vrfVerify error", "err", err, "height", height, "round", round, "step", step, "seed")
		return err
	}
	if m.SortHash.Num >= 3 || m.SortHash.Num < 0 {
		return fmt.Errorf("sort number > 3", "num", m.SortHash.Num)
	}
	data := fmt.Sprintf("%x+%d+%d", m.Proof.VrfHash, m.SortHash.Index, m.SortHash.Num)
	hash := hash2([]byte(data))
	if string(hash) != string(m.SortHash.Hash) {
		return fmt.Errorf("sort hash error")
	}

	// sz := pt.Pos33VoterSize
	// if step == 0 {
	// 	sz = pt.Pos33MakerSize
	// }
	// minDiff := float64(sz) / float64(n.allCount(height-pt.Pos33SortBlocks))
	// if m.Proof.Diff < minDiff {
	// 	return fmt.Errorf("diff too low")
	// }

	y := new(big.Int).SetBytes(hash)
	z := new(big.Float).SetInt(y)
	if new(big.Float).Quo(z, fmax).Cmp(big.NewFloat(m.Proof.Diff)) > 0 {
		plog.Error("verifySort diff error", "height", height, "step", step, "round", round, "diff", m.Proof.Diff*1000000, "addr", address.PubKeyToAddr(m.Proof.Pubkey))
		return errDiff
	}

	return nil
}

func hash2(data []byte) []byte {
	return crypto.Sha256(crypto.Sha256(data))
}

func (n *node) getMakerSorts(height int64, round int) []*pt.Pos33SortMsg {
	sortHeight := height - pt.Pos33SortBlocks
	seed, err := n.getSortSeed(sortHeight)
	if err != nil {
		plog.Error("bp error", "err", err)
		return nil
	}
	var pss [3]*pt.Pos33SortMsg
	for i := 0; i < 3; i++ {
		var minSort *pt.Pos33SortMsg
		_, ok := n.cms[height]
		if !ok {
			return nil
		}
		_, ok = n.cms[height][round]
		if !ok {
			return nil
		}
		for _, s := range n.cms[height][round][i] {
			if s == nil {
				continue
			}
			err := n.checkSort(s, seed, 0)
			if err != nil {
				plog.Error("checkSort error", "err", err)
				continue
			}
			if minSort == nil {
				minSort = s
			}
			if string(minSort.SortHash.Hash) > string(s.SortHash.Hash) {
				minSort = s
			}
		}
		if len(pss) == 0 {
			continue
		}
		pss[i] = minSort
	}
	return pss[:]
}
