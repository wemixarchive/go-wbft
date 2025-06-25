package api

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type MockAttackBuilder struct {
	mock.Mock
}

func (m *MockAttackBuilder) WithType(attackType AttackType) AttackBuilder {
	args := m.Called(attackType)
	return args.Get(0).(AttackBuilder)
}

func (m *MockAttackBuilder) WithTiming(sequence, round uint64) AttackBuilder {
	args := m.Called(sequence, round)
	return args.Get(0).(AttackBuilder)
}

func (m *MockAttackBuilder) WithTargets(targets []common.Address) AttackBuilder {
	args := m.Called(targets)
	return args.Get(0).(AttackBuilder)
}

func (m *MockAttackBuilder) WithOptions(options map[string]interface{}) AttackBuilder {
	args := m.Called(options)
	return args.Get(0).(AttackBuilder)
}

func (m *MockAttackBuilder) Build() (Attack, error) {
	args := m.Called()
	return args.Get(0).(Attack), args.Error(1)
}

type MockAttack struct {
	mock.Mock
}

func (m *MockAttack) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	args := m.Called(sequence, round, msgCode)
	return args.Bool(0)
}

func (m *MockAttack) Execute(ctx AttackContext) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAttack) RequiresData() []DataRequirement {
	args := m.Called()
	return args.Get(0).([]DataRequirement)
}

// Test: API should receive attacks configuration request
func TestAPI_ConfigureAttack(t *testing.T) {
	// Given
	mockBuilder := new(MockAttackBuilder)
	mockAttack := new(MockAttack)

	api := NewByzantineAPI(mockBuilder)

	params := AttackParams{
		Type:     AttackTypeDoublePrepare,
		Sequence: 100,
		Round:    0,
		Target:   []common.Address{common.HexToAddress("0x1234567890123456789012345678901234567890")},
		Options: map[string]interface{}{
			"delay": uint64(1000),
		},
	}

	// Mock expectations
	mockBuilder.On("WithType", AttackTypeDoublePrepare).Return(mockBuilder)
	mockBuilder.On("WithTiming", uint64(100), uint64(0)).Return(mockBuilder)
	mockBuilder.On("WithTargets", params.Target).Return(mockBuilder)
	mockBuilder.On("WithOptions", params.Options).Return(mockBuilder)
	mockBuilder.On("Build").Return(mockAttack, nil)
	mockAttack.On("ID").Return("attacks-123")

	// When
	attackID, err := api.ConfigureAttack(params)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, "attacks-123", attackID)
	mockBuilder.AssertExpectations(t)
	mockAttack.AssertExpectations(t)
}

// Test: API should handle builder errors
func TestAPI_ConfigureAttack_BuilderError(t *testing.T) {
	// Given
	mockBuilder := new(MockAttackBuilder)
	api := NewByzantineAPI(mockBuilder)

	params := AttackParams{
		Type:     AttackTypeDoublePrepare,
		Sequence: 100,
		Round:    0,
	}

	// Mock expectations for builder error
	mockBuilder.On("WithType", AttackTypeDoublePrepare).Return(mockBuilder)
	mockBuilder.On("WithTiming", uint64(100), uint64(0)).Return(mockBuilder)
	mockBuilder.On("WithTargets", []common.Address(nil)).Return(mockBuilder)
	mockBuilder.On("WithOptions", map[string]interface{}(nil)).Return(mockBuilder)
	mockBuilder.On("Build").Return((*MockAttack)(nil), assert.AnError)

	// When
	attackID, err := api.ConfigureAttack(params)

	// Then
	assert.Error(t, err)
	assert.Empty(t, attackID)
}

// Test: API should validate parameters
func TestAPI_ConfigureAttack_InvalidParams(t *testing.T) {
	// Given
	mockBuilder := new(MockAttackBuilder)
	api := NewByzantineAPI(mockBuilder)

	// When - empty attacks type
	params := AttackParams{
		Sequence: 100,
		Round:    0,
	}

	attackID, err := api.ConfigureAttack(params)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "attacks type is required")
	assert.Empty(t, attackID)
}
