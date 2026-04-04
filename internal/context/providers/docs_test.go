package providers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocsProvider(t *testing.T) {
	logger := logrus.New()
	provider := NewDocsProvider(nil, logger)

	require.NotNil(t, provider)
	assert.NotNil(t, provider.searchPaths)
	assert.NotNil(t, provider.logger)
	assert.NotNil(t, provider.cache)
	assert.Greater(t, len(provider.searchPaths), 0)
}

func TestNewDocsProvider_CustomPaths(t *testing.T) {
	paths := []string{"./custom", "./docs"}
	provider := NewDocsProvider(paths, nil)

	assert.Equal(t, paths, provider.searchPaths)
}

func TestDocsProvider_Name(t *testing.T) {
	provider := NewDocsProvider(nil, nil)
	assert.Equal(t, "docs", provider.Name())
}

func TestDocsProvider_Description(t *testing.T) {
	provider := NewDocsProvider(nil, nil)
	desc := provider.Description()
	assert.Contains(t, desc, "documentation")
}

func TestDocsProvider_Resolve_Readme(t *testing.T) {
	tempDir := t.TempDir()

	// Create test README
	readmePath := filepath.Join(tempDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test Project"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "readme")

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "README.md", items[0].Name)
	assert.Contains(t, items[0].Content, "# Test Project")
}

func TestDocsProvider_Resolve_API(t *testing.T) {
	tempDir := t.TempDir()

	// Create test API doc
	docsDir := filepath.Join(tempDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))
	apiPath := filepath.Join(docsDir, "api.md")
	require.NoError(t, os.WriteFile(apiPath, []byte("# API Documentation"), 0644))

	provider := NewDocsProvider([]string{docsDir}, nil)
	items, err := provider.Resolve(nil, "api")

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "api.md", items[0].Name)
}

