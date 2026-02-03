// Package messaging provides adapters bridging HelixAgent's internal messaging
// types to the generic digital.vasic.messaging module.
//
// # Overview
//
// The adapter package maintains backward compatibility with existing code using
// dev.helix.agent/internal/messaging while delegating core broker operations
// to digital.vasic.messaging. This allows HelixAgent to use the extracted
// generic messaging module while preserving all HelixAgent-specific types
// like Task, Event, TaskQueueBroker, and EventStreamBroker.
//
// # Type Mapping
//
// The adapter provides bidirectional conversion between internal and generic types:
//
//   - messaging.Message <-> broker.Message
//   - messaging.BrokerError <-> broker.BrokerError
//   - messaging.BrokerType <-> broker.BrokerType
//   - messaging.MessageHandler <-> broker.Handler
//   - messaging.Subscription <-> broker.Subscription
//
// # Usage
//
// Create an adapter wrapping a generic broker:
//
//	import (
//	    msgadapter "dev.helix.agent/internal/adapters/messaging"
//	    "digital.vasic.messaging/pkg/broker"
//	)
//
//	// Create a generic in-memory broker
//	genericBroker := broker.NewInMemoryBroker()
//
//	// Wrap it with the adapter
//	adapter := msgadapter.NewBrokerAdapter(genericBroker)
//
//	// Use it as an internal messaging.MessageBroker
//	adapter.Connect(ctx)
//	adapter.Publish(ctx, "topic", internalMessage)
//
// # Specialized Adapters
//
// For Kafka and RabbitMQ, use the specialized adapters:
//
//	// Kafka
//	kafkaAdapter := msgadapter.NewKafkaProducerAdapter(kafkaConfig)
//	kafkaConsumer := msgadapter.NewKafkaConsumerAdapter(kafkaConfig)
//
//	// RabbitMQ
//	rabbitAdapter := msgadapter.NewRabbitMQProducerAdapter(rabbitConfig)
//	rabbitConsumer := msgadapter.NewRabbitMQConsumerAdapter(rabbitConfig)
//
// # In-Memory Broker
//
// For testing and development:
//
//	inmemoryAdapter := msgadapter.NewInMemoryBrokerAdapter()
package messaging
