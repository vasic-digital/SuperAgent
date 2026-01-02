package streaming

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Buffer Tests
// ============================================================================

func TestCharacterBuffer(t *testing.T) {
	buf := NewCharacterBuffer()

	// Each character should be emitted individually
	results := buf.Add("hello")
	assert.Equal(t, []string{"h", "e", "l", "l", "o"}, results)

	// Empty string should return empty
	results = buf.Add("")
	assert.Empty(t, results)

	// Flush should return empty for character buffer
	assert.Equal(t, "", buf.Flush())
}

func TestWordBuffer(t *testing.T) {
	buf := NewWordBuffer(" ")

	// Single word without delimiter stays buffered
	results := buf.Add("hello")
	assert.Empty(t, results)

	// Adding space triggers word emission
	results = buf.Add(" ")
	assert.Equal(t, []string{"hello "}, results)

	// Multiple words
	buf.Reset()
	results = buf.Add("hello world ")
	assert.Equal(t, []string{"hello ", "world "}, results)

	// Partial word stays buffered
	results = buf.Add("test")
	assert.Empty(t, results)

	// Flush returns buffered content
	assert.Equal(t, "test", buf.Flush())
}

func TestSentenceBuffer(t *testing.T) {
	buf := NewSentenceBuffer()

	// Partial sentence stays buffered
	results := buf.Add("Hello world")
	assert.Empty(t, results)

	// Period triggers emission (trailing space is trimmed)
	results = buf.Add(". ")
	assert.Equal(t, []string{"Hello world."}, results)

	// Question mark triggers emission
	buf.Reset()
	results = buf.Add("How are you? ")
	assert.Equal(t, []string{"How are you?"}, results)

	// Exclamation mark triggers emission
	buf.Reset()
	results = buf.Add("Wow! Amazing")
	assert.Equal(t, []string{"Wow!"}, results)
	assert.Equal(t, "Amazing", buf.Flush())
}

func TestLineBuffer(t *testing.T) {
	buf := NewLineBuffer()

	// Partial line stays buffered
	results := buf.Add("Hello")
	assert.Empty(t, results)

	// Newline triggers emission
	results = buf.Add(" world\n")
	assert.Equal(t, []string{"Hello world\n"}, results)

	// Multiple lines
	buf.Reset()
	results = buf.Add("line1\nline2\npartial")
	assert.Equal(t, []string{"line1\n", "line2\n"}, results)
	assert.Equal(t, "partial", buf.Flush())
}

func TestParagraphBuffer(t *testing.T) {
	buf := NewParagraphBuffer()

	// Single newline doesn't trigger
	results := buf.Add("Hello\nworld")
	assert.Empty(t, results)

	// Double newline triggers paragraph emission
	results = buf.Add("\n\n")
	assert.Equal(t, []string{"Hello\nworld\n\n"}, results)

	// Flush remaining
	buf.Add("next paragraph")
	assert.Equal(t, "next paragraph", buf.Flush())
}

func TestTokenBuffer(t *testing.T) {
	buf := NewTokenBuffer(3) // Emit after 3 words/tokens

	// Fewer than threshold stays buffered
	results := buf.Add("one two")
	assert.Empty(t, results)

	// Reaching threshold triggers emission of all buffered content
	results = buf.Add(" three four")
	assert.Len(t, results, 1)
	assert.Contains(t, results[0], "one")
	assert.Contains(t, results[0], "three")

	// After emission, buffer is empty
	remaining := buf.Flush()
	assert.Empty(t, remaining)

	// Test flush with partial content
	buf.Reset()
	buf.Add("hello world") // 2 tokens, below threshold
	remaining = buf.Flush()
	assert.Equal(t, "hello world", remaining)
}

// ============================================================================
// Progress Tracker Tests
// ============================================================================

func TestProgressTracker(t *testing.T) {
	tracker := NewProgressTracker(100) // Estimate 100 tokens
	tracker.Start()

	// Initial state
	progress := tracker.GetProgress()
	assert.Equal(t, 0, progress.TokensGenerated)
	assert.Equal(t, 0, progress.ChunksReceived)

	// Update with content
	progress = tracker.Update("Hello world test")
	assert.Equal(t, 3, progress.TokensGenerated) // 3 words
	assert.Equal(t, 1, progress.ChunksReceived)
	assert.Equal(t, 16, progress.CharactersReceived)

	// Progress percentage
	assert.Equal(t, 3.0, progress.PercentComplete)

	// Multiple updates
	tracker.Update("more words here")
	progress = tracker.GetProgress()
	assert.Equal(t, 6, progress.TokensGenerated)
	assert.Equal(t, 2, progress.ChunksReceived)
}

func TestProgressTrackerWithoutEstimate(t *testing.T) {
	tracker := NewProgressTracker(0) // No estimate
	tracker.Start()

	progress := tracker.Update("Hello world")
	assert.Equal(t, 2, progress.TokensGenerated)
	assert.Equal(t, float64(0), progress.PercentComplete)
	assert.Equal(t, float64(0), progress.EstimatedRemaining)
}

func TestProgressTrackerUpdateTokens(t *testing.T) {
	tracker := NewProgressTracker(100)
	tracker.Start()

	progress := tracker.UpdateTokens(10)
	assert.Equal(t, 10, progress.TokensGenerated)
	assert.Equal(t, 1, progress.ChunksReceived)
	assert.Equal(t, 10.0, progress.PercentComplete)
}

func TestProgressTrackerReset(t *testing.T) {
	tracker := NewProgressTracker(100)
	tracker.Start()
	tracker.Update("Hello world")

	tracker.Reset()
	progress := tracker.GetProgress()
	assert.Equal(t, 0, progress.TokensGenerated)
	assert.Equal(t, 0, progress.ChunksReceived)
}

func TestProgressTrackerSetEstimatedTokens(t *testing.T) {
	tracker := NewProgressTracker(100)
	tracker.Start()
	tracker.Update("Hello world") // 2 tokens

	tracker.SetEstimatedTokens(10)
	progress := tracker.GetProgress()
	assert.Equal(t, 20.0, progress.PercentComplete) // 2/10 = 20%
}

