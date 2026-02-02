package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dev.helix.agent/internal/formatters"
	"github.com/sirupsen/logrus"
)

// ServiceFormatter is a formatter that runs as an HTTP service
type ServiceFormatter struct {
	metadata   *formatters.FormatterMetadata
	serviceURL string
	httpClient *http.Client
	logger     *logrus.Logger
}

// ServiceFormatRequest represents the HTTP request to a formatter service
type ServiceFormatRequest struct {
	Content string                 `json:"content"`
	Options map[string]interface{} `json:"options"`
}

// ServiceFormatResponse represents the HTTP response from a formatter service
type ServiceFormatResponse struct {
	Success   bool   `json:"success"`
	Content   string `json:"content"`
	Changed   bool   `json:"changed"`
	Formatter string `json:"formatter"`
	Error     string `json:"error,omitempty"`
}

// ServiceHealthResponse represents the health check response
type ServiceHealthResponse struct {
	Status    string `json:"status"`
	Formatter string `json:"formatter"`
	Version   string `json:"version"`
	Error     string `json:"error,omitempty"`
}

// NewServiceFormatter creates a new service-based formatter
func NewServiceFormatter(
	metadata *formatters.FormatterMetadata,
	serviceURL string,
	timeout time.Duration,
	logger *logrus.Logger,
) *ServiceFormatter {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ServiceFormatter{
		metadata:   metadata,
		serviceURL: serviceURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// Name returns the formatter name
func (s *ServiceFormatter) Name() string {
	return s.metadata.Name
}

// Version returns the formatter version
func (s *ServiceFormatter) Version() string {
	return s.metadata.Version
}

// Languages returns supported languages
func (s *ServiceFormatter) Languages() []string {
	return s.metadata.Languages
}

// SupportsStdin returns whether stdin is supported (always true for service formatters)
func (s *ServiceFormatter) SupportsStdin() bool {
	return s.metadata.SupportsStdin
}

// SupportsInPlace returns whether in-place formatting is supported
func (s *ServiceFormatter) SupportsInPlace() bool {
	return s.metadata.SupportsInPlace
}

// SupportsCheck returns whether check mode is supported
func (s *ServiceFormatter) SupportsCheck() bool {
	return s.metadata.SupportsCheck
}

// SupportsConfig returns whether configuration is supported
func (s *ServiceFormatter) SupportsConfig() bool {
	return s.metadata.SupportsConfig
}

// Format formats the given code via HTTP service
func (s *ServiceFormatter) Format(ctx context.Context, req *formatters.FormatRequest) (*formatters.FormatResult, error) {
	start := time.Now()

	// Build HTTP request
	serviceReq := ServiceFormatRequest{
		Content: req.Content,
		Options: req.Config,
	}

	jsonData, err := json.Marshal(serviceReq)
	if err != nil {
		return &formatters.FormatResult{
			Success: false,
			Error:   fmt.Errorf("failed to marshal request: %w", err),
		}, err
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/format", s.serviceURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return &formatters.FormatResult{
			Success: false,
			Error:   fmt.Errorf("failed to create request: %w", err),
		}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute HTTP request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return &formatters.FormatResult{
			Success: false,
			Error:   fmt.Errorf("HTTP request failed: %w", err),
		}, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &formatters.FormatResult{
			Success: false,
			Error:   fmt.Errorf("failed to read response: %w", err),
		}, err
	}

	// Parse response
	var serviceResp ServiceFormatResponse
	if err := json.Unmarshal(body, &serviceResp); err != nil {
		wrappedErr := fmt.Errorf("failed to parse response: %w", err)
		return &formatters.FormatResult{
			Success: false,
			Error:   wrappedErr,
		}, wrappedErr
	}

	// Check for service-level errors
	if !serviceResp.Success {
		err := fmt.Errorf("formatter service error: %s", serviceResp.Error)
		return &formatters.FormatResult{
			Success: false,
			Error:   err,
		}, err
	}

	// Return successful result
	return &formatters.FormatResult{
		Success:       true,
		Content:       serviceResp.Content,
		Changed:       serviceResp.Changed,
		FormatterName: s.Name(),
		Duration:      time.Since(start),
	}, nil
}

// FormatBatch formats multiple files in a batch
func (s *ServiceFormatter) FormatBatch(ctx context.Context, reqs []*formatters.FormatRequest) ([]*formatters.FormatResult, error) {
	results := make([]*formatters.FormatResult, len(reqs))

	for i, req := range reqs {
		result, err := s.Format(ctx, req)
		if err != nil {
			results[i] = &formatters.FormatResult{
				Success: false,
				Error:   err,
			}
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// HealthCheck checks if the service is healthy
func (s *ServiceFormatter) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/health", s.serviceURL),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	// Parse health response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read health response: %w", err)
	}

	var healthResp ServiceHealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		return fmt.Errorf("failed to parse health response: %w", err)
	}

	if healthResp.Status != "healthy" {
		return fmt.Errorf("service unhealthy: %s", healthResp.Error)
	}

	s.logger.Debugf("Health check passed for %s (version: %s)", s.Name(), healthResp.Version)

	return nil
}

// ValidateConfig validates the formatter configuration
func (s *ServiceFormatter) ValidateConfig(config map[string]interface{}) error {
	// Service formatters typically don't require config validation
	return nil
}

// DefaultConfig returns the default configuration
func (s *ServiceFormatter) DefaultConfig() map[string]interface{} {
	return make(map[string]interface{})
}

// GetMetadata returns formatter metadata
func (s *ServiceFormatter) GetMetadata() *formatters.FormatterMetadata {
	return s.metadata
}
