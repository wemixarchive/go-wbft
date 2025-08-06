package eth

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"math/big"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
)

// Byzantine Attack For ethHandler

func (h *ethHandler) SetByzantineHook(hook btypes.ConsensusHook) {
	(*handler)(h).byzantineHook = hook
	log.Debug("[BYZ] Byzantine hook integrated with eth backend handler")
}

func (h *ethHandler) ByzantineHook() btypes.ConsensusHook {
	if (*handler)(h).byzantineHook == nil {
		log.Warn("[BYZ] Byzantine hook is not set, returning nil")
		return nil
	}
	return (*handler)(h).byzantineHook
}

// Byzantine Attack For Ethereum backend

// SetByzantineHook sets the Byzantine hook
func (s *Ethereum) SetByzantineHook(hook btypes.ConsensusHook) {
	if s.handler != nil {
		s.handler.byzantineHook = hook
		log.Debug("[BYZ] Byzantine hook integrated with eth backend handler")
	} else {
		log.Warn("[BYZ] handler is nil. can't set byzantine hook")
	}
}

// createBlockWithMissingSeals creates a modified block with certain seals removed based on the omit command
func (h *handler) createBlockWithMissingSeals(block *types.Block, cmd uint64) *types.Block {
	// Get the original header
	header := block.Header()

	// Create a new header with deep copy of all fields
	newHeader := &types.Header{
		ParentHash:  header.ParentHash,
		UncleHash:   header.UncleHash,
		Coinbase:    header.Coinbase,
		Root:        header.Root,
		TxHash:      header.TxHash,
		ReceiptHash: header.ReceiptHash,
		Bloom:       header.Bloom,
		Difficulty:  new(big.Int).Set(header.Difficulty),
		Number:      new(big.Int).Set(header.Number),
		GasLimit:    header.GasLimit,
		GasUsed:     header.GasUsed,
		Time:        header.Time,
		Extra:       make([]byte, len(header.Extra)),
		MixDigest:   header.MixDigest,
		Nonce:       header.Nonce,
	}
	if header.BaseFee != nil {
		newHeader.BaseFee = new(big.Int).Set(header.BaseFee)
	}
	copy(newHeader.Extra, header.Extra)

	wbftExtra, err := types.ExtractWBFTExtra(newHeader)
	if err != nil {
		log.Warn("[BYZ] Failed to extract WBFTExtra", "err", err)
		return nil
	}

	// Modify seals based on command
	// Note: WBFT ignores PreparedSeal and CommittedSeal when calculating block hash
	// So we need to modify other fields to create a different hash
	switch cmd {
	case btypes.OmitCommandPrepareSeal:
		log.Trace("[BYZ] Setting PreparedSeal to nil")
		wbftExtra.PreparedSeal = nil
	case btypes.OmitCommandCommitSeal:
		log.Trace("[BYZ] Setting CommittedSeal to nil")
		wbftExtra.CommittedSeal = nil
	default:
		log.Warn("[BYZ] Unknown omit command", "cmd", cmd)
		return nil
	}

	payload, err := rlp.EncodeToBytes(wbftExtra)
	if err != nil {
		log.Warn("[BYZ] Failed to encode WBFTExtra", "err", err)
		return nil
	}
	newHeader.Extra = make([]byte, len(payload))
	copy(newHeader.Extra, payload)

	// Create a new block with the modified header
	return types.NewBlockWithHeader(newHeader).WithBody(block.Transactions(), block.Uncles())
}
