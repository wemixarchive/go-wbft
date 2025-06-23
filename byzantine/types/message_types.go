package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	wbftmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// MessageCode represents QBFT message types
type MessageCode uint64

const (
	MessageCodePrePrepare            MessageCode = 1 << 0 // 1
	MessageCodePrepare               MessageCode = 1 << 1 // 2
	MessageCodeCommit                MessageCode = 1 << 2 // 4
	MessageCodeRoundChange           MessageCode = 1 << 3 // 8
	MessageCodeRoundChangePrePrepare MessageCode = 1 << 4 // 16
	MessageCodePropagation           MessageCode = 1 << 5 // 32
)

// MessageDirection represents message direction
type MessageDirection uint64

const (
	MessageDirectionSend    MessageDirection = 1
	MessageDirectionReceive MessageDirection = 2
	MessageDirectionBoth    MessageDirection = 3
)

//// Message represents a QBFT message
//type Message interface {
//	// Basic properties
//	Code() MessageCode
//	Sequence() uint64
//	Round() uint64
//	From() common.Address
//	To() []common.Address
//
//	// Message content
//	Payload() []byte
//	Signature() []byte
//
//	// Validation
//	Validate() error
//	VerifySignature() error
//
//	// Encoding/decoding
//	Encode() ([]byte, error)
//	Decode([]byte) error
//
//	// Block reference
//	Block() *types.Block
//
//	// QBFT message conversion
//	QBFTMessage() qbftmessage.QBFTMessage
//}

// BaseMessage represents a base message structure
type BaseMessage struct {
	MsgCode      MessageCode
	MsgSeq       uint64
	MsgRound     uint64
	MsgFrom      common.Address
	MsgTo        []common.Address
	MsgData      []byte
	MsgSignature []byte
}

// Code returns the message code
func (m *BaseMessage) Code() MessageCode {
	return m.MsgCode
}

// Sequence returns the message sequence
func (m *BaseMessage) Sequence() uint64 {
	return m.MsgSeq
}

// Round returns the message round
func (m *BaseMessage) Round() uint64 {
	return m.MsgRound
}

// From returns the message sender
func (m *BaseMessage) From() common.Address {
	return m.MsgFrom
}

// To returns the message recipients
func (m *BaseMessage) To() []common.Address {
	return m.MsgTo
}

// Payload returns the message payload
func (m *BaseMessage) Payload() []byte {
	return m.MsgData
}

// Signature returns the message signature
func (m *BaseMessage) Signature() []byte {
	return m.MsgSignature
}

// Validate performs basic message validation
func (m *BaseMessage) Validate() error {
	// Basic validation logic
	if m.MsgCode == 0 {
		return ErrInvalidMessageCode
	}
	if m.MsgSeq == 0 {
		return ErrInvalidSequence
	}
	if m.MsgFrom == (common.Address{}) {
		return ErrInvalidSender
	}
	return nil
}

// VerifySignature verifies the message signature
func (m *BaseMessage) VerifySignature() error {
	// Signature verification logic
	if len(m.MsgSignature) == 0 {
		return ErrMissingSignature
	}
	return nil
}

// Encode encodes the message to bytes
func (m *BaseMessage) Encode() ([]byte, error) {
	data := struct {
		Code      MessageCode
		Sequence  uint64
		Round     uint64
		From      common.Address
		To        []common.Address
		Data      []byte
		Signature []byte
	}{
		Code:      m.MsgCode,
		Sequence:  m.MsgSeq,
		Round:     m.MsgRound,
		From:      m.MsgFrom,
		To:        m.MsgTo,
		Data:      m.MsgData,
		Signature: m.MsgSignature,
	}
	return rlp.EncodeToBytes(data)
}

// Decode decodes the message from bytes
func (m *BaseMessage) Decode(data []byte) error {
	var decoded struct {
		Code      MessageCode
		Sequence  uint64
		Round     uint64
		From      common.Address
		To        []common.Address
		Data      []byte
		Signature []byte
	}
	if err := rlp.DecodeBytes(data, &decoded); err != nil {
		return err
	}

	m.MsgCode = decoded.Code
	m.MsgSeq = decoded.Sequence
	m.MsgRound = decoded.Round
	m.MsgFrom = decoded.From

	if decoded.To != nil {
		m.MsgTo = make([]common.Address, len(decoded.To))
		copy(m.MsgTo, decoded.To)
	} else {
		m.MsgTo = nil
	}

	if decoded.Data != nil {
		m.MsgData = make([]byte, len(decoded.Data))
		copy(m.MsgData, decoded.Data)
	} else {
		m.MsgData = nil
	}

	if decoded.Signature != nil {
		m.MsgSignature = make([]byte, len(decoded.Signature))
		copy(m.MsgSignature, decoded.Signature)
	} else {
		m.MsgSignature = nil
	}

	return nil
}

