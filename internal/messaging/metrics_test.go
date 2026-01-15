package messaging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBrokerMetrics(t *testing.T) {
	m := NewBrokerMetrics()

	assert.NotNil(t, m)
	assert.False(t, m.StartTime.IsZero())
	assert.Equal(t, int64(0), m.MessagesPublished.Load())
	assert.Equal(t, int64(0), m.MessagesReceived.Load())
}

func TestBrokerMetrics_RecordConnectionAttempt(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordConnectionAttempt()

	assert.Equal(t, int64(1), m.ConnectionAttempts.Load())
}

func TestBrokerMetrics_RecordConnectionSuccess(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordConnectionSuccess()

	assert.Equal(t, int64(1), m.ConnectionSuccesses.Load())
	assert.Equal(t, int64(1), m.CurrentConnections.Load())
}

func TestBrokerMetrics_RecordConnectionFailure(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordConnectionFailure()

	assert.Equal(t, int64(1), m.ConnectionFailures.Load())
	assert.Equal(t, int64(1), m.ConnectionErrors.Load())
	assert.Equal(t, int64(1), m.TotalErrors.Load())
}

func TestBrokerMetrics_RecordDisconnection(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordConnectionSuccess()
	m.RecordDisconnection()

	assert.Equal(t, int64(0), m.CurrentConnections.Load())
}

func TestBrokerMetrics_RecordReconnectionAttempt(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordReconnectionAttempt()

	assert.Equal(t, int64(1), m.ReconnectionAttempts.Load())
	assert.False(t, m.LastReconnectTime.IsZero())
}

func TestBrokerMetrics_RecordPublish(t *testing.T) {
	m := NewBrokerMetrics()

	// Success
	m.RecordPublish(100, 10*time.Millisecond, true)
	assert.Equal(t, int64(1), m.MessagesPublished.Load())
	assert.Equal(t, int64(1), m.PublishSuccesses.Load())
	assert.Equal(t, int64(100), m.BytesPublished.Load())
	assert.False(t, m.GetLastPublishTime().IsZero())

	// Failure
	m.RecordPublish(50, 5*time.Millisecond, false)
	assert.Equal(t, int64(2), m.MessagesPublished.Load())
	assert.Equal(t, int64(1), m.PublishFailures.Load())
	assert.Equal(t, int64(1), m.PublishErrors.Load())
}

func TestBrokerMetrics_RecordBatchPublish(t *testing.T) {
	m := NewBrokerMetrics()

	// Success
	m.RecordBatchPublish(5, 500, 20*time.Millisecond, true)
	assert.Equal(t, int64(1), m.BatchesPublished.Load())
	assert.Equal(t, int64(5), m.MessagesPublished.Load())
	assert.Equal(t, int64(5), m.PublishSuccesses.Load())
	assert.Equal(t, int64(500), m.BytesPublished.Load())

	// Failure
	m.RecordBatchPublish(3, 300, 10*time.Millisecond, false)
	assert.Equal(t, int64(2), m.BatchesPublished.Load())
	assert.Equal(t, int64(8), m.MessagesPublished.Load())
	assert.Equal(t, int64(3), m.PublishFailures.Load())
}

func TestBrokerMetrics_RecordPublishConfirmation(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordPublishConfirmation()

	assert.Equal(t, int64(1), m.PublishConfirmations.Load())
}

func TestBrokerMetrics_RecordPublishTimeout(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordPublishTimeout()

	assert.Equal(t, int64(1), m.PublishTimeouts.Load())
	assert.Equal(t, int64(1), m.PublishErrors.Load())
	assert.Equal(t, int64(1), m.TotalErrors.Load())
}

func TestBrokerMetrics_RecordReceive(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordReceive(200, 15*time.Millisecond)

	assert.Equal(t, int64(1), m.MessagesReceived.Load())
	assert.Equal(t, int64(200), m.BytesReceived.Load())
	assert.False(t, m.GetLastReceiveTime().IsZero())
}

func TestBrokerMetrics_RecordProcessed(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordProcessed()

	assert.Equal(t, int64(1), m.MessagesProcessed.Load())
}

func TestBrokerMetrics_RecordFailed(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordFailed()

	assert.Equal(t, int64(1), m.MessagesFailed.Load())
	assert.Equal(t, int64(1), m.SubscribeErrors.Load())
	assert.Equal(t, int64(1), m.TotalErrors.Load())
	assert.False(t, m.GetLastErrorTime().IsZero())
}

func TestBrokerMetrics_RecordRetry(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordRetry()

	assert.Equal(t, int64(1), m.MessagesRetried.Load())
}

func TestBrokerMetrics_RecordDeadLettered(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordDeadLettered()

	assert.Equal(t, int64(1), m.MessagesDeadLettered.Load())
}

func TestBrokerMetrics_RecordSubscription(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordSubscription()

	assert.Equal(t, int64(1), m.ActiveSubscriptions.Load())
}

func TestBrokerMetrics_RecordUnsubscription(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordSubscription()
	m.RecordUnsubscription()

	assert.Equal(t, int64(0), m.ActiveSubscriptions.Load())
}

func TestBrokerMetrics_RecordQueueDeclared(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordQueueDeclared()

	assert.Equal(t, int64(1), m.QueuesDeclared.Load())
}

func TestBrokerMetrics_RecordTopicCreated(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordTopicCreated()

	assert.Equal(t, int64(1), m.TopicsCreated.Load())
}

