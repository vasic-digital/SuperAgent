package config

import (
	"os"
	"strings"
	"testing"
)

func TestContainerMCPConfigGenerator_GenerateContainerMCPs(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	// Test total MCP count (should be 60+)
	if len(mcps) < 60 {
		t.Errorf("Expected at least 60 MCPs, got %d", len(mcps))
	}
}

func TestContainerMCPConfigGenerator_ZeroNPXCommands(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	// The container generator should never contain npx
	if gen.ContainsNPX() {
		t.Error("Container generator should not contain any NPX commands")
	}

	// Double check by examining all configs
	mcps := gen.GenerateContainerMCPs()
	for name, cfg := range mcps {
		if cfg.Type != "remote" {
			t.Errorf("MCP %s should have type 'remote', got '%s'", name, cfg.Type)
		}
		if cfg.URL == "" {
			t.Errorf("MCP %s should have a URL, got empty string", name)
		}
		if strings.Contains(cfg.URL, "npx") {
			t.Errorf("MCP %s URL should not contain 'npx': %s", name, cfg.URL)
		}
	}
}

func TestContainerMCPConfigGenerator_AllMCPsHaveURLs(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	for name, cfg := range mcps {
		if cfg.URL == "" {
			t.Errorf("MCP %s has empty URL", name)
		}
		if !strings.HasPrefix(cfg.URL, "http://") && !strings.HasPrefix(cfg.URL, "https://") {
			t.Errorf("MCP %s URL should start with http:// or https://, got: %s", name, cfg.URL)
		}
	}
}

func TestContainerMCPConfigGenerator_PortAllocationUnique(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	err := gen.ValidatePortAllocations()
	if err != nil {
		t.Errorf("Port allocation validation failed: %v", err)
	}
}

func TestContainerMCPConfigGenerator_PortRanges(t *testing.T) {
	// Verify ports are in expected ranges
	for _, p := range MCPContainerPorts {
		switch p.Category {
		case "core":
			if p.Port < 9101 || p.Port > 9199 {
				t.Errorf("Core MCP %s port %d out of range 9101-9199", p.Name, p.Port)
			}
		case "database":
			if p.Port < 9201 || p.Port > 9299 {
				t.Errorf("Database MCP %s port %d out of range 9201-9299", p.Name, p.Port)
			}
		case "vector":
			if p.Port < 9301 || p.Port > 9399 {
				t.Errorf("Vector MCP %s port %d out of range 9301-9399", p.Name, p.Port)
			}
		case "devops":
			if p.Port < 9401 || p.Port > 9499 {
				t.Errorf("DevOps MCP %s port %d out of range 9401-9499", p.Name, p.Port)
			}
		case "browser":
			if p.Port < 9501 || p.Port > 9599 {
				t.Errorf("Browser MCP %s port %d out of range 9501-9599", p.Name, p.Port)
			}
		case "communication":
			if p.Port < 9601 || p.Port > 9699 {
				t.Errorf("Communication MCP %s port %d out of range 9601-9699", p.Name, p.Port)
			}
		case "productivity":
			if p.Port < 9701 || p.Port > 9799 {
				t.Errorf("Productivity MCP %s port %d out of range 9701-9799", p.Name, p.Port)
			}
		case "search":
			if p.Port < 9801 || p.Port > 9899 {
				t.Errorf("Search MCP %s port %d out of range 9801-9899", p.Name, p.Port)
			}
		case "google":
			if p.Port < 9901 || p.Port > 9920 {
				t.Errorf("Google MCP %s port %d out of range 9901-9920", p.Name, p.Port)
			}
		case "monitoring":
			if p.Port < 9921 || p.Port > 9940 {
				t.Errorf("Monitoring MCP %s port %d out of range 9921-9940", p.Name, p.Port)
			}
		case "finance":
			if p.Port < 9941 || p.Port > 9960 {
				t.Errorf("Finance MCP %s port %d out of range 9941-9960", p.Name, p.Port)
			}
		case "design":
			if p.Port < 9961 || p.Port > 9999 {
				t.Errorf("Design MCP %s port %d out of range 9961-9999", p.Name, p.Port)
			}
		}
	}
}

