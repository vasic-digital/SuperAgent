// Package servers provides an LSP server registry for multiple language servers.
package servers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLSPServerRegistry(t *testing.T) {
	tests := []struct {
		name             string
		config           RegistryConfig
		expectedPathsLen int
	}{
		{
			name:             "default configuration",
			config:           RegistryConfig{},
			expectedPathsLen: 4, // Default search paths
		},
		{
			name: "custom logger",
			config: RegistryConfig{
				Logger: logrus.New(),
			},
			expectedPathsLen: 4,
		},
		{
			name: "custom search paths",
			config: RegistryConfig{
				SearchPaths: []string{"/custom/path1", "/custom/path2"},
			},
			expectedPathsLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewLSPServerRegistry(tt.config)
			assert.NotNil(t, registry)
			assert.NotNil(t, registry.servers)
			assert.NotNil(t, registry.byLanguage)
			assert.NotNil(t, registry.logger)
			assert.Len(t, registry.searchPaths, tt.expectedPathsLen)
		})
	}
}

func TestLSPServerRegistry_loadDefaultServers(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Verify default servers are loaded
	expectedServers := []string{
		"gopls",
		"rust-analyzer",
		"pylsp",
		"pyright",
		"typescript-language-server",
		"clangd",
		"jdtls",
		"omnisharp",
		"phpactor",
		"solargraph",
		"elixir-ls",
		"haskell-language-server",
		"bash-language-server",
		"yaml-language-server",
		"docker-langserver",
		"terraform-ls",
		"lua-language-server",
		"lemminx",
	}

	for _, id := range expectedServers {
		server, err := registry.Get(id)
		assert.NoError(t, err, "Expected server %s to be registered", id)
		assert.NotNil(t, server)
	}
}

func TestLSPServerRegistry_Get(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	tests := []struct {
		name        string
		serverID    string
		expectError bool
	}{
		{
			name:        "get existing server",
			serverID:    "gopls",
			expectError: false,
		},
		{
			name:        "get non-existent server",
			serverID:    "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := registry.Get(tt.serverID)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, server)
				assert.Equal(t, tt.serverID, server.ID)
			}
		})
	}
}

func TestLSPServerRegistry_GetByLanguage(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Manually set binary paths to simulate available servers
	registry.mu.Lock()
	for _, server := range registry.servers {
		if server.Language == "python" {
			server.Binary = "/usr/bin/" + server.Command // Simulate found binary
		}
	}
	registry.mu.Unlock()

	tests := []struct {
		name        string
		language    string
		expectCount int // 0 if unknown language
	}{
		{
			name:        "get python servers",
			language:    "python",
			expectCount: 2, // pylsp and pyright
		},
		{
			name:        "get unknown language",
			language:    "unknown",
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servers := registry.GetByLanguage(tt.language)
			if tt.expectCount == 0 {
				assert.Empty(t, servers)
			} else {
				assert.Len(t, servers, tt.expectCount)
			}
		})
	}
}

func TestLSPServerRegistry_GetPreferredByLanguage(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Set up servers with different priorities
	registry.mu.Lock()
	registry.servers["test-low"] = &LSPServerDefinition{
		ID:       "test-low",
		Name:     "Test Low Priority",
		Language: "testlang",
		Priority: 50,
		Enabled:  true,
		Binary:   "/usr/bin/test-low",
	}
	registry.servers["test-high"] = &LSPServerDefinition{
		ID:       "test-high",
		Name:     "Test High Priority",
		Language: "testlang",
		Priority: 100,
		Enabled:  true,
		Binary:   "/usr/bin/test-high",
	}
	registry.byLanguage["testlang"] = []*LSPServerDefinition{
		registry.servers["test-low"],
		registry.servers["test-high"],
	}
	registry.mu.Unlock()

	preferred := registry.GetPreferredByLanguage("testlang")
	assert.NotNil(t, preferred)
	assert.Equal(t, "test-high", preferred.ID)

	// Test with no available servers
	noServers := registry.GetPreferredByLanguage("nonexistent")
	assert.Nil(t, noServers)
}

func TestLSPServerRegistry_GetByFilePattern(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Set binary for a few servers to simulate availability
	registry.mu.Lock()
	if server, ok := registry.servers["gopls"]; ok {
		server.Binary = "/usr/bin/gopls"
	}
	if server, ok := registry.servers["pylsp"]; ok {
		server.Binary = "/usr/bin/pylsp"
	}
	registry.mu.Unlock()

	tests := []struct {
		name        string
		filename    string
		expectMatch bool
	}{
		{
			name:        "match go file",
			filename:    "main.go",
			expectMatch: true,
		},
		{
			name:        "match python file",
			filename:    "script.py",
			expectMatch: true,
		},
		{
			name:        "no match for random file",
			filename:    "file.xyz",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servers := registry.GetByFilePattern(tt.filename)
			if tt.expectMatch {
				assert.NotEmpty(t, servers)
			} else {
				assert.Empty(t, servers)
			}
		})
	}
}

