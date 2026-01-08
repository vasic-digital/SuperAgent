package services

import (
	"testing"
	"time"

	"github.com/helixagent/helixagent/internal/models"
)

func TestProviderConfig_DefaultValues(t *testing.T) {
	cfg := &ProviderConfig{}

	if cfg.Name != "" {
		t.Errorf("Name should be empty, got %v", cfg.Name)
	}
	if cfg.Type != "" {
		t.Errorf("Type should be empty, got %v", cfg.Type)
	}
	if cfg.Enabled != false {
		t.Errorf("Enabled should be false, got %v", cfg.Enabled)
	}
	if cfg.APIKey != "" {
		t.Errorf("APIKey should be empty, got %v", cfg.APIKey)
	}
	if cfg.BaseURL != "" {
		t.Errorf("BaseURL should be empty, got %v", cfg.BaseURL)
	}
	if cfg.Models != nil {
		t.Errorf("Models should be nil, got %v", cfg.Models)
	}
	if cfg.Timeout != 0 {
		t.Errorf("Timeout should be 0, got %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 0 {
		t.Errorf("MaxRetries should be 0, got %v", cfg.MaxRetries)
	}
	if cfg.HealthCheckURL != "" {
		t.Errorf("HealthCheckURL should be empty, got %v", cfg.HealthCheckURL)
	}
	if cfg.Weight != 0.0 {
		t.Errorf("Weight should be 0.0, got %v", cfg.Weight)
	}
	if cfg.Tags != nil {
		t.Errorf("Tags should be nil, got %v", cfg.Tags)
	}
	if cfg.Capabilities != nil {
		t.Errorf("Capabilities should be nil, got %v", cfg.Capabilities)
	}
	if cfg.CustomSettings != nil {
		t.Errorf("CustomSettings should be nil, got %v", cfg.CustomSettings)
	}
}

func TestProviderConfig_WithValues(t *testing.T) {
	cfg := &ProviderConfig{
		Name:    "openai",
		Type:    "openai",
		Enabled: true,
		APIKey:  "test-key",
		BaseURL: "https://api.openai.com/v1",
		Models: []ModelConfig{
			{
				ID:      "gpt-4",
				Name:    "GPT-4",
				Enabled: true,
				Weight:  1.0,
			},
		},
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		HealthCheckURL: "https://api.openai.com/health",
		Weight:         0.8,
		Tags:           []string{"premium", "chat"},
		Capabilities: map[string]string{
			"streaming": "true",
			"vision":    "false",
		},
		CustomSettings: map[string]any{
			"temperature": 0.7,
		},
	}

	if cfg.Name != "openai" {
		t.Errorf("Name: got %v, want 'openai'", cfg.Name)
	}
	if cfg.Type != "openai" {
		t.Errorf("Type: got %v, want 'openai'", cfg.Type)
	}
	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("APIKey: got %v, want 'test-key'", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("BaseURL: got %v, want 'https://api.openai.com/v1'", cfg.BaseURL)
	}
	if len(cfg.Models) != 1 {
		t.Errorf("Models length: got %v, want 1", len(cfg.Models))
	}
	if cfg.Models[0].ID != "gpt-4" {
		t.Errorf("Models[0].ID: got %v, want 'gpt-4'", cfg.Models[0].ID)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout: got %v, want 30s", cfg.Timeout)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries: got %v, want 3", cfg.MaxRetries)
	}
	if cfg.HealthCheckURL != "https://api.openai.com/health" {
		t.Errorf("HealthCheckURL: got %v, want 'https://api.openai.com/health'", cfg.HealthCheckURL)
	}
	if cfg.Weight != 0.8 {
		t.Errorf("Weight: got %v, want 0.8", cfg.Weight)
	}
	if len(cfg.Tags) != 2 {
		t.Errorf("Tags length: got %v, want 2", len(cfg.Tags))
	}
	if cfg.Tags[0] != "premium" {
		t.Errorf("Tags[0]: got %v, want 'premium'", cfg.Tags[0])
	}
	if len(cfg.Capabilities) != 2 {
		t.Errorf("Capabilities length: got %v, want 2", len(cfg.Capabilities))
	}
	if cfg.Capabilities["streaming"] != "true" {
		t.Errorf("Capabilities['streaming']: got %v, want 'true'", cfg.Capabilities["streaming"])
	}
	if len(cfg.CustomSettings) != 1 {
		t.Errorf("CustomSettings length: got %v, want 1", len(cfg.CustomSettings))
	}
	if cfg.CustomSettings["temperature"] != 0.7 {
		t.Errorf("CustomSettings['temperature']: got %v, want 0.7", cfg.CustomSettings["temperature"])
	}
}