// Block returns the associated block
func (m *BaseMessage) Block() *types.Block {
	return nil
}

// QBFTMessage returns the underlying QBFT message
func (m *BaseMessage) QBFTMessage() wbftmessage.WBFTMessage {
	return nil
}

func (m *BaseMessage) Clone() Message {
	return &BaseMessage{
		MsgCode:      m.MsgCode,
		MsgSeq:       m.MsgSeq,
		MsgRound:     m.MsgRound,
		MsgFrom:      m.MsgFrom,
		MsgTo:        append([]common.Address{}, m.MsgTo...),
		MsgData:      append([]byte{}, m.MsgData...),
		MsgSignature: append([]byte{}, m.MsgSignature...),
	}
}

// MessageInfo struct for message tracking
type MessageInfo struct {
	Type      MessageCode
	Sequence  uint64
	Round     uint64
	From      common.Address
	To        []common.Address
	Timestamp int64
	Size      int
}

// ParseMessageCode converts string to MessageCode
func ParseMessageCode(code string) MessageCode {
	switch code {
	case "PrePrepare":
		return MessageCodePrePrepare
	case "Prepare":
		return MessageCodePrepare
	case "Commit":
		return MessageCodeCommit
	case "RoundChange":
		return MessageCodeRoundChange
	case "RoundChangePrePrepare":
		return MessageCodeRoundChangePrePrepare
	case "Propagation":
		return MessageCodePropagation
	default:
		return 0
	}
}

// String converts MessageCode to string
func (m MessageCode) String() string {
	switch m {
	case MessageCodePrePrepare:
		return "PrePrepare"
	case MessageCodePrepare:
		return "Prepare"
	case MessageCodeCommit:
		return "Commit"
	case MessageCodeRoundChange:
		return "RoundChange"
	case MessageCodeRoundChangePrePrepare:
		return "RoundChangePrePrepare"
	case MessageCodePropagation:
		return "Propagation"
	default:
		return "Unknown"
	}
}

// HasCode checks if the message code has a specific code
func (m MessageCode) HasCode(code MessageCode) bool {
	return m&code != 0
}

// QBFTMessageWrapper wraps a QBFT message to implement the Message interface
type QBFTMessageWrapper struct {
	msg     wbftmessage.WBFTMessage
	block   *types.Block
	targets []common.Address
}

// NewQBFTMessageWrapper creates a new QBFT message wrapper
func NewQBFTMessageWrapper(msg wbftmessage.WBFTMessage, block *types.Block) *QBFTMessageWrapper {
	return &QBFTMessageWrapper{
		msg:   msg,
		block: block,
	}
}

// Code returns the message code
func (w *QBFTMessageWrapper) Code() MessageCode {
	return MessageCode(w.msg.Code())
}

// Sequence returns the message sequence
func (w *QBFTMessageWrapper) Sequence() uint64 {
	return w.msg.View().Sequence.Uint64()
}

// Round returns the message round
func (w *QBFTMessageWrapper) Round() uint64 {
	return w.msg.View().Round.Uint64()
}

// From returns the sender address
func (w *QBFTMessageWrapper) From() common.Address {
	return w.msg.Source()
}

// To returns the recipient addresses (empty for broadcast)
func (w *QBFTMessageWrapper) To() []common.Address {
	// Return actual targets if available in wrapper context
	if w.targets != nil {
		return w.targets
	}
	// Default to empty for broadcast
	return []common.Address{}
}

// Payload returns the message payload
func (w *QBFTMessageWrapper) Payload() []byte {
	switch msg := w.msg.(type) {
	case *wbftmessage.Preprepare:
		if msg.Proposal != nil {
			payload, _ := rlp.EncodeToBytes(msg.Proposal)
			return payload
		}
		return nil
	case *wbftmessage.Prepare:
		return msg.Digest.Bytes()
	case *wbftmessage.Commit:
		return msg.Digest.Bytes()
	case *wbftmessage.RoundChange:
		data := struct {
			PreparedRound  *big.Int
			PreparedDigest common.Hash
			PreparedBlock  *types.Block
		}{
			PreparedRound:  msg.PreparedRound,
			PreparedDigest: msg.PreparedDigest,
			PreparedBlock:  msg.PreparedBlock,
		}
		payload, _ := rlp.EncodeToBytes(data)
		return payload
	default:
		return nil
	}
}

