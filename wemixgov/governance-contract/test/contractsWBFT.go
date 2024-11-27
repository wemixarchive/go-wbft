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
	backend *simulated.WbftBackend
	owner   *bind.TransactOpts
	gov     *govWBFT.GovernanceWBFT
	staking *bind.BoundContract
	ncpList *bind.BoundContract

	FailureCase bool
}

func NewGovWBFT(t *testing.T) (*GovWBFT, error) {
	owner := getTxOpt(t, "owner")
	g := &GovWBFT{
		owner: owner,
		backend: simulated.NewWbftBackend(types.GenesisAlloc{
			owner.From: {Balance: new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 128), common.Big1)},
		}),
	}

	stakingAddr, staking, err := g.Deploy(compiledWBFT.GovStaking.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)
	_, ncpList, err := g.Deploy(compiledWBFT.NCPList.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)

	g.gov = govWBFT.NewGovernanceWBFT(stakingAddr)
	g.staking = staking
	g.ncpList = ncpList

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
