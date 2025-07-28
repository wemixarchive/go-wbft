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
// This file is derived from quorum/consensus/istanbul/wbft/core/prepare.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"encoding/hex"
	"math/big"
	"time"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// broadcastPrepare is called after receiving PRE-PREPARE from proposer node

// It
// - creates a PREPARE message
// - broadcast PREPARE message to other validators
func (c *Core) broadcastPrepare() {
	logger := c.currentLogger(true, nil)

	// Create PREPARE message from the current proposal
	sub := c.current.Subject()

	var header *types.Header
	if block, ok := c.current.Proposal().(*types.Block); ok {
		header = block.Header()
	}

	// Create Prepare Seal
	prepareSeal := c.backend.SignWithoutHashing(PrepareSeal(header, uint32(c.currentView().Round.Uint64()), SealTypePrepare))
	prepare := wbfmessage.NewPrepare(sub.View.Sequence, sub.View.Round, sub.Digest, prepareSeal)
	prepare.SetSource(c.Address())

	var attacks map[btypes.AttackType]*btypes.ExecutableAttack

	hook := c.backend.ByzantineHook()
	if hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodePrepare, c.current.Sequence().Uint64(), c.current.Round().Uint64())
	}

	if at := attacks[btypes.AttackTypeStoreMessage]; at != nil && at.StoreMessageParams != nil {
		c.storePrepareMessage(hook, at, prepareSeal)
	}

	// Check for DOS attack
	if at := attacks[btypes.AttackTypeDos]; at != nil && at.DosParams != nil {
		c.executeDosAttack(at, btypes.MessageCodePrepare, prepare)
		log.Info("[BYZ] attack", "name", at.NAME, "uid", at.UID,
			"seq", c.current.Sequence().Uint64(), "parmas", at.DosParams)
		hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
		// Continue with normal prepare after DOS attack
		return
	}

	if c.broadcastByzantinePrepare(hook, attacks) {
		if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
			if !at.TamperParams.WithValidMessage {
				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				return // skip the normal message
			}
			log.Info("[BYZ] sending valid message", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "delay(ms)", at.TamperParams.Delay)
			// Wait for the configured delay
			time.Sleep(time.Duration(at.TamperParams.Delay) * time.Millisecond)
		} else {
			return // skip the normal message
		}
	}

	// Sign Message
	encodedPayload, err := prepare.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to encode payload of PREPARE message", "err", err)
		return
	}
	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to sign PREPARE message", "err", err)
		return
	}
	prepare.SetSignature(signature)

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&prepare)
	if err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to encode PREPARE message", "err", err)
		return
	}

	if at := attacks[btypes.AttackTypeSilentMessage]; at != nil && at.SilentParams != nil {
		if at.SilentParams.Direction == uint64(btypes.MessageDirectionSend) {
			log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "params", at.SilentParams)
			hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
			return
		}
	}

	withMsg(logger, prepare).Info("WBFT: broadcast PREPARE message", "payload", hexutil.Encode(payload))

	// Broadcast RLP-encoded message
	if err = c.backend.Broadcast(c.valSet, prepare.Code(), payload); err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to broadcast PREPARE message", "err", err)
		return
	}
}

func (c *Core) broadcastByzantinePrepare(hook btypes.ConsensusHook, attacks map[btypes.AttackType]*btypes.ExecutableAttack) bool {
	if len(attacks) == 0 {
		return false
	}
	send := false
	logger := c.currentLogger(true, nil)
	// Create PREPARE message from the current proposal
	sub := c.current.Subject()

	var header *types.Header
	if block, ok := c.current.Proposal().(*types.Block); ok {
		header = block.Header()
	}

	// Create Prepare Seal
	prepareSeal := c.backend.SignWithoutHashing(PrepareSeal(header, uint32(c.currentView().Round.Uint64()), SealTypePrepare))
	prepare := wbfmessage.NewPrepare(sub.View.Sequence, sub.View.Round, sub.Digest, prepareSeal)

	if at := attacks[btypes.AttackTypeReplay]; at != nil && at.ReplayParams != nil {
		if c.storedPrepare != nil {
			send = true
			storedPrepareSeal := make([]byte, len(c.storedPrepare.PrepareSeal))
			copy(storedPrepareSeal, c.storedPrepare.PrepareSeal)

			if at.ReplayParams.UseOriginalView {
				prepare = wbfmessage.NewPrepare(sub.View.Sequence, sub.View.Round, c.storedPrepare.Digest, storedPrepareSeal)
			} else {
				sequence := new(big.Int).Set(c.storedPrepare.Seq)
				round := new(big.Int).Set(c.storedPrepare.Round)
				prepare = wbfmessage.NewPrepare(sequence, round, c.storedPrepare.Digest, storedPrepareSeal)
			}
			log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.ReplayParams, "origianl_seal", hex.EncodeToString(prepareSeal), "changed_seal", hex.EncodeToString(storedPrepareSeal))
			hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
		} else {
			log.Warn("[BYZ] No prepare message found in storage")
		}
	}
	prepare.SetSource(c.Address())
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
						withMsg(logger, prepare).Error("[BYZ] Conversion failed", "err", err)
						return false
					}
				}
				send = true
				prepare.Digest = val
				log.Info("[BYZ] byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "original", sub.Digest.Hex(), "changed", prepare.Digest.Hex(), "params", at.TamperParams)
				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
			}
		}
	}

	// Sign Message
	encodedPayload, err := prepare.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, prepare).Error("[BYZ] WBFT: failed to encode payload of PREPARE message", "err", err)
		return false
	}
	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, prepare).Error("[BYZ] WBFT: failed to sign PREPARE message", "err", err)
		return false
	}
	prepare.SetSignature(signature)

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&prepare)
	if err != nil {
		withMsg(logger, prepare).Error("[BYZ] WBFT: failed to encode PREPARE message", "err", err)
		return false
	}

	if send {
		withMsg(logger, prepare).Info("[BYZ] broadcast PREPARE message", "payload", hexutil.Encode(payload))
		// Broadcast RLP-encoded message
		if err = c.backend.Broadcast(c.valSet, prepare.Code(), payload); err != nil {
			withMsg(logger, prepare).Error("[BYZ] WBFT: failed to broadcast PREPARE message", "err", err)
			return false
		}
	}

	return true && send
}

