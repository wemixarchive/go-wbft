package api

import (
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// API request/response models

// RegisterAttackRequest represents a request to register an attack
type RegisterAttackRequest struct {
	Type       types.AttackType       `json:"type"`
	Name       string                 `json:"name"`
	Sequence   uint64                 `json:"sequence"`
	Round      uint64                 `json:"round"`
	Code       types.MessageCode      `json:"code"`
	Targets    []common.Address       `json:"targets,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// RegisterAttackResponse represents a response to register attack request
type RegisterAttackResponse struct {
	UID    uint64 `json:"uid"`
	Status string `json:"status"`
}

// ListAttacksResponse represents a response to list attacks request
type ListAttacksResponse struct {
	Attacks []types.AttackConfig `json:"attacks"`
}

// CancelAttackRequest represents a request to cancel attacks
type CancelAttackRequest struct {
	UIDs []uint64 `json:"uid"`
}

// SetMessagePolicyRequest represents a request for set message policy
type SetMessagePolicyRequest struct {
	Sequence uint64            `json:"sequence"`
	Round    uint64            `json:"round"`
	Code     types.MessageCode `json:"code"`
	Fields   []types.Field     `json:"Fields"`
	Targets  []common.Address  `json:"targets,omitempty"`
}

// TamperedMessageRequest represents a request for tampered message attack
type TamperedMessageRequest struct {
	Sequence     uint64            `json:"sequence"`
	Round        uint64            `json:"round"`
	Code         types.MessageCode `json:"code"`
	TamperFields []types.Field     `json:"tamperFields"`
	Targets      []common.Address  `json:"targets,omitempty"`
}

// FakeMessageRequest represents a request for fake message attack
type FakeMessageRequest struct {
	Sequence    uint64            `json:"sequence"`
	Round       uint64            `json:"round"`
	Code        types.MessageCode `json:"code"`
	FakeMessage []byte            `json:"fakeMessage,omitempty"`
	Targets     []common.Address  `json:"targets,omitempty"`
}

// OmitMessageRequest represents a request for omit message attack
type OmitMessageRequest struct {
	Sequence uint64            `json:"sequence"`
	Round    uint64            `json:"round"`
	Code     types.MessageCode `json:"code"`
	Cmd      uint64            `json:"cmd"`
	Cnt      uint64            `json:"cnt"`
	Targets  []common.Address  `json:"targets,omitempty"`
}

// RoleSpoofedMessageRequest represents a request for role spoofed attack
type RoleSpoofedMessageRequest struct {
	Sequence    uint64            `json:"sequence"`
	Round       uint64            `json:"round"`
	Code        types.MessageCode `json:"code"`
	FakeMessage []byte            `json:"fakeMessage,omitempty"`
	Targets     []common.Address  `json:"targets,omitempty"`
}

// ReplayMessageRequest represents a request for replay message attack
type ReplayMessageRequest struct {
	Sequence        uint64            `json:"sequence"`
	Round           uint64            `json:"round"`
	Code            types.MessageCode `json:"code"`
	UseOriginalView bool              `json:"useOriginalView"`
	Targets         []common.Address  `json:"targets,omitempty"`
}

// StoreMessageRequest represents a request for store message
type StoreMessageRequest struct {
	Sequence uint64            `json:"sequence"`
	Round    uint64            `json:"round"`
	Code     types.MessageCode `json:"code"`
}

// ServiceStatusResponse represents service status response
type ServiceStatusResponse struct {
	Status types.ServiceStatus `json:"status"`
}

// MetricsResponse represents metrics response
type MetricsResponse struct {
	Metrics types.Metrics `json:"metrics"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
