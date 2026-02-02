# analytics

Package `analytics` provides ClickHouse-based time-series analytics for HelixAgent. It stores and queries debate metrics, provider performance statistics, and conversation metrics, enabling historical analysis and trend reporting across LLM providers.

## Architecture

The package centers on `ClickHouseAnalytics`, which wraps a `database/sql` connection to ClickHouse using the `clickhouse-go/v2` driver. It operates on two primary tables: `debate_metrics` and `conversation_metrics`.

### Key Types

- **`ClickHouseAnalytics`** -- Main client managing the ClickHouse connection and all query operations.
- **`DebateMetrics`** -- Per-round debate data: provider, model, response time, tokens, confidence, win status.
- **`ProviderStats`** -- Aggregated provider statistics with percentile latencies (P95/P99), win rate, and error rate.
- **`ConversationMetrics`** -- Per-conversation data: message count, entity count, tokens, duration, LLMs used.
- **`ClickHouseConfig`** -- Connection configuration (host, port, database, credentials, TLS).

## Public API

```go
// Construction
NewClickHouseAnalytics(config ClickHouseConfig, logger *logrus.Logger) (*ClickHouseAnalytics, error)

// Debate metrics
StoreDebateMetrics(ctx context.Context, metrics DebateMetrics) error
StoreDebateMetricsBatch(ctx context.Context, metricsList []DebateMetrics) error
GetDebateAnalytics(ctx context.Context, debateID string) (map[string]interface{}, error)

// Provider analytics
GetProviderPerformance(ctx context.Context, window time.Duration) ([]ProviderStats, error)
GetProviderTrends(ctx context.Context, provider string, interval string, periods int) ([]ProviderStats, error)
GetProviderAnalytics(ctx context.Context, provider string, window time.Duration) (*ProviderStats, error)
GetTopProviders(ctx context.Context, limit int, sortBy string) ([]ProviderStats, error)

// Conversation metrics
StoreConversationMetrics(ctx context.Context, metrics ConversationMetrics) error
GetConversationTrends(ctx context.Context, interval string, periods int) ([]map[string]interface{}, error)

// Custom queries (SELECT only)
ExecuteQuery(ctx context.Context, query string, parameters map[string]interface{}) ([]map[string]interface{}, error)

// Lifecycle
Close() error
```

## Configuration

Set via `ClickHouseConfig` or environment variables consumed by the bigdata integration layer:

| Field      | Default               | Env Var                    |
|------------|-----------------------|----------------------------|
| Host       | `localhost`           | `CLICKHOUSE_HOST`          |
| Port       | `9000`                | `CLICKHOUSE_PORT`          |
| Database   | `helixagent_analytics`| `CLICKHOUSE_DATABASE`      |
| Username   | `default`             | `CLICKHOUSE_USER`          |
| Password   | (empty)               | `CLICKHOUSE_PASSWORD`      |
| TLS        | `false`               | --                         |

## Usage

```go
client, err := analytics.NewClickHouseAnalytics(analytics.ClickHouseConfig{
    Host:     "localhost",
    Port:     9000,
    Database: "helixagent_analytics",
}, logger)
defer client.Close()

// Store debate metrics
err = client.StoreDebateMetrics(ctx, analytics.DebateMetrics{
    DebateID: "debate-123", Round: 1, Provider: "claude",
    ResponseTimeMs: 450.5, ConfidenceScore: 0.92,
})

// Query provider performance over the last 24 hours
stats, err := client.GetProviderPerformance(ctx, 24*time.Hour)
```

## Testing

```bash
go test -v -run TestClickHouse ./internal/analytics/
```

Tests require a running ClickHouse instance. Use `make test-infra-start` to provision one, or set `CLICKHOUSE_HOST` / `CLICKHOUSE_PORT` to point at an existing instance.
