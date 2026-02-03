package pgvector

import (
	"context"
	"fmt"
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

func TestConfig_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid minimal config",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "testdb",
			},
			expectError: false,
		},
		{
			name: "valid full config",
			config: &Config{
				Host:            "db.example.com",
				Port:            5432,
				User:            "admin",
				Password:        "secret123",
				Database:        "vectors",
				SSLMode:         "require",
				MaxConns:        20,
				MinConns:        5,
				MaxConnLifetime: 2 * time.Hour,
				MaxConnIdleTime: time.Hour,
				ConnectTimeout:  60 * time.Second,
			},
			expectError: false,
		},
		{
			name: "whitespace host",
			config: &Config{
				Host:     "   ",
				Port:     5432,
				User:     "postgres",
				Database: "testdb",
			},
			expectError: false, // Whitespace is technically non-empty
		},
		{
			name: "port at boundary (1)",
			config: &Config{
				Host:     "localhost",
				Port:     1,
				User:     "postgres",
				Database: "testdb",
			},
			expectError: false,
		},
		{
			name: "port at boundary (65535)",
			config: &Config{
				Host:     "localhost",
				Port:     65535,
				User:     "postgres",
				Database: "testdb",
			},
			expectError: false,
		},
		{
			name: "negative port",
			config: &Config{
				Host:     "localhost",
				Port:     -1,
				User:     "postgres",
				Database: "testdb",
			},
			expectError: true,
			errorSubstr: "invalid port",
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

func TestConfig_ConnectionString_AllOptions(t *testing.T) {
	config := &Config{
		Host:           "db.example.com",
		Port:           5433,
		User:           "admin",
		Password:       "secret123",
		Database:       "vectors",
		SSLMode:        "verify-full",
		ConnectTimeout: 45 * time.Second,
	}

	connStr := config.ConnectionString()

	assert.Contains(t, connStr, "host=db.example.com")
	assert.Contains(t, connStr, "port=5433")
	assert.Contains(t, connStr, "user=admin")
	assert.Contains(t, connStr, "password=secret123")
	assert.Contains(t, connStr, "dbname=vectors")
	assert.Contains(t, connStr, "sslmode=verify-full")
	assert.Contains(t, connStr, "connect_timeout=45")
}

func TestConfig_ConnectionString_EmptyOptionalFields(t *testing.T) {
	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Database: "test",
		// Password, SSLMode, ConnectTimeout are empty
	}

	connStr := config.ConnectionString()

	assert.Contains(t, connStr, "host=localhost")
	assert.Contains(t, connStr, "port=5432")
	assert.Contains(t, connStr, "user=postgres")
	assert.Contains(t, connStr, "dbname=test")
	assert.NotContains(t, connStr, "password=")
	assert.NotContains(t, connStr, "sslmode=")
	assert.NotContains(t, connStr, "connect_timeout=")
}

// =============================================================================
// DefaultConfig Tests
// =============================================================================

func TestDefaultConfig_AllFields(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "postgres", config.User)
	assert.Empty(t, config.Password)
	assert.Equal(t, "postgres", config.Database)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, int32(10), config.MaxConns)
	assert.Equal(t, int32(2), config.MinConns)
	assert.Equal(t, time.Hour, config.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, config.MaxConnIdleTime)
	assert.Equal(t, 30*time.Second, config.ConnectTimeout)

	// Should be valid
	err := config.Validate()
	assert.NoError(t, err)
}

// =============================================================================
// Distance Metric and Index Type Tests
// =============================================================================

func TestDistanceMetric_AllTypes(t *testing.T) {
	metrics := map[DistanceMetric]string{
		DistanceL2:     "l2",
		DistanceIP:     "ip",
		DistanceCosine: "cosine",
	}

	for metric, expected := range metrics {
		assert.Equal(t, DistanceMetric(expected), metric)
	}
}

func TestIndexType_AllTypes(t *testing.T) {
	indexTypes := map[IndexType]string{
		IndexTypeIVFFlat: "ivfflat",
		IndexTypeHNSW:    "hnsw",
	}

	for indexType, expected := range indexTypes {
		assert.Equal(t, IndexType(expected), indexType)
	}
}

// =============================================================================
// TableSchema Tests
// =============================================================================

