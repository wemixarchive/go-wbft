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

func TotalStaking(state StateReader) *big.Int {
	return state.GetState(GovStakingAddress, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
}

func ValidatorLength(state StateReader) uint64 {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.Length(state, GovStakingAddress)
}

func IsValidator(state StateReader, validator common.Address) bool {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.Contains(state, GovStakingAddress, validator)
}

func Validators(state StateReader) []common.Address {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.Values(state, GovStakingAddress)
}

func ValidatorAt(state StateReader, index *big.Int) common.Address {
	validatorSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET))
	return validatorSet.At(state, GovStakingAddress, index)
}

func ValidatorByStaker(state StateReader, staker common.Address) common.Address {
	validator := state.GetState(GovStakingAddress, CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_BY_STAKER), staker))
	return HashToAddress(validator)
}

func ValidatorInfo(state StateReader, validator common.Address) Validator {
	baseSlot := validatorInfoSlot(validator)

	return Validator{
		Staker:    getStaker(state, baseSlot),
		Reward:    HashToAddress(state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(1)))),
		Staking:   getStaking(state, baseSlot),
		Delegated: state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(3))).Big(),
	}
}

func ValidatorInfoMap(state StateReader) map[common.Address]Validator {
	validatorInfos := make(map[common.Address]Validator)
	validators := Validators(state)
	for _, v := range validators {
		validatorInfos[v] = ValidatorInfo(state, v)
	}
	return validatorInfos
}

func getStaker(state StateReader, baseSlot common.Hash) common.Address {
	return HashToAddress(state.GetState(GovStakingAddress, baseSlot))
}

func getStaking(state StateReader, baseSlot common.Hash) *big.Int {
	return state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(2))).Big()
}

func validatorInfoSlot(validator common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_INFO), validator)
}
