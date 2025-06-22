package api

import (
	"github.com/ethereum/go-ethereum/byzantine/attack"
	"github.com/ethereum/go-ethereum/common"
)

// Import types from attack package
type (
	AttackType      = attack.AttackType
	AttackParams    = attack.AttackParams
	AttackInfo      = attack.AttackInfo
	Attack          = attack.Attack
	AttackBuilder   = attack.AttackBuilder
	AttackContext   = attack.AttackContext
	DataRequirement = attack.DataRequirement
)

// Re-export attack types for convenience
const (
	AttackTypeDoublePrepare   = attack.AttackTypeDoublePrepare
	AttackTypeDoubleCommit    = attack.AttackTypeDoubleCommit
	AttackTypeSilentProposer  = attack.AttackTypeSilentProposer
	AttackTypeSilentValidator = attack.AttackTypeSilentValidator
	AttackTypeTamperedHeader  = attack.AttackTypeTamperedHeader
	AttackTypeFakeTransaction = attack.AttackTypeFakeTransaction
	// ... other attack types
)

// API Parameter Types

// SilentMessageParams represents parameters for silent message attack
type SilentMessageParams struct {
	Sequence  uint64           `json:"sequence"`
	Round     uint64           `json:"round"`
	Code      uint64           `json:"code"`
	Direction uint64           `json:"direction"`
	Targets   []common.Address `json:"targets,omitempty"`
}

// TamperedMessageParams represents parameters for tampered message attack
type TamperedMessageParams struct {
	Sequence         uint64           `json:"sequence"`
	Round            uint64           `json:"round"`
	Code             string           `json:"code"`
	TamperFields     []TamperField    `json:"tamperFields"`
	WithValidMessage bool             `json:"withValidMessage"`
	Delay            uint64           `json:"delay"`
	Targets          []common.Address `json:"targets,omitempty"`
}

// TamperField represents a field to tamper
type TamperField struct {
	Target string      `json:"target"`
	Value  interface{} `json:"value"`
}

// FakeMessageParams represents parameters for fake message attack
type FakeMessageParams struct {
	Sequence    uint64           `json:"sequence"`
	Round       uint64           `json:"round"`
	Code        string           `json:"code"`
	FakeMessage []byte           `json:"fakeMessage,omitempty"`
	Targets     []common.Address `json:"targets,omitempty"`
}

// OmitMessageParams represents parameters for omit message attack
type OmitMessageParams struct {
	Sequence uint64           `json:"sequence"`
	Round    uint64           `json:"round"`
	Code     string           `json:"code"`
	Cmd      uint64           `json:"cmd"`
	Cnt      uint64           `json:"cnt"`
	Targets  []common.Address `json:"targets,omitempty"`
}

// RoleSpoofParams represents parameters for role spoofing attack
type RoleSpoofParams struct {
	Sequence    uint64           `json:"sequence"`
	Round       uint64           `json:"round"`
	Code        string           `json:"code"`
	FakeMessage []byte           `json:"fakeMessage,omitempty"`
	Targets     []common.Address `json:"targets,omitempty"`
}

// ReplayMessageParams represents parameters for replay attack
type ReplayMessageParams struct {
	OriSequence     uint64           `json:"ori_sequence"`
	OriRound        uint64           `json:"ori_round"`
	Sequence        uint64           `json:"sequence"`
	Round           uint64           `json:"round"`
	UseOriginalView bool             `json:"useOriginalView"`
	Code            string           `json:"code"`
	Targets         []common.Address `json:"targets,omitempty"`
}

// UpgradeGovContractMessageParams represents parameters for gov contract upgrade
type UpgradeGovContractMessageParams struct{}
