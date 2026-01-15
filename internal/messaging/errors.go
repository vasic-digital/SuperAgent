package messaging

import (
	"errors"
	"fmt"
)

// ErrorCode represents a messaging error code.
type ErrorCode string

const (
	// Connection errors
	ErrCodeConnectionFailed    ErrorCode = "CONNECTION_FAILED"
	ErrCodeConnectionClosed    ErrorCode = "CONNECTION_CLOSED"
	ErrCodeConnectionTimeout   ErrorCode = "CONNECTION_TIMEOUT"
	ErrCodeReconnectionFailed  ErrorCode = "RECONNECTION_FAILED"
	ErrCodeAuthenticationFailed ErrorCode = "AUTHENTICATION_FAILED"

	// Publish errors
	ErrCodePublishFailed       ErrorCode = "PUBLISH_FAILED"
	ErrCodePublishTimeout      ErrorCode = "PUBLISH_TIMEOUT"
	ErrCodePublishRejected     ErrorCode = "PUBLISH_REJECTED"
	ErrCodeConfirmTimeout      ErrorCode = "CONFIRM_TIMEOUT"
	ErrCodeSerializationFailed ErrorCode = "SERIALIZATION_FAILED"

	// Subscribe errors
	ErrCodeSubscribeFailed ErrorCode = "SUBSCRIBE_FAILED"
	ErrCodeConsumerCanceled ErrorCode = "CONSUMER_CANCELED"
	ErrCodeHandlerError    ErrorCode = "HANDLER_ERROR"

	// Queue/Topic errors
	ErrCodeQueueNotFound       ErrorCode = "QUEUE_NOT_FOUND"
	ErrCodeQueueDeclareFailed  ErrorCode = "QUEUE_DECLARE_FAILED"
	ErrCodeTopicNotFound       ErrorCode = "TOPIC_NOT_FOUND"
	ErrCodeTopicCreateFailed   ErrorCode = "TOPIC_CREATE_FAILED"
	ErrCodeExchangeNotFound    ErrorCode = "EXCHANGE_NOT_FOUND"

	// Message errors
	ErrCodeMessageTooLarge    ErrorCode = "MESSAGE_TOO_LARGE"
	ErrCodeMessageExpired     ErrorCode = "MESSAGE_EXPIRED"
	ErrCodeMessageInvalid     ErrorCode = "MESSAGE_INVALID"
	ErrCodeAckFailed          ErrorCode = "ACK_FAILED"
	ErrCodeNackFailed         ErrorCode = "NACK_FAILED"
	ErrCodeDeadLetterFailed   ErrorCode = "DEAD_LETTER_FAILED"

	// Configuration errors
	ErrCodeInvalidConfig ErrorCode = "INVALID_CONFIG"
	ErrCodeMissingConfig ErrorCode = "MISSING_CONFIG"

	// General errors
	ErrCodeBrokerUnavailable ErrorCode = "BROKER_UNAVAILABLE"
	ErrCodeOperationCanceled ErrorCode = "OPERATION_CANCELED"
	ErrCodeUnknown           ErrorCode = "UNKNOWN_ERROR"
)

// Common sentinel errors for easy comparison.
var (
	// Connection errors
	ErrConnectionFailed     = errors.New("connection failed")
	ErrConnectionClosed     = errors.New("connection closed")
	ErrConnectionTimeout    = errors.New("connection timeout")
	ErrReconnectionFailed   = errors.New("reconnection failed")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrNotConnected         = errors.New("not connected to broker")

	// Publish errors
	ErrPublishFailed       = errors.New("publish failed")
	ErrPublishTimeout      = errors.New("publish timeout")
	ErrPublishRejected     = errors.New("publish rejected")
	ErrConfirmTimeout      = errors.New("publisher confirm timeout")
	ErrSerializationFailed = errors.New("serialization failed")

	// Subscribe errors
	ErrSubscribeFailed  = errors.New("subscribe failed")
	ErrConsumerCanceled = errors.New("consumer canceled")
	ErrHandlerError     = errors.New("message handler error")

	// Queue/Topic errors
	ErrQueueNotFound       = errors.New("queue not found")
	ErrQueueDeclareFailed  = errors.New("queue declaration failed")
	ErrTopicNotFound       = errors.New("topic not found")
	ErrTopicCreateFailed   = errors.New("topic creation failed")
	ErrExchangeNotFound    = errors.New("exchange not found")

	// Message errors
	ErrMessageTooLarge  = errors.New("message too large")
	ErrMessageExpired   = errors.New("message expired")
	ErrMessageInvalid   = errors.New("invalid message")
	ErrAckFailed        = errors.New("acknowledgment failed")
	ErrNackFailed       = errors.New("negative acknowledgment failed")
	ErrDeadLetterFailed = errors.New("dead letter failed")

	// Configuration errors
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrMissingConfig = errors.New("missing configuration")

	// General errors
	ErrBrokerUnavailable = errors.New("broker unavailable")
	ErrOperationCanceled = errors.New("operation canceled")
)

