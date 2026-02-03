package iceberg

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

// ==================== NewClient Comprehensive Tests ====================

func TestNewClient_AllScenarios(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		logger      *logrus.Logger
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config and nil logger",
			config:      nil,
			logger:      nil,
			expectError: false,
		},
		{
			name: "valid rest catalog",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultWriteFormat: "parquet",
				DefaultCompression: "zstd",
			},
			logger:      logrus.New(),
			expectError: false,
		},
		{
			name: "valid hive catalog",
			config: &Config{
				CatalogType: "hive",
				CatalogURI:  "thrift://localhost:9083",
				Warehouse:   "s3://warehouse",
				Timeout:     30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "valid glue catalog",
			config: &Config{
				CatalogType: "glue",
				CatalogURI:  "https://glue.us-east-1.amazonaws.com",
				Warehouse:   "s3://warehouse",
				Timeout:     30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "valid jdbc catalog",
			config: &Config{
				CatalogType: "jdbc",
				CatalogURI:  "jdbc:postgresql://localhost:5432/iceberg",
				Warehouse:   "s3://warehouse",
				Timeout:     30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "invalid catalog type",
			config: &Config{
				CatalogType: "invalid",
				CatalogURI:  "http://localhost:8181",
				Warehouse:   "s3://warehouse",
				Timeout:     30 * time.Second,
			},
			expectError: true,
			errorMsg:    "invalid config",
		},
		{
			name: "empty catalog URI",
			config: &Config{
				CatalogType: "rest",
				CatalogURI:  "",
				Warehouse:   "s3://warehouse",
				Timeout:     30 * time.Second,
			},
			expectError: true,
			errorMsg:    "catalog_uri is required",
		},
		{
			name: "empty warehouse",
			config: &Config{
				CatalogType: "rest",
				CatalogURI:  "http://localhost:8181",
				Warehouse:   "",
				Timeout:     30 * time.Second,
			},
			expectError: true,
			errorMsg:    "warehouse is required",
		},
		{
			name: "zero timeout",
			config: &Config{
				CatalogType: "rest",
				CatalogURI:  "http://localhost:8181",
				Warehouse:   "s3://warehouse",
				Timeout:     0,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "negative timeout",
			config: &Config{
				CatalogType: "rest",
				CatalogURI:  "http://localhost:8181",
				Warehouse:   "s3://warehouse",
				Timeout:     -1 * time.Second,
			},
			expectError: true,
			errorMsg:    "timeout must be positive",
		},
		{
			name: "invalid write format",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultWriteFormat: "csv",
			},
			expectError: true,
			errorMsg:    "invalid default_write_format",
		},
		{
			name: "invalid compression",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultCompression: "brotli",
			},
			expectError: true,
			errorMsg:    "invalid default_compression",
		},
		{
			name: "all valid write formats - parquet",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultWriteFormat: "parquet",
			},
			expectError: false,
		},
		{
			name: "all valid write formats - avro",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultWriteFormat: "avro",
			},
			expectError: false,
		},
		{
			name: "all valid write formats - orc",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultWriteFormat: "orc",
			},
			expectError: false,
		},
		{
			name: "all valid compressions - gzip",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultCompression: "gzip",
			},
			expectError: false,
		},
		{
			name: "all valid compressions - snappy",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultCompression: "snappy",
			},
			expectError: false,
		},
		{
			name: "all valid compressions - lz4",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultCompression: "lz4",
			},
			expectError: false,
		},
		{
			name: "all valid compressions - none",
			config: &Config{
				CatalogType:        "rest",
				CatalogURI:         "http://localhost:8181",
				Warehouse:          "s3://warehouse",
				Timeout:            30 * time.Second,
				DefaultCompression: "none",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config, tt.logger)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.False(t, client.IsConnected())
			}
		})
	}
}

// ==================== HealthCheck Coverage Tests ====================

func TestHealthCheck_RequestCreationError(t *testing.T) {
	// Test with invalid URL that causes request creation to fail
	config := &Config{
		CatalogType: "rest",
		CatalogURI:  "http://[::1]:namedport", // Invalid URL
		Warehouse:   "s3://warehouse",
		Timeout:     1 * time.Second,
	}

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.HealthCheck(context.Background())
	require.Error(t, err)
	// The error could be either "failed to create request" or "request failed" depending on how the URL is parsed
	errMsg := err.Error()
	hasExpectedError := strings.Contains(errMsg, "failed to create request") || strings.Contains(errMsg, "request failed")
	assert.True(t, hasExpectedError, "Expected error to contain 'failed to create request' or 'request failed', got: %s", errMsg)
}

