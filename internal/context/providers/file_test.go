package providers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileProvider(t *testing.T) {
	provider := NewFileProvider("/tmp")
	
	require.NotNil(t, provider)
	assert.Equal(t, "/tmp", provider.basePath)
	assert.Equal(t, int64(1024*1024), provider.maxSize)
}

func TestFileProvider_Name(t *testing.T) {
	provider := NewFileProvider("/tmp")
	assert.Equal(t, "file", provider.Name())
}

func TestFileProvider_Description(t *testing.T) {
	provider := NewFileProvider("/tmp")
	assert.NotEmpty(t, provider.Description())
}

func TestFileProvider_Resolve_SingleFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello world"), 0644))
	
	provider := NewFileProvider(tempDir)
	ctx := context.Background()
	
	items, err := provider.Resolve(ctx, "test.txt")
	
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "test.txt", items[0].Name)
	assert.Equal(t, "hello world", items[0].Content)
}

func TestFileProvider_Resolve_Directory(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content2"), 0644))
	
	provider := NewFileProvider(tempDir)
	ctx := context.Background()
	
	items, err := provider.Resolve(ctx, ".")
	
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestFileProvider_Resolve_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	
	provider := NewFileProvider(tempDir)
	ctx := context.Background()
	
	_, err := provider.Resolve(ctx, "nonexistent.txt")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestFileProvider_Resolve_PathEscape(t *testing.T) {
	tempDir := t.TempDir()
	
	provider := NewFileProvider(tempDir)
	ctx := context.Background()
	
	_, err := provider.Resolve(ctx, "../etc/passwd")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path escapes")
}

func TestFileProvider_Resolve_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create large file
	testFile := filepath.Join(tempDir, "large.txt")
	largeContent := make([]byte, 2000)
	require.NoError(t, os.WriteFile(testFile, largeContent, 0644))
	
	provider := NewFileProvider(tempDir).WithMaxSize(1000)
	ctx := context.Background()
	
	items, err := provider.Resolve(ctx, "large.txt")
	
	require.NoError(t, err)
	assert.Contains(t, items[0].Content, "exceeds max size")
}

func TestFileProvider_WithMaxSize(t *testing.T) {
	provider := NewFileProvider("/tmp").WithMaxSize(5000)
	assert.Equal(t, int64(5000), provider.maxSize)
}

func TestFileProvider_FindFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test files
	os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("text"), 0644)
	
	provider := NewFileProvider(tempDir)
	
	matches, err := provider.FindFiles("*.go")
	
	require.NoError(t, err)
	assert.Len(t, matches, 2)
}

func TestFileProvider_GetFileInfo(t *testing.T) {
	tempDir := t.TempDir()
	
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))
	
	provider := NewFileProvider(tempDir)
	
	info, err := provider.GetFileInfo("test.txt")
	
	require.NoError(t, err)
	assert.False(t, info.IsDir())
}

func TestFileProvider_GetFileInfo_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewFileProvider(tempDir)
	
	_, err := provider.GetFileInfo("nonexistent.txt")
	assert.Error(t, err)
}

func TestFileProvider_ReadFile(t *testing.T) {
	tempDir := t.TempDir()
	
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))
	
	provider := NewFileProvider(tempDir)
	
	content, err := provider.ReadFile("test.txt")
	
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))
}

func TestFileProvider_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	
	testFile := filepath.Join(tempDir, "exists.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("content"), 0644))
	
	provider := NewFileProvider(tempDir)
	
	assert.True(t, provider.FileExists("exists.txt"))
	assert.False(t, provider.FileExists("nonexistent.txt"))
}

func TestFileProvider_ListDirectory(t *testing.T) {
	tempDir := t.TempDir()
	
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte{}, 0644)
	
	provider := NewFileProvider(tempDir)
	
	names, err := provider.ListDirectory(".")
	
	require.NoError(t, err)
	assert.Len(t, names, 2)
}

func TestIsTextFile(t *testing.T) {
	// Text files
	assert.True(t, isTextFile("main.go"))
	assert.True(t, isTextFile("script.js"))
	assert.True(t, isTextFile("style.css"))
	assert.True(t, isTextFile("readme.md"))
	assert.True(t, isTextFile("config.json"))
	
	// Non-text files
	assert.False(t, isTextFile("image.png"))
	assert.False(t, isTextFile("archive.zip"))
	assert.False(t, isTextFile("binary.exe"))
}
