package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// ---- NewAIDebateConfigLoader ----

func TestNewAIDebateConfigLoader_WithPath(t *testing.T) {
	loader := NewAIDebateConfigLoader("/some/path/config.yaml")
	assert.NotNil(t, loader)
	assert.Equal(t, "/some/path/config.yaml", loader.configPath)
	assert.Nil(t, loader.config)
}

func TestNewAIDebateConfigLoader_EmptyPath(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	assert.NotNil(t, loader)
	assert.Equal(t, "", loader.configPath)
	assert.Nil(t, loader.config)
}

// ---- Load ----

func TestAIDebateConfigLoader_Load_EmptyPath(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	cfg, err := loader.Load()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration path is required")
}

func TestAIDebateConfigLoader_Load_NonexistentFile(t *testing.T) {
	loader := NewAIDebateConfigLoader("/nonexistent/path/config.yaml")
	cfg, err := loader.Load()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file does not exist")
}

func TestAIDebateConfigLoader_Load_InvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_invalid_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid yaml content: [unclosed")
	require.NoError(t, err)
	tmpFile.Close()

	loader := NewAIDebateConfigLoader(tmpFile.Name())
	cfg, loadErr := loader.Load()
	assert.Nil(t, cfg)
	require.Error(t, loadErr)
	assert.Contains(t, loadErr.Error(), "failed to parse YAML")
}

func TestAIDebateConfigLoader_Load_ValidationFailure(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_invalid_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Enabled but only 1 participant => validation fails
	yamlContent := `
enabled: true
maximal_repeat_rounds: 3
debate_timeout: 300000
consensus_threshold: 0.75
quality_threshold: 0.7
max_response_time: 30000
max_context_length: 32000
debate_strategy: "structured"
voting_strategy: "confidence_weighted"
participants:
  - name: "OnlyOne"
    role: "Analyst"
    enabled: true
    llms:
      - name: "LLM1"
        provider: "claude"
        model: "claude-3"
        enabled: true
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
    response_timeout: 30000
    weight: 1.0
    priority: 1
    debate_style: "analytical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
`
	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	loader := NewAIDebateConfigLoader(tmpFile.Name())
	cfg, loadErr := loader.Load()
	assert.Nil(t, cfg)
	require.Error(t, loadErr)
	assert.Contains(t, loadErr.Error(), "configuration validation failed")
}

func TestAIDebateConfigLoader_Load_ValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_valid_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	yamlContent := validTwoParticipantYAML()
	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	loader := NewAIDebateConfigLoader(tmpFile.Name())
	cfg, loadErr := loader.Load()
	require.NoError(t, loadErr)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 3, cfg.MaximalRepeatRounds)
	assert.Len(t, cfg.Participants, 2)

	// Verify config is stored internally
	assert.Equal(t, cfg, loader.GetConfig())
}

