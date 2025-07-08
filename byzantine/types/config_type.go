package types

import (
	_ "encoding/json"
)

// ByzantineConfig represents the configuration for the Byzantine module
type ByzantineConfig struct {
	Enabled bool           `json:"enabled"`
	Attacks []AttackConfig `json:"attacks,omitempty"`
}
