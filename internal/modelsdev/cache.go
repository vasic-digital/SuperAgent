package modelsdev

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Cache provides in-memory caching for Models.dev data
type Cache struct {
	models          map[string]*CachedModel
	providers       map[string]*CachedProvider
	modelsByProvider map[string][]string // provider ID -> model IDs

	mu              sync.RWMutex
	config          CacheConfig

	// Statistics
	hits            int64
	misses          int64
	lastRefresh     time.Time

	// Cleanup management
	stopCleanup     chan struct{}
	cleanupDone     chan struct{}
}

// NewCache creates a new cache instance
func NewCache(config *CacheConfig) *Cache {
	if config == nil {
		defaultConfig := DefaultCacheConfig()
		config = &defaultConfig
	}

	// Ensure CleanupInterval has a valid value
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 10 * time.Minute
	}

	c := &Cache{
		models:           make(map[string]*CachedModel),
		providers:        make(map[string]*CachedProvider),
		modelsByProvider: make(map[string][]string),
		config:           *config,
		stopCleanup:      make(chan struct{}),
		cleanupDone:      make(chan struct{}),
	}

	// Start background cleanup goroutine
	go c.cleanupLoop()

	return c
}

// GetModel retrieves a model from cache
func (c *Cache) GetModel(ctx context.Context, modelID string) (*Model, bool) {
	c.mu.RLock()
	cached, exists := c.models[modelID]
	c.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	if cached.IsExpired() {
		atomic.AddInt64(&c.misses, 1)
		// Async cleanup of expired entry
		go c.removeExpiredModel(modelID)
		return nil, false
	}

	atomic.AddInt64(&c.hits, 1)
	atomic.AddInt64(&cached.HitCount, 1)
	return cached.Model, true
}

// SetModel stores a model in cache
func (c *Cache) SetModel(ctx context.Context, model *Model) {
	if model == nil || model.ID == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.models) >= c.config.MaxModels {
		c.evictOldestModels(len(c.models) - c.config.MaxModels + 1)
	}

	now := time.Now()
	c.models[model.ID] = &CachedModel{
		Model:     model,
		CachedAt:  now,
		ExpiresAt: now.Add(c.config.ModelTTL),
		HitCount:  0,
	}

	// Update provider index
	if model.Provider != "" {
		c.modelsByProvider[model.Provider] = appendIfMissing(c.modelsByProvider[model.Provider], model.ID)
	}
}

// SetModels stores multiple models in cache
func (c *Cache) SetModels(ctx context.Context, models []Model) {
	if len(models) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(c.config.ModelTTL)

	for i := range models {
		model := &models[i]
		if model.ID == "" {
			continue
		}

		// Check capacity and evict if needed
		if len(c.models) >= c.config.MaxModels {
			c.evictOldestModels(1)
		}

		c.models[model.ID] = &CachedModel{
			Model:     model,
			CachedAt:  now,
			ExpiresAt: expiresAt,
			HitCount:  0,
		}

		// Update provider index
		if model.Provider != "" {
			c.modelsByProvider[model.Provider] = appendIfMissing(c.modelsByProvider[model.Provider], model.ID)
		}
	}
}

// GetProvider retrieves a provider from cache
func (c *Cache) GetProvider(ctx context.Context, providerID string) (*Provider, bool) {
	c.mu.RLock()
	cached, exists := c.providers[providerID]
	c.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	if cached.IsExpired() {
		atomic.AddInt64(&c.misses, 1)
		go c.removeExpiredProvider(providerID)
		return nil, false
	}

	atomic.AddInt64(&c.hits, 1)
	atomic.AddInt64(&cached.HitCount, 1)
	return cached.Provider, true
}

// SetProvider stores a provider in cache
func (c *Cache) SetProvider(ctx context.Context, provider *Provider) {
	if provider == nil || provider.ID == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.providers) >= c.config.MaxProviders {
		c.evictOldestProviders(len(c.providers) - c.config.MaxProviders + 1)
	}

	now := time.Now()
	c.providers[provider.ID] = &CachedProvider{
		Provider:  provider,
		CachedAt:  now,
		ExpiresAt: now.Add(c.config.ProviderTTL),
		HitCount:  0,
	}
}

// SetProviders stores multiple providers in cache
func (c *Cache) SetProviders(ctx context.Context, providers []Provider) {
	if len(providers) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(c.config.ProviderTTL)

	for i := range providers {
		provider := &providers[i]
		if provider.ID == "" {
			continue
		}

		// Check capacity and evict if needed
		if len(c.providers) >= c.config.MaxProviders {
			c.evictOldestProviders(1)
		}

		c.providers[provider.ID] = &CachedProvider{
			Provider:  provider,
			CachedAt:  now,
			ExpiresAt: expiresAt,
			HitCount:  0,
		}
	}
}

// GetModelsByProvider returns all cached models for a provider
func (c *Cache) GetModelsByProvider(ctx context.Context, providerID string) ([]*Model, bool) {
	c.mu.RLock()
	modelIDs, exists := c.modelsByProvider[providerID]
	if !exists || len(modelIDs) == 0 {
		c.mu.RUnlock()
		return nil, false
	}

	models := make([]*Model, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		if cached, ok := c.models[modelID]; ok && !cached.IsExpired() {
			models = append(models, cached.Model)
		}
	}
	c.mu.RUnlock()

	if len(models) == 0 {
		return nil, false
	}

	return models, true
}

