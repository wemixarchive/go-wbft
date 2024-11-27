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

func Test_gov(t *testing.T) {
	var (
		ctx        = context.Background()
		minStaking = towei(500000)
		validators = make([]common.Address, 0)
	)

	g, err := NewGovWBFT(t)
	require.NoError(t, err)

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}
	v1 := NewTestValidator()
	{
		_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, new(big.Int).Mul(minStaking, big.NewInt(2)), &v1.Staker.Address))
		require.NoError(t, err)
	}
	v2 := NewTestValidator()
	{
		_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, new(big.Int).Mul(minStaking, big.NewInt(2)), &v2.Staker.Address))
		require.NoError(t, err)
	}

	t.Run("New Validtor", func(t *testing.T) {
		require.True(t, g.gov.TotalStaking(stateDB).Cmp(common.Big0) == 0)
		require.Equal(t, validators, g.gov.Validators(stateDB))

		g.FailureCase = true
		_, err := g.NewValidator(t, v1, new(big.Int).Sub(minStaking, big.NewInt(1)))
		t.Log(err)
		ExpectedRevert(t, err, "out of bounds")

		_, err = g.NewValidator(t, v1, minStaking)
		require.NoError(t, err)

		require.NoError(t, err)
		validators = append(validators, v1.Validator.Address)

		require.Equal(t, minStaking, g.gov.TotalStaking(stateDB))
		require.Equal(t, validators, g.gov.Validators(stateDB))

		_, err = g.NewValidator(t, v2, minStaking)
		require.NoError(t, err)

		require.NoError(t, err)
		validators = append(validators, v2.Validator.Address)

		require.Equal(t, new(big.Int).Mul(minStaking, big.NewInt(2)), g.gov.TotalStaking(stateDB))
		require.Equal(t, validators, g.gov.Validators(stateDB))
	})

}

func (g *GovWBFT) NewValidator(t *testing.T, v *TestValidator, amount *big.Int) (*types.Receipt, error) {
	tx, err := g.staking.Transact(
		NewTxOptsWithValue(t, v.Staker, amount),
		"newValidator",
		amount,
		v.Validator.Address,
		v.Reward.Address,
	)

	if g.FailureCase {
		g.FailureCase = false
		return nil, g.ExpectedFail(tx, err)
	}
	return g.ExpectedOk(tx, err)
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
