// Package dlq provides dead letter queue processing for the messaging system.
package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"dev.helix.agent/internal/messaging"
	"go.uber.org/zap"
)

// ProcessorConfig configures the DLQ processor
type ProcessorConfig struct {
	// MaxRetries is the maximum number of times to retry a message
	MaxRetries int
	// RetryDelay is the initial delay between retries
	RetryDelay time.Duration
	// RetryBackoffMultiplier multiplies the delay on each retry
	RetryBackoffMultiplier float64
	// MaxRetryDelay caps the maximum retry delay
	MaxRetryDelay time.Duration
	// ProcessingTimeout is the timeout for processing a single message
	ProcessingTimeout time.Duration
	// BatchSize is the number of messages to process in a batch
	BatchSize int
	// PollInterval is how often to poll the DLQ for messages
	PollInterval time.Duration
}

// DefaultConfig returns sensible defaults for DLQ processing
func DefaultConfig() ProcessorConfig {
	return ProcessorConfig{
		MaxRetries:             3,
		RetryDelay:             1 * time.Second,
		RetryBackoffMultiplier: 2.0,
		MaxRetryDelay:          30 * time.Second,
		ProcessingTimeout:      30 * time.Second,
		BatchSize:              10,
		PollInterval:           5 * time.Second,
	}
}

