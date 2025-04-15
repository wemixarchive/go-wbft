package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

func TestOperatorSampleDeploy(t *testing.T) {
	var (
		operatorContractSingleOwner = getTxOpt(t, "operatorContractOwner")
	)
	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		operatorContractSingleOwner.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	defer g.backend.Close()

	owners := []common.Address{operatorContractSingleOwner.From}
	operatorSampleAddr := g.DeployOperatorSample(t, owners, new(big.Int))
	t.Log(operatorSampleAddr)
}

func TestOperatorContractSingleOwner(t *testing.T) {
	var (
		operatorContractSingleOwner = getTxOpt(t, "operatorContractOwner")
		minStaking                  = towei(500000)
		feeRate                     = new(big.Int).SetUint64(1500)
		ctx                         = context.Background()
		rewardAmount                = towei(10)
	)
	// initiate gov
	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		operatorContractSingleOwner.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
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
		// disribute reward manually
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		_, err := g.ExpectedOk(g.ClaimViaOperatorContract(operatorContractSingleOwner, s1, true))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, govwbft.GovStakingAddress, nil), new(big.Int).Add(minStaking, rewardAmount))
		fmt.Println(govwbft.StakerInfo(stateDB, s1.Staker.Address).AccRewardPerStaking)
		fmt.Println(govwbft.StakerInfo(stateDB, s1.Staker.Address).LastRewardBalance)
	})

	t.Run("Claim for Reward, not restake", func(t *testing.T) {
		// disribute reward manually
		fmt.Println(govwbft.StakerInfo(stateDB, s1.Staker.Address).AccRewardPerStaking)
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		_, err := g.ExpectedOk(g.ClaimViaOperatorContract(operatorContractSingleOwner, s1, false))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, operatorSampleAddr, nil), rewardAmount)

	})

	t.Run("Withdraw reward", func(t *testing.T) {

	})

	t.Run("Delegators add stake", func(t *testing.T) {

	})

	t.Run("Change Fee Recipient", func(t *testing.T) {
		// change fee recipeint to this contract

	})

	t.Run("Delegator undelegates", func(t *testing.T) {

	})

	t.Run("Withdraw fee", func(t *testing.T) {

	})

	t.Run("Untake and withdraw undstaked amount", func(t *testing.T) {

	})

}
