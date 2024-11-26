package core

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	"github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/consensus/qbft/testutils"
	"github.com/ethereum/go-ethereum/consensus/qbft/validator"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	testAddress  = "70524d664ffe731100208a0154e556f9bb679ae6"
	testAddress2 = "b37866a925bccd69cfa98d43b510f1d23d78a851"
)

func TestToPriority(t *testing.T) {
	type testMessage struct {
		message       messages.QBFTMessage
		expectedIndex int
	}

	queue := prque.New[int64, testMessage](nil)
	digest := []byte("abcd")
	testMessages := []testMessage{
		{
			createPrepareMsg(common.Big2, common.Big0, digest),
			3,
		},
		{
			createPrepareMsg(common.Big1, common.Big3, digest),
			4,
		},
		{
			createPrepareMsg(common.Big1, common.Big1, digest),
			5,
		},
		{
			createCommitMsg(common.Big2, common.Big0, digest),
			2,
		},
		{
			createPrepareMsg(common.Big3, common.Big3, digest),
			0,
		},
		{
			createCommitMsg(common.Big3, common.Big2, digest),
			1,
		},
	}

	// insert the test messages to priority queue
	for _, tm := range testMessages {
		view := tm.message.View()
		queue.Push(tm, toPriority(&view))
	}

	// check if test messages have index as expected
	idx := 0
	for !queue.Empty() {
		tm, _ := queue.Pop()
		if tm.expectedIndex != idx {
			t.Errorf("unexpected index of message. have %d, want %d", idx, tm.expectedIndex)
		}
		idx++
	}
}

// TestProcessExtraSeal is to test core.ProcessExtraSeal
// 1. if mesasge have lower view than LatestView, drop it
// 2. if wrong digest, drop it
// 3. message other than prepare/commit, drop it
func TestProcessExtraSeal(t *testing.T) {
	_, nodeKeys := testutils.GenesisAndKeys(1)
	address := crypto.PubkeyToAddress(nodeKeys[0].PublicKey)

	// Make core instance  with empty backend
	core := &Core{
		config:             nil,
		address:            address,
		state:              StateAcceptRequest,
		handlerWg:          new(sync.WaitGroup),
		logger:             log.New("address", address),
		backend:            nil,
		backlogs:           make(map[common.Address]*prque.Prque[int64, messages.QBFTMessage]),
		backlogsMu:         new(sync.Mutex),
		extraSeals:         prque.New[int64, messages.QBFTMessage](nil),
		extraSealsMu:       new(sync.Mutex),
		pendingRequests:    prque.New[int64, *Request](nil),
		pendingRequestsMu:  new(sync.Mutex),
		consensusTimestamp: time.Time{},
	}
	core.validateFn = core.checkValidatorSignature

	// Set core current view and state as seqeunce 3, round 0, state AcceptRequest
	core.current = newRoundState(
		&qbft.View{
			Round:    common.Big0,
			Sequence: common.Big3,
		}, validator.NewSet([]common.Address{
			common.BytesToAddress([]byte("1234567894")),
			common.BytesToAddress([]byte("1234567895")),
		}, qbft.NewRoundRobinProposerPolicy()),
		nil, nil, nil, nil, func(hash common.Hash) bool { return false })

	core.priorRound = common.Big2

	// make lastproposal
	lastProposal := makeLastProposal(common.Big2)
	invalidLastProposla := makeLastProposal(common.Big3)

	testExtraSealMessages := []messages.QBFTMessage{
		// 1. 2 valid extra seal messages
		createPrepareMsg(common.Big2, common.Big2, lastProposal.Hash().Bytes()),
		createCommitMsg(common.Big2, common.Big2, lastProposal.Hash().Bytes()),
		// 2. 3 extra seal messages with lower view than latestView
		createPrepareMsg(common.Big2, common.Big1, lastProposal.Hash().Bytes()),
		createCommitMsg(common.Big1, common.Big2, lastProposal.Hash().Bytes()),
		createCommitMsg(common.Big1, common.Big2, lastProposal.Hash().Bytes()),
		// 3. 1 extra seal message with invalid digest, with latestView
		createCommitMsg(common.Big2, common.Big2, invalidLastProposla.Hash().Bytes()),
		// 4. preprepare messag
		createPreprepareMsg(common.Big2, common.Big2, lastProposal),
	}

	for _, tm := range testExtraSealMessages {
		core.addToExtraSeal(tm)
	}

	preparedSeal, committedSeal := core.ProcessExtraSeal(lastProposal)
	if len(preparedSeal) != 1 {
		t.Errorf("unexpected length of preparedSeal. want %d, have %d", 1, len(preparedSeal))
	}
	if len(committedSeal) != 1 {
		t.Errorf("unexpected length of committedSeal. want %d, have %d", 1, len(committedSeal))
	}
	if core.ExtraSealsLen() != 0 {
		t.Errorf("core.extraSeals should be empty after processing")
	}

}

func createPreprepareMsg(sequence, round *big.Int, proposal qbft.Proposal) *messages.Preprepare {
	return messages.NewPreprepare(sequence, round, proposal)
}

func createPrepareMsg(sequence, round *big.Int, digest []byte) *messages.Prepare {
	return messages.NewPrepare(sequence, round, common.BytesToHash(digest), []byte("prepareSeal"))
}

func createCommitMsg(sequence, round *big.Int, digest []byte) *messages.Commit {
	return messages.NewCommit(sequence, round, common.BytesToHash(digest), []byte("commitSeal"))
}

func makeLastProposal(blockNumber *big.Int) *types.Block {
	header := &types.Header{
		ParentHash: common.BytesToHash([]byte("parentBlockHash")),
		Number:     blockNumber,
		GasLimit:   0,
		GasUsed:    0,
		Time:       uint64(12345678),
		Difficulty: types.QBFTDefaultDifficulty,
	}
	block := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	return block
}
