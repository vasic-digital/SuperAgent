package search

import (
	"context"
	"time"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchOptions(t *testing.T) {
	opts := SearchOptions{
		Limit:         10,
		Timeout:       30 * time.Second,
		IncludeAnswer: true,
		RecencyDays:   7,
		SafeSearch:    true,
	}

	assert.Equal(t, 10, opts.Limit)
	assert.Equal(t, 30 * time.Second, opts.Timeout)
	assert.True(t, opts.IncludeAnswer)
	assert.Equal(t, 7, opts.RecencyDays)
	assert.True(t, opts.SafeSearch)
}

func TestSearchResult(t *testing.T) {
	result := &SearchResult{
		Query:      "test query",
		Answer:     "test answer",
		Results:    []SearchItem{},
		TotalCount: 0,
		Provider:   "test",
	}

	assert.Equal(t, "test query", result.Query)
	assert.Equal(t, "test answer", result.Answer)
	assert.Empty(t, result.Results)
	assert.Equal(t, 0, result.TotalCount)
	assert.Equal(t, "test", result.Provider)
}

func TestSearchItem(t *testing.T) {
	item := SearchItem{
		Title:       "Test Title",
		URL:         "https://example.com",
		Snippet:     "Test snippet",
		Content:     "Test content",
		PublishedAt: "2024-01-01",
		Score:       0.95,
	}

	assert.Equal(t, "Test Title", item.Title)
	assert.Equal(t, "https://example.com", item.URL)
	assert.Equal(t, "Test snippet", item.Snippet)
	assert.Equal(t, "Test content", item.Content)
	assert.Equal(t, "2024-01-01", item.PublishedAt)
	assert.Equal(t, 0.95, item.Score)
}

func TestProviderInterface(t *testing.T) {
	// Test that Provider interface can be implemented
	var _ Provider = &mockProvider{}
}

type mockProvider struct{}

func (m *mockProvider) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	return &SearchResult{
		Query:    query,
		Provider: "mock",
		Results:  []SearchItem{},
	}, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}
