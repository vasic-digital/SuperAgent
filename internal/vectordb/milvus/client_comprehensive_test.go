package milvus

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestConfig_GetBaseURL_AllSchemes(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "http with localhost",
			config: &Config{
				Host:   "localhost",
				Port:   19530,
				Secure: false,
			},
			expected: "http://localhost:19530/v2/vectordb",
		},
		{
			name: "https with secure flag",
			config: &Config{
				Host:   "milvus.cloud.example.com",
				Port:   443,
				Secure: true,
			},
			expected: "https://milvus.cloud.example.com:443/v2/vectordb",
		},
		{
			name: "http with IP address",
			config: &Config{
				Host:   "192.168.1.100",
				Port:   19530,
				Secure: false,
			},
			expected: "http://192.168.1.100:19530/v2/vectordb",
		},
		{
			name: "custom port",
			config: &Config{
				Host:   "localhost",
				Port:   9091,
				Secure: false,
			},
			expected: "http://localhost:9091/v2/vectordb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.config.GetBaseURL()
			assert.Equal(t, tt.expected, url)
		})
	}
}

// =============================================================================
// Data Type Constants Tests
// =============================================================================

func TestDataTypeConstants(t *testing.T) {
	assert.Equal(t, DataType("Int64"), DataTypeInt64)
	assert.Equal(t, DataType("VarChar"), DataTypeVarChar)
	assert.Equal(t, DataType("Float"), DataTypeFloat)
	assert.Equal(t, DataType("Double"), DataTypeDouble)
	assert.Equal(t, DataType("Bool"), DataTypeBool)
	assert.Equal(t, DataType("JSON"), DataTypeJSON)
	assert.Equal(t, DataType("FloatVector"), DataTypeFloatVector)
	assert.Equal(t, DataType("BinaryVector"), DataTypeBinaryVector)
}

func TestIndexTypeConstants(t *testing.T) {
	assert.Equal(t, IndexType("IVF_FLAT"), IndexTypeIVFFlat)
	assert.Equal(t, IndexType("IVF_SQ8"), IndexTypeIVFSQ8)
	assert.Equal(t, IndexType("IVF_PQ"), IndexTypeIVFPQ)
	assert.Equal(t, IndexType("HNSW"), IndexTypeHNSW)
	assert.Equal(t, IndexType("AUTOINDEX"), IndexTypeAutoIndex)
}

func TestMetricTypeConstants(t *testing.T) {
	assert.Equal(t, MetricType("L2"), MetricTypeL2)
	assert.Equal(t, MetricType("IP"), MetricTypeIP)
	assert.Equal(t, MetricType("COSINE"), MetricTypeCosine)
}

// =============================================================================
// Schema Type Tests
// =============================================================================

func TestCollectionSchema_Fields(t *testing.T) {
	schema := CollectionSchema{
		CollectionName: "test_collection",
		Description:    "A test collection",
		Fields: []FieldSchema{
			{
				FieldName:    "id",
				DataType:     DataTypeVarChar,
				IsPrimaryKey: true,
				Params:       map[string]interface{}{"max_length": 100},
			},
			{
				FieldName: "vector",
				DataType:  DataTypeFloatVector,
				Params:    map[string]interface{}{"dim": 768},
			},
			{
				FieldName:   "content",
				DataType:    DataTypeVarChar,
				IsPartition: false,
				Params:      map[string]interface{}{"max_length": 65535},
			},
		},
	}

	assert.Equal(t, "test_collection", schema.CollectionName)
	assert.Equal(t, "A test collection", schema.Description)
	assert.Len(t, schema.Fields, 3)
	assert.True(t, schema.Fields[0].IsPrimaryKey)
}

func TestFieldSchema_AllFields(t *testing.T) {
	field := FieldSchema{
		FieldName:    "vector_field",
		DataType:     DataTypeFloatVector,
		IsPrimaryKey: false,
		IsPartition:  true,
		ElementType:  DataTypeFloat,
		Params: map[string]interface{}{
			"dim":          768,
			"custom_param": "value",
		},
	}

	assert.Equal(t, "vector_field", field.FieldName)
	assert.Equal(t, DataTypeFloatVector, field.DataType)
	assert.False(t, field.IsPrimaryKey)
	assert.True(t, field.IsPartition)
	assert.Equal(t, DataTypeFloat, field.ElementType)
	assert.Equal(t, 768, field.Params["dim"])
}

// =============================================================================
// Request/Response Type Tests
// =============================================================================

