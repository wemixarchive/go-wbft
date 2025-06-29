package storage

import (
	"encoding/json"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

// InMemoryHistoryStorage implements HistoryStorage interface using in-memory storage
type InMemoryHistoryStorage struct {
	mu      sync.RWMutex
	configs map[string]types.AttackConfig
	results map[string][]types.AttackResult
	config  types.StorageConfig
}

// NewInMemoryHistoryStorage creates a new in-memory history storage
func NewInMemoryHistoryStorage(config types.StorageConfig) *InMemoryHistoryStorage {
	return &InMemoryHistoryStorage{
		configs: make(map[string]types.AttackConfig),
		results: make(map[string][]types.AttackResult),
		config:  config,
	}
}

// SaveAttackConfig saves an attack configuration
func (s *InMemoryHistoryStorage) SaveAttackConfig(config types.AttackConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs[config.UID] = config
	return nil
}

// SaveAttackResult saves an attack result
func (s *InMemoryHistoryStorage) SaveAttackResult(result types.AttackResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results[result.UID] = append(s.results[result.UID], result)
	return nil
}

// GetAttackHistory retrieves attack history by UID
func (s *InMemoryHistoryStorage) GetAttackHistory(uid string) ([]types.AttackResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results, exists := s.results[uid]
	if !exists {
		return nil, nil
	}

	// Return a copy to prevent external modification
	resultsCopy := make([]types.AttackResult, len(results))
	copy(resultsCopy, results)

	return resultsCopy, nil
}

// GetAllHistory retrieves all attack history
func (s *InMemoryHistoryStorage) GetAllHistory(limit int) ([]types.AttackResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allResults := make([]types.AttackResult, 0)

	for _, results := range s.results {
		allResults = append(allResults, results...)
		if limit > 0 && len(allResults) >= limit {
			break
		}
	}

	if limit > 0 && len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// Clear clears all history
func (s *InMemoryHistoryStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs = make(map[string]types.AttackConfig)
	s.results = make(map[string][]types.AttackResult)

	return nil
}

// Export exports all data as JSON
func (s *InMemoryHistoryStorage) Export() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := struct {
		Configs map[string]types.AttackConfig   `json:"configs"`
		Results map[string][]types.AttackResult `json:"results"`
	}{
		Configs: s.configs,
		Results: s.results,
	}

	return json.MarshalIndent(data, "", "  ")
}
