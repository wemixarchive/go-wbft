package manager

import (
	"context"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
	"sync"
)

// ChainHandler implements chain of responsibility pattern for attack processing
type ChainHandler struct {
	manager *AttackManager
	mu      sync.RWMutex
}

// NewChainHandler creates a new chain handler
func NewChainHandler(manager *AttackManager) *ChainHandler {
	return &ChainHandler{
		manager: manager,
	}
}

// ProcessEvent processes an event through the chain of attacks
func (h *ChainHandler) ProcessEvent(ctx context.Context, event types.Event) error {
	activeAttacks := h.manager.GetActiveAttacks()

	var wg sync.WaitGroup
	errors := make(chan error, len(activeAttacks))

	for _, attack := range activeAttacks {
		wg.Add(1)
		go func(attack types.Attack) {
			defer wg.Done()

			// Check if attack should be executed
			if attack.CheckExecuteCondition(ctx, event) {
				log.Debug("Processing event", "event", event)
			}
		}(attack)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Collect errors
	var firstError error
	for err := range errors {
		if firstError == nil {
			firstError = err
		}
	}

	return firstError
}

// Chain represents a chain of handlers
type Chain struct {
	handlers []Handler
}

// Handler represents a handler in the chain
type Handler interface {
	Handle(ctx context.Context, event types.Event, next HandlerFunc) error
}

// HandlerFunc represents a handler function
type HandlerFunc func(ctx context.Context, event types.Event) error

// NewChain creates a new chain
func NewChain(handlers ...Handler) *Chain {
	return &Chain{
		handlers: handlers,
	}
}

// Execute executes the chain
func (c *Chain) Execute(ctx context.Context, event types.Event) error {
	return c.execute(ctx, event, 0)
}

// execute recursively executes handlers
func (c *Chain) execute(ctx context.Context, event types.Event, index int) error {
	if index >= len(c.handlers) {
		return nil
	}

	return c.handlers[index].Handle(ctx, event, func(ctx context.Context, event types.Event) error {
		return c.execute(ctx, event, index+1)
	})
}
