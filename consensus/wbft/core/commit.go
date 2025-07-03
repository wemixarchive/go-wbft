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
	"time"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
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

	if hook := c.backend.ByzantineHook(); hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodeCommit, c.current.Sequence().Uint64(), c.current.Round().Uint64())
	}

	if c.broadcastByzantineCommit(attacks) {
		if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
			if !at.TamperParams.WithValidMessage {
				return // skip the normal message
			}
			// Wait for the configured delay
			time.Sleep(time.Duration(at.TamperParams.Delay) * time.Millisecond)
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

	if at := attacks[btypes.AttackTypeSilentMessage]; at != nil && at.SilentParams != nil {
		if at.SilentParams.Direction == 1 || at.SilentParams.Direction == 3 {
			log.Info("[BYZ] attack silent: blocked outgoing message", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "round", c.current.Round().Uint64(), "msgCode", btypes.MessageCodeCommit)
			return
		}
	}

	withMsg(logger, commit).Info("WBFT: broadcast COMMIT message", "payload", hexutil.Encode(payload))

	// Broadcast RLP-encoded message
	if err = c.backend.Broadcast(c.valSet, commit.Code(), payload); err != nil {
		withMsg(logger, commit).Error("WBFT: failed to broadcast COMMIT message", "err", err)
		return
	}
}

func (c *Core) broadcastByzantineCommit(attacks map[btypes.AttackType]*btypes.ExecutableAttack) bool {
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
	commit.SetSource(c.Address())

	if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
		for _, field := range at.TamperParams.TamperFields {
			// Implementation depends on actual message structure
			// This is just a placeholder
			switch field.Target {
			case btypes.TamperDigest:
				val, err := field.ValueToHash()
				if err != nil {
					withMsg(logger, commit).Error("[BYZ] Conversion failed", "err", err)
					return false
				} else {
					send = true
					log.Info("[BYZ] attack tamper: Digest", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "round", c.current.Round().Uint64(), "msgCode", btypes.MessageCodeCommit, "original", commit.Digest.Hex(), "changed", val.Hex())
					commit.Digest = val
				}
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
			withMsg(logger, commit).Error("[BYZ] WBFT: failed to broadcast COMMIT message", "err", err)
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

	// Check digest
	if commit.Digest != c.current.Proposal().Hash() {
		logger.Error("WBFT: invalid COMMIT message digest", "digest", commit.Digest, "proposal", c.current.Proposal().Hash().String())
		return errInvalidMessage
	}

	// Check commitSeal
	block, ok := c.current.Proposal().(*types.Block)
	if !ok {
		logger.Error("WBFT: failed to cast proposal from COMMIT message to *types.Block")
		return errInvalidMessage
	}

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

		// Byzantine hook: Allow modification of seals before commit
		if hook := c.backend.ByzantineHook(); hook != nil {
			// Convert to interface slices for hook
			preparedInterface := make([]interface{}, len(preparedSeals))
			committedInterface := make([]interface{}, len(committedSeals))
			for i, seal := range preparedSeals {
				preparedInterface[i] = seal
			}
			for i, seal := range committedSeals {
				committedInterface[i] = seal
			}

			// Call hook
			modPrepared, modCommitted, err := hook.BeforeBlockCommit(proposal, preparedInterface, committedInterface)
			if err == nil {
				// Convert back to SealData
				preparedSeals = make([]wbft.SealData, len(modPrepared))
				for i, seal := range modPrepared {
					if s, ok := seal.(wbft.SealData); ok {
						preparedSeals[i] = s
					}
				}
				committedSeals = make([]wbft.SealData, len(modCommitted))
				for i, seal := range modCommitted {
					if s, ok := seal.(wbft.SealData); ok {
						committedSeals[i] = s
					}
				}
			}
		}

		// Commit proposal to database
		if err := c.backend.Commit(proposal, preparedSeals, committedSeals, c.currentView().Round); err != nil {
			c.currentLogger(true, nil).Error("WBFT: error committing proposal", "err", err)
			c.broadcastNextRoundChange()
			return
		}
	}
}
