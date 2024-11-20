// Modification Copyright 2024 The Wemix Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/consensus/istanbul/backend/engine_test.go (2024.07.25).
// Modified and improved for the wemix development.

package backend

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	qbftcore "github.com/ethereum/go-ethereum/consensus/qbft/core"
	qbftengine "github.com/ethereum/go-ethereum/consensus/qbft/engine"
	"github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/consensus/qbft/testutils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/fetcher"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
)

var blockEnqueueChannel chan *types.Block

type fakeBroadcaster struct {
	blockFetcher *fetcher.BlockFetcher
}

type otherNode struct {
	address    common.Address
	privateKey *ecdsa.PrivateKey
	balance    *big.Int
}

func makeFakeBroadcaster(chain *core.BlockChain) *fakeBroadcaster {
	blockEnqueueChannel = make(chan *types.Block)
	validator := func(header *types.Header) error {
		return chain.Engine().VerifyHeader(chain, header)
	}

	broadcastBlock := func(block *types.Block, propagate bool) {}

	heighter := func() uint64 {
		return chain.CurrentBlock().Number.Uint64()
	}

	//inserter := func(blocks types.Blocks) (int, error) {
	//	idx, err := chain.InsertChain(blocks)
	//	if err == nil {
	//		header := blocks[len(blocks)-1].Header()
	//		chain.SetFinalized(header)
	//		chain.SetSafe(header)
	//	}
	//	// ## Wemix END
	//	return idx, err
	//}

	fb := fakeBroadcaster{
		blockFetcher: fetcher.NewBlockFetcher(false, nil, chain.GetBlockByHash, validator, broadcastBlock, heighter, nil, nil, nil),
	}
	fb.blockFetcher.Start()
	return &fb
}

func (fb *fakeBroadcaster) Enqueue(id string, block *types.Block) {
	go func() { blockEnqueueChannel <- block }()
}

func (fb *fakeBroadcaster) FindPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	return m
}

func newBlockchainFromConfig(genesis *core.Genesis, nodeKeys []*ecdsa.PrivateKey, cfg *qbft.Config) (*core.BlockChain, *Backend, []otherNode) {
	memDB := rawdb.NewMemoryDatabase()

	// Use the first key as private key
	backend := New(cfg, nodeKeys[0], memDB)

	genesis.MustCommit(memDB, triedb.NewDatabase(memDB, triedb.HashDefaults))

	blockchain, err := core.NewBlockChain(memDB, nil, genesis, nil, backend, vm.Config{}, nil, nil)
	if err != nil {
		panic(err)
	}

	state, err := blockchain.StateAt(blockchain.Genesis().Root())
	if state == nil || err != nil {
		panic(err)
	}

	// Make virtual node struct for simulation
	nodes := make([]otherNode, 0)
	for i := 0; i < len(nodeKeys); i++ {
		address := crypto.PubkeyToAddress(nodeKeys[i].PublicKey)
		b := state.GetBalance(address).ToBig()
		nodes = append(nodes, otherNode{address, nodeKeys[i], b})
	}

	fb := makeFakeBroadcaster(blockchain)
	backend.broadcaster = fb

	backend.Start(blockchain, blockchain.CurrentFullBlock, rawdb.HasBadBlock)

	snap, err := backend.snapshot(blockchain, 0, common.Hash{}, nil)
	if err != nil {
		panic(err)
	}
	if snap == nil {
		panic("failed to get snapshot")
	}
	proposerAddr := snap.ValSet.GetProposer().Address()

	// find proposer key
	for _, key := range nodeKeys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		if addr.String() == proposerAddr.String() {
			backend.privateKey = key
			backend.address = addr
			backend.qbftEngine = qbftengine.NewEngine(backend.config, addr, backend.Sign)
		}
	}

	return blockchain, backend, nodes
}

// in this test, we can set n to 1, and it means we can process Istanbul and commit a
// block by one node. Otherwise, if n is larger than 1, we have to generate
// other fake events to process Istanbul.
func newBlockChain(n int) (*core.BlockChain, *Backend, []otherNode) {
	genesis, nodeKeys := testutils.GenesisAndKeys(n)

	config := copyConfig(qbft.DefaultConfig)

	return newBlockchainFromConfig(genesis, nodeKeys, config)
}

// copyConfig create a copy of qbft.Config, so that changing it does not update the original
func copyConfig(config *qbft.Config) *qbft.Config {
	cpy := *config
	return &cpy
}

func makeHeader(parent *types.Block, config *qbft.Config) *types.Header {
	blockNumber := parent.Number().Add(parent.Number(), common.Big1)
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     blockNumber,
		GasLimit:   core.CalcGasLimit(parent.GasLimit(), parent.GasLimit()),
		GasUsed:    0,
		Time:       parent.Time() + config.GetConfig(blockNumber).BlockPeriod,
		Difficulty: types.QBFTDefaultDifficulty,
	}
	return header
}

