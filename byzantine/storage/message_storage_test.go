package storage

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

func TestInMemoryMessageStorage(t *testing.T) {
	config := types.StorageConfig{
		MessageRetention: time.Hour,
		PruneInterval:    time.Minute,
	}

	storage := NewInMemoryMessageStorage(config)
	defer storage.Close()

	t.Run("Store and Retrieve", func(t *testing.T) {
		// Create test message
		msg := &types.StoredMessage{
			Message: &types.QBFTMessage{
				Code:     types.MessageCodePrepare,
				Sequence: 100,
				Round:    1,
				Hash:     common.HexToHash("0x1234"),
			},
			ReceivedAt: time.Now(),
			FromPeer:   common.HexToAddress("0xabcd"),
		}

		// Store message
		err := storage.Store(msg)
		require.NoError(t, err)

		// Retrieve by hash
		retrieved, err := storage.GetByHash(msg.Message.Hash)
		require.NoError(t, err)
		assert.Equal(t, msg.Message.Sequence, retrieved.Message.Sequence)

		// Retrieve by sequence and round
		messages, err := storage.GetBySequenceRound(100, 1)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, msg.Message.Hash, messages[0].Message.Hash)
	})

	t.Run("Prune Old Messages", func(t *testing.T) {
		// Create old message
		oldMsg := &types.StoredMessage{
			Message: &types.QBFTMessage{
				Code:     types.MessageCodePrepare,
				Sequence: 50,
				Round:    1,
				Hash:     common.HexToHash("0x5678"),
			},
			ReceivedAt: time.Now().Add(-2 * time.Hour),
			FromPeer:   common.HexToAddress("0xefgh"),
		}

		err := storage.Store(oldMsg)
		require.NoError(t, err)

		// Prune messages older than 1 hour
		err = storage.Prune(time.Now().Add(-time.Hour))
		require.NoError(t, err)

		// Old message should be removed
		_, err = storage.GetByHash(oldMsg.Message.Hash)
		assert.Error(t, err)
	})
}
