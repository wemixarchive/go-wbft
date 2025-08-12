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
// This file is derived from quorum/consensus/istanbul/wbft/core/commit.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"encoding/hex"
	"math/big"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// broadcastCommit is called when receiving quorum of PREPARE message

// It
// - creates a COMMIT message from current proposal
// - broadcast COMMIT message to other validators
func (c *Core) broadcastCommit() {
	var err error

	logger := c.currentLogger(true, nil)

	sub := c.current.Subject()

	var header *types.Header
	if block, ok := c.current.Proposal().(*types.Block); ok {
		header = block.Header()
	}

	// Create Commit Seal
	commitSeal := c.backend.SignWithoutHashing(PrepareSeal(header, uint32(c.currentView().Round.Uint64()), SealTypeCommit))
	commit := wbfmessage.NewCommit(sub.View.Sequence, sub.View.Round, sub.Digest, commitSeal)
	commit.SetSource(c.Address())

	var attacks map[btypes.AttackType]*btypes.ExecutableAttack

	hook := c.backend.ByzantineHook()
	if hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodeCommit, c.current.Sequence().Uint64(), c.current.Round().Uint64())
	}

	if at := attacks[btypes.AttackTypeStoreMessage]; at != nil && at.StoreMessageParams != nil {
		c.storeCommitMessage(hook, at, commitSeal)
	}

	// Check for DOS attack
	if at := attacks[btypes.AttackTypeDos]; at != nil && at.DosParams != nil {
		c.executeDosAttack(at, btypes.MessageCodeCommit, commit)
		log.Info("[BYZ] attack", "name", at.NAME, "uid", at.UID,
			"seq", c.current.Sequence().Uint64(), "parmas", at.DosParams)
		hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
		// Continue with normal commit after DOS attack
		return
	}

	if c.broadcastByzantineCommit(hook, attacks) {
		if !c.handleMessagePolicyAttack(hook, attacks) {
			return // skip the original message
		}
	}

	// Sign Message
	encodedPayload, err := commit.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, commit).Error("WBFT: failed to encode payload of COMMIT message", "err", err)
		return
	}

	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, commit).Error("WBFT: failed to sign COMMIT message", "err", err)
		return
	}
	commit.SetSignature(signature)

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&commit)
	if err != nil {
		withMsg(logger, commit).Error("WBFT: failed to encode COMMIT message", "err", err)
		return
	}

	modifiedValSet := c.valSet.Copy()
	if c.IsExecuteAttack(attacks, btypes.AttackTypeMessagePolicy) {
		isExistSpecificTargets := false
		isExecuteDropMessage := false
		messagePolicyParams := attacks[btypes.AttackTypeMessagePolicy].MessagePolicyParams
		for _, field := range messagePolicyParams.Fields {
			switch field.Target {
			case btypes.TargetMsgPolicyTargets:
				targetList := field.Value.([]string)
				for _, v := range targetList {
					if v != "" {
						// specific address
						validators := modifiedValSet.List()
						for _, validator := range validators {
							if !messagePolicyParams.IsTargetPeer(validator.Address().String()) {
								modifiedValSet.RemoveValidator(validator.Address())
							}
						}
					}
				}

				if modifiedValSet.Size() != 0 && modifiedValSet.Size() != c.valSet.Size() {
					isExistSpecificTargets = true
				}
			case btypes.TargetMsgPolicyDirection:
				if v, err := field.ValueToUint64(); err != nil {
					log.Error("[BYZ] Failed to parse field value", "err", err)
				} else {
					if v == uint64(btypes.MessageDirectionSend) || v == uint64(btypes.MessageDirectionBoth) {
						isExecuteDropMessage = true
					}
				}
			}
		}
		if isExecuteDropMessage {
			log.Info("[BYZ] byzantine attack triggered",
				"name", attacks[btypes.AttackTypeMessagePolicy].NAME,
				"uid", attacks[btypes.AttackTypeMessagePolicy].UID,
				"seq", c.CurrentView().Sequence.Uint64(),
				"params", messagePolicyParams)
			c.MarkAttackExecuted(attacks[btypes.AttackTypeMessagePolicy].UID, c.CurrentView().Sequence.Uint64())
			if isExistSpecificTargets {
				targetList := modifiedValSet.List()
				for _, target := range targetList {
					c.valSet.RemoveValidator(target.Address())
				}
			} else {
				// NOTE:
				// If a message drop exists and no specific targets exist,
				// the message drop should be performed for all targets.
				return
			}
		}
	}

	withMsg(logger, commit).Info("WBFT: broadcast COMMIT message", "payload", hexutil.Encode(payload))

	// Broadcast RLP-encoded message
	if err = c.backend.Broadcast(c.valSet, commit.Code(), payload); err != nil {
		withMsg(logger, commit).Error("WBFT: failed to broadcast COMMIT message", "err", err)
		return
	}
}

