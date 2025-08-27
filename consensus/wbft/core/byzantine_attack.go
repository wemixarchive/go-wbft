package core

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
)

func (c *Core) GetByzantineHook() btypes.ConsensusHook {
	return c.backend.ByzantineHook()
}

func (c *Core) GetByzantineAttacks(msgCode btypes.MessageCode, sequence, round uint64) map[btypes.AttackType]*btypes.ExecutableAttack {
	hook := c.GetByzantineHook()
	if hook == nil {
		log.Warn("BYZ: byzantine hook is nil")
		return nil
	}
	return hook.GetExecutableAttacks(msgCode, sequence, round)
}

func (c *Core) MarkAttackExecuted(uid string, sequence uint64) error {
	if hook := c.GetByzantineHook(); hook != nil {
		return c.GetByzantineHook().MarkAttackExecuted(uid, sequence)
	}
	return errors.New("BYZ: not found by byzantine hook")
}

func (c *Core) IsExecuteAttack(attacks map[btypes.AttackType]*btypes.ExecutableAttack, attackType btypes.AttackType) bool {
	if at := attacks[attackType]; at != nil {
		switch attackType {
		case btypes.AttackTypeMessagePolicy:
			if at.MessagePolicyParams != nil {
				return true
			}
		case btypes.AttackTypeTamperedMessage:
			if at.TamperParams != nil {
				return true
			}
		case btypes.AttackTypeFakeMessage:
			if at.FakeParams != nil {
				return true
			}
		case btypes.AttackTypeOmitMessage:
			if at.OmitParams != nil {
				return true
			}
		case btypes.AttackTypeRoleSpoofed:
			if at.RoleSpoofParams != nil {
				return true
			}
		case btypes.AttackTypeReplay:
			if at.ReplayParams != nil {
				return true
			}
		case btypes.AttackTypeStoreMessage:
			if at.StoreMessageParams != nil {
				return true
			}
		case btypes.AttackTypeDos:
			if at.DosParams != nil {
				return true
			}
		default:
			return false
		}
	}
	return false
}

