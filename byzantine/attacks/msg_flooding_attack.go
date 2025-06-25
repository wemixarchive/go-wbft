package attacks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// MessageFloodingAttack implements message flooding (DDoS) attack
type MessageFloodingAttack struct {
	*registry.BaseAttack
	floodRate     uint64 // messages per second
	floodDuration time.Duration
	messageSize   uint64
	floodType     string // "valid", "invalid", "mixed"
	targets       []common.Address
}

// NewMessageFloodingAttack creates a new message flooding attack
func NewMessageFloodingAttack(config types.AttackConfig) (*MessageFloodingAttack, error) {
	targets, err := registry.ParseTargets(config)
	if err != nil {
		return nil, err
	}

	// Parse flooding parameters
	floodRate := registry.GetUint64Parameter(config, "floodRate", 100)
	floodDurationMs := registry.GetUint64Parameter(config, "floodDuration", 5000)
	messageSize := registry.GetUint64Parameter(config, "messageSize", 1024)

	floodType := "mixed"
	if typeParam, exists := config.Parameters["floodType"]; exists {
		if ft, ok := typeParam.(string); ok {
			floodType = ft
		}
	}

	return &MessageFloodingAttack{
		BaseAttack:    registry.NewBaseAttack(config),
		floodRate:     floodRate,
		floodDuration: time.Duration(floodDurationMs) * time.Millisecond,
		messageSize:   messageSize,
		floodType:     floodType,
		targets:       targets,
	}, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *MessageFloodingAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		return false
	}

	// Can trigger on any message event
	return event.Type == types.EventTypeMessageSent ||
		event.Type == types.EventTypeMessageReceived ||
		event.Type == types.EventTypeRoundChange
}

// Execute performs the message flooding attack
func (a *MessageFloodingAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()
	config := a.GetConfig()

	// Calculate flooding parameters
	totalMessages := uint64(a.floodDuration.Seconds() * float64(a.floodRate))
	interval := time.Second / time.Duration(a.floodRate)

	// Create flood context with timeout
	floodCtx, cancel := context.WithTimeout(ctx, a.floodDuration)
	defer cancel()

	// Track sent messages
	var sentCount uint64
	var errorCount uint64
	var mu sync.Mutex

	// Create worker pool for flooding
	workerCount := 10
	if a.floodRate > 1000 {
		workerCount = 20
	}

	messageChan := make(chan int, workerCount*2)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for msgNum := range messageChan {
				message := a.generateFloodMessage(config.Code, event, msgNum)
				if err := a.sendMessage(message, a.targets); err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
				} else {
					mu.Lock()
					sentCount++
					mu.Unlock()
				}
			}
		}()
	}

	// Generate messages at specified rate
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	msgNum := 0
	for {
		select {
		case <-floodCtx.Done():
			close(messageChan)
			wg.Wait()

			return &types.AttackResult{
				UID:        a.GetUID(),
				Success:    true,
				ExecutedAt: time.Now(),
				Duration:   time.Since(startTime),
				Details: map[string]interface{}{
					"total_messages_sent": sentCount,
					"errors":              errorCount,
					"flood_rate":          a.floodRate,
					"flood_duration":      a.floodDuration.String(),
					"message_size":        a.messageSize,
					"flood_type":          a.floodType,
					"targets":             len(a.targets),
				},
			}, nil

		case <-ticker.C:
			if msgNum < int(totalMessages) {
				select {
				case messageChan <- msgNum:
					msgNum++
				default:
					// Channel full, skip this tick
				}
			}
		}
	}
}

// generateFloodMessage generates a flood message based on type
func (a *MessageFloodingAttack) generateFloodMessage(messageCode types.MessageCode, event types.Event, msgNum int) []byte {
	var message []byte

	switch a.floodType {
	case "valid":
		message = a.generateValidMessage(messageCode, event, msgNum)
	case "invalid":
		message = a.generateInvalidMessage(messageCode, event, msgNum)
	case "mixed":
		if msgNum%2 == 0 {
			message = a.generateValidMessage(messageCode, event, msgNum)
		} else {
			message = a.generateInvalidMessage(messageCode, event, msgNum)
		}
	default:
		message = a.generateRandomMessage(a.messageSize)
	}

	return message
}

// generateValidMessage generates a valid-looking message
func (a *MessageFloodingAttack) generateValidMessage(messageCode types.MessageCode, event types.Event, msgNum int) []byte {
	msg := &types.QBFTMessage{
		Code:      messageCode,
		Sequence:  event.Sequence,
		Round:     event.Round + uint64(msgNum%3), // Vary round slightly
		Address:   common.HexToAddress(fmt.Sprintf("0x%040d", msgNum)),
		Signature: make([]byte, 65), // Valid signature size
	}

	// Add some padding to reach target message size
	serialized := serializeQBFTMessage(msg)
	if uint64(len(serialized)) < a.messageSize {
		padding := make([]byte, a.messageSize-uint64(len(serialized)))
		serialized = append(serialized, padding...)
	}

	return serialized
}

// generateInvalidMessage generates an invalid message
func (a *MessageFloodingAttack) generateInvalidMessage(messageCode types.MessageCode, event types.Event, msgNum int) []byte {
	// Create message with invalid fields
	msg := &types.QBFTMessage{
		Code:      messageCode,
		Sequence:  999999 + uint64(msgNum), // Invalid sequence
		Round:     999999 + uint64(msgNum), // Invalid round
		Address:   common.HexToAddress("0xDEADBEEF"),
		Signature: []byte("invalid_signature"),
	}

	return serializeQBFTMessage(msg)
}

// generateRandomMessage generates random bytes
func (a *MessageFloodingAttack) generateRandomMessage(size uint64) []byte {
	message := make([]byte, size)
	// Fill with pattern instead of random for efficiency
	for i := range message {
		message[i] = byte(i % 256)
	}
	return message
}

// sendMessage sends a message to targets
func (a *MessageFloodingAttack) sendMessage(content []byte, targets []common.Address) error {
	// Implementation depends on actual network layer
	// This would send messages as fast as possible
	return nil
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeMessageFlooding, func(config types.AttackConfig) (types.Attack, error) {
		return NewMessageFloodingAttack(config)
	})
	if err != nil {
		panic(err)
	}
}
