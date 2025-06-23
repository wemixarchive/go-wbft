package types

// ConsensusState defines QBFT consensus states
type ConsensusState int

const (
	StateIdle ConsensusState = iota
	StatePreprepared
	StatePrepared
	StateCommitted
	StateRoundChange
)

// CollectedData represents data collected by the EventCollector
type CollectedData struct {
	Type     string
	Sequence uint64
	Round    uint64
	Data     interface{}
}

// StateTransition represents a consensus state transition
type StateTransition struct {
	Sequence uint64
	Round    uint64
	OldState ConsensusState
	NewState ConsensusState
}
