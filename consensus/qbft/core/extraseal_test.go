package core

import (
	"crypto/ecdsa"
	"errors"
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
	signer        *ecdsa.PrivateKey
	signerAddress common.Address
)

func init() {
	_, nodeKeys := testutils.GenesisAndKeys(1)
	signer = nodeKeys[0]
	signerAddress = crypto.PubkeyToAddress(signer.PublicKey)
}

func TestToPriority(t *testing.T) {
	type testMessage struct {
		message       messages.QBFTMessage
		expectedIndex int
	}

	queue := prque.New[int64, testMessage](nil)
	header := types.Header{}
	digest := []byte("abcd")
	testMessages := []testMessage{
		{
			createPrepareMsg(&header, common.Big2, common.Big0, digest),
			3,
		},
		{
			createPrepareMsg(&header, common.Big1, common.Big3, digest),
			4,
		},
		{
			createPrepareMsg(&header, common.Big1, common.Big1, digest),
			5,
		},
		{
			createCommitMsg(&header, common.Big2, common.Big0, digest),
			2,
		},
		{
			createPrepareMsg(&header, common.Big3, common.Big3, digest),
			0,
		},
		{
			createCommitMsg(&header, common.Big3, common.Big2, digest),
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

// makeCoreForTest returns core object with empty backend.
// Its purpose is to test pure qbft/core functions.
// not recommended for testing logic that includes qbft/backend function
func makeCoreForTest(priorRound, currentRound, currentSequence *big.Int, lastProposal *types.Block) *Core {
	// set core with empty backend.
	// current state is StateAcceptRequest.
	core := &Core{
		config:             nil,
		address:            signerAddress,
		state:              StateAcceptRequest,
		handlerWg:          new(sync.WaitGroup),
		logger:             log.New("address", signerAddress),
		backend:            nil,
		backlogs:           make(map[common.Address]*prque.Prque[int64, messages.QBFTMessage]),
		backlogsMu:         new(sync.Mutex),
		extraSeals:         prque.New[int64, messages.QBFTMessage](nil),
		extraSealsMu:       new(sync.Mutex),
		pendingRequests:    prque.New[int64, *Request](nil),
		pendingRequestsMu:  new(sync.Mutex),
		consensusTimestamp: time.Time{},
		priorState:         priorState{new(sync.RWMutex), common.Big0, nil},
	}
	core.validateFn = core.checkValidatorSignature
	// Set core current view and proposal
	core.current = newRoundState(
		&qbft.View{
			Round:    currentRound,
			Sequence: currentSequence,
		}, validator.NewSet([]common.Address{
			common.BytesToAddress([]byte("1234567894")),
			common.BytesToAddress([]byte("1234567895")),
		}, qbft.NewRoundRobinProposerPolicy()),
		nil, nil, nil, nil, func(hash common.Hash) bool { return false })
	core.updatePriorState(priorRound, lastProposal)
	return core
}

// TestAddToExtraSeal is to test core.addToExtraSeal function
// test case is when core.state is StateAcceptedRequest
func TestAddToExtraSeal(t *testing.T) {
	// make proposals
	lastProposal := makeProposal(common.Big2)
	invalidLastProposal := makeProposal(common.Big3)

	// make core instance  with empty backend
	core := makeCoreForTest(common.Big2, common.Big0, common.Big3, lastProposal)

	type testMessage struct {
		message       messages.QBFTMessage
		expectedError error
	}

	malformedSeal := func(commitMsg *messages.Commit) *messages.Commit {
		seal := []byte("malformedSeal")
		commitMsg.CommitSeal = seal
		return commitMsg
	}

	testExtraSealMessages := []testMessage{
		{
			createPrepareMsg(lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes()),
			nil,
		},
		{
			createCommitMsg(lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes()),
			nil,
		},
		{
			createPreprepareMsg(common.Big2, common.Big2, lastProposal),
			errInvalidExtraSealMessage,
		},
		{
			// fail to verify digest -  message's proposal is not same with currentProposal
			createPrepareMsg(invalidLastProposal.Header(), common.Big2, common.Big2, invalidLastProposal.Hash().Bytes()),
			errInvalidMessage,
		},
		{
			// fail to verify seal -  seal doesn't match with message
			malformedSeal(createCommitMsg(lastProposal.Header(), common.Big2, common.Big1, lastProposal.Hash().Bytes())),
			errInvalidSeal,
		},
	}

	for _, tm := range testExtraSealMessages {
		err := core.addToExtraSeal(tm.message)
		if !errors.Is(err, tm.expectedError) {
			t.Errorf("unexpected error adding to extraSeal. want %v, have %v", tm.expectedError, err)
		}
	}
}

// TestProcessExtraSeal is to test core.ProcessExtraSeal
// 1. if mesasge have lower view than LatestView, drop it
// 2. if wrong digest, drop it
// 3. message other than prepare/commit, drop it
func TestProcessExtraSeal(t *testing.T) {
	// make proposals
	lastProposal := makeProposal(common.Big2)

	// make core instance  with empty backend
	core := makeCoreForTest(common.Big2, common.Big0, common.Big3, lastProposal)

	// assume situation when consensus enters new round and preparing for new block.
	// set core.current.proposal. Proposal's block number should be current.Sequence -1
	core.current.SetPreprepare(createPreprepareMsg(common.Big2, common.Big2, lastProposal))

	testExtraSealMessages := []messages.QBFTMessage{
		// 1. 2 valid extra seal messages
		createPrepareMsg(lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes()),
		createCommitMsg(lastProposal.Header(), common.Big2, common.Big2, lastProposal.Hash().Bytes()),
		// 2. 3 extra seal messages with lower view than latestView
		createPrepareMsg(lastProposal.Header(), common.Big2, common.Big1, lastProposal.Hash().Bytes()),
		createCommitMsg(lastProposal.Header(), common.Big1, common.Big2, lastProposal.Hash().Bytes()),
		createCommitMsg(lastProposal.Header(), common.Big1, common.Big2, lastProposal.Hash().Bytes()),
	}

	for _, tm := range testExtraSealMessages {
		if err := core.addToExtraSeal(tm); err != nil {
			t.Errorf("error adding to core.extraSeals : %v", err)
		}
	}

	preparedSeal, committedSeal := core.ProcessExtraSeal(lastProposal, common.Big2)
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

func createPrepareMsg(header *types.Header, sequence, round *big.Int, digest []byte) *messages.Prepare {
	seal, _ := crypto.Sign(PrepareSeal(header, uint32(round.Uint64()), SealTypePrepare), signer)
	prepare := messages.NewPrepare(sequence, round, common.BytesToHash(digest), seal)
	prepare.SetSource(signerAddress)
	return prepare
}

func createCommitMsg(header *types.Header, sequence, round *big.Int, digest []byte) *messages.Commit {
	seal, _ := crypto.Sign(PrepareSeal(header, uint32(round.Uint64()), SealTypeCommit), signer)
	commit := messages.NewCommit(sequence, round, common.BytesToHash(digest), seal)
	commit.SetSource(signerAddress)
	return commit
}

func makeProposal(blockNumber *big.Int) *types.Block {
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