func makeBlock(chain *core.BlockChain, engine *Backend, parent *types.Block) *types.Block {
	block := makeBlockWithoutSeal(chain, engine, parent)
	state, _ := chain.State()
	block, _ = engine.FinalizeAndAssemble(chain, block.Header(), state, nil, nil, nil, nil)
	resultCh := make(chan *types.Block, 10)
	engine.Seal(chain, block, resultCh, nil)
	blk := <-resultCh
	return blk
}

func makeBlockWithoutSeal(chain *core.BlockChain, engine *Backend, parent *types.Block) *types.Block {
	header := makeHeader(parent, engine.config)
	engine.Prepare(chain, header)
	block := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	return block
}

func TestQBFTPrepare(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	header := makeHeader(chain.Genesis(), engine.config)
	err := engine.Prepare(chain, header)
	if err != nil {
		t.Errorf("error mismatch: have %v, want nil", err)
	}

	header.ParentHash = common.BytesToHash([]byte("1234567890"))
	err = engine.Prepare(chain, header)
	if err != consensus.ErrUnknownAncestor {
		t.Errorf("error mismatch: have %v, want %v", err, consensus.ErrUnknownAncestor)
	}
}

func TestSealStopChannel(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	stop := make(chan struct{}, 1)
	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})
	eventLoop := func() {
		ev := <-eventSub.Chan()
		_, ok := ev.Data.(qbft.RequestEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
		}
		stop <- struct{}{}
		eventSub.Unsubscribe()
	}
	resultCh := make(chan *types.Block, 10)
	go func() {
		err := engine.Seal(chain, block, resultCh, stop)
		if err != nil {
			t.Errorf("error mismatch: have %v, want nil", err)
		}
	}()
	go eventLoop()

	finalBlock := <-resultCh
	if finalBlock != nil {
		t.Errorf("block mismatch: have %v, want nil", finalBlock)
	}
}

func TestSealCommittedOtherHash(t *testing.T) {
	chain, engine, _ := newBlockChain(2)
	defer engine.Stop()
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	otherBlock := makeBlockWithoutSeal(chain, engine, block)
	expectedPreparedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)
	expectedCommittedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)

	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})
	defer eventSub.Unsubscribe()
	blockOutputChannel := make(chan *types.Block)
	stopChannel := make(chan struct{})

	if err := engine.Seal(chain, block, blockOutputChannel, stopChannel); err != nil {
		t.Error(err.Error())
	}

	ev := <-eventSub.Chan()
	if _, ok := ev.Data.(qbft.RequestEvent); !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}

	if err := engine.Commit(otherBlock, [][]byte{expectedPreparedSeal}, [][]byte{expectedCommittedSeal}, big.NewInt(0)); err != nil {
		t.Error(err.Error())
	}

	select {
	case <-blockOutputChannel:
		t.Error("Wrong block found!")
	case <-time.After(time.Second):
		//no block found, stop the sealing
		close(stopChannel)
	}

	output := <-blockOutputChannel
	if output != nil {
		t.Error("Block not nil!")
	}
}

func updateQBFTBlock(block *types.Block, addr common.Address) *types.Block {
	header := block.Header()
	header.Coinbase = addr
	return block.WithSeal(header)
}

func TestSealCommitted(t *testing.T) {
	chain, engine, _ := newBlockChain(2)
	defer engine.Stop()
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	expectedBlock := updateQBFTBlock(block, engine.Address())
	expectedPreparedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)
	expectedCommittedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)

	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})
	defer eventSub.Unsubscribe()
	resultCh := make(chan *types.Block, 10)

	if err := engine.Seal(chain, block, resultCh, make(chan struct{})); err != nil {
		t.Error(err.Error())
	}

	ev := <-eventSub.Chan()
	if _, ok := ev.Data.(qbft.RequestEvent); !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}

	if err := engine.Commit(expectedBlock, [][]byte{expectedPreparedSeal}, [][]byte{expectedCommittedSeal}, big.NewInt(0)); err != nil {
		t.Error(err.Error())
	}

	finalBlock := <-resultCh
	if finalBlock.Hash() != expectedBlock.Hash() {
		t.Errorf("hash mismatch: have %v, want %v", finalBlock.Hash(), expectedBlock.Hash())
	}
}

