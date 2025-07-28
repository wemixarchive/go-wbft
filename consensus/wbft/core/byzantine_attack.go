package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mrand "math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"

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

func (c *Core) applyByzantineRoundChangeMsg(at *btypes.ExecutableAttack, originRoundChange *wbfmessage.RoundChange) *wbfmessage.RoundChange {
	var fakedRoundChange *wbfmessage.RoundChange = nil

	if at != nil && at.FakeParams != nil {
		for _, field := range at.FakeParams.Fields {
			switch field.Target {
			case btypes.TargetMsgProposal:
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

// executeDosAttack executes DOS attack by flooding messages
func (c *Core) executeDosAttack(attack *btypes.ExecutableAttack, msgCode btypes.MessageCode, originalMsg wbfmessage.WBFTMessage) {
	if attack == nil || attack.DosParams == nil {
		return
	}

	// Generate and send multiple messages based on fields configuration
	for _, field := range attack.DosParams.Fields {
		switch field.Target {
		case "valid":
			c.sendValidDosMessage(msgCode, field, originalMsg)
		case "invalid":
			c.sendInvalidDosMessage(msgCode, field, originalMsg)
		default:
			log.Warn("[BYZ] Unknown DOS field target", "target", field.Target)
		}
	}
}

// sendValidDosMessage sends a valid message for DOS attack
func (c *Core) sendValidDosMessage(msgCode btypes.MessageCode, field btypes.Field, originalMsg wbfmessage.WBFTMessage) {
	switch msgCode {
	case btypes.MessageCodePrePrepare:
		if preprepare, ok := originalMsg.(*wbfmessage.Preprepare); ok {
			c.sendValidDosPrePrepare(field, preprepare)
		}
	case btypes.MessageCodePrepare:
		if prepare, ok := originalMsg.(*wbfmessage.Prepare); ok {
			c.sendValidDosPrepare(field, prepare)
		}
	case btypes.MessageCodeCommit:
		if commit, ok := originalMsg.(*wbfmessage.Commit); ok {
			c.sendValidDosCommit(field, commit)
		}
	case btypes.MessageCodeRoundChange:
		if roundChange, ok := originalMsg.(*wbfmessage.RoundChange); ok {
			c.sendValidDosRoundChange(field, roundChange)
		}
	default:
	}
}

// sendInvalidDosMessage sends an invalid message for DOS attack
func (c *Core) sendInvalidDosMessage(msgCode btypes.MessageCode, field btypes.Field, originalMsg wbfmessage.WBFTMessage) {
	switch msgCode {
	case btypes.MessageCodePrePrepare:
		if preprepare, ok := originalMsg.(*wbfmessage.Preprepare); ok {
			c.sendInvalidDosPrePrepare(field, preprepare)
		}
	case btypes.MessageCodePrepare:
		if prepare, ok := originalMsg.(*wbfmessage.Prepare); ok {
			c.sendInvalidDosPrepare(field, prepare)
		}
	case btypes.MessageCodeCommit:
		if commit, ok := originalMsg.(*wbfmessage.Commit); ok {
			c.sendInvalidDosCommit(field, commit)
		}
	case btypes.MessageCodeRoundChange:
		if roundChange, ok := originalMsg.(*wbfmessage.RoundChange); ok {
			c.sendInvalidDosRoundChange(field, roundChange)
		}
	default:
	}
}

// sendValidDosPrepare sends valid prepare messages for DOS
func (c *Core) sendValidDosPrepare(field btypes.Field, originalPrepare *wbfmessage.Prepare) {
	var messages []wbfmessage.WBFTMessage

	// Modify message based on field value
	switch field.Value {
	case "sequence":
		// Send with incremented sequence
		for i := uint64(1); i <= 1000; i++ {
			modifiedPrepare := originalPrepare.DeepCopy()
			modifiedPrepare.Sequence = originalPrepare.Sequence.Add(originalPrepare.Sequence, big.NewInt(int64(i)))
			messages = append(messages, modifiedPrepare)
		}
	case "round":
		// Send with different rounds
		for i := uint64(1); i <= 1000; i++ {
			modifiedPrepare := originalPrepare.DeepCopy()
			modifiedPrepare.Round = originalPrepare.Round.Add(originalPrepare.Round, big.NewInt(int64(i)))
			messages = append(messages, modifiedPrepare)
		}
	default:
		// Send multiple copies of the same message
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
	}
}

// sendInvalidDosPrepare sends invalid prepare messages for DOS
func (c *Core) sendInvalidDosPrepare(field btypes.Field, originalPrepare *wbfmessage.Prepare) {
	// Pre-generate messages for faster flooding
	var messages [][]byte

	switch field.Value {
	case "signature":
		// Pre-generate messages with invalid signatures
		for i := 0; i < 1000; i++ {
			invalidPrepare := originalPrepare.DeepCopy()
			invalidSig := make([]byte, 65)
			rand.Read(invalidSig)
			invalidPrepare.SetSignature(invalidSig)

			if data, err := rlp.EncodeToBytes(invalidPrepare); err == nil {
				messages = append(messages, data)
			}
		}

	case "blockHash":
		// Pre-generate messages with random block hashes
		for i := 0; i < 1000; i++ {
			prepare := originalPrepare.DeepCopy()
			randomHash := common.Hash{}
			rand.Read(randomHash[:])
			prepare.Digest = randomHash

			// Sign the message
			if payload, err := prepare.EncodePayloadForSigning(); err == nil {
				if sig, err := c.backend.Sign(payload); err == nil {
					prepare.SetSignature(sig)
					if data, err := rlp.EncodeToBytes(prepare); err == nil {
						messages = append(messages, data)
					}
				}
			}
		}

	case "random":
		// Pre-generate completely random messages
		for i := 0; i < 1000; i++ {
			prepare := &wbfmessage.Prepare{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: big.NewInt(int64(mrand.Uint64())),
					Round:    big.NewInt(int64(mrand.Uint64())),
				},
				Digest:      common.Hash{},
				PrepareSeal: make([]byte, 65),
			}
			prepare.SetSource(c.Address())

			if data, err := rlp.EncodeToBytes(prepare); err == nil {
				messages = append(messages, data)
			}
		}
	}

	// Broadcast all messages rapidly
	for _, data := range messages {
		c.backend.Broadcast(c.valSet, originalPrepare.Code(), data)
	}
}

// sendValidDosCommit sends valid commit messages for DOS
func (c *Core) sendValidDosCommit(field btypes.Field, originalCommit *wbfmessage.Commit) {
	var messages []wbfmessage.WBFTMessage

	switch field.Value {
	case "sequence":
		for i := uint64(1); i <= 1000; i++ {
			modifiedCommit := &wbfmessage.Commit{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: new(big.Int).Add(originalCommit.Sequence, big.NewInt(int64(i))),
					Round:    originalCommit.Round,
				},
				Digest:     originalCommit.Digest,
				CommitSeal: originalCommit.CommitSeal,
			}
			modifiedCommit.SetSource(c.Address())
			messages = append(messages, modifiedCommit)
		}
	case "round":
		for i := uint64(1); i <= 1000; i++ {
			modifiedCommit := &wbfmessage.Commit{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: originalCommit.Sequence,
					Round:    new(big.Int).Add(originalCommit.Round, big.NewInt(int64(i))),
				},
				Digest:     originalCommit.Digest,
				CommitSeal: originalCommit.CommitSeal,
			}
			modifiedCommit.SetSource(c.Address())
			messages = append(messages, modifiedCommit)
		}
	default:

	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
	}
}

// sendInvalidDosCommit sends invalid commit messages for DOS
func (c *Core) sendInvalidDosCommit(field btypes.Field, originalCommit *wbfmessage.Commit) {
	// Pre-generate messages for faster flooding
	var messages [][]byte

	switch field.Value {
	case "signature":
		// Pre-generate messages with invalid signatures
		for i := 0; i < 50; i++ {
			commit := &wbfmessage.Commit{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: originalCommit.Sequence,
					Round:    originalCommit.Round,
				},
				Digest:     originalCommit.Digest,
				CommitSeal: make([]byte, 65),
			}
			commit.SetSource(c.Address())

			// Invalid signature and seal
			invalidSig := make([]byte, 65)
			rand.Read(invalidSig)
			commit.SetSignature(invalidSig)
			rand.Read(commit.CommitSeal)

			if data, err := rlp.EncodeToBytes(commit); err == nil {
				messages = append(messages, data)
			}
		}

	case "blockHash":
		// Pre-generate messages with random block hashes
		for i := 0; i < 50; i++ {
			randomHash := common.Hash{}
			rand.Read(randomHash[:])

			commit := &wbfmessage.Commit{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: originalCommit.Sequence,
					Round:    originalCommit.Round,
				},
				Digest:     randomHash,
				CommitSeal: originalCommit.CommitSeal,
			}
			commit.SetSource(c.Address())

			// Sign the message
			if payload, err := commit.EncodePayloadForSigning(); err == nil {
				if sig, err := c.backend.Sign(payload); err == nil {
					commit.SetSignature(sig)
					if data, err := rlp.EncodeToBytes(commit); err == nil {
						messages = append(messages, data)
					}
				}
			}
		}

	case "random":
		// Pre-generate completely random messages
		for i := 0; i < 50; i++ {
			commit := &wbfmessage.Commit{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: big.NewInt(int64(mrand.Uint64())),
					Round:    big.NewInt(int64(mrand.Uint64())),
				},
				Digest:     common.Hash{},
				CommitSeal: make([]byte, 65),
			}
			commit.SetSource(c.Address())

			if data, err := rlp.EncodeToBytes(commit); err == nil {
				messages = append(messages, data)
			}
		}
	}

	// Broadcast all messages rapidly
	for _, data := range messages {
		c.backend.Broadcast(c.valSet, originalCommit.Code(), data)
	}
}

