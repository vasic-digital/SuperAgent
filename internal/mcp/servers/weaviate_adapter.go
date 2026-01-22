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

// WeaviateAdapter provides MCP-compatible interface to Weaviate.
type WeaviateAdapter struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	mu         sync.RWMutex
	connected  bool
}

// WeaviateAdapterConfig holds configuration for WeaviateAdapter.
type WeaviateAdapterConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// WeaviateClass represents a Weaviate class (collection).
type WeaviateClass struct {
	Class        string                 `json:"class"`
	Description  string                 `json:"description,omitempty"`
	Properties   []WeaviateProperty     `json:"properties,omitempty"`
	Vectorizer   string                 `json:"vectorizer,omitempty"`
	ModuleConfig map[string]interface{} `json:"moduleConfig,omitempty"`
}

// WeaviateProperty represents a property in a Weaviate class.
type WeaviateProperty struct {
	Name        string   `json:"name"`
	DataType    []string `json:"dataType"`
	Description string   `json:"description,omitempty"`
}

// WeaviateObject represents an object in Weaviate.
type WeaviateObject struct {
	ID         string                 `json:"id,omitempty"`
	Class      string                 `json:"class"`
	Properties map[string]interface{} `json:"properties"`
	Vector     []float32              `json:"vector,omitempty"`
}

// WeaviateSearchResult represents a search result from Weaviate.
type WeaviateSearchResult struct {
	ID         string                 `json:"id"`
	Class      string                 `json:"class"`
	Properties map[string]interface{} `json:"properties"`
	Vector     []float32              `json:"vector,omitempty"`
	Distance   float32                `json:"_additional,omitempty"`
	Certainty  float32                `json:"certainty,omitempty"`
}

// WeaviateSchema represents the Weaviate schema.
type WeaviateSchema struct {
	Classes []WeaviateClass `json:"classes"`
}

// NewWeaviateAdapter creates a new Weaviate adapter.
func NewWeaviateAdapter(config WeaviateAdapterConfig) *WeaviateAdapter {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &WeaviateAdapter{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Connect establishes connection to Weaviate.
func (a *WeaviateAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Test connection with ready endpoint
	resp, err := a.doRequest(ctx, "GET", "/v1/.well-known/ready", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Weaviate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Weaviate health check failed: status %d", resp.StatusCode)
	}

	a.connected = true
	return nil
}

// IsConnected returns whether the adapter is connected.
func (a *WeaviateAdapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connected
}

// Health checks the health of Weaviate connection.
func (a *WeaviateAdapter) Health(ctx context.Context) error {
	resp, err := a.doRequest(ctx, "GET", "/v1/.well-known/ready", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

// GetSchema retrieves the full Weaviate schema.
func (a *WeaviateAdapter) GetSchema(ctx context.Context) (*WeaviateSchema, error) {
	resp, err := a.doRequest(ctx, "GET", "/v1/schema", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get schema: status %d", resp.StatusCode)
	}

	var schema WeaviateSchema
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", err)
	}

	return &schema, nil
}

// ListClasses lists all classes in Weaviate.
func (a *WeaviateAdapter) ListClasses(ctx context.Context) ([]WeaviateClass, error) {
	schema, err := a.GetSchema(ctx)
	if err != nil {
		return nil, err
	}
	return schema.Classes, nil
}

// CreateClass creates a new class in Weaviate.
func (a *WeaviateAdapter) CreateClass(ctx context.Context, class *WeaviateClass) error {
	resp, err := a.doRequest(ctx, "POST", "/v1/schema", class)
	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create class: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteClass deletes a class from Weaviate.
func (a *WeaviateAdapter) DeleteClass(ctx context.Context, className string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/v1/schema/%s", className), nil)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete class: status %d", resp.StatusCode)
	}

	return nil
}

// GetClass retrieves a class by name.
func (a *WeaviateAdapter) GetClass(ctx context.Context, className string) (*WeaviateClass, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/v1/schema/%s", className), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("class not found: %s", className)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get class: status %d", resp.StatusCode)
	}

	var class WeaviateClass
	if err := json.NewDecoder(resp.Body).Decode(&class); err != nil {
		return nil, fmt.Errorf("failed to decode class: %w", err)
	}

	return &class, nil
}

