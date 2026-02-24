package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// ---- Struct types ----

func TestMultiProviderConfig_StructFields(t *testing.T) {
	cfg := MultiProviderConfig{
		Server: MultiServerConfig{
			Host:    "0.0.0.0",
			Port:    8080,
			Debug:   true,
			Timeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "admin",
			Password: "secret",
			Name:     "testdb",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "redis-pass",
		},
		Providers: map[string]*LLMProvider{
			"claude": {
				Name:       "claude",
				Type:       "llm",
				Enabled:    true,
				APIKey:     "test-key",
				BaseURL:    "https://api.anthropic.com",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				Weight:     1.0,
			},
		},
		Ensemble: &MultiEnsembleConfig{
			Strategy:            "confidence_weighted",
			MinProviders:        2,
			MaxProviders:        5,
			ConfidenceThreshold: 0.75,
			FallbackToBest:      true,
			Timeout:             60 * time.Second,
		},
		OpenAI: &OpenAIConfig{
			Enabled:           true,
			BasePath:          "/v1",
			ExposeAllModels:   true,
			EnsembleModelName: "helixagent-debate",
			EnableStreaming:   true,
			EnableFunctions:   true,
			EnableTools:       true,
		},
		MCP: &MCPConfig{
			Enabled:              true,
			ExposeAllTools:       true,
			UnifiedToolNamespace: true,
		},
		LSP: &LSPConfig{
			Enabled:               true,
			ExposeAllCapabilities: true,
		},
		Memory: &MemoryConfig{
			Enabled:          true,
			Provider:         "mem0",
			MaxContextLength: 32000,
			RetentionDays:    30,
		},
		Logging: &LoggingConfig{
			Level:                 "info",
			Format:                "json",
			EnableRequestLogging:  true,
			EnableResponseLogging: false,
		},
		Metrics: &MultiMetricsConfig{
			Enabled: true,
			Prometheus: &MultiPrometheusConfig{
				Enabled: true,
				Port:    9090,
				Path:    "/metrics",
			},
		},
	}

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.True(t, cfg.Server.Debug)
	assert.Equal(t, 30*time.Second, cfg.Server.Timeout)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "admin", cfg.Database.User)
	assert.Equal(t, "secret", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, "redis-pass", cfg.Redis.Password)

	require.NotNil(t, cfg.Providers["claude"])
	assert.Equal(t, "claude", cfg.Providers["claude"].Name)
	assert.True(t, cfg.Providers["claude"].Enabled)
	assert.Equal(t, "test-key", cfg.Providers["claude"].APIKey)

	assert.Equal(t, "confidence_weighted", cfg.Ensemble.Strategy)
	assert.Equal(t, 2, cfg.Ensemble.MinProviders)
	assert.Equal(t, 5, cfg.Ensemble.MaxProviders)

	assert.True(t, cfg.OpenAI.Enabled)
	assert.Equal(t, "/v1", cfg.OpenAI.BasePath)

	assert.True(t, cfg.MCP.Enabled)
	assert.True(t, cfg.LSP.Enabled)

	assert.Equal(t, "mem0", cfg.Memory.Provider)
	assert.Equal(t, 30, cfg.Memory.RetentionDays)

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)

	assert.True(t, cfg.Metrics.Enabled)
	assert.Equal(t, 9090, cfg.Metrics.Prometheus.Port)
}

func TestMultiServerConfig_Defaults(t *testing.T) {
	cfg := MultiServerConfig{}
	assert.Equal(t, "", cfg.Host)
	assert.Equal(t, 0, cfg.Port)
	assert.False(t, cfg.Debug)
	assert.Equal(t, time.Duration(0), cfg.Timeout)
}

