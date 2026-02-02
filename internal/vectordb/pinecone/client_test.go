package pinecone

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: "https://test-index.pinecone.io",
		Timeout:   30 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.False(t, client.IsConnected())
}

func TestNewClient_MissingAPIKey(t *testing.T) {
	config := &Config{
		IndexHost: "https://test-index.pinecone.io",
	}

	_, err := NewClient(config, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
}

func TestNewClient_MissingIndexHost(t *testing.T) {
	config := &Config{
		APIKey: "test-api-key",
	}

	_, err := NewClient(config, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index host")
}

func TestClient_Connect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/describe_index_stats", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("Api-Key"))

		response := DescribeIndexStatsResponse{
			Dimension:        1536,
			TotalVectorCount: 1000,
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	assert.NoError(t, err)
	assert.True(t, client.IsConnected())
}

func TestClient_Connect_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "invalid api key"}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "invalid-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

func TestClient_Close(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	require.NoError(t, err)
	assert.True(t, client.IsConnected())

	err = client.Close()
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestClient_Upsert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}

		assert.Equal(t, "/vectors/upsert", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req UpsertRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Len(t, req.Vectors, 2)
		assert.Equal(t, "test-ns", req.Namespace)

		response := UpsertResponse{UpsertedCount: 2}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	vectors := []Vector{
		{ID: "vec1", Values: make([]float32, 1536), Metadata: map[string]interface{}{"key": "value1"}},
		{ID: "vec2", Values: make([]float32, 1536), Metadata: map[string]interface{}{"key": "value2"}},
	}

	result, err := client.Upsert(context.Background(), vectors, "test-ns")
	assert.NoError(t, err)
	assert.Equal(t, 2, result.UpsertedCount)
}

func TestClient_Upsert_NotConnected(t *testing.T) {
	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: "https://test.pinecone.io",
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	_, err = client.Upsert(context.Background(), []Vector{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_Upsert_EmptyVectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	result, err := client.Upsert(context.Background(), []Vector{}, "")
	assert.NoError(t, err)
	assert.Equal(t, 0, result.UpsertedCount)
}

func TestClient_Upsert_GeneratesID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}

		var req UpsertRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		// ID should be generated
		assert.NotEmpty(t, req.Vectors[0].ID)

		_ = json.NewEncoder(w).Encode(UpsertResponse{UpsertedCount: 1})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	vectors := []Vector{
		{Values: make([]float32, 1536)}, // No ID
	}

	_, err = client.Upsert(context.Background(), vectors, "")
	assert.NoError(t, err)
}

func TestClient_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}

		assert.Equal(t, "/query", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req QueryRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, 10, req.TopK)

		response := QueryResponse{
			Matches: []ScoredVector{
				{ID: "vec1", Score: 0.95},
				{ID: "vec2", Score: 0.85},
			},
			Namespace: "test-ns",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	req := &QueryRequest{
		Vector:          make([]float32, 1536),
		TopK:            10,
		Namespace:       "test-ns",
		IncludeMetadata: true,
	}

	result, err := client.Query(context.Background(), req)
	assert.NoError(t, err)
	assert.Len(t, result.Matches, 2)
	assert.Equal(t, "vec1", result.Matches[0].ID)
	assert.Equal(t, float32(0.95), result.Matches[0].Score)
}

func TestClient_Query_NotConnected(t *testing.T) {
	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: "https://test.pinecone.io",
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	_, err = client.Query(context.Background(), &QueryRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_Query_DefaultTopK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}

		var req QueryRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, 10, req.TopK) // Default

		_ = json.NewEncoder(w).Encode(QueryResponse{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	req := &QueryRequest{
		Vector: make([]float32, 1536),
		TopK:   0, // Should default to 10
	}

	_, err = client.Query(context.Background(), req)
	assert.NoError(t, err)
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}

		assert.Equal(t, "/vectors/delete", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req DeleteRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, []string{"vec1", "vec2"}, req.IDs)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	req := &DeleteRequest{
		IDs:       []string{"vec1", "vec2"},
		Namespace: "test-ns",
	}

	err = client.Delete(context.Background(), req)
	assert.NoError(t, err)
}

