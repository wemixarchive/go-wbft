package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var (
	GovConstAddress   = common.HexToAddress("0x1000")
	GovStakingAddress = common.HexToAddress("0x1001")
	GovNCPAddress     = common.HexToAddress("0x1002")
)

func IsNCPValidator(stateDB StateDB, validator common.Address) bool {
	if !IsValidator(stateDB, validator) {
		return false
	}
	return IsNCP(stateDB, getStaker(stateDB, validatorInfoSlot(validator)))
}

func NCPValidators(stateDB StateDB) []common.Address {
	validators := make([]common.Address, 0)
	ncps := NCPList(stateDB)
	for _, ncp := range ncps {
		v := ValidatorByStaker(stateDB, ncp)
		if v != (common.Address{}) {
			validators = append(validators, v)
		}
	}
	return validators
}

func NCPTotalStaking(stateDB StateDB) *big.Int {
	totalStaking := new(big.Int)
	validators := NCPValidators(stateDB)
	for _, v := range validators {
		totalStaking.Add(totalStaking, getStaking(stateDB, validatorInfoSlot(v)))
	}
	return totalStaking
}

func NCPValidatorInfoMap(stateDB StateDB) map[common.Address]Validator {
	validatorInfos := make(map[common.Address]Validator)
	validators := NCPValidators(stateDB)
	for _, v := range validators {
		validatorInfos[v] = ValidatorInfo(stateDB, v)
	}
	return validatorInfos
}
