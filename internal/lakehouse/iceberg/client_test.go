package iceberg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		client, err := NewClient(nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.False(t, client.IsConnected())
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			CatalogType:        "rest",
			CatalogURI:         "http://iceberg.example.com:8181",
			Warehouse:          "s3://warehouse",
			Timeout:            60 * time.Second,
			DefaultWriteFormat: "parquet",
			DefaultCompression: "zstd",
		}
		client, err := NewClient(config, logrus.New())
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with invalid config - empty catalog URI", func(t *testing.T) {
		config := &Config{
			CatalogType: "rest",
			CatalogURI:  "",
			Warehouse:   "s3://warehouse",
			Timeout:     30 * time.Second,
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("with invalid config - empty warehouse", func(t *testing.T) {
		config := &Config{
			CatalogType: "rest",
			CatalogURI:  "http://localhost:8181",
			Warehouse:   "",
			Timeout:     30 * time.Second,
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClientConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/config" {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"defaults":  map[string]string{},
					"overrides": map[string]string{},
				})
				return
			}
			if r.URL.Path == "/v1/namespaces" {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"namespaces": [][]string{},
				})
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := DefaultConfig()
		config.CatalogURI = server.URL
		client, _ := NewClient(config, nil)

		err := client.Connect(context.Background())
		require.NoError(t, err)
		assert.True(t, client.IsConnected())
	})

	t.Run("connection failure", func(t *testing.T) {
		config := DefaultConfig()
		config.CatalogURI = "http://localhost:59999"
		config.Timeout = 100 * time.Millisecond

		client, _ := NewClient(config, nil)
		err := client.Connect(context.Background())
		require.Error(t, err)
		assert.False(t, client.IsConnected())
	})
}

func TestClientClose(t *testing.T) {
	client, _ := NewClient(nil, nil)
	err := client.Close()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestClientHealthCheck(t *testing.T) {
	t.Run("connection failure", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.HealthCheck(context.Background())
		require.Error(t, err)
		// HealthCheck attempts to connect to the catalog, so it will fail with request error
		assert.Contains(t, err.Error(), "request failed")
	})
}

func TestCreateNamespace(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.CreateNamespace(context.Background(), "test-ns", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestListNamespaces(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		namespaces, err := client.ListNamespaces(context.Background())
		require.Error(t, err)
		assert.Nil(t, namespaces)
	})
}

func TestGetNamespace(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		ns, err := client.GetNamespace(context.Background(), "test-ns")
		require.Error(t, err)
		assert.Nil(t, ns)
	})
}

func TestDropNamespace(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.DropNamespace(context.Background(), "test-ns")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestCreateTable(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		tableConfig := DefaultTableConfig("test-ns", "test-table")
		err := client.CreateTable(context.Background(), tableConfig)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestListTables(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		tables, err := client.ListTables(context.Background(), "test-ns")
		require.Error(t, err)
		assert.Nil(t, tables)
	})
}

func TestGetTable(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		table, err := client.GetTable(context.Background(), "test-ns", "test-table")
		require.Error(t, err)
		assert.Nil(t, table)
	})
}

func TestDropTable(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.DropTable(context.Background(), "test-ns", "test-table", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestRenameTable(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.RenameTable(context.Background(), "test-ns", "old-table", "new-table")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestTableExists(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		exists, err := client.TableExists(context.Background(), "test-ns", "test-table")
		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestGetSnapshots(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		snapshots, err := client.GetSnapshots(context.Background(), "test-ns", "test-table")
		require.Error(t, err)
		assert.Nil(t, snapshots)
	})
}

func TestGetCurrentSnapshot(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		snapshot, err := client.GetCurrentSnapshot(context.Background(), "test-ns", "test-table")
		require.Error(t, err)
		assert.Nil(t, snapshot)
	})
}

func TestSchema(t *testing.T) {
	schema := NewSchema().
		AddField(1, "id", "string", true).
		AddField(2, "name", "string", true).
		AddFieldWithDoc(3, "created_at", "timestamp", true, "Creation timestamp")

	assert.Equal(t, 0, schema.SchemaID)
	assert.Len(t, schema.Fields, 3)

	assert.Equal(t, 1, schema.Fields[0].ID)
	assert.Equal(t, "id", schema.Fields[0].Name)
	assert.Equal(t, "string", schema.Fields[0].Type)
	assert.True(t, schema.Fields[0].Required)

	assert.Equal(t, "Creation timestamp", schema.Fields[2].Doc)
}

