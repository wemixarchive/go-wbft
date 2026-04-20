// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Regression tests for GovStaking security fixes. Each test intentionally
// triggers the attack precondition and asserts the contract enforces the
// correct outcome post-fix. Side-effect preservation tests assert the
// normal/legitimate paths still succeed.
//
// Scenarios covered:
//
//   * claim() — operator can only claim on behalf of the staker they are
//     registered for; cross-staker drains must fail.
//   * claim() — operator's own delegation reward on another staker remains
//     claimable after the scope fix.
//   * withdraw() — shortening the unbonding period via governance must not
//     let users drain earlier, still-locked credentials.
//   * withdraw() — when a single account accumulates staker credentials and
//     delegator credentials with different unbonding periods, the older
//     locked credential must not be drained via the bulk path.
//   * withdraw() — auto mode preserves the sequential-lock policy (stops at
//     the first not-yet-mature credential).
//   * withdraw() — explicit-count mode succeeds when every credential in the
//     range is mature, and rejects partial-mature ranges.
//   * receive() — only a registered active rewardee vault may send coin to
//     GovStaking; direct EOA sends revert.
//
// Related documents:
//   - docs/secrity-review/fix-report-govstaking-security.md
//   - docs/secrity-review/govstaking-logic-flow.md

package test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// claim() — operator permission scope
// -----------------------------------------------------------------------------

// TestClaim_OperatorCannotDrainAnotherStakersReward asserts that an operator
// cannot claim rewards from a staker they are not registered as the operator
// for. The previous implementation used `isOperator(msg.sender)`, which
// returned true for any registered operator and allowed cross-staker drains.
// The corrected implementation uses `stakerByOperator[msg.sender] == _staker`
// so that operator-on-behalf-of-staker claims are scoped per-staker.
func TestClaim_OperatorCannotDrainAnotherStakersReward(t *testing.T) {
	var (
		ctx          = context.TODO()
		feeRate      = new(big.Int).SetUint64(1000)
		minStaking   = towei(500000)
		rewardAmount = towei(10)

		sX = NewTestStaker() // victim staker (operator OpX)
		sY = NewTestStaker() // attacker's own staker (operator OpY)
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		sY.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) common.Hash {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	// Register both stakers; each pays its own minStaking from its operator's balance.
	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.RegisterStaker(t, sY, minStaking, feeRate))
	require.NoError(t, err)

	// Accrue reward on victim sX (send coin to sX's rewardee vault).
	distributeReward(t, g, stateDB, rewardAmount, sX.Staker.Address)

	// Precondition: OpY has no stake or delegation in sX.
	stakerOfOpY := govwbft.StakerByOperator(TestGovStakingAddress, stateDB, sY.Operator.Address)
	require.Equal(t, sY.Staker.Address, stakerOfOpY, "OpY must only be registered as operator of sY")

	t.Run("cross-staker claim must revert", func(t *testing.T) {
		// stakerByOperator[OpY] == sY.Staker != sX.Staker therefore
		// _user = msg.sender = OpY. OpY has neither stakingAmount nor
		// pendingReward under sX, so the "no reward to claim" guard fires.
		err := g.ExpectedFail(g.Claim(t, sY.Operator, sX.Staker.Address, false))
		ExpectedRevert(t, err, "no reward to claim")

		// Victim's pending reward must remain claimable by the legitimate
		// operator (proves no drain happened).
		beforeOpX := g.balanceAt(t, ctx, sX.Operator.Address, nil)
		receipt, err := g.ExpectedOk(g.Claim(t, sX.Operator, sX.Staker.Address, false))
		require.NoError(t, err)
		afterOpX := g.balanceAt(t, ctx, sX.Operator.Address, nil)
		gasCost := calcTxGasCost(receipt)
		// OpX gets the full reward minus gas.
		expectedDelta := new(big.Int).Sub(rewardAmount, gasCost)
		require.Equal(t, expectedDelta, new(big.Int).Sub(afterOpX, beforeOpX),
			"legitimate operator receives reward net of gas")
	})
}