// ==================== doRequest Coverage Tests ====================

func TestDoRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay response to allow context cancellation
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL
	config.Timeout = 5 * time.Second

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.GetCatalogConfig(ctx)
	require.Error(t, err)
}

func TestDoRequest_ResponseReadError(t *testing.T) {
	// This is difficult to test without modifying the code
	// The read error happens when reading response body fails
	// We can simulate partial data that causes read issues

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set content length but close connection early
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write([]byte(`{}`))
		// Flush and close (simulates incomplete response)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}))
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	// This may or may not error depending on buffering
	// Just ensure it doesn't panic
	_, _ = client.GetCatalogConfig(context.Background())
}

// ==================== Schema Tests ====================

func TestSchema_AllFieldTypes(t *testing.T) {
	fieldTypes := []struct {
		name     string
		typ      string
		required bool
	}{
		{"boolean_field", "boolean", true},
		{"int_field", "int", true},
		{"long_field", "long", true},
		{"float_field", "float", false},
		{"double_field", "double", false},
		{"decimal_field", "decimal(10,2)", false},
		{"date_field", "date", true},
		{"time_field", "time", false},
		{"timestamp_field", "timestamp", true},
		{"timestamptz_field", "timestamptz", false},
		{"string_field", "string", true},
		{"uuid_field", "uuid", false},
		{"binary_field", "binary", false},
		{"fixed_field", "fixed[16]", false},
	}

	schema := NewSchema()
	for i, ft := range fieldTypes {
		if i%2 == 0 {
			schema.AddField(i+1, ft.name, ft.typ, ft.required)
		} else {
			schema.AddFieldWithDoc(i+1, ft.name, ft.typ, ft.required, "Test documentation for "+ft.name)
		}
	}

	assert.Len(t, schema.Fields, len(fieldTypes))

	for i, ft := range fieldTypes {
		assert.Equal(t, ft.name, schema.Fields[i].Name)
		assert.Equal(t, ft.typ, schema.Fields[i].Type)
		assert.Equal(t, ft.required, schema.Fields[i].Required)
	}
}

// ==================== Partition Transform Tests ====================

func TestPartitionTransform_AllTypes(t *testing.T) {
	transforms := []PartitionTransform{
		TransformIdentity,
		TransformYear,
		TransformMonth,
		TransformDay,
		TransformHour,
		TransformBucket,
		TransformTruncate,
	}

	for _, transform := range transforms {
		t.Run(string(transform), func(t *testing.T) {
			field := NewPartitionField(1, "test_field", transform)
			assert.Equal(t, transform, field.Transform)
			assert.Equal(t, 1, field.SourceID)
			assert.Equal(t, "test_field", field.Name)
		})
	}
}

func TestPartitionField_BucketWithWidth(t *testing.T) {
	widths := []int{4, 8, 16, 32, 64, 128, 256}

	for _, width := range widths {
		field := NewPartitionField(1, "bucket_field", TransformBucket).WithWidth(width)
		assert.Equal(t, width, field.Width)
		assert.Equal(t, TransformBucket, field.Transform)
	}
}

func TestPartitionField_TruncateWithWidth(t *testing.T) {
	widths := []int{10, 100, 1000}

	for _, width := range widths {
		field := NewPartitionField(1, "truncate_field", TransformTruncate).WithWidth(width)
		assert.Equal(t, width, field.Width)
		assert.Equal(t, TransformTruncate, field.Transform)
	}
}

// ==================== SortField Tests ====================

