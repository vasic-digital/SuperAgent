// Package rabbitmq provides a RabbitMQ message broker implementation.
package rabbitmq

import (
	"fmt"
	"time"

	"dev.helix.agent/internal/messaging"
)

// Config holds configuration for the RabbitMQ broker.
type Config struct {
	// Host is the RabbitMQ server host.
	Host string `json:"host" yaml:"host"`
	// Port is the RabbitMQ server port.
	Port int `json:"port" yaml:"port"`
	// Username for authentication.
	Username string `json:"username" yaml:"username"`
	// Password for authentication.
	Password string `json:"password" yaml:"password"`
	// VirtualHost is the RabbitMQ virtual host.
	VirtualHost string `json:"virtual_host" yaml:"virtual_host"`

	// TLS Configuration
	TLS         bool   `json:"tls" yaml:"tls"`
	TLSCertFile string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty"`
	TLSKeyFile  string `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty"`
	TLSCAFile   string `json:"tls_ca_file,omitempty" yaml:"tls_ca_file,omitempty"`
	TLSInsecure bool   `json:"tls_insecure,omitempty" yaml:"tls_insecure,omitempty"`

	// Connection settings
	ConnectTimeout       time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	ReconnectInterval    time.Duration `json:"reconnect_interval" yaml:"reconnect_interval"`
	MaxReconnectAttempts int           `json:"max_reconnect_attempts" yaml:"max_reconnect_attempts"`
	Heartbeat            time.Duration `json:"heartbeat" yaml:"heartbeat"`

	// Channel settings
	ChannelMax int `json:"channel_max" yaml:"channel_max"`
	FrameSize  int `json:"frame_size" yaml:"frame_size"`

	// Publisher settings
	PublisherConfirm        bool          `json:"publisher_confirm" yaml:"publisher_confirm"`
	PublisherConfirmTimeout time.Duration `json:"publisher_confirm_timeout" yaml:"publisher_confirm_timeout"`
	PublishTimeout          time.Duration `json:"publish_timeout" yaml:"publish_timeout"`

	// Consumer settings
	PrefetchCount int  `json:"prefetch_count" yaml:"prefetch_count"`
	PrefetchSize  int  `json:"prefetch_size" yaml:"prefetch_size"`
	GlobalQos     bool `json:"global_qos" yaml:"global_qos"`

	// Queue defaults
	DefaultQueueDurable    bool `json:"default_queue_durable" yaml:"default_queue_durable"`
	DefaultQueueAutoDelete bool `json:"default_queue_auto_delete" yaml:"default_queue_auto_delete"`
	DefaultQueueExclusive  bool `json:"default_queue_exclusive" yaml:"default_queue_exclusive"`

	// Exchange defaults
	DefaultExchangeType    string `json:"default_exchange_type" yaml:"default_exchange_type"`
	DefaultExchangeDurable bool   `json:"default_exchange_durable" yaml:"default_exchange_durable"`

	// Dead letter configuration
	DeadLetterExchange   string `json:"dead_letter_exchange" yaml:"dead_letter_exchange"`
	DeadLetterRoutingKey string `json:"dead_letter_routing_key" yaml:"dead_letter_routing_key"`
	DeadLetterQueue      string `json:"dead_letter_queue" yaml:"dead_letter_queue"`

	// Retry configuration
	RetryExchange string        `json:"retry_exchange" yaml:"retry_exchange"`
	RetryQueue    string        `json:"retry_queue" yaml:"retry_queue"`
	RetryDelay    time.Duration `json:"retry_delay" yaml:"retry_delay"`
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
}

// DefaultConfig returns the default RabbitMQ configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:                    "localhost",
		Port:                    5672,
		Username:                "guest",
		Password:                "guest",
		VirtualHost:             "/",
		TLS:                     false,
		ConnectTimeout:          30 * time.Second,
		ReconnectInterval:       5 * time.Second,
		MaxReconnectAttempts:    0, // Infinite
		Heartbeat:               60 * time.Second,
		ChannelMax:              0,  // Use server default
		FrameSize:               0,  // Use server default
		PublisherConfirm:        true,
		PublisherConfirmTimeout: 30 * time.Second,
		PublishTimeout:          30 * time.Second,
		PrefetchCount:           10,
		PrefetchSize:            0,
		GlobalQos:               false,
		DefaultQueueDurable:     true,
		DefaultQueueAutoDelete:  false,
		DefaultQueueExclusive:   false,
		DefaultExchangeType:     "direct",
		DefaultExchangeDurable:  true,
		DeadLetterExchange:      "helixagent.dlx",
		DeadLetterRoutingKey:    "dead-letter",
		DeadLetterQueue:         "helixagent.dlq",
		RetryExchange:           "helixagent.retry",
		RetryQueue:              "helixagent.retry.queue",
		RetryDelay:              5 * time.Second,
		MaxRetries:              3,
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Host == "" {
		return messaging.ConfigError("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return messaging.ConfigError("invalid port number")
	}
	if c.VirtualHost == "" {
		c.VirtualHost = "/"
	}
	if c.TLS && c.TLSCertFile == "" && !c.TLSInsecure {
		return messaging.ConfigError("TLS certificate file is required when TLS is enabled")
	}
	if c.PrefetchCount < 0 {
		return messaging.ConfigError("prefetch count must be non-negative")
	}
	return nil
}

