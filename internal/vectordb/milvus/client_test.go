// Package milvus provides a client for Milvus vector database.
package milvus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid config",
			config: &Config{
				Host:    "localhost",
				Port:    19530,
				DBName:  "default",
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: &Config{
				Host: "",
				Port: 19530,
			},
			wantErr:   true,
			errSubstr: "host is required",
		},
		{
			name: "invalid port",
			config: &Config{
				Host: "localhost",
				Port: 0,
			},
			wantErr:   true,
			errSubstr: "invalid port",
		},
		{
			name: "with token auth",
			config: &Config{
				Host:   "localhost",
				Port:   19530,
				Token:  "test-token",
				DBName: "default",
			},
			wantErr: false,
		},
		{
			name: "with basic auth",
			config: &Config{
				Host:     "localhost",
				Port:     19530,
				Username: "user",
				Password: "pass",
				DBName:   "default",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config, nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, client)
			assert.False(t, client.IsConnected())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 19530, config.Port)
	assert.Equal(t, "default", config.DBName)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid config",
			config: &Config{
				Host: "localhost",
				Port: 19530,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: &Config{
				Host: "",
				Port: 19530,
			},
			wantErr:   true,
			errSubstr: "host is required",
		},
		{
			name: "zero port",
			config: &Config{
				Host: "localhost",
				Port: 0,
			},
			wantErr:   true,
			errSubstr: "invalid port",
		},
		{
			name: "negative port",
			config: &Config{
				Host: "localhost",
				Port: -1,
			},
			wantErr:   true,
			errSubstr: "invalid port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigGetBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "http",
			config: &Config{
				Host:   "localhost",
				Port:   19530,
				Secure: false,
			},
			expected: "http://localhost:19530/v2/vectordb",
		},
		{
			name: "https",
			config: &Config{
				Host:   "milvus.example.com",
				Port:   443,
				Secure: true,
			},
			expected: "https://milvus.example.com:443/v2/vectordb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.config.GetBaseURL()
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []string{"collection1"},
			})
		}))
		defer server.Close()

		client := createTestClient(t, server)
		err := client.Connect(context.Background())
		require.NoError(t, err)
		assert.True(t, client.IsConnected())
	})

	t.Run("connection failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    500,
				"message": "internal error",
			})
		}))
		defer server.Close()

		client := createTestClient(t, server)
		err := client.Connect(context.Background())
		assert.Error(t, err)
		assert.False(t, client.IsConnected())
	})

	t.Run("with token authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			assert.Equal(t, "Bearer test-token", auth)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []string{},
			})
		}))
		defer server.Close()

		client := createTestClientWithToken(t, server, "test-token")
		err := client.Connect(context.Background())
		require.NoError(t, err)
	})

	t.Run("with basic authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "testuser", user)
			assert.Equal(t, "testpass", pass)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []string{},
			})
		}))
		defer server.Close()

		client := createTestClientWithBasicAuth(t, server, "testuser", "testpass")
		err := client.Connect(context.Background())
		require.NoError(t, err)
	})
}

func TestClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": []string{},
		})
	}))
	defer server.Close()

	client := createTestClient(t, server)
	_ = client.Connect(context.Background())
	assert.True(t, client.IsConnected())

	err := client.Close()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestListCollections(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.True(t, strings.HasSuffix(r.URL.Path, "/collections/list"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []string{"collection1", "collection2", "collection3"},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		collections, err := client.ListCollections(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"collection1", "collection2", "collection3"}, collections)
	})

	t.Run("error response", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call (connect) succeeds
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			// Second call (actual list) fails
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		_, err := client.ListCollections(context.Background())
		assert.Error(t, err)
	})
}

