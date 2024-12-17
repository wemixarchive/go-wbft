package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var (
	GovConstAddress   = common.HexToAddress(params.GOV_CONST_ADDRESS)
	GovStakingAddress = common.HexToAddress(params.GOV_STAKING_ADDRESS)
	GovNCPAddress     = common.HexToAddress(params.GOV_NCP_ADDRESS)
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

func IsNCPValidator(state StateReader, validator common.Address) bool {
	if !IsValidator(state, validator) {
		return false
	}
	return IsNCP(state, getStaker(state, validatorInfoSlot(validator)))
}

func NCPValidators(state StateReader) []common.Address {
	validators := make([]common.Address, 0)
	ncps := NCPList(state)
	for _, ncp := range ncps {
		v := ValidatorByStaker(state, ncp)
		if v != (common.Address{}) {
			validators = append(validators, v)
		}
	}
	return validators
}

func NCPTotalStaking(state StateReader) *big.Int {
	totalStaking := new(big.Int)
	validators := NCPValidators(state)
	for _, v := range validators {
		totalStaking.Add(totalStaking, getStaking(state, validatorInfoSlot(v)))
	}
	return totalStaking
}

func NCPValidatorInfoMap(state StateReader) map[common.Address]Validator {
	validatorInfos := make(map[common.Address]Validator)
	validators := NCPValidators(state)
	for _, v := range validators {
		validatorInfos[v] = ValidatorInfo(state, v)
	}
	return validatorInfos
}