func (c *Core) broadcastByzantineCommit(hook btypes.ConsensusHook, attacks map[btypes.AttackType]*btypes.ExecutableAttack) bool {
	if len(attacks) == 0 {
		return false
	}
	send := false
	var err error
	logger := c.currentLogger(true, nil)
	sub := c.current.Subject()

	var header *types.Header
	if block, ok := c.current.Proposal().(*types.Block); ok {
		header = block.Header()
	}

	// Create Commit Seal
	commitSeal := c.backend.SignWithoutHashing(PrepareSeal(header, uint32(c.currentView().Round.Uint64()), SealTypeCommit))
	commit := wbfmessage.NewCommit(sub.View.Sequence, sub.View.Round, sub.Digest, commitSeal)

	if at := attacks[btypes.AttackTypeReplay]; at != nil && at.ReplayParams != nil {
		if c.storedCommit != nil {
			send = true
			storedCommitSeal := make([]byte, len(c.storedCommit.CommitSeal))
			copy(storedCommitSeal, c.storedCommit.CommitSeal)

			if at.ReplayParams.UseOriginalView {
				commit = wbfmessage.NewCommit(sub.View.Sequence, sub.View.Round, c.storedCommit.Digest, storedCommitSeal)
			} else {
				sequence := new(big.Int).Set(c.storedCommit.Seq)
				round := new(big.Int).Set(c.storedCommit.Round)
				commit = wbfmessage.NewCommit(sequence, round, c.storedCommit.Digest, storedCommitSeal)
			}
			log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.ReplayParams, "origianl_seal", hex.EncodeToString(commitSeal), "changed_seal", hex.EncodeToString(storedCommitSeal))
			hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
		} else {
			log.Warn("[BYZ] No commit message found in storage")
		}
	}
	commit.SetSource(c.Address())

	if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
		for _, field := range at.TamperParams.Fields {
			// Implementation depends on actual message structure
			// This is just a placeholder
			switch field.Target {
			case btypes.TargetMsgDigest:
				var val common.Hash
				var err error
				if field.Value == nil {
					// use the current Digest if Value is nil
					val = sub.Digest
				} else {
					val, err = field.ValueToHash()
					if err != nil {
						withMsg(logger, commit).Error("[BYZ] Conversion failed", "err", err)
						return false
					}
				}
				send = true
				commit.Digest = val
				log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "original", sub.Digest.Hex(), "changed", commit.Digest.Hex(), "params", at.TamperParams)
				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
			}
		}
	}

	// Sign Message
	encodedPayload, err := commit.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, commit).Error("[BYZ] WBFT: failed to encode payload of COMMIT message", "err", err)
		return false
	}

	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, commit).Error("[BYZ] WBFT: failed to sign COMMIT message", "err", err)
		return false
	}
	commit.SetSignature(signature)

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&commit)
	if err != nil {
		withMsg(logger, commit).Error("[BYZ] WBFT: failed to encode COMMIT message", "err", err)
		return false
	}

	if send {
		withMsg(logger, commit).Info("[BYZ] WBFT: broadcast COMMIT message", "payload", hexutil.Encode(payload))
		// Broadcast RLP-encoded message
		if err = c.backend.Broadcast(c.valSet, commit.Code(), payload); err != nil {
			withMsg(logger, commit).Error("[BYZ] failed to broadcast COMMIT message", "err", err)
			return false
		}
	}
	return true && send
}

