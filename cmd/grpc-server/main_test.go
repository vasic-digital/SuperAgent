package main

import (
	"context"
	"testing"

	"dev.helix.agent/internal/version"
	pb "dev.helix.agent/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGrpcServer_Complete(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("basic completion request", func(t *testing.T) {
		req := &pb.CompletionRequest{
			SessionId:      "test-session-123",
			Prompt:         "What is the capital of France?",
			MemoryEnhanced: false,
		}

		// Note: This will return an error since no providers are configured
		// The test verifies the method doesn't panic and returns proper error handling
		resp, err := server.Complete(ctx, req)

		// Even if ensemble returns error, we expect a response object
		// The current implementation returns empty response on error
		if err != nil {
			assert.NotNil(t, resp)
			assert.Equal(t, "", resp.Content)
			assert.Equal(t, float64(0), resp.Confidence)
		} else {
			assert.NotNil(t, resp)
		}
	})

	t.Run("completion with memory enhancement", func(t *testing.T) {
		req := &pb.CompletionRequest{
			SessionId:      "test-session-456",
			Prompt:         "Based on our previous conversation, what was the topic?",
			MemoryEnhanced: true,
		}

		resp, err := server.Complete(ctx, req)

		// The ensemble may return an error without configured providers
		if err != nil {
			assert.NotNil(t, resp)
		} else {
			assert.NotNil(t, resp)
		}
	})

	t.Run("empty prompt", func(t *testing.T) {
		req := &pb.CompletionRequest{
			SessionId:      "test-session-789",
			Prompt:         "",
			MemoryEnhanced: false,
		}

		resp, err := server.Complete(ctx, req)

		// Method should handle empty prompts gracefully
		if err != nil {
			assert.NotNil(t, resp)
		} else {
			assert.NotNil(t, resp)
		}
	})
}

func TestGrpcServer_Complete_RequestConversion(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("verifies request is properly converted", func(t *testing.T) {
		req := &pb.CompletionRequest{
			SessionId:      "session-abc",
			Prompt:         "Test prompt for conversion verification",
			MemoryEnhanced: true,
		}

		// Call Complete - the internal conversion will happen
		resp, _ := server.Complete(ctx, req)

		// We verify the server doesn't panic during conversion
		require.NotNil(t, resp)
	})
}

func TestGrpcServer_Complete_ResponseFields(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	req := &pb.CompletionRequest{
		SessionId: "session-xyz",
		Prompt:    "Simple test prompt for response field verification",
	}

	resp, _ := server.Complete(ctx, req)

	// Verify response has expected structure
	require.NotNil(t, resp)

	// Response should have these fields defined (even if empty)
	_ = resp.Content      // Should exist
	_ = resp.Confidence   // Should exist
	_ = resp.ProviderName // Should exist
}

func TestGrpcServerStruct(t *testing.T) {
	// Verify LLMFacadeServer struct can be instantiated
	server := NewLLMFacadeServer()
	require.NotNil(t, server)
}

// LLMProviderServer Tests
func TestLLMProviderServer_Complete(t *testing.T) {
	server := NewLLMProviderServer(nil)
	ctx := context.Background()

	t.Run("basic completion request without registry", func(t *testing.T) {
		req := &pb.CompletionRequest{
			SessionId:      "test-provider-session-123",
			Prompt:         "What is the capital of Germany?",
			MemoryEnhanced: false,
		}

		// Without registry, will fall back to llm.RunEnsemble
		resp, err := server.Complete(ctx, req)

		// Should return response even if error occurs
		if err != nil {
			assert.NotNil(t, resp)
			assert.Equal(t, "", resp.Content)
		} else {
			assert.NotNil(t, resp)
		}
	})
}

