// Package servers provides MCP server adapters for various services.
package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ReplicateConfig contains configuration for Replicate adapter.
type ReplicateConfig struct {
	APIToken string        `json:"api_token"`
	BaseURL  string        `json:"base_url"`
	Timeout  time.Duration `json:"timeout"`
}

// ReplicateModel represents a model on Replicate.
type ReplicateModel struct {
	URL            string                 `json:"url"`
	Owner          string                 `json:"owner"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Visibility     string                 `json:"visibility"`
	GithubURL      string                 `json:"github_url,omitempty"`
	PaperURL       string                 `json:"paper_url,omitempty"`
	LicenseURL     string                 `json:"license_url,omitempty"`
	RunCount       int                    `json:"run_count"`
	CoverImageURL  string                 `json:"cover_image_url,omitempty"`
	DefaultExample *ReplicatePrediction   `json:"default_example,omitempty"`
	LatestVersion  *ReplicateModelVersion `json:"latest_version,omitempty"`
}

// ReplicateModelVersion represents a version of a model.
type ReplicateModelVersion struct {
	ID            string                 `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	CogVersion    string                 `json:"cog_version"`
	OpenAPISchema map[string]interface{} `json:"openapi_schema,omitempty"`
}

// ReplicatePrediction represents a prediction (generation) request/result.
type ReplicatePrediction struct {
	ID          string                   `json:"id"`
	Version     string                   `json:"version"`
	Status      string                   `json:"status"`
	Input       map[string]interface{}   `json:"input"`
	Output      interface{}              `json:"output,omitempty"`
	Error       string                   `json:"error,omitempty"`
	Logs        string                   `json:"logs,omitempty"`
	Metrics     *ReplicateMetrics        `json:"metrics,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	StartedAt   *time.Time               `json:"started_at,omitempty"`
	CompletedAt *time.Time               `json:"completed_at,omitempty"`
	URLs        *ReplicatePredictionURLs `json:"urls,omitempty"`
}

// ReplicateMetrics contains prediction metrics.
type ReplicateMetrics struct {
	PredictTime float64 `json:"predict_time,omitempty"`
}

// ReplicatePredictionURLs contains URLs related to a prediction.
type ReplicatePredictionURLs struct {
	Get    string `json:"get"`
	Cancel string `json:"cancel"`
}

// ReplicateCollection represents a collection of models.
type ReplicateCollection struct {
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description string           `json:"description"`
	Models      []ReplicateModel `json:"models,omitempty"`
}

// ReplicateWebhook represents a webhook configuration.
type ReplicateWebhook struct {
	URL    string   `json:"url"`
	Events []string `json:"events,omitempty"`
}

// ReplicateAdapter implements ServerAdapter for Replicate API.
type ReplicateAdapter struct {
	mu        sync.RWMutex
	config    ReplicateConfig
	client    *http.Client
	connected bool
	baseURL   string
}

// NewReplicateAdapter creates a new Replicate adapter.
func NewReplicateAdapter(config ReplicateConfig) *ReplicateAdapter {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.replicate.com/v1"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &ReplicateAdapter{
		config:  config,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Connect establishes connection to Replicate API.
func (a *ReplicateAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Verify token by getting account info
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/account", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Replicate: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid API token")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to authenticate: %s", string(body))
	}

	a.connected = true
	return nil
}

// Close closes the adapter connection.
func (a *ReplicateAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	return nil
}

// Health checks if the adapter is healthy.
func (a *ReplicateAdapter) Health(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected {
		return fmt.Errorf("not connected")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/account", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// GetModel retrieves a specific model.
func (a *ReplicateAdapter) GetModel(ctx context.Context, owner, name string) (*ReplicateModel, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/models/%s/%s", a.baseURL, owner, name)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("model not found: %s/%s", owner, name)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get model: %s", string(body))
	}

	var model ReplicateModel
	if err := json.NewDecoder(resp.Body).Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to decode model: %w", err)
	}

	return &model, nil
}

// GetModelVersion retrieves a specific version of a model.
func (a *ReplicateAdapter) GetModelVersion(ctx context.Context, owner, name, versionID string) (*ReplicateModelVersion, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/models/%s/%s/versions/%s", a.baseURL, owner, name, versionID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model version: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get model version: %s", string(body))
	}

	var version ReplicateModelVersion
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return nil, fmt.Errorf("failed to decode version: %w", err)
	}

	return &version, nil
}

// ListModels lists models with optional filtering.
func (a *ReplicateAdapter) ListModels(ctx context.Context, cursor string) ([]ReplicateModel, string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := a.baseURL + "/models"
	if cursor != "" {
		endpoint += "?cursor=" + url.QueryEscape(cursor)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list models: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to list models: %s", string(body))
	}

	var result struct {
		Results []ReplicateModel `json:"results"`
		Next    string           `json:"next,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Results, result.Next, nil
}

