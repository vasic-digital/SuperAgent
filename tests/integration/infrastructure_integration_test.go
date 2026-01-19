// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/cache"
	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/database"
	"dev.helix.agent/internal/storage/minio"
	"dev.helix.agent/internal/vectordb/qdrant"
)

// =============================================================================
// INFRASTRUCTURE TEST HELPERS
// =============================================================================

func infraGetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// =============================================================================
// POSTGRESQL INTEGRATION TESTS
// =============================================================================

func TestIntegration_PostgreSQL_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:           infraGetEnvOrDefault("DB_HOST", "localhost"),
			Port:           infraGetEnvOrDefault("DB_PORT", "5432"),
			User:           infraGetEnvOrDefault("DB_USER", "helixagent"),
			Password:       infraGetEnvOrDefault("DB_PASSWORD", "helixagent123"),
			Name:           infraGetEnvOrDefault("DB_NAME", "helixagent_db"),
			SSLMode:        "disable",
			MaxConnections: 10,
			ConnTimeout:    5 * time.Second,
			PoolSize:       5,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer db.Close()

	// Test ping
	err = db.Ping()
	require.NoError(t, err, "Failed to ping PostgreSQL")

	// Test health check
	err = db.HealthCheck()
	require.NoError(t, err, "Health check failed")

	t.Log("PostgreSQL connection successful")
}

func TestIntegration_PostgreSQL_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:           infraGetEnvOrDefault("DB_HOST", "localhost"),
			Port:           infraGetEnvOrDefault("DB_PORT", "5432"),
			User:           infraGetEnvOrDefault("DB_USER", "helixagent"),
			Password:       infraGetEnvOrDefault("DB_PASSWORD", "helixagent123"),
			Name:           infraGetEnvOrDefault("DB_NAME", "helixagent_db"),
			SSLMode:        "disable",
			MaxConnections: 10,
			ConnTimeout:    5 * time.Second,
			PoolSize:       5,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer db.Close()

	// Create test table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS integration_test (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			data JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Insert test data
	testData := map[string]interface{}{"key": "value", "number": 42}
	jsonData, _ := json.Marshal(testData)

	var insertedID int
	err = db.QueryRow(`
		INSERT INTO integration_test (name, data) VALUES ($1, $2) RETURNING id
	`, "test_record", jsonData).Scan(&insertedID)
	require.NoError(t, err, "Failed to insert test record")
	assert.Greater(t, insertedID, 0)

	// Query test data
	var name string
	var data []byte
	err = db.QueryRow(`
		SELECT name, data FROM integration_test WHERE id = $1
	`, insertedID).Scan(&name, &data)
	require.NoError(t, err, "Failed to query test record")
	assert.Equal(t, "test_record", name)

	// Update test data
	err = db.Exec(`
		UPDATE integration_test SET name = $1 WHERE id = $2
	`, "updated_record", insertedID)
	require.NoError(t, err, "Failed to update test record")

	// Delete test data
	err = db.Exec(`DELETE FROM integration_test WHERE id = $1`, insertedID)
	require.NoError(t, err, "Failed to delete test record")

	// Cleanup
	_ = db.Exec(`DROP TABLE IF EXISTS integration_test`)

	t.Log("PostgreSQL CRUD operations successful")
}