// CreateObject creates a new object in Weaviate.
func (a *WeaviateAdapter) CreateObject(ctx context.Context, obj *WeaviateObject) (*WeaviateObject, error) {
	resp, err := a.doRequest(ctx, "POST", "/v1/objects", obj)
	if err != nil {
		return nil, fmt.Errorf("failed to create object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create object: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var created WeaviateObject
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("failed to decode created object: %w", err)
	}

	return &created, nil
}

// BatchCreateObjects creates multiple objects in Weaviate.
func (a *WeaviateAdapter) BatchCreateObjects(ctx context.Context, objects []WeaviateObject) error {
	body := map[string]interface{}{
		"objects": objects,
	}

	resp, err := a.doRequest(ctx, "POST", "/v1/batch/objects", body)
	if err != nil {
		return fmt.Errorf("failed to batch create objects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to batch create objects: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetObject retrieves an object by ID.
func (a *WeaviateAdapter) GetObject(ctx context.Context, className, id string) (*WeaviateObject, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/v1/objects/%s/%s", className, id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("object not found: %s/%s", className, id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get object: status %d", resp.StatusCode)
	}

	var obj WeaviateObject
	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		return nil, fmt.Errorf("failed to decode object: %w", err)
	}

	return &obj, nil
}

// DeleteObject deletes an object by ID.
func (a *WeaviateAdapter) DeleteObject(ctx context.Context, className, id string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/v1/objects/%s/%s", className, id), nil)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete object: status %d", resp.StatusCode)
	}

	return nil
}

