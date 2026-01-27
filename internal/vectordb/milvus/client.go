// Package milvus provides a client for Milvus vector database.
package milvus

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

// Client provides an interface to interact with Milvus vector database.
type Client struct {
	config     *Config
	httpClient *http.Client
	logger     *logrus.Logger
	mu         sync.RWMutex
	connected  bool
}

// Config holds Milvus configuration.
type Config struct {
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Username string        `json:"username,omitempty"`
	Password string        `json:"password,omitempty"`
	DBName   string        `json:"db_name"`
	Secure   bool          `json:"secure"`
	Token    string        `json:"token,omitempty"` // For Zilliz Cloud
	Timeout  time.Duration `json:"timeout"`
}

// DefaultConfig returns default Milvus configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:    "localhost",
		Port:    19530,
		DBName:  "default",
		Timeout: 30 * time.Second,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("invalid port")
	}
	return nil
}

// GetBaseURL returns the base URL for the Milvus REST API.
func (c *Config) GetBaseURL() string {
	scheme := "http"
	if c.Secure {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/v2/vectordb", scheme, c.Host, c.Port)
}

// NewClient creates a new Milvus client.
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

// Connect verifies connectivity to Milvus.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.healthCheckLocked(ctx); err != nil {
		return fmt.Errorf("failed to connect to Milvus: %w", err)
	}

	c.connected = true
	c.logger.Info("Connected to Milvus")
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

// HealthCheck checks the health of Milvus.
func (c *Client) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.healthCheckLocked(ctx)
}

func (c *Client) healthCheckLocked(ctx context.Context) error {
	_, err := c.ListCollections(ctx)
	return err
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.config.GetBaseURL(), path)

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
	req.Header.Set("Accept", "application/json")

	// Add authentication
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	} else if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

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

	// Check for API-level errors
	var apiResp struct {
		Code    int    `json:"code"`
		Message string `json:"message,omitempty"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Code != 0 {
		return nil, fmt.Errorf("API error %d: %s", apiResp.Code, apiResp.Message)
	}

	return respBody, nil
}

// CollectionSchema defines a collection schema.
type CollectionSchema struct {
	CollectionName string        `json:"collectionName"`
	Description    string        `json:"description,omitempty"`
	Fields         []FieldSchema `json:"fields"`
}

// FieldSchema defines a field in a collection.
type FieldSchema struct {
	FieldName    string                 `json:"fieldName"`
	DataType     DataType               `json:"dataType"`
	IsPrimaryKey bool                   `json:"isPrimaryKey,omitempty"`
	IsPartition  bool                   `json:"isPartitionKey,omitempty"`
	ElementType  DataType               `json:"elementDataType,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
}

// DataType represents Milvus data types.
type DataType string

const (
	DataTypeInt64        DataType = "Int64"
	DataTypeVarChar      DataType = "VarChar"
	DataTypeFloat        DataType = "Float"
	DataTypeDouble       DataType = "Double"
	DataTypeBool         DataType = "Bool"
	DataTypeJSON         DataType = "JSON"
	DataTypeFloatVector  DataType = "FloatVector"
	DataTypeBinaryVector DataType = "BinaryVector"
)

// IndexType represents Milvus index types.
type IndexType string

const (
	IndexTypeIVFFlat   IndexType = "IVF_FLAT"
	IndexTypeIVFSQ8    IndexType = "IVF_SQ8"
	IndexTypeIVFPQ     IndexType = "IVF_PQ"
	IndexTypeHNSW      IndexType = "HNSW"
	IndexTypeAutoIndex IndexType = "AUTOINDEX"
)

// MetricType represents distance metrics.
type MetricType string

const (
	MetricTypeL2     MetricType = "L2"
	MetricTypeIP     MetricType = "IP"
	MetricTypeCosine MetricType = "COSINE"
)

// CreateCollectionRequest represents a create collection request.
type CreateCollectionRequest struct {
	DBName         string           `json:"dbName,omitempty"`
	CollectionName string           `json:"collectionName"`
	Schema         CollectionSchema `json:"schema,omitempty"`
	Dimension      int              `json:"dimension,omitempty"` // For quick setup
	MetricType     MetricType       `json:"metricType,omitempty"`
	PrimaryField   string           `json:"primaryFieldName,omitempty"`
	VectorField    string           `json:"vectorFieldName,omitempty"`
	IDType         DataType         `json:"idType,omitempty"`
}

