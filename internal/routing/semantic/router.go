// Package semantic provides a semantic router for intelligent query routing
// based on intent classification and semantic similarity.
package semantic

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Router provides semantic routing for queries to appropriate handlers
type Router struct {
	routes    []*Route
	encoder   Encoder
	cache     *SemanticCache
	config    *RouterConfig
	logger    *logrus.Logger
	mu        sync.RWMutex
}

// Route defines a semantic route
type Route struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Utterances  []string     `json:"utterances"`  // Example phrases for this route
	Handler     RouteHandler `json:"-"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ModelTier   ModelTier    `json:"model_tier"`
	Embedding   []float32    `json:"-"` // Averaged embedding of utterances
	Score       float64      `json:"-"` // Match score (set during routing)
}

// RouteHandler handles matched routes
type RouteHandler func(ctx context.Context, query string) (*RouteResult, error)

// RouteResult contains the result of route handling
type RouteResult struct {
	Content     string                 `json:"content"`
	Model       string                 `json:"model,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CacheKey    string                 `json:"cache_key,omitempty"`
	Latency     time.Duration          `json:"latency"`
}

// ModelTier indicates the complexity/cost tier for a route
type ModelTier string

const (
	ModelTierSimple   ModelTier = "simple"   // Fast, cheap models (haiku, mini)
	ModelTierStandard ModelTier = "standard" // Balanced models (sonnet, gpt-4)
	ModelTierComplex  ModelTier = "complex"  // Powerful models (opus, o1)
)

// Encoder generates embeddings for semantic matching
type Encoder interface {
	Encode(ctx context.Context, texts []string) ([][]float32, error)
	GetDimension() int
}

// RouterConfig configures the semantic router
type RouterConfig struct {
	// Threshold for route matching (0-1)
	ScoreThreshold float64 `json:"score_threshold"`
	// Maximum number of candidate routes
	TopK int `json:"top_k"`
	// Enable semantic caching
	EnableCache bool `json:"enable_cache"`
	// Cache TTL
	CacheTTL time.Duration `json:"cache_ttl"`
	// Fallback route name
	FallbackRoute string `json:"fallback_route"`
	// Aggregation method for multi-utterance routes
	AggregationMethod AggregationMethod `json:"aggregation_method"`
}

// AggregationMethod defines how to aggregate utterance embeddings
type AggregationMethod string

const (
	AggregationMean AggregationMethod = "mean"
	AggregationMax  AggregationMethod = "max"
)

// DefaultRouterConfig returns default configuration
func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		ScoreThreshold:    0.7,
		TopK:              5,
		EnableCache:       true,
		CacheTTL:          30 * time.Minute,
		FallbackRoute:     "",
		AggregationMethod: AggregationMean,
	}
}

// NewRouter creates a new semantic router
func NewRouter(encoder Encoder, config *RouterConfig, logger *logrus.Logger) *Router {
	if config == nil {
		config = DefaultRouterConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	r := &Router{
		routes:  make([]*Route, 0),
		encoder: encoder,
		config:  config,
		logger:  logger,
	}

	if config.EnableCache {
		r.cache = NewSemanticCache(config.CacheTTL, encoder)
	}

	return r
}

// AddRoute adds a route to the router
func (r *Router) AddRoute(ctx context.Context, route *Route) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if route.Name == "" {
		return fmt.Errorf("route name is required")
	}

	if len(route.Utterances) == 0 {
		return fmt.Errorf("route must have at least one utterance")
	}

	// Generate embeddings for utterances
	embeddings, err := r.encoder.Encode(ctx, route.Utterances)
	if err != nil {
		return fmt.Errorf("failed to encode utterances: %w", err)
	}

	// Aggregate embeddings
	route.Embedding = r.aggregateEmbeddings(embeddings)

	r.routes = append(r.routes, route)

	r.logger.WithFields(logrus.Fields{
		"route":      route.Name,
		"utterances": len(route.Utterances),
	}).Debug("Route added")

	return nil
}