func (c *Core) byzantineSendPreprepareFromNonProposer() error {
	// Byzantine logic: Check if we have a role spoof attack configured
	sequence := c.current.Sequence()
	round := c.current.Round()
	attacks := c.GetByzantineAttacks(btypes.MessageCodeRCPrePrepare, sequence.Uint64(), round.Uint64())
	if attacks == nil {
		return errors.New("BYZ: byzantine hook is nil")
	}

	if c.IsExecuteAttack(attacks, btypes.AttackTypeRoleSpoofed) {
		roleSpoofParams := attacks[btypes.AttackTypeRoleSpoofed].RoleSpoofParams
		currentRound := c.currentView().Round

		log.Info("BYZ, WBFT: received quorum of ROUND-CHANGE messages")
		log.Info("BYZ: Non-proposer attempting to send PrePrepare after round change",
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
				log.Warn("BYZ: Cannot execute role spoof: no proposal available")
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
			log.Error("BYZ: Invalid ROUND-CHANGE message justification", "err", err)
			return err
		}

		r := &Request{
			Proposal:        proposal,
			RCMessages:      roundChangeMessages,
			PrepareMessages: prepareMessages,
		}

		c.sendByzantinePreprepareMsg(r, attacks)

		if len(attacks) == 0 {
			return fmt.Errorf("no byzantine attack configured for PrePrepare message")
		}

		logger := c.currentLogger(true, nil)

		// Creates PRE-PREPARE message
		curView := c.currentView()

		preprepare := wbfmessage.NewPreprepare(sequence, round, r.Proposal)
		preprepare.SetSource(c.Address())

		roleSpoofAttackExecute := false
		for _, field := range roleSpoofParams.Fields {
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
			withMsg(logger, preprepare).Error("BYZ, WBFT: failed to encode payload of PRE-PREPARE message", "err", err)
			return fmt.Errorf("BYZ, WBFT: failed to encode payload of PRE-PREPARE message")
		}
		signature, err := c.backend.Sign(encodedPayload)
		if err != nil {
			withMsg(logger, preprepare).Error("BYZ, WBFT: failed to sign PRE-PREPARE message", "err", err)
			return fmt.Errorf("BYZ: failed to sign PRE-PREPARE message")
		}
		preprepare.SetSignature(signature)

		// Extend PRE-PREPARE message with ROUND-CHANGE justification
		if r.RCMessages != nil {
			preprepare.JustificationRoundChanges = make([]*wbfmessage.SignedRoundChangePayload, 0)
			for _, m := range r.RCMessages.Values() {
				preprepare.JustificationRoundChanges = append(preprepare.JustificationRoundChanges, &m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
				withMsg(logger, preprepare).Trace("BYZ, WBFT: add ROUND-CHANGE justification", "rc", m.(*wbfmessage.RoundChange).SignedRoundChangePayload)
			}
			withMsg(logger, preprepare).Trace("BYZ, WBFT: extended PRE-PREPARE message with ROUND-CHANGE justifications", "justifications", preprepare.JustificationRoundChanges)
		}

		// Extend PRE-PREPARE message with PREPARE justification
		if r.PrepareMessages != nil {
			preprepare.JustificationPrepares = r.PrepareMessages
			withMsg(logger, preprepare).Trace("BYZ, WBFT: extended PRE-PREPARE message with PREPARE justification", "justification", preprepare.JustificationPrepares)
		}

		// RLP-encode message
		payload, err := rlp.EncodeToBytes(&preprepare)
		if err != nil {
			withMsg(logger, preprepare).Error("BYZ, WBFT: failed to encode PRE-PREPARE message", "err", err)
			return fmt.Errorf("BYZ: failed to encode PRE-PREPARE message")
		}

		logger = withMsg(logger, preprepare).New("block.number", preprepare.Proposal.Number().Uint64(), "block.hash", preprepare.Proposal.Hash().String())

		if roleSpoofAttackExecute {
			logger.Info("BYZ: broadcast PRE-PREPARE message", "payload", hexutil.Encode(payload))
			// Broadcast RLP-encoded message
			if err = c.backend.Broadcast(c.valSet, preprepare.Code(), payload); err != nil {
				logger.Error("BYZ, WBFT: failed to broadcast PRE-PREPARE message", "err", err)
				return fmt.Errorf("BYZ: failed to broadcast PRE-PREPARE message")
			}

			c.current.preprepareSent = curView.Round
			log.Info("BYZ: byzantine attack triggered",
				"name", attacks[btypes.AttackTypeRoleSpoofed].NAME,
				"uid", attacks[btypes.AttackTypeRoleSpoofed].UID,
				"seq", c.current.Sequence().Uint64(),
				"params", roleSpoofParams,
				"actual_proposer", c.valSet.GetProposer().Address(),
				"byzantine_node", c.Address(),
				"has_rc_messages", r.RCMessages != nil)

			if err := c.MarkAttackExecuted(attacks[btypes.AttackTypeRoleSpoofed].UID, c.current.Sequence().Uint64()); err != nil {
				log.Warn("BYZ: Attack Executed Mark Error", "err", err)
			}

			return nil
		}
		return fmt.Errorf("BYZ: Role spoof attack executed, but no PRE-PREPARE message sent")
	}
	return nil
}

func (c *Core) applyByzantineRoundChangeMsg(at *btypes.ExecutableAttack, originRoundChange *wbfmessage.RoundChange) *wbfmessage.RoundChange {
	var fakedRoundChange *wbfmessage.RoundChange = nil

	if at != nil && at.FakeParams != nil {
		for _, field := range at.FakeParams.Fields {
			switch field.Target {
			case btypes.TargetMsgProposal:
				value, err := field.ValueToUint64()
				if err != nil {
					log.Error("BYZ: Failed to parse field value for ROUND-CHANGE", "err", err)
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
						log.Error("BYZ: No origin round change message found, cannot fake proposal")
					}
				}
			}
		}
	}
	return fakedRoundChange
}

// executeDosAttack executes DOS attack by flooding messages
func (c *Core) executeDosAttack(attack *btypes.ExecutableAttack, msgCode btypes.MessageCode, originalMsg wbfmessage.WBFTMessage) {
	if attack == nil || attack.DosParams == nil {
		return
	}

	params := attack.DosParams
	count := params.Cnt
	if count == 0 {
		return
	}

	// Execute based on cmd value
	switch params.Cmd {
	case 0:
		// Send valid messages
		c.sendValidDosMessages(msgCode, originalMsg, count, params.Delay)
	case 1, 2:
		// Send invalid messages
		c.sendInvalidDosMessages(msgCode, originalMsg, count, params.Delay, params.Cmd)
	default:
		log.Warn("BYZ: Unknown DOS cmd", "cmd", params.Cmd)
	}
}

