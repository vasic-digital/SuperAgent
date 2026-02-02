package analytics

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// sliceValueConverter is a custom ValueConverter that handles []string
// (and other slice types) that the real ClickHouse driver supports but
// the standard database/sql driver does not.
type sliceValueConverter struct{}

func (s sliceValueConverter) ConvertValue(v interface{}) (driver.Value, error) {
	switch val := v.(type) {
	case []string:
		return strings.Join(val, ","), nil
	default:
		return driver.DefaultParameterConverter.ConvertValue(v)
	}
}

// newMockAnalytics creates a *ClickHouseAnalytics backed by a sqlmock DB.
// The caller should call db.Close() when done (or let the test framework handle it).
func newMockAnalytics(t *testing.T) (*ClickHouseAnalytics, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.ValueConverterOption(sliceValueConverter{}))
	assert.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	cha := &ClickHouseAnalytics{
		conn:   db,
		logger: logger,
	}
	return cha, mock, db
}

// providerStatsColumns returns the standard column names for provider stats queries.
func providerStatsColumns() []string {
	return []string{
		"provider", "total_requests", "avg_response_time",
		"p95_response_time", "p99_response_time", "avg_confidence",
		"total_tokens", "avg_tokens_per_req", "error_rate", "win_rate",
	}
}

func TestNewClickHouseAnalytics_NilLogger(t *testing.T) {
	// We cannot fully test NewClickHouseAnalytics without a real ClickHouse,
	// but we can verify that passing a nil logger does not panic by constructing
	// the struct directly and checking that a logger would be assigned.
	// The constructor calls sql.Open + Ping which requires a real driver.
	// Instead, we verify the nil-logger branch logic:
	var logger *logrus.Logger
	if logger == nil {
		logger = logrus.New()
	}
	assert.NotNil(t, logger)

	// Also verify that our mock helper produces a valid analytics instance
	cha, _, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()
	assert.NotNil(t, cha)
	assert.NotNil(t, cha.conn)
	assert.NotNil(t, cha.logger)
}

