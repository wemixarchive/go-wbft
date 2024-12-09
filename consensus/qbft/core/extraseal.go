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
	var (
		block    *types.Block
		ok       bool
		sealType SealType
	)

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
		sealType = SealTypePrepare
		// Check digest
		if prepareMsg.Digest != block.Hash() {
			logger.Error("QBFT: invalid extra PREPARE message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(block.Header(), uint32(prepareMsg.CommonPayload.Round.Uint64()), sealType,
			prepareMsg.PrepareSeal, prepareMsg.Source()); err != nil {
			return errInvalidSeal
		}
	} else if commitMsg, ok := msg.(*qbftmessage.Commit); ok {
		sealType = SealTypeCommit
		// Check digest
		if commitMsg.Digest != block.Hash() {
			logger.Error("QBFT: invalid extra COMMIT message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(block.Header(), uint32(commitMsg.CommonPayload.Round.Uint64()), sealType,
			commitMsg.CommitSeal, commitMsg.Source()); err != nil {
			return errInvalidSeal
		}
	} else {
		return errInvalidExtraSealMessage
	}

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	// store seals
	extraSeal, ok := c.extraSeals[msg.Source()]
	if !ok {
		extraSeal = make(map[SealType]qbftmessage.QBFTMessage)
		c.extraSeals[msg.Source()] = extraSeal
	}

	if !ok {
		extraSeal[sealType] = msg
	} else if existingView, incomingView := extraSeal[sealType].View(), msg.View(); existingView.Cmp(&incomingView) < 0 {
		extraSeal[sealType] = msg
	}

	if extraSeal[sealType] != nil {
		if existingView, incomingView := extraSeal[sealType].View(), msg.View(); existingView.Cmp(&incomingView) >= 0 {
			return nil
		}
	}

	logger.Info("QBFT: new extra seal message")
	return nil
}

// ProcessExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend preparing new block
func (c *Core) ProcessExtraSeal(lastProposal qbft.Proposal, priorRound *big.Int) (map[common.Hash][]byte, map[common.Hash][]byte) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	preparedSeal := make(map[common.Hash][]byte)
	committedSeal := make(map[common.Hash][]byte)
	// latestView is view to process extra seal
	latestView := qbft.View{
		Round:    priorRound,
		Sequence: lastProposal.Number(),
	}

	for _, seal := range c.extraSeals {
		if seal[SealTypePrepare] != nil {
			prepareMsg := seal[SealTypePrepare].(*qbftmessage.Prepare)
			view := prepareMsg.View()
			if latestView.Cmp(&view) == 0 && prepareMsg.Digest == lastProposal.Hash() {
				preparedSeal[common.BytesToHash(prepareMsg.PrepareSeal[:])] = prepareMsg.PrepareSeal[:]
			}
		}

		if seal[SealTypeCommit] != nil {
			commitMsg := seal[SealTypeCommit].(*qbftmessage.Commit)
			view := commitMsg.View()
			if latestView.Cmp(&view) == 0 && commitMsg.Digest == lastProposal.Hash() {
				preparedSeal[common.BytesToHash(commitMsg.CommitSeal[:])] = commitMsg.CommitSeal[:]
			}
		}
	}

	return preparedSeal, committedSeal
}

func toPriority(view *qbft.View) int64 {
	return int64(view.Sequence.Uint64()*1000 + view.Round.Uint64())
}
