package wemixgov

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/wpoa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
	"github.com/pkg/errors"
)

var (
	errInvalidEnode   = errors.New("invalid enode")
	errNotInitialized = errors.New("not initialized")
	errNotFound       = errors.New("not found")
)

type WemixGov struct {
	cli                  bind.ContractBackend
	bootAccount          common.Address
	coinbaseEnodeCache   *sync.Map
	height2enode         *LruCache
	blockBuildParamsLock sync.Mutex
	blockBuildParams     *wpoa.BlockBuildParameters
}

func NewWemixGovClient(rpcCli *rpc.Client) *WemixGov {
	wg := &WemixGov{}
	wg.cli = ethclient.NewClient(rpcCli)
	wg.coinbaseEnodeCache = &sync.Map{}
	wg.height2enode = NewLruCache(10000, true)
	wg.blockBuildParams = &wpoa.BlockBuildParameters{}
	return wg
}

func (wg *WemixGov) SetBootAccount(bootAccount common.Address) error {
	wg.bootAccount = bootAccount
	return nil
}

// cached governance data to derive miner's enode
type coinbaseEnodeEntry struct {
	modifiedBlock  *big.Int
	nodes          []*wpoa.WemixNode
	coinbase2enode map[string][]byte // string(common.Address[:]) => []byte
	enode2index    map[string]int    // string([]byte) => int
}

