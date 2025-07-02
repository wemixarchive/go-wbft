package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MessageDirection represents message direction
type MessageDirection uint64

const (
	MessageDirectionSend    MessageDirection = 1
	MessageDirectionReceive MessageDirection = 2
	MessageDirectionBoth    MessageDirection = 3
)

// MessageCode represents WBFT message types
type MessageCode uint64

const (
	MessageCodePrePrepare  MessageCode = 1 << iota // 1
	MessageCodePrepare                             // 2
	MessageCodeCommit                              // 4
	MessageCodeRoundChange                         // 8
	MessageCodePropagation                         // 16
)

// MessageCodeToWBFT Byzantine message code to WBFT code mapping
var MessageCodeToWBFT = map[MessageCode]uint64{
	MessageCodePrePrepare:  WBFTPrePrepareCode,
	MessageCodePrepare:     WBFTPrepareCode,
	MessageCodeCommit:      WBFTCommitCode,
	MessageCodeRoundChange: WBFTRoundChangeCode,
}

func (mc *MessageCode) UnmarshalJSON(data []byte) error {
	var val interface{}
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	switch v := val.(type) {
	case float64:
		*mc = MessageCode(uint64(v))
	case string:
		*mc = ParseStringToMessageCode(v)
		if *mc == 0 {
			return fmt.Errorf("invalid message code string: %s", v)
		}
	default:
		return fmt.Errorf("invalid message code type: %T", v)
	}
	return nil
}

func (mc *MessageCode) Has(flag MessageCode) bool {
	return (*mc & flag) != 0
}

func (mc *MessageCode) Add(flag MessageCode) MessageCode {
	return *mc | flag
}

func (mc *MessageCode) Remove(flag MessageCode) MessageCode {
	return *mc & ^flag
}

func (mc *MessageCode) String() string {
	if *mc == 0 {
		return "none"
	}

	var codes []string
	if mc.Has(MessageCodePrePrepare) {
		codes = append(codes, "PrePrepare")
	}
	if mc.Has(MessageCodePrepare) {
		codes = append(codes, "Prepare")
	}
	if mc.Has(MessageCodeCommit) {
		codes = append(codes, "Commit")
	}
	if mc.Has(MessageCodeRoundChange) {
		codes = append(codes, "RoundChange")
	}
	if mc.Has(MessageCodePropagation) {
		codes = append(codes, "Propagation")
	}

	return strings.Join(codes, "|")
}

// ParseMessageCode parses various formats of message code
func ParseMessageCode(val interface{}) MessageCode {
	switch v := val.(type) {
	case float64:
		return MessageCode(uint64(v))
	case int:
		return MessageCode(uint64(v))
	case uint64:
		return MessageCode(v)
	case string:
		return ParseStringToMessageCode(v)
	case MessageCode:
		return v
	default:
		return 0
	}
}

// ParseStringToMessageCode converts string to MessageCode
func ParseStringToMessageCode(code string) MessageCode {
	if num, err := strconv.ParseUint(code, 10, 64); err == nil {
		return MessageCode(num)
	}
	v := strings.ToLower(code)
	switch v {
	case "PrePrepare":
		return MessageCodePrePrepare
	case "Prepare":
		return MessageCodePrepare
	case "Commit":
		return MessageCodeCommit
	case "RoundChange":
		return MessageCodeRoundChange
	case "Propagation":
		return MessageCodePropagation
	default:
		if strings.Contains(v, "|") {
			var result MessageCode
			parts := strings.Split(v, "|")
			for _, part := range parts {
				result = result.Add(ParseStringToMessageCode(strings.TrimSpace(part)))
			}
			return result
		}
		return 0
	}
}

func ValidateMessageCode(code MessageCode) bool {
	validMask := MessageCodePrePrepare | MessageCodePrepare | MessageCodeCommit |
		MessageCodeRoundChange | MessageCodePropagation
	return code != 0 && (code & ^validMask) == 0
}

// QBFTMessage represents a QBFT consensus message
type QBFTMessage struct {
	Code          MessageCode    `json:"code"`
	Sequence      uint64         `json:"sequence"`
	Round         uint64         `json:"round"`
	Address       common.Address `json:"address"`
	Signature     []byte         `json:"signature"`
	CommittedSeal []byte         `json:"committed_seal,omitempty"`
	Proposal      *types.Block   `json:"proposal,omitempty"`
	Hash          common.Hash    `json:"hash"`
}

// StoredMessage represents a message stored for potential replay
type StoredMessage struct {
	Message    *QBFTMessage   `json:"message"`
	ReceivedAt time.Time      `json:"received_at"`
	FromPeer   common.Address `json:"from_peer"`
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

	TamperDigest TamperTarget = "Digest"
	TamperReward TamperTarget = "Reward"

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
