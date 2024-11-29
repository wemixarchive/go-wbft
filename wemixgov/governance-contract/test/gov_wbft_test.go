package test

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

type TestStateDB struct {
	getState func(addr common.Address, hash common.Hash) common.Hash
}

func (db *TestStateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	return db.getState(addr, hash)
}

func TestGovWithoutNCP(t *testing.T) {
	var (
		ctx        = context.Background()
		minStaking = towei(500000)
		validators = make([]common.Address, 0)
	)

	g, err := NewGovWBFT(t, nil)
	require.NoError(t, err)

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}
	v1 := NewTestValidator()
	{
		_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, new(big.Int).Mul(minStaking, big.NewInt(3)), &v1.Staker.Address))
		require.NoError(t, err)
	}
	v2 := NewTestValidator()
	{
		_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, new(big.Int).Mul(minStaking, big.NewInt(3)), &v2.Staker.Address))
		require.NoError(t, err)
	}

	t.Run("New Validtor", func(t *testing.T) {
		t.Run("add validator", func(t *testing.T) {
			require.True(t, g.gov.TotalStaking(stateDB).Cmp(common.Big0) == 0)
			require.Equal(t, validators, g.gov.Validators(stateDB))

			_, err = g.ExpectedOk(g.newValidatorTx(t, v1, minStaking))
			require.NoError(t, err)
			validators = append(validators, v1.Validator.Address)

			require.Equal(t, minStaking, g.gov.TotalStaking(stateDB))
			require.Equal(t, validators, g.gov.Validators(stateDB))
		})

		t.Run("failure case", func(t *testing.T) {
			err := g.ExpectedFail(g.newValidatorTx(t, v2, new(big.Int).Sub(minStaking, big.NewInt(1))))
			ExpectedRevert(t, err, "out of bounds")

			err = g.ExpectedFail(g.newValidatorTx(t, &TestValidator{v2.Staker, v2.Staker, v2.Staker}, minStaking))
			ExpectedRevert(t, err, "staker cannot be validator or reward")

			err = g.ExpectedFail(g.newValidatorTx(t, &TestValidator{v2.Validator, v2.Staker, v2.Validator}, minStaking))
			ExpectedRevert(t, err, "validator cannot be reward")

			err = g.ExpectedFail(g.newValidatorTx(t, &TestValidator{v2.Validator, v1.Staker, v2.Reward}, minStaking))
			ExpectedRevert(t, err, "staker is already registered")

			err = g.ExpectedFail(g.newValidatorTx(t, &TestValidator{v1.Validator, v2.Staker, v2.Reward}, minStaking))
			ExpectedRevert(t, err, "validator exists")
		})

		t.Run("add another validator", func(t *testing.T) {
			require.Equal(t, minStaking, g.gov.TotalStaking(stateDB))
			require.Equal(t, validators, g.gov.Validators(stateDB))

			_, err = g.ExpectedOk(g.newValidatorTx(t, v2, minStaking))
			require.NoError(t, err)
			validators = append(validators, v2.Validator.Address)

			require.Equal(t, new(big.Int).Mul(minStaking, common.Big2), g.gov.TotalStaking(stateDB))
			require.Equal(t, validators, g.gov.Validators(stateDB))
		})
	})
}

func (g *GovWBFT) newValidatorTx(t *testing.T, v *TestValidator, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContract.Transact(
		NewTxOptsWithValue(t, v.Staker, amount),
		"newValidator",
		amount,
		v.Validator.Address,
		v.Reward.Address,
	)
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
