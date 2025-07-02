package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"

	"github.com/ethereum/go-ethereum/log"
)

// ConfigLoader loads Byzantine configuration
type ConfigLoader struct {
	paramRegistry *registry.ParameterParserRegistry
	uidGenerator  types.UIDGenerator
}

// NewConfigLoader creates a new config loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		paramRegistry: registry.NewParameterParserRegistry(),
		uidGenerator:  types.NewUIDGenerator(),
	}
}

// LoadConfig loads Byzantine configuration from file
func (cl *ConfigLoader) LoadConfig(configPath string) (*types.ByzantineConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return cl.LoadConfigFromReader(file)
}

// LoadConfigFromReader loads Byzantine configuration from io.Reader
func (cl *ConfigLoader) LoadConfigFromReader(reader io.Reader) (*types.ByzantineConfig, error) {
	var rawConfig struct {
		Enabled    bool              `json:"enabled"`
		Attacks    []json.RawMessage `json:"attacks,omitempty"`
		Storage    json.RawMessage   `json:"storage,omitempty"`
		Monitoring json.RawMessage   `json:"monitoring,omitempty"`
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&rawConfig); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	config := &types.ByzantineConfig{
		Enabled: rawConfig.Enabled,
	}

	// Parse attacks
	attacks, err := cl.parseAttacks(rawConfig.Attacks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse attacks: %w", err)
	}
	config.Attacks = attacks

	// Parse storage config
	if len(rawConfig.Storage) > 0 {
		if err := json.Unmarshal(rawConfig.Storage, &config.StorageConfig); err != nil {
			return nil, fmt.Errorf("failed to parse storage config: %w", err)
		}
	}

	// Parse monitoring config
	if len(rawConfig.Monitoring) > 0 {
		if err := json.Unmarshal(rawConfig.Monitoring, &config.Monitoring); err != nil {
			return nil, fmt.Errorf("failed to parse monitoring config: %w", err)
		}
	}

	// Validate the entire configuration
	if err := cl.validateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// parseAttackConfig parses a single attack configuration
func (cl *ConfigLoader) parseAttackConfig(raw json.RawMessage) (types.AttackConfig, error) {
	var config types.AttackConfig

	// Use the custom UnmarshalJSON which handles basic parsing
	if err := json.Unmarshal(raw, &config); err != nil {
		return config, fmt.Errorf("failed to unmarshal attack config: %w", err)
	}

	// Now parse the parameters using the registry (no import cycle)
	if config.RawParameters != nil {
		parser, exists := cl.paramRegistry.GetParser(config.Type)
		if exists {
			// Use ParseJSON for direct JSON parsing
			parsedParams, err := parser.ParseJSON(config.RawParameters)
			if err != nil {
				return config, fmt.Errorf("failed to parse %s parameters: %w", config.Type, err)
			}
			config.ParsedParameters = parsedParams
		}
	} else if config.Parameters != nil {
		// Fallback to map-based parsing for backward compatibility
		parser, exists := cl.paramRegistry.GetParser(config.Type)
		if exists {
			parsedParams, err := parser.Parse(config.Parameters)
			if err != nil {
				return config, fmt.Errorf("failed to parse %s parameters: %w", config.Type, err)
			}
			config.ParsedParameters = parsedParams
			log.Debug("[byzantine]", "type", config.Type, "parameters", config.Parameters)
			log.Debug("[byzantine]", "type", config.Type, "parsedParameters", config.ParsedParameters)
		}
	}

	return config, nil
}

// parseAttacks parses attack configurations with type-safe parameter handling
func (cl *ConfigLoader) parseAttacks(rawAttacks []json.RawMessage) ([]types.AttackConfig, error) {
	attacks := make([]types.AttackConfig, 0, len(rawAttacks))

	for i, rawAttack := range rawAttacks {
		attack, err := cl.parseAttackConfig(rawAttack)
		if err != nil {
			return nil, fmt.Errorf("failed to parse attack[%d]: %w", i, err)
		}
		attacks = append(attacks, attack)
	}

	return attacks, nil
}

//
//// parseAttackConfig parses a single attack configuration
//func (cl *ConfigLoader) parseAttackConfig(raw json.RawMessage) (types.AttackConfig, error) {
//	// First, unmarshal to get basic fields and determine attack type
//	var basicConfig struct {
//		Name       string                 `json:"name"`
//		Type       string                 `json:"type"`
//		Enabled    bool                   `json:"enabled"`
//		Sequence   uint64                 `json:"sequence"`
//		Round      uint64                 `json:"round"`
//		Parameters map[string]interface{} `json:"parameters,omitempty"`
//	}
//
//	if err := json.Unmarshal(raw, &basicConfig); err != nil {
//		return types.AttackConfig{}, fmt.Errorf("failed to unmarshal basic config: %w", err)
//	}
//
//	// Create attack config
//	config := types.AttackConfig{
//		Name:      basicConfig.Name,
//		Type:      types.StringToAttackType(basicConfig.Type),
//		Enabled:   basicConfig.Enabled,
//		Sequence:  basicConfig.Sequence,
//		Round:     basicConfig.Round,
//		Status:    types.AttackStatusPending,
//		CreatedAt: time.Now(),
//	}
//
//	// Store raw parameters for potential debugging
//	config.Parameters = basicConfig.Parameters
//
//	// Generate UID based on type, code, sequence, and round
//	config.UID = cl.uidGenerator.Generate(config.Type, config.Sequence, config.Round)
//
//	// Parse type-specific parameters
//	if basicConfig.Parameters != nil {
//		parsedParams, err := cl.paramRegistry.ParseParameters(config.Type, basicConfig.Parameters)
//		if err != nil {
//			return config, fmt.Errorf("failed to parse parameters for %s: %w", config.Name, err)
//		}
//
//		// For improved config, we would set ParsedParameters
//		// However, since we're using the existing AttackConfig, we need to
//		// ensure parameters are properly structured
//		config.Parameters = cl.restructureParameters(config.Type, parsedParams)
//		config.ParsedParameters = parsedParams
//	}
//
//	// Validate individual attack config
//	if err := cl.validateAttackConfig(&config); err != nil {
//		return config, fmt.Errorf("validation failed for %s: %w", config.Name, err)
//	}
//
//	log.Debug("Parsed attack configuration",
//		"name", config.Name,
//		"type", config.Type,
//		"uid", config.UID,
//		"sequence", config.Sequence,
//		"round", config.Round,
//		"code", config.Parameters["code"],
//		"parameters", config.Parameters)
//
//	return config, nil
//}

// restructureParameters converts parsed parameters back to map for storage
func (cl *ConfigLoader) restructureParameters(attackType types.AttackType, parsedParams interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	switch attackType {
	case types.AttackTypeSilentMessage:
		if params, ok := parsedParams.(*types.SilentAttackParams); ok {
			result["code"] = params.Code
			result["direction"] = params.Direction
			result["targets"] = params.Targets
		}

	case types.AttackTypeTamperedMessage:
		if params, ok := parsedParams.(*types.TamperAttackParams); ok {
			result["code"] = params.Code
			tamperFields := make([]map[string]interface{}, len(params.TamperFields))
			for i, field := range params.TamperFields {
				tamperFields[i] = map[string]interface{}{
					"target": field.Target,
					"value":  field.Value,
				}
			}
			result["tamperFields"] = tamperFields
			result["withValidMessage"] = params.WithValidMessage
			result["delay"] = params.Delay
			result["targets"] = params.Targets
		}

	case types.AttackTypeFakeMessage:
		if params, ok := parsedParams.(*types.FakeAttackParams); ok {
			result["code"] = params.Code
			if len(params.FakeMessage) > 0 {
				result["fakeMessage"] = string(params.FakeMessage)
			}
			result["targets"] = params.Targets
		}

	case types.AttackTypeOmitMessage:
		if params, ok := parsedParams.(*types.OmitAttackParams); ok {
			result["code"] = params.Code
			result["cmd"] = params.Cmd
			result["cnt"] = params.Cnt
			result["targets"] = params.Targets
		}

	case types.AttackTypeRoleSpoofed:
		if params, ok := parsedParams.(*types.RoleSpoofAttackParams); ok {
			result["code"] = params.Code
			if len(params.FakeMessage) > 0 {
				result["fakeMessage"] = string(params.FakeMessage)
			}
			result["targets"] = params.Targets
		}

	case types.AttackTypeReplay:
		if params, ok := parsedParams.(*types.ReplayAttackParams); ok {
			result["code"] = params.Code
			result["ori_sequence"] = params.OriSequence
			result["ori_round"] = params.OriRound
			result["useOriginalView"] = params.UseOriginalView
			result["targets"] = params.Targets
		}
	}

	return result
}

// validateConfig validates the entire Byzantine configuration
func (cl *ConfigLoader) validateConfig(config *types.ByzantineConfig) error {
	// Check for duplicate UIDs
	uidMap := make(map[string]string) // UID -> attack name
	for _, attack := range config.Attacks {
		if existingName, exists := uidMap[attack.UID]; exists {
			return fmt.Errorf("duplicate UID %s detected between attacks '%s' and '%s'",
				attack.UID, existingName, attack.Name)
		}
		uidMap[attack.UID] = attack.Name
	}

	// Validate storage config
	if config.StorageConfig.MessageRetention < 0 {
		return fmt.Errorf("invalid message retention: %s", config.StorageConfig.MessageRetention)
	}
	if config.StorageConfig.MaxStorageSize < 0 {
		return fmt.Errorf("invalid max storage size: %d", config.StorageConfig.MaxStorageSize)
	}

	// Validate monitoring config
	if config.Monitoring.Enabled && config.Monitoring.MetricsPort <= 0 {
		return fmt.Errorf("invalid metrics port: %d", config.Monitoring.MetricsPort)
	}

	return nil
}

// validateAttackConfig validates individual attack configuration
func (cl *ConfigLoader) validateAttackConfig(config *types.AttackConfig) error {
	// Basic validation
	if config.Name == "" {
		return fmt.Errorf("attack name is required")
	}

	if config.Type == "" {
		return fmt.Errorf("attack type is required")
	}

	// UID validation
	if config.UID == "" {
		return fmt.Errorf("attack UID is empty")
	}

	// Verify UID matches expected format
	expectedUID := cl.uidGenerator.GenerateWithRange(config.Type, config.SequenceStart, config.SequenceEnd, config.Round)
	if config.UID != expectedUID {
		return fmt.Errorf("UID mismatch: got %s, expected %s", config.UID, expectedUID)
	}

	// Type-specific parameter validation is done in the parser

	return nil
}

// SaveConfig saves Byzantine configuration to file
func (cl *ConfigLoader) SaveConfig(config *types.ByzantineConfig, configPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
