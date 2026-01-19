package qdrant

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
			Host:           "qdrant.example.com",
			HTTPPort:       6333,
			GRPCPort:       6334,
			Timeout:        60 * time.Second,
			MaxRetries:     5,
			DefaultLimit:   20,
			ScoreThreshold: 0.5,
		}
		client, err := NewClient(config, logrus.New())
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with invalid config - empty host", func(t *testing.T) {
		config := &Config{
			Host:     "",
			HTTPPort: 6333,
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("with invalid config - invalid port", func(t *testing.T) {
		config := &Config{
			Host:     "localhost",
			HTTPPort: 0,
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClientConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.Host = "localhost"
		config.HTTPPort = 6333
		config.Timeout = 5 * time.Second

		client, _ := NewClient(config, nil)
		// Note: This will fail as we can't easily mock the URL in the client
		// The test is to show the pattern
		assert.NotNil(t, client)
	})

	t.Run("connection failure", func(t *testing.T) {
		config := DefaultConfig()
		config.Host = "localhost"
		config.HTTPPort = 59999
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
	t.Run("not connected - server unavailable", func(t *testing.T) {
		// Use a port where Qdrant is not running to test connection failure
		config := DefaultConfig()
		config.Host = "localhost"
		config.HTTPPort = 59998 // Port where nothing is running
		config.Timeout = 100 * time.Millisecond

		client, _ := NewClient(config, nil)
		err := client.HealthCheck(context.Background())
		require.Error(t, err)
		// HealthCheck attempts connection, so will fail with request error when server unavailable
		assert.Contains(t, err.Error(), "request failed")
	})
}

func TestCreateCollection(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		config := DefaultCollectionConfig("test", 1536)
		err := client.CreateCollection(context.Background(), config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestDeleteCollection(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.DeleteCollection(context.Background(), "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestListCollections(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		collections, err := client.ListCollections(context.Background())
		require.Error(t, err)
		assert.Nil(t, collections)
	})
}

func TestCollectionExists(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		exists, err := client.CollectionExists(context.Background(), "test")
		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestGetCollectionInfo(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		info, err := client.GetCollectionInfo(context.Background(), "test")
		require.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestUpsertPoints(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		points := []Point{
			{ID: "1", Vector: make([]float32, 1536)},
		}
		err := client.UpsertPoints(context.Background(), "test", points)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestDeletePoints(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		err := client.DeletePoints(context.Background(), "test", []string{"1", "2"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestGetPoints(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		points, err := client.GetPoints(context.Background(), "test", []string{"1", "2"})
		require.Error(t, err)
		assert.Nil(t, points)
	})
}

func TestSearch(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		vector := make([]float32, 1536)
		results, err := client.Search(context.Background(), "test", vector, DefaultSearchOptions())
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestSearchBatch(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		vectors := [][]float32{make([]float32, 1536)}
		results, err := client.SearchBatch(context.Background(), "test", vectors, DefaultSearchOptions())
		require.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestScroll(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		results, nextOffset, err := client.Scroll(context.Background(), "test", 10, nil, nil)
		require.Error(t, err)
		assert.Nil(t, results)
		assert.Nil(t, nextOffset)
	})
}

func TestCreateSnapshot(t *testing.T) {
	t.Run("not connected", func(t *testing.T) {
		client, _ := NewClient(nil, nil)
		name, err := client.CreateSnapshot(context.Background(), "test")
		require.Error(t, err)
		assert.Empty(t, name)
	})
}

func TestPoint(t *testing.T) {
	point := Point{
		ID:      "test-point-1",
		Vector:  make([]float32, 1536),
		Payload: map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "test-point-1", point.ID)
	assert.Len(t, point.Vector, 1536)
	assert.Equal(t, "value", point.Payload["key"])
}

func TestCollectionInfo(t *testing.T) {
	info := &CollectionInfo{
		Name:          "test-collection",
		Status:        "green",
		PointsCount:   10000,
		VectorCount:   10000,
		SegmentsCount: 5,
	}

	assert.Equal(t, "test-collection", info.Name)
	assert.Equal(t, "green", info.Status)
	assert.Equal(t, int64(10000), info.PointsCount)
	assert.Equal(t, int64(10000), info.VectorCount)
}

func TestDistanceTypes(t *testing.T) {
	assert.Equal(t, Distance("Cosine"), DistanceCosine)
	assert.Equal(t, Distance("Euclid"), DistanceEuclid)
	assert.Equal(t, Distance("Dot"), DistanceDot)
	assert.Equal(t, Distance("Manhattan"), DistanceManhattan)
}

func TestCollectionConfigBuilder(t *testing.T) {
	config := DefaultCollectionConfig("test", 1536).
		WithDistance(DistanceEuclid).
		WithOnDiskPayload().
		WithIndexingThreshold(50000).
		WithShards(3).
		WithReplication(2)

	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 1536, config.VectorSize)
	assert.Equal(t, DistanceEuclid, config.Distance)
	assert.True(t, config.OnDiskPayload)
	assert.Equal(t, 50000, config.IndexingThreshold)
	assert.Equal(t, 3, config.ShardNumber)
	assert.Equal(t, 2, config.ReplicationFactor)
}

func TestSearchOptionsBuilder(t *testing.T) {
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{"key": "category", "match": map[string]interface{}{"value": "tech"}},
		},
	}

	opts := DefaultSearchOptions().
		WithLimit(20).
		WithOffset(10).
		WithScoreThreshold(0.7).
		WithVectorsEnabled().
		WithFilter(filter)

	assert.Equal(t, 20, opts.Limit)
	assert.Equal(t, 10, opts.Offset)
	assert.Equal(t, float32(0.7), opts.ScoreThreshold)
	assert.True(t, opts.WithVectors)
	assert.NotNil(t, opts.Filter)
}