// BrokerError represents a messaging broker error with detailed information.
type BrokerError struct {
	// Code is the error code.
	Code ErrorCode `json:"code"`
	// Message is the human-readable error message.
	Message string `json:"message"`
	// Cause is the underlying error.
	Cause error `json:"-"`
	// BrokerType is the type of broker that produced the error.
	BrokerType BrokerType `json:"broker_type,omitempty"`
	// Topic is the topic/queue involved (if applicable).
	Topic string `json:"topic,omitempty"`
	// MessageID is the message ID involved (if applicable).
	MessageID string `json:"message_id,omitempty"`
	// Retryable indicates if the operation can be retried.
	Retryable bool `json:"retryable"`
	// Details contains additional error details.
	Details map[string]interface{} `json:"details,omitempty"`
}

// NewBrokerError creates a new BrokerError.
func NewBrokerError(code ErrorCode, message string, cause error) *BrokerError {
	return &BrokerError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Retryable: isRetryable(code),
	}
}

// Error implements the error interface.
func (e *BrokerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *BrokerError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target.
func (e *BrokerError) Is(target error) bool {
	if t, ok := target.(*BrokerError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Cause, target)
}

// WithBrokerType sets the broker type.
func (e *BrokerError) WithBrokerType(bt BrokerType) *BrokerError {
	e.BrokerType = bt
	return e
}

// WithTopic sets the topic/queue name.
func (e *BrokerError) WithTopic(topic string) *BrokerError {
	e.Topic = topic
	return e
}

// WithMessageID sets the message ID.
func (e *BrokerError) WithMessageID(id string) *BrokerError {
	e.MessageID = id
	return e
}

// WithDetail adds a detail to the error.
func (e *BrokerError) WithDetail(key string, value interface{}) *BrokerError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details to the error.
func (e *BrokerError) WithDetails(details map[string]interface{}) *BrokerError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// isRetryable determines if an error code represents a retryable error.
func isRetryable(code ErrorCode) bool {
	switch code {
	case ErrCodeConnectionFailed,
		ErrCodeConnectionTimeout,
		ErrCodeReconnectionFailed,
		ErrCodePublishTimeout,
		ErrCodeConfirmTimeout,
		ErrCodeBrokerUnavailable:
		return true
	default:
		return false
	}
}

// ConnectionError creates a connection error.
func ConnectionError(message string, cause error) *BrokerError {
	return NewBrokerError(ErrCodeConnectionFailed, message, cause)
}

// ConnectionTimeoutError creates a connection timeout error.
func ConnectionTimeoutError(cause error) *BrokerError {
	return NewBrokerError(ErrCodeConnectionTimeout, "connection timeout", cause)
}

// AuthenticationError creates an authentication error.
func AuthenticationError(message string, cause error) *BrokerError {
	return NewBrokerError(ErrCodeAuthenticationFailed, message, cause)
}

// PublishError creates a publish error.
func PublishError(topic string, cause error) *BrokerError {
	return NewBrokerError(ErrCodePublishFailed, "failed to publish message", cause).
		WithTopic(topic)
}

// PublishTimeoutError creates a publish timeout error.
func PublishTimeoutError(topic string) *BrokerError {
	return NewBrokerError(ErrCodePublishTimeout, "publish timeout", nil).
		WithTopic(topic)
}

// SubscribeError creates a subscribe error.
func SubscribeError(topic string, cause error) *BrokerError {
	return NewBrokerError(ErrCodeSubscribeFailed, "failed to subscribe", cause).
		WithTopic(topic)
}

// HandlerError creates a handler error.
func HandlerError(messageID string, cause error) *BrokerError {
	return NewBrokerError(ErrCodeHandlerError, "message handler failed", cause).
		WithMessageID(messageID)
}

// QueueError creates a queue error.
func QueueError(queue string, cause error) *BrokerError {
	return NewBrokerError(ErrCodeQueueDeclareFailed, "queue operation failed", cause).
		WithTopic(queue)
}

// TopicError creates a topic error.
func TopicError(topic string, cause error) *BrokerError {
	return NewBrokerError(ErrCodeTopicCreateFailed, "topic operation failed", cause).
		WithTopic(topic)
}

// MessageError creates a message error.
func MessageError(code ErrorCode, messageID string, cause error) *BrokerError {
	return NewBrokerError(code, "message operation failed", cause).
		WithMessageID(messageID)
}

// SerializationError creates a serialization error.
func SerializationError(cause error) *BrokerError {
	return NewBrokerError(ErrCodeSerializationFailed, "serialization failed", cause)
}

// ConfigError creates a configuration error.
func ConfigError(message string) *BrokerError {
	return NewBrokerError(ErrCodeInvalidConfig, message, nil)
}

// IsBrokerError checks if an error is a BrokerError.
func IsBrokerError(err error) bool {
	var brokerErr *BrokerError
	return errors.As(err, &brokerErr)
}

// GetBrokerError extracts a BrokerError from an error chain.
func GetBrokerError(err error) *BrokerError {
	var brokerErr *BrokerError
	if errors.As(err, &brokerErr) {
		return brokerErr
	}
	return nil
}

// IsRetryableError checks if an error is retryable.
func IsRetryableError(err error) bool {
	if brokerErr := GetBrokerError(err); brokerErr != nil {
		return brokerErr.Retryable
	}
	return false
}

// IsConnectionError checks if an error is a connection error.
func IsConnectionError(err error) bool {
	if brokerErr := GetBrokerError(err); brokerErr != nil {
		switch brokerErr.Code {
		case ErrCodeConnectionFailed,
			ErrCodeConnectionClosed,
			ErrCodeConnectionTimeout,
			ErrCodeReconnectionFailed,
			ErrCodeAuthenticationFailed:
			return true
		}
	}
	return errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrConnectionClosed) ||
		errors.Is(err, ErrConnectionTimeout)
}