func TestVerifyHeaderForChainedBlock(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()

	montblancBlock := makeBlock(chain, engine, chain.Genesis())
	_, err := chain.InsertChain(types.Blocks{montblancBlock})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	qbftBlock := makeBlock(chain, engine, montblancBlock)
	_, err = chain.InsertChain(types.Blocks{qbftBlock})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	//create chain consists of genesisblock - montblanc hardfork block - regular qbft block
	testCases := []struct {
		block                  *types.Block
		headerManipulationFunc func(*types.Block) *types.Header
		expectedError          error
	}{
		{
			montblancBlock,
			func(block *types.Block) *types.Header { return block.Header() },
			nil,
		},
		{
			qbftBlock,
			func(block *types.Block) *types.Header { return block.Header() },
			nil,
		},
		{
			qbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevPreparedSeal([][]byte{})); err != nil {
					return nil
				}
				return header
			},
			qbftcommon.ErrEmptyPrevPreparedSeals,
		},
		{
			qbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevCommittedSeal([][]byte{})); err != nil {
					return nil
				}
				return header
			},
			// for qbftBlock, invalid preparedSeal occurs when validating preparedSeal because header is changed and gets wrong signer address
			qbftcommon.ErrInvalidPreparedSeals,
		},
		{
			montblancBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevPreparedSeal([][]byte{})); err != nil {
					return nil
				}
				return header
			},
			nil,
		},
		{
			montblancBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevCommittedSeal([][]byte{})); err != nil {
					return nil
				}
				return header
			},
			// for montblanc block, invalid preparedSeal "does not" occurs when validating preparedSeal because header is  "not" changed
			nil,
		},
		{
			qbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.PreparedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyPreparedSeals,
		},
		{
			qbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.CommittedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyCommittedSeals,
		},
		{
			montblancBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.PreparedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyPreparedSeals,
		},
		{
			montblancBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.CommittedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyCommittedSeals,
		},
	}

	for i, tc := range testCases {
		headerToTest := tc.headerManipulationFunc(tc.block)
		err := engine.VerifyHeader(chain, headerToTest)
		if !errors.Is(err, tc.expectedError) {
			t.Errorf("error mismatch for case %d: have %v, want %v", i, err, tc.expectedError)
		}
	}
}

func TestVerifyHeaderForSingleBlock(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()

	// qbftcommon.ErrEmptyPrevPreparedSeals case
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	block = updateQBFTBlock(block, engine.Address())
	err := engine.VerifyHeader(chain, block.Header())

	if !(chain.Config().MontBlancBlock.Cmp(block.Number()) < 0) {
		if err != qbftcommon.ErrEmptyPreparedSeals {
			t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrEmptyPreparedSeals)
		}
	} else {
		if err != qbftcommon.ErrEmptyPrevPreparedSeals {
			t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrEmptyPrevPreparedSeals)
		}
	}

	// short extra data
	header := block.Header()
	header.Extra = []byte{}
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidExtraDataFormat {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidExtraDataFormat)
	}
	// incorrect extra format
	header.Extra = []byte("0000000000000000000000000000000012300000000000000000000000000000000000000000000000000000000000000000")
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidExtraDataFormat {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidExtraDataFormat)
	}

	// invalid uncles hash
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.UncleHash = common.BytesToHash([]byte("123456789"))
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidUncleHash {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidUncleHash)
	}

	// invalid difficulty
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Difficulty = big.NewInt(2)
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidDifficulty {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidDifficulty)
	}

	// invalid timestamp
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = chain.Genesis().Time() + (engine.config.GetConfig(block.Number()).BlockPeriod - 1)
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidTimestamp {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidTimestamp)
	}

	// future block
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = uint64(time.Now().Unix() + 10)
	err = engine.VerifyHeader(chain, header)
	if err != consensus.ErrFutureBlock {
		t.Errorf("error mismatch: have %v, want %v", err, consensus.ErrFutureBlock)
	}

	// future block which is within AllowedFutureBlockTime
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = new(big.Int).Add(big.NewInt(time.Now().Unix()), new(big.Int).SetUint64(10)).Uint64()
	priorValue := engine.config.AllowedFutureBlockTime
	engine.config.AllowedFutureBlockTime = 10
	err = engine.VerifyHeader(chain, header)
	engine.config.AllowedFutureBlockTime = priorValue //restore changed value
	if err == consensus.ErrFutureBlock {
		t.Errorf("error mismatch: have %v, want nil", err)
	}
}

func TestVerifyHeaders(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	genesis := chain.Genesis()

	// success case
	headers := []*types.Header{}
	blocks := []*types.Block{}
	size := 100

	for i := 0; i < size; i++ {
		var b *types.Block
		if i == 0 {
			b = makeBlockWithoutSeal(chain, engine, genesis)
			b = updateQBFTBlock(b, engine.Address())
		} else {
			b = makeBlockWithoutSeal(chain, engine, blocks[i-1])
			b = updateQBFTBlock(b, engine.Address())
		}
		blocks = append(blocks, b)
		headers = append(headers, blocks[i].Header())
	}
	// now = func() time.Time {
	// 	return time.Unix(int64(headers[size-1].Time), 0)
	// }
	_, results := engine.VerifyHeaders(chain, headers)
	const timeoutDura = 2 * time.Second
	timeout := time.NewTimer(timeoutDura)
	index := 0
OUT1:
	for {
		select {
		case err := <-results:
			if err != nil {
				if err != qbftcommon.ErrEmptyPrevPreparedSeals && err != qbftcommon.ErrEmptyPreparedSeals && err != qbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					t.Errorf("error mismatch: have %v, want qbftcommon.ErrEmptyCommittedSeals|qbftcommon.ErrInvalidCommittedSeals|ErrUnknownAncestor", err)
					break OUT1
				}
			}
			index++
			if index == size {
				break OUT1
			}
		case <-timeout.C:
			break OUT1
		}
	}
	_, results = engine.VerifyHeaders(chain, headers)
	timeout = time.NewTimer(timeoutDura)
