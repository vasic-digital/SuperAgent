// Package providers provides context providers for HelixAgent
// Inspired by Continue's context provider system
package providers

import (
	"context"
	"fmt"
	"time"
)

// ContextItem represents a single context item
type ContextItem struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Source      string    `json:"source"`
	Timestamp   time.Time `json:"timestamp"`
	Score       float64   `json:"score,omitempty"` // Relevance score
}

// Provider interface for context providers
type Provider interface {
	Name() string
	Description() string
	Resolve(ctx context.Context, query string) ([]ContextItem, error)
}

// Registry manages context providers
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider
func (r *Registry) Register(provider Provider) {
	r.providers[provider.Name()] = provider
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

// List returns all registered provider names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Resolve resolves context from all providers matching the query
func (r *Registry) Resolve(ctx context.Context, query string) ([]ContextItem, error) {
	var allItems []ContextItem
	
	for _, provider := range r.providers {
		items, err := provider.Resolve(ctx, query)
		if err != nil {
			// Log error but continue with other providers
			continue
		}
		allItems = append(allItems, items...)
	}
	
	return allItems, nil
}

// ResolveWithProvider resolves context from a specific provider
func (r *Registry) ResolveWithProvider(ctx context.Context, providerName, query string) ([]ContextItem, error) {
	provider, ok := r.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}
	
	return provider.Resolve(ctx, query)
}

// Manager manages context resolution and caching
type Manager struct {
	registry *Registry
	cache    map[string]cachedItems
	ttl      time.Duration
}

type cachedItems struct {
	items     []ContextItem
	timestamp time.Time
}

// NewManager creates a new context manager
func NewManager(ttl time.Duration) *Manager {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	
	return &Manager{
		registry: NewRegistry(),
		cache:    make(map[string]cachedItems),
		ttl:      ttl,
	}
}

// Register registers a provider
func (m *Manager) Register(provider Provider) {
	m.registry.Register(provider)
}

// GetRegistry returns the provider registry
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// Resolve resolves context with caching
func (m *Manager) Resolve(ctx context.Context, query string) ([]ContextItem, error) {
	// Check cache
	if cached, ok := m.cache[query]; ok {
		if time.Since(cached.timestamp) < m.ttl {
			return cached.items, nil
		}
	}
	
	// Resolve from providers
	items, err := m.registry.Resolve(ctx, query)
	if err != nil {
		return nil, err
	}
	
	// Cache results
	m.cache[query] = cachedItems{
		items:     items,
		timestamp: time.Now(),
	}
	
	return items, nil
}

// ClearCache clears the context cache
func (m *Manager) ClearCache() {
	m.cache = make(map[string]cachedItems)
}

// ClearCacheFor clears cache for a specific query
func (m *Manager) ClearCacheFor(query string) {
	delete(m.cache, query)
}

// FormatContext formats context items for LLM consumption
func FormatContext(items []ContextItem) string {
	if len(items) == 0 {
		return ""
	}
	
	var result string
	for _, item := range items {
		result += fmt.Sprintf("\n## %s\n", item.Name)
		if item.Description != "" {
			result += fmt.Sprintf("*%s*\n", item.Description)
		}
		result += fmt.Sprintf("Source: %s\n", item.Source)
		result += fmt.Sprintf("```\n%s\n```\n", item.Content)
	}
	
	return result
}

// FilterByScore filters context items by minimum score
func FilterByScore(items []ContextItem, minScore float64) []ContextItem {
	var filtered []ContextItem
	for _, item := range items {
		if item.Score >= minScore {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Limit limits the number of context items
func Limit(items []ContextItem, max int) []ContextItem {
	if len(items) <= max {
		return items
	}
	return items[:max]
}

// SortByScore sorts items by score (highest first)
func SortByScore(items []ContextItem) {
	// Simple bubble sort for small slices
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Score > items[i].Score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// Combine combines multiple context items into one
func Combine(items []ContextItem, name string) ContextItem {
	var content string
	for _, item := range items {
		content += item.Content + "\n\n"
	}
	
	return ContextItem{
		Name:      name,
		Content:   content,
		Source:    "combined",
		Timestamp: time.Now(),
	}
}

// DefaultRegistry creates a registry with default providers
func DefaultRegistry(basePath string) *Registry {
	registry := NewRegistry()
	
	// Register default providers
	registry.Register(NewFileProvider(basePath))
	registry.Register(NewURLProvider())
	
	return registry
}