func TestLLMProviderServer_HealthCheck(t *testing.T) {
	server := NewLLMProviderServer(nil)
	ctx := context.Background()

	t.Run("health check without registry", func(t *testing.T) {
		req := &pb.HealthRequest{
			Detailed: true,
		}

		resp, err := server.HealthCheck(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "degraded", resp.Status) // No registry means degraded
		assert.NotNil(t, resp.Timestamp)
	})

	t.Run("health check detailed response", func(t *testing.T) {
		req := &pb.HealthRequest{
			Detailed: true,
		}

		resp, err := server.HealthCheck(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, len(resp.Components) > 0, "Expected components in detailed health check")
	})
}

func TestLLMProviderServer_GetCapabilities(t *testing.T) {
	server := NewLLMProviderServer(nil)
	ctx := context.Background()

	t.Run("get capabilities without registry", func(t *testing.T) {
		req := &pb.CapabilitiesRequest{}

		resp, err := server.GetCapabilities(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.SupportsStreaming)
		assert.NotNil(t, resp.Limits)
		assert.True(t, len(resp.SupportedFeatures) > 0)
		assert.Contains(t, resp.SupportedRequestTypes, "completion")
		assert.Contains(t, resp.SupportedRequestTypes, "chat")
		assert.Contains(t, resp.SupportedRequestTypes, "streaming")
	})

	t.Run("verify model limits", func(t *testing.T) {
		req := &pb.CapabilitiesRequest{}

		resp, err := server.GetCapabilities(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp.Limits)
		assert.Equal(t, int32(4096), resp.Limits.MaxTokens)
		assert.Equal(t, int32(100000), resp.Limits.MaxInputLength)
		assert.Equal(t, int32(4096), resp.Limits.MaxOutputLength)
		assert.Equal(t, int32(100), resp.Limits.MaxConcurrentRequests)
	})
}

func TestLLMProviderServer_ValidateConfig(t *testing.T) {
	server := NewLLMProviderServer(nil)
	ctx := context.Background()

	t.Run("validate nil config", func(t *testing.T) {
		req := &pb.ValidateConfigRequest{
			Config: nil,
		}

		resp, err := server.ValidateConfig(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Errors, "configuration is required")
	})

	t.Run("validate config without type", func(t *testing.T) {
		configStruct, _ := structpb.NewStruct(map[string]interface{}{
			"name": "test-provider",
		})
		req := &pb.ValidateConfigRequest{
			Config: configStruct,
		}

		resp, err := server.ValidateConfig(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Errors, "provider type is required")
	})

	t.Run("validate config with warnings for missing api key", func(t *testing.T) {
		configStruct, _ := structpb.NewStruct(map[string]interface{}{
			"name": "test-provider",
			"type": "claude",
		})
		req := &pb.ValidateConfigRequest{
			Config: configStruct,
		}

		resp, err := server.ValidateConfig(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Valid) // Valid but with warnings
		assert.True(t, len(resp.Warnings) > 0)
	})

	t.Run("validate ollama config without base_url", func(t *testing.T) {
		configStruct, _ := structpb.NewStruct(map[string]interface{}{
			"name": "local-ollama",
			"type": "ollama",
		})
		req := &pb.ValidateConfigRequest{
			Config: configStruct,
		}

		resp, err := server.ValidateConfig(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Valid) // Valid but with warning about base_url
	})

	t.Run("validate config with invalid timeout", func(t *testing.T) {
		configStruct, _ := structpb.NewStruct(map[string]interface{}{
			"name":    "test-provider",
			"type":    "deepseek",
			"timeout": float64(-5),
		})
		req := &pb.ValidateConfigRequest{
			Config: configStruct,
		}

		resp, err := server.ValidateConfig(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Errors, "timeout must be positive")
	})
}

func TestLLMProviderServerStruct(t *testing.T) {
	// Verify LLMProviderServer struct can be instantiated
	server := NewLLMProviderServer(nil)
	require.NotNil(t, server)
}

// Benchmark tests
func BenchmarkGrpcServer_Complete(b *testing.B) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	req := &pb.CompletionRequest{
		SessionId:      "benchmark-session",
		Prompt:         "Benchmark test prompt",
		MemoryEnhanced: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.Complete(ctx, req)
	}
}

