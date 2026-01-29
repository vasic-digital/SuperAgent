package formatters

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// FormatterExecutor executes formatting requests with middleware
type FormatterExecutor struct {
	registry   *FormatterRegistry
	cache      *FormatterCache
	middleware []Middleware
	logger     *logrus.Logger
	config     *ExecutorConfig
}

// ExecutorConfig configures the executor
type ExecutorConfig struct {
	DefaultTimeout time.Duration
	MaxRetries     int
	EnableCache    bool
	EnableMetrics  bool
	EnableTracing  bool
}

// Middleware wraps formatter execution
type Middleware func(next ExecuteFunc) ExecuteFunc

// ExecuteFunc is the execution function signature
type ExecuteFunc func(ctx context.Context, formatter Formatter, req *FormatRequest) (*FormatResult, error)

// NewFormatterExecutor creates a new formatter executor
func NewFormatterExecutor(registry *FormatterRegistry, config *ExecutorConfig, logger *logrus.Logger) *FormatterExecutor {
	executor := &FormatterExecutor{
		registry:   registry,
		logger:     logger,
		config:     config,
		middleware: make([]Middleware, 0),
	}

	// Initialize cache if enabled
	if config.EnableCache {
		executor.cache = NewFormatterCache(&CacheConfig{
			TTL:         3600 * time.Second,
			MaxSize:     10000,
			CleanupFreq: 300 * time.Second,
		}, logger)
	}

	return executor
}

// Execute executes a formatting request
func (e *FormatterExecutor) Execute(ctx context.Context, req *FormatRequest) (*FormatResult, error) {
	start := time.Now()

	// Determine formatter
	var formatter Formatter
	var err error

	if req.Language != "" {
		// Get preferred formatter for language
		formatters := e.registry.GetByLanguage(req.Language)
		if len(formatters) == 0 {
			return nil, fmt.Errorf("no formatters available for language: %s", req.Language)
		}
		formatter = formatters[0]
	} else if req.FilePath != "" {
		// Detect formatter from file path
		formatter, err = e.registry.DetectFormatter(req.FilePath, req.Content)
		if err != nil {
			return nil, fmt.Errorf("unable to detect formatter: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either language or file_path must be specified")
	}

	e.logger.Debugf("Executing formatter %s for language %s", formatter.Name(), req.Language)

	// Build execution chain with middleware
	executeFunc := e.buildChain(formatter)

	// Execute with chain
	result, err := executeFunc(ctx, formatter, req)
	if err != nil {
		e.logger.Errorf("Formatter execution failed: %v", err)
		return nil, err
	}

	result.Duration = time.Since(start)

	e.logger.Debugf("Formatter execution completed in %v: changed=%v", result.Duration, result.Changed)

	return result, nil
}

// ExecuteBatch executes multiple formatting requests
func (e *FormatterExecutor) ExecuteBatch(ctx context.Context, reqs []*FormatRequest) ([]*FormatResult, error) {
	results := make([]*FormatResult, len(reqs))
	errors := make([]error, len(reqs))

	// Execute all requests in parallel
	type resultPair struct {
		index  int
		result *FormatResult
		err    error
	}

	resultChan := make(chan resultPair, len(reqs))

	for i, req := range reqs {
		go func(index int, request *FormatRequest) {
			result, err := e.Execute(ctx, request)
			resultChan <- resultPair{
				index:  index,
				result: result,
				err:    err,
			}
		}(i, req)
	}

	// Collect results
	for i := 0; i < len(reqs); i++ {
		pair := <-resultChan
		results[pair.index] = pair.result
		errors[pair.index] = pair.err
	}

	// Check if any failed
	var firstError error
	for _, err := range errors {
		if err != nil {
			firstError = err
			break
		}
	}

	return results, firstError
}

// Use adds middleware to the execution chain
func (e *FormatterExecutor) Use(middleware ...Middleware) {
	e.middleware = append(e.middleware, middleware...)
}

// buildChain builds the execution chain with middleware
func (e *FormatterExecutor) buildChain(formatter Formatter) ExecuteFunc {
	// Base execution function
	base := func(ctx context.Context, f Formatter, req *FormatRequest) (*FormatResult, error) {
		return f.Format(ctx, req)
	}

	// Apply middleware in reverse order
	for i := len(e.middleware) - 1; i >= 0; i-- {
		base = e.middleware[i](base)
	}

	return base
}

// TimeoutMiddleware adds timeout handling
func TimeoutMiddleware(defaultTimeout time.Duration) Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, formatter Formatter, req *FormatRequest) (*FormatResult, error) {
			timeout := req.Timeout
			if timeout == 0 {
				timeout = defaultTimeout
			}

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			resultChan := make(chan *FormatResult, 1)
			errorChan := make(chan error, 1)

			go func() {
				result, err := next(ctx, formatter, req)
				if err != nil {
					errorChan <- err
				} else {
					resultChan <- result
				}
			}()

			select {
			case result := <-resultChan:
				return result, nil
			case err := <-errorChan:
				return nil, err
			case <-ctx.Done():
				return nil, fmt.Errorf("formatter execution timed out after %v", timeout)
			}
		}
	}
}

