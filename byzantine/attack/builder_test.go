package attack

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

// Test: Builder should create DoublePrepare attack
func TestAttackBuilder_BuildDoublePrepareAttack(t *testing.T) {
	// Given
	builder := NewAttackBuilder()

	targets := []common.Address{
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		common.HexToAddress("0x2345678901234567890123456789012345678901"),
	}

	options := map[string]interface{}{
		"delay":            uint64(1000),
		"withValidMessage": true,
	}

	// When
	attack, err := builder.
		WithType(AttackTypeDoublePrepare).
		WithTiming(100, 0).
		WithTargets(targets).
		WithOptions(options).
		Build()

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, attack)
	assert.NotEmpty(t, attack.ID())

	// Verify attack properties
	assert.True(t, attack.ShouldExecute(100, 0, uint64(MessageCodePrepare)))
	assert.False(t, attack.ShouldExecute(99, 0, uint64(MessageCodePrepare)))
	assert.False(t, attack.ShouldExecute(100, 1, uint64(MessageCodePrepare)))
}

// Test: Builder should validate attack type
func TestAttackBuilder_InvalidAttackType(t *testing.T) {
	// Given
	builder := NewAttackBuilder()

	// When
	attack, err := builder.
		WithTiming(100, 0).
		Build()

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attack type is required")
	assert.Nil(t, attack)
}

// Test: Builder should create different attack types
func TestAttackBuilder_DifferentAttackTypes(t *testing.T) {
	tests := []struct {
		name       string
		attackType AttackType
		msgCode    uint64
	}{
		{
			name:       "DoublePrepare",
			attackType: AttackTypeDoublePrepare,
			msgCode:    uint64(MessageCodePrepare),
		},
		{
			name:       "DoubleCommit",
			attackType: AttackTypeDoubleCommit,
			msgCode:    uint64(MessageCodeCommit),
		},
		{
			name:       "SilentProposer",
			attackType: AttackTypeSilentProposer,
			msgCode:    uint64(MessageCodePrePrepare),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			builder := NewAttackBuilder()

			// When
			attack, err := builder.
				WithType(tt.attackType).
				WithTiming(100, 0).
				Build()

			// Then
			assert.NoError(t, err)
			assert.NotNil(t, attack)
			assert.True(t, attack.ShouldExecute(100, 0, tt.msgCode))
		})
	}
}

// Test: Builder should handle complex options
func TestAttackBuilder_ComplexOptions(t *testing.T) {
	// Given
	builder := NewAttackBuilder()

	options := map[string]interface{}{
		"tamperFields": []TamperField{
			{
				Target: "Proposal.Header.Coinbase",
				Value:  common.HexToAddress("0xdeadbeef"),
			},
			{
				Target: "Proposal.Header.Number",
				Value:  uint64(999),
			},
		},
		"messageCount": 10,
		"interval":     uint64(100),
	}

	// When
	attack, err := builder.
		WithType(AttackTypeTamperedHeader).
		WithTiming(100, 0).
		WithOptions(options).
		Build()

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, attack)

	// Verify data requirements
	dataReqs := attack.RequiresData()
	assert.NotEmpty(t, dataReqs)
}

// Test: Builder should be reusable
func TestAttackBuilder_Reusable(t *testing.T) {
	// Given
	builder := NewAttackBuilder()

	// When - Build first attack
	attack1, err1 := builder.
		WithType(AttackTypeDoublePrepare).
		WithTiming(100, 0).
		Build()

	// When - Build second attack with same builder
	attack2, err2 := builder.
		WithType(AttackTypeDoubleCommit).
		WithTiming(200, 1).
		Build()

	// Then
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, attack1.ID(), attack2.ID())
}

// Test: Builder should handle timing configuration
func TestAttackBuilder_TimingConfiguration(t *testing.T) {
	// Given
	builder := NewAttackBuilder()

	// When
	attack, err := builder.
		WithType(AttackTypeDoublePrepare).
		WithTiming(100, 2).
		Build()

	// Then
	assert.NoError(t, err)

	// Should execute only at specific sequence and round
	assert.True(t, attack.ShouldExecute(100, 2, uint64(MessageCodePrepare)))
	assert.False(t, attack.ShouldExecute(100, 1, uint64(MessageCodePrepare)))
	assert.False(t, attack.ShouldExecute(101, 2, uint64(MessageCodePrepare)))
}