// GetAllModels returns all cached models
func (c *Cache) GetAllModels(ctx context.Context) []*Model {
	c.mu.RLock()
	defer c.mu.RUnlock()

	models := make([]*Model, 0, len(c.models))
	for _, cached := range c.models {
		if !cached.IsExpired() {
			models = append(models, cached.Model)
		}
	}
	return models
}

// GetAllProviders returns all cached providers
func (c *Cache) GetAllProviders(ctx context.Context) []*Provider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providers := make([]*Provider, 0, len(c.providers))
	for _, cached := range c.providers {
		if !cached.IsExpired() {
			providers = append(providers, cached.Provider)
		}
	}
	return providers
}

// InvalidateModel removes a model from cache
func (c *Cache) InvalidateModel(ctx context.Context, modelID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, exists := c.models[modelID]; exists {
		// Remove from provider index
		if cached.Model.Provider != "" {
			c.modelsByProvider[cached.Model.Provider] = removeString(c.modelsByProvider[cached.Model.Provider], modelID)
		}
		delete(c.models, modelID)
	}
}

// InvalidateProvider removes a provider and its models from cache
func (c *Cache) InvalidateProvider(ctx context.Context, providerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.providers, providerID)

	// Remove all models for this provider
	if modelIDs, exists := c.modelsByProvider[providerID]; exists {
		for _, modelID := range modelIDs {
			delete(c.models, modelID)
		}
		delete(c.modelsByProvider, providerID)
	}
}

// InvalidateAll clears the entire cache
func (c *Cache) InvalidateAll(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.models = make(map[string]*CachedModel)
	c.providers = make(map[string]*CachedProvider)
	c.modelsByProvider = make(map[string][]string)
}

// UpdateLastRefresh updates the last refresh timestamp
func (c *Cache) UpdateLastRefresh() {
	c.mu.Lock()
	c.lastRefresh = time.Now()
	c.mu.Unlock()
}

// Stats returns current cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)

	var hitRate float64
	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	var oldestEntry time.Time
	for _, cached := range c.models {
		if oldestEntry.IsZero() || cached.CachedAt.Before(oldestEntry) {
			oldestEntry = cached.CachedAt
		}
	}
	for _, cached := range c.providers {
		if oldestEntry.IsZero() || cached.CachedAt.Before(oldestEntry) {
			oldestEntry = cached.CachedAt
		}
	}

	return CacheStats{
		ModelCount:       len(c.models),
		ProviderCount:    len(c.providers),
		TotalHits:        hits,
		TotalMisses:      misses,
		HitRate:          hitRate,
		LastRefresh:      c.lastRefresh,
		OldestEntry:      oldestEntry,
		MemoryUsageBytes: c.estimateMemoryUsage(),
	}
}

// Close stops the cache cleanup goroutine
func (c *Cache) Close() error {
	close(c.stopCleanup)
	<-c.cleanupDone
	return nil
}

// Internal methods

func (c *Cache) cleanupLoop() {
	defer close(c.cleanupDone)

	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCleanup:
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Clean up expired models
	for modelID, cached := range c.models {
		if now.After(cached.ExpiresAt) {
			if cached.Model.Provider != "" {
				c.modelsByProvider[cached.Model.Provider] = removeString(c.modelsByProvider[cached.Model.Provider], modelID)
			}
			delete(c.models, modelID)
		}
	}

	// Clean up expired providers
	for providerID, cached := range c.providers {
		if now.After(cached.ExpiresAt) {
			delete(c.providers, providerID)
		}
	}

	// Clean up empty provider indices
	for providerID, modelIDs := range c.modelsByProvider {
		if len(modelIDs) == 0 {
			delete(c.modelsByProvider, providerID)
		}
	}
}

func (c *Cache) removeExpiredModel(modelID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, exists := c.models[modelID]; exists && cached.IsExpired() {
		if cached.Model.Provider != "" {
			c.modelsByProvider[cached.Model.Provider] = removeString(c.modelsByProvider[cached.Model.Provider], modelID)
		}
		delete(c.models, modelID)
	}
}

func (c *Cache) removeExpiredProvider(providerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, exists := c.providers[providerID]; exists && cached.IsExpired() {
		delete(c.providers, providerID)
	}
}

func (c *Cache) evictOldestModels(count int) {
	// Simple LRU eviction based on CachedAt time
	type modelEntry struct {
		id       string
		cachedAt time.Time
	}

	entries := make([]modelEntry, 0, len(c.models))
	for id, cached := range c.models {
		entries = append(entries, modelEntry{id: id, cachedAt: cached.CachedAt})
	}

	// Sort by CachedAt (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].cachedAt.Before(entries[i].cachedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		modelID := entries[i].id
		if cached, exists := c.models[modelID]; exists {
			if cached.Model.Provider != "" {
				c.modelsByProvider[cached.Model.Provider] = removeString(c.modelsByProvider[cached.Model.Provider], modelID)
			}
			delete(c.models, modelID)
		}
	}
}

func (c *Cache) evictOldestProviders(count int) {
	type providerEntry struct {
		id       string
		cachedAt time.Time
	}

	entries := make([]providerEntry, 0, len(c.providers))
	for id, cached := range c.providers {
		entries = append(entries, providerEntry{id: id, cachedAt: cached.CachedAt})
	}

	// Sort by CachedAt (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].cachedAt.Before(entries[i].cachedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		delete(c.providers, entries[i].id)
	}
}

func (c *Cache) estimateMemoryUsage() int64 {
	// Rough estimate: ~500 bytes per model, ~200 bytes per provider
	return int64(len(c.models)*500 + len(c.providers)*200)
}

// Helper functions

func appendIfMissing(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func removeString(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
