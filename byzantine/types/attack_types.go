package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// AttackType represents different types of Byzantine attacks
type AttackType string

const (
	AttackTypeSilentMessage   AttackType = AttackSilent
	AttackTypeTamperedMessage AttackType = AttackTamper
	AttackTypeFakeMessage     AttackType = AttackFake
	AttackTypeOmitMessage     AttackType = AttackOmit
	AttackTypeRoleSpoofed     AttackType = AttackRoleSpoof
	AttackTypeReplay          AttackType = AttackReplay
	AttackTypeStoreMessage    AttackType = AttackStore
)

// Slice of all AttackType constants for iteration
var AllAttackTypes = []AttackType{
	AttackTypeSilentMessage,
	AttackTypeTamperedMessage,
	AttackTypeFakeMessage,
	AttackTypeOmitMessage,
	AttackTypeRoleSpoofed,
	AttackTypeReplay,
	AttackTypeStoreMessage,
}

// AttackStatus represents the status of an attack
type AttackStatus string

const (
	AttackStatusPending   AttackStatus = "pending"
	AttackStatusActive    AttackStatus = "active"
	AttackStatusExecuted  AttackStatus = "executed"
	AttackStatusCompleted AttackStatus = "completed"
	AttackStatusFailed    AttackStatus = "failed"
	AttackStatusCancelled AttackStatus = "cancelled"
)

// StringToAttackType converts string to AttackType
func StringToAttackType(s string) AttackType {
	switch s {
	case "silent":
		return AttackTypeSilentMessage
	case "tamper":
		return AttackTypeTamperedMessage
	case "fake":
		return AttackTypeFakeMessage
	case "omit":
		return AttackTypeOmitMessage
	case "roleSpoof":
		return AttackTypeRoleSpoofed
	case "replay":
		return AttackTypeReplay
	case "store":
		return AttackTypeStoreMessage
	default:
		return AttackType(s) // fallback
	}
}

func AttackTypeToString(attackType AttackType) string {
	switch attackType {
	case AttackTypeSilentMessage:
		return AttackSilent
	case AttackTypeTamperedMessage:
		return AttackTamper
	case AttackTypeFakeMessage:
		return AttackFake
	case AttackTypeOmitMessage:
		return AttackOmit
	case AttackTypeRoleSpoofed:
		return AttackRoleSpoof
	case AttackTypeReplay:
		return AttackReplay
	case AttackTypeStoreMessage:
		return AttackStore
	default:
		return "unknown"
	}
}

