package formatters

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// InitializeFormattersSystem initializes the complete formatters system
// Note: Formatters must be registered after initialization using RegisterFormatter()
func InitializeFormattersSystem(config *Config, logger *logrus.Logger) (*FormatterRegistry, *FormatterExecutor, *HealthChecker, error) {
	// Create registry
	registry := NewFormatterRegistry(config.ToRegistryConfig(), logger)

	// Create executor
	executor := NewFormatterExecutor(registry, config.ToExecutorConfig(), logger)

	// Add middleware
	executor.Use(TimeoutMiddleware(config.DefaultTimeout))
	executor.Use(RetryMiddleware(3))

	if config.CacheEnabled {
		cache := NewFormatterCache(config.ToCacheConfig(), logger)
		executor.Use(CacheMiddleware(cache))
	}

	executor.Use(ValidationMiddleware())

	if config.Metrics {
		executor.Use(MetricsMiddleware())
	}

	if config.Tracing {
		executor.Use(TracingMiddleware())
	}

	// Create health checker
	health := NewHealthChecker(registry, logger, 10*time.Second)

	// Start registry
	ctx := context.Background()
	if err := registry.Start(ctx); err != nil {
		return nil, nil, nil, err
	}

	logger.Info("Formatters system initialized successfully")

	return registry, executor, health, nil
}

// RegisterFormatter is a helper to register a formatter
// This avoids import cycles by having callers pass in the formatter instance
func RegisterFormatter(registry *FormatterRegistry, formatter Formatter, metadata *FormatterMetadata) error {
	return registry.Register(formatter, metadata)
}
