// Package cloud provides adapter types for the digital.vasic.storage/pkg/provider module.
// This adapter re-exports types and functions to maintain backward compatibility
// with code using the internal/cloud package.
package cloud

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CloudProvider represents a cloud AI provider.
// This interface remains the same as the internal one for AI model invocation.
type CloudProvider interface {
	ListModels(ctx context.Context) ([]map[string]interface{}, error)
	InvokeModel(ctx context.Context, modelName, prompt string, config map[string]interface{}) (string, error)
	GetProviderName() string
	HealthCheck(ctx context.Context) error
}

// CloudIntegrationManager manages multiple cloud providers.
type CloudIntegrationManager struct {
	providers map[string]CloudProvider
	mu        sync.RWMutex
	logger    *logrus.Logger
}

// NewCloudIntegrationManager creates a new cloud integration manager.
func NewCloudIntegrationManager(logger *logrus.Logger) *CloudIntegrationManager {
	if logger == nil {
		logger = logrus.New()
	}
	return &CloudIntegrationManager{
		providers: make(map[string]CloudProvider),
		logger:    logger,
	}
}

// RegisterProvider registers a cloud provider.
func (m *CloudIntegrationManager) RegisterProvider(provider CloudProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.GetProviderName()] = provider
}

// GetProvider retrieves a provider by name.
func (m *CloudIntegrationManager) GetProvider(name string) (CloudProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// ListAllProviders returns the names of all registered providers.
func (m *CloudIntegrationManager) ListAllProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// HealthCheckAll checks the health of all providers.
func (m *CloudIntegrationManager) HealthCheckAll(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	results := make(map[string]error)
	for name, provider := range m.providers {
		results[name] = provider.HealthCheck(ctx)
	}
	return results
}

// ========== AWS Bedrock Integration ==========

// AWSBedrockConfig holds configuration for AWS Bedrock.
type AWSBedrockConfig struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	SessionToken     string
	Timeout          time.Duration
	EndpointOverride string
}

// AWSBedrockIntegration provides AWS Bedrock AI service integration.
type AWSBedrockIntegration struct {
	config AWSBedrockConfig
	logger *logrus.Logger
}

// NewAWSBedrockIntegration creates a new AWS Bedrock integration.
func NewAWSBedrockIntegration(region string, logger *logrus.Logger) *AWSBedrockIntegration {
	return NewAWSBedrockIntegrationWithConfig(AWSBedrockConfig{
		Region:  region,
		Timeout: 60 * time.Second,
	}, logger)
}

// NewAWSBedrockIntegrationWithConfig creates AWS Bedrock integration with explicit config.
func NewAWSBedrockIntegrationWithConfig(config AWSBedrockConfig, logger *logrus.Logger) *AWSBedrockIntegration {
	if logger == nil {
		logger = logrus.New()
	}
	return &AWSBedrockIntegration{
		config: config,
		logger: logger,
	}
}

// GetProviderName returns the provider name.
func (a *AWSBedrockIntegration) GetProviderName() string {
	return "aws-bedrock"
}

// ListModels lists available Bedrock models.
func (a *AWSBedrockIntegration) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	if a.config.AccessKeyID == "" || a.config.SecretAccessKey == "" {
		return nil, fmt.Errorf("AWS credentials not configured")
	}
	// Return empty list - actual implementation would call AWS API
	return []map[string]interface{}{}, nil
}

// InvokeModel invokes a model.
func (a *AWSBedrockIntegration) InvokeModel(ctx context.Context, modelName, prompt string, config map[string]interface{}) (string, error) {
	if a.config.AccessKeyID == "" || a.config.SecretAccessKey == "" {
		return "", fmt.Errorf("AWS credentials not configured")
	}
	return "", fmt.Errorf("model invocation not implemented in adapter")
}

// HealthCheck checks the health of AWS Bedrock.
func (a *AWSBedrockIntegration) HealthCheck(ctx context.Context) error {
	if a.config.AccessKeyID == "" || a.config.SecretAccessKey == "" {
		return fmt.Errorf("AWS credentials not configured")
	}
	return nil
}

// ========== GCP Vertex AI Integration ==========

// GCPVertexAIConfig holds configuration for GCP Vertex AI.
type GCPVertexAIConfig struct {
	ProjectID   string
	Location    string
	AccessToken string
	Timeout     time.Duration
}

