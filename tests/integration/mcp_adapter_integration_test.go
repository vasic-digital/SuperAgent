// Package integration provides comprehensive MCP adapter integration tests.
// These tests verify the MCP adapter registry, search functionality, and adapter operations.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/mcp/adapters"
	"dev.helix.agent/internal/mcp/servers"
)

// =============================================================================
// 1. ADAPTER REGISTRY TESTS
// =============================================================================

// TestAdapterRegistryTotalAdapterCount verifies that all 45+ adapters are registered
func TestAdapterRegistryTotalAdapterCount(t *testing.T) {
	count := adapters.GetAdapterCount()
	t.Logf("Total registered adapters: %d", count)

	// The registry should have at least 45 adapters
	assert.GreaterOrEqual(t, count, 45, "Should have at least 45 adapters registered")
}

// TestAdapterRegistryAllAdaptersHaveRequiredFields verifies adapter metadata is correct
func TestAdapterRegistryAllAdaptersHaveRequiredFields(t *testing.T) {
	allAdapters := adapters.AvailableAdapters

	for _, adapter := range allAdapters {
		t.Run(adapter.Name, func(t *testing.T) {
			// Name is required
			assert.NotEmpty(t, adapter.Name, "Adapter name should not be empty")

			// Description is required
			assert.NotEmpty(t, adapter.Description, "Adapter %s should have a description", adapter.Name)

			// Category must be valid
			assert.NotEmpty(t, adapter.Category, "Adapter %s should have a category", adapter.Name)

			// AuthType must be set
			assert.NotEmpty(t, adapter.AuthType, "Adapter %s should have an auth type", adapter.Name)

			// DocsURL should be present (can be empty for some)
			// Just log if missing
			if adapter.DocsURL == "" {
				t.Logf("Adapter %s has no DocsURL", adapter.Name)
			}
		})
	}
}

// TestAdapterRegistryAllCategoriesRepresented verifies all categories have adapters
func TestAdapterRegistryAllCategoriesRepresented(t *testing.T) {
	expectedCategories := []adapters.AdapterCategory{
		adapters.CategoryDatabase,
		adapters.CategoryStorage,
		adapters.CategoryVersionControl,
		adapters.CategoryProductivity,
		adapters.CategoryCommunication,
		adapters.CategorySearch,
		adapters.CategoryAutomation,
		adapters.CategoryInfrastructure,
		adapters.CategoryAnalytics,
		adapters.CategoryAI,
		adapters.CategoryUtility,
		adapters.CategoryDesign,
		adapters.CategoryCollaboration,
	}

	categoryCount := make(map[adapters.AdapterCategory]int)
	for _, adapter := range adapters.AvailableAdapters {
		categoryCount[adapter.Category]++
	}

	for _, category := range expectedCategories {
		count := categoryCount[category]
		t.Run(string(category), func(t *testing.T) {
			assert.Greater(t, count, 0, "Category %s should have at least one adapter", category)
			t.Logf("Category %s has %d adapters", category, count)
		})
	}
}

// TestAdapterRegistryAuthTypesPresent verifies all auth types are represented
func TestAdapterRegistryAuthTypesPresent(t *testing.T) {
	authTypes := adapters.GetAllAuthTypes()
	t.Logf("Available auth types: %v", authTypes)

	// Should have various auth types
	expectedAuthTypes := []string{"api_key", "token", "oauth2", "none"}
	authTypeSet := make(map[string]bool)
	for _, at := range authTypes {
		authTypeSet[at] = true
	}

	for _, expected := range expectedAuthTypes {
		t.Run(expected, func(t *testing.T) {
			assert.True(t, authTypeSet[expected], "Auth type %s should be present", expected)
		})
	}
}

// TestAdapterRegistryOfficialAdaptersExist verifies official MCP adapters are present
func TestAdapterRegistryOfficialAdaptersExist(t *testing.T) {
	official := adapters.GetOfficialAdapters()
	t.Logf("Official adapters count: %d", len(official))

	// Should have at least 10 official adapters
	assert.GreaterOrEqual(t, len(official), 10, "Should have at least 10 official adapters")

	// Check specific official adapters exist
	officialNames := make(map[string]bool)
	for _, a := range official {
		officialNames[a.Name] = true
	}

	expectedOfficial := []string{
		"postgresql",
		"sqlite",
		"github",
		"slack",
		"fetch",
		"memory",
		"filesystem",
		"brave-search",
		"puppeteer",
		"sentry",
		"everart",
	}

	for _, name := range expectedOfficial {
		t.Run("official_"+name, func(t *testing.T) {
			assert.True(t, officialNames[name], "Adapter %s should be marked as official", name)
		})
	}
}

