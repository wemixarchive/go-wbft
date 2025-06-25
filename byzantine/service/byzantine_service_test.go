package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

func TestByzantineService(t *testing.T) {
	config := types.ByzantineConfig{
		Enabled: true,
		StorageConfig: types.StorageConfig{
			MessageRetention: time.Hour,
			HistoryRetention: 24 * time.Hour,
			PruneInterval:    time.Minute,
		},
	}

	service, err := NewByzantineService(&config)
	require.NoError(t, err)

	t.Run("Start and Stop", func(t *testing.T) {
		// Start service
		err := service.Start()
		require.NoError(t, err)

		// Check status
		status := service.GetStatus()
		assert.True(t, status.Running)
		assert.NotNil(t, status.StartedAt)

		// Stop service
		err = service.Stop()
		require.NoError(t, err)

		// Check status after stop
		status = service.GetStatus()
		assert.False(t, status.Running)
		assert.Nil(t, status.StartedAt)
	})

	t.Run("Register Attack", func(t *testing.T) {
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		attackConfig := types.AttackConfig{
			Type:     types.AttackTypeDoubleVote,
			Name:     "test_double_vote",
			Sequence: 100,
			Round:    1,
			Code:     types.MessageCodePrepare,
		}

		uid, err := service.RegisterAttack(attackConfig)
		require.NoError(t, err)
		assert.Greater(t, uid, uint64(0))

		// List attacks
		attacks := service.ListAttacks()
		assert.Len(t, attacks, 1)
		assert.Equal(t, uid, attacks[0].UID)
		assert.Equal(t, types.AttackStatusPending, attacks[0].Status)
	})

	t.Run("Cancel Attack", func(t *testing.T) {
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		// Register attack
		attackConfig := types.AttackConfig{
			Type:     types.AttackTypeSilentMessage,
			Name:     "test_silent",
			Sequence: 200,
			Round:    2,
			Code:     types.MessageCodePrePrepare,
		}

		uid, err := service.RegisterAttack(attackConfig)
		require.NoError(t, err)

		// Cancel attack
		err = service.CancelAttack(uid)
		require.NoError(t, err)

		// Check that attack is removed
		attacks := service.ListAttacks()
		assert.Len(t, attacks, 0)
	})
}

func TestByzantineServiceLifecycle(t *testing.T) {
	config := &types.ByzantineConfig{
		Enabled: true,
		// ... config setup ...
	}

	service, err := NewByzantineService(config)
	require.NoError(t, err)

	// Start should initialize context
	err = service.Start()
	require.NoError(t, err)
	require.NotNil(t, service.ctx)
	require.NotNil(t, service.cancel)

	// Stop should clean up context
	err = service.Stop()
	require.NoError(t, err)
	require.Nil(t, service.ctx)
	require.Nil(t, service.cancel)
}