OUT2:
	for {
		select {
		case err := <-results:
			if err != nil {
				if err != qbftcommon.ErrEmptyPrevPreparedSeals && err != qbftcommon.ErrEmptyPreparedSeals && err != qbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					t.Errorf("error mismatch: have %v, want qbftcommon.ErrEmptyCommittedSeals|qbftcommon.ErrInvalidCommittedSeals|ErrUnknownAncestor", err)
					break OUT2
				}
			}
		case <-timeout.C:
			break OUT2
		}
	}
	// error header cases
	headers[2].Number = big.NewInt(100)
	_, results = engine.VerifyHeaders(chain, headers)
	timeout = time.NewTimer(timeoutDura)
	index = 0
	errors := 0
	expectedErrors := 0
OUT3:
	for {
		select {
		case err := <-results:
			if err != nil {
				if err != qbftcommon.ErrEmptyPrevPreparedSeals && err != qbftcommon.ErrEmptyPreparedSeals && err != qbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					errors++
				}
			}
			index++
			if index == size {
				if errors != expectedErrors {
					t.Errorf("error mismatch: have %v, want %v", errors, expectedErrors)
				}
				break OUT3
			}
		case <-timeout.C:
			break OUT3
		}
	}
}

// Test that the next block has the correct previous prepared seals and previous committed seals.
// First, create a block with a seal and commit it.
// Then, when creating the next block, it should have the correct previous prepared seals and previous committed seals.
func TestPrevSeals(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	// Create an insert a new block into the chain.
	block := makeBlock(chain, engine, chain.Genesis())

	blockExtra, err := types.ExtractQBFTExtra(block.Header())
	if err != nil {
		t.Error(err.Error())
	}

	_, err = chain.InsertChain(types.Blocks{block})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	if err = engine.NewChainHead(); err != nil {
		t.Errorf("Error posting NewChainHead Event: %v", err)
	}

	nextBlock := makeBlockWithoutSeal(chain, engine, block)

	nextBlockExtra, err := types.ExtractQBFTExtra(nextBlock.Header())
	if err != nil {
		t.Error(err.Error())
	}

	if len(nextBlockExtra.PrevPreparedSeal) != 1 {
		t.Errorf("prev prepared seals mismatch: have %v, want 1", len(nextBlockExtra.PrevPreparedSeal))
	}

	if !bytes.Equal(blockExtra.PreparedSeal[0], nextBlockExtra.PrevPreparedSeal[0]) {
		t.Errorf("prev prepared seals mismatch: have %v, want %v", nextBlockExtra.PrevPreparedSeal[0], blockExtra.PreparedSeal[0])
	}

	if len(nextBlockExtra.PrevCommittedSeal) != 1 {
		t.Errorf("committed seals mismatch: have %v, want 1", len(nextBlockExtra.PrevCommittedSeal))
	}

	if !bytes.Equal(blockExtra.CommittedSeal[0], nextBlockExtra.PrevCommittedSeal[0]) {
		t.Errorf("committed seals mismatch: have %v, want %v", nextBlockExtra.PrevCommittedSeal[0], blockExtra.CommittedSeal[0])
	}
}

func postMsgEventToBackend(qbftEngine *Backend, message messages.QBFTMessage, payload []byte) error {
	if err := qbftEngine.istanbulEventMux.Post(qbft.MessageEvent{
		Code:    message.Code(),
		Payload: payload,
	}); err != nil {
		return err
	}
	return nil
}

