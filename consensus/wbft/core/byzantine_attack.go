package core

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
)

func (c *Core) byzantineSendPreprepareFromNonProposer() error {
	// Byzantine logic: Check if we have a role spoof attack configured
	var attacks map[btypes.AttackType]*btypes.ExecutableAttack
	hook := c.backend.ByzantineHook()
	sequence := c.current.Sequence()
	round := c.current.Round()
	if hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodePrePrepare, sequence.Uint64(), round.Uint64())
	}

	if at := attacks[btypes.AttackTypeRoleSpoofed]; at != nil && at.RoleSpoofParams != nil {
		currentRound := c.currentView().Round

		log.Info("[BYZ] WBFT: received quorum of ROUND-CHANGE messages")
		log.Info("[BYZ] Non-proposer attempting to send PrePrepare after round change",
			"sequence", c.current.Sequence().Uint64(),
			"round", currentRound.Uint64(),
			"actual_proposer", c.valSet.GetProposer().Address(),
			"byzantine_node", c.Address())

		// Prepare the same data as proposer would
		_, proposal := c.highestPrepared(currentRound)
		if proposal == nil {
			if c.current != nil && c.current.pendingRequest != nil {
				proposal = c.current.pendingRequest.Proposal
			} else {
				log.Warn("[BYZ] Cannot execute role spoof: no proposal available")
				return nil
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
			log.Error("[BYZ] Invalid ROUND-CHANGE message justification", "err", err)
			return err
		}

		r := &Request{
			Proposal:        proposal,
			RCMessages:      roundChangeMessages,
			PrepareMessages: prepareMessages,
		}

		c.sendByzantinePreprepareMsg(hook, r, attacks)

		if len(attacks) == 0 {
			return fmt.Errorf("no byzantine attack configured for PrePrepare message")
		}

		logger := c.currentLogger(true, nil)

		// Creates PRE-PREPARE message
		curView := c.currentView()

		preprepare := wbfmessage.NewPreprepare(sequence, round, r.Proposal)
		preprepare.SetSource(c.Address())

		roleSpoofAttackExecute := false
		for _, field := range at.RoleSpoofParams.Fields {
			switch field.Target {
			case btypes.RoleProposer:
				if !c.IsProposer() {
					if r.RCMessages != nil {
						roleSpoofAttackExecute = true
						// When attacked with a tamper attack, the value exists in value.
						switch field.Value {
						case "":
						default:
						}
					}
				}
			}
		}

		// Sign payload
		encodedPayload, err := preprepare.EncodePayloadForSigning()
		if err != nil {
			withMsg(logger, preprepare).Error("[BYZ] WBFT: failed to encode payload of PRE-PREPARE message", "err", err)
			return fmt.Errorf("[BYZ] WBFT: failed to encode payload of PRE-PREPARE message")
		}
		signature, err := c.backend.Sign(encodedPayload)
		if err != nil {
			withMsg(logger, preprepare).Error("[BYZ] WBFT: failed to sign PRE-PREPARE message", "err", err)
			return fmt.Errorf("[BYZ] failed to sign PRE-PREPARE message")
		}
		preprepare.SetSignature(signature)

		// Extend PRE-PREPARE message with ROUND-CHANGE justification
		if r.RCMessages != nil {
			preprepare.JustificationRoundChanges = make([]*wbfmessage.SignedRoundChangePayload, 0)
			for _, m := range r.RCMessages.Values() {
				preprepare.JustificationRoundChanges = append(preprepare.JustificationRoundChanges, &m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
				withMsg(logger, preprepare).Trace("[BYZ] WBFT: add ROUND-CHANGE justification", "rc", m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
			}
			withMsg(logger, preprepare).Trace("[BYZ] WBFT: extended PRE-PREPARE message with ROUND-CHANGE justifications", "justifications", preprepare.JustificationRoundChanges)
		}

		// Extend PRE-PREPARE message with PREPARE justification
		if r.PrepareMessages != nil {
			preprepare.JustificationPrepares = r.PrepareMessages
			withMsg(logger, preprepare).Trace("[BYZ] WBFT: extended PRE-PREPARE message with PREPARE justification", "justification", preprepare.JustificationPrepares)
		}

		// RLP-encode message
		payload, err := rlp.EncodeToBytes(&preprepare)
		if err != nil {
			withMsg(logger, preprepare).Error("[BYZ] WBFT: failed to encode PRE-PREPARE message", "err", err)
			return fmt.Errorf("[BYZ] failed to encode PRE-PREPARE message")
		}

		logger = withMsg(logger, preprepare).New("block.number", preprepare.Proposal.Number().Uint64(), "block.hash", preprepare.Proposal.Hash().String())

		if roleSpoofAttackExecute {
			logger.Info("[BYZ] broadcast PRE-PREPARE message", "payload", hexutil.Encode(payload))
			// Broadcast RLP-encoded message
			if err = c.backend.Broadcast(c.valSet, preprepare.Code(), payload); err != nil {
				logger.Error("[BYZ] WBFT: failed to broadcast PRE-PREPARE message", "err", err)
				return fmt.Errorf("[BYZ] failed to broadcast PRE-PREPARE message")
			}

			c.current.preprepareSent = curView.Round
			log.Info("[BYZ] attack",
				"name", at.NAME,
				"uid", at.UID,
				"seq", c.current.Sequence().Uint64(),
				"params", at.RoleSpoofParams,
				"actual_proposer", c.valSet.GetProposer().Address(),
				"byzantine_node", c.Address(),
				"has_rc_messages", r.RCMessages != nil)

			if hook != nil {
				err := hook.MarkAttackExecuted(at.UID, c.current.Sequence().Uint64())
				if err != nil {
					log.Warn("[BYZ] Attack Executed Mark Error", "err", err)
				}
			}
			return nil
		}
		return fmt.Errorf("[BYZ] Role spoof attack executed, but no PRE-PREPARE message sent")
	}
	return nil
}

func (c *Core) sendByzantineRoundChangeMsg(at *btypes.ExecutableAttack, originRoundChange *wbfmessage.RoundChange) bool {
	if at != nil && at.FakeParams != nil {
		logger := c.currentLogger(true, nil)

		fakedRoundChange := c.applyByzantineRoundChangeMsg(at, originRoundChange)
		if fakedRoundChange == nil {
			log.Error("[BYZ] Failed to create faked ROUND-CHANGE message")
			return false
		}

		// Sign message
		encodedPayload, err := fakedRoundChange.EncodePayloadForSigning()
		if err != nil {
			withMsg(logger, fakedRoundChange).Error("[BYZ] WBFT: failed to encode ROUND-CHANGE message", "err", err)
			return false
		}
		signature, err := c.backend.Sign(encodedPayload)
		if err != nil {
			withMsg(logger, fakedRoundChange).Error("[BYZ] WBFT: failed to sign ROUND-CHANGE message", "err", err)
			return false
		}
		fakedRoundChange.SetSignature(signature)

		// Extend ROUND-CHANGE message with PREPARE justification
		// Check if we need to manipulate the justification based on attack code
		if c.WBFTPreparedPrepares != nil {
			for _, field := range at.FakeParams.Fields {
				switch field.Target {
				case btypes.FakeTargetJustification:
					value, err := field.ValueToUint64()
					if err != nil {
						log.Error("[BYZ] Failed to parse field value for ROUND-CHANGE", "err", err)
						return false
					}
					switch value {
					case 0: // Remove justification entirely
						fakedRoundChange.Justification = nil
						log.Info("[BYZ] Removed PREPARE justification from ROUND-CHANGE message")
					case 1: // Reduce justification (remove some PREPARE messages)
						if len(c.WBFTPreparedPrepares) > 1 {
							// Keep only first PREPARE message (insufficient for quorum)
							fakedRoundChange.Justification = c.WBFTPreparedPrepares[:1]
							log.Info("[BYZ] Reduced PREPARE justification",
								"original_count", len(c.WBFTPreparedPrepares),
								"reduced_count", 1)
						} else {
							fakedRoundChange.Justification = c.WBFTPreparedPrepares
						}
					case 2: // Tamper with justification (modify digest in PREPARE messages)
						tamperedJustification := make([]*wbfmessage.Prepare, 0, len(c.WBFTPreparedPrepares))
						for _, prepare := range c.WBFTPreparedPrepares {
							// Create a copy and modify the digest
							tamperedPrepare := prepare.DeepCopy()
							// Flip some bits in the digest
							for i := 0; i < 4; i++ {
								tamperedPrepare.Digest[i] = ^tamperedPrepare.Digest[i]
							}
							tamperedJustification = append(tamperedJustification, tamperedPrepare)
						}
						fakedRoundChange.Justification = tamperedJustification
						log.Info("[BYZ] Tampered with PREPARE justification digests")
					case 3:
						// Create fake justification when there's none
						// This creates invalid justification with wrong sequence/round
						fakeJustification := make([]*wbfmessage.Prepare, 0, c.valSet.QuorumSize())
						for i := 0; i < c.valSet.QuorumSize(); i++ {
							fakePrepare := &wbfmessage.Prepare{
								CommonPayload: wbfmessage.CommonPayload{
									Sequence: fakedRoundChange.Sequence,
									Round:    fakedRoundChange.Round, // Wrong round for justification
								},
								Digest: fakedRoundChange.PreparedDigest,
							}
							fakeJustification = append(fakeJustification, fakePrepare)
						}
						fakedRoundChange.Justification = fakeJustification
						log.Info("[BYZ] Created fake PREPARE justification")
					default:
						log.Warn("[BYZ] Unknown justification manipulation value", "value", value)
						return false
					}
				default:
					// Normal behavior - use original justification
					fakedRoundChange.Justification = c.WBFTPreparedPrepares
					withMsg(logger, fakedRoundChange).Debug("[BYZ] WBFT: extended ROUND-CHANGE message with PREPARE justification", "justification", fakedRoundChange.Justification)
				}
			}
		}

		// RLP-encode message
		data, err := rlp.EncodeToBytes(fakedRoundChange)
		if err != nil {
			withMsg(logger, fakedRoundChange).Error("[BYZ] WBFT: failed to encode ROUND-CHANGE message", "err", err)
			return false
		}

		withMsg(logger, fakedRoundChange).Info("[BYZ] WBFT: broadcast ROUND-CHANGE message", "payload", hexutil.Encode(data))

		// Broadcast RLP-encoded message
		if err = c.backend.Broadcast(c.valSet, fakedRoundChange.Code(), data); err != nil {
			withMsg(logger, fakedRoundChange).Error("[BYZ] WBFT: failed to broadcast ROUND-CHANGE message", "err", err)
			return false
		}
		log.Info("[BYZ] attack", "name", at.NAME,
			"uid", at.UID,
			"seq", c.current.Sequence().Uint64(),
			"code", at.FakeParams.Code,
			"parmas", at.FakeParams,
			"origin digest", originRoundChange.PreparedDigest,
			"fake digest", fakedRoundChange.PreparedDigest,
			"original_pr", originRoundChange.PreparedRound,
			"fake_pr", fakedRoundChange.PreparedRound)
		return true
	}
	return false
}

func (c *Core) applyByzantineRoundChangeMsg(at *btypes.ExecutableAttack, originRoundChange *wbfmessage.RoundChange) *wbfmessage.RoundChange {
	var fakedRoundChange *wbfmessage.RoundChange = nil

	if at != nil && at.FakeParams != nil {
		for _, field := range at.FakeParams.Fields {
			switch field.Target {
			case btypes.FakeTargetProposal:
				value, err := field.ValueToUint64()
				if err != nil {
					log.Error("[BYZ] Failed to parse field value for ROUND-CHANGE", "err", err)
					return nil
				}

				switch value {
				case 0:
					if originRoundChange != nil {
						proposal := c.current.Proposal().DeepCopy()
						fakedRequest := c.createNewProposal(proposal)
						preparedRound := originRoundChange.PreparedRound
						if preparedRound != nil {
							preparedRound = nil
						}
						fakedRoundChange = wbfmessage.NewRoundChange(originRoundChange.Sequence,
							originRoundChange.Round, originRoundChange.PreparedRound, fakedRequest)
						fakedRoundChange.PreparedDigest = originRoundChange.PreparedDigest
					} else {
						log.Error("[BYZ] No origin round change message found, cannot fake proposal")
					}
				}
			}
		}
	}
	return fakedRoundChange
}
