package minio

import (
	"bytes"
	"context"
	"io"
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