func TestLLMProvider_AllFields(t *testing.T) {
	provider := LLMProvider{
		Name:       "openai",
		Type:       "llm",
		Enabled:    true,
		APIKey:     "sk-test-key",
		BaseURL:    "https://api.openai.com/v1",
		Timeout:    45 * time.Second,
		MaxRetries: 5,
		Weight:     1.5,
		Tags:       []string{"fast", "reliable"},
		Capabilities: map[string]string{
			"streaming": "true",
			"tools":     "true",
		},
		CustomSettings: map[string]interface{}{
			"org_id": "org-123",
		},
		Models: []ModelConfig{
			{
				ID:           "gpt-4",
				Name:         "GPT-4",
				Enabled:      true,
				Weight:       1.0,
				Capabilities: []string{"code_generation", "reasoning"},
				CustomParams: map[string]interface{}{
					"max_context": 128000,
				},
			},
		},
	}

	assert.Equal(t, "openai", provider.Name)
	assert.Equal(t, "sk-test-key", provider.APIKey)
	assert.Len(t, provider.Tags, 2)
	assert.Contains(t, provider.Tags, "fast")
	assert.Len(t, provider.Models, 1)
	assert.Equal(t, "gpt-4", provider.Models[0].ID)
	assert.Contains(t, provider.Models[0].Capabilities, "code_generation")
}

func TestModelConfig_Fields(t *testing.T) {
	m := ModelConfig{
		ID:           "model-1",
		Name:         "Test Model",
		Enabled:      true,
		Weight:       0.8,
		Capabilities: []string{"chat", "completion"},
		CustomParams: map[string]interface{}{
			"temperature": 0.7,
		},
	}

	assert.Equal(t, "model-1", m.ID)
	assert.Equal(t, "Test Model", m.Name)
	assert.True(t, m.Enabled)
	assert.Equal(t, 0.8, m.Weight)
	assert.Len(t, m.Capabilities, 2)
	assert.Equal(t, 0.7, m.CustomParams["temperature"])
}

func TestMultiEnsembleConfig_Fields(t *testing.T) {
	ec := MultiEnsembleConfig{
		Strategy:            "weighted",
		MinProviders:        1,
		MaxProviders:        10,
		ConfidenceThreshold: 0.8,
		FallbackToBest:      true,
		Timeout:             120 * time.Second,
		PreferredProviders:  []string{"claude", "openai"},
		ProviderWeights: map[string]float64{
			"claude": 1.5,
			"openai": 1.0,
		},
	}

	assert.Equal(t, "weighted", ec.Strategy)
	assert.Equal(t, 1, ec.MinProviders)
	assert.Equal(t, 10, ec.MaxProviders)
	assert.Equal(t, 0.8, ec.ConfidenceThreshold)
	assert.True(t, ec.FallbackToBest)
	assert.Len(t, ec.PreferredProviders, 2)
	assert.Equal(t, 1.5, ec.ProviderWeights["claude"])
}

func TestOpenAIConfig_Fields(t *testing.T) {
	oc := OpenAIConfig{
		Enabled:           true,
		BasePath:          "/v1",
		ExposeAllModels:   true,
		EnsembleModelName: "helixagent-debate",
		EnableStreaming:   true,
		EnableFunctions:   true,
		EnableTools:       true,
	}

	assert.True(t, oc.Enabled)
	assert.Equal(t, "/v1", oc.BasePath)
	assert.True(t, oc.ExposeAllModels)
	assert.Equal(t, "helixagent-debate", oc.EnsembleModelName)
	assert.True(t, oc.EnableStreaming)
	assert.True(t, oc.EnableFunctions)
	assert.True(t, oc.EnableTools)
}

func TestMCPConfig_Fields(t *testing.T) {
	mc := MCPConfig{
		Enabled:              true,
		ExposeAllTools:       true,
		UnifiedToolNamespace: true,
	}

	assert.True(t, mc.Enabled)
	assert.True(t, mc.ExposeAllTools)
	assert.True(t, mc.UnifiedToolNamespace)
}

