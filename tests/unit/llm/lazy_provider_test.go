package llm

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/llm"
)

// mockProvider implements LLMProvider for testing
type mockProvider struct {
	name         string
	healthStatus bool
	completeFunc func(ctx context.Context, req *llm.LLMRequest) (*llm.LLMResponse, error)
}

func (m *mockProvider) Complete(ctx context.Context, req *llm.LLMRequest) (*llm.LLMResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, req)
	}
	return &llm.LLMResponse{
		Content: "mock response",
		Model:   m.name,
	}, nil
}

func (m *mockProvider) CompleteStream(ctx context.Context, req *llm.LLMRequest) (<-chan *llm.StreamChunk, error) {
	ch := make(chan *llm.StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- &llm.StreamChunk{Content: "mock stream"}
	}()
	return ch, nil
}

func (m *mockProvider) HealthCheck(ctx context.Context) error {
	if !m.healthStatus {
		return errors.New("unhealthy")
	}
	return nil
}

func (m *mockProvider) GetCapabilities() *llm.ProviderCapabilities {
	return &llm.ProviderCapabilities{
		MaxTokens:  4096,
		Streaming:  true,
		Vision:     false,
		Functions:  true,
	}
}

func (m *mockProvider) ValidateConfig() error {
	return nil
}

func (m *mockProvider) Name() string {
	return m.name
}

func TestLazyProvider_BasicOperation(t *testing.T) {
	var initCount int32

	factory := func() (llm.LLMProvider, error) {
		atomic.AddInt32(&initCount, 1)
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	// Should not be initialized yet
	assert.False(t, lazy.IsInitialized())
	assert.Equal(t, int32(0), atomic.LoadInt32(&initCount))

	// Get the provider
	provider, err := lazy.Get()
	require.NoError(t, err)
	require.NotNil(t, provider)

	// Should be initialized now
	assert.True(t, lazy.IsInitialized())
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCount))

	// Second get should not reinitialize
	provider2, err := lazy.Get()
	require.NoError(t, err)
	assert.Equal(t, provider, provider2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCount))
}

func TestLazyProvider_FactoryError(t *testing.T) {
	expectedErr := errors.New("factory error")

	factory := func() (llm.LLMProvider, error) {
		return nil, expectedErr
	}

	lazy := llm.NewLazyProvider(factory)

	provider, err := lazy.Get()
	assert.Nil(t, provider)
	assert.Equal(t, expectedErr, err)

	// Subsequent calls should return the same error
	provider2, err2 := lazy.Get()
	assert.Nil(t, provider2)
	assert.Equal(t, expectedErr, err2)
}

func TestLazyProvider_ConcurrentAccess(t *testing.T) {
	var initCount int32
	initStarted := make(chan struct{})
	initDone := make(chan struct{})

	factory := func() (llm.LLMProvider, error) {
		close(initStarted)
		<-initDone // Wait for signal to complete
		atomic.AddInt32(&initCount, 1)
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	var wg sync.WaitGroup
	results := make(chan llm.LLMProvider, 100)

	// Start many concurrent getters
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			provider, err := lazy.Get()
			if err == nil {
				results <- provider
			}
		}()
	}

	// Wait for initialization to start
	<-initStarted

	// Let initialization complete
	close(initDone)

	wg.Wait()
	close(results)

	// All should get the same provider
	var firstProvider llm.LLMProvider
	count := 0
	for provider := range results {
		count++
		if firstProvider == nil {
			firstProvider = provider
		} else {
			assert.Equal(t, firstProvider, provider)
		}
	}

	assert.Equal(t, 100, count)
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCount))
}

func TestLazyProvider_Complete(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{
			name:         "test",
			healthStatus: true,
			completeFunc: func(ctx context.Context, req *llm.LLMRequest) (*llm.LLMResponse, error) {
				return &llm.LLMResponse{
					Content: "response to: " + req.Prompt,
					Model:   "test",
				}, nil
			},
		}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	req := &llm.LLMRequest{Prompt: "hello"}
	resp, err := lazy.Complete(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "response to: hello", resp.Content)
}

func TestLazyProvider_CompleteWithError(t *testing.T) {
	expectedErr := errors.New("completion error")

	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{
			name:         "test",
			healthStatus: true,
			completeFunc: func(ctx context.Context, req *llm.LLMRequest) (*llm.LLMResponse, error) {
				return nil, expectedErr
			},
		}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	req := &llm.LLMRequest{Prompt: "hello"}
	resp, err := lazy.Complete(context.Background(), req)

	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
}

func TestLazyProvider_CompleteStream(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	req := &llm.LLMRequest{Prompt: "hello"}
	ch, err := lazy.CompleteStream(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, ch)

	// Read from the stream
	var chunks []string
	for chunk := range ch {
		chunks = append(chunks, chunk.Content)
	}

	assert.Contains(t, chunks, "mock stream")
}

