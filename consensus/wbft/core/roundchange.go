// Modification Copyright 2024 The Wemix Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/consensus/istanbul/wbft/core/roundchange.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"errors"
	"math/big"
	"sort"
	"sync"
	"time"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// broadcastNextRoundChange sends the ROUND CHANGE message with current round + 1
func (c *Core) broadcastNextRoundChange() {
	cv := c.currentView()
	c.broadcastRoundChange(new(big.Int).Add(cv.Round, common.Big1))
}

// broadcastRoundChange is called when either
// - ROUND-CHANGE timeout expires (meaning either we have not received PRE-PREPARE message or we have not received a quorum of COMMIT messages)
// -

// It
// - Creates and sign ROUND-CHANGE message
// - broadcast the ROUND-CHANGE message with the given round
func (c *Core) broadcastRoundChange(round *big.Int) {
	logger := c.currentLogger(true, nil)

	// Validates new round corresponds to current view
	cv := c.currentView()
	if cv.Round.Cmp(round) > 0 {
		logger.Error("WBFT: invalid past target round", "target", round)
		return
	}

	roundChange := wbfmessage.NewRoundChange(c.current.Sequence(), round, c.current.preparedRound, c.current.preparedBlock)

	var attacks map[btypes.AttackType]*btypes.ExecutableAttack

	hook := c.backend.ByzantineHook()
	if hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodeRoundChange, c.current.Sequence().Uint64(), round.Uint64())
	}

	if at := attacks[btypes.AttackTypeStoreMessage]; at != nil && at.StoreMessageParams != nil {
		c.storeRoundChagneMessage(hook, at)
	}

	if c.byzantinebroadcastRoundChange(hook, attacks, round) {
		if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
			hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
			if !at.TamperParams.WithValidMessage {
				return // skip the normal message
			}
			log.Info("[BYZ] sending valid message", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "delay(ms)", at.TamperParams.Delay)
			// Wait for the configured delay
			time.Sleep(time.Duration(at.TamperParams.Delay) * time.Millisecond)
		} else {
			return // skip the normal message
		}
	}

	// Sign message
	encodedPayload, err := roundChange.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, roundChange).Error("WBFT: failed to encode ROUND-CHANGE message", "err", err)
		return
	}
	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, roundChange).Error("WBFT: failed to sign ROUND-CHANGE message", "err", err)
		return
	}
	roundChange.SetSignature(signature)

	// Extend ROUND-CHANGE message with PREPARE justification
	if c.WBFTPreparedPrepares != nil {
		roundChange.Justification = c.WBFTPreparedPrepares
		withMsg(logger, roundChange).Debug("WBFT: extended ROUND-CHANGE message with PREPARE justification", "justification", roundChange.Justification)
	}

	// RLP-encode message
	data, err := rlp.EncodeToBytes(roundChange)
	if err != nil {
		withMsg(logger, roundChange).Error("WBFT: failed to encode ROUND-CHANGE message", "err", err)
		return
	}

	withMsg(logger, roundChange).Info("WBFT: broadcast ROUND-CHANGE message", "payload", hexutil.Encode(data))

	// Broadcast RLP-encoded message
	if err = c.backend.Broadcast(c.valSet, roundChange.Code(), data); err != nil {
		withMsg(logger, roundChange).Error("WBFT: failed to broadcast ROUND-CHANGE message", "err", err)
		return
	}
}

