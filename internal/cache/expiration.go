package cache

import (
	"context"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

// ValidatorFunc is a function that validates cache entries
// Returns true if the entry is still valid, false if it should be expired
type ValidatorFunc func(key string, value interface{}, age time.Duration) bool

// ExpirationConfig holds configuration for the expiration manager
type ExpirationConfig struct {
	// CleanupInterval is how often to run cleanup
	CleanupInterval time.Duration
	// DefaultTTL is the default time-to-live for entries
	DefaultTTL time.Duration
	// MaxAge is the maximum age for any entry
	MaxAge time.Duration
	// EnableValidation enables custom validation on access
	EnableValidation bool
	// ValidationInterval is how often to run full validation
	ValidationInterval time.Duration
}

// DefaultExpirationConfig returns sensible defaults
func DefaultExpirationConfig() *ExpirationConfig {
	return &ExpirationConfig{
		CleanupInterval:    time.Minute,
		DefaultTTL:         30 * time.Minute,
		MaxAge:             24 * time.Hour,
		EnableValidation:   true,
		ValidationInterval: 5 * time.Minute,
	}
}

// ExpirationManager handles TTL, validation, and cleanup for cache entries
type ExpirationManager struct {
	cache       *TieredCache
	config      *ExpirationConfig
	validators  map[string]ValidatorFunc
	validatorMu sync.RWMutex
	metrics     *ExpirationMetrics
	ctx         context.Context
	cancel      context.CancelFunc
}

// ExpirationMetrics tracks expiration manager statistics
type ExpirationMetrics struct {
	ExpiredByTTL        int64
	ExpiredByValidation int64
	ForceExpired        int64
	ValidationRuns      int64
	ValidationErrors    int64
	CleanupRuns         int64
	CleanupDuration     int64 // microseconds
}

// NewExpirationManager creates a new expiration manager
func NewExpirationManager(cache *TieredCache, config *ExpirationConfig) *ExpirationManager {
	if config == nil {
		config = DefaultExpirationConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	em := &ExpirationManager{
		cache:      cache,
		config:     config,
		validators: make(map[string]ValidatorFunc),
		metrics:    &ExpirationMetrics{},
		ctx:        ctx,
		cancel:     cancel,
	}

	return em
}

// Start starts the background cleanup and validation goroutines
func (m *ExpirationManager) Start() {
	go m.cleanupLoop()
	if m.config.EnableValidation {
		go m.validationLoop()
	}
}

// Stop stops the expiration manager
func (m *ExpirationManager) Stop() {
	m.cancel()
}

// RegisterValidator adds a custom validator for keys matching a pattern
// Pattern uses glob-style matching (* for any characters)
func (m *ExpirationManager) RegisterValidator(pattern string, fn ValidatorFunc) {
	m.validatorMu.Lock()
	defer m.validatorMu.Unlock()
	m.validators[pattern] = fn
}

// UnregisterValidator removes a validator
func (m *ExpirationManager) UnregisterValidator(pattern string) {
	m.validatorMu.Lock()
	defer m.validatorMu.Unlock()
	delete(m.validators, pattern)
}

// ValidateEntry checks if a cached entry is still valid
func (m *ExpirationManager) ValidateEntry(key string, value interface{}, age time.Duration) bool {
	if !m.config.EnableValidation {
		return true
	}

	// Check max age
	if age > m.config.MaxAge {
		return false
	}

	// Find matching validator
	m.validatorMu.RLock()
	defer m.validatorMu.RUnlock()

	for pattern, validator := range m.validators {
		if matchPattern(pattern, key) {
			if !validator(key, value, age) {
				atomic.AddInt64(&m.metrics.ExpiredByValidation, 1)
				return false
			}
		}
	}

	return true
}

// ForceExpire immediately expires entries matching the pattern
func (m *ExpirationManager) ForceExpire(ctx context.Context, pattern string) (int, error) {
	count, err := m.cache.InvalidatePrefix(ctx, pattern)
	if err != nil {
		return 0, err
	}

	atomic.AddInt64(&m.metrics.ForceExpired, int64(count))
	return count, nil
}

// ForceExpireByTag immediately expires entries with the given tag
func (m *ExpirationManager) ForceExpireByTag(ctx context.Context, tag string) (int, error) {
	count, err := m.cache.InvalidateByTag(ctx, tag)
	if err != nil {
		return 0, err
	}

	atomic.AddInt64(&m.metrics.ForceExpired, int64(count))
	return count, nil
}

// SetTTL sets a new TTL for an existing entry (by re-setting with new TTL)
func (m *ExpirationManager) SetTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return m.cache.Set(ctx, key, value, ttl)
}

// Metrics returns current metrics
func (m *ExpirationManager) Metrics() *ExpirationMetrics {
	return &ExpirationMetrics{
		ExpiredByTTL:        atomic.LoadInt64(&m.metrics.ExpiredByTTL),
		ExpiredByValidation: atomic.LoadInt64(&m.metrics.ExpiredByValidation),
		ForceExpired:        atomic.LoadInt64(&m.metrics.ForceExpired),
		ValidationRuns:      atomic.LoadInt64(&m.metrics.ValidationRuns),
		ValidationErrors:    atomic.LoadInt64(&m.metrics.ValidationErrors),
		CleanupRuns:         atomic.LoadInt64(&m.metrics.CleanupRuns),
		CleanupDuration:     atomic.LoadInt64(&m.metrics.CleanupDuration),
	}
}

func (m *ExpirationManager) cleanupLoop() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.runCleanup()
		}
	}
}

