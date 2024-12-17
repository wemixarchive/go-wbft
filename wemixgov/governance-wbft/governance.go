package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var (
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
