// Package servers provides MCP server adapters for various services.
package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisAdapterConfig holds configuration for Redis MCP adapter
type RedisAdapterConfig struct {
	// Host is the Redis host
	Host string `json:"host,omitempty"`
	// Port is the Redis port
	Port int `json:"port,omitempty"`
	// Password is the Redis password
	Password string `json:"password,omitempty"`
	// Database is the Redis database number
	Database int `json:"database,omitempty"`
	// PoolSize is the connection pool size
	PoolSize int `json:"pool_size,omitempty"`
	// DialTimeout is the connection timeout
	DialTimeout time.Duration `json:"dial_timeout,omitempty"`
	// ReadTimeout is the read timeout
	ReadTimeout time.Duration `json:"read_timeout,omitempty"`
	// WriteTimeout is the write timeout
	WriteTimeout time.Duration `json:"write_timeout,omitempty"`
	// ReadOnly restricts operations to read-only commands
	ReadOnly bool `json:"read_only"`
	// KeyPrefix prefixes all keys with this string
	KeyPrefix string `json:"key_prefix,omitempty"`
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries,omitempty"`
}

// DefaultRedisAdapterConfig returns default configuration
func DefaultRedisAdapterConfig() RedisAdapterConfig {
	return RedisAdapterConfig{
		Host:         "localhost",
		Port:         6379,
		Database:     0,
		PoolSize:     10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		ReadOnly:     false,
		MaxRetries:   3,
	}
}

// RedisAdapter implements MCP adapter for Redis operations
type RedisAdapter struct {
	config      RedisAdapterConfig
	client      *redis.Client
	initialized bool
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// NewRedisAdapter creates a new Redis MCP adapter
func NewRedisAdapter(config RedisAdapterConfig, logger *logrus.Logger) *RedisAdapter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6379
	}
	if config.PoolSize <= 0 {
		config.PoolSize = 10
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = 3 * time.Second
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 3
	}

	return &RedisAdapter{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the Redis adapter
func (r *RedisAdapter) Initialize(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Password:     r.config.Password,
		DB:           r.config.Database,
		PoolSize:     r.config.PoolSize,
		DialTimeout:  r.config.DialTimeout,
		ReadTimeout:  r.config.ReadTimeout,
		WriteTimeout: r.config.WriteTimeout,
		MaxRetries:   r.config.MaxRetries,
	})

	// Test connection
	if err := r.client.Ping(ctx).Err(); err != nil {
		r.client.Close()
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	r.initialized = true
	r.logger.WithFields(logrus.Fields{
		"host":     r.config.Host,
		"port":     r.config.Port,
		"database": r.config.Database,
	}).Info("Redis adapter initialized")

	return nil
}

// Health returns health status
func (r *RedisAdapter) Health(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return fmt.Errorf("Redis adapter not initialized")
	}

	return r.client.Ping(ctx).Err()
}

// Close closes the adapter and Redis connection
func (r *RedisAdapter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.client != nil {
		if err := r.client.Close(); err != nil {
			return err
		}
		r.client = nil
	}

	r.initialized = false
	return nil
}

// prefixKey adds the key prefix if configured
func (r *RedisAdapter) prefixKey(key string) string {
	if r.config.KeyPrefix != "" {
		return r.config.KeyPrefix + key
	}
	return key
}

// Get gets a value by key
func (r *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return "", fmt.Errorf("adapter not initialized")
	}

	val, err := r.client.Get(ctx, r.prefixKey(key)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Set sets a value with optional expiration
func (r *RedisAdapter) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return fmt.Errorf("write operations not allowed in read-only mode")
	}

	return r.client.Set(ctx, r.prefixKey(key), value, expiration).Err()
}

// Delete deletes a key
func (r *RedisAdapter) Delete(ctx context.Context, keys ...string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return 0, fmt.Errorf("write operations not allowed in read-only mode")
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.prefixKey(key)
	}

	return r.client.Del(ctx, prefixedKeys...).Result()
}

// Exists checks if keys exist
func (r *RedisAdapter) Exists(ctx context.Context, keys ...string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.prefixKey(key)
	}

	return r.client.Exists(ctx, prefixedKeys...).Result()
}

// Keys returns keys matching a pattern
func (r *RedisAdapter) Keys(ctx context.Context, pattern string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	keys, err := r.client.Keys(ctx, r.prefixKey(pattern)).Result()
	if err != nil {
		return nil, err
	}

	// Remove prefix from results
	if r.config.KeyPrefix != "" {
		for i, key := range keys {
			keys[i] = strings.TrimPrefix(key, r.config.KeyPrefix)
		}
	}

	return keys, nil
}

// Expire sets expiration on a key
func (r *RedisAdapter) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return false, fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return false, fmt.Errorf("write operations not allowed in read-only mode")
	}

	return r.client.Expire(ctx, r.prefixKey(key), expiration).Result()
}