func TestModelConfig_DefaultValues(t *testing.T) {
	cfg := &ModelConfig{}

	if cfg.ID != "" {
		t.Errorf("ID should be empty, got %v", cfg.ID)
	}
	if cfg.Name != "" {
		t.Errorf("Name should be empty, got %v", cfg.Name)
	}
	if cfg.Enabled != false {
		t.Errorf("Enabled should be false, got %v", cfg.Enabled)
	}
	if cfg.Weight != 0.0 {
		t.Errorf("Weight should be 0.0, got %v", cfg.Weight)
	}
	if cfg.Capabilities != nil {
		t.Errorf("Capabilities should be nil, got %v", cfg.Capabilities)
	}
	if cfg.CustomParams != nil {
		t.Errorf("CustomParams should be nil, got %v", cfg.CustomParams)
	}
}

func TestModelConfig_WithValues(t *testing.T) {
	cfg := &ModelConfig{
		ID:      "gpt-4-turbo",
		Name:    "GPT-4 Turbo",
		Enabled: true,
		Weight:  0.9,
		Capabilities: []string{
			"chat",
			"completion",
			"function_calling",
		},
		CustomParams: map[string]any{
			"max_tokens":  4096,
			"temperature": 0.7,
		},
	}

	if cfg.ID != "gpt-4-turbo" {
		t.Errorf("ID: got %v, want 'gpt-4-turbo'", cfg.ID)
	}
	if cfg.Name != "GPT-4 Turbo" {
		t.Errorf("Name: got %v, want 'GPT-4 Turbo'", cfg.Name)
	}
	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.Weight != 0.9 {
		t.Errorf("Weight: got %v, want 0.9", cfg.Weight)
	}
	if len(cfg.Capabilities) != 3 {
		t.Errorf("Capabilities length: got %v, want 3", len(cfg.Capabilities))
	}
	if cfg.Capabilities[0] != "chat" {
		t.Errorf("Capabilities[0]: got %v, want 'chat'", cfg.Capabilities[0])
	}
	if len(cfg.CustomParams) != 2 {
		t.Errorf("CustomParams length: got %v, want 2", len(cfg.CustomParams))
	}
	if cfg.CustomParams["max_tokens"] != 4096 {
		t.Errorf("CustomParams['max_tokens']: got %v, want 4096", cfg.CustomParams["max_tokens"])
	}
}

func TestRegistryConfig_DefaultValues(t *testing.T) {
	cfg := &RegistryConfig{}

	if cfg.DefaultTimeout != 0 {
		t.Errorf("DefaultTimeout should be 0, got %v", cfg.DefaultTimeout)
	}
	if cfg.MaxRetries != 0 {
		t.Errorf("MaxRetries should be 0, got %v", cfg.MaxRetries)
	}
	if cfg.HealthCheck.Enabled != false {
		t.Errorf("HealthCheck.Enabled should be false, got %v", cfg.HealthCheck.Enabled)
	}
	if cfg.HealthCheck.Interval != 0 {
		t.Errorf("HealthCheck.Interval should be 0, got %v", cfg.HealthCheck.Interval)
	}
	if cfg.HealthCheck.Timeout != 0 {
		t.Errorf("HealthCheck.Timeout should be 0, got %v", cfg.HealthCheck.Timeout)
	}
	if cfg.HealthCheck.FailureThreshold != 0 {
		t.Errorf("HealthCheck.FailureThreshold should be 0, got %v", cfg.HealthCheck.FailureThreshold)
	}
	if cfg.Providers != nil {
		t.Errorf("Providers should be nil, got %v", cfg.Providers)
	}
	if cfg.Ensemble != nil {
		t.Errorf("Ensemble should be nil, got %v", cfg.Ensemble)
	}
	if cfg.Routing != nil {
		t.Errorf("Routing should be nil, got %v", cfg.Routing)
	}
}