func makeQBFTMessagePayload(message messages.QBFTMessage, node otherNode) ([]byte, error) {
	// set source of the message
	message.SetSource(node.address)
	encodedPayload, err := message.EncodePayloadForSigning()
	if err != nil {
		return nil, err
	}
	hashData := crypto.Keccak256(encodedPayload)
	signature, err := crypto.Sign(hashData, node.privateKey)
	if err != nil {
		return nil, err
	}
	message.SetSignature(signature)
	payload, err := rlp.EncodeToBytes(&message)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func nodeSendPreprepareMsg(qbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	preprepare := messages.NewPreprepare(sequence, round, targetBlock)
	payload, err := makeQBFTMessagePayload(preprepare, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(qbftEngine, preprepare, payload)

	return nil
}

func nodeSendPrepareMsg(qbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	prepareSeal, err := crypto.Sign(qbftcore.PrepareCommittedSeal(targetBlock.Header(), uint32(round.Uint64())), node.privateKey)
	if err != nil {
		return err
	}
	prepare := messages.NewPrepare(sequence, round, targetBlock.Hash(), prepareSeal)
	payload, err := makeQBFTMessagePayload(prepare, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(qbftEngine, prepare, payload)

	return nil
}

func nodeSendCommitMsg(qbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	commitSeal, err := crypto.Sign(qbftcore.PrepareCommittedSeal(targetBlock.Header(), uint32(round.Uint64())), node.privateKey)
	if err != nil {
		return err
	}
	prepare := messages.NewCommit(sequence, round, targetBlock.Hash(), commitSeal)
	payload, err := makeQBFTMessagePayload(prepare, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(qbftEngine, prepare, payload)

	return nil
}

func contains(slice []common.Address, item common.Address) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func TestSimulation(t *testing.T) {
	const (
		PreprepareCode  = 0x12
		PrepareCode     = 0x13
		CommitCode      = 0x14
		RoundChangeCode = 0x15
	)

	for i := 0; i < 48; i++ {
		chain, engine, nodes := newBlockChain(5)
		block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
		currState, _ := chain.State()
		block, _ = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)

		eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{}, qbft.MessageEvent{})

		resultCh := make(chan *types.Block, 10)
		err := engine.Seal(chain, block, resultCh, nil)
		if err != nil {
			t.Errorf("error %v", err)
		}

		ev := <-eventSub.Chan()
		request, ok := ev.Data.(qbft.RequestEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v, want request event", reflect.TypeOf(ev.Data))
		}

		proposedBlock, ok := request.Proposal.(*types.Block)
		if !ok {
			t.Errorf("unexpected proposal comes: %v, want %v", reflect.TypeOf(request.Proposal), request.Proposal)
		}

		quorumSize := engine.Validators(chain.Genesis()).QuorumSize()
		// Preprepare
		ev = <-eventSub.Chan()
		msgEv, ok := ev.Data.(qbft.MessageEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
			return
		}

		if msgEv.Code != PreprepareCode {
			m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
			t.Errorf("unexpected code comes: %v, sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, PreprepareCode, proposedBlock.Number())
			return
		}

		//decode msg
		m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
		t.Logf("Preprepare message comes for block %v, round %v", m.View().Sequence, m.View().Round)

		// Prepare
		totalPrepareMessages := 0
		ev = <-eventSub.Chan()
		msgEv, ok = ev.Data.(qbft.MessageEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
			return
		}

		if msgEv.Code != PrepareCode {
			m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
			t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, PrepareCode, proposedBlock.Number())
			return
		}
		totalPrepareMessages++

		//decode msg
		m, _ = messages.Decode(msgEv.Code, msgEv.Payload)
		t.Logf("Local prepare message comes for block %v, round %v", m.View().Sequence, m.View().Round)

		// send another prepare message
		for _, node := range nodes {
			if totalPrepareMessages >= quorumSize {
				break
			}

			if node.address == engine.Address() {
				continue
			}

			t.Log("sending another prepare message", node.address)
			if err := nodeSendPrepareMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
				t.Errorf("failed to send prepare msg. err :  %v", err)
			}

			ev = <-eventSub.Chan()
			msgEv, ok = ev.Data.(qbft.MessageEvent)
			if !ok {
				t.Errorf("unexpected event comes: %v, want event message", reflect.TypeOf(ev.Data))
				return
			}

			if msgEv.Code != PrepareCode {
				m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
				t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, PrepareCode, proposedBlock.Number())
				return
			}
			totalPrepareMessages++
			t.Logf("another prepare message comes, total count : %d", totalPrepareMessages)
		}

		// Commit
		totalCommitMessages := 0
		ev = <-eventSub.Chan()
		msgEv, ok = ev.Data.(qbft.MessageEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
			return
		}

		if msgEv.Code != CommitCode {
			m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
			t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, CommitCode, proposedBlock.Number())
			return
		}
		totalCommitMessages++

		m, _ = messages.Decode(msgEv.Code, msgEv.Payload)
		t.Logf("Local commit message comes for block %v, round %v", m.View().Sequence, m.View().Round)

		// send another commit message
		for _, node := range nodes {
			if totalCommitMessages >= quorumSize {
				break
			}

			if node.address == engine.Address() {
				continue
			}

			t.Log("sending another commit message", node.address)
			if err := nodeSendCommitMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
				t.Errorf("failed to send commit msg. err :  %v", err)
			}

			ev = <-eventSub.Chan()
			msgEv, ok = ev.Data.(qbft.MessageEvent)
			if !ok {
				t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
				return
			}

			if msgEv.Code != CommitCode {
				m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
				t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, CommitCode, proposedBlock.Number())
				return
			}
			totalCommitMessages++
			t.Logf("another commit message comes, total count : %d", totalCommitMessages)
		}

		// check whether the block is inserted into the chain
		finalBlock1 := <-resultCh
		if finalBlock1.Hash() != proposedBlock.Hash() {
			t.Errorf("hash mismatch: have %v, want %v", finalBlock1.Hash(), proposedBlock.Hash())
		}

		_, err = chain.InsertChain(types.Blocks{finalBlock1})
		if err != nil {
			t.Errorf("Error inserting block: %v", err)
		}

		if err = engine.NewChainHead(); err != nil {
			t.Errorf("Error posting NewChainHead Event: %v", err)
		}

		actualProposer1 := engine.Address()
		expectedProposer1 := engine.GetProposer(1)
		if actualProposer1 != expectedProposer1 {
			t.Errorf("proposer mismatch: have %v, want %v", actualProposer1.Hex(), expectedProposer1.Hex())
		}

		// get last block from the chain and compare with final block
		lastBlock := chain.CurrentBlock()
		if lastBlock.Hash() != finalBlock1.Hash() {
			t.Errorf("hash mismatch: have %v, want %v", lastBlock.Hash(), finalBlock1.Hash())
		}

		for j, node := range nodes {
			if node.address == expectedProposer1 {
				state, err := chain.StateAt(finalBlock1.Root())
				if state == nil || err != nil {
					panic(err)
				}

				balance := state.GetBalance(node.address).ToBig()
				expectedBalance := node.balance

				blockReward := chain.Config().GetBlockReward(finalBlock1.Number())
				expectedBalance = big.NewInt(0).Add(expectedBalance, &blockReward)
				if balance.Cmp(expectedBalance) != 0 {
					t.Errorf("balance mismatch: have %v, want %v", balance, expectedBalance)
				}
				nodes[j].balance = balance
				break
			}
		}

		// wait until the engine is ready to accept the next block
		for {
			if engine.core.GetState() == 0 { // StateAcceptRequest
				break
			}
		}

		quorumSize = engine.Validators(finalBlock1).QuorumSize()

		// make second block
		block = makeBlockWithoutSeal(chain, engine, finalBlock1)
		currState, err = chain.State()
		if err != nil {
			t.Errorf("failed to get current state. err :  %v", err)
			return
		}
		block, err = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)
		if err != nil {
			t.Errorf("failed to finalize and assemble. err :  %v", err)
			return
		}

		var actualProposer2 common.Address
		// send preprepare message
		for _, node := range nodes {
			if engine.Validators(finalBlock1).IsProposer(node.address) {
				actualProposer2 = node.address

				header := block.Header()
				header.Coinbase = node.address
				currState, err = chain.State()
				if err != nil {
					t.Errorf("failed to get current state. err :  %v", err)
					return
				}
				engine.Finalize(chain, header, currState, nil, nil, nil)
				proposedBlock = block.WithSeal(header)

				t.Log("sending preprepare message", node.address)
				if err := nodeSendPreprepareMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("failed to send preprepare msg. err :  %v", err)
				}
				break
			}
		}

		// Preprepare
		ev = <-eventSub.Chan()
		msgEv, ok = ev.Data.(qbft.MessageEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
			return
		}

		if msgEv.Code != PreprepareCode {
			m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
			t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, PreprepareCode, proposedBlock.Number())
			return
		}

		//decode msg
		m, _ = messages.Decode(msgEv.Code, msgEv.Payload)
		t.Logf("Preprepare message comes for block %v, round %v", m.View().Sequence, m.View().Round)

		// Prepare
		totalPrepareMessages = 0
		ev = <-eventSub.Chan()
		msgEv, ok = ev.Data.(qbft.MessageEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
			return
		}

		if msgEv.Code != PrepareCode {
			m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
			t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, PrepareCode, proposedBlock.Number())
			return
		}
		totalPrepareMessages++

		//decode msg
		m, _ = messages.Decode(msgEv.Code, msgEv.Payload)
		t.Logf("Local prepare message comes for block %v, round %v", m.View().Sequence, m.View().Round)

		// send another prepare message
		for _, node := range nodes {
			if totalPrepareMessages >= quorumSize {
				break
			}

			if node.address == engine.Address() {
				continue
			}

			t.Log("sending another prepare message", node.address)
			if err := nodeSendPrepareMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
				t.Errorf("failed to send prepare msg. err :  %v", err)
			}

			ev = <-eventSub.Chan()
			msgEv, ok = ev.Data.(qbft.MessageEvent)
			if !ok {
				t.Errorf("unexpected event comes: %v, want event message", reflect.TypeOf(ev.Data))
				return
			}

			if msgEv.Code != PrepareCode {
				m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
				t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, PrepareCode, proposedBlock.Number())
				return
			}
			totalPrepareMessages++
			t.Logf("another prepare message comes, total count : %d", totalPrepareMessages)
		}

		// Commit
		totalCommitMessages = 0
		ev = <-eventSub.Chan()
		msgEv, ok = ev.Data.(qbft.MessageEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
			return
		}

		if msgEv.Code != CommitCode {
			m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
			t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, CommitCode, proposedBlock.Number())
			return
		}
		totalCommitMessages++

		m, _ = messages.Decode(msgEv.Code, msgEv.Payload)
		t.Logf("Local commit message comes for block %v, round %v", m.View().Sequence, m.View().Round)

		// send another commit message
		for _, node := range nodes {
			if totalCommitMessages >= quorumSize {
				break
			}

			if node.address == engine.Address() {
				continue
			}

			t.Log("sending another commit message", node.address)
			if err := nodeSendCommitMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
				t.Errorf("failed to send commit msg. err :  %v", err)
			}

			ev = <-eventSub.Chan()
			msgEv, ok = ev.Data.(qbft.MessageEvent)
			if !ok {
				t.Errorf("unexpected event comes: %v, want message event", reflect.TypeOf(ev.Data))
				return
			}

			if msgEv.Code != CommitCode {
				m, _ := messages.Decode(msgEv.Code, msgEv.Payload)
				t.Errorf("unexpected code comes: %v , sequence %v, want %v, sequence %v", msgEv.Code, m.View().Sequence, CommitCode, proposedBlock.Number())
				return
			}
			totalCommitMessages++
			t.Logf("another commit message comes, total count : %d", totalCommitMessages)
		}

		finalBlock2 := <-blockEnqueueChannel
		if finalBlock2.Hash() != proposedBlock.Hash() {
			t.Errorf("hash mismatch: have %v, want %v", finalBlock2.Hash(), proposedBlock.Hash())
		}
		_, err = chain.InsertChain(types.Blocks{finalBlock2})
		if err != nil {
			t.Errorf("Error inserting block: %v", err)
		}

		if err = engine.NewChainHead(); err != nil {
			t.Errorf("Error posting NewChainHead Event: %v", err)
		}

		expectedProposer2 := engine.GetProposer(2)
		if actualProposer2 != expectedProposer2 {
			t.Errorf("proposer mismatch: have %v, want %v", actualProposer2.Hex(), expectedProposer2.Hex())
		}

		// get last block from the chain and compare with final block
		lastBlock = chain.CurrentBlock()
		if lastBlock.Hash() != finalBlock2.Hash() {
			t.Errorf("hash mismatch: have %v, want %v", lastBlock.Hash(), finalBlock2.Hash())
		}

		blockExtra, err := types.ExtractQBFTExtra(finalBlock2.Header())
		if err != nil {
			t.Error(err.Error())
		}

		h := types.CopyHeader(finalBlock1.Header())
		proposalSeal := h.QBFTHashWithRoundNumber(0).Bytes()

		var prepareRewardees []common.Address
		var commitRewardees []common.Address

		// get prev prepared address
		for _, seal := range blockExtra.PrevPreparedSeal {
			addr, err := qbft.GetSignatureAddressNoHashing(proposalSeal, seal)
			if err != nil {
				t.Errorf("failed to get signature address. err :  %v", err)
			}
			prepareRewardees = append(prepareRewardees, addr)
		}

		// get prev committed address
		for _, seal := range blockExtra.PrevCommittedSeal {
			addr, err := qbft.GetSignatureAddressNoHashing(proposalSeal, seal)
			if err != nil {
				t.Errorf("failed to get signature address. err :  %v", err)
			}
			commitRewardees = append(commitRewardees, addr)
		}

		state, err := chain.StateAt(finalBlock2.Root())
		if state == nil || err != nil {
			panic(err)
		}

		for _, node := range nodes {
			balance := state.GetBalance(node.address).ToBig()
			expectedBalance := node.balance

			if node.address == expectedProposer2 {
				blockReward := chain.Config().GetBlockReward(finalBlock2.Number())
				expectedBalance = big.NewInt(0).Add(expectedBalance, &blockReward)
			}

			// check prepare reward
			if contains(prepareRewardees, node.address) {
				prepareReward := chain.Config().GetPrepareReward(finalBlock2.Number())
				expectedBalance = big.NewInt(0).Add(expectedBalance, &prepareReward)
			}

			// check commit reward
			if contains(commitRewardees, node.address) {
				commitReward := chain.Config().GetCommitReward(finalBlock2.Number())
				expectedBalance = big.NewInt(0).Add(expectedBalance, &commitReward)
			}

			if balance.Cmp(expectedBalance) != 0 {
				t.Errorf("balance mismatch: have %v, want %v", balance, expectedBalance)
			}
		}
		engine.Stop()
		eventSub.Unsubscribe()
	}
}

