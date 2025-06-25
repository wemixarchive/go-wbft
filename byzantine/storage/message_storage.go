package storage

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// InMemoryMessageStorage implements MessageStorage interface using in-memory storage
type InMemoryMessageStorage struct {
	mu          sync.RWMutex
	messages    map[common.Hash]*types.StoredMessage
	index       map[uint64]map[uint64][]common.Hash // sequence -> round -> hashes
	config      types.StorageConfig
	pruneCancel context.CancelFunc
}

// NewInMemoryMessageStorage creates a new in-memory message storage
func NewInMemoryMessageStorage(config types.StorageConfig) *InMemoryMessageStorage {
	storage := &InMemoryMessageStorage{
		messages: make(map[common.Hash]*types.StoredMessage),
		index:    make(map[uint64]map[uint64][]common.Hash),
		config:   config,
	}

	// Start pruning goroutine
	ctx, cancel := context.WithCancel(context.Background())
	storage.pruneCancel = cancel
	go storage.startPruning(ctx)

	return storage
}

// Store stores a message
func (s *InMemoryMessageStorage) Store(message *types.StoredMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := message.Message.Hash
	s.messages[hash] = message

	// Update index
	sequence := message.Message.Sequence
	round := message.Message.Round

	if s.index[sequence] == nil {
		s.index[sequence] = make(map[uint64][]common.Hash)
	}
	s.index[sequence][round] = append(s.index[sequence][round], hash)

	return nil
}

// GetByHash retrieves a message by hash
func (s *InMemoryMessageStorage) GetByHash(hash common.Hash) (*types.StoredMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	message, exists := s.messages[hash]
	if !exists {
		return nil, ErrMessageNotFound
	}

	return message, nil
}

// GetBySequenceRound retrieves messages by sequence and round
func (s *InMemoryMessageStorage) GetBySequenceRound(sequence, round uint64) ([]*types.StoredMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hashes, exists := s.index[sequence][round]
	if !exists {
		return nil, nil
	}

	messages := make([]*types.StoredMessage, 0, len(hashes))
	for _, hash := range hashes {
		if msg, exists := s.messages[hash]; exists {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// GetRecentMessages retrieves recent messages
func (s *InMemoryMessageStorage) GetRecentMessages(limit int) ([]*types.StoredMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make([]*types.StoredMessage, 0, limit)

	// Sort by received time and get recent ones
	// This is a simplified implementation
	for _, msg := range s.messages {
		messages = append(messages, msg)
		if len(messages) >= limit {
			break
		}
	}

	return messages, nil
}

// Prune removes old messages
func (s *InMemoryMessageStorage) Prune(before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for hash, msg := range s.messages {
		if msg.ReceivedAt.Before(before) {
			delete(s.messages, hash)

			// Remove from index
			sequence := msg.Message.Sequence
			round := msg.Message.Round
			if rounds, exists := s.index[sequence]; exists {
				if hashes, exists := rounds[round]; exists {
					// Remove hash from slice
					for i, h := range hashes {
						if h == hash {
							rounds[round] = append(hashes[:i], hashes[i+1:]...)
							break
						}
					}
					if len(rounds[round]) == 0 {
						delete(rounds, round)
					}
				}
				if len(rounds) == 0 {
					delete(s.index, sequence)
				}
			}
		}
	}

	return nil
}

// startPruning runs periodic pruning
func (s *InMemoryMessageStorage) startPruning(ctx context.Context) {
	if s.config.PruneInterval <= 0 {
		log.Warn("Invalid prune interval, using default 1 hour",
			"configured", s.config.PruneInterval)
		s.config.PruneInterval = time.Hour
	}

	log.Info("Starting message pruning",
		"interval", s.config.PruneInterval,
		"retention", s.config.MessageRetention)

	ticker := time.NewTicker(s.config.PruneInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pruneTime := time.Now().Add(-s.config.MessageRetention)
			_ = s.Prune(pruneTime)
		}
	}
}

// Close closes the storage
func (s *InMemoryMessageStorage) Close() error {
	if s.pruneCancel != nil {
		s.pruneCancel()
	}
	return nil
}

// Errors
var (
	ErrMessageNotFound = errors.New("message not found")
)
