package core

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
)

// addToExtraSeal adds the message to extraSeals which is read when making block
func (c *Core) addToExtraSeal(msg qbftmessage.QBFTMessage) {
	logger := c.currentLogger(true, msg)

	logger.Trace("QBFT: new extra seal message", "extra_seal_size", c.extraSeals.Size())

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	view := msg.View()
	c.extraSeals.Push(msg, toPriority(&view))
	fmt.Println(":!!!!!!!!!!!!!!!!!! seal added")
}

// ProcessExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend preparing new block
func (c *Core) ProcessExtraSeal() ([][]byte, [][]byte) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	var preparedSeal [][]byte
	var committedSeal [][]byte
	latestView := qbft.View{
		Round:    common.Big0,
		Sequence: common.Big0,
	}

	for !(c.extraSeals.Empty()) {
		msg, _ := c.extraSeals.Pop()
		code := msg.Code()
		view := msg.View()

		// store latestView among extraSeals
		// when view has lower view than latestView,
		// discard all remaining seals in queue
		if latestView.Cmp(&view) < 0 {
			latestView = view
		} else if latestView.Cmp(&view) > 0 {
			c.extraSeals.Reset()
			break
		}

		lastProposal, _ := c.backend.LastProposal()

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