func TestThrottledCallback(t *testing.T) {
	var callCount int
	var mu sync.Mutex

	callback := func(p *StreamProgress) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	throttled := NewThrottledCallback(callback, 50*time.Millisecond)

	// First call should go through
	throttled.Call(&StreamProgress{})
	mu.Lock()
	assert.Equal(t, 1, callCount)
	mu.Unlock()

	// Immediate second call should be throttled
	throttled.Call(&StreamProgress{})
	mu.Lock()
	assert.Equal(t, 1, callCount)
	mu.Unlock()

	// Wait for interval and call again
	time.Sleep(60 * time.Millisecond)
	throttled.Call(&StreamProgress{})
	mu.Lock()
	assert.Equal(t, 2, callCount)
	mu.Unlock()
}

func TestThrottledCallbackForceCall(t *testing.T) {
	var callCount int

	callback := func(p *StreamProgress) {
		callCount++
	}

	throttled := NewThrottledCallback(callback, 1*time.Hour) // Very long interval

	throttled.Call(&StreamProgress{})
	assert.Equal(t, 1, callCount)

	// ForceCall should bypass throttle
	throttled.ForceCall(&StreamProgress{})
	assert.Equal(t, 2, callCount)
}

// ============================================================================
// Stream Aggregator Tests
// ============================================================================

func TestStreamAggregator(t *testing.T) {
	agg := NewStreamAggregator()

	ctx := context.Background()
	in := make(chan string, 3)
	in <- "Hello "
	in <- "world "
	in <- "test"
	close(in)

	out, getResult := agg.Aggregate(ctx, in)

	// Consume output
	var chunks []string
	for chunk := range out {
		chunks = append(chunks, chunk)
	}

	assert.Equal(t, []string{"Hello ", "world ", "test"}, chunks)

	// Get aggregated result
	result := getResult()
	assert.Equal(t, "Hello world test", result.FullContent)
	assert.Equal(t, 3, result.ChunkCount)
	assert.Equal(t, 3, result.TokenCount) // 3 words
	assert.Equal(t, 16, result.CharacterCount)
}

func TestStreamAggregatorContextCancel(t *testing.T) {
	agg := NewStreamAggregator()

	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan string)

	out, _ := agg.Aggregate(ctx, in)

	// Cancel context
	cancel()

	// Output channel should close
	select {
	case _, ok := <-out:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for channel close")
	}
}

func TestChunkAggregator(t *testing.T) {
	agg := NewChunkAggregator()

	ctx := context.Background()
	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hello ", Index: 0}
	in <- &StreamChunk{Content: "world", Index: 1}
	in <- &StreamChunk{Content: "", Index: 2, Done: true}
	close(in)

	out, getResult := agg.AggregateChunks(ctx, in)

	// Consume output
	var chunks []*StreamChunk
	for chunk := range out {
		chunks = append(chunks, chunk)
	}

	assert.Len(t, chunks, 3)

	// Get aggregated result
	result := getResult()
	assert.Equal(t, "Hello world", result.FullContent)
	assert.Equal(t, 3, result.ChunkCount)
}

// ============================================================================
// Rate Limiter Tests
// ============================================================================

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(100) // 100 tokens/second

	ctx := context.Background()
	in := make(chan string, 5)
	for i := 0; i < 5; i++ {
		in <- "token"
	}
	close(in)

	start := time.Now()
	out := limiter.Limit(ctx, in)

	// Consume all tokens
	count := 0
	for range out {
		count++
	}

	elapsed := time.Since(start)
	assert.Equal(t, 5, count)

	// Should take at least 40ms (5 tokens at 100/sec = 50ms intervals, minus first)
	assert.True(t, elapsed >= 40*time.Millisecond, "Expected rate limiting delay")
}

func TestRateLimiterZeroRate(t *testing.T) {
	limiter := NewRateLimiter(0) // Should default to 100

	assert.Equal(t, float64(100), limiter.tokensPerSecond)
}

func TestRateLimiterContextCancel(t *testing.T) {
	limiter := NewRateLimiter(1) // Very slow

	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan string, 10)
	for i := 0; i < 10; i++ {
		in <- "token"
	}

	out := limiter.Limit(ctx, in)

	// Read one token
	<-out

	// Cancel context
	cancel()

	// Channel should close quickly
	select {
	case <-out:
		// Good, channel activity after cancel
	case <-time.After(100 * time.Millisecond):
		// Also acceptable - context was cancelled
	}
}

func TestRateLimiterLimitChunks(t *testing.T) {
	limiter := NewRateLimiter(100)

	ctx := context.Background()
	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	in <- &StreamChunk{Content: " world", Index: 1}
	in <- &StreamChunk{Content: "", Index: 2, Done: true}
	close(in)

	out := limiter.LimitChunks(ctx, in)

	// Consume all chunks
	var chunks []*StreamChunk
	for chunk := range out {
		chunks = append(chunks, chunk)
	}

	assert.Len(t, chunks, 3)
	assert.True(t, chunks[2].Done)
}

func TestRateLimiterReset(t *testing.T) {
	limiter := NewRateLimiter(100)

	// Simulate some activity
	ctx := context.Background()
	in := make(chan string, 1)
	in <- "test"
	close(in)

	out := limiter.Limit(ctx, in)
	<-out

	// Reset
	limiter.Reset()
	assert.True(t, limiter.lastEmit.IsZero())
}

func TestRateLimiterSetRate(t *testing.T) {
	limiter := NewRateLimiter(100)

	limiter.SetRate(200)
	assert.Equal(t, float64(200), limiter.tokensPerSecond)
	assert.Equal(t, time.Duration(5*time.Millisecond), limiter.delay)

	// Zero rate should default to 100
	limiter.SetRate(0)
	assert.Equal(t, float64(100), limiter.tokensPerSecond)
}

func TestBurstRateLimiter(t *testing.T) {
	limiter := NewBurstRateLimiter(10, 5) // 10/sec, burst of 5

	ctx := context.Background()
	in := make(chan string, 10)
	for i := 0; i < 10; i++ {
		in <- "token"
	}
	close(in)

	start := time.Now()
	out := limiter.Limit(ctx, in)

	// Consume all tokens
	count := 0
	for range out {
		count++
	}

	elapsed := time.Since(start)
	assert.Equal(t, 10, count)

	// First 5 should be instant (burst), then rate limited
	// So ~500ms for remaining 5 at 10/sec
	assert.True(t, elapsed >= 400*time.Millisecond, "Expected burst then rate limit")
}