// sendValidDosMessages sends valid messages for DOS attack
func (c *Core) sendValidDosMessages(msgCode btypes.MessageCode, originalMsg wbfmessage.WBFTMessage, count uint64, delay uint64) {
	switch msgCode {
	case btypes.MessageCodePrePrepare:
		if preprepare, ok := originalMsg.(*wbfmessage.Preprepare); ok {
			c.sendValidDosPrePrepare(preprepare, count, delay)
		}
	case btypes.MessageCodePrepare:
		if prepare, ok := originalMsg.(*wbfmessage.Prepare); ok {
			c.sendValidDosPrepare(prepare, count, delay)
		}
	case btypes.MessageCodeCommit:
		if commit, ok := originalMsg.(*wbfmessage.Commit); ok {
			c.sendValidDosCommit(commit, count, delay)
		}
	case btypes.MessageCodeRoundChange:
		if roundChange, ok := originalMsg.(*wbfmessage.RoundChange); ok {
			c.sendValidDosRoundChange(roundChange, count, delay)
		}
	default:
	}
}

// sendInvalidDosMessages sends invalid messages for DOS attack
func (c *Core) sendInvalidDosMessages(msgCode btypes.MessageCode, originalMsg wbfmessage.WBFTMessage, count, delay, cmd uint64) {
	switch msgCode {
	case btypes.MessageCodePrePrepare:
		if preprepare, ok := originalMsg.(*wbfmessage.Preprepare); ok {
			c.sendInvalidDosPrePrepare(preprepare, count, delay, cmd)
		}
	case btypes.MessageCodePrepare:
		if prepare, ok := originalMsg.(*wbfmessage.Prepare); ok {
			c.sendInvalidDosPrepare(prepare, count, delay, cmd)
		}
	case btypes.MessageCodeCommit:
		if commit, ok := originalMsg.(*wbfmessage.Commit); ok {
			c.sendInvalidDosCommit(commit, count, delay, cmd)
		}
	case btypes.MessageCodeRoundChange:
		if roundChange, ok := originalMsg.(*wbfmessage.RoundChange); ok {
			c.sendInvalidDosRoundChange(roundChange, count, delay, cmd)
		}
	default:
	}
}

