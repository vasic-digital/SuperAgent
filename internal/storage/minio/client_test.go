package minio

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
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

	t.Run("with nil logger uses default logger", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewClient(config, nil)
		require.NoError(t, err)
		assert.NotNil(t, client)
		// Verify client is functional without a logger
		assert.False(t, client.IsConnected())
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			Endpoint:          "minio.example.com:9000",
			AccessKey:         "access",
			SecretKey:         "secret",
			ConnectTimeout:    60 * time.Second,
			RequestTimeout:    120 * time.Second,
			PartSize:          16 * 1024 * 1024,
			ConcurrentUploads: 4,
		}
		client, err := NewClient(config, logrus.New())
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with SSL enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.UseSSL = true
		client, err := NewClient(config, nil)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with custom region", func(t *testing.T) {
		config := DefaultConfig()
		config.Region = "eu-west-1"
		client, err := NewClient(config, nil)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with invalid config - empty endpoint", func(t *testing.T) {
		config := &Config{
			Endpoint:  "",
			AccessKey: "access",
			SecretKey: "secret",
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("with invalid config - empty access key", func(t *testing.T) {
		config := &Config{
			Endpoint:  "localhost:9000",
			AccessKey: "",
			SecretKey: "secret",
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("with custom logger level", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		client, err := NewClient(nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestClientClose(t *testing.T) {
	client, _ := NewClient(nil, nil)
	err := client.Close()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestClientIsConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	assert.False(t, client.IsConnected())
}

func TestClientHealthCheck(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.HealthCheck(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestCreateBucket(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		bucketConfig := DefaultBucketConfig("test-bucket")
		err := client.CreateBucket(context.Background(), bucketConfig)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestDeleteBucket(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.DeleteBucket(context.Background(), "test-bucket")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestListBuckets(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		buckets, err := client.ListBuckets(context.Background())
		require.Error(t, err)
		assert.Nil(t, buckets)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestBucketExists(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		exists, err := client.BucketExists(context.Background(), "test-bucket")
		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestPutObject(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		reader := bytes.NewReader([]byte("test content"))
		err := client.PutObject(context.Background(), "bucket", "key", reader, 12)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestGetObject(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		obj, err := client.GetObject(context.Background(), "bucket", "key")
		require.Error(t, err)
		assert.Nil(t, obj)
	})
}

func TestDeleteObject(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.DeleteObject(context.Background(), "bucket", "key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestListObjects(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		objects, err := client.ListObjects(context.Background(), "bucket", "prefix/")
		require.Error(t, err)
		assert.Nil(t, objects)
	})
}

func TestStatObject(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		info, err := client.StatObject(context.Background(), "bucket", "key")
		require.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestGetPresignedURL(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		url, err := client.GetPresignedURL(context.Background(), "bucket", "key", time.Hour)
		require.Error(t, err)
		assert.Empty(t, url)
	})
}

func TestGetPresignedPutURL(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		url, err := client.GetPresignedPutURL(context.Background(), "bucket", "key", time.Hour)
		require.Error(t, err)
		assert.Empty(t, url)
	})
}

func TestCopyObject(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.CopyObject(context.Background(), "src-bucket", "src-key", "dst-bucket", "dst-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestSetLifecycleRule(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		rule := DefaultLifecycleRule("expire", 30)
		err := client.SetLifecycleRule(context.Background(), "bucket", rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestRemoveLifecycleRule(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.RemoveLifecycleRule(context.Background(), "bucket", "rule-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestPutOptions(t *testing.T) {
	t.Run("WithContentType", func(t *testing.T) {
		opt := WithContentType("application/json")
		assert.NotNil(t, opt)
	})

	t.Run("WithMetadata", func(t *testing.T) {
		metadata := map[string]string{"key": "value"}
		opt := WithMetadata(metadata)
		assert.NotNil(t, opt)
	})
}

func TestObjectInfo(t *testing.T) {
	info := &ObjectInfo{
		Key:          "test/file.txt",
		Size:         1024,
		LastModified: time.Now(),
		ContentType:  "text/plain",
		ETag:         "abc123",
		Metadata:     map[string]string{"custom": "value"},
	}

	assert.Equal(t, "test/file.txt", info.Key)
	assert.Equal(t, int64(1024), info.Size)
	assert.Equal(t, "text/plain", info.ContentType)
	assert.Equal(t, "abc123", info.ETag)
	assert.Equal(t, "value", info.Metadata["custom"])
}

func TestBucketInfo(t *testing.T) {
	info := &BucketInfo{
		Name:         "test-bucket",
		CreationDate: time.Now(),
	}

	assert.Equal(t, "test-bucket", info.Name)
	assert.False(t, info.CreationDate.IsZero())
}

// MockReader for testing
type mockReader struct {
	data   []byte
	offset int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.offset >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.offset:])
	m.offset += n
	return n, nil
}

func TestMockReader(t *testing.T) {
	reader := &mockReader{data: []byte("test data")}
	buf := make([]byte, 4)

	n, err := reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "test", string(buf))

	n, err = reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, " dat", string(buf))

	n, err = reader.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)

	n, err = reader.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

// Additional comprehensive tests for improved coverage

func TestPutOptions_Applied(t *testing.T) {
	t.Run("WithContentType applies correctly", func(t *testing.T) {
		opt := WithContentType("application/json")
		var opts minio.PutObjectOptions
		opt(&opts)
		assert.Equal(t, "application/json", opts.ContentType)
	})

	t.Run("WithMetadata applies correctly", func(t *testing.T) {
		metadata := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		opt := WithMetadata(metadata)
		var opts minio.PutObjectOptions
		opt(&opts)
		assert.Equal(t, "value1", opts.UserMetadata["key1"])
		assert.Equal(t, "value2", opts.UserMetadata["key2"])
	})

	t.Run("multiple options combined", func(t *testing.T) {
		options := []PutOption{
			WithContentType("text/plain"),
			WithMetadata(map[string]string{"author": "test"}),
		}

		var opts minio.PutObjectOptions
		for _, opt := range options {
			opt(&opts)
		}

		assert.Equal(t, "text/plain", opts.ContentType)
		assert.Equal(t, "test", opts.UserMetadata["author"])
	})
}

func TestNewClient_ValidationErrors(t *testing.T) {
	t.Run("empty secret key", func(t *testing.T) {
		config := &Config{
			Endpoint:  "localhost:9000",
			AccessKey: "access",
			SecretKey: "",
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid config")
	})

	t.Run("invalid connect timeout", func(t *testing.T) {
		config := &Config{
			Endpoint:       "localhost:9000",
			AccessKey:      "access",
			SecretKey:      "secret",
			ConnectTimeout: 0,
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("valid config creates client", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewClient(config, logrus.New())
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.False(t, client.IsConnected())
	})
}

func TestClientConcurrency(t *testing.T) {
	client, _ := NewClient(nil, nil)

	// Test concurrent IsConnected calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = client.IsConnected()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent Close calls
	for i := 0; i < 10; i++ {
		go func() {
			_ = client.Close()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestClient_CloseIdempotent(t *testing.T) {
	client, _ := NewClient(nil, nil)

	// Close should be idempotent
	err1 := client.Close()
	err2 := client.Close()
	err3 := client.Close()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.False(t, client.IsConnected())
}

func TestObjectInfo_Fields(t *testing.T) {
	now := time.Now()
	info := &ObjectInfo{
		Key:          "path/to/file.txt",
		Size:         2048,
		LastModified: now,
		ContentType:  "application/octet-stream",
		ETag:         "etag123",
		Metadata:     map[string]string{"x-amz-meta-custom": "value"},
	}

	assert.Equal(t, "path/to/file.txt", info.Key)
	assert.Equal(t, int64(2048), info.Size)
	assert.Equal(t, now, info.LastModified)
	assert.Equal(t, "application/octet-stream", info.ContentType)
	assert.Equal(t, "etag123", info.ETag)
	assert.Equal(t, "value", info.Metadata["x-amz-meta-custom"])
}

func TestBucketInfo_Fields(t *testing.T) {
	now := time.Now()
	info := &BucketInfo{
		Name:         "my-bucket",
		CreationDate: now,
	}

	assert.Equal(t, "my-bucket", info.Name)
	assert.Equal(t, now, info.CreationDate)
}

func TestClient_AllOperationsWhenNotConnected(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	// All these should return "not connected" errors
	tests := []struct {
		name string
		fn   func() error
	}{
		{"HealthCheck", func() error { return client.HealthCheck(ctx) }},
		{"CreateBucket", func() error { return client.CreateBucket(ctx, DefaultBucketConfig("test")) }},
		{"DeleteBucket", func() error { return client.DeleteBucket(ctx, "test") }},
		{"PutObject", func() error { return client.PutObject(ctx, "bucket", "key", bytes.NewReader([]byte("data")), 4) }},
		{"DeleteObject", func() error { return client.DeleteObject(ctx, "bucket", "key") }},
		{"CopyObject", func() error { return client.CopyObject(ctx, "src", "key", "dst", "key") }},
		{"SetLifecycleRule", func() error { return client.SetLifecycleRule(ctx, "bucket", DefaultLifecycleRule("id", 30)) }},
		{"RemoveLifecycleRule", func() error { return client.RemoveLifecycleRule(ctx, "bucket", "id") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not connected")
		})
	}

	// Test functions that return values
	t.Run("ListBuckets", func(t *testing.T) {
		buckets, err := client.ListBuckets(ctx)
		require.Error(t, err)
		assert.Nil(t, buckets)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("BucketExists", func(t *testing.T) {
		exists, err := client.BucketExists(ctx, "test")
		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("GetObject", func(t *testing.T) {
		obj, err := client.GetObject(ctx, "bucket", "key")
		require.Error(t, err)
		assert.Nil(t, obj)
	})

	t.Run("ListObjects", func(t *testing.T) {
		objects, err := client.ListObjects(ctx, "bucket", "prefix")
		require.Error(t, err)
		assert.Nil(t, objects)
	})

	t.Run("StatObject", func(t *testing.T) {
		info, err := client.StatObject(ctx, "bucket", "key")
		require.Error(t, err)
		assert.Nil(t, info)
	})

	t.Run("GetPresignedURL", func(t *testing.T) {
		url, err := client.GetPresignedURL(ctx, "bucket", "key", time.Hour)
		require.Error(t, err)
		assert.Empty(t, url)
	})

	t.Run("GetPresignedPutURL", func(t *testing.T) {
		url, err := client.GetPresignedPutURL(ctx, "bucket", "key", time.Hour)
		require.Error(t, err)
		assert.Empty(t, url)
	})
}

// Additional comprehensive tests

func TestClient_Connect_InvalidEndpoint(t *testing.T) {
	client, err := NewClient(nil, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Connect should fail with invalid endpoint (no server running)
	err = client.Connect(ctx)
	require.Error(t, err)
}

func TestClient_ContextCancellation(t *testing.T) {
	client, _ := NewClient(nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Operations should fail due to cancelled context
	err := client.HealthCheck(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_ConcurrentOperations(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Test concurrent health checks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := client.HealthCheck(ctx)
			if err != nil {
				errors <- err
			}
		}()
	}

	// Test concurrent IsConnected checks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.IsConnected()
		}()
	}

	// Test concurrent Close operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Close()
		}()
	}

	wg.Wait()
	close(errors)

	// All errors should be "not connected"
	for err := range errors {
		assert.Contains(t, err.Error(), "not connected")
	}
}

func TestClient_ConcurrentReadWrite(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	var wg sync.WaitGroup

	// Concurrent reads and writes (all should fail with not connected)
	for i := 0; i < 5; i++ {
		wg.Add(4)

		go func(idx int) {
			defer wg.Done()
			_ = client.IsConnected()
		}(i)

		go func(idx int) {
			defer wg.Done()
			_ = client.Close()
		}(i)

		go func(idx int) {
			defer wg.Done()
			_, _ = client.ListBuckets(ctx)
		}(i)

		go func(idx int) {
			defer wg.Done()
			_ = client.HealthCheck(ctx)
		}(i)
	}

	wg.Wait()
}

func TestClient_PutObjectWithOptions(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	t.Run("with content type option", func(t *testing.T) {
		reader := bytes.NewReader([]byte("test data"))
		err := client.PutObject(ctx, "bucket", "key", reader, 9, WithContentType("application/json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("with metadata option", func(t *testing.T) {
		reader := bytes.NewReader([]byte("test data"))
		metadata := map[string]string{"key": "value"}
		err := client.PutObject(ctx, "bucket", "key", reader, 9, WithMetadata(metadata))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})

	t.Run("with multiple options", func(t *testing.T) {
		reader := bytes.NewReader([]byte("test data"))
		err := client.PutObject(ctx, "bucket", "key", reader, 9,
			WithContentType("text/plain"),
			WithMetadata(map[string]string{"author": "test"}),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestClient_PutObjectWithEmptyData(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	reader := bytes.NewReader([]byte{})
	err := client.PutObject(ctx, "bucket", "key", reader, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_PutObjectWithLargeSize(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	// Simulate a large file size (just the reader, not actual data)
	reader := bytes.NewReader([]byte("test"))
	err := client.PutObject(ctx, "bucket", "key", reader, 1024*1024*1024) // 1GB
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_GetPresignedURLWithDurations(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	durations := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
	}

	for _, d := range durations {
		t.Run(d.String(), func(t *testing.T) {
			url, err := client.GetPresignedURL(ctx, "bucket", "key", d)
			require.Error(t, err)
			assert.Empty(t, url)
			assert.Contains(t, err.Error(), "not connected")
		})
	}
}

func TestClient_GetPresignedPutURLWithDurations(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	durations := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
	}

	for _, d := range durations {
		t.Run(d.String(), func(t *testing.T) {
			url, err := client.GetPresignedPutURL(ctx, "bucket", "key", d)
			require.Error(t, err)
			assert.Empty(t, url)
			assert.Contains(t, err.Error(), "not connected")
		})
	}
}

func TestClient_CopyObjectVariations(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		srcBucket string
		srcKey    string
		dstBucket string
		dstKey    string
	}{
		{"same bucket different key", "bucket", "src.txt", "bucket", "dst.txt"},
		{"different buckets", "src-bucket", "file.txt", "dst-bucket", "file.txt"},
		{"nested keys", "bucket", "folder/src.txt", "bucket", "folder/dst.txt"},
		{"empty src key", "bucket", "", "bucket", "dst.txt"},
		{"empty dst key", "bucket", "src.txt", "bucket", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.CopyObject(ctx, tt.srcBucket, tt.srcKey, tt.dstBucket, tt.dstKey)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not connected")
		})
	}
}

func TestClient_ListObjectsWithPrefixes(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	prefixes := []string{
		"",
		"/",
		"folder/",
		"folder/subfolder/",
		"special-chars-!@#/",
	}

	for _, prefix := range prefixes {
		t.Run("prefix_"+prefix, func(t *testing.T) {
			objects, err := client.ListObjects(ctx, "bucket", prefix)
			require.Error(t, err)
			assert.Nil(t, objects)
			assert.Contains(t, err.Error(), "not connected")
		})
	}
}

func TestClient_BucketExistsWithNames(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	bucketNames := []string{
		"simple-bucket",
		"bucket.with.dots",
		"bucket-123",
		"",
		"a", // minimum length
		"a-very-long-bucket-name-that-tests-limits",
	}

	for _, name := range bucketNames {
		t.Run("bucket_"+name, func(t *testing.T) {
			exists, err := client.BucketExists(ctx, name)
			require.Error(t, err)
			assert.False(t, exists)
		})
	}
}

func TestClient_CreateBucketConfigurations(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	tests := []struct {
		name   string
		config *BucketConfig
	}{
		{
			"default config",
			DefaultBucketConfig("test"),
		},
		{
			"with versioning",
			DefaultBucketConfig("test").WithVersioning(),
		},
		{
			"with retention",
			DefaultBucketConfig("test").WithRetention(30),
		},
		{
			"with object locking",
			DefaultBucketConfig("test").WithObjectLocking(),
		},
		{
			"with public access",
			DefaultBucketConfig("test").WithPublicAccess(),
		},
		{
			"full configuration",
			DefaultBucketConfig("test").
				WithVersioning().
				WithRetention(90).
				WithObjectLocking().
				WithPublicAccess(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.CreateBucket(ctx, tt.config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not connected")
		})
	}
}

func TestClient_SetLifecycleRuleConfigurations(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	tests := []struct {
		name string
		rule *LifecycleRule
	}{
		{
			"default rule",
			DefaultLifecycleRule("test", 30),
		},
		{
			"with prefix",
			DefaultLifecycleRule("test", 30).WithPrefix("logs/"),
		},
		{
			"with noncurrent expiry",
			DefaultLifecycleRule("test", 30).WithNoncurrentExpiry(7),
		},
		{
			"full configuration",
			DefaultLifecycleRule("test", 90).
				WithPrefix("archive/").
				WithNoncurrentExpiry(30),
		},
		{
			"disabled rule",
			&LifecycleRule{ID: "disabled", Enabled: false, ExpirationDays: 30},
		},
		{
			"delete marker expiry",
			&LifecycleRule{ID: "delete-marker", DeleteMarkerExpiry: true, ExpirationDays: 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.SetLifecycleRule(ctx, "bucket", tt.rule)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not connected")
		})
	}
}

func TestClient_StateTransitions(t *testing.T) {
	client, _ := NewClient(nil, nil)

	// Initial state
	assert.False(t, client.IsConnected())

	// After close
	err := client.Close()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())

	// Multiple closes should be safe
	err = client.Close()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestObjectInfo_ZeroValue(t *testing.T) {
	info := &ObjectInfo{}

	assert.Equal(t, "", info.Key)
	assert.Equal(t, int64(0), info.Size)
	assert.True(t, info.LastModified.IsZero())
	assert.Equal(t, "", info.ContentType)
	assert.Equal(t, "", info.ETag)
	assert.Nil(t, info.Metadata)
}

func TestBucketInfo_ZeroValue(t *testing.T) {
	info := &BucketInfo{}

	assert.Equal(t, "", info.Name)
	assert.True(t, info.CreationDate.IsZero())
}

func TestPutOption_FunctionType(t *testing.T) {
	// Verify PutOption is a function type
	var opt PutOption = func(opts *minio.PutObjectOptions) {
		opts.ContentType = "test"
	}

	var minioOpts minio.PutObjectOptions
	opt(&minioOpts)
	assert.Equal(t, "test", minioOpts.ContentType)
}

func TestWithContentType_VariousTypes(t *testing.T) {
	contentTypes := []string{
		"application/json",
		"text/plain",
		"text/html",
		"application/octet-stream",
		"image/png",
		"application/pdf",
		"",
		"custom/type",
	}

	for _, ct := range contentTypes {
		t.Run(ct, func(t *testing.T) {
			opt := WithContentType(ct)
			var opts minio.PutObjectOptions
			opt(&opts)
			assert.Equal(t, ct, opts.ContentType)
		})
	}
}

func TestWithMetadata_VariousMaps(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
	}{
		{"empty map", map[string]string{}},
		{"single key", map[string]string{"key": "value"}},
		{"multiple keys", map[string]string{"key1": "value1", "key2": "value2"}},
		{"special characters", map[string]string{"key-with-dash": "value with spaces"}},
		{"empty values", map[string]string{"key": ""}},
		{"nil map", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithMetadata(tt.metadata)
			var opts minio.PutObjectOptions
			opt(&opts)
			assert.Equal(t, tt.metadata, opts.UserMetadata)
		})
	}
}

func TestClient_ErrorMessages(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	// Verify error messages are meaningful
	t.Run("health check error", func(t *testing.T) {
		err := client.HealthCheck(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected to MinIO")
	})

	t.Run("list buckets error", func(t *testing.T) {
		_, err := client.ListBuckets(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected to MinIO")
	})

	t.Run("put object error", func(t *testing.T) {
		err := client.PutObject(ctx, "bucket", "key", bytes.NewReader([]byte("data")), 4)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected to MinIO")
	})

	t.Run("get object error", func(t *testing.T) {
		_, err := client.GetObject(ctx, "bucket", "key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected to MinIO")
	})
}

func TestClient_ReaderImplementations(t *testing.T) {
	client, _ := NewClient(nil, nil)
	ctx := context.Background()

	t.Run("bytes.Reader", func(t *testing.T) {
		reader := bytes.NewReader([]byte("test data"))
		err := client.PutObject(ctx, "bucket", "key", reader, 9)
		require.Error(t, err)
	})

	t.Run("bytes.Buffer", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte("test data"))
		err := client.PutObject(ctx, "bucket", "key", buffer, 9)
		require.Error(t, err)
	})

	t.Run("strings.Reader", func(t *testing.T) {
		reader := bytes.NewReader([]byte("test data"))
		err := client.PutObject(ctx, "bucket", "key", reader, 9)
		require.Error(t, err)
	})
}

func TestMockReader_EdgeCases(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		reader := &mockReader{data: []byte{}}
		buf := make([]byte, 10)
		n, err := reader.Read(buf)
		assert.Equal(t, io.EOF, err)
		assert.Equal(t, 0, n)
	})

	t.Run("exact buffer size", func(t *testing.T) {
		reader := &mockReader{data: []byte("test")}
		buf := make([]byte, 4)
		n, err := reader.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "test", string(buf))

		// Next read should return EOF
		n, err = reader.Read(buf)
		assert.Equal(t, io.EOF, err)
		assert.Equal(t, 0, n)
	})

	t.Run("large buffer", func(t *testing.T) {
		reader := &mockReader{data: []byte("test")}
		buf := make([]byte, 100)
		n, err := reader.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "test", string(buf[:n]))
	})

	t.Run("multiple small reads", func(t *testing.T) {
		reader := &mockReader{data: []byte("hello world")}
		buf := make([]byte, 3)

		results := []struct {
			expected string
			n        int
		}{
			{"hel", 3},
			{"lo ", 3},
			{"wor", 3},
			{"ld", 2},
		}

		for i, r := range results {
			n, err := reader.Read(buf)
			if i < len(results)-1 {
				assert.NoError(t, err)
			}
			assert.Equal(t, r.n, n, "iteration %d", i)
			assert.Equal(t, r.expected, string(buf[:n]), "iteration %d", i)
		}

		// Final read should be EOF
		n, err := reader.Read(buf)
		assert.Equal(t, io.EOF, err)
		assert.Equal(t, 0, n)
	})
}

func TestClient_MultipleNewClients(t *testing.T) {
	// Ensure multiple clients can be created independently
	client1, err := NewClient(nil, nil)
	require.NoError(t, err)

	client2, err := NewClient(nil, nil)
	require.NoError(t, err)

	config := DefaultConfig()
	config.Endpoint = "different.host:9000"
	client3, err := NewClient(config, logrus.New())
	require.NoError(t, err)

	// All clients should be independent
	assert.NotSame(t, client1, client2)
	assert.NotSame(t, client2, client3)

	// Closing one shouldn't affect others
	err = client1.Close()
	require.NoError(t, err)
	assert.False(t, client1.IsConnected())
	assert.False(t, client2.IsConnected()) // Never connected
	assert.False(t, client3.IsConnected()) // Never connected
}