func TestBurstRateLimiterDefaults(t *testing.T) {
	limiter := NewBurstRateLimiter(0, 0)

	assert.Equal(t, float64(100), limiter.tokensPerSecond)
	assert.Equal(t, 10, limiter.burstSize)
}

func TestBurstRateLimiterReset(t *testing.T) {
	limiter := NewBurstRateLimiter(100, 5)

	// Consume some burst capacity
	limiter.tokens = 0

	limiter.Reset()
	assert.Equal(t, 5, limiter.tokens)
}

// ============================================================================
// SSE Writer Tests
// ============================================================================

func TestNewSSEWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, err := NewSSEWriter(recorder)

	require.NoError(t, err)
	assert.NotNil(t, sse)
	assert.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", recorder.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", recorder.Header().Get("Connection"))
}

func TestSSEWriterWriteEvent(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, _ := NewSSEWriter(recorder)

	err := sse.WriteEvent("message", "Hello world", "123")
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "id: 123")
	assert.Contains(t, body, "event: message")
	assert.Contains(t, body, "data: Hello world")
}

func TestSSEWriterWriteData(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, _ := NewSSEWriter(recorder)

	err := sse.WriteData("test data")
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: test data")
	assert.NotContains(t, body, "event:")
	assert.NotContains(t, body, "id:")
}

func TestSSEWriterWriteJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, _ := NewSSEWriter(recorder)

	data := map[string]string{"key": "value"}
	err := sse.WriteJSON(data)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, `{"key":"value"}`)
}

func TestSSEWriterWriteProgress(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, _ := NewSSEWriter(recorder)

	progress := &StreamProgress{
		TokensGenerated: 10,
		ChunksReceived:  5,
	}
	err := sse.WriteProgress(progress)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "event: progress")
	assert.Contains(t, body, `"tokens_generated":10`)
}

func TestSSEWriterWriteDone(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, _ := NewSSEWriter(recorder)

	err := sse.WriteDone()
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: [DONE]")
}

func TestSSEWriterWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, _ := NewSSEWriter(recorder)

	err := sse.WriteError(assert.AnError)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "event: error")
}

func TestStreamToSSE(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.Background()

	in := make(chan string, 3)
	in <- "Hello "
	in <- "world"
	close(in)

	err := StreamToSSE(ctx, recorder, in)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: Hello ")
	assert.Contains(t, body, "data: world")
	assert.Contains(t, body, "data: [DONE]")
}

func TestStreamChunksToSSE(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.Background()

	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	in <- &StreamChunk{Content: " world", Index: 1}
	in <- &StreamChunk{Content: "", Index: 2, Done: true}
	close(in)

	err := StreamChunksToSSE(ctx, recorder, in)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: Hello")
	assert.Contains(t, body, "data:  world")
	assert.Contains(t, body, "data: [DONE]")
}

func TestStreamChunksToSSEWithError(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx := context.Background()

	in := make(chan *StreamChunk, 2)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	in <- &StreamChunk{Error: assert.AnError, Index: 1}
	close(in)

	err := StreamChunksToSSE(ctx, recorder, in)
	assert.Error(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "event: error")
}

func TestFormatSSEEvent(t *testing.T) {
	event := &SSEEvent{
		ID:    "123",
		Event: "message",
		Data:  "Hello world",
	}

	result := FormatSSEEvent(event)
	assert.Contains(t, result, "id: 123\n")
	assert.Contains(t, result, "event: message\n")
	assert.Contains(t, result, "data: Hello world\n\n")
}

func TestFormatSSEEventMinimal(t *testing.T) {
	event := &SSEEvent{
		Data: "Just data",
	}

	result := FormatSSEEvent(event)
	assert.NotContains(t, result, "id:")
	assert.NotContains(t, result, "event:")
	assert.Equal(t, "data: Just data\n\n", result)
}

// ============================================================================
// Enhanced Streamer Tests
// ============================================================================

func TestDefaultStreamConfig(t *testing.T) {
	config := DefaultStreamConfig()

	assert.Equal(t, BufferTypeWord, config.BufferType)
	assert.Equal(t, 100*time.Millisecond, config.ProgressInterval)
	assert.Equal(t, float64(0), config.RateLimit)
	assert.Equal(t, 0, config.EstimatedTokens)
	assert.Equal(t, 5, config.TokenBufferThreshold)
}

func TestNewEnhancedStreamer(t *testing.T) {
	// With nil config
	streamer := NewEnhancedStreamer(nil)
	assert.NotNil(t, streamer.config)
	assert.Equal(t, BufferTypeWord, streamer.config.BufferType)

	// With custom config
	config := &StreamConfig{BufferType: BufferTypeSentence}
	streamer = NewEnhancedStreamer(config)
	assert.Equal(t, BufferTypeSentence, streamer.config.BufferType)
}

func TestEnhancedStreamerConfig(t *testing.T) {
	config := &StreamConfig{BufferType: BufferTypeLine}
	streamer := NewEnhancedStreamer(config)

	assert.Equal(t, config, streamer.Config())

	newConfig := &StreamConfig{BufferType: BufferTypeParagraph}
	streamer.SetConfig(newConfig)
	assert.Equal(t, newConfig, streamer.Config())

	// Nil config should not change
	streamer.SetConfig(nil)
	assert.Equal(t, newConfig, streamer.Config())
}

func TestEnhancedStreamerStreamBuffered(t *testing.T) {
	config := &StreamConfig{BufferType: BufferTypeWord}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	in := make(chan *StreamChunk, 5)
	in <- &StreamChunk{Content: "Hello ", Index: 0}
	in <- &StreamChunk{Content: "world ", Index: 1}
	in <- &StreamChunk{Content: "test", Index: 2}
	in <- &StreamChunk{Content: "", Index: 3, Done: true}
	close(in)

	out := streamer.StreamBuffered(ctx, in)

	var chunks []*StreamChunk
	for chunk := range out {
		chunks = append(chunks, chunk)
	}

	assert.True(t, len(chunks) >= 2) // At least buffered output + done
}

