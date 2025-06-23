package types

import "errors"

// Attack management errors
var (
	ErrAttackNotFound              = errors.New("attack not found")
	ErrAttackAlreadyExists         = errors.New("attack already exists")
	ErrInvalidParams               = errors.New("invalid parameters")
	ErrMessageNotFound             = errors.New("message not found")
	ErrAttackTypeAlreadyRegistered = errors.New("attack type already registered")
	ErrUnknownAttackType           = errors.New("unknown attack type")
)

// Common errors
var (
	ErrNotImplemented = errors.New("not implemented")
)

// Hook-related errors
var (
	ErrHookExists    = errors.New("hook already exists")
	ErrHookNotFound  = errors.New("hook not found")
	ErrInvalidHook   = errors.New("invalid hook")
	ErrHookExecution = errors.New("hook execution failed")
)

// Message-related errors
var (
	ErrInvalidMessageCode = errors.New("invalid message code")
	ErrInvalidSequence    = errors.New("invalid sequence number")
	ErrInvalidSender      = errors.New("invalid sender address")
	ErrMissingSignature   = errors.New("missing message signature")
	ErrMessageConversion  = errors.New("message conversion failed")
	ErrInvalidView        = errors.New("invalid view")
	ErrPayloadEncoding    = errors.New("payload encoding failed")
)

// Attack-related errors
var (
	ErrInvalidAttackType = errors.New("invalid attack type")
	ErrAttackNotActive   = errors.New("attack is not active")
	ErrAttackExecution   = errors.New("attack execution failed")
)

// Integration-related errors
var (
	ErrInvalidIntegration = errors.New("invalid integration")
	ErrEngineNotSet       = errors.New("QBFT engine not set")
)

// ErrUnsupportedMessageType is returned when a message type is not supported for replay
var ErrUnsupportedMessageType = errors.New("unsupported message type for replay")
