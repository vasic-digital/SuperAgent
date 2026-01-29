package formatters

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// System represents the complete formatters system
type System struct {
	Config   *Config
	Registry *FormatterRegistry
	Executor *FormatterExecutor
	Health   *HealthChecker
	Logger   *logrus.Logger
}

// NewSystem creates and initializes a new formatters system
func NewSystem(config *Config, logger *logrus.Logger) (*System, error) {
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
		return nil, err
	}

	logger.Info("Formatters system initialized successfully")

	return &System{
		Config:   config,
		Registry: registry,
		Executor: executor,
		Health:   health,
		Logger:   logger,
	}, nil
}

// Shutdown gracefully shuts down the formatters system
func (s *System) Shutdown() error {
	ctx := context.Background()
	return s.Registry.Stop(ctx)
}
