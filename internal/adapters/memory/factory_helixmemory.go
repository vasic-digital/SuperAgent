// Package memory provides adapters between HelixAgent's internal/memory types
// and the extracted digital.vasic.memory module.
//
//go:build helixmemory

package memory

import (
	"log"

	helixcfg "digital.vasic.helixmemory/pkg/config"
	cogneeclient "digital.vasic.helixmemory/pkg/clients/cognee"
	graphiticlient "digital.vasic.helixmemory/pkg/clients/graphiti"
	lettaclient "digital.vasic.helixmemory/pkg/clients/letta"
	mem0client "digital.vasic.helixmemory/pkg/clients/mem0"
	helixprovider "digital.vasic.helixmemory/pkg/provider"
)

// NewOptimalStoreAdapter creates a StoreAdapter backed by the HelixMemory
// unified cognitive memory engine. It initializes all four backends (Mem0,
// Cognee, Letta, Graphiti) and registers them with the UnifiedMemoryProvider.
// The resulting adapter implements digital.vasic.memory/pkg/store.MemoryStore
// via the HelixMemory fusion engine.
//
// Configuration is read from HELIX_MEMORY_* environment variables.
// See HelixMemory/pkg/config for available settings.
func NewOptimalStoreAdapter() *StoreAdapter {
	cfg := helixcfg.FromEnv()

	// Create the unified provider (orchestrator)
	unified := helixprovider.New(cfg)

	// Register all four memory backends
	unified.RegisterProvider(mem0client.NewClient(cfg))
	unified.RegisterProvider(cogneeclient.NewClient(cfg))
	unified.RegisterProvider(lettaclient.NewClient(cfg))
	unified.RegisterProvider(graphiticlient.NewClient(cfg))

	// Create HelixMemory's MemoryStore adapter (implements digital.vasic.memory)
	helixStore := helixprovider.NewMemoryStoreAdapter(unified)

	log.Printf("[memory] Using HelixMemory unified engine (%d backends)",
		len(unified.AvailableProviders()))

	// Wrap it in HelixAgent's StoreAdapter
	return NewStoreAdapter(helixStore)
}

// IsHelixMemoryEnabled returns whether HelixMemory is compiled into the binary.
func IsHelixMemoryEnabled() bool {
	return true
}

// MemoryBackendName returns the name of the active memory backend.
func MemoryBackendName() string {
	return "digital.vasic.helixmemory"
}

// NewHelixMemoryProvider creates a standalone UnifiedMemoryProvider for
// direct access to HelixMemory features (temporal queries, core memory,
// consolidation, power features). Use this when you need more than
// the basic MemoryStore interface.
func NewHelixMemoryProvider() *helixprovider.UnifiedMemoryProvider {
	cfg := helixcfg.FromEnv()
	unified := helixprovider.New(cfg)

	unified.RegisterProvider(mem0client.NewClient(cfg))
	unified.RegisterProvider(cogneeclient.NewClient(cfg))
	unified.RegisterProvider(lettaclient.NewClient(cfg))
	unified.RegisterProvider(graphiticlient.NewClient(cfg))

	return unified
}
