// Package lazy provides lazy initialization patterns for HelixAgent
// This package implements thread-safe lazy loading for expensive resources
package lazy

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Initializer defines the interface for lazy initialization
type Initializer[T any] interface {
	Get(ctx context.Context) (T, error)
	IsInitialized() bool
	Reset()
}

// Loader implements thread-safe lazy loading
type Loader[T any] struct {
	mu          sync.RWMutex
	instance    T
	factory     func() (T, error)
	initialized bool
	err         error
	initTime    time.Time
	lastAccess  time.Time
	metrics     *Metrics
	ttl         time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

// Metrics tracks lazy loader performance
type Metrics struct {
	InitCount     int64
	InitErrors    int64
	AccessCount   int64
	AvgInitTime   time.Duration
	TotalInitTime time.Duration
}

// Config configures lazy loader behavior
type Config struct {
	// TTL for cached instances (0 = no expiration)
	TTL time.Duration

	// MaxRetries for initialization (0 = unlimited)
	MaxRetries int

	// RetryDelay between retries
	RetryDelay time.Duration

	// EnableMetrics tracks performance metrics
	EnableMetrics bool
}

// New creates a new lazy loader with the given factory
func New[T any](factory func() (T, error), config *Config) *Loader[T] {
	if config == nil {
		config = &Config{}
	}

	ctx, cancel := context.WithCancel(context.Background())

	loader := &Loader[T]{
		factory: factory,
		ttl:     config.TTL,
		ctx:     ctx,
		cancel:  cancel,
	}

	if config.EnableMetrics {
		loader.metrics = &Metrics{}
	}

	return loader
}

// Get returns the initialized instance, creating it if necessary
func (l *Loader[T]) Get(ctx context.Context) (T, error) {
	// Fast path: check if already initialized
	l.mu.RLock()
	if l.initialized && l.err == nil {
		// Check TTL expiration
		if l.ttl > 0 && time.Since(l.initTime) > l.ttl {
			l.mu.RUnlock()
			l.Reset()
			return l.Get(ctx)
		}

		instance := l.instance
		l.lastAccess = time.Now()
		if l.metrics != nil {
			l.metrics.AccessCount++
		}
		l.mu.RUnlock()
		return instance, nil
	}
	l.mu.RUnlock()

	// Slow path: initialize
	return l.initialize(ctx)
}

// initialize performs the actual initialization
func (l *Loader[T]) initialize(ctx context.Context) (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if l.initialized {
		return l.instance, l.err
	}

	start := time.Now()

	// Perform initialization with context support
	done := make(chan struct{})
	go func() {
		l.instance, l.err = l.factory()
		close(done)
	}()

	select {
	case <-done:
		// Initialization completed
	case <-ctx.Done():
		return l.instance, fmt.Errorf("initialization cancelled: %w", ctx.Err())
	case <-l.ctx.Done():
		return l.instance, fmt.Errorf("loader closed: %w", l.ctx.Err())
	}

	if l.err == nil {
		l.initialized = true
		l.initTime = time.Now()
		l.lastAccess = l.initTime
	}

	// Update metrics
	if l.metrics != nil {
		l.metrics.InitCount++
		initDuration := time.Since(start)
		l.metrics.TotalInitTime += initDuration
		l.metrics.AvgInitTime = l.metrics.TotalInitTime / time.Duration(l.metrics.InitCount)

		if l.err != nil {
			l.metrics.InitErrors++
		}
	}

	return l.instance, l.err
}

// IsInitialized returns true if the instance has been initialized
func (l *Loader[T]) IsInitialized() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.initialized && l.err == nil
}

// Reset clears the initialization state, allowing re-initialization
func (l *Loader[T]) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	var zero T
	l.instance = zero
	l.initialized = false
	l.err = nil
	l.initTime = time.Time{}
}

// Close shuts down the loader and prevents further initialization
func (l *Loader[T]) Close() error {
	l.cancel()
	return nil
}

// GetMetrics returns loader metrics (if enabled)
func (l *Loader[T]) GetMetrics() *Metrics {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.metrics == nil {
		return nil
	}

	// Return a copy
	return &Metrics{
		InitCount:     l.metrics.InitCount,
		InitErrors:    l.metrics.InitErrors,
		AccessCount:   l.metrics.AccessCount,
		AvgInitTime:   l.metrics.AvgInitTime,
		TotalInitTime: l.metrics.TotalInitTime,
	}
}

// Registry manages multiple lazy loaders
type Registry struct {
	mu      sync.RWMutex
	loaders map[string]interface{}
}

// NewRegistry creates a new lazy loader registry
func NewRegistry() *Registry {
	return &Registry{
		loaders: make(map[string]interface{}),
	}
}

// Register adds a lazy loader to the registry
func (r *Registry) Register(name string, loader interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loaders[name] = loader
}

// Get retrieves a lazy loader by name
func (r *Registry) Get(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loader, ok := r.loaders[name]
	return loader, ok
}

// InitializeAll initializes all registered loaders concurrently
func (r *Registry) InitializeAll(ctx context.Context) error {
	r.mu.RLock()
	loaders := make(map[string]interface{}, len(r.loaders))
	for k, v := range r.loaders {
		loaders[k] = v
	}
	r.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(loaders))

	for name, loader := range loaders {
		wg.Add(1)
		go func(n string, l interface{}) {
			defer wg.Done()

			switch typed := l.(type) {
			case *Loader[interface{}]:
				if _, err := typed.Get(ctx); err != nil {
					errChan <- fmt.Errorf("failed to initialize %s: %w", n, err)
				}
			}
		}(name, loader)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to initialize %d loaders", len(errs))
	}

	return nil
}

// CloseAll closes all registered loaders
func (r *Registry) CloseAll() error {
	r.mu.RLock()
	loaders := make(map[string]interface{}, len(r.loaders))
	for k, v := range r.loaders {
		loaders[k] = v
	}
	r.mu.RUnlock()

	var errs []error
	for name, loader := range loaders {
		switch typed := loader.(type) {
		case *Loader[interface{}]:
			if err := typed.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close %s: %w", name, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d loaders", len(errs))
	}

	return nil
}

// Warmup pre-initializes the loader to avoid first-access latency
func (l *Loader[T]) Warmup(ctx context.Context) error {
	_, err := l.Get(ctx)
	return err
}

// WaitFor blocks until the loader is initialized or context is cancelled
func (l *Loader[T]) WaitFor(ctx context.Context) error {
	for {
		if l.IsInitialized() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// Continue polling
		}
	}
}