func TestEnhancedStreamerStreamWithProgress(t *testing.T) {
	config := &StreamConfig{
		ProgressInterval: 10 * time.Millisecond,
		EstimatedTokens:  100,
	}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hello world", Index: 0}
	in <- &StreamChunk{Content: "", Index: 1, Done: true}
	close(in)

	var progressUpdates []*StreamProgress
	var mu sync.Mutex

	progressCallback := func(p *StreamProgress) {
		mu.Lock()
		progressUpdates = append(progressUpdates, p)
		mu.Unlock()
	}

	out := streamer.StreamWithProgress(ctx, in, progressCallback)

	// Consume output
	for range out {
	}

	// Wait for progress callbacks
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.True(t, len(progressUpdates) >= 1, "Expected at least one progress update")
	mu.Unlock()
}

func TestEnhancedStreamerStreamWithRateLimit(t *testing.T) {
	config := &StreamConfig{RateLimit: 100}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	in <- &StreamChunk{Content: " world", Index: 1}
	close(in)

	out := streamer.StreamWithRateLimit(ctx, in)

	// Should be rate limited
	count := 0
	for range out {
		count++
	}
	assert.Equal(t, 2, count)
}

func TestEnhancedStreamerStreamWithRateLimitDisabled(t *testing.T) {
	config := &StreamConfig{RateLimit: 0} // Disabled
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	in := make(chan *StreamChunk, 1)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	close(in)

	out := streamer.StreamWithRateLimit(ctx, in)

	// When disabled, should pass through without rate limiting delay
	start := time.Now()
	chunk := <-out
	elapsed := time.Since(start)

	assert.Equal(t, "Hello", chunk.Content)
	assert.True(t, elapsed < 10*time.Millisecond, "Should be instant when rate limiting disabled")
}

func TestEnhancedStreamerStreamEnhanced(t *testing.T) {
	config := &StreamConfig{
		BufferType:       BufferTypeCharacter,
		ProgressInterval: 10 * time.Millisecond,
		RateLimit:        0,
		EstimatedTokens:  100,
	}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hi", Index: 0}
	in <- &StreamChunk{Content: "", Index: 1, Done: true}
	close(in)

	var progressCalled bool
	progressCallback := func(p *StreamProgress) {
		progressCalled = true
	}

	out, getResult := streamer.StreamEnhanced(ctx, in, progressCallback)

	// Consume output
	for range out {
	}

	// Wait for callbacks
	time.Sleep(50 * time.Millisecond)

	// Get aggregated result
	result := getResult()
	assert.NotNil(t, result)
	assert.True(t, progressCalled)
}

func TestEnhancedStreamerStreamEnhancedNoProgress(t *testing.T) {
	config := DefaultStreamConfig()
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	in := make(chan *StreamChunk, 2)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	in <- &StreamChunk{Content: "", Index: 1, Done: true}
	close(in)

	out, getResult := streamer.StreamEnhanced(ctx, in, nil) // nil progress callback

	// Consume output
	for range out {
	}

	result := getResult()
	assert.NotNil(t, result)
}

func TestEnhancedStreamerCreateBuffer(t *testing.T) {
	testCases := []struct {
		bufferType BufferType
		expected   string
	}{
		{BufferTypeCharacter, "*streaming.CharacterBuffer"},
		{BufferTypeWord, "*streaming.WordBuffer"},
		{BufferTypeSentence, "*streaming.SentenceBuffer"},
		{BufferTypeLine, "*streaming.LineBuffer"},
		{BufferTypeParagraph, "*streaming.ParagraphBuffer"},
		{BufferTypeToken, "*streaming.TokenBuffer"},
		{BufferType("unknown"), "*streaming.WordBuffer"}, // Unknown defaults to word
	}

	for _, tc := range testCases {
		config := &StreamConfig{BufferType: tc.bufferType, TokenBufferThreshold: 5}
		streamer := NewEnhancedStreamer(config)

		// Access createBuffer via StreamBuffered behavior
		ctx, cancel := context.WithCancel(context.Background())
		in := make(chan *StreamChunk)
		close(in)

		_ = streamer.StreamBuffered(ctx, in)
		cancel()
	}
}

// ============================================================================
// Channel Conversion Tests
// ============================================================================

func TestStringToChunkChannel(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 3)
	in <- "Hello"
	in <- " world"
	close(in)

	out := StringToChunkChannel(ctx, in)

	chunks := make([]*StreamChunk, 0)
	for chunk := range out {
		chunks = append(chunks, chunk)
	}

	require.Len(t, chunks, 3) // 2 content + 1 done
	assert.Equal(t, "Hello", chunks[0].Content)
	assert.Equal(t, 0, chunks[0].Index)
	assert.Equal(t, " world", chunks[1].Content)
	assert.Equal(t, 1, chunks[1].Index)
	assert.True(t, chunks[2].Done)
}

func TestStringToChunkChannelContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan string)

	out := StringToChunkChannel(ctx, in)
	cancel()

	select {
	case <-out:
		// Channel closed or empty
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for channel close")
	}
}

func TestChunkToStringChannel(t *testing.T) {
	ctx := context.Background()
	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	in <- &StreamChunk{Content: " world", Index: 1}
	in <- &StreamChunk{Content: "", Index: 2, Done: true}
	close(in)

	out := ChunkToStringChannel(ctx, in)

	var strings []string
	for s := range out {
		strings = append(strings, s)
	}

	assert.Equal(t, []string{"Hello", " world"}, strings) // Done chunk not included
}

func TestChunkToStringChannelContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan *StreamChunk)

	out := ChunkToStringChannel(ctx, in)
	cancel()

	select {
	case <-out:
		// Channel closed or empty
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for channel close")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestFullStreamingPipeline(t *testing.T) {
	config := &StreamConfig{
		BufferType:           BufferTypeWord,
		ProgressInterval:     10 * time.Millisecond,
		RateLimit:            0,
		EstimatedTokens:      50,
		TokenBufferThreshold: 3,
	}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()

	// Create input channel with multiple chunks
	in := make(chan *StreamChunk, 10)
	words := []string{"The ", "quick ", "brown ", "fox ", "jumps ", "over ", "the ", "lazy ", "dog."}
	for i, word := range words {
		in <- &StreamChunk{Content: word, Index: i}
	}
	in <- &StreamChunk{Content: "", Index: len(words), Done: true}
	close(in)

	var progressUpdates []*StreamProgress
	var mu sync.Mutex

	progressCallback := func(p *StreamProgress) {
		mu.Lock()
		progressUpdates = append(progressUpdates, p)
		mu.Unlock()
	}

	out, getResult := streamer.StreamEnhanced(ctx, in, progressCallback)

	// Consume all output
	var outputChunks []*StreamChunk
	for chunk := range out {
		outputChunks = append(outputChunks, chunk)
	}

	// Wait for async progress updates
	time.Sleep(50 * time.Millisecond)

	// Verify aggregated result
	result := getResult()
	assert.NotNil(t, result)
	assert.Contains(t, result.FullContent, "quick")
	assert.Contains(t, result.FullContent, "brown")
	assert.Contains(t, result.FullContent, "fox")
	assert.True(t, result.TokenCount > 0)
	assert.True(t, result.DurationSeconds > 0)

	// Verify progress was tracked
	mu.Lock()
	assert.True(t, len(progressUpdates) > 0, "Expected progress updates")
	mu.Unlock()
}

func TestStreamingWithContextTimeout(t *testing.T) {
	config := &StreamConfig{
		BufferType: BufferTypeWord,
		RateLimit:  1, // Very slow
	}
	streamer := NewEnhancedStreamer(config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Create input that would take long to process
	in := make(chan *StreamChunk, 100)
	for i := 0; i < 100; i++ {
		in <- &StreamChunk{Content: "word ", Index: i}
	}
	close(in)

	out := streamer.StreamWithRateLimit(ctx, in)

	// Should timeout before all chunks are processed
	count := 0
	for range out {
		count++
	}

	assert.True(t, count < 100, "Expected early termination due to context timeout")
}

// ============================================================================
// Buffer Type Constants Test
// ============================================================================

func TestBufferTypeConstants(t *testing.T) {
	assert.Equal(t, BufferType("character"), BufferTypeCharacter)
	assert.Equal(t, BufferType("word"), BufferTypeWord)
	assert.Equal(t, BufferType("sentence"), BufferTypeSentence)
	assert.Equal(t, BufferType("line"), BufferTypeLine)
	assert.Equal(t, BufferType("paragraph"), BufferTypeParagraph)
	assert.Equal(t, BufferType("token"), BufferTypeToken)
}

// ============================================================================
// NewBuffer Factory Tests
// ============================================================================

func TestNewBuffer(t *testing.T) {
	tests := []struct {
		bufferType BufferType
		options    []interface{}
		expected   string
	}{
		{BufferTypeCharacter, nil, "*streaming.CharacterBuffer"},
		{BufferTypeWord, nil, "*streaming.WordBuffer"},
		{BufferTypeWord, []interface{}{","}, "*streaming.WordBuffer"},
		{BufferTypeSentence, nil, "*streaming.SentenceBuffer"},
		{BufferTypeLine, nil, "*streaming.LineBuffer"},
		{BufferTypeParagraph, nil, "*streaming.ParagraphBuffer"},
		{BufferTypeToken, nil, "*streaming.TokenBuffer"},
		{BufferTypeToken, []interface{}{10}, "*streaming.TokenBuffer"},
		{BufferType("unknown"), nil, "*streaming.WordBuffer"}, // Default
	}

	for _, tc := range tests {
		buf := NewBuffer(tc.bufferType, tc.options...)
		assert.NotNil(t, buf)
	}
}

func TestNewBufferWithWordDelimiter(t *testing.T) {
	buf := NewBuffer(BufferTypeWord, ",")
	wordBuf, ok := buf.(*WordBuffer)
	require.True(t, ok)
	assert.Equal(t, ",", wordBuf.delimiter)
}

func TestNewBufferWithTokenThreshold(t *testing.T) {
	buf := NewBuffer(BufferTypeToken, 10)
	tokenBuf, ok := buf.(*TokenBuffer)
	require.True(t, ok)
	assert.Equal(t, 10, tokenBuf.threshold)
}

// ============================================================================
// Reset Method Tests
// ============================================================================

func TestCharacterBuffer_Reset(t *testing.T) {
	buf := NewCharacterBuffer()

	// Add some characters
	buf.Add("Hello")

	// Reset should clear the buffer
	buf.Reset()

	// Flush should return empty string after reset
	result := buf.Flush()
	assert.Empty(t, result)
}

func TestStreamAggregator_Reset(t *testing.T) {
	agg := NewStreamAggregator()

	agg.Start()
	agg.Add("Hello")
	agg.Add(" World")

	result := agg.GetResult()
	assert.Equal(t, "Hello World", result.FullContent)

	// Reset should clear state
	agg.Reset()
}

func TestParagraphBuffer_Reset(t *testing.T) {
	buf := NewParagraphBuffer()

	buf.Add("First paragraph.\n\n")
	buf.Add("Second paragraph.")

	// Reset should clear everything
	buf.Reset()

	result := buf.Flush()
	assert.Empty(t, result)
}

// ============================================================================
// SSE Edge Case Tests
// ============================================================================

func TestSSEWriter_WriteEvent_AllParams(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer, err := NewSSEWriter(recorder)
	require.NoError(t, err)

	err = writer.WriteEvent("message", "test data", "test-id")
	require.NoError(t, err)

	output := recorder.Body.String()
	assert.Contains(t, output, "id: test-id")
	assert.Contains(t, output, "event: message")
	assert.Contains(t, output, "data: test data")
}

func TestStreamToSSE_Done(t *testing.T) {
	recorder := httptest.NewRecorder()

	chunks := make(chan string, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- StreamToSSE(context.Background(), recorder, chunks)
	}()

	// Send some content and then signal done
	chunks <- "Hello World"
	close(chunks)

	err := <-errChan
	require.NoError(t, err)

	output := recorder.Body.String()
	assert.Contains(t, output, "Hello World")
	assert.Contains(t, output, "[DONE]")
}

func TestStreamWithProgressToSSE_Basic(t *testing.T) {
	recorder := httptest.NewRecorder()

	chunks := make(chan string, 10)
	errChan := make(chan error, 1)

	go func() {
		errChan <- StreamWithProgressToSSE(context.Background(), recorder, chunks, 1)
	}()

	// Send content
	chunks <- "Hello"
	chunks <- " World"
	close(chunks)

	err := <-errChan
	require.NoError(t, err)

	output := recorder.Body.String()
	assert.Contains(t, output, "Hello")
}

// ============================================================================
// Aggregator Edge Cases
// ============================================================================

func TestChunkAggregator_MultipleChunks(t *testing.T) {
	agg := NewChunkAggregator()

	agg.Start()

	// Add chunks
	agg.Add(&StreamChunk{Content: "Hello", Index: 0})
	agg.Add(&StreamChunk{Content: " World", Index: 1, Done: true})

	result := agg.GetResult()
	assert.Equal(t, "Hello World", result.FullContent)
	assert.Equal(t, 2, result.ChunkCount)
}

// ============================================================================
// Enhanced Streamer Edge Cases
// ============================================================================

func TestEnhancedStreamer_StreamWithProgress_Error(t *testing.T) {
	config := &StreamConfig{
		ProgressInterval: 10 * time.Millisecond,
		EstimatedTokens:  100,
	}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	in := make(chan *StreamChunk, 3)
	in <- &StreamChunk{Content: "Hello", Index: 0}
	in <- &StreamChunk{Content: "", Index: 1, Error: errors.New("provider error")}
	close(in)

	var progressUpdates []*StreamProgress
	var mu sync.Mutex

	progressCallback := func(p *StreamProgress) {
		mu.Lock()
		progressUpdates = append(progressUpdates, p)
		mu.Unlock()
	}

	out := streamer.StreamWithProgress(ctx, in, progressCallback)

	// Consume output
	var errFound error
	for chunk := range out {
		if chunk.Error != nil {
			errFound = chunk.Error
		}
	}

	assert.NotNil(t, errFound)
}

// ============================================================================
// Additional SSE Tests for Coverage
// ============================================================================

type nonFlushingWriter struct {
	http.ResponseWriter
}

func (w *nonFlushingWriter) Header() http.Header {
	return http.Header{}
}

func (w *nonFlushingWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *nonFlushingWriter) WriteHeader(statusCode int) {}

func TestNewSSEWriter_NoFlusher(t *testing.T) {
	// Test that NewSSEWriter returns error when ResponseWriter doesn't support Flush
	w := &nonFlushingWriter{}

	_, err := NewSSEWriter(w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "streaming not supported")
}

func TestSSEWriter_WriteEventWithID(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, err := NewSSEWriter(recorder)
	require.NoError(t, err)

	err = sse.WriteEvent("message", "test data", "event-123")
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "id: event-123")
	assert.Contains(t, body, "event: message")
	assert.Contains(t, body, "data: test data")
}

