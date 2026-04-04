// Package providers provides documentation context provider
package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DocsProvider provides context from documentation files
type DocsProvider struct {
	searchPaths []string
	logger      *logrus.Logger
	maxSize     int64
	cache       map[string]*docsCacheEntry
}

type docsCacheEntry struct {
	content   string
	timestamp time.Time
}

// NewDocsProvider creates a new documentation provider
func NewDocsProvider(searchPaths []string, logger *logrus.Logger) *DocsProvider {
	if logger == nil {
		logger = logrus.New()
	}

	if len(searchPaths) == 0 {
		// Default search paths
		searchPaths = []string{
			"./docs",
			"./documentation",
			"./doc",
			"./wiki",
			"./.github",
		}
	}

	return &DocsProvider{
		searchPaths: searchPaths,
		logger:      logger,
		maxSize:     1024 * 1024, // 1MB
		cache:       make(map[string]*docsCacheEntry),
	}
}

// Name returns the provider name
func (d *DocsProvider) Name() string {
	return "docs"
}

// Description returns the provider description
func (d *DocsProvider) Description() string {
	return "Provides context from documentation files (README, docs, etc.)"
}

// Resolve resolves documentation queries
func (d *DocsProvider) Resolve(ctx context.Context, query string) ([]ContextItem, error) {
	// Query can be:
	// - "readme" -> find README files
	// - "api" -> find API documentation
	// - "setup" -> find setup/installation docs
	// - "contributing" -> find contribution guidelines
	// - specific filename like "README.md"

	query = strings.ToLower(strings.TrimSpace(query))

	// Check cache first
	if cached := d.getFromCache(query); cached != nil {
		return []ContextItem{*cached}, nil
	}

	// Handle specific file requests
	if strings.HasSuffix(query, ".md") || strings.HasSuffix(query, ".txt") {
		return d.findSpecificFile(query)
	}

	// Handle semantic queries
	switch query {
	case "readme", "overview":
		return d.findReadmeFiles()
	case "api", "reference":
		return d.findAPIDocs()
	case "setup", "install", "installation":
		return d.findSetupDocs()
	case "contributing", "contribute":
		return d.findContributingDocs()
	case "license", "licensing":
		return d.findLicenseFiles()
	case "changelog", "changes", "history":
		return d.findChangelogFiles()
	default:
		// Search for files containing query in name
		return d.searchDocs(query)
	}
}

// findReadmeFiles finds README files
func (d *DocsProvider) findReadmeFiles() ([]ContextItem, error) {
	patterns := []string{
		"README.md", "README.txt", "README",
		"readme.md", "readme.txt", "readme",
	}
	return d.findByPatterns(patterns)
}

// findAPIDocs finds API documentation
func (d *DocsProvider) findAPIDocs() ([]ContextItem, error) {
	patterns := []string{
		"API.md", "API_REFERENCE.md", "api.md",
		"docs/api.md", "docs/api/**",
		"reference.md", "REFERENCE.md",
	}
	return d.findByPatterns(patterns)
}

// findSetupDocs finds setup/installation documentation
func (d *DocsProvider) findSetupDocs() ([]ContextItem, error) {
	patterns := []string{
		"SETUP.md", "INSTALL.md", "GETTING_STARTED.md",
		"setup.md", "install.md", "getting_started.md",
		"docs/setup.md", "docs/installation.md",
	}
	return d.findByPatterns(patterns)
}

// findContributingDocs finds contribution guidelines
func (d *DocsProvider) findContributingDocs() ([]ContextItem, error) {
	patterns := []string{
		"CONTRIBUTING.md", "contributing.md",
		"CODE_OF_CONDUCT.md", "code_of_conduct.md",
		"SECURITY.md", "security.md",
	}
	return d.findByPatterns(patterns)
}

// findLicenseFiles finds license files
func (d *DocsProvider) findLicenseFiles() ([]ContextItem, error) {
	patterns := []string{
		"LICENSE", "LICENSE.md", "LICENSE.txt",
		"license", "license.md", "license.txt",
		"COPYING", "copying",
	}
	return d.findByPatterns(patterns)
}

// findChangelogFiles finds changelog files
func (d *DocsProvider) findChangelogFiles() ([]ContextItem, error) {
	patterns := []string{
		"CHANGELOG.md", "changelog.md",
		"CHANGES.md", "changes.md",
		"HISTORY.md", "history.md",
		"NEWS.md", "news.md",
	}
	return d.findByPatterns(patterns)
}

// findSpecificFile finds a specific file
func (d *DocsProvider) findSpecificFile(filename string) ([]ContextItem, error) {
	for _, searchPath := range d.searchPaths {
		fullPath := filepath.Join(searchPath, filename)
		
		// Also check without search path (root level)
		paths := []string{fullPath, filename}
		
		for _, path := range paths {
			if item := d.readFile(path); item != nil {
				return []ContextItem{*item}, nil
			}
		}
	}
	
	return nil, fmt.Errorf("file not found: %s", filename)
}