func TestBrokerMetrics_RecordSerializationError(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordSerializationError()

	assert.Equal(t, int64(1), m.SerializationErrors.Load())
	assert.Equal(t, int64(1), m.TotalErrors.Load())
}

func TestBrokerMetrics_RecordError(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordError()

	assert.Equal(t, int64(1), m.TotalErrors.Load())
	assert.False(t, m.GetLastErrorTime().IsZero())
}

func TestBrokerMetrics_GetAveragePublishLatency(t *testing.T) {
	m := NewBrokerMetrics()

	// No data
	assert.Equal(t, time.Duration(0), m.GetAveragePublishLatency())

	// With data
	m.RecordPublish(100, 10*time.Millisecond, true)
	m.RecordPublish(100, 20*time.Millisecond, true)
	avg := m.GetAveragePublishLatency()
	assert.True(t, avg > 0)
}

func TestBrokerMetrics_GetAverageSubscribeLatency(t *testing.T) {
	m := NewBrokerMetrics()

	// No data
	assert.Equal(t, time.Duration(0), m.GetAverageSubscribeLatency())

	// With data
	m.RecordReceive(100, 10*time.Millisecond)
	m.RecordReceive(100, 20*time.Millisecond)
	avg := m.GetAverageSubscribeLatency()
	assert.True(t, avg > 0)
}

func TestBrokerMetrics_GetUptime(t *testing.T) {
	m := NewBrokerMetrics()
	time.Sleep(10 * time.Millisecond)
	uptime := m.GetUptime()

	assert.True(t, uptime >= 10*time.Millisecond)
}

func TestBrokerMetrics_GetPublishSuccessRate(t *testing.T) {
	m := NewBrokerMetrics()

	// No data
	assert.Equal(t, 1.0, m.GetPublishSuccessRate())

	// With data
	m.RecordPublish(100, 10*time.Millisecond, true)
	m.RecordPublish(100, 10*time.Millisecond, true)
	m.RecordPublish(100, 10*time.Millisecond, false)
	rate := m.GetPublishSuccessRate()
	assert.InDelta(t, 0.666, rate, 0.01)
}

func TestBrokerMetrics_GetProcessingSuccessRate(t *testing.T) {
	m := NewBrokerMetrics()

	// No data
	assert.Equal(t, 1.0, m.GetProcessingSuccessRate())

	// With data
	m.RecordReceive(100, 10*time.Millisecond)
	m.RecordReceive(100, 10*time.Millisecond)
	m.RecordProcessed()
	rate := m.GetProcessingSuccessRate()
	assert.InDelta(t, 0.5, rate, 0.01)
}

func TestBrokerMetrics_Reset(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordPublish(100, 10*time.Millisecond, true)
	m.RecordReceive(100, 10*time.Millisecond)
	m.RecordConnectionSuccess()

	m.Reset()

	assert.Equal(t, int64(0), m.MessagesPublished.Load())
	assert.Equal(t, int64(0), m.MessagesReceived.Load())
	assert.Equal(t, int64(0), m.CurrentConnections.Load())
}

func TestBrokerMetrics_Clone(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordPublish(100, 10*time.Millisecond, true)
	m.RecordReceive(100, 10*time.Millisecond)
	m.RecordConnectionSuccess()

	clone := m.Clone()

	assert.Equal(t, m.MessagesPublished.Load(), clone.MessagesPublished.Load())
	assert.Equal(t, m.MessagesReceived.Load(), clone.MessagesReceived.Load())
	assert.Equal(t, m.CurrentConnections.Load(), clone.CurrentConnections.Load())

	// Verify independence
	m.RecordPublish(100, 10*time.Millisecond, true)
	assert.NotEqual(t, m.MessagesPublished.Load(), clone.MessagesPublished.Load())
}

func TestNewMetricsSnapshot(t *testing.T) {
	m := NewBrokerMetrics()
	m.RecordPublish(100, 10*time.Millisecond, true)

	snapshot := NewMetricsSnapshot(BrokerTypeRabbitMQ, m)

	assert.Equal(t, BrokerTypeRabbitMQ, snapshot.BrokerType)
	assert.NotNil(t, snapshot.Metrics)
	assert.False(t, snapshot.CollectedAt.IsZero())
}

func TestMetricsSnapshot_WithQueueStats(t *testing.T) {
	m := NewBrokerMetrics()
	snapshot := NewMetricsSnapshot(BrokerTypeRabbitMQ, m)

	stats := []QueueStats{
		{Name: "queue1", Messages: 100},
		{Name: "queue2", Messages: 200},
	}
	snapshot.WithQueueStats(stats)

	assert.Len(t, snapshot.QueueStats, 2)
	assert.Equal(t, "queue1", snapshot.QueueStats[0].Name)
}

func TestMetricsSnapshot_WithTopicStats(t *testing.T) {
	m := NewBrokerMetrics()
	snapshot := NewMetricsSnapshot(BrokerTypeKafka, m)

	stats := []TopicMetadata{
		{Name: "topic1", Partitions: 3},
		{Name: "topic2", Partitions: 5},
	}
	snapshot.WithTopicStats(stats)

	assert.Len(t, snapshot.TopicStats, 2)
	assert.Equal(t, "topic1", snapshot.TopicStats[0].Name)
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector(100 * time.Millisecond)
	assert.NotNil(t, collector)
}

func TestMetricsCollector_Collect(t *testing.T) {
	collector := NewMetricsCollector(100 * time.Millisecond)

	// Create a mock broker with metrics
	// Since we don't have a real broker, we'll just test the collector structure
	snapshots := collector.Collect()
	assert.Empty(t, snapshots)
}
