package byzantine

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type ByzantineHooks interface {
	// Proposal 단계
	BeforeCreateProposal(proposal *types.Block) (*types.Block, error)
	AfterReceiveProposal(proposal *types.Block) error

	// Prepare 단계
	//BeforeSendPrepare(prepare *PrepareMessage) (*PrepareMessage, error)
	//AfterReceivePrepare(prepare *PrepareMessage) error
	//
	//// Commit 단계
	//BeforeSendCommit(commit *CommitMessage) (*CommitMessage, error)
	//AfterReceiveCommit(commit *CommitMessage) error

	// RoundChange 단계
	OnRoundChange(round uint64) error

	// Block propagation
	BeforeBlockBroadcast(block *types.Block) (*types.Block, error)
}