func TestSortField_AllCombinations(t *testing.T) {
	tests := []struct {
		name         string
		direction    SortDirection
		nullOrder    NullOrder
		expectedNull NullOrder
		overrideNull bool
	}{
		{
			name:         "asc default nulls",
			direction:    SortAsc,
			expectedNull: NullsLast,
		},
		{
			name:         "desc default nulls",
			direction:    SortDesc,
			expectedNull: NullsFirst,
		},
		{
			name:         "asc with nulls first",
			direction:    SortAsc,
			nullOrder:    NullsFirst,
			expectedNull: NullsFirst,
			overrideNull: true,
		},
		{
			name:         "desc with nulls last",
			direction:    SortDesc,
			nullOrder:    NullsLast,
			expectedNull: NullsLast,
			overrideNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewSortField(1, tt.direction)
			if tt.overrideNull {
				field = field.WithNullOrder(tt.nullOrder)
			}
			assert.Equal(t, tt.expectedNull, field.NullOrder)
			assert.Equal(t, tt.direction, field.Direction)
			assert.Equal(t, "identity", field.Transform)
		})
	}
}

// ==================== TableConfig Tests ====================

func TestTableConfig_BuilderPattern(t *testing.T) {
	schema := NewSchema().
		AddField(1, "id", "long", true).
		AddField(2, "name", "string", true).
		AddField(3, "timestamp", "timestamp", true).
		AddFieldWithDoc(4, "data", "binary", false, "Raw data blob")

	config := DefaultTableConfig("production", "events").
		WithSchema(schema).
		WithPartition(NewPartitionField(1, "id_bucket", TransformBucket).WithWidth(256)).
		WithPartition(NewPartitionField(3, "ts_day", TransformDay)).
		WithPartition(NewPartitionField(3, "ts_hour", TransformHour)).
		WithSortOrder(NewSortField(3, SortDesc)).
		WithSortOrder(NewSortField(1, SortAsc)).
		WithProperty("write.format.default", "parquet").
		WithProperty("write.parquet.compression-codec", "zstd").
		WithProperty("write.target-file-size-bytes", "134217728")

	assert.Equal(t, "production", config.Namespace)
	assert.Equal(t, "events", config.Name)
	assert.Equal(t, "production.events", config.FullName())
	assert.NotNil(t, config.Schema)
	assert.Len(t, config.Schema.Fields, 4)
	assert.Len(t, config.PartitionSpec, 3)
	assert.Len(t, config.SortOrder, 2)
	assert.Len(t, config.Properties, 3)
}

func TestTableConfig_EmptyInitialization(t *testing.T) {
	config := &TableConfig{
		Namespace: "ns",
		Name:      "table",
	}

	// Test WithProperty initializes the map
	config = config.WithProperty("key", "value")
	assert.NotNil(t, config.Properties)
	assert.Equal(t, "value", config.Properties["key"])
}

func TestTableConfig_MultipleProperties(t *testing.T) {
	config := DefaultTableConfig("ns", "table").
		WithProperty("prop1", "value1").
		WithProperty("prop2", "value2").
		WithProperty("prop3", "value3").
		WithProperty("prop1", "updated") // Override

	assert.Equal(t, "updated", config.Properties["prop1"])
	assert.Equal(t, "value2", config.Properties["prop2"])
	assert.Equal(t, "value3", config.Properties["prop3"])
}

// ==================== SnapshotConfig Tests ====================

func TestSnapshotConfig_CustomValues(t *testing.T) {
	config := &SnapshotConfig{
		OlderThanDays:    14,
		RetainLast:       5,
		MaxConcurrentOps: 8,
		DryRun:           true,
	}

	assert.Equal(t, 14, config.OlderThanDays)
	assert.Equal(t, 5, config.RetainLast)
	assert.Equal(t, 8, config.MaxConcurrentOps)
	assert.True(t, config.DryRun)
}

// ==================== CompactionConfig Tests ====================

func TestCompactionConfig_CustomValues(t *testing.T) {
	config := &CompactionConfig{
		TargetFileSizeBytes:    268435456, // 256MB
		MinInputFiles:          10,
		MaxConcurrentOps:       8,
		PartialProgressEnabled: false,
	}

	assert.Equal(t, int64(268435456), config.TargetFileSizeBytes)
	assert.Equal(t, 10, config.MinInputFiles)
	assert.Equal(t, 8, config.MaxConcurrentOps)
	assert.False(t, config.PartialProgressEnabled)
}

// ==================== Namespace Type Tests ====================

