package collector

import (
	"container/ring"
	"github.com/ethereum/go-ethereum/byzantine/attack"
	"sync"
)

// Import types from attack package (would be properly imported in real implementation)
type (
	DataRequirement = attack.DataRequirement
	DataFilter      = attack.DataFilter
	MessageCode     = attack.MessageCode
)

// EventCollector interface for collecting QBFT events
type EventCollector interface {
	StartCollection(dataReqs []DataRequirement) error
	StopCollection(attackID string) error
	GetHistoricalData(filter DataFilter) []CollectedData
	GetStateHistory(sequence uint64) []StateTransition
}

// messageStore stores messages with a size limit
type messageStore struct {
	messages *ring.Ring
	mu       sync.RWMutex
}

func newMessageStore(maxSize int) *messageStore {
	return &messageStore{
		messages: ring.New(maxSize),
	}
}

func (s *messageStore) add(msg *QBFTMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages.Value = msg
	s.messages = s.messages.Next()
}

func (s *messageStore) getFiltered(filter DataFilter) []*QBFTMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*QBFTMessage

	s.messages.Do(func(v interface{}) {
		if v == nil {
			return
		}

		msg := v.(*QBFTMessage)

		// Apply filters
		if filter.FromSequence > 0 && msg.Sequence < filter.FromSequence {
			return
		}
		if filter.ToSequence > 0 && msg.Sequence > filter.ToSequence {
			return
		}

		// Filter by message type
		if len(filter.MessageTypes) > 0 {
			found := false
			for _, msgType := range filter.MessageTypes {
				if msg.Code == msgType {
					found = true
					break
				}
			}
			if !found {
				return
			}
		}

		// Filter by validator
		if len(filter.Validators) > 0 {
			found := false
			for _, val := range filter.Validators {
				if msg.Address == val {
					found = true
					break
				}
			}
			if !found {
				return
			}
		}

		result = append(result, msg)
	})

	return result
}

// stateStore stores state transitions
type stateStore struct {
	transitions map[uint64][]StateTransition
	mu          sync.RWMutex
}

func newStateStore() *stateStore {
	return &stateStore{
		transitions: make(map[uint64][]StateTransition),
	}
}

func (s *stateStore) add(transition StateTransition) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.transitions[transition.Sequence] = append(s.transitions[transition.Sequence], transition)
}

func (s *stateStore) get(sequence uint64) []StateTransition {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.transitions[sequence]
}

// eventCollector implements EventCollector interface
type eventCollector struct {
	source       QBFTEventSource
	messageStore *messageStore
	stateStore   *stateStore
	eventChan    chan QBFTEvent
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex

	dataReqs []DataRequirement
}

// NewEventCollector creates a new event collector
func NewEventCollector(source QBFTEventSource) EventCollector {
	return &eventCollector{
		source:     source,
		stateStore: newStateStore(),
		eventChan:  make(chan QBFTEvent, 1000),
		stopChan:   make(chan struct{}),
	}
}

// StartCollection starts collecting events based on data requirements
func (c *eventCollector) StartCollection(dataReqs []DataRequirement) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store data requirements
	c.dataReqs = dataReqs

	// Initialize message store based on requirements
	maxMessages := 100 // default
	for _, req := range dataReqs {
		if req.Type == "messages" && req.MaxRecords > 0 {
			maxMessages = req.MaxRecords
			break
		}
	}
	c.messageStore = newMessageStore(maxMessages)

	// Subscribe to events
	if err := c.source.Subscribe(c.eventChan); err != nil {
		return err
	}

	// Start event processing
	c.wg.Add(1)
	go c.processEvents()

	return nil
}

// StopCollection stops collecting events
func (c *eventCollector) StopCollection(attackID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Signal stop
	close(c.stopChan)

	// Unsubscribe from events
	if err := c.source.Unsubscribe(c.eventChan); err != nil {
		return err
	}

	// Wait for processing to complete
	c.wg.Wait()

	return nil
}

// processEvents processes incoming events
func (c *eventCollector) processEvents() {
	defer c.wg.Done()

	for {
		select {
		case event := <-c.eventChan:
			c.handleEvent(event)
		case <-c.stopChan:
			return
		}
	}
}

// handleEvent handles a single event
func (c *eventCollector) handleEvent(event QBFTEvent) {
	switch event.Type {
	case EventTypeMessage:
		if event.Message != nil {
			// Check if we should collect this message
			for _, req := range c.dataReqs {
				if req.Type == "messages" {
					// Apply filter if specified
					if c.matchesFilter(event.Message, req.Filter) {
						c.messageStore.add(event.Message)
					}
					break
				}
			}
		}

	case EventTypeStateChange:
		// Check if we should collect state changes
		for _, req := range c.dataReqs {
			if req.Type == "state" {
				transition := StateTransition{
					Sequence: event.Sequence,
					Round:    event.Round,
					OldState: event.OldState,
					NewState: event.NewState,
				}
				c.stateStore.add(transition)
				break
			}
		}
	}
}

// matchesFilter checks if a message matches the filter criteria
func (c *eventCollector) matchesFilter(msg *QBFTMessage, filter DataFilter) bool {
	// Check sequence range
	if filter.FromSequence > 0 && msg.Sequence < filter.FromSequence {
		return false
	}
	if filter.ToSequence > 0 && msg.Sequence > filter.ToSequence {
		return false
	}

	// Check message types
	if len(filter.MessageTypes) > 0 {
		found := false
		for _, msgType := range filter.MessageTypes {
			if msg.Code == msgType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check validators
	if len(filter.Validators) > 0 {
		found := false
		for _, val := range filter.Validators {
			if msg.Address == val {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GetHistoricalData retrieves historical data based on filter
func (c *eventCollector) GetHistoricalData(filter DataFilter) []CollectedData {
	messages := c.messageStore.getFiltered(filter)

	result := make([]CollectedData, 0, len(messages))
	for _, msg := range messages {
		result = append(result, CollectedData{
			Type:     "message",
			Sequence: msg.Sequence,
			Round:    msg.Round,
			Data:     msg,
		})
	}

	return result
}

// GetStateHistory retrieves state transition history for a sequence
func (c *eventCollector) GetStateHistory(sequence uint64) []StateTransition {
	return c.stateStore.get(sequence)
}
