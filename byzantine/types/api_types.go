package types

import "github.com/ethereum/go-ethereum/common"

// AttackAPIConfig represents unified attack configuration for API
// This mirrors AttackConfig but with explicit parameter types
type AttackAPIConfig struct {
	Name              string                 `json:"name"`
	Type              string                 `json:"type"`
	Enabled           bool                   `json:"enabled"`
	SequenceStart     uint64                 `json:"seq_s"`
	SequenceEnd       uint64                 `json:"seq_e"`
	Round             uint64                 `json:"round"`
	MaxExecutionCount uint64                 `json:"max_execution_count,omitempty"`
	Parameters        map[string]interface{} `json:"parameters"`
}

// RegisterAttacksParams represents parameters for registering multiple attacks
type RegisterAttacksParams struct {
	Attacks []AttackAPIConfig `json:"attacks"`
}

// RegisterAttackResponse represents the response for attack registration
type RegisterAttackResponse struct {
	UID     string `json:"uid"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BatchRegisterResponse represents the response for batch attack registration
type BatchRegisterResponse struct {
	Results []RegisterAttackResponse `json:"results"`
	Total   int                      `json:"total"`
	Success int                      `json:"success"`
	Failed  int                      `json:"failed"`
}

// AttackStatusResponse represents attack status information
type AttackStatusResponse struct {
	UID            string                 `json:"uid"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Status         string                 `json:"status"`
	Enabled        bool                   `json:"enabled"`
	SequenceRange  string                 `json:"sequence_range"`
	Round          uint64                 `json:"round"`
	ExecutionCount uint64                 `json:"execution_count"`
	Parameters     map[string]interface{} `json:"parameters"`
	CreatedAt      int64                  `json:"created_at"`
	ExecutedAt     *int64                 `json:"executed_at,omitempty"`
}

// SilentMessageParams represents parameters for silent message attacks
type SilentMessageParams struct {
	Enabled       bool             `json:"enabled"`
	Sequence      uint64           `json:"sequence,omitempty"`
	SequenceStart uint64           `json:"seq_s,omitempty"`
	SequenceEnd   uint64           `json:"seq_e,omitempty"`
	Round         uint64           `json:"round"`
	Code          uint64           `json:"code"`
	Direction     uint64           `json:"direction"`
	Targets       []common.Address `json:"targets,omitempty"`
}

// TamperedMessageParams represents parameters for tampered message attacks
type TamperedMessageParams struct {
	Enabled          bool             `json:"enabled"`
	Sequence         uint64           `json:"sequence,omitempty"`
	SequenceStart    uint64           `json:"seq_s,omitempty"`
	SequenceEnd      uint64           `json:"seq_e,omitempty"`
	Round            uint64           `json:"round"`
	Code             uint64           `json:"code"`
	Fields           []Field          `json:"fields"`
	WithValidMessage bool             `json:"withValidMessage"`
	Delay            uint64           `json:"delay"`
	Targets          []common.Address `json:"targets,omitempty"`
}

// FakeMessageParams represents parameters for fake message attacks
type FakeMessageParams struct {
	Enabled       bool             `json:"enabled"`
	Sequence      uint64           `json:"sequence,omitempty"`
	SequenceStart uint64           `json:"seq_s,omitempty"`
	SequenceEnd   uint64           `json:"seq_e,omitempty"`
	Round         uint64           `json:"round"`
	Code          uint64           `json:"code"`
	FakeMessage   []byte           `json:"fakeMessage,omitempty"`
	Targets       []common.Address `json:"targets,omitempty"`
}

// OmitMessageParams represents parameters for omit message attacks
type OmitMessageParams struct {
	Enabled       bool             `json:"enabled"`
	Sequence      uint64           `json:"sequence,omitempty"`
	SequenceStart uint64           `json:"seq_s,omitempty"`
	SequenceEnd   uint64           `json:"seq_e,omitempty"`
	Round         uint64           `json:"round"`
	Code          uint64           `json:"code"`
	Cmd           uint64           `json:"cmd"`
	Cnt           uint64           `json:"cnt"`
	Targets       []common.Address `json:"targets,omitempty"`
}

// RoleSpoofParams represents parameters for role spoofing attacks
type RoleSpoofParams struct {
	Enabled       bool             `json:"enabled"`
	Sequence      uint64           `json:"sequence,omitempty"`
	SequenceStart uint64           `json:"seq_s,omitempty"`
	SequenceEnd   uint64           `json:"seq_e,omitempty"`
	Round         uint64           `json:"round"`
	Code          uint64           `json:"code"`
	FakeMessage   []byte           `json:"fakeMessage,omitempty"`
	Targets       []common.Address `json:"targets,omitempty"`
}

// ReplayMessageParams represents parameters for replay attacks
type ReplayMessageParams struct {
	Enabled         bool             `json:"enabled"`
	Sequence        uint64           `json:"sequence,omitempty"`
	SequenceStart   uint64           `json:"seq_s,omitempty"`
	SequenceEnd     uint64           `json:"seq_e,omitempty"`
	Round           uint64           `json:"round"`
	Code            uint64           `json:"code"`
	UseOriginalView bool             `json:"useOriginalView"`
	Targets         []common.Address `json:"targets,omitempty"`
}

// StoreMessageParams represents parameters for replay attacks
type StoreMessageParams struct {
	Enabled       bool   `json:"enabled"`
	Sequence      uint64 `json:"sequence,omitempty"`
	SequenceStart uint64 `json:"seq_s,omitempty"`
	SequenceEnd   uint64 `json:"seq_e,omitempty"`
	Round         uint64 `json:"round"`
	Code          uint64 `json:"code"`
}

// UpgradeGovContractMessageParams represents parameters for gov contract upgrade
type UpgradeGovContractMessageParams struct{}