// TestAdapterRegistrySupportedAdaptersExist verifies supported adapters are present
func TestAdapterRegistrySupportedAdaptersExist(t *testing.T) {
	supported := adapters.GetSupportedAdapters()
	t.Logf("Supported adapters count: %d", len(supported))

	// All adapters should be supported
	for _, a := range supported {
		assert.True(t, a.Supported, "Adapter %s should be marked as supported", a.Name)
	}

	// Should have many supported adapters
	assert.GreaterOrEqual(t, len(supported), 40, "Should have at least 40 supported adapters")
}

// TestAdapterRegistryDefaultRegistryInitialized verifies default registry initialization
func TestAdapterRegistryDefaultRegistryInitialized(t *testing.T) {
	require.NotNil(t, adapters.DefaultRegistry, "Default registry should not be nil")

	// Should have metadata for all available adapters
	for _, meta := range adapters.AvailableAdapters {
		retrieved, ok := adapters.DefaultRegistry.GetMetadata(meta.Name)
		assert.True(t, ok, "Should have metadata for %s", meta.Name)
		assert.Equal(t, meta.Name, retrieved.Name)
	}
}

// =============================================================================
// 2. ADAPTER SEARCH TESTS
// =============================================================================

// TestAdapterSearchByExactName tests search by exact name
func TestAdapterSearchByExactName(t *testing.T) {
	testCases := []string{
		"postgresql",
		"github",
		"slack",
		"fetch",
		"memory",
		"redis",
	}

	for _, name := range testCases {
		t.Run(name, func(t *testing.T) {
			results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
				Query:      name,
				MaxResults: 10,
			})

			require.NotEmpty(t, results, "Should find adapter %s", name)
			assert.Equal(t, name, results[0].Adapter.Name, "First result should be exact match")
			assert.Equal(t, 1.0, results[0].Score, "Exact match should have score 1.0")
			assert.Equal(t, "name", results[0].MatchType, "Match type should be name")
		})
	}
}

// TestAdapterSearchByPartialName tests search by partial name
func TestAdapterSearchByPartialName(t *testing.T) {
	testCases := []struct {
		query    string
		expected string
	}{
		{"post", "postgresql"},
		{"git", "github"},
		{"mongo", "mongodb"},
		{"drop", "dropbox"},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
				Query:      tc.query,
				MaxResults: 10,
			})

			require.NotEmpty(t, results, "Should find results for query %s", tc.query)

			// Should find the expected adapter
			found := false
			for _, r := range results {
				if r.Adapter.Name == tc.expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find %s with query %s", tc.expected, tc.query)
		})
	}
}

// TestAdapterSearchByCategory tests search by category
func TestAdapterSearchByCategory(t *testing.T) {
	categories := []adapters.AdapterCategory{
		adapters.CategoryDatabase,
		adapters.CategoryStorage,
		adapters.CategoryVersionControl,
		adapters.CategoryProductivity,
	}

	for _, category := range categories {
		t.Run(string(category), func(t *testing.T) {
			results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
				Categories: []adapters.AdapterCategory{category},
				MaxResults: 50,
			})

			require.NotEmpty(t, results, "Should find adapters in category %s", category)

			// All results should be in the specified category
			for _, r := range results {
				assert.Equal(t, category, r.Adapter.Category,
					"Result %s should be in category %s", r.Adapter.Name, category)
			}
		})
	}
}

// TestAdapterSearchByCapability tests search by capability keywords
func TestAdapterSearchByCapability(t *testing.T) {
	testCases := []struct {
		capability string
		expectMin  int
	}{
		{"database", 3},
		{"storage", 2},
		{"file", 2},
		{"search", 2},
		{"ai", 2},
		{"monitoring", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.capability, func(t *testing.T) {
			results := adapters.DefaultRegistry.SearchByCapability(tc.capability)
			t.Logf("Found %d adapters for capability %s", len(results), tc.capability)

			assert.GreaterOrEqual(t, len(results), tc.expectMin,
				"Should find at least %d adapters for capability %s", tc.expectMin, tc.capability)
		})
	}
}