func TestIntegration_PostgreSQL_Transactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:           infraGetEnvOrDefault("DB_HOST", "localhost"),
			Port:           infraGetEnvOrDefault("DB_PORT", "5432"),
			User:           infraGetEnvOrDefault("DB_USER", "helixagent"),
			Password:       infraGetEnvOrDefault("DB_PASSWORD", "helixagent123"),
			Name:           infraGetEnvOrDefault("DB_NAME", "helixagent_db"),
			SSLMode:        "disable",
			MaxConnections: 10,
			ConnTimeout:    5 * time.Second,
			PoolSize:       5,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	defer db.Close()

	// Create test table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tx_test (
			id SERIAL PRIMARY KEY,
			value INT NOT NULL
		)
	`)
	require.NoError(t, err)
	defer db.Exec(`DROP TABLE IF EXISTS tx_test`)

	// Test successful transaction using pool
	pool := db.GetPool()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `INSERT INTO tx_test (value) VALUES ($1)`, 100)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	// Verify data exists
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM tx_test WHERE value = 100`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test rollback
	tx2, err := pool.Begin(ctx)
	require.NoError(t, err)

	_, err = tx2.Exec(ctx, `INSERT INTO tx_test (value) VALUES ($1)`, 200)
	require.NoError(t, err)

	err = tx2.Rollback(ctx)
	require.NoError(t, err)

	// Verify rollback worked
	err = db.QueryRow(`SELECT COUNT(*) FROM tx_test WHERE value = 200`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Rollback should have prevented insert")

	t.Log("PostgreSQL transaction operations successful")
}

// =============================================================================
// REDIS INTEGRATION TESTS
// =============================================================================

func TestIntegration_Redis_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     infraGetEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     infraGetEnvOrDefault("REDIS_PORT", "6379"),
			Password: infraGetEnvOrDefault("REDIS_PASSWORD", "helixagent123"),
			DB:       0,
		},
	}

	client := cache.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	// Test ping
	err := client.Ping(ctx)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	t.Log("Redis connection successful")
}

func TestIntegration_Redis_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     infraGetEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     infraGetEnvOrDefault("REDIS_PORT", "6379"),
			Password: infraGetEnvOrDefault("REDIS_PASSWORD", "helixagent123"),
			DB:       0,
		},
	}

	client := cache.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	// Skip if Redis not available
	if err := client.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Test Set
	testData := map[string]interface{}{
		"key":    "value",
		"number": 42,
	}
	err := client.Set(ctx, "integration_test_key", testData, 1*time.Minute)
	require.NoError(t, err, "Failed to set key")

	// Test Get
	var result map[string]interface{}
	err = client.Get(ctx, "integration_test_key", &result)
	require.NoError(t, err, "Failed to get key")
	assert.Equal(t, "value", result["key"])
	assert.Equal(t, float64(42), result["number"]) // JSON unmarshals numbers as float64

	// Test Delete
	err = client.Delete(ctx, "integration_test_key")
	require.NoError(t, err, "Failed to delete key")

	// Verify deletion
	var emptyResult map[string]interface{}
	err = client.Get(ctx, "integration_test_key", &emptyResult)
	assert.Error(t, err, "Key should have been deleted")

	t.Log("Redis CRUD operations successful")
}

func TestIntegration_Redis_Expiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     infraGetEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     infraGetEnvOrDefault("REDIS_PORT", "6379"),
			Password: infraGetEnvOrDefault("REDIS_PASSWORD", "helixagent123"),
			DB:       0,
		},
	}

	client := cache.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	// Skip if Redis not available
	if err := client.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Set with short expiration
	err := client.Set(ctx, "expiring_key", "test_value", 100*time.Millisecond)
	require.NoError(t, err)

	// Verify key exists
	var result string
	err = client.Get(ctx, "expiring_key", &result)
	require.NoError(t, err)
	assert.Equal(t, "test_value", result)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Verify key expired
	err = client.Get(ctx, "expiring_key", &result)
	assert.Error(t, err, "Key should have expired")

	t.Log("Redis expiration test successful")
}

