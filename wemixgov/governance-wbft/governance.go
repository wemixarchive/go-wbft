package govwbft

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
)

func init() {
	// to avoid import cycle
	params.CheckGovContractVersions = checkGovContractVersions
}

func checkGovContractVersions(govContracts *params.GovContracts) error {
	if GovContractCodes[gov.CONTRACT_GOV_CONFIG][govContracts.GovConfig.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govConfig`: unsupported version %s", govContracts.GovConfig.Version)
	}
	if GovContractCodes[gov.CONTRACT_GOV_STAKING][govContracts.GovStaking.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govStaking`: unsupported version %s", govContracts.GovStaking.Version)
	}
	if GovContractCodes[gov.CONTRACT_GOV_REWARDEE_IMP][govContracts.GovRewardeeImp.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govRewardeeImp`: unsupported version %s", govContracts.GovRewardeeImp.Version)
	}
	if govContracts.GovNCP != nil && GovContractCodes[gov.CONTRACT_GOV_NCP][govContracts.GovNCP.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govNCP`: unsupported version %s", govContracts.GovNCP.Version)
	}
	return nil
}

func InitializeNCP(govNCPAddress common.Address, ncps []common.Address) []params.StateParam {
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
				Address: govNCPAddress,
				Key:     CalculateMappingSlot(indexSlot, ncp),
				Value:   common.BigToHash(newLength),
			},
			// set value slot
			params.StateParam{
				Address: govNCPAddress,
				Key:     CalculateDynamicSlot(valueSlot, new(big.Int).SetUint64(currentIdx)),
				Value:   common.BytesToHash(ncp.Bytes()),
			},
		)
		duplicated[ncp] = struct{}{}
		currentIdx++
	}
	if newLength.Sign() > 0 {
		param = append(param, params.StateParam{
			Address: govNCPAddress,
			Key:     valueSlot,
			Value:   common.BigToHash(newLength),
		})
	}
	return param
}

func NCPStakers(govStakingAddress, govNCPAddress common.Address, state StateReader) []common.Address {
	stakers := make([]common.Address, 0)
	ncps := NCPList(govNCPAddress, state)
	for _, ncp := range ncps {
		v := StakerByOperator(govStakingAddress, state, ncp)
		if v != (common.Address{}) {
			stakers = append(stakers, v)
		}
	}
	return stakers
}

func NCPTotalStaking(govStakingAddress, govNCPAddress common.Address, state StateReader) *big.Int {
	totalStaking := new(big.Int)
	stakers := NCPStakers(govStakingAddress, govNCPAddress, state)
	for _, v := range stakers {
		totalStaking.Add(totalStaking, GetTotalStaked(govStakingAddress, state, v))
	}
	return totalStaking
}

func NCPStakerInfoMap(govStakingAddress, govNCPAddress common.Address, state StateReader) map[common.Address]Staker {
	stakerInfos := make(map[common.Address]Staker)
	stakers := NCPStakers(govStakingAddress, govNCPAddress, state)
	for _, v := range stakers {
		stakerInfos[v] = StakerInfo(govStakingAddress, state, v)
	}
	return stakerInfos
}