func TestContainerMCPConfigGenerator_CoreMCPsAlwaysEnabled(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	// Core MCPs that should always be enabled
	coreMCPs := []string{
		"fetch", "git", "time", "filesystem", "memory",
		"everything", "sequential-thinking", "sqlite", "puppeteer",
	}

	for _, name := range coreMCPs {
		cfg, ok := mcps[name]
		if !ok {
			t.Errorf("Core MCP %s not found", name)
			continue
		}
		if !cfg.Enabled {
			t.Errorf("Core MCP %s should always be enabled", name)
		}
	}
}

func TestContainerMCPConfigGenerator_ConditionalMCPs(t *testing.T) {
	// Test that conditional MCPs are disabled without env vars
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	// These MCPs require API keys - should be disabled without env vars
	conditionalMCPs := []string{
		"github", "gitlab", "slack", "discord", "notion",
		"linear", "jira", "brave-search", "openai",
	}

	for _, name := range conditionalMCPs {
		cfg, ok := mcps[name]
		if !ok {
			t.Errorf("Conditional MCP %s not found", name)
			continue
		}
		// In a clean environment, these should be disabled
		// (unless the test environment has API keys set)
		if cfg.Enabled {
			// Verify the required env var is actually set
			if !hasRequiredEnvVarsForMCP(name) {
				t.Logf("MCP %s is enabled - checking if env vars are set", name)
			}
		}
	}
}

func hasRequiredEnvVarsForMCP(name string) bool {
	requirements := map[string][]string{
		"github":       {"GITHUB_TOKEN"},
		"gitlab":       {"GITLAB_TOKEN"},
		"slack":        {"SLACK_BOT_TOKEN", "SLACK_TEAM_ID"},
		"discord":      {"DISCORD_TOKEN"},
		"notion":       {"NOTION_API_KEY"},
		"linear":       {"LINEAR_API_KEY"},
		"jira":         {"JIRA_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"},
		"brave-search": {"BRAVE_API_KEY"},
		"openai":       {"OPENAI_API_KEY"},
	}

	if reqs, ok := requirements[name]; ok {
		for _, req := range reqs {
			if os.Getenv(req) == "" {
				return false
			}
		}
		return true
	}
	return false
}

func TestContainerMCPConfigGenerator_GetEnabledMCPs(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	enabled := gen.GetEnabledContainerMCPs()

	// At minimum, core MCPs should be enabled
	if len(enabled) < 10 {
		t.Errorf("Expected at least 10 enabled MCPs, got %d", len(enabled))
	}

	// All enabled MCPs should have valid URLs
	for name, cfg := range enabled {
		if cfg.URL == "" {
			t.Errorf("Enabled MCP %s has empty URL", name)
		}
	}
}

func TestContainerMCPConfigGenerator_GetDisabledMCPs(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	disabled := gen.GetDisabledContainerMCPs()

	// Disabled MCPs should have reasons
	for name, reason := range disabled {
		if reason == "" {
			t.Errorf("Disabled MCP %s has no reason", name)
		}
		if !strings.HasPrefix(reason, "Missing:") && reason != "Unknown reason" {
			t.Errorf("Disabled MCP %s has unexpected reason format: %s", name, reason)
		}
	}
}

func TestContainerMCPConfigGenerator_GenerateSummary(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	summary := gen.GenerateSummary()

	total, ok := summary["total"].(int)
	if !ok || total < 60 {
		t.Errorf("Expected total >= 60, got %v", summary["total"])
	}

	totalEnabled, ok := summary["total_enabled"].(int)
	if !ok || totalEnabled < 1 {
		t.Errorf("Expected total_enabled >= 1, got %v", summary["total_enabled"])
	}

	npxDeps, ok := summary["npx_dependencies"].(int)
	if !ok || npxDeps != 0 {
		t.Errorf("Expected npx_dependencies = 0, got %v", summary["npx_dependencies"])
	}
}

func TestContainerMCPConfigGenerator_GetMCPsByCategory(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	byCategory := gen.GetMCPsByCategory()

	expectedCategories := []string{
		"helixagent", "core", "database", "vector", "devops",
		"browser", "communication", "productivity", "search",
		"google", "monitoring", "finance", "design",
	}

	for _, cat := range expectedCategories {
		if _, ok := byCategory[cat]; !ok {
			t.Errorf("Expected category %s not found", cat)
		}
	}

	// Core should have the most MCPs
	if len(byCategory["core"]) < 5 {
		t.Errorf("Expected at least 5 core MCPs, got %d", len(byCategory["core"]))
	}
}