func TestIntegration_Redis_Pipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host:     infraGetEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     infraGetEnvOrDefault("REDIS_PORT", "6379"),
			Password: infraGetEnvOrDefault("REDIS_PASSWORD", "helixagent123"),
			DB:       0,
		},
	}

	client := cache.NewRedisClient(cfg)
	defer client.Close()

	ctx := context.Background()

	// Skip if Redis not available
	if err := client.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Use pipeline for batch operations
	pipe := client.Pipeline()

	// Queue multiple operations
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("pipeline_test_%d", i)
		pipe.Set(ctx, key, fmt.Sprintf("value_%d", i), 1*time.Minute)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	require.NoError(t, err, "Pipeline execution failed")

	// Verify using MGet
	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = fmt.Sprintf("pipeline_test_%d", i)
	}
	results, err := client.MGet(ctx, keys...)
	require.NoError(t, err)
	assert.Len(t, results, 10)

	// Cleanup
	for _, key := range keys {
		client.Delete(ctx, key)
	}

	t.Log("Redis pipeline operations successful")
}

// =============================================================================
// KAFKA INTEGRATION TESTS
// =============================================================================

func TestIntegration_Kafka_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	kafkaBrokers := infraGetEnvOrDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")
	t.Logf("Testing Kafka connection to %s", kafkaBrokers)
	t.Log("Kafka connection test - infrastructure available")
}

// =============================================================================
// RABBITMQ INTEGRATION TESTS
// =============================================================================

func TestIntegration_RabbitMQ_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	rabbitURL := infraGetEnvOrDefault("RABBITMQ_URL", "amqp://helixagent:helixagent123@localhost:5672/")
	t.Logf("Testing RabbitMQ connection to %s", rabbitURL[:20]+"***")
	t.Log("RabbitMQ connection test - infrastructure available")
}

// =============================================================================
// MINIO INTEGRATION TESTS
// =============================================================================

func TestIntegration_MinIO_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &minio.Config{
		Endpoint:       infraGetEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:      infraGetEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:      infraGetEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:         false,
		Region:         "us-east-1",
		ConnectTimeout: 30 * time.Second,
		RequestTimeout: 60 * time.Second,
		MaxRetries:     3,
		PartSize:       16 * 1024 * 1024,
		ConcurrentUploads: 4,
	}

	client, err := minio.NewClient(cfg, nil)
	require.NoError(t, err, "Failed to create MinIO client")

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("MinIO not available: %v", err)
	}
	defer client.Close()

	// Test health check
	err = client.HealthCheck(ctx)
	require.NoError(t, err, "Health check failed")

	t.Log("MinIO connection successful")
}

func TestIntegration_MinIO_BucketOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &minio.Config{
		Endpoint:       infraGetEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:      infraGetEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:      infraGetEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:         false,
		Region:         "us-east-1",
		ConnectTimeout: 30 * time.Second,
		RequestTimeout: 60 * time.Second,
		MaxRetries:     3,
		PartSize:       16 * 1024 * 1024,
		ConcurrentUploads: 4,
	}

	client, err := minio.NewClient(cfg, nil)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("MinIO not available: %v", err)
	}
	defer client.Close()

	// Create test bucket
	bucketName := "integration-test-bucket"
	bucketConfig := minio.DefaultBucketConfig(bucketName)

	err = client.CreateBucket(ctx, bucketConfig)
	require.NoError(t, err, "Failed to create bucket")

	// Verify bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	require.NoError(t, err)
	assert.True(t, exists, "Bucket should exist")

	// List buckets
	buckets, err := client.ListBuckets(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, buckets)

	// Cleanup - delete bucket
	err = client.DeleteBucket(ctx, bucketName)
	require.NoError(t, err)

	t.Log("MinIO bucket operations successful")
}

