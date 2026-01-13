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
	"dev.helix.agent/internal/models"
)

// mockProvider implements LLMProvider for testing
type mockProvider struct {
	name         string
	healthStatus bool
	completeFunc func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error)
}

func (m *mockProvider) Complete(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, req)
	}
	return &models.LLMResponse{
		Content:      "mock response",
		ProviderName: m.name,
	}, nil
}

func (m *mockProvider) CompleteStream(ctx context.Context, req *models.LLMRequest) (<-chan *models.LLMResponse, error) {
	ch := make(chan *models.LLMResponse, 1)
	go func() {
		defer close(ch)
		ch <- &models.LLMResponse{Content: "mock stream"}
	}()
	return ch, nil
}

func (m *mockProvider) HealthCheck() error {
	if !m.healthStatus {
		return errors.New("unhealthy")
	}
	return nil
}

func (m *mockProvider) GetCapabilities() *models.ProviderCapabilities {
	return &models.ProviderCapabilities{
		SupportsStreaming:       true,
		SupportsVision:          false,
		SupportsFunctionCalling: true,
		Limits: models.ModelLimits{
			MaxTokens: 4096,
		},
	}
}

func (m *mockProvider) ValidateConfig(config map[string]interface{}) (bool, []string) {
	return true, nil
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

	lazy := llm.NewLazyProvider("test", factory, nil)

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

	config := &llm.LazyProviderConfig{
		RetryAttempts: 1,
		InitTimeout:   1 * time.Second,
		RetryDelay:    10 * time.Millisecond,
	}
	lazy := llm.NewLazyProvider("test", factory, config)

	provider, err := lazy.Get()
	assert.Nil(t, provider)
	assert.Error(t, err)

	// Subsequent calls should return the same error
	provider2, err2 := lazy.Get()
	assert.Nil(t, provider2)
	assert.Error(t, err2)
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

	lazy := llm.NewLazyProvider("test", factory, nil)

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
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return &models.LLMResponse{
					Content:      "response to: " + req.Prompt,
					ProviderName: "test",
				}, nil
			},
		}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	req := &models.LLMRequest{Prompt: "hello"}
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
			completeFunc: func(ctx context.Context, req *models.LLMRequest) (*models.LLMResponse, error) {
				return nil, expectedErr
			},
		}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	req := &models.LLMRequest{Prompt: "hello"}
	resp, err := lazy.Complete(context.Background(), req)

	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
}

func TestLazyProvider_CompleteStream(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	req := &models.LLMRequest{Prompt: "hello"}
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

	lazy := llm.NewLazyProvider("test", factory, nil)

	err := lazy.HealthCheck()
	assert.NoError(t, err)
}

func TestLazyProvider_HealthCheckUnhealthy(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: false}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	err := lazy.HealthCheck()
	assert.Error(t, err)
}

func TestLazyProvider_GetCapabilities(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	caps := lazy.GetCapabilities()
	require.NotNil(t, caps)
	assert.Equal(t, 4096, caps.Limits.MaxTokens)
	assert.True(t, caps.SupportsStreaming)
}

func TestLazyProvider_Name(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test-provider", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test-provider", factory, nil)

	name := lazy.Name()
	assert.Equal(t, "test-provider", name)
}

func TestLazyProvider_InitTime(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		time.Sleep(50 * time.Millisecond)
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	// Init time should be zero before initialization
	assert.Equal(t, time.Duration(0), lazy.InitializationTime())

	// Get triggers initialization
	lazy.Get()

	// Init time should be around 50ms
	initTime := lazy.InitializationTime()
	assert.True(t, initTime >= 50*time.Millisecond)
	assert.True(t, initTime < 500*time.Millisecond)
}

func TestLazyProvider_FactoryNotCalledUntilNeeded(t *testing.T) {
	var factoryCalled bool

	factory := func() (llm.LLMProvider, error) {
		factoryCalled = true
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

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
	registry := llm.NewLazyProviderRegistry(nil, nil)

	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	registry.Register("test", factory)

	lazy, ok := registry.Get("test")
	require.True(t, ok)
	require.NotNil(t, lazy)

	provider, err := lazy.Get()
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestLazyProviderRegistry_NotFound(t *testing.T) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

	lazy, ok := registry.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, lazy)
}