func TestLSPServerRegistry_List(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	ids := registry.List()
	assert.NotEmpty(t, ids)
	assert.Contains(t, ids, "gopls")
	assert.Contains(t, ids, "rust-analyzer")
}

func TestLSPServerRegistry_ListAvailable(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Initially no binaries found (since we don't have them installed in test env)
	// Let's manually set one
	registry.mu.Lock()
	if server, ok := registry.servers["gopls"]; ok {
		server.Binary = "/usr/bin/gopls"
	}
	registry.mu.Unlock()

	available := registry.ListAvailable()
	assert.Contains(t, available, "gopls")
}

func TestLSPServerRegistry_ListLanguages(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	languages := registry.ListLanguages()
	assert.NotEmpty(t, languages)
	assert.Contains(t, languages, "go")
	assert.Contains(t, languages, "python")
	assert.Contains(t, languages, "rust")
}

func TestLSPServerRegistry_Register(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	tests := []struct {
		name        string
		server      *LSPServerDefinition
		expectError bool
	}{
		{
			name: "register valid server",
			server: &LSPServerDefinition{
				ID:       "custom-server",
				Name:     "Custom Server",
				Language: "custom",
				Command:  "custom-lsp",
				Enabled:  true,
			},
			expectError: false,
		},
		{
			name: "register without ID",
			server: &LSPServerDefinition{
				Name:    "No ID Server",
				Command: "no-id-lsp",
			},
			expectError: true,
		},
		{
			name: "register without command",
			server: &LSPServerDefinition{
				ID:   "no-command",
				Name: "No Command Server",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.server)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify it was registered
				server, err := registry.Get(tt.server.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.server.ID, server.ID)
			}
		})
	}
}

func TestLSPServerRegistry_Unregister(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Register a custom server first
	server := &LSPServerDefinition{
		ID:       "to-unregister",
		Name:     "To Unregister",
		Language: "custom",
		Command:  "custom-lsp",
	}
	registry.Register(server)

	// Unregister it
	err := registry.Unregister("to-unregister")
	assert.NoError(t, err)

	// Verify it's gone
	_, err = registry.Get("to-unregister")
	assert.Error(t, err)

	// Try to unregister non-existent server
	err = registry.Unregister("nonexistent")
	assert.Error(t, err)
}

func TestLSPServerRegistry_Refresh(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Just verify refresh doesn't panic
	registry.Refresh()
}

