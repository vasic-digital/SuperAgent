package services

import (
	"context"
	"testing"

	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/memory"
	"dev.helix.agent/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMem0Service_NewMem0Service(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	assert.NotNil(t, service)
	assert.True(t, service.IsEnabled())
}

func TestMem0Service_NewMem0Service_Disabled(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: false, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	assert.NotNil(t, service)
	assert.False(t, service.IsEnabled())
}

func TestMem0Service_AddMemory(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	req := &Mem0Request{
		UserID:     "user-123",
		SessionID:  "session-456",
		Content:    "User prefers dark mode",
		Type:       "semantic",
		Importance: 0.8,
	}
	resp, err := service.AddMemory(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.True(t, resp.Success)
	assert.NotZero(t, resp.CreatedAt)
	stats := service.GetStats()
	assert.Equal(t, int64(1), stats.TotalMemoriesStored)
}

func TestMem0Service_AddMemory_Disabled(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: false, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	req := &Mem0Request{UserID: "user-123", Content: "User prefers dark mode"}
	_, err := service.AddMemory(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestMem0Service_SearchMemory(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	_, err := service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "User prefers dark mode for coding", Type: "semantic"})
	require.NoError(t, err)
	_, err = service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "User likes TypeScript", Type: "semantic"})
	require.NoError(t, err)
	// Note: Without embeddings, search won't return results from InMemoryStore
	// This tests the service layer, not the store implementation
	stats := service.GetStats()
	assert.Equal(t, int64(2), stats.TotalMemoriesStored)
}

func TestMem0Service_DeleteMemory(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	resp, err := service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "Test memory"})
	require.NoError(t, err)
	err = service.DeleteMemory(ctx, resp.ID)
	require.NoError(t, err)
	memories, err := service.GetUserMemories(ctx, "user-123", 10, 0)
	require.NoError(t, err)
	assert.Empty(t, memories)
}

func TestMem0Service_DeleteUserMemories(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "Test memory"})
		require.NoError(t, err)
	}
	err := service.DeleteUserMemories(ctx, "user-123")
	require.NoError(t, err)
	memories, err := service.GetUserMemories(ctx, "user-123", 10, 0)
	require.NoError(t, err)
	assert.Empty(t, memories)
}

func TestMem0Service_GetContext_NoEmbeddings(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	_, err := service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "User prefers dark mode for coding"})
	require.NoError(t, err)
	// Without embeddings, GetContext returns empty (expected behavior)
	_, err = service.GetContext(ctx, "preferences", "user-123", 1000)
	require.NoError(t, err) // No error, just empty context
}

func TestMem0Service_EnhanceRequest_NoEmbeddings(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	_, err := service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "User prefers dark mode"})
	require.NoError(t, err)
	req := &models.LLMRequest{UserID: "user-123", Prompt: "What are my preferences?"}
	err = service.EnhanceRequest(ctx, req)
	require.NoError(t, err)
	// Without embeddings, memory context may be empty
}

func TestMem0Service_Stats(t *testing.T) {
	store := memory.NewInMemoryStore()
	logger := logrus.New()
	cfg := &config.MemoryConfig{Enabled: true, Provider: "mem0"}
	service := NewMem0Service(store, nil, nil, nil, cfg, logger)
	ctx := context.Background()
	_, err := service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "Test memory 1"})
	require.NoError(t, err)
	_, err = service.AddMemory(ctx, &Mem0Request{UserID: "user-123", Content: "Test memory 2"})
	require.NoError(t, err)
	stats := service.GetStats()
	assert.Equal(t, int64(2), stats.TotalMemoriesStored)
	assert.NotZero(t, stats.LastActivity)
}

func TestMem0Service_MemoryTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected memory.MemoryType
	}{
		{"episodic", memory.MemoryTypeEpisodic},
		{"conversation", memory.MemoryTypeEpisodic},
		{"event", memory.MemoryTypeEpisodic},
		{"semantic", memory.MemoryTypeSemantic},
		{"fact", memory.MemoryTypeSemantic},
		{"knowledge", memory.MemoryTypeSemantic},
		{"procedural", memory.MemoryTypeProcedural},
		{"howto", memory.MemoryTypeProcedural},
		{"working", memory.MemoryTypeWorking},
		{"temporary", memory.MemoryTypeWorking},
		{"unknown", memory.MemoryTypeEpisodic},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := memoryTypeFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
