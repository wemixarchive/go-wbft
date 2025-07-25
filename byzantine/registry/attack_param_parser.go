package registry

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

// ParameterParserRegistry manages parameter parsers for different attack types
type ParameterParserRegistry struct {
	parsers map[types.AttackType]types.AttackParamsParser
}

// NewParameterParserRegistry creates a new parameter parser registry
func NewParameterParserRegistry() *ParameterParserRegistry {
	registry := &ParameterParserRegistry{
		parsers: make(map[types.AttackType]types.AttackParamsParser),
	}

	// Register default parsers
	registry.RegisterParser(types.AttackTypeSilentMessage, &types.SilentAttackParams{})
	registry.RegisterParser(types.AttackTypeTamperedMessage, &types.TamperAttackParams{})
	registry.RegisterParser(types.AttackTypeFakeMessage, &types.FakeAttackParams{})
	registry.RegisterParser(types.AttackTypeOmitMessage, &types.OmitAttackParams{})
	registry.RegisterParser(types.AttackTypeRoleSpoofed, &types.RoleSpoofAttackParams{})
	registry.RegisterParser(types.AttackTypeReplay, &types.ReplayAttackParams{})
	registry.RegisterParser(types.AttackTypeStoreMessage, &types.StoreAttackParams{})
	registry.RegisterParser(types.AttackTypeDos, &types.DosAttackParams{})
	return registry
}

// RegisterParser registers a parser for an attack type
func (r *ParameterParserRegistry) RegisterParser(attackType types.AttackType, parser types.AttackParamsParser) {
	r.parsers[attackType] = parser
}

// GetParser returns the parser for an attack type
func (r *ParameterParserRegistry) GetParser(attackType types.AttackType) (types.AttackParamsParser, bool) {
	parser, exists := r.parsers[attackType]
	return parser, exists
}

// ParseParameters parses parameters for a given attack type
func (r *ParameterParserRegistry) ParseParameters(attackType types.AttackType, raw map[string]interface{}) (interface{}, error) {
	parser, exists := r.GetParser(attackType)
	if !exists {
		return nil, fmt.Errorf("no parser registered for attack type: %s", attackType)
	}

	params, err := parser.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if err := parser.Validate(params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	return params, nil
}
