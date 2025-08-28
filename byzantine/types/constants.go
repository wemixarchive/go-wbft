package types

import (
	wbftmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/log"
)

// AttackType by string
const (
	AttackPolicy    = "policy"
	AttackTamper    = "tamper"
	AttackFake      = "fake"
	AttackOmit      = "omit"
	AttackRoleSpoof = "roleSpoof"
	AttackReplay    = "replay"
	AttackStore     = "store"
	AttackDos       = "dos"
)

// WBFTMessage codes mapping
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
		log.Warn("BYZ: unknown Omit command", "code", code, "cmd", cmd)
	}
	return ""
}

// Target defines the common set of fields that can be manipulated
// across various Byzantine attack scenarios.
const (
	// Block targets
	TargetBlockReward string = "block.reward"

	// Header targets
	TargetHeaderParentHash  string = "header.parentHash"
	TargetHeaderUncleHash   string = "header.uncleHash"
	TargetHeaderCoinbase    string = "header.coinbase"
	TargetHeaderRoot        string = "header.root"
	TargetHeaderTxHash      string = "header.txHash"
	TargetHeaderReceiptHash string = "header.receiptHash"
	TargetHeaderBloom       string = "header.bloom"
	TargetHeaderDifficulty  string = "header.difficulty"
	TargetHeaderNumber      string = "header.number"
	TargetHeaderGasLimit    string = "header.gasLimit"
	TargetHeaderGasUsed     string = "header.gasUsed"
	TargetHeaderTime        string = "header.time"
	TargetHeaderExtra       string = "header.extra"
	TargetHeaderMixDigest   string = "header.mixDigest"
	TargetHeaderNonce       string = "header.nonce"
	TargetHeaderEpochInfo   string = "header.epochinfo"

	// Transaction targets
	TargetTxCount string = "tx.count"
	TargetTxSign  string = "tx.sign"
	TargetTxNonce string = "tx.nonce"
	TargetTxData  string = "tx.data"
	TargetTxValue string = "tx.value"

	// Consensus message targets
	TargetMsgRound          string = "msg.round"
	TargetMsgSequence       string = "msg.sequence"
	TargetMsgPreparedRound  string = "msg.pr"
	TargetMsgPreparedDigest string = "msg.pd"
	TargetMsgDigest         string = "msg.digest"
	TargetMsgJustification  string = "msg.justification"
	TargetMsgProposal       string = "msg.proposal"

	// Seal targets
	TargetPrevCommitSeal  string = "seal.prevcommit"
	TargetPrevPrePareSeal string = "seal.prevprepare"
	TargetCommitSeal      string = "seal.commit"
	TargetPrePareSeal     string = "seal.prepare"

	TargetMsgPolicyDirection    string = "policy.direction"
	TargetMsgPolicySendOriginal string = "policy.original"
	TargetMsgPolicyDelay        string = "policy.delay"
	TargetMsgPolicyTargets      string = "policy.targets"

	TargetSpoofedRole string = "spoofed_role"
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