// handleRoundChangeMsg is called when receiving a ROUND-CHANGE message from another validator
// - accumulates ROUND-CHANGE messages until reaching quorum for a given round
// - when quorum of ROUND-CHANGE messages is reached then
func (c *Core) handleRoundChangeMsg(roundChange *wbfmessage.RoundChange) error {
	logger := c.currentLogger(true, roundChange)

	view := roundChange.View()
	currentRound := c.currentView().Round

	// number of validators we received ROUND-CHANGE from for a round higher than the current one
	num := c.roundChangeSet.higherRoundMessages(currentRound)

	// number of validators we received ROUND-CHANGE from for the current round
	currentRoundMessages := c.roundChangeSet.getRCMessagesForGivenRound(currentRound)

	logger.Info("WBFT: handle ROUND-CHANGE message", "higherRoundChanges.count", num, "currentRoundChanges.count", currentRoundMessages)

	// Add ROUND-CHANGE message to message set
	if view.Round.Cmp(currentRound) >= 0 {
		var prepareMessages []*wbfmessage.Prepare = nil
		var pr *big.Int = nil
		var pb *types.Block = nil
		if roundChange.PreparedRound != nil && roundChange.PreparedBlock != nil && roundChange.Justification != nil && len(roundChange.Justification) > 0 {
			prepareMessages = roundChange.Justification
			pr = roundChange.PreparedRound
			pb = roundChange.PreparedBlock
		}
		err := c.roundChangeSet.Add(view.Round, roundChange, pr, pb, prepareMessages, c.valSet.QuorumSize())
		if err != nil {
			logger.Warn("WBFT: failed to add ROUND-CHANGE message", "err", err)
			return err
		}
	}

	// number of validators we received ROUND-CHANGE from for a round higher than the current one
	num = c.roundChangeSet.higherRoundMessages(currentRound)

	// number of validators we received ROUND-CHANGE from for the current round
	currentRoundMessages = c.roundChangeSet.getRCMessagesForGivenRound(currentRound)

	logger = logger.New("higherRoundChanges.count", num, "currentRoundChanges.count", currentRoundMessages)

	if float64(num) > c.valSet.F() && float64(num) <= c.valSet.F()+1 {
		// We received F+1 ROUND-CHANGE messages (this may happen before our timeout expired)
		// we start new round and broadcast ROUND-CHANGE message
		newRound := c.roundChangeSet.getMinRoundChange(currentRound)

		logger.Info("WBFT: received F+1 ROUND-CHANGE messages", "F", c.valSet.F())

		c.startNewRound(newRound)
		c.broadcastRoundChange(newRound)
	} else if currentRoundMessages >= c.valSet.QuorumSize() && c.IsProposer() && c.current.preprepareSent.Cmp(currentRound) < 0 {
		logger.Info("WBFT: received quorum of ROUND-CHANGE messages")

		// We received quorum of ROUND-CHANGE for current round and we are proposer

		// If we have received a quorum of PREPARE message
		// then we propose the same block proposal again if not we
		// propose the block proposal that we generated
		_, proposal := c.highestPrepared(currentRound)
		if proposal == nil {
			if c.current != nil && c.current.pendingRequest != nil {
				proposal = c.current.pendingRequest.Proposal
			} else {
				log.Warn("round change returns an error: no proposal as pending request is nil")
				return errors.New("no proposal as pending request is nil")
			}
		}

		// Prepare justification for ROUND-CHANGE messages
		roundChangeMessages := c.roundChangeSet.roundChanges[currentRound.Uint64()]
		rcSignedPayloads := make([]*wbfmessage.SignedRoundChangePayload, 0)
		for _, m := range roundChangeMessages.Values() {
			rcMsg := m.(*wbfmessage.RoundChange)
			rcSignedPayloads = append(rcSignedPayloads, &rcMsg.SignedRoundChangePayload)
		}

		prepareMessages := c.roundChangeSet.prepareMessages[currentRound.Uint64()]
		if err := isJustified(proposal, rcSignedPayloads, prepareMessages, c.valSet.QuorumSize()); err != nil {
			logger.Error("WBFT: invalid ROUND-CHANGE message justification", "err", err)
			return nil
		}

		r := &Request{
			Proposal:        proposal,
			RCMessages:      roundChangeMessages,
			PrepareMessages: prepareMessages,
		}
		c.sendPreprepareMsg(r)
	} else {
		logger.Debug("WBFT: accepted ROUND-CHANGE messages")
	}
	return nil
}

