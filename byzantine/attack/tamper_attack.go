package attack

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// TamperField represents a field to tamper with its new value
type TamperField struct {
	Target string      // Field path to tamper (e.g., "Proposal.Header.Coinbase")
	Value  interface{} // New value to set
}

// TamperAttack implements an attack where message fields are tampered
type TamperAttack struct {
	*baseAttack
	messageCode  types.MessageCode
	tamperFields []string
}

// NewTamperAttack creates a new tamper attack
func NewTamperAttack(config *types.AttackConfig) (types.Attack, error) {
	//sequence := config.Sequence
	//round := config.Round

	// Parse message code
	codeStr, ok := config.Params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'code' parameter")
	}
	messageCode := types.ParseMessageCode(codeStr)

	// Parse tamper fields
	tamperFields, ok := config.Params["tamperFields"].([]string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'tamperFields' parameter")
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

	attack := &TamperAttack{
		// TODO:
		// should refactoring
		baseAttack:   nil,
		messageCode:  messageCode,
		tamperFields: tamperFields,
	}

	return attack, nil
}

// Type returns the attack type
func (t *TamperAttack) Type() types.AttackType {
	return ""
}

// Execute executes the tamper attack
func (t *TamperAttack) Execute(ctx *types.AttackContext) error {
	// Check if this is the target message type
	//if ctx.Message == nil || ctx.Message.Code() != t.messageCode {
	//	return nil
	//}

	log.Info("Executing tamper attack",
		//"name", t.Name(),
		//"sequence", ctx.Sequence,
		//"round", ctx.Round,
		"messageType", t.messageCode,
		"tamperFields", t.tamperFields)
	//"targets", len(t.GetTargets()))

	// The actual tampering will be handled by the interceptor
	//t.RecordExecution(true, nil)

	return nil
}

// GetMessageCode returns the target message code
func (t *TamperAttack) GetMessageCode() types.MessageCode {
	return t.messageCode
}

// GetTamperFields returns the fields to tamper
func (t *TamperAttack) GetTamperFields() []string {
	return t.tamperFields
}

// ShouldSendValidMessage returns whether to send a valid message
func (t *TamperAttack) ShouldSendValidMessage() bool {
	return false
}

// GetDelay returns the delay for valid message
func (t *TamperAttack) GetDelay() time.Duration {
	return 0
}

// ApplyTamper applies tampering to a message payload
func (t *TamperAttack) ApplyTamper(originalPayload []byte) ([]byte, error) {
	// This is a placeholder for the actual tampering logic
	// In a real implementation, this would parse the message,
	// apply the specified field changes, and re-encode

	log.Debug("Applying tamper to message",
		"originalSize", len(originalPayload),
		"tamperFields", len(t.tamperFields))

	// For now, return the original payload
	// Real implementation would modify specific fields
	return originalPayload, nil
}

// ValidateTamperTarget validates if a tamper target is valid
func ValidateTamperTarget(target string) error {
	validTargets := map[string]bool{
		// PrePrepare message targets
		string(types.TamperProposalHeaderCoinbase):    true,
		string(types.TamperProposalHeaderNumber):      true,
		string(types.TamperProposalHeaderTime):        true,
		string(types.TamperProposalHeaderParentHash):  true,
		string(types.TamperProposalHeaderStateRoot):   true,
		string(types.TamperProposalHeaderTxHash):      true,
		string(types.TamperProposalHeaderReceiptHash): true,
		string(types.TamperProposalHeaderBloom):       true,
		string(types.TamperProposalHeaderDifficulty):  true,
		string(types.TamperProposalHeaderGasLimit):    true,
		string(types.TamperProposalHeaderGasUsed):     true,
		string(types.TamperProposalHeaderExtra):       true,
		string(types.TamperProposalHeaderMixDigest):   true,
		string(types.TamperProposalHeaderNonce):       true,

		// Transaction targets
		string(types.TamperTransactionSign):    true,
		string(types.TamperTransactionBalance): true,
		string(types.TamperTransactionNonce):   true,
		string(types.TamperTransactionData):    true,

		// Common message targets
		string(types.TamperMessageRound):     true,
		string(types.TamperMessageSequence):  true,
		string(types.TamperMessageSealType):  true,
		string(types.TamperMessageHeader):    true,
		string(types.TamperMessageSignature): true,

		// RoundChange specific
		string(types.TamperRoundChangePreparedRound):  true,
		string(types.TamperRoundChangePreparedDigest): true,

		// Seal targets
		string(types.TamperSealPrepare):  true,
		string(types.TamperSealCommit):   true,
		string(types.TamperSealPrevious): true,
	}

	if !validTargets[target] {
		return fmt.Errorf("invalid tamper target: %s", target)
	}

	return nil
}

func (t *TamperAttack) ID() string                                   { return "" }
func (t *TamperAttack) SetID(id uint64)                              {}
func (t *TamperAttack) CheckCondition(ctx *types.AttackContext) bool { return false }
func (t *TamperAttack) Configure(params interface{}) error           { return nil }
func (t *TamperAttack) Validate() error                              { return nil }
func (t *TamperAttack) Reset() error                                 { return nil }
func (t *TamperAttack) ToJSON() ([]byte, error)                      { return nil, nil }
func (t *TamperAttack) FromJSON(data []byte) error                   { return nil }

func (s *TamperAttack) Config() types.AttackConfig                                { return types.AttackConfig{} }
func (s *TamperAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool { return false }
func (s *TamperAttack) RequiresData() []types.DataRequirement                     { return nil }

// TamperAttackFactory creates tamper attacks
func TamperAttackFactory(config *types.AttackConfig) (types.Attack, error) {
	return NewTamperAttack(config)
}