func TestLazyProvider_HealthCheck(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	err := lazy.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestLazyProvider_HealthCheckUnhealthy(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: false}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	err := lazy.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestLazyProvider_GetCapabilities(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	caps := lazy.GetCapabilities()
	require.NotNil(t, caps)
	assert.Equal(t, 4096, caps.MaxTokens)
	assert.True(t, caps.Streaming)
}

func TestLazyProvider_Name(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test-provider", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	name := lazy.Name()
	assert.Equal(t, "test-provider", name)
}

func TestLazyProvider_InitTime(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		time.Sleep(50 * time.Millisecond)
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	// Init time should be zero before initialization
	assert.Equal(t, time.Duration(0), lazy.InitTime())

	// Get triggers initialization
	lazy.Get()

	// Init time should be around 50ms
	initTime := lazy.InitTime()
	assert.True(t, initTime >= 50*time.Millisecond)
	assert.True(t, initTime < 200*time.Millisecond)
}

func TestLazyProvider_FactoryNotCalledUntilNeeded(t *testing.T) {
	var factoryCalled bool

	factory := func() (llm.LLMProvider, error) {
		factoryCalled = true
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	// Just creating the lazy provider should not call factory
	assert.False(t, factoryCalled)

	// Checking IsInitialized should not call factory
	assert.False(t, lazy.IsInitialized())
	assert.False(t, factoryCalled)

	// Getting provider should call factory
	lazy.Get()
	assert.True(t, factoryCalled)
}

func TestLazyProviderRegistry_BasicOperation(t *testing.T) {
	registry := llm.NewLazyProviderRegistry()

	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	registry.Register("test", factory)

	provider, err := registry.Get("test")
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestLazyProviderRegistry_NotFound(t *testing.T) {
	registry := llm.NewLazyProviderRegistry()

	provider, err := registry.Get("nonexistent")
	assert.Nil(t, provider)
	assert.Error(t, err)
}

func TestLazyProviderRegistry_LazyInit(t *testing.T) {
	registry := llm.NewLazyProviderRegistry()

	var initCounts [3]int32

	for i := 0; i < 3; i++ {
		idx := i
		factory := func() (llm.LLMProvider, error) {
			atomic.AddInt32(&initCounts[idx], 1)
			return &mockProvider{name: "test", healthStatus: true}, nil
		}
		registry.Register("provider-"+string(rune('a'+i)), factory)
	}

	// None should be initialized yet
	for i := 0; i < 3; i++ {
		assert.Equal(t, int32(0), atomic.LoadInt32(&initCounts[i]))
	}

	// Get only one provider
	registry.Get("provider-a")

	// Only that one should be initialized
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCounts[0]))
	assert.Equal(t, int32(0), atomic.LoadInt32(&initCounts[1]))
	assert.Equal(t, int32(0), atomic.LoadInt32(&initCounts[2]))
}

func TestLazyProviderRegistry_Preload(t *testing.T) {
	registry := llm.NewLazyProviderRegistry()

	var initCounts [3]int32

	for i := 0; i < 3; i++ {
		idx := i
		factory := func() (llm.LLMProvider, error) {
			atomic.AddInt32(&initCounts[idx], 1)
			return &mockProvider{name: "test", healthStatus: true}, nil
		}
		registry.Register("provider-"+string(rune('a'+idx)), factory)
	}

	// Preload specific providers
	err := registry.Preload("provider-a", "provider-c")
	require.NoError(t, err)

	// a and c should be initialized, b should not
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCounts[0]))
	assert.Equal(t, int32(0), atomic.LoadInt32(&initCounts[1]))
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCounts[2]))
}

func TestLazyProviderRegistry_List(t *testing.T) {
	registry := llm.NewLazyProviderRegistry()

	for i := 0; i < 3; i++ {
		factory := func() (llm.LLMProvider, error) {
			return &mockProvider{name: "test", healthStatus: true}, nil
		}
		registry.Register("provider-"+string(rune('a'+i)), factory)
	}

	names := registry.List()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "provider-a")
	assert.Contains(t, names, "provider-b")
	assert.Contains(t, names, "provider-c")
}

func BenchmarkLazyProvider_Get(b *testing.B) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lazy.Get()
	}
}

func BenchmarkLazyProvider_GetParallel(b *testing.B) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider(factory)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lazy.Get()
		}
	})
}

func BenchmarkLazyProviderRegistry_Get(b *testing.B) {
	registry := llm.NewLazyProviderRegistry()

	providers := []string{"provider-a", "provider-b", "provider-c", "provider-d"}
	for _, name := range providers {
		factory := func() (llm.LLMProvider, error) {
			return &mockProvider{name: "test", healthStatus: true}, nil
		}
		registry.Register(name, factory)
	}

	// Preload all
	registry.Preload(providers...)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			registry.Get(providers[i%len(providers)])
			i++
		}
	})
}