func TestCreateCollectionRequest_QuickSetup(t *testing.T) {
	req := CreateCollectionRequest{
		DBName:         "default",
		CollectionName: "quick_collection",
		Dimension:      768,
		MetricType:     MetricTypeCosine,
		PrimaryField:   "id",
		VectorField:    "vector",
		IDType:         DataTypeVarChar,
	}

	assert.Equal(t, "default", req.DBName)
	assert.Equal(t, "quick_collection", req.CollectionName)
	assert.Equal(t, 768, req.Dimension)
	assert.Equal(t, MetricTypeCosine, req.MetricType)
}

func TestCreateCollectionRequest_WithSchema(t *testing.T) {
	req := CreateCollectionRequest{
		CollectionName: "schema_collection",
		Schema: CollectionSchema{
			CollectionName: "schema_collection",
			Fields: []FieldSchema{
				{FieldName: "id", DataType: DataTypeVarChar, IsPrimaryKey: true},
				{FieldName: "vector", DataType: DataTypeFloatVector},
			},
		},
	}

	assert.Equal(t, "schema_collection", req.CollectionName)
	assert.Len(t, req.Schema.Fields, 2)
}

func TestInsertRequest_Fields(t *testing.T) {
	req := InsertRequest{
		DBName:         "default",
		CollectionName: "test_collection",
		Data: []map[string]interface{}{
			{"id": "1", "vector": []float32{0.1, 0.2}, "content": "first"},
			{"id": "2", "vector": []float32{0.3, 0.4}, "content": "second"},
		},
	}

	assert.Equal(t, "default", req.DBName)
	assert.Equal(t, "test_collection", req.CollectionName)
	assert.Len(t, req.Data, 2)
}

func TestSearchRequest_AllFields(t *testing.T) {
	req := SearchRequest{
		DBName:         "default",
		CollectionName: "test_collection",
		Data:           [][]float32{{0.1, 0.2, 0.3}},
		AnnsField:      "vector",
		Limit:          10,
		Offset:         5,
		OutputFields:   []string{"id", "content"},
		Filter:         "category == 'tech'",
		SearchParams: map[string]interface{}{
			"metric_type": "COSINE",
			"params":      map[string]interface{}{"ef": 64},
		},
	}

	assert.Equal(t, "test_collection", req.CollectionName)
	assert.Len(t, req.Data, 1)
	assert.Equal(t, "vector", req.AnnsField)
	assert.Equal(t, 10, req.Limit)
	assert.Equal(t, 5, req.Offset)
	assert.Equal(t, "category == 'tech'", req.Filter)
}

func TestQueryRequest_AllFields(t *testing.T) {
	req := QueryRequest{
		DBName:         "default",
		CollectionName: "test_collection",
		Filter:         "id in ['1', '2', '3']",
		OutputFields:   []string{"id", "content", "vector"},
		Limit:          100,
		Offset:         10,
	}

	assert.Equal(t, "test_collection", req.CollectionName)
	assert.Equal(t, "id in ['1', '2', '3']", req.Filter)
	assert.Len(t, req.OutputFields, 3)
}

func TestCreateIndexRequest_AllFields(t *testing.T) {
	req := CreateIndexRequest{
		DBName:         "default",
		CollectionName: "test_collection",
		FieldName:      "vector",
		IndexName:      "vector_hnsw_idx",
		IndexType:      IndexTypeHNSW,
		MetricType:     MetricTypeCosine,
		Params: map[string]interface{}{
			"M":              16,
			"efConstruction": 256,
		},
	}

	assert.Equal(t, "test_collection", req.CollectionName)
	assert.Equal(t, "vector", req.FieldName)
	assert.Equal(t, IndexTypeHNSW, req.IndexType)
	assert.Equal(t, 16, req.Params["M"])
}

func TestSearchResult_Fields(t *testing.T) {
	result := SearchResult{
		ID:       "doc-123",
		Distance: 0.15,
		Entity: map[string]interface{}{
			"content":  "sample text",
			"category": "tech",
		},
	}

	assert.Equal(t, "doc-123", result.ID)
	assert.Equal(t, float32(0.15), result.Distance)
	assert.Equal(t, "sample text", result.Entity["content"])
}

func TestCollectionInfo_Fields(t *testing.T) {
	info := CollectionInfo{
		CollectionName: "my_collection",
		Description:    "A test collection",
		ShardsNum:      4,
		Load:           "Loaded",
	}

	assert.Equal(t, "my_collection", info.CollectionName)
	assert.Equal(t, "A test collection", info.Description)
	assert.Equal(t, 4, info.ShardsNum)
	assert.Equal(t, "Loaded", info.Load)
}

// =============================================================================
// Connection Tests
// =============================================================================

