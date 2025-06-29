package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// ByzantineConfig represents the configuration for the Byzantine module
type ByzantineConfig struct {
	Enabled       bool             `json:"enabled"`
	Attacks       []AttackConfig   `json:"attacks,omitempty"`
	StorageConfig StorageConfig    `json:"storage,omitempty"`
	Monitoring    MonitoringConfig `json:"monitoring,omitempty"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	MessageRetention time.Duration `json:"message_retention"`
	HistoryRetention time.Duration `json:"history_retention"`
	MaxStorageSize   int64         `json:"max_storage_size"`
	PruneInterval    time.Duration `json:"prune_interval"`
}

// UnmarshalJSON implements custom JSON unmarshalling for StorageConfig
func (s *StorageConfig) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with string fields
	var temp struct {
		MessageRetention string `json:"message_retention"`
		HistoryRetention string `json:"history_retention"`
		MaxStorageSize   int64  `json:"max_storage_size"`
		PruneInterval    string `json:"prune_interval"`
	}

	// Unmarshal into temporary struct
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Parse duration strings
	if temp.MessageRetention != "" {
		duration, err := time.ParseDuration(temp.MessageRetention)
		if err != nil {
			return fmt.Errorf("invalid message_retention duration: %w", err)
		}
		s.MessageRetention = duration
	}

	if temp.HistoryRetention != "" {
		duration, err := time.ParseDuration(temp.HistoryRetention)
		if err != nil {
			return fmt.Errorf("invalid history_retention duration: %w", err)
		}
		s.HistoryRetention = duration
	}

	if temp.PruneInterval != "" {
		duration, err := time.ParseDuration(temp.PruneInterval)
		if err != nil {
			return fmt.Errorf("invalid prune_interval duration: %w", err)
		}
		s.PruneInterval = duration
	}

	// Copy non-duration fields
	s.MaxStorageSize = temp.MaxStorageSize

	return nil
}

// MarshalJSON implements custom JSON marshaling for StorageConfig
func (s StorageConfig) MarshalJSON() ([]byte, error) {
	// Convert durations back to strings for JSON
	type Alias StorageConfig
	return json.Marshal(&struct {
		MessageRetention string `json:"message_retention"`
		HistoryRetention string `json:"history_retention"`
		PruneInterval    string `json:"prune_interval"`
		*Alias
	}{
		MessageRetention: s.MessageRetention.String(),
		HistoryRetention: s.HistoryRetention.String(),
		PruneInterval:    s.PruneInterval.String(),
		Alias:            (*Alias)(&s),
	})
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled         bool           `json:"enabled"`
	MetricsPort     int            `json:"metrics_port"`
	LogLevel        string         `json:"log_level"`
	AlertThresholds map[string]int `json:"alert_thresholds"`
}