// CreatePrediction creates a new prediction.
func (a *ReplicateAdapter) CreatePrediction(ctx context.Context, version string, input map[string]interface{}, webhook *ReplicateWebhook) (*ReplicatePrediction, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]interface{}{
		"version": version,
		"input":   input,
	}
	if webhook != nil {
		payload["webhook"] = webhook.URL
		if len(webhook.Events) > 0 {
			payload["webhook_events_filter"] = webhook.Events
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/predictions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create prediction: %s", string(respBody))
	}

	var prediction ReplicatePrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, fmt.Errorf("failed to decode prediction: %w", err)
	}

	return &prediction, nil
}

// GetPrediction retrieves a prediction by ID.
func (a *ReplicateAdapter) GetPrediction(ctx context.Context, predictionID string) (*ReplicatePrediction, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/predictions/%s", a.baseURL, predictionID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("prediction not found: %s", predictionID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get prediction: %s", string(body))
	}

	var prediction ReplicatePrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, fmt.Errorf("failed to decode prediction: %w", err)
	}

	return &prediction, nil
}

// CancelPrediction cancels a running prediction.
func (a *ReplicateAdapter) CancelPrediction(ctx context.Context, predictionID string) (*ReplicatePrediction, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/predictions/%s/cancel", a.baseURL, predictionID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel prediction: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to cancel prediction: %s", string(body))
	}

	var prediction ReplicatePrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, fmt.Errorf("failed to decode prediction: %w", err)
	}

	return &prediction, nil
}

