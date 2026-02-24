// Package memory provides adapters between HelixAgent's internal/memory types
// and the extracted digital.vasic.memory module.
//
// This is the opt-out fallback implementation using the standard Memory module.
// It is only active when you explicitly build with:
//
//	go build -tags nohelixmemory
//
//go:build nohelixmemory

package memory

import "log"

// NewOptimalStoreAdapter creates a StoreAdapter using the standard Memory module.
// This opt-out fallback is only active when building with -tags nohelixmemory.
// The caller must provide their own modmem.MemoryStore implementation via
// NewStoreAdapter directly, or use this function which returns nil (no default
// store available without explicit configuration).
//
// To restore the default HelixMemory unified engine, simply build without
// the nohelixmemory tag: go build
func NewOptimalStoreAdapter() *StoreAdapter {
	log.Println("[memory] Using standard Memory module (opt-out fallback)")
	return nil
}

// IsHelixMemoryEnabled returns whether HelixMemory is compiled into the binary.
func IsHelixMemoryEnabled() bool {
	return false
}

// MemoryBackendName returns the name of the active memory backend.
func MemoryBackendName() string {
	return "digital.vasic.memory"
}
