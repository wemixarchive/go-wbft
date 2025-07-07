package types

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ConsensusContext - Consensus context for WBFT messages
type ConsensusContext struct {
	MessageType  AttackType
	MessageCode  MessageCode
	Sequence     uint64
	Round        uint64
	From         common.Address
	Block        *types.Block     // for block-related attacks
	Seals        [][]byte         // for seal manipulation
	MessageData  interface{}      // flexible data container
	Validators   []common.Address // current validator set
	ProposerAddr common.Address   // current proposer

	// Additional metadata
	Direction   string
	IsProposer  bool
	BlockNumber uint64
	Timestamp   time.Time
}

type Proposal struct {
	Number       *big.Int       `json:"number"`
	Hash         common.Hash    `json:"hash"`
	ParentHash   common.Hash    `json:"parentHash"`
	Coinbase     common.Address `json:"coinbase"`
	Timestamp    uint64         `json:"timestamp"`
	Transactions []Transaction  `json:"transactions"`
	Header       *BlockHeader   `json:"header"`

	// QBFT specific
	PrevSeals       []Seal            `json:"prevSeals"`
	RoundChanges    []RoundChangeInfo `json:"roundChanges,omitempty"`
	PrepareMessages []PrepareInfo     `json:"prepareMessages,omitempty"`
}

type PrepareMessage struct {
	View      View           `json:"view"`
	Digest    common.Hash    `json:"digest"`
	Signer    common.Address `json:"signer"`
	Signature []byte         `json:"signature"`
}

type CommitMessage struct {
	View      View           `json:"view"`
	Digest    common.Hash    `json:"digest"`
	Signer    common.Address `json:"signer"`
	Signature []byte         `json:"signature"`

	// Committed seal
	CommittedSeal []byte `json:"committedSeal"`
}

type RoundChangeMessage struct {
	View           View           `json:"view"`
	PreparedRound  *uint64        `json:"preparedRound,omitempty"`
	PreparedDigest *common.Hash   `json:"preparedDigest,omitempty"`
	Signer         common.Address `json:"signer"`
	Signature      []byte         `json:"signature"`
}

// View - QBFT view (sequence + round)
type View struct {
	Sequence uint64 `json:"sequence"`
	Round    uint64 `json:"round"`
}

type BlockHeader struct {
	ParentHash  common.Hash    `json:"parentHash"`
	Coinbase    common.Address `json:"miner"`
	Root        common.Hash    `json:"stateRoot"`
	TxHash      common.Hash    `json:"transactionsRoot"`
	ReceiptHash common.Hash    `json:"receiptsRoot"`
	Number      *big.Int       `json:"number"`
	GasLimit    uint64         `json:"gasLimit"`
	GasUsed     uint64         `json:"gasUsed"`
	Time        uint64         `json:"timestamp"`
	Extra       []byte         `json:"extraData"`
	MixDigest   common.Hash    `json:"mixHash"`

	// Block hash
	Hash common.Hash `json:"hash"`
}

type Transaction struct {
	Hash     common.Hash     `json:"hash"`
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Value    *big.Int        `json:"value"`
	Gas      uint64          `json:"gas"`
	GasPrice *big.Int        `json:"gasPrice"`
	Nonce    uint64          `json:"nonce"`
	Data     []byte          `json:"input"`

	// Signature values
	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
}

type Seal struct {
	Validator common.Address `json:"validator"`
	Signature []byte         `json:"signature"`
}

// RoundChangeInfo - PrePrepare에 포함되는 Round Change 정보
type RoundChangeInfo struct {
	Message   RoundChangeMessage `json:"message"`
	Validator common.Address     `json:"validator"`
}

// PrepareInfo - PrePrepare에 포함되는 Prepare 정보
type PrepareInfo struct {
	Message   PrepareMessage `json:"message"`
	Validator common.Address `json:"validator"`
}

// === Message Implementations ===

// PrePrepareMessage - PrePrepare 메시지 구현
type PrePrepareMessage struct {
	View      View           `json:"view"`
	Proposal  Proposal       `json:"proposal"`
	Signer    common.Address `json:"signer"`
	Signature []byte         `json:"signature"`
}

func (m *PrePrepareMessage) Code() MessageCode {
	return MessageCodePrePrepare
}

func (m *PrePrepareMessage) Sequence() uint64 {
	return m.View.Sequence
}

func (m *PrePrepareMessage) Round() uint64 {
	return m.View.Round
}

func (m *PrePrepareMessage) Hash() common.Hash {
	return m.Proposal.Hash
}

func (m *PrePrepareMessage) Payload() []byte {
	// Implementation for getting raw payload
	return nil
}

func (m *PrePrepareMessage) Encode() ([]byte, error) {
	// Implementation for encoding
	return nil, nil
}

func (m *PrePrepareMessage) Decode(data []byte) error {
	// Implementation for decoding
	return nil
}

func (m *PrePrepareMessage) Validate() error {
	if m.Proposal.Number == nil {
		return ErrInvalidProposal
	}
	return nil
}

// Similar implementations for PrepareMessage, CommitMessage, RoundChangeMessage...
func (m *PrepareMessage) Code() MessageCode       { return MessageCodePrepare }
func (m *PrepareMessage) Sequence() uint64        { return m.View.Sequence }
func (m *PrepareMessage) Round() uint64           { return m.View.Round }
func (m *PrepareMessage) Hash() common.Hash       { return m.Digest }
func (m *PrepareMessage) Payload() []byte         { return nil }
func (m *PrepareMessage) Encode() ([]byte, error) { return nil, nil }
func (m *PrepareMessage) Decode([]byte) error     { return nil }
func (m *PrepareMessage) Validate() error         { return nil }

func (m *CommitMessage) Code() MessageCode       { return MessageCodeCommit }
func (m *CommitMessage) Sequence() uint64        { return m.View.Sequence }
func (m *CommitMessage) Round() uint64           { return m.View.Round }
func (m *CommitMessage) Hash() common.Hash       { return m.Digest }
func (m *CommitMessage) Payload() []byte         { return nil }
func (m *CommitMessage) Encode() ([]byte, error) { return nil, nil }
func (m *CommitMessage) Decode([]byte) error     { return nil }
func (m *CommitMessage) Validate() error         { return nil }

func (m *RoundChangeMessage) Code() MessageCode { return MessageCodeRoundChange }
func (m *RoundChangeMessage) Sequence() uint64  { return m.View.Sequence }
func (m *RoundChangeMessage) Round() uint64     { return m.View.Round }
func (m *RoundChangeMessage) Hash() common.Hash {
	if m.PreparedDigest != nil {
		return *m.PreparedDigest
	}
	return common.Hash{}
}
func (m *RoundChangeMessage) Payload() []byte         { return nil }
func (m *RoundChangeMessage) Encode() ([]byte, error) { return nil, nil }
func (m *RoundChangeMessage) Decode([]byte) error     { return nil }
func (m *RoundChangeMessage) Validate() error         { return nil }