func TestAIDebateConfigLoader_Load_UnreadableFile(t *testing.T) {
	// Create a directory to simulate unreadable path
	tmpDir, err := os.MkdirTemp("", "test_unreadable_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Trying to read a directory as a file should fail
	loader := NewAIDebateConfigLoader(tmpDir)
	cfg, loadErr := loader.Load()
	assert.Nil(t, cfg)
	require.Error(t, loadErr)
	assert.Contains(t, loadErr.Error(), "failed to read configuration file")
}

// ---- LoadFromString ----

func TestAIDebateConfigLoader_LoadFromString_ValidConfig(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	cfg, err := loader.LoadFromString(validTwoParticipantYAML())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Len(t, cfg.Participants, 2)
}

func TestAIDebateConfigLoader_LoadFromString_InvalidYAML(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	cfg, err := loader.LoadFromString("invalid yaml: [broken")
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestAIDebateConfigLoader_LoadFromString_ValidationFails(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	yamlContent := `
enabled: true
maximal_repeat_rounds: 0
debate_timeout: 300000
consensus_threshold: 0.75
quality_threshold: 0.7
max_response_time: 30000
max_context_length: 32000
participants:
  - name: "P1"
    role: "Analyst"
    enabled: true
    llms:
      - name: "LLM1"
        provider: "claude"
        model: "claude-3"
        enabled: true
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
    response_timeout: 30000
    weight: 1.0
    priority: 1
    debate_style: "analytical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
`
	cfg, err := loader.LoadFromString(yamlContent)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestAIDebateConfigLoader_LoadFromString_StoresConfig(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	assert.Nil(t, loader.GetConfig())

	cfg, err := loader.LoadFromString(validTwoParticipantYAML())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	stored := loader.GetConfig()
	assert.Equal(t, cfg, stored)
}

// ---- GetConfig ----

func TestAIDebateConfigLoader_GetConfig_NilBeforeLoad(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	assert.Nil(t, loader.GetConfig())
}

func TestAIDebateConfigLoader_GetConfig_AfterLoad(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	_, err := loader.LoadFromString(validTwoParticipantYAML())
	require.NoError(t, err)
	assert.NotNil(t, loader.GetConfig())
}

// ---- Reload ----

func TestAIDebateConfigLoader_Reload_ReloadsFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_reload_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(validTwoParticipantYAML())
	require.NoError(t, err)
	tmpFile.Close()

	loader := NewAIDebateConfigLoader(tmpFile.Name())

	cfg1, err := loader.Load()
	require.NoError(t, err)
	assert.Equal(t, 3, cfg1.MaximalRepeatRounds)

	// Modify file and reload
	newYAML := strings.Replace(validTwoParticipantYAML(), "maximal_repeat_rounds: 3", "maximal_repeat_rounds: 5", 1)
	err = os.WriteFile(tmpFile.Name(), []byte(newYAML), 0600)
	require.NoError(t, err)

	cfg2, err := loader.Reload()
	require.NoError(t, err)
	assert.Equal(t, 5, cfg2.MaximalRepeatRounds)
}

func TestAIDebateConfigLoader_Reload_EmptyPathError(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	cfg, err := loader.Reload()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration path is required")
}

// ---- Save ----

func TestAIDebateConfigLoader_Save_EmptyPath(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	cfg := GetDefaultConfig()
	err := loader.Save(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration path is required")
}

func TestAIDebateConfigLoader_Save_InvalidConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_save_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "out.yaml")
	loader := NewAIDebateConfigLoader(outPath)

	// Config with only 1 participant should fail validation
	invalidCfg := &AIDebateConfig{
		Enabled:             true,
		MaximalRepeatRounds: 3,
		DebateTimeout:       300000,
		ConsensusThreshold:  0.75,
		QualityThreshold:    0.7,
		MaxResponseTime:     30000,
		MaxContextLength:    32000,
		DebateStrategy:      "structured",
		VotingStrategy:      "confidence_weighted",
		Participants: []DebateParticipant{
			{
				Name:               "P1",
				Role:               "Analyst",
				Enabled:            true,
				ResponseTimeout:    30000,
				Weight:             1.0,
				Priority:           1,
				DebateStyle:        "analytical",
				ArgumentationStyle: "logical",
				PersuasionLevel:    0.5,
				OpennessToChange:   0.5,
				QualityThreshold:   0.7,
				MinResponseLength:  50,
				MaxResponseLength:  1000,
				LLMs: []LLMConfiguration{
					{
						Name: "LLM1", Provider: "claude", Model: "claude-3",
						Enabled: true, Timeout: 30000, MaxTokens: 1000, Temperature: 0.7,
					},
				},
			},
		},
	}

	err = loader.Save(invalidCfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestAIDebateConfigLoader_Save_ValidConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_save_valid_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "subdir", "out.yaml")
	loader := NewAIDebateConfigLoader(outPath)

	cfg := GetDefaultConfig()
	err = loader.Save(cfg)
	require.NoError(t, err)

	// Verify file was created
	_, statErr := os.Stat(outPath)
	assert.NoError(t, statErr)

	// Verify config was stored on the loader
	assert.Equal(t, cfg, loader.config)

	// Verify we can reload the saved config
	loader2 := NewAIDebateConfigLoader(outPath)
	cfg2, err := loader2.Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.MaximalRepeatRounds, cfg2.MaximalRepeatRounds)
	assert.Equal(t, cfg.DebateStrategy, cfg2.DebateStrategy)
	assert.Len(t, cfg2.Participants, len(cfg.Participants))
}

func TestAIDebateConfigLoader_Save_CreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_save_mkdir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	deepPath := filepath.Join(tmpDir, "a", "b", "c", "config.yaml")
	loader := NewAIDebateConfigLoader(deepPath)

	cfg := GetDefaultConfig()
	err = loader.Save(cfg)
	require.NoError(t, err)

	_, statErr := os.Stat(deepPath)
	assert.NoError(t, statErr)
}

// ---- substituteEnvVars ----

