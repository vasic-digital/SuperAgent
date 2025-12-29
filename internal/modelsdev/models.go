package modelsdev

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type ListModelsOptions struct {
	Provider   string
	Search     string
	ModelType  string
	Capability string
	Page       int
	Limit      int
}

type ModelInfo struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Provider      string                 `json:"provider"`
	DisplayName   string                 `json:"display_name"`
	Description   string                 `json:"description"`
	ContextWindow int                    `json:"context_window"`
	MaxTokens     int                    `json:"max_tokens"`
	Pricing       *ModelPricing          `json:"pricing"`
	Capabilities  ModelCapabilities      `json:"capabilities"`
	Performance   *ModelPerformance      `json:"performance"`
	Tags          []string               `json:"tags"`
	Categories    []string               `json:"categories"`
	Family        string                 `json:"family"`
	Version       string                 `json:"version"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type ModelPricing struct {
	InputPrice  float64 `json:"input_price"`
	OutputPrice float64 `json:"output_price"`
	Currency    string  `json:"currency"`
	Unit        string  `json:"unit"`
}

type ModelCapabilities struct {
	Vision          bool `json:"vision"`
	FunctionCalling bool `json:"function_calling"`
	Streaming       bool `json:"streaming"`
	JSONMode        bool `json:"json_mode"`
	ImageGeneration bool `json:"image_generation"`
	Audio           bool `json:"audio"`
	CodeGeneration  bool `json:"code_generation"`
	Reasoning       bool `json:"reasoning"`
	ToolUse         bool `json:"tool_use"`
}

type ModelPerformance struct {
	BenchmarkScore   float64            `json:"benchmark_score"`
	PopularityScore  int                `json:"popularity_score"`
	ReliabilityScore float64            `json:"reliability_score"`
	Benchmarks       map[string]float64 `json:"benchmarks"`
}

type ModelsListResponse struct {
	Models []ModelInfo `json:"models"`
	Total  int         `json:"total"`
	Page   int         `json:"page"`
	Limit  int         `json:"limit"`
}

type ModelDetailsResponse struct {
	Model      ModelInfo        `json:"model"`
	Benchmarks []ModelBenchmark `json:"benchmarks,omitempty"`
}

type ModelBenchmark struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Score           float64 `json:"score"`
	Rank            int     `json:"rank"`
	NormalizedScore float64 `json:"normalized_score"`
	Date            string  `json:"date"`
}

type ProviderInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	ModelsCount int      `json:"models_count"`
	Website     string   `json:"website"`
	APIDocsURL  string   `json:"api_docs_url"`
	Features    []string `json:"features"`
}

type ProvidersListResponse struct {
	Providers []ProviderInfo `json:"providers"`
	Total     int            `json:"total"`
}

func (c *Client) ListModels(ctx context.Context, opts *ListModelsOptions) (*ModelsListResponse, error) {
	if opts == nil {
		opts = &ListModelsOptions{}
	}

	path := "/models"
	query := url.Values{}

	if opts.Provider != "" {
		query.Set("provider", opts.Provider)
	}
	if opts.Search != "" {
		query.Set("search", opts.Search)
	}
	if opts.ModelType != "" {
		query.Set("type", opts.ModelType)
	}
	if opts.Capability != "" {
		query.Set("capability", opts.Capability)
	}
	if opts.Page > 0 {
		query.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var response ModelsListResponse
	if err := c.doGet(ctx, path, &response); err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	return &response, nil
}

func (c *Client) GetModel(ctx context.Context, modelID string) (*ModelDetailsResponse, error) {
	if modelID == "" {
		return nil, fmt.Errorf("model ID is required")
	}

	path := fmt.Sprintf("/models/%s", url.PathEscape(modelID))

	var response ModelDetailsResponse
	if err := c.doGet(ctx, path, &response); err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	return &response, nil
}

func (c *Client) SearchModels(ctx context.Context, query string, opts *ListModelsOptions) (*ModelsListResponse, error) {
	if opts == nil {
		opts = &ListModelsOptions{}
	}
	opts.Search = query
	return c.ListModels(ctx, opts)
}

func (c *Client) ListProviders(ctx context.Context) (*ProvidersListResponse, error) {
	path := "/providers"

	var response ProvidersListResponse
	if err := c.doGet(ctx, path, &response); err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	return &response, nil
}

func (c *Client) GetProvider(ctx context.Context, providerID string) (*ProviderInfo, error) {
	if providerID == "" {
		return nil, fmt.Errorf("provider ID is required")
	}

	path := fmt.Sprintf("/providers/%s", url.PathEscape(providerID))

	var response ProviderInfo
	if err := c.doGet(ctx, path, &response); err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	return &response, nil
}

func (c *Client) ListProviderModels(ctx context.Context, providerID string, opts *ListModelsOptions) (*ModelsListResponse, error) {
	if providerID == "" {
		return nil, fmt.Errorf("provider ID is required")
	}

	if opts == nil {
		opts = &ListModelsOptions{}
	}
	opts.Provider = providerID

	return c.ListModels(ctx, opts)
}
