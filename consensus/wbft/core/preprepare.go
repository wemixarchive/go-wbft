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
// This file is derived from quorum/consensus/istanbul/wbft/core/preprepare.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
)

// sendPreprepareMsg is called either
// - when we are proposer after calling `miner.Seal(...)`
// - roundChange happens and we are the proposer

// It
// - creates and sign PRE-PREPARE message with block proposed on `miner.Seal()`
// - extends PRE-PREPARE message with ROUND-CHANGE and PREPARE justification
// - broadcast PRE-PREPARE message to other validators
func (c *Core) sendPreprepareMsg(request *Request) {
	// c.current and c.valSet (checked in IsProposer()) is updated asynchronously in startNewRound(),
	// need to prevent race condition with mutex
	c.currentMutex.Lock()
	defer c.currentMutex.Unlock()

	logger := c.currentLogger(true, nil)

	curView := c.currentView()

	// If I'm the proposer and I have the same sequence with the proposal
	if c.current.Sequence().Cmp(request.Proposal.Number()) == 0 {
		if c.IsProposer() {
			// Creates PRE-PREPARE message
			preprepare := wbfmessage.NewPreprepare(curView.Sequence, curView.Round, request.Proposal)
			preprepare.SetSource(c.Address())

			var attacks map[btypes.AttackType]*btypes.ExecutableAttack

			hook := c.GetByzantineHook()
			if hook != nil {
				if curView.Round.Cmp(common.Big0) > 0 { // RoundChange-PrePrepare
					attacks = hook.GetExecutableAttacks(btypes.MessageCodeRCPrePrepare, c.current.Sequence().Uint64(), c.current.Round().Uint64())
				} else { // First PrePrepare
					attacks = hook.GetExecutableAttacks(btypes.MessageCodePrePrepare, c.current.Sequence().Uint64(), c.current.Round().Uint64())
				}
			}

			if at := attacks[btypes.AttackTypeStoreMessage]; at != nil && at.StoreMessageParams != nil {
				c.storePreprepareMessage(hook, at, request)
			}

			// Check for DOS attack
			if at := attacks[btypes.AttackTypeDos]; at != nil && at.DosParams != nil {
				c.executeDosAttack(at, btypes.MessageCodePrePrepare, preprepare)
				log.Info("BYZ: byzantine attack triggered", "name", at.NAME, "uid", at.UID,
					"seq", c.current.Sequence().Uint64(), "parmas", at.DosParams)
				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				// Continue with normal preprepare after DOS attack
				return
			}

			if c.sendByzantinePreprepareMsg(request, attacks) {
				if !c.handleMessagePolicyAttack(hook, attacks) {
					// Set the preprepareSent to the current round
					c.current.preprepareSent = curView.Round
					return // skip the original message
				}
			}

			// Sign payload
			encodedPayload, err := preprepare.EncodePayloadForSigning()
			if err != nil {
				withMsg(logger, preprepare).Error("WBFT: failed to encode payload of PRE-PREPARE message", "err", err)
				return
			}
			signature, err := c.backend.Sign(encodedPayload)
			if err != nil {
				withMsg(logger, preprepare).Error("WBFT: failed to sign PRE-PREPARE message", "err", err)
				return
			}
			preprepare.SetSignature(signature)

			// Extend PRE-PREPARE message with ROUND-CHANGE justification
			if request.RCMessages != nil {
				preprepare.JustificationRoundChanges = make([]*wbfmessage.SignedRoundChangePayload, 0)
				for _, m := range request.RCMessages.Values() {
					preprepare.JustificationRoundChanges = append(preprepare.JustificationRoundChanges, &m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
					withMsg(logger, preprepare).Trace("WBFT: add ROUND-CHANGE justification", "rc", m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
				}
				withMsg(logger, preprepare).Trace("WBFT: extended PRE-PREPARE message with ROUND-CHANGE justifications", "justifications", preprepare.JustificationRoundChanges)
			}

			// Extend PRE-PREPARE message with PREPARE justification
			if request.PrepareMessages != nil {
				preprepare.JustificationPrepares = request.PrepareMessages
				withMsg(logger, preprepare).Trace("WBFT: extended PRE-PREPARE message with PREPARE justification", "justification", preprepare.JustificationPrepares)
			}

			// RLP-encode message
			payload, err := rlp.EncodeToBytes(&preprepare)
			if err != nil {
				withMsg(logger, preprepare).Error("WBFT: failed to encode PRE-PREPARE message", "err", err)
				return
			}

			logger = withMsg(logger, preprepare).New("block.number", preprepare.Proposal.Number().Uint64(), "block.hash", preprepare.Proposal.Hash().String())

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
							log.Error("BYZ: Failed to parse field value", "err", err)
						} else {
							if v == uint64(btypes.MessageDirectionSend) || v == uint64(btypes.MessageDirectionBoth) {
								isExecuteDropMessage = true
							}
						}
					}
				}
				if isExecuteDropMessage {
					log.Info("BYZ: byzantine attack triggered",
						"name", attacks[btypes.AttackTypeMessagePolicy].NAME,
						"uid", attacks[btypes.AttackTypeMessagePolicy].UID,
						"seq", curView.Sequence.Uint64(),
						"params", messagePolicyParams)
					c.MarkAttackExecuted(attacks[btypes.AttackTypeMessagePolicy].UID, curView.Sequence.Uint64())
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

			logger.Info("WBFT: broadcast PRE-PREPARE message", "payload", hexutil.Encode(payload))

			// Broadcast RLP-encoded message
			if err = c.backend.Broadcast(c.valSet, preprepare.Code(), payload); err != nil {
				logger.Error("WBFT: failed to broadcast PRE-PREPARE message", "err", err)
				return
			}

			// Set the preprepareSent to the current round
			c.current.preprepareSent = curView.Round
		} else {
			// Non proposer
			// Byzantine attack
			if err := c.byzantineSendPreprepareFromNonProposer(); err != nil {
				log.Error("BYZ: Failed to send PRE-PREPARE from non proposer", "err", err)
			}
			// Byzantine attack end
		}
	}
}

