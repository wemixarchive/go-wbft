package types

import wbftmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"

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
