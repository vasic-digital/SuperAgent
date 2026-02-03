// Package qdrant provides adapter types for the digital.vasic.vectordb module.
// This adapter re-exports types and functions to maintain backward compatibility
// with code using the internal/vectordb/qdrant package.
package qdrant

import (
	"context"
	"fmt"
	"sync"
	"time"

	"digital.vasic.vectordb/pkg/client"
	extqdrant "digital.vasic.vectordb/pkg/qdrant"
	"github.com/sirupsen/logrus"
)

// ScoredPoint represents a search result from Qdrant.
type ScoredPoint struct {
	ID      string                 `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload,omitempty"`
	Vector  []float32              `json:"vector,omitempty"`
}

// Point represents a vector point in Qdrant.
type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// SearchOptions configures vector search parameters.
type SearchOptions struct {
	Limit          int
	ScoreThreshold float32
	WithPayload    bool
	WithVectors    bool
	Filter         map[string]interface{}
}

// DefaultSearchOptions returns search options with sensible defaults.
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		Limit:       10,
		WithPayload: true,
		WithVectors: false,
	}
}

// WithLimit sets the limit and returns the options for chaining.
func (o *SearchOptions) WithLimit(limit int) *SearchOptions {
	o.Limit = limit
	return o
}

// CollectionConfig configures a Qdrant collection.
type CollectionConfig struct {
	Name       string
	VectorSize int
	Distance   DistanceMetric
}

// DistanceMetric represents the distance metric for vector similarity.
type DistanceMetric string

const (
	DistanceCosine    DistanceMetric = "Cosine"
	DistanceDot       DistanceMetric = "Dot"
	DistanceEuclidean DistanceMetric = "Euclid"
)

// CollectionInfo holds information about a collection.
type CollectionInfo struct {
	Name         string `json:"name"`
	VectorSize   int    `json:"vector_size"`
	PointsCount  int64  `json:"points_count"`
	Distance     string `json:"distance"`
	OptimizersOk bool   `json:"optimizers_ok"`
}

// Config holds Qdrant client configuration.
type Config struct {
	Host    string
	Port    int
	APIKey  string
	Timeout time.Duration
	UseTLS  bool
}

// DefaultConfig returns a default Qdrant configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:    "localhost",
		Port:    6333,
		Timeout: 30 * time.Second,
		UseTLS:  false,
	}
}

// DefaultCollectionConfig creates a default collection configuration.
func DefaultCollectionConfig(name string, vectorSize int) *CollectionConfig {
	return &CollectionConfig{
		Name:       name,
		VectorSize: vectorSize,
		Distance:   DistanceCosine,
	}
}

// Client wraps the extracted module's Qdrant client to provide backward compatibility.
type Client struct {
	extClient *extqdrant.Client
	logger    *logrus.Logger
	mu        sync.RWMutex
	connected bool
	config    *Config
}

// NewClient creates a new Qdrant client adapter.
func NewClient(config *Config, logger *logrus.Logger) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	extConfig := &extqdrant.Config{
		Host:     config.Host,
		HTTPPort: config.Port,
		GRPCPort: config.Port + 1, // gRPC port is typically HTTP port + 1
		APIKey:   config.APIKey,
		Timeout:  config.Timeout,
		TLS:      config.UseTLS,
	}

	extClient, err := extqdrant.NewClient(extConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	return &Client{
		extClient: extClient,
		logger:    logger,
		config:    config,
	}, nil
}

// Connect connects to Qdrant.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.extClient.Connect(ctx); err != nil {
		return err
	}
	c.connected = true
	c.logger.Info("Connected to Qdrant via adapter")
	return nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return c.extClient.Close()
}

// HealthCheck checks the health of Qdrant.
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.extClient.Connect(ctx) // Connect acts as health check
}

// CreateCollection creates a new collection.
func (c *Client) CreateCollection(ctx context.Context, config *CollectionConfig) error {
	metric := client.DistanceCosine
	switch config.Distance {
	case DistanceDot:
		metric = client.DistanceDotProduct
	case DistanceEuclidean:
		metric = client.DistanceEuclidean
	}

	extConfig := client.CollectionConfig{
		Name:      config.Name,
		Dimension: config.VectorSize,
		Metric:    metric,
	}
	return c.extClient.CreateCollection(ctx, extConfig)
}