// TTL gets the time to live for a key
func (r *RedisAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	return r.client.TTL(ctx, r.prefixKey(key)).Result()
}

// HSet sets a hash field
func (r *RedisAdapter) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return fmt.Errorf("write operations not allowed in read-only mode")
	}

	return r.client.HSet(ctx, r.prefixKey(key), values).Err()
}

// HGet gets a hash field
func (r *RedisAdapter) HGet(ctx context.Context, key, field string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return "", fmt.Errorf("adapter not initialized")
	}

	val, err := r.client.HGet(ctx, r.prefixKey(key), field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// HGetAll gets all hash fields
func (r *RedisAdapter) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	return r.client.HGetAll(ctx, r.prefixKey(key)).Result()
}

// LPush pushes values to the beginning of a list
func (r *RedisAdapter) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return 0, fmt.Errorf("write operations not allowed in read-only mode")
	}

	return r.client.LPush(ctx, r.prefixKey(key), values...).Result()
}

// LRange gets a range of list elements
func (r *RedisAdapter) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	return r.client.LRange(ctx, r.prefixKey(key), start, stop).Result()
}

// SAdd adds members to a set
func (r *RedisAdapter) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return 0, fmt.Errorf("write operations not allowed in read-only mode")
	}

	return r.client.SAdd(ctx, r.prefixKey(key), members...).Result()
}

// SMembers gets all set members
func (r *RedisAdapter) SMembers(ctx context.Context, key string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	return r.client.SMembers(ctx, r.prefixKey(key)).Result()
}

// Incr increments a key
func (r *RedisAdapter) Incr(ctx context.Context, key string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return 0, fmt.Errorf("write operations not allowed in read-only mode")
	}

	return r.client.Incr(ctx, r.prefixKey(key)).Result()
}

// Decr decrements a key
func (r *RedisAdapter) Decr(ctx context.Context, key string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	if r.config.ReadOnly {
		return 0, fmt.Errorf("write operations not allowed in read-only mode")
	}

	return r.client.Decr(ctx, r.prefixKey(key)).Result()
}

// RedisInfo represents Redis server info
type RedisInfo struct {
	Server      map[string]string `json:"server"`
	Clients     map[string]string `json:"clients"`
	Memory      map[string]string `json:"memory"`
	Stats       map[string]string `json:"stats"`
	Keyspace    map[string]string `json:"keyspace"`
}

// Info returns Redis server information
func (r *RedisAdapter) Info(ctx context.Context, sections ...string) (*RedisInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return nil, fmt.Errorf("adapter not initialized")
	}

	section := "all"
	if len(sections) > 0 {
		section = sections[0]
	}

	result, err := r.client.Info(ctx, section).Result()
	if err != nil {
		return nil, err
	}

	info := &RedisInfo{
		Server:   make(map[string]string),
		Clients:  make(map[string]string),
		Memory:   make(map[string]string),
		Stats:    make(map[string]string),
		Keyspace: make(map[string]string),
	}

	// Parse info output
	lines := strings.Split(result, "\n")
	currentSection := "server"
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "# ") {
				currentSection = strings.ToLower(strings.TrimPrefix(line, "# "))
			}
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		switch currentSection {
		case "server":
			info.Server[key] = val
		case "clients":
			info.Clients[key] = val
		case "memory":
			info.Memory[key] = val
		case "stats":
			info.Stats[key] = val
		case "keyspace":
			info.Keyspace[key] = val
		}
	}

	return info, nil
}

// DBSize returns the number of keys in the database
func (r *RedisAdapter) DBSize(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized || r.client == nil {
		return 0, fmt.Errorf("adapter not initialized")
	}

	return r.client.DBSize(ctx).Result()
}