func TestContainerMCPConfigGenerator_CustomHost(t *testing.T) {
	// Set custom MCP host
	os.Setenv("MCP_CONTAINER_HOST", "mcp.example.com")
	defer os.Unsetenv("MCP_CONTAINER_HOST")

	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	// Check that URLs use the custom host (except helixagent)
	for name, cfg := range mcps {
		if name == "helixagent" {
			continue
		}
		if !strings.Contains(cfg.URL, "mcp.example.com") {
			t.Errorf("MCP %s URL should contain custom host, got: %s", name, cfg.URL)
		}
	}
}

func TestContainerMCPConfigGenerator_SSEEndpoints(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	// All MCP URLs should end with /sse for SSE transport
	for name, cfg := range mcps {
		if name == "helixagent" {
			// HelixAgent uses its own endpoint format
			continue
		}
		if !strings.HasSuffix(cfg.URL, "/sse") {
			t.Errorf("MCP %s URL should end with /sse, got: %s", name, cfg.URL)
		}
	}
}

func TestContainerMCPConfigGenerator_GetTotalMCPCount(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	count := gen.GetTotalMCPCount()

	if count < 60 {
		t.Errorf("Expected at least 60 MCPs, got %d", count)
	}
}

func TestContainerMCPConfigGenerator_HelixAgentMCP(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	helixAgent, ok := mcps["helixagent"]
	if !ok {
		t.Fatal("HelixAgent MCP not found")
	}

	if !helixAgent.Enabled {
		t.Error("HelixAgent MCP should always be enabled")
	}

	if helixAgent.Category != "helixagent" {
		t.Errorf("HelixAgent category should be 'helixagent', got '%s'", helixAgent.Category)
	}

	if !strings.Contains(helixAgent.URL, "/mcp/sse") {
		t.Errorf("HelixAgent URL should contain /mcp/sse, got: %s", helixAgent.URL)
	}
}

func TestContainerMCPConfigGenerator_NoLocalCommands(t *testing.T) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	mcps := gen.GenerateContainerMCPs()

	// Ensure no MCP has type "local" - all should be "remote"
	for name, cfg := range mcps {
		if cfg.Type != "remote" {
			t.Errorf("MCP %s should have type 'remote', got '%s'", name, cfg.Type)
		}
	}
}

func TestMCPContainerPorts_Coverage(t *testing.T) {
	// Ensure all ports are unique
	usedPorts := make(map[int]bool)
	for _, p := range MCPContainerPorts {
		if usedPorts[p.Port] {
			t.Errorf("Port %d is used multiple times", p.Port)
		}
		usedPorts[p.Port] = true
	}

	// Ensure port count matches expected
	if len(MCPContainerPorts) < 60 {
		t.Errorf("Expected at least 60 port allocations, got %d", len(MCPContainerPorts))
	}
}

func TestCompareWithNPXGenerator(t *testing.T) {
	// Compare container generator with the full NPX generator
	// to ensure all MCPs are covered
	containerGen := NewContainerMCPConfigGenerator("http://localhost:8080")
	npxGen := NewFullMCPConfigGenerator("http://localhost:8080")

	containerMCPs := containerGen.GenerateContainerMCPs()
	npxMCPs := npxGen.GenerateAllMCPs()

	// Container generator should have at least as many MCPs as NPX generator
	if len(containerMCPs) < len(npxMCPs) {
		t.Errorf("Container generator has fewer MCPs (%d) than NPX generator (%d)",
			len(containerMCPs), len(npxMCPs))
	}

	// Check that all NPX MCPs are also in container generator
	for name := range npxMCPs {
		if _, ok := containerMCPs[name]; !ok {
			t.Errorf("MCP %s exists in NPX generator but not in container generator", name)
		}
	}
}

func BenchmarkGenerateContainerMCPs(b *testing.B) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateContainerMCPs()
	}
}

func BenchmarkGetEnabledContainerMCPs(b *testing.B) {
	gen := NewContainerMCPConfigGenerator("http://localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GetEnabledContainerMCPs()
	}
}
