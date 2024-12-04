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
			require.True(t, g.gov.TotalStaking(stateDB).Sign() == 0)
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
				validators = removeElement(validators, v2.Validator.Address)

				require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
				require.Equal(t, validators, g.gov.Validators(stateDB))
				require.True(t, g.gov.ValidatorInfo(stateDB, v2.Validator.Address).Staking.Sign() == 0)

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
				validators = removeElement(validators, v1.Validator.Address)

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
			require.True(t, totalStaking.Sign() == 0)
			require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateAmount, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})
	})
}

var (
	ProposalType_None       = common.Big0
	ProposalType_NCPAdd     = common.Big1
	ProposalType_NCPRemoval = common.Big2

	Voting_Period = time.Duration(604800) * time.Second
)

func TestGovWithNCP(t *testing.T) {
	var (
		ctx             = context.TODO()
		minStaking      = towei(500000)
		totalStaking    = new(big.Int)
		ncpTotalStaking = new(big.Int)
		validators      = make([]common.Address, 0)
		ncps            = make([]common.Address, 0)
		ncpValidators   = make([]common.Address, 0)

		ncp1 = NewTestValidator()
		ncp2 = NewTestValidator()
		ncp3 = NewTestValidator()
		ncp4 = NewTestValidator()
	)

	ncps = append(ncps, ncp1.Staker.Address, ncp2.Staker.Address)
	g, err := NewGovWBFT(t, ncps, types.GenesisAlloc{
		ncp1.Staker.Address: {Balance: MAX_UINT_128},
		ncp2.Staker.Address: {Balance: MAX_UINT_128},
		ncp3.Staker.Address: {Balance: MAX_UINT_128},
		ncp4.Staker.Address: {Balance: MAX_UINT_128},
	})
	require.NoError(t, err)

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	checkNCPValidator := func() {
		require.Equal(t, totalStaking, g.gov.TotalStaking(stateDB))
		require.Equal(t, validators, g.gov.Validators(stateDB))
		require.Equal(t, ncps, g.gov.NCPList(stateDB))
		require.Equal(t, ncpTotalStaking, g.gov.NCPTotalStaking(stateDB))
		require.Equal(t, ncpValidators, g.gov.NCPValidators(stateDB))
	}

	t.Run("deployment failure", func(t *testing.T) {
		_, _, _, err := compiledWBFT.GovNCP.Deploy(g.backend.Client(), g.owner, []common.Address{})
		ExpectedRevert(t, err, "at least one ncp required")
	})

	t.Run("NCP Staking", func(t *testing.T) {
		require.True(t, g.gov.TotalStaking(stateDB).Sign() == 0)
		require.True(t, g.gov.NCPTotalStaking(stateDB).Sign() == 0)
		require.Equal(t, validators, g.gov.Validators(stateDB))
		require.Equal(t, ncps, g.gov.NCPList(stateDB))
		require.Equal(t, ncpValidators, g.gov.NCPValidators(stateDB))

		t.Run("NCP staking", func(t *testing.T) {
			defer checkNCPValidator()
			_, err := g.ExpectedOk(g.RegisterValidator(t, ncp1, minStaking))
			require.NoError(t, err)

			validators = append(validators, ncp1.Validator.Address)
			ncpValidators = append(ncpValidators, ncp1.Validator.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)
			ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, minStaking)
		})

		t.Run("non-NCP staking", func(t *testing.T) {
			defer checkNCPValidator()
			_, err := g.ExpectedOk(g.RegisterValidator(t, ncp3, minStaking))
			require.NoError(t, err)

			validators = append(validators, ncp3.Validator.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)
		})

		t.Run("stake more", func(t *testing.T) {
			defer checkNCPValidator()
			// ncp stake more
			{
				_, err := g.ExpectedOk(g.Stake(t, ncp1.Staker, minStaking))
				require.NoError(t, err)

				totalStaking = totalStaking.Add(totalStaking, minStaking)
				ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, minStaking)
			}

			// non-ncp stake more
			{
				_, err := g.ExpectedOk(g.Stake(t, ncp3.Staker, minStaking))
				require.NoError(t, err)
				totalStaking = totalStaking.Add(totalStaking, minStaking)
			}
		})
	})

	t.Run("Add NCP", func(t *testing.T) {
		var proposalEvent map[string]interface{}

		t.Run("new proposal to add ncp", func(t *testing.T) {
			defer checkNCPValidator()

			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
			require.NoError(t, err)

			proposalEvent = findEvent("NewProposal", receipt.Logs)
			require.Equal(t, ProposalType_NCPAdd, proposalEvent["proposalType"].(*big.Int))
		})

		t.Run("failure case", func(t *testing.T) {
			defer checkNCPValidator()

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp4.Staker, ncp4.Staker.Address)),
				"msg.sender is not ncp",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp1.Staker, ncp2.Staker.Address)),
				"ncp exists",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp1.Staker, ncp4.Staker.Address)),
				"previous vote is in progress",
			)
		})
		t.Run("vote & add ncp", func(t *testing.T) {
			defer checkNCPValidator()

			_, err := g.ExpectedOk(g.Vote(t, ncp1.Staker, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			// Vote not finalized
			checkNCPValidator()

			_, err = g.ExpectedOk(g.Vote(t, ncp2.Staker, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			ncps = append(ncps, ncp3.Staker.Address)
			ncpValidators = append(ncpValidators, ncp3.Validator.Address)
			ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, g.gov.ValidatorInfo(stateDB, ncp3.Validator.Address).Staking)
		})
	})

	t.Run("Remove NCP", func(t *testing.T) {
		var proposalEvent map[string]interface{}

		t.Run("new proposal to remove ncp", func(t *testing.T) {
			defer checkNCPValidator()

			receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp1.Staker, ncp3.Staker.Address))
			require.NoError(t, err)

			proposalEvent = findEvent("NewProposal", receipt.Logs)
			require.Equal(t, ProposalType_NCPRemoval, proposalEvent["proposalType"].(*big.Int))
		})

		t.Run("failure case", func(t *testing.T) {
			defer checkNCPValidator()

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToRemoveNCP(t, ncp1.Staker, ncp4.Staker.Address)),
				"invalid ncp",
			)
		})
		t.Run("vote & remove ncp", func(t *testing.T) {
			defer checkNCPValidator()

			_, err := g.ExpectedOk(g.Vote(t, ncp1.Staker, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			// not finalized
			checkNCPValidator()

			_, err = g.ExpectedOk(g.Vote(t, ncp2.Staker, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			ncps = removeElement(ncps, ncp3.Staker.Address)
			ncpValidators = removeElement(ncpValidators, ncp3.Validator.Address)
			ncpTotalStaking = ncpTotalStaking.Sub(ncpTotalStaking, g.gov.ValidatorInfo(stateDB, ncp3.Validator.Address).Staking)
		})
	})

	t.Run("Cancel Proposal", func(t *testing.T) {
		defer checkNCPValidator()

		t.Run("cancel by proposer", func(t *testing.T) {
			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			ExpectedRevert(t,
				g.ExpectedFail(g.CancelProposal(t, ncp2.Staker, proposalEvent["id"].(*big.Int))),
				"cannot cancel",
			)

			receipt, err = g.ExpectedOk(g.CancelProposal(t, ncp1.Staker, proposalEvent["id"].(*big.Int)))
			require.NoError(t, err)
			require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])
		})

		t.Run("canceled due to timeout", func(t *testing.T) {
			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			ExpectedRevert(t,
				g.ExpectedFail(g.CancelProposal(t, ncp2.Staker, proposalEvent["id"].(*big.Int))),
				"cannot cancel",
			)

			g.backend.AdjustTime(Voting_Period)

			receipt, err = g.ExpectedOk(g.CancelProposal(t, ncp2.Staker, proposalEvent["id"].(*big.Int)))
			require.NoError(t, err)
			require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])
		})

		t.Run("timeout & new proposal", func(t *testing.T) {
			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address)),
				"previous vote is in progress",
			)

			g.backend.AdjustTime(Voting_Period)

			receipt, err = g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
			require.NoError(t, err)
			require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])

			// cancel for next test
			{
				proposalEvent := findEvent("NewProposal", receipt.Logs)
				receipt, err := g.ExpectedOk(g.CancelProposal(t, ncp1.Staker, proposalEvent["id"].(*big.Int)))
				require.NoError(t, err)
				require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])
			}
		})
	})

	t.Run("Vote", func(t *testing.T) {
		t.Run("failure case", func(t *testing.T) {
			defer checkNCPValidator()

			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			_, err = g.ExpectedOk(g.Vote(t, ncp1.Staker, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			ExpectedRevert(t,
				g.ExpectedFail(g.Vote(t, ncp1.Staker, proposalEvent["id"].(*big.Int), true)),
				"already voted",
			)

			g.backend.AdjustTime(Voting_Period)

			ExpectedRevert(t,
				g.ExpectedFail(g.Vote(t, ncp2.Staker, proposalEvent["id"].(*big.Int), true)),
				"already closed vote",
			)
		})

		t.Run("majority", func(t *testing.T) {
			t.Run("2 ncp", func(t *testing.T) {
				t.Run("reject", func(t *testing.T) {
					defer checkNCPValidator()

					// 1 NCP is required for reject
					receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp1.Staker, proposalEvent["id"].(*big.Int), false))
					require.NoError(t, err)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, false, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
				t.Run("accept", func(t *testing.T) {
					defer checkNCPValidator()

					// 2 NCP is required for accept
					receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Staker, ncp3.Staker.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					_, err = g.ExpectedOk(g.Vote(t, ncp1.Staker, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp2.Staker, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					ncps = append(ncps, ncp3.Staker.Address)
					ncpValidators = append(ncpValidators, ncp3.Validator.Address)
					ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, g.gov.ValidatorInfo(stateDB, ncp3.Validator.Address).Staking)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, true, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
			})
			t.Run("3 ncp", func(t *testing.T) {
				t.Run("reject", func(t *testing.T) {
					defer checkNCPValidator()

					// 2 NCP is required for reject
					receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp2.Staker, ncp3.Staker.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					_, err = g.ExpectedOk(g.Vote(t, ncp1.Staker, proposalEvent["id"].(*big.Int), false))
					require.NoError(t, err)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp2.Staker, proposalEvent["id"].(*big.Int), false))
					require.NoError(t, err)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, false, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
				t.Run("accept", func(t *testing.T) {
					defer checkNCPValidator()

					// 2 NCP is required for accept
					receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp3.Staker, ncp3.Staker.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					_, err = g.ExpectedOk(g.Vote(t, ncp2.Staker, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp3.Staker, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					ncps = removeElement(ncps, ncp3.Staker.Address)
					ncpValidators = removeElement(ncpValidators, ncp3.Validator.Address)
					ncpTotalStaking = ncpTotalStaking.Sub(ncpTotalStaking, g.gov.ValidatorInfo(stateDB, ncp3.Validator.Address).Staking)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, true, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
			})
		})
	})
}

func removeElement(slice []common.Address, value common.Address) []common.Address {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