func TestPartitionField(t *testing.T) {
	field := NewPartitionField(1, "date_day", TransformDay)
	assert.Equal(t, 1, field.SourceID)
	assert.Equal(t, "date_day", field.Name)
	assert.Equal(t, TransformDay, field.Transform)
	assert.Equal(t, 0, field.Width)

	fieldWithWidth := NewPartitionField(2, "bucket", TransformBucket).WithWidth(16)
	assert.Equal(t, 16, fieldWithWidth.Width)
}

func TestSortField(t *testing.T) {
	ascField := NewSortField(1, SortAsc)
	assert.Equal(t, 1, ascField.SourceID)
	assert.Equal(t, SortAsc, ascField.Direction)
	assert.Equal(t, NullsLast, ascField.NullOrder)
	assert.Equal(t, "identity", ascField.Transform)

	descField := NewSortField(2, SortDesc)
	assert.Equal(t, SortDesc, descField.Direction)
	assert.Equal(t, NullsFirst, descField.NullOrder)

	customNull := NewSortField(3, SortAsc).WithNullOrder(NullsFirst)
	assert.Equal(t, NullsFirst, customNull.NullOrder)
}

func TestTableConfig(t *testing.T) {
	schema := NewSchema().
		AddField(1, "id", "string", true)

	config := DefaultTableConfig("test-ns", "test-table").
		WithSchema(schema).
		WithPartition(NewPartitionField(1, "id_hash", TransformBucket).WithWidth(16)).
		WithSortOrder(NewSortField(1, SortAsc)).
		WithProperty("custom.key", "custom.value")

	assert.Equal(t, "test-ns", config.Namespace)
	assert.Equal(t, "test-table", config.Name)
	assert.Equal(t, "test-ns.test-table", config.FullName())
	assert.NotNil(t, config.Schema)
	assert.Len(t, config.PartitionSpec, 1)
	assert.Len(t, config.SortOrder, 1)
	assert.Equal(t, "custom.value", config.Properties["custom.key"])
}

func TestNamespaceType(t *testing.T) {
	ns := &Namespace{
		Name: []string{"test-namespace"},
		Properties: map[string]string{
			"owner":       "test-user",
			"description": "Test namespace",
		},
	}

	assert.Equal(t, "test-namespace", ns.Name[0])
	assert.Equal(t, "test-user", ns.Properties["owner"])
}

func TestTableMetadataType(t *testing.T) {
	snapshotID := int64(12345)
	info := &TableMetadata{
		FormatVersion:     2,
		TableUUID:         "test-uuid-1234",
		Location:          "s3://bucket/warehouse/test-ns/test-table",
		CurrentSnapshotID: &snapshotID,
		Properties: map[string]string{
			"write.format.default": "parquet",
		},
	}

	assert.Equal(t, 2, info.FormatVersion)
	assert.Equal(t, "test-uuid-1234", info.TableUUID)
	assert.Equal(t, int64(12345), *info.CurrentSnapshotID)
}

func TestSnapshotType(t *testing.T) {
	parentID := int64(12345678901233)
	info := &Snapshot{
		SnapshotID:       12345678901234,
		ParentSnapshotID: &parentID,
		SequenceNumber:   100,
		TimestampMs:      time.Now().UnixMilli(),
		ManifestList:     "s3://bucket/metadata/snap-12345.avro",
		Summary: map[string]string{
			"added-records":    "1000",
			"added-data-files": "5",
		},
	}

	assert.Equal(t, int64(12345678901234), info.SnapshotID)
	assert.Equal(t, int64(100), info.SequenceNumber)
	assert.Equal(t, "1000", info.Summary["added-records"])
}

func TestSnapshotConfig(t *testing.T) {
	config := DefaultSnapshotConfig()

	assert.Equal(t, 7, config.OlderThanDays)
	assert.Equal(t, 10, config.RetainLast)
	assert.Equal(t, 4, config.MaxConcurrentOps)
	assert.False(t, config.DryRun)
}

func TestCompactionConfig(t *testing.T) {
	config := DefaultCompactionConfig()

	assert.Equal(t, int64(134217728), config.TargetFileSizeBytes) // 128MB
	assert.Equal(t, 5, config.MinInputFiles)
	assert.Equal(t, 4, config.MaxConcurrentOps)
	assert.True(t, config.PartialProgressEnabled)
}
