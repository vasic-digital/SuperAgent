package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// ConfigGenerator Tests
// ========================================

func TestNewConfigGenerator(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	assert.Equal(t, "http://localhost:7061/v1", gen.baseURL)
	assert.Equal(t, "test-key", gen.apiKey)
	assert.Equal(t, "helixagent-debate", gen.model)
	assert.Equal(t, 120, gen.timeout)
	assert.Equal(t, 8192, gen.maxTokens)
}

func TestConfigGenerator_SetTimeout(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "model")
	gen.SetTimeout(300)
	assert.Equal(t, 300, gen.timeout)
}

func TestConfigGenerator_SetMaxTokens(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "model")
	gen.SetMaxTokens(4096)
	assert.Equal(t, 4096, gen.maxTokens)
}

func TestConfigGenerator_Validate_Success(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	err := gen.validate()
	assert.NoError(t, err)
}

func TestConfigGenerator_Validate_EmptyBaseURL(t *testing.T) {
	gen := NewConfigGenerator("", "test-key", "model")
	err := gen.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "baseURL is required")
}

func TestConfigGenerator_Validate_EmptyModel(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "")
	err := gen.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model is required")
}

func TestConfigGenerator_Validate_InvalidURL(t *testing.T) {
	gen := NewConfigGenerator("://invalid", "test-key", "model")
	err := gen.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid baseURL")
}

// ========================================
// OpenCode Config Generation Tests
// ========================================

func TestConfigGenerator_GenerateOpenCodeConfig(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	gen.SetTimeout(180)

	config, err := gen.GenerateOpenCodeConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://opencode.ai/config.json", config.Schema)
	assert.Contains(t, config.Provider, "helixagent")

	provider := config.Provider["helixagent"]
	assert.Equal(t, "@ai-sdk/openai-compatible", provider.NPM)
	assert.Equal(t, "http://localhost:7061/v1", provider.Options.BaseURL)
	assert.Equal(t, "test-key", provider.Options.APIKey)
	assert.Equal(t, 180000, provider.Options.Timeout) // Converted to ms
}

func TestConfigGenerator_GenerateOpenCodeConfig_JSON(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	jsonData, err := gen.GenerateJSON(AgentTypeOpenCode)
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, parsed, "$schema")
	assert.Contains(t, parsed, "provider")

	providers := parsed["provider"].(map[string]interface{})
	assert.Contains(t, providers, "helixagent")
}

// ========================================
// Crush Config Generation Tests
// ========================================

func TestConfigGenerator_GenerateCrushConfig(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	gen.SetMaxTokens(4096)

	config, err := gen.GenerateCrushConfig()
	require.NoError(t, err)

	assert.Equal(t, "https://charm.land/crush.json", config.Schema)
	assert.Contains(t, config.Providers, "helixagent")

	provider := config.Providers["helixagent"]
	assert.Equal(t, "openai-compat", provider.Type)
	assert.Equal(t, "http://localhost:7061/v1", provider.BaseURL)
	assert.Equal(t, "test-key", provider.APIKey)
	assert.Len(t, provider.Models, 1)
	assert.Equal(t, "helixagent-debate", provider.Models[0].ID)
	assert.Equal(t, 4096, provider.Models[0].DefaultMaxTokens)
}

func TestConfigGenerator_GenerateCrushConfig_JSON(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	jsonData, err := gen.GenerateJSON(AgentTypeCrush)
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, parsed, "$schema")
	assert.Contains(t, parsed, "providers")

	providers := parsed["providers"].(map[string]interface{})
	assert.Contains(t, providers, "helixagent")

	helixagent := providers["helixagent"].(map[string]interface{})
	assert.Equal(t, "openai-compat", helixagent["type"])
}

// ========================================
// HelixCode Config Generation Tests
// ========================================

func TestConfigGenerator_GenerateHelixCodeConfig(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	gen.SetTimeout(300).SetMaxTokens(16384)

	config, err := gen.GenerateHelixCodeConfig()
	require.NoError(t, err)

	assert.Contains(t, config.Providers, "helixagent")

	provider := config.Providers["helixagent"]
	assert.Equal(t, "openai-compatible", provider.Type)
	assert.Equal(t, "http://localhost:7061/v1", provider.BaseURL)
	assert.Equal(t, "test-key", provider.APIKey)
	assert.Equal(t, "helixagent-debate", provider.Model)
	assert.Equal(t, 16384, provider.MaxTokens)
	assert.Equal(t, 300, provider.Timeout)

	assert.Equal(t, "helixagent", config.Settings.DefaultProvider)
	assert.True(t, config.Settings.StreamingEnabled)
}