// TestClaim_OperatorCanClaimOwnDelegationOnAnotherStaker asserts that an
// operator who also holds a legitimate delegation to a different staker can
// still claim their own delegation rewards. The scope fix must not break the
// cross-staker delegation case where the call is legitimate (operator is
// acting as a plain delegator on a staker they do not operate).
func TestClaim_OperatorCanClaimOwnDelegationOnAnotherStaker(t *testing.T) {
	var (
		ctx          = context.TODO()
		feeRate      = new(big.Int).SetUint64(1000)
		minStaking   = towei(500000)
		delegateAmt  = towei(100000)
		rewardAmount = towei(10)

		sX = NewTestStaker() // operator OpX; OpX will delegate to sY
		sY = NewTestStaker() // operator OpY
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		sY.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) common.Hash {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.RegisterStaker(t, sY, minStaking, feeRate))
	require.NoError(t, err)

	// OpX (operator of sX) delegates to sY. Allowed because
	// delegate() only forbids self-delegation by sY's own operator.
	_, err = g.ExpectedOk(g.Delegate(t, sX.Operator, sY.Staker.Address, delegateAmt))
	require.NoError(t, err)

	// Accrue reward on sY; OpX's delegation now has pending reward.
	distributeReward(t, g, stateDB, rewardAmount, sY.Staker.Address)

	// OpX calls claim(sY, false). stakerByOperator[OpX]=sX.Staker, compared
	// against _staker=sY.Staker, so _user = msg.sender = OpX. OpX has
	// stakingAmount > 0 in userRewardInfo[sY][OpX] → claim succeeds and pays
	// OpX their delegator reward.
	beforeOpX := g.balanceAt(t, ctx, sX.Operator.Address, nil)
	receipt, err := g.ExpectedOk(g.Claim(t, sX.Operator, sY.Staker.Address, false))
	require.NoError(t, err)
	afterOpX := g.balanceAt(t, ctx, sX.Operator.Address, nil)

	// OpX's net balance must increase (some reward minus gas).
	gasCost := calcTxGasCost(receipt)
	require.True(t, afterOpX.Cmp(new(big.Int).Sub(beforeOpX, gasCost)) > 0,
		"OpX received at least some delegation reward net of gas")
}

// -----------------------------------------------------------------------------
// withdraw() — governance-shortened unbonding period
// -----------------------------------------------------------------------------

// TestWithdraw_RejectsDrainAfterUnbondingPeriodShortened asserts that
// shortening the staker unbonding period via governance after credentials are
// already in-flight does not let the earlier (still-locked) credential be
// drained via the bulk path `withdraw(_withdrawalCount > 0)`. The previous
// implementation only checked the last credential's withdrawableTime; the
// corrected implementation enforces per-credential expiry inside the loop.
func TestWithdraw_RejectsDrainAfterUnbondingPeriodShortened(t *testing.T) {
	var (
		feeRate    = new(big.Int).SetUint64(1000)
		minStaking = towei(500000)

		sX = NewTestStaker()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g) // unbondingPeriodStaker = 604800s (7d)
	defer g.backend.Close()

	// Register + top up so there is something to unstake while staying active.
	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)
	topUp := towei(10)
	_, err = g.ExpectedOk(g.Stake(t, sX.Operator, topUp))
	require.NoError(t, err)

	// credential[0] with the ORIGINAL long period.
	_, err = g.ExpectedOk(g.Unstake(t, sX.Operator, towei(5)))
	require.NoError(t, err)

	// Governance shortens unbondingPeriodStaker to 1 hour.
	const shortPeriod int64 = 3600
	g.backend.CommitWithState(&params.GovContracts{
		GovConfig: &params.GovContract{
			Address: TestGovConfigAddress,
			Version: govwbft.GOV_CONTRACT_VERSION_1,
			Params: map[string]string{
				govwbft.GOV_CONFIG_PARAM_MINIMUM_STAKING:     minStaking.String(),
				govwbft.GOV_CONFIG_PARAM_MAXIMUM_STAKING:     MAX_UINT_128.String(),
				govwbft.GOV_CONFIG_PARAM_UNBONDING_STAKER:    big.NewInt(shortPeriod).String(),
				govwbft.GOV_CONFIG_PARAM_UNBONDING_DELEGATOR: "259200",
				govwbft.GOV_CONFIG_PARAM_FEE_PRECISION:       "10000",
				govwbft.GOV_CONFIG_PARAM_CHANGE_FEE_DELAY:    "604800",
			},
		},
	}, nil)

	// credential[1] with the SHORTER period, created after the governance change.
	_, err = g.ExpectedOk(g.Unstake(t, sX.Operator, towei(1)))
	require.NoError(t, err)

	// Advance past the short period — credential[1] is mature, credential[0] is not.
	g.adjustTime(time.Duration(shortPeriod+60) * time.Second)

	t.Run("explicit bulk withdraw must revert", func(t *testing.T) {
		// The later, short-period credential matured first, but the earlier
		// long-period credential is still locked. Per-credential expiry
		// enforcement inside the loop must abort the whole call.
		err := g.ExpectedFail(g.Withdraw(t, sX.Operator, common.Big2))
		ExpectedRevert(t, err, "withdrawal time not reached")
	})

	t.Run("auto withdraw breaks at first immature credential", func(t *testing.T) {
		// credential[0] is still locked, so withdraw(0) must break
		// immediately without returning any amount and without advancing
		// withdrawalIndex. This is the intentional sequential-lock policy:
		// the user must wait for credential[0] to mature before any later
		// credential can be withdrawn.
		receipt, err := g.ExpectedOk(g.Withdraw(t, sX.Operator, common.Big0))
		require.NoError(t, err)
		require.Empty(t, findEvents("Withdrawn", receipt.Logs),
			"no Withdrawn event when first credential is immature")
	})
}

