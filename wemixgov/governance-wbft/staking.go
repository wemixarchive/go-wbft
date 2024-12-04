package governancewbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_TOTAL_STAKING       = "0x0"
	SLOT_VALIDATOR_SET       = "0x1" // ,0x2
	SLOT_VALIDATOR_INFO      = "0x3"
	SLOT_VALIDATOR_BY_STAKER = "0x4"
)

type Validator struct {
	Staker    common.Address
	Reward    common.Address
	Staking   *big.Int
	Delegated *big.Int
}

type GovStaking struct {
	Address      common.Address
	validatorSet *EnumerableSet[common.Address]
}

func NewGovStaking(address common.Address) *GovStaking {
	return &GovStaking{
		Address:      address,
		validatorSet: NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET)),
	}
}

func (gs *GovStaking) TotalStaking(stateDB StateDB) *big.Int {
	return stateDB.GetState(gs.Address, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
}

func (gs *GovStaking) ValidatorLength(stateDB StateDB) uint64 {
	return gs.validatorSet.Length(stateDB, gs.Address)
}

func (gs *GovStaking) IsValidator(stateDB StateDB, validator common.Address) bool {
	return gs.validatorSet.Contains(stateDB, gs.Address, validator)
}

func (gs *GovStaking) Validators(stateDB StateDB) []common.Address {
	return gs.validatorSet.Values(stateDB, gs.Address)
}

func (gs *GovStaking) ValidatorAt(stateDB StateDB, index *big.Int) common.Address {
	return gs.validatorSet.At(stateDB, gs.Address, index)
}

func (gs *GovStaking) ValidatorByStaker(stateDB StateDB, staker common.Address) common.Address {
	validator := stateDB.GetState(gs.Address, CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_BY_STAKER), staker))
	return HashToAddress(validator)
}

func (gs *GovStaking) ValidatorInfo(stateDB StateDB, validator common.Address) Validator {
	baseSlot := gs.validatorInfoSlot(validator)

	return Validator{
		Staker:    gs.getStaker(stateDB, baseSlot),
		Reward:    HashToAddress(stateDB.GetState(gs.Address, IncrementHash(baseSlot, big.NewInt(1)))),
		Staking:   gs.getStaking(stateDB, baseSlot),
		Delegated: stateDB.GetState(gs.Address, IncrementHash(baseSlot, big.NewInt(3))).Big(),
	}
}

func (gs *GovStaking) ValidatorInfoMap(stateDB StateDB) map[common.Address]Validator {
	validatorInfos := make(map[common.Address]Validator)
	validators := gs.Validators(stateDB)
	for _, v := range validators {
		validatorInfos[v] = gs.ValidatorInfo(stateDB, v)
	}
	return validatorInfos
}

func (gs *GovStaking) getStaker(stateDB StateDB, baseSlot common.Hash) common.Address {
	return HashToAddress(stateDB.GetState(gs.Address, baseSlot))
}

func (gs *GovStaking) getStaking(stateDB StateDB, baseSlot common.Hash) *big.Int {
	return stateDB.GetState(gs.Address, IncrementHash(baseSlot, big.NewInt(2))).Big()
}

func (gs *GovStaking) validatorInfoSlot(validator common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_INFO), validator)
}
