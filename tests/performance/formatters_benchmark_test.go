//go:build performance
// +build performance

package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/formatters"
)

// =============================================================================
// Formatter benchmark helpers
// =============================================================================

// benchFormatter is a minimal in-process formatter used exclusively for
// benchmarks.  It performs no I/O and returns the input content unchanged so
// that measured latency reflects registry/executor overhead alone.
type benchFormatter struct {
	*formatters.BaseFormatter
}

func newBenchFormatter(name string, languages []string) *benchFormatter {
	meta := &formatters.FormatterMetadata{
		Name:          name,
		Type:          formatters.FormatterTypeNative,
		Version:       "1.0.0",
		Languages:     languages,
		SupportsStdin: true,
		SupportsCheck: true,
	}
	return &benchFormatter{BaseFormatter: formatters.NewBaseFormatter(meta)}
}

func (f *benchFormatter) Format(
	_ context.Context, req *formatters.FormatRequest,
) (*formatters.FormatResult, error) {
	return &formatters.FormatResult{
		Content:          req.Content,
		Changed:          false,
		FormatterName:    f.Name(),
		FormatterVersion: f.Version(),
		Success:          true,
	}, nil
}

func (f *benchFormatter) FormatBatch(
	ctx context.Context, reqs []*formatters.FormatRequest,
) ([]*formatters.FormatResult, error) {
	out := make([]*formatters.FormatResult, len(reqs))
	for i, r := range reqs {
		res, err := f.Format(ctx, r)
		if err != nil {
			return nil, err
		}
		out[i] = res
	}
	return out, nil
}

func (f *benchFormatter) HealthCheck(_ context.Context) error { return nil }

// silentLogger returns a logrus logger with output suppressed.
func silentLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.PanicLevel)
	return l
}

// newBenchRegistry creates a FormatterRegistry pre-populated with formatters
// covering several common languages.
func newBenchRegistry() *formatters.FormatterRegistry {
	logger := silentLogger()
	reg := formatters.NewFormatterRegistry(&formatters.RegistryConfig{
		DefaultTimeout: 5 * time.Second,
		MaxConcurrent:  8,
		EnableCaching:  false,
	}, logger)

	specs := []struct {
		name  string
		langs []string
	}{
		{"gofmt", []string{"go"}},
		{"black", []string{"python"}},
		{"prettier", []string{"javascript", "typescript", "json", "yaml", "markdown"}},
		{"rustfmt", []string{"rust"}},
		{"clang-format", []string{"c", "cpp", "java"}},
	}

	for _, s := range specs {
		f := newBenchFormatter(s.name, s.langs)
		meta := &formatters.FormatterMetadata{
			Name:      s.name,
			Type:      formatters.FormatterTypeNative,
			Version:   "1.0.0",
			Languages: s.langs,
		}
		_ = reg.Register(f, meta)
	}

	return reg
}

// newBenchExecutor creates a FormatterExecutor backed by newBenchRegistry.
func newBenchExecutor() *formatters.FormatterExecutor {
	reg := newBenchRegistry()
	logger := silentLogger()
	exec := formatters.NewFormatterExecutor(reg, &formatters.ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
		MaxRetries:     0,
		EnableCache:    false,
	}, logger)
	return exec
}

// =============================================================================
// Registry benchmarks
// =============================================================================

// BenchmarkFormatterRegistry_Get measures the cost of a name-based registry
// lookup for an already-registered formatter.
func BenchmarkFormatterRegistry_Get(b *testing.B) {
	reg := newBenchRegistry()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = reg.Get("gofmt")
	}
}

// BenchmarkFormatterRegistry_GetByLanguage measures the cost of a language-based
// registry lookup that may return multiple formatters.
func BenchmarkFormatterRegistry_GetByLanguage(b *testing.B) {
	reg := newBenchRegistry()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.GetByLanguage("javascript")
	}
}

// BenchmarkFormatterRegistry_DetectLanguage measures file-extension to language
// detection performance.
func BenchmarkFormatterRegistry_DetectLanguage(b *testing.B) {
	reg := newBenchRegistry()
	extensions := []string{
		"main.go", "app.py", "index.js", "lib.rs",
		"component.ts", "Main.java", "service.yaml",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.DetectLanguageFromPath(extensions[i%len(extensions)])
	}
}

// BenchmarkFormatterRegistry_DetectFormatter measures the combined detect +
// lookup path used by the executor when no language is specified.
func BenchmarkFormatterRegistry_DetectFormatter(b *testing.B) {
	reg := newBenchRegistry()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = reg.DetectFormatter("main.go", "package main")
	}
}

