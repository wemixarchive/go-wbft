// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

package wemix

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftBackend "github.com/ethereum/go-ethereum/consensus/wbft/backend"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

type DalgonaConsensus struct {
	legacy consensus.Engine
	wbft   *wbftBackend.Backend
	stopCh chan struct{}
}

func NewDalgonaEngine(legacyEngine consensus.Engine, config *wbft.Config, privateKey *ecdsa.PrivateKey, db ethdb.Database) consensus.Engine {
	wbft := wbftBackend.New(config, privateKey, db)

	result := &DalgonaConsensus{
		legacy: legacyEngine,
		wbft:   wbft,
	}
	return result
}

func (we *DalgonaConsensus) Start(
	config *params.ChainConfig,
	chain consensus.ChainHeaderReader,
	currentBlock func() *types.Block,
	subscribeChainHead func(ch chan<- core.ChainHeadEvent) event.Subscription,
	notifyNewRound func(waitTime time.Duration, round *big.Int)) {
	we.stopCh = make(chan struct{})
	we.wbft.Start(chain, currentBlock, rawdb.HasBadBlock, notifyNewRound)
}

func (we *DalgonaConsensus) Stop() {
	we.wbft.Stop()
	close(we.stopCh)
}

func (we *DalgonaConsensus) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

func (we *DalgonaConsensus) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	// The fact that this engine is used implies that the Dalgona hard fork is configured. (Dalgona is not nil)
	if chain.Config().IsDalgona(header.Number) {
		return we.wbft.VerifyHeader(chain, header)
	}
	return we.legacy.VerifyHeader(chain, header)
}

// VerifyHeaders verifies a batch of headers. If the batch spans the Dalgona
// fork boundary, it splits the headers into legacy and wbft slices and verifies
// each with the appropriate engine sequentially.
func (we *DalgonaConsensus) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	if chain.Config().IsDalgona(headers[0].Number) {
		return we.wbft.VerifyHeaders(chain, headers)
	} else if !chain.Config().IsDalgona(headers[len(headers)-1].Number) {
		return we.legacy.VerifyHeaders(chain, headers)
	}

	// Mixed batch spanning the Dalgona fork boundary.
	// Split into legacy and wbft slices so each engine can do batch verification
	// (which passes preceding headers as parents), avoiding ErrUnknownAncestor
	// that would occur if each header were verified individually via VerifyHeader.
	dalgonaBlock := chain.Config().DalgonaBlock
	splitIdx := int(new(big.Int).Sub(dalgonaBlock, headers[0].Number).Int64())

	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		// Verify legacy headers first. Sequential execution ensures results are
		// forwarded in order and the last legacy header is available as the parent
		// seed for the wbft batch below.
		legacyAbort, legacyResults := we.legacy.VerifyHeaders(chain, headers[:splitIdx])
		defer close(legacyAbort) // propagate abort to sub-engine goroutine on exit
		for i := 0; i < splitIdx; i++ {
			select {
			case <-abort:
				return
			case err := <-legacyResults:
				results <- err
			}
		}

		// Verify wbft headers with the last legacy header injected as the explicit
		// parent of headers[0], since it is not yet committed to the chain DB
		// during snap sync.
		wbftAbort, wbftResults := we.wbft.VerifyHeadersWithParent(chain, headers[splitIdx:], headers[splitIdx-1])
		defer close(wbftAbort) // propagate abort to sub-engine goroutine on exit
		for i := splitIdx; i < len(headers); i++ {
			select {
			case <-abort:
				return
			case err := <-wbftResults:
				results <- err
			}
		}
	}()

	return abort, results
}

func (we *DalgonaConsensus) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if chain.Config().IsDalgona(block.Number()) {
		return we.wbft.VerifyUncles(chain, block)
	}
	return we.legacy.VerifyUncles(chain, block)
}

func (we *DalgonaConsensus) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	if chain.Config().IsDalgona(header.Number) {
		return we.wbft.Prepare(chain, header)
	}
	return we.legacy.Prepare(chain, header)
}

func (we *DalgonaConsensus) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, withdrawals []*types.Withdrawal) error {
	if chain.Config().IsDalgona(header.Number) {
		return we.wbft.Finalize(chain, header, state, txs, uncles, withdrawals)
	} else {
		return we.legacy.Finalize(chain, header, state, txs, uncles, withdrawals)
	}
}

func (we *DalgonaConsensus) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	if chain.Config().IsDalgona(header.Number) {
		return we.wbft.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts, withdrawals)
	}
	return we.legacy.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts, withdrawals)
}

func (we *DalgonaConsensus) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	if chain.Config().IsDalgona(block.Number()) {
		return we.wbft.Seal(chain, block, results, stop)
	}
	return we.legacy.Seal(chain, block, results, stop)
}

func (we *DalgonaConsensus) SealHash(header *types.Header) common.Hash {
	// Wpoa does not support SealHash, so we suppose this block to be a wbft block
	return we.wbft.SealHash(header)
}

func (we *DalgonaConsensus) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	num := new(big.Int).Set(parent.Number)
	num = num.Add(num, big.NewInt(1))
	if chain.Config().IsDalgona(num) {
		return we.wbft.CalcDifficulty(chain, time, parent)
	}
	return we.legacy.CalcDifficulty(chain, time, parent)
}

func (we *DalgonaConsensus) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	// legacy has no consensus API
	return we.wbft.APIs(chain)
}

func (we *DalgonaConsensus) Close() error {
	err := we.legacy.Close()
	if err != nil {
		return err
	}
	return we.wbft.Close()
}

// CallEngineSpecific implements consensus.Engine
func (we *DalgonaConsensus) CallEngineSpecific(method string, args ...interface{}) interface{} {
	return we.wbft.CallEngineSpecific(method, args)
}

func (we *DalgonaConsensus) NewChainHead() error {
	return we.wbft.NewChainHead()
}

func (we *DalgonaConsensus) HandleMsg(address common.Address, data p2p.Msg) (bool, error) {
	return we.wbft.HandleMsg(address, data)
}

func (we *DalgonaConsensus) SetBroadcaster(broadcaster consensus.Broadcaster) {
	we.wbft.SetBroadcaster(broadcaster)
}
