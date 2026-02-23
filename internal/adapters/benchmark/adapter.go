// Package benchmark provides an adapter bridging HelixAgent internals to the
// digital.vasic.benchmark extracted module.
package benchmark

import (
	"context"

	"github.com/sirupsen/logrus"

	extbenchmark "digital.vasic.benchmark/benchmark"
)

// Adapter bridges HelixAgent to the digital.vasic.benchmark module.
// It wraps a StandardBenchmarkRunner and BenchmarkSystem, exposing a
// simplified interface for HelixAgent use.
type Adapter struct {
	system *extbenchmark.BenchmarkSystem
	logger *logrus.Logger
}

// New creates a new benchmark Adapter with the provided logger.
// If logger is nil, a default logrus logger is used.
func New(logger *logrus.Logger) *Adapter {
	if logger == nil {
		logger = logrus.New()
	}
	cfg := extbenchmark.DefaultBenchmarkSystemConfig()
	system := extbenchmark.NewBenchmarkSystem(cfg, logger)
	return &Adapter{
		system: system,
		logger: logger,
	}
}

// Initialize sets up the BenchmarkSystem with the given LLM provider.
// providerName and modelName identify the provider for run labelling.
func (a *Adapter) Initialize(providerSvc extbenchmark.ProviderServiceForBenchmark, providerName, modelName string) error {
	pa := extbenchmark.NewProviderAdapterForBenchmark(providerSvc, providerName, modelName, a.logger)
	return a.system.Initialize(pa)
}

// SetDebateService registers a debate service for benchmark evaluation.
func (a *Adapter) SetDebateService(svc extbenchmark.DebateServiceForBenchmark) {
	a.system.SetDebateService(svc)
}

// SetVerifierService registers a verifier service for provider score–based selection.
func (a *Adapter) SetVerifierService(svc extbenchmark.VerifierServiceForBenchmark) {
	a.system.SetVerifierService(svc)
}

// GetRunner returns the underlying BenchmarkRunner for direct use.
func (a *Adapter) GetRunner() extbenchmark.BenchmarkRunner {
	return a.system.GetRunner()
}

// ListBenchmarks returns all available built-in and registered benchmarks.
func (a *Adapter) ListBenchmarks(ctx context.Context) ([]*extbenchmark.Benchmark, error) {
	runner := a.system.GetRunner()
	if runner == nil {
		return nil, nil
	}
	return runner.ListBenchmarks(ctx)
}

// RunBenchmarkWithBestProvider selects the best-scoring provider via the
// registered VerifierService and runs the requested benchmark type.
func (a *Adapter) RunBenchmarkWithBestProvider(
	ctx context.Context,
	benchmarkType extbenchmark.BenchmarkType,
	config *extbenchmark.BenchmarkConfig,
) (*extbenchmark.BenchmarkRun, error) {
	if config == nil {
		config = extbenchmark.DefaultBenchmarkConfig()
	}
	return a.system.RunBenchmarkWithBestProvider(ctx, benchmarkType, config)
}

// CompareProviders runs the benchmark for each named provider and returns the
// set of runs for comparison.
func (a *Adapter) CompareProviders(
	ctx context.Context,
	benchmarkType extbenchmark.BenchmarkType,
	providers []string,
	config *extbenchmark.BenchmarkConfig,
) ([]*extbenchmark.BenchmarkRun, error) {
	if config == nil {
		config = extbenchmark.DefaultBenchmarkConfig()
	}
	return a.system.CompareProviders(ctx, benchmarkType, providers, config)
}

// GenerateLeaderboard builds a leaderboard from all completed runs of the
// given benchmark type.
func (a *Adapter) GenerateLeaderboard(
	ctx context.Context,
	benchmarkType extbenchmark.BenchmarkType,
) (*extbenchmark.Leaderboard, error) {
	return a.system.GenerateLeaderboard(ctx, benchmarkType)
}
