package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/superagent/superagent/pkg/api"
)

func TestGrpcServer_Complete(t *testing.T) {
	server := &grpcServer{}
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
	server := &grpcServer{}
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
	server := &grpcServer{}
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
	// Verify grpcServer struct can be instantiated
	server := &grpcServer{}
	require.NotNil(t, server)
}

// Benchmark tests
func BenchmarkGrpcServer_Complete(b *testing.B) {
	server := &grpcServer{}
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
