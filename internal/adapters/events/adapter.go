// Package eventsadapter bridges internal/events with the application.
// It provides a global event bus for publishing domain events
// (provider failures, debate completions, cache invalidations, etc.)
// using sync.Once for lazy initialization.
package eventsadapter

import (
	"sync"

	"dev.helix.agent/internal/events"
)

// globalBus is the lazily-initialized singleton event bus.
var (
	globalBus  *events.EventBus
	initOnce   sync.Once
	shutdownMu sync.Mutex
)

// Initialize creates the global event bus with the given config.
// It is safe to call multiple times; only the first call has effect.
// Pass nil to use the default configuration.
func Initialize(config *events.BusConfig) {
	initOnce.Do(func() {
		globalBus = events.NewEventBus(config)
	})
}

// GetBus returns the initialized event bus.
// If Initialize has not been called, it initializes the bus with
// default configuration (lazy initialization).
func GetBus() *events.EventBus {
	initOnce.Do(func() {
		globalBus = events.NewEventBus(nil)
	})
	return globalBus
}

// Shutdown gracefully shuts down the global event bus and resets
// the singleton so that Initialize can be called again.
func Shutdown() error {
	shutdownMu.Lock()
	defer shutdownMu.Unlock()

	if globalBus == nil {
		return nil
	}

	err := globalBus.Close()
	globalBus = nil
	// Reset sync.Once so the bus can be re-initialized (e.g. in tests).
	initOnce = sync.Once{}
	return err
}

// IsInitialized reports whether the global event bus has been created.
func IsInitialized() bool {
	shutdownMu.Lock()
	defer shutdownMu.Unlock()
	return globalBus != nil
}
