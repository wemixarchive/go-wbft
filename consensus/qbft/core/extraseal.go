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
		logger.Warn("QBFT: extra seal from self")
		return
	}

	logger.Trace("QBFT: new extra seal message", "extra_seal_size", len(c.backlogs))

	c.extrasealsMu.Lock()
	defer c.extrasealsMu.Unlock()

	extraseals := c.extraseals[src]
	if extraseals == nil {
		extraseals = prque.New[int64, qbftmessage.QBFTMessage](nil)
		c.extraseals[src] = extraseals
	}
	view := msg.View()
	extraseals.Push(msg, toPriority(msg.Code(), &view))
}

// processBacklog lookup for future messages that have been backlogged and post it on
// the event channel so main handler loop can handle it

// It is called on every state change
func (c *Core) processExtraSeal() {
	c.extrasealsMu.Lock()
	defer c.extrasealsMu.Unlock()

	for srcAddress, extraseal := range c.extraseals {
		if extraseal == nil {
			continue
		}
		_, src := c.valSet.GetByAddress(srcAddress)
		if src == nil {
			// validator is not available
			delete(c.extraseals, srcAddress)
			continue
		}
		logger := c.logger.New("from", src, "state", c.state)
		isFuture := false

		logger.Trace("QBFT: process extraseal")

		// We stop processing if
		//   1. extra seal message is empty
		//   2. The first message in queue is a future message
		for !(extraseal.Empty() || isFuture) {
			msg, prio := extraseal.Pop()

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
					logger.Trace("QBFT: stop processing extraseal", "msg", msg)
					extraseal.Push(msg, prio)
					isFuture = true
					break
				}
				logger.Trace("QBFT: skip extra seal message", "msg", msg, "err", err)
				continue
			}
			logger.Trace("QBFT: post extraseal event", "msg", msg)

			event.src = src
			go c.sendEvent(event)
		}
	}
}
