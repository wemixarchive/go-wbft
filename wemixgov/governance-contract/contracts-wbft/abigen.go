package main

import (
	"flag"
	"fmt"
	"path/filepath"

	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
	compile "github.com/ethereum/go-ethereum/wemixgov/governance-contract"
)

const pkg string = "gov"

var (
	rootFlag         = flag.String("root", "../contracts-wbft", "")
	openZeppelinFlag = flag.String("openZeppelin", "../contracts", "")
)

func main() {
	flag.Parse()
	root := *rootFlag
	openZeppelin := *openZeppelinFlag
	outDir := filepath.Join(root, "../../bind")
	if contracts, err := compile.Compile(openZeppelin,
		filepath.Join(root, "GovStaking.sol"),
		filepath.Join(root, "NCPList.sol"),
	); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_govStaking_abi.go"), gov.CONTRACT_GOV_STAKING); err != nil {
		panic(err)
	} else if err := contracts.BindContracts(pkg, filepath.Join(outDir, "gen_ncpList_abi.go"), gov.CONTRACT_NCP_LIST); err != nil {
		panic(err)
	} else {
		fmt.Println("success!")
	}
}
