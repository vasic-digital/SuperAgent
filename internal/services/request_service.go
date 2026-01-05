package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/superagent/superagent/internal/models"
)

// RequestService handles LLM request routing and load balancing
type RequestService struct {
	providers map[string]LLMProvider
	ensemble  *EnsembleService
	memory    *MemoryService
	strategy  RoutingStrategy
	mu        sync.RWMutex
}

// RoutingStrategy defines different routing approaches
type RoutingStrategy interface {
	SelectProvider(providers map[string]LLMProvider, req *models.LLMRequest) (string, error)
}

// ProviderHealth tracks health and performance metrics for providers
type ProviderHealth struct {
	Name          string
	Healthy       bool
	LastCheck     time.Time
	ResponseTime  int64   // Average response time in milliseconds
	SuccessRate   float64 // Success rate (0.0 to 1.0)
	ErrorCount    int64
	TotalRequests int64
	LastError     string
	Weight        float64 // Dynamic weight based on performance
	mu            sync.RWMutex
}

// Routing strategies

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
	counter int64
	mu      sync.Mutex
}

// ProviderMetrics tracks performance metrics for a provider with rolling window
type ProviderMetrics struct {
	SuccessCount   int64
	FailureCount   int64
	TotalLatencyMs int64
	LatencyHistory []int64 // Rolling window of recent latencies
	LastUpdated    time.Time
	mu             sync.RWMutex
}

// GetSuccessRate returns the success rate (0.0 to 1.0)
func (pm *ProviderMetrics) GetSuccessRate() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	total := pm.SuccessCount + pm.FailureCount
	if total == 0 {
		return 1.0 // Default to 1.0 for new providers
	}
	return float64(pm.SuccessCount) / float64(total)
}

// GetAverageLatency returns the average latency in milliseconds
func (pm *ProviderMetrics) GetAverageLatency() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if len(pm.LatencyHistory) == 0 {
		return 1000.0 // Default latency for new providers
	}
	var sum int64
	for _, lat := range pm.LatencyHistory {
		sum += lat
	}
	return float64(sum) / float64(len(pm.LatencyHistory))
}

// RecordSuccess records a successful request with latency
func (pm *ProviderMetrics) RecordSuccess(latencyMs int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.SuccessCount++
	pm.TotalLatencyMs += latencyMs
	pm.LastUpdated = time.Now()
	// Maintain rolling window of last 100 latencies
	pm.LatencyHistory = append(pm.LatencyHistory, latencyMs)
	if len(pm.LatencyHistory) > 100 {
		pm.LatencyHistory = pm.LatencyHistory[1:]
	}
}

// RecordFailure records a failed request
func (pm *ProviderMetrics) RecordFailure() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.FailureCount++
	pm.LastUpdated = time.Now()
}

// MetricsRegistry is a thread-safe registry for provider metrics
type MetricsRegistry struct {
	metrics map[string]*ProviderMetrics
	mu      sync.RWMutex
}

// GlobalMetricsRegistry is the singleton metrics registry
var GlobalMetricsRegistry = &MetricsRegistry{
	metrics: make(map[string]*ProviderMetrics),
}

// GetMetrics returns metrics for a provider, creating if necessary
func (mr *MetricsRegistry) GetMetrics(providerName string) *ProviderMetrics {
	mr.mu.RLock()
	if pm, exists := mr.metrics[providerName]; exists {
		mr.mu.RUnlock()
		return pm
	}
	mr.mu.RUnlock()

	mr.mu.Lock()
	defer mr.mu.Unlock()
	// Double-check after acquiring write lock
	if pm, exists := mr.metrics[providerName]; exists {
		return pm
	}
	pm := &ProviderMetrics{
		LatencyHistory: make([]int64, 0, 100),
		LastUpdated:    time.Now(),
	}
	mr.metrics[providerName] = pm
	return pm
}

// RecordRequest records the outcome of a request to the metrics registry
func (mr *MetricsRegistry) RecordRequest(providerName string, success bool, latencyMs int64) {
	pm := mr.GetMetrics(providerName)
	if success {
		pm.RecordSuccess(latencyMs)
	} else {
		pm.RecordFailure()
	}
}

// WeightedStrategy implements weighted load balancing based on performance
type WeightedStrategy struct {
	metricsRegistry *MetricsRegistry
}

