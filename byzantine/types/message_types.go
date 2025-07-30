package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

func WBFTCodeToByzantineCode(code uint64) (MessageCode, error) {
	switch code {
	case WBFTPrePrepareCode:
		return MessageCodePrePrepare, nil
	case WBFTPrepareCode:
		return MessageCodePrepare, nil
	case WBFTCommitCode:
		return MessageCodeCommit, nil
	case WBFTRoundChangeCode:
		return MessageCodeRoundChange, nil
	default:
		return 0, fmt.Errorf("unknown message code: %d", code)
	}
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

// WBFTMessage represents a QBFT consensus message
type WBFTMessage struct {
	Code          MessageCode    `json:"code"`
	Sequence      uint64         `json:"sequence"`
	Round         uint64         `json:"round"`
	Address       common.Address `json:"address"`
	Signature     []byte         `json:"signature"`
	CommittedSeal []byte         `json:"committed_seal,omitempty"`
	Proposal      *types.Block   `json:"proposal,omitempty"`
	Hash          common.Hash    `json:"hash"`
}