// handleCommitMsg is called when receiving a COMMIT message from another validator

// It
// - validates COMMIT message digest matches the current block proposal
// - accumulates valid COMMIT messages until reaching quorum
// - when quorum of COMMIT messages is reached then update state and commits
func (c *Core) handleCommitMsg(commit *wbfmessage.Commit) error {
	logger := c.currentLogger(true, commit)

	logger.Info("WBFT: handle COMMIT message", "commits.count", c.current.WBFTCommits.Size(), "quorum", c.valSet.QuorumSize())

	var attacks map[btypes.AttackType]*btypes.ExecutableAttack
	hook := c.backend.ByzantineHook()
	sequence := c.current.Sequence()
	round := c.current.Round()
	if hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodeCommit, sequence.Uint64(), round.Uint64())
	}

	modifiedValSet := c.valSet.Copy()
	if c.IsExecuteAttack(attacks, btypes.AttackTypeMessagePolicy) {
		isExistSpecificTargets := false
		isExecuteDropMessage := false
		messagePolicyParams := attacks[btypes.AttackTypeMessagePolicy].MessagePolicyParams
		for _, field := range messagePolicyParams.Fields {
			switch field.Target {
			case btypes.TargetMsgPolicyTargets:
				targetList := field.Value.([]string)
				for _, v := range targetList {
					if v != "" {
						// specific address
						validators := modifiedValSet.List()
						for _, validator := range validators {
							if !messagePolicyParams.IsTargetPeer(validator.Address().String()) {
								modifiedValSet.RemoveValidator(validator.Address())
							}
						}
					}
				}

				if modifiedValSet.Size() != 0 && modifiedValSet.Size() != c.valSet.Size() {
					isExistSpecificTargets = true
				}
			case btypes.TargetMsgPolicyDirection:
				if v, err := field.ValueToUint64(); err != nil {
					log.Error("[BYZ] Failed to parse field value", "err", err)
				} else {
					if v == uint64(btypes.MessageDirectionReceive) || v == uint64(btypes.MessageDirectionBoth) {
						isExecuteDropMessage = true
					}
				}
			}
		}
		if isExecuteDropMessage {
			log.Info("[BYZ] byzantine attack triggered",
				"name", attacks[btypes.AttackTypeMessagePolicy].NAME,
				"uid", attacks[btypes.AttackTypeMessagePolicy].UID,
				"seq", sequence.Uint64(),
				"params", messagePolicyParams)
			c.MarkAttackExecuted(attacks[btypes.AttackTypeMessagePolicy].UID, sequence.Uint64())
			if isExistSpecificTargets {
				targetList := modifiedValSet.List()
				for _, target := range targetList {
					if target.Address() == commit.Source() {
						return nil
					}
				}
				log.Info("[BYZ] modified", "valset", c.valSet.AddressList())
			} else {
				// NOTE:
				// If a message drop exists and no specific targets exist,
				// the message drop should be performed for all targets.
				return nil
			}
		}
	}

	// Check digest
	if commit.Digest != c.current.Proposal().Hash() {
		logger.Error("WBFT: invalid COMMIT message digest", "digest", commit.Digest, "proposal", c.current.Proposal().Hash().String())
		return errInvalidMessage
	}

	block, ok := c.current.Proposal().(*types.Block)
	if !ok {
		logger.Error("WBFT: failed to cast proposal from COMMIT message to *types.Block")
		return errInvalidMessage
	}

	// Check commitSeal
	if verifySeal(c.valSet, block.Header(), uint32(commit.CommonPayload.Round.Uint64()), SealTypeCommit,
		commit.CommitSeal, commit.Source()) != nil {
		logger.Error("WBFT: failed to verify seal from COMMIT message", "from", commit.Source())
		return errInvalidMessage
	}

	// Add to received msgs
	if err := c.current.WBFTCommits.Add(commit); err != nil {
		c.logger.Error("WBFT: failed to save COMMIT message", "err", err)
		return err
	}

	logger = logger.New("commits.count", c.current.WBFTCommits.Size(), "quorum", c.valSet.QuorumSize())

	// If we reached threshold
	if c.current.WBFTCommits.Size() >= c.valSet.QuorumSize() {
		logger.Info("WBFT: received quorum of COMMIT messages")
		c.commitWBFT()
	} else {
		logger.Debug("WBFT: accepted new COMMIT messages")
	}

	return nil
}

