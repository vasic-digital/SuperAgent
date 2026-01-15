package messaging

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBrokerError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewBrokerError(ErrCodeConnectionFailed, "connection failed", cause)

	assert.Equal(t, ErrCodeConnectionFailed, err.Code)
	assert.Equal(t, "connection failed", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.True(t, err.Retryable) // Connection errors are retryable
}

func TestBrokerError_Error(t *testing.T) {
	// With cause
	cause := errors.New("underlying error")
	err := NewBrokerError(ErrCodeConnectionFailed, "connection failed", cause)
	assert.Contains(t, err.Error(), "CONNECTION_FAILED")
	assert.Contains(t, err.Error(), "connection failed")
	assert.Contains(t, err.Error(), "underlying error")

	// Without cause
	err2 := NewBrokerError(ErrCodeConnectionFailed, "connection failed", nil)
	assert.Contains(t, err2.Error(), "CONNECTION_FAILED")
	assert.Contains(t, err2.Error(), "connection failed")
}

func TestBrokerError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewBrokerError(ErrCodeConnectionFailed, "connection failed", cause)

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, cause, unwrapped)
}

func TestBrokerError_Is(t *testing.T) {
	err1 := NewBrokerError(ErrCodeConnectionFailed, "error 1", nil)
	err2 := NewBrokerError(ErrCodeConnectionFailed, "error 2", nil)
	err3 := NewBrokerError(ErrCodePublishFailed, "error 3", nil)

	// Same code
	assert.True(t, errors.Is(err1, err2))
	// Different code
	assert.False(t, errors.Is(err1, err3))

	// With underlying error
	cause := ErrConnectionFailed
	err4 := NewBrokerError(ErrCodeConnectionFailed, "error 4", cause)
	assert.True(t, errors.Is(err4, ErrConnectionFailed))
}

func TestBrokerError_WithBrokerType(t *testing.T) {
	err := NewBrokerError(ErrCodeConnectionFailed, "error", nil).
		WithBrokerType(BrokerTypeRabbitMQ)

	assert.Equal(t, BrokerTypeRabbitMQ, err.BrokerType)
}

func TestBrokerError_WithTopic(t *testing.T) {
	err := NewBrokerError(ErrCodePublishFailed, "error", nil).
		WithTopic("test.topic")

	assert.Equal(t, "test.topic", err.Topic)
}

func TestBrokerError_WithMessageID(t *testing.T) {
	err := NewBrokerError(ErrCodeHandlerError, "error", nil).
		WithMessageID("msg-123")

	assert.Equal(t, "msg-123", err.MessageID)
}

func TestBrokerError_WithDetail(t *testing.T) {
	err := NewBrokerError(ErrCodeConnectionFailed, "error", nil).
		WithDetail("key1", "value1").
		WithDetail("key2", 123)

	assert.Equal(t, "value1", err.Details["key1"])
	assert.Equal(t, 123, err.Details["key2"])
}

func TestBrokerError_WithDetails(t *testing.T) {
	details := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	err := NewBrokerError(ErrCodeConnectionFailed, "error", nil).
		WithDetails(details)

	assert.Equal(t, "value1", err.Details["key1"])
	assert.Equal(t, 123, err.Details["key2"])
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected bool
	}{
		{ErrCodeConnectionFailed, true},
		{ErrCodeConnectionTimeout, true},
		{ErrCodeReconnectionFailed, true},
		{ErrCodePublishTimeout, true},
		{ErrCodeConfirmTimeout, true},
		{ErrCodeBrokerUnavailable, true},
		{ErrCodePublishFailed, false},
		{ErrCodeSubscribeFailed, false},
		{ErrCodeMessageInvalid, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := NewBrokerError(tt.code, "error", nil)
			assert.Equal(t, tt.expected, err.Retryable)
		})
	}
}

