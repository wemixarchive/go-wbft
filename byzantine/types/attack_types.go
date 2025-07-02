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
)

// Slice of all AttackType constants for iteration
var AllAttackTypes = []AttackType{
	AttackTypeSilentMessage,
	AttackTypeTamperedMessage,
	AttackTypeFakeMessage,
	AttackTypeOmitMessage,
	AttackTypeRoleSpoofed,
	AttackTypeReplay,
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
	default:
		return AttackType(s) // fallback
	}
}

func AttachTypeToString(attackType AttackType) string {
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
	default:
		return "unknown"
	}
}

// AttackConfig represents the configuration for an attack
type AttackConfig struct {
	UID      string       `json:"uid"`
	Name     string       `json:"name"`
	Type     AttackType   `json:"type"`
	Enabled  bool         `json:"enabled"`
	Sequence uint64       `json:"sequence"`
	Round    uint64       `json:"round"`
	Status   AttackStatus `json:"status,omitempty"`
	//Targets          []common.Address       `json:"targets,omitempty"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	ParsedParameters interface{}            `json:"-"`
	CreatedAt        time.Time              `json:"created_at"`
	ExecutedAt       *time.Time             `json:"executed_at,omitempty"`
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

	params := &SilentAttackParams{
		Code:      ac.Parameters["code"].(MessageCode),
		Direction: ac.Parameters["direction"].(uint64),
		Targets:   ac.Parameters["target"].([]common.Address),
	}

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
	if len(params.TamperFields) == 0 && !params.WithValidMessage {
		return fmt.Errorf("tamper attack must have either tamperFields or withValidMessage=true (tamperFields: %d, withValidMessage: %t)", len(params.TamperFields), params.WithValidMessage)
	}
	for i, field := range params.TamperFields {
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
	// Cmd and Cnt can be 0, which might be valid
	return validateTargets(params.Targets)
}

func validateRoleSpoofParams(params *RoleSpoofAttackParams) error {
	// FakeMessage can be empty, as it might be generated later
	return validateTargets(params.Targets)
}

func validateReplayParams(params *ReplayAttackParams) error {
	if params.OriSequence == 0 {
		return errors.New("original sequence must be specified")
	}
	return validateTargets(params.Targets)
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
	Enabled         bool
	Status          AttackStatus
	SilentParams    *SilentAttackParams
	TamperParams    *TamperAttackParams
	FakeParams      *FakeAttackParams
	OmitParams      *OmitAttackParams
	RoleSpoofParams *RoleSpoofAttackParams
	ReplayParams    *ReplayAttackParams
}