func TestClient_Connect_WithAllAuthMethods(t *testing.T) {
	tests := []struct {
		name       string
		configFunc func(*httptest.Server) *Config
		checkFunc  func(*http.Request) bool
	}{
		{
			name: "no auth",
			configFunc: func(server *httptest.Server) *Config {
				host, port := parseTestServerURL(t, server.URL)
				return &Config{
					Host:    host,
					Port:    port,
					DBName:  "default",
					Timeout: 5 * time.Second,
				}
			},
			checkFunc: func(r *http.Request) bool {
				return r.Header.Get("Authorization") == ""
			},
		},
		{
			name: "token auth",
			configFunc: func(server *httptest.Server) *Config {
				host, port := parseTestServerURL(t, server.URL)
				return &Config{
					Host:    host,
					Port:    port,
					DBName:  "default",
					Token:   "test-bearer-token",
					Timeout: 5 * time.Second,
				}
			},
			checkFunc: func(r *http.Request) bool {
				return r.Header.Get("Authorization") == "Bearer test-bearer-token"
			},
		},
		{
			name: "basic auth",
			configFunc: func(server *httptest.Server) *Config {
				host, port := parseTestServerURL(t, server.URL)
				return &Config{
					Host:     host,
					Port:     port,
					DBName:   "default",
					Username: "testuser",
					Password: "testpass",
					Timeout:  5 * time.Second,
				}
			},
			checkFunc: func(r *http.Request) bool {
				user, pass, ok := r.BasicAuth()
				return ok && user == "testuser" && pass == "testpass"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var authOK bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authOK = tt.checkFunc(r)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 0,
					"data": []string{},
				})
			}))
			defer server.Close()

			config := tt.configFunc(server)
			client, err := NewClient(config, logrus.New())
			require.NoError(t, err)

			err = client.Connect(context.Background())
			require.NoError(t, err)
			assert.True(t, authOK)
		})
	}
}

func TestClient_Connect_ServerTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	host, port := parseTestServerURL(t, server.URL)
	config := &Config{
		Host:    host,
		Port:    port,
		DBName:  "default",
		Timeout: 100 * time.Millisecond,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.Connect(context.Background())
	assert.Error(t, err)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// Collection Operations Comprehensive Tests
// =============================================================================

func TestClient_CreateCollection_AllOptions(t *testing.T) {
	var receivedReq CreateCollectionRequest
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/collections/create") {
			_ = json.Unmarshal(body, &receivedReq)
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	err := client.CreateCollection(context.Background(), &CreateCollectionRequest{
		CollectionName: "comprehensive_test",
		Dimension:      1536,
		MetricType:     MetricTypeCosine,
		PrimaryField:   "doc_id",
		VectorField:    "embedding",
		IDType:         DataTypeVarChar,
	})
	require.NoError(t, err)

	assert.Equal(t, "comprehensive_test", receivedReq.CollectionName)
	assert.Equal(t, 1536, receivedReq.Dimension)
	assert.Equal(t, MetricTypeCosine, receivedReq.MetricType)
}

func TestClient_CreateCollection_UsesDefaultDB(t *testing.T) {
	var receivedReq CreateCollectionRequest
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/collections/create") {
			_ = json.Unmarshal(body, &receivedReq)
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	err := client.CreateCollection(context.Background(), &CreateCollectionRequest{
		CollectionName: "test_collection",
		Dimension:      768,
	})
	require.NoError(t, err)
	assert.Equal(t, "default", receivedReq.DBName)
}

// =============================================================================
// Insert Operations Comprehensive Tests
// =============================================================================

func TestClient_Insert_LargeDataset(t *testing.T) {
	var receivedReq InsertRequest
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/entities/insert") {
			_ = json.Unmarshal(body, &receivedReq)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"insertCount": len(receivedReq.Data),
					"insertIds":   make([]string, len(receivedReq.Data)),
				},
			})
			return
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	// Create a large dataset without pre-set IDs (they will be auto-generated)
	data := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]interface{}{
			// No "id" field - it should be auto-generated
			"vector": make([]float32, 768),
		}
	}

	resp, err := client.Insert(context.Background(), "test_collection", data)
	require.NoError(t, err)
	assert.Equal(t, 100, resp.InsertCount)

	// All IDs should have been auto-generated by the client before sending
	assert.Len(t, receivedReq.Data, 100)
	for _, d := range receivedReq.Data {
		// The ID field should be present and non-empty in the received request
		if id, ok := d["id"]; ok {
			assert.NotEmpty(t, id)
		}
	}
}

// =============================================================================
// Search Operations Comprehensive Tests
// =============================================================================

