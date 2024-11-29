package test

import (
	"math/big"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	compile "github.com/ethereum/go-ethereum/wemixgov/governance-contract"
	govWBFT "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
	"github.com/stretchr/testify/require"
)

var (
	compiledWBFT compiledContractWBFT
)

func init() {
	compiledWBFT.Compile("../contracts-wbft", "../contracts")
}

type compiledContractWBFT struct {
	GovStaking,
	NCPList *bindContract
}

func (c *compiledContractWBFT) Compile(root, openzeppelinPath string) {
	if contracts, err := compile.Compile(openzeppelinPath,
		filepath.Join(root, "GovStaking.sol"),
		filepath.Join(root, "NCPList.sol"),
	); err != nil {
		panic(err)
	} else {
		if c.GovStaking, err = newBindContract(contracts["GovStaking"]); err != nil {
			panic(err)
		} else if c.NCPList, err = newBindContract(contracts["NCPList"]); err != nil {
			panic(err)
		}
	}
}

type GovWBFT struct {
	backend         *simulated.WbftBackend
	owner           *bind.TransactOpts
	gov             *govWBFT.Governance
	stakingContract *bind.BoundContract
	ncpListContract *bind.BoundContract
}

func NewGovWBFT(t *testing.T, ncpList []common.Address) (*GovWBFT, error) {
	owner := getTxOpt(t, "owner")
	g := &GovWBFT{
		owner: owner,
		backend: simulated.NewWbftBackend(types.GenesisAlloc{
			owner.From: {Balance: new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 128), common.Big1)},
		}),
	}

	stakingAddr, stakingContract, err := g.Deploy(compiledWBFT.GovStaking.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)
	ncpAddr, ncpListContract, err := g.Deploy(compiledWBFT.NCPList.Deploy(g.backend.Client(), g.owner, ncpList))
	require.NoError(t, err)

	g.gov = govWBFT.NewGovernance(stakingAddr, ncpAddr)
	g.stakingContract = stakingContract
	g.ncpListContract = ncpListContract

	return g, nil

}

func (g *GovWBFT) Deploy(address common.Address, tx *types.Transaction, contract *bind.BoundContract, txErr error) (common.Address, *bind.BoundContract, error) {
	if txErr != nil {
		return common.Address{}, nil, txErr
	}
	_, err := g.ExpectedOk(tx, txErr)
	return address, contract, err
}

func (g *GovWBFT) ExpectedOk(tx *types.Transaction, txErr error) (*types.Receipt, error) {
	return expectedOk(g.backend, tx, txErr)
}

func (g *GovWBFT) ExpectedFail(tx *types.Transaction, txErr error) error {
	_, err := expectedFail(g.backend, tx, txErr)
	return err
}