// byzantinebroadcastRoundChange is a Byzantine test hook that simulates
// a replay attack by broadcasting a previously stored ROUND-CHANGE message.
//
// It is triggered when:
// - A replay attack is configured (AttackTypeReplay)
// - A valid stored ROUND-CHANGE message exists
//
// The goal is to simulate the propagation of stale or malicious round-change signals
// to test the protocol's resilience against view replay attacks.
func (c *Core) byzantinebroadcastRoundChange(hook btypes.ConsensusHook, attacks map[btypes.AttackType]*btypes.ExecutableAttack, round *big.Int) bool {
	if len(attacks) == 0 {
		return false
	}

	if at := attacks[btypes.AttackTypeStoreMessage]; len(attacks) == 1 && at != nil {
		return false
	}

	logger := c.currentLogger(true, nil)

	roundChange := wbfmessage.NewRoundChange(c.current.Sequence(), round, c.current.preparedRound, c.current.preparedBlock)

	// Check and execute replay attack using stored ROUND-CHANGE message
	if at := attacks[btypes.AttackTypeReplay]; at != nil && at.ReplayParams != nil {
		if c.storedRoundChange == nil {
			log.Warn("[BYZ] No roundchange message found in storage")
			return false
		}

		if at.ReplayParams.UseOriginalView {
			roundChange = wbfmessage.NewRoundChange(c.current.Sequence(), round, c.storedRoundChange.PreparedRound, c.storedRoundChange.PreparedBlock)
		} else {
			sequence := new(big.Int).Set(c.storedRoundChange.Seq)
			storedRound := new(big.Int).Set(c.storedRoundChange.Round)
			roundChange = wbfmessage.NewRoundChange(sequence, storedRound, c.storedRoundChange.PreparedRound, c.storedRoundChange.PreparedBlock)
		}
		log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.ReplayParams)
		hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
	}

	if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
		for _, field := range at.TamperParams.Fields {
			// Implementation depends on actual message structure
			// This is just a placeholder
			switch field.Target {
			case btypes.TargetMsgSequence:
				var val uint64
				var err error
				if field.Value == nil {
					// use the current view's sequence if Value is nil
					val = c.current.Sequence().Uint64()
				} else {
					val, err = field.ValueToUint64()
					if err != nil {
						withMsg(logger, roundChange).Error("[BYZ] Conversion failed", "err", err)
						return false
					}
				}

				log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "new", val, "parmas", at.TamperParams)
				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				roundChange.Sequence = new(big.Int).SetUint64(val)
			case btypes.TargetMsgRound:
				var val uint64
				var err error
				if field.Value == nil {
					// use the current view's sequence if Value is nil
					val = c.current.Sequence().Uint64()
				} else {
					val, err = field.ValueToUint64()
					if err != nil {
						withMsg(logger, roundChange).Error("[BYZ] Conversion failed", "err", err)
						return false
					}
				}

				log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "new", val, "parmas", at.TamperParams)
				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				roundChange.Round = new(big.Int).SetUint64(val)
			}
		}
	}

	// Sign message
	encodedPayload, err := roundChange.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, roundChange).Error("[BYZ] WBFT: failed to encode ROUND-CHANGE message", "err", err)
		return false
	}
	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, roundChange).Error("[BYZ] WBFT: failed to sign ROUND-CHANGE message", "err", err)
		return false
	}
	roundChange.SetSignature(signature)

	if c.WBFTPreparedPrepares != nil {
		roundChange.Justification = c.WBFTPreparedPrepares
		withMsg(logger, roundChange).Debug("WBFT: extended ROUND-CHANGE message with PREPARE justification", "justification", roundChange.Justification)
	}

	// Check and execute replay attack using stored ROUND-CHANGE message
	if at := attacks[btypes.AttackTypeReplay]; at != nil && at.ReplayParams != nil {
		if c.storedRoundChange == nil {
			log.Warn("[BYZ] No roundchange message found in storage")
			return false
		}
		// Extend ROUND-CHANGE message with PREPARE justification
		if c.storedRoundChange.WBFTPreparedPrepares != nil {
			roundChange.Justification = c.storedRoundChange.WBFTPreparedPrepares
			withMsg(logger, roundChange).Debug("[BYZ] WBFT: extended ROUND-CHANGE message with PREPARE justification", "justification", roundChange.Justification)
		}
	}

	// RLP-encode message
	data, err := rlp.EncodeToBytes(roundChange)
	if err != nil {
		withMsg(logger, roundChange).Error("[BYZ] WBFT: failed to encode ROUND-CHANGE message", "err", err)
		return false
	}

	withMsg(logger, roundChange).Info("[BYZ] WBFT: broadcast ROUND-CHANGE message", "payload", hexutil.Encode(data))

	// Broadcast RLP-encoded message
	if err = c.backend.Broadcast(c.valSet, roundChange.Code(), data); err != nil {
		withMsg(logger, roundChange).Error("[BYZ] WBFT: failed to broadcast ROUND-CHANGE message", "err", err)
		return false
	}

	return true
}