// Route routes a query to the best matching route
func (r *Router) Route(ctx context.Context, query string) (*Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.routes) == 0 {
		return nil, fmt.Errorf("no routes configured")
	}

	// Check cache first
	if r.cache != nil {
		if cached := r.cache.Get(query); cached != nil {
			r.logger.Debug("Cache hit for query")
			return cached, nil
		}
	}

	// Encode query
	queryEmbeddings, err := r.encoder.Encode(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	if len(queryEmbeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	queryEmbedding := queryEmbeddings[0]

	// Calculate similarity scores
	type routeScore struct {
		route *Route
		score float64
	}

	scores := make([]routeScore, len(r.routes))
	for i, route := range r.routes {
		score := cosineSimilarity(queryEmbedding, route.Embedding)
		scores[i] = routeScore{route: route, score: score}
	}

	// Sort by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Get best match
	if scores[0].score < r.config.ScoreThreshold {
		// No match above threshold, use fallback
		if r.config.FallbackRoute != "" {
			for _, route := range r.routes {
				if route.Name == r.config.FallbackRoute {
					route.Score = scores[0].score
					return route, nil
				}
			}
		}
		return nil, fmt.Errorf("no route matched with sufficient confidence (best: %.2f, threshold: %.2f)",
			scores[0].score, r.config.ScoreThreshold)
	}

	bestRoute := scores[0].route
	bestRoute.Score = scores[0].score

	// Cache the result
	if r.cache != nil {
		r.cache.Set(query, bestRoute)
	}

	r.logger.WithFields(logrus.Fields{
		"query": query[:min(50, len(query))],
		"route": bestRoute.Name,
		"score": bestRoute.Score,
	}).Debug("Query routed")

	return bestRoute, nil
}

// RouteWithCandidates returns top-K candidate routes
func (r *Router) RouteWithCandidates(ctx context.Context, query string) ([]*Route, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.routes) == 0 {
		return nil, fmt.Errorf("no routes configured")
	}

	// Encode query
	queryEmbeddings, err := r.encoder.Encode(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	queryEmbedding := queryEmbeddings[0]

	// Calculate similarity scores and create copies
	type routeScore struct {
		route *Route
		score float64
	}

	scores := make([]routeScore, len(r.routes))
	for i, route := range r.routes {
		score := cosineSimilarity(queryEmbedding, route.Embedding)
		scores[i] = routeScore{route: route, score: score}
	}

	// Sort by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Get top-K
	k := r.config.TopK
	if k > len(scores) {
		k = len(scores)
	}

	candidates := make([]*Route, k)
	for i := 0; i < k; i++ {
		route := *scores[i].route // Copy
		route.Score = scores[i].score
		candidates[i] = &route
	}

	return candidates, nil
}

// RemoveRoute removes a route by name
func (r *Router) RemoveRoute(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.Name == name {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			r.logger.WithField("route", name).Debug("Route removed")
			return
		}
	}
}

// ListRoutes returns all configured routes
func (r *Router) ListRoutes() []*Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]*Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

// ClearCache clears the semantic cache
func (r *Router) ClearCache() {
	if r.cache != nil {
		r.cache.Clear()
	}
}

// aggregateEmbeddings aggregates multiple embeddings into one
func (r *Router) aggregateEmbeddings(embeddings [][]float32) []float32 {
	if len(embeddings) == 0 {
		return nil
	}

	dim := len(embeddings[0])
	result := make([]float32, dim)

	switch r.config.AggregationMethod {
	case AggregationMax:
		for i := 0; i < dim; i++ {
			maxVal := embeddings[0][i]
			for _, emb := range embeddings[1:] {
				if emb[i] > maxVal {
					maxVal = emb[i]
				}
			}
			result[i] = maxVal
		}
	default: // Mean
		for i := 0; i < dim; i++ {
			sum := float32(0)
			for _, emb := range embeddings {
				sum += emb[i]
			}
			result[i] = sum / float32(len(embeddings))
		}
	}

	return result
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
