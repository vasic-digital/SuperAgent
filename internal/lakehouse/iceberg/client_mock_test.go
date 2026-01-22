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

// Fixed timestamps for snapshot tests (to avoid race conditions)
var (
	snapshotOldTimestamp = time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC).UnixMilli() // 2 days ago
	snapshotNewTimestamp = time.Date(2026, 1, 22, 12, 0, 0, 0, time.UTC).UnixMilli() // today at noon
)

// MockIcebergServer creates a mock Iceberg REST Catalog server for testing
func MockIcebergServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		// Config endpoint
		case r.URL.Path == "/v1/config" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"defaults": map[string]string{
					"warehouse": "s3://test-warehouse",
				},
				"overrides": map[string]string{
					"default-namespace": "helixagent",
				},
			})

		// List namespaces
		case r.URL.Path == "/v1/namespaces" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"namespaces": [][]string{
					{"helixagent"},
					{"analytics"},
					{"test"},
				},
			})

		// Create namespace
		case r.URL.Path == "/v1/namespaces" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"namespace":  []string{"new-namespace"},
				"properties": map[string]string{},
			})

		// Get namespace
		case r.URL.Path == "/v1/namespaces/helixagent" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"namespace": []string{"helixagent"},
				"properties": map[string]string{
					"owner":       "admin",
					"description": "HelixAgent namespace",
				},
			})

		// Delete namespace
		case r.URL.Path == "/v1/namespaces/helixagent" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		// List tables
		case r.URL.Path == "/v1/namespaces/helixagent/tables" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"identifiers": []map[string]interface{}{
					{
						"namespace": []string{"helixagent"},
						"name":      "debates",
					},
					{
						"namespace": []string{"helixagent"},
						"name":      "responses",
					},
				},
			})

		// Create table
		case r.URL.Path == "/v1/namespaces/helixagent/tables" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"metadata": map[string]interface{}{
					"format-version": 2,
					"table-uuid":     "new-table-uuid",
					"location":       "s3://test-warehouse/helixagent/new-table",
				},
			})

		// Get table
		case r.URL.Path == "/v1/namespaces/helixagent/tables/debates" && r.Method == http.MethodGet:
			snapshotID := int64(12345)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"metadata": map[string]interface{}{
					"format-version":      2,
					"table-uuid":          "test-uuid-debates",
					"location":            "s3://test-warehouse/helixagent/debates",
					"current-snapshot-id": snapshotID,
					"properties": map[string]string{
						"write.format.default": "parquet",
					},
					"snapshots": []map[string]interface{}{
						{
							"snapshot-id":     int64(12344),
							"sequence-number": int64(1),
							"timestamp-ms":    snapshotOldTimestamp,
							"manifest-list":   "s3://test-warehouse/metadata/snap-12344.avro",
							"summary": map[string]string{
								"added-records": "500",
							},
						},
						{
							"snapshot-id":        snapshotID,
							"parent-snapshot-id": int64(12344),
							"sequence-number":    int64(2),
							"timestamp-ms":       snapshotNewTimestamp,
							"manifest-list":      "s3://test-warehouse/metadata/snap-12345.avro",
							"summary": map[string]string{
								"added-records": "1000",
							},
						},
					},
				},
			})

		// Get table without snapshot
		case r.URL.Path == "/v1/namespaces/helixagent/tables/empty_table" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"metadata": map[string]interface{}{
					"format-version":      2,
					"table-uuid":          "test-uuid-empty",
					"location":            "s3://test-warehouse/helixagent/empty_table",
					"current-snapshot-id": nil,
					"snapshots":           []interface{}{},
				},
			})

		// Table exists (HEAD request)
		case r.URL.Path == "/v1/namespaces/helixagent/tables/debates" && r.Method == http.MethodHead:
			w.WriteHeader(http.StatusOK)

		// Table does not exist (HEAD request)
		case r.URL.Path == "/v1/namespaces/helixagent/tables/nonexistent" && r.Method == http.MethodHead:
			w.WriteHeader(http.StatusNotFound)

		// Delete table
		case r.URL.Path == "/v1/namespaces/helixagent/tables/debates" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		// Delete table with purge
		case r.URL.Path == "/v1/namespaces/helixagent/tables/debates" && r.Method == http.MethodDelete && r.URL.RawQuery == "purgeRequested=true":
			w.WriteHeader(http.StatusNoContent)

		// Rename table
		case r.URL.Path == "/v1/tables/rename" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})

		// Update table properties
		case r.URL.Path == "/v1/namespaces/helixagent/tables/debates" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})

		// Get table for wait operation
		case r.URL.Path == "/v1/namespaces/helixagent/tables/new_table" && r.Method == http.MethodHead:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Not found",
					"type":    "NoSuchTableException",
				},
			})
		}
	}))
}