func makeBlockThroughConsensus(chain *core.BlockChain, engine *Backend, nodes []otherNode, parentBlock *types.Block) (*types.Block, error) {
	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})
	defer eventSub.Unsubscribe()

	block := makeBlockWithoutSeal(chain, engine, parentBlock)
	currState, _ := chain.State()
	block, _ = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)
	resultCh := make(chan *types.Block, 10)
	stopCh := make(chan struct{})
	go func() {
		engine.Seal(chain, block, resultCh, stopCh)
	}()

	ev := <-eventSub.Chan()
	request, ok := ev.Data.(qbft.RequestEvent)
	if !ok {
		return nil, fmt.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}

	proposedBlock, _ := request.Proposal.(*types.Block)
	for _, node := range nodes {
		go func(node otherNode) error {
			ticker := time.NewTicker(300 * time.Millisecond)
			defer ticker.Stop()

			executed := make(map[string]bool)
			for {
				<-ticker.C
				consensusState := engine.core.GetState().String()
				if executed[consensusState] {
					continue
				}

				switch consensusState {
				case "Accept request":
					if engine.Validators(parentBlock).IsProposer(node.address) {
						header := proposedBlock.Header()
						header.Coinbase = node.address
						statedb, err := chain.State()
						if err != nil {
							return fmt.Errorf("failed to get statedb err %v", err)
						}
						engine.Finalize(chain, header, statedb, nil, nil, nil)
						proposedBlock = proposedBlock.WithSeal(header)
						if err := nodeSendPreprepareMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
							return fmt.Errorf("failed to send preprepare msg. err :  %v", err)
						}
						executed[consensusState] = true
					} else {
						executed[consensusState] = true
					}
				case "Preprepared":
					if err := nodeSendPrepareMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
						return fmt.Errorf("failed to send prepare msg. err :  %v", err)
					}
					executed[consensusState] = true
				case "Prepared":
					if err := nodeSendCommitMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
						return fmt.Errorf("failed to send commit msg. err :  %v", err)
					}
					executed[consensusState] = true
				}
			}
		}(node)
	}

	select {
	case blockFromResultCh := <-resultCh:
		return blockFromResultCh, nil
	case blockFromEnquequeCh := <-blockEnqueueChannel:
		stopCh <- struct{}{}
		return blockFromEnquequeCh, nil
	}
}

