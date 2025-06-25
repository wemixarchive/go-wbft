package adapter

import (
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

// EventObserver implements the observer pattern for events
type EventObserver struct {
	mu          sync.RWMutex
	subscribers map[types.EventType][]types.EventHandler
}

// NewEventObserver creates a new event observer
func NewEventObserver() *EventObserver {
	return &EventObserver{
		subscribers: make(map[types.EventType][]types.EventHandler),
	}
}

// Publish publishes an event to all subscribers
func (o *EventObserver) Publish(event types.Event) error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Get handlers for this event type
	handlers, exists := o.subscribers[event.Type]
	if !exists || len(handlers) == 0 {
		return nil
	}

	// Call each handler in a separate goroutine
	var wg sync.WaitGroup
	errors := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h types.EventHandler) {
			defer wg.Done()
			if err := h(event); err != nil {
				errors <- err
			}
		}(handler)
	}

	wg.Wait()
	close(errors)

	// Return first error if any
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

// Subscribe subscribes to events of a specific type
func (o *EventObserver) Subscribe(eventType types.EventType, handler types.EventHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.subscribers[eventType] = append(o.subscribers[eventType], handler)
	return nil
}

// Unsubscribe removes a subscription
func (o *EventObserver) Unsubscribe(eventType types.EventType, handler types.EventHandler) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	handlers, exists := o.subscribers[eventType]
	if !exists {
		return nil
	}

	// Find and remove the handler
	for i, h := range handlers {
		// Compare function pointers
		if &h == &handler {
			o.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// Clear removes all subscriptions
func (o *EventObserver) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.subscribers = make(map[types.EventType][]types.EventHandler)
}

// EventCollector collects and aggregates events
type EventCollector struct {
	mu        sync.RWMutex
	events    []types.Event
	maxEvents int
}

// NewEventCollector creates a new event collector
func NewEventCollector(maxEvents int) *EventCollector {
	return &EventCollector{
		events:    make([]types.Event, 0),
		maxEvents: maxEvents,
	}
}

// Collect collects an event
func (c *EventCollector) Collect(event types.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.events = append(c.events, event)

	// Trim if exceeds max
	if len(c.events) > c.maxEvents {
		c.events = c.events[len(c.events)-c.maxEvents:]
	}
}

// GetEvents returns collected events
func (c *EventCollector) GetEvents() []types.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	events := make([]types.Event, len(c.events))
	copy(events, c.events)

	return events
}

// Clear clears all collected events
func (c *EventCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.events = make([]types.Event, 0)
}
