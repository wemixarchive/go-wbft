package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
)

var (
	GovConfigAddress      = common.HexToAddress(params.GOV_CONFIG_ADDRESS)
	GovStakingAddress     = common.HexToAddress(params.GOV_STAKING_ADDRESS)
	GovNCPAddress         = common.HexToAddress(params.GOV_NCP_ADDRESS)
	GovRewardeeImpAddress = common.HexToAddress(params.GOV_REWARDEE_IMP_ADDRESS)
)

func InitializeNCP(ncps []common.Address) []params.StateParam {
	param := make([]params.StateParam, 0)

	valueSlot := common.HexToHash(SLOT_NCP_LIST)
	indexSlot := IncrementHash(valueSlot, big.NewInt(1))
	duplicated := make(map[common.Address]struct{})

	currentIdx := uint64(0)
	newLength := new(big.Int)
	for _, ncp := range ncps {
		if _, ok := duplicated[ncp]; ok {
			continue
		}
		newLength = new(big.Int).SetUint64(currentIdx + 1)

		param = append(param,
			// set index slot
			params.StateParam{
				Address: GovNCPAddress,
				Key:     CalculateMappingSlot(indexSlot, ncp),
				Value:   common.BigToHash(newLength),
			},
			// set value slot
			params.StateParam{
				Address: GovNCPAddress,
				Key:     CalculateDynamicSlot(valueSlot, new(big.Int).SetUint64(currentIdx)),
				Value:   common.BytesToHash(ncp.Bytes()),
			},
		)
		duplicated[ncp] = struct{}{}
		currentIdx++
	}
	if newLength.Sign() > 0 {
		param = append(param, params.StateParam{
			Address: GovNCPAddress,
			Key:     valueSlot,
			Value:   common.BigToHash(newLength),
		})
	}
	return param
}

func IsNCPStaker(state StateReader, staker common.Address) bool {
	if !IsStaker(state, staker) {
		return false
	}
	return IsNCP(state, getOperator(state, stakerInfoSlot(staker)))
}

func NCPStakers(state StateReader) []common.Address {
	stakers := make([]common.Address, 0)
	ncps := NCPList(state)
	for _, ncp := range ncps {
		v := StakerByOperator(state, ncp)
		if v != (common.Address{}) {
			stakers = append(stakers, v)
		}
	}
	return stakers
}

func NCPTotalStaking(state StateReader) *big.Int {
	totalStaking := new(big.Int)
	stakers := NCPStakers(state)
	for _, v := range stakers {
		totalStaking.Add(totalStaking, GetTotalStaked(state, v))
	}
	return totalStaking
}

func NCPStakerInfoMap(state StateReader) map[common.Address]Staker {
	stakerInfos := make(map[common.Address]Staker)
	stakers := NCPStakers(state)
	for _, v := range stakers {
		stakerInfos[v] = StakerInfo(state, v)
	}
	return stakerInfos
}

func BuildGovTransitionParams(config *params.ChainConfig) (
	codes []params.CodeParam,
	states []params.StateParam,
) {
	codes = []params.CodeParam{
		{Address: GovConfigAddress, Code: GovConfigContract},
		{Address: GovStakingAddress, Code: GovStakingContract},
		{Address: GovRewardeeImpAddress, Code: GovRewardeeImpContract},
	}

	gp := config.QBFT.GovParams
	if gp.MinimumStaking == nil {
		x, ok := new(big.Int).SetString("69e10de76676d0800000", 16)
		if !ok {
			panic("invalid hexadecimal literal")
		}
		gp.MinimumStaking = (*math.HexOrDecimal256)(x)
	}
	if gp.MaximumStaking == nil {
		gp.MaximumStaking = (*math.HexOrDecimal256)(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))
	}
	if gp.UnbondingStaker == 0 {
		gp.UnbondingStaker = 604800 // 7 days
	}
	if gp.UnbondingDelegator == 0 {
		gp.UnbondingDelegator = 259200 // 3 days
	}
	if gp.FeePrecision == 0 {
		gp.FeePrecision = 10000 // 0.01%
	}
	if gp.ChangeFeeDelay == 0 {
		gp.ChangeFeeDelay = 604800 // 7 days
	}
	if gp.MinStakers == 0 {
		gp.MinStakers = 1
	}

	states = []params.StateParam{
		{Address: GovConfigAddress, Key: common.BigToHash(big.NewInt(0)), Value: common.BigToHash((*big.Int)(gp.MinimumStaking))},
		{Address: GovConfigAddress, Key: common.BigToHash(big.NewInt(1)), Value: common.BigToHash((*big.Int)(gp.MaximumStaking))},
		{Address: GovConfigAddress, Key: common.BigToHash(big.NewInt(2)), Value: common.BigToHash(new(big.Int).SetUint64(gp.UnbondingStaker))},
		{Address: GovConfigAddress, Key: common.BigToHash(big.NewInt(3)), Value: common.BigToHash(new(big.Int).SetUint64(gp.UnbondingDelegator))},
		{Address: GovConfigAddress, Key: common.BigToHash(big.NewInt(4)), Value: common.BigToHash(new(big.Int).SetUint64(gp.FeePrecision))},
		{Address: GovConfigAddress, Key: common.BigToHash(big.NewInt(5)), Value: common.BigToHash(new(big.Int).SetUint64(gp.ChangeFeeDelay))},
		{Address: GovConfigAddress, Key: common.BigToHash(big.NewInt(6)), Value: common.BigToHash(new(big.Int).SetUint64(gp.MinStakers))},
	}

	return
}
