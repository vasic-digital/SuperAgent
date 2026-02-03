package pinecone

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Config Comprehensive Tests
// =============================================================================

func TestConfig_Validate_Comprehensive(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid config",
			config: &Config{
				APIKey:    "test-api-key",
				IndexHost: "https://my-index.pinecone.io",
				Timeout:   30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "empty API key",
			config: &Config{
				APIKey:    "",
				IndexHost: "https://my-index.pinecone.io",
			},
			expectError: true,
			errorSubstr: "API key",
		},
		{
			name: "whitespace API key",
			config: &Config{
				APIKey:    "   ",
				IndexHost: "https://my-index.pinecone.io",
			},
			expectError: false, // Whitespace is technically a non-empty string
		},
		{
			name: "empty index host",
			config: &Config{
				APIKey:    "test-key",
				IndexHost: "",
			},
			expectError: true,
			errorSubstr: "index host",
		},
		{
			name: "index host with port",
			config: &Config{
				APIKey:    "test-key",
				IndexHost: "https://my-index.pinecone.io:443",
			},
			expectError: false,
		},
		{
			name: "index host without protocol",
			config: &Config{
				APIKey:    "test-key",
				IndexHost: "my-index.pinecone.io",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				if tt.errorSubstr != "" {
					assert.Contains(t, err.Error(), tt.errorSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	config := DefaultConfig()

	assert.Empty(t, config.APIKey)
	assert.Empty(t, config.Environment)
	assert.Empty(t, config.ProjectID)
	assert.Empty(t, config.IndexHost)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

// =============================================================================
// Client Creation Tests
// =============================================================================

func TestNewClient_Comprehensive(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		logger      *logrus.Logger
		expectError bool
		errorSubstr string
	}{
		{
			name:        "nil config",
			config:      nil,
			logger:      nil,
			expectError: true, // Default config has empty API key and index host
		},
		{
			name: "valid config with nil logger",
			config: &Config{
				APIKey:    "test-key",
				IndexHost: "https://test.pinecone.io",
				Timeout:   time.Second,
			},
			logger:      nil,
			expectError: false,
		},
		{
			name: "valid config with custom logger",
			config: &Config{
				APIKey:    "test-key",
				IndexHost: "https://test.pinecone.io",
				Timeout:   time.Second,
			},
			logger:      logrus.New(),
			expectError: false,
		},
		{
			name: "invalid config",
			config: &Config{
				APIKey:    "",
				IndexHost: "",
			},
			logger:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config, tt.logger)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.False(t, client.IsConnected())
			}
		})
	}
}

// =============================================================================
// Vector Type Tests
// =============================================================================

func TestVector_Fields(t *testing.T) {
	v := Vector{
		ID:     "vec-123",
		Values: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		Metadata: map[string]interface{}{
			"category": "tech",
			"count":    42,
			"active":   true,
		},
	}

	assert.Equal(t, "vec-123", v.ID)
	assert.Len(t, v.Values, 5)
	assert.Equal(t, "tech", v.Metadata["category"])
	assert.Equal(t, 42, v.Metadata["count"])
	assert.Equal(t, true, v.Metadata["active"])
}

func TestScoredVector_Fields(t *testing.T) {
	sv := ScoredVector{
		ID:     "result-1",
		Score:  0.95,
		Values: []float32{0.1, 0.2},
		Metadata: map[string]interface{}{
			"text": "sample",
		},
	}

	assert.Equal(t, "result-1", sv.ID)
	assert.Equal(t, float32(0.95), sv.Score)
	assert.Len(t, sv.Values, 2)
	assert.Equal(t, "sample", sv.Metadata["text"])
}

// =============================================================================
// Request/Response Type Tests
// =============================================================================

func TestUpsertRequest_Fields(t *testing.T) {
	req := UpsertRequest{
		Vectors: []Vector{
			{ID: "v1", Values: []float32{0.1}},
			{ID: "v2", Values: []float32{0.2}},
		},
		Namespace: "test-ns",
	}

	assert.Len(t, req.Vectors, 2)
	assert.Equal(t, "test-ns", req.Namespace)
}

func TestQueryRequest_Fields(t *testing.T) {
	req := QueryRequest{
		Vector:          []float32{0.1, 0.2, 0.3},
		ID:              "query-id",
		TopK:            10,
		Namespace:       "test-ns",
		Filter:          map[string]interface{}{"key": "value"},
		IncludeValues:   true,
		IncludeMetadata: true,
	}

	assert.Len(t, req.Vector, 3)
	assert.Equal(t, "query-id", req.ID)
	assert.Equal(t, 10, req.TopK)
	assert.Equal(t, "test-ns", req.Namespace)
	assert.True(t, req.IncludeValues)
	assert.True(t, req.IncludeMetadata)
}

func TestDeleteRequest_Fields(t *testing.T) {
	req := DeleteRequest{
		IDs:       []string{"id1", "id2"},
		DeleteAll: false,
		Namespace: "test-ns",
		Filter:    map[string]interface{}{"category": "old"},
	}

	assert.Len(t, req.IDs, 2)
	assert.False(t, req.DeleteAll)
	assert.Equal(t, "test-ns", req.Namespace)
}

func TestUpdateRequest_Fields(t *testing.T) {
	req := UpdateRequest{
		ID:          "vec-123",
		Values:      []float32{0.5, 0.6, 0.7},
		SetMetadata: map[string]interface{}{"updated": true},
		Namespace:   "test-ns",
	}

	assert.Equal(t, "vec-123", req.ID)
	assert.Len(t, req.Values, 3)
	assert.Equal(t, true, req.SetMetadata["updated"])
	assert.Equal(t, "test-ns", req.Namespace)
}

func TestDescribeIndexStatsResponse_Fields(t *testing.T) {
	resp := DescribeIndexStatsResponse{
		Namespaces: map[string]NamespaceStats{
			"ns1": {VectorCount: 1000},
			"ns2": {VectorCount: 2000},
		},
		Dimension:        1536,
		IndexFullness:    0.5,
		TotalVectorCount: 3000,
	}

	assert.Len(t, resp.Namespaces, 2)
	assert.Equal(t, 1536, resp.Dimension)
	assert.Equal(t, 0.5, resp.IndexFullness)
	assert.Equal(t, int64(3000), resp.TotalVectorCount)
}

// =============================================================================
// Connection Tests
// =============================================================================

func TestClient_Connect_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Slow response
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-key",
		IndexHost: server.URL,
		Timeout:   100 * time.Millisecond,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

func TestClient_Connect_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-key",
		IndexHost: server.URL,
		Timeout:   30 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = client.Connect(ctx)
	assert.Error(t, err)
}