// UpdateObject updates an object.
func (a *WeaviateAdapter) UpdateObject(ctx context.Context, obj *WeaviateObject) error {
	resp, err := a.doRequest(ctx, "PUT", fmt.Sprintf("/v1/objects/%s/%s", obj.Class, obj.ID), obj)
	if err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update object: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// VectorSearch performs a vector similarity search using GraphQL.
func (a *WeaviateAdapter) VectorSearch(ctx context.Context, className string, vector []float32, limit int, certainty float32, properties []string) ([]WeaviateSearchResult, error) {
	// Build GraphQL query
	propsStr := "id"
	if len(properties) > 0 {
		for _, p := range properties {
			propsStr += " " + p
		}
	}

	query := fmt.Sprintf(`{
		Get {
			%s(
				nearVector: {
					vector: %v
					certainty: %f
				}
				limit: %d
			) {
				%s
				_additional {
					id
					certainty
					distance
				}
			}
		}
	}`, className, vector, certainty, limit, propsStr)

	body := map[string]interface{}{
		"query": query,
	}

	resp, err := a.doRequest(ctx, "POST", "/v1/graphql", body)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	// Parse results from GraphQL response
	var results []WeaviateSearchResult
	if data, ok := response["data"].(map[string]interface{}); ok {
		if get, ok := data["Get"].(map[string]interface{}); ok {
			if items, ok := get[className].([]interface{}); ok {
				for _, item := range items {
					if m, ok := item.(map[string]interface{}); ok {
						result := WeaviateSearchResult{
							Class:      className,
							Properties: make(map[string]interface{}),
						}
						for k, v := range m {
							if k == "_additional" {
								if add, ok := v.(map[string]interface{}); ok {
									if id, ok := add["id"].(string); ok {
										result.ID = id
									}
									if cert, ok := add["certainty"].(float64); ok {
										result.Certainty = float32(cert)
									}
									if dist, ok := add["distance"].(float64); ok {
										result.Distance = float32(dist)
									}
								}
							} else {
								result.Properties[k] = v
							}
						}
						results = append(results, result)
					}
				}
			}
		}
	}

	return results, nil
}

// HybridSearch performs a hybrid (vector + keyword) search.
func (a *WeaviateAdapter) HybridSearch(ctx context.Context, className, query string, vector []float32, limit int, alpha float32, properties []string) ([]WeaviateSearchResult, error) {
	// Build GraphQL query for hybrid search
	propsStr := "id"
	if len(properties) > 0 {
		for _, p := range properties {
			propsStr += " " + p
		}
	}

	graphqlQuery := fmt.Sprintf(`{
		Get {
			%s(
				hybrid: {
					query: "%s"
					vector: %v
					alpha: %f
				}
				limit: %d
			) {
				%s
				_additional {
					id
					score
				}
			}
		}
	}`, className, query, vector, alpha, limit, propsStr)

	body := map[string]interface{}{
		"query": graphqlQuery,
	}

	resp, err := a.doRequest(ctx, "POST", "/v1/graphql", body)
	if err != nil {
		return nil, fmt.Errorf("failed to hybrid search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to hybrid search: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	// Parse results (similar to VectorSearch)
	var results []WeaviateSearchResult
	if data, ok := response["data"].(map[string]interface{}); ok {
		if get, ok := data["Get"].(map[string]interface{}); ok {
			if items, ok := get[className].([]interface{}); ok {
				for _, item := range items {
					if m, ok := item.(map[string]interface{}); ok {
						result := WeaviateSearchResult{
							Class:      className,
							Properties: make(map[string]interface{}),
						}
						for k, v := range m {
							if k == "_additional" {
								if add, ok := v.(map[string]interface{}); ok {
									if id, ok := add["id"].(string); ok {
										result.ID = id
									}
								}
							} else {
								result.Properties[k] = v
							}
						}
						results = append(results, result)
					}
				}
			}
		}
	}

	return results, nil
}

// Close closes the adapter connection.
func (a *WeaviateAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	return nil
}

// doRequest performs an HTTP request to Weaviate.
func (a *WeaviateAdapter) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
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
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	return a.httpClient.Do(req)
}

// GetMCPTools returns the MCP tool definitions for Weaviate.
func (a *WeaviateAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "weaviate_list_classes",
			Description: "List all classes (collections) in Weaviate",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "weaviate_create_class",
			Description: "Create a new class (collection) in Weaviate",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"class": map[string]interface{}{
						"type":        "string",
						"description": "Name of the class to create",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Description of the class",
					},
					"properties": map[string]interface{}{
						"type":        "array",
						"description": "Array of property definitions",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name":     map[string]interface{}{"type": "string"},
								"dataType": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
							},
						},
					},
					"vectorizer": map[string]interface{}{
						"type":        "string",
						"description": "Vectorizer module to use",
					},
				},
				"required": []string{"class"},
			},
		},
		{
			Name:        "weaviate_delete_class",
			Description: "Delete a class from Weaviate",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"class": map[string]interface{}{
						"type":        "string",
						"description": "Name of the class to delete",
					},
				},
				"required": []string{"class"},
			},
		},
		{
			Name:        "weaviate_create_object",
			Description: "Create a new object in a Weaviate class",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"class": map[string]interface{}{
						"type":        "string",
						"description": "Name of the class",
					},
					"properties": map[string]interface{}{
						"type":        "object",
						"description": "Object properties",
					},
					"vector": map[string]interface{}{
						"type":        "array",
						"description": "Optional vector embedding",
						"items":       map[string]interface{}{"type": "number"},
					},
				},
				"required": []string{"class", "properties"},
			},
		},
		{
			Name:        "weaviate_vector_search",
			Description: "Search for similar objects using vector similarity",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"class": map[string]interface{}{
						"type":        "string",
						"description": "Name of the class to search",
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
					"certainty": map[string]interface{}{
						"type":        "number",
						"description": "Minimum certainty threshold (0-1)",
						"default":     0.7,
					},
					"properties": map[string]interface{}{
						"type":        "array",
						"description": "Properties to return",
						"items":       map[string]interface{}{"type": "string"},
					},
				},
				"required": []string{"class", "vector"},
			},
		},
		{
			Name:        "weaviate_hybrid_search",
			Description: "Perform hybrid search (vector + keyword) in Weaviate",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"class": map[string]interface{}{
						"type":        "string",
						"description": "Name of the class to search",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Text query for keyword search",
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
					"alpha": map[string]interface{}{
						"type":        "number",
						"description": "Balance between vector (1) and keyword (0) search",
						"default":     0.5,
					},
				},
				"required": []string{"class", "query", "vector"},
			},
		},
		{
			Name:        "weaviate_delete_object",
			Description: "Delete an object from Weaviate",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"class": map[string]interface{}{
						"type":        "string",
						"description": "Name of the class",
					},
					"id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the object to delete",
					},
				},
				"required": []string{"class", "id"},
			},
		},
	}
}