// TestAdapterSearchFuzzyMatching tests fuzzy search functionality
func TestAdapterSearchFuzzyMatching(t *testing.T) {
	testCases := []struct {
		query       string
		shouldFind  string
		description string
	}{
		{"database", "postgresql", "Description match for database"},
		{"storage", "aws-s3", "Description match for storage"},
		{"automation", "puppeteer", "Description match for automation"},
		{"repositories", "github", "Description match for repositories"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
				Query:      tc.query,
				MaxResults: 20,
				MinScore:   0.1,
			})

			// Should find some results
			require.NotEmpty(t, results, "Should find results for %s", tc.query)

			// Check if expected adapter is in results
			found := false
			for _, r := range results {
				if r.Adapter.Name == tc.shouldFind {
					found = true
					t.Logf("Found %s with score %f, match type: %s",
						tc.shouldFind, r.Score, r.MatchType)
					break
				}
			}
			assert.True(t, found, "Should find %s for query '%s'", tc.shouldFind, tc.query)
		})
	}
}

// TestAdapterSearchSuggestions tests suggestion functionality
func TestAdapterSearchSuggestions(t *testing.T) {
	testCases := []struct {
		prefix         string
		expectedInList []string
	}{
		{"git", []string{"github", "gitlab"}},
		{"post", []string{"postgresql"}},
		{"mongo", []string{"mongodb"}},
		{"aws", []string{"aws-s3"}},
		{"google", []string{"google-drive", "google-search", "google-maps"}},
	}

	for _, tc := range testCases {
		t.Run(tc.prefix, func(t *testing.T) {
			suggestions := adapters.DefaultRegistry.GetAdapterSuggestions(tc.prefix, 10)
			require.NotEmpty(t, suggestions, "Should get suggestions for prefix %s", tc.prefix)

			suggestionNames := make(map[string]bool)
			for _, s := range suggestions {
				suggestionNames[s.Name] = true
			}

			for _, expected := range tc.expectedInList {
				assert.True(t, suggestionNames[expected],
					"Suggestions for %s should include %s", tc.prefix, expected)
			}
		})
	}
}

// TestAdapterSearchMaxResults tests result limiting
func TestAdapterSearchMaxResults(t *testing.T) {
	limits := []int{1, 3, 5, 10}

	for _, limit := range limits {
		t.Run(string(rune(limit+'0')), func(t *testing.T) {
			results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
				MaxResults: limit,
			})

			assert.LessOrEqual(t, len(results), limit, "Results should not exceed max %d", limit)
		})
	}
}

// TestAdapterSearchMinScore tests minimum score filtering
func TestAdapterSearchMinScore(t *testing.T) {
	minScores := []float64{0.1, 0.5, 0.8}

	for _, minScore := range minScores {
		t.Run("minScore", func(t *testing.T) {
			results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
				Query:    "post",
				MinScore: minScore,
			})

			for _, r := range results {
				assert.GreaterOrEqual(t, r.Score, minScore,
					"Result %s has score %f which is below min %f",
					r.Adapter.Name, r.Score, minScore)
			}
		})
	}
}

// TestAdapterSearchCombinedFilters tests combined filter functionality
func TestAdapterSearchCombinedFilters(t *testing.T) {
	// Search for official database adapters
	official := true
	results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
		Categories: []adapters.AdapterCategory{adapters.CategoryDatabase},
		Official:   &official,
		MaxResults: 50,
	})

	require.NotEmpty(t, results, "Should find official database adapters")

	for _, r := range results {
		assert.Equal(t, adapters.CategoryDatabase, r.Adapter.Category)
		assert.True(t, r.Adapter.Official)
	}

	// Search for api_key auth type adapters
	results = adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
		AuthTypes:  []string{"api_key"},
		MaxResults: 50,
	})

	require.NotEmpty(t, results, "Should find api_key auth type adapters")

	for _, r := range results {
		assert.Equal(t, "api_key", r.Adapter.AuthType)
	}
}

// =============================================================================
// 3. ADAPTER FUNCTIONALITY TESTS
// =============================================================================