func TestCreateCollection(t *testing.T) {
	t.Run("quick setup", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/collections/create"))
			var req CreateCollectionRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req.CollectionName)
			assert.Equal(t, 768, req.Dimension)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.CreateCollection(context.Background(), &CreateCollectionRequest{
			CollectionName: "test_collection",
			Dimension:      768,
			MetricType:     MetricTypeCosine,
		})
		require.NoError(t, err)
	})

	t.Run("with schema", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			var req CreateCollectionRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req.CollectionName)
			assert.Len(t, req.Schema.Fields, 2)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.CreateCollection(context.Background(), &CreateCollectionRequest{
			CollectionName: "test_collection",
			Schema: CollectionSchema{
				CollectionName: "test_collection",
				Fields: []FieldSchema{
					{FieldName: "id", DataType: DataTypeVarChar, IsPrimaryKey: true},
					{FieldName: "vector", DataType: DataTypeFloatVector, Params: map[string]interface{}{"dim": 768}},
				},
			},
		})
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		err := client.CreateCollection(context.Background(), &CreateCollectionRequest{
			CollectionName: "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestDropCollection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/collections/drop"))
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req["collectionName"])
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.DropCollection(context.Background(), "test_collection")
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		err := client.DropCollection(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestDescribeCollection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/collections/describe"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"collectionName": "test_collection",
					"description":    "Test collection",
					"fields": []map[string]interface{}{
						{"fieldName": "id", "dataType": "VarChar", "isPrimaryKey": true},
						{"fieldName": "vector", "dataType": "FloatVector", "params": map[string]interface{}{"dim": 768}},
					},
					"shardsNum": 2,
					"load":      "Loaded",
				},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		info, err := client.DescribeCollection(context.Background(), "test_collection")
		require.NoError(t, err)
		assert.Equal(t, "test_collection", info.CollectionName)
		assert.Len(t, info.Fields, 2)
		assert.Equal(t, 2, info.ShardsNum)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		_, err := client.DescribeCollection(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestInsert(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/entities/insert"))
			var req InsertRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req.CollectionName)
			assert.Len(t, req.Data, 2)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"insertCount": 2,
					"insertIds":   []string{"id1", "id2"},
				},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		data := []map[string]interface{}{
			{"id": "id1", "vector": []float32{0.1, 0.2, 0.3}},
			{"id": "id2", "vector": []float32{0.4, 0.5, 0.6}},
		}
		resp, err := client.Insert(context.Background(), "test_collection", data)
		require.NoError(t, err)
		assert.Equal(t, 2, resp.InsertCount)
		assert.Len(t, resp.InsertIDs, 2)
	})

	t.Run("empty data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			t.Error("should not make request for empty data")
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		resp, err := client.Insert(context.Background(), "test_collection", []map[string]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, 0, resp.InsertCount)
	})

	t.Run("auto-generates ids", func(t *testing.T) {
		var receivedData []map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			var req InsertRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			receivedData = req.Data
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{"insertCount": 1, "insertIds": []string{"auto-id"}},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		data := []map[string]interface{}{
			{"vector": []float32{0.1, 0.2, 0.3}}, // No ID
		}
		_, err := client.Insert(context.Background(), "test_collection", data)
		require.NoError(t, err)
		assert.NotEmpty(t, receivedData[0]["id"])
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		_, err := client.Insert(context.Background(), "test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestSearch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/entities/search"))
			var req SearchRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req.CollectionName)
			assert.Equal(t, 5, req.Limit)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": [][]map[string]interface{}{
					{
						{"id": "id1", "distance": 0.1, "entity": map[string]interface{}{"text": "hello"}},
						{"id": "id2", "distance": 0.2, "entity": map[string]interface{}{"text": "world"}},
					},
				},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		resp, err := client.Search(context.Background(), &SearchRequest{
			CollectionName: "test_collection",
			Data:           [][]float32{{0.1, 0.2, 0.3}},
			Limit:          5,
			OutputFields:   []string{"text"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Results, 1)
		assert.Len(t, resp.Results[0], 2)
	})

	t.Run("with filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			var req SearchRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "category == 'news'", req.Filter)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": [][]map[string]interface{}{{}}})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		_, err := client.Search(context.Background(), &SearchRequest{
			CollectionName: "test_collection",
			Data:           [][]float32{{0.1, 0.2, 0.3}},
			Filter:         "category == 'news'",
			Limit:          10,
		})
		require.NoError(t, err)
	})

	t.Run("default limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			var req SearchRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, 10, req.Limit) // Default
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": [][]map[string]interface{}{{}}})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		_, err := client.Search(context.Background(), &SearchRequest{
			CollectionName: "test_collection",
			Data:           [][]float32{{0.1, 0.2, 0.3}},
		})
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		_, err := client.Search(context.Background(), &SearchRequest{CollectionName: "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestDelete(t *testing.T) {
	t.Run("by ids", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/entities/delete"))
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req["collectionName"])
			ids := req["ids"].([]interface{})
			assert.Len(t, ids, 2)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.Delete(context.Background(), "test_collection", "", []string{"id1", "id2"})
		require.NoError(t, err)
	})

	t.Run("by filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "category == 'old'", req["filter"])
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.Delete(context.Background(), "test_collection", "category == 'old'", nil)
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		err := client.Delete(context.Background(), "test", "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/entities/get"))
			var req GetRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, []string{"id1", "id2"}, req.IDs)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []map[string]interface{}{
					{"id": "id1", "text": "hello"},
					{"id": "id2", "text": "world"},
				},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		results, err := client.Get(context.Background(), "test_collection", []string{"id1", "id2"}, []string{"text"})
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("empty ids", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			t.Error("should not make request for empty ids")
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		results, err := client.Get(context.Background(), "test_collection", []string{}, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		_, err := client.Get(context.Background(), "test", []string{"id1"}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestQuery(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/entities/query"))
			var req QueryRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req.CollectionName)
			assert.Equal(t, "category == 'news'", req.Filter)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": []map[string]interface{}{
					{"id": "id1", "category": "news"},
					{"id": "id2", "category": "news"},
				},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		results, err := client.Query(context.Background(), &QueryRequest{
			CollectionName: "test_collection",
			Filter:         "category == 'news'",
			OutputFields:   []string{"category"},
			Limit:          100,
		})
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		_, err := client.Query(context.Background(), &QueryRequest{CollectionName: "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestCreateIndex(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/indexes/create"))
			var req CreateIndexRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req.CollectionName)
			assert.Equal(t, "vector", req.FieldName)
			assert.Equal(t, IndexTypeHNSW, req.IndexType)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.CreateIndex(context.Background(), &CreateIndexRequest{
			CollectionName: "test_collection",
			FieldName:      "vector",
			IndexName:      "vector_index",
			IndexType:      IndexTypeHNSW,
			MetricType:     MetricTypeCosine,
			Params: map[string]interface{}{
				"M":              16,
				"efConstruction": 256,
			},
		})
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		err := client.CreateIndex(context.Background(), &CreateIndexRequest{CollectionName: "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestLoadCollection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/collections/load"))
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req["collectionName"])
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.LoadCollection(context.Background(), "test_collection")
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		err := client.LoadCollection(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestReleaseCollection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/collections/release"))
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test_collection", req["collectionName"])
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.ReleaseCollection(context.Background(), "test_collection")
		require.NoError(t, err)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		err := client.ReleaseCollection(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestGetLoadState(t *testing.T) {
	t.Run("loaded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/collections/list") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			assert.True(t, strings.HasSuffix(r.URL.Path, "/collections/get_load_state"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"loadState": "LoadStateLoaded",
				},
			})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		state, err := client.GetLoadState(context.Background(), "test_collection")
		require.NoError(t, err)
		assert.Equal(t, "LoadStateLoaded", state)
	})

	t.Run("not connected", func(t *testing.T) {
		config := DefaultConfig()
		client, _ := NewClient(config, logrus.New())
		_, err := client.GetLoadState(context.Background(), "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.HealthCheck(context.Background())
		require.NoError(t, err)
	})

	t.Run("unhealthy", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				// First call (connect) succeeds
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
				return
			}
			// Subsequent calls fail
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := createConnectedClient(t, server)
		err := client.HealthCheck(context.Background())
		assert.Error(t, err)
	})
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    100,
			"message": "Collection not found",
		})
	}))
	defer server.Close()

	client := createTestClient(t, server)
	_, err := client.ListCollections(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 100")
	assert.Contains(t, err.Error(), "Collection not found")
}