// commitWBFT is called once quorum of commits is reached
// - computes committedSeals from each received commit messages
// - then commits block proposal to database with committed seals
// - broadcast round change
func (c *Core) commitWBFT() {
	c.setState(StateCommitted)

	proposal := c.current.Proposal()
	if proposal != nil {
		// Compute prepared seals
		preparedSeals := make([]wbft.SealData, c.current.WBFTPrepares.Size())
		for i, msg := range c.current.WBFTPrepares.Values() {
			idx, _ := c.valSet.GetByAddress(msg.Source())
			if idx < 0 {
				continue
			}
			preparedSeals[i] = wbft.SealData{
				Sealer: uint32(idx),
				Seal:   make([]byte, types.IstanbulExtraSeal),
			}
			prepareMsg := msg.(*wbfmessage.Prepare)
			copy(preparedSeals[i].Seal[:], prepareMsg.PrepareSeal[:])
		}

		// Compute committed seals
		committedSeals := make([]wbft.SealData, c.current.WBFTCommits.Size())
		for i, msg := range c.current.WBFTCommits.Values() {
			idx, _ := c.valSet.GetByAddress(msg.Source())
			if idx < 0 {
				continue
			}
			committedSeals[i] = wbft.SealData{
				Sealer: uint32(idx),
				Seal:   make([]byte, types.IstanbulExtraSeal),
			}
			commitMsg := msg.(*wbfmessage.Commit)
			copy(committedSeals[i].Seal[:], commitMsg.CommitSeal[:])
		}

		// Commit proposal to database
		if err := c.backend.Commit(proposal, preparedSeals, committedSeals, c.currentView().Round); err != nil {
			c.currentLogger(true, nil).Error("WBFT: error committing proposal", "err", err)
			c.broadcastNextRoundChange()
			return
		}
	}
}

func (c *Core) storeCommitMessage(hook btypes.ConsensusHook, attack *btypes.ExecutableAttack, CommitSeal []byte) {
	// Create PREPARE message from the current proposal
	sub := c.current.Subject()

	c.storedCommit = &wbfmessage.StoredCommit{
		Seq:        new(big.Int).Set(sub.View.Sequence),
		Round:      new(big.Int).Set(sub.View.Round),
		Digest:     sub.Digest,
		CommitSeal: make([]byte, len(CommitSeal)),
	}
	copy(c.storedCommit.CommitSeal, CommitSeal)

	log.Info("[BYZ] byzantine message stored",
		"name", attack.NAME,
		"uid", attack.UID,
		"seq", c.storedCommit.Seq,
		"round", c.storedCommit.Round,
		"params", attack.StoreMessageParams,
		"Digest", c.storedCommit.Digest.Hex(),
		"CommitSeal", c.storedCommit.CommitSeal)

	hook.MarkAttackExecuted(attack.UID, sub.View.Sequence.Uint64())
}
