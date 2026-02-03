package messaging

import (
	"context"
	"time"

	"digital.vasic.messaging/pkg/broker"

	"dev.helix.agent/internal/messaging"
)

// InternalToGenericMessage converts an internal Message to a generic Message.
func InternalToGenericMessage(m *messaging.Message) *broker.Message {
	if m == nil {
		return nil
	}

	gm := broker.NewMessage(m.Type, m.Payload)
	gm.ID = m.ID
	gm.Topic = m.Type
	gm.Timestamp = m.Timestamp

	// Copy headers
	for k, v := range m.Headers {
		gm.SetHeader(k, v)
	}

	// Add trace and correlation info as headers
	if m.TraceID != "" {
		gm.SetHeader("x-trace-id", m.TraceID)
	}
	if m.CorrelationID != "" {
		gm.SetHeader("x-correlation-id", m.CorrelationID)
	}
	if m.ReplyTo != "" {
		gm.SetHeader("x-reply-to", m.ReplyTo)
	}

	return gm
}

// GenericToInternalMessage converts a generic Message to an internal Message.
func GenericToInternalMessage(gm *broker.Message) *messaging.Message {
	if gm == nil {
		return nil
	}

	m := messaging.NewMessage(gm.Topic, gm.Value)
	m.ID = gm.ID
	m.Timestamp = gm.Timestamp

	// Copy headers
	for k, v := range gm.Headers {
		switch k {
		case "x-trace-id":
			m.TraceID = v
		case "x-correlation-id":
			m.CorrelationID = v
		case "x-reply-to":
			m.ReplyTo = v
		default:
			m.SetHeader(k, v)
		}
	}

	return m
}

// InternalToGenericError converts internal BrokerError to generic BrokerError.
func InternalToGenericError(e *messaging.BrokerError) *broker.BrokerError {
	if e == nil {
		return nil
	}

	return broker.NewBrokerError(
		broker.ErrorCode(e.Code),
		e.Message,
		e.Cause,
	)
}

// GenericToInternalError converts generic BrokerError to internal BrokerError.
func GenericToInternalError(e *broker.BrokerError) *messaging.BrokerError {
	if e == nil {
		return nil
	}

	return messaging.NewBrokerError(
		messaging.ErrorCode(e.Code),
		e.Message,
		e.Cause,
	)
}

// InternalToGenericBrokerType converts internal BrokerType to generic BrokerType.
func InternalToGenericBrokerType(bt messaging.BrokerType) broker.BrokerType {
	switch bt {
	case messaging.BrokerTypeKafka:
		return broker.BrokerTypeKafka
	case messaging.BrokerTypeRabbitMQ:
		return broker.BrokerTypeRabbitMQ
	case messaging.BrokerTypeInMemory:
		return broker.BrokerTypeInMemory
	default:
		return broker.BrokerTypeInMemory
	}
}

// GenericToInternalBrokerType converts generic BrokerType to internal BrokerType.
func GenericToInternalBrokerType(bt broker.BrokerType) messaging.BrokerType {
	switch bt {
	case broker.BrokerTypeKafka:
		return messaging.BrokerTypeKafka
	case broker.BrokerTypeRabbitMQ:
		return messaging.BrokerTypeRabbitMQ
	case broker.BrokerTypeInMemory:
		return messaging.BrokerTypeInMemory
	default:
		return messaging.BrokerTypeInMemory
	}
}

// InternalToGenericHandler wraps an internal MessageHandler to a generic Handler.
func InternalToGenericHandler(h messaging.MessageHandler) broker.Handler {
	return func(ctx context.Context, msg *broker.Message) error {
		return h(ctx, GenericToInternalMessage(msg))
	}
}

// GenericToInternalHandler wraps a generic Handler to an internal MessageHandler.
func GenericToInternalHandler(h broker.Handler) messaging.MessageHandler {
	return func(ctx context.Context, msg *messaging.Message) error {
		return h(ctx, InternalToGenericMessage(msg))
	}
}

// MessageBatch converts a slice of internal messages to generic messages.
func MessageBatch(msgs []*messaging.Message) []*broker.Message {
	result := make([]*broker.Message, len(msgs))
	for i, m := range msgs {
		result[i] = InternalToGenericMessage(m)
	}
	return result
}

// InternalMessageBatch converts a slice of generic messages to internal messages.
func InternalMessageBatch(msgs []*broker.Message) []*messaging.Message {
	result := make([]*messaging.Message, len(msgs))
	for i, m := range msgs {
		result[i] = GenericToInternalMessage(m)
	}
	return result
}

// Timestamp helpers to preserve precision across conversions.

// TimestampToMillis converts a time.Time to milliseconds since epoch.
func TimestampToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// MillisToTimestamp converts milliseconds since epoch to time.Time.
func MillisToTimestamp(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