// URL returns the AMQP connection URL.
func (c *Config) URL() string {
	scheme := "amqp"
	if c.TLS {
		scheme = "amqps"
	}
	return fmt.Sprintf("%s://%s:%s@%s:%d%s",
		scheme, c.Username, c.Password, c.Host, c.Port, c.VirtualHost)
}

// ExchangeType represents the type of RabbitMQ exchange.
type ExchangeType string

const (
	// ExchangeDirect routes messages to queues based on exact routing key match.
	ExchangeDirect ExchangeType = "direct"
	// ExchangeFanout routes messages to all bound queues.
	ExchangeFanout ExchangeType = "fanout"
	// ExchangeTopic routes messages based on routing key patterns.
	ExchangeTopic ExchangeType = "topic"
	// ExchangeHeaders routes messages based on header attributes.
	ExchangeHeaders ExchangeType = "headers"
)

// String returns the string representation of ExchangeType.
func (e ExchangeType) String() string {
	return string(e)
}

// IsValid checks if the exchange type is valid.
func (e ExchangeType) IsValid() bool {
	switch e {
	case ExchangeDirect, ExchangeFanout, ExchangeTopic, ExchangeHeaders:
		return true
	default:
		return false
	}
}

// ExchangeConfig holds configuration for an exchange.
type ExchangeConfig struct {
	Name       string       `json:"name" yaml:"name"`
	Type       ExchangeType `json:"type" yaml:"type"`
	Durable    bool         `json:"durable" yaml:"durable"`
	AutoDelete bool         `json:"auto_delete" yaml:"auto_delete"`
	Internal   bool         `json:"internal" yaml:"internal"`
	NoWait     bool         `json:"no_wait" yaml:"no_wait"`
	Args       map[string]interface{} `json:"args,omitempty" yaml:"args,omitempty"`
}

// DefaultExchangeConfig returns the default exchange configuration.
func DefaultExchangeConfig(name string, exchangeType ExchangeType) *ExchangeConfig {
	return &ExchangeConfig{
		Name:       name,
		Type:       exchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}
}

// QueueConfig holds configuration for a queue.
type QueueConfig struct {
	Name                 string                 `json:"name" yaml:"name"`
	Durable              bool                   `json:"durable" yaml:"durable"`
	AutoDelete           bool                   `json:"auto_delete" yaml:"auto_delete"`
	Exclusive            bool                   `json:"exclusive" yaml:"exclusive"`
	NoWait               bool                   `json:"no_wait" yaml:"no_wait"`
	Args                 map[string]interface{} `json:"args,omitempty" yaml:"args,omitempty"`
	DeadLetterExchange   string                 `json:"dead_letter_exchange,omitempty" yaml:"dead_letter_exchange,omitempty"`
	DeadLetterRoutingKey string                 `json:"dead_letter_routing_key,omitempty" yaml:"dead_letter_routing_key,omitempty"`
	MessageTTL           int64                  `json:"message_ttl,omitempty" yaml:"message_ttl,omitempty"`
	MaxLength            int64                  `json:"max_length,omitempty" yaml:"max_length,omitempty"`
	MaxLengthBytes       int64                  `json:"max_length_bytes,omitempty" yaml:"max_length_bytes,omitempty"`
	MaxPriority          *int                   `json:"max_priority,omitempty" yaml:"max_priority,omitempty"`
}

// DefaultQueueConfig returns the default queue configuration.
func DefaultQueueConfig(name string) *QueueConfig {
	return &QueueConfig{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	}
}

