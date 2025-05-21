package main

import (
	"flag"
	"fmt"
	"path/filepath"

	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
	compile "github.com/ethereum/go-ethereum/wemixgov/governance-contract"
)

var (
	rootFlag         = flag.String("root", "../contracts-wbft", "")
	openZeppelinFlag = flag.String("openZeppelin", "../contracts", "")
)

func main() {
	flag.Parse()
	root := *rootFlag
	versions := []string{gov.GOV_CONTRACT_VERSION_1, gov.GOV_CONTRACT_VERSION_2}
	srcFiles := [][]string{
		{ // v1
			filepath.Join(filepath.Join(root, versions[0]), "GovStaking.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovNCP.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovConfig.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovRewardee.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "GovRewardeeImp.sol"),
			filepath.Join(filepath.Join(root, versions[0]), "OperatorSample.sol"),
		},
		{ // v2
			filepath.Join(filepath.Join(root, versions[1]), "GovStaking.sol"),
		},
	}
	contractBins := [][]string{
		{ // v1
			gov.CONTRACT_GOV_STAKING,
			gov.CONTRACT_GOV_NCP,
			gov.CONTRACT_GOV_CONFIG,
			gov.CONTRACT_GOV_REWARDEE,
			gov.CONTRACT_GOV_REWARDEE_IMP,
			gov.CONTRACT_OPERATOR_SAMPLE,
		},
		{ // v2
			gov.CONTRACT_GOV_STAKING,
		},
	}
	openZeppelin := *openZeppelinFlag

	for i, version := range versions {
		codeDir := filepath.Join(root, "../../governance-wbft/govcontracts/"+version)
		if compiledContracts, err := compile.Compile(openZeppelin, srcFiles[i]...,
		); err != nil {
			panic(err)
		} else if err := compiledContracts.ExportContractCode(codeDir, contractBins[i]); err != nil {
			panic(err)
		}
	}
	fmt.Println("success!")
}
