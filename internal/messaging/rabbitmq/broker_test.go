package rabbitmq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"dev.helix.agent/internal/messaging"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5672, cfg.Port)
	assert.Equal(t, "guest", cfg.Username)
	assert.Equal(t, "guest", cfg.Password)
	assert.Equal(t, "/", cfg.VirtualHost)
	assert.False(t, cfg.TLS)
	assert.Equal(t, 30*time.Second, cfg.ConnectTimeout)
	assert.Equal(t, 5*time.Second, cfg.ReconnectInterval)
	assert.Equal(t, 0, cfg.MaxReconnectAttempts)
	assert.Equal(t, 60*time.Second, cfg.Heartbeat)
	assert.True(t, cfg.PublisherConfirm)
	assert.Equal(t, 30*time.Second, cfg.PublisherConfirmTimeout)
	assert.Equal(t, 10, cfg.PrefetchCount)
	assert.True(t, cfg.DefaultQueueDurable)
	assert.False(t, cfg.DefaultQueueAutoDelete)
	assert.Equal(t, "direct", cfg.DefaultExchangeType)
	assert.True(t, cfg.DefaultExchangeDurable)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty host",
			cfg: &Config{
				Host: "",
				Port: 5672,
			},
			wantErr: true,
		},
		{
			name: "invalid port zero",
			cfg: &Config{
				Host: "localhost",
				Port: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid port too high",
			cfg: &Config{
				Host: "localhost",
				Port: 70000,
			},
			wantErr: true,
		},
		{
			name: "TLS without cert",
			cfg: &Config{
				Host:        "localhost",
				Port:        5672,
				TLS:         true,
				TLSCertFile: "",
				TLSInsecure: false,
			},
			wantErr: true,
		},
		{
			name: "TLS with cert",
			cfg: &Config{
				Host:        "localhost",
				Port:        5672,
				TLS:         true,
				TLSCertFile: "/path/to/cert",
			},
			wantErr: false,
		},
		{
			name: "TLS insecure without cert",
			cfg: &Config{
				Host:        "localhost",
				Port:        5672,
				TLS:         true,
				TLSInsecure: true,
			},
			wantErr: false,
		},
		{
			name: "negative prefetch count",
			cfg: &Config{
				Host:          "localhost",
				Port:          5672,
				PrefetchCount: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_URL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name: "standard URL",
			cfg: &Config{
				Host:        "localhost",
				Port:        5672,
				Username:    "guest",
				Password:    "guest",
				VirtualHost: "/",
			},
			expected: "amqp://guest:guest@localhost:5672/",
		},
		{
			name: "TLS URL",
			cfg: &Config{
				Host:        "localhost",
				Port:        5671,
				Username:    "user",
				Password:    "pass",
				VirtualHost: "/vhost",
				TLS:         true,
			},
			expected: "amqps://user:pass@localhost:5671/vhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cfg.URL())
		})
	}
}

func TestExchangeType_String(t *testing.T) {
	tests := []struct {
		et       ExchangeType
		expected string
	}{
		{ExchangeDirect, "direct"},
		{ExchangeFanout, "fanout"},
		{ExchangeTopic, "topic"},
		{ExchangeHeaders, "headers"},
	}

	for _, tt := range tests {
		t.Run(string(tt.et), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.et.String())
		})
	}
}

func TestExchangeType_IsValid(t *testing.T) {
	tests := []struct {
		et      ExchangeType
		isValid bool
	}{
		{ExchangeDirect, true},
		{ExchangeFanout, true},
		{ExchangeTopic, true},
		{ExchangeHeaders, true},
		{ExchangeType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.et), func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.et.IsValid())
		})
	}
}

func TestDefaultExchangeConfig(t *testing.T) {
	cfg := DefaultExchangeConfig("test.exchange", ExchangeTopic)

	assert.Equal(t, "test.exchange", cfg.Name)
	assert.Equal(t, ExchangeTopic, cfg.Type)
	assert.True(t, cfg.Durable)
	assert.False(t, cfg.AutoDelete)
	assert.False(t, cfg.Internal)
	assert.False(t, cfg.NoWait)
}

func TestDefaultQueueConfig(t *testing.T) {
	cfg := DefaultQueueConfig("test.queue")

	assert.Equal(t, "test.queue", cfg.Name)
	assert.True(t, cfg.Durable)
	assert.False(t, cfg.AutoDelete)
	assert.False(t, cfg.Exclusive)
	assert.False(t, cfg.NoWait)
}

func TestQueueConfig_ToArgs(t *testing.T) {
	maxPriority := 10
	cfg := &QueueConfig{
		Name:                 "test.queue",
		DeadLetterExchange:   "dlx",
		DeadLetterRoutingKey: "dlx.key",
		MessageTTL:           3600000,
		MaxLength:            1000,
		MaxLengthBytes:       1024 * 1024,
		MaxPriority:          &maxPriority,
		Args: map[string]interface{}{
			"custom-arg": "value",
		},
	}

	args := cfg.ToArgs()

	assert.Equal(t, "dlx", args["x-dead-letter-exchange"])
	assert.Equal(t, "dlx.key", args["x-dead-letter-routing-key"])
	assert.Equal(t, int64(3600000), args["x-message-ttl"])
	assert.Equal(t, int64(1000), args["x-max-length"])
	assert.Equal(t, int64(1024*1024), args["x-max-length-bytes"])
	assert.Equal(t, 10, args["x-max-priority"])
	assert.Equal(t, "value", args["custom-arg"])
}