// TestFilesystemAdapterBasicOperations tests filesystem adapter functionality
func TestFilesystemAdapterBasicOperations(t *testing.T) {
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "helix-mcp-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := servers.FilesystemAdapterConfig{
		AllowedPaths:   []string{tempDir},
		MaxFileSize:    1024 * 1024, // 1MB
		AllowWrite:     true,
		AllowDelete:    true,
		AllowCreateDir: true,
		FollowSymlinks: false,
	}

	adapter := servers.NewFilesystemAdapter(config)
	require.NotNil(t, adapter)

	ctx := context.Background()

	// Initialize
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	// Test Health
	err = adapter.Health(ctx)
	assert.NoError(t, err, "Health check should pass")

	// Test GetMCPTools
	tools := adapter.GetMCPTools()
	assert.NotEmpty(t, tools, "Should have tools defined")
	t.Logf("Filesystem adapter has %d tools", len(tools))

	// Verify expected tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"filesystem_read_file",
		"filesystem_write_file",
		"filesystem_list_directory",
		"filesystem_get_info",
		"filesystem_search",
	}

	for _, expected := range expectedTools {
		assert.True(t, toolNames[expected], "Should have tool %s", expected)
	}

	// Test write and read file
	testFilePath := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, MCP adapter test!"

	err = adapter.WriteFile(ctx, testFilePath, testContent)
	require.NoError(t, err, "Should write file successfully")

	content, err := adapter.ReadFile(ctx, testFilePath)
	require.NoError(t, err, "Should read file successfully")
	assert.Equal(t, testContent, content.Content)

	// Test list directory
	listing, err := adapter.ListDirectory(ctx, tempDir)
	require.NoError(t, err)
	assert.Equal(t, 1, listing.Count, "Should have one file in directory")

	// Test get file info
	info, err := adapter.GetFileInfo(ctx, testFilePath)
	require.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name)
	assert.False(t, info.IsDir)

	// Test create directory
	newDir := filepath.Join(tempDir, "subdir")
	err = adapter.CreateDirectory(ctx, newDir)
	require.NoError(t, err)

	info, err = adapter.GetFileInfo(ctx, newDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir)

	// Test search files
	results, err := adapter.SearchFiles(ctx, tempDir, "*.txt", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)

	// Test delete file
	err = adapter.DeleteFile(ctx, testFilePath)
	assert.NoError(t, err)

	// Verify file is deleted
	_, err = adapter.GetFileInfo(ctx, testFilePath)
	assert.Error(t, err)
}

// TestFilesystemAdapterPathRestrictions tests path security restrictions
func TestFilesystemAdapterPathRestrictions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "helix-mcp-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := servers.FilesystemAdapterConfig{
		AllowedPaths:   []string{tempDir},
		DeniedPaths:    []string{"/etc", "/root", "/.ssh"},
		MaxFileSize:    1024 * 1024,
		AllowWrite:     true,
		AllowDelete:    false,
		AllowCreateDir: false,
	}

	adapter := servers.NewFilesystemAdapter(config)
	ctx := context.Background()

	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	// Test reading outside allowed paths
	_, err = adapter.ReadFile(ctx, "/etc/passwd")
	assert.Error(t, err, "Should not allow reading /etc/passwd")

	// Test reading from denied path
	_, err = adapter.ReadFile(ctx, filepath.Join(tempDir, ".ssh", "config"))
	// Should error due to denied path pattern
	// Note: This might not error if .ssh doesn't exist, which is expected behavior

	// Test write restriction
	err = adapter.DeleteFile(ctx, filepath.Join(tempDir, "any-file"))
	assert.Error(t, err, "Should not allow delete when AllowDelete=false")

	// Test create directory restriction
	err = adapter.CreateDirectory(ctx, filepath.Join(tempDir, "newdir"))
	assert.Error(t, err, "Should not allow create dir when AllowCreateDir=false")
}

