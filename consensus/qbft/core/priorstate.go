package core

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/consensus/qbft"
)

// priorState stores prior consensus state
// for collecting extra seals during StateAcceptRequest
type priorState struct {
	mu           *sync.RWMutex
	round        *big.Int
	proposal     qbft.Proposal
	validatorSet qbft.ValidatorSet
}

func (c *Core) updatePriorState(priorRound *big.Int, priorProposal qbft.Proposal, priorValSet qbft.ValidatorSet) {
	c.priorState.mu.Lock()
	defer c.priorState.mu.Unlock()
	c.priorState.round = priorRound
	if priorProposal != nil {
		c.priorState.proposal = priorProposal
	}
	if priorValSet != nil {
		c.priorState.validatorSet = priorValSet
	}
}
func (p *priorState) Round() *big.Int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.round
}

func (p *priorState) Proposal() qbft.Proposal {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.proposal
}

func (p *priorState) Validators() qbft.ValidatorSet {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.validatorSet
}