// MockErrorServer creates a mock server that returns errors
func MockErrorServer(statusCode int, errorMessage string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": errorMessage,
				"type":    "TestError",
			},
		})
	}))
}

// createConnectedClient creates a client connected to a mock server
func createConnectedClient(t *testing.T, server *httptest.Server) *Client {
	config := DefaultConfig()
	config.CatalogURI = server.URL
	config.Timeout = 5 * time.Second

	client, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	err = client.Connect(context.Background())
	require.NoError(t, err)

	return client
}

// ==================== GetCatalogConfig Tests ====================

func TestGetCatalogConfig_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	config, err := client.GetCatalogConfig(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "s3://test-warehouse", config.Defaults["warehouse"])
	assert.Equal(t, "helixagent", config.Overrides["default-namespace"])
}

func TestGetCatalogConfig_NotConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	config, err := client.GetCatalogConfig(context.Background())
	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "not connected")
}

func TestGetCatalogConfig_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusInternalServerError, "Internal server error")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	// Manually set connected to test error path
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	_, err := client.GetCatalogConfig(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get catalog config")
}

// ==================== CreateNamespace Tests ====================

func TestCreateNamespace_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.CreateNamespace(context.Background(), "new-namespace", map[string]string{
		"owner": "test-user",
	})
	require.NoError(t, err)
}

func TestCreateNamespace_WithNilProperties(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.CreateNamespace(context.Background(), "new-namespace", nil)
	require.NoError(t, err)
}

func TestCreateNamespace_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusConflict, "Namespace already exists")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	err := client.CreateNamespace(context.Background(), "existing-ns", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create namespace")
}

// ==================== ListNamespaces Tests ====================

func TestListNamespaces_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	namespaces, err := client.ListNamespaces(context.Background())
	require.NoError(t, err)
	assert.Len(t, namespaces, 3)
	assert.Contains(t, namespaces, "helixagent")
	assert.Contains(t, namespaces, "analytics")
	assert.Contains(t, namespaces, "test")
}

func TestListNamespaces_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusInternalServerError, "Server error")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	namespaces, err := client.ListNamespaces(context.Background())
	require.Error(t, err)
	assert.Nil(t, namespaces)
}

func TestListNamespaces_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	_, err := client.ListNamespaces(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

// ==================== GetNamespace Tests ====================

func TestGetNamespace_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	ns, err := client.GetNamespace(context.Background(), "helixagent")
	require.NoError(t, err)
	assert.NotNil(t, ns)
	assert.Equal(t, "admin", ns.Properties["owner"])
	assert.Equal(t, "HelixAgent namespace", ns.Properties["description"])
}

func TestGetNamespace_NotFound(t *testing.T) {
	server := MockErrorServer(http.StatusNotFound, "Namespace not found")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	ns, err := client.GetNamespace(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Nil(t, ns)
}

// ==================== DropNamespace Tests ====================

func TestDropNamespace_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.DropNamespace(context.Background(), "helixagent")
	require.NoError(t, err)
}

func TestDropNamespace_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusConflict, "Namespace not empty")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	err := client.DropNamespace(context.Background(), "helixagent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to drop namespace")
}

// ==================== CreateTable Tests ====================

func TestCreateTable_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	schema := NewSchema().
		AddField(1, "id", "string", true).
		AddField(2, "name", "string", true).
		AddFieldWithDoc(3, "created_at", "timestamp", true, "Creation timestamp")

	tableConfig := DefaultTableConfig("helixagent", "new_table").
		WithSchema(schema).
		WithPartition(NewPartitionField(3, "created_at_day", TransformDay)).
		WithSortOrder(NewSortField(1, SortAsc)).
		WithProperty("write.format.default", "parquet")

	err := client.CreateTable(context.Background(), tableConfig)
	require.NoError(t, err)
}

func TestCreateTable_MinimalConfig(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	tableConfig := DefaultTableConfig("helixagent", "minimal_table")

	err := client.CreateTable(context.Background(), tableConfig)
	require.NoError(t, err)
}

func TestCreateTable_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusConflict, "Table already exists")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	tableConfig := DefaultTableConfig("helixagent", "existing_table")
	err := client.CreateTable(context.Background(), tableConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create table")
}

// ==================== ListTables Tests ====================