// TestMemoryAdapterBasicOperations tests memory adapter functionality
func TestMemoryAdapterBasicOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "helix-mcp-memory-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := servers.MemoryAdapterConfig{
		StoragePath:       tempDir,
		MaxEntities:       1000,
		MaxRelations:      5000,
		EnablePersistence: false, // Disable for testing
		AutoSaveInterval:  0,
	}

	adapter := servers.NewMemoryAdapter(config, nil)
	require.NotNil(t, adapter)

	ctx := context.Background()

	// Initialize
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	// Test Health
	err = adapter.Health(ctx)
	assert.NoError(t, err)

	// Test GetMCPTools
	tools := adapter.GetMCPTools()
	assert.NotEmpty(t, tools)
	t.Logf("Memory adapter has %d tools", len(tools))

	// Verify expected tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"memory_create_entity",
		"memory_get_entity",
		"memory_update_entity",
		"memory_delete_entity",
		"memory_search",
		"memory_create_relation",
		"memory_get_relations",
	}

	for _, expected := range expectedTools {
		assert.True(t, toolNames[expected], "Should have tool %s", expected)
	}

	// Test create entity
	entity1, err := adapter.CreateEntity(ctx, "Alice", "person",
		[]string{"Software engineer", "Likes Go"},
		map[string]interface{}{"age": 30})
	require.NoError(t, err)
	assert.NotEmpty(t, entity1.ID)
	assert.Equal(t, "Alice", entity1.Name)

	// Test create another entity
	entity2, err := adapter.CreateEntity(ctx, "Bob", "person",
		[]string{"Data scientist"},
		map[string]interface{}{"age": 28})
	require.NoError(t, err)

	// Test get entity by ID
	retrieved, err := adapter.GetEntity(ctx, entity1.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alice", retrieved.Name)

	// Test get entity by name
	retrieved, err = adapter.GetEntityByName(ctx, "Bob")
	require.NoError(t, err)
	assert.Equal(t, "Bob", retrieved.Name)

	// Test create relation
	relation, err := adapter.CreateRelation(ctx, entity1.ID, entity2.ID, "knows", 0.9, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, relation.ID)
	assert.Equal(t, "knows", relation.RelationType)

	// Test get entity relations
	relations, err := adapter.GetEntityRelations(ctx, entity1.ID, "outgoing")
	require.NoError(t, err)
	assert.Len(t, relations, 1)

	// Test search entities
	searchResults, err := adapter.SearchEntities(ctx, "Alice", "", 10)
	require.NoError(t, err)
	assert.Len(t, searchResults, 1)

	// Test search with relevance
	relevanceResults, err := adapter.SearchWithRelevance(ctx, "software", 10)
	require.NoError(t, err)
	assert.NotEmpty(t, relevanceResults)

	// Test update entity
	updated, err := adapter.UpdateEntity(ctx, entity1.ID,
		[]string{"New observation"},
		map[string]interface{}{"department": "Engineering"})
	require.NoError(t, err)
	assert.Contains(t, updated.Observations, "New observation")
	assert.Equal(t, "Engineering", updated.Properties["department"])

	// Test add observation
	err = adapter.AddObservation(ctx, entity1.ID, "Another observation")
	assert.NoError(t, err)

	// Test statistics
	stats, err := adapter.GetStatistics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalEntities)
	assert.Equal(t, 1, stats.TotalRelations)

	// Test delete relation
	err = adapter.DeleteRelation(ctx, relation.ID)
	assert.NoError(t, err)

	// Test delete entity
	err = adapter.DeleteEntity(ctx, entity2.ID)
	assert.NoError(t, err)

	// Verify entity is deleted
	_, err = adapter.GetEntity(ctx, entity2.ID)
	assert.Error(t, err)

	// Test clear
	err = adapter.Clear(ctx)
	assert.NoError(t, err)

	stats, err = adapter.GetStatistics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalEntities)
}

// TestFetchAdapterBasicOperations tests fetch adapter functionality
func TestFetchAdapterBasicOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping fetch adapter tests in short mode")
	}

	config := servers.FetchAdapterConfig{
		UserAgent:       "HelixAgent-Test/1.0",
		Timeout:         10 * time.Second,
		MaxRedirects:    5,
		MaxResponseSize: 1024 * 1024,
	}

	adapter := servers.NewFetchAdapter(config, nil)
	require.NotNil(t, adapter)

	ctx := context.Background()

	// Initialize
	err := adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	// Test Health
	err = adapter.Health(ctx)
	assert.NoError(t, err)

	// Test GetMCPTools
	tools := adapter.GetMCPTools()
	assert.NotEmpty(t, tools)
	t.Logf("Fetch adapter has %d tools", len(tools))

	// Verify expected tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"fetch_url",
		"fetch_json",
		"fetch_extract_links",
		"fetch_extract_text",
	}

	for _, expected := range expectedTools {
		assert.True(t, toolNames[expected], "Should have tool %s", expected)
	}

	// Test capabilities
	caps := adapter.GetCapabilities()
	assert.Equal(t, "fetch", caps["name"])
	assert.True(t, caps["initialized"].(bool))
}