func TestNamespaceConfig_Usage(t *testing.T) {
	config := NamespaceConfig{
		Name: "production",
		Properties: map[string]string{
			"owner":       "data-engineering",
			"description": "Production namespace for analytics",
			"location":    "s3://data-lake/production",
		},
	}

	assert.Equal(t, "production", config.Name)
	assert.Equal(t, "data-engineering", config.Properties["owner"])
}

// ==================== TableIdentifier Tests ====================

func TestTableIdentifier_Usage(t *testing.T) {
	id := TableIdentifier{
		Namespace: []string{"production", "analytics"},
		Name:      "user_events",
	}

	assert.Equal(t, "user_events", id.Name)
	assert.Len(t, id.Namespace, 2)
	assert.Equal(t, "production", id.Namespace[0])
	assert.Equal(t, "analytics", id.Namespace[1])
}

// ==================== Snapshot Type Tests ====================

func TestSnapshot_AllFields(t *testing.T) {
	parentID := int64(12344)
	schemaID := 1
	snapshot := Snapshot{
		SnapshotID:       12345,
		ParentSnapshotID: &parentID,
		SequenceNumber:   100,
		TimestampMs:      1706000000000,
		ManifestList:     "s3://bucket/metadata/snap-12345.avro",
		Summary: map[string]string{
			"added-records":          "1000",
			"added-data-files":       "5",
			"added-files-size":       "1073741824",
			"deleted-records":        "0",
			"deleted-data-files":     "0",
			"total-records":          "50000",
			"total-data-files":       "100",
			"total-equality-deletes": "0",
		},
		SchemaID: &schemaID,
	}

	assert.Equal(t, int64(12345), snapshot.SnapshotID)
	assert.NotNil(t, snapshot.ParentSnapshotID)
	assert.Equal(t, int64(12344), *snapshot.ParentSnapshotID)
	assert.Equal(t, int64(100), snapshot.SequenceNumber)
	assert.NotNil(t, snapshot.SchemaID)
	assert.Equal(t, 1, *snapshot.SchemaID)
	assert.Len(t, snapshot.Summary, 8)
}

func TestSnapshot_MinimalFields(t *testing.T) {
	snapshot := Snapshot{
		SnapshotID:     12345,
		SequenceNumber: 1,
		TimestampMs:    1706000000000,
		ManifestList:   "s3://bucket/metadata/snap-12345.avro",
		Summary:        map[string]string{},
	}

	assert.Nil(t, snapshot.ParentSnapshotID)
	assert.Nil(t, snapshot.SchemaID)
	assert.Empty(t, snapshot.Summary)
}

// ==================== SnapshotLogEntry Tests ====================

func TestSnapshotLogEntry_Usage(t *testing.T) {
	entry := SnapshotLogEntry{
		TimestampMs: 1706000000000,
		SnapshotID:  12345,
	}

	assert.Equal(t, int64(1706000000000), entry.TimestampMs)
	assert.Equal(t, int64(12345), entry.SnapshotID)
}

// ==================== MetadataLogEntry Tests ====================

func TestMetadataLogEntry_Usage(t *testing.T) {
	entry := MetadataLogEntry{
		TimestampMs:  1706000000000,
		MetadataFile: "s3://bucket/metadata/v1.metadata.json",
	}

	assert.Equal(t, int64(1706000000000), entry.TimestampMs)
	assert.Equal(t, "s3://bucket/metadata/v1.metadata.json", entry.MetadataFile)
}

// ==================== TableMetadata Tests ====================

func TestTableMetadata_AllFields(t *testing.T) {
	snapshotID := int64(12345)
	metadata := TableMetadata{
		FormatVersion:      2,
		TableUUID:          "f79c3e09-677c-4bbd-a479-3f349cb785e7",
		Location:           "s3://bucket/warehouse/db/table",
		LastUpdatedMs:      1706000000000,
		LastColumnID:       10,
		CurrentSchemaID:    1,
		Schemas:            []Schema{{SchemaID: 0}, {SchemaID: 1}},
		DefaultSpecID:      0,
		PartitionSpecs:     []interface{}{map[string]interface{}{"spec-id": 0}},
		LastPartitionID:    1000,
		DefaultSortOrderID: 0,
		SortOrders:         []interface{}{map[string]interface{}{"order-id": 0}},
		Properties: map[string]string{
			"write.format.default": "parquet",
		},
		CurrentSnapshotID: &snapshotID,
		Snapshots: []Snapshot{
			{SnapshotID: 12345, SequenceNumber: 1, TimestampMs: 1706000000000},
		},
		SnapshotLog: []SnapshotLogEntry{
			{TimestampMs: 1706000000000, SnapshotID: 12345},
		},
		MetadataLog: []MetadataLogEntry{
			{TimestampMs: 1706000000000, MetadataFile: "s3://bucket/metadata/v1.json"},
		},
	}

	assert.Equal(t, 2, metadata.FormatVersion)
	assert.Equal(t, "f79c3e09-677c-4bbd-a479-3f349cb785e7", metadata.TableUUID)
	assert.NotNil(t, metadata.CurrentSnapshotID)
	assert.Len(t, metadata.Schemas, 2)
	assert.Len(t, metadata.Snapshots, 1)
	assert.Len(t, metadata.SnapshotLog, 1)
	assert.Len(t, metadata.MetadataLog, 1)
}

