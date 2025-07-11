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
		case OmitCommandPrevCommitSeal:
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

// FakeTarget constants for fake attack
const (
	FakeTargetPrevCommitSeal   string = "prevCommitSeal"
	FakeTargetPrevPrePareSeal  string = "prevPrePareSeal"
	FakeTargetCommitSeal       string = "commitSeal"
	FakeTargetPrePareSeal      string = "prePareSeal"
	FakeTargetTransactionCount string = "Transaction.Count"
)

// TamperTarget represents fields that can be tampered
const (
	// PrePrepare message targets
	TamperProposalHeaderCoinbase    string = "Header.Coinbase"
	TamperProposalHeaderNumber      string = "Header.Number"
	TamperProposalHeaderTime        string = "Header.Time"
	TamperProposalHeaderParentHash  string = "Header.ParentHash"
	TamperProposalHeaderStateRoot   string = "Header.StateRoot"
	TamperProposalHeaderTxHash      string = "Header.TxHash"
	TamperProposalHeaderReceiptHash string = "Header.ReceiptHash"
	TamperProposalHeaderBloom       string = "Header.Bloom"
	TamperProposalHeaderDifficulty  string = "Header.Difficulty"
	TamperProposalHeaderGasLimit    string = "Header.GasLimit"
	TamperProposalHeaderGasUsed     string = "Header.GasUsed"
	TamperProposalHeaderExtra       string = "Header.Extra"
	TamperProposalHeaderMixDigest   string = "Header.MixDigest"
	TamperProposalHeaderNonce       string = "Header.Nonce"

	TamperDigest string = "Digest"
	TamperReward string = "Reward"

	// Transaction targets
	TamperTransactionSign    string = "Transaction.Sign"
	TamperTransactionBalance string = "Transaction.Balance"
	TamperTransactionNonce   string = "Transaction.Nonce"
	TamperTransactionData    string = "Transaction.Data"
	TamperTransactionValue   string = "Transaction.Value"

	// Common message targets
	TamperMessageRound     string = "Message.Round"
	TamperMessageSequence  string = "Message.Sequence"
	TamperMessageSealType  string = "Message.SealType"
	TamperMessageHeader    string = "Message.Header"
	TamperMessageSignature string = "Message.Signature"

	// RoundChange specific
	TamperRoundChangePreparedRound  string = "RoundChange.PreparedRound"
	TamperRoundChangePreparedDigest string = "RoundChange.PreparedDigest"

	// Seal targets
	TamperSealPrepare  string = "Seal.Prepare"
	TamperSealCommit   string = "Seal.Commit"
	TamperSealPrevious string = "Seal.Previous"
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