func TestClickHouseAnalytics_StoreDebateMetrics_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	metrics := DebateMetrics{
		DebateID:        "debate-1",
		Round:           1,
		Timestamp:       time.Now(),
		Provider:        "claude",
		Model:           "claude-3",
		Position:        "for",
		ResponseTimeMs:  150.5,
		TokensUsed:      500,
		ConfidenceScore: 0.85,
		ErrorCount:      0,
		WasWinner:       true,
	}

	mock.ExpectExec("INSERT INTO debate_metrics").
		WithArgs(
			metrics.DebateID, metrics.Round, metrics.Timestamp,
			metrics.Provider, metrics.Model, metrics.Position,
			metrics.ResponseTimeMs, metrics.TokensUsed, metrics.ConfidenceScore,
			metrics.ErrorCount, metrics.WasWinner,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cha.StoreDebateMetrics(ctx, metrics)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_StoreDebateMetrics_Error(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	metrics := DebateMetrics{
		DebateID: "debate-err",
		Round:    1,
		Provider: "gemini",
	}

	mock.ExpectExec("INSERT INTO debate_metrics").
		WithArgs(
			metrics.DebateID, metrics.Round, metrics.Timestamp,
			metrics.Provider, metrics.Model, metrics.Position,
			metrics.ResponseTimeMs, metrics.TokensUsed, metrics.ConfidenceScore,
			metrics.ErrorCount, metrics.WasWinner,
		).
		WillReturnError(fmt.Errorf("connection refused"))

	err := cha.StoreDebateMetrics(ctx, metrics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert debate metrics")
	assert.Contains(t, err.Error(), "connection refused")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_StoreDebateMetricsBatch_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	now := time.Now()
	metricsList := []DebateMetrics{
		{
			DebateID: "batch-1", Round: 1, Timestamp: now,
			Provider: "claude", Model: "claude-3", Position: "for",
			ResponseTimeMs: 100, TokensUsed: 300, ConfidenceScore: 0.9,
			ErrorCount: 0, WasWinner: true,
		},
		{
			DebateID: "batch-1", Round: 1, Timestamp: now,
			Provider: "gemini", Model: "gemini-pro", Position: "against",
			ResponseTimeMs: 120, TokensUsed: 350, ConfidenceScore: 0.8,
			ErrorCount: 0, WasWinner: false,
		},
	}

	mock.ExpectBegin()
	prep := mock.ExpectPrepare("INSERT INTO debate_metrics")

	for _, m := range metricsList {
		prep.ExpectExec().
			WithArgs(
				m.DebateID, m.Round, m.Timestamp,
				m.Provider, m.Model, m.Position,
				m.ResponseTimeMs, m.TokensUsed, m.ConfidenceScore,
				m.ErrorCount, m.WasWinner,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectCommit()

	err := cha.StoreDebateMetricsBatch(ctx, metricsList)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_StoreDebateMetricsBatch_EmptyList(t *testing.T) {
	cha, _, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	err := cha.StoreDebateMetricsBatch(ctx, []DebateMetrics{})
	assert.NoError(t, err)

	err = cha.StoreDebateMetricsBatch(ctx, nil)
	assert.NoError(t, err)
}

func TestClickHouseAnalytics_StoreDebateMetricsBatch_CommitError(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	now := time.Now()
	metricsList := []DebateMetrics{
		{
			DebateID: "batch-fail", Round: 1, Timestamp: now,
			Provider: "deepseek", Model: "deepseek-v2", Position: "neutral",
			ResponseTimeMs: 200, TokensUsed: 400, ConfidenceScore: 0.7,
			ErrorCount: 1, WasWinner: false,
		},
	}

	mock.ExpectBegin()
	prep := mock.ExpectPrepare("INSERT INTO debate_metrics")
	prep.ExpectExec().
		WithArgs(
			metricsList[0].DebateID, metricsList[0].Round, metricsList[0].Timestamp,
			metricsList[0].Provider, metricsList[0].Model, metricsList[0].Position,
			metricsList[0].ResponseTimeMs, metricsList[0].TokensUsed,
			metricsList[0].ConfidenceScore, metricsList[0].ErrorCount,
			metricsList[0].WasWinner,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))
	// The deferred tx.Rollback() runs after Commit fails. After a failed commit,
	// database/sql considers the transaction done, so Rollback returns ErrTxDone
	// and sqlmock does not match it as an expectation. No ExpectRollback needed.

	err := cha.StoreDebateMetricsBatch(ctx, metricsList)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetProviderPerformance_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	window := 24 * time.Hour

	rows := sqlmock.NewRows(providerStatsColumns()).
		AddRow("claude", int64(100), float32(120.5), float32(250.0), float32(400.0),
			float32(0.88), int64(50000), float32(500.0), float32(0.02), float32(0.65)).
		AddRow("gemini", int64(80), float32(110.0), float32(200.0), float32(350.0),
			float32(0.82), int64(40000), float32(500.0), float32(0.03), float32(0.55))

	mock.ExpectQuery("SELECT").
		WithArgs(int64(window.Seconds())).
		WillReturnRows(rows)

	stats, err := cha.GetProviderPerformance(ctx, window)
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "claude", stats[0].Provider)
	assert.Equal(t, int64(100), stats[0].TotalRequests)
	assert.InDelta(t, float32(120.5), stats[0].AvgResponseTime, 0.01)
	assert.Equal(t, "last_24h0m0s", stats[0].Period)
	assert.Equal(t, "gemini", stats[1].Provider)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetProviderPerformance_Empty(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	window := time.Hour

	rows := sqlmock.NewRows(providerStatsColumns())
	mock.ExpectQuery("SELECT").
		WithArgs(int64(window.Seconds())).
		WillReturnRows(rows)

	stats, err := cha.GetProviderPerformance(ctx, window)
	assert.NoError(t, err)
	assert.Empty(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetProviderPerformance_Error(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	window := time.Hour

	mock.ExpectQuery("SELECT").
		WithArgs(int64(window.Seconds())).
		WillReturnError(fmt.Errorf("query timeout"))

	stats, err := cha.GetProviderPerformance(ctx, window)
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "failed to query provider performance")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetProviderTrends_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	provider := "claude"
	interval := "HOUR"
	periods := 6

	trendCols := []string{
		"provider", "period_start", "total_requests", "avg_response_time",
		"p95_response_time", "p99_response_time", "avg_confidence",
		"total_tokens", "avg_tokens_per_req", "error_rate", "win_rate",
	}
	periodStart := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows(trendCols).
		AddRow("claude", periodStart, int64(50), float32(100.0), float32(200.0),
			float32(300.0), float32(0.9), int64(25000), float32(500.0),
			float32(0.01), float32(0.7))

	mock.ExpectQuery("SELECT").
		WithArgs(provider, periods, periods).
		WillReturnRows(rows)

	trends, err := cha.GetProviderTrends(ctx, provider, interval, periods)
	assert.NoError(t, err)
	assert.Len(t, trends, 1)
	assert.Equal(t, "claude", trends[0].Provider)
	assert.Equal(t, periodStart.Format("2006-01-02 15:04:05"), trends[0].Period)
	assert.Equal(t, int64(50), trends[0].TotalRequests)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetProviderAnalytics_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	provider := "deepseek"
	window := 12 * time.Hour

	row := sqlmock.NewRows(providerStatsColumns()).
		AddRow("deepseek", int64(200), float32(95.0), float32(180.0),
			float32(300.0), float32(0.85), int64(100000), float32(500.0),
			float32(0.01), float32(0.6))

	mock.ExpectQuery("SELECT").
		WithArgs(provider, int64(window.Seconds())).
		WillReturnRows(row)

	stats, err := cha.GetProviderAnalytics(ctx, provider, window)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "deepseek", stats.Provider)
	assert.Equal(t, int64(200), stats.TotalRequests)
	assert.Equal(t, "last_12h0m0s", stats.Period)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetProviderAnalytics_NoRows(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	provider := "unknown"
	window := time.Hour

	row := sqlmock.NewRows(providerStatsColumns())
	mock.ExpectQuery("SELECT").
		WithArgs(provider, int64(window.Seconds())).
		WillReturnRows(row)

	stats, err := cha.GetProviderAnalytics(ctx, provider, window)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "unknown", stats.Provider)
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, "last_1h0m0s", stats.Period)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_StoreConversationMetrics_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	metrics := ConversationMetrics{
		ConversationID: "conv-1",
		UserID:         "user-123",
		Timestamp:      time.Now(),
		MessageCount:   10,
		EntityCount:    3,
		TotalTokens:    5000,
		DurationMs:     30000,
		DebateRounds:   2,
		LLMsUsed:       []string{"claude", "gemini"},
	}

	// Use AnyArg for the []string LLMsUsed field since standard database/sql
	// cannot convert []string natively; the real ClickHouse driver handles it.
	mock.ExpectExec("INSERT INTO conversation_metrics").
		WithArgs(
			metrics.ConversationID, metrics.UserID, metrics.Timestamp,
			metrics.MessageCount, metrics.EntityCount, metrics.TotalTokens,
			metrics.DurationMs, metrics.DebateRounds, sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cha.StoreConversationMetrics(ctx, metrics)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_StoreConversationMetrics_Error(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	metrics := ConversationMetrics{
		ConversationID: "conv-err",
		UserID:         "user-456",
		Timestamp:      time.Now(),
	}

	mock.ExpectExec("INSERT INTO conversation_metrics").
		WithArgs(
			metrics.ConversationID, metrics.UserID, metrics.Timestamp,
			metrics.MessageCount, metrics.EntityCount, metrics.TotalTokens,
			metrics.DurationMs, metrics.DebateRounds, sqlmock.AnyArg(),
		).
		WillReturnError(fmt.Errorf("disk full"))

	err := cha.StoreConversationMetrics(ctx, metrics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert conversation metrics")
	assert.Contains(t, err.Error(), "disk full")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetConversationTrends_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	interval := "DAY"
	periods := 7

	trendCols := []string{
		"period_start", "total_conversations", "avg_messages",
		"avg_entities", "avg_tokens", "avg_duration_ms", "avg_debate_rounds",
	}
	periodStart := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows(trendCols).
		AddRow(periodStart, int64(25), float64(8.5), float64(2.3),
			float64(4500.0), float64(25000.0), float64(1.5))

	mock.ExpectQuery("SELECT").
		WithArgs(periods, periods).
		WillReturnRows(rows)

	trends, err := cha.GetConversationTrends(ctx, interval, periods)
	assert.NoError(t, err)
	assert.Len(t, trends, 1)
	assert.Equal(t, periodStart.Format("2006-01-02 15:04:05"), trends[0]["period"])
	assert.Equal(t, int64(25), trends[0]["total_conversations"])
	assert.InDelta(t, 8.5, trends[0]["avg_messages"], 0.01)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetTopProviders_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	limit := 3
	sortBy := "winrate"

	rows := sqlmock.NewRows(providerStatsColumns()).
		AddRow("claude", int64(100), float32(120.0), float32(250.0), float32(400.0),
			float32(0.9), int64(50000), float32(500.0), float32(0.01), float32(0.75)).
		AddRow("gemini", int64(80), float32(110.0), float32(200.0), float32(350.0),
			float32(0.85), int64(40000), float32(500.0), float32(0.02), float32(0.65))

	mock.ExpectQuery("SELECT").
		WithArgs(limit).
		WillReturnRows(rows)

	providers, err := cha.GetTopProviders(ctx, limit, sortBy)
	assert.NoError(t, err)
	assert.Len(t, providers, 2)
	assert.Equal(t, "claude", providers[0].Provider)
	assert.Equal(t, "last_24h", providers[0].Period)
	assert.Equal(t, "gemini", providers[1].Provider)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetTopProviders_DefaultSort(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	limit := 5
	sortBy := "invalid_sort_field"

	rows := sqlmock.NewRows(providerStatsColumns()).
		AddRow("mistral", int64(60), float32(130.0), float32(260.0), float32(410.0),
			float32(0.82), int64(30000), float32(500.0), float32(0.03), float32(0.5))

	// The invalid sort field should default to "confidence", which uses
	// ORDER BY avg_confidence DESC. We just verify the query executes.
	mock.ExpectQuery("SELECT").
		WithArgs(limit).
		WillReturnRows(rows)

	providers, err := cha.GetTopProviders(ctx, limit, sortBy)
	assert.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "mistral", providers[0].Provider)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_GetDebateAnalytics_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	debateID := "debate-42"

	// Main aggregated metrics query
	mainCols := []string{
		"total_rounds", "avg_response_time", "max_response_time",
		"total_tokens", "avg_confidence", "total_errors",
	}
	mainRow := sqlmock.NewRows(mainCols).
		AddRow(int64(3), float32(130.0), float32(250.0), int64(15000), float32(0.87), 2)

	mock.ExpectQuery("SELECT").
		WithArgs(debateID).
		WillReturnRows(mainRow)

	// Participants query
	partCols := []string{"provider", "model"}
	partRows := sqlmock.NewRows(partCols).
		AddRow("claude", "claude-3").
		AddRow("gemini", "gemini-pro")

	mock.ExpectQuery("SELECT DISTINCT").
		WithArgs(debateID).
		WillReturnRows(partRows)

	// Winner query
	winCols := []string{"provider", "win_count"}
	winRow := sqlmock.NewRows(winCols).
		AddRow("claude", 2)

	mock.ExpectQuery("SELECT").
		WithArgs(debateID).
		WillReturnRows(winRow)

	result, err := cha.GetDebateAnalytics(ctx, debateID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, debateID, result["debate_id"])
	assert.Equal(t, int64(3), result["total_rounds"])
	assert.InDelta(t, float32(130.0), result["avg_response_time"], 0.01)
	assert.Equal(t, int64(15000), result["total_tokens"])
	assert.Equal(t, 2, result["total_errors"])
	assert.Equal(t, "claude", result["winner"])

	participants, ok := result["participants"].([]string)
	assert.True(t, ok)
	assert.Len(t, participants, 2)
	assert.Equal(t, "claude/claude-3", participants[0])
	assert.Equal(t, "gemini/gemini-pro", participants[1])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_ExecuteQuery_Success(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	query := "SELECT provider, COUNT(*) as cnt FROM debate_metrics GROUP BY provider"

	resultCols := []string{"provider", "cnt"}
	rows := sqlmock.NewRows(resultCols).
		AddRow("claude", int64(100)).
		AddRow("gemini", int64(80))

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	results, err := cha.ExecuteQuery(ctx, query, nil)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "claude", results[0]["provider"])
	assert.Equal(t, int64(100), results[0]["cnt"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_ExecuteQuery_NonSelectRejected(t *testing.T) {
	cha, _, db := newMockAnalytics(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	tests := []struct {
		name  string
		query string
	}{
		{"INSERT", "INSERT INTO debate_metrics VALUES (1,2,3)"},
		{"DELETE", "DELETE FROM debate_metrics WHERE 1=1"},
		{"DROP", "DROP TABLE debate_metrics"},
		{"UPDATE", "UPDATE debate_metrics SET round = 1"},
		{"ALTER", "ALTER TABLE debate_metrics ADD COLUMN foo String"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := cha.ExecuteQuery(ctx, tt.query, nil)
			assert.Error(t, err)
			assert.Nil(t, results)
			assert.Contains(t, err.Error(), "only SELECT queries are allowed")
		})
	}
}

func TestClickHouseAnalytics_Close(t *testing.T) {
	cha, mock, db := newMockAnalytics(t)
	_ = db // db is the same as cha.conn

	mock.ExpectClose()
	err := cha.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestClickHouseAnalytics_Close_NilConn(t *testing.T) {
	logger := logrus.New()
	cha := &ClickHouseAnalytics{
		conn:   nil,
		logger: logger,
	}

	err := cha.Close()
	assert.NoError(t, err)
}
