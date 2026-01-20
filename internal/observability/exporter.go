package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// ExporterConfig contains configuration for trace exporters
type ExporterConfig struct {
	Type        ExporterType
	Endpoint    string
	Headers     map[string]string
	Insecure    bool
	ServiceName string
	Environment string
	Version     string
}

// SetupTraceExporter initializes the trace exporter based on configuration
func SetupTraceExporter(ctx context.Context, config *ExporterConfig) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	switch config.Type {
	case ExporterOTLP:
		exporter, err = setupOTLPExporter(ctx, config)
	case ExporterConsole:
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	case ExporterNone:
		// No-op exporter, just set up the provider without exporting
		return setupNoOpProvider(config)
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.Version),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set as global provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

func setupOTLPExporter(ctx context.Context, config *ExporterConfig) (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Endpoint),
	}

	if config.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if len(config.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(config.Headers))
	}

	return otlptracehttp.New(ctx, opts...)
}

func setupNoOpProvider(config *ExporterConfig) (*sdktrace.TracerProvider, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.Version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.NeverSample()),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

// ShutdownTraceExporter gracefully shuts down the trace provider
func ShutdownTraceExporter(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

// LangfuseConfig contains configuration for Langfuse integration
type LangfuseConfig struct {
	PublicKey  string
	SecretKey  string
	BaseURL    string
	FlushAt    int
	FlushAfter int // seconds
}

// SetupLangfuseExporter sets up Langfuse as the observability backend
// Langfuse supports OTLP protocol, so we can use the OTLP exporter
func SetupLangfuseExporter(ctx context.Context, config *LangfuseConfig) (*sdktrace.TracerProvider, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://cloud.langfuse.com"
	}

	return SetupTraceExporter(ctx, &ExporterConfig{
		Type:     ExporterOTLP,
		Endpoint: config.BaseURL + "/api/public/otel/v1/traces",
		Headers: map[string]string{
			"Authorization": "Basic " + config.PublicKey + ":" + config.SecretKey,
		},
		ServiceName: "helixagent",
		Environment: "production",
	})
}