// sendValidDosPrePrepare sends valid preprepare messages for DOS (only if proposer)
func (c *Core) sendValidDosPrePrepare(originalPreprepare *wbfmessage.Preprepare, count, delay uint64) {
	if !c.IsProposer() {
		log.Warn("BYZ: Cannot send PrePrepare DOS attack: not proposer")
		return
	}

	var messages []wbfmessage.WBFTMessage

	for i := uint64(0); i < count; i++ {
		originalPreprepare.SetSource(c.Address())
		messages = append(messages, originalPreprepare)
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// sendInvalidDosPrePrepare sends invalid preprepare messages for DOS
func (c *Core) sendInvalidDosPrePrepare(originalPreprepare *wbfmessage.Preprepare, count, delay, cmd uint64) {
	var messages []wbfmessage.WBFTMessage

	for i := uint64(0); i < count; i++ {
		switch cmd {
		case 1:
			// Increment sequence number
			modifiedPreprepare := wbfmessage.NewPreprepare(
				big.NewInt(0).Add(originalPreprepare.Sequence, big.NewInt(int64(i+1))),
				originalPreprepare.Round,
				originalPreprepare.Proposal)
			modifiedPreprepare.SetSource(c.Address())
			messages = append(messages, modifiedPreprepare)
		case 2:
			// Increment round number
			modifiedPreprepare := wbfmessage.NewPreprepare(
				originalPreprepare.Sequence,
				big.NewInt(0).Add(originalPreprepare.Round, big.NewInt(int64(i+1))),
				originalPreprepare.Proposal)
			modifiedPreprepare.SetSource(c.Address())
			messages = append(messages, modifiedPreprepare)
		default:
			log.Warn("BYZ: Invalid command for PrePrepare DOS attack", "cmd", cmd)
			return
		}
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// sendValidDosPrepare sends valid prepare messages for DOS
func (c *Core) sendValidDosPrepare(originalPrepare *wbfmessage.Prepare, count, delay uint64) {
	var messages []wbfmessage.WBFTMessage

	// Simply resend the same valid prepare message
	for i := uint64(0); i < count; i++ {
		originalPrepare.SetSource(c.Address())
		messages = append(messages, originalPrepare)
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// sendInvalidDosPrepare sends invalid prepare messages for DOS
func (c *Core) sendInvalidDosPrepare(originalPrepare *wbfmessage.Prepare, count, delay, cmd uint64) {
	var messages []wbfmessage.WBFTMessage

	for i := uint64(0); i < count; i++ {
		switch cmd {
		case 1:
			// Increment sequence number
			modifiedPrepare := originalPrepare.DeepCopy()
			modifiedPrepare.Sequence = big.NewInt(0).Add(originalPrepare.Sequence, big.NewInt(int64(i+1)))
			modifiedPrepare.SetSource(c.Address())
			messages = append(messages, modifiedPrepare)
		case 2:
			// Increment round number
			modifiedPrepare := originalPrepare.DeepCopy()
			modifiedPrepare.Round = big.NewInt(0).Add(originalPrepare.Round, big.NewInt(int64(i+1)))
			modifiedPrepare.SetSource(c.Address())
			messages = append(messages, modifiedPrepare)
		default:
			log.Warn("BYZ: Invalid command for PrePrepare DOS attack", "cmd", cmd)
			return
		}
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// sendValidDosCommit sends valid commit messages for DOS
func (c *Core) sendValidDosCommit(originalCommit *wbfmessage.Commit, count, delay uint64) {
	var messages []wbfmessage.WBFTMessage

	// Simply resend the same valid prepare message
	for i := uint64(0); i < count; i++ {
		originalCommit.SetSource(c.Address())
		messages = append(messages, originalCommit)
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// sendInvalidDosCommit sends invalid commit messages for DOS
func (c *Core) sendInvalidDosCommit(originalCommit *wbfmessage.Commit, count, delay, cmd uint64) {
	var messages []wbfmessage.WBFTMessage

	for i := uint64(0); i < count; i++ {
		switch cmd {
		case 1:
			// Increment sequence number
			modifiedCommit := wbfmessage.NewCommit(
				big.NewInt(0).Add(originalCommit.Sequence, big.NewInt(int64(i+1))),
				originalCommit.Round,
				originalCommit.Digest,
				originalCommit.CommitSeal,
			)
			modifiedCommit.SetSource(c.Address())
			messages = append(messages, modifiedCommit)
		case 2:
			// Increment round number
			modifiedCommit := wbfmessage.NewCommit(
				originalCommit.Sequence,
				big.NewInt(0).Add(originalCommit.Round, big.NewInt(int64(i+1))),
				originalCommit.Digest,
				originalCommit.CommitSeal)
			modifiedCommit.SetSource(c.Address())
			messages = append(messages, modifiedCommit)
		default:
			log.Warn("BYZ: Invalid command for PrePrepare DOS attack", "cmd", cmd)
			return
		}
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// sendValidDosRoundChange sends valid round change messages for DOS
func (c *Core) sendValidDosRoundChange(originalRoundChange *wbfmessage.RoundChange, count, delay uint64) {
	var messages []wbfmessage.WBFTMessage

	// Simply resend the same valid prepare message
	for i := uint64(0); i < count; i++ {
		originalRoundChange.SetSource(c.Address())
		messages = append(messages, originalRoundChange)
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// sendInvalidDosRoundChange sends invalid round change messages for DOS
func (c *Core) sendInvalidDosRoundChange(originalRoundChange *wbfmessage.RoundChange, count, delay, cmd uint64) {
	var messages []wbfmessage.WBFTMessage

	for i := uint64(0); i < count; i++ {
		switch cmd {
		case 1:
			// Increment sequence number
			modifiedRoundChange := wbfmessage.NewRoundChange(
				big.NewInt(0).Add(originalRoundChange.Sequence, big.NewInt(int64(i+1))),
				originalRoundChange.Round,
				originalRoundChange.PreparedRound,
				nil)
			modifiedRoundChange.PreparedBlock = originalRoundChange.PreparedBlock
			modifiedRoundChange.PreparedDigest = originalRoundChange.PreparedDigest
			modifiedRoundChange.SetSource(c.Address())
			if originalRoundChange.Justification != nil {
				// Copy justification if exists
				modifiedRoundChange.Justification = make([]*wbfmessage.Prepare, len(originalRoundChange.Justification))
				copy(modifiedRoundChange.Justification, originalRoundChange.Justification)
			}

			messages = append(messages, modifiedRoundChange)
		case 2:
			// Increment round number
			modifiedRoundChange := wbfmessage.NewRoundChange(
				originalRoundChange.Sequence,
				big.NewInt(0).Add(originalRoundChange.Round, big.NewInt(int64(i+1))),
				originalRoundChange.PreparedRound,
				nil)
			modifiedRoundChange.PreparedBlock = originalRoundChange.PreparedBlock
			modifiedRoundChange.PreparedDigest = originalRoundChange.PreparedDigest
			modifiedRoundChange.SetSource(c.Address())
			if originalRoundChange.Justification != nil {
				// Copy justification if exists
				modifiedRoundChange.Justification = make([]*wbfmessage.Prepare, len(originalRoundChange.Justification))
				copy(modifiedRoundChange.Justification, originalRoundChange.Justification)
			}

			messages = append(messages, modifiedRoundChange)
		default:
			log.Warn("BYZ: Invalid command for PrePrepare DOS attack", "cmd", cmd)
			return
		}
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// signAndBroadcastMessage signs and broadcasts a message
func (c *Core) signAndBroadcastMessage(msg wbfmessage.WBFTMessage) {
	// Get message with signature interface
	msgWithSig, ok := msg.(interface {
		EncodePayloadForSigning() ([]byte, error)
		SetSignature([]byte)
		Code() uint64
	})
	if !ok {
		log.Error("BYZ: Message does not support signature operations")
		return
	}

	// Encode payload
	payload, err := msgWithSig.EncodePayloadForSigning()
	if err != nil {
		log.Error("BYZ: Failed to encode message", "err", err)
		return
	}

	// Sign
	sig, err := c.backend.Sign(payload)
	if err != nil {
		log.Error("BYZ: Failed to sign message", "err", err)
		return
	}

	msgWithSig.SetSignature(sig)

	// RLP encode and broadcast
	data, err := rlp.EncodeToBytes(msg)
	if err != nil {
		log.Error("BYZ: Failed to RLP encode message", "err", err)
		return
	}

	if err := c.backend.Broadcast(c.valSet, msgWithSig.Code(), data); err != nil {
		log.Error("BYZ: Failed to broadcast message", "err", err)
	}
}

func (c *Core) handleMessagePolicyAttack(hook btypes.ConsensusHook, attacks map[btypes.AttackType]*btypes.ExecutableAttack) bool {
	var at *btypes.ExecutableAttack

	var (
		shouldSend bool
		delayMs    int
	)

	if at = attacks[btypes.AttackTypeMessagePolicy]; at == nil || at.MessagePolicyParams == nil {
		return shouldSend
	}

	curView := c.currentView()
	sequence := new(big.Int).Set(curView.Sequence)
	round := new(big.Int).Set(curView.Round)

	for _, field := range at.MessagePolicyParams.Fields {
		switch field.Target {
		case btypes.TargetMsgPolicySendOriginal:
			if v, ok := field.Value.(bool); ok {
				shouldSend = v
			} else {
				log.Error("BYZ: Invalid value for policy.original", "value", field.Value)
			}
		case btypes.TargetMsgPolicyDelay:
			if v, ok := field.Value.(uint64); ok {
				delayMs = int(v)
			} else {
				log.Error("BYZ: Invalid value for policy.delay", "value", field.Value)
			}
		}
	}

	if !shouldSend {
		return shouldSend // skip sending original message
	}

	log.Info("BYZ: byzantine attack triggered",
		"name", at.NAME,
		"uid", at.UID,
		"seq", sequence,
		"round", round,
		"delay(ms)", delayMs,
	)

	hook.MarkAttackExecuted(at.UID, sequence.Uint64())

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	return shouldSend
}