// ==================== Config S3 Settings Tests ====================

func TestConfig_S3Settings(t *testing.T) {
	config := &Config{
		CatalogType:       "rest",
		CatalogURI:        "http://localhost:8181",
		Warehouse:         "s3://my-bucket/warehouse",
		S3Endpoint:        "http://localhost:9000",
		S3AccessKey:       "minioadmin",
		S3SecretKey:       "minioadmin",
		S3PathStyleAccess: true,
		S3Region:          "us-west-2",
		Timeout:           30 * time.Second,
	}

	err := config.Validate()
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:9000", config.S3Endpoint)
	assert.Equal(t, "minioadmin", config.S3AccessKey)
	assert.True(t, config.S3PathStyleAccess)
	assert.Equal(t, "us-west-2", config.S3Region)
}

// ==================== CatalogConfig Tests ====================

func TestCatalogConfig_Usage(t *testing.T) {
	config := CatalogConfig{
		Defaults: map[string]string{
			"warehouse":         "s3://default-warehouse",
			"io-impl":           "org.apache.iceberg.aws.s3.S3FileIO",
			"default-namespace": "default",
		},
		Overrides: map[string]string{
			"write.format.default": "parquet",
		},
	}

	assert.Equal(t, "s3://default-warehouse", config.Defaults["warehouse"])
	assert.Equal(t, "parquet", config.Overrides["write.format.default"])
}

// ==================== Error Handling Tests ====================

func TestClient_ConcurrentOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add small delay to simulate network latency
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/config" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"defaults":  map[string]string{},
				"overrides": map[string]string{},
			})
			return
		}
		if r.URL.Path == "/v1/namespaces" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"namespaces": [][]string{{"ns1"}, {"ns2"}},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL
	config.Timeout = 5 * time.Second

	client, err := NewClient(config, nil)
	require.NoError(t, err)
	err = client.Connect(context.Background())
	require.NoError(t, err)

	// Run concurrent operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = client.ListNamespaces(context.Background())
			_ = client.IsConnected()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Ensure client is still in good state
	assert.True(t, client.IsConnected())
}

// ==================== Field Struct Tests ====================

func TestField_AllProperties(t *testing.T) {
	field := Field{
		ID:       1,
		Name:     "user_id",
		Type:     "long",
		Required: true,
		Doc:      "Unique identifier for the user",
	}

	assert.Equal(t, 1, field.ID)
	assert.Equal(t, "user_id", field.Name)
	assert.Equal(t, "long", field.Type)
	assert.True(t, field.Required)
	assert.Equal(t, "Unique identifier for the user", field.Doc)
}

func TestField_OmitEmptyDoc(t *testing.T) {
	field := Field{
		ID:       1,
		Name:     "simple_field",
		Type:     "string",
		Required: false,
	}

	// Marshal to JSON and verify Doc is omitted
	data, err := json.Marshal(field)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "doc")
}

// ==================== Namespace Type Deep Tests ====================

func TestNamespace_NestedNamespace(t *testing.T) {
	ns := Namespace{
		Name: []string{"level1", "level2", "level3"},
		Properties: map[string]string{
			"location": "s3://bucket/level1/level2/level3",
		},
	}

	assert.Len(t, ns.Name, 3)
	assert.Equal(t, "level1", ns.Name[0])
	assert.Equal(t, "level3", ns.Name[2])
}
