// Package gptcache provides semantic caching for LLM queries to reduce redundant API calls.
package gptcache

import (
	"time"
)

// CacheRequest represents a request to cache or retrieve an LLM response.
type CacheRequest struct {
	// Query is the original query text.
	Query string `json:"query"`
	// Embedding is the vector embedding of the query.
	Embedding []float64 `json:"embedding,omitempty"`
	// Model is the LLM model used.
	Model string `json:"model,omitempty"`
	// Provider is the LLM provider.
	Provider string `json:"provider,omitempty"`
	// Parameters are the generation parameters.
	Parameters *GenerationParameters `json:"parameters,omitempty"`
	// Metadata is additional metadata for the request.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CacheResponse represents a cached LLM response.
type CacheResponse struct {
	// Response is the LLM response text.
	Response string `json:"response"`
	// CacheHit indicates if this was retrieved from cache.
	CacheHit bool `json:"cache_hit"`
	// Similarity is the similarity score if from cache (0-1).
	Similarity float64 `json:"similarity,omitempty"`
	// CacheID is the ID of the cache entry.
	CacheID string `json:"cache_id,omitempty"`
	// GeneratedAt is when the response was originally generated.
	GeneratedAt time.Time `json:"generated_at,omitempty"`
	// Metadata is additional metadata from the cache entry.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GenerationParameters holds LLM generation parameters.
type GenerationParameters struct {
	// Temperature controls randomness (0-2).
	Temperature float64 `json:"temperature,omitempty"`
	// MaxTokens is the maximum tokens to generate.
	MaxTokens int `json:"max_tokens,omitempty"`
	// TopP is nucleus sampling probability.
	TopP float64 `json:"top_p,omitempty"`
	// TopK is the number of top tokens to consider.
	TopK int `json:"top_k,omitempty"`
	// Stop is a list of stop sequences.
	Stop []string `json:"stop,omitempty"`
	// FrequencyPenalty penalizes frequent tokens.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	// PresencePenalty penalizes tokens that have appeared.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
}

// EmbeddingProvider defines the interface for generating embeddings.
type EmbeddingProvider interface {
	// Embed generates an embedding vector for the given text.
	Embed(text string) ([]float64, error)
	// EmbedBatch generates embeddings for multiple texts.
	EmbedBatch(texts []string) ([][]float64, error)
	// Dimension returns the embedding dimension.
	Dimension() int
}

// CacheBackend defines the interface for cache storage backends.
type CacheBackend interface {
	// Get retrieves an entry by ID.
	Get(id string) (*CacheEntry, error)
	// Set stores an entry.
	Set(entry *CacheEntry) error
	// Delete removes an entry.
	Delete(id string) error
	// FindSimilar finds entries similar to the given embedding.
	FindSimilar(embedding []float64, threshold float64, limit int) ([]*CacheHit, error)
	// Clear removes all entries.
	Clear() error
	// Size returns the number of entries.
	Size() int
}

// QueryNormalizer normalizes queries before caching.
type QueryNormalizer interface {
	// Normalize preprocesses a query for caching.
	Normalize(query string) string
}

// DefaultQueryNormalizer provides basic query normalization.
type DefaultQueryNormalizer struct{}

// Normalize normalizes a query by trimming whitespace.
func (n *DefaultQueryNormalizer) Normalize(query string) string {
	return query // Basic implementation; extend for more sophisticated normalization
}

// CacheEventType defines types of cache events.
type CacheEventType string

const (
	// CacheEventHit indicates a cache hit.
	CacheEventHit CacheEventType = "hit"
	// CacheEventMiss indicates a cache miss.
	CacheEventMiss CacheEventType = "miss"
	// CacheEventSet indicates an entry was stored.
	CacheEventSet CacheEventType = "set"
	// CacheEventEvict indicates an entry was evicted.
	CacheEventEvict CacheEventType = "evict"
	// CacheEventDelete indicates an entry was deleted.
	CacheEventDelete CacheEventType = "delete"
	// CacheEventClear indicates the cache was cleared.
	CacheEventClear CacheEventType = "clear"
)

// CacheEvent represents a cache operation event.
type CacheEvent struct {
	// Type is the event type.
	Type CacheEventType `json:"type"`
	// EntryID is the ID of the affected entry (if applicable).
	EntryID string `json:"entry_id,omitempty"`
	// Query is the query (if applicable).
	Query string `json:"query,omitempty"`
	// Similarity is the similarity score (for hits).
	Similarity float64 `json:"similarity,omitempty"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}

// CacheEventHandler handles cache events.
type CacheEventHandler func(event *CacheEvent)

// SemanticMatchResult represents the result of a semantic match search.
type SemanticMatchResult struct {
	// Entries are the matched cache entries.
	Entries []*CacheEntry `json:"entries"`
	// Similarities are the similarity scores for each entry.
	Similarities []float64 `json:"similarities"`
	// SearchTime is how long the search took.
	SearchTime time.Duration `json:"search_time"`
}

// PersistenceConfig holds configuration for cache persistence.
type PersistenceConfig struct {
	// Enabled indicates if persistence is enabled.
	Enabled bool `json:"enabled"`
	// Path is the file path for persistence.
	Path string `json:"path"`
	// SyncInterval is how often to sync to disk.
	SyncInterval time.Duration `json:"sync_interval"`
	// Compression indicates if compression should be used.
	Compression bool `json:"compression"`
}

// ClusterConfig holds configuration for distributed caching.
type ClusterConfig struct {
	// Enabled indicates if clustering is enabled.
	Enabled bool `json:"enabled"`
	// Nodes are the addresses of cluster nodes.
	Nodes []string `json:"nodes"`
	// ReplicationFactor is the number of replicas.
	ReplicationFactor int `json:"replication_factor"`
	// PartitionCount is the number of partitions.
	PartitionCount int `json:"partition_count"`
}

// CacheMetrics holds cache performance metrics.
type CacheMetrics struct {
	// TotalQueries is the total number of queries.
	TotalQueries int64 `json:"total_queries"`
	// CacheHits is the number of cache hits.
	CacheHits int64 `json:"cache_hits"`
	// CacheMisses is the number of cache misses.
	CacheMisses int64 `json:"cache_misses"`
	// HitRate is the cache hit rate.
	HitRate float64 `json:"hit_rate"`
	// AverageSimilarity is the average similarity for hits.
	AverageSimilarity float64 `json:"average_similarity"`
	// AverageLatencyMs is the average lookup latency in milliseconds.
	AverageLatencyMs float64 `json:"average_latency_ms"`
	// CacheSize is the current cache size.
	CacheSize int `json:"cache_size"`
	// EvictionCount is the number of evictions.
	EvictionCount int64 `json:"eviction_count"`
	// LastUpdated is when metrics were last updated.
	LastUpdated time.Time `json:"last_updated"`
}

// QueryContext provides additional context for cache lookups.
type QueryContext struct {
	// ConversationID groups queries in a conversation.
	ConversationID string `json:"conversation_id,omitempty"`
	// SessionID identifies the session.
	SessionID string `json:"session_id,omitempty"`
	// UserID identifies the user.
	UserID string `json:"user_id,omitempty"`
	// Tags are tags for filtering.
	Tags []string `json:"tags,omitempty"`
	// Timestamp is when the query was made.
	Timestamp time.Time `json:"timestamp"`
}

// CachePolicy defines caching behavior rules.
type CachePolicy struct {
	// MinQueryLength is the minimum query length to cache.
	MinQueryLength int `json:"min_query_length"`
	// MaxQueryLength is the maximum query length to cache.
	MaxQueryLength int `json:"max_query_length"`
	// MinResponseLength is the minimum response length to cache.
	MinResponseLength int `json:"min_response_length"`
	// ExcludedModels are models that should not be cached.
	ExcludedModels []string `json:"excluded_models,omitempty"`
	// ExcludedProviders are providers that should not be cached.
	ExcludedProviders []string `json:"excluded_providers,omitempty"`
	// ExcludedPatterns are regex patterns to exclude from caching.
	ExcludedPatterns []string `json:"excluded_patterns,omitempty"`
	// CacheableTemperatureMax is the max temperature for caching (higher = more random).
	CacheableTemperatureMax float64 `json:"cacheable_temperature_max"`
}

// DefaultCachePolicy returns a default cache policy.
func DefaultCachePolicy() *CachePolicy {
	return &CachePolicy{
		MinQueryLength:          5,
		MaxQueryLength:          100000,
		MinResponseLength:       10,
		ExcludedModels:          nil,
		ExcludedProviders:       nil,
		ExcludedPatterns:        nil,
		CacheableTemperatureMax: 0.5, // Only cache low-temperature (deterministic) responses
	}
}

// ShouldCache checks if a request should be cached based on the policy.
func (p *CachePolicy) ShouldCache(req *CacheRequest, responseLength int) bool {
	if len(req.Query) < p.MinQueryLength || len(req.Query) > p.MaxQueryLength {
		return false
	}
	if responseLength < p.MinResponseLength {
		return false
	}
	if req.Parameters != nil && req.Parameters.Temperature > p.CacheableTemperatureMax {
		return false
	}
	for _, model := range p.ExcludedModels {
		if req.Model == model {
			return false
		}
	}
	for _, provider := range p.ExcludedProviders {
		if req.Provider == provider {
			return false
		}
	}
	return true
}

// IndexType defines the type of index for similarity search.
type IndexType string

const (
	// IndexTypeFlat uses brute-force search (exact but slow).
	IndexTypeFlat IndexType = "flat"
	// IndexTypeIVF uses inverted file index (faster but approximate).
	IndexTypeIVF IndexType = "ivf"
	// IndexTypeHNSW uses hierarchical navigable small world graphs.
	IndexTypeHNSW IndexType = "hnsw"
	// IndexTypeLSH uses locality-sensitive hashing.
	IndexTypeLSH IndexType = "lsh"
)

// IndexConfig holds configuration for the similarity index.
type IndexConfig struct {
	// Type is the index type.
	Type IndexType `json:"type"`
	// NumPartitions is the number of partitions for IVF.
	NumPartitions int `json:"num_partitions,omitempty"`
	// NumNeighbors is the number of neighbors for HNSW.
	NumNeighbors int `json:"num_neighbors,omitempty"`
	// NumBits is the number of bits for LSH.
	NumBits int `json:"num_bits,omitempty"`
	// NumTables is the number of hash tables for LSH.
	NumTables int `json:"num_tables,omitempty"`
}

// DefaultIndexConfig returns a default index configuration.
func DefaultIndexConfig() *IndexConfig {
	return &IndexConfig{
		Type:          IndexTypeFlat,
		NumPartitions: 100,
		NumNeighbors:  32,
		NumBits:       128,
		NumTables:     8,
	}
}