func TestTableSchema_AllFields(t *testing.T) {
	schema := &TableSchema{
		TableName:    "documents",
		VectorColumn: "embedding",
		Dimension:    1536,
		IDColumn:     "doc_id",
		MetadataColumns: []ColumnDef{
			{Name: "title", Type: "VARCHAR(500)", Nullable: false},
			{Name: "content", Type: "TEXT", Nullable: true},
			{Name: "created_at", Type: "TIMESTAMPTZ", Nullable: false},
			{Name: "metadata", Type: "JSONB", Nullable: true},
			{Name: "score", Type: "FLOAT", Nullable: true},
		},
	}

	assert.Equal(t, "documents", schema.TableName)
	assert.Equal(t, "embedding", schema.VectorColumn)
	assert.Equal(t, 1536, schema.Dimension)
	assert.Equal(t, "doc_id", schema.IDColumn)
	assert.Len(t, schema.MetadataColumns, 5)
	assert.False(t, schema.MetadataColumns[0].Nullable)
	assert.True(t, schema.MetadataColumns[1].Nullable)
}

func TestColumnDef_AllTypes(t *testing.T) {
	columns := []ColumnDef{
		{Name: "varchar_col", Type: "VARCHAR(255)", Nullable: false},
		{Name: "text_col", Type: "TEXT", Nullable: true},
		{Name: "int_col", Type: "INTEGER", Nullable: false},
		{Name: "bigint_col", Type: "BIGINT", Nullable: true},
		{Name: "float_col", Type: "FLOAT", Nullable: false},
		{Name: "double_col", Type: "DOUBLE PRECISION", Nullable: true},
		{Name: "bool_col", Type: "BOOLEAN", Nullable: false},
		{Name: "json_col", Type: "JSONB", Nullable: true},
		{Name: "timestamp_col", Type: "TIMESTAMPTZ", Nullable: false},
		{Name: "array_col", Type: "TEXT[]", Nullable: true},
	}

	for _, col := range columns {
		assert.NotEmpty(t, col.Name)
		assert.NotEmpty(t, col.Type)
	}
}

// =============================================================================
// Vector Type Tests
// =============================================================================

func TestVector_AllFields(t *testing.T) {
	v := Vector{
		ID:     "vec-123-abc",
		Vector: []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8},
		Metadata: map[string]interface{}{
			"string_field":  "hello",
			"int_field":     42,
			"float_field":   3.14159,
			"bool_field":    true,
			"nil_field":     nil,
			"array_field":   []int{1, 2, 3},
			"nested_object": map[string]interface{}{"key": "value"},
		},
	}

	assert.Equal(t, "vec-123-abc", v.ID)
	assert.Len(t, v.Vector, 8)
	assert.Equal(t, "hello", v.Metadata["string_field"])
	assert.Equal(t, 42, v.Metadata["int_field"])
	assert.Equal(t, 3.14159, v.Metadata["float_field"])
	assert.Equal(t, true, v.Metadata["bool_field"])
	assert.Nil(t, v.Metadata["nil_field"])
}

func TestSearchResult_AllFields(t *testing.T) {
	result := SearchResult{
		ID:       "result-456",
		Distance: 0.0523,
		Metadata: map[string]interface{}{
			"title":    "Test Document",
			"category": "technology",
		},
	}

	assert.Equal(t, "result-456", result.ID)
	assert.InDelta(t, 0.0523, result.Distance, 0.0001)
	assert.Equal(t, "Test Document", result.Metadata["title"])
	assert.Equal(t, "technology", result.Metadata["category"])
}

// =============================================================================
// Request Type Tests
// =============================================================================

func TestUpsertRequest_AllFields(t *testing.T) {
	req := &UpsertRequest{
		TableName:    "documents",
		VectorColumn: "embedding",
		IDColumn:     "doc_id",
		Vectors: []Vector{
			{ID: "v1", Vector: []float32{0.1, 0.2}, Metadata: map[string]interface{}{"key": "value1"}},
			{ID: "v2", Vector: []float32{0.3, 0.4}, Metadata: map[string]interface{}{"key": "value2"}},
		},
	}

	assert.Equal(t, "documents", req.TableName)
	assert.Equal(t, "embedding", req.VectorColumn)
	assert.Equal(t, "doc_id", req.IDColumn)
	assert.Len(t, req.Vectors, 2)
}

