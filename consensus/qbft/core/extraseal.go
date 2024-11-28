package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
)

// addToExtraSeal adds the message to extraSeals which is read when making block
func (c *Core) addToExtraSeal(msg qbftmessage.QBFTMessage) error {
	logger := c.currentLogger(true, msg)
	block, ok := c.current.Proposal().(*types.Block)
	if !ok {
		return errInvalidMessage
	}

	// validate seal
	if prepareMsg, ok := msg.(*qbftmessage.Prepare); !ok {
		if commitMsg, ok := msg.(*qbftmessage.Commit); !ok {
			return errInvalidExtraSealMessage
		} else {
			// Check digest
			if commitMsg.Digest != c.current.Proposal().Hash() {
				logger.Error("QBFT: invalid COMMIT message digest")
				return errInvalidMessage
			}

			if err := verifySeal(block.Header(), uint32(commitMsg.CommonPayload.Round.Uint64()), SealTypeCommit,
				commitMsg.CommitSeal, commitMsg.Source()); err != nil {
				return errInvalidSeal
			}
		}
	} else {
		// Check digest
		if prepareMsg.Digest != c.current.Proposal().Hash() {
			logger.Error("QBFT: invalid PREPARE message digest")
			return errInvalidMessage
		}

		if err := verifySeal(block.Header(), uint32(prepareMsg.CommonPayload.Round.Uint64()), SealTypePrepare,
			prepareMsg.PrepareSeal, prepareMsg.Source()); err != nil {
			return errInvalidSeal
		}
	}
	logger.Info("QBFT: new extra seal message", "extra_seal_size", c.extraSeals.Size())
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	view := msg.View()
	c.extraSeals.Push(msg, toPriority(&view))
	return nil
}

// ProcessExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend preparing new block
func (c *Core) ProcessExtraSeal(lastProposal qbft.Proposal, priorRound *big.Int) ([][]byte, [][]byte) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	var preparedSeal [][]byte
	var committedSeal [][]byte
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
				preparedSeal = append(preparedSeal, prepareMsg.PrepareSeal[:])
			}
		} else if code == qbftmessage.CommitCode {
			commitMsg := msg.(*qbftmessage.Commit)
			if commitMsg.Digest == lastProposal.Hash() {
				committedSeal = append(committedSeal, commitMsg.CommitSeal[:])
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
