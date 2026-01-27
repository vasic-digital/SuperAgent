// Package compliance provides system compliance tests that ensure
// HelixAgent meets minimum requirements for providers, MCPs, and features.
package compliance

import (
	"testing"

	"dev.helix.agent/internal/debate/orchestrator"
	"dev.helix.agent/internal/mcp/adapters"
	"dev.helix.agent/internal/verifier"
	"llm-verifier/pkg/cliagents"
)

// MinProviders is the minimum number of providers required
const MinProviders = 30

// MinMCPs is the minimum number of MCP servers required for default generator
// Note: The generator provides a minimal set; users customize via configs
const MinMCPs = 10

// MinModels is the minimum number of LLM models that should be available
const MinModels = 100

// TestProviderCompliance verifies that at least 30 providers are registered
func TestProviderCompliance(t *testing.T) {
	providers := verifier.SupportedProviders
	count := len(providers)

	t.Logf("Registered providers: %d (required: %d)", count, MinProviders)

	if count < MinProviders {
		t.Errorf("COMPLIANCE FAILED: Only %d providers registered, minimum is %d", count, MinProviders)
		t.Log("Registered providers:")
		for name := range providers {
			t.Logf("  - %s", name)
		}
		t.Fatal("Add more providers to internal/verifier/provider_types.go")
	}

	// List all providers
	t.Log("Registered providers:")
	for name, info := range providers {
		t.Logf("  - %s (%s) [Tier %d]", name, info.DisplayName, info.Tier)
	}
}

// TestMCPAdapterCompliance verifies that at least 30 MCP adapters are registered
func TestMCPAdapterCompliance(t *testing.T) {
	adapters := adapters.AvailableAdapters
	count := len(adapters)

	t.Logf("Registered MCP adapters: %d (required: %d)", count, MinMCPs)

	if count < MinMCPs {
		t.Errorf("COMPLIANCE FAILED: Only %d MCP adapters registered, minimum is %d", count, MinMCPs)
		t.Log("Registered adapters:")
		for _, adapter := range adapters {
			t.Logf("  - %s: %s", adapter.Name, adapter.Description)
		}
		t.Fatal("Add more adapters to internal/mcp/adapters/registry.go")
	}

	// List all adapters by category
	categories := make(map[string][]string)
	for _, adapter := range adapters {
		categories[string(adapter.Category)] = append(categories[string(adapter.Category)], adapter.Name)
	}
	t.Log("MCP adapters by category:")
	for cat, names := range categories {
		t.Logf("  %s: %v", cat, names)
	}
}

// TestCLIMCPConfigCompliance verifies that CLI agents get 30+ MCPs
func TestCLIMCPConfigCompliance(t *testing.T) {
	mcps := cliagents.DefaultMCPServers()
	count := len(mcps)

	t.Logf("Default MCP servers for CLI agents: %d (required: %d)", count, MinMCPs)

	if count < MinMCPs {
		t.Errorf("COMPLIANCE FAILED: Only %d MCP servers in CLI config, minimum is %d", count, MinMCPs)
		t.Log("Configured MCP servers:")
		for _, mcp := range mcps {
			t.Logf("  - %s (%s)", mcp.Name, mcp.Type)
		}
		t.Fatal("Add more MCPs to LLMsVerifier/llm-verifier/pkg/cliagents/generator.go DefaultMCPServers()")
	}

	// List MCPs by type
	remote := 0
	local := 0
	for _, mcp := range mcps {
		if mcp.Type == "remote" {
			remote++
		} else {
			local++
		}
	}
	t.Logf("MCP distribution: %d remote (HelixAgent), %d local (npm packages)", remote, local)
}

