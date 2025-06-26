package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockByzantineService implements types.ByzantineService for testing
type MockByzantineService struct {
	mock.Mock
}

// Start implements types.ByzantineService
func (m *MockByzantineService) Start() error {
	args := m.Called()
	return args.Error(0)
}

// Stop implements types.ByzantineService
func (m *MockByzantineService) Stop() error {
	args := m.Called()
	return args.Error(0)
}

// GetStatus implements types.ByzantineService
func (m *MockByzantineService) GetStatus() types.ServiceStatus {
	args := m.Called()
	return args.Get(0).(types.ServiceStatus)
}

// Configure implements types.ByzantineService
func (m *MockByzantineService) Configure(config types.ByzantineConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

// GetMetrics implements types.ByzantineService
func (m *MockByzantineService) GetMetrics() types.Metrics {
	args := m.Called()
	return args.Get(0).(types.Metrics)
}

// RegisterAttack implements types.ByzantineService
func (m *MockByzantineService) RegisterAttack(config types.AttackConfig) (uint64, error) {
	args := m.Called(config)
	return args.Get(0).(uint64), args.Error(1)
}

// CancelAttack implements types.ByzantineService
func (m *MockByzantineService) CancelAttack(uid uint64) error {
	args := m.Called(uid)
	return args.Error(0)
}

// ListAttacks implements types.ByzantineService
func (m *MockByzantineService) ListAttacks() []types.AttackConfig {
	args := m.Called()
	return args.Get(0).([]types.AttackConfig)
}

// GetAttackHistory implements types.ByzantineService
func (m *MockByzantineService) GetAttackHistory(uid uint64) ([]types.AttackResult, error) {
	args := m.Called(uid)
	return args.Get(0).([]types.AttackResult), args.Error(1)
}

// GetAttackManager implements types.ByzantineService
func (m *MockByzantineService) GetAttackManager() types.AttackManager {
	args := m.Called()
	if val := args.Get(0); val != nil {
		return val.(types.AttackManager)
	}
	return nil
}

// GetMessageStorage implements types.ByzantineService
func (m *MockByzantineService) GetMessageStorage() types.MessageStorage {
	args := m.Called()
	if val := args.Get(0); val != nil {
		return val.(types.MessageStorage)
	}
	return nil
}

// GetHistoryStorage implements types.ByzantineService
func (m *MockByzantineService) GetHistoryStorage() types.HistoryStorage {
	args := m.Called()
	if val := args.Get(0); val != nil {
		return val.(types.HistoryStorage)
	}
	return nil
}

// GetConsensusHook implements types.ByzantineService
func (m *MockByzantineService) GetConsensusHook() types.ConsensusHook {
	args := m.Called()
	if val := args.Get(0); val != nil {
		return val.(types.ConsensusHook)
	}
	return nil
}

// Helper function to setup mock service with common expectations
func setupMockService() *MockByzantineService {
	mockService := new(MockByzantineService)

	// Handler creation requires these three methods to return something
	// They can return nil for most tests
	mockService.On("GetAttackManager").Return(nil)
	mockService.On("GetMessageStorage").Return(nil)
	mockService.On("GetHistoryStorage").Return(nil)

	return mockService
}

// Helper function to create mock service with custom attack manager
func setupMockServiceWithAttackManager(attackManager types.AttackManager) *MockByzantineService {
	mockService := new(MockByzantineService)

	mockService.On("GetAttackManager").Return(attackManager)
	mockService.On("GetMessageStorage").Return(nil)
	mockService.On("GetHistoryStorage").Return(nil)

	return mockService
}

// Additional mock types that might be needed

// MockAttackManager implements types.AttackManager
type MockAttackManager struct {
	mock.Mock
}

func (m *MockAttackManager) RegisterAttack(attack types.Attack) error {
	args := m.Called(attack)
	return args.Error(0)
}

func (m *MockAttackManager) UnregisterAttack(uid uint64) error {
	args := m.Called(uid)
	return args.Error(0)
}

func (m *MockAttackManager) GetAttack(uid uint64) (types.Attack, error) {
	args := m.Called(uid)
	if val := args.Get(0); val != nil {
		return val.(types.Attack), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAttackManager) ListAttacks() []types.Attack {
	args := m.Called()
	return args.Get(0).([]types.Attack)
}

func (m *MockAttackManager) ProcessEvent(ctx context.Context, event types.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAttackManager) GetActiveAttacks() []types.Attack {
	args := m.Called()
	return args.Get(0).([]types.Attack)
}

func (m *MockAttackManager) EvaluateAndExecuteAttacks(ctx context.Context, event types.Event) (types.AttackDecision, error) {
	args := m.Called(ctx, event)
	return args.Get(0).(types.AttackDecision), args.Error(1)
}

func (m *MockAttackManager) ProcessEventAsync(ctx context.Context, event types.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// MockMessageStorage implements types.MessageStorage
type MockMessageStorage struct {
	mock.Mock
}

func (m *MockMessageStorage) Store(message *types.StoredMessage) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageStorage) GetByHash(hash common.Hash) (*types.StoredMessage, error) {
	args := m.Called(hash)
	if val := args.Get(0); val != nil {
		return val.(*types.StoredMessage), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMessageStorage) GetBySequenceRound(sequence, round uint64) ([]*types.StoredMessage, error) {
	args := m.Called(sequence, round)
	return args.Get(0).([]*types.StoredMessage), args.Error(1)
}

func (m *MockMessageStorage) GetRecentMessages(limit int) ([]*types.StoredMessage, error) {
	args := m.Called(limit)
	return args.Get(0).([]*types.StoredMessage), args.Error(1)
}

func (m *MockMessageStorage) Prune(before time.Time) error {
	args := m.Called(before)
	return args.Error(0)
}

// MockHistoryStorage implements types.HistoryStorage
type MockHistoryStorage struct {
	mock.Mock
}

func (m *MockHistoryStorage) SaveAttackConfig(config types.AttackConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockHistoryStorage) SaveAttackResult(result types.AttackResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockHistoryStorage) GetAttackHistory(uid uint64) ([]types.AttackResult, error) {
	args := m.Called(uid)
	return args.Get(0).([]types.AttackResult), args.Error(1)
}

func (m *MockHistoryStorage) GetAllHistory(limit int) ([]types.AttackResult, error) {
	args := m.Called(limit)
	return args.Get(0).([]types.AttackResult), args.Error(1)
}

func (m *MockHistoryStorage) Clear() error {
	args := m.Called()
	return args.Error(0)
}

func TestAPI_ByzantineTests(t *testing.T) {
	// Given
	mockService := setupMockService()
	api := NewPublicByzantineAPI(mockService)

	attacks := []types.AttackConfig{
		{
			UID:      1,
			Name:     "test_attack",
			Type:     types.AttackTypeSilentMessage,
			Sequence: 100,
			Round:    0,
			Code:     types.MessageCodePrePrepare,
			Status:   types.AttackStatusActive,
		},
	}

	mockService.On("ListAttacks").Return(attacks)

	// When
	result, err := api.ByzantineTests()

	// Then
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, uint64(1), result[0].UID)
}

func TestAPI_StopByzantineTests(t *testing.T) {
	// 성공 케이스
	t.Run("success", func(t *testing.T) {
		mockService := setupMockService()
		api := NewPublicByzantineAPI(mockService)

		uids := []uint64{1, 2, 3}
		for _, uid := range uids {
			mockService.On("CancelAttack", uid).Return(nil)
		}

		err := api.StopByzantineTests(uids)
		assert.NoError(t, err)
	})

	// 부분 실패 케이스
	t.Run("partial failure", func(t *testing.T) {
		mockService := setupMockService()
		api := NewPublicByzantineAPI(mockService)

		mockService.On("CancelAttack", uint64(1)).Return(nil)
		mockService.On("CancelAttack", uint64(2)).Return(errors.New("cancel failed"))
		mockService.On("CancelAttack", uint64(3)).Return(nil)

		err := api.StopByzantineTests([]uint64{1, 2, 3})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stopped 2/3 attacks")
	})
}

func TestAPI_SilentMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := setupMockService()
		api := NewPublicByzantineAPI(mockService)

		params := map[string]interface{}{
			"sequence":  uint64(100),
			"round":     uint64(0),
			"code":      uint64(1),
			"direction": uint64(1),
			"targets":   []string{"0x1234567890123456789012345678901234567890"},
		}

		mockService.On("RegisterAttack", mock.MatchedBy(func(config types.AttackConfig) bool {
			return config.Type == types.AttackTypeSilentMessage &&
				config.Sequence == 100 &&
				config.Parameters["direction"] == uint64(1)
		})).Return(uint64(1), nil)

		err := api.SilentMessage(params)
		assert.NoError(t, err)
	})

	t.Run("invalid parameters", func(t *testing.T) {
		mockService := setupMockService()
		api := NewPublicByzantineAPI(mockService)

		testCases := []struct {
			name   string
			params map[string]interface{}
			errMsg string
		}{
			{
				name: "missing sequence",
				params: map[string]interface{}{
					"round":     uint64(0),
					"code":      uint64(1),
					"direction": uint64(1),
				},
				errMsg: "sequence number is required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := api.SilentMessage(tc.params)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})
}