func TestConfigGenerator_GenerateHelixCodeConfig_JSON(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	jsonData, err := gen.GenerateJSON(AgentTypeHelixCode)
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, parsed, "providers")
	assert.Contains(t, parsed, "settings")
}

// ========================================
// GenerateConfig Generic Tests
// ========================================

func TestConfigGenerator_GenerateConfig_AllTypes(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	tests := []struct {
		agentType AgentType
		valid     bool
	}{
		{AgentTypeOpenCode, true},
		{AgentTypeCrush, true},
		{AgentTypeHelixCode, true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.agentType), func(t *testing.T) {
			config, err := gen.GenerateConfig(tt.agentType)
			if tt.valid {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			} else {
				assert.Error(t, err)
				assert.Nil(t, config)
			}
		})
	}
}

// ========================================
// ConfigValidator Tests
// ========================================

func TestNewConfigValidator(t *testing.T) {
	v := NewConfigValidator()
	assert.NotNil(t, v)
}

// ========================================
// OpenCode Validation Tests
// ========================================

func TestConfigValidator_ValidateOpenCodeConfig_Valid(t *testing.T) {
	v := NewConfigValidator()

	config := &OpenCodeConfig{
		Schema: "https://opencode.ai/config.json",
		Provider: map[string]OpenCodeProvider{
			"helixagent": {
				NPM: "@ai-sdk/openai-compatible",
				Options: OpenCodeProviderOptions{
					BaseURL: "http://localhost:7061/v1",
					APIKey:  "test-key",
					Timeout: 120000,
				},
			},
		},
	}

	result := v.ValidateOpenCodeConfig(config)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestConfigValidator_ValidateOpenCodeConfig_MissingProvider(t *testing.T) {
	v := NewConfigValidator()

	config := &OpenCodeConfig{
		Provider: map[string]OpenCodeProvider{},
	}

	result := v.ValidateOpenCodeConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "provider section is required")
}

func TestConfigValidator_ValidateOpenCodeConfig_MissingBaseURL(t *testing.T) {
	v := NewConfigValidator()

	config := &OpenCodeConfig{
		Provider: map[string]OpenCodeProvider{
			"test": {
				Options: OpenCodeProviderOptions{
					BaseURL: "",
				},
			},
		},
	}

	result := v.ValidateOpenCodeConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "baseURL is required")
}

func TestConfigValidator_ValidateOpenCodeConfig_InvalidMCPType(t *testing.T) {
	v := NewConfigValidator()

	config := &OpenCodeConfig{
		Provider: map[string]OpenCodeProvider{
			"test": {
				Options: OpenCodeProviderOptions{
					BaseURL: "http://localhost:7061",
				},
			},
		},
		MCP: map[string]OpenCodeMCP{
			"server1": {
				Type: "invalid",
			},
		},
	}

	result := v.ValidateOpenCodeConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "type must be 'local' or 'remote'")
}

// ========================================
// Crush Validation Tests
// ========================================

func TestConfigValidator_ValidateCrushConfig_Valid(t *testing.T) {
	v := NewConfigValidator()

	config := &CrushConfig{
		Schema: "https://charm.land/crush.json",
		Providers: map[string]CrushProvider{
			"helixagent": {
				Type:    "openai-compat",
				BaseURL: "http://localhost:7061/v1",
				APIKey:  "test-key",
				Models: []CrushModel{
					{ID: "helixagent-debate"},
				},
			},
		},
	}

	result := v.ValidateCrushConfig(config)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestConfigValidator_ValidateCrushConfig_InvalidType(t *testing.T) {
	v := NewConfigValidator()

	config := &CrushConfig{
		Providers: map[string]CrushProvider{
			"test": {
				Type: "invalid-type",
			},
		},
	}

	result := v.ValidateCrushConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "invalid type")
}

