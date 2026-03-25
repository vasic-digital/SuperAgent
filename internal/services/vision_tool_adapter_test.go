package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockVisionProvider implements VisionProvider for testing.
type mockVisionProvider struct {
	analyzeImageDataFn func(ctx context.Context, imageData []byte, prompt string) (string, error)
	analyzeImageURLFn  func(ctx context.Context, imageURL string, prompt string) (string, error)
}

func (m *mockVisionProvider) AnalyzeImageData(
	ctx context.Context,
	imageData []byte,
	prompt string,
) (string, error) {
	if m.analyzeImageDataFn != nil {
		return m.analyzeImageDataFn(ctx, imageData, prompt)
	}
	return "mock image analysis result", nil
}

func (m *mockVisionProvider) AnalyzeImageURL(
	ctx context.Context,
	imageURL string,
	prompt string,
) (string, error) {
	if m.analyzeImageURLFn != nil {
		return m.analyzeImageURLFn(ctx, imageURL, prompt)
	}
	return "mock URL analysis result", nil
}

func TestVisionToolAdapter_NewVisionToolAdapter(t *testing.T) {
	t.Run("creates adapter with provider and logger", func(t *testing.T) {
		provider := &mockVisionProvider{}
		logger := newTestLogger()

		adapter := NewVisionToolAdapter(provider, logger)

		require.NotNil(t, adapter)
		assert.Equal(t, provider, adapter.provider)
		assert.Equal(t, logger, adapter.logger)
	})

	t.Run("creates adapter with nil provider", func(t *testing.T) {
		logger := newTestLogger()

		adapter := NewVisionToolAdapter(nil, logger)

		require.NotNil(t, adapter)
		assert.Nil(t, adapter.provider)
	})

	t.Run("creates adapter with nil logger uses default", func(t *testing.T) {
		provider := &mockVisionProvider{}

		adapter := NewVisionToolAdapter(provider, nil)

		require.NotNil(t, adapter)
		assert.NotNil(t, adapter.logger)
	})
}

func TestVisionToolAdapter_AnalyzeImage(t *testing.T) {
	ctx := context.Background()

	t.Run("successful image analysis", func(t *testing.T) {
		provider := &mockVisionProvider{
			analyzeImageDataFn: func(_ context.Context, data []byte, prompt string) (string, error) {
				return fmt.Sprintf("analyzed %d bytes: %s", len(data), prompt), nil
			},
		}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		result, err := adapter.AnalyzeImage(ctx, []byte("fake-image-data"), "describe this")

		require.NoError(t, err)
		assert.Equal(t, "analyzed 15 bytes: describe this", result)
	})

	t.Run("nil provider returns error", func(t *testing.T) {
		adapter := NewVisionToolAdapter(nil, newTestLogger())

		result, err := adapter.AnalyzeImage(ctx, []byte("fake-data"), "prompt")

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "vision provider not configured")
	})

	t.Run("empty image data returns error", func(t *testing.T) {
		provider := &mockVisionProvider{}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		result, err := adapter.AnalyzeImage(ctx, []byte{}, "prompt")

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "image data is empty")
	})

	t.Run("empty prompt uses default", func(t *testing.T) {
		var capturedPrompt string
		provider := &mockVisionProvider{
			analyzeImageDataFn: func(_ context.Context, _ []byte, prompt string) (string, error) {
				capturedPrompt = prompt
				return "ok", nil
			},
		}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		_, err := adapter.AnalyzeImage(ctx, []byte("data"), "")

		require.NoError(t, err)
		assert.Equal(t, "Describe this image in detail.", capturedPrompt)
	})

	t.Run("provider error is wrapped", func(t *testing.T) {
		provider := &mockVisionProvider{
			analyzeImageDataFn: func(_ context.Context, _ []byte, _ string) (string, error) {
				return "", fmt.Errorf("model unavailable")
			},
		}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		result, err := adapter.AnalyzeImage(ctx, []byte("data"), "prompt")

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "vision analysis failed")
		assert.Contains(t, err.Error(), "model unavailable")
	})
}

func TestVisionToolAdapter_AnalyzeURL(t *testing.T) {
	ctx := context.Background()

	t.Run("successful URL analysis", func(t *testing.T) {
		provider := &mockVisionProvider{
			analyzeImageURLFn: func(_ context.Context, url string, prompt string) (string, error) {
				return fmt.Sprintf("analyzed %s: %s", url, prompt), nil
			},
		}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		result, err := adapter.AnalyzeURL(ctx, "https://example.com/image.png", "describe this")

		require.NoError(t, err)
		assert.Equal(t, "analyzed https://example.com/image.png: describe this", result)
	})

	t.Run("nil provider returns error", func(t *testing.T) {
		adapter := NewVisionToolAdapter(nil, newTestLogger())

		result, err := adapter.AnalyzeURL(ctx, "https://example.com/img.png", "prompt")

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "vision provider not configured")
	})

	t.Run("empty URL returns error", func(t *testing.T) {
		provider := &mockVisionProvider{}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		result, err := adapter.AnalyzeURL(ctx, "", "prompt")

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "image URL is empty")
	})

	t.Run("empty prompt uses default", func(t *testing.T) {
		var capturedPrompt string
		provider := &mockVisionProvider{
			analyzeImageURLFn: func(_ context.Context, _ string, prompt string) (string, error) {
				capturedPrompt = prompt
				return "ok", nil
			},
		}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		_, err := adapter.AnalyzeURL(ctx, "https://example.com/img.png", "")

		require.NoError(t, err)
		assert.Equal(t, "Describe this image in detail.", capturedPrompt)
	})

	t.Run("provider error is wrapped", func(t *testing.T) {
		provider := &mockVisionProvider{
			analyzeImageURLFn: func(_ context.Context, _ string, _ string) (string, error) {
				return "", fmt.Errorf("timeout")
			},
		}
		adapter := NewVisionToolAdapter(provider, newTestLogger())

		result, err := adapter.AnalyzeURL(ctx, "https://example.com/img.png", "prompt")

		assert.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "vision URL analysis failed")
		assert.Contains(t, err.Error(), "timeout")
	})
}

func TestTruncatePrompt(t *testing.T) {
	t.Run("short prompt unchanged", func(t *testing.T) {
		result := truncatePrompt("hello", 100)
		assert.Equal(t, "hello", result)
	})

	t.Run("exact length unchanged", func(t *testing.T) {
		result := truncatePrompt("12345", 5)
		assert.Equal(t, "12345", result)
	})

	t.Run("long prompt truncated with ellipsis", func(t *testing.T) {
		result := truncatePrompt("hello world", 5)
		assert.Equal(t, "hello...", result)
	})

	t.Run("empty prompt unchanged", func(t *testing.T) {
		result := truncatePrompt("", 100)
		assert.Equal(t, "", result)
	})
}