func TestAIDebateConfigLoader_SubstituteEnvVars_FullCoverage(t *testing.T) {
	os.Setenv("TEST_LOADER_KEY", "substituted-key")
	os.Setenv("TEST_LOADER_URL", "https://api.example.com")
	os.Setenv("TEST_LOADER_HEALTH", "https://health.example.com")
	os.Setenv("TEST_LOADER_DATASET", "my-dataset")
	os.Setenv("TEST_LOADER_CUSTOM", "custom-value")
	defer func() {
		os.Unsetenv("TEST_LOADER_KEY")
		os.Unsetenv("TEST_LOADER_URL")
		os.Unsetenv("TEST_LOADER_HEALTH")
		os.Unsetenv("TEST_LOADER_DATASET")
		os.Unsetenv("TEST_LOADER_CUSTOM")
	}()

	yamlContent := `
enabled: true
maximal_repeat_rounds: 3
debate_timeout: 300000
consensus_threshold: 0.75
quality_threshold: 0.7
max_response_time: 30000
max_context_length: 32000
debate_strategy: "structured"
voting_strategy: "confidence_weighted"
enable_cognee: true
cognee_config:
  enabled: true
  dataset_name: "${TEST_LOADER_DATASET}"
  max_enhancement_time: 10000
  enhancement_strategy: "hybrid"
  enhance_responses: true
  analyze_consensus: true
  generate_insights: true
participants:
  - name: "P1"
    role: "Analyst"
    enabled: true
    llms:
      - name: "LLM1"
        provider: "claude"
        model: "claude-3"
        enabled: true
        api_key: "${TEST_LOADER_KEY}"
        base_url: "${TEST_LOADER_URL}"
        health_check_url: "${TEST_LOADER_HEALTH}"
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
        custom_params:
          proxy: "${TEST_LOADER_CUSTOM}"
          count: 42
    response_timeout: 30000
    weight: 1.0
    priority: 1
    debate_style: "analytical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
    enable_cognee: true
    cognee_settings:
      enhance_responses: true
      analyze_sentiment: true
      extract_entities: true
      generate_summary: true
      dataset_name: "${TEST_LOADER_DATASET}_participant"
  - name: "P2"
    role: "Critic"
    enabled: true
    llms:
      - name: "LLM2"
        provider: "deepseek"
        model: "deepseek-coder"
        enabled: true
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
    response_timeout: 30000
    weight: 1.0
    priority: 2
    debate_style: "critical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
`

	loader := NewAIDebateConfigLoader("")
	cfg, err := loader.LoadFromString(yamlContent)
	require.NoError(t, err)

	// Cognee dataset name substituted
	assert.Equal(t, "my-dataset", cfg.CogneeConfig.DatasetName)

	// LLM fields substituted
	llm := cfg.Participants[0].LLMs[0]
	assert.Equal(t, "substituted-key", llm.APIKey)
	assert.Equal(t, "https://api.example.com", llm.BaseURL)
	assert.Equal(t, "https://health.example.com", llm.HealthCheckURL)

	// Custom params: string substituted, non-string untouched
	assert.Equal(t, "custom-value", llm.CustomParams["proxy"])

	// Cognee participant dataset substituted
	assert.Equal(t, "my-dataset_participant", cfg.Participants[0].CogneeSettings.DatasetName)
}

func TestAIDebateConfigLoader_SubstituteEnvVars_NilCognee(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	cfg := &AIDebateConfig{
		CogneeConfig: nil,
		Participants: []DebateParticipant{
			{
				LLMs:           []LLMConfiguration{},
				CogneeSettings: nil,
			},
		},
	}
	err := loader.substituteEnvVars(cfg)
	assert.NoError(t, err)
}

// ---- applyDefaults ----

