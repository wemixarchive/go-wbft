package types

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

// AttackConfig represents generic attack configuration
type AttackConfig struct {
	// Identidy
	UID      uint64         `json:"uid"`
	Name     string         `json:"name"`
	Type     AttackType     `json:"type"`
	Category AttackCategory `json:"category,omitempty"`
	Severity AttackSeverity `json:"severity,omitempty"`

	// Execution parameters
	Enabled  bool   `json:"enabled"`
	Sequence uint64 `json:"sequence"`
	Round    uint64 `json:"round"`

	// Attack-specific parameters
	Params map[string]interface{} `json:"params,omitempty"`

	// Advanced options
	RepeatCount      int  `json:"repeatCount,omitempty"`
	RandomDelay      bool `json:"randomDelay,omitempty"`
	CoordinationMode bool `json:"coordinationMode,omitempty"`

	// Target validators for the attack
	Targets []string `json:"targets,omitempty"`

	// Message-specific options
	MessageCode uint64 `json:"messageCode,omitempty"`
	Delay       uint64 `json:"delay,omitempty"`
}

func (ac *AttackConfig) ConvertTypeToString() string {
	switch ac.Type {
	case AttackTypeSilent:
		return "silent"
	case AttackTypeTamper:
		return "tamper"
	case AttackTypeFake:
		return "fake"
	case AttackTypeOmit:
		return "omit"
	case AttackTypeRoleSpoof:
		return "roleSpoof"
	case AttackTypeReplay:
		return "replay"
	default:
		return "unknown"
	}
}

// Validate validates a single attack configuration
func (a *AttackConfig) Validate() error {
	if a.Name == "" {
		return errors.New("attack name is required")
	}

	if a.Type == "" {
		return errors.New("attack type is required")
	}

	// Validate attack type
	validTypes := map[string]bool{
		"doublePrepare": true, "doubleCommit": true,
		"silentProposer": true, "silentValidator": true,
		"tamperedHeader": true, "fakeTransaction": true,
		"messageFlood": true, "replayAttack": true,
	}

	if !validTypes[string(a.Type)] {
		return fmt.Errorf("invalid attack type: %s", a.Type)
	}

	// Validate targets are valid addresses
	for i, target := range a.Targets {
		if !common.IsHexAddress(target) {
			return fmt.Errorf("invalid target address[%d]: %s", i, target)
		}
	}

	return nil
}

// GetTargetAddresses converts string targets to common.Address
func (ac *AttackConfig) GetTargetAddresses() []common.Address {
	addresses := make([]common.Address, 0, len(ac.Targets))
	for _, target := range ac.Targets {
		if common.IsHexAddress(target) {
			addresses = append(addresses, common.HexToAddress(target))
		}
	}
	return addresses
}

// AttackParams contains parameters for configuring an attack
type AttackParams struct {
	Type     AttackType
	Sequence uint64
	Round    uint64
	Targets  []common.Address
	Options  map[string]interface{}
}

// AttackContext provides context for attack execution
type AttackContext struct {
	CurrentSequence uint64
	CurrentRound    uint64
	MessageCode     uint64
}

// AttackParams for API
//type AttackParams struct {
//	Type     string                 `json:"type"`
//	Sequence uint64                 `json:"sequence"`
//	Round    uint64                 `json:"round"`
//	Targets  []common.Address       `json:"targets"`
//	Options  map[string]interface{} `json:"options"`
//}

// AttackInfo for API responses
type AttackInfo struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Type     AttackType       `json:"type"`
	Sequence uint64           `json:"sequence"`
	Round    uint64           `json:"round"`
	Status   AttackStatus     `json:"status"`
	Targets  []common.Address `json:"targets"`
}

// AttackType represents different types of Byzantine attacks
type AttackType string

const (
	// Basic attack types
	AttackTypeSilent    AttackType = "silent"
	AttackTypeTamper    AttackType = "tamper"
	AttackTypeFake      AttackType = "fake"
	AttackTypeOmit      AttackType = "omit"
	AttackTypeRoleSpoof AttackType = "roleSpoof"
	AttackTypeReplay    AttackType = "replay"

	// Advanced attack types
	AttackTypeFlood   AttackType = "flood"
	AttackTypeDiverge AttackType = "diverge"

	// Coordinated attack types
	AttackTypeCoordinatedSilent AttackType = "coordinatedSilent"
	AttackTypePartitionAttack   AttackType = "partitionAttack"

	// Advance attach type
	AttackTypeDoublePrepare   AttackType = "doublePrepare"
	AttackTypeDoubleCommit    AttackType = "doubleCommit"
	AttackTypeSilentProposer  AttackType = "silentProposer"
	AttackTypeSilentValidator AttackType = "silentValidator"
	AttackTypeTamperedHeader  AttackType = "tamperedHeader"
	AttackTypeFakeTransaction AttackType = "fakeTransaction"
)

// AttackStatus represents the status of an attack
type AttackStatus string

const (
	AttackStatusPending  AttackStatus = "pending"
	AttackStatusActive   AttackStatus = "active"
	AttackStatusExecuted AttackStatus = "executed"
	AttackStatusStopped  AttackStatus = "stopped"
	AttackStatusFailed   AttackStatus = "failed"
)

// AttackCategory represents attack categories
type AttackCategory string

const (
	AttackCategorySafety    AttackCategory = "safety"
	AttackCategoryLiveness  AttackCategory = "liveness"
	AttackCategoryIntegrity AttackCategory = "integrity"
	AttackCategoryRole      AttackCategory = "role"
	AttackCategoryReplay    AttackCategory = "replay"
	AttackCategoryNetwork   AttackCategory = "network"
)

// AttackSeverity represents attack severity levels
type AttackSeverity string

const (
	AttackSeverityLow      AttackSeverity = "low"
	AttackSeverityMedium   AttackSeverity = "medium"
	AttackSeverityHigh     AttackSeverity = "high"
	AttackSeverityCritical AttackSeverity = "critical"
)

// DataRequirement specifies what data an attack needs
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

// MessageCode defines QBFT message types
type MessageCode uint64

const (
	MessageCodePrePrepare MessageCode = 1 << iota
	MessageCodePrepare
	MessageCodeCommit
	MessageCodeRoundChange
	MessageCodeRoundChangePrePrepare
	MessageCodePropagation
)