// GetMCPTools returns the list of MCP tools provided by this adapter
func (r *RedisAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "redis_get",
			Description: "Get a value by key",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Key to get",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "redis_set",
			Description: "Set a value with optional expiration",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Key to set",
					},
					"value": map[string]interface{}{
						"type":        "string",
						"description": "Value to set",
					},
					"ttl_seconds": map[string]interface{}{
						"type":        "integer",
						"description": "Time to live in seconds (0 for no expiration)",
					},
				},
				"required": []string{"key", "value"},
			},
		},
		{
			Name:        "redis_delete",
			Description: "Delete one or more keys",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keys": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Keys to delete",
					},
				},
				"required": []string{"keys"},
			},
		},
		{
			Name:        "redis_exists",
			Description: "Check if keys exist",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keys": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Keys to check",
					},
				},
				"required": []string{"keys"},
			},
		},
		{
			Name:        "redis_keys",
			Description: "Find keys matching a pattern",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Pattern to match (e.g., 'user:*')",
					},
				},
				"required": []string{"pattern"},
			},
		},
		{
			Name:        "redis_hset",
			Description: "Set hash fields",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Hash key",
					},
					"fields": map[string]interface{}{
						"type":        "object",
						"description": "Field-value pairs to set",
					},
				},
				"required": []string{"key", "fields"},
			},
		},
		{
			Name:        "redis_hgetall",
			Description: "Get all hash fields",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Hash key",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "redis_lpush",
			Description: "Push values to the beginning of a list",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "List key",
					},
					"values": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Values to push",
					},
				},
				"required": []string{"key", "values"},
			},
		},
		{
			Name:        "redis_lrange",
			Description: "Get a range of list elements",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "List key",
					},
					"start": map[string]interface{}{
						"type":        "integer",
						"description": "Start index",
						"default":     0,
					},
					"stop": map[string]interface{}{
						"type":        "integer",
						"description": "Stop index (-1 for all)",
						"default":     -1,
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "redis_sadd",
			Description: "Add members to a set",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Set key",
					},
					"members": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Members to add",
					},
				},
				"required": []string{"key", "members"},
			},
		},
		{
			Name:        "redis_smembers",
			Description: "Get all set members",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Set key",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "redis_incr",
			Description: "Increment a key's integer value",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Key to increment",
					},
				},
				"required": []string{"key"},
			},
		},
		{
			Name:        "redis_info",
			Description: "Get Redis server information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"section": map[string]interface{}{
						"type":        "string",
						"description": "Info section (server, clients, memory, stats, etc.)",
						"default":     "all",
					},
				},
			},
		},
		{
			Name:        "redis_dbsize",
			Description: "Get the number of keys in the database",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (r *RedisAdapter) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	r.mu.RLock()
	initialized := r.initialized
	r.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	switch toolName {
	case "redis_get":
		key, _ := params["key"].(string)
		return r.Get(ctx, key)

	case "redis_set":
		key, _ := params["key"].(string)
		value, _ := params["value"].(string)
		var ttl time.Duration
		if ttlSeconds, ok := params["ttl_seconds"].(float64); ok && ttlSeconds > 0 {
			ttl = time.Duration(ttlSeconds) * time.Second
		}
		return nil, r.Set(ctx, key, value, ttl)

	case "redis_delete":
		var keys []string
		if k, ok := params["keys"].([]interface{}); ok {
			for _, v := range k {
				if s, ok := v.(string); ok {
					keys = append(keys, s)
				}
			}
		}
		return r.Delete(ctx, keys...)

	case "redis_exists":
		var keys []string
		if k, ok := params["keys"].([]interface{}); ok {
			for _, v := range k {
				if s, ok := v.(string); ok {
					keys = append(keys, s)
				}
			}
		}
		return r.Exists(ctx, keys...)

	case "redis_keys":
		pattern, _ := params["pattern"].(string)
		return r.Keys(ctx, pattern)

	case "redis_hset":
		key, _ := params["key"].(string)
		fields, _ := params["fields"].(map[string]interface{})
		return nil, r.HSet(ctx, key, fields)

	case "redis_hgetall":
		key, _ := params["key"].(string)
		return r.HGetAll(ctx, key)

	case "redis_lpush":
		key, _ := params["key"].(string)
		var values []interface{}
		if v, ok := params["values"].([]interface{}); ok {
			values = v
		}
		return r.LPush(ctx, key, values...)

	case "redis_lrange":
		key, _ := params["key"].(string)
		start := int64(0)
		stop := int64(-1)
		if s, ok := params["start"].(float64); ok {
			start = int64(s)
		}
		if s, ok := params["stop"].(float64); ok {
			stop = int64(s)
		}
		return r.LRange(ctx, key, start, stop)

	case "redis_sadd":
		key, _ := params["key"].(string)
		var members []interface{}
		if m, ok := params["members"].([]interface{}); ok {
			members = m
		}
		return r.SAdd(ctx, key, members...)

	case "redis_smembers":
		key, _ := params["key"].(string)
		return r.SMembers(ctx, key)

	case "redis_incr":
		key, _ := params["key"].(string)
		return r.Incr(ctx, key)

	case "redis_info":
		section, _ := params["section"].(string)
		if section == "" {
			section = "all"
		}
		return r.Info(ctx, section)

	case "redis_dbsize":
		return r.DBSize(ctx)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetCapabilities returns adapter capabilities
func (r *RedisAdapter) GetCapabilities() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"name":        "redis",
		"host":        r.config.Host,
		"port":        r.config.Port,
		"database":    r.config.Database,
		"read_only":   r.config.ReadOnly,
		"key_prefix":  r.config.KeyPrefix,
		"tools":       len(r.GetMCPTools()),
		"initialized": r.initialized,
	}
}

// MarshalJSON implements custom JSON marshaling
func (r *RedisAdapter) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"initialized":  r.initialized,
		"capabilities": r.GetCapabilities(),
	})
}

// parseIntOrDefault parses an int from a string or returns a default
func parseIntOrDefault(s string, defaultVal int64) int64 {
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}
	return defaultVal
}