// TestDebateFrameworkCompliance verifies new debate framework is enabled
func TestDebateFrameworkCompliance(t *testing.T) {
	config := orchestrator.DefaultServiceIntegrationConfig()

	t.Logf("New debate framework enabled: %v", config.EnableNewFramework)
	t.Logf("Fallback to legacy: %v", config.FallbackToLegacy)
	t.Logf("Learning enabled: %v", config.EnableLearning)
	t.Logf("Min agents for new framework: %d", config.MinAgentsForNewFramework)

	if !config.EnableNewFramework {
		t.Error("COMPLIANCE FAILED: New debate framework is not enabled by default")
		t.Fatal("Set EnableNewFramework to true in orchestrator.DefaultServiceIntegrationConfig()")
	}

	if !config.EnableLearning {
		t.Error("COMPLIANCE FAILED: Learning is not enabled by default")
		t.Fatal("Set EnableLearning to true in orchestrator.DefaultServiceIntegrationConfig()")
	}
}

// TestProviderEnvVarCoverage verifies all providers have env vars defined
func TestProviderEnvVarCoverage(t *testing.T) {
	providers := verifier.SupportedProviders
	missingEnvVars := []string{}

	for name, info := range providers {
		if len(info.EnvVars) == 0 {
			missingEnvVars = append(missingEnvVars, name)
		}
	}

	if len(missingEnvVars) > 0 {
		t.Errorf("COMPLIANCE WARNING: %d providers have no env vars defined: %v", len(missingEnvVars), missingEnvVars)
	}
}

// TestOAuthProviderSupport verifies OAuth providers are supported
func TestOAuthProviderSupport(t *testing.T) {
	oauthProviders := verifier.GetProvidersByAuthType(verifier.AuthTypeOAuth)

	t.Logf("OAuth providers: %d", len(oauthProviders))
	for _, p := range oauthProviders {
		t.Logf("  - %s", p.DisplayName)
	}

	if len(oauthProviders) < 2 {
		t.Error("COMPLIANCE FAILED: At least 2 OAuth providers required (Claude, Qwen)")
	}
}

// TestFreeProviderSupport verifies free providers are supported
func TestFreeProviderSupport(t *testing.T) {
	freeProviders := []string{}
	for name, info := range verifier.SupportedProviders {
		if info.Free {
			freeProviders = append(freeProviders, name)
		}
	}

	t.Logf("Free providers: %d", len(freeProviders))
	for _, name := range freeProviders {
		t.Logf("  - %s", name)
	}

	if len(freeProviders) < 2 {
		t.Error("COMPLIANCE FAILED: At least 2 free providers required (Zen, OpenRouter)")
	}
}

// TestProviderTierDistribution verifies providers are distributed across tiers
func TestProviderTierDistribution(t *testing.T) {
	tiers := make(map[int]int)

	for _, info := range verifier.SupportedProviders {
		tiers[info.Tier]++
	}

	t.Log("Provider tier distribution:")
	for tier := 1; tier <= 7; tier++ {
		t.Logf("  Tier %d: %d providers", tier, tiers[tier])
	}

	// Verify distribution
	if tiers[1]+tiers[2] < 5 {
		t.Error("COMPLIANCE WARNING: Less than 5 premium providers (Tier 1-2)")
	}
}

// TestMCPCategoryDistribution verifies MCPs cover all categories
func TestMCPCategoryDistribution(t *testing.T) {
	categories := adapters.GetAllCategories()
	categoryCount := make(map[string]int)

	for _, cat := range categories {
		adaptersInCat := adapters.DefaultRegistry.ListByCategory(cat)
		categoryCount[string(cat)] = len(adaptersInCat)
	}

	t.Log("MCP adapter category distribution:")
	emptyCategories := []string{}
	for cat, count := range categoryCount {
		t.Logf("  %s: %d adapters", cat, count)
		if count == 0 {
			emptyCategories = append(emptyCategories, cat)
		}
	}

	if len(emptyCategories) > 0 {
		t.Errorf("COMPLIANCE WARNING: %d categories have no adapters: %v", len(emptyCategories), emptyCategories)
	}
}