func TestLSPConfig_Fields(t *testing.T) {
	lc := LSPConfig{
		Enabled:               true,
		ExposeAllCapabilities: false,
	}

	assert.True(t, lc.Enabled)
	assert.False(t, lc.ExposeAllCapabilities)
}

func TestMemoryConfig_Fields(t *testing.T) {
	mc := MemoryConfig{
		Enabled:          true,
		Provider:         "mem0",
		MaxContextLength: 64000,
		RetentionDays:    90,
	}

	assert.True(t, mc.Enabled)
	assert.Equal(t, "mem0", mc.Provider)
	assert.Equal(t, 64000, mc.MaxContextLength)
	assert.Equal(t, 90, mc.RetentionDays)
}

func TestLoggingConfig_Fields(t *testing.T) {
	lc := LoggingConfig{
		Level:                 "debug",
		Format:                "text",
		EnableRequestLogging:  true,
		EnableResponseLogging: true,
	}

	assert.Equal(t, "debug", lc.Level)
	assert.Equal(t, "text", lc.Format)
	assert.True(t, lc.EnableRequestLogging)
	assert.True(t, lc.EnableResponseLogging)
}

func TestMetricsConfig_Fields(t *testing.T) {
	mc := MetricsConfig{
		Enabled: true,
		Prometheus: &PrometheusConfig{
			Enabled: true,
			Path:    "/metrics",
			Port:    "9090",
		},
	}

	assert.True(t, mc.Enabled)
	assert.True(t, mc.Prometheus.Enabled)
	assert.Equal(t, "/metrics", mc.Prometheus.Path)
}

func TestMultiMetricsConfig_Fields(t *testing.T) {
	mc := MultiMetricsConfig{
		Enabled: true,
		Prometheus: &MultiPrometheusConfig{
			Enabled: true,
			Port:    9090,
			Path:    "/metrics",
		},
	}

	assert.True(t, mc.Enabled)
	assert.True(t, mc.Prometheus.Enabled)
	assert.Equal(t, 9090, mc.Prometheus.Port)
	assert.Equal(t, "/metrics", mc.Prometheus.Path)
}

func TestMultiPrometheusConfig_Fields(t *testing.T) {
	pc := MultiPrometheusConfig{
		Enabled: true,
		Port:    9091,
		Path:    "/custom-metrics",
	}

	assert.True(t, pc.Enabled)
	assert.Equal(t, 9091, pc.Port)
	assert.Equal(t, "/custom-metrics", pc.Path)
}

// ---- LoadMultiProviderConfig ----