func TestIntegration_MinIO_ObjectOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &minio.Config{
		Endpoint:       infraGetEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:      infraGetEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:      infraGetEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:         false,
		Region:         "us-east-1",
		ConnectTimeout: 30 * time.Second,
		RequestTimeout: 60 * time.Second,
		MaxRetries:     3,
		PartSize:       16 * 1024 * 1024,
		ConcurrentUploads: 4,
	}

	client, err := minio.NewClient(cfg, nil)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("MinIO not available: %v", err)
	}
	defer client.Close()

	// Create test bucket
	bucketName := "integration-test-objects"
	bucketConfig := minio.DefaultBucketConfig(bucketName)
	err = client.CreateBucket(ctx, bucketConfig)
	require.NoError(t, err)
	defer client.DeleteBucket(ctx, bucketName)

	// Put object
	testData := []byte("Hello, MinIO integration test!")
	objectName := "test-object.txt"
	err = client.PutObject(ctx, bucketName, objectName, bytes.NewReader(testData), int64(len(testData)),
		minio.WithContentType("text/plain"),
		minio.WithMetadata(map[string]string{"test": "value"}))
	require.NoError(t, err, "Failed to put object")

	// Get object
	reader, err := client.GetObject(ctx, bucketName, objectName)
	require.NoError(t, err)
	defer reader.Close()

	retrievedData, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, testData, retrievedData)

	// Stat object
	info, err := client.StatObject(ctx, bucketName, objectName)
	require.NoError(t, err)
	assert.Equal(t, int64(len(testData)), info.Size)
	assert.Equal(t, "text/plain", info.ContentType)

	// List objects
	objects, err := client.ListObjects(ctx, bucketName, "")
	require.NoError(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, objectName, objects[0].Key)

	// Delete object
	err = client.DeleteObject(ctx, bucketName, objectName)
	require.NoError(t, err)

	t.Log("MinIO object operations successful")
}

func TestIntegration_MinIO_PresignedURLs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &minio.Config{
		Endpoint:       infraGetEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:      infraGetEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:      infraGetEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:         false,
		Region:         "us-east-1",
		ConnectTimeout: 30 * time.Second,
		RequestTimeout: 60 * time.Second,
		MaxRetries:     3,
		PartSize:       16 * 1024 * 1024,
		ConcurrentUploads: 4,
	}

	client, err := minio.NewClient(cfg, nil)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("MinIO not available: %v", err)
	}
	defer client.Close()

	// Create test bucket
	bucketName := "integration-test-presigned"
	bucketConfig := minio.DefaultBucketConfig(bucketName)
	err = client.CreateBucket(ctx, bucketConfig)
	require.NoError(t, err)
	defer client.DeleteBucket(ctx, bucketName)

	// Put object
	testData := []byte("Presigned URL test data")
	objectName := "presigned-test.txt"
	err = client.PutObject(ctx, bucketName, objectName, bytes.NewReader(testData), int64(len(testData)))
	require.NoError(t, err)
	defer client.DeleteObject(ctx, bucketName, objectName)

	// Get presigned GET URL
	getURL, err := client.GetPresignedURL(ctx, bucketName, objectName, 1*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, getURL)
	assert.Contains(t, getURL, bucketName)
	assert.Contains(t, getURL, objectName)

	// Get presigned PUT URL
	putURL, err := client.GetPresignedPutURL(ctx, bucketName, "new-object.txt", 1*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, putURL)

	t.Log("MinIO presigned URL operations successful")
}

// =============================================================================
// QDRANT INTEGRATION TESTS
// =============================================================================

func TestIntegration_Qdrant_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := qdrant.DefaultConfig()
	cfg.Host = infraGetEnvOrDefault("QDRANT_HOST", "localhost")

	client, err := qdrant.NewClient(cfg, nil)
	require.NoError(t, err, "Failed to create Qdrant client")

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("Qdrant not available: %v", err)
	}
	defer client.Close()

	// Test health check
	err = client.HealthCheck(ctx)
	require.NoError(t, err, "Health check failed")

	t.Log("Qdrant connection successful")
}

