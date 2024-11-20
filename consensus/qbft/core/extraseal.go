package core

import (
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
)

// it adds the message to extraSeals which is read when making block
func (c *Core) addToExtraSeal(msg qbftmessage.QBFTMessage) {
	logger := c.currentLogger(true, msg)

	src := msg.Source()
	if src == c.Address() {
		logger.Warn("QBFT: extra seal from self")
		return
	}
	logger.Trace("QBFT: new extra seal message", "extra_seal_size", c.extraSeals.Size())

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	view := msg.View()
	c.extraSeals.Push(msg, toPriority(&view))
}

// processExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend making new block
func (c *Core) ProcessExtraSeal() ([][]byte, [][]byte) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	var preparedSeal [][]byte
	var committedSeal [][]byte
	var latestView qbft.View

	for !(c.extraSeals.Empty()) {
		msg, _ := c.extraSeals.Pop()
		code := msg.Code()
		view := msg.View()

		// store latestView among extraSeals
		if latestView.Cmp(&view) < 0 {
			latestView = view
		} else if latestView.Cmp(&view) > 0 {
			// if view has lower view than latestView,
			// discard all remaining seals in queue
			c.extraSeals.Reset()
			break
		} else {
			if code == qbftmessage.PrepareCode {
				prepareMsg := msg.(*qbftmessage.Prepare)
				if prepareMsg.Digest == c.current.Proposal().Hash() {
					preparedSeal = append(preparedSeal, prepareMsg.PrepareSeal[:])
				}

			} else if code == qbftmessage.CommitCode {
				commitMsg := msg.(*qbftmessage.Commit)
				if commitMsg.Digest == c.current.Proposal().Hash() {
					committedSeal = append(committedSeal, commitMsg.CommitSeal[:])
				}
			}
		}
	}
	return preparedSeal, committedSeal
}

func toPriority(view *qbft.View) int64 {
	return int64(view.Sequence.Uint64()*1000 + view.Round.Uint64())
}
