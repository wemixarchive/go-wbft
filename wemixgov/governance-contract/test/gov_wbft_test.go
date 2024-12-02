package test

import (
	"context"
	"math/big"
	"testing"
	"time"

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
		ctx          = context.TODO()
		minStaking   = towei(500000)
		totalStaking = new(big.Int)
		validators   = make([]common.Address, 0)

		v1        = NewTestValidator()
		v2        = NewTestValidator()
		delegator = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		v1.Staker.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		v2.Staker.Address: {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		delegator.Address: {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
	})
	require.NoError(t, err)
	defer g.backend.Close()

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	checkGovBalanceFn := func() {
		require.Equal(t, totalStaking, g.balanceAt(t, ctx, g.gov.GovStaking.Address, nil))
	}

	t.Run("New Validtor", func(t *testing.T) {
		defer checkGovBalanceFn()
		t.Run("add validator", func(t *testing.T) {
			require.True(t, g.gov.TotalStaking(stateDB).Cmp(common.Big0) == 0)
			require.True(t, len(g.gov.Validators(stateDB)) == 0)
			beforeBalance := g.balanceAt(t, ctx, v1.Staker.Address, nil)

			receipt, err := g.ExpectedOk(g.RegisterValidator(t, v1, minStaking))
			require.NoError(t, err)
			validators = append(validators, v1.Validator.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)

			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
			require.Equal(t, validators, g.gov.Validators(stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Staker.Address, nil))
		})

		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t,
					"registerValidator",
					v2.Staker, minStaking,
					new(big.Int).Sub(minStaking, big.NewInt(1)),
					v2.Validator.Address,
					v2.Reward.Address,
				)),
				"amount and msg.value mismatch",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, v2, new(big.Int).Sub(minStaking, big.NewInt(1)))),
				"out of bounds",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, v2, new(big.Int).Add(MAX_UINT_128, big.NewInt(1)))),
				"out of bounds",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, &TestValidator{v2.Staker, v2.Staker, v2.Staker}, minStaking)),
				"staker cannot be validator or reward",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, &TestValidator{&EOA{Address: common.Address{}}, v2.Staker, v2.Validator}, minStaking)),
				"zero address",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, &TestValidator{v2.Validator, v2.Staker, v2.Validator}, minStaking)),
				"validator cannot be reward",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, &TestValidator{v2.Validator, v1.Staker, v2.Reward}, minStaking)),
				"staker is already registered",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, &TestValidator{v1.Staker, v2.Staker, v2.Reward}, minStaking)),
				"validator is already registered",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, &TestValidator{v2.Validator, v2.Staker, v1.Reward}, minStaking)),
				"reward is already registered",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterValidator(t, &TestValidator{v1.Validator, v2.Staker, v2.Reward}, minStaking)),
				"validator exists",
			)
		})

		t.Run("add another validator", func(t *testing.T) {
			require.Equal(t, minStaking, g.gov.TotalStaking(stateDB))
			require.Equal(t, validators, g.gov.Validators(stateDB))
			beforeBalance := g.balanceAt(t, ctx, v2.Staker.Address, nil)

			receipt, err := g.ExpectedOk(g.RegisterValidator(t, v2, minStaking))
			require.NoError(t, err)

			validators = append(validators, v2.Validator.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)

			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
			require.Equal(t, validators, g.gov.Validators(stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v2.Staker.Address, nil))
		})
	})

	t.Run("Stake", func(t *testing.T) {
		defer checkGovBalanceFn()
		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t, "stake", v1.Staker, common.Big2, minStaking)),
				"amount and msg.value mismatch",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Stake(t, delegator, minStaking)),
				"unregistered validator",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Stake(t, v1.Staker, MAX_UINT_128)),
				"exceeded the maximum",
			)
		})

		t.Run("stake more", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, v1.Staker.Address, nil)

			receipt, err := g.ExpectedOk(g.Stake(t, v1.Staker, minStaking))
			require.NoError(t, err)

			totalStaking = totalStaking.Add(totalStaking, minStaking)
			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
			require.Equal(t, new(big.Int).Mul(minStaking, common.Big2), g.gov.ValidatorInfo(stateDB, v1.Validator.Address).Staking)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Staker.Address, nil))
		})
	})

	t.Run("Unstake & Withdraw", func(t *testing.T) {
		defer checkGovBalanceFn()
		var unstakeEvent map[string]interface{}

		t.Run("unstake failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, delegator, minStaking)),
				"unregistered validator",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v2.Staker, common.Big0)),
				"amount is zero",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v2.Staker, new(big.Int).Add(minStaking, common.Big1))),
				"insufficient balance",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v2.Staker, new(big.Int).Sub(minStaking, common.Big1))),
				"amount must equal balance to remove validator",
			)
		})

		t.Run("unstake", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, v1.Staker.Address, nil)

			receipt, err := g.ExpectedOk(g.Unstake(t, v1.Staker, minStaking))
			require.NoError(t, err)

			totalStaking = totalStaking.Sub(totalStaking, minStaking)

			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
			require.Equal(t, minStaking, g.gov.ValidatorInfo(stateDB, v1.Validator.Address).Staking)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, gasCost)
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Staker.Address, nil))

			unstakeEvent = findEvent("NewCredential", receipt.Logs)
			require.NotNil(t, unstakeEvent)
		})

		t.Run("withdraw failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t, "withdraw", v1.Staker, nil, common.Big0)),
				"invalid credential",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, v2.Staker, unstakeEvent["credentialID"].(*big.Int))),
				"msg.sender is not requester",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, v1.Staker, unstakeEvent["credentialID"].(*big.Int))),
				"not yet time to withdraw",
			)
		})

		t.Run("withdraw", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, v1.Staker.Address, nil)

			unbonding := unstakeEvent["unbonding"].(*big.Int)
			g.backend.AdjustTime(time.Duration(unbonding.Int64()) * time.Second)
			receipt, err := g.ExpectedOk(g.Withdraw(t, v1.Staker, unstakeEvent["credentialID"].(*big.Int)))
			require.NoError(t, err)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Staker.Address, nil))
		})

		t.Run("remove validator", func(t *testing.T) {
			{ // unstake
				receipt, err := g.ExpectedOk(g.Unstake(t, v2.Staker, minStaking))
				require.NoError(t, err)

				totalStaking = totalStaking.Sub(totalStaking, minStaking)
				validators = removeValidator(validators, v2.Validator.Address)

				require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
				require.Equal(t, validators, g.gov.Validators(stateDB))
				require.True(t, g.gov.ValidatorInfo(stateDB, v2.Validator.Address).Staking.Cmp(common.Big0) == 0)

				unstakeEvent = findEvent("NewCredential", receipt.Logs)
				require.NotNil(t, unstakeEvent)
			}

			{ // withdraw
				beforeBalance := g.balanceAt(t, ctx, v2.Staker.Address, nil)

				unbonding := unstakeEvent["unbonding"].(*big.Int)
				g.backend.AdjustTime(time.Duration(unbonding.Int64()) * time.Second)
				receipt, err := g.ExpectedOk(g.Withdraw(t, v2.Staker, unstakeEvent["credentialID"].(*big.Int)))
				require.NoError(t, err)

				gasCost := calcTxGasCost(receipt)
				expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
				require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v2.Staker.Address, nil))
			}
		})
	})

	t.Run("Delegate", func(t *testing.T) {
		defer checkGovBalanceFn()
		delegateAmount := towei(100_000)
		t.Run("failure case", func(t *testing.T) {
			{
				_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, minStaking, &v1.Validator.Address))
				require.NoError(t, err)
				_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, minStaking, &v1.Reward.Address))
				require.NoError(t, err)
			}
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, v1.Validator, v1.Validator.Address, delegateAmount)),
				"validator cannot delegate",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, v1.Staker, v1.Validator.Address, delegateAmount)),
				"staker(reward) cannot delegate",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, v1.Reward, v1.Validator.Address, delegateAmount)),
				"staker(reward) cannot delegate",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, delegator, v2.Validator.Address, delegateAmount)),
				"unregistered validator",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, delegator, v1.Validator.Address, MAX_UINT_128)),
				"exceeded the maximum",
			)
		})

		t.Run("delegate", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)
			beforeInfo_v1 := g.gov.ValidatorInfo(stateDB, v1.Validator.Address)

			receipt, err := g.ExpectedOk(g.Delegate(t, delegator, v1.Validator.Address, delegateAmount))
			require.NoError(t, err)

			totalStaking = totalStaking.Add(totalStaking, delegateAmount)
			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))

			afterInfo_v1 := g.gov.ValidatorInfo(stateDB, v1.Validator.Address)
			require.Equal(t, delegateAmount, new(big.Int).Sub(afterInfo_v1.Staking, beforeInfo_v1.Staking))
			require.Equal(t, delegateAmount, new(big.Int).Sub(afterInfo_v1.Delegated, beforeInfo_v1.Delegated))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(delegateAmount, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})
	})

	t.Run("Undelegate & Withdraw", func(t *testing.T) {
		defer checkGovBalanceFn()
		var (
			undelegateEvent  map[string]interface{}
			undelegateAmount = towei(50_000)
		)

		t.Run("undelegate", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)
			beforeInfo_v1 := g.gov.ValidatorInfo(stateDB, v1.Validator.Address)

			receipt, err := g.ExpectedOk(g.Unelegate(t, delegator, v1.Validator.Address, undelegateAmount))
			require.NoError(t, err)

			totalStaking = totalStaking.Sub(totalStaking, undelegateAmount)
			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))

			afterInfo_v1 := g.gov.ValidatorInfo(stateDB, v1.Validator.Address)
			require.Equal(t, undelegateAmount, new(big.Int).Sub(beforeInfo_v1.Staking, afterInfo_v1.Staking))
			require.Equal(t, undelegateAmount, new(big.Int).Sub(beforeInfo_v1.Delegated, afterInfo_v1.Delegated))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, gasCost)
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))

			undelegateEvent = findEvent("NewCredential", receipt.Logs)
			require.NotNil(t, undelegateEvent)
		})

		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, delegator, undelegateEvent["credentialID"].(*big.Int))),
				"not yet time to withdraw",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unelegate(t, delegator, v1.Validator.Address, new(big.Int).Add(undelegateAmount, common.Big1))),
				"insufficient balance",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unelegate(t, delegator, v2.Validator.Address, undelegateAmount)),
				"insufficient balance",
			)

			// try unstake, including the delegated amount
			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v1.Staker, g.gov.ValidatorInfo(stateDB, v1.Validator.Address).Staking)),
				"insufficient balance",
			)
		})

		t.Run("withdraw", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)

			unbonding := undelegateEvent["unbonding"].(*big.Int)
			g.backend.AdjustTime(time.Duration(unbonding.Int64()) * time.Second)
			receipt, err := g.ExpectedOk(g.Withdraw(t, delegator, undelegateEvent["credentialID"].(*big.Int)))
			require.NoError(t, err)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateEvent["amount"].(*big.Int), gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})

		t.Run("undelegate to removed validator", func(t *testing.T) {
			// unstake and remove validator
			{
				receipt, err := g.ExpectedOk(g.Unstake(t, v1.Staker, minStaking))
				require.NoError(t, err)

				totalStaking = totalStaking.Sub(totalStaking, minStaking)
				validators = removeValidator(validators, v1.Validator.Address)

				require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
				require.Equal(t, validators, g.gov.Validators(stateDB))

				unstakeEvent := findEvent("NewCredential", receipt.Logs)
				require.NotNil(t, unstakeEvent)

				beforeBalance := g.balanceAt(t, ctx, v1.Staker.Address, nil)

				unbonding := unstakeEvent["unbonding"].(*big.Int)
				g.backend.AdjustTime(time.Duration(unbonding.Int64()) * time.Second)
				withdrawReceipt, err := g.ExpectedOk(g.Withdraw(t, v1.Staker, unstakeEvent["credentialID"].(*big.Int)))
				require.NoError(t, err)

				gasCost := calcTxGasCost(withdrawReceipt)
				expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
				require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Staker.Address, nil))
			}
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)

			receipt, err := g.ExpectedOk(g.Unelegate(t, delegator, v1.Validator.Address, undelegateAmount))
			require.NoError(t, err)

			totalStaking = totalStaking.Sub(totalStaking, undelegateAmount)
			require.True(t, totalStaking.Cmp(common.Big0) == 0)
			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateAmount, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})
	})
}

func removeValidator(validators []common.Address, validator common.Address) []common.Address {
	for i, v := range validators {
		if v == validator {
			return append(validators[:i], validators[i+1:]...)
		}
	}
	return validators
}
