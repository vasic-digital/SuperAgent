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

// ChromaAdapter provides MCP-compatible interface to ChromaDB.
type ChromaAdapter struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
	mu         sync.RWMutex
	connected  bool
}

// ChromaAdapterConfig holds configuration for ChromaAdapter.
type ChromaAdapterConfig struct {
	BaseURL   string
	AuthToken string
	Timeout   time.Duration
}

// ChromaCollection represents a ChromaDB collection.
type ChromaCollection struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChromaDocument represents a document in ChromaDB.
type ChromaDocument struct {
	ID        string                 `json:"id"`
	Document  string                 `json:"document,omitempty"`
	Embedding []float32              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChromaQueryResult represents a query result from ChromaDB.
type ChromaQueryResult struct {
	IDs        [][]string                 `json:"ids"`
	Documents  [][]string                 `json:"documents,omitempty"`
	Embeddings [][][]float32              `json:"embeddings,omitempty"`
	Metadatas  [][]map[string]interface{} `json:"metadatas,omitempty"`
	Distances  [][]float32                `json:"distances,omitempty"`
}

// NewChromaAdapter creates a new ChromaDB adapter.
func NewChromaAdapter(config ChromaAdapterConfig) *ChromaAdapter {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &ChromaAdapter{
		baseURL:   config.BaseURL,
		authToken: config.AuthToken,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Connect establishes connection to ChromaDB.
func (a *ChromaAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Test connection with heartbeat
	resp, err := a.doRequest(ctx, "GET", "/api/v1/heartbeat", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to ChromaDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ChromaDB heartbeat failed: status %d", resp.StatusCode)
	}

	a.connected = true
	return nil
}

// IsConnected returns whether the adapter is connected.
func (a *ChromaAdapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connected
}

// Health checks the health of ChromaDB connection.
func (a *ChromaAdapter) Health(ctx context.Context) error {
	resp, err := a.doRequest(ctx, "GET", "/api/v1/heartbeat", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

// ListCollections lists all collections in ChromaDB.
func (a *ChromaAdapter) ListCollections(ctx context.Context) ([]ChromaCollection, error) {
	resp, err := a.doRequest(ctx, "GET", "/api/v1/collections", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer resp.Body.Close()

	var collections []ChromaCollection
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return nil, fmt.Errorf("failed to decode collections: %w", err)
	}

	return collections, nil
}

// CreateCollection creates a new collection in ChromaDB.
func (a *ChromaAdapter) CreateCollection(ctx context.Context, name string, metadata map[string]interface{}) (*ChromaCollection, error) {
	body := map[string]interface{}{
		"name": name,
	}
	if metadata != nil {
		body["metadata"] = metadata
	}

	resp, err := a.doRequest(ctx, "POST", "/api/v1/collections", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create collection: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var collection ChromaCollection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("failed to decode collection: %w", err)
	}

	return &collection, nil
}

// DeleteCollection deletes a collection from ChromaDB.
func (a *ChromaAdapter) DeleteCollection(ctx context.Context, name string) error {
	resp, err := a.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/collections/%s", name), nil)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete collection: status %d", resp.StatusCode)
	}

	return nil
}

// GetCollection retrieves a collection by name.
func (a *ChromaAdapter) GetCollection(ctx context.Context, name string) (*ChromaCollection, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/collections/%s", name), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("collection not found: %s", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get collection: status %d", resp.StatusCode)
	}

	var collection ChromaCollection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("failed to decode collection: %w", err)
	}

	return &collection, nil
}