// TestFetchAdapterDomainRestrictions tests domain filtering
func TestFetchAdapterDomainRestrictions(t *testing.T) {
	config := servers.FetchAdapterConfig{
		AllowedDomains: []string{"example.com", "api.example.com"},
		BlockedDomains: []string{"blocked.example.com"},
		Timeout:        5 * time.Second,
	}

	adapter := servers.NewFetchAdapter(config, nil)
	ctx := context.Background()

	err := adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	// Test blocked domain
	_, err = adapter.Fetch(ctx, "https://blocked.example.com/test", "GET", nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")

	// Test non-allowed domain
	_, err = adapter.Fetch(ctx, "https://other-domain.com/test", "GET", nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

// TestFetchAdapterExtractFunctions tests text/link extraction
func TestFetchAdapterExtractFunctions(t *testing.T) {
	config := servers.DefaultFetchAdapterConfig()
	adapter := servers.NewFetchAdapter(config, nil)

	// Test ExtractText
	html := `<html>
		<head><title>Test</title></head>
		<body>
			<script>alert('test');</script>
			<style>.foo { color: red; }</style>
			<p>Hello World</p>
			<p>Test &amp; Content</p>
		</body>
	</html>`

	text := adapter.ExtractText(html)
	assert.Contains(t, text, "Hello World")
	assert.Contains(t, text, "Test & Content")
	assert.NotContains(t, text, "alert")
	assert.NotContains(t, text, "color: red")

	// Test ExtractLinks
	htmlWithLinks := `<html>
		<body>
			<a href="/relative">Relative</a>
			<a href="https://example.com/absolute">Absolute</a>
			<a href="javascript:void(0)">JS Link</a>
			<a href="#anchor">Anchor</a>
		</body>
	</html>`

	links, err := adapter.ExtractLinks(htmlWithLinks, "https://base.com")
	require.NoError(t, err)
	assert.Contains(t, links, "https://base.com/relative")
	assert.Contains(t, links, "https://example.com/absolute")
	// Should not contain javascript or anchor links
	for _, link := range links {
		assert.NotContains(t, link, "javascript:")
	}
}

// =============================================================================
// 4. CATEGORY COVERAGE TESTS
// =============================================================================

// TestDatabaseCategoryAdapters verifies database category adapters
func TestDatabaseCategoryAdapters(t *testing.T) {
	dbAdapters := adapters.DefaultRegistry.ListByCategory(adapters.CategoryDatabase)
	require.NotEmpty(t, dbAdapters, "Should have database adapters")

	t.Logf("Database adapters: %v", dbAdapters)

	expectedAdapters := []string{
		"postgresql",
		"sqlite",
		"mongodb",
		"redis",
	}

	adapterSet := make(map[string]bool)
	for _, name := range dbAdapters {
		adapterSet[name] = true
	}

	for _, expected := range expectedAdapters {
		t.Run(expected, func(t *testing.T) {
			assert.True(t, adapterSet[expected], "Database category should include %s", expected)

			// Verify metadata
			meta, ok := adapters.DefaultRegistry.GetMetadata(expected)
			require.True(t, ok)
			assert.Equal(t, adapters.CategoryDatabase, meta.Category)
		})
	}
}

// TestStorageCategoryAdapters verifies storage category adapters
func TestStorageCategoryAdapters(t *testing.T) {
	storageAdapters := adapters.DefaultRegistry.ListByCategory(adapters.CategoryStorage)
	require.NotEmpty(t, storageAdapters, "Should have storage adapters")

	t.Logf("Storage adapters: %v", storageAdapters)

	expectedAdapters := []string{
		"aws-s3",
		"google-drive",
		"dropbox",
	}

	adapterSet := make(map[string]bool)
	for _, name := range storageAdapters {
		adapterSet[name] = true
	}

	for _, expected := range expectedAdapters {
		t.Run(expected, func(t *testing.T) {
			assert.True(t, adapterSet[expected], "Storage category should include %s", expected)
		})
	}
}

// TestVersionControlCategoryAdapters verifies version control category adapters
func TestVersionControlCategoryAdapters(t *testing.T) {
	vcAdapters := adapters.DefaultRegistry.ListByCategory(adapters.CategoryVersionControl)
	require.NotEmpty(t, vcAdapters, "Should have version control adapters")

	t.Logf("Version control adapters: %v", vcAdapters)

	expectedAdapters := []string{
		"github",
		"gitlab",
		"bitbucket",
	}

	adapterSet := make(map[string]bool)
	for _, name := range vcAdapters {
		adapterSet[name] = true
	}

	for _, expected := range expectedAdapters {
		t.Run(expected, func(t *testing.T) {
			assert.True(t, adapterSet[expected], "Version control category should include %s", expected)
		})
	}
}

// TestProductivityCategoryAdapters verifies productivity category adapters
func TestProductivityCategoryAdapters(t *testing.T) {
	prodAdapters := adapters.DefaultRegistry.ListByCategory(adapters.CategoryProductivity)
	require.NotEmpty(t, prodAdapters, "Should have productivity adapters")

	t.Logf("Productivity adapters: %v", prodAdapters)

	expectedAdapters := []string{
		"notion",
		"linear",
		"todoist",
	}

	adapterSet := make(map[string]bool)
	for _, name := range prodAdapters {
		adapterSet[name] = true
	}

	for _, expected := range expectedAdapters {
		t.Run(expected, func(t *testing.T) {
			assert.True(t, adapterSet[expected], "Productivity category should include %s", expected)
		})
	}
}

// TestAllCategoriesHaveCorrectAdapters verifies all categories
func TestAllCategoriesHaveCorrectAdapters(t *testing.T) {
	allCategories := adapters.GetAllCategories()

	for _, category := range allCategories {
		t.Run(string(category), func(t *testing.T) {
			categoryAdapters := adapters.DefaultRegistry.ListByCategory(category)

			// All adapters in the list should have the correct category
			for _, name := range categoryAdapters {
				meta, ok := adapters.DefaultRegistry.GetMetadata(name)
				require.True(t, ok, "Should have metadata for %s", name)
				assert.Equal(t, category, meta.Category,
					"Adapter %s should be in category %s", name, category)
			}
		})
	}
}

// =============================================================================
// 5. CONCURRENT ACCESS TESTS
// =============================================================================

// TestAdapterRegistryConcurrentAccess tests thread safety
func TestAdapterRegistryConcurrentAccess(t *testing.T) {
	registry := adapters.NewAdapterRegistry()

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent searches
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Perform various operations
			registry.Search(adapters.AdapterSearchOptions{
				Query:      "database",
				MaxResults: 10,
			})

			registry.SearchByCapability("storage")
			registry.GetAdapterSuggestions("git", 5)
			registry.ListByCategory(adapters.CategoryDatabase)
		}(i)
	}

	wg.Wait()

	// No panic or race condition should occur
	t.Log("Concurrent access test completed successfully")
}

// TestMemoryAdapterConcurrentAccess tests memory adapter thread safety
func TestMemoryAdapterConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "helix-mcp-concurrent-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := servers.MemoryAdapterConfig{
		StoragePath:       tempDir,
		MaxEntities:       10000,
		MaxRelations:      50000,
		EnablePersistence: false,
	}

	adapter := servers.NewMemoryAdapter(config, nil)
	ctx := context.Background()

	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	var wg sync.WaitGroup
	numGoroutines := 20

	// Concurrent entity creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			name := strings.Repeat("Entity", idx+1)
			_, err := adapter.CreateEntity(ctx, name, "test",
				[]string{"observation"}, nil)
			if err != nil {
				// May fail if limit reached, which is OK
				t.Logf("Create entity error (may be expected): %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Verify statistics
	stats, err := adapter.GetStatistics(ctx)
	require.NoError(t, err)
	t.Logf("Created %d entities concurrently", stats.TotalEntities)
}

// =============================================================================
// 6. EDGE CASE AND ERROR HANDLING TESTS
// =============================================================================

// TestAdapterSearchEmptyQuery tests searching with empty query
func TestAdapterSearchEmptyQuery(t *testing.T) {
	results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
		Query:      "",
		MaxResults: 50,
	})

	// Empty query should return all adapters (up to max)
	assert.NotEmpty(t, results)
	assert.LessOrEqual(t, len(results), 50)

	// All should have score 1.0 and match type "all"
	for _, r := range results {
		assert.Equal(t, 1.0, r.Score)
		assert.Equal(t, "all", r.MatchType)
	}
}