// Signature returns the message signature
func (w *QBFTMessageWrapper) Signature() []byte {
	switch msg := w.msg.(type) {
	case *wbftmessage.Preprepare:
		return msg.Signature()
	case *wbftmessage.Prepare:
		return msg.Signature()
	case *wbftmessage.Commit:
		return msg.Signature()
	case *wbftmessage.RoundChange:
		return msg.Signature()
	default:
		return nil
	}
}

// Validate performs basic message validation
func (w *QBFTMessageWrapper) Validate() error {
	// QBFT messages are already validated
	return nil
}

// VerifySignature verifies the message signature
func (w *QBFTMessageWrapper) VerifySignature() error {
	// QBFT messages are already signature verified
	return nil
}

// Encode encodes the message to bytes
func (w *QBFTMessageWrapper) Encode() ([]byte, error) {
	return rlp.EncodeToBytes(w.msg)
}

// Decode decodes the message from bytes
func (w *QBFTMessageWrapper) Decode(data []byte) error {
	return rlp.DecodeBytes(data, w.msg)
}

// Block returns the associated block
func (w *QBFTMessageWrapper) Block() *types.Block {
	if pre, ok := w.msg.(*wbftmessage.Preprepare); ok && pre.Proposal != nil {
		if block, ok := pre.Proposal.(*types.Block); ok {
			return block
		}
	}
	return w.block
}

// QBFTMessage returns the wrapped QBFT message
func (w *QBFTMessageWrapper) QBFTMessage() wbftmessage.WBFTMessage {
	return w.msg
}

func (w *QBFTMessageWrapper) Clone() Message {
	if w == nil || w.msg == nil {
		return nil
	}

	var clonedMsg wbftmessage.WBFTMessage

	switch msg := w.msg.(type) {
	case *wbftmessage.Preprepare:
		cloned := wbftmessage.NewPreprepare(
			new(big.Int).Set(msg.Sequence),
			new(big.Int).Set(msg.Round),
			msg.Proposal,
		)

		if msg.JustificationRoundChanges != nil {
			cloned.JustificationRoundChanges = make([]*wbftmessage.SignedRoundChangePayload, len(msg.JustificationRoundChanges))
			for i, rc := range msg.JustificationRoundChanges {
				if rc != nil {
					// Copy SignedRoundChangePayload
					clonedRC := &wbftmessage.SignedRoundChangePayload{
						CommonPayload: wbftmessage.CommonPayload{
							Sequence: new(big.Int).Set(rc.Sequence),
							Round:    new(big.Int).Set(rc.Round),
						},
						PreparedRound:  nil,
						PreparedDigest: rc.PreparedDigest,
					}

					// Copy PreparedRound
					if rc.PreparedRound != nil {
						clonedRC.PreparedRound = new(big.Int).Set(rc.PreparedRound)
					}

					// Copy private field (signature, source)
					clonedRC.SetSignature(append([]byte(nil), rc.Signature()...))
					clonedRC.SetSource(rc.Source())

					cloned.JustificationRoundChanges[i] = clonedRC
				}
			}
		}

		if msg.JustificationPrepares != nil {
			cloned.JustificationPrepares = make([]*wbftmessage.Prepare, len(msg.JustificationPrepares))
			for i, p := range msg.JustificationPrepares {
				if p != nil {
					clonedPrepare := wbftmessage.NewPrepare(
						new(big.Int).Set(p.Sequence),
						new(big.Int).Set(p.Round),
						p.Digest,
						p.PrepareSeal,
					)
					clonedPrepare.SetSignature(append([]byte(nil), p.Signature()...))
					clonedPrepare.SetSource(p.Source())
					cloned.JustificationPrepares[i] = clonedPrepare
				}
			}
		}

		// Copy private field
		cloned.SetSignature(append([]byte(nil), msg.Signature()...))
		cloned.SetSource(msg.Source())

		clonedMsg = cloned

	case *wbftmessage.Prepare:
		cloned := wbftmessage.NewPrepare(
			new(big.Int).Set(msg.Sequence),
			new(big.Int).Set(msg.Round),
			msg.Digest,
			append([]byte(nil), msg.PrepareSeal...),
		)

		// Copy private field
		cloned.SetSignature(append([]byte(nil), msg.Signature()...))
		cloned.SetSource(msg.Source())

		clonedMsg = cloned

	case *wbftmessage.Commit:
		cloned := wbftmessage.NewCommit(
			new(big.Int).Set(msg.Sequence),
			new(big.Int).Set(msg.Round),
			msg.Digest,
			append([]byte(nil), msg.CommitSeal...),
		)

		// Copy private field
		cloned.SetSignature(append([]byte(nil), msg.Signature()...))
		cloned.SetSource(msg.Source())

		clonedMsg = cloned

	case *wbftmessage.RoundChange:
		var preparedRound *big.Int
		if msg.PreparedRound != nil {
			preparedRound = new(big.Int).Set(msg.PreparedRound)
		}

		cloned := wbftmessage.NewRoundChange(
			new(big.Int).Set(msg.Sequence),
			new(big.Int).Set(msg.Round),
			preparedRound,
			msg.PreparedBlock,
		)

		// Justification 깊은 복사
		if msg.Justification != nil {
			cloned.Justification = make([]*wbftmessage.Prepare, len(msg.Justification))
			for i, p := range msg.Justification {
				if p != nil {
					clonedPrepare := wbftmessage.NewPrepare(
						new(big.Int).Set(p.Sequence),
						new(big.Int).Set(p.Round),
						p.Digest,
						p.PrepareSeal,
					)
					clonedPrepare.SetSignature(append([]byte(nil), p.Signature()...))
					clonedPrepare.SetSource(p.Source())
					cloned.Justification[i] = clonedPrepare
				}
			}
		}

		// private 필드 복사
		cloned.SetSignature(append([]byte(nil), msg.Signature()...))
		cloned.SetSource(msg.Source())

		clonedMsg = cloned

	default:
		// 알 수 없는 메시지 타입의 경우 nil 반환
		return nil
	}

	// Wrapper 복사
	return &QBFTMessageWrapper{
		msg:     clonedMsg,
		block:   w.block,                                     // Block은 immutable이므로 shallow copy
		targets: append([]common.Address(nil), w.targets...), // targets 깊은 복사
	}
}

