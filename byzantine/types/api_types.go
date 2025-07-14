package types

import "github.com/ethereum/go-ethereum/common"

// SilentMessageParams represents parameters for silent message attacks
type SilentMessageParams struct {
	Sequence  uint64           `json:"sequence"`
	Round     uint64           `json:"round"`
	Code      uint64           `json:"code"`
	Direction uint64           `json:"direction"`
	Targets   []common.Address `json:"targets,omitempty"`
}

// TamperedMessageParams represents parameters for tampered message attacks
type TamperedMessageParams struct {
	Sequence         uint64           `json:"sequence"`
	Round            uint64           `json:"round"`
	Code             uint64           `json:"code"`
	Fields           []Field          `json:"fields"`
	WithValidMessage bool             `json:"withValidMessage"`
	Delay            uint64           `json:"delay"`
	Targets          []common.Address `json:"targets,omitempty"`
}

// FakeMessageParams represents parameters for fake message attacks
type FakeMessageParams struct {
	Sequence    uint64           `json:"sequence"`
	Round       uint64           `json:"round"`
	Code        uint64           `json:"code"`
	FakeMessage []byte           `json:"fakeMessage,omitempty"`
	Targets     []common.Address `json:"targets,omitempty"`
}

// OmitMessageParams represents parameters for omit message attacks
type OmitMessageParams struct {
	Sequence uint64           `json:"sequence"`
	Round    uint64           `json:"round"`
	Code     uint64           `json:"code"`
	Cmd      uint64           `json:"cmd"`
	Cnt      uint64           `json:"cnt"`
	Targets  []common.Address `json:"targets,omitempty"`
}

// RoleSpoofParams represents parameters for role spoofing attacks
type RoleSpoofParams struct {
	Sequence    uint64           `json:"sequence"`
	Round       uint64           `json:"round"`
	Code        uint64           `json:"code"`
	FakeMessage []byte           `json:"fakeMessage,omitempty"`
	Targets     []common.Address `json:"targets,omitempty"`
}

// ReplayMessageParams represents parameters for replay attacks
type ReplayMessageParams struct {
	Sequence        uint64           `json:"sequence"`
	Round           uint64           `json:"round"`
	Code            uint64           `json:"code"`
	UseOriginalView bool             `json:"useOriginalView"`
	Targets         []common.Address `json:"targets,omitempty"`
}

// StoreMessageParams represents parameters for replay attacks
type StoreMessageParams struct {
	Sequence uint64 `json:"sequence"`
	Round    uint64 `json:"round"`
	Code     uint64 `json:"code"`
}

// UpgradeGovContractMessageParams represents parameters for gov contract upgrade
type UpgradeGovContractMessageParams struct{}