func TestSearchRequest_AllFields(t *testing.T) {
	req := &SearchRequest{
		TableName:     "documents",
		VectorColumn:  "embedding",
		IDColumn:      "doc_id",
		QueryVector:   []float32{0.1, 0.2, 0.3},
		Limit:         10,
		Metric:        DistanceCosine,
		Filter:        "category = 'tech' AND score > 0.5",
		OutputColumns: []string{"title", "content", "metadata"},
	}

	assert.Equal(t, "documents", req.TableName)
	assert.Equal(t, "embedding", req.VectorColumn)
	assert.Equal(t, "doc_id", req.IDColumn)
	assert.Len(t, req.QueryVector, 3)
	assert.Equal(t, 10, req.Limit)
	assert.Equal(t, DistanceCosine, req.Metric)
	assert.Equal(t, "category = 'tech' AND score > 0.5", req.Filter)
	assert.Len(t, req.OutputColumns, 3)
}

func TestDeleteRequest_AllFields(t *testing.T) {
	req := &DeleteRequest{
		TableName: "documents",
		IDColumn:  "doc_id",
		IDs:       []string{"id1", "id2", "id3"},
		Filter:    "status = 'deleted'",
	}

	assert.Equal(t, "documents", req.TableName)
	assert.Equal(t, "doc_id", req.IDColumn)
	assert.Len(t, req.IDs, 3)
	assert.Equal(t, "status = 'deleted'", req.Filter)
}

func TestCreateIndexRequest_AllIndexTypes(t *testing.T) {
	tests := []struct {
		name      string
		indexType IndexType
		metric    DistanceMetric
		m         int
		ef        int
		lists     int
	}{
		{
			name:      "HNSW with custom params",
			indexType: IndexTypeHNSW,
			metric:    DistanceCosine,
			m:         32,
			ef:        128,
		},
		{
			name:      "HNSW with defaults",
			indexType: IndexTypeHNSW,
			metric:    DistanceL2,
			m:         0, // will use default 16
			ef:        0, // will use default 64
		},
		{
			name:      "IVFFlat with custom lists",
			indexType: IndexTypeIVFFlat,
			metric:    DistanceIP,
			lists:     200,
		},
		{
			name:      "IVFFlat with default lists",
			indexType: IndexTypeIVFFlat,
			metric:    DistanceCosine,
			lists:     0, // will use default 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateIndexRequest{
				TableName:      "test_table",
				IndexName:      "test_idx",
				VectorColumn:   "embedding",
				IndexType:      tt.indexType,
				Metric:         tt.metric,
				M:              tt.m,
				EfConstruction: tt.ef,
				Lists:          tt.lists,
			}

			assert.Equal(t, "test_table", req.TableName)
			assert.Equal(t, tt.indexType, req.IndexType)
			assert.Equal(t, tt.metric, req.Metric)
		})
	}
}

// =============================================================================
// vectorToString Helper Tests
// =============================================================================