func (m *ExpirationManager) runCleanup() {
	start := time.Now()
	atomic.AddInt64(&m.metrics.CleanupRuns, 1)

	// The L1 cache handles its own cleanup via the tiered cache
	// L2 (Redis) handles TTL automatically
	// This is primarily for triggering any additional cleanup logic

	duration := time.Since(start).Microseconds()
	atomic.AddInt64(&m.metrics.CleanupDuration, duration)
}

func (m *ExpirationManager) validationLoop() {
	ticker := time.NewTicker(m.config.ValidationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.runValidation()
		}
	}
}

func (m *ExpirationManager) runValidation() {
	atomic.AddInt64(&m.metrics.ValidationRuns, 1)

	// Validation is typically done on access, not proactively
	// This loop can be used for background validation if needed
}

// matchPattern performs simple glob matching
func matchPattern(pattern, s string) bool {
	// Convert glob pattern to regex
	// * matches any characters
	regexPattern := "^" + regexp.QuoteMeta(pattern) + "$"
	regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")

	matched, _ := regexp.MatchString(regexPattern, s)
	return matched
}

// StandardValidators provides common validation functions

// ProviderHealthValidator validates provider health cache entries
func ProviderHealthValidator(maxAge time.Duration) ValidatorFunc {
	return func(key string, value interface{}, age time.Duration) bool {
		// Provider health data should be refreshed frequently
		return age < maxAge
	}
}

// SessionValidator validates session cache entries
func SessionValidator(maxAge time.Duration) ValidatorFunc {
	return func(key string, value interface{}, age time.Duration) bool {
		return age < maxAge
	}
}

// LLMResponseValidator validates LLM response cache entries
func LLMResponseValidator(maxAge time.Duration) ValidatorFunc {
	return func(key string, value interface{}, age time.Duration) bool {
		// LLM responses can be cached longer
		return age < maxAge
	}
}

// MCPResultValidator validates MCP tool result cache entries
func MCPResultValidator(toolSpecificTTLs map[string]time.Duration, defaultTTL time.Duration) ValidatorFunc {
	return func(key string, value interface{}, age time.Duration) bool {
		// Check for tool-specific TTL
		for tool, ttl := range toolSpecificTTLs {
			if matchPattern("*"+tool+"*", key) {
				return age < ttl
			}
		}
		return age < defaultTTL
	}
}

// NeverCacheValidator always returns false (useful for memory/state tools)
func NeverCacheValidator() ValidatorFunc {
	return func(key string, value interface{}, age time.Duration) bool {
		return false
	}
}

// AlwaysValidValidator always returns true (no additional validation)
func AlwaysValidValidator() ValidatorFunc {
	return func(key string, value interface{}, age time.Duration) bool {
		return true
	}
}