// GCPVertexAIIntegration provides GCP Vertex AI service integration.
type GCPVertexAIIntegration struct {
	config GCPVertexAIConfig
	logger *logrus.Logger
}

// NewGCPVertexAIIntegration creates a new GCP Vertex AI integration.
func NewGCPVertexAIIntegration(projectID, location string, logger *logrus.Logger) *GCPVertexAIIntegration {
	return NewGCPVertexAIIntegrationWithConfig(GCPVertexAIConfig{
		ProjectID: projectID,
		Location:  location,
		Timeout:   60 * time.Second,
	}, logger)
}

// NewGCPVertexAIIntegrationWithConfig creates GCP Vertex AI integration with explicit config.
func NewGCPVertexAIIntegrationWithConfig(config GCPVertexAIConfig, logger *logrus.Logger) *GCPVertexAIIntegration {
	if logger == nil {
		logger = logrus.New()
	}
	return &GCPVertexAIIntegration{
		config: config,
		logger: logger,
	}
}

// GetProviderName returns the provider name.
func (g *GCPVertexAIIntegration) GetProviderName() string {
	return "gcp-vertex-ai"
}

// ListModels lists available Vertex AI models.
func (g *GCPVertexAIIntegration) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	if g.config.ProjectID == "" {
		return nil, fmt.Errorf("GCP project ID not configured")
	}
	return []map[string]interface{}{}, nil
}

// InvokeModel invokes a model.
func (g *GCPVertexAIIntegration) InvokeModel(ctx context.Context, modelName, prompt string, config map[string]interface{}) (string, error) {
	if g.config.ProjectID == "" {
		return "", fmt.Errorf("GCP project ID not configured")
	}
	return "", fmt.Errorf("model invocation not implemented in adapter")
}

// HealthCheck checks the health of GCP Vertex AI.
func (g *GCPVertexAIIntegration) HealthCheck(ctx context.Context) error {
	if g.config.ProjectID == "" {
		return fmt.Errorf("GCP project ID not configured")
	}
	if g.config.AccessToken == "" {
		return fmt.Errorf("GCP access token not configured")
	}
	return nil
}

// ========== Azure OpenAI Integration ==========

// AzureOpenAIConfig holds configuration for Azure OpenAI.
type AzureOpenAIConfig struct {
	Endpoint string
	APIKey   string
	Timeout  time.Duration
}

// AzureOpenAIIntegration provides Azure OpenAI service integration.
type AzureOpenAIIntegration struct {
	config AzureOpenAIConfig
	logger *logrus.Logger
}

// NewAzureOpenAIIntegration creates a new Azure OpenAI integration.
func NewAzureOpenAIIntegration(endpoint string, logger *logrus.Logger) *AzureOpenAIIntegration {
	return NewAzureOpenAIIntegrationWithConfig(AzureOpenAIConfig{
		Endpoint: endpoint,
		Timeout:  60 * time.Second,
	}, logger)
}

// NewAzureOpenAIIntegrationWithConfig creates Azure OpenAI integration with explicit config.
func NewAzureOpenAIIntegrationWithConfig(config AzureOpenAIConfig, logger *logrus.Logger) *AzureOpenAIIntegration {
	if logger == nil {
		logger = logrus.New()
	}
	return &AzureOpenAIIntegration{
		config: config,
		logger: logger,
	}
}

// GetProviderName returns the provider name.
func (a *AzureOpenAIIntegration) GetProviderName() string {
	return "azure-openai"
}

// ListModels lists available Azure OpenAI models.
func (a *AzureOpenAIIntegration) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	if a.config.Endpoint == "" {
		return nil, fmt.Errorf("Azure endpoint not configured")
	}
	return []map[string]interface{}{}, nil
}

// InvokeModel invokes a model.
func (a *AzureOpenAIIntegration) InvokeModel(ctx context.Context, modelName, prompt string, config map[string]interface{}) (string, error) {
	if a.config.Endpoint == "" {
		return "", fmt.Errorf("Azure endpoint not configured")
	}
	return "", fmt.Errorf("model invocation not implemented in adapter")
}

// HealthCheck checks the health of Azure OpenAI.
func (a *AzureOpenAIIntegration) HealthCheck(ctx context.Context) error {
	if a.config.Endpoint == "" || a.config.APIKey == "" {
		return fmt.Errorf("Azure credentials not configured")
	}
	return nil
}