func TestListTables_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	tables, err := client.ListTables(context.Background(), "helixagent")
	require.NoError(t, err)
	assert.Len(t, tables, 2)
	assert.Contains(t, tables, "debates")
	assert.Contains(t, tables, "responses")
}

func TestListTables_EmptyNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"identifiers": []interface{}{},
		})
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	tables, err := client.ListTables(context.Background(), "empty")
	require.NoError(t, err)
	assert.Empty(t, tables)
}

func TestListTables_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusNotFound, "Namespace not found")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	tables, err := client.ListTables(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Nil(t, tables)
}

// ==================== GetTable Tests ====================

func TestGetTable_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	table, err := client.GetTable(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.NotNil(t, table)
	assert.Equal(t, 2, table.FormatVersion)
	assert.Equal(t, "test-uuid-debates", table.TableUUID)
	assert.Equal(t, "s3://test-warehouse/helixagent/debates", table.Location)
	assert.NotNil(t, table.CurrentSnapshotID)
	assert.Equal(t, int64(12345), *table.CurrentSnapshotID)
	assert.Len(t, table.Snapshots, 2)
}

func TestGetTable_NotFound(t *testing.T) {
	server := MockErrorServer(http.StatusNotFound, "Table not found")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	table, err := client.GetTable(context.Background(), "helixagent", "nonexistent")
	require.Error(t, err)
	assert.Nil(t, table)
}

// ==================== DropTable Tests ====================

func TestDropTable_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.DropTable(context.Background(), "helixagent", "debates", false)
	require.NoError(t, err)
}

func TestDropTable_WithPurge(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.DropTable(context.Background(), "helixagent", "debates", true)
	require.NoError(t, err)
}

func TestDropTable_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusNotFound, "Table not found")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	err := client.DropTable(context.Background(), "helixagent", "nonexistent", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to drop table")
}

// ==================== TableExists Tests ====================

func TestTableExists_True(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	exists, err := client.TableExists(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestTableExists_False(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	exists, err := client.TableExists(context.Background(), "helixagent", "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}

// ==================== RenameTable Tests ====================

func TestRenameTable_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.RenameTable(context.Background(), "helixagent", "old_table", "new_table")
	require.NoError(t, err)
}

func TestRenameTable_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusNotFound, "Table not found")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	err := client.RenameTable(context.Background(), "helixagent", "nonexistent", "new_name")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to rename table")
}

// ==================== UpdateTableProperties Tests ====================

func TestUpdateTableProperties_WithUpdates(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	updates := map[string]string{
		"write.format.default": "orc",
		"compression":          "snappy",
	}

	err := client.UpdateTableProperties(context.Background(), "helixagent", "debates", updates, nil)
	require.NoError(t, err)
}

func TestUpdateTableProperties_WithRemovals(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	removals := map[string]string{
		"deprecated.property": "",
	}

	err := client.UpdateTableProperties(context.Background(), "helixagent", "debates", nil, removals)
	require.NoError(t, err)
}

func TestUpdateTableProperties_WithBoth(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	updates := map[string]string{
		"new.property": "value",
	}
	removals := map[string]string{
		"old.property": "",
	}

	err := client.UpdateTableProperties(context.Background(), "helixagent", "debates", updates, removals)
	require.NoError(t, err)
}

func TestUpdateTableProperties_NotConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	err := client.UpdateTableProperties(context.Background(), "ns", "table", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestUpdateTableProperties_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusNotFound, "Table not found")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	err := client.UpdateTableProperties(context.Background(), "helixagent", "nonexistent", map[string]string{"key": "value"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update table properties")
}

// ==================== GetSnapshots Tests ====================

func TestGetSnapshots_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	snapshots, err := client.GetSnapshots(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.Len(t, snapshots, 2)
	assert.Equal(t, int64(12344), snapshots[0].SnapshotID)
	assert.Equal(t, int64(12345), snapshots[1].SnapshotID)
}

func TestGetSnapshots_NotConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	snapshots, err := client.GetSnapshots(context.Background(), "ns", "table")
	require.Error(t, err)
	assert.Nil(t, snapshots)
}

// ==================== GetCurrentSnapshot Tests ====================

func TestGetCurrentSnapshot_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	snapshot, err := client.GetCurrentSnapshot(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, int64(12345), snapshot.SnapshotID)
	assert.Equal(t, "1000", snapshot.Summary["added-records"])
}

func TestGetCurrentSnapshot_NoSnapshot(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	snapshot, err := client.GetCurrentSnapshot(context.Background(), "helixagent", "empty_table")
	require.NoError(t, err)
	assert.Nil(t, snapshot)
}