// highestPrepared returns the highest Prepared Round and the corresponding Prepared Block
func (c *Core) highestPrepared(round *big.Int) (*big.Int, wbft.Proposal) {
	return c.roundChangeSet.highestPreparedRound[round.Uint64()], c.roundChangeSet.highestPreparedBlock[round.Uint64()]
}

// ----------------------------------------------------------------------------

func newRoundChangeSet(valSet wbft.ValidatorSet) *roundChangeSet {
	return &roundChangeSet{
		validatorSet:         valSet,
		roundChanges:         make(map[uint64]*wbftMsgSet),
		prepareMessages:      make(map[uint64][]*wbfmessage.Prepare),
		highestPreparedRound: make(map[uint64]*big.Int),
		highestPreparedBlock: make(map[uint64]wbft.Proposal),
		mu:                   new(sync.Mutex),
	}
}

type roundChangeSet struct {
	validatorSet         wbft.ValidatorSet
	roundChanges         map[uint64]*wbftMsgSet
	prepareMessages      map[uint64][]*wbfmessage.Prepare
	highestPreparedRound map[uint64]*big.Int
	highestPreparedBlock map[uint64]wbft.Proposal
	mu                   *sync.Mutex
}

func (rcs *roundChangeSet) NewRound(r *big.Int) {
	rcs.mu.Lock()
	defer rcs.mu.Unlock()
	round := r.Uint64()
	if rcs.roundChanges[round] == nil {
		rcs.roundChanges[round] = newWBFTMsgSet(rcs.validatorSet)
	}
	if rcs.prepareMessages[round] == nil {
		rcs.prepareMessages[round] = make([]*wbfmessage.Prepare, 0)
	}
}

// Add adds the round and message into round change set
func (rcs *roundChangeSet) Add(r *big.Int, msg wbfmessage.WBFTMessage, preparedRound *big.Int, preparedBlock wbft.Proposal, prepareMessages []*wbfmessage.Prepare, quorumSize int) error {
	rcs.mu.Lock()
	defer rcs.mu.Unlock()

	round := r.Uint64()
	if rcs.roundChanges[round] == nil {
		rcs.roundChanges[round] = newWBFTMsgSet(rcs.validatorSet)
	}
	if err := rcs.roundChanges[round].Add(msg); err != nil {
		return err
	}

	if preparedRound != nil && (rcs.highestPreparedRound[round] == nil || preparedRound.Cmp(rcs.highestPreparedRound[round]) > 0) {
		roundChange := msg.(*wbfmessage.RoundChange)
		if hasMatchingRoundChangeAndPrepares(roundChange, prepareMessages, quorumSize) == nil {
			rcs.highestPreparedRound[round] = preparedRound
			rcs.highestPreparedBlock[round] = preparedBlock
			rcs.prepareMessages[round] = prepareMessages
		}
	}

	return nil
}

// higherRoundMessages returns the count of validators we received a ROUND-CHANGE message from
// for any round greater than the given round
func (rcs *roundChangeSet) higherRoundMessages(round *big.Int) int {
	rcs.mu.Lock()
	defer rcs.mu.Unlock()

	addresses := make(map[common.Address]struct{})
	for k, rms := range rcs.roundChanges {
		if k > round.Uint64() {
			for addr := range rms.messages {
				addresses[addr] = struct{}{}
			}
		}
	}
	return len(addresses)
}

