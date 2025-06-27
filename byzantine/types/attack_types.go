package types

import (
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

// AttackInfo represents attack information
type AttackInfo struct {
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

func (ac *AttackConfig) ConvertTypeToString() string {
	switch ac.Type {
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

// AttackInfo for API responses
//type AttackInfo struct {
//	ID       string           `json:"id"`
//	Name     string           `json:"name"`
//	Type     AttackType       `json:"type"`
//	Sequence uint64           `json:"sequence"`
//	Round    uint64           `json:"round"`
//	Status   AttackStatus     `json:"status"`
//	Targets  []common.Address `json:"targets"`
//}

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
