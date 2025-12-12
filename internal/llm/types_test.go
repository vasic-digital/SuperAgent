package llm

import (
	"testing"
)

func TestProviderCapabilities_DefaultValues(t *testing.T) {
	cap := &ProviderCapabilities{}

	if cap.SupportedModels != nil {
		t.Errorf("SupportedModels should be nil, got %v", cap.SupportedModels)
	}
	if cap.SupportedFeatures != nil {
		t.Errorf("SupportedFeatures should be nil, got %v", cap.SupportedFeatures)
	}
	if cap.SupportedRequestTypes != nil {
		t.Errorf("SupportedRequestTypes should be nil, got %v", cap.SupportedRequestTypes)
	}
	if cap.SupportsStreaming != false {
		t.Errorf("SupportsStreaming should be false, got %v", cap.SupportsStreaming)
	}
	if cap.SupportsFunctionCalling != false {
		t.Errorf("SupportsFunctionCalling should be false, got %v", cap.SupportsFunctionCalling)
	}
	if cap.SupportsVision != false {
		t.Errorf("SupportsVision should be false, got %v", cap.SupportsVision)
	}
	if cap.SupportsTools != false {
		t.Errorf("SupportsTools should be false, got %v", cap.SupportsTools)
	}
	if cap.SupportsSearch != false {
		t.Errorf("SupportsSearch should be false, got %v", cap.SupportsSearch)
	}
	if cap.SupportsReasoning != false {
		t.Errorf("SupportsReasoning should be false, got %v", cap.SupportsReasoning)
	}
	if cap.SupportsCodeCompletion != false {
		t.Errorf("SupportsCodeCompletion should be false, got %v", cap.SupportsCodeCompletion)
	}
	if cap.SupportsCodeAnalysis != false {
		t.Errorf("SupportsCodeAnalysis should be false, got %v", cap.SupportsCodeAnalysis)
	}
	if cap.SupportsRefactoring != false {
		t.Errorf("SupportsRefactoring should be false, got %v", cap.SupportsRefactoring)
	}
	if cap.Limits.MaxTokens != 0 {
		t.Errorf("Limits.MaxTokens should be 0, got %v", cap.Limits.MaxTokens)
	}
	if cap.Limits.MaxInputLength != 0 {
		t.Errorf("Limits.MaxInputLength should be 0, got %v", cap.Limits.MaxInputLength)
	}
	if cap.Limits.MaxOutputLength != 0 {
		t.Errorf("Limits.MaxOutputLength should be 0, got %v", cap.Limits.MaxOutputLength)
	}
	if cap.Limits.MaxConcurrentRequests != 0 {
		t.Errorf("Limits.MaxConcurrentRequests should be 0, got %v", cap.Limits.MaxConcurrentRequests)
	}
	if cap.Metadata != nil {
		t.Errorf("Metadata should be nil, got %v", cap.Metadata)
	}
}

