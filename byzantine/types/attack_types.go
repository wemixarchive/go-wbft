package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

// AttackType represents different types of Byzantine attacks
type AttackType string

const (
	// Basic attacks types

	AttackTypeSilentMessage   AttackType = AttackSilent
	AttackTypeTamperedMessage AttackType = AttackTamper
	AttackTypeFakeMessage     AttackType = AttackFake
	AttackTypeOmitMessage     AttackType = AttackOmit
	AttackTypeRoleSpoofed     AttackType = AttackRoleSpoof
	AttackTypeReplay          AttackType = AttackReplay
)

// AttackStatus represents the status of an attacks
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
	UID        uint64                 `json:"uid"`
	Name       string                 `json:"name"`
	Type       AttackType             `json:"type"`
	Enabled    bool                   `json:"enabled"`
	Sequence   uint64                 `json:"sequence"`
	Round      uint64                 `json:"round"`
	Code       MessageCode            `json:"code,omitempty"`
	Status     AttackStatus           `json:"status,omitempty"`
	Targets    []common.Address       `json:"targets,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	ExecutedAt *time.Time             `json:"executed_at,omitempty"`
}

// MarshalJSON implements custom JSON marshaling to ensure deep nested structures are properly serialized
func (ac AttackConfig) MarshalJSON() ([]byte, error) {
	type Alias AttackConfig
	
	var serializedParams interface{}
	if ac.Parameters != nil {
		serializedParams = ensureSerializable(ac.Parameters)
	}

	return json.Marshal(&struct {
		*Alias
		Parameters interface{} `json:"parameters,omitempty"`
	}{
		Alias:      (*Alias)(&ac),
		Parameters: serializedParams,
	})
}

func (ac *AttackConfig) UnmarshalJSON(data []byte) error {
	type Alias AttackConfig
	aux := &struct {
		Type string      `json:"type"`
		Code interface{} `json:"code,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(ac),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ac.Type = StringToAttackType(aux.Type)

	if aux.Code != nil {
		ac.Code = ParseMessageCode(aux.Code)
	}

	//if ac.Parameters != nil {
	//	if paramMap, ok := ac.Parameters.(map[string]interface{}); ok {
	//		if codeVal, exists := paramMap["code"]; exists && codeVal != nil {
	//			ac.Code = ParseMessageCode(codeVal)
	//		}
	//
	//		if targetsVal, exists := paramMap["targets"]; exists {
	//			ac.Targets = ParseAddresses(targetsVal)
	//		}
	//	}
	//}

	if !ac.Enabled && aux.Alias.Enabled == false {
		ac.Enabled = true
	}

	return nil
}

// Validate validates a single attacks configuration
func (a *AttackConfig) Validate() error {
	if a.Name == "" {
		return errors.New("attacks name is required")
	}

	if a.Type == "" {
		return errors.New("attacks type is required")
	}

	// Validate attacks type
	validTypes := map[string]bool{
		"doublePrepare": true, "doubleCommit": true,
		"silentProposer": true, "silentValidator": true,
		"tamperedHeader": true, "fakeTransaction": true,
		"messageFlood": true, "replayAttack": true,
	}

	if !validTypes[string(a.Type)] {
		return fmt.Errorf("invalid attacks type: %s", a.Type)
	}

	// Validate targets are valid addresses
	//for i, target := range a.Targets {
	//	if !common.IsHexAddress(target) {
	//		return fmt.Errorf("invalid target address[%d]: %s", i, target)
	//	}
	//}

	return nil
}

// GetTargetAddresses converts string targets to common.Address
func (ac *AttackConfig) GetTargetAddresses() []common.Address {
	addresses := make([]common.Address, 0, len(ac.Targets))
	//for _, target := range ac.Targets {
	//	if common.IsHexAddress(target) {
	//		addresses = append(addresses, common.HexToAddress(target))
	//	}
	//}
	return addresses
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

// AttackResult represents the result of an attacks execution
type AttackResult struct {
	UID          uint64        `json:"uid"`
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
	ShouldBlock bool
	AttackUID   uint64
	AttackType  AttackType
	Reason      string
	Result      *AttackResult
}

// AttackParams contains parameters for configuring an attacks
type AttackParams struct {
	Type     AttackType
	Sequence uint64
	Round    uint64
	Targets  []common.Address
	Options  map[string]interface{}
}

// AttackContext provides context for attacks execution
type AttackContext struct {
	CurrentSequence uint64
	CurrentRound    uint64
	MessageCode     uint64
}

// AttackCategory represents attacks categories
type AttackCategory string

const (
	AttackCategorySafety    AttackCategory = "safety"
	AttackCategoryLiveness  AttackCategory = "liveness"
	AttackCategoryIntegrity AttackCategory = "integrity"
	AttackCategoryRole      AttackCategory = "role"
	AttackCategoryReplay    AttackCategory = "replay"
	AttackCategoryNetwork   AttackCategory = "network"
)

// AttackSeverity represents attacks severity levels
type AttackSeverity string

const (
	AttackSeverityLow      AttackSeverity = "low"
	AttackSeverityMedium   AttackSeverity = "medium"
	AttackSeverityHigh     AttackSeverity = "high"
	AttackSeverityCritical AttackSeverity = "critical"
)

// DataRequirement specifies what data an attacks needs
type DataRequirement struct {
	Type       string // "messages", "blocks", "state"
	Filter     DataFilter
	MaxRecords int
}

// DataFilter for querying historical data
type DataFilter struct {
	FromSequence uint64
	ToSequence   uint64
	MessageTypes []uint64
	Validators   []common.Address
}

// TamperMessageParams for tampered message attacks
type TamperMessageParams struct {
	Sequence         uint64
	Round            uint64
	Code             string
	TamperFields     []TamperField
	WithValidMessage bool
	Delay            uint64
	Targets          []common.Address
}

// TamperField represents a field to be tampered in a message
type TamperField struct {
	Target string      `json:"target"` // e.g., "Proposal.Header.Coinbase"
	Value  interface{} `json:"value"`  // New value for the field
}

// ConditionContext provides context for condition evaluation
type ConditionContext struct {
	Sequence    uint64
	Round       uint64
	MessageType string
	Role        string
	Validators  []common.Address
	Self        common.Address
}
