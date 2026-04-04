// Package providers provides file-based context
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

// FileProvider provides context from files
type FileProvider struct {
	basePath string
	logger   *logrus.Logger
	maxSize  int64 // Max file size in bytes
}

// NewFileProvider creates a new file provider
func NewFileProvider(basePath string) *FileProvider {
	return &FileProvider{
		basePath: basePath,
		logger:   logrus.New(),
		maxSize:  1024 * 1024, // 1MB default
	}
}

// Name returns the provider name
func (f *FileProvider) Name() string {
	return "file"
}

// Description returns the provider description
func (f *FileProvider) Description() string {
	return "Provides context from local files"
}

// Resolve resolves file context
func (f *FileProvider) Resolve(ctx context.Context, query string) ([]ContextItem, error) {
	// Clean the path
	query = filepath.Clean(query)
	
	// Check for path escape
	fullPath := filepath.Join(f.basePath, query)
	if !strings.HasPrefix(fullPath, f.basePath) {
		return nil, fmt.Errorf("path escapes base directory")
	}
	
	// Check if path exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", query)
		}
		return nil, err
	}
	
	// Handle directory
	if info.IsDir() {
		return f.resolveDirectory(ctx, fullPath, query)
	}
	
	// Handle single file
	return f.resolveFile(fullPath, query)
}

// resolveFile resolves a single file
func (f *FileProvider) resolveFile(fullPath, relPath string) ([]ContextItem, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	
	// Check size
	if info.Size() > f.maxSize {
		return []ContextItem{{
			Name:        filepath.Base(relPath),
			Description: fmt.Sprintf("File too large (%d bytes)", info.Size()),
			Content:     fmt.Sprintf("[File exceeds max size of %d bytes]", f.maxSize),
			Source:      relPath,
			Timestamp:   time.Now(),
		}}, nil
	}
	
	// Read file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	
	return []ContextItem{{
		Name:        filepath.Base(relPath),
		Description: fmt.Sprintf("%d bytes", len(content)),
		Content:     string(content),
		Source:      relPath,
		Timestamp:   info.ModTime(),
	}}, nil
}

// resolveDirectory resolves a directory
func (f *FileProvider) resolveDirectory(ctx context.Context, fullPath, relPath string) ([]ContextItem, error) {
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	var items []ContextItem
	
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return items, ctx.Err()
		default:
		}
		
		if entry.IsDir() {
			continue // Skip subdirectories for now
		}
		
		// Check if it's a text file
		if !isTextFile(entry.Name()) {
			continue
		}
		
		entryPath := filepath.Join(fullPath, entry.Name())
		entryRelPath := filepath.Join(relPath, entry.Name())
		
		entryItems, err := f.resolveFile(entryPath, entryRelPath)
		if err != nil {
			f.logger.WithError(err).WithField("file", entry.Name()).Debug("Failed to resolve file")
			continue
		}
		
		items = append(items, entryItems...)
	}
	
	return items, nil
}

// isTextFile checks if a file is likely a text file
func isTextFile(name string) bool {
	textExtensions := map[string]bool{
		".go":    true,
		".js":    true,
		".ts":    true,
		".py":    true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".rs":    true,
		".rb":    true,
		".php":   true,
		".swift": true,
		".kt":    true,
		".md":    true,
		".txt":   true,
		".json":  true,
		".yaml":  true,
		".yml":   true,
		".xml":   true,
		".html":  true,
		".css":   true,
		".scss":  true,
		".sql":   true,
		".sh":    true,
		".bash":  true,
		".zsh":   true,
		".vim":   true,
		".lua":   true,
	}
	
	ext := strings.ToLower(filepath.Ext(name))
	return textExtensions[ext]
}

// WithMaxSize sets the max file size
func (f *FileProvider) WithMaxSize(size int64) *FileProvider {
	f.maxSize = size
	return f
}

// FindFiles finds files matching a pattern
func (f *FileProvider) FindFiles(pattern string) ([]string, error) {
	var matches []string
	
	err := filepath.Walk(f.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Check if file matches pattern
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}
		
		if matched {
			relPath, _ := filepath.Rel(f.basePath, path)
			matches = append(matches, relPath)
		}
		
		return nil
	})
	
	return matches, err
}

// GetFileInfo returns file information
func (f *FileProvider) GetFileInfo(relPath string) (os.FileInfo, error) {
	fullPath := filepath.Join(f.basePath, relPath)
	fullPath = filepath.Clean(fullPath)
	
	if !strings.HasPrefix(fullPath, f.basePath) {
		return nil, fmt.Errorf("path escapes base directory")
	}
	
	return os.Stat(fullPath)
}

// ReadFile reads a file and returns its content
func (f *FileProvider) ReadFile(relPath string) ([]byte, error) {
	fullPath := filepath.Join(f.basePath, relPath)
	fullPath = filepath.Clean(fullPath)
	
	if !strings.HasPrefix(fullPath, f.basePath) {
		return nil, fmt.Errorf("path escapes base directory")
	}
	
	return os.ReadFile(fullPath)
}

// FileExists checks if a file exists
func (f *FileProvider) FileExists(relPath string) bool {
	fullPath := filepath.Join(f.basePath, relPath)
	fullPath = filepath.Clean(fullPath)
	
	if !strings.HasPrefix(fullPath, f.basePath) {
		return false
	}
	
	_, err := os.Stat(fullPath)
	return err == nil
}

// ListDirectory lists files in a directory
func (f *FileProvider) ListDirectory(relPath string) ([]string, error) {
	fullPath := filepath.Join(f.basePath, relPath)
	fullPath = filepath.Clean(fullPath)
	
	if !strings.HasPrefix(fullPath, f.basePath) {
		return nil, fmt.Errorf("path escapes base directory")
	}
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	
	return names, nil
}