// -----------------------------------------------------------------------------
// withdraw() — staker/delegator mixed credentials
// -----------------------------------------------------------------------------

// TestWithdraw_RejectsDrainOnMixedStakerDelegatorCredentials reproduces the
// dual-role condition: the same EOA is the operator of one staker and a
// delegator to a different staker. Because unbondingPeriodStaker is longer
// than unbondingPeriodDelegator in standard deployments, credential[0] (from
// unstake) matures *later* than credential[1] (from undelegate). The previous
// last-credential-only expiry check allowed draining credential[0] early; the
// corrected per-credential check must revert on any locked credential inside
// the requested range.
func TestWithdraw_RejectsDrainOnMixedStakerDelegatorCredentials(t *testing.T) {
	var (
		feeRate     = new(big.Int).SetUint64(1000)
		minStaking  = towei(500000)
		delegateAmt = towei(50000)

		sX = NewTestStaker() // operator OpX plays dual role (unstake + undelegate)
		sY = NewTestStaker() // another staker; OpX delegates here
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		sY.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g) // staker=604800 (7d), delegator=259200 (3d)
	defer g.backend.Close()

	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.RegisterStaker(t, sY, minStaking, feeRate))
	require.NoError(t, err)

	// Add some stake so we can unstake a portion without deactivating.
	_, err = g.ExpectedOk(g.Stake(t, sX.Operator, towei(20)))
	require.NoError(t, err)

	// credential[OpX][0]: from unstake → unbondingPeriodStaker (7d)
	_, err = g.ExpectedOk(g.Unstake(t, sX.Operator, towei(10)))
	require.NoError(t, err)

	// OpX delegates to sY, then undelegates → credential[OpX][1]: unbondingPeriodDelegator (3d)
	_, err = g.ExpectedOk(g.Delegate(t, sX.Operator, sY.Staker.Address, delegateAmt))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.Undelegate(t, sX.Operator, sY.Staker.Address, delegateAmt))
	require.NoError(t, err)

	// Advance 3d + slack → credential[1] mature, credential[0] still locked.
	g.adjustTime(time.Duration(259200+60) * time.Second)

	t.Run("explicit bulk withdraw must revert", func(t *testing.T) {
		err := g.ExpectedFail(g.Withdraw(t, sX.Operator, common.Big2))
		ExpectedRevert(t, err, "withdrawal time not reached")
	})

	t.Run("auto withdraw must not advance past the locked credential[0]", func(t *testing.T) {
		receipt, err := g.ExpectedOk(g.Withdraw(t, sX.Operator, common.Big0))
		require.NoError(t, err)
		require.Empty(t, findEvents("Withdrawn", receipt.Logs),
			"sequential policy: credential[0] must gate credential[1] until it matures")
	})

	t.Run("after credential[0] matures, explicit withdraw succeeds", func(t *testing.T) {
		// Wait the remaining 4 days to mature credential[0] (7d total for staker period).
		g.adjustTime(time.Duration(4*24*3600) * time.Second)

		receipt, err := g.ExpectedOk(g.Withdraw(t, sX.Operator, common.Big2))
		require.NoError(t, err)

		events := findEvents("Withdrawn", receipt.Logs)
		require.Len(t, events, 2, "both mature credentials drained in order")
	})
}

// -----------------------------------------------------------------------------
// Side-effect preservation — withdraw()
// -----------------------------------------------------------------------------

