package wpoa

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

var (
	wemixPoA *WemixPoA
)

func SetWemixPoA(wpoa *WemixPoA) {
	wemixPoA = wpoa
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