// TamperTarget represents fields that can be tampered
type TamperTarget string

const (
	// PrePrepare message targets
	TamperProposalHeaderCoinbase    TamperTarget = "Proposal.Header.Coinbase"
	TamperProposalHeaderNumber      TamperTarget = "Proposal.Header.Number"
	TamperProposalHeaderTime        TamperTarget = "Proposal.Header.Time"
	TamperProposalHeaderParentHash  TamperTarget = "Proposal.Header.ParentHash"
	TamperProposalHeaderStateRoot   TamperTarget = "Proposal.Header.StateRoot"
	TamperProposalHeaderTxHash      TamperTarget = "Proposal.Header.TxHash"
	TamperProposalHeaderReceiptHash TamperTarget = "Proposal.Header.ReceiptHash"
	TamperProposalHeaderBloom       TamperTarget = "Proposal.Header.Bloom"
	TamperProposalHeaderDifficulty  TamperTarget = "Proposal.Header.Difficulty"
	TamperProposalHeaderGasLimit    TamperTarget = "Proposal.Header.GasLimit"
	TamperProposalHeaderGasUsed     TamperTarget = "Proposal.Header.GasUsed"
	TamperProposalHeaderExtra       TamperTarget = "Proposal.Header.Extra"
	TamperProposalHeaderMixDigest   TamperTarget = "Proposal.Header.MixDigest"
	TamperProposalHeaderNonce       TamperTarget = "Proposal.Header.Nonce"

	// Transaction targets
	TamperTransactionSign    TamperTarget = "Transaction.Sign"
	TamperTransactionBalance TamperTarget = "Transaction.Balance"
	TamperTransactionNonce   TamperTarget = "Transaction.Nonce"
	TamperTransactionData    TamperTarget = "Transaction.Data"

	// Common message targets
	TamperMessageRound     TamperTarget = "Message.Round"
	TamperMessageSequence  TamperTarget = "Message.Sequence"
	TamperMessageSealType  TamperTarget = "Message.SealType"
	TamperMessageHeader    TamperTarget = "Message.Header"
	TamperMessageSignature TamperTarget = "Message.Signature"

	// RoundChange specific
	TamperRoundChangePreparedRound  TamperTarget = "RoundChange.PreparedRound"
	TamperRoundChangePreparedDigest TamperTarget = "RoundChange.PreparedDigest"

	// Seal targets
	TamperSealPrepare  TamperTarget = "Seal.Prepare"
	TamperSealCommit   TamperTarget = "Seal.Commit"
	TamperSealPrevious TamperTarget = "Seal.Previous"
)

// OmitCommand represents what to omit in messages
type OmitCommand uint64

const (
	// PrePrepare omit commands
	OmitPrevPrepareSeal OmitCommand = 1
	OmitPrevCommitSeal  OmitCommand = 2

	// Propagation omit commands
	OmitPrepareSeal OmitCommand = 3
	OmitCommitSeal  OmitCommand = 4

	// RoundChange-PrePrepare omit commands
	OmitRoundChangeMessages OmitCommand = 5
	OmitPrepareMessages     OmitCommand = 6
)