// DeleteCollection deletes a collection.
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	return c.extClient.DeleteCollection(ctx, name)
}

// CollectionExists checks if a collection exists.
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	collections, err := c.extClient.ListCollections(ctx)
	if err != nil {
		return false, err
	}
	for _, col := range collections {
		if col == name {
			return true, nil
		}
	}
	return false, nil
}

// ListCollections lists all collections.
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	return c.extClient.ListCollections(ctx)
}

// GetCollectionInfo returns information about a collection.
func (c *Client) GetCollectionInfo(ctx context.Context, name string) (*CollectionInfo, error) {
	// The extracted module doesn't have a direct equivalent, so we provide minimal info
	exists, err := c.CollectionExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("collection %s not found", name)
	}
	return &CollectionInfo{
		Name:         name,
		OptimizersOk: true,
	}, nil
}

// UpsertPoints upserts points into a collection.
func (c *Client) UpsertPoints(ctx context.Context, collection string, points []Point) error {
	vectors := make([]client.Vector, len(points))
	for i, p := range points {
		vectors[i] = client.Vector{
			ID:       p.ID,
			Values:   p.Vector,
			Metadata: p.Payload,
		}
	}
	return c.extClient.Upsert(ctx, collection, vectors)
}

// DeletePoints deletes points from a collection.
func (c *Client) DeletePoints(ctx context.Context, collection string, ids []string) error {
	return c.extClient.Delete(ctx, collection, ids)
}

// GetPoints retrieves points by their IDs.
func (c *Client) GetPoints(ctx context.Context, collection string, ids []string) ([]Point, error) {
	vectors, err := c.extClient.Get(ctx, collection, ids)
	if err != nil {
		return nil, err
	}
	points := make([]Point, len(vectors))
	for i, v := range vectors {
		points[i] = Point{
			ID:      v.ID,
			Vector:  v.Values,
			Payload: v.Metadata,
		}
	}
	return points, nil
}

// GetPoint retrieves a single point by ID.
func (c *Client) GetPoint(ctx context.Context, collection, id string) (*Point, error) {
	points, err := c.GetPoints(ctx, collection, []string{id})
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("point not found: %s", id)
	}
	return &points[0], nil
}

// Search performs vector similarity search.
func (c *Client) Search(ctx context.Context, collection string, vector []float32, opts *SearchOptions) ([]ScoredPoint, error) {
	if opts == nil {
		opts = DefaultSearchOptions()
	}

	query := client.SearchQuery{
		Vector:   vector,
		TopK:     opts.Limit,
		MinScore: float64(opts.ScoreThreshold),
		Filter:   opts.Filter,
	}

	results, err := c.extClient.Search(ctx, collection, query)
	if err != nil {
		return nil, err
	}

	scoredPoints := make([]ScoredPoint, len(results))
	for i, r := range results {
		scoredPoints[i] = ScoredPoint{
			ID:      r.ID,
			Score:   r.Score,
			Payload: r.Metadata,
		}
		if opts.WithVectors && r.Vector != nil {
			scoredPoints[i].Vector = r.Vector
		}
	}
	return scoredPoints, nil
}

// SearchBatch performs batch vector similarity search.
func (c *Client) SearchBatch(ctx context.Context, collection string, vectors [][]float32, opts *SearchOptions) ([][]ScoredPoint, error) {
	results := make([][]ScoredPoint, len(vectors))
	for i, v := range vectors {
		r, err := c.Search(ctx, collection, v, opts)
		if err != nil {
			return nil, err
		}
		results[i] = r
	}
	return results, nil
}

// CountPoints counts points in a collection with optional filter.
func (c *Client) CountPoints(ctx context.Context, collection string, filter map[string]interface{}) (int64, error) {
	// The extracted module doesn't have a direct count method
	// Return 0 as a fallback
	return 0, nil
}