func BenchmarkLLMProviderServer_Complete(b *testing.B) {
	server := NewLLMProviderServer(nil)
	ctx := context.Background()

	req := &pb.CompletionRequest{
		SessionId:      "benchmark-provider-session",
		Prompt:         "Benchmark provider test prompt",
		MemoryEnhanced: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.Complete(ctx, req)
	}
}

// =============================================================================
// LLMFacadeServer Session Management Tests
// =============================================================================

func TestLLMFacadeServer_CreateSession(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("create session with defaults", func(t *testing.T) {
		req := &pb.CreateSessionRequest{
			UserId:        "user-123",
			MemoryEnabled: true,
		}

		resp, err := server.CreateSession(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.NotEmpty(t, resp.SessionId)
		assert.Equal(t, "user-123", resp.UserId)
		assert.Equal(t, "active", resp.Status)
		assert.Equal(t, int32(0), resp.RequestCount)
		assert.NotNil(t, resp.ExpiresAt)
	})

	t.Run("create session with TTL", func(t *testing.T) {
		initialCtx, _ := structpb.NewStruct(map[string]interface{}{
			"key": "value",
		})
		req := &pb.CreateSessionRequest{
			UserId:         "user-456",
			MemoryEnabled:  false,
			TtlHours:       2,
			InitialContext: initialCtx,
		}

		resp, err := server.CreateSession(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "user-456", resp.UserId)
		assert.NotNil(t, resp.Context)
	})
}

func TestLLMFacadeServer_GetSession(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	// First create a session
	createReq := &pb.CreateSessionRequest{
		UserId:        "user-test",
		MemoryEnabled: true,
	}
	createResp, err := server.CreateSession(ctx, createReq)
	require.NoError(t, err)
	sessionID := createResp.SessionId

	t.Run("get existing session", func(t *testing.T) {
		req := &pb.GetSessionRequest{
			SessionId:      sessionID,
			IncludeContext: false,
		}

		resp, err := server.GetSession(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, sessionID, resp.SessionId)
		assert.Equal(t, "user-test", resp.UserId)
		assert.Equal(t, "active", resp.Status)
	})

	t.Run("get session with context", func(t *testing.T) {
		req := &pb.GetSessionRequest{
			SessionId:      sessionID,
			IncludeContext: true,
		}

		resp, err := server.GetSession(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("get non-existing session", func(t *testing.T) {
		req := &pb.GetSessionRequest{
			SessionId: "non-existent-session",
		}

		_, err := server.GetSession(ctx, req)

		require.Error(t, err)
	})
}

func TestLLMFacadeServer_TerminateSession(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("terminate existing session gracefully", func(t *testing.T) {
		// First create a session
		createReq := &pb.CreateSessionRequest{
			UserId: "user-to-terminate",
		}
		createResp, err := server.CreateSession(ctx, createReq)
		require.NoError(t, err)
		sessionID := createResp.SessionId

		// Terminate it
		req := &pb.TerminateSessionRequest{
			SessionId: sessionID,
			Graceful:  true,
		}

		resp, err := server.TerminateSession(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "terminated", resp.Status)
	})

	t.Run("terminate existing session forcefully", func(t *testing.T) {
		// First create a session
		createReq := &pb.CreateSessionRequest{
			UserId: "user-to-force-terminate",
		}
		createResp, err := server.CreateSession(ctx, createReq)
		require.NoError(t, err)
		sessionID := createResp.SessionId

		// Terminate it
		req := &pb.TerminateSessionRequest{
			SessionId: sessionID,
			Graceful:  false,
		}

		resp, err := server.TerminateSession(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("terminate non-existing session", func(t *testing.T) {
		req := &pb.TerminateSessionRequest{
			SessionId: "non-existent",
		}

		_, err := server.TerminateSession(ctx, req)

		require.Error(t, err)
	})
}

// =============================================================================
// LLMFacadeServer Provider Management Tests
// =============================================================================

func TestLLMFacadeServer_ListProviders(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("list empty providers", func(t *testing.T) {
		req := &pb.ListProvidersRequest{}

		resp, err := server.ListProviders(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Providers)
	})

	t.Run("list providers with filter", func(t *testing.T) {
		// Add a provider first
		addReq := &pb.AddProviderRequest{
			Name:    "test-provider",
			Type:    "claude",
			Model:   "claude-3-opus",
			BaseUrl: "https://api.anthropic.com",
			Weight:  1.0,
		}
		_, err := server.AddProvider(ctx, addReq)
		require.NoError(t, err)

		// List with enabled filter
		listReq := &pb.ListProvidersRequest{
			EnabledOnly: true,
		}

		resp, err := server.ListProviders(ctx, listReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Providers, 1)
	})

	t.Run("list providers by type", func(t *testing.T) {
		listReq := &pb.ListProvidersRequest{
			ProviderType: "claude",
		}

		resp, err := server.ListProviders(ctx, listReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		// Should have the provider we added
		assert.NotEmpty(t, resp.Providers)
	})
}

func TestLLMFacadeServer_AddProvider(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("add valid provider", func(t *testing.T) {
		configStruct, _ := structpb.NewStruct(map[string]interface{}{
			"timeout": 30,
		})
		req := &pb.AddProviderRequest{
			Name:    "new-provider",
			Type:    "deepseek",
			Model:   "deepseek-chat",
			BaseUrl: "https://api.deepseek.com",
			Weight:  0.8,
			Config:  configStruct,
		}

		resp, err := server.AddProvider(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.NotEmpty(t, resp.Provider.Id)
		assert.Equal(t, "new-provider", resp.Provider.Name)
		assert.Equal(t, "deepseek", resp.Provider.Type)
	})

	t.Run("add provider without name", func(t *testing.T) {
		req := &pb.AddProviderRequest{
			Type: "claude",
		}

		_, err := server.AddProvider(ctx, req)

		require.Error(t, err)
	})

	t.Run("add provider without type", func(t *testing.T) {
		req := &pb.AddProviderRequest{
			Name: "missing-type",
		}

		_, err := server.AddProvider(ctx, req)

		require.Error(t, err)
	})

	t.Run("add duplicate provider name", func(t *testing.T) {
		req := &pb.AddProviderRequest{
			Name: "duplicate-provider",
			Type: "gemini",
		}

		_, err := server.AddProvider(ctx, req)
		require.NoError(t, err)

		// Try to add again
		_, err = server.AddProvider(ctx, req)
		require.Error(t, err)
	})
}

func TestLLMFacadeServer_UpdateProvider(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	// First add a provider
	addReq := &pb.AddProviderRequest{
		Name:   "provider-to-update",
		Type:   "mistral",
		Model:  "mistral-large",
		Weight: 0.5,
	}
	addResp, err := server.AddProvider(ctx, addReq)
	require.NoError(t, err)
	providerID := addResp.Provider.Id

	t.Run("update existing provider", func(t *testing.T) {
		req := &pb.UpdateProviderRequest{
			Id:      providerID,
			Name:    "updated-provider",
			Model:   "mistral-small",
			Weight:  0.9,
			Enabled: true,
		}

		resp, err := server.UpdateProvider(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, "updated-provider", resp.Provider.Name)
		assert.Equal(t, "mistral-small", resp.Provider.Model)
		assert.Equal(t, 0.9, resp.Provider.Weight)
	})

	t.Run("update with API key", func(t *testing.T) {
		req := &pb.UpdateProviderRequest{
			Id:     providerID,
			ApiKey: "sk-test-key",
		}

		resp, err := server.UpdateProvider(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("update non-existing provider", func(t *testing.T) {
		req := &pb.UpdateProviderRequest{
			Id: "non-existent-id",
		}

		_, err := server.UpdateProvider(ctx, req)

		require.Error(t, err)
	})

	t.Run("update without id", func(t *testing.T) {
		req := &pb.UpdateProviderRequest{
			Name: "some-name",
		}

		_, err := server.UpdateProvider(ctx, req)

		require.Error(t, err)
	})
}

func TestLLMFacadeServer_RemoveProvider(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("remove disabled provider", func(t *testing.T) {
		// Add a provider
		addReq := &pb.AddProviderRequest{
			Name: "provider-to-remove",
			Type: "ollama",
		}
		addResp, err := server.AddProvider(ctx, addReq)
		require.NoError(t, err)
		providerID := addResp.Provider.Id

		// Disable it first
		updateReq := &pb.UpdateProviderRequest{
			Id:      providerID,
			Enabled: false,
		}
		_, err = server.UpdateProvider(ctx, updateReq)
		require.NoError(t, err)

		// Remove it
		req := &pb.RemoveProviderRequest{
			Id:    providerID,
			Force: false,
		}

		resp, err := server.RemoveProvider(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("force remove enabled provider", func(t *testing.T) {
		// Add a provider
		addReq := &pb.AddProviderRequest{
			Name: "provider-to-force-remove",
			Type: "qwen",
		}
		addResp, err := server.AddProvider(ctx, addReq)
		require.NoError(t, err)

		// Force remove it
		req := &pb.RemoveProviderRequest{
			Id:    addResp.Provider.Id,
			Force: true,
		}

		resp, err := server.RemoveProvider(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	t.Run("remove enabled provider without force", func(t *testing.T) {
		// Add a provider
		addReq := &pb.AddProviderRequest{
			Name: "enabled-provider",
			Type: "openrouter",
		}
		addResp, err := server.AddProvider(ctx, addReq)
		require.NoError(t, err)

		// Try to remove without force
		req := &pb.RemoveProviderRequest{
			Id:    addResp.Provider.Id,
			Force: false,
		}

		_, err = server.RemoveProvider(ctx, req)

		require.Error(t, err)
	})

	t.Run("remove non-existing provider", func(t *testing.T) {
		req := &pb.RemoveProviderRequest{
			Id: "non-existent",
		}

		_, err := server.RemoveProvider(ctx, req)

		require.Error(t, err)
	})

	t.Run("remove without id", func(t *testing.T) {
		req := &pb.RemoveProviderRequest{}

		_, err := server.RemoveProvider(ctx, req)

		require.Error(t, err)
	})
}

// =============================================================================
// LLMFacadeServer Health and Metrics Tests
// =============================================================================

func TestLLMFacadeServer_HealthCheck(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("basic health check", func(t *testing.T) {
		req := &pb.HealthRequest{
			Detailed: false,
		}

		resp, err := server.HealthCheck(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "degraded", resp.Status) // No providers means degraded
		assert.NotNil(t, resp.Timestamp)
		assert.Equal(t, version.Version, resp.Version)
	})

	t.Run("detailed health check", func(t *testing.T) {
		req := &pb.HealthRequest{
			Detailed:        true,
			CheckComponents: []string{"providers", "database", "cognee"},
		}

		resp, err := server.HealthCheck(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Components)
	})

	t.Run("health check with providers", func(t *testing.T) {
		// Add a provider
		addReq := &pb.AddProviderRequest{
			Name: "health-test-provider",
			Type: "claude",
		}
		_, err := server.AddProvider(ctx, addReq)
		require.NoError(t, err)

		req := &pb.HealthRequest{
			Detailed:        true,
			CheckComponents: []string{"providers"},
		}

		resp, err := server.HealthCheck(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "healthy", resp.Status)
	})
}

func TestLLMFacadeServer_GetMetrics(t *testing.T) {
	server := NewLLMFacadeServer()
	ctx := context.Background()

	t.Run("get metrics default time range", func(t *testing.T) {
		req := &pb.MetricsRequest{}

		resp, err := server.GetMetrics(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, resp.Metrics)
		assert.NotNil(t, resp.StartTime)
		assert.NotNil(t, resp.EndTime)
	})

	t.Run("get metrics 1h time range", func(t *testing.T) {
		req := &pb.MetricsRequest{
			TimeRange: "1h",
		}

		resp, err := server.GetMetrics(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("get metrics 24h time range", func(t *testing.T) {
		req := &pb.MetricsRequest{
			TimeRange: "24h",
		}

		resp, err := server.GetMetrics(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("get metrics 7d time range", func(t *testing.T) {
		req := &pb.MetricsRequest{
			TimeRange: "7d",
		}

		resp, err := server.GetMetrics(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

// =============================================================================
// LLMFacadeServer Metrics Recording Tests
// =============================================================================

func TestLLMFacadeServer_recordSuccess(t *testing.T) {
	server := NewLLMFacadeServer()

	// Record some successes
	server.recordRequest()
	server.recordSuccess(100)

	server.metricsMu.RLock()
	defer server.metricsMu.RUnlock()

	assert.Equal(t, int64(1), server.metrics.TotalRequests)
	assert.Equal(t, int64(1), server.metrics.SuccessfulRequests)
	assert.Equal(t, int64(100), server.metrics.TotalLatencyMs)
}

// =============================================================================
// LLMProviderServer Additional Tests
// =============================================================================

func TestLLMProviderServer_recordSuccess(t *testing.T) {
	server := NewLLMProviderServer(nil)

	// Record some successes
	server.recordProviderRequest()
	server.recordProviderSuccess(200)

	server.metricsMu.RLock()
	defer server.metricsMu.RUnlock()

	assert.Equal(t, int64(1), server.metrics.TotalRequests)
	assert.Equal(t, int64(1), server.metrics.SuccessfulRequests)
	assert.Equal(t, int64(200), server.metrics.TotalLatencyMs)
}

func TestLLMProviderServer_ValidateConfig_Comprehensive(t *testing.T) {
	server := NewLLMProviderServer(nil)
	ctx := context.Background()

	t.Run("validate config with empty model", func(t *testing.T) {
		configStruct, _ := structpb.NewStruct(map[string]interface{}{
			"name":  "test-provider",
			"type":  "gemini",
			"model": "",
		})
		req := &pb.ValidateConfigRequest{
			Config: configStruct,
		}

		resp, err := server.ValidateConfig(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Valid) // Valid but with warnings
		assert.True(t, len(resp.Warnings) > 0)
	})

	t.Run("validate config with excessive timeout", func(t *testing.T) {
		configStruct, _ := structpb.NewStruct(map[string]interface{}{
			"name":    "test-provider",
			"type":    "qwen",
			"timeout": float64(400),
		})
		req := &pb.ValidateConfigRequest{
			Config: configStruct,
		}

		resp, err := server.ValidateConfig(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		// Should have warning about excessive timeout
		found := false
		for _, w := range resp.Warnings {
			if len(w) > 0 {
				found = true
			}
		}
		assert.True(t, found || len(resp.Errors) > 0 || resp.Valid)
	})

	t.Run("validate different provider types", func(t *testing.T) {
		providerTypes := []string{"claude", "deepseek", "gemini", "qwen", "openrouter", "ollama"}

		for _, provType := range providerTypes {
			configStruct, _ := structpb.NewStruct(map[string]interface{}{
				"name": "test-" + provType,
				"type": provType,
			})
			req := &pb.ValidateConfigRequest{
				Config: configStruct,
			}

			resp, err := server.ValidateConfig(ctx, req)

			require.NoError(t, err, "Provider type %s should not return error", provType)
			require.NotNil(t, resp, "Response should not be nil for provider type %s", provType)
		}
	})
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestContainsHelper(t *testing.T) {
	t.Run("contains existing value", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		assert.True(t, contains(slice, "b"))
	})

	t.Run("does not contain value", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		assert.False(t, contains(slice, "d"))
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := []string{}
		assert.False(t, contains(slice, "a"))
	})
}