// CreateCollection creates a new collection.
func (c *Client) CreateCollection(ctx context.Context, req *CreateCollectionRequest) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Milvus")
	}

	if req.DBName == "" {
		req.DBName = c.config.DBName
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/collections/create", req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	c.logger.WithField("collection", req.CollectionName).Info("Collection created")
	return nil
}

// DropCollection drops a collection.
func (c *Client) DropCollection(ctx context.Context, collectionName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Milvus")
	}

	req := map[string]interface{}{
		"dbName":         c.config.DBName,
		"collectionName": collectionName,
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/collections/drop", req)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	c.logger.WithField("collection", collectionName).Info("Collection dropped")
	return nil
}

// ListCollections returns all collections.
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	req := map[string]interface{}{
		"dbName": c.config.DBName,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/collections/list", req)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	var response struct {
		Data []string `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

// DescribeCollection returns collection details.
func (c *Client) DescribeCollection(ctx context.Context, collectionName string) (*CollectionInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Milvus")
	}

	req := map[string]interface{}{
		"dbName":         c.config.DBName,
		"collectionName": collectionName,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/collections/describe", req)
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %w", err)
	}

	var response struct {
		Data CollectionInfo `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response.Data, nil
}

// CollectionInfo represents collection information.
type CollectionInfo struct {
	CollectionName string `json:"collectionName"`
	Description    string `json:"description"`
	Fields         []struct {
		FieldName    string                 `json:"fieldName"`
		DataType     DataType               `json:"dataType"`
		IsPrimaryKey bool                   `json:"isPrimaryKey"`
		Params       map[string]interface{} `json:"params"`
	} `json:"fields"`
	Indexes []struct {
		FieldName  string                 `json:"fieldName"`
		IndexName  string                 `json:"indexName"`
		IndexType  IndexType              `json:"indexType"`
		MetricType MetricType             `json:"metricType"`
		Params     map[string]interface{} `json:"params"`
	} `json:"indexes"`
	Load      string `json:"load"`
	ShardsNum int    `json:"shardsNum"`
}

// Entity represents a vector entity in Milvus.
type Entity struct {
	ID     interface{}            `json:"id,omitempty"`
	Vector []float32              `json:"vector,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// InsertRequest represents an insert request.
type InsertRequest struct {
	DBName         string                   `json:"dbName,omitempty"`
	CollectionName string                   `json:"collectionName"`
	Data           []map[string]interface{} `json:"data"`
}

// InsertResponse represents an insert response.
type InsertResponse struct {
	InsertCount int      `json:"insertCount"`
	InsertIDs   []string `json:"insertIds"`
}

// Insert inserts vectors into a collection.
func (c *Client) Insert(ctx context.Context, collectionName string, data []map[string]interface{}) (*InsertResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Milvus")
	}

	if len(data) == 0 {
		return &InsertResponse{}, nil
	}

	// Ensure all records have IDs
	for i := range data {
		if _, ok := data[i]["id"]; !ok {
			data[i]["id"] = uuid.New().String()
		}
	}

	req := InsertRequest{
		DBName:         c.config.DBName,
		CollectionName: collectionName,
		Data:           data,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/entities/insert", req)
	if err != nil {
		return nil, fmt.Errorf("failed to insert entities: %w", err)
	}

	var response struct {
		Data InsertResponse `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"collection": collectionName,
		"count":      response.Data.InsertCount,
	}).Debug("Entities inserted")

	return &response.Data, nil
}

// SearchRequest represents a search request.
type SearchRequest struct {
	DBName         string                 `json:"dbName,omitempty"`
	CollectionName string                 `json:"collectionName"`
	Data           [][]float32            `json:"data"` // Query vectors
	AnnsField      string                 `json:"annsField,omitempty"`
	Limit          int                    `json:"limit"`
	Offset         int                    `json:"offset,omitempty"`
	OutputFields   []string               `json:"outputFields,omitempty"`
	Filter         string                 `json:"filter,omitempty"`
	SearchParams   map[string]interface{} `json:"searchParams,omitempty"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	ID       interface{}            `json:"id"`
	Distance float32                `json:"distance"`
	Entity   map[string]interface{} `json:"entity,omitempty"`
}

// SearchResponse represents search results.
type SearchResponse struct {
	Results [][]SearchResult `json:"results"`
}

// Search performs vector similarity search.
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Milvus")
	}

	if req.DBName == "" {
		req.DBName = c.config.DBName
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/entities/search", req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	var response struct {
		Data [][]SearchResult `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &SearchResponse{Results: response.Data}, nil
}

