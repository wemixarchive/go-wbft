package types

import (
	wbftmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/log"
)

// AttackType by string
const (
	AttackSilent    = "silent"
	AttackTamper    = "tamper"
	AttackFake      = "fake"
	AttackOmit      = "omit"
	AttackRoleSpoof = "roleSpoof"
	AttackReplay    = "replay"
)

// QBFTMessage codes mapping
const (
	WBFTPrePrepareCode  = wbftmessage.PreprepareCode  // 0x12
	WBFTPrepareCode     = wbftmessage.PrepareCode     // 0x13
	WBFTCommitCode      = wbftmessage.CommitCode      // 0x14
	WBFTRoundChangeCode = wbftmessage.RoundChangeCode // 0x15
)

// OmitCommand constants for omit attack
const (
	// PrePrepare

	OmitCommandPrevPrepareSeal uint64 = 1
	OmitCommandPrevCommitSeal  uint64 = 2
	OmitCommandRoundChange     uint64 = 3
	OmitCommandPrepareMessage  uint64 = 4

	// Propagation

	OmitCommandPrepareSeal uint64 = 1
	OmitCommandCommitSeal  uint64 = 2
)

func ParseOmitCommand(code MessageCode, cmd uint64) string {
	switch code {
	case MessageCodePrePrepare:
		switch cmd {
		case OmitCommandPrevPrepareSeal:
			return "Omit Prev Prepare Seal"
		case OmitCommandCommitSeal:
			return "Omit Prev Commit Seal"
		case OmitCommandRoundChange:
			return "Omit Round Change Seal"
		case OmitCommandPrepareMessage:
			return "Omit Prepare Message Seal"
		}
	case MessageCodePropagation:
		switch cmd {
		case OmitCommandPrepareSeal:
			return "Omit Prepare Seal"
		case OmitCommandCommitSeal:
			return "Omit Commit Seal"
		}
	default:
		log.Warn("[BYZ] unknown Omit command", "code", code, "cmd", cmd)
	}
	return ""
}

// Node roles
const (
	RoleValidator = "validator"
	RoleProposer  = "proposer"
	RoleObserver  = "observer"
)

// Default limits
const (
	DefaultMaxMessageHistorySize = 1000
	DefaultMaxConcurrentAttacks  = 100
	DefaultMessageTimeout        = 30 // seconds
)

// Direction of event
const (
	DirectionSend    = "send"
	DirectionReceive = "receive"
)