// sendValidDosPrePrepare sends valid preprepare messages for DOS (only if proposer)
func (c *Core) sendValidDosPrePrepare(field btypes.Field, originalPreprepare *wbfmessage.Preprepare) {
	if !c.IsProposer() {
		log.Warn("[BYZ] Cannot send PrePrepare DOS attack: not proposer")
		return
	}

	var messages []wbfmessage.WBFTMessage

	switch field.Value {
	case "sequence":
		// This is dangerous as it can break consensus
		log.Warn("[BYZ] DOS PrePrepare with sequence modification not recommended")
	case "round":
		// Send with future rounds
		for i := uint64(1); i <= 1000; i++ {
			// Create preprepare with future round
			modifiedPreprepare := &wbfmessage.Preprepare{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: originalPreprepare.Sequence,
					Round:    new(big.Int).Add(originalPreprepare.Round, big.NewInt(int64(i))),
				},
				Proposal:                  originalPreprepare.Proposal,
				JustificationRoundChanges: originalPreprepare.JustificationRoundChanges,
				JustificationPrepares:     originalPreprepare.JustificationPrepares,
			}
			modifiedPreprepare.SetSource(c.Address())
			messages = append(messages, modifiedPreprepare)
		}
	default:
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
	}
}

// sendInvalidDosPrePrepare sends invalid preprepare messages for DOS
func (c *Core) sendInvalidDosPrePrepare(field btypes.Field, originalPreprepare *wbfmessage.Preprepare) {
	var messages [][]byte

	// Create a fake block for invalid preprepare
	fakeBlock := types.NewBlock(
		&types.Header{
			Number: originalPreprepare.Sequence,
			Time:   uint64(time.Now().Unix()),
		},
		nil, nil, nil, trie.NewStackTrie(nil),
	)

	switch field.Value {
	case "signature":
		for i := 0; i < 1000; i++ {
			// Create preprepare with invalid signature
			preprepare := &wbfmessage.Preprepare{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: originalPreprepare.Sequence,
					Round:    originalPreprepare.Round,
				},
				Proposal: originalPreprepare.Proposal,
			}
			preprepare.SetSource(c.Address())

			// Invalid signature
			invalidSig := make([]byte, 65)
			rand.Read(invalidSig)
			preprepare.SetSignature(invalidSig)

			if data, err := rlp.EncodeToBytes(preprepare); err == nil {
				messages = append(messages, data)
			}
		}

	case "blockHash":
		// Send with mismatched block hash
		for i := 0; i < 1000; i++ {
			modifiedBlock := types.NewBlock(
				&types.Header{
					Number:     originalPreprepare.Sequence,
					Time:       uint64(time.Now().Unix()),
					ParentHash: common.Hash{byte(i)}, // Different parent hash each time
				},
				nil, nil, nil, trie.NewStackTrie(nil),
			)
			preprepare := &wbfmessage.Preprepare{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: originalPreprepare.Sequence,
					Round:    originalPreprepare.Round,
				},
				Proposal: modifiedBlock,
			}
			preprepare.SetSource(c.Address())

			if data, err := rlp.EncodeToBytes(preprepare); err == nil {
				messages = append(messages, data)
			}
		}

	case "random":
		for i := 0; i < 1000; i++ {
			preprepare := &wbfmessage.Preprepare{
				CommonPayload: wbfmessage.CommonPayload{
					Sequence: big.NewInt(int64(mrand.Uint64())),
					Round:    big.NewInt(int64(mrand.Uint64())),
				},
				Proposal: fakeBlock,
			}
			preprepare.SetSource(c.Address())

			if data, err := rlp.EncodeToBytes(preprepare); err == nil {
				messages = append(messages, data)
			}
		}
	}

	for _, data := range messages {
		c.backend.Broadcast(c.valSet, originalPreprepare.Code(), data)
	}
}

