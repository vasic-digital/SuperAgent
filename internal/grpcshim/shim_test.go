package grpcshim

import (
	"testing"
	"time"

	"dev.helix.agent/internal/models"
)

func TestShimCompleteRequest_ToLLMRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  *ShimCompleteRequest
		expected *models.LLMRequest
	}{
		{
			name: "basic request conversion",
			request: &ShimCompleteRequest{
				Prompt:         "test prompt",
				SessionID:      "session-123",
				UserID:         "user-456",
				MemoryEnhanced: true,
				ModelParams: models.ModelParameters{
					Model:       "gpt-4",
					Temperature: 0.7,
					MaxTokens:   1000,
				},
				EnsembleConfig: &models.EnsembleConfig{
					Strategy: "round_robin",
				},
				RequestType: "completion",
			},
			expected: &models.LLMRequest{
				ID:             "session-123:shim",
				SessionID:      "session-123",
				UserID:         "user-456",
				Prompt:         "test prompt",
				MemoryEnhanced: true,
				ModelParams: models.ModelParameters{
					Model:       "gpt-4",
					Temperature: 0.7,
					MaxTokens:   1000,
				},
				EnsembleConfig: &models.EnsembleConfig{
					Strategy: "round_robin",
				},
				RequestType: "completion",
			},
		},
		{
			name: "request without ensemble config",
			request: &ShimCompleteRequest{
				Prompt:      "simple prompt",
				SessionID:   "session-789",
				UserID:      "user-999",
				RequestType: "chat",
			},
			expected: &models.LLMRequest{
				ID:             "session-789:shim",
				SessionID:      "session-789",
				UserID:         "user-999",
				Prompt:         "simple prompt",
				MemoryEnhanced: false,
				ModelParams:    models.ModelParameters{},
				EnsembleConfig: nil,
				RequestType:    "chat",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ToLLMRequest()

			if result.ID != tt.expected.ID {
				t.Errorf("ID: got %v, want %v", result.ID, tt.expected.ID)
			}
			if result.SessionID != tt.expected.SessionID {
				t.Errorf("SessionID: got %v, want %v", result.SessionID, tt.expected.SessionID)
			}
			if result.UserID != tt.expected.UserID {
				t.Errorf("UserID: got %v, want %v", result.UserID, tt.expected.UserID)
			}
			if result.Prompt != tt.expected.Prompt {
				t.Errorf("Prompt: got %v, want %v", result.Prompt, tt.expected.Prompt)
			}
			if result.MemoryEnhanced != tt.expected.MemoryEnhanced {
				t.Errorf("MemoryEnhanced: got %v, want %v", result.MemoryEnhanced, tt.expected.MemoryEnhanced)
			}
			if result.ModelParams.Model != tt.expected.ModelParams.Model {
				t.Errorf("ModelParams.Model: got %v, want %v", result.ModelParams.Model, tt.expected.ModelParams.Model)
			}
			if result.RequestType != tt.expected.RequestType {
				t.Errorf("RequestType: got %v, want %v", result.RequestType, tt.expected.RequestType)
			}

			if tt.expected.EnsembleConfig == nil {
				if result.EnsembleConfig != nil {
					t.Errorf("EnsembleConfig: expected nil, got %v", result.EnsembleConfig)
				}
			} else {
				if result.EnsembleConfig == nil {
					t.Errorf("EnsembleConfig: expected %v, got nil", tt.expected.EnsembleConfig)
				} else if result.EnsembleConfig.Strategy != tt.expected.EnsembleConfig.Strategy {
					t.Errorf("EnsembleConfig.Strategy: got %v, want %v", result.EnsembleConfig.Strategy, tt.expected.EnsembleConfig.Strategy)
				}
			}

			if result.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}

			if time.Since(result.CreatedAt) > time.Second {
				t.Error("CreatedAt should be recent")
			}
		})
	}
}

func TestShimCompleteRequest_ZeroValues(t *testing.T) {
	req := &ShimCompleteRequest{}
	result := req.ToLLMRequest()

	if result.ID != ":shim" {
		t.Errorf("ID: got %v, want ':shim'", result.ID)
	}
	if result.SessionID != "" {
		t.Errorf("SessionID: got %v, want empty", result.SessionID)
	}
	if result.UserID != "" {
		t.Errorf("UserID: got %v, want empty", result.UserID)
	}
	if result.Prompt != "" {
		t.Errorf("Prompt: got %v, want empty", result.Prompt)
	}
	if result.MemoryEnhanced != false {
		t.Errorf("MemoryEnhanced: got %v, want false", result.MemoryEnhanced)
	}
	if result.RequestType != "" {
		t.Errorf("RequestType: got %v, want empty", result.RequestType)
	}
	if result.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}