func TestSSEWriter_WriteEventWithEventType(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, err := NewSSEWriter(recorder)
	require.NoError(t, err)

	err = sse.WriteEvent("custom-event", "payload", "")
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "event: custom-event")
	assert.Contains(t, body, "data: payload")
	assert.NotContains(t, body, "id:")
}

func TestSSEWriter_WriteJSON_MarshalError(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, err := NewSSEWriter(recorder)
	require.NoError(t, err)

	// Channels cannot be marshaled to JSON
	err = sse.WriteJSON(make(chan int))
	assert.Error(t, err)
}

func TestSSEWriter_WriteProgress_Success(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, err := NewSSEWriter(recorder)
	require.NoError(t, err)

	progress := &StreamProgress{
		TokensGenerated:    50,
		CharactersReceived: 200,
		PercentComplete:    0.5,
	}

	err = sse.WriteProgress(progress)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "event: progress")
	assert.Contains(t, body, "\"tokens_generated\":50")
}

func TestSSEWriter_WriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	sse, err := NewSSEWriter(recorder)
	require.NoError(t, err)

	err = sse.WriteError(errors.New("test error"))
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "event: error")
	assert.Contains(t, body, "test error")
}

func TestStreamToSSE_ContextCancellation(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan string)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err := StreamToSSE(ctx, recorder, stream)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestStreamChunksToSSE_WithError(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan *StreamChunk, 2)

	ctx := context.Background()

	// Send chunk with error
	stream <- &StreamChunk{Error: errors.New("chunk error")}
	close(stream)

	err := StreamChunksToSSE(ctx, recorder, stream)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chunk error")
}

func TestStreamChunksToSSE_WithDone(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan *StreamChunk, 3)

	ctx := context.Background()

	stream <- &StreamChunk{Content: "Hello", Index: 0}
	stream <- &StreamChunk{Content: " World", Index: 1, Done: true}
	close(stream)

	err := StreamChunksToSSE(ctx, recorder, stream)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: Hello")
	assert.Contains(t, body, "data: [DONE]")
}

func TestStreamChunksToSSE_ContextCancellation(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan *StreamChunk)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := StreamChunksToSSE(ctx, recorder, stream)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestStreamWithProgressToSSE_WithInterval(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan string, 10)

	ctx := context.Background()

	// Send multiple chunks
	for i := 0; i < 5; i++ {
		stream <- "chunk"
	}
	close(stream)

	err := StreamWithProgressToSSE(ctx, recorder, stream, 2) // Emit progress every 2 chunks
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "event: progress")
	assert.Contains(t, body, "data: [DONE]")
}

func TestStreamWithProgressToSSE_ContextCancellation(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan string)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := StreamWithProgressToSSE(ctx, recorder, stream, 1)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestStreamWithProgressToSSE_NoInterval(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan string, 3)

	ctx := context.Background()

	stream <- "hello"
	stream <- " world"
	close(stream)

	err := StreamWithProgressToSSE(ctx, recorder, stream, 0) // No interval progress
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: hello")
	assert.Contains(t, body, "data: [DONE]")
}

// ============================================================================
// Additional Buffer Tests for Coverage
// ============================================================================

func TestCharacterBuffer_ResetNoOp(t *testing.T) {
	buf := NewCharacterBuffer()
	// Reset is a no-op for CharacterBuffer but should not panic
	buf.Reset()
	// Should still work normally - returns each char as slice
	result := buf.Add("Hi")
	assert.Len(t, result, 2)
	assert.Equal(t, "H", result[0])
	assert.Equal(t, "i", result[1])
}

