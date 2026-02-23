package cloud_test

import (
	"context"
	"testing"
	"time"

	adapter "dev.helix.agent/internal/adapters/cloud"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// CloudIntegrationManager Tests
// ============================================================================

func TestNewCloudIntegrationManager(t *testing.T) {
	manager := adapter.NewCloudIntegrationManager(nil)
	require.NotNil(t, manager)

	// With logger
	logger := logrus.New()
	manager2 := adapter.NewCloudIntegrationManager(logger)
	require.NotNil(t, manager2)
}

func TestCloudIntegrationManager_RegisterAndGet(t *testing.T) {
	manager := adapter.NewCloudIntegrationManager(nil)

	awsProvider := adapter.NewAWSBedrockIntegration("us-east-1", nil)
	manager.RegisterProvider(awsProvider)

	provider, err := manager.GetProvider("aws-bedrock")
	require.NoError(t, err)
	assert.Equal(t, "aws-bedrock", provider.GetProviderName())
}

func TestCloudIntegrationManager_GetProvider_NotFound(t *testing.T) {
	manager := adapter.NewCloudIntegrationManager(nil)

	_, err := manager.GetProvider("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestCloudIntegrationManager_ListAllProviders(t *testing.T) {
	manager := adapter.NewCloudIntegrationManager(nil)

	// Initially empty
	names := manager.ListAllProviders()
	assert.Empty(t, names)

	// Register providers
	manager.RegisterProvider(adapter.NewAWSBedrockIntegration("us-east-1", nil))
	manager.RegisterProvider(adapter.NewGCPVertexAIIntegration("my-project", "us-central1", nil))

	names = manager.ListAllProviders()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "aws-bedrock")
	assert.Contains(t, names, "gcp-vertex-ai")
}

func TestCloudIntegrationManager_HealthCheckAll(t *testing.T) {
	manager := adapter.NewCloudIntegrationManager(nil)
	manager.RegisterProvider(adapter.NewAWSBedrockIntegration("us-east-1", nil))

	ctx := context.Background()
	results := manager.HealthCheckAll(ctx)

	assert.Contains(t, results, "aws-bedrock")
	// No credentials → expect an error
	assert.Error(t, results["aws-bedrock"])
}

// ============================================================================
// AWSBedrockIntegration Tests
// ============================================================================

func TestNewAWSBedrockIntegration(t *testing.T) {
	provider := adapter.NewAWSBedrockIntegration("us-east-1", nil)
	require.NotNil(t, provider)
	assert.Equal(t, "aws-bedrock", provider.GetProviderName())
}

func TestNewAWSBedrockIntegrationWithConfig(t *testing.T) {
	cfg := adapter.AWSBedrockConfig{
		Region:          "eu-west-1",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		Timeout:         30 * time.Second,
	}
	provider := adapter.NewAWSBedrockIntegrationWithConfig(cfg, nil)
	require.NotNil(t, provider)
	assert.Equal(t, "aws-bedrock", provider.GetProviderName())
}

func TestAWSBedrockIntegration_ListModels_NoCredentials(t *testing.T) {
	provider := adapter.NewAWSBedrockIntegration("us-east-1", nil)
	_, err := provider.ListModels(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AWS credentials")
}

func TestAWSBedrockIntegration_ListModels_WithCredentials(t *testing.T) {
	cfg := adapter.AWSBedrockConfig{
		Region:          "us-east-1",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
	}
	provider := adapter.NewAWSBedrockIntegrationWithConfig(cfg, nil)
	models, err := provider.ListModels(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, models)
}

func TestAWSBedrockIntegration_InvokeModel_NoCredentials(t *testing.T) {
	provider := adapter.NewAWSBedrockIntegration("us-east-1", nil)
	_, err := provider.InvokeModel(context.Background(), "model", "prompt", nil)
	assert.Error(t, err)
}

func TestAWSBedrockIntegration_HealthCheck_NoCredentials(t *testing.T) {
	provider := adapter.NewAWSBedrockIntegration("us-east-1", nil)
	err := provider.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestAWSBedrockIntegration_HealthCheck_WithCredentials(t *testing.T) {
	cfg := adapter.AWSBedrockConfig{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
	}
	provider := adapter.NewAWSBedrockIntegrationWithConfig(cfg, nil)
	err := provider.HealthCheck(context.Background())
	assert.NoError(t, err)
}

// ============================================================================
// GCPVertexAIIntegration Tests
// ============================================================================

func TestNewGCPVertexAIIntegration(t *testing.T) {
	provider := adapter.NewGCPVertexAIIntegration("my-project", "us-central1", nil)
	require.NotNil(t, provider)
	assert.Equal(t, "gcp-vertex-ai", provider.GetProviderName())
}

func TestGCPVertexAIIntegration_ListModels_NoProjectID(t *testing.T) {
	provider := adapter.NewGCPVertexAIIntegration("", "us-central1", nil)
	_, err := provider.ListModels(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project ID")
}

func TestGCPVertexAIIntegration_ListModels_WithProjectID(t *testing.T) {
	provider := adapter.NewGCPVertexAIIntegration("my-project", "us-central1", nil)
	models, err := provider.ListModels(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, models)
}

func TestGCPVertexAIIntegration_InvokeModel_NoProjectID(t *testing.T) {
	provider := adapter.NewGCPVertexAIIntegration("", "us-central1", nil)
	_, err := provider.InvokeModel(context.Background(), "model", "prompt", nil)
	assert.Error(t, err)
}

func TestGCPVertexAIIntegration_HealthCheck_MissingConfig(t *testing.T) {
	provider := adapter.NewGCPVertexAIIntegration("", "", nil)
	err := provider.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestGCPVertexAIIntegration_HealthCheck_NoAccessToken(t *testing.T) {
	provider := adapter.NewGCPVertexAIIntegration("my-project", "us-central1", nil)
	err := provider.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access token")
}

func TestGCPVertexAIIntegration_HealthCheck_WithToken(t *testing.T) {
	cfg := adapter.GCPVertexAIConfig{
		ProjectID:   "my-project",
		Location:    "us-central1",
		AccessToken: "token-123",
	}
	provider := adapter.NewGCPVertexAIIntegrationWithConfig(cfg, nil)
	err := provider.HealthCheck(context.Background())
	assert.NoError(t, err)
}

// ============================================================================
// AzureOpenAIIntegration Tests
// ============================================================================

func TestNewAzureOpenAIIntegration(t *testing.T) {
	provider := adapter.NewAzureOpenAIIntegration("https://myresource.openai.azure.com", nil)
	require.NotNil(t, provider)
	assert.Equal(t, "azure-openai", provider.GetProviderName())
}

func TestAzureOpenAIIntegration_ListModels_NoEndpoint(t *testing.T) {
	provider := adapter.NewAzureOpenAIIntegration("", nil)
	_, err := provider.ListModels(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

func TestAzureOpenAIIntegration_ListModels_WithEndpoint(t *testing.T) {
	provider := adapter.NewAzureOpenAIIntegration("https://myresource.openai.azure.com", nil)
	models, err := provider.ListModels(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, models)
}

func TestAzureOpenAIIntegration_HealthCheck_MissingConfig(t *testing.T) {
	provider := adapter.NewAzureOpenAIIntegration("", nil)
	err := provider.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestAzureOpenAIIntegration_HealthCheck_NoAPIKey(t *testing.T) {
	provider := adapter.NewAzureOpenAIIntegration("https://myresource.openai.azure.com", nil)
	err := provider.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestAzureOpenAIIntegration_HealthCheck_WithAPIKey(t *testing.T) {
	cfg := adapter.AzureOpenAIConfig{
		Endpoint: "https://myresource.openai.azure.com",
		APIKey:   "my-api-key",
	}
	provider := adapter.NewAzureOpenAIIntegrationWithConfig(cfg, nil)
	err := provider.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestAzureOpenAIIntegration_InvokeModel_NoEndpoint(t *testing.T) {
	provider := adapter.NewAzureOpenAIIntegration("", nil)
	_, err := provider.InvokeModel(context.Background(), "gpt-4", "hello", nil)
	assert.Error(t, err)
}