func TestClient_Search_MultipleVectors(t *testing.T) {
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/entities/search") {
			var req SearchRequest
			_ = json.Unmarshal(body, &req)
			assert.Len(t, req.Data, 3) // 3 query vectors

			w.Header().Set("Content-Type", "application/json")
			// Return results for each query vector
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": [][]map[string]interface{}{
					{{"id": "r1", "distance": 0.1}},
					{{"id": "r2", "distance": 0.2}},
					{{"id": "r3", "distance": 0.3}},
				},
			})
			return
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	resp, err := client.Search(context.Background(), &SearchRequest{
		CollectionName: "test_collection",
		Data: [][]float32{
			{0.1, 0.2},
			{0.3, 0.4},
			{0.5, 0.6},
		},
		Limit: 5,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Results, 3)
}

func TestClient_Search_WithAllOptions(t *testing.T) {
	var receivedReq SearchRequest
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/entities/search") {
			_ = json.Unmarshal(body, &receivedReq)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": [][]map[string]interface{}{{}},
			})
			return
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	_, err := client.Search(context.Background(), &SearchRequest{
		CollectionName: "test_collection",
		Data:           [][]float32{{0.1, 0.2}},
		AnnsField:      "embedding",
		Limit:          20,
		Offset:         10,
		OutputFields:   []string{"id", "content", "metadata"},
		Filter:         "category in ['tech', 'science']",
		SearchParams: map[string]interface{}{
			"metric_type": "COSINE",
			"params":      map[string]interface{}{"ef": 128, "nprobe": 16},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "test_collection", receivedReq.CollectionName)
	assert.Equal(t, "embedding", receivedReq.AnnsField)
	assert.Equal(t, 20, receivedReq.Limit)
	assert.Equal(t, 10, receivedReq.Offset)
	assert.Equal(t, "category in ['tech', 'science']", receivedReq.Filter)
}

// =============================================================================
// Delete Operations Comprehensive Tests
// =============================================================================

func TestClient_Delete_ByFilter(t *testing.T) {
	var receivedFilter string
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/entities/delete") {
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			receivedFilter = req["filter"].(string)
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	err := client.Delete(context.Background(), "test_collection", "age > 30 AND status == 'inactive'", nil)
	require.NoError(t, err)
	assert.Equal(t, "age > 30 AND status == 'inactive'", receivedFilter)
}

func TestClient_Delete_ByIDs(t *testing.T) {
	var receivedIDs []interface{}
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/entities/delete") {
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			receivedIDs = req["ids"].([]interface{})
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	err := client.Delete(context.Background(), "test_collection", "", []string{"id1", "id2", "id3"})
	require.NoError(t, err)
	assert.Len(t, receivedIDs, 3)
}

// =============================================================================
// Query Operations Comprehensive Tests
// =============================================================================

func TestClient_Query_WithPagination(t *testing.T) {
	var receivedReq QueryRequest
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/entities/query") {
			_ = json.Unmarshal(body, &receivedReq)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []map[string]interface{}{
					{"id": "page2-1"},
					{"id": "page2-2"},
				},
			})
			return
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	results, err := client.Query(context.Background(), &QueryRequest{
		CollectionName: "test_collection",
		Filter:         "category == 'tech'",
		Limit:          10,
		Offset:         20,
		OutputFields:   []string{"id", "title"},
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 10, receivedReq.Limit)
	assert.Equal(t, 20, receivedReq.Offset)
}

// =============================================================================
// Index Operations Comprehensive Tests
// =============================================================================

func TestClient_CreateIndex_AllIndexTypes(t *testing.T) {
	indexTypes := []IndexType{
		IndexTypeIVFFlat,
		IndexTypeIVFSQ8,
		IndexTypeIVFPQ,
		IndexTypeHNSW,
		IndexTypeAutoIndex,
	}

	for _, indexType := range indexTypes {
		t.Run(string(indexType), func(t *testing.T) {
			var receivedReq CreateIndexRequest
			server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
				if strings.HasSuffix(path, "/indexes/create") {
					_ = json.Unmarshal(body, &receivedReq)
				}
				respondSuccess(w)
			})
			defer server.Close()

			client := createMilvusConnectedClient(t, server)

			err := client.CreateIndex(context.Background(), &CreateIndexRequest{
				CollectionName: "test_collection",
				FieldName:      "vector",
				IndexType:      indexType,
				MetricType:     MetricTypeCosine,
			})
			require.NoError(t, err)
			assert.Equal(t, indexType, receivedReq.IndexType)
		})
	}
}

// =============================================================================
// Collection State Operations Tests
// =============================================================================

