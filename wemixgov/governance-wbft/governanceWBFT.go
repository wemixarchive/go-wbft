package governancewbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_TOTAL_STAKING  = "0x0"
	SLOT_VALIDATOR_SET  = "0x1" // ,0x2
	SLOT_VALIDATOR_INFO = "0x3"
)

type Validator struct {
	Staker    common.Address
	Reward    common.Address
	Staking   *big.Int
	Delegated *big.Int
}

type GovernanceWBFT struct {
	Address      common.Address
	validatorSet *EnumerableSet[common.Address]
}

func NewGovernanceWBFT(address common.Address) *GovernanceWBFT {
	return &GovernanceWBFT{
		Address:      address,
		validatorSet: NewAddressSet(common.HexToHash(SLOT_VALIDATOR_SET)),
	}
}

func (g *GovernanceWBFT) TotalStaking(stateDB StateDB) *big.Int {
	return stateDB.GetState(g.Address, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
}

func (g *GovernanceWBFT) ValidatorLength(stateDB StateDB) uint64 {
	return g.validatorSet.Length(stateDB, g.Address)
}

func (g *GovernanceWBFT) IsValidator(stateDB StateDB, validator common.Address) bool {
	return g.validatorSet.Contains(stateDB, g.Address, validator)
}

func (g *GovernanceWBFT) Validators(stateDB StateDB) []common.Address {
	return g.validatorSet.Values(stateDB, g.Address)
}

func (g *GovernanceWBFT) ValidatorAt(stateDB StateDB, index *big.Int) common.Address {
	return g.validatorSet.At(stateDB, g.Address, index)
}

func (g *GovernanceWBFT) ValidatorInfo(stateDB StateDB, validator common.Address) Validator {
	baseSlot := g.validatorInfoSlot(validator)

	return Validator{
		Staker:    HashToAddress(stateDB.GetState(g.Address, IncrementHash(baseSlot, big.NewInt(0)))),
		Reward:    g.getReward(stateDB, baseSlot),
		Staking:   stateDB.GetState(g.Address, IncrementHash(baseSlot, big.NewInt(2))).Big(),
		Delegated: stateDB.GetState(g.Address, IncrementHash(baseSlot, big.NewInt(3))).Big(),
	}
}

func (g *GovernanceWBFT) ValidatorRewardMap(stateDB StateDB) map[common.Address]common.Address {
	rewards := make(map[common.Address]common.Address)
	validators := g.Validators(stateDB)
	for _, v := range validators {
		rewards[v] = g.getReward(stateDB, g.validatorInfoSlot(v))
	}
	return rewards
}

func (g *GovernanceWBFT) getReward(stateDB StateDB, baseSlot common.Hash) common.Address {
	return HashToAddress(stateDB.GetState(g.Address, IncrementHash(baseSlot, big.NewInt(1))))
}

func (g *GovernanceWBFT) validatorInfoSlot(validator common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_INFO), validator)
}