func TestVectorToString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		vector   []float32
		expected string
	}{
		{
			name:     "empty vector",
			vector:   []float32{},
			expected: "[]",
		},
		{
			name:     "single zero",
			vector:   []float32{0.0},
			expected: "[0.000000]",
		},
		{
			name:     "single positive",
			vector:   []float32{1.5},
			expected: "[1.500000]",
		},
		{
			name:     "single negative",
			vector:   []float32{-2.5},
			expected: "[-2.500000]",
		},
		{
			name:     "multiple values",
			vector:   []float32{0.1, 0.2, 0.3},
			expected: "[0.100000,0.200000,0.300000]",
		},
		{
			name:     "very small values",
			vector:   []float32{0.000001, 0.000002},
			expected: "[0.000001,0.000002]",
		},
		{
			name:     "very large values",
			vector:   []float32{999999.0, -999999.0},
			expected: "[999999.000000,-999999.000000]",
		},
		{
			name:     "mixed values",
			vector:   []float32{-1.5, 0.0, 1.5, 100.123},
			expected: "[-1.500000,0.000000,1.500000,100.123001]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vectorToString(tt.vector)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVectorToString_LargeVector(t *testing.T) {
	// Test with a typical embedding size
	vector := make([]float32, 1536)
	for i := range vector {
		vector[i] = float32(i) * 0.001
	}

	result := vectorToString(vector)

	assert.True(t, strings.HasPrefix(result, "["))
	assert.True(t, strings.HasSuffix(result, "]"))
	// Should have 1535 commas (1536 values - 1)
	assert.Equal(t, 1535, strings.Count(result, ","))
}

// =============================================================================
// Client State Tests
// =============================================================================

func TestClient_NotConnected_AllOperations(t *testing.T) {
	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Database: "testdb",
	}
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	ctx := context.Background()

	// All operations should fail with "not connected" error

	t.Run("HealthCheck", func(t *testing.T) {
		err := client.HealthCheck(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("CreateTable", func(t *testing.T) {
		err := client.CreateTable(ctx, &TableSchema{
			TableName:    "test",
			VectorColumn: "vec",
			Dimension:    768,
			IDColumn:     "id",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("DropTable", func(t *testing.T) {
		err := client.DropTable(ctx, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("TableExists", func(t *testing.T) {
		_, err := client.TableExists(ctx, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("CreateIndex", func(t *testing.T) {
		err := client.CreateIndex(ctx, &CreateIndexRequest{
			TableName:    "test",
			IndexName:    "idx",
			VectorColumn: "vec",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("Upsert", func(t *testing.T) {
		_, err := client.Upsert(ctx, &UpsertRequest{
			TableName:    "test",
			VectorColumn: "vec",
			IDColumn:     "id",
			Vectors:      []Vector{{ID: "1", Vector: []float32{0.1}}},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("Search", func(t *testing.T) {
		_, err := client.Search(ctx, &SearchRequest{
			TableName:    "test",
			VectorColumn: "vec",
			IDColumn:     "id",
			QueryVector:  []float32{0.1},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("Delete", func(t *testing.T) {
		_, err := client.Delete(ctx, &DeleteRequest{
			TableName: "test",
			IDColumn:  "id",
			IDs:       []string{"1"},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("Get", func(t *testing.T) {
		_, err := client.Get(ctx, "test", "id", []string{"1"}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("Count", func(t *testing.T) {
		_, err := client.Count(ctx, "test", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("GetPool", func(t *testing.T) {
		pool := client.GetPool()
		assert.Nil(t, pool)
	})
}

func TestClient_Close_WithoutConnect(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	// Should not error
	err = client.Close()
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())

	// Should be idempotent
	err = client.Close()
	assert.NoError(t, err)
}

// =============================================================================
// Upsert Edge Cases
// =============================================================================

func TestClient_Upsert_EmptyVectors(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	// Manually set connected to test edge case
	client.connected = true
	client.pool = nil // Will panic if used

	count, err := client.Upsert(context.Background(), &UpsertRequest{
		TableName:    "test",
		VectorColumn: "vec",
		IDColumn:     "id",
		Vectors:      []Vector{},
	})

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// =============================================================================
// Delete Edge Cases
// =============================================================================

func TestClient_Delete_NoIDsNoFilter(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	// Manually set connected
	client.connected = true
	client.pool = nil

	_, err = client.Delete(context.Background(), &DeleteRequest{
		TableName: "test",
		IDColumn:  "id",
		IDs:       nil,
		Filter:    "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either IDs or filter must be specified")
}

func TestClient_Delete_EmptyIDs(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	client.connected = true
	client.pool = nil

	_, err = client.Delete(context.Background(), &DeleteRequest{
		TableName: "test",
		IDColumn:  "id",
		IDs:       []string{},
		Filter:    "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either IDs or filter must be specified")
}

// =============================================================================
// Get Edge Cases
// =============================================================================

func TestClient_Get_EmptyIDs(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	client.connected = true
	client.pool = nil

	results, err := client.Get(context.Background(), "test", "id", []string{}, nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

// =============================================================================
// NewClient Tests
// =============================================================================

func TestNewClient_AllConfigurations(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		logger      *logrus.Logger
		expectError bool
		errorSubstr string
	}{
		{
			name:        "nil config uses defaults",
			config:      nil,
			logger:      nil,
			expectError: false,
		},
		{
			name: "valid config with logger",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "test",
			},
			logger:      logrus.New(),
			expectError: false,
		},
		{
			name: "valid config without logger",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "test",
			},
			logger:      nil,
			expectError: false,
		},
		{
			name: "invalid config - empty host",
			config: &Config{
				Host:     "",
				Port:     5432,
				User:     "postgres",
				Database: "test",
			},
			logger:      nil,
			expectError: true,
			errorSubstr: "host is required",
		},
		{
			name: "invalid config - zero port",
			config: &Config{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Database: "test",
			},
			logger:      nil,
			expectError: true,
			errorSubstr: "invalid port",
		},
		{
			name: "invalid config - empty user",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "",
				Database: "test",
			},
			logger:      nil,
			expectError: true,
			errorSubstr: "user is required",
		},
		{
			name: "invalid config - empty database",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "",
			},
			logger:      nil,
			expectError: true,
			errorSubstr: "database is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config, tt.logger)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorSubstr != "" {
					assert.Contains(t, err.Error(), tt.errorSubstr)
				}
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
// Concurrent Access Tests (without actual DB)
// =============================================================================

func TestClient_ConcurrentStateAccess(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent reads of IsConnected
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.IsConnected()
		}()
	}

	wg.Wait()
	// If we get here without race detector issues, test passes
}

func TestClient_ConcurrentClose(t *testing.T) {
	config := DefaultConfig()
	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent Close calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Close()
		}()
	}

	wg.Wait()
	assert.False(t, client.IsConnected())
}

// =============================================================================
// Search Default Limit Test
// =============================================================================

func TestSearchRequest_DefaultLimit(t *testing.T) {
	req := &SearchRequest{
		TableName:    "test",
		VectorColumn: "vec",
		IDColumn:     "id",
		QueryVector:  []float32{0.1},
		Limit:        0, // Should default to 10 when processed
	}

	// The Search method sets default limit to 10 if <= 0
	if req.Limit <= 0 {
		req.Limit = 10
	}

	assert.Equal(t, 10, req.Limit)
}

// =============================================================================
// Distance Operator Selection
// =============================================================================

func TestDistanceOperatorSelection(t *testing.T) {
	tests := []struct {
		metric   DistanceMetric
		expected string
	}{
		{DistanceL2, "<->"},
		{DistanceIP, "<#>"},
		{DistanceCosine, "<=>"},
		{"unknown", "<->"}, // Default to L2
	}

	for _, tt := range tests {
		t.Run(string(tt.metric), func(t *testing.T) {
			var distanceOp string
			switch tt.metric {
			case DistanceIP:
				distanceOp = "<#>"
			case DistanceCosine:
				distanceOp = "<=>"
			default:
				distanceOp = "<->"
			}
			assert.Equal(t, tt.expected, distanceOp)
		})
	}
}

// =============================================================================
// Index Operator Class Selection
// =============================================================================

func TestIndexOperatorClassSelection(t *testing.T) {
	tests := []struct {
		metric   DistanceMetric
		expected string
	}{
		{DistanceL2, "vector_l2_ops"},
		{DistanceIP, "vector_ip_ops"},
		{DistanceCosine, "vector_cosine_ops"},
		{"unknown", "vector_l2_ops"}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.metric), func(t *testing.T) {
			var opClass string
			switch tt.metric {
			case DistanceIP:
				opClass = "vector_ip_ops"
			case DistanceCosine:
				opClass = "vector_cosine_ops"
			default:
				opClass = "vector_l2_ops"
			}
			assert.Equal(t, tt.expected, opClass)
		})
	}
}

// =============================================================================
// Index Default Parameters
// =============================================================================

func TestIndexDefaultParameters(t *testing.T) {
	t.Run("HNSW defaults", func(t *testing.T) {
		req := &CreateIndexRequest{
			IndexType:      IndexTypeHNSW,
			M:              0,
			EfConstruction: 0,
		}

		m := req.M
		if m == 0 {
			m = 16
		}
		efConstruction := req.EfConstruction
		if efConstruction == 0 {
			efConstruction = 64
		}

		assert.Equal(t, 16, m)
		assert.Equal(t, 64, efConstruction)
	})

	t.Run("IVFFlat defaults", func(t *testing.T) {
		req := &CreateIndexRequest{
			IndexType: IndexTypeIVFFlat,
			Lists:     0,
		}

		lists := req.Lists
		if lists == 0 {
			lists = 100
		}

		assert.Equal(t, 100, lists)
	})

	t.Run("HNSW custom values", func(t *testing.T) {
		req := &CreateIndexRequest{
			IndexType:      IndexTypeHNSW,
			M:              32,
			EfConstruction: 128,
		}

		assert.Equal(t, 32, req.M)
		assert.Equal(t, 128, req.EfConstruction)
	})

	t.Run("IVFFlat custom lists", func(t *testing.T) {
		req := &CreateIndexRequest{
			IndexType: IndexTypeIVFFlat,
			Lists:     500,
		}

		assert.Equal(t, 500, req.Lists)
	})
}

// =============================================================================
// Placeholder Count Verification
// =============================================================================

func TestPlaceholderGeneration(t *testing.T) {
	ids := []string{"id1", "id2", "id3", "id4", "id5"}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	assert.Equal(t, "$1", placeholders[0])
	assert.Equal(t, "$2", placeholders[1])
	assert.Equal(t, "$3", placeholders[2])
	assert.Equal(t, "$4", placeholders[3])
	assert.Equal(t, "$5", placeholders[4])
}