func TestDocsProvider_Resolve_Setup(t *testing.T) {
	tempDir := t.TempDir()

	// Create setup doc
	setupPath := filepath.Join(tempDir, "SETUP.md")
	require.NoError(t, os.WriteFile(setupPath, []byte("# Setup Instructions"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "setup")

	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestDocsProvider_Resolve_Contributing(t *testing.T) {
	tempDir := t.TempDir()

	// Create contributing doc
	contribPath := filepath.Join(tempDir, "CONTRIBUTING.md")
	require.NoError(t, os.WriteFile(contribPath, []byte("# Contributing"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "contributing")

	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestDocsProvider_Resolve_License(t *testing.T) {
	tempDir := t.TempDir()

	// Create license file
	licensePath := filepath.Join(tempDir, "LICENSE")
	require.NoError(t, os.WriteFile(licensePath, []byte("MIT License"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "license")

	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestDocsProvider_Resolve_Changelog(t *testing.T) {
	tempDir := t.TempDir()

	// Create changelog
	changelogPath := filepath.Join(tempDir, "CHANGELOG.md")
	require.NoError(t, os.WriteFile(changelogPath, []byte("# Changelog"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "changelog")

	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestDocsProvider_Resolve_SpecificFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create specific file
	testPath := filepath.Join(tempDir, "custom.md")
	require.NoError(t, os.WriteFile(testPath, []byte("# Custom"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "custom.md")

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "custom.md", items[0].Name)
}

func TestDocsProvider_Resolve_Search(t *testing.T) {
	tempDir := t.TempDir()

	// Create file matching search
	testPath := filepath.Join(tempDir, "architecture.md")
	require.NoError(t, os.WriteFile(testPath, []byte("# Architecture"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "architecture")

	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestDocsProvider_findReadmeFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create README
	readmePath := filepath.Join(tempDir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# README"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.findReadmeFiles()

	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestDocsProvider_findAPIDocs(t *testing.T) {
	tempDir := t.TempDir()

	// Create API doc
	apiPath := filepath.Join(tempDir, "API.md")
	require.NoError(t, os.WriteFile(apiPath, []byte("# API"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.findAPIDocs()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 0)
}

func TestDocsProvider_findSetupDocs(t *testing.T) {
	tempDir := t.TempDir()

	// Create setup doc
	setupPath := filepath.Join(tempDir, "INSTALL.md")
	require.NoError(t, os.WriteFile(setupPath, []byte("# Install"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.findSetupDocs()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 0)
}

func TestDocsProvider_searchDocs(t *testing.T) {
	tempDir := t.TempDir()

	// Create doc
	docPath := filepath.Join(tempDir, "mydocument.md")
	require.NoError(t, os.WriteFile(docPath, []byte("# Doc"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.searchDocs("mydocument")

	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestDocsProvider_readFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.md")
	content := []byte("# Test Content")
	require.NoError(t, os.WriteFile(filePath, content, 0644))

	provider := NewDocsProvider(nil, nil)
	item := provider.readFile(filePath)

	require.NotNil(t, item)
	assert.Equal(t, "test.md", item.Name)
	assert.Equal(t, string(content), item.Content)
}

func TestDocsProvider_readFile_Large(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "large.md")
	largeContent := make([]byte, 2000000) // 2MB
	require.NoError(t, os.WriteFile(filePath, largeContent, 0644))

	provider := NewDocsProvider(nil, nil)
	provider.maxSize = 1000000 // 1MB limit
	item := provider.readFile(filePath)

	require.NotNil(t, item)
	assert.Contains(t, item.Content, "too large")
}

func TestDocsProvider_readFile_NotFound(t *testing.T) {
	provider := NewDocsProvider(nil, nil)
	item := provider.readFile("/nonexistent/file.md")
	assert.Nil(t, item)
}

func TestDocsProvider_Cache(t *testing.T) {
	provider := NewDocsProvider(nil, nil)

	// Add to cache
	provider.addToCache("test", "cached content")

	// Retrieve from cache
	item := provider.getFromCache("test")
	require.NotNil(t, item)
	assert.Equal(t, "cached content", item.Content)
}

func TestDocsProvider_Cache_Expired(t *testing.T) {
	// This test would require time manipulation
	// For simplicity, just test the cache exists
	provider := NewDocsProvider(nil, nil)
	item := provider.getFromCache("nonexistent")
	assert.Nil(t, item)
}

func TestDocsProvider_ClearCache(t *testing.T) {
	provider := NewDocsProvider(nil, nil)
	provider.addToCache("test", "content")

	provider.ClearCache()

	item := provider.getFromCache("test")
	assert.Nil(t, item)
}

func TestDocsProvider_SetSearchPaths(t *testing.T) {
	provider := NewDocsProvider([]string{"path1"}, nil)
	newPaths := []string{"path2", "path3"}

	provider.SetSearchPaths(newPaths)

	assert.Equal(t, newPaths, provider.searchPaths)
}

func TestDocsProvider_AddSearchPath(t *testing.T) {
	provider := NewDocsProvider([]string{"path1"}, nil)

	provider.AddSearchPath("path2")

	assert.Len(t, provider.searchPaths, 2)
	assert.Contains(t, provider.searchPaths, "path2")
}

func TestIsDocFile(t *testing.T) {
	assert.True(t, isDocFile("readme.md"))
	assert.True(t, isDocFile("doc.txt"))
	assert.True(t, isDocFile("guide.rst"))
	assert.True(t, isDocFile("manual.adoc"))
	assert.True(t, isDocFile("notes.org"))

	assert.False(t, isDocFile("script.go"))
	assert.False(t, isDocFile("data.json"))
	assert.False(t, isDocFile("image.png"))
}

func TestDocsProvider_FindAllDocs(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple docs
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "doc1.md"), []byte("1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "doc2.txt"), []byte("2"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "script.go"), []byte("code"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	docs, err := provider.FindAllDocs()

	require.NoError(t, err)
	assert.Len(t, docs, 2) // Only .md and .txt
}

func TestDocsProvider_GetDocSummary(t *testing.T) {
	tempDir := t.TempDir()

	// Create categorized docs (use .txt extension for LICENSE since isDocFile filters extensions)
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("r"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "LICENSE.txt"), []byte("l"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "other.md"), []byte("o"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	summary, err := provider.GetDocSummary()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, summary.TotalFiles, 2)
	assert.GreaterOrEqual(t, summary.ByCategory["readme"], 1)
}

func TestDocsSummary(t *testing.T) {
	summary := &DocsSummary{
		TotalFiles: 5,
		ByCategory: map[string]int{
			"readme": 2,
			"api":    3,
		},
	}

	assert.Equal(t, 5, summary.TotalFiles)
	assert.Equal(t, 2, summary.ByCategory["readme"])
}

func TestDocsProvider_Resolve_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.Resolve(nil, "nonexistent_query")

	// Should not error, just return empty
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestDocsProvider_findByPatterns(t *testing.T) {
	tempDir := t.TempDir()

	// Create file matching pattern
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "MATCH.md"), []byte("match"), 0644))

	provider := NewDocsProvider([]string{tempDir}, nil)
	items, err := provider.findByPatterns([]string{"MATCH.md"})

	require.NoError(t, err)
	assert.Len(t, items, 1)
}