func TestAIDebateConfigLoader_ApplyDefaults_ZeroValues(t *testing.T) {
	loader := NewAIDebateConfigLoader("")

	cfg := &AIDebateConfig{
		Participants: []DebateParticipant{
			{
				Name: "P1",
				LLMs: []LLMConfiguration{
					{Name: "LLM1"},
				},
			},
		},
	}

	loader.applyDefaults(cfg)

	// Global defaults
	assert.Equal(t, 3, cfg.MaximalRepeatRounds)
	assert.Equal(t, 300000, cfg.DebateTimeout)
	assert.Equal(t, 0.75, cfg.ConsensusThreshold)
	assert.Equal(t, 0.7, cfg.QualityThreshold)
	assert.Equal(t, 30000, cfg.MaxResponseTime)
	assert.Equal(t, 32000, cfg.MaxContextLength)
	assert.Equal(t, 2592000000, cfg.MemoryRetention)
	assert.Equal(t, "structured", cfg.DebateStrategy)
	assert.Equal(t, "confidence_weighted", cfg.VotingStrategy)
	assert.Equal(t, "detailed", cfg.ResponseFormat)

	// Cognee defaults
	assert.True(t, cfg.EnableCognee)
	require.NotNil(t, cfg.CogneeConfig)
	assert.True(t, cfg.CogneeConfig.Enabled)
	assert.Equal(t, "ai_debate_enhancement", cfg.CogneeConfig.DatasetName)
	assert.Equal(t, "hybrid", cfg.CogneeConfig.EnhancementStrategy)

	// Participant defaults
	p := cfg.Participants[0]
	assert.Equal(t, 30000, p.ResponseTimeout)
	assert.Equal(t, 1.0, p.Weight)
	assert.Equal(t, 1, p.Priority)
	assert.Equal(t, 0.7, p.QualityThreshold)
	assert.Equal(t, 50, p.MinResponseLength)
	assert.Equal(t, 2000, p.MaxResponseLength)
	assert.Equal(t, "balanced", p.DebateStyle)
	assert.Equal(t, "logical", p.ArgumentationStyle)
	assert.Equal(t, 0.5, p.PersuasionLevel)
	assert.Equal(t, 0.5, p.OpennessToChange)
	assert.True(t, p.EnableCognee)
	require.NotNil(t, p.CogneeSettings)
	assert.Equal(t, "participant_p1_enhancement", p.CogneeSettings.DatasetName)

	// LLM defaults
	llm := p.LLMs[0]
	assert.Equal(t, 30000, llm.Timeout)
	assert.Equal(t, 3, llm.MaxRetries)
	assert.Equal(t, 0.7, llm.Temperature)
	assert.Equal(t, 1000, llm.MaxTokens)
	assert.Equal(t, 1.0, llm.Weight)
	assert.Equal(t, 10, llm.RateLimitRPS)
	assert.Equal(t, 30000, llm.RequestTimeout)
	assert.Equal(t, 60000, llm.HealthCheckInterval)
}

func TestAIDebateConfigLoader_ApplyDefaults_PreservesExistingValues(t *testing.T) {
	loader := NewAIDebateConfigLoader("")

	cfg := &AIDebateConfig{
		MaximalRepeatRounds: 5,
		DebateTimeout:       600000,
		ConsensusThreshold:  0.9,
		QualityThreshold:    0.8,
		MaxResponseTime:     60000,
		MaxContextLength:    64000,
		MemoryRetention:     1000000,
		DebateStrategy:      "adversarial",
		VotingStrategy:      "majority",
		ResponseFormat:      "concise",
		EnableCognee:        true,
		CogneeConfig: &CogneeDebateConfig{
			Enabled:     true,
			DatasetName: "custom",
		},
		Participants: []DebateParticipant{
			{
				Name:               "P1",
				ResponseTimeout:    60000,
				Weight:             2.0,
				Priority:           3,
				QualityThreshold:   0.9,
				MinResponseLength:  100,
				MaxResponseLength:  5000,
				DebateStyle:        "creative",
				ArgumentationStyle: "socratic",
				PersuasionLevel:    0.8,
				OpennessToChange:   0.9,
				EnableCognee:       true,
				CogneeSettings: &CogneeParticipantConfig{
					DatasetName: "existing",
				},
				LLMs: []LLMConfiguration{
					{
						Name:                "LLM1",
						Timeout:             60000,
						MaxRetries:          5,
						Temperature:         0.3,
						MaxTokens:           4000,
						Weight:              2.0,
						RateLimitRPS:        20,
						RequestTimeout:      50000,
						HealthCheckInterval: 120000,
					},
				},
			},
		},
	}

	loader.applyDefaults(cfg)

	// All values should be preserved
	assert.Equal(t, 5, cfg.MaximalRepeatRounds)
	assert.Equal(t, 600000, cfg.DebateTimeout)
	assert.Equal(t, 0.9, cfg.ConsensusThreshold)
	assert.Equal(t, 0.8, cfg.QualityThreshold)
	assert.Equal(t, 60000, cfg.MaxResponseTime)
	assert.Equal(t, 64000, cfg.MaxContextLength)
	assert.Equal(t, 1000000, cfg.MemoryRetention)
	assert.Equal(t, "adversarial", cfg.DebateStrategy)
	assert.Equal(t, "majority", cfg.VotingStrategy)
	assert.Equal(t, "concise", cfg.ResponseFormat)

	p := cfg.Participants[0]
	assert.Equal(t, 60000, p.ResponseTimeout)
	assert.Equal(t, 2.0, p.Weight)
	assert.Equal(t, 3, p.Priority)
	assert.Equal(t, 0.9, p.QualityThreshold)
	assert.Equal(t, 100, p.MinResponseLength)
	assert.Equal(t, 5000, p.MaxResponseLength)
	assert.Equal(t, "creative", p.DebateStyle)
	assert.Equal(t, "socratic", p.ArgumentationStyle)
	assert.Equal(t, 0.8, p.PersuasionLevel)
	assert.Equal(t, 0.9, p.OpennessToChange)
	assert.Equal(t, "existing", p.CogneeSettings.DatasetName)

	llm := p.LLMs[0]
	assert.Equal(t, 60000, llm.Timeout)
	assert.Equal(t, 5, llm.MaxRetries)
	assert.Equal(t, 0.3, llm.Temperature)
	assert.Equal(t, 4000, llm.MaxTokens)
	assert.Equal(t, 2.0, llm.Weight)
	assert.Equal(t, 20, llm.RateLimitRPS)
	assert.Equal(t, 50000, llm.RequestTimeout)
	assert.Equal(t, 120000, llm.HealthCheckInterval)
}

