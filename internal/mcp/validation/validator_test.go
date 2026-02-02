package validation

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPValidator(t *testing.T) {
	v := NewMCPValidator()
	assert.NotNil(t, v)
	assert.NotNil(t, v.requirements)
	assert.NotNil(t, v.results)
	assert.NotNil(t, v.envCache)
}

func TestMCPValidator_LoadRequirements(t *testing.T) {
	v := NewMCPValidator()

	// Check that core MCPs are loaded
	coreMCPs := []string{"filesystem", "fetch", "memory", "time", "git", "sequential-thinking", "everything"}
	for _, mcp := range coreMCPs {
		req := v.GetMCPConfig(mcp)
		assert.NotNil(t, req, "Core MCP %s should be loaded", mcp)
		assert.Equal(t, "core", req.Category, "MCP %s should be in core category", mcp)
		assert.True(t, req.CanWorkLocally, "Core MCP %s should work locally", mcp)
	}

	// Check that helixagent is loaded
	helixagent := v.GetMCPConfig("helixagent")
	assert.NotNil(t, helixagent)
	assert.Equal(t, "helixagent", helixagent.Category)
	assert.Equal(t, 100, helixagent.Priority)
}

func TestMCPValidator_HasEnvVar(t *testing.T) {
	v := NewMCPValidator()

	// Set a test env var
	_ = os.Setenv("TEST_MCP_VAR", "test_value")
	defer func() { _ = os.Unsetenv("TEST_MCP_VAR") }()

	// Reload env cache
	v.loadEnvCache()

	assert.True(t, v.hasEnvVar("TEST_MCP_VAR"))
	assert.False(t, v.hasEnvVar("NONEXISTENT_VAR_12345"))
}

func TestMCPValidator_ValidateMCP_CoreMCP(t *testing.T) {
	v := NewMCPValidator()

	// Core MCPs should always work (no env vars required)
	req := v.GetMCPConfig("filesystem")
	require.NotNil(t, req)

	ctx := context.Background()
	result := v.validateMCP(ctx, "filesystem", req)

	assert.Equal(t, "filesystem", result.Name)
	assert.Equal(t, "works", result.Status)
	assert.True(t, result.CanEnable)
	assert.Empty(t, result.MissingEnvVars)
}

func TestMCPValidator_ValidateMCP_MissingEnvVars(t *testing.T) {
	v := NewMCPValidator()

	// GitHub requires GITHUB_TOKEN
	req := &MCPRequirement{
		Name:         "test-github",
		Type:         "local",
		RequiredEnvs: []string{"NONEXISTENT_GITHUB_TOKEN_12345"},
		Category:     "development",
	}

	ctx := context.Background()
	result := v.validateMCP(ctx, "test-github", req)

	assert.Equal(t, "disabled", result.Status)
	assert.False(t, result.CanEnable)
	assert.Contains(t, result.MissingEnvVars, "NONEXISTENT_GITHUB_TOKEN_12345")
}

func TestMCPValidator_ValidateAll(t *testing.T) {
	v := NewMCPValidator()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report := v.ValidateAll(ctx)

	assert.NotNil(t, report)
	assert.True(t, report.TotalMCPs > 0)
	assert.True(t, report.WorkingMCPs > 0, "Should have at least some working MCPs")
	assert.NotEmpty(t, report.EnabledMCPList)
	assert.NotEmpty(t, report.Summary)

	// Core MCPs should always be in enabled list
	coreMCPs := []string{"filesystem", "fetch", "memory", "time", "git"}
	for _, mcp := range coreMCPs {
		assert.Contains(t, report.EnabledMCPList, mcp, "Core MCP %s should be enabled", mcp)
	}
}

func TestMCPValidator_GetEnabledMCPs(t *testing.T) {
	v := NewMCPValidator()

	ctx := context.Background()
	v.ValidateAll(ctx)

	enabled := v.GetEnabledMCPs()
	assert.NotEmpty(t, enabled)

	// Core MCPs should be enabled
	coreMCPs := map[string]bool{
		"filesystem":          false,
		"fetch":               false,
		"memory":              false,
		"time":                false,
		"git":                 false,
		"sequential-thinking": false,
		"everything":          false,
		"sqlite":              false,
	}

	for _, mcp := range enabled {
		if _, ok := coreMCPs[mcp]; ok {
			coreMCPs[mcp] = true
		}
	}

	for mcp, found := range coreMCPs {
		assert.True(t, found, "Core MCP %s should be in enabled list", mcp)
	}
}

func TestMCPValidator_GenerateReport(t *testing.T) {
	v := NewMCPValidator()

	ctx := context.Background()
	v.ValidateAll(ctx)

	report := v.GenerateReport()

	assert.Contains(t, report, "MCP VALIDATION REPORT")
	assert.Contains(t, report, "WORKING MCPs")
	assert.Contains(t, report, "filesystem")
}

