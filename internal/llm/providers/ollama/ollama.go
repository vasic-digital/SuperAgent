package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/superagent/superagent/internal/models"
)

// OllamaProvider implements the LLMProvider interface for local Ollama models
type OllamaProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt,omitempty"`
	Stream  bool          `json:"stream,omitempty"`
	Options OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions represents options for Ollama generation
type OllamaOptions struct {
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	MaxTokens   int      `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Context  []int  `json:"context,omitempty"`
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama2"
	}

	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Ollama can be slow for first requests
		},
	}
}

// Complete implements the LLMProvider interface
func (o *OllamaProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	ollamaReq := OllamaRequest{
		Model:  o.model,
		Prompt: req.Prompt,
		Stream: false,
		Options: OllamaOptions{
			Temperature: req.ModelParams.Temperature,
			TopP:        req.ModelParams.TopP,
			MaxTokens:   req.ModelParams.MaxTokens,
			Stop:        req.ModelParams.StopSequences,
		},
	}

	resp, err := o.makeRequest(ctx, ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to complete request: %w", err)
	}

	return o.convertResponse(resp, req.ID)
}

// CompleteStream implements streaming completion
func (o *OllamaProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	ch := make(chan *models.LLMResponse)

	go func() {
		defer close(ch)

		ollamaReq := OllamaRequest{
			Model:  o.model,
			Prompt: req.Prompt,
			Stream: true,
			Options: OllamaOptions{
				Temperature: req.ModelParams.Temperature,
				TopP:        req.ModelParams.TopP,
				MaxTokens:   req.ModelParams.MaxTokens,
				Stop:        req.ModelParams.StopSequences,
			},
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", nil)
		if err != nil {
			ch <- &models.LLMResponse{
				RequestID:    req.ID,
				ProviderID:   "ollama",
				ProviderName: "Ollama",
				Content:      fmt.Sprintf("Error: %v", err),
				Confidence:   0.0,
				FinishReason: "error",
				CreatedAt:    time.Now(),
			}
			return
		}

		jsonData, err := json.Marshal(ollamaReq)
		if err != nil {
			ch <- &models.LLMResponse{
				RequestID:    req.ID,
				ProviderID:   "ollama",
				ProviderName: "Ollama",
				Content:      fmt.Sprintf("Error: %v", err),
				Confidence:   0.0,
				FinishReason: "error",
				CreatedAt:    time.Now(),
			}
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Body = io.NopCloser(bytes.NewBuffer(jsonData))

		response, err := o.httpClient.Do(httpReq)
		if err != nil {
			ch <- &models.LLMResponse{
				RequestID:    req.ID,
				ProviderID:   "ollama",
				ProviderName: "Ollama",
				Content:      fmt.Sprintf("Error: %v", err),
				Confidence:   0.0,
				FinishReason: "error",
				CreatedAt:    time.Now(),
			}
			return
		}
		defer response.Body.Close()

		decoder := json.NewDecoder(response.Body)
		fullContent := ""

		for {
			var streamResp OllamaResponse
			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			fullContent += streamResp.Response

			chunkResp := &models.LLMResponse{
				RequestID:      req.ID,
				ProviderID:     "ollama",
				ProviderName:   "Ollama",
				Content:        streamResp.Response,
				Confidence:     0.8,
				TokensUsed:     1,
				ResponseTime:   time.Now().UnixMilli(),
				FinishReason:   "",
				Selected:       false,
				SelectionScore: 0.0,
				CreatedAt:      time.Now(),
			}

			select {
			case ch <- chunkResp:
			case <-ctx.Done():
				return
			}

			if streamResp.Done {
				// Send final response
				finalResp := &models.LLMResponse{
					RequestID:      req.ID,
					ProviderID:     "ollama",
					ProviderName:   "Ollama",
					Content:        "",
					Confidence:     0.8,
					TokensUsed:     len(fullContent) / 4,
					ResponseTime:   time.Now().UnixMilli(),
					FinishReason:   "stop",
					Selected:       false,
					SelectionScore: 0.0,
					CreatedAt:      time.Now(),
				}
				ch <- finalResp
				break
			}
		}
	}()

	return ch, nil
}

// HealthCheck implements health checking for the Ollama provider
func (o *OllamaProvider) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetCapabilities returns the capabilities of the Ollama provider
func (o *OllamaProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportedModels: []string{
			"llama2",
			"llama2:13b",
			"llama2:70b",
			"codellama",
			"mistral",
			"vicuna",
			"orca-mini",
		},
		SupportedFeatures: []string{
			"text_completion",
			"chat",
			"streaming",
		},
		SupportedRequestTypes: []string{
			"text_completion",
			"chat",
		},
		SupportsStreaming:       true,
		SupportsFunctionCalling: false,
		SupportsVision:          false,
		Limits: models.ModelLimits{
			MaxTokens:             4096,
			MaxInputLength:        4096,
			MaxOutputLength:       4096,
			MaxConcurrentRequests: 1, // Ollama typically handles one request at a time
		},
		Metadata: map[string]string{
			"provider":     "Ollama",
			"model_family": "Local Models",
			"api_version":  "v1",
			"local":        "true",
		},
	}
}

// ValidateConfig validates the provider configuration
func (o *OllamaProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	var errors []string

	if o.baseURL == "" {
		errors = append(errors, "base URL is required")
	}

	if o.model == "" {
		errors = append(errors, "model is required")
	}

	return len(errors) == 0, errors
}

// convertResponse converts Ollama API response to internal format
func (o *OllamaProvider) convertResponse(resp *OllamaResponse, requestID string) (*models.LLMResponse, error) {
	return &models.LLMResponse{
		ID:           fmt.Sprintf("ollama-%d", time.Now().Unix()),
		RequestID:    requestID,
		ProviderID:   "ollama",
		ProviderName: "Ollama",
		Content:      resp.Response,
		Confidence:   0.8,                    // Ollama doesn't provide confidence scores
		TokensUsed:   len(resp.Response) / 4, // Rough estimate
		ResponseTime: time.Now().UnixMilli(),
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"model":   resp.Model,
			"context": len(resp.Context),
		},
		Selected:       false,
		SelectionScore: 0.0,
		CreatedAt:      time.Now(),
	}, nil
}

// makeRequest sends a request to the Ollama API
func (o *OllamaProvider) makeRequest(ctx context.Context, req OllamaRequest) (*OllamaResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &ollamaResp, nil
}