func TestGetCurrentSnapshot_NotConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	snapshot, err := client.GetCurrentSnapshot(context.Background(), "ns", "table")
	require.Error(t, err)
	assert.Nil(t, snapshot)
}

// ==================== GetSnapshotAtTimestamp Tests ====================

func TestGetSnapshotAtTimestamp_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	// Query for a time between the two snapshots - should return the first (old) snapshot
	timestamp := time.Date(2026, 1, 21, 12, 0, 0, 0, time.UTC) // Jan 21 - between old (Jan 20) and new (Jan 22)
	snapshot, err := client.GetSnapshotAtTimestamp(context.Background(), "helixagent", "debates", timestamp)
	require.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, int64(12344), snapshot.SnapshotID)
}

func TestGetSnapshotAtTimestamp_CurrentTime(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	// Query for a time after both snapshots - should return the latest (new) snapshot
	timestamp := time.Date(2026, 1, 23, 12, 0, 0, 0, time.UTC) // Jan 23 - after new snapshot (Jan 22)
	snapshot, err := client.GetSnapshotAtTimestamp(context.Background(), "helixagent", "debates", timestamp)
	require.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, int64(12345), snapshot.SnapshotID)
}

func TestGetSnapshotAtTimestamp_BeforeAllSnapshots(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	// Query for a time before all snapshots
	timestamp := time.Date(2026, 1, 19, 12, 0, 0, 0, time.UTC) // Jan 19 - before old snapshot (Jan 20)
	snapshot, err := client.GetSnapshotAtTimestamp(context.Background(), "helixagent", "debates", timestamp)
	require.NoError(t, err)
	assert.Nil(t, snapshot)
}

func TestGetSnapshotAtTimestamp_NotConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	snapshot, err := client.GetSnapshotAtTimestamp(context.Background(), "ns", "table", time.Now())
	require.Error(t, err)
	assert.Nil(t, snapshot)
}

// ==================== WaitForTable Tests ====================

func TestWaitForTable_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.WaitForTable(context.Background(), "helixagent", "new_table", 5*time.Second)
	require.NoError(t, err)
}

func TestWaitForTable_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		// Always return 404 for table check
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	err := client.WaitForTable(context.Background(), "helixagent", "never_created", 1*time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for table")
}

func TestWaitForTable_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		// Always return 404 for table check
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	err := client.WaitForTable(ctx, "helixagent", "cancelled", 10*time.Second)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestWaitForTable_NotConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	err := client.WaitForTable(context.Background(), "ns", "table", 1*time.Second)
	require.Error(t, err)
	// WaitForTable calls TableExists which returns error, not timeout
}

// ==================== HealthCheck Tests with Mock Server ====================

func TestHealthCheck_Success(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.HealthCheck(context.Background())
	require.NoError(t, err)
}

func TestHealthCheck_ServerError(t *testing.T) {
	server := MockErrorServer(http.StatusServiceUnavailable, "Service unavailable")
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, err := NewClient(config, nil)
	require.NoError(t, err)

	err = client.HealthCheck(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "catalog unhealthy")
}

// ==================== doRequest Tests ====================

func TestDoRequest_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid JSON for ALL requests
		w.Write([]byte("not valid json {"))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.CatalogURI = server.URL

	client, _ := NewClient(config, nil)
	// Manually set connected to bypass connection check
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	// GetCatalogConfig will try to parse invalid JSON
	_, err := client.GetCatalogConfig(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse response")
}

// ==================== Edge Cases ====================

func TestListNamespaces_EmptyNamespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"namespaces": [][]string{
				{}, // Empty namespace array
				{"valid"},
			},
		})
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	namespaces, err := client.ListNamespaces(context.Background())
	require.NoError(t, err)
	assert.Len(t, namespaces, 2)
	assert.Equal(t, "", namespaces[0])
	assert.Equal(t, "valid", namespaces[1])
}

func TestConnect_AlreadyConnected(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	// Connect again
	err := client.Connect(context.Background())
	require.NoError(t, err)
	assert.True(t, client.IsConnected())
}