func TestClient_Delete_NotConnected(t *testing.T) {
	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: "https://test.pinecone.io",
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.Delete(context.Background(), &DeleteRequest{IDs: []string{"test"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_Fetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}

		assert.Contains(t, r.URL.Path, "/vectors/fetch")
		assert.Contains(t, r.URL.RawQuery, "ids=vec1")
		assert.Contains(t, r.URL.RawQuery, "ids=vec2")
		assert.Equal(t, http.MethodGet, r.Method)

		response := FetchResponse{
			Vectors: map[string]Vector{
				"vec1": {ID: "vec1", Values: make([]float32, 1536)},
				"vec2": {ID: "vec2", Values: make([]float32, 1536)},
			},
			Namespace: "test-ns",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	result, err := client.Fetch(context.Background(), []string{"vec1", "vec2"}, "test-ns")
	assert.NoError(t, err)
	assert.Len(t, result.Vectors, 2)
}

func TestClient_Fetch_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	result, err := client.Fetch(context.Background(), []string{}, "")
	assert.NoError(t, err)
	assert.Empty(t, result.Vectors)
}

func TestClient_Fetch_NotConnected(t *testing.T) {
	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: "https://test.pinecone.io",
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	_, err = client.Fetch(context.Background(), []string{"test"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_DescribeIndexStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/describe_index_stats", r.URL.Path)

		response := DescribeIndexStatsResponse{
			Dimension:        1536,
			TotalVectorCount: 5000,
			IndexFullness:    0.25,
			Namespaces: map[string]NamespaceStats{
				"ns1": {VectorCount: 3000},
				"ns2": {VectorCount: 2000},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	stats, err := client.DescribeIndexStats(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, 1536, stats.Dimension)
	assert.Equal(t, int64(5000), stats.TotalVectorCount)
	assert.Len(t, stats.Namespaces, 2)
}

func TestClient_DescribeIndexStats_NotConnected(t *testing.T) {
	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: "https://test.pinecone.io",
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	_, err = client.DescribeIndexStats(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_ListNamespaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := DescribeIndexStatsResponse{
			Namespaces: map[string]NamespaceStats{
				"ns1": {VectorCount: 1000},
				"ns2": {VectorCount: 2000},
				"ns3": {VectorCount: 500},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	namespaces, err := client.ListNamespaces(context.Background())
	assert.NoError(t, err)
	assert.Len(t, namespaces, 3)
}

func TestClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/describe_index_stats" {
			_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
			return
		}

		assert.Equal(t, "/vectors/update", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req UpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "vec1", req.ID)
		assert.Equal(t, "new-value", req.SetMetadata["key"])

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	require.NoError(t, client.Connect(context.Background()))

	req := &UpdateRequest{
		ID:          "vec1",
		SetMetadata: map[string]interface{}{"key": "new-value"},
		Namespace:   "test-ns",
	}

	err = client.Update(context.Background(), req)
	assert.NoError(t, err)
}

func TestClient_Update_NotConnected(t *testing.T) {
	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: "https://test.pinecone.io",
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.Update(context.Background(), &UpdateRequest{ID: "test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_HealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(DescribeIndexStatsResponse{})
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-api-key",
		IndexHost: server.URL,
		Timeout:   5 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	// Connect first
	require.NoError(t, client.Connect(context.Background()))

	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{"valid", &Config{APIKey: "key", IndexHost: "https://test.pinecone.io"}, false},
		{"missing api key", &Config{IndexHost: "https://test.pinecone.io"}, true},
		{"missing index host", &Config{APIKey: "key"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, 30*time.Second, config.Timeout)
}