// HealthBasedStrategy implements health-based routing
type HealthBasedStrategy struct {
	circuitBreakers *CircuitBreakerPattern
	providerRegistry interface {
		GetCircuitBreaker(name string) *CircuitBreaker
	}
}

// LatencyBasedStrategy implements latency-based routing
type LatencyBasedStrategy struct {
	metricsRegistry *MetricsRegistry
}

// RandomStrategy implements random provider selection
type RandomStrategy struct{}

func NewRequestService(strategy string, ensemble *EnsembleService, memory *MemoryService) *RequestService {
	var routingStrategy RoutingStrategy

	switch strategy {
	case "round_robin":
		routingStrategy = &RoundRobinStrategy{}
	case "weighted":
		routingStrategy = &WeightedStrategy{
			metricsRegistry: GlobalMetricsRegistry,
		}
	case "health_based":
		routingStrategy = &HealthBasedStrategy{
			circuitBreakers: NewCircuitBreakerPattern(),
		}
	case "latency_based":
		routingStrategy = &LatencyBasedStrategy{
			metricsRegistry: GlobalMetricsRegistry,
		}
	case "random":
		routingStrategy = &RandomStrategy{}
	default:
		routingStrategy = &WeightedStrategy{
			metricsRegistry: GlobalMetricsRegistry,
		} // Default
	}

	return &RequestService{
		providers: make(map[string]LLMProvider),
		ensemble:  ensemble,
		memory:    memory,
		strategy:  routingStrategy,
	}
}

func (r *RequestService) RegisterProvider(name string, provider LLMProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

func (r *RequestService) RemoveProvider(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
}

func (r *RequestService) GetProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

func (r *RequestService) ProcessRequest(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	r.mu.RLock()
	providers := make(map[string]LLMProvider)
	for k, v := range r.providers {
		providers[k] = v
	}
	r.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Enhance request with memory if enabled
	if r.memory != nil && req.MemoryEnhanced {
		if err := r.memory.EnhanceRequest(ctx, req); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Memory enhancement failed: %v\n", err)
		}
	}

	// Check if ensemble is requested and we have multiple providers
	if req.EnsembleConfig != nil && len(providers) >= req.EnsembleConfig.MinProviders {
		result, err := r.ensemble.RunEnsemble(ctx, req)
		if err != nil {
			// Fall back to single provider if ensemble fails
			return r.processSingleProvider(ctx, req, providers)
		}
		return result.Selected, nil
	}

	// Process with single provider
	return r.processSingleProvider(ctx, req, providers)
}

func (r *RequestService) ProcessRequestStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	r.mu.RLock()
	providers := make(map[string]LLMProvider)
	for k, v := range r.providers {
		providers[k] = v
	}
	r.mu.RUnlock()

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Enhance request with memory if enabled
	if r.memory != nil && req.MemoryEnhanced {
		if err := r.memory.EnhanceRequest(ctx, req); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Memory enhancement failed: %v\n", err)
		}
	}

	// Check if ensemble is requested and we have multiple providers
	if req.EnsembleConfig != nil && len(providers) >= req.EnsembleConfig.MinProviders {
		return r.ensemble.RunEnsembleStream(ctx, req)
	}

	// Process with single provider
	return r.processSingleProviderStream(ctx, req, providers)
}