// ListPredictions lists recent predictions.
func (a *ReplicateAdapter) ListPredictions(ctx context.Context, cursor string) ([]ReplicatePrediction, string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := a.baseURL + "/predictions"
	if cursor != "" {
		endpoint += "?cursor=" + url.QueryEscape(cursor)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list predictions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to list predictions: %s", string(body))
	}

	var result struct {
		Results []ReplicatePrediction `json:"results"`
		Next    string                `json:"next,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Results, result.Next, nil
}

// WaitForPrediction waits for a prediction to complete.
func (a *ReplicateAdapter) WaitForPrediction(ctx context.Context, predictionID string, pollInterval time.Duration) (*ReplicatePrediction, error) {
	if pollInterval == 0 {
		pollInterval = 1 * time.Second
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			prediction, err := a.GetPrediction(ctx, predictionID)
			if err != nil {
				return nil, err
			}

			switch prediction.Status {
			case "succeeded", "failed", "canceled":
				return prediction, nil
			case "starting", "processing":
				// Continue polling
			default:
				return nil, fmt.Errorf("unknown prediction status: %s", prediction.Status)
			}
		}
	}
}

// RunModel is a convenience method that creates a prediction and waits for it.
func (a *ReplicateAdapter) RunModel(ctx context.Context, owner, name string, input map[string]interface{}) (*ReplicatePrediction, error) {
	// Get the model to find latest version
	model, err := a.GetModel(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	if model.LatestVersion == nil {
		return nil, fmt.Errorf("model has no versions: %s/%s", owner, name)
	}

	// Create prediction
	prediction, err := a.CreatePrediction(ctx, model.LatestVersion.ID, input, nil)
	if err != nil {
		return nil, err
	}

	// Wait for completion
	return a.WaitForPrediction(ctx, prediction.ID, 2*time.Second)
}

// GenerateImage is a convenience method for generating images with Stable Diffusion.
func (a *ReplicateAdapter) GenerateImage(ctx context.Context, prompt string, options map[string]interface{}) (*ReplicatePrediction, error) {
	input := map[string]interface{}{
		"prompt": prompt,
	}

	// Merge options
	for k, v := range options {
		input[k] = v
	}

	// Use Stable Diffusion XL by default
	return a.RunModel(ctx, "stability-ai", "sdxl", input)
}

// GetCollection retrieves a collection by slug.
func (a *ReplicateAdapter) GetCollection(ctx context.Context, slug string) (*ReplicateCollection, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/collections/%s", a.baseURL, slug)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("collection not found: %s", slug)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get collection: %s", string(body))
	}

	var collection ReplicateCollection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("failed to decode collection: %w", err)
	}

	return &collection, nil
}

// ListCollections lists available collections.
func (a *ReplicateAdapter) ListCollections(ctx context.Context, cursor string) ([]ReplicateCollection, string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := a.baseURL + "/collections"
	if cursor != "" {
		endpoint += "?cursor=" + url.QueryEscape(cursor)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+a.config.APIToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list collections: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to list collections: %s", string(body))
	}

	var result struct {
		Results []ReplicateCollection `json:"results"`
		Next    string                `json:"next,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Results, result.Next, nil
}

// GetMCPTools returns the MCP tool definitions for Replicate.
func (a *ReplicateAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "replicate_get_model",
			Description: "Get information about a Replicate model",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Model owner username",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Model name",
					},
				},
				"required": []string{"owner", "name"},
			},
		},
		{
			Name:        "replicate_list_models",
			Description: "List available Replicate models",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cursor": map[string]interface{}{
						"type":        "string",
						"description": "Pagination cursor",
					},
				},
			},
		},
		{
			Name:        "replicate_create_prediction",
			Description: "Create a new prediction (run a model)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Model version ID",
					},
					"input": map[string]interface{}{
						"type":        "object",
						"description": "Input parameters for the model",
					},
					"webhook_url": map[string]interface{}{
						"type":        "string",
						"description": "Optional webhook URL for notifications",
					},
				},
				"required": []string{"version", "input"},
			},
		},
		{
			Name:        "replicate_get_prediction",
			Description: "Get the status and result of a prediction",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prediction_id": map[string]interface{}{
						"type":        "string",
						"description": "Prediction ID",
					},
				},
				"required": []string{"prediction_id"},
			},
		},
		{
			Name:        "replicate_cancel_prediction",
			Description: "Cancel a running prediction",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prediction_id": map[string]interface{}{
						"type":        "string",
						"description": "Prediction ID to cancel",
					},
				},
				"required": []string{"prediction_id"},
			},
		},
		{
			Name:        "replicate_list_predictions",
			Description: "List recent predictions",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cursor": map[string]interface{}{
						"type":        "string",
						"description": "Pagination cursor",
					},
				},
			},
		},
		{
			Name:        "replicate_run_model",
			Description: "Run a model and wait for the result",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Model owner username",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Model name",
					},
					"input": map[string]interface{}{
						"type":        "object",
						"description": "Input parameters for the model",
					},
				},
				"required": []string{"owner", "name", "input"},
			},
		},
		{
			Name:        "replicate_generate_image",
			Description: "Generate an image using Stable Diffusion",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"prompt": map[string]interface{}{
						"type":        "string",
						"description": "Text prompt describing the image to generate",
					},
					"negative_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Negative prompt (what to avoid)",
					},
					"width": map[string]interface{}{
						"type":        "integer",
						"description": "Image width in pixels",
					},
					"height": map[string]interface{}{
						"type":        "integer",
						"description": "Image height in pixels",
					},
					"num_outputs": map[string]interface{}{
						"type":        "integer",
						"description": "Number of images to generate",
					},
					"guidance_scale": map[string]interface{}{
						"type":        "number",
						"description": "Guidance scale for generation",
					},
					"num_inference_steps": map[string]interface{}{
						"type":        "integer",
						"description": "Number of inference steps",
					},
				},
				"required": []string{"prompt"},
			},
		},
		{
			Name:        "replicate_get_collection",
			Description: "Get a collection of models",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"slug": map[string]interface{}{
						"type":        "string",
						"description": "Collection slug/identifier",
					},
				},
				"required": []string{"slug"},
			},
		},
		{
			Name:        "replicate_list_collections",
			Description: "List available model collections",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cursor": map[string]interface{}{
						"type":        "string",
						"description": "Pagination cursor",
					},
				},
			},
		},
	}
}
