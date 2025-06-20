package types

// AttackStatus represents the status of an attack
type AttackStatus string

const (
	AttackStatusPending  AttackStatus = "pending"
	AttackStatusActive   AttackStatus = "active"
	AttackStatusExecuted AttackStatus = "executed"
	AttackStatusStopped  AttackStatus = "stopped"
	AttackStatusFailed   AttackStatus = "failed"
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
