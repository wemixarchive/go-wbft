package byzantine

import (
	"github.com/ethereum/go-ethereum/log"
)

// eventCollector implements EventCollector interface
type eventCollector struct {
	// TODO: Add actual implementation
}

// NewEventCollector creates a new event collector
func NewEventCollector() EventCollector {
	return &eventCollector{}
}

// Start starts event collection
func (c *eventCollector) Start() error {
	log.Info("Starting event collector")
	// TODO: Implement event collection from consensus
	return nil
}

// Stop stops event collection
func (c *eventCollector) Stop() error {
	log.Info("Stopping event collector")
	// TODO: Stop event collection
	return nil
}