func TestDefaultBindingConfig(t *testing.T) {
	cfg := DefaultBindingConfig("test.queue", "test.exchange", "routing.key")

	assert.Equal(t, "test.queue", cfg.Queue)
	assert.Equal(t, "test.exchange", cfg.Exchange)
	assert.Equal(t, "routing.key", cfg.RoutingKey)
	assert.False(t, cfg.NoWait)
}

func TestDefaultTopologyConfig(t *testing.T) {
	cfg := DefaultTopologyConfig()

	assert.Len(t, cfg.Exchanges, 4)
	assert.Len(t, cfg.Queues, 6)
	assert.Len(t, cfg.Bindings, 6)

	// Check exchange names
	exchangeNames := make([]string, len(cfg.Exchanges))
	for i, ex := range cfg.Exchanges {
		exchangeNames[i] = ex.Name
	}
	assert.Contains(t, exchangeNames, messaging.ExchangeTasks)
	assert.Contains(t, exchangeNames, messaging.ExchangeEvents)
	assert.Contains(t, exchangeNames, messaging.ExchangeNotifications)
	assert.Contains(t, exchangeNames, messaging.ExchangeDeadLetter)

	// Check queue names
	queueNames := make([]string, len(cfg.Queues))
	for i, q := range cfg.Queues {
		queueNames[i] = q.Name
	}
	assert.Contains(t, queueNames, messaging.QueueBackgroundTasks)
	assert.Contains(t, queueNames, messaging.QueueLLMRequests)
	assert.Contains(t, queueNames, messaging.QueueDebateRounds)
	assert.Contains(t, queueNames, messaging.QueueVerification)
	assert.Contains(t, queueNames, messaging.QueueNotifications)
	assert.Contains(t, queueNames, messaging.QueueDeadLetter)
}

func TestNewBroker(t *testing.T) {
	// With nil config
	broker := NewBroker(nil)
	assert.NotNil(t, broker)
	assert.NotNil(t, broker.config)
	assert.Equal(t, "localhost", broker.config.Host)

	// With custom config
	cfg := &Config{
		Host: "custom-host",
		Port: 5673,
	}
	broker2 := NewBroker(cfg)
	assert.Equal(t, "custom-host", broker2.config.Host)
	assert.Equal(t, 5673, broker2.config.Port)
}

func TestBroker_BrokerType(t *testing.T) {
	broker := NewBroker(nil)
	assert.Equal(t, messaging.BrokerTypeRabbitMQ, broker.BrokerType())
}

func TestBroker_IsConnected_WhenNotConnected(t *testing.T) {
	broker := NewBroker(nil)
	assert.False(t, broker.IsConnected())
}

func TestBroker_GetMetrics(t *testing.T) {
	broker := NewBroker(nil)
	metrics := broker.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestBroker_SetTopology(t *testing.T) {
	broker := NewBroker(nil)

	customTopology := &TopologyConfig{
		Exchanges: []ExchangeConfig{
			*DefaultExchangeConfig("custom.exchange", ExchangeTopic),
		},
	}

	broker.SetTopology(customTopology)
	assert.Equal(t, customTopology, broker.topology)
}

func TestNewTaskQueueBroker(t *testing.T) {
	cfg := DefaultConfig()
	broker := NewTaskQueueBroker(cfg)

	assert.NotNil(t, broker)
	assert.NotNil(t, broker.Broker)
	assert.NotNil(t, broker.taskMetrics)
	assert.NotNil(t, broker.consumers)
}

func TestTaskQueueBroker_GetTaskMetrics(t *testing.T) {
	broker := NewTaskQueueBroker(nil)
	metrics := broker.GetTaskMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TasksEnqueued)
	assert.Equal(t, int64(0), metrics.TasksDequeued)
	assert.Equal(t, int64(0), metrics.TasksCompleted)
	assert.Equal(t, int64(0), metrics.TasksFailed)
	assert.Equal(t, int64(0), metrics.TasksDeadLetter)
	assert.Equal(t, int64(0), metrics.TasksRetried)
}

func TestGenerateSubscriptionID(t *testing.T) {
	id1 := generateSubscriptionID()
	id2 := generateSubscriptionID()

	assert.Contains(t, id1, "sub-")
	assert.Contains(t, id2, "sub-")
	// IDs should be different
	assert.NotEqual(t, id1, id2)
}

func TestRandomString(t *testing.T) {
	s1 := randomString(8)
	s2 := randomString(8)

	assert.Len(t, s1, 8)
	assert.Len(t, s2, 8)
	// Strings should be different (with high probability)
	// Note: There's a small chance they could be equal, but very unlikely
}