// sendValidDosRoundChange sends valid round change messages for DOS
func (c *Core) sendValidDosRoundChange(field btypes.Field, originalRoundChange *wbfmessage.RoundChange) {
	var messages []wbfmessage.WBFTMessage

	switch field.Value {
	case "sequence":
		// Not recommended for round change
		log.Warn("[BYZ] DOS RoundChange with sequence modification not implemented")
	case "round":
		// Send round changes for future rounds
		for i := uint64(1); i <= 10000; i++ {
			// Create round change with future round
			modifiedRoundChange := wbfmessage.NewRoundChange(originalRoundChange.Sequence,
				new(big.Int).Add(originalRoundChange.Round, big.NewInt(int64(i))),
				originalRoundChange.PreparedRound,
				nil)
			modifiedRoundChange.PreparedBlock = originalRoundChange.PreparedBlock
			modifiedRoundChange.PreparedDigest = originalRoundChange.PreparedDigest
			modifiedRoundChange.SetSource(c.Address())

			if originalRoundChange.Justification != nil {
				// Copy justification if exists
				modifiedRoundChange.Justification = make([]*wbfmessage.Prepare, len(originalRoundChange.Justification))
				copy(modifiedRoundChange.Justification, originalRoundChange.Justification)
			} else {
				// No justification, set to nil
				modifiedRoundChange.Justification = nil
			}

			messages = append(messages, modifiedRoundChange)
		}
	default:
		// Send multiple round changes for same round
	}

	for _, msg := range messages {
		c.signAndBroadcastMessage(msg)
	}
}