func TestAIDebateConfigLoader_ApplyDefaults_RequestTimeoutFromTimeout(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	cfg := &AIDebateConfig{
		Participants: []DebateParticipant{
			{
				LLMs: []LLMConfiguration{
					{
						Name:    "LLM1",
						Timeout: 45000,
						// RequestTimeout is 0 => should be set to Timeout
					},
				},
			},
		},
	}

	loader.applyDefaults(cfg)
	assert.Equal(t, 45000, cfg.Participants[0].LLMs[0].RequestTimeout)
}

// ---- GetDefaultConfig ----

func TestGetDefaultConfig_Structure(t *testing.T) {
	cfg := GetDefaultConfig()
	require.NotNil(t, cfg)

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 3, cfg.MaximalRepeatRounds)
	assert.Equal(t, 300000, cfg.DebateTimeout)
	assert.Equal(t, 0.75, cfg.ConsensusThreshold)
	assert.True(t, cfg.EnableCognee)
	assert.True(t, cfg.EnableMemory)
	assert.True(t, cfg.EnableDebateLogging)
	assert.True(t, cfg.LogDebateDetails)
	assert.True(t, cfg.MetricsEnabled)
	assert.False(t, cfg.EnableStreaming)
	assert.Equal(t, "structured", cfg.DebateStrategy)
	assert.Equal(t, "confidence_weighted", cfg.VotingStrategy)
	assert.Equal(t, "detailed", cfg.ResponseFormat)
	assert.Equal(t, 32000, cfg.MaxContextLength)
	assert.Equal(t, 0.7, cfg.QualityThreshold)
	assert.Equal(t, 30*1000, cfg.MaxResponseTime)
	assert.Equal(t, 30*24*60*60*1000, cfg.MemoryRetention)
}

func TestGetDefaultConfig_CogneeConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	require.NotNil(t, cfg.CogneeConfig)

	cc := cfg.CogneeConfig
	assert.True(t, cc.Enabled)
	assert.True(t, cc.EnhanceResponses)
	assert.True(t, cc.AnalyzeConsensus)
	assert.True(t, cc.GenerateInsights)
	assert.Equal(t, "ai_debate_enhancement", cc.DatasetName)
	assert.Equal(t, 10*1000, cc.MaxEnhancementTime)
	assert.Equal(t, "hybrid", cc.EnhancementStrategy)
	assert.True(t, cc.MemoryIntegration)
	assert.True(t, cc.ContextualAnalysis)
}

