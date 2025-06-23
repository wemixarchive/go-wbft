package attack

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// FakeAttack implements an attack where fake messages are generated and sent
type FakeAttack struct {
	*baseAttack

	// Attack configuration
	messageCode types.MessageCode
	fakeMessage []byte // The fake message content
}

// NewFakeAttack creates a new fake attack
func NewFakeAttack(config *types.AttackConfig) (types.Attack, error) {
	//sequence := config.Sequence
	//round := config.Round

	// Parse message code
	codeStr, ok := config.Params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'code' parameter")
	}
	messageCode := types.ParseMessageCode(codeStr)

	// Parse fake message
	fakeMessage, ok := config.Params["fakeMessage"].([]byte)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'fakeMessage' parameter")
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

	attack := &FakeAttack{
		// TODO:
		// should refactoring
		baseAttack:  nil,
		messageCode: messageCode,
		fakeMessage: fakeMessage,
	}

	return attack, nil
}

// Type returns the attack type
func (f *FakeAttack) Type() types.AttackType { return "" }

// Execute executes the fake attack
func (f *FakeAttack) Execute(ctx *types.AttackContext) error {

	// Check if this is the target message type
	//if ctx.Message == nil || ctx.Message.Code() != f.messageCode {
	//	return nil
	//}

	log.Info("Executing fake attack",
		//"name", f.Name(),
		//"sequence", ctx.Sequence,
		//"round", ctx.Round,
		"messageType", f.messageCode,
		"fakeMessageSize", len(f.fakeMessage))
	//"targets", len(f.GetTargets()))

	// Log fake message details for debugging
	log.Debug("Fake message details",
		"messageHex", common.Bytes2Hex(f.fakeMessage),
		"messageLength", len(f.fakeMessage))

	// Validate the fake message
	if err := f.validateFakeMessage(); err != nil {
		log.Error("Invalid fake message", "error", err)
		//f.RecordExecution(false, err)
		return err
	}

	// The actual fake message injection will be handled by the interceptor
	// This just marks the execution and provides the fake message data
	//f.RecordExecution(true, nil)

	return nil
}

// GetMessageCode returns the target message code
func (f *FakeAttack) GetMessageCode() types.MessageCode {
	return f.messageCode
}

// GetFakeMessage returns the fake message content
func (f *FakeAttack) GetFakeMessage() []byte {
	return f.fakeMessage
}

// SetFakeMessage sets the fake message content
func (f *FakeAttack) SetFakeMessage(message []byte) {
	f.fakeMessage = message
}

// validateFakeMessage performs basic validation on the fake message
func (f *FakeAttack) validateFakeMessage() error {
	if len(f.fakeMessage) == 0 {
		return fmt.Errorf("fake message cannot be empty")
	}

	const maxMessageSize = 32 * 1024 * 1024 // 32MB limit
	if len(f.fakeMessage) > maxMessageSize {
		return fmt.Errorf("fake message too large: %d bytes (max %d)",
			len(f.fakeMessage), maxMessageSize)
	}

	switch {
	case f.messageCode.HasCode(types.MessageCodePrePrepare):
		return f.validatePrePrepareMessage()
	case f.messageCode.HasCode(types.MessageCodePrepare):
		return f.validatePrepareMessage()
	case f.messageCode.HasCode(types.MessageCodeCommit):
		return f.validateCommitMessage()
	case f.messageCode.HasCode(types.MessageCodeRoundChange):
		return f.validateRoundChangeMessage()
	case f.messageCode.HasCode(types.MessageCodePropagation):
		return f.validatePropagationMessage()
	default:
		// Unknown message type - allow but warn
		log.Warn("Unknown message type for fake attack",
			"messageType", f.messageCode)
	}

	return nil
}

// validatePrePrepareMessage validates fake PrePrepare message
func (f *FakeAttack) validatePrePrepareMessage() error {
	// Basic structure validation for PrePrepare
	const minPrePrepareSize = 100 // Minimum realistic size
	if len(f.fakeMessage) < minPrePrepareSize {
		log.Warn("Fake PrePrepare message may be too small",
			"size", len(f.fakeMessage),
			"minimum", minPrePrepareSize)
	}
	return nil
}

// validatePrepareMessage validates fake Prepare message
func (f *FakeAttack) validatePrepareMessage() error {
	// Basic structure validation for Prepare
	const minPrepareSize = 64 // Minimum realistic size
	if len(f.fakeMessage) < minPrepareSize {
		log.Warn("Fake Prepare message may be too small",
			"size", len(f.fakeMessage),
			"minimum", minPrepareSize)
	}
	return nil
}

// validateCommitMessage validates fake Commit message
func (f *FakeAttack) validateCommitMessage() error {
	// Basic structure validation for Commit
	const minCommitSize = 64 // Minimum realistic size
	if len(f.fakeMessage) < minCommitSize {
		log.Warn("Fake Commit message may be too small",
			"size", len(f.fakeMessage),
			"minimum", minCommitSize)
	}
	return nil
}

// validateRoundChangeMessage validates fake RoundChange message
func (f *FakeAttack) validateRoundChangeMessage() error {
	// Basic structure validation for RoundChange
	const minRoundChangeSize = 32 // Minimum realistic size
	if len(f.fakeMessage) < minRoundChangeSize {
		log.Warn("Fake RoundChange message may be too small",
			"size", len(f.fakeMessage),
			"minimum", minRoundChangeSize)
	}
	return nil
}

// validatePropagationMessage validates fake Propagation message
func (f *FakeAttack) validatePropagationMessage() error {
	// Basic structure validation for Propagation
	const minPropagationSize = 100 // Minimum realistic size
	if len(f.fakeMessage) < minPropagationSize {
		log.Warn("Fake Propagation message may be too small",
			"size", len(f.fakeMessage),
			"minimum", minPropagationSize)
	}
	return nil
}

// GenerateRandomFakeMessage generates a random fake message for testing
func (f *FakeAttack) GenerateRandomFakeMessage(size int) {
	if size <= 0 {
		size = 128 // Default size
	}

	fakeMessage := make([]byte, size)
	for i := range fakeMessage {
		fakeMessage[i] = byte(i % 256)
	}

	f.fakeMessage = fakeMessage
	log.Debug("Generated random fake message", "size", size)
}

func (f *FakeAttack) ID() string                                   { return "" }
func (f *FakeAttack) SetID(id uint64)                              {}
func (f *FakeAttack) CheckCondition(ctx *types.AttackContext) bool { return false }
func (f *FakeAttack) Configure(params interface{}) error           { return nil }
func (f *FakeAttack) Validate() error                              { return nil }
func (f *FakeAttack) Reset() error                                 { return nil }
func (f *FakeAttack) ToJSON() ([]byte, error)                      { return nil, nil }
func (f *FakeAttack) FromJSON(data []byte) error                   { return nil }

func (s *FakeAttack) Config() types.AttackConfig                                { return types.AttackConfig{} }
func (s *FakeAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool { return false }
func (s *FakeAttack) RequiresData() []types.DataRequirement                     { return nil }

// FakeAttackFactory creates fake attacks
func FakeAttackFactory(config *types.AttackConfig) (types.Attack, error) {
	return NewFakeAttack(config)
}
