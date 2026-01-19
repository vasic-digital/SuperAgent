// Package servers provides MCP server adapters for various services.
package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// QdrantAdapter provides MCP-compatible interface to Qdrant.
type QdrantAdapter struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	mu         sync.RWMutex
	connected  bool
}

// QdrantAdapterConfig holds configuration for QdrantAdapter.
type QdrantAdapterConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// QdrantCollection represents a Qdrant collection.
type QdrantCollection struct {
	Name   string                  `json:"name"`
	Status string                  `json:"status"`
	Config *QdrantCollectionConfig `json:"config,omitempty"`
}

// QdrantCollectionConfig represents collection configuration.
type QdrantCollectionConfig struct {
	VectorSize uint64 `json:"vector_size"`
	Distance   string `json:"distance"` // Cosine, Euclidean, Dot
}

// QdrantPoint represents a point (vector) in Qdrant.
type QdrantPoint struct {
	ID      interface{}            `json:"id"` // Can be uint64 or string
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// QdrantSearchResult represents a search result from Qdrant.
type QdrantSearchResult struct {
	ID      interface{}            `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload,omitempty"`
	Vector  []float32              `json:"vector,omitempty"`
}

// QdrantSearchResponse represents the search response.
type QdrantSearchResponse struct {
	Result []QdrantSearchResult `json:"result"`
	Status string               `json:"status"`
	Time   float64              `json:"time"`
}

// QdrantCollectionsResponse represents the collections list response.
type QdrantCollectionsResponse struct {
	Result struct {
		Collections []QdrantCollection `json:"collections"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// NewQdrantAdapter creates a new Qdrant adapter.
func NewQdrantAdapter(config QdrantAdapterConfig) *QdrantAdapter {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &QdrantAdapter{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Connect establishes connection to Qdrant.
func (a *QdrantAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Test connection with readyz endpoint
	resp, err := a.doRequest(ctx, "GET", "/readyz", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Qdrant: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Qdrant health check failed: status %d", resp.StatusCode)
	}

	a.connected = true
	return nil
}

// IsConnected returns whether the adapter is connected.
func (a *QdrantAdapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connected
}

// Health checks the health of Qdrant connection.
func (a *QdrantAdapter) Health(ctx context.Context) error {
	resp, err := a.doRequest(ctx, "GET", "/readyz", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

// ListCollections lists all collections in Qdrant.
func (a *QdrantAdapter) ListCollections(ctx context.Context) ([]QdrantCollection, error) {
	resp, err := a.doRequest(ctx, "GET", "/collections", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer resp.Body.Close()

	var response QdrantCollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode collections: %w", err)
	}

	return response.Result.Collections, nil
}

// CreateCollection creates a new collection in Qdrant.
func (a *QdrantAdapter) CreateCollection(ctx context.Context, name string, vectorSize uint64, distance string) error {
	if distance == "" {
		distance = "Cosine"
	}

	body := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     vectorSize,
			"distance": distance,
		},
	}

	resp, err := a.doRequest(ctx, "PUT", fmt.Sprintf("/collections/%s", name), body)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteCollection deletes a collection from Qdrant.
func (a *QdrantAdapter) DeleteCollection(ctx context.Context, name string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/collections/%s", name), nil)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete collection: status %d", resp.StatusCode)
	}

	return nil
}

// CollectionExists checks if a collection exists.
func (a *QdrantAdapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/collections/%s", name), nil)
	if err != nil {
		return false, fmt.Errorf("failed to check collection: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetCollectionInfo retrieves collection information.
func (a *QdrantAdapter) GetCollectionInfo(ctx context.Context, name string) (*QdrantCollection, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/collections/%s", name), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("collection not found: %s", name)
	}

	var response struct {
		Result QdrantCollection `json:"result"`
		Status string           `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode collection info: %w", err)
	}

	return &response.Result, nil
}

// UpsertPoints inserts or updates points in a collection.
func (a *QdrantAdapter) UpsertPoints(ctx context.Context, collectionName string, points []QdrantPoint) error {
	body := map[string]interface{}{
		"points": points,
	}

	resp, err := a.doRequest(ctx, "PUT", fmt.Sprintf("/collections/%s/points", collectionName), body)
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upsert points: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeletePoints deletes points from a collection by IDs.
func (a *QdrantAdapter) DeletePoints(ctx context.Context, collectionName string, ids []interface{}) error {
	body := map[string]interface{}{
		"points": ids,
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/delete", collectionName), body)
	if err != nil {
		return fmt.Errorf("failed to delete points: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete points: status %d", resp.StatusCode)
	}

	return nil
}

// Search performs similarity search in a collection.
func (a *QdrantAdapter) Search(ctx context.Context, collectionName string, vector []float32, limit int, filter map[string]interface{}, withPayload bool, withVector bool) ([]QdrantSearchResult, error) {
	body := map[string]interface{}{
		"vector":       vector,
		"limit":        limit,
		"with_payload": withPayload,
		"with_vector":  withVector,
	}
	if filter != nil {
		body["filter"] = filter
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/search", collectionName), body)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response QdrantSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return response.Result, nil
}

// SearchBatch performs multiple similarity searches.
func (a *QdrantAdapter) SearchBatch(ctx context.Context, collectionName string, searches []struct {
	Vector      []float32
	Limit       int
	Filter      map[string]interface{}
	WithPayload bool
}) ([][]QdrantSearchResult, error) {
	searchRequests := make([]map[string]interface{}, len(searches))
	for i, s := range searches {
		searchRequests[i] = map[string]interface{}{
			"vector":       s.Vector,
			"limit":        s.Limit,
			"with_payload": s.WithPayload,
		}
		if s.Filter != nil {
			searchRequests[i]["filter"] = s.Filter
		}
	}

	body := map[string]interface{}{
		"searches": searchRequests,
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/search/batch", collectionName), body)
	if err != nil {
		return nil, fmt.Errorf("failed to search batch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search batch: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Result [][]QdrantSearchResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode batch search results: %w", err)
	}

	return response.Result, nil
}

// GetPoints retrieves points by IDs.
func (a *QdrantAdapter) GetPoints(ctx context.Context, collectionName string, ids []interface{}, withPayload bool, withVector bool) ([]QdrantPoint, error) {
	body := map[string]interface{}{
		"ids":          ids,
		"with_payload": withPayload,
		"with_vector":  withVector,
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points", collectionName), body)
	if err != nil {
		return nil, fmt.Errorf("failed to get points: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get points: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Result []QdrantPoint `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode points: %w", err)
	}

	return response.Result, nil
}

// CountPoints returns the number of points in a collection.
func (a *QdrantAdapter) CountPoints(ctx context.Context, collectionName string) (uint64, error) {
	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/count", collectionName), map[string]interface{}{})
	if err != nil {
		return 0, fmt.Errorf("failed to count points: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to count points: status %d", resp.StatusCode)
	}

	var response struct {
		Result struct {
			Count uint64 `json:"count"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode count: %w", err)
	}

	return response.Result.Count, nil
}

// Scroll iterates through all points in a collection.
func (a *QdrantAdapter) Scroll(ctx context.Context, collectionName string, offset interface{}, limit int, withPayload bool, withVector bool, filter map[string]interface{}) ([]QdrantPoint, interface{}, error) {
	body := map[string]interface{}{
		"limit":        limit,
		"with_payload": withPayload,
		"with_vector":  withVector,
	}
	if offset != nil {
		body["offset"] = offset
	}
	if filter != nil {
		body["filter"] = filter
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/collections/%s/points/scroll", collectionName), body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scroll: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("failed to scroll: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Result struct {
			Points     []QdrantPoint `json:"points"`
			NextPageOffset interface{} `json:"next_page_offset"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, nil, fmt.Errorf("failed to decode scroll results: %w", err)
	}

	return response.Result.Points, response.Result.NextPageOffset, nil
}

// Close closes the adapter connection.
func (a *QdrantAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	return nil
}

// doRequest performs an HTTP request to Qdrant.
func (a *QdrantAdapter) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("api-key", a.apiKey)
	}

	return a.httpClient.Do(req)
}

// GetMCPTools returns the MCP tool definitions for Qdrant.
func (a *QdrantAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "qdrant_list_collections",
			Description: "List all collections in Qdrant",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "qdrant_create_collection",
			Description: "Create a new collection in Qdrant",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection to create",
					},
					"vector_size": map[string]interface{}{
						"type":        "integer",
						"description": "Size of vectors (e.g., 1536 for OpenAI)",
					},
					"distance": map[string]interface{}{
						"type":        "string",
						"description": "Distance metric (Cosine, Euclidean, Dot)",
						"enum":        []string{"Cosine", "Euclidean", "Dot"},
						"default":     "Cosine",
					},
				},
				"required": []string{"name", "vector_size"},
			},
		},
		{
			Name:        "qdrant_delete_collection",
			Description: "Delete a collection from Qdrant",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection to delete",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "qdrant_upsert_points",
			Description: "Insert or update points in a Qdrant collection",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection",
					},
					"points": map[string]interface{}{
						"type":        "array",
						"description": "Array of points to upsert",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":      map[string]interface{}{"type": []string{"string", "integer"}},
								"vector":  map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
								"payload": map[string]interface{}{"type": "object"},
							},
							"required": []string{"id", "vector"},
						},
					},
				},
				"required": []string{"collection", "points"},
			},
		},
		{
			Name:        "qdrant_search",
			Description: "Search for similar vectors in a Qdrant collection",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection to search",
					},
					"vector": map[string]interface{}{
						"type":        "array",
						"description": "Query vector",
						"items":       map[string]interface{}{"type": "number"},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to return",
						"default":     10,
					},
					"filter": map[string]interface{}{
						"type":        "object",
						"description": "Optional filter conditions",
					},
					"with_payload": map[string]interface{}{
						"type":        "boolean",
						"description": "Include payload in results",
						"default":     true,
					},
				},
				"required": []string{"collection", "vector"},
			},
		},
		{
			Name:        "qdrant_delete_points",
			Description: "Delete points from a Qdrant collection by IDs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection",
					},
					"ids": map[string]interface{}{
						"type":        "array",
						"description": "Array of point IDs to delete",
						"items":       map[string]interface{}{"type": []string{"string", "integer"}},
					},
				},
				"required": []string{"collection", "ids"},
			},
		},
		{
			Name:        "qdrant_count_points",
			Description: "Get the number of points in a Qdrant collection",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection",
					},
				},
				"required": []string{"collection"},
			},
		},
	}
}