// BenchmarkFormatterRegistry_List measures the overhead of listing all registered
// formatter names.
func BenchmarkFormatterRegistry_List(b *testing.B) {
	reg := newBenchRegistry()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.List()
	}
}

// BenchmarkFormatterRegistry_Register measures the cost of registering a new
// formatter into an existing registry.
func BenchmarkFormatterRegistry_Register(b *testing.B) {
	logger := silentLogger()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reg := formatters.NewFormatterRegistry(&formatters.RegistryConfig{
			DefaultTimeout: 5 * time.Second,
		}, logger)
		f := newBenchFormatter(fmt.Sprintf("fmt-%d", i), []string{"go"})
		meta := &formatters.FormatterMetadata{
			Name:      fmt.Sprintf("fmt-%d", i),
			Type:      formatters.FormatterTypeNative,
			Version:   "1.0.0",
			Languages: []string{"go"},
		}
		b.StartTimer()
		_ = reg.Register(f, meta)
	}
}

// =============================================================================
// Executor benchmarks
// =============================================================================

// BenchmarkFormatterExecutor_Execute measures single-request execution overhead.
func BenchmarkFormatterExecutor_Execute(b *testing.B) {
	exec := newBenchExecutor()
	ctx := context.Background()
	req := &formatters.FormatRequest{
		Content:  "package main\n\nfunc main() {}\n",
		Language: "go",
		Timeout:  5 * time.Second,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = exec.Execute(ctx, req)
	}
}

// BenchmarkFormatterExecutor_Execute_Parallel measures parallel execution
// throughput.
func BenchmarkFormatterExecutor_Execute_Parallel(b *testing.B) {
	exec := newBenchExecutor()
	ctx := context.Background()
	req := &formatters.FormatRequest{
		Content:  "package main\n\nfunc main() {}\n",
		Language: "go",
		Timeout:  5 * time.Second,
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = exec.Execute(ctx, req)
		}
	})
}

// BenchmarkFormatterExecutor_ExecuteBatch_10 measures batch execution for 10
// requests submitted in a single call.
func BenchmarkFormatterExecutor_ExecuteBatch_10(b *testing.B) {
	exec := newBenchExecutor()
	ctx := context.Background()

	reqs := make([]*formatters.FormatRequest, 10)
	for i := range reqs {
		reqs[i] = &formatters.FormatRequest{
			Content:  fmt.Sprintf("package main\n// batch %d\nfunc main() {}\n", i),
			Language: "go",
			Timeout:  5 * time.Second,
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = exec.ExecuteBatch(ctx, reqs)
	}
}

// =============================================================================
// Cache benchmarks
// =============================================================================

// BenchmarkFormatterCache_Set measures the cost of inserting an entry into the
// in-process formatter cache.
func BenchmarkFormatterCache_Set(b *testing.B) {
	logger := silentLogger()
	cache := formatters.NewFormatterCache(&formatters.CacheConfig{
		TTL:         time.Minute,
		MaxSize:     100000,
		CleanupFreq: time.Minute,
	}, logger)
	defer cache.Stop()

	result := &formatters.FormatResult{
		Content:       "package main\n\nfunc main() {}\n",
		FormatterName: "gofmt",
		Success:       true,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &formatters.FormatRequest{
			Content:  fmt.Sprintf("content-%d", i),
			Language: "go",
		}
		cache.Set(req, result)
	}
}

// BenchmarkFormatterCache_Get_Hit measures the cost of a cache hit lookup.
func BenchmarkFormatterCache_Get_Hit(b *testing.B) {
	logger := silentLogger()
	cache := formatters.NewFormatterCache(&formatters.CacheConfig{
		TTL:         time.Minute,
		MaxSize:     100000,
		CleanupFreq: time.Minute,
	}, logger)
	defer cache.Stop()

	req := &formatters.FormatRequest{
		Content:  "package main\n\nfunc main() {}\n",
		Language: "go",
	}
	result := &formatters.FormatResult{
		Content:       req.Content,
		FormatterName: "gofmt",
		Success:       true,
	}
	cache.Set(req, result)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(req)
	}
}

// BenchmarkFormatterCache_Get_Miss measures the cost of a cache miss lookup.
func BenchmarkFormatterCache_Get_Miss(b *testing.B) {
	logger := silentLogger()
	cache := formatters.NewFormatterCache(&formatters.CacheConfig{
		TTL:         time.Minute,
		MaxSize:     100000,
		CleanupFreq: time.Minute,
	}, logger)
	defer cache.Stop()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &formatters.FormatRequest{
			Content:  fmt.Sprintf("miss-%d", i),
			Language: "go",
		}
		_, _ = cache.Get(req)
	}
}
