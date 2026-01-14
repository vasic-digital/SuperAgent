// Package verifier provides integration with LLMsVerifier for model verification,
// scoring, and health monitoring capabilities.
package verifier

import (
	"time"

	"dev.helix.agent/internal/llm"
)

// ProviderAuthType defines how a provider authenticates
type ProviderAuthType string

const (
	// AuthTypeAPIKey is for providers that use API key authentication (Bearer token)
	AuthTypeAPIKey ProviderAuthType = "api_key"

	// AuthTypeOAuth is for providers that use OAuth2 authentication (Claude, Qwen)
	AuthTypeOAuth ProviderAuthType = "oauth"

	// AuthTypeFree is for providers that offer free models without authentication
	AuthTypeFree ProviderAuthType = "free"

	// AuthTypeAnonymous is for providers that work with anonymous access (device ID)
	AuthTypeAnonymous ProviderAuthType = "anonymous"

	// AuthTypeLocal is for self-hosted providers (Ollama)
	AuthTypeLocal ProviderAuthType = "local"
)

// ProviderStatus represents the health status of a provider
type ProviderStatus string

const (
	// StatusUnknown indicates the provider has not been verified yet
	StatusUnknown ProviderStatus = "unknown"

	// StatusHealthy indicates the provider is verified and working
	StatusHealthy ProviderStatus = "healthy"

	// StatusVerified indicates the provider passed verification (alias for Healthy)
	StatusVerified ProviderStatus = "verified"

	// StatusUnhealthy indicates the provider failed verification
	StatusUnhealthy ProviderStatus = "unhealthy"

	// StatusFailed indicates verification failed for this provider
	StatusFailed ProviderStatus = "failed"

	// StatusDegraded indicates partial verification success
	StatusDegraded ProviderStatus = "degraded"

	// StatusRateLimited indicates the provider is temporarily unavailable due to rate limits
	StatusRateLimited ProviderStatus = "rate_limited"

	// StatusAuthFailed indicates authentication failed for this provider
	StatusAuthFailed ProviderStatus = "auth_failed"

	// StatusUnavailable indicates the provider is not accessible
	StatusUnavailable ProviderStatus = "unavailable"
)