func TestLoadMultiProviderConfig_ValidFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_multi_provider_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `
server:
  host: "0.0.0.0"
  port: 8080
  debug: true
  timeout: 30s
database:
  host: "localhost"
  port: "5432"
  user: "admin"
  password: "secret"
  name: "testdb"
redis:
  host: "localhost"
  port: "6379"
  password: "redis-pass"
providers:
  claude:
    name: "claude"
    type: "llm"
    enabled: true
    api_key: "test-key"
    base_url: "https://api.anthropic.com"
    timeout: 30s
    max_retries: 3
    weight: 1.0
    tags: ["primary", "fast"]
    models:
      - id: "claude-3-sonnet"
        name: "Claude 3 Sonnet"
        enabled: true
        weight: 1.0
  deepseek:
    name: "deepseek"
    type: "llm"
    enabled: true
    api_key: "ds-key"
    timeout: 25s
    max_retries: 2
    weight: 0.9
ensemble:
  strategy: "confidence_weighted"
  min_providers: 2
  max_providers: 5
  confidence_threshold: 0.75
  fallback_to_best: true
  timeout: 60s
  preferred_providers: ["claude", "deepseek"]
  provider_weights:
    claude: 1.5
    deepseek: 1.0
openai_compatible:
  enabled: true
  base_path: "/v1"
  expose_all_models: true
  ensemble_model_name: "helixagent-debate"
  enable_streaming: true
mcp:
  enabled: true
  expose_all_tools: true
lsp:
  enabled: true
memory:
  enabled: true
  provider: "mem0"
  max_context_length: 32000
  retention_days: 30
logging:
  level: "info"
  format: "json"
  enable_request_logging: true
metrics:
  enabled: true
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"
`

	err = os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

	cfg, err := LoadMultiProviderConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.True(t, cfg.Server.Debug)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "admin", cfg.Database.User)

	assert.Equal(t, "localhost", cfg.Redis.Host)

	require.Len(t, cfg.Providers, 2)
	require.NotNil(t, cfg.Providers["claude"])
	assert.Equal(t, "test-key", cfg.Providers["claude"].APIKey)
	assert.Equal(t, "claude", cfg.Providers["claude"].Name)
	assert.Len(t, cfg.Providers["claude"].Tags, 2)
	assert.Len(t, cfg.Providers["claude"].Models, 1)
	assert.Equal(t, "claude-3-sonnet", cfg.Providers["claude"].Models[0].ID)

	require.NotNil(t, cfg.Ensemble)
	assert.Equal(t, "confidence_weighted", cfg.Ensemble.Strategy)
	assert.Equal(t, 2, cfg.Ensemble.MinProviders)
	assert.Len(t, cfg.Ensemble.PreferredProviders, 2)
	assert.Equal(t, 1.5, cfg.Ensemble.ProviderWeights["claude"])

	require.NotNil(t, cfg.OpenAI)
	assert.True(t, cfg.OpenAI.Enabled)
	assert.Equal(t, "/v1", cfg.OpenAI.BasePath)

	require.NotNil(t, cfg.MCP)
	assert.True(t, cfg.MCP.Enabled)

	require.NotNil(t, cfg.LSP)
	assert.True(t, cfg.LSP.Enabled)

	require.NotNil(t, cfg.Memory)
	assert.Equal(t, "mem0", cfg.Memory.Provider)

	require.NotNil(t, cfg.Logging)
	assert.Equal(t, "info", cfg.Logging.Level)

	require.NotNil(t, cfg.Metrics)
	assert.True(t, cfg.Metrics.Enabled)
}

func TestLoadMultiProviderConfig_NonexistentFile(t *testing.T) {
	cfg, err := LoadMultiProviderConfig("/nonexistent/path/config.yaml")
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadMultiProviderConfig_InvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_invalid_multi_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid yaml: [[[")
	require.NoError(t, err)
	tmpFile.Close()

	cfg, loadErr := LoadMultiProviderConfig(tmpFile.Name())
	assert.Nil(t, cfg)
	require.Error(t, loadErr)
	assert.Contains(t, loadErr.Error(), "failed to parse config file")
}

func TestLoadMultiProviderConfig_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_empty_multi_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg, loadErr := LoadMultiProviderConfig(tmpFile.Name())
	require.NoError(t, loadErr)
	require.NotNil(t, cfg)
	// Empty file results in zero-value config
	assert.Equal(t, "", cfg.Server.Host)
	assert.Equal(t, 0, cfg.Server.Port)
	assert.Nil(t, cfg.Providers)
}

func TestLoadMultiProviderConfig_MinimalConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_minimal_multi_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "minimal.yaml")
	yamlContent := `
server:
  port: 7061
`

	err = os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

	cfg, loadErr := LoadMultiProviderConfig(configPath)
	require.NoError(t, loadErr)
	require.NotNil(t, cfg)
	assert.Equal(t, 7061, cfg.Server.Port)
	assert.Nil(t, cfg.Providers)
	assert.Nil(t, cfg.Ensemble)
	assert.Nil(t, cfg.OpenAI)
}

// ---- substituteEnvVars (multi-provider) ----