func (r *RequestService) processSingleProvider(ctx context.Context, req *models.LLMRequest, providers map[string]LLMProvider) (*models.LLMResponse, error) {
	// Select provider based on routing strategy
	providerName, err := r.strategy.SelectProvider(providers, req)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	provider, exists := providers[providerName]
	if !exists {
		return nil, fmt.Errorf("selected provider %s not found", providerName)
	}

	// Track request timing for metrics
	startTime := time.Now()

	// Execute request
	resp, err := provider.Complete(ctx, req)
	latencyMs := time.Since(startTime).Milliseconds()

	// Record metrics
	GlobalMetricsRegistry.RecordRequest(providerName, err == nil, latencyMs)

	// Update circuit breaker for health-based strategy
	if hbs, ok := r.strategy.(*HealthBasedStrategy); ok && hbs.circuitBreakers != nil {
		cb := hbs.circuitBreakers.GetCircuitBreaker(providerName)
		if err != nil {
			cb.mu.Lock()
			cb.FailureCount++
			cb.LastFailTime = time.Now()
			if cb.FailureCount >= cb.FailureThreshold {
				cb.State = RequestStateOpen
			}
			cb.mu.Unlock()
		} else {
			cb.mu.Lock()
			cb.SuccessCount++
			if cb.State == RequestStateHalfOpen {
				cb.State = RequestStateClosed
				cb.FailureCount = 0
			}
			cb.mu.Unlock()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("provider %s failed: %w", providerName, err)
	}

	// Add provider metadata
	resp.ProviderID = providerName
	resp.ProviderName = providerName

	return resp, nil
}

func (r *RequestService) processSingleProviderStream(ctx context.Context, req *models.LLMRequest, providers map[string]LLMProvider) (<-chan *models.LLMResponse, error) {
	// Select provider based on routing strategy
	providerName, err := r.strategy.SelectProvider(providers, req)
	if err != nil {
		return nil, fmt.Errorf("failed to select provider: %w", err)
	}

	provider, exists := providers[providerName]
	if !exists {
		return nil, fmt.Errorf("selected provider %s not found", providerName)
	}

	// Track request timing for metrics
	startTime := time.Now()

	// Execute streaming request
	streamChan, err := provider.CompleteStream(ctx, req)
	if err != nil {
		latencyMs := time.Since(startTime).Milliseconds()
		GlobalMetricsRegistry.RecordRequest(providerName, false, latencyMs)
		return nil, fmt.Errorf("provider %s failed: %w", providerName, err)
	}

	// Wrap responses with provider info and record metrics on completion
	wrappedChan := make(chan *models.LLMResponse)
	go func() {
		defer close(wrappedChan)
		hasResponses := false
		for resp := range streamChan {
			hasResponses = true
			resp.ProviderID = providerName
			resp.ProviderName = providerName
			wrappedChan <- resp
		}
		// Record metrics after stream completes
		latencyMs := time.Since(startTime).Milliseconds()
		GlobalMetricsRegistry.RecordRequest(providerName, hasResponses, latencyMs)
	}()

	return wrappedChan, nil
}

// Routing strategy implementations

// RoundRobinStrategy
func (s *RoundRobinStrategy) SelectProvider(providers map[string]LLMProvider, req *models.LLMRequest) (string, error) {
	if len(providers) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}

	selected := names[s.counter%int64(len(names))]
	s.counter++
	return selected, nil
}

// WeightedStrategy
func (s *WeightedStrategy) SelectProvider(providers map[string]LLMProvider, req *models.LLMRequest) (string, error) {
	if len(providers) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	// Get the metrics registry (use global if not set)
	registry := s.metricsRegistry
	if registry == nil {
		registry = GlobalMetricsRegistry
	}

	// Calculate dynamic weights based on performance metrics
	weights := make(map[string]float64)
	totalWeight := 0.0

	for name := range providers {
		metrics := registry.GetMetrics(name)

		// Base weight calculation:
		// - Success rate contributes 60% of the weight
		// - Inverse latency contributes 40% (faster = higher weight)
		successRate := metrics.GetSuccessRate()
		avgLatency := metrics.GetAverageLatency()

		// Normalize latency: lower latency = higher score (max 1.0 for latency < 100ms)
		latencyScore := 1.0
		if avgLatency > 0 {
			latencyScore = math.Min(1.0, 1000.0/avgLatency) // 1000ms baseline
		}

		// Calculate composite weight
		weight := (successRate * 0.6) + (latencyScore * 0.4)

		// Ensure minimum weight of 0.1 to give underperforming providers a chance
		weight = math.Max(0.1, weight)

		// Apply preference weights from ensemble config
		if req != nil && req.EnsembleConfig != nil {
			for i, preferred := range req.EnsembleConfig.PreferredProviders {
				if name == preferred {
					// Boost preferred providers by 50-100% based on position
					weight *= 1.5 + (0.5 * (1.0 - float64(i)/float64(len(req.EnsembleConfig.PreferredProviders))))
					break
				}
			}
		}

		weights[name] = weight
		totalWeight += weight
	}

	// Select based on weighted random selection
	random := rand.Float64() * totalWeight
	current := 0.0

	for name, weight := range weights {
		current += weight
		if random <= current {
			return name, nil
		}
	}

	// Fallback to first provider
	for name := range providers {
		return name, nil
	}

	return "", fmt.Errorf("failed to select provider")
}