// AddDocuments adds documents to a collection.
func (a *ChromaAdapter) AddDocuments(ctx context.Context, collectionName string, docs []ChromaDocument) error {
	ids := make([]string, len(docs))
	documents := make([]string, len(docs))
	embeddings := make([][]float32, 0)
	metadatas := make([]map[string]interface{}, len(docs))

	for i, doc := range docs {
		ids[i] = doc.ID
		documents[i] = doc.Document
		metadatas[i] = doc.Metadata
		if doc.Embedding != nil {
			embeddings = append(embeddings, doc.Embedding)
		}
	}

	body := map[string]interface{}{
		"ids":       ids,
		"documents": documents,
		"metadatas": metadatas,
	}
	if len(embeddings) == len(docs) {
		body["embeddings"] = embeddings
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/collections/%s/add", collectionName), body)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add documents: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Query performs a similarity search in a collection.
func (a *ChromaAdapter) Query(ctx context.Context, collectionName string, queryEmbeddings [][]float32, nResults int, where map[string]interface{}) (*ChromaQueryResult, error) {
	body := map[string]interface{}{
		"query_embeddings": queryEmbeddings,
		"n_results":        nResults,
		"include":          []string{"documents", "metadatas", "distances"},
	}
	if where != nil {
		body["where"] = where
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/collections/%s/query", collectionName), body)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to query collection: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ChromaQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode query result: %w", err)
	}

	return &result, nil
}

// DeleteDocuments deletes documents from a collection by IDs.
func (a *ChromaAdapter) DeleteDocuments(ctx context.Context, collectionName string, ids []string) error {
	body := map[string]interface{}{
		"ids": ids,
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/collections/%s/delete", collectionName), body)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete documents: status %d", resp.StatusCode)
	}

	return nil
}

// UpdateDocuments updates documents in a collection.
func (a *ChromaAdapter) UpdateDocuments(ctx context.Context, collectionName string, docs []ChromaDocument) error {
	ids := make([]string, len(docs))
	documents := make([]string, len(docs))
	embeddings := make([][]float32, 0)
	metadatas := make([]map[string]interface{}, len(docs))

	for i, doc := range docs {
		ids[i] = doc.ID
		documents[i] = doc.Document
		metadatas[i] = doc.Metadata
		if doc.Embedding != nil {
			embeddings = append(embeddings, doc.Embedding)
		}
	}

	body := map[string]interface{}{
		"ids":       ids,
		"documents": documents,
		"metadatas": metadatas,
	}
	if len(embeddings) == len(docs) {
		body["embeddings"] = embeddings
	}

	resp, err := a.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/collections/%s/update", collectionName), body)
	if err != nil {
		return fmt.Errorf("failed to update documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update documents: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Count returns the number of documents in a collection.
func (a *ChromaAdapter) Count(ctx context.Context, collectionName string) (int64, error) {
	resp, err := a.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/collections/%s/count", collectionName), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to count documents: status %d", resp.StatusCode)
	}

	var count int64
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return 0, fmt.Errorf("failed to decode count: %w", err)
	}

	return count, nil
}

// Close closes the adapter connection.
func (a *ChromaAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	return nil
}

// doRequest performs an HTTP request to ChromaDB.
func (a *ChromaAdapter) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
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
	if a.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.authToken)
	}

	return a.httpClient.Do(req)
}

// GetMCPTools returns the MCP tool definitions for ChromaDB.
func (a *ChromaAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "chroma_list_collections",
			Description: "List all collections in ChromaDB",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "chroma_create_collection",
			Description: "Create a new collection in ChromaDB",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection to create",
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Optional metadata for the collection",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "chroma_delete_collection",
			Description: "Delete a collection from ChromaDB",
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
			Name:        "chroma_add_documents",
			Description: "Add documents to a ChromaDB collection",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection",
					},
					"documents": map[string]interface{}{
						"type":        "array",
						"description": "Array of documents to add",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":        map[string]interface{}{"type": "string"},
								"document":  map[string]interface{}{"type": "string"},
								"embedding": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
								"metadata":  map[string]interface{}{"type": "object"},
							},
							"required": []string{"id", "document"},
						},
					},
				},
				"required": []string{"collection", "documents"},
			},
		},
		{
			Name:        "chroma_query",
			Description: "Query documents from a ChromaDB collection using similarity search",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection to query",
					},
					"query_embeddings": map[string]interface{}{
						"type":        "array",
						"description": "Query embedding vectors",
						"items": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"type": "number"},
						},
					},
					"n_results": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to return",
						"default":     10,
					},
					"where": map[string]interface{}{
						"type":        "object",
						"description": "Optional filter conditions",
					},
				},
				"required": []string{"collection", "query_embeddings"},
			},
		},
		{
			Name:        "chroma_delete_documents",
			Description: "Delete documents from a ChromaDB collection by IDs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Name of the collection",
					},
					"ids": map[string]interface{}{
						"type":        "array",
						"description": "Array of document IDs to delete",
						"items":       map[string]interface{}{"type": "string"},
					},
				},
				"required": []string{"collection", "ids"},
			},
		},
		{
			Name:        "chroma_count",
			Description: "Get the number of documents in a ChromaDB collection",
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

// MCPTool represents an MCP tool definition.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}
