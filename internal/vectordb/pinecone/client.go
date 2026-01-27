// Package pinecone provides a client for Pinecone vector database.
package pinecone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Client provides an interface to interact with Pinecone vector database.
type Client struct {
	config     *Config
	httpClient *http.Client
	logger     *logrus.Logger
	mu         sync.RWMutex
	connected  bool
}

// Config holds Pinecone configuration.
type Config struct {
	APIKey      string        `json:"api_key"`
	Environment string        `json:"environment"` // e.g., "us-west1-gcp"
	ProjectID   string        `json:"project_id"`  // Optional for serverless
	IndexHost   string        `json:"index_host"`  // Full index host URL
	Timeout     time.Duration `json:"timeout"`
}

// DefaultConfig returns default Pinecone configuration.
func DefaultConfig() *Config {
	return &Config{
		Timeout: 30 * time.Second,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.IndexHost == "" {
		return fmt.Errorf("index host is required")
	}
	return nil
}

// NewClient creates a new Pinecone client.
func NewClient(config *Config, logger *logrus.Logger) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:    logger,
		connected: false,
	}, nil
}

// Connect verifies connectivity to Pinecone.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.describeIndexStats(ctx); err != nil {
		return fmt.Errorf("failed to connect to Pinecone: %w", err)
	}

	c.connected = true
	c.logger.Info("Connected to Pinecone")
	return nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return nil
}

// IsConnected returns whether the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// HealthCheck checks the health of Pinecone.
func (c *Client) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.describeIndexStats(ctx)
}

func (c *Client) describeIndexStats(ctx context.Context) error {
	_, err := c.doRequest(ctx, http.MethodPost, "/describe_index_stats", nil)
	return err
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.config.IndexHost, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Vector represents a vector in Pinecone.
type Vector struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ScoredVector represents a search result with score.
type ScoredVector struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Values   []float32              `json:"values,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpsertRequest represents an upsert request.
type UpsertRequest struct {
	Vectors   []Vector `json:"vectors"`
	Namespace string   `json:"namespace,omitempty"`
}

// UpsertResponse represents an upsert response.
type UpsertResponse struct {
	UpsertedCount int `json:"upsertedCount"`
}

// Upsert inserts or updates vectors in Pinecone.
func (c *Client) Upsert(ctx context.Context, vectors []Vector, namespace string) (*UpsertResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pinecone")
	}

	if len(vectors) == 0 {
		return &UpsertResponse{}, nil
	}

	// Ensure all vectors have IDs
	for i := range vectors {
		if vectors[i].ID == "" {
			vectors[i].ID = uuid.New().String()
		}
	}

	reqBody := UpsertRequest{
		Vectors:   vectors,
		Namespace: namespace,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/vectors/upsert", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert vectors: %w", err)
	}

	var result UpsertResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"namespace": namespace,
		"count":     result.UpsertedCount,
	}).Debug("Vectors upserted")

	return &result, nil
}

// QueryRequest represents a query request.
type QueryRequest struct {
	Vector          []float32              `json:"vector,omitempty"`
	ID              string                 `json:"id,omitempty"`
	TopK            int                    `json:"topK"`
	Namespace       string                 `json:"namespace,omitempty"`
	Filter          map[string]interface{} `json:"filter,omitempty"`
	IncludeValues   bool                   `json:"includeValues"`
	IncludeMetadata bool                   `json:"includeMetadata"`
}

// QueryResponse represents a query response.
type QueryResponse struct {
	Matches   []ScoredVector `json:"matches"`
	Namespace string         `json:"namespace"`
}

// Query performs a vector similarity search.
func (c *Client) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pinecone")
	}

	if req.TopK <= 0 {
		req.TopK = 10
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/query", req)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	var result QueryResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DeleteRequest represents a delete request.
type DeleteRequest struct {
	IDs       []string               `json:"ids,omitempty"`
	DeleteAll bool                   `json:"deleteAll,omitempty"`
	Namespace string                 `json:"namespace,omitempty"`
	Filter    map[string]interface{} `json:"filter,omitempty"`
}

// Delete removes vectors from Pinecone.
func (c *Client) Delete(ctx context.Context, req *DeleteRequest) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Pinecone")
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/vectors/delete", req)
	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"namespace": req.Namespace,
		"ids":       len(req.IDs),
		"deleteAll": req.DeleteAll,
	}).Debug("Vectors deleted")

	return nil
}

// FetchRequest represents a fetch request.
type FetchRequest struct {
	IDs       []string `json:"ids"`
	Namespace string   `json:"namespace,omitempty"`
}

// FetchResponse represents a fetch response.
type FetchResponse struct {
	Vectors   map[string]Vector `json:"vectors"`
	Namespace string            `json:"namespace"`
}

// Fetch retrieves vectors by IDs.
func (c *Client) Fetch(ctx context.Context, ids []string, namespace string) (*FetchResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pinecone")
	}

	if len(ids) == 0 {
		return &FetchResponse{Vectors: make(map[string]Vector)}, nil
	}

	// Build query string
	path := "/vectors/fetch?"
	for i, id := range ids {
		if i > 0 {
			path += "&"
		}
		path += "ids=" + id
	}
	if namespace != "" {
		path += "&namespace=" + namespace
	}

	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vectors: %w", err)
	}

	var result FetchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DescribeIndexStatsResponse represents index stats.
type DescribeIndexStatsResponse struct {
	Namespaces       map[string]NamespaceStats `json:"namespaces"`
	Dimension        int                       `json:"dimension"`
	IndexFullness    float64                   `json:"indexFullness"`
	TotalVectorCount int64                     `json:"totalVectorCount"`
}

// NamespaceStats represents stats for a namespace.
type NamespaceStats struct {
	VectorCount int64 `json:"vectorCount"`
}

// DescribeIndexStats returns statistics about the index.
func (c *Client) DescribeIndexStats(ctx context.Context, filter map[string]interface{}) (*DescribeIndexStatsResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pinecone")
	}

	var body interface{}
	if filter != nil {
		body = map[string]interface{}{"filter": filter}
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/describe_index_stats", body)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index stats: %w", err)
	}

	var result DescribeIndexStatsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ListNamespaces returns all namespaces in the index.
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	stats, err := c.DescribeIndexStats(ctx, nil)
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(stats.Namespaces))
	for ns := range stats.Namespaces {
		namespaces = append(namespaces, ns)
	}
	return namespaces, nil
}

// UpdateVector updates a vector's metadata.
type UpdateRequest struct {
	ID          string                 `json:"id"`
	Values      []float32              `json:"values,omitempty"`
	SetMetadata map[string]interface{} `json:"setMetadata,omitempty"`
	Namespace   string                 `json:"namespace,omitempty"`
}

// Update updates a vector's values or metadata.
func (c *Client) Update(ctx context.Context, req *UpdateRequest) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Pinecone")
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/vectors/update", req)
	if err != nil {
		return fmt.Errorf("failed to update vector: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"id":        req.ID,
		"namespace": req.Namespace,
	}).Debug("Vector updated")

	return nil
}