// UnifiedProvider represents a provider with all verification data
// This is the single source of truth for provider information
type UnifiedProvider struct {
	// Identity
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	DisplayName  string           `json:"display_name,omitempty"`
	Type         string           `json:"type"`
	AuthType     ProviderAuthType `json:"auth_type"`

	// Verification Results
	Verified     bool             `json:"verified"`
	VerifiedAt   time.Time        `json:"verified_at,omitempty"`
	Score        float64          `json:"score"`
	ScoreSuffix  string           `json:"score_suffix,omitempty"`
	TestResults  map[string]bool  `json:"test_results,omitempty"`
	CodeVisible  bool             `json:"code_visible"`

	// Models Available
	Models       []UnifiedModel   `json:"models"`
	DefaultModel string           `json:"default_model"`
	PrimaryModel *UnifiedModel    `json:"primary_model,omitempty"`

	// Health Status
	Status       ProviderStatus   `json:"status"`
	LastHealthCheck time.Time    `json:"last_health_check,omitempty"`
	HealthCheckError string      `json:"health_check_error,omitempty"`

	// OAuth-specific (only for AuthTypeOAuth)
	OAuthTokenExpiry  time.Time   `json:"oauth_token_expiry,omitempty"`
	OAuthAutoRefresh  bool        `json:"oauth_auto_refresh,omitempty"`

	// Configuration
	BaseURL      string           `json:"base_url,omitempty"`
	APIKey       string           `json:"-"` // Not serialized for security
	Tier         int              `json:"tier"`
	Priority     int              `json:"priority"`

	// Provider Instance (not serialized)
	Instance     llm.LLMProvider  `json:"-"`

	// Error tracking
	ErrorMessage     string       `json:"error_message,omitempty"`
	ConsecutiveFails int          `json:"consecutive_fails,omitempty"`
	ErrorCount       int          `json:"error_count,omitempty"`

	// Health tracking (aliases for compatibility)
	LastHealthAt     time.Time    `json:"last_health_at,omitempty"`

	// Metadata for additional information
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// UnifiedModel represents a model with its verification score
type UnifiedModel struct {
	// Identity
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name,omitempty"`
	Provider    string  `json:"provider"`

	// Verification
	Score       float64       `json:"score"`
	ScoreSuffix string        `json:"score_suffix,omitempty"`
	Verified    bool          `json:"verified"`
	VerifiedAt  time.Time     `json:"verified_at,omitempty"`
	Latency     time.Duration `json:"latency,omitempty"`

	// Capabilities (multiple representations for flexibility)
	ContextWindow     int      `json:"context_window,omitempty"`
	MaxOutputTokens   int      `json:"max_output_tokens,omitempty"`
	SupportsStreaming bool     `json:"supports_streaming,omitempty"`
	SupportsTools     bool     `json:"supports_tools,omitempty"`
	SupportsFunctions bool     `json:"supports_functions,omitempty"`
	SupportsVision    bool     `json:"supports_vision,omitempty"`
	Capabilities      []string `json:"capabilities,omitempty"` // String list of capabilities

	// Pricing (per 1M tokens)
	CostPerInputToken  float64 `json:"cost_per_input_token,omitempty"`
	CostPerOutputToken float64 `json:"cost_per_output_token,omitempty"`

	// Test Results
	TestResults map[string]bool `json:"test_results,omitempty"`

	// Metadata for additional information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StartupResult contains the results of the startup verification pipeline
type StartupResult struct {
	// Counts
	TotalProviders   int `json:"total_providers"`
	VerifiedCount    int `json:"verified_count"`
	FailedCount      int `json:"failed_count"`
	SkippedCount     int `json:"skipped_count"`

	// Provider breakdown by auth type
	APIKeyProviders  int `json:"api_key_providers"`
	OAuthProviders   int `json:"oauth_providers"`
	FreeProviders    int `json:"free_providers"`

	// Timing
	StartedAt        time.Time     `json:"started_at"`
	CompletedAt      time.Time     `json:"completed_at"`
	DurationMs       int64         `json:"duration_ms"`

	// Results
	Providers        []*UnifiedProvider `json:"providers"`
	RankedProviders  []*UnifiedProvider `json:"ranked_providers"`

	// Debate Team (15 LLMs)
	DebateTeam       *DebateTeamResult `json:"debate_team,omitempty"`

	// Errors
	Errors           []StartupError `json:"errors,omitempty"`
}

// DebateTeamResult contains the selected 15 LLMs for the AI debate team
type DebateTeamResult struct {
	// Team composition (5 positions x 3 LLMs)
	Positions    []*DebatePosition `json:"positions"`
	TotalLLMs    int               `json:"total_llms"`

	// Selection criteria
	MinScore     float64           `json:"min_score"`
	OAuthFirst   bool              `json:"oauth_first"`

	// Timing
	SelectedAt   time.Time         `json:"selected_at"`
}

// DebatePosition represents a position in the AI debate team
type DebatePosition struct {
	Position     int              `json:"position"`
	Role         string           `json:"role"`
	Primary      *DebateLLM       `json:"primary"`
	Fallback1    *DebateLLM       `json:"fallback_1,omitempty"`
	Fallback2    *DebateLLM       `json:"fallback_2,omitempty"`
}

// DebateLLM represents an LLM selected for the debate team
type DebateLLM struct {
	Provider     string           `json:"provider"`
	ProviderType string           `json:"provider_type"`
	ModelID      string           `json:"model_id"`
	ModelName    string           `json:"model_name"`
	AuthType     ProviderAuthType `json:"auth_type"`
	Score        float64          `json:"score"`
	Verified     bool             `json:"verified"`
	IsOAuth      bool             `json:"is_oauth"`
}

// StartupError represents an error during startup verification
type StartupError struct {
	Provider    string `json:"provider"`
	ModelID     string `json:"model_id,omitempty"`
	Phase       string `json:"phase"`
	Error       string `json:"error"`
	Recoverable bool   `json:"recoverable"`
}

// ProviderDiscoveryResult contains information about a discovered provider
type ProviderDiscoveryResult struct {
	ID          string           `json:"id"`
	Type        string           `json:"type"`
	AuthType    ProviderAuthType `json:"auth_type"`
	Discovered  bool             `json:"discovered"`
	Source      string           `json:"source"` // "env", "oauth", "auto"
	Credentials string           `json:"credentials,omitempty"` // Redacted
	BaseURL     string           `json:"base_url,omitempty"`
	Models      []string         `json:"models,omitempty"`
}

// StartupConfig holds configuration for the startup verification pipeline
type StartupConfig struct {
	// Verification settings
	ParallelVerification  bool          `yaml:"parallel_verification" json:"parallel_verification"`
	MaxConcurrency        int           `yaml:"max_concurrency" json:"max_concurrency"`
	VerificationTimeout   time.Duration `yaml:"verification_timeout" json:"verification_timeout"`
	HealthCheckTimeout    time.Duration `yaml:"health_check_timeout" json:"health_check_timeout"`

	// Score thresholds
	MinScore              float64       `yaml:"min_score" json:"min_score"`
	RequireCodeVisibility bool          `yaml:"require_code_visibility" json:"require_code_visibility"`

	// Debate team settings
	DebateTeamSize        int           `yaml:"debate_team_size" json:"debate_team_size"`
	PositionCount         int           `yaml:"position_count" json:"position_count"`
	FallbacksPerPosition  int           `yaml:"fallbacks_per_position" json:"fallbacks_per_position"`

	// OAuth settings
	OAuthPriorityBoost    float64       `yaml:"oauth_priority_boost" json:"oauth_priority_boost"`
	TrustOAuthOnFailure   bool          `yaml:"trust_oauth_on_failure" json:"trust_oauth_on_failure"`
	OAuthTokenRefreshMins int           `yaml:"oauth_token_refresh_mins" json:"oauth_token_refresh_mins"`

	// Free provider settings
	FreeProviderBaseScore float64       `yaml:"free_provider_base_score" json:"free_provider_base_score"`
	EnableFreeProviders   bool          `yaml:"enable_free_providers" json:"enable_free_providers"`

	// Fallback strategy
	OAuthPrimaryNonOAuthFallback bool `yaml:"oauth_primary_non_oauth_fallback" json:"oauth_primary_non_oauth_fallback"`

	// Caching
	CacheVerificationResults bool          `yaml:"cache_verification_results" json:"cache_verification_results"`
	CacheTTL                 time.Duration `yaml:"cache_ttl" json:"cache_ttl"`
}

// DefaultStartupConfig returns sensible defaults for startup configuration
func DefaultStartupConfig() *StartupConfig {
	return &StartupConfig{
		ParallelVerification:         true,
		MaxConcurrency:               10,
		VerificationTimeout:          30 * time.Second,
		HealthCheckTimeout:           10 * time.Second,
		MinScore:                     5.0,
		RequireCodeVisibility:        true,
		DebateTeamSize:               15,
		PositionCount:                5,
		FallbacksPerPosition:         2,
		OAuthPriorityBoost:           0.5,
		TrustOAuthOnFailure:          true,
		OAuthTokenRefreshMins:        10,
		FreeProviderBaseScore:        6.5,
		EnableFreeProviders:          true,
		OAuthPrimaryNonOAuthFallback: true,
		CacheVerificationResults:     true,
		CacheTTL:                     6 * time.Hour,
	}
}

// ProviderTypeInfo contains metadata about a provider type
type ProviderTypeInfo struct {
	Type        string           `json:"type"`
	DisplayName string           `json:"display_name"`
	AuthType    ProviderAuthType `json:"auth_type"`
	Tier        int              `json:"tier"`
	Priority    int              `json:"priority"`
	EnvVars     []string         `json:"env_vars"`
	BaseURL     string           `json:"base_url"`
	Models      []string         `json:"models"`
	Free        bool             `json:"free"`
}

// SupportedProviders defines all provider types supported by the system
var SupportedProviders = map[string]*ProviderTypeInfo{
	// Tier 1: Premium (OAuth2 + API Key)
	"claude": {
		Type:        "claude",
		DisplayName: "Claude (Anthropic)",
		AuthType:    AuthTypeOAuth, // Primary auth is OAuth, fallback to API key
		Tier:        1,
		Priority:    1,
		EnvVars:     []string{"ANTHROPIC_API_KEY", "CLAUDE_API_KEY"},
		BaseURL:     "https://api.anthropic.com/v1/messages",
		Models:      []string{"claude-opus-4-5-20251101", "claude-sonnet-4-5-20250929", "claude-haiku-4-5-20251001"},
		Free:        false,
	},
	"qwen": {
		Type:        "qwen",
		DisplayName: "Qwen (Alibaba)",
		AuthType:    AuthTypeOAuth,
		Tier:        2,
		Priority:    3,
		EnvVars:     []string{"QWEN_API_KEY", "DASHSCOPE_API_KEY"},
		BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		Models:      []string{"qwen-max", "qwen-plus", "qwen-turbo", "qwen-coder-turbo"},
		Free:        false,
	},

	// Tier 2: High-quality API key providers
	"gemini": {
		Type:        "gemini",
		DisplayName: "Gemini (Google)",
		AuthType:    AuthTypeAPIKey,
		Tier:        2,
		Priority:    2,
		EnvVars:     []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"},
		BaseURL:     "https://generativelanguage.googleapis.com/v1beta",
		Models:      []string{"gemini-2.0-flash", "gemini-1.5-pro", "gemini-1.5-flash"},
		Free:        false,
	},
	"deepseek": {
		Type:        "deepseek",
		DisplayName: "DeepSeek",
		AuthType:    AuthTypeAPIKey,
		Tier:        2,
		Priority:    3,
		EnvVars:     []string{"DEEPSEEK_API_KEY"},
		BaseURL:     "https://api.deepseek.com/v1/chat/completions",
		Models:      []string{"deepseek-chat", "deepseek-coder", "deepseek-reasoner"},
		Free:        false,
	},
	"mistral": {
		Type:        "mistral",
		DisplayName: "Mistral AI",
		AuthType:    AuthTypeAPIKey,
		Tier:        2,
		Priority:    3,
		EnvVars:     []string{"MISTRAL_API_KEY"},
		BaseURL:     "https://api.mistral.ai/v1/chat/completions",
		Models:      []string{"mistral-large-latest", "mistral-medium-latest", "codestral-latest"},
		Free:        false,
	},

	// Tier 3: Fast inference
	"groq": {
		Type:        "groq",
		DisplayName: "Groq",
		AuthType:    AuthTypeAPIKey,
		Tier:        3,
		Priority:    4,
		EnvVars:     []string{"GROQ_API_KEY"},
		BaseURL:     "https://api.groq.com/openai/v1/chat/completions",
		Models:      []string{"llama-3.1-70b-versatile", "llama-3.1-8b-instant", "mixtral-8x7b-32768"},
		Free:        false,
	},
	"cerebras": {
		Type:        "cerebras",
		DisplayName: "Cerebras",
		AuthType:    AuthTypeAPIKey,
		Tier:        3,
		Priority:    4,
		EnvVars:     []string{"CEREBRAS_API_KEY"},
		BaseURL:     "https://api.cerebras.ai/v1/chat/completions",
		Models:      []string{"llama-3.3-70b"},
		Free:        false,
	},

	// Tier 4: Aggregators
	"openrouter": {
		Type:        "openrouter",
		DisplayName: "OpenRouter",
		AuthType:    AuthTypeAPIKey,
		Tier:        4,
		Priority:    5,
		EnvVars:     []string{"OPENROUTER_API_KEY"},
		BaseURL:     "https://openrouter.ai/api/v1/chat/completions",
		Models:      []string{}, // Many models available
		Free:        true,       // Has free tier
	},

	// Tier 5: Free providers
	"zen": {
		Type:        "zen",
		DisplayName: "OpenCode Zen",
		AuthType:    AuthTypeFree,
		Tier:        5,
		Priority:    4,
		EnvVars:     []string{"OPENCODE_API_KEY"}, // Optional
		BaseURL:     "https://opencode.ai/zen/v1/chat/completions",
		Models:      []string{"opencode/grok-code", "opencode/big-pickle", "opencode/glm-4.7-free", "opencode/gpt-5-nano"},
		Free:        true,
	},

	// Tier 6: Self-hosted
	"ollama": {
		Type:        "ollama",
		DisplayName: "Ollama (Local)",
		AuthType:    AuthTypeLocal,
		Tier:        6,
		Priority:    20,
		EnvVars:     []string{"OLLAMA_BASE_URL", "OLLAMA_HOST"},
		BaseURL:     "http://localhost:11434",
		Models:      []string{"llama3.2", "codellama", "mistral"},
		Free:        true,
	},

	// Tier 7: Other API key providers
	"zai": {
		Type:        "zai",
		DisplayName: "ZAI",
		AuthType:    AuthTypeAPIKey,
		Tier:        7,
		Priority:    6,
		EnvVars:     []string{"ZAI_API_KEY"},
		BaseURL:     "https://api.zai.ai/v1/chat/completions",
		Models:      []string{},
		Free:        false,
	},
}

// GetProviderInfo returns info for a provider type
func GetProviderInfo(providerType string) (*ProviderTypeInfo, bool) {
	info, ok := SupportedProviders[providerType]
	return info, ok
}

// IsOAuthProvider checks if a provider type uses OAuth authentication
func IsOAuthProvider(providerType string) bool {
	info, ok := SupportedProviders[providerType]
	if !ok {
		return false
	}
	return info.AuthType == AuthTypeOAuth
}

// IsFreeProvider checks if a provider type is free
func IsFreeProvider(providerType string) bool {
	info, ok := SupportedProviders[providerType]
	if !ok {
		return false
	}
	return info.Free
}

// GetProvidersByAuthType returns all providers of a specific auth type
func GetProvidersByAuthType(authType ProviderAuthType) []*ProviderTypeInfo {
	var result []*ProviderTypeInfo
	for _, info := range SupportedProviders {
		if info.AuthType == authType {
			result = append(result, info)
		}
	}
	return result
}

// GetProvidersByTier returns all providers in a specific tier
func GetProvidersByTier(tier int) []*ProviderTypeInfo {
	var result []*ProviderTypeInfo
	for _, info := range SupportedProviders {
		if info.Tier == tier {
			result = append(result, info)
		}
	}
	return result
}