func TestProviderCapabilities_WithValues(t *testing.T) {
	cap := &ProviderCapabilities{
		SupportedModels:         []string{"gpt-4", "gpt-3.5-turbo"},
		SupportedFeatures:       []string{"chat", "completion", "embeddings"},
		SupportedRequestTypes:   []string{"text", "chat", "code"},
		SupportsStreaming:       true,
		SupportsFunctionCalling: true,
		SupportsVision:          true,
		SupportsTools:           true,
		SupportsSearch:          true,
		SupportsReasoning:       true,
		SupportsCodeCompletion:  true,
		SupportsCodeAnalysis:    true,
		SupportsRefactoring:     true,
		Limits: ModelLimits{
			MaxTokens:             4096,
			MaxInputLength:        8192,
			MaxOutputLength:       2048,
			MaxConcurrentRequests: 10,
		},
		Metadata: map[string]string{
			"provider": "openai",
			"version":  "1.0",
		},
	}

	if len(cap.SupportedModels) != 2 {
		t.Errorf("SupportedModels length: got %v, want 2", len(cap.SupportedModels))
	}
	if cap.SupportedModels[0] != "gpt-4" {
		t.Errorf("SupportedModels[0]: got %v, want 'gpt-4'", cap.SupportedModels[0])
	}
	if len(cap.SupportedFeatures) != 3 {
		t.Errorf("SupportedFeatures length: got %v, want 3", len(cap.SupportedFeatures))
	}
	if len(cap.SupportedRequestTypes) != 3 {
		t.Errorf("SupportedRequestTypes length: got %v, want 3", len(cap.SupportedRequestTypes))
	}
	if !cap.SupportsStreaming {
		t.Error("SupportsStreaming should be true")
	}
	if !cap.SupportsFunctionCalling {
		t.Error("SupportsFunctionCalling should be true")
	}
	if !cap.SupportsVision {
		t.Error("SupportsVision should be true")
	}
	if !cap.SupportsTools {
		t.Error("SupportsTools should be true")
	}
	if !cap.SupportsSearch {
		t.Error("SupportsSearch should be true")
	}
	if !cap.SupportsReasoning {
		t.Error("SupportsReasoning should be true")
	}
	if !cap.SupportsCodeCompletion {
		t.Error("SupportsCodeCompletion should be true")
	}
	if !cap.SupportsCodeAnalysis {
		t.Error("SupportsCodeAnalysis should be true")
	}
	if !cap.SupportsRefactoring {
		t.Error("SupportsRefactoring should be true")
	}
	if cap.Limits.MaxTokens != 4096 {
		t.Errorf("Limits.MaxTokens: got %v, want 4096", cap.Limits.MaxTokens)
	}
	if cap.Limits.MaxInputLength != 8192 {
		t.Errorf("Limits.MaxInputLength: got %v, want 8192", cap.Limits.MaxInputLength)
	}
	if cap.Limits.MaxOutputLength != 2048 {
		t.Errorf("Limits.MaxOutputLength: got %v, want 2048", cap.Limits.MaxOutputLength)
	}
	if cap.Limits.MaxConcurrentRequests != 10 {
		t.Errorf("Limits.MaxConcurrentRequests: got %v, want 10", cap.Limits.MaxConcurrentRequests)
	}
	if len(cap.Metadata) != 2 {
		t.Errorf("Metadata length: got %v, want 2", len(cap.Metadata))
	}
	if cap.Metadata["provider"] != "openai" {
		t.Errorf("Metadata['provider']: got %v, want 'openai'", cap.Metadata["provider"])
	}
}

func TestModelLimits_DefaultValues(t *testing.T) {
	limits := ModelLimits{}

	if limits.MaxTokens != 0 {
		t.Errorf("MaxTokens should be 0, got %v", limits.MaxTokens)
	}
	if limits.MaxInputLength != 0 {
		t.Errorf("MaxInputLength should be 0, got %v", limits.MaxInputLength)
	}
	if limits.MaxOutputLength != 0 {
		t.Errorf("MaxOutputLength should be 0, got %v", limits.MaxOutputLength)
	}
	if limits.MaxConcurrentRequests != 0 {
		t.Errorf("MaxConcurrentRequests should be 0, got %v", limits.MaxConcurrentRequests)
	}
}

func TestModelLimits_WithValues(t *testing.T) {
	limits := ModelLimits{
		MaxTokens:             1000,
		MaxInputLength:        2000,
		MaxOutputLength:       500,
		MaxConcurrentRequests: 5,
	}

	if limits.MaxTokens != 1000 {
		t.Errorf("MaxTokens: got %v, want 1000", limits.MaxTokens)
	}
	if limits.MaxInputLength != 2000 {
		t.Errorf("MaxInputLength: got %v, want 2000", limits.MaxInputLength)
	}
	if limits.MaxOutputLength != 500 {
		t.Errorf("MaxOutputLength: got %v, want 500", limits.MaxOutputLength)
	}
	if limits.MaxConcurrentRequests != 5 {
		t.Errorf("MaxConcurrentRequests: got %v, want 5", limits.MaxConcurrentRequests)
	}
}

func TestProviderCapabilities_JSONTags(t *testing.T) {
	cap := &ProviderCapabilities{
		SupportedModels:       []string{"test"},
		SupportedFeatures:     []string{"test"},
		SupportedRequestTypes: []string{"test"},
		Limits: ModelLimits{
			MaxTokens: 100,
		},
		Metadata: map[string]string{"key": "value"},
	}

	if cap.SupportedModels == nil {
		t.Error("SupportedModels should not be nil for JSON marshaling test")
	}
	if cap.SupportedFeatures == nil {
		t.Error("SupportedFeatures should not be nil for JSON marshaling test")
	}
	if cap.SupportedRequestTypes == nil {
		t.Error("SupportedRequestTypes should not be nil for JSON marshaling test")
	}
	if cap.Metadata == nil {
		t.Error("Metadata should not be nil for JSON marshaling test")
	}
}