// =============================================================================
// Upsert Tests
// =============================================================================

func TestClient_Upsert_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "Invalid vector dimensions"}`))
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	vectors := []Vector{
		{ID: "v1", Values: []float32{0.1, 0.2}},
	}
	_, err := client.Upsert(context.Background(), vectors, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert")
}

func TestClient_Upsert_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	vectors := []Vector{
		{ID: "v1", Values: []float32{0.1, 0.2}},
	}
	_, err := client.Upsert(context.Background(), vectors, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

// =============================================================================
// Query Tests
// =============================================================================

func TestClient_Query_WithID(t *testing.T) {
	var receivedReq QueryRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedReq)
		_ = json.NewEncoder(w).Encode(QueryResponse{
			Matches: []ScoredVector{{ID: "match1", Score: 0.9}},
		})
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	req := &QueryRequest{
		ID:   "query-by-id",
		TopK: 5,
	}
	result, err := client.Query(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "query-by-id", receivedReq.ID)
	assert.Len(t, result.Matches, 1)
}

func TestClient_Query_WithFilter(t *testing.T) {
	var receivedReq QueryRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedReq)
		_ = json.NewEncoder(w).Encode(QueryResponse{})
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	filter := map[string]interface{}{
		"$and": []map[string]interface{}{
			{"category": map[string]interface{}{"$eq": "tech"}},
			{"price": map[string]interface{}{"$lt": 100}},
		},
	}
	req := &QueryRequest{
		Vector:          []float32{0.1, 0.2, 0.3},
		TopK:            10,
		Filter:          filter,
		IncludeMetadata: true,
	}
	_, err := client.Query(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, receivedReq.Filter)
}

func TestClient_Query_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	_, err := client.Query(context.Background(), &QueryRequest{
		Vector: []float32{0.1},
		TopK:   5,
	})
	assert.Error(t, err)
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestClient_Delete_WithFilter(t *testing.T) {
	var receivedReq DeleteRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedReq)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	filter := map[string]interface{}{"category": "old"}
	err := client.Delete(context.Background(), &DeleteRequest{
		Filter:    filter,
		Namespace: "test-ns",
	})
	require.NoError(t, err)
	assert.NotNil(t, receivedReq.Filter)
}

func TestClient_Delete_All(t *testing.T) {
	var receivedReq DeleteRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedReq)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	err := client.Delete(context.Background(), &DeleteRequest{
		DeleteAll: true,
		Namespace: "test-ns",
	})
	require.NoError(t, err)
	assert.True(t, receivedReq.DeleteAll)
}

func TestClient_Delete_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	err := client.Delete(context.Background(), &DeleteRequest{IDs: []string{"id1"}})
	assert.Error(t, err)
}

// =============================================================================
// Fetch Tests
// =============================================================================

func TestClient_Fetch_WithNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		assert.Contains(t, r.URL.RawQuery, "namespace=test-ns")
		_ = json.NewEncoder(w).Encode(FetchResponse{
			Vectors: map[string]Vector{
				"v1": {ID: "v1", Values: []float32{0.1}},
			},
		})
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	result, err := client.Fetch(context.Background(), []string{"v1"}, "test-ns")
	require.NoError(t, err)
	assert.Len(t, result.Vectors, 1)
}

func TestClient_Fetch_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	_, err := client.Fetch(context.Background(), []string{"v1"}, "")
	assert.Error(t, err)
}

// =============================================================================
// Update Tests
// =============================================================================

func TestClient_Update_WithValues(t *testing.T) {
	var receivedReq UpdateRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedReq)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	err := client.Update(context.Background(), &UpdateRequest{
		ID:     "v1",
		Values: []float32{0.5, 0.6, 0.7},
	})
	require.NoError(t, err)
	assert.Equal(t, "v1", receivedReq.ID)
	assert.Len(t, receivedReq.Values, 3)
}

func TestClient_Update_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	err := client.Update(context.Background(), &UpdateRequest{ID: "nonexistent"})
	assert.Error(t, err)
}

// =============================================================================
// DescribeIndexStats Tests
// =============================================================================

func TestClient_DescribeIndexStats_WithFilter(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			_ = json.Unmarshal(body, &receivedBody)
		}
		_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{
			Dimension:        768,
			TotalVectorCount: 1000,
		})
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	filter := map[string]interface{}{"category": "tech"}
	stats, err := client.DescribeIndexStats(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, 768, stats.Dimension)
	assert.NotNil(t, receivedBody["filter"])
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestClient_ConcurrentOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}
		if r.URL.Path == "/query" {
			_ = json.NewEncoder(w).Encode(QueryResponse{
				Matches: []ScoredVector{{ID: "result", Score: 0.9}},
			})
			return
		}
		if r.URL.Path == "/vectors/upsert" {
			_ = json.NewEncoder(w).Encode(UpsertResponse{UpsertedCount: 1})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)

	var wg sync.WaitGroup
	numGoroutines := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			if idx%2 == 0 {
				// Query
				_, err := client.Query(context.Background(), &QueryRequest{
					Vector: []float32{0.1, 0.2},
					TopK:   5,
				})
				assert.NoError(t, err)
			} else {
				// Upsert
				_, err := client.Upsert(context.Background(), []Vector{
					{ID: "v1", Values: []float32{0.1, 0.2}},
				}, "")
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()
	assert.True(t, client.IsConnected())
}

// =============================================================================
// Headers Tests
// =============================================================================

func TestClient_APIKeyHeader(t *testing.T) {
	var receivedAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("Api-Key")
		_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "my-secret-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	_ = client.Connect(context.Background())
	assert.Equal(t, "my-secret-api-key", receivedAPIKey)
}

func TestClient_ContentTypeHeader(t *testing.T) {
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
	}))
	defer server.Close()

	client := createConnectedTestClient(t, server)
	_ = client.HealthCheck(context.Background())
	assert.Equal(t, "application/json", receivedContentType)
}

// =============================================================================
// Helper Functions
// =============================================================================

func createConnectedTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	err = client.Connect(context.Background())
	require.NoError(t, err)

	return client
}