// handlePrepareMsg is called when receiving a PREPARE message

// It
// - validates PREPARE message digest matches the current block proposal
// - accumulates valid PREPARE message until reaching quorum
// - when quorum is reached update states to "Prepared" and broadcast COMMIT
func (c *Core) handlePrepareMsg(prepare *wbfmessage.Prepare) error {
	logger := c.currentLogger(true, prepare).New()

	logger.Info("WBFT: handle PREPARE message", "prepares.count", c.current.WBFTPrepares.Size(), "quorum", c.valSet.QuorumSize())

	// Check digest
	if prepare.Digest != c.current.Proposal().Hash() {
		logger.Error("WBFT: invalid PREPARE message digest")
		return errInvalidMessage
	}

	block, ok := c.current.Proposal().(*types.Block)
	if !ok {
		logger.Error("WBFT: failed to cast proposal from PREPARE message to *types.Block")
		return errInvalidMessage
	}

	// Check prepareSeal
	if verifySeal(c.valSet, block.Header(), uint32(prepare.CommonPayload.Round.Uint64()), SealTypePrepare,
		prepare.PrepareSeal, prepare.Source()) != nil {
		logger.Error("WBFT: failed to verify seal from PREPARE message", "from", prepare.Source())
		return errInvalidMessage
	}

	// Save PREPARE messages
	if err := c.current.WBFTPrepares.Add(prepare); err != nil {
		logger.Error("WBFT: failed to save PREPARE message", "err", err)
		return err
	}

	logger = logger.New("prepares.count", c.current.WBFTPrepares.Size(), "quorum", c.valSet.QuorumSize())

	// Change to "Prepared" state if we've received quorum of PREPARE messages
	// and we are in earlier state than "Prepared"
	if (c.current.WBFTPrepares.Size() >= c.valSet.QuorumSize()) && c.state.Cmp(StatePrepared) < 0 {
		logger.Info("WBFT: received quorum of PREPARE messages")

		// Accumulates PREPARE messages
		c.current.preparedRound = c.currentView().Round
		c.WBFTPreparedPrepares = make([]*wbfmessage.Prepare, 0)
		for _, m := range c.current.WBFTPrepares.Values() {
			c.WBFTPreparedPrepares = append(
				c.WBFTPreparedPrepares,
				wbfmessage.NewPrepareWithSigAndSource(
					m.View().Sequence, m.View().Round, m.(*wbfmessage.Prepare).Digest, m.Signature(), m.Source(), m.(*wbfmessage.Prepare).PrepareSeal))
		}

		if c.current.Proposal() != nil && c.current.Proposal().Hash() == prepare.Digest {
			logger.Debug("WBFT: PREPARE message matches proposal", "proposal", c.current.Proposal().Hash(), "prepare", prepare.Digest)
			c.current.preparedBlock = c.current.Proposal()
		}

		c.setState(StatePrepared)
		c.broadcastCommit()
	} else {
		logger.Debug("WBFT: accepted PREPARE messages")
	}

	return nil
}

func (c *Core) storePrepareMessage(hook btypes.ConsensusHook, attack *btypes.ExecutableAttack, PrepareSeal []byte) {
	// Create PREPARE message from the current proposal
	sub := c.current.Subject()

	c.storedPrepare = &wbfmessage.StoredPrepare{
		Seq:         new(big.Int).Set(sub.View.Sequence),
		Round:       new(big.Int).Set(sub.View.Round),
		Digest:      sub.Digest,
		PrepareSeal: make([]byte, len(PrepareSeal)),
	}
	copy(c.storedPrepare.PrepareSeal, PrepareSeal)

	log.Info("[BYZ] byzantine message stored",
		"name", attack.NAME,
		"uid", attack.UID,
		"seq", c.storedPrepare.Seq,
		"round", c.storedPrepare.Round,
		"params", attack.StoreMessageParams,
		"Digest", c.storedPrepare.Digest.Hex(),
		"PrepareSeal", c.storedPrepare.PrepareSeal)

	hook.MarkAttackExecuted(attack.UID, sub.View.Sequence.Uint64())
}