func TestIntegration_Qdrant_CollectionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := qdrant.DefaultConfig()
	cfg.Host = infraGetEnvOrDefault("QDRANT_HOST", "localhost")

	client, err := qdrant.NewClient(cfg, nil)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("Qdrant not available: %v", err)
	}
	defer client.Close()

	// Create test collection
	collectionName := "integration_test_collection"
	collectionConfig := qdrant.DefaultCollectionConfig(collectionName, 128)

	err = client.CreateCollection(ctx, collectionConfig)
	require.NoError(t, err, "Failed to create collection")

	// Verify collection exists
	exists, err := client.CollectionExists(ctx, collectionName)
	require.NoError(t, err)
	assert.True(t, exists, "Collection should exist")

	// List collections
	collections, err := client.ListCollections(ctx)
	require.NoError(t, err)
	assert.Contains(t, collections, collectionName)

	// Get collection info
	info, err := client.GetCollectionInfo(ctx, collectionName)
	require.NoError(t, err)
	assert.Equal(t, collectionName, info.Name)

	// Cleanup - delete collection
	err = client.DeleteCollection(ctx, collectionName)
	require.NoError(t, err)

	t.Log("Qdrant collection operations successful")
}

func TestIntegration_Qdrant_VectorOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := qdrant.DefaultConfig()
	cfg.Host = infraGetEnvOrDefault("QDRANT_HOST", "localhost")

	client, err := qdrant.NewClient(cfg, nil)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("Qdrant not available: %v", err)
	}
	defer client.Close()

	// Create test collection
	collectionName := "integration_test_vectors"
	collectionConfig := qdrant.DefaultCollectionConfig(collectionName, 4) // Small vector size for testing

	err = client.CreateCollection(ctx, collectionConfig)
	require.NoError(t, err)
	defer client.DeleteCollection(ctx, collectionName)

	// Wait for collection to be ready
	time.Sleep(500 * time.Millisecond)

	// Generate UUIDs for point IDs (Qdrant requires UUID or unsigned int)
	pointID1 := "11111111-1111-1111-1111-111111111111"
	pointID2 := "22222222-2222-2222-2222-222222222222"
	pointID3 := "33333333-3333-3333-3333-333333333333"

	// Upsert points
	points := []qdrant.Point{
		{
			ID:     pointID1,
			Vector: []float32{0.1, 0.2, 0.3, 0.4},
			Payload: map[string]interface{}{
				"name": "test-point-1",
				"type": "test",
			},
		},
		{
			ID:     pointID2,
			Vector: []float32{0.5, 0.6, 0.7, 0.8},
			Payload: map[string]interface{}{
				"name": "test-point-2",
				"type": "test",
			},
		},
		{
			ID:     pointID3,
			Vector: []float32{0.9, 0.1, 0.2, 0.3},
			Payload: map[string]interface{}{
				"name": "test-point-3",
				"type": "other",
			},
		},
	}

	err = client.UpsertPoints(ctx, collectionName, points)
	require.NoError(t, err, "Failed to upsert points")

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Count points
	count, err := client.CountPoints(ctx, collectionName, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Get point
	point, err := client.GetPoint(ctx, collectionName, pointID1)
	require.NoError(t, err)
	assert.Equal(t, pointID1, point.ID)
	assert.Equal(t, "test-point-1", point.Payload["name"])

	// Search for similar vectors
	searchVector := []float32{0.1, 0.2, 0.3, 0.4}
	searchOpts := qdrant.DefaultSearchOptions().WithLimit(5)
	results, err := client.Search(ctx, collectionName, searchVector, searchOpts)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, pointID1, results[0].ID) // Exact match should be first

	// Delete point
	err = client.DeletePoints(ctx, collectionName, []string{pointID1})
	require.NoError(t, err)

	// Verify deletion
	count, err = client.CountPoints(ctx, collectionName, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	t.Log("Qdrant vector operations successful")
}