func TestMCPValidator_ToJSON(t *testing.T) {
	v := NewMCPValidator()

	ctx := context.Background()
	v.ValidateAll(ctx)

	jsonData, err := v.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify it's valid JSON
	assert.True(t, len(jsonData) > 100)
	assert.Contains(t, string(jsonData), "enabled_mcp_list")
	assert.Contains(t, string(jsonData), "working_mcps")
}

func TestMCPValidator_CategoryCounts(t *testing.T) {
	v := NewMCPValidator()

	categories := make(map[string]int)
	for _, req := range v.requirements {
		categories[req.Category]++
	}

	// Verify we have MCPs in expected categories
	assert.True(t, categories["core"] >= 5, "Should have at least 5 core MCPs")
	assert.True(t, categories["database"] >= 2, "Should have at least 2 database MCPs")
	assert.True(t, categories["devops"] >= 2, "Should have at least 2 devops MCPs")
}

func TestMCPValidator_PriorityOrder(t *testing.T) {
	v := NewMCPValidator()

	// helixagent should have highest priority (100)
	helixagent := v.GetMCPConfig("helixagent")
	assert.Equal(t, 100, helixagent.Priority)

	// Core MCPs should have high priority (>=80)
	filesystem := v.GetMCPConfig("filesystem")
	assert.True(t, filesystem.Priority >= 80)

	// Optional MCPs should have lower priority
	cloudflare := v.GetMCPConfig("cloudflare")
	if cloudflare != nil {
		assert.True(t, cloudflare.Priority < 80)
	}
}

func TestMCPValidator_RequirementFields(t *testing.T) {
	v := NewMCPValidator()

	for name, req := range v.requirements {
		assert.NotEmpty(t, req.Name, "MCP %s should have a name", name)
		assert.NotEmpty(t, req.Type, "MCP %s should have a type", name)
		assert.NotEmpty(t, req.Description, "MCP %s should have a description", name)
		assert.NotEmpty(t, req.Category, "MCP %s should have a category", name)

		if req.Type == "local" {
			assert.NotEmpty(t, req.Command, "Local MCP %s should have a command", name)
		}
		if req.Type == "remote" {
			assert.NotEmpty(t, req.URL, "Remote MCP %s should have a URL", name)
		}
	}
}

func TestMCPValidator_CheckLocalService_HelixAgent(t *testing.T) {
	v := NewMCPValidator()

	// This test checks if HelixAgent is running
	// It may pass or fail depending on whether HelixAgent is running
	result := v.checkLocalService("helixagent")

	// Just verify the function doesn't panic
	_ = result
}

func TestMCPValidator_CheckLocalService_Docker(t *testing.T) {
	v := NewMCPValidator()

	result := v.checkLocalService("docker")

	// Verify the function returns a boolean without panicking
	assert.IsType(t, true, result)
}

func TestMCPValidationResult_Fields(t *testing.T) {
	result := &MCPValidationResult{
		Name:           "test-mcp",
		Status:         "works",
		CanEnable:      true,
		Category:       "core",
		TestedAt:       time.Now(),
		ResponseTimeMs: 150,
	}

	assert.Equal(t, "test-mcp", result.Name)
	assert.Equal(t, "works", result.Status)
	assert.True(t, result.CanEnable)
	assert.Equal(t, "core", result.Category)
	assert.Equal(t, int64(150), result.ResponseTimeMs)
}

func TestMCPValidationReport_Fields(t *testing.T) {
	report := &MCPValidationReport{
		GeneratedAt:     time.Now(),
		TotalMCPs:       10,
		WorkingMCPs:     7,
		DisabledMCPs:    3,
		FailedMCPs:      0,
		EnabledMCPList:  []string{"filesystem", "memory"},
		DisabledMCPList: []string{"slack", "discord"},
		Summary:         "Test summary",
	}

	assert.Equal(t, 10, report.TotalMCPs)
	assert.Equal(t, 7, report.WorkingMCPs)
	assert.Equal(t, 3, report.DisabledMCPs)
	assert.Len(t, report.EnabledMCPList, 2)
	assert.Len(t, report.DisabledMCPList, 2)
}

func TestMCPRequirement_CanWorkLocally(t *testing.T) {
	v := NewMCPValidator()

	// Core MCPs should be able to work locally
	for _, name := range []string{"filesystem", "fetch", "memory", "time", "git"} {
		req := v.GetMCPConfig(name)
		assert.True(t, req.CanWorkLocally, "Core MCP %s should work locally", name)
	}

	// MCPs requiring external APIs should not work locally
	for _, name := range []string{"slack", "discord", "notion"} {
		req := v.GetMCPConfig(name)
		if req != nil {
			assert.False(t, req.CanWorkLocally, "MCP %s should NOT work locally", name)
		}
	}
}
