package govwbft

import (
	_ "embed"
	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
)

var (
	//go:embed govcontracts/v1/GovStaking
	GovStakingContractV1 string
	//go:embed govcontracts/v1/GovNCP
	GovNCPContractV1 string
	//go:embed govcontracts/v1/GovConfig
	GovConfigContractV1 string
	//go:embed govcontracts/v1/GovRewardeeImp
	GovRewardeeImpContractV1 string

	//go:embed govcontracts/v2/GovStaking
	GovStakingContractV2 string

	GovContractCodes map[string]map[string]string
)

func init() {
	GovContractCodes = make(map[string]map[string]string)

	GovContractCodes[gov.CONTRACT_GOV_CONFIG] = make(map[string]string)
	GovContractCodes[gov.CONTRACT_GOV_NCP] = make(map[string]string)
	GovContractCodes[gov.CONTRACT_GOV_STAKING] = make(map[string]string)
	GovContractCodes[gov.CONTRACT_GOV_REWARDEE_IMP] = make(map[string]string)

	GovContractCodes[gov.CONTRACT_GOV_CONFIG][gov.GOV_CONTRACT_VERSION_1] = GovConfigContractV1
	GovContractCodes[gov.CONTRACT_GOV_NCP][gov.GOV_CONTRACT_VERSION_1] = GovNCPContractV1
	GovContractCodes[gov.CONTRACT_GOV_STAKING][gov.GOV_CONTRACT_VERSION_1] = GovStakingContractV1
	GovContractCodes[gov.CONTRACT_GOV_REWARDEE_IMP][gov.GOV_CONTRACT_VERSION_1] = GovRewardeeImpContractV1
	GovContractCodes[gov.CONTRACT_GOV_STAKING][gov.GOV_CONTRACT_VERSION_2] = GovStakingContractV2
}
