package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/consensus/qbft/messages"
)

func TestToPriority(t *testing.T) {
	type testMessage struct {
		message       messages.QBFTMessage
		expectedIndex int
	}

	queue := prque.New[int64, testMessage](nil)

	testMessages := []testMessage{
		{
			createPrepareMsg(common.Big2, common.Big0),
			3,
		},
		{
			createPrepareMsg(common.Big1, common.Big3),
			4,
		},
		{
			createPrepareMsg(common.Big1, common.Big1),
			5,
		},
		{
			createCommitMsg(common.Big2, common.Big0),
			2,
		},
		{
			createPrepareMsg(common.Big3, common.Big3),
			0,
		},
		{
			createCommitMsg(common.Big3, common.Big2),
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

func TestAddingExtraSeals(t *testing.T) {
	// adding preparedSeal when prepared state - success case, reject case
	// adding committeSeal when committed state - success case, reject case
	// adding prepared/committedSeal when AcceptRequest state - success case, reject case
}

func TestProcessExtraSeal(t *testing.T) {
	// ProcessExtraSeal 동작 테스트
	// 1. latestView 보다 이전 View 가진 메세지 버림 처리 테스트
	// 2. digest 안맞으면 버림 처리 테스트
	// 3. prepare/commit 메세지가 아닌 메세지가 들어올 경우 테스트
}

func createPrepareMsg(sequence, round *big.Int) *messages.Prepare {
	return &messages.Prepare{
		CommonPayload: messages.CommonPayload{
			Sequence: sequence,
			Round:    round,
		},
	}
}

func createCommitMsg(sequence, round *big.Int) *messages.Prepare {
	return &messages.Prepare{
		CommonPayload: messages.CommonPayload{
			Sequence: sequence,
			Round:    round,
		},
	}
}