func TestSubstituteEnvVars_ProviderAPIKeys(t *testing.T) {
	os.Setenv("TEST_CLAUDE_KEY", "claude-api-key-123")
	os.Setenv("TEST_CLAUDE_URL", "https://custom-api.anthropic.com")
	defer func() {
		os.Unsetenv("TEST_CLAUDE_KEY")
		os.Unsetenv("TEST_CLAUDE_URL")
	}()

	tmpDir, err := os.MkdirTemp("", "test_env_sub_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `
server:
  port: 8080
database:
  host: "localhost"
redis:
  host: "localhost"
providers:
  claude:
    name: "claude"
    enabled: true
    api_key: "${TEST_CLAUDE_KEY}"
    base_url: "${TEST_CLAUDE_URL}"
`

	err = os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

	cfg, loadErr := LoadMultiProviderConfig(configPath)
	require.NoError(t, loadErr)
	require.NotNil(t, cfg.Providers["claude"])
	assert.Equal(t, "claude-api-key-123", cfg.Providers["claude"].APIKey)
	assert.Equal(t, "https://custom-api.anthropic.com", cfg.Providers["claude"].BaseURL)
}

func TestSubstituteEnvVars_DatabaseCredentials(t *testing.T) {
	os.Setenv("TEST_DB_HOST", "db.example.com")
	os.Setenv("TEST_DB_USER", "dbuser")
	os.Setenv("TEST_DB_PASS", "dbpass123")
	os.Setenv("TEST_DB_NAME", "production")
	os.Setenv("TEST_REDIS_HOST", "redis.example.com")
	os.Setenv("TEST_REDIS_PASS", "redispass456")
	defer func() {
		os.Unsetenv("TEST_DB_HOST")
		os.Unsetenv("TEST_DB_USER")
		os.Unsetenv("TEST_DB_PASS")
		os.Unsetenv("TEST_DB_NAME")
		os.Unsetenv("TEST_REDIS_HOST")
		os.Unsetenv("TEST_REDIS_PASS")
	}()

	tmpDir, err := os.MkdirTemp("", "test_db_env_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `
server:
  port: 8080
database:
  host: "${TEST_DB_HOST}"
  user: "${TEST_DB_USER}"
  password: "${TEST_DB_PASS}"
  name: "${TEST_DB_NAME}"
redis:
  host: "${TEST_REDIS_HOST}"
  password: "${TEST_REDIS_PASS}"
`

	err = os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

	cfg, loadErr := LoadMultiProviderConfig(configPath)
	require.NoError(t, loadErr)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "dbuser", cfg.Database.User)
	assert.Equal(t, "dbpass123", cfg.Database.Password)
	assert.Equal(t, "production", cfg.Database.Name)
	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, "redispass456", cfg.Redis.Password)
}

func TestSubstituteEnvVars_ModelCustomParams(t *testing.T) {
	os.Setenv("TEST_MODEL_PARAM", "custom-value")
	defer os.Unsetenv("TEST_MODEL_PARAM")

	tmpDir, err := os.MkdirTemp("", "test_model_env_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `
server:
  port: 8080
database:
  host: "localhost"
redis:
  host: "localhost"
providers:
  claude:
    name: "claude"
    enabled: true
    api_key: "static-key"
    models:
      - id: "model-1"
        name: "Model 1"
        enabled: true
        custom_params:
          proxy: "${TEST_MODEL_PARAM}"
          count: 42
`

	err = os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

	cfg, loadErr := LoadMultiProviderConfig(configPath)
	require.NoError(t, loadErr)
	require.NotNil(t, cfg.Providers["claude"])
	require.Len(t, cfg.Providers["claude"].Models, 1)
	assert.Equal(t, "custom-value", cfg.Providers["claude"].Models[0].CustomParams["proxy"])
}

func TestSubstituteEnvVars_NoProviders(t *testing.T) {
	// substituteEnvVars should not panic with nil/empty providers
	cfg := &MultiProviderConfig{
		Providers: nil,
	}
	err := substituteEnvVars(cfg)
	assert.NoError(t, err)
}

func TestSubstituteEnvVars_EmptyProviders(t *testing.T) {
	cfg := &MultiProviderConfig{
		Providers: map[string]*LLMProvider{},
	}
	err := substituteEnvVars(cfg)
	assert.NoError(t, err)
}

// ---- YAML/JSON parsing with all struct types ----

func TestLoadMultiProviderConfig_FullYAMLParsing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_full_yaml_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "full.yaml")
	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9090
  debug: false
  timeout: 45s
database:
  host: "db.local"
  port: "15432"
  user: "helixagent"
  password: "pass123"
  name: "helix_db"
redis:
  host: "redis.local"
  port: "16379"
  password: "redis123"
providers:
  openai:
    name: "openai"
    type: "llm"
    enabled: true
    api_key: "sk-test"
    base_url: "https://api.openai.com/v1"
    timeout: 30s
    max_retries: 3
    weight: 1.0
    tags: ["reliable"]
    capabilities:
      streaming: "true"
      tools: "true"
    custom_settings:
      org_id: "org-123"
    models:
      - id: "gpt-4"
        name: "GPT-4"
        enabled: true
        weight: 1.0
        capabilities: ["chat", "code"]
ensemble:
  strategy: "majority"
  min_providers: 3
  max_providers: 7
  confidence_threshold: 0.85
  fallback_to_best: false
  timeout: 90s
  preferred_providers: ["openai"]
  provider_weights:
    openai: 2.0
openai_compatible:
  enabled: true
  base_path: "/v1"
  expose_all_models: false
  ensemble_model_name: "custom-ensemble"
  enable_streaming: true
  enable_functions: true
  enable_tools: false
mcp:
  enabled: false
  expose_all_tools: false
  unified_tool_namespace: true
lsp:
  enabled: true
  expose_all_capabilities: false
memory:
  enabled: true
  provider: "mem0"
  max_context_length: 64000
  retention_days: 60
logging:
  level: "debug"
  format: "text"
  enable_request_logging: true
  enable_response_logging: true
metrics:
  enabled: true
  prometheus:
    enabled: true
    port: 9091
    path: "/custom-metrics"
`

	err = os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

	cfg, loadErr := LoadMultiProviderConfig(configPath)
	require.NoError(t, loadErr)

	// Verify server
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.False(t, cfg.Server.Debug)
	assert.Equal(t, 45*time.Second, cfg.Server.Timeout)

	// Verify database
	assert.Equal(t, "db.local", cfg.Database.Host)
	assert.Equal(t, "15432", cfg.Database.Port)

	// Verify redis
	assert.Equal(t, "redis.local", cfg.Redis.Host)
	assert.Equal(t, "16379", cfg.Redis.Port)

	// Verify providers
	openai := cfg.Providers["openai"]
	require.NotNil(t, openai)
	assert.Equal(t, "openai", openai.Name)
	assert.Equal(t, "llm", openai.Type)
	assert.True(t, openai.Enabled)
	assert.Equal(t, "sk-test", openai.APIKey)
	assert.Equal(t, "https://api.openai.com/v1", openai.BaseURL)
	assert.Equal(t, 30*time.Second, openai.Timeout)
	assert.Equal(t, 3, openai.MaxRetries)
	assert.Equal(t, 1.0, openai.Weight)
	assert.Contains(t, openai.Tags, "reliable")
	assert.Equal(t, "true", openai.Capabilities["streaming"])
	assert.Len(t, openai.Models, 1)
	assert.Equal(t, "gpt-4", openai.Models[0].ID)
	assert.Contains(t, openai.Models[0].Capabilities, "chat")

	// Verify ensemble
	assert.Equal(t, "majority", cfg.Ensemble.Strategy)
	assert.Equal(t, 3, cfg.Ensemble.MinProviders)
	assert.Equal(t, 7, cfg.Ensemble.MaxProviders)
	assert.Equal(t, 0.85, cfg.Ensemble.ConfidenceThreshold)
	assert.False(t, cfg.Ensemble.FallbackToBest)
	assert.Equal(t, 90*time.Second, cfg.Ensemble.Timeout)

	// Verify openai_compatible
	assert.True(t, cfg.OpenAI.Enabled)
	assert.False(t, cfg.OpenAI.ExposeAllModels)
	assert.Equal(t, "custom-ensemble", cfg.OpenAI.EnsembleModelName)
	assert.True(t, cfg.OpenAI.EnableStreaming)
	assert.True(t, cfg.OpenAI.EnableFunctions)
	assert.False(t, cfg.OpenAI.EnableTools)

	// Verify mcp
	assert.False(t, cfg.MCP.Enabled)
	assert.False(t, cfg.MCP.ExposeAllTools)
	assert.True(t, cfg.MCP.UnifiedToolNamespace)

	// Verify lsp
	assert.True(t, cfg.LSP.Enabled)
	assert.False(t, cfg.LSP.ExposeAllCapabilities)

	// Verify memory
	assert.True(t, cfg.Memory.Enabled)
	assert.Equal(t, "mem0", cfg.Memory.Provider)
	assert.Equal(t, 64000, cfg.Memory.MaxContextLength)
	assert.Equal(t, 60, cfg.Memory.RetentionDays)

	// Verify logging
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.True(t, cfg.Logging.EnableRequestLogging)
	assert.True(t, cfg.Logging.EnableResponseLogging)

	// Verify metrics
	assert.True(t, cfg.Metrics.Enabled)
	assert.Equal(t, 9091, cfg.Metrics.Prometheus.Port)
	assert.Equal(t, "/custom-metrics", cfg.Metrics.Prometheus.Path)
}

// ---- Default values ----

func TestMultiProviderConfig_DefaultValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_defaults_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "defaults.yaml")
	// Minimal YAML with only required server section
	err = os.WriteFile(configPath, []byte("server:\n  port: 8080\n"), 0600)
	require.NoError(t, err)

	cfg, loadErr := LoadMultiProviderConfig(configPath)
	require.NoError(t, loadErr)

	// Optional sections should be nil
	assert.Nil(t, cfg.Ensemble)
	assert.Nil(t, cfg.OpenAI)
	assert.Nil(t, cfg.MCP)
	assert.Nil(t, cfg.LSP)
	assert.Nil(t, cfg.Memory)
	assert.Nil(t, cfg.Logging)
	assert.Nil(t, cfg.Metrics)
	assert.Nil(t, cfg.Providers)
}

// ---- Multiple providers ----

func TestLoadMultiProviderConfig_MultipleProviders(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_multi_prov_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "multi.yaml")
	yamlContent := `
server:
  port: 8080
database:
  host: "localhost"
redis:
  host: "localhost"
providers:
  claude:
    name: "claude"
    enabled: true
    api_key: "claude-key"
    weight: 1.5
  deepseek:
    name: "deepseek"
    enabled: true
    api_key: "deepseek-key"
    weight: 1.2
  gemini:
    name: "gemini"
    enabled: false
    api_key: "gemini-key"
    weight: 1.0
  openrouter:
    name: "openrouter"
    enabled: true
    api_key: "or-key"
    weight: 0.8
`

	err = os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

	cfg, loadErr := LoadMultiProviderConfig(configPath)
	require.NoError(t, loadErr)
	assert.Len(t, cfg.Providers, 4)

	assert.True(t, cfg.Providers["claude"].Enabled)
	assert.Equal(t, 1.5, cfg.Providers["claude"].Weight)
	assert.True(t, cfg.Providers["deepseek"].Enabled)
	assert.False(t, cfg.Providers["gemini"].Enabled)
	assert.True(t, cfg.Providers["openrouter"].Enabled)
}
