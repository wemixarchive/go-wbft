package core

import (
	"errors"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/log"
)

func (c *Core) byzantineSendPreprepareFromNonProposer() error {
	// Byzantine logic: Check if we have a role spoof attack configured
	var attacks map[btypes.AttackType]*btypes.ExecutableAttack
	hook := c.backend.ByzantineHook()
	if hook != nil {
		attacks = hook.GetExecutableAttacks(btypes.MessageCodePrePrepare, c.current.Sequence().Uint64(), c.current.Round().Uint64())
	}

	if at := attacks[btypes.AttackTypeRoleSpoofed]; at != nil && at.RoleSpoofParams != nil {
		currentRound := c.currentView().Round

		log.Info("[BYZ] WBFT: received quorum of ROUND-CHANGE messages")
		log.Info("[BYZ] Non-proposer attempting to send PrePrepare after round change",
			"sequence", c.current.Sequence().Uint64(),
			"round", currentRound.Uint64(),
			"actual_proposer", c.valSet.GetProposer().Address(),
			"byzantine_node", c.Address())

		// Prepare the same data as proposer would
		_, proposal := c.highestPrepared(currentRound)
		if proposal == nil {
			if c.current != nil && c.current.pendingRequest != nil {
				proposal = c.current.pendingRequest.Proposal
			} else {
				log.Warn("[BYZ] Cannot execute role spoof: no proposal available")
				return nil
			}
		}

		// Prepare justification for ROUND-CHANGE messages
		roundChangeMessages := c.roundChangeSet.roundChanges[currentRound.Uint64()]
		rcSignedPayloads := make([]*wbfmessage.SignedRoundChangePayload, 0)
		for _, m := range roundChangeMessages.Values() {
			rcMsg := m.(*wbfmessage.RoundChange)
			rcSignedPayloads = append(rcSignedPayloads, &rcMsg.SignedRoundChangePayload)
		}

		prepareMessages := c.roundChangeSet.prepareMessages[currentRound.Uint64()]
		if err := isJustified(proposal, rcSignedPayloads, prepareMessages, c.valSet.QuorumSize()); err != nil {
			log.Error("[BYZ] Invalid ROUND-CHANGE message justification", "err", err)
			return nil
		}

		r := &Request{
			Proposal:        proposal,
			RCMessages:      roundChangeMessages,
			PrepareMessages: prepareMessages,
		}

		if c.sendByzantinePreprepareMsg(hook, r, attacks) {
			log.Info("[BYZ] send Byzantine prepares message")
		}
	}
	return errors.New("[BYZ] no exist available attack")
}