func (wg *WemixGov) GetGovInfo(blockNumber *big.Int) (wpoa.WemixGovInfo, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	contracts, err := wg.getRegGovEnvContracts(ctx, blockNumber)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	opts := &bind.CallOpts{Context: ctx, BlockNumber: blockNumber}
	result := wpoa.WemixGovInfo{}

	contractAddresses := contracts.Address()
	result.Registry = contractAddresses.Registry
	result.Gov = contractAddresses.Gov
	result.Staking = contractAddresses.Staking

	result.ModifiedBlock, err = contracts.GovImp.ModifiedBlock(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	result.BlockInterval, err = contracts.EnvStorageImp.GetBlockCreationTime(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	result.BlocksPer, err = contracts.EnvStorageImp.GetBlocksPer(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	result.BlockReward, err = contracts.EnvStorageImp.GetBlockRewardAmount(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	result.MaxPriorityFeePerGas, err = contracts.EnvStorageImp.GetMaxPriorityFeePerGas(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	result.MaxBaseFee, err = contracts.EnvStorageImp.GetMaxBaseFee(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	result.GasLimit, result.BaseFeeMaxChangeRate, result.BaseFeeMaxChangeRate, err = contracts.EnvStorageImp.GetGasLimitAndBaseFee(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}

	result.DefaultBriocheBlockReward = new(big.Int).Set(params.DefaultBriocheBlockReward)

	nodes := make([]*wpoa.WemixNode, 0)
	nodeLength, err := contracts.GovImp.GetNodeLength(opts)
	if err != nil {
		return wpoa.WemixGovInfo{}, err
	}
	count := nodeLength.Int64()
	for i := int64(1); i <= count; i++ {
		node, err := contracts.GovImp.GetNode(opts, big.NewInt(i))
		if err != nil {
			return wpoa.WemixGovInfo{}, err
		}
		member, err := contracts.GovImp.GetMember(opts, big.NewInt(i))
		if err != nil {
			return wpoa.WemixGovInfo{}, err
		}

		sid := hex.EncodeToString(node.Enode)
		if len(sid) != 128 {
			return wpoa.WemixGovInfo{}, errInvalidEnode
		}
		idv4, _ := toIdv4(sid)
		nodes = append(nodes, &wpoa.WemixNode{
			Name:  string(node.Name),
			Enode: sid,
			Ip:    string(node.Ip),
			Id:    idv4,
			Port:  int(node.Port.Int64()),
			Addr:  member,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	result.Nodes = nodes
	return result, nil
}

func (wg *WemixGov) GetLegacyBlockRewardAmount(height *big.Int) (*big.Int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := &bind.CallOpts{Context: ctx, BlockNumber: height}

	contracts, err := wg.getRegGovEnvContracts(ctx, height)
	if err != nil {
		return nil, err
	}
	rewardAmount, err := contracts.EnvStorageImp.GetBlockRewardAmount(opts)
	if err != nil {
		return nil, err
	}
	return rewardAmount, nil
}

func (wg *WemixGov) GetMaxPriorityFeePerGas(height *big.Int) (*big.Int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	contracts, err := wg.getRegGovEnvContracts(ctx, height)
	if err != nil {
		return nil, err
	}
	fee, err := contracts.EnvStorageImp.GetMaxPriorityFeePerGas(nil)
	if err != nil {
		return nil, err
	}
	return fee, nil
}

func (wg *WemixGov) GetRewardParams(height *big.Int) (*wpoa.RewardParameters, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rp := &wpoa.RewardParameters{}
	contracts, err := wg.getRegGovEnvContracts(ctx, height)
	if err != nil {
		return nil, err
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: height}

	rp.RewardAmount, err = contracts.EnvStorageImp.GetBlockRewardAmount(opts)
	if err != nil {
		return nil, err
	}

	distributionMethod1, distributionMethod2, distributionMethod3, distributionMethod4, err := contracts.EnvStorageImp.GetBlockRewardDistributionMethod(opts)
	if err != nil {
		return nil, err
	}
	rp.DistributionMethod = []*big.Int{distributionMethod1, distributionMethod2, distributionMethod3, distributionMethod4}

	staker, err := contracts.Registry.GetContractAddress(opts, toBytes32(gov.DOMAIN_StakingReward))
	if err != nil {
		return nil, errors.Wrap(err, gov.DOMAIN_Staking)
	}
	rp.Staker = &staker

	ecoSystem, err := contracts.Registry.GetContractAddress(opts, toBytes32(gov.DOMAIN_Ecosystem))
	if err != nil {
		return nil, errors.Wrap(err, gov.DOMAIN_Ecosystem)
	}
	rp.EcoSystem = &ecoSystem

	maintenance, err := contracts.Registry.GetContractAddress(opts, toBytes32(gov.DOMAIN_Maintenance))
	if err != nil {
		return nil, errors.Wrap(err, gov.DOMAIN_Maintenance)
	}
	rp.Maintenance = &maintenance

	feeCollector, err := contracts.Registry.GetContractAddress(opts, toBytes32(gov.DOMAIN_FeeCollector))
	if err != nil {
		rp.FeeCollector = nil
	} else {
		rp.FeeCollector = &feeCollector
	}

	blocksPer, err := contracts.EnvStorageImp.GetBlocksPer(opts)
	if err != nil {
		return nil, err
	}
	rp.BlocksPer = blocksPer.Int64()

	if countBig, err := contracts.GovImp.GetMemberLength(opts); err != nil {
		return nil, err
	} else {
		count := countBig.Int64()
		for i := int64(1); i <= count; i++ {
			index := big.NewInt(i)
			if member, err := contracts.GovImp.GetMember(opts, index); err != nil {
				return nil, err
			} else if reward, err := contracts.GovImp.GetReward(opts, index); err != nil {
				return nil, err
			} else if stake, err := contracts.StakingImp.LockedBalanceOf(opts, member); err != nil {
				return nil, err
			} else {
				rp.Members = append(rp.Members, &wpoa.WemixMember{
					Staker: member,
					Reward: reward,
					Stake:  stake,
				})
			}
		}
	}
	return rp, nil
}

func (wg *WemixGov) VerifyBlockSig(height *big.Int, chain consensus.ChainHeaderReader, coinbase common.Address, nodeId []byte, hash common.Hash, sig []byte, checkMinerLimit bool) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get nodeid from the coinbase
	num := new(big.Int).Sub(height, common.Big1)
	contracts, err := wg.getRegGovEnvContracts(ctx, num)
	if err != nil {
		return err == errNotInitialized || errors.Is(err, errNotFound)
	} else if count, err := contracts.GovImp.GetMemberLength(&bind.CallOpts{Context: ctx, BlockNumber: num}); err != nil || count.Sign() == 0 {
		return err == errNotInitialized || count.Sign() == 0
	}
	gov := contracts.GovImp
	// if minerNodeId is given, i.e. present in block header, use it,
	// otherwise, derive it from the codebase
	var data []byte
	if len(nodeId) == 0 {
		nodeId, err = wg.coinbaseExists(ctx, height, gov, &coinbase)
		if err != nil || len(nodeId) == 0 {
			return false
		}
		data = append(height.Bytes(), hash.Bytes()...)
		data = crypto.Keccak256(data)
	} else {
		if _, err := wg.enodeExists(ctx, height, gov, nodeId); err != nil {
			return false
		}
		data = hash.Bytes()
	}
	pubKey, err := crypto.Ecrecover(data, sig)
	if err != nil || len(pubKey) < 1 || !bytes.Equal(nodeId, pubKey[1:]) {
		return false
	}
	// check miner limit
	if !checkMinerLimit {
		return true
	}
	ok, err := wg.verifyMinerLimit(ctx, chain, height, gov, &coinbase, nodeId)
	return err == nil && ok
}

func (wg *WemixGov) GetBlockBuildParameters(height *big.Int) (blockInterval int64, maxBaseFee, gasLimit *big.Int, baseFeeMaxChangeRate, gasTargetPercentage int64, err error) {
	err = errNotInitialized

	wg.blockBuildParamsLock.Lock()
	if wg.blockBuildParams != nil && wg.blockBuildParams.Height == height.Uint64() {
		// use cached
		blockInterval = wg.blockBuildParams.BlockInterval
		maxBaseFee = wg.blockBuildParams.MaxBaseFee
		gasLimit = wg.blockBuildParams.GasLimit
		baseFeeMaxChangeRate = wg.blockBuildParams.BaseFeeMaxChangeRate
		gasTargetPercentage = wg.blockBuildParams.GasTargetPercentage
		wg.blockBuildParamsLock.Unlock()
		err = nil
		return
	}
	wg.blockBuildParamsLock.Unlock()

	// default values
	blockInterval = 15
	maxBaseFee = big.NewInt(0)
	gasLimit = big.NewInt(0)
	baseFeeMaxChangeRate = 0
	gasTargetPercentage = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		env *gov.EnvStorageImp
		gov *gov.GovImp
	)
	if contracts, err2 := wg.getRegGovEnvContracts(ctx, height); err2 != nil {
		err = errNotInitialized
		return
	} else {
		env, gov = contracts.EnvStorageImp, contracts.GovImp
	}

	opts := &bind.CallOpts{Context: ctx, BlockNumber: height}
	if count, err2 := gov.GetMemberLength(opts); err2 != nil || count.Sign() == 0 {
		err = errNotInitialized
		return
	}
	if v, err2 := env.GetBlockCreationTime(opts); err2 != nil {
		err = errNotInitialized
		return
	} else {
		blockInterval = v.Int64()
	}

	if GasLimit, BaseFeeMaxChangeRate, GasTargetPercentage, err2 := env.GetGasLimitAndBaseFee(opts); err2 != nil {
		err = errNotInitialized
		return
	} else {
		gasLimit = GasLimit
		baseFeeMaxChangeRate = BaseFeeMaxChangeRate.Int64()
		gasTargetPercentage = GasTargetPercentage.Int64()
	}

	if maxBaseFee, err = env.GetMaxBaseFee(opts); err != nil {
		err = errNotInitialized
		return
	}

	// cache it
	wg.blockBuildParamsLock.Lock()
	wg.blockBuildParams = &wpoa.BlockBuildParameters{
		Height:               height.Uint64(),
		BlockInterval:        blockInterval,
		MaxBaseFee:           maxBaseFee,
		GasLimit:             gasLimit,
		BaseFeeMaxChangeRate: baseFeeMaxChangeRate,
		GasTargetPercentage:  gasTargetPercentage,
	}
	wg.blockBuildParamsLock.Unlock()
	err = nil
	return
}

func (wg *WemixGov) verifyMinerLimit(ctx context.Context, chain consensus.ChainHeaderReader, height *big.Int, gov *gov.GovImp, coinbase *common.Address, enode []byte) (bool, error) {
	// parent block number
	prev := new(big.Int).Sub(height, common.Big1)
	e, err := wg.getCoinbaseEnodeCache(ctx, prev, gov)
	if err != nil {
		return false, err
	}
	// if count <= 2, not enforced
	if len(e.nodes) <= 2 {
		return true, nil
	}
	// if enode is not given, derive it from the coinbase
	if len(enode) == 0 {
		enode2, ok := e.coinbase2enode[string(coinbase[:])]
		if !ok {
			return false, nil
		}
		enode = enode2
	}
	// the enode should not appear within the last (member count / 2) blocks
	limit := len(e.nodes) / 2
	if limit > int(height.Int64()-e.modifiedBlock.Int64()-1) {
		limit = int(height.Int64() - e.modifiedBlock.Int64() - 1)
	}
	for h := new(big.Int).Set(prev); limit > 0; h, limit = h.Sub(h, common.Big1), limit-1 {
		blockMinerEnode, err := wg.getBlockMiner(chain, e, h)
		if err != nil {
			return false, err
		}
		if bytes.Equal(enode[:], blockMinerEnode[:]) {
			return false, nil
		}
	}
	return true, nil
}

// return block's miner node id
func (wg *WemixGov) getBlockMiner(chain consensus.ChainHeaderReader, entry *coinbaseEnodeEntry, height *big.Int) ([]byte, error) {
	// if already cached, use it
	if enode := wg.height2enode.Get(height.Uint64()); enode != nil {
		return enode.([]byte), nil
	}
	block := chain.GetHeaderByNumber(height.Uint64())
	if len(block.MinerNodeId) == 0 {
		enode, ok := entry.coinbase2enode[string(block.Coinbase[:])]
		if !ok {
			return nil, nil
		}
		wg.height2enode.Put(height.Uint64(), enode)
		return enode, nil
	} else {
		if _, ok := entry.enode2index[string(block.MinerNodeId)]; !ok {
			return nil, nil
		}
		wg.height2enode.Put(height.Uint64(), block.MinerNodeId)
		return block.MinerNodeId, nil
	}
}

func (wg *WemixGov) getRegGovEnvContracts(ctx context.Context, height *big.Int) (*gov.GovContracts, error) {
	if ctx == nil {
		var cancel func()
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: height}
	return gov.GetGovContractsByOwner(opts, wg.cli, wg.bootAccount)
}

// returns coinbase's enode if exists in governance at given height - 1
func (wg *WemixGov) coinbaseExists(ctx context.Context, height *big.Int, gov *gov.GovImp, coinbase *common.Address) ([]byte, error) {
	e, err := wg.getCoinbaseEnodeCache(ctx, new(big.Int).Sub(height, common.Big1), gov)
	if err != nil {
		return nil, err
	}
	enode, ok := e.coinbase2enode[string(coinbase[:])]
	if !ok {
		return nil, nil
	}
	return enode, nil
}

func (wg *WemixGov) getCoinbaseEnodeCache(ctx context.Context, height *big.Int, gov *gov.GovImp) (*coinbaseEnodeEntry, error) {
	opts := &bind.CallOpts{Context: ctx, BlockNumber: height}
	modifiedBlock, err := gov.ModifiedBlock(opts)
	if err != nil {
		return nil, err
	}
	if modifiedBlock.Sign() == 0 {
		return nil, errNotInitialized
	}

	// if found in cache, use it
	if e, ok := wg.coinbaseEnodeCache.Load(modifiedBlock.Int64()); ok {
		return e.(*coinbaseEnodeEntry), nil
	}
	// otherwise, load it from the governance
	var (
		count       *big.Int
		addr        common.Address
		name, enode []byte
		e           = &coinbaseEnodeEntry{
			modifiedBlock:  modifiedBlock,
			coinbase2enode: map[string][]byte{},
			enode2index:    map[string]int{},
		}
	)
	if count, err = gov.GetNodeLength(opts); err != nil {
		return nil, err
	}
	for i := int64(1); i <= count.Int64(); i++ {
		ix := big.NewInt(i)
		if addr, err = gov.GetReward(opts, ix); err != nil {
			return nil, err
		}

		if output, err := gov.GetNode(opts, ix); err != nil {
			return nil, err
		} else {
			name, enode = output.Name, output.Enode
		}

		idv4, _ := toIdv4(hex.EncodeToString(enode))
		e.nodes = append(e.nodes, &wpoa.WemixNode{
			Name:  string(name),
			Enode: string(enode), // note that this is not in hex unlike wemixAdmin
			Id:    idv4,
			Addr:  addr,
		})
		e.coinbase2enode[string(addr[:])] = enode
		e.enode2index[string(enode)] = int(i) // 1-based, not 0-based
	}
	wg.coinbaseEnodeCache.Store(modifiedBlock.Int64(), e)
	return e, nil
}

// returns true if enode exists in governance at given height-1
func (wg *WemixGov) enodeExists(ctx context.Context, height *big.Int, gov *gov.GovImp, enode []byte) (common.Address, error) {
	e, err := wg.getCoinbaseEnodeCache(ctx, new(big.Int).Sub(height, common.Big1), gov)
	if err != nil {
		return common.Address{}, err
	}
	ix, ok := e.enode2index[string(enode)]
	if !ok {
		return common.Address{}, errNotFound
	}
	return e.nodes[ix-1].Addr, nil
}

func toIdv4(id string) (string, error) {
	if len(id) == 64 {
		return id, nil
	} else if len(id) == 128 {
		idv4, err := enode.ParseV4(fmt.Sprintf("enode://%v@127.0.0.1:8589", id))
		if err != nil {
			return "", err
		} else {
			return idv4.ID().String(), nil
		}
	} else {
		return "", fmt.Errorf("invalid V5 Identifier")
	}
}

func toBytes32(b string) [32]byte {
	var b32 [32]byte
	if len(b) > len(b32) {
		b = b[len(b)-len(b32):]
	}
	copy(b32[:], []byte(b))
	return b32
}
