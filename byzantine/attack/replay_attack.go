package attack

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// ReplayAttack implements an attack where old messages are replayed
type ReplayAttack struct {
	*baseAttack

	// Attack configuration
	originalSequence uint64
	originalRound    uint64
	useOriginalView  bool
	messageCode      types.MessageCode
}

// NewReplayAttack creates a new replay attack
func NewReplayAttack(config *types.AttackConfig) (types.Attack, error) {
	// Extract parameters
	//sequence := config.Sequence
	//round := config.Round

	// Parse original sequence
	origSequence, ok := config.Params["ori_sequence"].(uint64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'ori_sequence' parameter")
	}

	// Parse original round
	origRound, ok := config.Params["ori_round"].(uint64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'ori_round' parameter")
	}

	// Parse use original view
	useOriginalView := false
	if use, ok := config.Params["useOriginalView"].(bool); ok {
		useOriginalView = use
	}

	// Parse message code
	messageCodeStr, ok := config.Params["code"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'code' parameter")
	}
	messageCode := types.ParseMessageCode(messageCodeStr)
	if messageCode == 0 {
		return nil, fmt.Errorf("invalid message code: %s", messageCodeStr)
	}

	// Parse targets
	//var targets []common.Address
	//if targetList, ok := config.Params["targets"].([]common.Address); ok {
	//	targets = targetList
	//}

	attack := &ReplayAttack{
		// TODO:
		// should refactoring
		baseAttack:       nil,
		originalSequence: origSequence,
		originalRound:    origRound,
		useOriginalView:  useOriginalView,
		messageCode:      messageCode,
	}

	return attack, nil
}

// Type returns the attack type
func (r *ReplayAttack) Type() types.AttackType { return "" }

// Execute executes the replay attack
func (r *ReplayAttack) Execute(ctx *types.AttackContext) error {
	// Check if this is the target message type
	//if ctx.Message == nil || ctx.Message.Code() != r.messageCode {
	//	return nil
	//}

	log.Info("Executing replay attack",
		//"name", r.Name(),
		//"sequence", ctx.Sequence,
		//"round", ctx.Round,
		"messageCode", r.messageCode.String(),
		"originalSequence", r.originalSequence,
		"originalRound", r.originalRound,
		"useOriginalView", r.useOriginalView)

	// The actual message replay will be handled by the interceptor
	// This just marks the execution
	//r.RecordExecution(true, nil)

	return nil
}

// GetOriginalSequence returns the original message sequence
func (r *ReplayAttack) GetOriginalSequence() uint64 {
	return r.originalSequence
}

// GetOriginalRound returns the original message round
func (r *ReplayAttack) GetOriginalRound() uint64 {
	return r.originalRound
}

// ShouldUseOriginalView returns whether to use original view
func (r *ReplayAttack) ShouldUseOriginalView() bool {
	return r.useOriginalView
}

// GetMessageCode returns the target message code
func (r *ReplayAttack) GetMessageCode() types.MessageCode {
	return r.messageCode
}

func (r *ReplayAttack) ID() string                                   { return "" }
func (r *ReplayAttack) SetID(id uint64)                              {}
func (r *ReplayAttack) CheckCondition(ctx *types.AttackContext) bool { return false }
func (r *ReplayAttack) Configure(params interface{}) error           { return nil }
func (r *ReplayAttack) Validate() error                              { return nil }
func (r *ReplayAttack) Reset() error                                 { return nil }
func (r *ReplayAttack) ToJSON() ([]byte, error)                      { return nil, nil }
func (r *ReplayAttack) FromJSON(data []byte) error                   { return nil }

func (s *ReplayAttack) Config() types.AttackConfig                                { return types.AttackConfig{} }
func (s *ReplayAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool { return false }
func (s *ReplayAttack) RequiresData() []types.DataRequirement                     { return nil }

// ReplayAttackFactory creates replay attacks
func ReplayAttackFactory(config *types.AttackConfig) (types.Attack, error) {
	return NewReplayAttack(config)
}
