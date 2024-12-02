package test

import (
	"context"
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

func NewGovWBFT(t *testing.T, ncpList []common.Address, alloc types.GenesisAlloc) (*GovWBFT, error) {
	if alloc == nil {
		alloc = make(types.GenesisAlloc)
	}
	owner := getTxOpt(t, "owner")
	alloc[owner.From] = types.Account{Balance: MAX_UINT_128}
	g := &GovWBFT{
		owner:   owner,
		backend: simulated.NewWbftBackend(alloc),
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

func (g *GovWBFT) RegisterValidator(t *testing.T, v *TestValidator, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "registerValidator", v.Staker, amount, amount, v.Validator.Address, v.Reward.Address)
}

func (g *GovWBFT) Stake(t *testing.T, staker *EOA, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "stake", staker, amount, amount)
}

func (g *GovWBFT) Unstake(t *testing.T, staker *EOA, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "unstake", staker, nil, amount)
}

func (g *GovWBFT) Delegate(t *testing.T, delegator *EOA, validator common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "delegate", delegator, amount, validator, amount)
}

func (g *GovWBFT) Unelegate(t *testing.T, delegator *EOA, validator common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "undelegate", delegator, nil, validator, amount)
}

func (g *GovWBFT) Withdraw(t *testing.T, sender *EOA, credentialID *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "withdraw", sender, nil, credentialID)
}

func (g *GovWBFT) stakingContractTx(t *testing.T, method string, sender *EOA, value *big.Int, params ...interface{}) (*types.Transaction, error) {
	return g.stakingContract.Transact(NewTxOptsWithValue(t, sender, value), method, params...)
}

func (g *GovWBFT) balanceAt(t *testing.T, ctx context.Context, addr common.Address, num *big.Int) *big.Int {
	balance, err := g.backend.Client().BalanceAt(ctx, addr, num)
	require.NoError(t, err)

	return balance
}

type TestValidator struct {
	Validator *EOA
	Staker    *EOA
	Reward    *EOA
}

func NewTestValidator() *TestValidator {
	return &TestValidator{
		Validator: NewEOA(),
		Staker:    NewEOA(),
		Reward:    NewEOA(),
	}
}