// AttackConfig represents the configuration for an attack
type AttackConfig struct {
	UID              string                 `json:"uid"`
	Name             string                 `json:"name"`
	Type             AttackType             `json:"type"`
	Enabled          bool                   `json:"enabled"`
	SequenceStart    uint64                 `json:"seq_s"`
	SequenceEnd      uint64                 `json:"seq_e"`
	Round            uint64                 `json:"round"`
	ExecutionCount   uint64                 `json:"-"`
	Status           AttackStatus           `json:"status,omitempty"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	RawParameters    json.RawMessage        `json:"-"`
	ParsedParameters interface{}            `json:"-"`
	CreatedAt        time.Time              `json:"created_at"`
	ExecutedAt       *time.Time             `json:"executed_at,omitempty"`
	LastExecutedSeq  uint64                 `json:"-"`
}

func (ac *AttackConfig) UnmarshalJSON(data []byte) error {
	type Alias AttackConfig
	aux := &struct {
		*Alias
		Type       string          `json:"type"`
		Parameters json.RawMessage `json:"parameters,omitempty"`
		Sequence   *uint64         `json:"sequence,omitempty"`
	}{
		Alias: (*Alias)(ac),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ac.Type = StringToAttackType(aux.Type)

	if aux.Sequence != nil {
		ac.SequenceStart = *aux.Sequence
		ac.SequenceEnd = 0
	}

	if len(aux.Parameters) > 0 {
		// Store raw parameters for later parsing
		ac.RawParameters = aux.Parameters
		// Also unmarshal to map for backward compatibility
		json.Unmarshal(aux.Parameters, &ac.Parameters)
	}

	uidGen := NewUIDGenerator()
	ac.UID = uidGen.GenerateWithRange(ac.Type, ac.SequenceStart, ac.SequenceEnd, ac.Round)

	if ac.Status == "" {
		ac.Status = AttackStatusPending
	}
	ac.CreatedAt = time.Now()

	return nil
}

// IsInSequenceRange checks if given sequence is in attack's range
func (ac *AttackConfig) IsInSequenceRange(sequence uint64) bool {
	if ac.SequenceEnd == 0 || ac.SequenceEnd == ac.SequenceStart {
		// Single sequence case
		return sequence == ac.SequenceStart
	}
	// Range case
	return sequence >= ac.SequenceStart && sequence <= ac.SequenceEnd
}

// CanExecute checks if attack can be executed
func (ac *AttackConfig) CanExecute() bool {
	// Check if enabled
	if !ac.Enabled {
		return false
	}

	// Check status
	switch ac.Status {
	case AttackStatusCompleted, AttackStatusCancelled, AttackStatusFailed:
		return false
	default:
		// Continue with execution limit check
	}

	// NOTE:
	// Currently not checking ExecutionCount.
	// Activated after byzantine attack development is complete.
	// Check already execution
	//if ac.ExecutionCount > uint64(0) {
	//	return false
	//}

	return true
}

// IncrementExecutionCount increments the execution count
func (ac *AttackConfig) IncrementExecutionCount(sequence uint64) {
	ac.ExecutionCount++
	ac.LastExecutedSeq = sequence

	if ac.ExecutionCount >= uint64(1) {
		ac.Status = AttackStatusCompleted
	}
}

// GetSequenceRange returns the sequence range as string for display
func (ac *AttackConfig) GetSequenceRange() string {
	if ac.SequenceEnd == 0 {
		return fmt.Sprintf("%d", ac.SequenceStart)
	}
	return fmt.Sprintf("%d-%d", ac.SequenceStart, ac.SequenceEnd)
}

// MarshalJSON implements custom JSON marshaling to ensure deep nes ted structures are properly serialized
func (ac *AttackConfig) MarshalJSON() ([]byte, error) {
	type Alias AttackConfig

	var serializedParams interface{}
	if ac.Parameters != nil {
		serializedParams = ensureSerializable(ac.Parameters)
	}

	return json.Marshal(&struct {
		*Alias
		Parameters interface{} `json:"parameters,omitempty"`
	}{
		Alias:      (*Alias)(ac),
		Parameters: serializedParams,
	})
}

// GetSilentParams returns parsed parameters for silent attack
func (ac *AttackConfig) GetSilentParams() (*SilentAttackParams, error) {
	if ac.Type != AttackTypeSilentMessage {
		return nil, fmt.Errorf("invalid attack type: expected %s, got %s", AttackTypeSilentMessage, ac.Type)
	}

	params := ac.ParsedParameters.(*SilentAttackParams)

	return params, nil
}

// GetTamperParams returns parsed parameters for tamper attack
func (ac *AttackConfig) GetTamperParams() (*TamperAttackParams, error) {
	if ac.Type != AttackTypeTamperedMessage {
		return nil, fmt.Errorf("invalid attack type: expected %s, got %s", AttackTypeTamperedMessage, ac.Type)
	}

	params := ac.ParsedParameters.(*TamperAttackParams)

	return params, nil
}

// GetFakeParams returns parsed parameters for fake attack
func (ac *AttackConfig) GetFakeParams() (*FakeAttackParams, error) {
	if ac.Type != AttackTypeFakeMessage {
		return nil, fmt.Errorf("invalid attack type: expected %s, got %s", AttackTypeFakeMessage, ac.Type)
	}
	params, ok := ac.ParsedParameters.(*FakeAttackParams)
	if !ok {
		return nil, errors.New("parameters not properly parsed")
	}
	return params, nil
}

// GetOmitParams returns parsed parameters for omit attack
func (ac *AttackConfig) GetOmitParams() (*OmitAttackParams, error) {
	if ac.Type != AttackTypeOmitMessage {
		return nil, fmt.Errorf("invalid attack type: expected %s, got %s", AttackTypeOmitMessage, ac.Type)
	}
	params, ok := ac.ParsedParameters.(*OmitAttackParams)
	if !ok {
		return nil, errors.New("parameters not properly parsed")
	}
	return params, nil
}

// GetRoleSpoofParams returns parsed parameters for role spoof attack
func (ac *AttackConfig) GetRoleSpoofParams() (*RoleSpoofAttackParams, error) {
	if ac.Type != AttackTypeRoleSpoofed {
		return nil, fmt.Errorf("invalid attack type: expected %s, got %s", AttackTypeRoleSpoofed, ac.Type)
	}
	params, ok := ac.ParsedParameters.(*RoleSpoofAttackParams)
	if !ok {
		return nil, errors.New("parameters not properly parsed")
	}
	return params, nil
}

// GetReplayParams returns parsed parameters for replay attack
func (ac *AttackConfig) GetReplayParams() (*ReplayAttackParams, error) {
	if ac.Type != AttackTypeReplay {
		return nil, fmt.Errorf("invalid attack type: expected %s, got %s", AttackTypeReplay, ac.Type)
	}
	params, ok := ac.ParsedParameters.(*ReplayAttackParams)
	if !ok {
		return nil, errors.New("parameters not properly parsed")
	}
	return params, nil
}

// GetReplayParams returns parsed parameters for replay attack
func (ac *AttackConfig) GetStoreParams() (*StoreAttackParams, error) {
	if ac.Type != AttackTypeStoreMessage {
		return nil, fmt.Errorf("invalid attack type: expected %s, got %s", AttackTypeStoreMessage, ac.Type)
	}
	params, ok := ac.ParsedParameters.(*StoreAttackParams)
	if !ok {
		return nil, errors.New("parameters not properly parsed")
	}
	return params, nil
}

// Validate validates the attack configuration including type-specific parameters
func (ac *AttackConfig) Validate() error {
	// Basic validation
	if ac.Name == "" {
		return errors.New("attack name is required")
	}
	if ac.Type == "" {
		return errors.New("attack type is required")
	}

	// Type-specific parameter validation
	if ac.ParsedParameters == nil {
		return errors.New("parameters not parsed")
	}

	switch ac.Type {
	case AttackTypeSilentMessage:
		params, err := ac.GetSilentParams()
		if err != nil {
			return err
		}
		return validateSilentParams(params)

	case AttackTypeTamperedMessage:
		params, err := ac.GetTamperParams()
		if err != nil {
			return err
		}
		return validateTamperParams(params)

	case AttackTypeFakeMessage:
		params, err := ac.GetFakeParams()
		if err != nil {
			return err
		}
		return validateFakeParams(params)

	case AttackTypeOmitMessage:
		params, err := ac.GetOmitParams()
		if err != nil {
			return err
		}
		return validateOmitParams(params)

	case AttackTypeRoleSpoofed:
		params, err := ac.GetRoleSpoofParams()
		if err != nil {
			return err
		}
		return validateRoleSpoofParams(params)

	case AttackTypeReplay:
		params, err := ac.GetReplayParams()
		if err != nil {
			return err
		}
		return validateReplayParams(params)

	case AttackTypeStoreMessage:
		params, err := ac.GetStoreParams()
		if err != nil {
			return err
		}
		return validateStoreParams(params)

	default:
		return fmt.Errorf("unknown attack type: %s", ac.Type)
	}
}

// GetTargetAddresses converts string targets to common.Address
func (ac *AttackConfig) GetTargetAddresses() []common.Address {
	addresses := make([]common.Address, 0, len(ac.Parameters["targets"].([]common.Address)))
	for _, target := range ac.Parameters["targets"].([]common.Address) {
		addresses = append(addresses, target)
	}
	return addresses
}

// Parameter validation functions
func validateSilentParams(params *SilentAttackParams) error {
	if params.Direction > 3 {
		return fmt.Errorf("invalid direction: %d (must be 0-3)", params.Direction)
	}
	return validateTargets(params.Targets)
}

func validateTamperParams(params *TamperAttackParams) error {
	for i, field := range params.Fields {
		if field.Target == "" {
			return fmt.Errorf("tamperField[%d] target is empty", i)
		}
		if field.Value == nil {
			return fmt.Errorf("tamperField[%d] value is nil", i)
		}
	}
	return validateTargets(params.Targets)
}

func validateFakeParams(params *FakeAttackParams) error {
	// FakeMessage can be empty, as it might be generated later
	return validateTargets(params.Targets)
}

func validateOmitParams(params *OmitAttackParams) error {
	// Cmd and Option can be 0, which might be valid
	return validateTargets(params.Targets)
}

func validateRoleSpoofParams(params *RoleSpoofAttackParams) error {
	// FakeMessage can be empty, as it might be generated later
	return validateTargets(params.Targets)
}

func validateReplayParams(params *ReplayAttackParams) error {
	return validateTargets(params.Targets)
}

func validateStoreParams(params *StoreAttackParams) error {
	return nil
}

func validateTargets(targets []common.Address) error {
	for i, target := range targets {
		if target == (common.Address{}) {
			return fmt.Errorf("invalid target address at index %d: zero address", i)
		}
	}
	return nil
}

func ensureSerializable(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = ensureSerializable(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = ensureSerializable(v)
		}
		return result
	default:
		return val
	}
}

// ParseAddresses parses addresses from various formats
func ParseAddresses(val interface{}) []common.Address {
	var addresses []common.Address

	switch v := val.(type) {
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok && common.IsHexAddress(str) {
				addresses = append(addresses, common.HexToAddress(str))
			}
		}
	case []string:
		for _, str := range v {
			if common.IsHexAddress(str) {
				addresses = append(addresses, common.HexToAddress(str))
			}
		}
	case []common.Address:
		addresses = v
	}

	return addresses
}

// AttackResult represents the result of an attack execution
type AttackResult struct {
	UID          string        `json:"uid"`
	Success      bool          `json:"success"`
	Error        error         `json:"error,omitempty"`
	Details      interface{}   `json:"details,omitempty"`
	ExecutedAt   time.Time     `json:"executed_at"`
	Duration     time.Duration `json:"duration"`
	BlockMessage bool          `json:"block_message,omitempty"`
	BlockReason  string        `json:"block_reason,omitempty"`
}

// AttackDecision represents the consolidated decision from attack evaluation
type AttackDecision struct {
	ShouldAttack bool
	AttackUID    string
	AttackType   AttackType
	Reason       string
	Result       *AttackResult
}

// AttackContext provides context for attacks execution
type AttackContext struct {
	CurrentSequence uint64
	CurrentRound    uint64
	MessageCode     uint64
}

// ExecutableAttack bundles the original AttackConfig with its
// concrete parameter struct (TamperAttackParams, FakeAttackParams, …).
type ExecutableAttack struct {
	Enabled            bool
	UID                string
	NAME               string
	Status             AttackStatus
	SilentParams       *SilentAttackParams
	TamperParams       *TamperAttackParams
	FakeParams         *FakeAttackParams
	OmitParams         *OmitAttackParams
	RoleSpoofParams    *RoleSpoofAttackParams
	ReplayParams       *ReplayAttackParams
	StoreMessageParams *StoreAttackParams
}
