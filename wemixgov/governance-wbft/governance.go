package governancewbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Governance struct {
	*GovStaking
	*GovNCP
}

func NewGovernance(stakingAddr, ncpAddr common.Address) *Governance {
	return &Governance{
		GovStaking: NewGovStaking(stakingAddr),
		GovNCP:     NewGovNCP(ncpAddr),
	}
}

func (g *Governance) IsNCPValidator(stateDB StateDB, validator common.Address) bool {
	if !g.IsValidator(stateDB, validator) {
		return false
	}
	return g.IsNCP(stateDB, g.getStaker(stateDB, g.validatorInfoSlot(validator)))
}

func (g *Governance) NCPValidators(stateDB StateDB) []common.Address {
	validators := make([]common.Address, 0)
	ncps := g.NCPList(stateDB)
	for _, ncp := range ncps {
		v := g.ValidatorByStaker(stateDB, ncp)
		if v != (common.Address{}) {
			validators = append(validators, v)
		}
	}
	return validators
}

func (g *Governance) NCPTotalStaking(stateDB StateDB) *big.Int {
	totalStaking := new(big.Int)
	validators := g.NCPValidators(stateDB)
	for _, v := range validators {
		totalStaking.Add(totalStaking, g.getStaking(stateDB, g.validatorInfoSlot(v)))
	}
	return totalStaking
}

func (g *Governance) NCPValidatorInfoMap(stateDB StateDB) map[common.Address]Validator {
	validatorInfos := make(map[common.Address]Validator)
	validators := g.NCPValidators(stateDB)
	for _, v := range validators {
		validatorInfos[v] = g.ValidatorInfo(stateDB, v)
	}
	return validatorInfos
}