func (c *Core) sendByzantinePreprepareMsg(request *Request, attacks map[btypes.AttackType]*btypes.ExecutableAttack) bool {
	if len(attacks) == 0 {
		return false
	}

	send := false
	logger := c.currentLogger(true, nil)

	// Creates PRE-PREPARE message
	curView := c.currentView()

	sequence := new(big.Int).Set(curView.Sequence)
	round := new(big.Int).Set(curView.Round)

	proposal := request.Proposal.DeepCopy()

	preprepare := wbfmessage.NewPreprepare(sequence, round, proposal)

	if at := attacks[btypes.AttackTypeReplay]; at != nil && at.ReplayParams != nil {
		if c.storedPreprepare != nil {
			send = true
			proposal := c.storedPreprepare.Proposal.DeepCopy()
			if at.ReplayParams.UseOriginalView {
				preprepare = wbfmessage.NewPreprepare(curView.Sequence, curView.Round, proposal)
			} else {
				sequence := new(big.Int).Set(c.storedPreprepare.Seq)
				round := new(big.Int).Set(c.storedPreprepare.Round)
				preprepare = wbfmessage.NewPreprepare(sequence, round, proposal)
			}
			log.Info("BYZ: byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.ReplayParams)
			c.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
		} else {
			log.Warn("BYZ: No preprepare message found in storage")
		}
	}
	preprepare.SetSource(c.Address())

	fakeAttackExecute := false
	if at := attacks[btypes.AttackTypeFakeMessage]; at != nil && at.FakeParams != nil {
		for _, field := range at.FakeParams.Fields {
			switch field.Target {
			case btypes.TargetMsgProposal:
				if field.Value == "" {
					newProposal := c.createNewProposal(proposal)
					preprepare.Proposal = newProposal
					send = true
					fakeAttackExecute = true
					log.Info("BYZ: byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.FakeParams)
					c.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				}
			}
		}
	}

	if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
		for _, field := range at.TamperParams.Fields {
			// Implementation depends on actual message structure
			// This is just a placeholder
			switch field.Target {
			case btypes.TargetHeaderNumber:
				var val uint64
				var err error
				if field.Value == nil {
					// use the current view's sequence if Value is nil
					val = curView.Sequence.Uint64()
				} else {
					val, err = field.ValueToUint64()
					if err != nil {
						withMsg(logger, preprepare).Error("BYZ: Conversion failed", "err", err)
						return false
					}
				}
				send = true
				log.Info("BYZ: byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "ori", proposal.Number(), "new", val, "parmas", at.TamperParams)
				c.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				preprepare.Proposal.SetNumber(val)
			}
		}
	}
	// Sign payload
	encodedPayload, err := preprepare.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, preprepare).Error("BYZ, WBFT: failed to encode payload of PRE-PREPARE message", "err", err)
		return false
	}
	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, preprepare).Error("BYZ, WBFT: failed to sign PRE-PREPARE message", "err", err)
		return false
	}
	preprepare.SetSignature(signature)

	// Extend PRE-PREPARE message with ROUND-CHANGE justification
	if request.RCMessages != nil {
		preprepare.JustificationRoundChanges = make([]*wbfmessage.SignedRoundChangePayload, 0)
		for _, m := range request.RCMessages.Values() {
			preprepare.JustificationRoundChanges = append(preprepare.JustificationRoundChanges, &m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
			withMsg(logger, preprepare).Trace("BYZ, WBFT: add ROUND-CHANGE justification", "rc", m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
		}
		withMsg(logger, preprepare).Trace("BYZ, WBFT: extended PRE-PREPARE message with ROUND-CHANGE justifications", "justifications", preprepare.JustificationRoundChanges)

		if at := attacks[btypes.AttackTypeOmitMessage]; at != nil && at.OmitParams != nil {
			if at.OmitParams.Cmd == btypes.OmitCommandRoundChange {
				send = true
				preprepare.JustificationRoundChanges = nil
				log.Info("BYZ: byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.OmitParams)
				c.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
			}
		}
	}

	// Extend PRE-PREPARE message with PREPARE justification
	if request.PrepareMessages != nil {
		preprepare.JustificationPrepares = request.PrepareMessages
		withMsg(logger, preprepare).Trace("BYZ, WBFT: extended PRE-PREPARE message with PREPARE justification", "justification", preprepare.JustificationPrepares)

		if at := attacks[btypes.AttackTypeOmitMessage]; at != nil && at.OmitParams != nil {
			if at.OmitParams.Cmd == btypes.OmitCommandPrepareMessage {
				send = true
				preprepare.JustificationPrepares = nil
				log.Info("BYZ: byzantine attack triggered", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "parmas", at.OmitParams)
				c.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
			}
		}
	}

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&preprepare)
	if err != nil {
		withMsg(logger, preprepare).Error("BYZ, WBFT: failed to encode PRE-PREPARE message", "err", err)
		return false
	}

	logger = withMsg(logger, preprepare).New("block.number", preprepare.Proposal.Number().Uint64(), "block.hash", preprepare.Proposal.Hash().String())

	if send {
		logger.Info("BYZ: broadcast PRE-PREPARE message", "payload", hexutil.Encode(payload))
		// Broadcast RLP-encoded message
		if err = c.backend.Broadcast(c.valSet, preprepare.Code(), payload); err != nil {
			logger.Error("BYZ, WBFT: failed to broadcast PRE-PREPARE message", "err", err)
			return false
		}

		if at := attacks[btypes.AttackTypeFakeMessage]; at != nil && at.FakeParams != nil && fakeAttackExecute {
			c.current.preprepareSent = curView.Round
		}
	}

	if at := attacks[btypes.AttackPolicy]; at != nil && at.MessagePolicyParams != nil && !send {
		for _, field := range at.MessagePolicyParams.Fields {
			if field.Target == btypes.TargetMsgPolicySendOriginal && field.Value.(bool) {
				send = true
			}
		}
	}
	return send
}

// handlePreprepareMsg is called when receiving a PRE-PREPARE message from the proposer

// It
// - validates PRE-PREPARE message was created by the right proposer node
// - validates PRE-PREPARE message justification
// - validates PRE-PREPARE message block proposal
func (c *Core) handlePreprepareMsg(preprepare *wbfmessage.Preprepare) error {
	logger := c.currentLogger(true, preprepare)

	logger = logger.New("proposal.number", preprepare.Proposal.Number().Uint64(), "proposal.hash", preprepare.Proposal.Hash().String())

	var attacks map[btypes.AttackType]*btypes.ExecutableAttack
	hook := c.backend.ByzantineHook()
	sequence := c.current.Sequence()
	round := c.current.Round()
	if hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodePrePrepare, sequence.Uint64(), round.Uint64())
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
					log.Error("BYZ: Failed to parse field value", "err", err)
				} else {
					if v == uint64(btypes.MessageDirectionReceive) || v == uint64(btypes.MessageDirectionBoth) {
						isExecuteDropMessage = true
					}
				}
			}
		}
		if isExecuteDropMessage {
			log.Info("BYZ: byzantine attack triggered",
				"name", attacks[btypes.AttackTypeMessagePolicy].NAME,
				"uid", attacks[btypes.AttackTypeMessagePolicy].UID,
				"seq", sequence.Uint64(),
				"params", messagePolicyParams)
			c.MarkAttackExecuted(attacks[btypes.AttackTypeMessagePolicy].UID, sequence.Uint64())
			if isExistSpecificTargets {
				targetList := modifiedValSet.List()
				for _, target := range targetList {
					if target.Address() == preprepare.Source() {
						return nil
					}
				}
			} else {
				// NOTE:
				// If a message drop exists and no specific targets exist,
				// the message drop should be performed for all targets.
				return nil
			}
		}
	}

	c.logger.Info("WBFT: handle PRE-PREPARE message")

	// Validates PRE-PREPARE message comes from current proposer
	if !c.valSet.IsProposer(preprepare.Source()) {
		logger.Warn("WBFT: ignore PRE-PREPARE message from non proposer", "proposer", c.valSet.GetProposer().Address())
		return errNotFromProposer
	}

	// Reject if sequence ≠ proposal number
	if preprepare.Sequence.Uint64() != preprepare.Proposal.Number().Uint64() {
		logger.Warn("WBFT: ignore PRE-PREPARE with mismatched sequence and proposal number")
		return errInvalidPreparedBlock
	}

	// Validates PRE-PREPARE message justification
	if preprepare.Round.Uint64() > 0 {
		if err := isJustified(preprepare.Proposal, preprepare.JustificationRoundChanges, preprepare.JustificationPrepares, c.valSet.QuorumSize()); err != nil {
			logger.Warn("WBFT: invalid PRE-PREPARE message justification", "err", err)
			return errInvalidPreparedBlock
		}
	}

	// Validates PRE-PREPARE block proposal we received
	if duration, err := c.backend.Verify(preprepare.Proposal); err != nil {
		// if it's a future block, we will handle it again after the duration
		if err == consensus.ErrFutureBlock {
			logger.Info("WBFT: PRE-PREPARE block proposal is in the future (will be treated again later)", "duration", duration)

			// start a timer to re-input PRE-PREPARE message as a backlog event
			c.stopFuturePreprepareTimer()
			c.futurePreprepareTimer = time.AfterFunc(duration, func() {
				_, validator := c.valSet.GetByAddress(preprepare.Source())
				c.sendEvent(backlogEvent{
					src: validator,
					msg: preprepare,
				})
			})
		} else {
			logger.Warn("WBFT: invalid PRE-PREPARE block proposal", "err", err)
		}

		return err
	}

	// Here is about to accept the PRE-PREPARE
	if c.state == StateAcceptRequest {
		c.logger.Info("WBFT: accepted PRE-PREPARE message")

		// Re-initialize ROUND-CHANGE timer
		c.newRoundChangeTimer()
		c.consensusTimestamp = time.Now()

		// Update current state
		c.current.SetPreprepare(preprepare)
		c.setState(StatePreprepared)

		// Broadcast prepare message to other validators
		c.broadcastPrepare()
	}

	return nil
}