// IsPublishError checks if an error is a publish error.
func IsPublishError(err error) bool {
	if brokerErr := GetBrokerError(err); brokerErr != nil {
		switch brokerErr.Code {
		case ErrCodePublishFailed,
			ErrCodePublishTimeout,
			ErrCodePublishRejected,
			ErrCodeConfirmTimeout:
			return true
		}
	}
	return errors.Is(err, ErrPublishFailed) ||
		errors.Is(err, ErrPublishTimeout) ||
		errors.Is(err, ErrPublishRejected)
}

// IsSubscribeError checks if an error is a subscribe error.
func IsSubscribeError(err error) bool {
	if brokerErr := GetBrokerError(err); brokerErr != nil {
		switch brokerErr.Code {
		case ErrCodeSubscribeFailed,
			ErrCodeConsumerCanceled,
			ErrCodeHandlerError:
			return true
		}
	}
	return errors.Is(err, ErrSubscribeFailed) ||
		errors.Is(err, ErrConsumerCanceled) ||
		errors.Is(err, ErrHandlerError)
}

// WrapError wraps an error with broker context.
func WrapError(err error, code ErrorCode, message string) *BrokerError {
	return NewBrokerError(code, message, err)
}

// MultiError represents multiple errors.
type MultiError struct {
	Errors []error
}

// NewMultiError creates a new MultiError.
func NewMultiError(errs ...error) *MultiError {
	return &MultiError{Errors: errs}
}

// Error implements the error interface.
func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors (%d): %v", len(e.Errors), e.Errors[0])
}

// Add adds an error to the MultiError.
func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if there are any errors.
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// ErrorOrNil returns nil if there are no errors.
func (e *MultiError) ErrorOrNil() error {
	if e.HasErrors() {
		return e
	}
	return nil
}

// Unwrap returns the first error (for errors.Is/errors.As compatibility).
func (e *MultiError) Unwrap() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}