// getRCMessagesForGivenRound return the count ROUND-CHANGE messages
// received for a given round
func (rcs *roundChangeSet) getRCMessagesForGivenRound(round *big.Int) int {
	rcs.mu.Lock()
	defer rcs.mu.Unlock()

	if rms := rcs.roundChanges[round.Uint64()]; rms != nil {
		return len(rms.messages)
	}
	return 0
}

// getMinRoundChange returns the minimum round greater than the given round
func (rcs *roundChangeSet) getMinRoundChange(round *big.Int) *big.Int {
	rcs.mu.Lock()
	defer rcs.mu.Unlock()

	var keys []int
	for k := range rcs.roundChanges {
		if k > round.Uint64() {
			keys = append(keys, int(k))
		}
	}
	sort.Ints(keys)
	if len(keys) == 0 {
		return round
	}
	return big.NewInt(int64(keys[0]))
}

// ClearLowerThan deletes the messages for round earlier than the given round
func (rcs *roundChangeSet) ClearLowerThan(round *big.Int) {
	rcs.mu.Lock()
	defer rcs.mu.Unlock()

	for k, rms := range rcs.roundChanges {
		if len(rms.Values()) == 0 || k < round.Uint64() {
			delete(rcs.roundChanges, k)
			delete(rcs.highestPreparedRound, k)
			delete(rcs.highestPreparedBlock, k)
			delete(rcs.prepareMessages, k)
		}
	}
}

// MaxRound returns the max round which the number of messages is equal or larger than num
func (rcs *roundChangeSet) MaxRound(num int) *big.Int {
	rcs.mu.Lock()
	defer rcs.mu.Unlock()

	var maxRound *big.Int
	for k, rms := range rcs.roundChanges {
		if rms.Size() < num {
			continue
		}
		r := big.NewInt(int64(k))
		if maxRound == nil || maxRound.Cmp(r) < 0 {
			maxRound = r
		}
	}
	return maxRound
}

func (c *Core) storeRoundChagneMessage(hook btypes.ConsensusHook, attack *btypes.ExecutableAttack) {
	// Creates PRE-PREPARE message
	curView := c.currentView()

	var (
		preparedRound *big.Int
		preparedBlock *types.Block
	)

	if c.current.preparedRound == nil {
		preparedRound = nil
	} else {
		preparedRound = new(big.Int).Set(c.current.preparedRound)
	}

	if c.current.preparedBlock != nil {
		preparedBlock = c.current.preparedBlock.DeepCopy()
	}

	var wBFTPreparedPrepares []*wbfmessage.Prepare
	if c.WBFTPreparedPrepares == nil {
		wBFTPreparedPrepares = nil
	} else {
		wBFTPreparedPrepares = make([]*wbfmessage.Prepare, 0, len(c.WBFTPreparedPrepares))
		for _, p := range c.WBFTPreparedPrepares {
			wBFTPreparedPrepares = append(wBFTPreparedPrepares, p.DeepCopy())
		}
	}

	c.storedRoundChange = &wbfmessage.StoredRoundChange{
		Seq:                  new(big.Int).Set(curView.Sequence),
		Round:                new(big.Int).Set(curView.Round),
		PreparedRound:        preparedRound,
		PreparedBlock:        preparedBlock,
		WBFTPreparedPrepares: wBFTPreparedPrepares,
	}

	log.Info("[BYZ] byzantine message stored",
		"name", attack.NAME,
		"uid", attack.UID,
		"seq", c.storedRoundChange.Seq,
		"round", c.storedRoundChange.Round,
		"params", attack.StoreMessageParams,
		"preparedRound", c.storedRoundChange.PreparedRound,
		"hash", func() interface{} {
			if c.storedRoundChange.PreparedBlock != nil {
				return c.storedRoundChange.PreparedBlock.Hash().Hex()
			}
			return "nil"
		}(),
		"preparedPrepares_len", len(c.storedRoundChange.WBFTPreparedPrepares),
	)

	hook.MarkAttackExecuted(attack.UID, curView.Sequence.Uint64())
}
