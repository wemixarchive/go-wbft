package types

import wbftmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"

// AttackType by string
const (
	AttackDoubleVote = "double_vote"
	AttackSilent     = "silent"
	AttackTamper     = "tamper"
	AttackFake       = "fake"
	AttackOmit       = "omit"
	AttackRoleSpoof  = "roleSpoof"
	AttackReplay     = "replay"
	AttackFlooding   = "flooding"
)

// QBFT Message coes mapping
const (
	// Map to actual QBFT message codes
	QBFTPrePrepareCode  = wbftmessage.PreprepareCode  // 0x12
	QBFTPrepareCode     = wbftmessage.PrepareCode     // 0x13
	QBFTCommitCode      = wbftmessage.CommitCode      // 0x14
	QBFTRoundChangeCode = wbftmessage.RoundChangeCode // 0x15
)

// Byzantine message code to QBFT code mapping
var MessageCodeToQBFT = map[MessageCode]uint64{
	MessageCodePrePrepare:            QBFTPrePrepareCode,
	MessageCodePrepare:               QBFTPrepareCode,
	MessageCodeCommit:                QBFTCommitCode,
	MessageCodeRoundChange:           QBFTRoundChangeCode,
	MessageCodeRoundChangePrePrepare: QBFTPrePrepareCode, // Special case
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