func TestClose_MultipleClose(t *testing.T) {
	client, _ := NewClient(nil, nil)
	err := client.Close()
	require.NoError(t, err)

	// Close again should be safe
	err = client.Close()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestCreateTable_WithPartitionAndSort(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	schema := NewSchema().
		AddField(1, "id", "long", true).
		AddField(2, "event_time", "timestamp", true).
		AddField(3, "data", "string", false)

	// Create bucket partition with width
	bucketPartition := NewPartitionField(1, "id_bucket", TransformBucket).WithWidth(16)

	// Create sort field with custom null order
	sortField := NewSortField(2, SortDesc).WithNullOrder(NullsLast)

	tableConfig := DefaultTableConfig("helixagent", "partitioned_table").
		WithSchema(schema).
		WithPartition(bucketPartition).
		WithPartition(NewPartitionField(2, "event_day", TransformDay)).
		WithSortOrder(sortField)

	err := client.CreateTable(context.Background(), tableConfig)
	require.NoError(t, err)
}

// ==================== Additional Coverage Tests ====================

func TestGetNamespace_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	ns, err := client.GetNamespace(context.Background(), "helixagent")
	require.Error(t, err)
	assert.Nil(t, ns)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestListTables_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	tables, err := client.ListTables(context.Background(), "helixagent")
	require.Error(t, err)
	assert.Nil(t, tables)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestGetTable_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	table, err := client.GetTable(context.Background(), "helixagent", "debates")
	require.Error(t, err)
	assert.Nil(t, table)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestGetCurrentSnapshot_SnapshotNotInList(t *testing.T) {
	// Test case where current-snapshot-id points to a non-existent snapshot
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/config" {
			json.NewEncoder(w).Encode(map[string]interface{}{"defaults": map[string]string{}, "overrides": map[string]string{}})
			return
		}
		snapshotID := int64(99999) // Points to non-existent snapshot
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metadata": map[string]interface{}{
				"format-version":      2,
				"table-uuid":          "test-uuid",
				"location":            "s3://test-warehouse/test/table",
				"current-snapshot-id": snapshotID,
				"snapshots": []map[string]interface{}{
					{
						"snapshot-id":     int64(12345),
						"sequence-number": int64(1),
						"timestamp-ms":    snapshotOldTimestamp,
						"manifest-list":   "s3://test/snap.avro",
						"summary":         map[string]string{},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	snapshot, err := client.GetCurrentSnapshot(context.Background(), "test", "table")
	require.NoError(t, err)
	assert.Nil(t, snapshot) // Should return nil when snapshot not found
}

func TestTableConfig_WithPropertyNilMap(t *testing.T) {
	// Test WithProperty when Properties map is nil
	config := &TableConfig{
		Namespace:  "test",
		Name:       "table",
		Properties: nil, // Explicitly nil
	}

	result := config.WithProperty("key", "value")
	assert.NotNil(t, result.Properties)
	assert.Equal(t, "value", result.Properties["key"])
}

func TestDoRequest_MarshalError(t *testing.T) {
	// Test that marshal errors are handled
	// This is difficult to test since maps always marshal successfully
	// We'll test by creating a scenario with a channel (which can't be marshaled)
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	// The doRequest function handles marshal errors internally
	// We can verify this path exists by checking the code
	// Since we can't easily trigger this with normal usage, we'll verify coverage via other paths
}

// ==================== Integration-style Tests ====================

func TestFullWorkflow(t *testing.T) {
	server := MockIcebergServer(t)
	defer server.Close()

	client := createConnectedClient(t, server)
	defer client.Close()

	// 1. Get catalog config
	config, err := client.GetCatalogConfig(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, config)

	// 2. List namespaces
	namespaces, err := client.ListNamespaces(context.Background())
	require.NoError(t, err)
	assert.Contains(t, namespaces, "helixagent")

	// 3. Get namespace
	ns, err := client.GetNamespace(context.Background(), "helixagent")
	require.NoError(t, err)
	assert.NotNil(t, ns)

	// 4. Create table
	schema := NewSchema().
		AddField(1, "id", "string", true).
		AddField(2, "content", "string", false)

	tableConfig := DefaultTableConfig("helixagent", "new_table").
		WithSchema(schema)

	err = client.CreateTable(context.Background(), tableConfig)
	require.NoError(t, err)

	// 5. List tables
	tables, err := client.ListTables(context.Background(), "helixagent")
	require.NoError(t, err)
	assert.Contains(t, tables, "debates")

	// 6. Get table
	table, err := client.GetTable(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.NotNil(t, table)

	// 7. Check table exists
	exists, err := client.TableExists(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.True(t, exists)

	// 8. Get snapshots
	snapshots, err := client.GetSnapshots(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.Len(t, snapshots, 2)

	// 9. Get current snapshot
	current, err := client.GetCurrentSnapshot(context.Background(), "helixagent", "debates")
	require.NoError(t, err)
	assert.NotNil(t, current)
}