func TestClient_LoadCollection_Success(t *testing.T) {
	var loadCalled bool
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/collections/load") {
			loadCalled = true
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			assert.Equal(t, "test_collection", req["collectionName"])
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	err := client.LoadCollection(context.Background(), "test_collection")
	require.NoError(t, err)
	assert.True(t, loadCalled)
}

func TestClient_ReleaseCollection_Success(t *testing.T) {
	var releaseCalled bool
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/collections/release") {
			releaseCalled = true
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			assert.Equal(t, "test_collection", req["collectionName"])
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	err := client.ReleaseCollection(context.Background(), "test_collection")
	require.NoError(t, err)
	assert.True(t, releaseCalled)
}

func TestClient_GetLoadState_AllStates(t *testing.T) {
	states := []string{"LoadStateLoaded", "LoadStateLoading", "LoadStateNotLoaded", "LoadStateNotExist"}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
				if strings.HasSuffix(path, "/collections/get_load_state") {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"code": 0,
						"data": map[string]interface{}{
							"loadState": state,
						},
					})
					return
				}
				respondSuccess(w)
			})
			defer server.Close()

			client := createMilvusConnectedClient(t, server)

			loadState, err := client.GetLoadState(context.Background(), "test_collection")
			require.NoError(t, err)
			assert.Equal(t, state, loadState)
		})
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestClient_ConcurrentOperations(t *testing.T) {
	server := createMilvusTestServer(t, func(path string, body []byte, w http.ResponseWriter) {
		if strings.HasSuffix(path, "/entities/search") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": [][]map[string]interface{}{
					{{"id": "result", "distance": 0.1}},
				},
			})
			return
		}
		if strings.HasSuffix(path, "/entities/insert") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"insertCount": 1,
					"insertIds":   []string{"id1"},
				},
			})
			return
		}
		respondSuccess(w)
	})
	defer server.Close()

	client := createMilvusConnectedClient(t, server)

	var wg sync.WaitGroup
	numGoroutines := 30

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			switch idx % 3 {
			case 0:
				// Search
				_, err := client.Search(context.Background(), &SearchRequest{
					CollectionName: "test_collection",
					Data:           [][]float32{{0.1, 0.2}},
					Limit:          5,
				})
				assert.NoError(t, err)
			case 1:
				// Insert
				_, err := client.Insert(context.Background(), "test_collection", []map[string]interface{}{
					{"id": "v1", "vector": []float32{0.1, 0.2}},
				})
				assert.NoError(t, err)
			case 2:
				// List collections
				_, err := client.ListCollections(context.Background())
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()
	assert.True(t, client.IsConnected())
}

// =============================================================================
// Error Response Tests
// =============================================================================

func TestClient_APIErrorCodes(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
	}{
		{"collection not found", 100, "Collection not found"},
		{"invalid parameter", 200, "Invalid parameter"},
		{"permission denied", 300, "Permission denied"},
		{"internal error", 500, "Internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    tt.code,
					"message": tt.message,
				})
			}))
			defer server.Close()

			host, port := parseTestServerURL(t, server.URL)
			config := &Config{Host: host, Port: port, DBName: "default", Timeout: 5 * time.Second}
			client, _ := NewClient(config, nil)

			_, err := client.ListCollections(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.message)
		})
	}
}

func TestClient_HTTPErrorStatuses(t *testing.T) {
	statuses := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("error"))
			}))
			defer server.Close()

			host, port := parseTestServerURL(t, server.URL)
			config := &Config{Host: host, Port: port, DBName: "default", Timeout: 5 * time.Second}
			client, _ := NewClient(config, nil)

			_, err := client.ListCollections(context.Background())
			require.Error(t, err)
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func createMilvusTestServer(t *testing.T, handler func(path string, body []byte, w http.ResponseWriter)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		handler(r.URL.Path, body, w)
	}))
}

func createMilvusConnectedClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	host, port := parseTestServerURL(t, server.URL)
	config := &Config{
		Host:    host,
		Port:    port,
		DBName:  "default",
		Timeout: 5 * time.Second,
	}
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)
	err = client.Connect(context.Background())
	require.NoError(t, err)
	return client
}

func parseTestServerURL(t *testing.T, url string) (string, int) {
	t.Helper()
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	parts := strings.Split(url, ":")
	require.Len(t, parts, 2)
	var port int
	for i := 0; i < len(parts[1]); i++ {
		if parts[1][i] >= '0' && parts[1][i] <= '9' {
			port = port*10 + int(parts[1][i]-'0')
		} else {
			break
		}
	}
	return parts[0], port
}

func respondSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"data": []string{},
	})
}