func TestGetDefaultConfig_Participants(t *testing.T) {
	cfg := GetDefaultConfig()
	require.Len(t, cfg.Participants, 3)

	tests := []struct {
		name          string
		expectedName  string
		expectedRole  string
		expectedLLMs  int
		expectedStyle string
	}{
		{"Strongest", "Strongest", "Primary Analyst", 3, "analytical"},
		{"MiddleOne", "Middle One", "Balanced Analyst", 2, "balanced"},
		{"Creative", "Creative Thinker", "Creative Analyst", 2, "creative"},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := cfg.Participants[i]
			assert.Equal(t, tt.expectedName, p.Name)
			assert.Equal(t, tt.expectedRole, p.Role)
			assert.True(t, p.Enabled)
			assert.Len(t, p.LLMs, tt.expectedLLMs)
			assert.Equal(t, tt.expectedStyle, p.DebateStyle)
			assert.True(t, p.EnableCognee)

			primary := p.GetPrimaryLLM()
			require.NotNil(t, primary)
			assert.True(t, primary.Enabled)
		})
	}
}

func TestGetDefaultConfig_ParticipantMaxRounds(t *testing.T) {
	cfg := GetDefaultConfig()

	// Only "Middle One" has participant-specific max rounds
	assert.Nil(t, cfg.Participants[0].MaximalRepeatRounds)
	require.NotNil(t, cfg.Participants[1].MaximalRepeatRounds)
	assert.Equal(t, 2, *cfg.Participants[1].MaximalRepeatRounds)
	assert.Nil(t, cfg.Participants[2].MaximalRepeatRounds)
}

func TestGetDefaultConfig_Validates(t *testing.T) {
	cfg := GetDefaultConfig()
	err := cfg.Validate()
	assert.NoError(t, err)
}

// ---- intPtr helper ----

func TestIntPtr(t *testing.T) {
	p := intPtr(42)
	require.NotNil(t, p)
	assert.Equal(t, 42, *p)
}

func TestIntPtr_Zero(t *testing.T) {
	p := intPtr(0)
	require.NotNil(t, p)
	assert.Equal(t, 0, *p)
}

func TestIntPtr_Negative(t *testing.T) {
	p := intPtr(-5)
	require.NotNil(t, p)
	assert.Equal(t, -5, *p)
}

// ---- Concurrent access ----

func TestAIDebateConfigLoader_ConcurrentLoadAndGetConfig(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	done := make(chan bool, 20)

	yamlContent := validTwoParticipantYAML()

	for i := 0; i < 10; i++ {
		go func() {
			_, _ = loader.LoadFromString(yamlContent)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		go func() {
			_ = loader.GetConfig()
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

// ---- Round-trip: Load -> Save -> Load ----

func TestAIDebateConfigLoader_RoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_roundtrip_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "roundtrip.yaml")

	original := GetDefaultConfig()

	saver := NewAIDebateConfigLoader(outPath)
	err = saver.Save(original)
	require.NoError(t, err)

	reloader := NewAIDebateConfigLoader(outPath)
	loaded, err := reloader.Load()
	require.NoError(t, err)

	assert.Equal(t, original.Enabled, loaded.Enabled)
	assert.Equal(t, original.MaximalRepeatRounds, loaded.MaximalRepeatRounds)
	assert.Equal(t, original.DebateTimeout, loaded.DebateTimeout)
	assert.Equal(t, original.DebateStrategy, loaded.DebateStrategy)
	assert.Equal(t, original.VotingStrategy, loaded.VotingStrategy)
	assert.Len(t, loaded.Participants, len(original.Participants))
}

// ---- Disabled config skips validation ----

func TestAIDebateConfigLoader_LoadFromString_DisabledSkipsValidation(t *testing.T) {
	loader := NewAIDebateConfigLoader("")
	yamlContent := `
enabled: false
maximal_repeat_rounds: 0
`
	cfg, err := loader.LoadFromString(yamlContent)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.False(t, cfg.Enabled)
}

// ---- helper ----

func validTwoParticipantYAML() string {
	return `
enabled: true
maximal_repeat_rounds: 3
debate_timeout: 300000
consensus_threshold: 0.75
quality_threshold: 0.7
max_response_time: 30000
max_context_length: 32000
debate_strategy: "structured"
voting_strategy: "confidence_weighted"
participants:
  - name: "P1"
    role: "Analyst"
    enabled: true
    llms:
      - name: "LLM1"
        provider: "claude"
        model: "claude-3-sonnet"
        enabled: true
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
    response_timeout: 30000
    weight: 1.0
    priority: 1
    debate_style: "analytical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
  - name: "P2"
    role: "Critic"
    enabled: true
    llms:
      - name: "LLM2"
        provider: "deepseek"
        model: "deepseek-coder"
        enabled: true
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
    response_timeout: 30000
    weight: 1.0
    priority: 2
    debate_style: "critical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
`
}