func TestLazyProviderRegistry_LazyInit(t *testing.T) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

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
	lazy, ok := registry.Get("provider-a")
	require.True(t, ok)
	lazy.Get()

	// Only that one should be initialized
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCounts[0]))
	assert.Equal(t, int32(0), atomic.LoadInt32(&initCounts[1]))
	assert.Equal(t, int32(0), atomic.LoadInt32(&initCounts[2]))
}

func TestLazyProviderRegistry_Preload(t *testing.T) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

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
	ctx := context.Background()
	err := registry.Preload(ctx, "provider-a", "provider-c")
	require.NoError(t, err)

	// a and c should be initialized, b should not
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCounts[0]))
	assert.Equal(t, int32(0), atomic.LoadInt32(&initCounts[1]))
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCounts[2]))
}

func TestLazyProviderRegistry_List(t *testing.T) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

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

func TestLazyProviderRegistry_GetProvider(t *testing.T) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	registry.Register("test", factory)

	provider, err := registry.GetProvider("test")
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestLazyProviderRegistry_GetProviderNotFound(t *testing.T) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

	provider, err := registry.GetProvider("nonexistent")
	assert.Nil(t, provider)
	assert.Error(t, err)
}

func TestLazyProviderRegistry_InitializedProviders(t *testing.T) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

	for i := 0; i < 3; i++ {
		factory := func() (llm.LLMProvider, error) {
			return &mockProvider{name: "test", healthStatus: true}, nil
		}
		registry.Register("provider-"+string(rune('a'+i)), factory)
	}

	// Initially none initialized
	initialized := registry.InitializedProviders()
	assert.Len(t, initialized, 0)

	// Initialize one
	registry.GetProvider("provider-a")

	initialized = registry.InitializedProviders()
	assert.Len(t, initialized, 1)
	assert.Contains(t, initialized, "provider-a")
}

func TestLazyProvider_Metrics(t *testing.T) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	// Initial metrics
	metrics := lazy.Metrics()
	assert.Equal(t, int64(0), metrics.AccessCount)
	assert.Equal(t, int64(0), metrics.InitializationCount)

	// Access triggers initialization
	lazy.Get()

	metrics = lazy.Metrics()
	assert.Equal(t, int64(1), metrics.AccessCount)
	assert.Equal(t, int64(1), metrics.InitializationCount)

	// Another access
	lazy.Get()
	metrics = lazy.Metrics()
	assert.Equal(t, int64(2), metrics.AccessCount)
	assert.Equal(t, int64(1), metrics.InitializationCount) // Still 1
}

func TestLazyProvider_Reset(t *testing.T) {
	var initCount int32

	factory := func() (llm.LLMProvider, error) {
		atomic.AddInt32(&initCount, 1)
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	// First init
	lazy.Get()
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCount))
	assert.True(t, lazy.IsInitialized())

	// Reset
	lazy.Reset()
	assert.False(t, lazy.IsInitialized())

	// Re-init
	lazy.Get()
	assert.Equal(t, int32(2), atomic.LoadInt32(&initCount))
	assert.True(t, lazy.IsInitialized())
}

func BenchmarkLazyProvider_Get(b *testing.B) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lazy.Get()
	}
}

func BenchmarkLazyProvider_GetParallel(b *testing.B) {
	factory := func() (llm.LLMProvider, error) {
		return &mockProvider{name: "test", healthStatus: true}, nil
	}

	lazy := llm.NewLazyProvider("test", factory, nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lazy.Get()
		}
	})
}

func BenchmarkLazyProviderRegistry_Get(b *testing.B) {
	registry := llm.NewLazyProviderRegistry(nil, nil)

	providers := []string{"provider-a", "provider-b", "provider-c", "provider-d"}
	for _, name := range providers {
		factory := func() (llm.LLMProvider, error) {
			return &mockProvider{name: "test", healthStatus: true}, nil
		}
		registry.Register(name, factory)
	}

	// Preload all
	registry.Preload(context.Background(), providers...)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			registry.Get(providers[i%len(providers)])
			i++
		}
	})
}