// sendInvalidDosRoundChange sends invalid round change messages for DOS
func (c *Core) sendInvalidDosRoundChange(field btypes.Field, originalRoundChange *wbfmessage.RoundChange) {
	// Pre-generate messages for faster flooding
	var messages [][]byte

	switch field.Value {
	case "signature":
		// Pre-generate messages with invalid signatures
		for i := 0; i < 1000; i++ {
			roundChange := &wbfmessage.RoundChange{
				SignedRoundChangePayload: wbfmessage.SignedRoundChangePayload{
					CommonPayload: wbfmessage.CommonPayload{
						Sequence: originalRoundChange.Sequence,
						Round:    originalRoundChange.Round,
					},
					PreparedRound:  originalRoundChange.PreparedRound,
					PreparedDigest: originalRoundChange.PreparedDigest,
				},
				PreparedBlock: originalRoundChange.PreparedBlock,
			}
			roundChange.SetSource(c.Address())

			// Invalid signature
			invalidSig := make([]byte, 65)
			rand.Read(invalidSig)
			roundChange.SetSignature(invalidSig)

			if data, err := rlp.EncodeToBytes(roundChange); err == nil {
				messages = append(messages, data)
			}
		}

	case "random":
		// Pre-generate completely random messages
		for i := 0; i < 1000; i++ {
			roundChange := &wbfmessage.RoundChange{
				SignedRoundChangePayload: wbfmessage.SignedRoundChangePayload{
					CommonPayload: wbfmessage.CommonPayload{
						Sequence: big.NewInt(int64(mrand.Uint64())),
						Round:    big.NewInt(int64(mrand.Uint64())),
					},
					PreparedRound:  nil,
					PreparedDigest: common.Hash{},
				},
				PreparedBlock: nil,
			}
			roundChange.SetSource(c.Address())

			if data, err := rlp.EncodeToBytes(roundChange); err == nil {
				messages = append(messages, data)
			}
		}

	default:
	}

	// Broadcast all messages rapidly
	for _, data := range messages {
		c.backend.Broadcast(c.valSet, originalRoundChange.Code(), data)
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
		log.Error("[BYZ] Message does not support signature operations")
		return
	}

	// Encode payload
	payload, err := msgWithSig.EncodePayloadForSigning()
	if err != nil {
		log.Error("[BYZ] Failed to encode message", "err", err)
		return
	}

	// Sign
	sig, err := c.backend.Sign(payload)
	if err != nil {
		log.Error("[BYZ] Failed to sign message", "err", err)
		return
	}

	msgWithSig.SetSignature(sig)

	// RLP encode and broadcast
	data, err := rlp.EncodeToBytes(msg)
	if err != nil {
		log.Error("[BYZ] Failed to RLP encode message", "err", err)
		return
	}

	if err := c.backend.Broadcast(c.valSet, msgWithSig.Code(), data); err != nil {
		log.Error("[BYZ] Failed to broadcast message", "err", err)
	}
}
