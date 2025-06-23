package attack

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// SilentAttack implements an attack where the node stays silent for specific messages
type SilentAttack struct {
	*baseAttack

	// Attack configuration
	messageCode types.MessageCode
	direction   types.MessageDirection // Direction (send/receive/both)
}

// NewSilentAttack creates a new silent attack
func NewSilentAttack(config *types.AttackConfig) (types.Attack, error) {
	//sequence := config.Sequence
	//round := config.Round

	// Parse message code
	//codeStr, ok := config.Params["code"].(string)
	//if !ok {
	//	return nil, fmt.Errorf("missing or invalid 'code' parameter")
	//}
	//messageCode := types.ParseMessageCode(codeStr)
	// Parse message code
	var messageCode types.MessageCode
	var messageCodes []types.MessageCode // 개별 메시지 코드들을 저장

	// uint64 타입으로 처리 (비트 조합)
	if codeUint, ok := config.Params["code"].(uint64); ok {
		messageCode = types.MessageCode(codeUint)

		// 비트 마스크로 개별 메시지 타입 추출
		if codeUint&uint64(types.MessageCodePrePrepare) != 0 {
			messageCodes = append(messageCodes, types.MessageCodePrePrepare)
			log.Debug("Silent attack includes PrePrepare")
		}
		if codeUint&uint64(types.MessageCodePrepare) != 0 {
			messageCodes = append(messageCodes, types.MessageCodePrepare)
			log.Debug("Silent attack includes Prepare")
		}
		if codeUint&uint64(types.MessageCodeCommit) != 0 {
			messageCodes = append(messageCodes, types.MessageCodeCommit)
			log.Debug("Silent attack includes Commit")
		}
		if codeUint&uint64(types.MessageCodeRoundChange) != 0 {
			messageCodes = append(messageCodes, types.MessageCodeRoundChange)
			log.Debug("Silent attack includes RoundChange")
		}
		if codeUint&uint64(types.MessageCodeRoundChangePrePrepare) != 0 {
			messageCodes = append(messageCodes, types.MessageCodeRoundChangePrePrepare)
			log.Debug("Silent attack includes RoundChangePrePrepare")
		}
		if codeUint&uint64(types.MessageCodePropagation) != 0 {
			messageCodes = append(messageCodes, types.MessageCodePropagation)
			log.Debug("Silent attack includes Propagation")
		}

		log.Info("Parsed message codes from uint64",
			"code", codeUint,
			"binary", fmt.Sprintf("0b%08b", codeUint),
			"messageTypes", messageCodes)

	} else {
		return nil, fmt.Errorf("missing or invalid 'code' parameter")
	}

	// Parse direction
	direction := types.MessageDirectionSend
	if dir, ok := config.Params["direction"].(string); ok {
		switch dir {
		case "send":
			direction = types.MessageDirectionSend
		case "receive":
			direction = types.MessageDirectionReceive
		case "both":
			direction = types.MessageDirectionBoth
		}
	}

	// Parse targets
	var targets []common.Address
	if targetList, ok := config.Params["targets"].([]common.Address); ok {
		targets = targetList
	} else if targetList, ok := config.Params["targets"].([]interface{}); ok {
		for _, target := range targetList {
			if addr, ok := target.(common.Address); ok {
				targets = append(targets, addr)
			} else if addrStr, ok := target.(string); ok {
				if common.IsHexAddress(addrStr) {
					targets = append(targets, common.HexToAddress(addrStr))
				}
			}
		}
	}

	attack := &SilentAttack{
		baseAttack:  nil, // TODO: should refactoring
		messageCode: messageCode,
		direction:   direction,
	}

	return attack, nil
}

// Type returns the attack type
func (s *SilentAttack) Type() types.AttackType {
	return ""
}

// Execute executes the silent attack

func (s *SilentAttack) Execute(ctx *types.AttackContext) error {
	// Check if this message type should be silenced
	//if ctx.Message == nil || ctx.Message.Code() != s.messageCode {
	//	return nil // This message type is not targeted
	//}

	// Check if we should silence based on direction
	shouldSilence := false
	switch s.direction {
	case types.MessageDirectionSend:
		shouldSilence = true // Always silence outgoing messages
	case types.MessageDirectionReceive:
		shouldSilence = false // This is handled by the interceptor for incoming
	case types.MessageDirectionBoth:
		shouldSilence = true // Silence both directions
	}

	if !shouldSilence {
		return nil
	}

	//log.Info("Executing silent attack",
	//	"name", s.Name(),
	//	"sequence", ctx.Sequence,
	//	"round", ctx.Round,
	//	"messageType", s.messageCode,
	//	"direction", s.direction,
	//	"targets", len(s.GetTargets()))
	//
	//// Record successful execution
	//s.RecordExecution(true, nil)

	// Return error to indicate message should be silenced
	return fmt.Errorf("message silenced by silent attack")
}

// GetMessageCode returns the message code bitmap
func (s *SilentAttack) GetMessageCode() types.MessageCode {
	return s.messageCode
}

// GetDirection returns the message direction
func (s *SilentAttack) GetDirection() types.MessageDirection {
	return s.direction
}

// ShouldSilenceMessage checks if a specific message should be silenced
func (s *SilentAttack) ShouldSilenceMessage(msgType types.MessageCode, direction types.MessageDirection) bool {
	// Check if message type matches
	if s.messageCode&msgType == 0 {
		return false
	}

	// Check direction
	switch s.direction {
	case types.MessageDirectionSend:
		return direction == types.MessageDirectionSend
	case types.MessageDirectionReceive:
		return direction == types.MessageDirectionReceive
	case types.MessageDirectionBoth:
		return true
	default:
		return false
	}
}

// ShouldSilence checks if a message should be silenced based on its code and direction
func (s *SilentAttack) ShouldSilence(msgCode types.MessageCode, direction types.MessageDirection) bool {
	// Check if message type matches
	if s.messageCode&msgCode == 0 {
		return false
	}

	// Check direction
	switch s.direction {
	case types.MessageDirectionSend:
		return direction == types.MessageDirectionSend
	case types.MessageDirectionReceive:
		return direction == types.MessageDirectionReceive
	case types.MessageDirectionBoth:
		return true
	default:
		return false
	}
}

func (s *SilentAttack) ID() string                                   { return "" }
func (s *SilentAttack) SetID(id uint64)                              {}
func (s *SilentAttack) CheckCondition(ctx *types.AttackContext) bool { return false }
func (s *SilentAttack) Configure(params interface{}) error           { return nil }
func (s *SilentAttack) Validate() error                              { return nil }
func (s *SilentAttack) Reset() error                                 { return nil }
func (s *SilentAttack) ToJSON() ([]byte, error)                      { return nil, nil }
func (s *SilentAttack) FromJSON(data []byte) error                   { return nil }

func (s *SilentAttack) Config() types.AttackConfig                                { return types.AttackConfig{} }
func (s *SilentAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool { return false }
func (s *SilentAttack) RequiresData() []types.DataRequirement                     { return nil }

// SilentAttackFactory creates silent attacks
func SilentAttackFactory(config *types.AttackConfig) (types.Attack, error) {
	return NewSilentAttack(config)
}