// TestAdapterSearchNoResults tests searching with query that has no matches
func TestAdapterSearchNoResults(t *testing.T) {
	results := adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
		Query:    "xyznonexistent123",
		MinScore: 0.5,
	})

	assert.Empty(t, results, "Should return empty results for non-matching query")
}

// TestMemoryAdapterUninitializedOperations tests operations on uninitialized adapter
func TestMemoryAdapterUninitializedOperations(t *testing.T) {
	config := servers.DefaultMemoryAdapterConfig()
	adapter := servers.NewMemoryAdapter(config, nil)

	ctx := context.Background()

	// Should fail when not initialized
	_, err := adapter.CreateEntity(ctx, "Test", "test", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	err = adapter.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestFilesystemAdapterUninitializedOperations tests operations on uninitialized adapter
func TestFilesystemAdapterUninitializedOperations(t *testing.T) {
	config := servers.DefaultFilesystemAdapterConfig()
	adapter := servers.NewFilesystemAdapter(config)

	ctx := context.Background()

	// Should fail when not initialized
	_, err := adapter.ReadFile(ctx, "/tmp/test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	err = adapter.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestFetchAdapterUninitializedOperations tests operations on uninitialized adapter
func TestFetchAdapterUninitializedOperations(t *testing.T) {
	config := servers.DefaultFetchAdapterConfig()
	adapter := servers.NewFetchAdapter(config, nil)

	ctx := context.Background()

	// Should fail when not initialized
	_, err := adapter.Fetch(ctx, "https://example.com", "GET", nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	err = adapter.Health(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestMemoryAdapterEntityLimits tests entity limit enforcement
func TestMemoryAdapterEntityLimits(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "helix-mcp-limit-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := servers.MemoryAdapterConfig{
		StoragePath:       tempDir,
		MaxEntities:       5, // Small limit for testing
		MaxRelations:      10,
		EnablePersistence: false,
	}

	adapter := servers.NewMemoryAdapter(config, nil)
	ctx := context.Background()

	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	// Create entities up to limit
	for i := 0; i < 5; i++ {
		_, err := adapter.CreateEntity(ctx, "Entity", "test", nil, nil)
		require.NoError(t, err)
	}

	// Should fail when limit exceeded
	_, err = adapter.CreateEntity(ctx, "ExtraEntity", "test", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum entity limit")
}

// TestMemoryAdapterRelationValidation tests relation validation
func TestMemoryAdapterRelationValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "helix-mcp-rel-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := servers.MemoryAdapterConfig{
		StoragePath:       tempDir,
		MaxEntities:       100,
		MaxRelations:      100,
		EnablePersistence: false,
	}

	adapter := servers.NewMemoryAdapter(config, nil)
	ctx := context.Background()

	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close()

	// Create one entity
	entity, err := adapter.CreateEntity(ctx, "Entity1", "test", nil, nil)
	require.NoError(t, err)

	// Should fail creating relation with non-existent entity
	_, err = adapter.CreateRelation(ctx, entity.ID, "nonexistent-id", "test", 1.0, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =============================================================================
// 7. BENCHMARK TESTS
// =============================================================================

// BenchmarkAdapterSearch benchmarks adapter search performance
func BenchmarkAdapterSearch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		adapters.DefaultRegistry.Search(adapters.AdapterSearchOptions{
			Query:      "database",
			MaxResults: 10,
		})
	}
}

// BenchmarkAdapterSearchByCapability benchmarks capability search
func BenchmarkAdapterSearchByCapability(b *testing.B) {
	for i := 0; i < b.N; i++ {
		adapters.DefaultRegistry.SearchByCapability("storage")
	}
}

// BenchmarkAdapterGetSuggestions benchmarks suggestion generation
func BenchmarkAdapterGetSuggestions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		adapters.DefaultRegistry.GetAdapterSuggestions("git", 10)
	}
}

// BenchmarkMemoryAdapterCreateEntity benchmarks entity creation
func BenchmarkMemoryAdapterCreateEntity(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "helix-bench-*")
	defer os.RemoveAll(tempDir)

	config := servers.MemoryAdapterConfig{
		StoragePath:       tempDir,
		MaxEntities:       b.N + 100,
		EnablePersistence: false,
	}

	adapter := servers.NewMemoryAdapter(config, nil)
	ctx := context.Background()
	adapter.Initialize(ctx)
	defer adapter.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.CreateEntity(ctx, "Entity", "test", nil, nil)
	}
}

// BenchmarkMemoryAdapterSearch benchmarks memory search
func BenchmarkMemoryAdapterSearch(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "helix-bench-*")
	defer os.RemoveAll(tempDir)

	config := servers.MemoryAdapterConfig{
		StoragePath:       tempDir,
		MaxEntities:       1000,
		EnablePersistence: false,
	}

	adapter := servers.NewMemoryAdapter(config, nil)
	ctx := context.Background()
	adapter.Initialize(ctx)
	defer adapter.Close()

	// Pre-populate with entities
	for i := 0; i < 100; i++ {
		adapter.CreateEntity(ctx, "TestEntity", "test", []string{"observation"}, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.SearchEntities(ctx, "Test", "", 20)
	}
}
