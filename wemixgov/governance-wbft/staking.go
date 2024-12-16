package govwbft

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

func TotalStaking(stateDB StateDB) *big.Int {
	return stateDB.GetState(GovStakingAddress, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
}

func ValidatorLength(stateDB StateDB) uint64 {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.Length(stateDB, GovStakingAddress)
}

func IsValidator(stateDB StateDB, validator common.Address) bool {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.Contains(stateDB, GovStakingAddress, validator)
}

func Validators(stateDB StateDB) []common.Address {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.Values(stateDB, GovStakingAddress)
}

func ValidatorAt(stateDB StateDB, index *big.Int) common.Address {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.At(stateDB, GovStakingAddress, index)
}

func ValidatorByStaker(stateDB StateDB, staker common.Address) common.Address {
	validator := stateDB.GetState(GovStakingAddress, CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_BY_STAKER), staker))
	return HashToAddress(validator)
}

func ValidatorInfo(stateDB StateDB, validator common.Address) Validator {
	baseSlot := validatorInfoSlot(validator)

	return Validator{
		Staker:    getStaker(stateDB, baseSlot),
		Reward:    HashToAddress(stateDB.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(1)))),
		Staking:   getStaking(stateDB, baseSlot),
		Delegated: stateDB.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(3))).Big(),
	}
}

func ValidatorInfoMap(stateDB StateDB) map[common.Address]Validator {
	validatorInfos := make(map[common.Address]Validator)
	validators := Validators(stateDB)
	for _, v := range validators {
		validatorInfos[v] = ValidatorInfo(stateDB, v)
	}
	return validatorInfos
}

func getStaker(stateDB StateDB, baseSlot common.Hash) common.Address {
	return HashToAddress(stateDB.GetState(GovStakingAddress, baseSlot))
}

func getStaking(stateDB StateDB, baseSlot common.Hash) *big.Int {
	return stateDB.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(2))).Big()
}

func validatorInfoSlot(validator common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_INFO), validator)
}