func TestConfigValidator_ValidateCrushConfig_OpenAICompatMissingBaseURL(t *testing.T) {
	v := NewConfigValidator()

	config := &CrushConfig{
		Providers: map[string]CrushProvider{
			"test": {
				Type: "openai-compat",
				// Missing BaseURL
			},
		},
	}

	result := v.ValidateCrushConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "base_url is required for openai-compat")
}

func TestConfigValidator_ValidateCrushConfig_MCPValidation(t *testing.T) {
	v := NewConfigValidator()

	tests := []struct {
		name        string
		mcpType     string
		command     string
		url         string
		expectValid bool
	}{
		{"stdio_valid", "stdio", "npx", "", true},
		{"stdio_missing_command", "stdio", "", "", false},
		{"http_valid", "http", "", "http://localhost:7061", true},
		{"http_missing_url", "http", "", "", false},
		{"sse_valid", "sse", "", "http://localhost:7061/sse", true},
		{"sse_missing_url", "sse", "", "", false},
		{"invalid_type", "invalid", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CrushConfig{
				Providers: map[string]CrushProvider{
					"test": {Type: "openai", BaseURL: "http://test"},
				},
				MCP: map[string]CrushMCP{
					"server": {
						Type:    tt.mcpType,
						Command: tt.command,
						URL:     tt.url,
					},
				},
			}

			result := v.ValidateCrushConfig(config)
			if tt.expectValid {
				assert.True(t, result.Valid, "Expected valid config for %s", tt.name)
			} else {
				assert.False(t, result.Valid, "Expected invalid config for %s", tt.name)
			}
		})
	}
}

// ========================================
// HelixCode Validation Tests
// ========================================

func TestConfigValidator_ValidateHelixCodeConfig_Valid(t *testing.T) {
	v := NewConfigValidator()

	config := &HelixCodeConfig{
		Providers: map[string]HelixCodeProvider{
			"helixagent": {
				Type:    "openai-compatible",
				BaseURL: "http://localhost:7061/v1",
				Model:   "helixagent-debate",
				Timeout: 120,
			},
		},
		Settings: HelixCodeSettings{
			DefaultProvider: "helixagent",
		},
	}

	result := v.ValidateHelixCodeConfig(config)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestConfigValidator_ValidateHelixCodeConfig_MissingBaseURL(t *testing.T) {
	v := NewConfigValidator()

	config := &HelixCodeConfig{
		Providers: map[string]HelixCodeProvider{
			"test": {
				Model: "model",
				// Missing BaseURL
			},
		},
	}

	result := v.ValidateHelixCodeConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "base_url is required")
}

func TestConfigValidator_ValidateHelixCodeConfig_MissingModel(t *testing.T) {
	v := NewConfigValidator()

	config := &HelixCodeConfig{
		Providers: map[string]HelixCodeProvider{
			"test": {
				BaseURL: "http://localhost:7061",
				// Missing Model
			},
		},
	}

	result := v.ValidateHelixCodeConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "model is required")
}

func TestConfigValidator_ValidateHelixCodeConfig_InvalidDefaultProvider(t *testing.T) {
	v := NewConfigValidator()

	config := &HelixCodeConfig{
		Providers: map[string]HelixCodeProvider{
			"test": {
				BaseURL: "http://localhost:7061",
				Model:   "model",
			},
		},
		Settings: HelixCodeSettings{
			DefaultProvider: "nonexistent",
		},
	}

	result := v.ValidateHelixCodeConfig(config)
	assert.False(t, result.Valid)
	assert.Contains(t, strings.Join(result.Errors, ","), "not found in providers")
}

// ========================================
// ValidateJSON Tests
// ========================================