// DeadLetterMessage represents a message in the dead letter queue
type DeadLetterMessage struct {
	ID              string                 `json:"id"`
	OriginalQueue   string                 `json:"original_queue"`
	OriginalTopic   string                 `json:"original_topic"`
	OriginalMessage *messaging.Message     `json:"original_message"`
	FailureReason   string                 `json:"failure_reason"`
	FailureDetails  map[string]interface{} `json:"failure_details,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	FirstFailure    time.Time              `json:"first_failure"`
	LastFailure     time.Time              `json:"last_failure"`
	NextRetry       time.Time              `json:"next_retry,omitempty"`
	Status          DLQStatus              `json:"status"`
}

// DLQStatus represents the status of a DLQ message
type DLQStatus string

const (
	StatusPending   DLQStatus = "pending"
	StatusRetrying  DLQStatus = "retrying"
	StatusProcessed DLQStatus = "processed"
	StatusDiscarded DLQStatus = "discarded"
	StatusExpired   DLQStatus = "expired"
)

// ProcessorMetrics tracks DLQ processor metrics
type ProcessorMetrics struct {
	MessagesProcessed  int64
	MessagesRetried    int64
	MessagesDiscarded  int64
	MessagesExpired    int64
	ProcessingErrors   int64
	CurrentQueueDepth  int64
	AverageProcessTime time.Duration
}

// Processor handles dead letter queue processing
type Processor struct {
	config  ProcessorConfig
	broker  messaging.MessageBroker
	logger  *zap.Logger
	metrics ProcessorMetrics

	handlers map[string]RetryHandler
	mu       sync.RWMutex

	// In-memory store for DLQ messages (for listing/reprocessing/discarding)
	messages   map[string]*DeadLetterMessage
	messagesMu sync.RWMutex

	running atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// RetryHandler handles retry logic for specific message types
type RetryHandler func(ctx context.Context, msg *DeadLetterMessage) error

// NewProcessor creates a new DLQ processor
func NewProcessor(broker messaging.MessageBroker, config ProcessorConfig, logger *zap.Logger) *Processor {
	return &Processor{
		config:   config,
		broker:   broker,
		logger:   logger,
		handlers: make(map[string]RetryHandler),
		messages: make(map[string]*DeadLetterMessage),
	}
}

// RegisterHandler registers a retry handler for a specific message type
func (p *Processor) RegisterHandler(messageType string, handler RetryHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[messageType] = handler
}

// Start begins processing the dead letter queue
func (p *Processor) Start(ctx context.Context) error {
	if p.running.Load() {
		return fmt.Errorf("processor already running")
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.running.Store(true)

	p.wg.Add(1)
	go p.processLoop()

	p.logger.Info("DLQ processor started",
		zap.Int("max_retries", p.config.MaxRetries),
		zap.Duration("poll_interval", p.config.PollInterval),
		zap.Int("batch_size", p.config.BatchSize))

	return nil
}

// Stop gracefully stops the processor
func (p *Processor) Stop() error {
	if !p.running.Load() {
		return nil
	}

	p.cancel()
	p.wg.Wait()
	p.running.Store(false)

	p.logger.Info("DLQ processor stopped",
		zap.Int64("messages_processed", atomic.LoadInt64(&p.metrics.MessagesProcessed)),
		zap.Int64("messages_retried", atomic.LoadInt64(&p.metrics.MessagesRetried)))

	return nil
}

// processLoop is the main processing loop
func (p *Processor) processLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.processBatch()
		}
	}
}

// processBatch processes a batch of DLQ messages
func (p *Processor) processBatch() {
	// Subscribe to DLQ and process messages
	ctx, cancel := context.WithTimeout(p.ctx, p.config.ProcessingTimeout*time.Duration(p.config.BatchSize))
	defer cancel()

	processed := 0
	messagesCh := make(chan *messaging.Message, p.config.BatchSize)

	// Create a temporary subscription to fetch messages
	sub, err := p.broker.Subscribe(ctx, "helixagent.dlq", func(ctx context.Context, msg *messaging.Message) error {
		select {
		case messagesCh <- msg:
		default:
			// Channel full, skip this message
		}
		return nil
	})
	if err != nil {
		p.logger.Error("Failed to subscribe to DLQ", zap.Error(err))
		return
	}
	defer func() { _ = sub.Unsubscribe() }()

	// Wait for messages or timeout
	timeout := time.After(p.config.PollInterval / 2)

	for processed < p.config.BatchSize {
		select {
		case msg := <-messagesCh:
			if err := p.processMessage(ctx, msg); err != nil {
				p.logger.Error("Failed to process DLQ message",
					zap.String("message_id", msg.ID),
					zap.Error(err))
				atomic.AddInt64(&p.metrics.ProcessingErrors, 1)
			} else {
				processed++
			}
		case <-timeout:
			return
		case <-ctx.Done():
			return
		}
	}
}

// processMessage processes a single DLQ message
func (p *Processor) processMessage(ctx context.Context, msg *messaging.Message) error {
	startTime := time.Now()

	// Parse DLQ message
	var dlqMsg DeadLetterMessage
	if err := json.Unmarshal(msg.Payload, &dlqMsg); err != nil {
		// If not a proper DLQ message, just log and discard
		p.logger.Warn("Invalid DLQ message format", zap.String("message_id", msg.ID))
		atomic.AddInt64(&p.metrics.MessagesDiscarded, 1)
		return nil
	}

	// Check retry count
	if dlqMsg.RetryCount >= p.config.MaxRetries {
		p.logger.Warn("Message exceeded max retries, discarding",
			zap.String("message_id", dlqMsg.ID),
			zap.Int("retry_count", dlqMsg.RetryCount),
			zap.String("failure_reason", dlqMsg.FailureReason))
		atomic.AddInt64(&p.metrics.MessagesDiscarded, 1)
		dlqMsg.Status = StatusDiscarded
		return p.updateDLQMessage(ctx, &dlqMsg)
	}

	// Check if retry is due
	if time.Now().Before(dlqMsg.NextRetry) {
		return nil // Not ready for retry yet
	}

	// Find handler for message type
	p.mu.RLock()
	handler, ok := p.handlers[dlqMsg.OriginalMessage.Type]
	p.mu.RUnlock()

	if !ok {
		// Use default retry handler
		handler = p.defaultRetryHandler
	}

	// Update status to retrying
	dlqMsg.Status = StatusRetrying
	dlqMsg.RetryCount++

	// Execute retry
	retryCtx, retryCancel := context.WithTimeout(ctx, p.config.ProcessingTimeout)
	defer retryCancel()

	if err := handler(retryCtx, &dlqMsg); err != nil {
		// Retry failed, schedule next retry
		delay := p.calculateRetryDelay(dlqMsg.RetryCount)
		dlqMsg.NextRetry = time.Now().Add(delay)
		dlqMsg.LastFailure = time.Now()
		dlqMsg.Status = StatusPending
		if dlqMsg.FailureDetails == nil {
			dlqMsg.FailureDetails = make(map[string]interface{})
		}
		dlqMsg.FailureDetails["last_error"] = err.Error()

		p.logger.Warn("Retry failed, scheduling next attempt",
			zap.String("message_id", dlqMsg.ID),
			zap.Int("retry_count", dlqMsg.RetryCount),
			zap.Duration("next_retry_in", delay),
			zap.Error(err))

		atomic.AddInt64(&p.metrics.MessagesRetried, 1)
		return p.updateDLQMessage(ctx, &dlqMsg)
	}

	// Retry succeeded
	dlqMsg.Status = StatusProcessed
	atomic.AddInt64(&p.metrics.MessagesProcessed, 1)

	p.logger.Info("Successfully reprocessed DLQ message",
		zap.String("message_id", dlqMsg.ID),
		zap.Int("retry_count", dlqMsg.RetryCount),
		zap.Duration("processing_time", time.Since(startTime)))

	return p.updateDLQMessage(ctx, &dlqMsg)
}

// defaultRetryHandler republishes the original message to its original queue
func (p *Processor) defaultRetryHandler(ctx context.Context, dlqMsg *DeadLetterMessage) error {
	// Republish to original topic/queue
	if dlqMsg.OriginalTopic != "" {
		return p.broker.Publish(ctx, dlqMsg.OriginalTopic, dlqMsg.OriginalMessage)
	}
	return fmt.Errorf("no original topic specified for message %s", dlqMsg.ID)
}

// calculateRetryDelay calculates the delay before next retry using exponential backoff
func (p *Processor) calculateRetryDelay(retryCount int) time.Duration {
	delay := p.config.RetryDelay
	for i := 1; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * p.config.RetryBackoffMultiplier)
		if delay > p.config.MaxRetryDelay {
			delay = p.config.MaxRetryDelay
			break
		}
	}
	return delay
}

// updateDLQMessage updates a DLQ message in the in-memory store
func (p *Processor) updateDLQMessage(ctx context.Context, dlqMsg *DeadLetterMessage) error {
	p.messagesMu.Lock()
	defer p.messagesMu.Unlock()

	// Store or update the message
	p.messages[dlqMsg.ID] = dlqMsg

	// Update queue depth metric
	pendingCount := int64(0)
	for _, msg := range p.messages {
		if msg.Status == StatusPending || msg.Status == StatusRetrying {
			pendingCount++
		}
	}
	atomic.StoreInt64(&p.metrics.CurrentQueueDepth, pendingCount)

	p.logger.Debug("DLQ message updated",
		zap.String("message_id", dlqMsg.ID),
		zap.String("status", string(dlqMsg.Status)),
		zap.Int("retry_count", dlqMsg.RetryCount))
	return nil
}

// GetMetrics returns current processor metrics
func (p *Processor) GetMetrics() ProcessorMetrics {
	return ProcessorMetrics{
		MessagesProcessed: atomic.LoadInt64(&p.metrics.MessagesProcessed),
		MessagesRetried:   atomic.LoadInt64(&p.metrics.MessagesRetried),
		MessagesDiscarded: atomic.LoadInt64(&p.metrics.MessagesDiscarded),
		MessagesExpired:   atomic.LoadInt64(&p.metrics.MessagesExpired),
		ProcessingErrors:  atomic.LoadInt64(&p.metrics.ProcessingErrors),
		CurrentQueueDepth: atomic.LoadInt64(&p.metrics.CurrentQueueDepth),
	}
}

// ReprocessMessage manually triggers reprocessing of a specific DLQ message
func (p *Processor) ReprocessMessage(ctx context.Context, messageID string) error {
	p.logger.Info("Manual reprocess requested", zap.String("message_id", messageID))

	// Fetch the message from the store
	p.messagesMu.RLock()
	dlqMsg, exists := p.messages[messageID]
	p.messagesMu.RUnlock()

	if !exists {
		return fmt.Errorf("message not found: %s", messageID)
	}

	// Check if message can be reprocessed
	if dlqMsg.Status == StatusProcessed || dlqMsg.Status == StatusDiscarded {
		return fmt.Errorf("message %s is already %s and cannot be reprocessed", messageID, dlqMsg.Status)
	}

	// Reset retry count for manual reprocess and set to retrying
	p.messagesMu.Lock()
	dlqMsg.Status = StatusRetrying
	dlqMsg.NextRetry = time.Time{} // Clear next retry to process immediately
	p.messagesMu.Unlock()

	// Find handler for message type
	p.mu.RLock()
	handler, ok := p.handlers[dlqMsg.OriginalMessage.Type]
	p.mu.RUnlock()

	if !ok {
		handler = p.defaultRetryHandler
	}

	// Execute retry with timeout
	retryCtx, retryCancel := context.WithTimeout(ctx, p.config.ProcessingTimeout)
	defer retryCancel()

	if err := handler(retryCtx, dlqMsg); err != nil {
		// Retry failed
		p.messagesMu.Lock()
		dlqMsg.RetryCount++
		dlqMsg.LastFailure = time.Now()
		dlqMsg.Status = StatusPending
		delay := p.calculateRetryDelay(dlqMsg.RetryCount)
		dlqMsg.NextRetry = time.Now().Add(delay)
		if dlqMsg.FailureDetails == nil {
			dlqMsg.FailureDetails = make(map[string]interface{})
		}
		dlqMsg.FailureDetails["last_error"] = err.Error()
		dlqMsg.FailureDetails["manual_reprocess"] = true
		p.messagesMu.Unlock()

		atomic.AddInt64(&p.metrics.MessagesRetried, 1)
		p.logger.Warn("Manual reprocess failed",
			zap.String("message_id", messageID),
			zap.Error(err))
		return p.updateDLQMessage(ctx, dlqMsg)
	}

	// Success
	p.messagesMu.Lock()
	dlqMsg.Status = StatusProcessed
	p.messagesMu.Unlock()

	atomic.AddInt64(&p.metrics.MessagesProcessed, 1)
	p.logger.Info("Manual reprocess succeeded", zap.String("message_id", messageID))
	return p.updateDLQMessage(ctx, dlqMsg)
}

// DiscardMessage permanently discards a DLQ message
func (p *Processor) DiscardMessage(ctx context.Context, messageID string, reason string) error {
	p.logger.Info("Manual discard requested",
		zap.String("message_id", messageID),
		zap.String("reason", reason))

	p.messagesMu.Lock()
	defer p.messagesMu.Unlock()

	dlqMsg, exists := p.messages[messageID]
	if !exists {
		return fmt.Errorf("message not found: %s", messageID)
	}

	// Check if already discarded
	if dlqMsg.Status == StatusDiscarded {
		return fmt.Errorf("message %s is already discarded", messageID)
	}

	// Update status
	dlqMsg.Status = StatusDiscarded
	if dlqMsg.FailureDetails == nil {
		dlqMsg.FailureDetails = make(map[string]interface{})
	}
	dlqMsg.FailureDetails["discard_reason"] = reason
	dlqMsg.FailureDetails["discarded_at"] = time.Now().Format(time.RFC3339)

	atomic.AddInt64(&p.metrics.MessagesDiscarded, 1)

	// Update queue depth
	pendingCount := int64(0)
	for _, msg := range p.messages {
		if msg.Status == StatusPending || msg.Status == StatusRetrying {
			pendingCount++
		}
	}
	atomic.StoreInt64(&p.metrics.CurrentQueueDepth, pendingCount)

	p.logger.Info("Message discarded",
		zap.String("message_id", messageID),
		zap.String("reason", reason))
	return nil
}

// ListMessages returns a list of messages currently in the DLQ
func (p *Processor) ListMessages(ctx context.Context, limit, offset int) ([]*DeadLetterMessage, error) {
	p.messagesMu.RLock()
	defer p.messagesMu.RUnlock()

	// Collect all messages into a slice for sorting
	allMessages := make([]*DeadLetterMessage, 0, len(p.messages))
	for _, msg := range p.messages {
		allMessages = append(allMessages, msg)
	}

	// Sort by last failure time (most recent first)
	for i := 0; i < len(allMessages)-1; i++ {
		for j := i + 1; j < len(allMessages); j++ {
			if allMessages[j].LastFailure.After(allMessages[i].LastFailure) {
				allMessages[i], allMessages[j] = allMessages[j], allMessages[i]
			}
		}
	}

	// Apply offset
	if offset > 0 {
		if offset >= len(allMessages) {
			return []*DeadLetterMessage{}, nil
		}
		allMessages = allMessages[offset:]
	}

	// Apply limit
	if limit > 0 && limit < len(allMessages) {
		allMessages = allMessages[:limit]
	}

	// Return copies to avoid external modifications
	result := make([]*DeadLetterMessage, len(allMessages))
	for i, msg := range allMessages {
		msgCopy := *msg
		result[i] = &msgCopy
	}

	return result, nil
}

// GetMessage retrieves a specific message by ID
func (p *Processor) GetMessage(ctx context.Context, messageID string) (*DeadLetterMessage, error) {
	p.messagesMu.RLock()
	defer p.messagesMu.RUnlock()

	msg, exists := p.messages[messageID]
	if !exists {
		return nil, fmt.Errorf("message not found: %s", messageID)
	}

	// Return a copy
	msgCopy := *msg
	return &msgCopy, nil
}

// AddMessage adds a new message to the DLQ (used when messages fail in other systems)
func (p *Processor) AddMessage(ctx context.Context, dlqMsg *DeadLetterMessage) error {
	if dlqMsg.ID == "" {
		return fmt.Errorf("message ID is required")
	}

	p.messagesMu.Lock()
	defer p.messagesMu.Unlock()

	// Check if message already exists
	if _, exists := p.messages[dlqMsg.ID]; exists {
		return fmt.Errorf("message %s already exists in DLQ", dlqMsg.ID)
	}

	// Initialize timestamps
	now := time.Now()
	if dlqMsg.FirstFailure.IsZero() {
		dlqMsg.FirstFailure = now
	}
	dlqMsg.LastFailure = now

	// Set initial status
	if dlqMsg.Status == "" {
		dlqMsg.Status = StatusPending
	}

	// Calculate next retry time
	if dlqMsg.NextRetry.IsZero() {
		delay := p.calculateRetryDelay(dlqMsg.RetryCount + 1)
		dlqMsg.NextRetry = now.Add(delay)
	}

	p.messages[dlqMsg.ID] = dlqMsg

	// Update queue depth
	atomic.AddInt64(&p.metrics.CurrentQueueDepth, 1)

	p.logger.Info("Message added to DLQ",
		zap.String("message_id", dlqMsg.ID),
		zap.String("original_topic", dlqMsg.OriginalTopic),
		zap.String("failure_reason", dlqMsg.FailureReason))

	return nil
}

// GetMessageCount returns the total number of messages in the DLQ
func (p *Processor) GetMessageCount(ctx context.Context) int {
	p.messagesMu.RLock()
	defer p.messagesMu.RUnlock()
	return len(p.messages)
}

// GetMessagesByStatus returns messages filtered by status
func (p *Processor) GetMessagesByStatus(ctx context.Context, status DLQStatus, limit int) ([]*DeadLetterMessage, error) {
	p.messagesMu.RLock()
	defer p.messagesMu.RUnlock()

	result := make([]*DeadLetterMessage, 0)
	for _, msg := range p.messages {
		if msg.Status == status {
			msgCopy := *msg
			result = append(result, &msgCopy)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// PurgeProcessed removes all processed messages from the store
func (p *Processor) PurgeProcessed(ctx context.Context) (int, error) {
	p.messagesMu.Lock()
	defer p.messagesMu.Unlock()

	count := 0
	for id, msg := range p.messages {
		if msg.Status == StatusProcessed || msg.Status == StatusDiscarded {
			delete(p.messages, id)
			count++
		}
	}

	p.logger.Info("Purged processed messages", zap.Int("count", count))
	return count, nil
}