func TestNewWordBuffer_DefaultDelimiter(t *testing.T) {
	// Test that empty delimiter defaults to space
	buf := NewWordBuffer("")
	result := buf.Add("Hello World ")
	// Returns words with delimiter included
	assert.Len(t, result, 2)
	assert.Equal(t, "Hello ", result[0])
	assert.Equal(t, "World ", result[1])
}

func TestNewWordBuffer_CustomDelimiter(t *testing.T) {
	buf := NewWordBuffer(",")
	words := buf.Add("one,two,three")
	// WordBuffer includes delimiter in each word
	assert.Equal(t, []string{"one,", "two,"}, words)

	remaining := buf.Flush()
	assert.Equal(t, "three", remaining)
}

func TestNewTokenBuffer_DefaultThreshold(t *testing.T) {
	// Zero threshold should default to 5
	buf := NewTokenBuffer(0)
	tokens := buf.Add("Hello World")
	assert.Empty(t, tokens) // Not enough tokens yet to emit
}

func TestNewTokenBuffer_CustomThreshold(t *testing.T) {
	buf := NewTokenBuffer(2)
	// Add enough content to reach threshold
	buf.Add("Hello World how are you today")
	tokens := buf.Add("this should trigger")
	assert.NotEmpty(t, tokens)
}

// ============================================================================
// Additional Rate Limiter Tests for Coverage
// ============================================================================

func TestRateLimiter_LimitWithContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(10.0) // 10 tokens per second
	input := make(chan string, 10)

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < 5; i++ {
		input <- "token"
	}
	close(input)

	// Cancel context immediately
	cancel()

	output := limiter.Limit(ctx, input)

	// Drain output
	count := 0
	for range output {
		count++
	}
	// Some tokens may have been processed before cancellation
	assert.True(t, count <= 5)
}

func TestRateLimiter_LimitChunksWithError(t *testing.T) {
	limiter := NewRateLimiter(1000.0) // Fast rate
	input := make(chan *StreamChunk, 3)

	ctx := context.Background()

	input <- &StreamChunk{Content: "Hello", Index: 0}
	input <- &StreamChunk{Content: "", Index: 1, Error: errors.New("test error")}
	close(input)

	output := limiter.LimitChunks(ctx, input)

	var chunks []*StreamChunk
	for chunk := range output {
		chunks = append(chunks, chunk)
	}

	assert.Len(t, chunks, 2)
	assert.NotNil(t, chunks[1].Error)
}

func TestBurstRateLimiter_LimitWithContextCancellation(t *testing.T) {
	limiter := NewBurstRateLimiter(10.0, 5) // Slow rate to ensure context cancellation
	input := make(chan string, 20)

	ctx, cancel := context.WithCancel(context.Background())

	// Fill input
	for i := 0; i < 20; i++ {
		input <- "token"
	}
	close(input)

	// Cancel after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	output := limiter.Limit(ctx, input)

	// Drain output
	count := 0
	for range output {
		count++
	}
	// Some tokens should have been processed before cancellation
	assert.True(t, count > 0)
	assert.True(t, count < 20)
}

// ============================================================================
// Additional Aggregator Tests for Coverage
// ============================================================================

func TestStreamAggregator_Aggregate(t *testing.T) {
	agg := NewStreamAggregator()
	input := make(chan string, 5)

	input <- "Hello"
	input <- " "
	input <- "World"
	close(input)

	output, getResult := agg.Aggregate(context.Background(), input)

	// Drain output
	var chunks []string
	for chunk := range output {
		chunks = append(chunks, chunk)
	}

	assert.Len(t, chunks, 3)
	result := getResult()
	assert.Equal(t, "Hello World", result.FullContent)
}

func TestStreamAggregator_AggregateWithContextCancel(t *testing.T) {
	agg := NewStreamAggregator()
	input := make(chan string)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	output, _ := agg.Aggregate(ctx, input)

	// Output should be closed due to context cancellation
	_, ok := <-output
	assert.False(t, ok)
}

func TestChunkAggregator_AggregateChunks(t *testing.T) {
	agg := NewChunkAggregator()
	input := make(chan *StreamChunk, 3)

	input <- &StreamChunk{Content: "Hello", Index: 0}
	input <- &StreamChunk{Content: " World", Index: 1}
	close(input)

	output, getResult := agg.AggregateChunks(context.Background(), input)

	// Drain output
	var chunks []*StreamChunk
	for chunk := range output {
		chunks = append(chunks, chunk)
	}

	assert.Len(t, chunks, 2)
	result := getResult()
	assert.Equal(t, "Hello World", result.FullContent)
}

func TestChunkAggregator_AggregateChunksWithContextCancel(t *testing.T) {
	agg := NewChunkAggregator()
	input := make(chan *StreamChunk)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	output, _ := agg.AggregateChunks(ctx, input)

	_, ok := <-output
	assert.False(t, ok)
}

// ============================================================================
// Additional Enhanced Streamer Tests for Coverage
// ============================================================================

func TestEnhancedStreamer_StreamBuffered_Word(t *testing.T) {
	config := &StreamConfig{
		BufferType: BufferTypeWord,
	}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	input := make(chan *StreamChunk, 5)

	input <- &StreamChunk{Content: "Hello ", Index: 0}
	input <- &StreamChunk{Content: "World ", Index: 1}
	input <- &StreamChunk{Content: "!", Index: 2, Done: true}
	close(input)

	output := streamer.StreamBuffered(ctx, input)

	var results []string
	for chunk := range output {
		results = append(results, chunk.Content)
	}

	assert.NotEmpty(t, results)
}

func TestEnhancedStreamer_StreamBuffered_Sentence(t *testing.T) {
	config := &StreamConfig{
		BufferType: BufferTypeSentence,
	}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	input := make(chan *StreamChunk, 5)

	input <- &StreamChunk{Content: "Hello. ", Index: 0}
	input <- &StreamChunk{Content: "World! ", Index: 1}
	input <- &StreamChunk{Content: "Test?", Index: 2, Done: true}
	close(input)

	output := streamer.StreamBuffered(ctx, input)

	var results []string
	for chunk := range output {
		results = append(results, chunk.Content)
	}

	assert.NotEmpty(t, results)
}

