package main

//ycc 这部分配置随代码发布，不能修改
var yccconfig = `
Title="ycc3_test"
coinSymbol="YCC"
TestNet=true
FixTime=false
version="6.3.0"


[log]
# 日志级别，支持debug(dbug)/info/warn/error(eror)/crit
loglevel = "info"
logConsoleLevel = "error"
# 日志文件名，可带目录，所有生成的日志文件都放到此目录下
logFile = "logs/chain33.log"
# 单个日志文件的最大值（单位：兆）
maxFileSize = 300
# 最多保存的历史日志文件个数
maxBackups = 100
# 最多保存的历史日志消息（单位：天）
maxAge = 28
# 日志文件名是否使用本地事件（否则使用UTC时间）
localTime = true
# 历史日志文件是否压缩（压缩格式为gz）
compress = true
# 是否打印调用源文件和行号
callerFile = true
# 是否打印调用方法
callerFunction = false

[blockchain]
defCacheSize=128
maxFetchBlockNum=128
timeoutSeconds=5
batchBlockNum=128
driver="leveldb"
dbPath="datadir"
dbCache=64
isStrongConsistency=false
singleMode=false
batchsync=false
isRecordBlockSequence=true
isParaChain=false
enableTxQuickIndex=true
enableReExecLocal=true
txHeight=false
disableShard=true

[p2p]
# p2p类型
types=["dht"]
# 是否启动P2P服务
enable=true
# 使用的数据库类型
driver="leveldb"
# 使用的数据库类型
dbPath="datadir/addrbook"
# 数据库缓存大小
dbCache=4
# GRPC请求日志文件
grpcLogFile="grpc33.log"
#waitPid 等待seed导入
waitPid=false

[p2p.sub.dht]
bootstraps=["/ip4/183.129.226.76/tcp/13801/p2p/16Uiu2HAmQiU11LNm4mQPXatnxKpGyYBVzJWD34VEcVKY1t5hroGQ"]
port=13801
channel=2
minLtBlockTxNum=20000

[rpc]
jrpcBindAddr="localhost:8803"
grpcBindAddr="localhost:8804"
whitelist=["127.0.0.1"]
jrpcFuncWhitelist=["*"]
grpcFuncWhitelist=["*"]

[mempool]
name="price"
poolCacheSize=1024000
minTxFeeRate=10000
maxTxNumPerAccount=10000

[consensus]
name="pos33"
minerstart=true
genesisBlockTime=1514533394
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
minerExecs=["pos33"]

[consensus.sub.pos33]
genesisBlockTime=1514533394
listenPort="10901"
bootPeers=["/ip4/183.129.226.76/tcp/10901/p2p/16Uiu2HAmErmNhtS145Lv5fe9FWrHSrNjPkp1eMLeLgi6t3sdr1of"]

[[consensus.sub.pos33.genesis]]
minerAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
returnAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
count=10000

[mver.consensus]
fundKeyAddr = "1Wj2mPoBwJMVwAQLKPNDseGpDNibDt9Vq"
powLimitBits="0x1f00ffff"
maxTxNumber=10000

[mver.consensus.ForkChainParamV1]
maxTxNumber=10000

[mver.consensus.ForkChainParamV2]
powLimitBits = "0x1f2fffff"

[mver.consensus.ForkTicketFundAddrV1]
fundKeyAddr = "1Wj2mPoBwJMVwAQLKPNDseGpDNibDt9Vq"

[mver.consensus.pos33]
ticketPrice = 10000
ticketWithdrawTime = 50

[store]
name="kvmvccmavl"
driver="leveldb"
storedbVersion="2.0.0"
dbPath="paradatadir/mavltree"
dbCache=128

[store.sub.mavl]
enableMavlPrefix=false
enableMVCC=false
enableMavlPrune=false
pruneHeight=10000
# 是否使能mavl数据载入内存
enableMemTree=true
# 是否使能mavl叶子节点数据载入内存
enableMemVal=true
# 缓存close ticket数目，该缓存越大同步速度越快，最大设置到1500000
tkCloseCacheLen=100000

[store.sub.kvmvccmavl]
enableMVCCIter=true
enableMavlPrefix=false
enableMVCC=false
enableMavlPrune=false
pruneMavlHeight=10000
enableMVCCPrune=false
pruneMVCCHeight=10000
# 是否使能mavl数据载入内存
enableMemTree=true
# 是否使能mavl叶子节点数据载入内存
enableMemVal=true
# 缓存close ticket数目，该缓存越大同步速度越快，最大设置到1500000
tkCloseCacheLen=100000
# 该参数针对平行链，如果平行链的ForkKvmvccmavl高度不为0,需要开启此功能,开启此功能需要从0开始执行区块
enableEmptyBlockHandle=false

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet"
dbCache=16
signType="secp256k1"

[wallet.sub.pos33]
minerdisable=false
minerwhitelist=["*"]

[wallet.sub.multisig]
rescanMultisigAddr=false

#系统中所有的fork,默认用chain33的测试网络的
#但是我们可以替换
[fork.system]
ForkChainParamV1= 0
ForkCheckTxDup=0
ForkBlockHash=10000000000000000
ForkMinerTime= 0
ForkTransferExec=0
ForkExecKey=0
ForkTxGroup=0
ForkResetTx0=0
ForkWithdraw=0
ForkExecRollback=0
ForkCheckBlockTime=0
ForkTxHeight=0
ForkTxGroupPara=0
ForkChainParamV2=0
ForkMultiSignAddress=0
ForkStateDBSet=0
ForkLocalDBAccess=0
ForkBlockCheck=0
ForkBase58AddressCheck=0
#平行链上使能平行链执行器如user.p.x.coins执行器的注册，缺省为0，对已有的平行链需要设置一个fork高度
ForkEnableParaRegExec=0
ForkCacheDriver=0
ForkTicketFundAddrV1=-1 #fork6.3
#主链和平行链都使用同一个fork高度
ForkRootHash=0 

[fork.sub.pos33]
Enable=0

[fork.sub.coins]
Enable=0

[fork.sub.ticket]
Enable=0
ForkTicketId =0
ForkTicketVrf =0

[fork.sub.retrieve]
Enable=0
ForkRetrive=0
ForkRetriveAsset=0

[fork.sub.hashlock]
Enable=0
ForkBadRepeatSecret=0

[fork.sub.manage]
Enable=0
ForkManageExec=0

[fork.sub.token]
Enable=0
ForkTokenBlackList= 0
ForkBadTokenSymbol= 0
ForkTokenPrice=0
ForkTokenSymbolWithNumber=0
ForkTokenCheck= 0

[fork.sub.trade]
Enable=0
ForkTradeBuyLimit= 0
ForkTradeAsset= 0
ForkTradeID = 0
ForkTradeFixAssetDB = 0
ForkTradePrice = 0

[fork.sub.paracross]
Enable=0
ForkParacrossWithdrawFromParachain=0
ForkParacrossCommitTx=0
ForkLoopCheckCommitTxDone=0
#仅平行链适用，自共识分阶段开启，缺省是0，若对应主链高度7000000之前开启过自共识，需要重新配置此分叉，并为之前自共识设置selfConsensEnablePreContract配置项
ForkParaSelfConsStages=0
ForkParaAssetTransferRbk=0

[fork.sub.evm]
Enable=0
ForkEVMState=0
ForkEVMABI=0
ForkEVMFrozen=0
ForkEVMKVHash=0

[fork.sub.blackwhite]
Enable=0
ForkBlackWhiteV2=0

[fork.sub.cert]
Enable=0

[fork.sub.guess]
Enable=0

[fork.sub.lottery]
Enable=0

[fork.sub.oracle]
Enable=0

[fork.sub.relay]
Enable=0

[fork.sub.norm]
Enable=0

[fork.sub.pokerbull]
Enable=0

[fork.sub.privacy]
Enable=0

[fork.sub.game]
Enable=0

[fork.sub.multisig]
Enable=0

[fork.sub.unfreeze]
Enable=0
ForkTerminatePart=0
ForkUnfreezeIDX= 0

[fork.sub.autonomy]
Enable=0

[fork.sub.jsvm]
Enable=0

[fork.sub.issuance]
Enable=0
ForkIssuanceTableUpdate=0

[fork.sub.collateralize]
Enable=0
ForkCollateralizeTableUpdate=0

#对已有的平行链如果不是从0开始同步数据，需要设置这个kvmvccmavl的对应平行链高度的fork，如果从0开始同步，statehash会跟以前mavl的不同
[fork.sub.store-kvmvccmavl]
ForkKvmvccmavl=0

[exec]
isFree=false
minExecFee=100000
maxExecFee=1000000000
enableStat=false
enableMVCC=false
alias=["token1:token","token2:token","token3:token"]

[metrics]
#是否使能发送metrics数据的发送
enableMetrics=false
#数据保存模式
dataEmitMode="influxdb"

[metrics.sub.influxdb]
#以纳秒为单位的发送间隔
duration=1000000000
url="http://influxdb:8086"
database="chain33metrics"
username=""
password=""
namespace=""
`
