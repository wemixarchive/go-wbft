package attack

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// OmitAttack implements an attack where parts of messages are omitted
type OmitAttack struct {
	*baseAttack

	// Attack configuration
	messageCode types.MessageCode
	omitCommand types.OmitCommand
	omitCount   uint64
}

// NewOmitAttack creates a new omit attack
func NewOmitAttack(config *types.AttackConfig) (types.Attack, error) {
	//sequence := config.Sequence
	//round := config.Round

	// Parse message code
	codeStr, ok := config.Params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'code' parameter")
	}
	messageCode := types.ParseMessageCode(codeStr)

	// Parse omit command
	omitCommand, ok := config.Params["omitCommand"].(types.OmitCommand)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'omitCommand' parameter")
	}

	// Parse omit count
	omitCount, ok := config.Params["omitCount"].(uint64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'omitCount' parameter")
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

	attack := &OmitAttack{
		// TODO:
		// should refactoring
		baseAttack:  nil,
		messageCode: messageCode,
		omitCommand: omitCommand,
		omitCount:   omitCount,
	}

	return attack, nil
}

// Type returns the attack type
func (o *OmitAttack) Type() types.AttackType { return "" }

// Execute executes the fake attack
func (o *OmitAttack) Execute(ctx *types.AttackContext) error {
	// Check if this is the target message type
	//if ctx.Message == nil || ctx.Message.Code() != o.messageCode {
	//	return nil
	//}

	log.Info("Executing omit attack",
		//"name", o.Name(),
		//"sequence", ctx.Sequence,
		//"round", ctx.Round,
		"messageType", o.messageCode,
		"omitCommand", o.omitCommand,
		"omitCount", o.omitCount)
	//"targets", len(o.GetTargets()))

	// Log omit details for debugging
	log.Debug("Omit attack details",
		"commandDescription", o.getOmitCommandDescription(),
		"itemsToOmit", o.omitCount)

	// The actual message omission will be handled by the interceptor
	// This just marks the execution and provides the omission parameters
	//o.RecordExecution(true, nil)

	return nil
}

// GetMessageCode returns the target message code
func (o *OmitAttack) GetMessageCode() types.MessageCode {
	return o.messageCode
}

// GetOmitCommand returns the omit command
func (o *OmitAttack) GetOmitCommand() types.OmitCommand {
	return o.omitCommand
}

// GetOmitCount returns the number of items to omit
func (o *OmitAttack) GetOmitCount() uint64 {
	return o.omitCount
}

// ShouldOmitField checks if a specific field should be omitted
func (o *OmitAttack) ShouldOmitField(fieldType string) bool {
	switch {
	case o.messageCode.HasCode(types.MessageCodePrePrepare):
		return o.shouldOmitPrePrepareField(fieldType)
	case o.messageCode.HasCode(types.MessageCodePropagation):
		return o.shouldOmitPropagationField(fieldType)
	case o.messageCode.HasCode(types.MessageCodeRoundChangePrePrepare):
		return o.shouldOmitRoundChangePrePrepareField(fieldType)
	default:
		return false
	}
}

// shouldOmitPrePrepareField checks omission for PrePrepare messages
func (o *OmitAttack) shouldOmitPrePrepareField(fieldType string) bool {
	switch o.omitCommand {
	case types.OmitPrevPrepareSeal:
		return fieldType == "PrevPrepareSeal"
	case types.OmitPrevCommitSeal:
		return fieldType == "PrevCommitSeal"
	default:
		return false
	}
}

// shouldOmitPropagationField checks omission for Propagation messages
func (o *OmitAttack) shouldOmitPropagationField(fieldType string) bool {
	switch o.omitCommand {
	case types.OmitPrepareSeal:
		return fieldType == "PrepareSeal"
	case types.OmitCommitSeal:
		return fieldType == "CommitSeal"
	default:
		return false
	}
}

// shouldOmitRoundChangePrePrepareField checks omission for RoundChange-PrePrepare messages
func (o *OmitAttack) shouldOmitRoundChangePrePrepareField(fieldType string) bool {
	switch o.omitCommand {
	case types.OmitRoundChangeMessages:
		return fieldType == "RoundChangeMessages"
	case types.OmitPrepareMessages:
		return fieldType == "PrepareMessages"
	default:
		return false
	}
}