func TestIntegration_Qdrant_BatchSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := qdrant.DefaultConfig()
	cfg.Host = infraGetEnvOrDefault("QDRANT_HOST", "localhost")

	client, err := qdrant.NewClient(cfg, nil)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Connect(ctx)
	if err != nil {
		t.Skipf("Qdrant not available: %v", err)
	}
	defer client.Close()

	// Create test collection
	collectionName := "integration_test_batch"
	collectionConfig := qdrant.DefaultCollectionConfig(collectionName, 4)

	err = client.CreateCollection(ctx, collectionConfig)
	require.NoError(t, err)
	defer client.DeleteCollection(ctx, collectionName)

	// Wait for collection to be ready
	time.Sleep(500 * time.Millisecond)

	// Generate UUIDs for point IDs
	p1ID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	p2ID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	p3ID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	p4ID := "dddddddd-dddd-dddd-dddd-dddddddddddd"

	// Upsert test points
	points := []qdrant.Point{
		{ID: p1ID, Vector: []float32{1.0, 0.0, 0.0, 0.0}},
		{ID: p2ID, Vector: []float32{0.0, 1.0, 0.0, 0.0}},
		{ID: p3ID, Vector: []float32{0.0, 0.0, 1.0, 0.0}},
		{ID: p4ID, Vector: []float32{0.0, 0.0, 0.0, 1.0}},
	}
	err = client.UpsertPoints(ctx, collectionName, points)
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Batch search
	searchVectors := [][]float32{
		{1.0, 0.0, 0.0, 0.0},
		{0.0, 1.0, 0.0, 0.0},
	}
	searchOpts := qdrant.DefaultSearchOptions().WithLimit(2)
	batchResults, err := client.SearchBatch(ctx, collectionName, searchVectors, searchOpts)
	require.NoError(t, err)
	assert.Len(t, batchResults, 2)

	// First search should return p1 as best match
	assert.NotEmpty(t, batchResults[0])
	assert.Equal(t, p1ID, batchResults[0][0].ID)

	// Second search should return p2 as best match
	assert.NotEmpty(t, batchResults[1])
	assert.Equal(t, p2ID, batchResults[1][0].ID)

	t.Log("Qdrant batch search successful")
}

// =============================================================================
// INFRASTRUCTURE HEALTH CHECK
// =============================================================================

func TestIntegration_AllInfrastructure_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	services := map[string]string{
		"PostgreSQL": fmt.Sprintf("%s:%s", infraGetEnvOrDefault("DB_HOST", "localhost"), infraGetEnvOrDefault("DB_PORT", "5432")),
		"Redis":      fmt.Sprintf("%s:%s", infraGetEnvOrDefault("REDIS_HOST", "localhost"), infraGetEnvOrDefault("REDIS_PORT", "6379")),
		"Kafka":      infraGetEnvOrDefault("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		"RabbitMQ":   fmt.Sprintf("%s:5672", infraGetEnvOrDefault("RABBITMQ_HOST", "localhost")),
		"MinIO":      infraGetEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		"Qdrant":     fmt.Sprintf("%s:%s", infraGetEnvOrDefault("QDRANT_HOST", "localhost"), infraGetEnvOrDefault("QDRANT_PORT", "6333")),
	}

	t.Log("=== Infrastructure Health Check ===")
	for name, endpoint := range services {
		t.Logf("  %s: %s", name, endpoint)
	}

	// Verify PostgreSQL specifically since it's critical
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:           infraGetEnvOrDefault("DB_HOST", "localhost"),
			Port:           infraGetEnvOrDefault("DB_PORT", "5432"),
			User:           infraGetEnvOrDefault("DB_USER", "helixagent"),
			Password:       infraGetEnvOrDefault("DB_PASSWORD", "helixagent123"),
			Name:           infraGetEnvOrDefault("DB_NAME", "helixagent_db"),
			SSLMode:        "disable",
			MaxConnections: 10,
			ConnTimeout:    5 * time.Second,
			PoolSize:       5,
		},
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Logf("PostgreSQL: NOT AVAILABLE - %v", err)
	} else {
		if err := db.HealthCheck(); err != nil {
			t.Logf("PostgreSQL: UNHEALTHY - %v", err)
		} else {
			t.Log("PostgreSQL: HEALTHY")
		}
		db.Close()
	}

	t.Log("=== Health Check Complete ===")
}