// ToArgs converts the queue config to AMQP arguments.
func (c *QueueConfig) ToArgs() map[string]interface{} {
	args := make(map[string]interface{})
	if c.Args != nil {
		for k, v := range c.Args {
			args[k] = v
		}
	}
	if c.DeadLetterExchange != "" {
		args["x-dead-letter-exchange"] = c.DeadLetterExchange
	}
	if c.DeadLetterRoutingKey != "" {
		args["x-dead-letter-routing-key"] = c.DeadLetterRoutingKey
	}
	if c.MessageTTL > 0 {
		args["x-message-ttl"] = c.MessageTTL
	}
	if c.MaxLength > 0 {
		args["x-max-length"] = c.MaxLength
	}
	if c.MaxLengthBytes > 0 {
		args["x-max-length-bytes"] = c.MaxLengthBytes
	}
	if c.MaxPriority != nil {
		args["x-max-priority"] = *c.MaxPriority
	}
	return args
}

// BindingConfig holds configuration for a queue-exchange binding.
type BindingConfig struct {
	Queue      string                 `json:"queue" yaml:"queue"`
	Exchange   string                 `json:"exchange" yaml:"exchange"`
	RoutingKey string                 `json:"routing_key" yaml:"routing_key"`
	NoWait     bool                   `json:"no_wait" yaml:"no_wait"`
	Args       map[string]interface{} `json:"args,omitempty" yaml:"args,omitempty"`
}

// DefaultBindingConfig returns the default binding configuration.
func DefaultBindingConfig(queue, exchange, routingKey string) *BindingConfig {
	return &BindingConfig{
		Queue:      queue,
		Exchange:   exchange,
		RoutingKey: routingKey,
		NoWait:     false,
	}
}

// TopologyConfig holds the complete topology configuration.
type TopologyConfig struct {
	Exchanges []ExchangeConfig `json:"exchanges" yaml:"exchanges"`
	Queues    []QueueConfig    `json:"queues" yaml:"queues"`
	Bindings  []BindingConfig  `json:"bindings" yaml:"bindings"`
}

// DefaultTopologyConfig returns the default HelixAgent topology.
func DefaultTopologyConfig() *TopologyConfig {
	maxPriority := 10
	return &TopologyConfig{
		Exchanges: []ExchangeConfig{
			*DefaultExchangeConfig(messaging.ExchangeTasks, ExchangeDirect),
			*DefaultExchangeConfig(messaging.ExchangeEvents, ExchangeTopic),
			*DefaultExchangeConfig(messaging.ExchangeNotifications, ExchangeFanout),
			*DefaultExchangeConfig(messaging.ExchangeDeadLetter, ExchangeDirect),
		},
		Queues: []QueueConfig{
			{
				Name:               messaging.QueueBackgroundTasks,
				Durable:            true,
				DeadLetterExchange: messaging.ExchangeDeadLetter,
				MaxPriority:        &maxPriority,
			},
			{
				Name:               messaging.QueueLLMRequests,
				Durable:            true,
				DeadLetterExchange: messaging.ExchangeDeadLetter,
				MaxPriority:        &maxPriority,
			},
			{
				Name:               messaging.QueueDebateRounds,
				Durable:            true,
				DeadLetterExchange: messaging.ExchangeDeadLetter,
				MaxPriority:        &maxPriority,
			},
			{
				Name:               messaging.QueueVerification,
				Durable:            true,
				DeadLetterExchange: messaging.ExchangeDeadLetter,
			},
			{
				Name:               messaging.QueueNotifications,
				Durable:            true,
				DeadLetterExchange: messaging.ExchangeDeadLetter,
			},
			{
				Name:    messaging.QueueDeadLetter,
				Durable: true,
			},
		},
		Bindings: []BindingConfig{
			*DefaultBindingConfig(messaging.QueueBackgroundTasks, messaging.ExchangeTasks, messaging.QueueBackgroundTasks),
			*DefaultBindingConfig(messaging.QueueLLMRequests, messaging.ExchangeTasks, messaging.QueueLLMRequests),
			*DefaultBindingConfig(messaging.QueueDebateRounds, messaging.ExchangeTasks, messaging.QueueDebateRounds),
			*DefaultBindingConfig(messaging.QueueVerification, messaging.ExchangeTasks, messaging.QueueVerification),
			*DefaultBindingConfig(messaging.QueueNotifications, messaging.ExchangeNotifications, ""),
			*DefaultBindingConfig(messaging.QueueDeadLetter, messaging.ExchangeDeadLetter, "dead-letter"),
		},
	}
}
