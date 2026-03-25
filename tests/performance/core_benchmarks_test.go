//go:build performance
// +build performance

// Package performance contains benchmark and load tests for critical components.
// This file provides 10 core benchmarks covering hot paths that are not
// covered by other benchmark files in this package.
package performance

import (
	"context"
	"fmt"
	"testing"

	"dev.helix.agent/internal/llm"
	"dev.helix.agent/internal/mcp/adapters"
	"dev.helix.agent/internal/services"
	"dev.helix.agent/internal/tools"
)

// =============================================================================
// 1. Provider selection from registry
// =============================================================================

// BenchmarkCore_ProviderRegistry_ListProviders measures the cost of
// enumerating registered providers from the in-memory registry map.
func BenchmarkCore_ProviderRegistry_ListProviders(b *testing.B) {
	cfg := &services.RegistryConfig{
		DisableAutoDiscovery: true,
	}
	reg := services.NewProviderRegistryWithoutAutoDiscovery(cfg, nil)

	// Pre-populate with 20 mock providers so the list is non-trivial.
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("bench-provider-%02d", i)
		_ = reg.RegisterProvider(name, &benchMockProvider{response: nil})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = reg.ListProviders()
	}
}

// BenchmarkCore_ProviderRegistry_GetProvider measures a hot-path read of a
// single already-initialized provider from the registry.
func BenchmarkCore_ProviderRegistry_GetProvider(b *testing.B) {
	cfg := &services.RegistryConfig{
		DisableAutoDiscovery: true,
	}
	reg := services.NewProviderRegistryWithoutAutoDiscovery(cfg, nil)
	_ = reg.RegisterProvider("bench-target", &benchMockProvider{response: nil})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = reg.GetProvider("bench-target")
	}
}

// =============================================================================
// 2. Circuit breaker state check
// =============================================================================

// BenchmarkCore_CircuitBreaker_GetState measures the RLock + state return in
// the closed (happy-path) circuit breaker.
func BenchmarkCore_CircuitBreaker_GetState(b *testing.B) {
	cb := llm.NewDefaultCircuitBreaker("bench-cb", &benchMockProvider{response: nil})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cb.GetState()
	}
}

// BenchmarkCore_CircuitBreaker_IsOpen measures the fast-path bool check used
// in every request admission decision.
func BenchmarkCore_CircuitBreaker_IsOpen(b *testing.B) {
	cb := llm.NewDefaultCircuitBreaker("bench-cb-open", &benchMockProvider{response: nil})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cb.IsOpen()
	}
}

// =============================================================================
// 3. Tool schema lookup and validation
// =============================================================================

// BenchmarkCore_ToolSchema_GetToolSchema measures a map lookup in the static
// global ToolSchemaRegistry.
func BenchmarkCore_ToolSchema_GetToolSchema(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = tools.GetToolSchema("Bash")
	}
}

// BenchmarkCore_ToolSchema_ValidateToolArgs measures the required-field
// validation loop for a well-formed call.
func BenchmarkCore_ToolSchema_ValidateToolArgs(b *testing.B) {
	args := map[string]interface{}{
		"command":     "echo hello",
		"description": "greet",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tools.ValidateToolArgs("Bash", args)
	}
}

// BenchmarkCore_ToolSchema_GetToolsByCategory measures filtering the global
// registry by category string.
func BenchmarkCore_ToolSchema_GetToolsByCategory(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = tools.GetToolsByCategory(tools.CategoryCore)
	}
}

// =============================================================================
// 4. MCP adapter registry resolution
// =============================================================================

// BenchmarkCore_MCPAdapterRegistry_Get measures eager-adapter lookup (fast
// path: single RLock + map read).
func BenchmarkCore_MCPAdapterRegistry_Get(b *testing.B) {
	reg := adapters.NewAdapterRegistry()
	stub := &benchMCPAdapter{name: "bench-mcp"}
	reg.Register("bench-mcp", stub, adapters.AdapterMetadata{
		Name:      "bench-mcp",
		Category:  adapters.CategoryUtility,
		Supported: true,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = reg.Get("bench-mcp")
	}
}

// BenchmarkCore_MCPAdapterRegistry_GetMetadata measures metadata-only lookup
// which avoids adapter initialization entirely.
func BenchmarkCore_MCPAdapterRegistry_GetMetadata(b *testing.B) {
	reg := adapters.NewAdapterRegistry()
	stub := &benchMCPAdapter{name: "bench-meta"}
	reg.Register("bench-meta", stub, adapters.AdapterMetadata{
		Name:      "bench-meta",
		Category:  adapters.CategorySearch,
		Supported: true,
	})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = reg.GetMetadata("bench-meta")
	}
}

// =============================================================================
// 5. Debate consensus detection
// =============================================================================

// BenchmarkCore_DebateOptimizer_ShouldTerminateEarly measures the consensus
// scoring loop used to trigger early exit when agents converge.
func BenchmarkCore_DebateOptimizer_ShouldTerminateEarly(b *testing.B) {
	cfg := services.DefaultDebateOptimizationConfig()
	cfg.EnableEarlyTermination = true
	cfg.EarlyTerminationThreshold = 0.95

	opt := services.NewDebatePerformanceOptimizer(cfg, nil, nil)

	// Build a response map that mirrors real debate output (5 positions).
	responses := map[services.DebateTeamPosition]string{
		services.PositionAnalyst:   "the solution involves refactoring the module boundary",
		services.PositionProposer:  "I propose refactoring module boundaries for clean separation",
		services.PositionCritic:    "refactoring module boundary addresses the core issue well",
		services.PositionSynthesis: "synthesising: module boundary refactoring is the right approach",
		services.PositionMediator:  "consensus: refactor module boundaries consistently across the system",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = opt.ShouldTerminateEarly(responses)
	}
}

// =============================================================================
// Helpers
// =============================================================================

// benchMCPAdapter is a minimal stub satisfying the adapters.MCPAdapter interface.
type benchMCPAdapter struct {
	name string
}

func (a *benchMCPAdapter) GetServerInfo() adapters.ServerInfo {
	return adapters.ServerInfo{Name: a.name, Version: "1.0.0"}
}

func (a *benchMCPAdapter) ListTools() []adapters.ToolDefinition {
	return nil
}

func (a *benchMCPAdapter) CallTool(
	_ context.Context, _ string, _ map[string]interface{},
) (*adapters.ToolResult, error) {
	return &adapters.ToolResult{}, nil
}