func TestRegistryConfig_WithValues(t *testing.T) {
	cfg := &RegistryConfig{
		DefaultTimeout: 60 * time.Second,
		MaxRetries:     5,
		HealthCheck: HealthCheckConfig{
			Enabled:          true,
			Interval:         30 * time.Second,
			Timeout:          5 * time.Second,
			FailureThreshold: 3,
		},
		Providers: map[string]*ProviderConfig{
			"openai": {
				Name: "openai",
				Type: "openai",
			},
		},
		Ensemble: &models.EnsembleConfig{
			Strategy: "confidence_weighted",
		},
		Routing: &RoutingConfig{
			Strategy: "weighted",
			Weights: map[string]float64{
				"openai": 0.8,
			},
		},
	}

	if cfg.DefaultTimeout != 60*time.Second {
		t.Errorf("DefaultTimeout: got %v, want 60s", cfg.DefaultTimeout)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries: got %v, want 5", cfg.MaxRetries)
	}
	if !cfg.HealthCheck.Enabled {
		t.Error("HealthCheck.Enabled should be true")
	}
	if cfg.HealthCheck.Interval != 30*time.Second {
		t.Errorf("HealthCheck.Interval: got %v, want 30s", cfg.HealthCheck.Interval)
	}
	if cfg.HealthCheck.Timeout != 5*time.Second {
		t.Errorf("HealthCheck.Timeout: got %v, want 5s", cfg.HealthCheck.Timeout)
	}
	if cfg.HealthCheck.FailureThreshold != 3 {
		t.Errorf("HealthCheck.FailureThreshold: got %v, want 3", cfg.HealthCheck.FailureThreshold)
	}
	if len(cfg.Providers) != 1 {
		t.Errorf("Providers length: got %v, want 1", len(cfg.Providers))
	}
	if cfg.Providers["openai"].Name != "openai" {
		t.Errorf("Providers['openai'].Name: got %v, want 'openai'", cfg.Providers["openai"].Name)
	}
	if cfg.Ensemble == nil {
		t.Error("Ensemble should not be nil")
	} else if cfg.Ensemble.Strategy != "confidence_weighted" {
		t.Errorf("Ensemble.Strategy: got %v, want 'confidence_weighted'", cfg.Ensemble.Strategy)
	}
	if cfg.Routing == nil {
		t.Error("Routing should not be nil")
	} else {
		if cfg.Routing.Strategy != "weighted" {
			t.Errorf("Routing.Strategy: got %v, want 'weighted'", cfg.Routing.Strategy)
		}
		if len(cfg.Routing.Weights) != 1 {
			t.Errorf("Routing.Weights length: got %v, want 1", len(cfg.Routing.Weights))
		}
		if cfg.Routing.Weights["openai"] != 0.8 {
			t.Errorf("Routing.Weights['openai']: got %v, want 0.8", cfg.Routing.Weights["openai"])
		}
	}
}

func TestHealthCheckConfig_DefaultValues(t *testing.T) {
	cfg := &HealthCheckConfig{}

	if cfg.Enabled != false {
		t.Errorf("Enabled should be false, got %v", cfg.Enabled)
	}
	if cfg.Interval != 0 {
		t.Errorf("Interval should be 0, got %v", cfg.Interval)
	}
	if cfg.Timeout != 0 {
		t.Errorf("Timeout should be 0, got %v", cfg.Timeout)
	}
	if cfg.FailureThreshold != 0 {
		t.Errorf("FailureThreshold should be 0, got %v", cfg.FailureThreshold)
	}
}

func TestHealthCheckConfig_WithValues(t *testing.T) {
	cfg := &HealthCheckConfig{
		Enabled:          true,
		Interval:         10 * time.Second,
		Timeout:          2 * time.Second,
		FailureThreshold: 5,
	}

	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("Interval: got %v, want 10s", cfg.Interval)
	}
	if cfg.Timeout != 2*time.Second {
		t.Errorf("Timeout: got %v, want 2s", cfg.Timeout)
	}
	if cfg.FailureThreshold != 5 {
		t.Errorf("FailureThreshold: got %v, want 5", cfg.FailureThreshold)
	}
}

func TestRoutingConfig_DefaultValues(t *testing.T) {
	cfg := &RoutingConfig{}

	if cfg.Strategy != "" {
		t.Errorf("Strategy should be empty, got %v", cfg.Strategy)
	}
	if cfg.Weights != nil {
		t.Errorf("Weights should be nil, got %v", cfg.Weights)
	}
}

func TestRoutingConfig_WithValues(t *testing.T) {
	cfg := &RoutingConfig{
		Strategy: "round_robin",
		Weights: map[string]float64{
			"provider1": 0.6,
			"provider2": 0.4,
		},
	}

	if cfg.Strategy != "round_robin" {
		t.Errorf("Strategy: got %v, want 'round_robin'", cfg.Strategy)
	}
	if len(cfg.Weights) != 2 {
		t.Errorf("Weights length: got %v, want 2", len(cfg.Weights))
	}
	if cfg.Weights["provider1"] != 0.6 {
		t.Errorf("Weights['provider1']: got %v, want 0.6", cfg.Weights["provider1"])
	}
	if cfg.Weights["provider2"] != 0.4 {
		t.Errorf("Weights['provider2']: got %v, want 0.4", cfg.Weights["provider2"])
	}
}