// RetryMiddleware adds retry logic
func RetryMiddleware(maxRetries int) Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, formatter Formatter, req *FormatRequest) (*FormatResult, error) {
			var lastErr error

			for attempt := 0; attempt <= maxRetries; attempt++ {
				result, err := next(ctx, formatter, req)
				if err == nil {
					return result, nil
				}

				lastErr = err

				// Don't retry on context cancellation
				if ctx.Err() != nil {
					break
				}

				// Wait before retry (exponential backoff)
				if attempt < maxRetries {
					waitTime := time.Duration(1<<uint(attempt)) * time.Second
					time.Sleep(waitTime)
				}
			}

			return nil, fmt.Errorf("formatter execution failed after %d retries: %w", maxRetries, lastErr)
		}
	}
}

// CacheMiddleware adds caching
func CacheMiddleware(cache *FormatterCache) Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, formatter Formatter, req *FormatRequest) (*FormatResult, error) {
			// Don't cache check-only requests
			if req.CheckOnly {
				return next(ctx, formatter, req)
			}

			// Check cache
			if cached, found := cache.Get(req); found {
				return cached, nil
			}

			// Execute
			result, err := next(ctx, formatter, req)
			if err != nil {
				return nil, err
			}

			// Store in cache
			cache.Set(req, result)

			return result, nil
		}
	}
}

// ValidationMiddleware adds pre/post validation
func ValidationMiddleware() Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, formatter Formatter, req *FormatRequest) (*FormatResult, error) {
			// Pre-validation
			if req.Content == "" {
				return nil, fmt.Errorf("empty content provided")
			}

			// Execute
			result, err := next(ctx, formatter, req)
			if err != nil {
				return nil, err
			}

			// Post-validation
			if result.Success && result.Content == "" {
				return nil, fmt.Errorf("formatter returned empty content")
			}

			return result, nil
		}
	}
}

// MetricsMiddleware adds metrics collection
func MetricsMiddleware() Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, formatter Formatter, req *FormatRequest) (*FormatResult, error) {
			start := time.Now()

			result, err := next(ctx, formatter, req)

			duration := time.Since(start)

			// TODO: Collect metrics
			// - formatter name
			// - language
			// - duration
			// - success/failure
			// - bytes processed

			_ = duration // Suppress unused warning for now

			return result, err
		}
	}
}

// TracingMiddleware adds distributed tracing
func TracingMiddleware() Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, formatter Formatter, req *FormatRequest) (*FormatResult, error) {
			// TODO: Start span
			// span := opentelemetry.StartSpan(ctx, "formatter.execute")
			// defer span.End()

			result, err := next(ctx, formatter, req)

			// TODO: Record span attributes
			// span.SetAttributes(
			//   "formatter.name", formatter.Name(),
			//   "formatter.language", req.Language,
			//   "formatter.success", err == nil,
			// )

			return result, err
		}
	}
}
