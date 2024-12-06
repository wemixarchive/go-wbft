package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
)

// addToExtraSeal adds the message to extraSeals which is read when making block
func (c *Core) addToExtraSeal(msg qbftmessage.QBFTMessage) error {
	logger := c.currentLogger(true, msg)
	var block *types.Block
	var ok bool

	if c.state == StateAcceptRequest {
		block, ok = c.priorState.Proposal().(*types.Block)
	} else {
		block, ok = c.current.Proposal().(*types.Block)
	}
	if !ok {
		return errInvalidMessage
	}

	// validate seal
	if prepareMsg, ok := msg.(*qbftmessage.Prepare); ok {
		// Check digest
		if prepareMsg.Digest != block.Hash() {
			logger.Error("QBFT: invalid PREPARE message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(block.Header(), uint32(prepareMsg.CommonPayload.Round.Uint64()), SealTypePrepare,
			prepareMsg.PrepareSeal, prepareMsg.Source()); err != nil {
			return errInvalidSeal
		}
	} else if commitMsg, ok := msg.(*qbftmessage.Commit); ok {
		// Check digest
		if commitMsg.Digest != block.Hash() {
			logger.Error("QBFT: invalid COMMIT message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(block.Header(), uint32(commitMsg.CommonPayload.Round.Uint64()), SealTypeCommit,
			commitMsg.CommitSeal, commitMsg.Source()); err != nil {
			return errInvalidSeal
		}
	} else {
		return errInvalidExtraSealMessage
	}

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	view := msg.View()
	c.extraSeals.Push(msg, toPriority(&view))
	logger.Info("QBFT: new extra seal message", "extra_seal_size", c.extraSeals.Size())
	return nil
}

// ProcessExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend preparing new block
func (c *Core) ProcessExtraSeal(lastProposal qbft.Proposal, priorRound *big.Int) (map[common.Hash][]byte, map[common.Hash][]byte) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	preparedSeal := make(map[common.Hash][]byte)
	committedSeal := make(map[common.Hash][]byte)
	latestView := qbft.View{
		Round:    priorRound,
		Sequence: lastProposal.Number(),
	}

	for !c.extraSeals.Empty() {
		msg, _ := c.extraSeals.Pop()
		code := msg.Code()
		view := msg.View()

		// when view has lower view than latestView,
		// discard all remaining seals in queue
		if latestView.Cmp(&view) > 0 {
			c.extraSeals.Reset()
			break
		}

		if code == qbftmessage.PrepareCode {
			prepareMsg := msg.(*qbftmessage.Prepare)
			if prepareMsg.Digest == lastProposal.Hash() {
				preparedSeal[common.BytesToHash(prepareMsg.PrepareSeal[:])] = prepareMsg.PrepareSeal[:]
			}
		} else if code == qbftmessage.CommitCode {
			commitMsg := msg.(*qbftmessage.Commit)
			if commitMsg.Digest == lastProposal.Hash() {
				committedSeal[common.BytesToHash(commitMsg.CommitSeal[:])] = commitMsg.CommitSeal[:]
			}
		}
	}

	return preparedSeal, committedSeal
}

func toPriority(view *qbft.View) int64 {
	return int64(view.Sequence.Uint64()*1000 + view.Round.Uint64())
}

func (c *Core) ExtraSealsLen() int {
	// used in test code
	return c.extraSeals.Size()
}
