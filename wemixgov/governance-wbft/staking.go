package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_TOTAL_STAKING      = "0x0"
	SLOT_STAKER_SET         = "0x1" // ,0x2
	SLOT_STAKER_INFO        = "0x3"
	SLOT_STAKER_BY_OPERATOR = "0x4"
	SLOT_USER_REWARD_INFO   = "0x9"
)

type Staker struct {
	Operator            common.Address
	Rewardee            common.Address
	FeeRecipient        common.Address
	TotalStaked         *big.Int
	Delegated           *big.Int
	FeeRate             *big.Int
	AccRewardPerStaking *big.Int
	AccFeePerStaking    *big.Int
	LastRewardBalance   *big.Int
}

type UserRewardInfo struct {
	StakingAmount    *big.Int
	PendingReward    *big.Int
	PendingFee       *big.Int
	RewardPerStaking *big.Int
	FeePerStaking    *big.Int
}

func TotalStaking(state StateReader) *big.Int {
	return state.GetState(GovStakingAddress, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
}

func StakerLength(state StateReader) uint64 {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Length(state, GovStakingAddress)
}

func IsStaker(state StateReader, staker common.Address) bool {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Contains(state, GovStakingAddress, staker)
}

func Stakers(state StateReader) []common.Address {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Values(state, GovStakingAddress)
}

func StakerAt(state StateReader, index *big.Int) common.Address {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.At(state, GovStakingAddress, index)
}

func StakerByOperator(state StateReader, operator common.Address) common.Address {
	staker := state.GetState(GovStakingAddress, CalculateMappingSlot(common.HexToHash(SLOT_STAKER_BY_OPERATOR), operator))
	return HashToAddress(staker)
}

func StakerInfo(state StateReader, staker common.Address) Staker {
	baseSlot := stakerInfoSlot(staker)

	stakerInfo := Staker{
		Operator:            getOperator(state, baseSlot),
		Rewardee:            HashToAddress(state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(1)))),
		FeeRecipient:        HashToAddress(state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(2)))),
		FeeRate:             state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(3))).Big(),
		TotalStaked:         state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(4))).Big(),
		AccRewardPerStaking: state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(5))).Big(),
		AccFeePerStaking:    state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(6))).Big(),
		LastRewardBalance:   state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(7))).Big(),
	}
	userInfo := UserInfo(state, staker, stakerInfo.Operator)
	x := new(big.Int).Set(stakerInfo.TotalStaked)
	stakerInfo.Delegated = x.Sub(x, userInfo.StakingAmount)
	return stakerInfo
}

func UserInfo(state StateReader, staker common.Address, user common.Address) UserRewardInfo {
	baseSlot := userInfoSlot(staker, user)

	return UserRewardInfo{
		StakingAmount:    state.GetState(GovStakingAddress, baseSlot).Big(),
		PendingReward:    state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(1))).Big(),
		PendingFee:       state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(2))).Big(),
		RewardPerStaking: state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(3))).Big(),
		FeePerStaking:    state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(4))).Big(),
	}
}

func StakerInfoMap(state StateReader) map[common.Address]Staker {
	stakerInfos := make(map[common.Address]Staker)
	stakers := Stakers(state)
	for _, v := range stakers {
		stakerInfos[v] = StakerInfo(state, v)
	}
	return stakerInfos
}

func GetTotalStaked(state StateReader, staker common.Address) *big.Int {
	return getTotalStaked(state, stakerInfoSlot(staker))
}

func getOperator(state StateReader, baseSlot common.Hash) common.Address {
	return HashToAddress(state.GetState(GovStakingAddress, baseSlot))
}

func getTotalStaked(state StateReader, baseSlot common.Hash) *big.Int {
	return state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(4))).Big()
}

func stakerInfoSlot(staker common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_STAKER_INFO), staker)
}

func userInfoSlot(staker common.Address, user common.Address) common.Hash {
	stakerMap := CalculateMappingSlot(common.HexToHash(SLOT_USER_REWARD_INFO), staker)
	return CalculateMappingSlot(stakerMap, user)
}