// TestWithdraw_ExplicitAllMatureStillWorks verifies that the post-fix
// withdraw() remains fully backward-compatible when every credential in the
// requested range is mature.
func TestWithdraw_ExplicitAllMatureStillWorks(t *testing.T) {
	var (
		feeRate    = new(big.Int).SetUint64(1000)
		minStaking = towei(500000)

		sX = NewTestStaker()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.Stake(t, sX.Operator, towei(30)))
	require.NoError(t, err)

	// Three sequential unstakes — all with the same unbondingPeriodStaker, so
	// withdrawableTime is monotonically increasing.
	for i := 0; i < 3; i++ {
		_, err = g.ExpectedOk(g.Unstake(t, sX.Operator, towei(10)))
		require.NoError(t, err)
	}

	// Advance beyond the staker unbonding period.
	g.adjustTime(time.Duration(604800+60) * time.Second)

	receipt, err := g.ExpectedOk(g.Withdraw(t, sX.Operator, big.NewInt(3)))
	require.NoError(t, err)
	require.Len(t, findEvents("Withdrawn", receipt.Logs), 3)
}

// TestWithdraw_AutoModeMixedMaturity verifies that auto mode (count==0) stops
// at the first immature credential and does not drain later ones — the
// intentional sequential-lock policy that sits alongside the per-credential
// expiry check.
func TestWithdraw_AutoModeMixedMaturity(t *testing.T) {
	var (
		feeRate    = new(big.Int).SetUint64(1000)
		minStaking = towei(500000)

		sX = NewTestStaker()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.Stake(t, sX.Operator, towei(20)))
	require.NoError(t, err)

	// Two unstakes.
	_, err = g.ExpectedOk(g.Unstake(t, sX.Operator, towei(10)))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.Unstake(t, sX.Operator, towei(5)))
	require.NoError(t, err)

	// Advance past first credential only (use the staker period / 2 plus
	// account for blocks between the two unstakes — but both have 7d period
	// so they mature together; instead use a small partial advance that
	// leaves both immature, confirming auto mode returns nothing).
	g.adjustTime(time.Hour)

	receipt, err := g.ExpectedOk(g.Withdraw(t, sX.Operator, common.Big0))
	require.NoError(t, err)
	require.Empty(t, findEvents("Withdrawn", receipt.Logs),
		"auto mode yields nothing when nothing is mature yet")

	// Now mature both.
	g.adjustTime(time.Duration(604800+60) * time.Second)
	receipt, err = g.ExpectedOk(g.Withdraw(t, sX.Operator, common.Big0))
	require.NoError(t, err)
	require.Len(t, findEvents("Withdrawn", receipt.Logs), 2,
		"auto mode drains all mature credentials")
}

// TestReceive_OnlyActiveRewardeeAllowed covers the GovStaking `receive()`
// guard: only an active staker's rewardee vault may send coin to the staking
// contract (used by the internal restake flow). Direct sends by an EOA or
// other contract must revert, preventing balance inflation attacks.
func TestReceive_OnlyActiveRewardeeAllowed(t *testing.T) {
	var (
		feeRate    = new(big.Int).SetUint64(1000)
		minStaking = towei(500000)
		sX         = NewTestStaker()
		attacker   = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		attacker.Address:    {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)

	// Direct coin send from an EOA must revert.
	govStakingAddr := TestGovStakingAddress
	tx, txErr := TransferCoin(
		g.backend.Client(),
		NewTxOptsWithValue(t, attacker, nil),
		big.NewInt(1),
		&govStakingAddr,
	)
	// TransferCoin may fail at gas estimation (pre-submit revert) or at commit
	// time — either way, expectedFail normalizes both into a revert error.
	_, err = expectedFail(g.backend, tx, txErr)
	ExpectedRevert(t, err, "only an active rewardee can send coin")
}

// TestWithdraw_OutOfRangeStillRejected is a regression check for the
// `_lastIndex <= credentialIndex` guard that was NOT removed by the fix.
func TestWithdraw_OutOfRangeStillRejected(t *testing.T) {
	var (
		feeRate    = new(big.Int).SetUint64(1000)
		minStaking = towei(500000)
		sX         = NewTestStaker()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		sX.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	_, err = g.ExpectedOk(g.RegisterStaker(t, sX, minStaking, feeRate))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.Stake(t, sX.Operator, towei(10)))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.Unstake(t, sX.Operator, towei(5)))
	require.NoError(t, err)

	// count=5 but only 1 credential exists — must revert.
	err = g.ExpectedFail(g.Withdraw(t, sX.Operator, big.NewInt(5)))
	ExpectedRevert(t, err, "out of max user credential index")
}