// HealthBasedStrategy
func (s *HealthBasedStrategy) SelectProvider(providers map[string]LLMProvider, req *models.LLMRequest) (string, error) {
	if len(providers) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	// Get circuit breaker pattern (create if not set)
	cbPattern := s.circuitBreakers
	if cbPattern == nil {
		cbPattern = NewCircuitBreakerPattern()
	}

	// Filter healthy providers based on circuit breaker state and health metrics
	healthyProviders := make([]string, 0)
	halfOpenProviders := make([]string, 0)

	for name := range providers {
		cb := cbPattern.GetCircuitBreaker(name)

		cb.mu.RLock()
		state := cb.State
		lastFailTime := cb.LastFailTime
		recoveryTimeout := cb.RecoveryTimeout
		cb.mu.RUnlock()

		switch state {
		case RequestStateClosed:
			// Provider is healthy
			healthyProviders = append(healthyProviders, name)
		case RequestStateHalfOpen:
			// Provider is recovering, give it a chance
			halfOpenProviders = append(halfOpenProviders, name)
		case RequestStateOpen:
			// Check if enough time has passed to try again
			if time.Since(lastFailTime) >= recoveryTimeout {
				halfOpenProviders = append(halfOpenProviders, name)
			}
			// Otherwise, skip this provider
		}
	}

	// Prefer fully healthy providers
	if len(healthyProviders) > 0 {
		// Select based on success rate from metrics
		registry := GlobalMetricsRegistry
		bestProvider := healthyProviders[0]
		bestScore := 0.0

		for _, name := range healthyProviders {
			metrics := registry.GetMetrics(name)
			score := metrics.GetSuccessRate()
			if score > bestScore {
				bestScore = score
				bestProvider = name
			}
		}
		return bestProvider, nil
	}

	// Fall back to half-open providers if no fully healthy ones
	if len(halfOpenProviders) > 0 {
		// Select randomly among recovering providers
		return halfOpenProviders[rand.Intn(len(halfOpenProviders))], nil
	}

	return "", fmt.Errorf("no healthy providers available")
}

// SetCircuitBreakers sets the circuit breaker pattern for this strategy
func (s *HealthBasedStrategy) SetCircuitBreakers(cb *CircuitBreakerPattern) {
	s.circuitBreakers = cb
}

// LatencyBasedStrategy
func (s *LatencyBasedStrategy) SelectProvider(providers map[string]LLMProvider, req *models.LLMRequest) (string, error) {
	if len(providers) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	// Get the metrics registry (use global if not set)
	registry := s.metricsRegistry
	if registry == nil {
		registry = GlobalMetricsRegistry
	}

	// Find provider with lowest average latency
	var bestProvider string
	lowestLatency := math.MaxFloat64

	// Collect providers with metrics
	providersWithMetrics := make([]string, 0)
	providersWithoutMetrics := make([]string, 0)

	for name := range providers {
		metrics := registry.GetMetrics(name)

		metrics.mu.RLock()
		hasHistory := len(metrics.LatencyHistory) > 0
		metrics.mu.RUnlock()

		if hasHistory {
			avgLatency := metrics.GetAverageLatency()
			providersWithMetrics = append(providersWithMetrics, name)

			if avgLatency < lowestLatency {
				lowestLatency = avgLatency
				bestProvider = name
			}
		} else {
			// Provider has no latency history yet
			providersWithoutMetrics = append(providersWithoutMetrics, name)
		}
	}

	// If we found a provider with the lowest latency, use it (with some randomization to allow exploration)
	if bestProvider != "" {
		// 10% of the time, pick a random provider to allow exploration
		if rand.Float64() < 0.1 {
			// Pick from all providers for exploration
			names := make([]string, 0, len(providers))
			for name := range providers {
				names = append(names, name)
			}
			return names[rand.Intn(len(names))], nil
		}
		return bestProvider, nil
	}

	// No providers with latency data yet, prefer those without metrics to build up data
	if len(providersWithoutMetrics) > 0 {
		return providersWithoutMetrics[rand.Intn(len(providersWithoutMetrics))], nil
	}

	// Fallback: select randomly from all providers
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names[rand.Intn(len(names))], nil
}

// RandomStrategy
func (s *RandomStrategy) SelectProvider(providers map[string]LLMProvider, req *models.LLMRequest) (string, error) {
	if len(providers) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	selected := names[rand.Intn(len(names))]
	return selected, nil
}

// ProviderHealth management

func (r *RequestService) UpdateProviderHealth(name string, healthy bool, responseTime int64, err error) {
	// This would be used to track provider health and performance
	// Implementation would involve maintaining a health registry
}

