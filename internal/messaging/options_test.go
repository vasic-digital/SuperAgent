package messaging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompressionType_String(t *testing.T) {
	tests := []struct {
		ct       CompressionType
		expected string
	}{
		{CompressionNone, "none"},
		{CompressionGzip, "gzip"},
		{CompressionSnappy, "snappy"},
		{CompressionLZ4, "lz4"},
		{CompressionZstd, "zstd"},
		{CompressionType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ct.String())
		})
	}
}

func TestDefaultPublishOptions(t *testing.T) {
	opts := DefaultPublishOptions()

	assert.False(t, opts.Mandatory)
	assert.False(t, opts.Immediate)
	assert.Empty(t, opts.Exchange)
	assert.Empty(t, opts.RoutingKey)
	assert.Equal(t, "application/json", opts.ContentType)
	assert.Equal(t, "utf-8", opts.ContentEncoding)
	assert.Equal(t, 30*time.Second, opts.Timeout)
	assert.True(t, opts.Confirm)
	assert.Nil(t, opts.Partition)
	assert.Nil(t, opts.Key)
	assert.Equal(t, CompressionNone, opts.Compression)
	assert.Equal(t, 100, opts.BatchSize)
	assert.Equal(t, 10, opts.LingerMs)
}

func TestApplyPublishOptions(t *testing.T) {
	partition := int32(5)
	opts := ApplyPublishOptions(
		WithMandatory(true),
		WithImmediate(true),
		WithExchange("test-exchange"),
		WithRoutingKey("test.key"),
		WithContentType("text/plain"),
		WithContentEncoding("ascii"),
		WithPublishTimeout(60*time.Second),
		WithConfirm(false),
		WithPartition(partition),
		WithMessageKey([]byte("key")),
		WithCompression(CompressionLZ4),
		WithBatchSize(200),
		WithLingerMs(20),
	)

	assert.True(t, opts.Mandatory)
	assert.True(t, opts.Immediate)
	assert.Equal(t, "test-exchange", opts.Exchange)
	assert.Equal(t, "test.key", opts.RoutingKey)
	assert.Equal(t, "text/plain", opts.ContentType)
	assert.Equal(t, "ascii", opts.ContentEncoding)
	assert.Equal(t, 60*time.Second, opts.Timeout)
	assert.False(t, opts.Confirm)
	assert.NotNil(t, opts.Partition)
	assert.Equal(t, partition, *opts.Partition)
	assert.Equal(t, []byte("key"), opts.Key)
	assert.Equal(t, CompressionLZ4, opts.Compression)
	assert.Equal(t, 200, opts.BatchSize)
	assert.Equal(t, 20, opts.LingerMs)
}

func TestDefaultSubscribeOptions(t *testing.T) {
	opts := DefaultSubscribeOptions()

	assert.Empty(t, opts.ConsumerTag)
	assert.False(t, opts.AutoAck)
	assert.False(t, opts.Exclusive)
	assert.False(t, opts.NoLocal)
	assert.False(t, opts.NoWait)
	assert.Nil(t, opts.QueueArgs)
	assert.Equal(t, 10, opts.Prefetch)
	assert.Equal(t, 0, opts.PrefetchSize)
	assert.Empty(t, opts.GroupID)
	assert.Equal(t, 10*time.Second, opts.SessionTimeout)
	assert.Equal(t, 3*time.Second, opts.HeartbeatInterval)
	assert.Equal(t, 500, opts.MaxPollRecords)
	assert.Equal(t, OffsetResetLatest, opts.OffsetReset)
	assert.Equal(t, 5*time.Second, opts.CommitInterval)
	assert.Nil(t, opts.Filter)
	assert.Equal(t, 100, opts.BufferSize)
	assert.True(t, opts.RetryOnError)
	assert.Equal(t, 3, opts.MaxRetries)
	assert.Equal(t, 1*time.Second, opts.RetryDelay)
}

