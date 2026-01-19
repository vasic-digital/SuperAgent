// Package streaming provides enhanced streaming capabilities for LLM responses.
package streaming

import (
	"time"
)

// StreamConfig holds streaming configuration.
// Note: This is defined in enhanced_streamer.go, but we extend it here with additional types.

// StreamState represents the state of a stream.
type StreamState string

const (
	// StreamStatePending indicates the stream hasn't started.
	StreamStatePending StreamState = "pending"
	// StreamStateActive indicates the stream is actively receiving data.
	StreamStateActive StreamState = "active"
	// StreamStatePaused indicates the stream is paused.
	StreamStatePaused StreamState = "paused"
	// StreamStateCompleted indicates the stream completed successfully.
	StreamStateCompleted StreamState = "completed"
	// StreamStateFailed indicates the stream failed.
	StreamStateFailed StreamState = "failed"
	// StreamStateCancelled indicates the stream was cancelled.
	StreamStateCancelled StreamState = "cancelled"
)

// StreamMetadata contains metadata about a stream.
type StreamMetadata struct {
	// ID is the stream identifier.
	ID string `json:"id"`
	// CreatedAt is when the stream was created.
	CreatedAt time.Time `json:"created_at"`
	// StartedAt is when streaming started.
	StartedAt time.Time `json:"started_at,omitempty"`
	// CompletedAt is when streaming completed.
	CompletedAt time.Time `json:"completed_at,omitempty"`
	// State is the current stream state.
	State StreamState `json:"state"`
	// Model is the LLM model being used.
	Model string `json:"model,omitempty"`
	// Provider is the LLM provider.
	Provider string `json:"provider,omitempty"`
	// EstimatedTokens is the estimated total tokens.
	EstimatedTokens int `json:"estimated_tokens,omitempty"`
	// Metadata contains additional metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StreamEvent represents an event during streaming.
type StreamEvent struct {
	// Type is the event type.
	Type StreamEventType `json:"type"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// StreamID is the stream identifier.
	StreamID string `json:"stream_id,omitempty"`
	// Data contains event-specific data.
	Data interface{} `json:"data,omitempty"`
	// Error contains error information if applicable.
	Error string `json:"error,omitempty"`
}

// StreamEventType defines types of stream events.
type StreamEventType string

const (
	// EventTypeStart indicates streaming started.
	EventTypeStart StreamEventType = "start"
	// EventTypeChunk indicates a chunk was received.
	EventTypeChunk StreamEventType = "chunk"
	// EventTypeProgress indicates progress was updated.
	EventTypeProgress StreamEventType = "progress"
	// EventTypeComplete indicates streaming completed.
	EventTypeComplete StreamEventType = "complete"
	// EventTypeError indicates an error occurred.
	EventTypeError StreamEventType = "error"
	// EventTypePause indicates streaming was paused.
	EventTypePause StreamEventType = "pause"
	// EventTypeResume indicates streaming was resumed.
	EventTypeResume StreamEventType = "resume"
	// EventTypeCancel indicates streaming was cancelled.
	EventTypeCancel StreamEventType = "cancel"
)

// StreamOptions configures stream behavior.
type StreamOptions struct {
	// BufferType determines how content is buffered.
	BufferType BufferType `json:"buffer_type"`
	// BufferSize is the buffer size for buffered types.
	BufferSize int `json:"buffer_size,omitempty"`
	// RateLimitTokensPerSecond limits output rate.
	RateLimitTokensPerSecond float64 `json:"rate_limit_tokens_per_second,omitempty"`
	// EnableProgress enables progress tracking.
	EnableProgress bool `json:"enable_progress"`
	// ProgressInterval is how often to emit progress.
	ProgressInterval time.Duration `json:"progress_interval,omitempty"`
	// EnableMetrics enables metrics collection.
	EnableMetrics bool `json:"enable_metrics"`
	// Timeout is the maximum stream duration.
	Timeout time.Duration `json:"timeout,omitempty"`
	// RetryOnError retries on transient errors.
	RetryOnError bool `json:"retry_on_error"`
	// MaxRetries is the maximum retry attempts.
	MaxRetries int `json:"max_retries,omitempty"`
}

// DefaultStreamOptions returns default stream options.
func DefaultStreamOptions() *StreamOptions {
	return &StreamOptions{
		BufferType:       BufferTypeWord,
		EnableProgress:   true,
		ProgressInterval: 100 * time.Millisecond,
		EnableMetrics:    true,
		Timeout:          5 * time.Minute,
		RetryOnError:     false,
		MaxRetries:       3,
	}
}

// StreamMetrics contains metrics for a stream.
type StreamMetrics struct {
	// TotalChunks is the total chunks received.
	TotalChunks int64 `json:"total_chunks"`
	// TotalTokens is the total tokens received.
	TotalTokens int64 `json:"total_tokens"`
	// TotalCharacters is the total characters received.
	TotalCharacters int64 `json:"total_characters"`
	// BytesReceived is the total bytes received.
	BytesReceived int64 `json:"bytes_received"`
	// DurationMs is the total duration in milliseconds.
	DurationMs int64 `json:"duration_ms"`
	// FirstChunkLatencyMs is the time to first chunk.
	FirstChunkLatencyMs int64 `json:"first_chunk_latency_ms"`
	// AverageChunkLatencyMs is the average inter-chunk latency.
	AverageChunkLatencyMs float64 `json:"average_chunk_latency_ms"`
	// TokensPerSecond is the throughput in tokens/second.
	TokensPerSecond float64 `json:"tokens_per_second"`
	// ChunksPerSecond is the throughput in chunks/second.
	ChunksPerSecond float64 `json:"chunks_per_second"`
	// ErrorCount is the number of errors encountered.
	ErrorCount int64 `json:"error_count"`
	// RetryCount is the number of retries.
	RetryCount int64 `json:"retry_count"`
}

// CalculateThroughput calculates throughput metrics.
func (m *StreamMetrics) CalculateThroughput() {
	if m.DurationMs > 0 {
		seconds := float64(m.DurationMs) / 1000.0
		m.TokensPerSecond = float64(m.TotalTokens) / seconds
		m.ChunksPerSecond = float64(m.TotalChunks) / seconds
	}
}

// ChunkInfo contains information about a chunk.
type ChunkInfo struct {
	// Index is the chunk index.
	Index int `json:"index"`
	// Content is the chunk content.
	Content string `json:"content"`
	// TokenCount is the number of tokens in this chunk.
	TokenCount int `json:"token_count"`
	// CharacterCount is the number of characters.
	CharacterCount int `json:"character_count"`
	// ByteCount is the number of bytes.
	ByteCount int `json:"byte_count"`
	// Timestamp is when the chunk was received.
	Timestamp time.Time `json:"timestamp"`
	// LatencyMs is the latency from previous chunk.
	LatencyMs int64 `json:"latency_ms,omitempty"`
}

// StreamResult contains the result of a completed stream.
type StreamResult struct {
	// Success indicates if streaming completed successfully.
	Success bool `json:"success"`
	// FullContent is the complete streamed content.
	FullContent string `json:"full_content"`
	// Chunks are the individual chunks.
	Chunks []ChunkInfo `json:"chunks"`
	// Metrics contains streaming metrics.
	Metrics *StreamMetrics `json:"metrics"`
	// Metadata contains stream metadata.
	Metadata *StreamMetadata `json:"metadata,omitempty"`
	// Error contains error information if failed.
	Error string `json:"error,omitempty"`
}

// StreamHandler handles stream events.
type StreamHandler interface {
	// OnStart is called when streaming starts.
	OnStart(metadata *StreamMetadata)
	// OnChunk is called for each chunk.
	OnChunk(chunk *ChunkInfo)
	// OnProgress is called for progress updates.
	OnProgress(progress *StreamProgress)
	// OnComplete is called when streaming completes.
	OnComplete(result *StreamResult)
	// OnError is called when an error occurs.
	OnError(err error)
}

// BaseStreamHandler provides a base implementation of StreamHandler.
type BaseStreamHandler struct{}

// OnStart handles stream start.
func (h *BaseStreamHandler) OnStart(metadata *StreamMetadata) {}

// OnChunk handles chunks.
func (h *BaseStreamHandler) OnChunk(chunk *ChunkInfo) {}

// OnProgress handles progress updates.
func (h *BaseStreamHandler) OnProgress(progress *StreamProgress) {}

// OnComplete handles stream completion.
func (h *BaseStreamHandler) OnComplete(result *StreamResult) {}

// OnError handles errors.
func (h *BaseStreamHandler) OnError(err error) {}

// StreamTransformer transforms stream content.
type StreamTransformer interface {
	// Transform transforms a chunk.
	Transform(chunk string) string
	// TransformChunk transforms a StreamChunk.
	TransformChunk(chunk *StreamChunk) *StreamChunk
}

// IdentityTransformer returns content unchanged.
type IdentityTransformer struct{}

// Transform returns the chunk unchanged.
func (t *IdentityTransformer) Transform(chunk string) string {
	return chunk
}

// TransformChunk returns the chunk unchanged.
func (t *IdentityTransformer) TransformChunk(chunk *StreamChunk) *StreamChunk {
	return chunk
}

// TrimTransformer trims whitespace from chunks.
type TrimTransformer struct{}

// Transform trims whitespace.
func (t *TrimTransformer) Transform(chunk string) string {
	return chunk // Don't trim individual chunks as it may affect word boundaries
}

// TransformChunk trims chunk content.
func (t *TrimTransformer) TransformChunk(chunk *StreamChunk) *StreamChunk {
	return chunk
}

// FilterTransformer filters out certain patterns.
type FilterTransformer struct {
	// FilterFunc returns true if the chunk should be filtered out.
	FilterFunc func(string) bool
}

// Transform filters the chunk.
func (t *FilterTransformer) Transform(chunk string) string {
	if t.FilterFunc != nil && t.FilterFunc(chunk) {
		return ""
	}
	return chunk
}

// TransformChunk filters the chunk.
func (t *FilterTransformer) TransformChunk(chunk *StreamChunk) *StreamChunk {
	if t.FilterFunc != nil && t.FilterFunc(chunk.Content) {
		chunk.Content = ""
	}
	return chunk
}

// StreamPipeline chains multiple transformers.
type StreamPipeline struct {
	transformers []StreamTransformer
}

// NewStreamPipeline creates a new stream pipeline.
func NewStreamPipeline(transformers ...StreamTransformer) *StreamPipeline {
	return &StreamPipeline{transformers: transformers}
}

// Transform applies all transformers in sequence.
func (p *StreamPipeline) Transform(chunk string) string {
	for _, t := range p.transformers {
		chunk = t.Transform(chunk)
	}
	return chunk
}

// TransformChunk applies all transformers to a chunk.
func (p *StreamPipeline) TransformChunk(chunk *StreamChunk) *StreamChunk {
	for _, t := range p.transformers {
		chunk = t.TransformChunk(chunk)
	}
	return chunk
}

// BackpressureStrategy defines how to handle backpressure.
type BackpressureStrategy string

const (
	// BackpressureBlock blocks until consumer catches up.
	BackpressureBlock BackpressureStrategy = "block"
	// BackpressureBuffer buffers excess chunks.
	BackpressureBuffer BackpressureStrategy = "buffer"
	// BackpressureDrop drops excess chunks.
	BackpressureDrop BackpressureStrategy = "drop"
	// BackpressureSample samples chunks when backpressured.
	BackpressureSample BackpressureStrategy = "sample"
)

// BackpressureConfig configures backpressure handling.
type BackpressureConfig struct {
	// Strategy is the backpressure strategy.
	Strategy BackpressureStrategy `json:"strategy"`
	// BufferSize is the buffer size for buffer strategy.
	BufferSize int `json:"buffer_size,omitempty"`
	// SampleRate is the sample rate for sample strategy.
	SampleRate float64 `json:"sample_rate,omitempty"`
}

// DefaultBackpressureConfig returns default backpressure config.
func DefaultBackpressureConfig() *BackpressureConfig {
	return &BackpressureConfig{
		Strategy:   BackpressureBlock,
		BufferSize: 100,
		SampleRate: 0.5,
	}
}

// StreamError represents a streaming error.
type StreamError struct {
	// Code is the error code.
	Code StreamErrorCode `json:"code"`
	// Message is the error message.
	Message string `json:"message"`
	// Recoverable indicates if the error can be recovered from.
	Recoverable bool `json:"recoverable"`
	// ChunkIndex is the chunk index where error occurred.
	ChunkIndex int `json:"chunk_index,omitempty"`
	// Timestamp is when the error occurred.
	Timestamp time.Time `json:"timestamp"`
}

// Error implements the error interface.
func (e *StreamError) Error() string {
	return e.Message
}

// StreamErrorCode defines stream error codes.
type StreamErrorCode string

const (
	// ErrorCodeTimeout indicates a timeout.
	ErrorCodeTimeout StreamErrorCode = "timeout"
	// ErrorCodeCancelled indicates cancellation.
	ErrorCodeCancelled StreamErrorCode = "cancelled"
	// ErrorCodeConnection indicates a connection error.
	ErrorCodeConnection StreamErrorCode = "connection"
	// ErrorCodeRateLimit indicates rate limiting.
	ErrorCodeRateLimit StreamErrorCode = "rate_limit"
	// ErrorCodeBackend indicates a backend error.
	ErrorCodeBackend StreamErrorCode = "backend"
	// ErrorCodeInvalid indicates invalid data.
	ErrorCodeInvalid StreamErrorCode = "invalid"
)

// ReconnectionStrategy defines how to handle reconnection.
type ReconnectionStrategy struct {
	// Enabled indicates if reconnection is enabled.
	Enabled bool `json:"enabled"`
	// MaxAttempts is the maximum reconnection attempts.
	MaxAttempts int `json:"max_attempts"`
	// InitialDelay is the initial delay before reconnecting.
	InitialDelay time.Duration `json:"initial_delay"`
	// MaxDelay is the maximum delay between attempts.
	MaxDelay time.Duration `json:"max_delay"`
	// BackoffMultiplier multiplies delay on each attempt.
	BackoffMultiplier float64 `json:"backoff_multiplier"`
}

// DefaultReconnectionStrategy returns default reconnection strategy.
func DefaultReconnectionStrategy() *ReconnectionStrategy {
	return &ReconnectionStrategy{
		Enabled:           true,
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// StreamCapabilities describes streaming capabilities.
type StreamCapabilities struct {
	// SupportsProgress indicates progress tracking is supported.
	SupportsProgress bool `json:"supports_progress"`
	// SupportsRateLimit indicates rate limiting is supported.
	SupportsRateLimit bool `json:"supports_rate_limit"`
	// SupportsBuffering indicates buffering is supported.
	SupportsBuffering bool `json:"supports_buffering"`
	// SupportsPause indicates pause/resume is supported.
	SupportsPause bool `json:"supports_pause"`
	// SupportsReconnect indicates reconnection is supported.
	SupportsReconnect bool `json:"supports_reconnect"`
	// MaxConcurrentStreams is the maximum concurrent streams.
	MaxConcurrentStreams int `json:"max_concurrent_streams"`
	// SupportedBufferTypes are the supported buffer types.
	SupportedBufferTypes []BufferType `json:"supported_buffer_types"`
}

// DefaultStreamCapabilities returns default streaming capabilities.
func DefaultStreamCapabilities() *StreamCapabilities {
	return &StreamCapabilities{
		SupportsProgress:     true,
		SupportsRateLimit:    true,
		SupportsBuffering:    true,
		SupportsPause:        false,
		SupportsReconnect:    true,
		MaxConcurrentStreams: 10,
		SupportedBufferTypes: []BufferType{
			BufferTypeCharacter,
			BufferTypeWord,
			BufferTypeSentence,
			BufferTypeLine,
			BufferTypeParagraph,
			BufferTypeToken,
		},
	}
}
