package llm

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/models"
)

// slowStreamProvider sends responses with a delay, allowing tests to cancel
// mid-stream and verify goroutine cleanup.
type slowStreamProvider struct {
	mu            sync.Mutex
	chunkCount    int
	chunkInterval time.Duration
}

func (p *slowStreamProvider) Complete(_ context.Context, _ *models.LLMRequest) (*models.LLMResponse, error) {
	return &models.LLMResponse{Content: "ok"}, nil
}

func (p *slowStreamProvider) CompleteStream(_ context.Context, _ *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	p.mu.Lock()
	count := p.chunkCount
	interval := p.chunkInterval
	p.mu.Unlock()

	ch := make(chan *models.LLMResponse)
	go func() {
		defer close(ch)
		for i := 0; i < count; i++ {
			if interval > 0 {
				time.Sleep(interval)
			}
			ch <- &models.LLMResponse{Content: "chunk"}
		}
	}()
	return ch, nil
}

func (p *slowStreamProvider) HealthCheck() error {
	return nil
}

func (p *slowStreamProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{}
}

func (p *slowStreamProvider) ValidateConfig(_ map[string]interface{}) (bool, []string) {
	return true, nil
}

// TestCircuitBreaker_CompleteStream_NoGoroutineLeak verifies that after a
// stream completes normally, the forwarding goroutine exits and does not leak.
func TestCircuitBreaker_CompleteStream_NoGoroutineLeak(t *testing.T) {
	// Let the runtime settle before capturing baseline.
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	provider := &slowStreamProvider{chunkCount: 5, chunkInterval: 10 * time.Millisecond}
	cb := NewDefaultCircuitBreaker("leak-test-normal", provider)
	req := &models.LLMRequest{ID: "test"}

	ch, err := cb.CompleteStream(context.Background(), req)
	require.NoError(t, err)

	// Drain the channel completely.
	count := 0
	for range ch {
		count++
	}
	assert.Equal(t, 5, count, "should receive all 5 chunks")

	// Allow goroutines to finish.
	time.Sleep(300 * time.Millisecond)
	runtime.GC()

	after := runtime.NumGoroutine()
	assert.LessOrEqual(t, after, baseline+3,
		"goroutine count should return near baseline after normal stream completion (baseline=%d, after=%d)", baseline, after)

	// Verify circuit breaker recorded a success.
	stats := cb.GetStats()
	assert.Equal(t, int64(1), stats.TotalSuccesses, "stream should be recorded as success")
}

// longStreamProvider sends chunks continuously until its stop channel is
// closed. Unlike a truly hanging provider, its goroutine can be cleaned up
// by the test so we can accurately measure CB wrapper goroutine leaks.
type longStreamProvider struct {
	mu   sync.Mutex
	stop chan struct{} // closed by the test to stop all active streams
}

func newLongStreamProvider() *longStreamProvider {
	return &longStreamProvider{stop: make(chan struct{})}
}

func (p *longStreamProvider) Complete(_ context.Context, _ *models.LLMRequest) (*models.LLMResponse, error) {
	return &models.LLMResponse{Content: "ok"}, nil
}

func (p *longStreamProvider) CompleteStream(_ context.Context, _ *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	p.mu.Lock()
	stop := p.stop
	p.mu.Unlock()

	ch := make(chan *models.LLMResponse)
	go func() {
		defer close(ch)
		for {
			select {
			case <-stop:
				return
			case ch <- &models.LLMResponse{Content: "chunk"}:
			}
		}
	}()
	return ch, nil
}

func (p *longStreamProvider) HealthCheck() error                                      { return nil }
func (p *longStreamProvider) GetCapabilities() *models.ProviderCapabilities            { return &models.ProviderCapabilities{} }
func (p *longStreamProvider) ValidateConfig(_ map[string]interface{}) (bool, []string) { return true, nil }

// TestCircuitBreaker_CompleteStream_ContextCancelNoLeak verifies that when the
// caller cancels the context mid-stream, the forwarding goroutine exits
// promptly and does not leak. We open 10 streams and cancel them all; without
// the fix each leaked stream leaves at least 1 goroutine stuck on wrappedCh,
// so the goroutine count would grow by 10+ above baseline.
func TestCircuitBreaker_CompleteStream_ContextCancelNoLeak(t *testing.T) {
	// Let the runtime settle before capturing baseline.
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	const iterations = 10
	provider := newLongStreamProvider()
	// Use a high failure threshold so context cancellation errors don't
	// trip the circuit open during the test loop.
	cfg := CircuitBreakerConfig{
		FailureThreshold:    iterations + 5,
		SuccessThreshold:    2,
		Timeout:             30 * time.Second,
		HalfOpenMaxRequests: 3,
	}
	cb := NewCircuitBreaker("leak-test-cancel", provider, cfg)
	req := &models.LLMRequest{ID: "test"}

	for i := 0; i < iterations; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		ch, err := cb.CompleteStream(ctx, req)
		require.NoError(t, err)

		// Read one chunk to prove the stream is active, then cancel.
		resp, ok := <-ch
		require.True(t, ok, "should receive at least one chunk")
		require.NotNil(t, resp)

		cancel()

		// Brief pause to let goroutine observe cancellation.
		time.Sleep(20 * time.Millisecond)
	}

	// Stop the provider so its goroutines exit, isolating CB wrapper leaks.
	close(provider.stop)

	// Allow goroutines to clean up after the last cancel.
	time.Sleep(300 * time.Millisecond)
	runtime.GC()

	after := runtime.NumGoroutine()
	// With the fix, the wrapper goroutines exit on ctx.Done(). Without the
	// fix, each iteration leaks at least 1 goroutine (stuck on wrappedCh send),
	// so after would be baseline + 10+. We allow a margin of 3 for runtime jitter.
	assert.LessOrEqual(t, after, baseline+3,
		"goroutine count should return near baseline after %d cancelled streams (baseline=%d, after=%d)",
		iterations, baseline, after)
}