// getOmitCommandDescription returns human-readable description of omit command
func (o *OmitAttack) getOmitCommandDescription() string {
	switch {
	case o.messageCode.HasCode(types.MessageCodePrePrepare):
		switch o.omitCommand {
		case types.OmitPrevPrepareSeal:
			return "Omit previous prepare seal"
		case types.OmitPrevCommitSeal:
			return "Omit previous commit seal"
		}
	case o.messageCode.HasCode(types.MessageCodePropagation):
		switch o.omitCommand {
		case types.OmitPrepareSeal:
			return "Omit prepare seal"
		case types.OmitCommitSeal:
			return "Omit commit seal"
		}
	case o.messageCode.HasCode(types.MessageCodeRoundChangePrePrepare):
		switch o.omitCommand {
		case types.OmitRoundChangeMessages:
			return "Omit round change messages"
		case types.OmitPrepareMessages:
			return "Omit prepare messages"
		}
	}
	return fmt.Sprintf("Unknown command %d for message type %s", o.omitCommand, o.messageCode)
}

// ApplyOmission applies the omission to a message payload
func (o *OmitAttack) ApplyOmission(originalPayload []byte) ([]byte, error) {
	// This is a placeholder for the actual omission logic
	// In a real implementation, this would parse the message,
	// remove the specified fields, and re-encode

	log.Debug("Applying omission to message",
		"originalSize", len(originalPayload),
		"omitCommand", o.omitCommand,
		"omitCount", o.omitCount)

	// For now, return the original payload
	// Real implementation would modify the message structure
	return originalPayload, nil
}

// validateOmitCommand validates if the omit command is valid for the message type
func validateOmitCommand(messageCode types.MessageCode, omitCommand types.OmitCommand) error {
	switch {
	case messageCode.HasCode(types.MessageCodePrePrepare):
		if omitCommand != types.OmitPrevPrepareSeal && omitCommand != types.OmitPrevCommitSeal {
			return fmt.Errorf("invalid omit command %d for PrePrepare message", omitCommand)
		}
	case messageCode.HasCode(types.MessageCodePropagation):
		if omitCommand != types.OmitPrepareSeal && omitCommand != types.OmitCommitSeal {
			return fmt.Errorf("invalid omit command %d for Propagation message", omitCommand)
		}
	case messageCode.HasCode(types.MessageCodeRoundChangePrePrepare):
		if omitCommand != types.OmitRoundChangeMessages && omitCommand != types.OmitPrepareMessages {
			return fmt.Errorf("invalid omit command %d for RoundChange-PrePrepare message", omitCommand)
		}
	default:
		return fmt.Errorf("omit attack not supported for message type: %s", messageCode)
	}
	return nil
}

// GetSupportedOmitCommands returns supported omit commands for a message type
func GetSupportedOmitCommands(messageCode types.MessageCode) []types.OmitCommand {
	switch {
	case messageCode.HasCode(types.MessageCodePrePrepare):
		return []types.OmitCommand{types.OmitPrevPrepareSeal, types.OmitPrevCommitSeal}
	case messageCode.HasCode(types.MessageCodePropagation):
		return []types.OmitCommand{types.OmitPrepareSeal, types.OmitCommitSeal}
	case messageCode.HasCode(types.MessageCodeRoundChangePrePrepare):
		return []types.OmitCommand{types.OmitRoundChangeMessages, types.OmitPrepareMessages}
	default:
		return []types.OmitCommand{}
	}
}

func (o *OmitAttack) ID() string                                   { return "" }
func (o *OmitAttack) SetID(id uint64)                              {}
func (o *OmitAttack) CheckCondition(ctx *types.AttackContext) bool { return false }
func (o *OmitAttack) Configure(params interface{}) error           { return nil }
func (o *OmitAttack) Validate() error                              { return nil }
func (o *OmitAttack) Reset() error                                 { return nil }
func (o *OmitAttack) ToJSON() ([]byte, error)                      { return nil, nil }
func (o *OmitAttack) FromJSON(data []byte) error                   { return nil }

func (s *OmitAttack) Config() types.AttackConfig                                { return types.AttackConfig{} }
func (s *OmitAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool { return false }
func (s *OmitAttack) RequiresData() []types.DataRequirement                     { return nil }

// OmitAttackFactory creates omit attacks
func OmitAttackFactory(config *types.AttackConfig) (types.Attack, error) {
	return NewOmitAttack(config)
}
