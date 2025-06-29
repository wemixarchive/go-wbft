package types

import "encoding/json"

type ConfigWrapper struct {
	Enabled       bool                  `json:"enabled"`
	Attacks       []AttackConfigWrapper `json:"attacks"`
	StorageConfig StorageConfig         `json:"storage"`
	Monitoring    MonitoringConfig      `json:"monitoring"`
}

type AttackConfigWrapper struct {
	UID      string      `json:"uid,omitempty"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Enabled  bool        `json:"enabled"`
	Sequence uint64      `json:"sequence"`
	Round    uint64      `json:"round"`
	Code     interface{} `json:"code,omitempty"`

	// Raw parameters for parsing
	Parameters json.RawMessage `json:"parameters"`
}
