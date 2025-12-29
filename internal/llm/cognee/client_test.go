package llm

import (
	"testing"
	"time"

	"github.com/superagent/superagent/internal/config"
	"github.com/superagent/superagent/internal/models"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-api-key",
		},
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.baseURL != "http://localhost:8080" {
		t.Fatalf("Expected baseURL to be 'http://localhost:8080', got %s", client.baseURL)
	}

	if client.apiKey != "test-api-key" {
		t.Fatalf("Expected apiKey to be 'test-api-key', got %s", client.apiKey)
	}

	if client.client == nil {
		t.Fatal("Expected HTTP client to be created")
	}

	if client.client.Timeout != 30*time.Second {
		t.Fatalf("Expected timeout to be 30s, got %v", client.client.Timeout)
	}
}

func TestNewClientWithEmptyConfig(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{},
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.baseURL != "" {
		t.Fatalf("Expected baseURL to be empty, got %s", client.baseURL)
	}

	if client.apiKey != "" {
		t.Fatalf("Expected apiKey to be empty, got %s", client.apiKey)
	}
}

func TestGetBaseURL(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: "http://test:8080",
		},
	}

	client := NewClient(cfg)
	baseURL := client.GetBaseURL()

	if baseURL != "http://test:8080" {
		t.Fatalf("Expected GetBaseURL to return 'http://test:8080', got %s", baseURL)
	}
}

func TestMemoryRequestFields(t *testing.T) {
	req := &MemoryRequest{
		Content:     "test content",
		DatasetName: "test-dataset",
		ContentType: "text/plain",
	}

	if req.Content != "test content" {
		t.Fatalf("Expected Content to be 'test content', got %s", req.Content)
	}

	if req.DatasetName != "test-dataset" {
		t.Fatalf("Expected DatasetName to be 'test-dataset', got %s", req.DatasetName)
	}

	if req.ContentType != "text/plain" {
		t.Fatalf("Expected ContentType to be 'text/plain', got %s", req.ContentType)
	}
}

func TestMemoryResponseFields(t *testing.T) {
	response := &MemoryResponse{
		VectorID: "vector-123",
		GraphNodes: map[string]interface{}{
			"node1": "value1",
		},
	}

	if response.VectorID != "vector-123" {
		t.Fatalf("Expected VectorID to be 'vector-123', got %s", response.VectorID)
	}

	if len(response.GraphNodes) != 1 {
		t.Fatalf("Expected GraphNodes to have 1 element, got %d", len(response.GraphNodes))
	}

	if response.GraphNodes["node1"] != "value1" {
		t.Fatalf("Expected GraphNodes['node1'] to be 'value1', got %v", response.GraphNodes["node1"])
	}
}

func TestSearchRequestFields(t *testing.T) {
	req := &SearchRequest{
		Query:       "test query",
		DatasetName: "test-dataset",
		Limit:       10,
	}

	if req.Query != "test query" {
		t.Fatalf("Expected Query to be 'test query', got %s", req.Query)
	}

	if req.DatasetName != "test-dataset" {
		t.Fatalf("Expected DatasetName to be 'test-dataset', got %s", req.DatasetName)
	}

	if req.Limit != 10 {
		t.Fatalf("Expected Limit to be 10, got %d", req.Limit)
	}
}

func TestSearchResponseFields(t *testing.T) {
	response := &SearchResponse{
		Results: []models.MemorySource{
			{
				DatasetName:    "test-dataset",
				Content:        "test content",
				RelevanceScore: 0.95,
				SourceType:     "text",
			},
		},
	}

	if len(response.Results) != 1 {
		t.Fatalf("Expected Results to have 1 element, got %d", len(response.Results))
	}

	if response.Results[0].DatasetName != "test-dataset" {
		t.Fatalf("Expected Results[0].DatasetName to be 'test-dataset', got %s", response.Results[0].DatasetName)
	}

	if response.Results[0].Content != "test content" {
		t.Fatalf("Expected Results[0].Content to be 'test content', got %s", response.Results[0].Content)
	}

	if response.Results[0].RelevanceScore != 0.95 {
		t.Fatalf("Expected Results[0].RelevanceScore to be 0.95, got %f", response.Results[0].RelevanceScore)
	}

	if response.Results[0].SourceType != "text" {
		t.Fatalf("Expected Results[0].SourceType to be 'text', got %s", response.Results[0].SourceType)
	}
}

func TestAutoContainerizeError(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: "http://invalid-url",
		},
	}

	client := NewClient(cfg)
	err := client.AutoContainerize()

	if err == nil {
		t.Fatal("Expected AutoContainerize to return an error")
	}

	expectedErr := "neither 'docker compose' nor 'docker-compose' found in PATH"
	if err.Error() != expectedErr {
		t.Fatalf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestStartCogneeContainerError(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: "http://invalid-url",
		},
	}

	client := NewClient(cfg)
	err := client.startCogneeContainer()

	if err == nil {
		t.Fatal("Expected startCogneeContainer to return an error")
	}

	expectedErr := "neither 'docker compose' nor 'docker-compose' found in PATH"
	if err.Error() != expectedErr {
		t.Fatalf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestTestConnection(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: "http://localhost:9999", // Non-existent URL
		},
	}

	client := NewClient(cfg)
	connected := client.testConnection()

	if connected {
		t.Fatal("Expected testConnection to return false for non-existent URL")
	}
}

func TestClientStructFields(t *testing.T) {
	cfg := &config.Config{
		Cognee: config.CogneeConfig{
			BaseURL: "http://test:8080",
			APIKey:  "test-key",
		},
	}

	client := NewClient(cfg)

	if client.baseURL != "http://test:8080" {
		t.Fatalf("Expected baseURL to be 'http://test:8080', got %s", client.baseURL)
	}

	if client.apiKey != "test-key" {
		t.Fatalf("Expected apiKey to be 'test-key', got %s", client.apiKey)
	}

	if client.client == nil {
		t.Fatal("Expected client to have HTTP client")
	}
}