func TestApplySubscribeOptions(t *testing.T) {
	filter := func(msg *Message) bool { return true }
	args := map[string]interface{}{"key": "value"}

	opts := ApplySubscribeOptions(
		WithConsumerTag("consumer-1"),
		WithAutoAck(true),
		WithExclusive(true),
		WithNoLocal(true),
		WithNoWait(true),
		WithQueueArgs(args),
		WithPrefetch(20),
		WithPrefetchSize(1024),
		WithGroupID("group-1"),
		WithSessionTimeout(20*time.Second),
		WithHeartbeatInterval(5*time.Second),
		WithMaxPollRecords(1000),
		WithOffsetReset(OffsetResetEarliest),
		WithCommitInterval(10*time.Second),
		WithMessageFilter(filter),
		WithBufferSize(500),
		WithRetryOnError(false),
		WithMaxSubscribeRetries(5),
		WithRetryDelay(2*time.Second),
	)

	assert.Equal(t, "consumer-1", opts.ConsumerTag)
	assert.True(t, opts.AutoAck)
	assert.True(t, opts.Exclusive)
	assert.True(t, opts.NoLocal)
	assert.True(t, opts.NoWait)
	assert.Equal(t, args, opts.QueueArgs)
	assert.Equal(t, 20, opts.Prefetch)
	assert.Equal(t, 1024, opts.PrefetchSize)
	assert.Equal(t, "group-1", opts.GroupID)
	assert.Equal(t, 20*time.Second, opts.SessionTimeout)
	assert.Equal(t, 5*time.Second, opts.HeartbeatInterval)
	assert.Equal(t, 1000, opts.MaxPollRecords)
	assert.Equal(t, OffsetResetEarliest, opts.OffsetReset)
	assert.Equal(t, 10*time.Second, opts.CommitInterval)
	assert.NotNil(t, opts.Filter)
	assert.Equal(t, 500, opts.BufferSize)
	assert.False(t, opts.RetryOnError)
	assert.Equal(t, 5, opts.MaxRetries)
	assert.Equal(t, 2*time.Second, opts.RetryDelay)
}

func TestDefaultQueueOptions(t *testing.T) {
	opts := DefaultQueueOptions()

	assert.True(t, opts.Durable)
	assert.False(t, opts.AutoDelete)
	assert.False(t, opts.Exclusive)
	assert.False(t, opts.NoWait)
	assert.Nil(t, opts.Args)
	assert.Empty(t, opts.DeadLetterExchange)
	assert.Empty(t, opts.DeadLetterRoutingKey)
	assert.Equal(t, int64(0), opts.MessageTTL)
	assert.Equal(t, int64(0), opts.MaxLength)
	assert.Equal(t, int64(0), opts.MaxLengthBytes)
	assert.Nil(t, opts.MaxPriority)
}

func TestApplyQueueOptions(t *testing.T) {
	maxPriority := 10
	opts := ApplyQueueOptions(
		WithDurable(false),
		WithAutoDelete(true),
		WithQueueExclusive(true),
		WithQueueNoWait(true),
		WithDeadLetterExchange("dlx"),
		WithDeadLetterRoutingKey("dlx.key"),
		WithMessageTTL(1*time.Hour),
		WithMaxLength(1000),
		WithMaxLengthBytes(1024*1024),
		WithMaxPriority(maxPriority),
	)

	assert.False(t, opts.Durable)
	assert.True(t, opts.AutoDelete)
	assert.True(t, opts.Exclusive)
	assert.True(t, opts.NoWait)
	assert.Equal(t, "dlx", opts.DeadLetterExchange)
	assert.Equal(t, "dlx.key", opts.DeadLetterRoutingKey)
	assert.Equal(t, int64(3600000), opts.MessageTTL)
	assert.Equal(t, int64(1000), opts.MaxLength)
	assert.Equal(t, int64(1024*1024), opts.MaxLengthBytes)
	assert.NotNil(t, opts.MaxPriority)
	assert.Equal(t, maxPriority, *opts.MaxPriority)
}

func TestDefaultStreamOptions(t *testing.T) {
	opts := DefaultStreamOptions()

	assert.Equal(t, int64(-1), opts.StartOffset)
	assert.Nil(t, opts.StartTime)
	assert.Equal(t, 1000, opts.BufferSize)
	assert.Equal(t, 1, opts.MinBytes)
	assert.Equal(t, 10*1024*1024, opts.MaxBytes)
	assert.Equal(t, 500*time.Millisecond, opts.MaxWait)
	assert.Nil(t, opts.Partition)
}

func TestApplyStreamOptions(t *testing.T) {
	startTime := time.Now()
	partition := int32(3)
	opts := ApplyStreamOptions(
		WithStartOffset(100),
		WithStartTime(startTime),
		WithStreamBufferSize(500),
		WithMinBytes(10),
		WithMaxBytes(5*1024*1024),
		WithMaxWait(1*time.Second),
		WithStreamPartition(partition),
	)

	assert.Equal(t, int64(100), opts.StartOffset)
	assert.NotNil(t, opts.StartTime)
	assert.Equal(t, startTime, *opts.StartTime)
	assert.Equal(t, 500, opts.BufferSize)
	assert.Equal(t, 10, opts.MinBytes)
	assert.Equal(t, 5*1024*1024, opts.MaxBytes)
	assert.Equal(t, 1*time.Second, opts.MaxWait)
	assert.NotNil(t, opts.Partition)
	assert.Equal(t, partition, *opts.Partition)
}

func TestOffsetReset_Values(t *testing.T) {
	assert.Equal(t, OffsetReset("earliest"), OffsetResetEarliest)
	assert.Equal(t, OffsetReset("latest"), OffsetResetLatest)
	assert.Equal(t, OffsetReset("none"), OffsetResetNone)
}