func TestLSPServerRegistry_findBinary(t *testing.T) {
	// Create a temporary directory with a mock binary
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "mock-lsp")
	file, err := os.Create(mockBinary)
	require.NoError(t, err)
	file.Close()
	os.Chmod(mockBinary, 0755)

	registry := NewLSPServerRegistry(RegistryConfig{
		SearchPaths: []string{tmpDir},
	})

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "find mock binary",
			command:  "mock-lsp",
			expected: true,
		},
		{
			name:     "find absolute path",
			command:  mockBinary,
			expected: true,
		},
		{
			name:     "not found",
			command:  "nonexistent-binary",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.findBinary(tt.command)
			if tt.expected {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestLSPServerRegistry_findBinary_LookPath(t *testing.T) {
	// Test that we can find system binaries using LookPath
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Try to find a common system binary
	if _, err := exec.LookPath("ls"); err == nil {
		result := registry.findBinary("ls")
		assert.NotEmpty(t, result)
	}
}

func TestLSPServerRegistry_GetStatuses(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	statuses := registry.GetStatuses()
	assert.NotEmpty(t, statuses)

	// Check that statuses have required fields
	for _, status := range statuses {
		assert.NotEmpty(t, status.ID)
		assert.NotEmpty(t, status.Name)
		assert.NotEmpty(t, status.Language)
		assert.False(t, status.CheckedAt.IsZero())
	}
}

func TestLSPServerRegistry_HealthCheck(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Create a temporary mock binary
	tmpDir := t.TempDir()
	mockBinary := filepath.Join(tmpDir, "test-lsp")
	file, err := os.Create(mockBinary)
	require.NoError(t, err)
	file.Close()

	// Register a custom server with the mock binary
	registry.Register(&LSPServerDefinition{
		ID:       "test-health",
		Name:     "Test Health",
		Language: "test",
		Command:  "test-lsp",
		Enabled:  true,
	})

	// Manually set the binary path
	registry.mu.Lock()
	registry.servers["test-health"].Binary = mockBinary
	registry.mu.Unlock()

	results := registry.HealthCheck(context.Background())
	assert.Contains(t, results, "test-health")
	assert.NoError(t, results["test-health"])
}

func TestLSPServerRegistry_HealthCheck_MissingBinary(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Register a server without a binary
	registry.Register(&LSPServerDefinition{
		ID:       "test-missing",
		Name:     "Test Missing",
		Language: "test",
		Command:  "missing-binary",
		Enabled:  true,
	})

	results := registry.HealthCheck(context.Background())
	assert.Error(t, results["test-missing"])
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filePath string
		expected string
	}{
		// Standard extensions
		{"main.go", "go"},
		{"lib.rs", "rust"},
		{"script.py", "python"},
		{"types.pyi", "python"},
		{"app.ts", "typescript"},
		{"app.tsx", "typescript"},
		{"app.js", "javascript"},
		{"app.jsx", "javascript"},
		{"main.c", "c"},
		{"header.h", "c"},
		{"main.cpp", "cpp"},
		{"header.hpp", "cpp"},
		{"main.cc", "cpp"},
		{"main.cxx", "cpp"},
		{"Main.java", "java"},
		{"Program.cs", "csharp"},
		{"index.php", "php"},
		{"app.rb", "ruby"},
		{"task.rake", "ruby"},
		{"app.ex", "elixir"},
		{"app.exs", "elixir"},
		{"Main.hs", "haskell"},
		{"Main.lhs", "haskell"},
		{"script.sh", "bash"},
		{"script.bash", "bash"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"schema.xml", "xml"},
		{"schema.xsd", "xml"},
		{"style.xsl", "xml"},
		{"main.tf", "terraform"},
		{"vars.tfvars", "terraform"},
		{"script.lua", "lua"},

		// Special filenames
		{"Dockerfile", "dockerfile"},
		{"app.dockerfile", "dockerfile"},
		{"Gemfile", "ruby"},
		{".bashrc", "bash"},
		{".bash_profile", "bash"},

		// Unknown
		{"file.xyz", "unknown"},
		{"noextension", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := DetectLanguage(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLSPServerDefinition_Capabilities(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Verify gopls has all expected capabilities
	gopls, err := registry.Get("gopls")
	require.NoError(t, err)

	assert.True(t, gopls.Capabilities.Completion)
	assert.True(t, gopls.Capabilities.Hover)
	assert.True(t, gopls.Capabilities.Definition)
	assert.True(t, gopls.Capabilities.References)
	assert.True(t, gopls.Capabilities.Diagnostics)
	assert.True(t, gopls.Capabilities.Rename)
	assert.True(t, gopls.Capabilities.CodeAction)
	assert.True(t, gopls.Capabilities.Formatting)
	assert.True(t, gopls.Capabilities.SignatureHelp)
}

func TestLSPServerRegistry_Concurrent(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := registry.Get("gopls")
			if err != nil {
				errChan <- err
			}
		}()
	}

	// Concurrent list operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = registry.List()
			_ = registry.ListAvailable()
			_ = registry.ListLanguages()
		}()
	}

	// Concurrent GetByLanguage
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			languages := []string{"go", "python", "rust", "typescript"}
			_ = registry.GetByLanguage(languages[idx%len(languages)])
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestLSPServerRegistry_ServersByLanguage(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Verify Python has multiple servers
	registry.mu.RLock()
	pythonServers := registry.byLanguage["python"]
	registry.mu.RUnlock()

	assert.Len(t, pythonServers, 2) // pylsp and pyright

	// Verify Go has one server
	registry.mu.RLock()
	goServers := registry.byLanguage["go"]
	registry.mu.RUnlock()

	assert.Len(t, goServers, 1) // gopls
}

func TestLSPServerRegistry_GetByFilePattern_Dockerfile(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Set binary for docker-langserver
	registry.mu.Lock()
	if server, ok := registry.servers["docker-langserver"]; ok {
		server.Binary = "/usr/bin/docker-langserver"
	}
	registry.mu.Unlock()

	// Test various Dockerfile patterns
	patterns := []string{"Dockerfile", "Dockerfile.prod", "app.dockerfile"}
	for _, pattern := range patterns {
		servers := registry.GetByFilePattern(pattern)
		assert.NotEmpty(t, servers, "Expected match for %s", pattern)
	}
}

func TestLSPServerRegistry_DisabledServer(t *testing.T) {
	registry := NewLSPServerRegistry(RegistryConfig{})

	// Disable a server
	registry.mu.Lock()
	if server, ok := registry.servers["gopls"]; ok {
		server.Binary = "/usr/bin/gopls" // Has binary
		server.Enabled = false           // But disabled
	}
	registry.mu.Unlock()

	// Should not appear in GetByLanguage
	servers := registry.GetByLanguage("go")
	for _, s := range servers {
		assert.NotEqual(t, "gopls", s.ID)
	}

	// Should not appear in ListAvailable
	available := registry.ListAvailable()
	assert.NotContains(t, available, "gopls")
}

func TestLSPServerRegistry_WithSearchPaths(t *testing.T) {
	// Create temp directories
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Create mock binary in second directory
	mockBinary := filepath.Join(tmpDir2, "custom-lsp")
	file, err := os.Create(mockBinary)
	require.NoError(t, err)
	file.Close()
	os.Chmod(mockBinary, 0755)

	registry := NewLSPServerRegistry(RegistryConfig{
		SearchPaths: []string{tmpDir1, tmpDir2},
	})

	// Register server
	registry.Register(&LSPServerDefinition{
		ID:       "custom",
		Name:     "Custom LSP",
		Language: "custom",
		Command:  "custom-lsp",
		Enabled:  true,
	})

	// Should find the binary
	server, _ := registry.Get("custom")
	assert.Equal(t, mockBinary, server.Binary)
}
