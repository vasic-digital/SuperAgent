// Package memory provides adapters between HelixAgent's internal/memory types
// and the extracted digital.vasic.memory module.
//
//go:build !helixmemory

package memory

import "log"

// NewOptimalStoreAdapter creates a StoreAdapter using the default Memory module.
// When building without the helixmemory tag, the standard digital.vasic.memory
// module is used. The caller must provide their own modmem.MemoryStore
// implementation via NewStoreAdapter directly, or use this function which
// returns nil (no default store available without explicit configuration).
//
// To use the enhanced HelixMemory system with Mem0, Cognee, Letta, and Graphiti
// fusion, build with: go build -tags helixmemory
func NewOptimalStoreAdapter() *StoreAdapter {
	log.Println("[memory] Using standard Memory module (default)")
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