func (r *RequestService) GetProviderHealth(name string) *ProviderHealth {
	// Return health information for a specific provider
	// Implementation would query the health registry
	return &ProviderHealth{
		Name:         name,
		Healthy:      true,
		LastCheck:    time.Now(),
		ResponseTime: 1000,
		SuccessRate:  0.95,
	}
}

func (r *RequestService) GetAllProviderHealth() map[string]*ProviderHealth {
	// Return health information for all providers
	health := make(map[string]*ProviderHealth)
	r.mu.RLock()
	for name := range r.providers {
		health[name] = &ProviderHealth{
			Name:         name,
			Healthy:      true,
			LastCheck:    time.Now(),
			ResponseTime: 1000,
			SuccessRate:  0.95,
		}
	}
	r.mu.RUnlock()
	return health
}

// Advanced routing features

// CircuitBreakerPattern implements circuit breaker for failing providers
type CircuitBreakerPattern struct {
	providers map[string]*RequestCircuitBreaker
	mu        sync.RWMutex
}

type RequestCircuitBreaker struct {
	Name             string
	State            RequestCircuitState
	FailureCount     int64
	LastFailTime     time.Time
	SuccessCount     int64
	Timeout          time.Duration
	FailureThreshold int64
	RecoveryTimeout  time.Duration
	mu               sync.RWMutex
}

type RequestCircuitState int

const (
	RequestStateClosed RequestCircuitState = iota
	RequestStateOpen
	RequestStateHalfOpen
)

func NewCircuitBreakerPattern() *CircuitBreakerPattern {
	return &CircuitBreakerPattern{
		providers: make(map[string]*RequestCircuitBreaker),
	}
}

func (c *CircuitBreakerPattern) GetCircuitBreaker(name string) *RequestCircuitBreaker {
	// First try with read lock
	c.mu.RLock()
	cb, exists := c.providers[name]
	c.mu.RUnlock()

	if exists {
		return cb
	}

	// Need to create new circuit breaker - acquire write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have created it)
	if cb, exists = c.providers[name]; exists {
		return cb
	}

	cb = &RequestCircuitBreaker{
		Name:             name,
		State:            RequestStateClosed,
		FailureThreshold: 5,
		Timeout:          60 * time.Second,
		RecoveryTimeout:  30 * time.Second,
	}
	c.providers[name] = cb

	return cb
}

func (cb *RequestCircuitBreaker) Call(ctx context.Context, operation func() (*models.LLMResponse, error)) (*models.LLMResponse, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.State {
	case RequestStateOpen:
		if time.Since(cb.LastFailTime) > cb.RecoveryTimeout {
			cb.State = RequestStateHalfOpen
		} else {
			return nil, fmt.Errorf("circuit breaker is open for provider %s", cb.Name)
		}
	case RequestStateHalfOpen:
		// Allow one request through
		resp, err := operation()
		if err != nil {
			cb.FailureCount++
			cb.LastFailTime = time.Now()
			cb.State = RequestStateOpen
			return resp, err
		}
		cb.SuccessCount++
		cb.State = RequestStateClosed
		return resp, nil
	case RequestStateClosed:
		// Normal operation
		resp, err := operation()
		if err != nil {
			cb.FailureCount++
			if cb.FailureCount >= cb.FailureThreshold {
				cb.State = RequestStateOpen
				cb.LastFailTime = time.Now()
			}
			return resp, err
		}
		cb.SuccessCount++
		return resp, nil
	}

	return nil, fmt.Errorf("unknown circuit breaker state")
}

// RetryPattern implements retry logic with exponential backoff
type RetryPattern struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

func NewRetryPattern(maxRetries int, initialDelay, maxDelay time.Duration, backoffFactor float64) *RetryPattern {
	return &RetryPattern{
		MaxRetries:    maxRetries,
		InitialDelay:  initialDelay,
		MaxDelay:      maxDelay,
		BackoffFactor: backoffFactor,
	}
}

func (r *RetryPattern) Execute(ctx context.Context, operation func() (*models.LLMResponse, error)) (*models.LLMResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		resp, err := operation()
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't wait on the last attempt
		if attempt < r.MaxRetries {
			delay := time.Duration(float64(r.InitialDelay) * math.Pow(r.BackoffFactor, float64(attempt)))
			if delay > r.MaxDelay {
				delay = r.MaxDelay
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return nil, lastErr
}
