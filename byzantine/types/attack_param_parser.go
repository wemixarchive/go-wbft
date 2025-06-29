package types

import (
	"fmt"
)

// ParameterParserRegistry manages parameter parsers for different attack types
type ParameterParserRegistry struct {
	parsers map[AttackType]AttackParamsParser
}

// NewParameterParserRegistry creates a new parameter parser registry
func NewParameterParserRegistry() *ParameterParserRegistry {
	registry := &ParameterParserRegistry{
		parsers: make(map[AttackType]AttackParamsParser),
	}

	// Register default parsers
	registry.RegisterParser(AttackTypeSilentMessage, &SilentAttackParams{})
	registry.RegisterParser(AttackTypeTamperedMessage, &TamperAttackParams{})
	registry.RegisterParser(AttackTypeFakeMessage, &FakeAttackParams{})
	registry.RegisterParser(AttackTypeOmitMessage, &OmitAttackParams{})
	registry.RegisterParser(AttackTypeRoleSpoofed, &RoleSpoofAttackParams{})
	registry.RegisterParser(AttackTypeReplay, &ReplayAttackParams{})

	return registry
}

// RegisterParser registers a parser for an attack type
func (r *ParameterParserRegistry) RegisterParser(attackType AttackType, parser AttackParamsParser) {
	r.parsers[attackType] = parser
}

// GetParser returns the parser for an attack type
func (r *ParameterParserRegistry) GetParser(attackType AttackType) (AttackParamsParser, bool) {
	parser, exists := r.parsers[attackType]
	return parser, exists
}

// ParseParameters parses parameters for a given attack type
func (r *ParameterParserRegistry) ParseParameters(attackType AttackType, raw map[string]interface{}) (interface{}, error) {
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
