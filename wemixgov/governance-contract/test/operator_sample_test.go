package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
	"github.com/stretchr/testify/require"
)

// TESTS

// 1. single operator - multisig 없이 register, add stake, unstake, claim
// 1-1. 각 amount 값 검증
// 2, withdraw - revert when owner tries to withdraw by .cal() or .transfer() or .send()
// 3. withdraw - withdraw when owner tries to submit tx that .cal() or .transfer() or .send()
// 4. withdraw - withdraw when owner tires to withdraw by right function
// 5. withdraw 후 각 amount 값 검증
// 6. addOwner, removeOwner, change Owner 테스트
// 7. multiple operator - multisig 동작 확인. ( 1~ 5) MultiSig 버전으로 확인

func TestOperatorContractSingleOwner(t *testing.T) {
	var (
		operatorContractSingleOwner = getTxOpt(t, "operatorContractOwner")
		minStaking                  = towei(500000)
		feeRate                     = new(big.Int).SetUint64(1500)
		ctx                         = context.Background()
		rewardAmount                = towei(10)
		delegator1                  = NewEOA()
	)
	// initiate gov
	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		operatorContractSingleOwner.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		delegator1.Address:               {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
	})
	require.NoError(t, err)
	defer g.backend.Close()

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	// deploy operatorSample
	owners := []common.Address{operatorContractSingleOwner.From}
	operatorSampleAddr := g.DeployOperatorSample(t, owners, new(big.Int))

	var s1 = NewTestStakerWithOperatorCA(&CA{Address: operatorSampleAddr})

	t.Run("Register Staker", func(t *testing.T) {
		// 1. owner sends value to operator contract
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), operatorContractSingleOwner, minStaking, &operatorSampleAddr))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, operatorSampleAddr, nil), minStaking)

		//2. owner executes registerStaker function of operator contract
		_, err := g.ExpectedOk(g.SingleOwnerRegisterStaker(operatorContractSingleOwner, s1, minStaking, feeRate))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, govwbft.GovStakingAddress, nil), minStaking)
		require.True(t, big.NewInt(0).Cmp(g.balanceAt(t, ctx, operatorSampleAddr, nil)) == 0)
	})

	t.Run("Claim for Reward, restake", func(t *testing.T) {
		// disribute reward manually - 10 ether
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		// restake the reward
		_, err := g.ExpectedOk(g.ClaimViaOperatorContract(operatorContractSingleOwner, s1, true))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, govwbft.GovStakingAddress, nil), new(big.Int).Add(minStaking, rewardAmount))
	})

	var claimedReward *big.Int

	t.Run("Claim for Reward, not restake", func(t *testing.T) {
		// disribute reward manually
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		// transfer reward from rewardee to operator contract
		receipt, err := g.ExpectedOk(g.ClaimViaOperatorContract(operatorContractSingleOwner, s1, false))
		require.NoError(t, err)
		userRewardEvent := findEvents("UserRewardUpdated", receipt.Logs)
		claimedReward = userRewardEvent[0]["pendingReward"].(*big.Int)
		require.Equal(t, userRewardEvent[0]["pendingReward"], g.balanceAt(t, ctx, operatorSampleAddr, nil))
	})

	t.Run("Withdraw reward", func(t *testing.T) {
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		receipt, err := g.ExpectedOk(g.WithdrawRewardAmount(operatorContractSingleOwner, operatorContractSingleOwner.From, claimedReward))
		require.NoError(t, err)
		gasUsed := calcTxGasCost(receipt)
		require.Equal(t, g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil), beforeBalance.Add(beforeBalance, new(big.Int).Sub(claimedReward, gasUsed)))
	})

	var claimedFee *big.Int

	t.Run("Get fee from undelegation", func(t *testing.T) {
		// 1. delegator1 delegates, distribute reward
		_, err := g.ExpectedOk(g.Delegate(t, delegator1, s1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)

		// 2. submit tx to set operatorContract as feeRecipient
		functionSignature := []byte("changeFeeRecipient(address)")
		methodID := crypto.Keccak256(functionSignature)[:4]
		args := abi.Arguments{
			{Type: mustParseType("address")},
		}
		packedArgs, err := args.Pack(operatorSampleAddr)
		if err != nil {
			panic(fmt.Sprintf("Failed to pack arguments: %v", err))
		}
		callData := append(methodID, packedArgs...)

		receipt, err := g.ExpectedOk(g.SubmitTransaction(operatorContractSingleOwner, govwbft.GovStakingAddress, new(big.Int), callData))
		require.NoError(t, err)
		txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

		// 3.confirm the submitted tx
		_, err = g.ExpectedOk(g.ConfirmTransaction(operatorContractSingleOwner, txId))
		require.NoError(t, err)

		// 4. execute the tx
		_, err = g.ExpectedOk(g.ExecuteTransaction(operatorContractSingleOwner, txId))
		require.NoError(t, err)

		// 5. Check if FeeRecipient has changed
		require.Equal(t, operatorSampleAddr, govwbft.StakerInfo(stateDB, s1.Staker.Address).FeeRecipient)

		beforeBalance := g.balanceAt(t, ctx, operatorSampleAddr, nil)
		// 6. Delegator claims reward, fee will be sent to operatorContract
		receipt, err = g.ExpectedOk(g.Claim(t, delegator1, s1.Staker.Address, false))
		pendingFee := findEvents("UserRewardUpdated", receipt.Logs)[0]["pendingFee"].(*big.Int)
		afterBalance := g.balanceAt(t, ctx, operatorSampleAddr, nil)
		require.Equal(t, afterBalance, beforeBalance.Add(beforeBalance, pendingFee))

		// 7. Check if feeAmount tracking works
		callOpts := new(bind.CallOpts)
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&claimedFee}, "feeAmount"))
		require.Equal(t, pendingFee, claimedFee)
	})

	t.Run("Withdraw fee", func(t *testing.T) {
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		// owner withdraw fee
		receipt, err := g.ExpectedOk(g.WithdrawFeeAmount(operatorContractSingleOwner, operatorContractSingleOwner.From, claimedFee))
		require.NoError(t, err)
		gasUsed := calcTxGasCost(receipt)
		afterBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		// check if balance change is valid
		require.Equal(t, afterBalance, beforeBalance.Add(beforeBalance, new(big.Int).Sub(claimedFee, gasUsed)))

		// remaining fee amount in operator contract should be 0
		callOpts := new(bind.CallOpts)
		var feeAmount *big.Int
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&feeAmount}, "feeAmount"))
		require.True(t, feeAmount.Cmp(new(big.Int)) == 0)
	})

	t.Run("Unstake and withdraw undstaked amount", func(t *testing.T) {
		// unstake the staked amount
		stakedAmt := govwbft.UserInfo(stateDB, s1.Staker.Address, operatorSampleAddr).StakingAmount
		receipt, err := g.ExpectedOk(g.SingleOwnerUnstake(operatorContractSingleOwner, stakedAmt))
		require.NoError(t, err)
		unbondingPeriod := findEvents("NewCredential", receipt.Logs)[0]["unbonding"].(*big.Int)
		g.adjustTime(time.Duration(unbondingPeriod.Int64()) * time.Second)

		_, err = g.ExpectedOk(g.WithdrawViaOperatorContract(operatorContractSingleOwner, new(big.Int)))
		require.NoError(t, err)

		callOpts := new(bind.CallOpts)
		var unstakedAmount *big.Int
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&unstakedAmount}, "unstakedAmount"))
		require.Equal(t, unstakedAmount, stakedAmt)

		// withdraw it
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		receipt, err = g.ExpectedOk(g.WithdrawUnstakedAmount(operatorContractSingleOwner, operatorContractSingleOwner.From, unstakedAmount))
		require.NoError(t, err)
		gasUsed := calcTxGasCost(receipt)
		require.Equal(t, g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil), beforeBalance.Add(beforeBalance, new(big.Int).Sub(unstakedAmount, gasUsed)))
	})

	t.Run("Withdraw remaining reward", func(t *testing.T) {
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		receipt, err := g.ExpectedOk(g.ClaimViaOperatorContract(operatorContractSingleOwner, s1, false))
		require.NoError(t, err)
		claimedReward := findEvents("UserRewardUpdated", receipt.Logs)[0]["pendingReward"].(*big.Int)

		callOpts := new(bind.CallOpts)
		var tracedReward *big.Int
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedReward}, "rewardAmount"))
		require.Equal(t, tracedReward, claimedReward)

		// withdraw it
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		receipt, err = g.ExpectedOk(g.WithdrawRewardAmount(operatorContractSingleOwner, operatorContractSingleOwner.From, claimedReward))
		require.NoError(t, err)
		gasUsed := calcTxGasCost(receipt)
		require.Equal(t, g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil), beforeBalance.Add(beforeBalance, new(big.Int).Sub(claimedReward, gasUsed)))
	})

}
