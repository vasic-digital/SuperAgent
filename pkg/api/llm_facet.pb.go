package api

import (
	"context"
)

// CompletionRequest mirrors the gRPC request payload (minimal surface for CI land)
type CompletionRequest struct {
	Prompt         string            `json:"prompt"`
	SessionID      string            `json:"session_id"`
	UserID         string            `json:"user_id"`
	MemoryEnhanced bool              `json:"memory_enhanced"`
	ModelParams    *ModelParameters  `json:"model_params,omitempty"`
	EnsembleConfig *EnsembleConfig   `json:"ensemble_config,omitempty"`
	Memory         map[string]string `json:"memory,omitempty"`
	RequestType    string            `json:"request_type"` // code_generation, reasoning, tool_use
	Priority       int32             `json:"priority"`
	CreatedAt      int64             `json:"created_at"`
}

// CompletionResponse mirrors the gRPC response payload (minimal surface for CI land)
type CompletionResponse struct {
	Response       string  `json:"response"`
	Confidence     float64 `json:"confidence"`
	TokensUsed     int32   `json:"tokens_used"`
	Selected       bool    `json:"selected"`
	SelectionScore float64 `json:"selection_score"`
	ProviderName   string  `json:"provider_name"`
	CreatedAt      int64   `json:"created_at"`
}

type ModelParameters struct {
	Model            string                 `json:"model"`
	Temperature      float64                `json:"temperature"`
	MaxTokens        int32                  `json:"max_tokens"`
	TopP             float64                `json:"top_p"`
	StopSequences    []string               `json:"stop_sequences"`
	ProviderSpecific map[string]interface{} `json:"provider_specific"`
}

type EnsembleConfig struct {
	Strategy            string   `json:"strategy"`
	MinProviders        int      `json:"min_providers"`
	ConfidenceThreshold float64  `json:"confidence_threshold"`
	FallbackToBest      bool     `json:"fallback_to_best"`
	Timeout             int32    `json:"timeout"`
	PreferredProviders  []string `json:"preferred_providers"`
}

type MemorySource struct {
	DatasetName    string  `json:"dataset_name"`
	Content        string  `json:"content"`
	RelevanceScore float64 `json:"relevance_score"`
	SourceType     string  `json:"source_type"`
}

// LLMFacadeServer defines the gRPC surface for the facade (minimal)
type LLMFacadeServer interface {
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
}

type UnimplementedLLMFacadeServer struct{}

func RegisterLLMFacadeServer(s interface{}, srv LLMFacadeServer) {}
