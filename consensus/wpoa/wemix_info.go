package wpoa

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var (
	wemixPoA *WemixPoA
)

func SetWemixPoA(wpoa *WemixPoA) {
	wemixPoA = wpoa
}

func StartWemix(currentBlock *types.Header) error {
	return wemixPoA.SetBootInfo(currentBlock)
}

func GetLegacyBlockRewardAmount(height *big.Int) (*big.Int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := &bind.CallOpts{Context: ctx, BlockNumber: height}

	contracts, err := wemixPoA.getRegGovEnvContracts(ctx, height)
	if err != nil {
		return nil, err
	}
	rewardAmount, err := contracts.EnvStorageImp.GetBlockRewardAmount(opts)
	if err != nil {
		return nil, err
	}
	return rewardAmount, nil
}

func SuggestGasPrice() *big.Int {
	defaultFee := big.NewInt(100 * params.GWei)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	contracts, err := wemixPoA.getRegGovEnvContracts(ctx, nil)
	if err != nil {
		return defaultFee
	}
	fee, err := contracts.EnvStorageImp.GetMaxPriorityFeePerGas(nil)
	if err != nil {
		return defaultFee
	} else {
		return fee
	}
}

func CalcBaseFee(config *params.ChainConfig, parent *types.Header) *big.Int {
	return wemixPoA.CalcBaseFee(config, parent)
}

func Info(block *types.Header) interface{} {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	contracts, err := wemixPoA.getRegGovEnvContracts(ctx, nil)
	if err != nil {
		return ""
	}
	govInfo, err := wemixPoA.GovInfo(block)
	if err != nil {
		return ""
	}

	ca := contracts.Address()
	info := &map[string]interface{}{
		"registry":                  ca.Registry,
		"governance":                ca.Gov,
		"staking":                   ca.Staking,
		"modifiedblock":             govInfo.ModifiedBlock,
		"blocksPer":                 govInfo.BlocksPer,
		"blockInterval":             govInfo.BlockInterval,
		"blockReward":               govInfo.BlockReward,
		"maxPriorityFeePerGas":      govInfo.MaxPriorityFeePerGas,
		"blockGasLimit":             govInfo.GasLimit,
		"maxBaseFee":                govInfo.MaxBaseFee,
		"baseFeeMaxChangeRate":      govInfo.BaseFeeMaxChangeRate,
		"gasTargetPercentage":       govInfo.GasTargetPercentage,
		"nodes":                     govInfo.Nodes,
		"defaultBriocheBlockReward": defaultBriocheBlockReward,
	}
	return info
}