// DeleteRequest represents a delete request.
type DeleteRequest struct {
	DBName         string   `json:"dbName,omitempty"`
	CollectionName string   `json:"collectionName"`
	Filter         string   `json:"filter,omitempty"`
	IDs            []string `json:"ids,omitempty"`
}

// Delete removes entities from a collection.
func (c *Client) Delete(ctx context.Context, collectionName string, filter string, ids []string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Milvus")
	}

	req := map[string]interface{}{
		"dbName":         c.config.DBName,
		"collectionName": collectionName,
	}

	if filter != "" {
		req["filter"] = filter
	}
	if len(ids) > 0 {
		req["ids"] = ids
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/entities/delete", req)
	if err != nil {
		return fmt.Errorf("failed to delete entities: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"collection": collectionName,
		"filter":     filter,
		"ids":        len(ids),
	}).Debug("Entities deleted")

	return nil
}

// GetRequest represents a get request.
type GetRequest struct {
	DBName         string   `json:"dbName,omitempty"`
	CollectionName string   `json:"collectionName"`
	IDs            []string `json:"ids"`
	OutputFields   []string `json:"outputFields,omitempty"`
}

// Get retrieves entities by IDs.
func (c *Client) Get(ctx context.Context, collectionName string, ids []string, outputFields []string) ([]map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Milvus")
	}

	if len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}

	req := GetRequest{
		DBName:         c.config.DBName,
		CollectionName: collectionName,
		IDs:            ids,
		OutputFields:   outputFields,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/entities/get", req)
	if err != nil {
		return nil, fmt.Errorf("failed to get entities: %w", err)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

// QueryRequest represents a query request.
type QueryRequest struct {
	DBName         string   `json:"dbName,omitempty"`
	CollectionName string   `json:"collectionName"`
	Filter         string   `json:"filter"`
	OutputFields   []string `json:"outputFields,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Offset         int      `json:"offset,omitempty"`
}

// Query performs a filtered query.
func (c *Client) Query(ctx context.Context, req *QueryRequest) ([]map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Milvus")
	}

	if req.DBName == "" {
		req.DBName = c.config.DBName
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/entities/query", req)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

// CreateIndexRequest represents a create index request.
type CreateIndexRequest struct {
	DBName         string                 `json:"dbName,omitempty"`
	CollectionName string                 `json:"collectionName"`
	FieldName      string                 `json:"fieldName"`
	IndexName      string                 `json:"indexName,omitempty"`
	IndexType      IndexType              `json:"indexType,omitempty"`
	MetricType     MetricType             `json:"metricType,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
}

// CreateIndex creates an index on a field.
func (c *Client) CreateIndex(ctx context.Context, req *CreateIndexRequest) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Milvus")
	}

	if req.DBName == "" {
		req.DBName = c.config.DBName
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/indexes/create", req)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"collection": req.CollectionName,
		"field":      req.FieldName,
		"indexType":  req.IndexType,
	}).Info("Index created")

	return nil
}

// LoadCollection loads a collection into memory.
func (c *Client) LoadCollection(ctx context.Context, collectionName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Milvus")
	}

	req := map[string]interface{}{
		"dbName":         c.config.DBName,
		"collectionName": collectionName,
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/collections/load", req)
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	c.logger.WithField("collection", collectionName).Info("Collection loaded")
	return nil
}

// ReleaseCollection releases a collection from memory.
func (c *Client) ReleaseCollection(ctx context.Context, collectionName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return fmt.Errorf("not connected to Milvus")
	}

	req := map[string]interface{}{
		"dbName":         c.config.DBName,
		"collectionName": collectionName,
	}

	_, err := c.doRequest(ctx, http.MethodPost, "/collections/release", req)
	if err != nil {
		return fmt.Errorf("failed to release collection: %w", err)
	}

	c.logger.WithField("collection", collectionName).Info("Collection released")
	return nil
}

// GetLoadState gets the load state of a collection.
func (c *Client) GetLoadState(ctx context.Context, collectionName string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return "", fmt.Errorf("not connected to Milvus")
	}

	req := map[string]interface{}{
		"dbName":         c.config.DBName,
		"collectionName": collectionName,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "/collections/get_load_state", req)
	if err != nil {
		return "", fmt.Errorf("failed to get load state: %w", err)
	}

	var response struct {
		Data struct {
			LoadState string `json:"loadState"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data.LoadState, nil
}