func TestConfigValidator_ValidateJSON_OpenCode(t *testing.T) {
	v := NewConfigValidator()

	jsonData := `{
		"$schema": "https://opencode.ai/config.json",
		"provider": {
			"helixagent": {
				"npm": "@ai-sdk/openai-compatible",
				"options": {
					"baseURL": "http://localhost:7061/v1",
					"apiKey": "test-key"
				}
			}
		}
	}`

	result, err := v.ValidateJSON(AgentTypeOpenCode, []byte(jsonData))
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigValidator_ValidateJSON_Crush(t *testing.T) {
	v := NewConfigValidator()

	jsonData := `{
		"$schema": "https://charm.land/crush.json",
		"providers": {
			"helixagent": {
				"type": "openai-compat",
				"base_url": "http://localhost:7061/v1",
				"models": [{"id": "helixagent-debate"}]
			}
		}
	}`

	result, err := v.ValidateJSON(AgentTypeCrush, []byte(jsonData))
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigValidator_ValidateJSON_HelixCode(t *testing.T) {
	v := NewConfigValidator()

	jsonData := `{
		"providers": {
			"helixagent": {
				"type": "openai-compatible",
				"base_url": "http://localhost:7061/v1",
				"model": "helixagent-debate"
			}
		}
	}`

	result, err := v.ValidateJSON(AgentTypeHelixCode, []byte(jsonData))
	require.NoError(t, err)
	assert.True(t, result.Valid)
}

func TestConfigValidator_ValidateJSON_InvalidJSON(t *testing.T) {
	v := NewConfigValidator()

	jsonData := `{invalid json}`

	_, err := v.ValidateJSON(AgentTypeOpenCode, []byte(jsonData))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestConfigValidator_ValidateJSON_UnsupportedAgent(t *testing.T) {
	v := NewConfigValidator()

	_, err := v.ValidateJSON("unknown", []byte(`{}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported agent type")
}

// ========================================
// ValidationResult Tests
// ========================================

func TestValidationResult_String_Valid(t *testing.T) {
	result := &ValidationResult{Valid: true}
	str := result.String()
	assert.Contains(t, str, "Configuration is valid")
}

func TestValidationResult_String_Invalid(t *testing.T) {
	result := &ValidationResult{
		Valid:    false,
		Errors:   []string{"error1", "error2"},
		Warnings: []string{"warning1"},
	}
	str := result.String()
	assert.Contains(t, str, "Configuration is INVALID")
	assert.Contains(t, str, "error1")
	assert.Contains(t, str, "error2")
	assert.Contains(t, str, "warning1")
}

// ========================================
// Integration Tests
// ========================================

func TestConfigGenerator_GenerateAndValidate_AllAgents(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	v := NewConfigValidator()

	agents := []AgentType{AgentTypeOpenCode, AgentTypeCrush, AgentTypeHelixCode}

	for _, agent := range agents {
		t.Run(string(agent), func(t *testing.T) {
			// Generate config
			jsonData, err := gen.GenerateJSON(agent)
			require.NoError(t, err)

			// Validate generated config
			result, err := v.ValidateJSON(agent, jsonData)
			require.NoError(t, err)
			assert.True(t, result.Valid, "Generated config should be valid: %v", result.Errors)
		})
	}
}

// ========================================
// Benchmark Tests
// ========================================

func BenchmarkConfigGenerator_GenerateOpenCodeConfig(b *testing.B) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateOpenCodeConfig()
	}
}

func BenchmarkConfigGenerator_GenerateCrushConfig(b *testing.B) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateCrushConfig()
	}
}

func BenchmarkConfigGenerator_GenerateHelixCodeConfig(b *testing.B) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateHelixCodeConfig()
	}
}

func BenchmarkConfigValidator_ValidateOpenCodeConfig(b *testing.B) {
	v := NewConfigValidator()
	config := &OpenCodeConfig{
		Provider: map[string]OpenCodeProvider{
			"helixagent": {
				Options: OpenCodeProviderOptions{
					BaseURL: "http://localhost:7061/v1",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateOpenCodeConfig(config)
	}
}

// ========================================
// OpenCode Agent Section Regression Tests
// These tests prevent the bug where OpenCode closes without opening
// due to invalid agent section format.
//
// CRITICAL: The agent section requires NAMED sub-objects like "default",
// NOT a direct model property at the agent level.
//
// WRONG (causes OpenCode to close immediately):
//   "agent": {"model": "provider/model"}
//
// CORRECT (works properly):
//   "agent": {"default": {"model": "provider/model"}}
// ========================================

func TestOpenCodeConfig_AgentSection_HasNamedSubObject(t *testing.T) {
	// This is the most critical test - ensures agent section uses named sub-objects
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	config, err := gen.GenerateOpenCodeConfig()
	require.NoError(t, err)

	// Agent section must exist
	require.NotNil(t, config.Agent, "Agent section must be present")
	require.NotEmpty(t, config.Agent, "Agent section must have at least one named sub-object")

	// Agent section must have a "default" entry (or at least one named entry)
	defaultAgent, exists := config.Agent["default"]
	assert.True(t, exists, "Agent section must have a 'default' named sub-object")
	assert.NotEmpty(t, defaultAgent.Model, "Agent 'default' must have a model")
}

func TestOpenCodeConfig_AgentSection_ModelFormat(t *testing.T) {
	// Verify the model format is "provider/model"
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	config, err := gen.GenerateOpenCodeConfig()
	require.NoError(t, err)

	defaultAgent := config.Agent["default"]
	assert.Equal(t, "helixagent/helixagent-debate", defaultAgent.Model,
		"Agent model must be in format 'provider/model'")
	assert.Contains(t, defaultAgent.Model, "/",
		"Agent model must contain '/' separator")
}

func TestOpenCodeConfig_AgentSection_JSONStructure(t *testing.T) {
	// This test verifies the JSON structure is correct at the raw JSON level
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	jsonData, err := gen.GenerateJSON(AgentTypeOpenCode)
	require.NoError(t, err)

	// Parse as raw JSON to verify structure
	var rawJSON map[string]interface{}
	err = json.Unmarshal(jsonData, &rawJSON)
	require.NoError(t, err)

	// Check agent section exists
	agentSection, ok := rawJSON["agent"]
	require.True(t, ok, "JSON must have 'agent' key")
	require.NotNil(t, agentSection, "Agent section must not be nil")

	// Agent section must be a map (object with named keys)
	agentMap, ok := agentSection.(map[string]interface{})
	require.True(t, ok, "Agent section must be a JSON object, not a primitive")
	require.NotEmpty(t, agentMap, "Agent section must have at least one entry")

	// First level must contain named objects, not direct model property
	_, hasDirectModel := agentMap["model"]
	assert.False(t, hasDirectModel,
		"CRITICAL: Agent section must NOT have direct 'model' property - this causes OpenCode to crash!")

	// Check for named sub-objects (like "default")
	defaultEntry, hasDefault := agentMap["default"]
	assert.True(t, hasDefault, "Agent section should have 'default' named sub-object")

	// The named sub-object should have the model property
	if hasDefault {
		defaultMap, ok := defaultEntry.(map[string]interface{})
		require.True(t, ok, "default entry must be a JSON object")
		_, hasModel := defaultMap["model"]
		assert.True(t, hasModel, "default sub-object must have 'model' property")
	}
}

func TestOpenCodeConfig_AgentSection_Validation_MissingModel(t *testing.T) {
	v := NewConfigValidator()

	config := &OpenCodeConfig{
		Provider: map[string]OpenCodeProvider{
			"helixagent": {
				Options: OpenCodeProviderOptions{
					BaseURL: "http://localhost:7061/v1",
				},
			},
		},
		Agent: map[string]OpenCodeAgent{
			"default": {
				Model: "", // Empty model should fail validation
			},
		},
	}

	result := v.ValidateOpenCodeConfig(config)
	assert.False(t, result.Valid, "Config with empty agent model should be invalid")
	assert.Contains(t, strings.Join(result.Errors, ","), "model is required")
}

func TestOpenCodeConfig_AgentSection_Validation_InvalidModelFormat(t *testing.T) {
	v := NewConfigValidator()

	config := &OpenCodeConfig{
		Provider: map[string]OpenCodeProvider{
			"helixagent": {
				Options: OpenCodeProviderOptions{
					BaseURL: "http://localhost:7061/v1",
				},
			},
		},
		Agent: map[string]OpenCodeAgent{
			"default": {
				Model: "model-without-provider", // Missing provider/ prefix should warn
			},
		},
	}

	result := v.ValidateOpenCodeConfig(config)
	// Should be valid but with warning about format
	assert.True(t, result.Valid)
	assert.Contains(t, strings.Join(result.Warnings, ","), "should be in format 'provider/model'")
}

func TestOpenCodeConfig_AgentSection_CompleteWorkingConfig(t *testing.T) {
	// Test a complete configuration that matches what actually works with OpenCode
	v := NewConfigValidator()

	// This is the exact structure that was verified to work with OpenCode
	config := &OpenCodeConfig{
		Schema: "https://opencode.ai/config.json",
		Provider: map[string]OpenCodeProvider{
			"helixagent": {
				NPM:  "@ai-sdk/openai-compatible",
				Name: "HelixAgent AI Debate Ensemble",
				Options: OpenCodeProviderOptions{
					BaseURL: "http://localhost:7061/v1",
					APIKey:  "helixagent-local",
					Timeout: 600000,
				},
				Models: map[string]OpenCodeModel{
					"helixagent-debate": {
						Name:       "HelixAgent Debate Ensemble",
						Attachment: true,
						Reasoning:  true,
						ToolCall:   true,
						Limit: &OpenCodeLimit{
							Context: 128000,
							Output:  8192,
						},
					},
				},
			},
		},
		Agent: map[string]OpenCodeAgent{
			"default": {
				Model: "helixagent/helixagent-debate",
			},
		},
	}

	result := v.ValidateOpenCodeConfig(config)
	assert.True(t, result.Valid, "Complete working config should be valid: %v", result.Errors)
	assert.Empty(t, result.Errors, "Complete working config should have no errors")
}

func TestOpenCodeConfig_ModelsSection_Structure(t *testing.T) {
	// Verify the models section is properly structured
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	config, err := gen.GenerateOpenCodeConfig()
	require.NoError(t, err)

	provider := config.Provider["helixagent"]
	require.NotNil(t, provider.Models, "Provider must have models section")

	model, exists := provider.Models["helixagent-debate"]
	assert.True(t, exists, "Model must exist with correct key")
	assert.NotEmpty(t, model.Name, "Model must have name")
	assert.NotNil(t, model.Limit, "Model must have limits")
	assert.Greater(t, model.Limit.Context, 0, "Context limit must be positive")
	assert.Greater(t, model.Limit.Output, 0, "Output limit must be positive")
}

func TestOpenCodeConfig_ProviderSection_RequiredFields(t *testing.T) {
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	config, err := gen.GenerateOpenCodeConfig()
	require.NoError(t, err)

	provider := config.Provider["helixagent"]
	assert.NotEmpty(t, provider.NPM, "Provider must have npm package")
	assert.NotEmpty(t, provider.Name, "Provider must have name")
	assert.NotEmpty(t, provider.Options.BaseURL, "Provider must have baseURL")
}

func TestOpenCodeConfig_ValidateJSON_WithAgentSection(t *testing.T) {
	v := NewConfigValidator()

	// Valid config with proper agent section
	validJSON := `{
		"$schema": "https://opencode.ai/config.json",
		"provider": {
			"helixagent": {
				"npm": "@ai-sdk/openai-compatible",
				"options": {
					"baseURL": "http://localhost:7061/v1",
					"apiKey": "test-key"
				}
			}
		},
		"agent": {
			"default": {
				"model": "helixagent/helixagent-debate"
			}
		}
	}`

	result, err := v.ValidateJSON(AgentTypeOpenCode, []byte(validJSON))
	require.NoError(t, err)
	assert.True(t, result.Valid, "Config with proper agent section should be valid: %v", result.Errors)
}

func TestOpenCodeConfig_GeneratedConfigCanBeReParsed(t *testing.T) {
	// Test that generated config can be serialized and deserialized without loss
	gen := NewConfigGenerator("http://localhost:7061/v1", "test-key", "helixagent-debate")
	v := NewConfigValidator()

	// Generate config
	originalConfig, err := gen.GenerateOpenCodeConfig()
	require.NoError(t, err)

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(originalConfig, "", "  ")
	require.NoError(t, err)

	// Deserialize back
	var parsedConfig OpenCodeConfig
	err = json.Unmarshal(jsonData, &parsedConfig)
	require.NoError(t, err)

	// Validate parsed config
	result := v.ValidateOpenCodeConfig(&parsedConfig)
	assert.True(t, result.Valid, "Re-parsed config should be valid: %v", result.Errors)

	// Verify agent section survived round-trip
	assert.NotEmpty(t, parsedConfig.Agent, "Agent section must survive round-trip")
	defaultAgent, exists := parsedConfig.Agent["default"]
	assert.True(t, exists, "default agent must survive round-trip")
	assert.Equal(t, "helixagent/helixagent-debate", defaultAgent.Model,
		"Agent model must survive round-trip")
}