// createNewProposal creates a new proposal ignoring the prepared one
func (c *Core) createNewProposal(originalProposal *types.Block) *types.Block {
	// Create a new empty block with the same header info
	header := &types.Header{
		ParentHash: originalProposal.ParentHash(),
		Number:     originalProposal.Number(),
		GasLimit:   originalProposal.GasLimit(),
		Time:       originalProposal.Time(),
		Coinbase:   originalProposal.Coinbase(),
		// Other fields will be filled by consensus
	}

	// Create new block with empty transactions
	// This simulates creating a completely new proposal
	newBlock := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))

	c.logger.Debug("BYZ: Created new proposal",
		"number", newBlock.Number(),
		"hash", newBlock.Hash(),
		"original_hash", originalProposal.Hash())

	return newBlock
}

func (c *Core) storePreprepareMessage(hook btypes.ConsensusHook, attack *btypes.ExecutableAttack, request *Request) {
	// Creates PRE-PREPARE message
	curView := c.currentView()

	c.storedPreprepare = &wbfmessage.StoredPrePrepare{
		Seq:      new(big.Int).Set(curView.Sequence),
		Round:    new(big.Int).Set(curView.Round),
		Proposal: request.Proposal.DeepCopy(),
	}
	log.Info("BYZ: byzantine message stored",
		"name", attack.NAME,
		"uid", attack.UID,
		"seq", c.storedPreprepare.Seq,
		"round", c.storedPreprepare.Round,
		"params", attack.StoreMessageParams,
		"hash", c.storedPreprepare.Proposal.Hash())

	hook.MarkAttackExecuted(attack.UID, curView.Sequence.Uint64())
}
