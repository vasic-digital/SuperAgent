package database

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// ProtocolRepository Extended Tests
// Tests for all CRUD operations on MCP, LSP, ACP servers, embedding config,
// protocol cache, and protocol metrics using nil pool + panic recovery.
// =============================================================================

// -----------------------------------------------------------------------------
// MCP Server Operations
// -----------------------------------------------------------------------------

func TestProtocolRepository_CreateMCPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	server := &MCPServer{
		ID:      "mcp-1",
		Name:    "filesystem",
		Type:    "local",
		Enabled: true,
		Tools:   json.RawMessage(`["read","write"]`),
	}

	err := safeCallError(func() error {
		return repo.CreateMCPServer(context.Background(), server)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_GetMCPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() (*MCPServer, error) {
		return repo.GetMCPServer(context.Background(), "mcp-1")
	})
	assert.Error(t, err)
}

func TestProtocolRepository_ListMCPServers_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*MCPServer, error) {
		return repo.ListMCPServers(context.Background(), false)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_ListMCPServers_EnabledOnly_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*MCPServer, error) {
		return repo.ListMCPServers(context.Background(), true)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_UpdateMCPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	server := &MCPServer{
		ID:      "mcp-1",
		Name:    "filesystem-updated",
		Type:    "local",
		Enabled: false,
	}

	err := safeCallError(func() error {
		return repo.UpdateMCPServer(context.Background(), server)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_DeleteMCPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	err := safeCallError(func() error {
		return repo.DeleteMCPServer(context.Background(), "mcp-1")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// LSP Server Operations
// -----------------------------------------------------------------------------

func TestProtocolRepository_CreateLSPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	server := &LSPServer{
		ID:           "lsp-1",
		Name:         "gopls",
		Language:     "go",
		Command:      "gopls",
		Enabled:      true,
		Workspace:    "/workspace",
		Capabilities: json.RawMessage(`{"hover":true}`),
	}

	err := safeCallError(func() error {
		return repo.CreateLSPServer(context.Background(), server)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_GetLSPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() (*LSPServer, error) {
		return repo.GetLSPServer(context.Background(), "lsp-1")
	})
	assert.Error(t, err)
}

func TestProtocolRepository_ListLSPServers_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*LSPServer, error) {
		return repo.ListLSPServers(context.Background(), false)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_ListLSPServers_EnabledOnly_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*LSPServer, error) {
		return repo.ListLSPServers(context.Background(), true)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_UpdateLSPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	server := &LSPServer{
		ID:       "lsp-1",
		Name:     "gopls-updated",
		Language: "go",
		Command:  "gopls serve",
		Enabled:  true,
	}

	err := safeCallError(func() error {
		return repo.UpdateLSPServer(context.Background(), server)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_DeleteLSPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	err := safeCallError(func() error {
		return repo.DeleteLSPServer(context.Background(), "lsp-1")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// ACP Server Operations
// -----------------------------------------------------------------------------

func TestProtocolRepository_CreateACPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	url := "http://localhost:4096"
	server := &ACPServer{
		ID:      "acp-1",
		Name:    "opencode",
		Type:    "remote",
		URL:     &url,
		Enabled: true,
		Tools:   json.RawMessage(`["execute"]`),
	}

	err := safeCallError(func() error {
		return repo.CreateACPServer(context.Background(), server)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_GetACPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() (*ACPServer, error) {
		return repo.GetACPServer(context.Background(), "acp-1")
	})
	assert.Error(t, err)
}

func TestProtocolRepository_ListACPServers_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*ACPServer, error) {
		return repo.ListACPServers(context.Background(), false)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_ListACPServers_EnabledOnly_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*ACPServer, error) {
		return repo.ListACPServers(context.Background(), true)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_UpdateACPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	server := &ACPServer{
		ID:      "acp-1",
		Name:    "opencode-updated",
		Type:    "remote",
		Enabled: false,
	}

	err := safeCallError(func() error {
		return repo.UpdateACPServer(context.Background(), server)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_DeleteACPServer_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	err := safeCallError(func() error {
		return repo.DeleteACPServer(context.Background(), "acp-1")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Embedding Config Operations
// -----------------------------------------------------------------------------

func TestProtocolRepository_GetEmbeddingConfig_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() (*EmbeddingConfig, error) {
		return repo.GetEmbeddingConfig(context.Background())
	})
	assert.Error(t, err)
}

func TestProtocolRepository_UpdateEmbeddingConfig_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	cfg := &EmbeddingConfig{
		ID:        1,
		Provider:  "openai",
		Model:     "text-embedding-3-small",
		Dimension: 1536,
	}

	err := safeCallError(func() error {
		return repo.UpdateEmbeddingConfig(context.Background(), cfg)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Protocol Cache Operations
// -----------------------------------------------------------------------------

func TestProtocolRepository_SetCache_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	err := safeCallError(func() error {
		return repo.SetCache(context.Background(), "test-key", json.RawMessage(`{"data":"test"}`), 5*time.Minute)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_GetCache_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() (*ProtocolCache, error) {
		return repo.GetCache(context.Background(), "test-key")
	})
	assert.Error(t, err)
}

func TestProtocolRepository_DeleteCache_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	err := safeCallError(func() error {
		return repo.DeleteCache(context.Background(), "test-key")
	})
	assert.Error(t, err)
}

func TestProtocolRepository_ClearExpiredCache_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() (int64, error) {
		return repo.ClearExpiredCache(context.Background())
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Protocol Metrics Operations
// -----------------------------------------------------------------------------

func TestProtocolRepository_RecordMetric_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)
	durationMs := 150
	metric := &ProtocolMetrics{
		ProtocolType: "mcp",
		Operation:    "tool_call",
		Status:       "success",
		DurationMs:   &durationMs,
		Metadata:     json.RawMessage(`{"tool":"read_file"}`),
	}

	err := safeCallError(func() error {
		return repo.RecordMetric(context.Background(), metric)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_GetMetrics_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*ProtocolMetrics, error) {
		return repo.GetMetrics(context.Background(), "mcp", time.Now().Add(-24*time.Hour), 100)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_GetMetrics_EmptyProtocolType_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() ([]*ProtocolMetrics, error) {
		return repo.GetMetrics(context.Background(), "", time.Now().Add(-24*time.Hour), 0)
	})
	assert.Error(t, err)
}

func TestProtocolRepository_GetMetricsSummary_NilPool(t *testing.T) {
	repo := NewProtocolRepository(nil)

	_, err := safeCallResult(func() (map[string]interface{}, error) {
		return repo.GetMetricsSummary(context.Background(), "mcp", time.Now().Add(-24*time.Hour))
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Model Struct Tests
// -----------------------------------------------------------------------------

func TestMCPServer_Fields(t *testing.T) {
	cmd := "npx -y @modelcontextprotocol/server-filesystem"
	server := &MCPServer{
		ID:      "mcp-fs",
		Name:    "filesystem",
		Type:    "local",
		Command: &cmd,
		Enabled: true,
		Tools:   json.RawMessage(`["read_file","write_file"]`),
	}

	assert.Equal(t, "mcp-fs", server.ID)
	assert.Equal(t, "filesystem", server.Name)
	assert.Equal(t, "local", server.Type)
	assert.NotNil(t, server.Command)
	assert.True(t, server.Enabled)
	assert.Nil(t, server.URL)
}

func TestLSPServer_Fields(t *testing.T) {
	server := &LSPServer{
		ID:           "lsp-gopls",
		Name:         "gopls",
		Language:     "go",
		Command:      "gopls serve",
		Enabled:      true,
		Workspace:    "/src",
		Capabilities: json.RawMessage(`{"completion":true,"hover":true}`),
	}

	assert.Equal(t, "go", server.Language)
	assert.Equal(t, "gopls serve", server.Command)
}

func TestACPServer_Fields(t *testing.T) {
	url := "http://localhost:4096"
	server := &ACPServer{
		ID:      "acp-opencode",
		Name:    "opencode",
		Type:    "remote",
		URL:     &url,
		Enabled: true,
	}

	assert.Equal(t, "remote", server.Type)
	assert.NotNil(t, server.URL)
}

func TestEmbeddingConfig_Fields(t *testing.T) {
	endpoint := "https://api.openai.com/v1/embeddings"
	key := "sk-test"
	cfg := &EmbeddingConfig{
		ID:          1,
		Provider:    "openai",
		Model:       "text-embedding-3-small",
		Dimension:   1536,
		APIEndpoint: &endpoint,
		APIKey:      &key,
	}

	assert.Equal(t, 1536, cfg.Dimension)
	assert.Equal(t, "openai", cfg.Provider)
}

func TestProtocolCache_Fields(t *testing.T) {
	cache := &ProtocolCache{
		CacheKey:  "mcp:servers:list",
		CacheData: json.RawMessage(`[{"id":"1"}]`),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	assert.Equal(t, "mcp:servers:list", cache.CacheKey)
	assert.NotNil(t, cache.CacheData)
}

func TestProtocolMetrics_Fields(t *testing.T) {
	serverID := "mcp-fs"
	durationMs := 250
	errMsg := "connection timeout"
	metric := &ProtocolMetrics{
		ID:           1,
		ProtocolType: "mcp",
		ServerID:     &serverID,
		Operation:    "tool_call",
		Status:       "error",
		DurationMs:   &durationMs,
		ErrorMessage: &errMsg,
		Metadata:     json.RawMessage(`{"tool":"read_file"}`),
	}

	assert.Equal(t, "mcp", metric.ProtocolType)
	assert.Equal(t, "error", metric.Status)
	assert.NotNil(t, metric.ErrorMessage)
}