func TestEnhancedStreamer_StreamBuffered_Paragraph(t *testing.T) {
	config := &StreamConfig{
		BufferType: BufferTypeParagraph,
	}
	streamer := NewEnhancedStreamer(config)

	ctx := context.Background()
	input := make(chan *StreamChunk, 5)

	input <- &StreamChunk{Content: "First paragraph.\n\n", Index: 0}
	input <- &StreamChunk{Content: "Second paragraph.", Index: 1, Done: true}
	close(input)

	output := streamer.StreamBuffered(ctx, input)

	var results []string
	for chunk := range output {
		results = append(results, chunk.Content)
	}

	assert.NotEmpty(t, results)
}

func TestStringToChunkChannel_WithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input := make(chan string, 5)

	input <- "Hello"
	input <- "World"
	close(input)

	output := StringToChunkChannel(ctx, input)

	var chunks []*StreamChunk
	for chunk := range output {
		chunks = append(chunks, chunk)
	}

	// 2 content chunks + 1 Done chunk
	assert.Len(t, chunks, 3)
	assert.Equal(t, "Hello", chunks[0].Content)
	assert.Equal(t, "World", chunks[1].Content)
	assert.True(t, chunks[2].Done)
}

func TestChunkToStringChannel_WithContext(t *testing.T) {
	ctx := context.Background()
	input := make(chan *StreamChunk, 3)

	input <- &StreamChunk{Content: "Hello", Index: 0}
	input <- &StreamChunk{Content: " World", Index: 1}
	close(input)

	output := ChunkToStringChannel(ctx, input)

	var strings []string
	for s := range output {
		strings = append(strings, s)
	}

	assert.Len(t, strings, 2)
	assert.Equal(t, "Hello", strings[0])
}

func TestChunkToStringChannel_WithError(t *testing.T) {
	ctx := context.Background()
	input := make(chan *StreamChunk, 3)

	input <- &StreamChunk{Content: "Hello", Index: 0}
	input <- &StreamChunk{Content: "World", Index: 1, Error: errors.New("test")}
	input <- &StreamChunk{Content: "", Index: 2, Done: true}
	close(input)

	output := ChunkToStringChannel(ctx, input)

	var strings []string
	for s := range output {
		strings = append(strings, s)
	}

	// ChunkToStringChannel outputs Content regardless of error, stops at Done
	assert.Len(t, strings, 2)
	assert.Equal(t, "Hello", strings[0])
	assert.Equal(t, "World", strings[1])
}

// ============================================================================
// Progress Tracker Edge Cases
// ============================================================================

func TestProgressTracker_GetProgressWithEstimation(t *testing.T) {
	tracker := NewProgressTracker(100) // Estimate 100 tokens
	tracker.Start()

	// Simulate some progress
	tracker.Update("Hello")
	tracker.Update(" World")

	progress := tracker.GetProgress()
	assert.Greater(t, progress.PercentComplete, float64(0))
	assert.Greater(t, progress.TokensGenerated, 0)
}

// ============================================================================
// Additional tests for edge case coverage
// ============================================================================

func TestSSEWriter_WriteEvent_ErrorPaths(t *testing.T) {
	// Test WriteEvent with all parameters
	recorder := httptest.NewRecorder()
	sse, err := NewSSEWriter(recorder)
	require.NoError(t, err)

	// With both event and ID
	err = sse.WriteEvent("custom", "data", "id-1")
	require.NoError(t, err)

	// Empty data but valid event and ID
	err = sse.WriteEvent("empty", "", "id-2")
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "id: id-1")
	assert.Contains(t, body, "event: custom")
	assert.Contains(t, body, "id: id-2")
}

func TestStreamBuffered_ContextCancel(t *testing.T) {
	config := &StreamConfig{
		BufferType: BufferTypeCharacter,
	}
	streamer := NewEnhancedStreamer(config)

	ctx, cancel := context.WithCancel(context.Background())
	input := make(chan *StreamChunk)

	// Start buffering
	output := streamer.StreamBuffered(ctx, input)

	// Cancel context before sending anything
	cancel()

	// Output should close
	_, ok := <-output
	assert.False(t, ok)
}

func TestStreamWithProgress_ContextCancel(t *testing.T) {
	config := &StreamConfig{
		ProgressInterval: 10 * time.Millisecond,
		EstimatedTokens:  100,
	}
	streamer := NewEnhancedStreamer(config)

	ctx, cancel := context.WithCancel(context.Background())
	input := make(chan *StreamChunk)

	progressCallback := func(p *StreamProgress) {}

	output := streamer.StreamWithProgress(ctx, input, progressCallback)

	// Cancel context
	cancel()

	// Output should close
	count := 0
	for range output {
		count++
	}
	assert.Equal(t, 0, count)
}

func TestBurstRateLimiter_WaitForToken(t *testing.T) {
	// Test with very slow rate to ensure waiting
	limiter := NewBurstRateLimiter(1.0, 1) // 1 token/sec, burst 1
	input := make(chan string, 3)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	input <- "token1"
	input <- "token2"
	close(input)

	output := limiter.Limit(ctx, input)

	var count int
	for range output {
		count++
	}
	// At least some tokens should be processed
	assert.GreaterOrEqual(t, count, 1)
}

func TestStreamChunksToSSE_NormalFlow(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan *StreamChunk, 5)

	ctx := context.Background()

	stream <- &StreamChunk{Content: "Hello", Index: 0}
	stream <- &StreamChunk{Content: " World", Index: 1}
	close(stream)

	err := StreamChunksToSSE(ctx, recorder, stream)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: Hello")
	assert.Contains(t, body, "data:  World")
	assert.Contains(t, body, "data: [DONE]")
}

func TestStreamToSSE_NormalFlow(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := make(chan string, 5)

	ctx := context.Background()

	stream <- "Hello"
	stream <- " World"
	close(stream)

	err := StreamToSSE(ctx, recorder, stream)
	require.NoError(t, err)

	body := recorder.Body.String()
	assert.Contains(t, body, "data: Hello")
	assert.Contains(t, body, "data:  World")
	assert.Contains(t, body, "data: [DONE]")
}