func TestContextCancellation(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call (connect) succeeds immediately
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": []string{}})
			return
		}
		// Subsequent calls are slow
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	client := createConnectedClient(t, server)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Search(ctx, &SearchRequest{CollectionName: "test"})
	// Context was cancelled, so request should fail
	assert.Error(t, err)
}

// Helper functions

func createTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	host, port := parseServerURL(t, server.URL)
	config := &Config{
		Host:    host,
		Port:    port,
		DBName:  "default",
		Timeout: 5 * time.Second,
	}
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)
	return client
}

func createTestClientWithToken(t *testing.T, server *httptest.Server, token string) *Client {
	t.Helper()
	host, port := parseServerURL(t, server.URL)
	config := &Config{
		Host:    host,
		Port:    port,
		DBName:  "default",
		Token:   token,
		Timeout: 5 * time.Second,
	}
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)
	return client
}

func createTestClientWithBasicAuth(t *testing.T, server *httptest.Server, username, password string) *Client {
	t.Helper()
	host, port := parseServerURL(t, server.URL)
	config := &Config{
		Host:     host,
		Port:     port,
		DBName:   "default",
		Username: username,
		Password: password,
		Timeout:  5 * time.Second,
	}
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)
	return client
}

func createConnectedClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client := createTestClient(t, server)
	err := client.Connect(context.Background())
	require.NoError(t, err)
	return client
}

func parseServerURL(t *testing.T, url string) (string, int) {
	t.Helper()
	// URL format: http://127.0.0.1:12345
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
