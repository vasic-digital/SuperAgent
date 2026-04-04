package providers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	require.NotNil(t, registry)
	assert.NotNil(t, registry.providers)
	assert.Empty(t, registry.providers)
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	provider := NewFileProvider("/tmp")
	
	registry.Register(provider)
	
	assert.Len(t, registry.providers, 1)
	assert.Contains(t, registry.providers, "file")
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	provider := NewFileProvider("/tmp")
	
	registry.Register(provider)
	
	p, ok := registry.Get("file")
	assert.True(t, ok)
	assert.NotNil(t, p)
	
	_, ok = registry.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()
	registry.Register(NewFileProvider("/tmp"))
	registry.Register(NewURLProvider())
	
	names := registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "file")
	assert.Contains(t, names, "url")
}

func TestRegistry_Resolve(t *testing.T) {
	registry := NewRegistry()
	
	// Add mock providers
	registry.Register(&mockProvider{
		name: "mock1",
		items: []ContextItem{
			{Name: "item1", Content: "content1"},
		},
	})
	registry.Register(&mockProvider{
		name: "mock2",
		items: []ContextItem{
			{Name: "item2", Content: "content2"},
		},
	})
	
	ctx := context.Background()
	items, err := registry.Resolve(ctx, "query")
	
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestRegistry_ResolveWithProvider(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockProvider{
		name: "mock",
		items: []ContextItem{
			{Name: "item", Content: "content"},
		},
	})
	
	ctx := context.Background()
	items, err := registry.ResolveWithProvider(ctx, "mock", "query")
	
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestRegistry_ResolveWithProvider_NotFound(t *testing.T) {
	registry := NewRegistry()
	
	ctx := context.Background()
	_, err := registry.ResolveWithProvider(ctx, "nonexistent", "query")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider not found")
}

func TestNewManager(t *testing.T) {
	manager := NewManager(time.Minute)
	
	require.NotNil(t, manager)
	assert.NotNil(t, manager.registry)
	assert.NotNil(t, manager.cache)
	assert.Equal(t, time.Minute, manager.ttl)
}

func TestNewManager_DefaultTTL(t *testing.T) {
	manager := NewManager(0)
	
	assert.Equal(t, 5*time.Minute, manager.ttl)
}

func TestManager_Register(t *testing.T) {
	manager := NewManager(0)
	provider := NewFileProvider("/tmp")
	
	manager.Register(provider)
	
	_, ok := manager.registry.Get("file")
	assert.True(t, ok)
}

func TestManager_Resolve_WithCache(t *testing.T) {
	manager := NewManager(time.Hour)
	manager.Register(&mockProvider{
		name: "mock",
		items: []ContextItem{
			{Name: "item", Content: "content"},
		},
	})
	
	ctx := context.Background()
	
	// First call should hit provider
	items1, err := manager.Resolve(ctx, "query")
	require.NoError(t, err)
	assert.Len(t, items1, 1)
	
	// Second call should hit cache
	items2, err := manager.Resolve(ctx, "query")
	require.NoError(t, err)
	assert.Equal(t, items1, items2)
}

func TestManager_ClearCache(t *testing.T) {
	manager := NewManager(time.Hour)
	manager.cache["query"] = cachedItems{
		items:     []ContextItem{{Name: "item"}},
		timestamp: time.Now(),
	}
	
	manager.ClearCache()
	
	assert.Empty(t, manager.cache)
}

func TestManager_ClearCacheFor(t *testing.T) {
	manager := NewManager(time.Hour)
	manager.cache["query1"] = cachedItems{items: []ContextItem{{}}}
	manager.cache["query2"] = cachedItems{items: []ContextItem{{}}}
	
	manager.ClearCacheFor("query1")
	
	assert.NotContains(t, manager.cache, "query1")
	assert.Contains(t, manager.cache, "query2")
}

func TestFormatContext(t *testing.T) {
	items := []ContextItem{
		{
			Name:        "file.go",
			Description: "Go source file",
			Content:     "package main",
			Source:      "/path/file.go",
		},
	}
	
	result := FormatContext(items)
	
	assert.Contains(t, result, "file.go")
	assert.Contains(t, result, "package main")
	assert.Contains(t, result, "/path/file.go")
}

func TestFormatContext_Empty(t *testing.T) {
	result := FormatContext([]ContextItem{})
	assert.Equal(t, "", result)
}

func TestFilterByScore(t *testing.T) {
	items := []ContextItem{
		{Name: "high", Score: 0.9},
		{Name: "medium", Score: 0.6},
		{Name: "low", Score: 0.3},
	}
	
	filtered := FilterByScore(items, 0.5)
	
	assert.Len(t, filtered, 2)
	assert.Equal(t, "high", filtered[0].Name)
	assert.Equal(t, "medium", filtered[1].Name)
}

func TestLimit(t *testing.T) {
	items := []ContextItem{
		{Name: "1"},
		{Name: "2"},
		{Name: "3"},
		{Name: "4"},
	}
	
	limited := Limit(items, 2)
	
	assert.Len(t, limited, 2)
	assert.Equal(t, "1", limited[0].Name)
	assert.Equal(t, "2", limited[1].Name)
}

func TestLimit_UnderLimit(t *testing.T) {
	items := []ContextItem{
		{Name: "1"},
		{Name: "2"},
	}
	
	limited := Limit(items, 5)
	
	assert.Len(t, limited, 2)
}

func TestSortByScore(t *testing.T) {
	items := []ContextItem{
		{Name: "low", Score: 0.3},
		{Name: "high", Score: 0.9},
		{Name: "medium", Score: 0.6},
	}
	
	SortByScore(items)
	
	assert.Equal(t, "high", items[0].Name)
	assert.Equal(t, "medium", items[1].Name)
	assert.Equal(t, "low", items[2].Name)
}

func TestCombine(t *testing.T) {
	items := []ContextItem{
		{Content: "part1"},
		{Content: "part2"},
	}
	
	combined := Combine(items, "combined")
	
	assert.Equal(t, "combined", combined.Name)
	assert.Contains(t, combined.Content, "part1")
	assert.Contains(t, combined.Content, "part2")
	assert.Equal(t, "combined", combined.Source)
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry("/tmp")
	
	names := registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "file")
	assert.Contains(t, names, "url")
}

// mockProvider is a test provider
type mockProvider struct {
	name  string
	items []ContextItem
	err   error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Description() string {
	return "mock provider"
}

func (m *mockProvider) Resolve(ctx context.Context, query string) ([]ContextItem, error) {
	return m.items, m.err
}