func TestMakingBlock(t *testing.T) {
	chain, engine, nodes := newBlockChain(4)
	parentBlock := chain.Genesis()

	for i := 0; i < 3; i++ {
		finalBlock, err := makeBlockThroughConsensus(chain, engine, nodes, parentBlock)
		if err != nil {
			t.Errorf("failed to make block1 through consensus. err %v", err)
		}

		if parentBlock.Hash() != finalBlock.ParentHash() {
			t.Errorf("parent hash mismatch: have %v, want %v", finalBlock.ParentHash(), parentBlock.Hash())
		}

		if _, err := chain.InsertChain(types.Blocks{finalBlock}); err != nil {
			t.Errorf("failed to insert block1. err %v", err)
		}

		// check if generated final block is included in chain properly
		block := chain.GetBlockByHash(finalBlock.Hash())
		if block == nil {
			t.Errorf("block number %v is not generated correctly", finalBlock.Hash())
		}
		parentBlock = finalBlock
	}
}

func TestLackingSealsFromPropagatedBlock(t *testing.T) {
	chain, engine, nodes := newBlockChain(4)
	// 1. generate block through consensus
	validBlock, err := makeBlockThroughConsensus(chain, engine, nodes, chain.Genesis())
	if err != nil {
		t.Errorf("failed to make valid block through consensus. err %v", err)
	}

	// 2. remove some preparedSeals from valid block ( len(preparedSeal) < 2F+1 ) and make malformed block
	header := validBlock.Header()
	qbftExtra, _ := types.ExtractQBFTExtra(header)

	qbftExtra.PreparedSeal = qbftExtra.PreparedSeal[2:]
	payload, err := rlp.EncodeToBytes(qbftExtra)
	if err != nil {
		t.Errorf("failed to encode qbftExtra. err %v", err)
	}
	header.Extra = payload
	malformedBlock := validBlock.WithSeal(header)

	// 3. test the case when malformed block with lack of preparedSeal is propagated
	_, err = chain.InsertChain(types.Blocks{malformedBlock})
	if !errors.Is(err, qbftcommon.ErrInvalidPreparedSeals) {
		t.Errorf("unexpected error. expect %v, got %v", qbftcommon.ErrInvalidPreparedSeals, err)
	}
}
