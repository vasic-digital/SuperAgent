package modelsdev

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	t.Run("creates limiter with correct values", func(t *testing.T) {
		limiter := NewRateLimiter(100, time.Minute)

		require.NotNil(t, limiter)
		assert.Equal(t, 100, limiter.maxTokens)
		assert.Equal(t, 100, limiter.tokens)
		assert.Equal(t, time.Minute, limiter.interval)
	})

	t.Run("creates limiter with small values", func(t *testing.T) {
		limiter := NewRateLimiter(1, time.Second)

		assert.Equal(t, 1, limiter.maxTokens)
		assert.Equal(t, 1, limiter.tokens)
	})
}

func TestRateLimiter_Wait_ContextCancellation(t *testing.T) {
	t.Run("respects context cancellation when exhausted", func(t *testing.T) {
		limiter := NewRateLimiter(1, time.Second)
		ctx, cancel := context.WithCancel(context.Background())

		// Exhaust the token
		err := limiter.Wait(ctx)
		require.NoError(t, err)

		// Cancel context and try to wait
		cancel()
		err = limiter.Wait(ctx)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("respects context timeout", func(t *testing.T) {
		limiter := NewRateLimiter(1, 10*time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Exhaust the token
		err := limiter.Wait(ctx)
		require.NoError(t, err)

		// Try to get another token - should timeout
		err = limiter.Wait(ctx)

		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestRateLimiter_Concurrent(t *testing.T) {
	t.Run("handles concurrent access", func(t *testing.T) {
		limiter := NewRateLimiter(10, time.Minute)
		ctx := context.Background()

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := limiter.Wait(ctx); err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	t.Run("refills tokens after interval", func(t *testing.T) {
		limiter := NewRateLimiter(3, 50*time.Millisecond)
		ctx := context.Background()

		// Exhaust all tokens
		for i := 0; i < 3; i++ {
			err := limiter.Wait(ctx)
			require.NoError(t, err)
		}

		// Verify tokens are exhausted
		assert.Equal(t, 0, limiter.tokens)

		// Wait for refill
		time.Sleep(60 * time.Millisecond)

		// Should be able to get tokens again
		err := limiter.Wait(ctx)
		require.NoError(t, err)

		// Tokens should be refilled to max - 1 (we just took one)
		assert.Equal(t, 2, limiter.tokens)
	})
}

func TestRateLimiter_WaitBlocksWhenExhausted(t *testing.T) {
	t.Run("blocks until tokens available", func(t *testing.T) {
		limiter := NewRateLimiter(1, 100*time.Millisecond)
		ctx := context.Background()

		// Exhaust the token
		err := limiter.Wait(ctx)
		require.NoError(t, err)

		// Start timing
		start := time.Now()

		// This should block until refill
		err = limiter.Wait(ctx)
		require.NoError(t, err)

		elapsed := time.Since(start)

		// Should have waited approximately 100ms (with some tolerance)
		assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond)
		assert.Less(t, elapsed, 200*time.Millisecond)
	})
}

func TestRateLimiter_MultipleRefills(t *testing.T) {
	t.Run("handles multiple refill cycles", func(t *testing.T) {
		limiter := NewRateLimiter(2, 50*time.Millisecond)
		ctx := context.Background()

		// First cycle
		for i := 0; i < 2; i++ {
			err := limiter.Wait(ctx)
			require.NoError(t, err)
		}

		time.Sleep(60 * time.Millisecond)

		// Second cycle
		for i := 0; i < 2; i++ {
			err := limiter.Wait(ctx)
			require.NoError(t, err)
		}

		time.Sleep(60 * time.Millisecond)

		// Third cycle
		for i := 0; i < 2; i++ {
			err := limiter.Wait(ctx)
			require.NoError(t, err)
		}
	})
}

func TestRateLimiter_EdgeCases(t *testing.T) {
	t.Run("single token limiter", func(t *testing.T) {
		limiter := NewRateLimiter(1, 50*time.Millisecond)
		ctx := context.Background()

		err := limiter.Wait(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, limiter.tokens)

		time.Sleep(60 * time.Millisecond)

		err = limiter.Wait(ctx)
		require.NoError(t, err)
	})

	t.Run("high token count", func(t *testing.T) {
		limiter := NewRateLimiter(1000, time.Minute)
		ctx := context.Background()

		for i := 0; i < 100; i++ {
			err := limiter.Wait(ctx)
			require.NoError(t, err)
		}

		assert.Equal(t, 900, limiter.tokens)
	})
}
