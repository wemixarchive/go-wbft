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

	// If I'm the proposer and I have the same sequence with the proposal
	if c.current.Sequence().Cmp(request.Proposal.Number()) == 0 && c.IsProposer() {
		// Creates PRE-PREPARE message
		curView := c.currentView()
		preprepare := wbfmessage.NewPreprepare(curView.Sequence, curView.Round, request.Proposal)
		preprepare.SetSource(c.Address())

		var attacks map[btypes.AttackType]*btypes.ExecutableAttack

		hook := c.backend.ByzantineHook()
		if hook != nil {
			attacks = hook.GetExecutableAttacks(btypes.MessageCodePrePrepare, c.current.Sequence().Uint64(), c.current.Round().Uint64())
		}

		if c.sendByzantinePreprepareMsg(hook, request, attacks) {
			if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
				if !at.TamperParams.WithValidMessage {
					hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
					return // skip the normal message
				}
				log.Info("[BYZ] sending valid message", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "delay(ms)", at.TamperParams.Delay)
				// Wait for the configured delay
				time.Sleep(time.Duration(at.TamperParams.Delay) * time.Millisecond)
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

		// Byzantine hook for fake attack - Invalid PrePrepare
		//if hook := c.backend.ByzantineHook(); hook != nil {
		//	if at := attacks[btypes.AttackTypeFakeMessage]; at != nil && at.FakeParams != nil {
		//		if at.FakeParams.FakeType == "invalidProposal" && at.FakeParams.IgnorePrepared {
		//			// Round > 0이고 prepared proposal이 있어도 무시
		//			if c.current.Round().Uint64() > 0 && request.PrepareMessages != nil {
		//				originalPrepares := len(request.PrepareMessages)
		//				request.PrepareMessages = nil
		//				preprepare.JustificationPrepares = nil
		//
		//				// Create new proposal ignoring the prepared one
		//				if block, ok := request.Proposal.(*types.Block); ok {
		//					newProposal := c.createNewProposal(block)
		//					request.Proposal = newProposal
		//					preprepare.Proposal = newProposal
		//				}
		//
		//				logger.Info("[BYZ] Ignoring prepared proposal, creating new one",
		//					"attack_uid", at.UID,
		//					"round", c.current.Round(),
		//					"ignored_prepares", originalPrepares)
		//
		//				// Mark attack as executed
		//				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
		//			}
		//		}
		//	}
		//}

		// RLP-encode message
		payload, err := rlp.EncodeToBytes(&preprepare)
		if err != nil {
			withMsg(logger, preprepare).Error("WBFT: failed to encode PRE-PREPARE message", "err", err)
			return
		}

		logger = withMsg(logger, preprepare).New("block.number", preprepare.Proposal.Number().Uint64(), "block.hash", preprepare.Proposal.Hash().String())

		if at := attacks[btypes.AttackTypeSilentMessage]; at != nil && at.SilentParams != nil {
			if at.SilentParams.Direction == 1 {
				log.Info("[BYZ] attack", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "params", at.SilentParams)
				hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				return
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
	}
}

func (c *Core) sendByzantinePreprepareMsg(hook btypes.ConsensusHook, request *Request, attacks map[btypes.AttackType]*btypes.ExecutableAttack) bool {
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
	preprepare.SetSource(c.Address())

	if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
		for _, field := range at.TamperParams.TamperFields {
			// Implementation depends on actual message structure
			// This is just a placeholder
			switch field.Target {
			case btypes.TamperProposalHeaderNumber:
				val, err := field.ValueToUint64()
				if err != nil {
					withMsg(logger, preprepare).Error("[BYZ] Conversion failed", "err", err)
					return false
				} else {
					send = true
					log.Info("[BYZ] attack", "name", at.NAME, "uid", at.UID, "seq", c.current.Sequence().Uint64(), "original", proposal.Number(), "parmas", at.TamperParams)
					hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
					proposal.SetNumber(val)
				}
			}
		}
	}
	// Sign payload
	encodedPayload, err := preprepare.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, preprepare).Error("[BYZ] WBFT: failed to encode payload of PRE-PREPARE message", "err", err)
		return false
	}
	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, preprepare).Error("[BYZ] WBFT: failed to sign PRE-PREPARE message", "err", err)
		return false
	}
	preprepare.SetSignature(signature)

	// Extend PRE-PREPARE message with ROUND-CHANGE justification
	if request.RCMessages != nil {
		preprepare.JustificationRoundChanges = make([]*wbfmessage.SignedRoundChangePayload, 0)
		for _, m := range request.RCMessages.Values() {
			preprepare.JustificationRoundChanges = append(preprepare.JustificationRoundChanges, &m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
			withMsg(logger, preprepare).Trace("[BYZ] WBFT: add ROUND-CHANGE justification", "rc", m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
		}
		withMsg(logger, preprepare).Trace("[BYZ] WBFT: extended PRE-PREPARE message with ROUND-CHANGE justifications", "justifications", preprepare.JustificationRoundChanges)
	}

	// Extend PRE-PREPARE message with PREPARE justification
	if request.PrepareMessages != nil {
		preprepare.JustificationPrepares = request.PrepareMessages
		withMsg(logger, preprepare).Trace("[BYZ] WBFT: extended PRE-PREPARE message with PREPARE justification", "justification", preprepare.JustificationPrepares)
	}

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&preprepare)
	if err != nil {
		withMsg(logger, preprepare).Error("[BYZ] WBFT: failed to encode PRE-PREPARE message", "err", err)
		return false
	}

	logger = withMsg(logger, preprepare).New("block.number", preprepare.Proposal.Number().Uint64(), "block.hash", preprepare.Proposal.Hash().String())

	if send {
		logger.Info("[BYZ] broadcast PRE-PREPARE message", "payload", hexutil.Encode(payload))
		// Broadcast RLP-encoded message
		if err = c.backend.Broadcast(c.valSet, preprepare.Code(), payload); err != nil {
			logger.Error("[BYZ] WBFT: failed to broadcast PRE-PREPARE message", "err", err)
			return false
		}
		if at := attacks[btypes.AttackTypeTamperedMessage]; at != nil && at.TamperParams != nil {
			if !at.TamperParams.WithValidMessage {
				// Set the preprepareSent to the current round
				c.current.preprepareSent = curView.Round
			}
		}
	}
	return true && send
}

// handlePreprepareMsg is called when receiving a PRE-PREPARE message from the proposer

// It
// - validates PRE-PREPARE message was created by the right proposer node
// - validates PRE-PREPARE message justification
// - validates PRE-PREPARE message block proposal
func (c *Core) handlePreprepareMsg(preprepare *wbfmessage.Preprepare) error {
	logger := c.currentLogger(true, preprepare)

	logger = logger.New("proposal.number", preprepare.Proposal.Number().Uint64(), "proposal.hash", preprepare.Proposal.Hash().String())

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

	c.logger.Info("[BYZ] Created new proposal",
		"number", newBlock.Number(),
		"hash", newBlock.Hash(),
		"original_hash", originalProposal.Hash())

	return newBlock
}