func TestConnectionError(t *testing.T) {
	cause := errors.New("underlying")
	err := ConnectionError("failed to connect", cause)

	assert.Equal(t, ErrCodeConnectionFailed, err.Code)
	assert.Equal(t, "failed to connect", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestConnectionTimeoutError(t *testing.T) {
	cause := errors.New("timeout")
	err := ConnectionTimeoutError(cause)

	assert.Equal(t, ErrCodeConnectionTimeout, err.Code)
	assert.Equal(t, "connection timeout", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestAuthenticationError(t *testing.T) {
	cause := errors.New("bad credentials")
	err := AuthenticationError("auth failed", cause)

	assert.Equal(t, ErrCodeAuthenticationFailed, err.Code)
	assert.Equal(t, "auth failed", err.Message)
}

func TestPublishError(t *testing.T) {
	cause := errors.New("network error")
	err := PublishError("test.topic", cause)

	assert.Equal(t, ErrCodePublishFailed, err.Code)
	assert.Equal(t, "test.topic", err.Topic)
	assert.Equal(t, cause, err.Cause)
}

func TestPublishTimeoutError(t *testing.T) {
	err := PublishTimeoutError("test.topic")

	assert.Equal(t, ErrCodePublishTimeout, err.Code)
	assert.Equal(t, "test.topic", err.Topic)
}

func TestSubscribeError(t *testing.T) {
	cause := errors.New("subscribe error")
	err := SubscribeError("test.queue", cause)

	assert.Equal(t, ErrCodeSubscribeFailed, err.Code)
	assert.Equal(t, "test.queue", err.Topic)
}

func TestHandlerError(t *testing.T) {
	cause := errors.New("handler panic")
	err := HandlerError("msg-123", cause)

	assert.Equal(t, ErrCodeHandlerError, err.Code)
	assert.Equal(t, "msg-123", err.MessageID)
}

func TestQueueError(t *testing.T) {
	cause := errors.New("queue error")
	err := QueueError("test.queue", cause)

	assert.Equal(t, ErrCodeQueueDeclareFailed, err.Code)
	assert.Equal(t, "test.queue", err.Topic)
}

func TestTopicError(t *testing.T) {
	cause := errors.New("topic error")
	err := TopicError("test.topic", cause)

	assert.Equal(t, ErrCodeTopicCreateFailed, err.Code)
	assert.Equal(t, "test.topic", err.Topic)
}

func TestMessageError(t *testing.T) {
	cause := errors.New("message error")
	err := MessageError(ErrCodeMessageTooLarge, "msg-123", cause)

	assert.Equal(t, ErrCodeMessageTooLarge, err.Code)
	assert.Equal(t, "msg-123", err.MessageID)
}

func TestSerializationError(t *testing.T) {
	cause := errors.New("json error")
	err := SerializationError(cause)

	assert.Equal(t, ErrCodeSerializationFailed, err.Code)
	assert.Equal(t, cause, err.Cause)
}

func TestConfigError(t *testing.T) {
	err := ConfigError("invalid port")

	assert.Equal(t, ErrCodeInvalidConfig, err.Code)
	assert.Equal(t, "invalid port", err.Message)
}

func TestIsBrokerError(t *testing.T) {
	brokerErr := NewBrokerError(ErrCodeConnectionFailed, "error", nil)
	stdErr := errors.New("standard error")

	assert.True(t, IsBrokerError(brokerErr))
	assert.False(t, IsBrokerError(stdErr))
}

func TestGetBrokerError(t *testing.T) {
	brokerErr := NewBrokerError(ErrCodeConnectionFailed, "error", nil)
	stdErr := errors.New("standard error")

	result := GetBrokerError(brokerErr)
	require.NotNil(t, result)
	assert.Equal(t, ErrCodeConnectionFailed, result.Code)

	result2 := GetBrokerError(stdErr)
	assert.Nil(t, result2)
}

func TestIsRetryableError(t *testing.T) {
	retryable := NewBrokerError(ErrCodeConnectionFailed, "error", nil)
	notRetryable := NewBrokerError(ErrCodeMessageInvalid, "error", nil)
	stdErr := errors.New("standard error")

	assert.True(t, IsRetryableError(retryable))
	assert.False(t, IsRetryableError(notRetryable))
	assert.False(t, IsRetryableError(stdErr))
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{NewBrokerError(ErrCodeConnectionFailed, "error", nil), true},
		{NewBrokerError(ErrCodeConnectionClosed, "error", nil), true},
		{NewBrokerError(ErrCodeConnectionTimeout, "error", nil), true},
		{NewBrokerError(ErrCodeReconnectionFailed, "error", nil), true},
		{NewBrokerError(ErrCodeAuthenticationFailed, "error", nil), true},
		{NewBrokerError(ErrCodePublishFailed, "error", nil), false},
		{ErrConnectionFailed, true},
		{ErrConnectionClosed, true},
		{ErrConnectionTimeout, true},
		{errors.New("standard error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.expected, IsConnectionError(tt.err))
		})
	}
}

func TestIsPublishError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{NewBrokerError(ErrCodePublishFailed, "error", nil), true},
		{NewBrokerError(ErrCodePublishTimeout, "error", nil), true},
		{NewBrokerError(ErrCodePublishRejected, "error", nil), true},
		{NewBrokerError(ErrCodeConfirmTimeout, "error", nil), true},
		{NewBrokerError(ErrCodeConnectionFailed, "error", nil), false},
		{ErrPublishFailed, true},
		{ErrPublishTimeout, true},
		{ErrPublishRejected, true},
		{errors.New("standard error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.expected, IsPublishError(tt.err))
		})
	}
}

func TestIsSubscribeError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{NewBrokerError(ErrCodeSubscribeFailed, "error", nil), true},
		{NewBrokerError(ErrCodeConsumerCanceled, "error", nil), true},
		{NewBrokerError(ErrCodeHandlerError, "error", nil), true},
		{NewBrokerError(ErrCodePublishFailed, "error", nil), false},
		{ErrSubscribeFailed, true},
		{ErrConsumerCanceled, true},
		{ErrHandlerError, true},
		{errors.New("standard error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.expected, IsSubscribeError(tt.err))
		})
	}
}

func TestWrapError(t *testing.T) {
	cause := errors.New("underlying error")
	err := WrapError(cause, ErrCodeConnectionFailed, "wrapped error")

	assert.Equal(t, ErrCodeConnectionFailed, err.Code)
	assert.Equal(t, "wrapped error", err.Message)
	assert.Equal(t, cause, err.Cause)
}

func TestMultiError(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		me := NewMultiError()
		assert.False(t, me.HasErrors())
		assert.Nil(t, me.ErrorOrNil())
		assert.Equal(t, "no errors", me.Error())
	})

	t.Run("single error", func(t *testing.T) {
		err := errors.New("error 1")
		me := NewMultiError(err)
		assert.True(t, me.HasErrors())
		assert.NotNil(t, me.ErrorOrNil())
		assert.Equal(t, "error 1", me.Error())
		assert.Equal(t, err, me.Unwrap())
	})

	t.Run("multiple errors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		me := NewMultiError(err1, err2)
		assert.True(t, me.HasErrors())
		assert.Len(t, me.Errors, 2)
		assert.Contains(t, me.Error(), "multiple errors")
	})

	t.Run("add error", func(t *testing.T) {
		me := NewMultiError()
		me.Add(errors.New("error 1"))
		assert.Len(t, me.Errors, 1)
		me.Add(nil) // Should not add nil
		assert.Len(t, me.Errors, 1)
	})
}