// searchDocs searches for documentation containing query
func (d *DocsProvider) searchDocs(query string) ([]ContextItem, error) {
	var items []ContextItem
	
	for _, searchPath := range d.searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}
		
		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Continue walking
			}
			
			if info.IsDir() {
				return nil
			}
			
			// Check if filename contains query
			name := strings.ToLower(info.Name())
			if strings.Contains(name, query) && isDocFile(path) {
				if item := d.readFile(path); item != nil {
					items = append(items, *item)
				}
			}
			
			return nil
		})
		
		if err != nil {
			d.logger.WithError(err).WithField("path", searchPath).Warn("Failed to walk docs path")
		}
	}
	
	return items, nil
}

// findByPatterns finds files matching any of the patterns
func (d *DocsProvider) findByPatterns(patterns []string) ([]ContextItem, error) {
	var items []ContextItem
	found := make(map[string]bool)
	
	for _, searchPath := range d.searchPaths {
		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(searchPath, pattern))
			if err != nil {
				continue
			}
			
			// Also check root level
			rootMatches, _ := filepath.Glob(pattern)
			matches = append(matches, rootMatches...)
			
			for _, match := range matches {
				if found[match] {
					continue
				}
				
				if item := d.readFile(match); item != nil {
					items = append(items, *item)
					found[match] = true
				}
			}
		}
	}
	
	return items, nil
}

// readFile reads a file and returns a ContextItem
func (d *DocsProvider) readFile(path string) *ContextItem {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	
	if info.Size() > d.maxSize {
		return &ContextItem{
			Name:        filepath.Base(path),
			Description: fmt.Sprintf("Documentation file (truncated, %d bytes)", info.Size()),
			Content:     fmt.Sprintf("[File too large: %s]", path),
			Source:      path,
			Timestamp:   info.ModTime(),
		}
	}
	
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	
	return &ContextItem{
		Name:        filepath.Base(path),
		Description: fmt.Sprintf("Documentation file (%d bytes)", len(content)),
		Content:     string(content),
		Source:      path,
		Timestamp:   info.ModTime(),
	}
}

// getFromCache retrieves from cache if valid
func (d *DocsProvider) getFromCache(query string) *ContextItem {
	entry, ok := d.cache[query]
	if !ok {
		return nil
	}
	
	// Cache expires after 5 minutes
	if time.Since(entry.timestamp) > 5*time.Minute {
		delete(d.cache, query)
		return nil
	}
	
	return &ContextItem{
		Name:        query,
		Description: "Cached documentation",
		Content:     entry.content,
		Source:      "cache",
		Timestamp:   entry.timestamp,
	}
}

// addToCache adds an item to cache
func (d *DocsProvider) addToCache(query, content string) {
	d.cache[query] = &docsCacheEntry{
		content:   content,
		timestamp: time.Now(),
	}
}

// ClearCache clears the documentation cache
func (d *DocsProvider) ClearCache() {
	d.cache = make(map[string]*docsCacheEntry)
}

// SetSearchPaths updates the search paths
func (d *DocsProvider) SetSearchPaths(paths []string) {
	d.searchPaths = paths
}

// AddSearchPath adds a search path
func (d *DocsProvider) AddSearchPath(path string) {
	d.searchPaths = append(d.searchPaths, path)
}

// isDocFile checks if a file is a documentation file
func isDocFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	docExtensions := map[string]bool{
		".md":    true,
		".txt":   true,
		".rst":   true,
		".adoc":  true,
		".org":   true,
		".wiki":  true,
		".mdown": true,
		".mkd":   true,
	}
	
	return docExtensions[ext]
}

// FindAllDocs finds all documentation files
func (d *DocsProvider) FindAllDocs() ([]string, error) {
	var docs []string
	
	for _, searchPath := range d.searchPaths {
		if _, err := os.Stat(searchPath); os.IsNotExist(err) {
			continue
		}
		
		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			
			if !info.IsDir() && isDocFile(path) {
				docs = append(docs, path)
			}
			
			return nil
		})
		
		if err != nil {
			return docs, err
		}
	}
	
	return docs, nil
}

// GetDocSummary gets a summary of available documentation
func (d *DocsProvider) GetDocSummary() (*DocsSummary, error) {
	docs, err := d.FindAllDocs()
	if err != nil {
		return nil, err
	}
	
	summary := &DocsSummary{
		TotalFiles: len(docs),
		ByCategory: make(map[string]int),
	}
	
	for _, doc := range docs {
		name := strings.ToLower(filepath.Base(doc))
		
		switch {
		case strings.HasPrefix(name, "readme"):
			summary.ByCategory["readme"]++
		case strings.HasPrefix(name, "api") || strings.Contains(doc, "/api/"):
			summary.ByCategory["api"]++
		case strings.HasPrefix(name, "changelog") || strings.HasPrefix(name, "changes"):
			summary.ByCategory["changelog"]++
		case strings.HasPrefix(name, "contributing"):
			summary.ByCategory["contributing"]++
		case strings.HasPrefix(name, "license") || strings.HasPrefix(name, "copying"):
			summary.ByCategory["license"]++
		default:
			summary.ByCategory["other"]++
		}
	}
	
	return summary, nil
}

// DocsSummary represents a summary of documentation
type DocsSummary struct {
	TotalFiles int            `json:"total_files"`
	ByCategory map[string]int `json:"by_category"`
}
