package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type invokeModelFunc func(ctx context.Context, modelName, prompt string, cfg map[string]interface{}) (string, error)

var (
	DefaultAWSBedrockInvoker  func(AWSBedrockConfig, *logrus.Logger) invokeModelFunc
	DefaultGCPVertexAIInvoker func(GCPVertexAIConfig, *logrus.Logger) invokeModelFunc
	DefaultAzureOpenAIInvoker func(AzureOpenAIConfig, *logrus.Logger) invokeModelFunc
)

type awsBedrockClient struct {
	config AWSBedrockConfig
	logger *logrus.Logger
	client *http.Client
}

func newAWSBedrockClient(config AWSBedrockConfig, logger *logrus.Logger) *awsBedrockClient {
	return &awsBedrockClient{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (c *awsBedrockClient) invokeModel(ctx context.Context, modelName, prompt string, cfg map[string]interface{}) (string, error) {
	if c.config.AccessKeyID == "" || c.config.SecretAccessKey == "" {
		return "", fmt.Errorf("AWS credentials not configured")
	}

	region := c.config.Region
	if region == "" {
		region = "us-east-1"
	}

	endpoint := c.config.EndpointOverride
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)
	}

	requestBody := map[string]interface{}{
		"prompt": prompt,
	}

	if maxTokens, ok := cfg["max_tokens"].(int); ok {
		requestBody["max_gen_len"] = maxTokens
	}
	if temperature, ok := cfg["temperature"].(float64); ok {
		requestBody["temperature"] = temperature
	}
	if topP, ok := cfg["top_p"].(float64); ok {
		requestBody["top_p"] = topP
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/model/%s/invoke", endpoint, modelName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.WithFields(logrus.Fields{
		"model":  modelName,
		"region": region,
	}).Debug("invoking AWS Bedrock model")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AWS Bedrock error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Completion string `json:"completion"`
		Generation string `json:"generation"`
		Output     struct {
			Text string `json:"text"`
		} `json:"output"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Completion != "" {
		return result.Completion, nil
	}
	if result.Generation != "" {
		return result.Generation, nil
	}
	if result.Output.Text != "" {
		return result.Output.Text, nil
	}

	return string(respBody), nil
}

type gcpVertexAIClient struct {
	config GCPVertexAIConfig
	logger *logrus.Logger
	client *http.Client
}

func newGCPVertexAIClient(config GCPVertexAIConfig, logger *logrus.Logger) *gcpVertexAIClient {
	return &gcpVertexAIClient{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (c *gcpVertexAIClient) invokeModel(ctx context.Context, modelName, prompt string, cfg map[string]interface{}) (string, error) {
	if c.config.ProjectID == "" {
		return "", fmt.Errorf("GCP project ID not configured")
	}
	if c.config.AccessToken == "" {
		return "", fmt.Errorf("GCP access token not configured")
	}

	location := c.config.Location
	if location == "" {
		location = "us-central1"
	}

	endpoint := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		location, c.config.ProjectID, location, modelName,
	)

	instances := []map[string]interface{}{
		{
			"prompt": prompt,
		},
	}

	if maxTokens, ok := cfg["max_tokens"].(int); ok {
		instances[0]["maxOutputTokens"] = maxTokens
	}
	if temperature, ok := cfg["temperature"].(float64); ok {
		instances[0]["temperature"] = temperature
	}
	if topP, ok := cfg["top_p"].(float64); ok {
		instances[0]["topP"] = topP
	}

	requestBody := map[string]interface{}{
		"instances": instances,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)

	c.logger.WithFields(logrus.Fields{
		"model":    modelName,
		"project":  c.config.ProjectID,
		"location": location,
	}).Debug("invoking GCP Vertex AI model")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GCP Vertex AI error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Predictions []struct {
			Content string `json:"content"`
		} `json:"predictions"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Predictions) > 0 && result.Predictions[0].Content != "" {
		return result.Predictions[0].Content, nil
	}

	return string(respBody), nil
}

type azureOpenAIClient struct {
	config AzureOpenAIConfig
	logger *logrus.Logger
	client *http.Client
}

func newAzureOpenAIClient(config AzureOpenAIConfig, logger *logrus.Logger) *azureOpenAIClient {
	return &azureOpenAIClient{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (c *azureOpenAIClient) invokeModel(ctx context.Context, modelName, prompt string, cfg map[string]interface{}) (string, error) {
	if c.config.Endpoint == "" {
		return "", fmt.Errorf("Azure endpoint not configured")
	}
	if c.config.APIKey == "" {
		return "", fmt.Errorf("Azure API key not configured")
	}

	messages := []map[string]string{
		{"role": "user", "content": prompt},
	}

	requestBody := map[string]interface{}{
		"messages": messages,
	}

	if maxTokens, ok := cfg["max_tokens"].(int); ok {
		requestBody["max_tokens"] = maxTokens
	}
	if temperature, ok := cfg["temperature"].(float64); ok {
		requestBody["temperature"] = temperature
	}
	if topP, ok := cfg["top_p"].(float64); ok {
		requestBody["top_p"] = topP
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview", c.config.Endpoint, modelName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.config.APIKey)

	c.logger.WithFields(logrus.Fields{
		"model":    modelName,
		"endpoint": c.config.Endpoint,
	}).Debug("invoking Azure OpenAI model")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Azure OpenAI error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) > 0 && result.Choices[0].Message.Content != "" {
		return result.Choices[0].Message.Content, nil
	}

	return string(respBody), nil
}

func init() {
	DefaultAWSBedrockInvoker = func(config AWSBedrockConfig, logger *logrus.Logger) invokeModelFunc {
		client := newAWSBedrockClient(config, logger)
		return client.invokeModel
	}
	DefaultGCPVertexAIInvoker = func(config GCPVertexAIConfig, logger *logrus.Logger) invokeModelFunc {
		client := newGCPVertexAIClient(config, logger)
		return client.invokeModel
	}
	DefaultAzureOpenAIInvoker = func(config AzureOpenAIConfig, logger *logrus.Logger) invokeModelFunc {
		client := newAzureOpenAIClient(config, logger)
		return client.invokeModel
	}
}
