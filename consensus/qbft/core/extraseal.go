package core

import (
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
)

// addToBacklog allows to postpone the processing of future messages

// it adds the message to backlog which is read on every state change
func (c *Core) addToExtraSeal(msg qbftmessage.QBFTMessage) {
	logger := c.currentLogger(true, msg)

	src := msg.Source()
	if src == c.Address() {
		logger.Warn("QBFT: backlog from self")
		return
	}

	logger.Trace("QBFT: new backlog message", "backlogs_size", len(c.backlogs))

	c.backlogsMu.Lock()
	defer c.backlogsMu.Unlock()

	backlog := c.backlogs[src]
	if backlog == nil {
		backlog = prque.New[int64, qbftmessage.QBFTMessage](nil)
		c.backlogs[src] = backlog
	}
	view := msg.View()
	backlog.Push(msg, toPriority(msg.Code(), &view))
}

// processBacklog lookup for future messages that have been backlogged and post it on
// the event channel so main handler loop can handle it

// It is called on every state change
func (c *Core) processExtraSeal() {
	c.backlogsMu.Lock()
	defer c.backlogsMu.Unlock()

	for srcAddress, backlog := range c.backlogs {
		if backlog == nil {
			continue
		}
		_, src := c.valSet.GetByAddress(srcAddress)
		if src == nil {
			// validator is not available
			delete(c.backlogs, srcAddress)
			continue
		}
		logger := c.logger.New("from", src, "state", c.state)
		isFuture := false

		logger.Trace("QBFT: process backlog")

		// We stop processing if
		//   1. backlog is empty
		//   2. The first message in queue is a future message
		for !(backlog.Empty() || isFuture) {
			msg, prio := backlog.Pop()

			var code uint64
			var view qbft.View
			var event backlogEvent

			code = msg.Code()
			view = msg.View()
			event.msg = msg

			// Push back if it's a future message
			err := c.checkMessage(code, &view)
			if err != nil {
				if err == errFutureMessage {
					// this is still a future message
					logger.Trace("QBFT: stop processing backlog", "msg", msg)
					backlog.Push(msg, prio)
					isFuture = true
					break
				}
				logger.Trace("QBFT: skip backlog message", "msg", msg, "err", err)
				continue
			}
			logger.Trace("QBFT: post backlog event", "msg", msg)

			event.src = src
			go c.sendEvent(event)
		}
	}
}
