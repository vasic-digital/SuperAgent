package minio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "localhost:9000", config.Endpoint)
	assert.Equal(t, "minioadmin", config.AccessKey)
	assert.Equal(t, "minioadmin123", config.SecretKey)
	assert.False(t, config.UseSSL)
	assert.Equal(t, "us-east-1", config.Region)
	assert.Equal(t, 30*time.Second, config.ConnectTimeout)
	assert.Equal(t, 60*time.Second, config.RequestTimeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, int64(16*1024*1024), config.PartSize)
	assert.Equal(t, 4, config.ConcurrentUploads)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
}

func TestDefaultConfig_IsValidByDefault(t *testing.T) {
	config := DefaultConfig()
	err := config.Validate()
	require.NoError(t, err)
}

func TestConfig_ZeroValue(t *testing.T) {
	config := &Config{}
	err := config.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required")
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		modify      func(*Config)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			modify:      func(c *Config) {},
			expectError: false,
		},
		{
			name: "empty endpoint",
			modify: func(c *Config) {
				c.Endpoint = ""
			},
			expectError: true,
			errorMsg:    "endpoint is required",
		},
		{
			name: "empty access key",
			modify: func(c *Config) {
				c.AccessKey = ""
			},
			expectError: true,
			errorMsg:    "access_key is required",
		},
		{
			name: "empty secret key",
			modify: func(c *Config) {
				c.SecretKey = ""
			},
			expectError: true,
			errorMsg:    "secret_key is required",
		},
		{
			name: "invalid connect timeout",
			modify: func(c *Config) {
				c.ConnectTimeout = 0
			},
			expectError: true,
			errorMsg:    "connect_timeout must be positive",
		},
		{
			name: "negative connect timeout",
			modify: func(c *Config) {
				c.ConnectTimeout = -1 * time.Second
			},
			expectError: true,
			errorMsg:    "connect_timeout must be positive",
		},
		{
			name: "invalid request timeout",
			modify: func(c *Config) {
				c.RequestTimeout = -1
			},
			expectError: true,
			errorMsg:    "request_timeout must be positive",
		},
		{
			name: "zero request timeout",
			modify: func(c *Config) {
				c.RequestTimeout = 0
			},
			expectError: true,
			errorMsg:    "request_timeout must be positive",
		},
		{
			name: "negative max retries",
			modify: func(c *Config) {
				c.MaxRetries = -1
			},
			expectError: true,
			errorMsg:    "max_retries cannot be negative",
		},
		{
			name: "zero max retries is valid",
			modify: func(c *Config) {
				c.MaxRetries = 0
			},
			expectError: false,
		},
		{
			name: "part size too small",
			modify: func(c *Config) {
				c.PartSize = 1024 // 1KB, less than 5MB minimum
			},
			expectError: true,
			errorMsg:    "part_size must be at least 5MB",
		},
		{
			name: "part size at minimum boundary",
			modify: func(c *Config) {
				c.PartSize = 5 * 1024 * 1024 // exactly 5MB
			},
			expectError: false,
		},
		{
			name: "part size below minimum boundary",
			modify: func(c *Config) {
				c.PartSize = 5*1024*1024 - 1 // one byte less than 5MB
			},
			expectError: true,
			errorMsg:    "part_size must be at least 5MB",
		},
		{
			name: "invalid concurrent uploads",
			modify: func(c *Config) {
				c.ConcurrentUploads = 0
			},
			expectError: true,
			errorMsg:    "concurrent_uploads must be at least 1",
		},
		{
			name: "negative concurrent uploads",
			modify: func(c *Config) {
				c.ConcurrentUploads = -5
			},
			expectError: true,
			errorMsg:    "concurrent_uploads must be at least 1",
		},
		{
			name: "concurrent uploads at minimum",
			modify: func(c *Config) {
				c.ConcurrentUploads = 1
			},
			expectError: false,
		},
		{
			name: "valid custom config",
			modify: func(c *Config) {
				c.Endpoint = "custom.minio.io:9000"
				c.AccessKey = "custom-access"
				c.SecretKey = "custom-secret"
				c.UseSSL = true
				c.Region = "eu-west-1"
				c.ConnectTimeout = 60 * time.Second
				c.RequestTimeout = 120 * time.Second
				c.MaxRetries = 5
				c.PartSize = 32 * 1024 * 1024
				c.ConcurrentUploads = 8
			},
			expectError: false,
		},
		{
			name: "whitespace only endpoint",
			modify: func(c *Config) {
				c.Endpoint = "   "
			},
			expectError: false, // Note: Validate() doesn't trim whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			tt.modify(config)

			err := config.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_ValidationOrder(t *testing.T) {
	// Test that validation fails on the first error in order
	config := &Config{
		Endpoint:          "", // First validation check
		AccessKey:         "", // Second validation check
		SecretKey:         "",
		ConnectTimeout:    0,
		RequestTimeout:    0,
		MaxRetries:        -1,
		PartSize:          0,
		ConcurrentUploads: 0,
	}

	err := config.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required")
}

func TestDefaultBucketConfig(t *testing.T) {
	config := DefaultBucketConfig("test-bucket")

	assert.Equal(t, "test-bucket", config.Name)
	assert.Equal(t, -1, config.RetentionDays)
	assert.False(t, config.Versioning)
	assert.False(t, config.ObjectLocking)
	assert.False(t, config.Public)
}

func TestDefaultBucketConfig_EmptyName(t *testing.T) {
	config := DefaultBucketConfig("")
	assert.Equal(t, "", config.Name)
}

func TestDefaultBucketConfig_SpecialCharacterNames(t *testing.T) {
	tests := []struct {
		name     string
		bucket   string
		expected string
	}{
		{"simple name", "my-bucket", "my-bucket"},
		{"with dots", "my.bucket.name", "my.bucket.name"},
		{"with numbers", "bucket-123", "bucket-123"},
		{"unicode name", "bucket-\u4e2d\u6587", "bucket-\u4e2d\u6587"},
		{"long name", "a-very-long-bucket-name-that-exceeds-normal-length-expectations",
			"a-very-long-bucket-name-that-exceeds-normal-length-expectations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultBucketConfig(tt.bucket)
			assert.Equal(t, tt.expected, config.Name)
		})
	}
}

func TestBucketConfigChaining(t *testing.T) {
	config := DefaultBucketConfig("test-bucket").
		WithRetention(30).
		WithVersioning().
		WithObjectLocking().
		WithPublicAccess()

	assert.Equal(t, "test-bucket", config.Name)
	assert.Equal(t, 30, config.RetentionDays)
	assert.True(t, config.Versioning)
	assert.True(t, config.ObjectLocking)
	assert.True(t, config.Public)
}

func TestBucketConfig_WithRetention(t *testing.T) {
	tests := []struct {
		name string
		days int
	}{
		{"positive retention", 30},
		{"zero retention", 0},
		{"negative retention", -1},
		{"large retention", 365},
		{"very large retention", 3650},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultBucketConfig("test").WithRetention(tt.days)
			assert.Equal(t, tt.days, config.RetentionDays)
		})
	}
}

func TestBucketConfig_WithVersioning(t *testing.T) {
	config := DefaultBucketConfig("test")
	assert.False(t, config.Versioning)

	result := config.WithVersioning()

	// Verify it returns the same pointer for chaining
	assert.Same(t, config, result)
	assert.True(t, config.Versioning)
}

func TestBucketConfig_WithObjectLocking(t *testing.T) {
	config := DefaultBucketConfig("test")
	assert.False(t, config.ObjectLocking)

	result := config.WithObjectLocking()

	assert.Same(t, config, result)
	assert.True(t, config.ObjectLocking)
}

func TestBucketConfig_WithPublicAccess(t *testing.T) {
	config := DefaultBucketConfig("test")
	assert.False(t, config.Public)

	result := config.WithPublicAccess()

	assert.Same(t, config, result)
	assert.True(t, config.Public)
}

func TestBucketConfig_ChainingOrder(t *testing.T) {
	// Test that chaining order doesn't matter
	config1 := DefaultBucketConfig("test1").
		WithRetention(30).
		WithVersioning().
		WithObjectLocking().
		WithPublicAccess()

	config2 := DefaultBucketConfig("test1").
		WithPublicAccess().
		WithObjectLocking().
		WithVersioning().
		WithRetention(30)

	assert.Equal(t, config1.Name, config2.Name)
	assert.Equal(t, config1.RetentionDays, config2.RetentionDays)
	assert.Equal(t, config1.Versioning, config2.Versioning)
	assert.Equal(t, config1.ObjectLocking, config2.ObjectLocking)
	assert.Equal(t, config1.Public, config2.Public)
}

func TestBucketConfig_PartialChaining(t *testing.T) {
	// Only apply some options
	config := DefaultBucketConfig("test").
		WithRetention(7).
		WithVersioning()

	assert.Equal(t, 7, config.RetentionDays)
	assert.True(t, config.Versioning)
	assert.False(t, config.ObjectLocking)
	assert.False(t, config.Public)
}

func TestDefaultLifecycleRule(t *testing.T) {
	rule := DefaultLifecycleRule("expire-old", 90)

	assert.Equal(t, "expire-old", rule.ID)
	assert.Equal(t, "", rule.Prefix)
	assert.True(t, rule.Enabled)
	assert.Equal(t, 90, rule.ExpirationDays)
	assert.Equal(t, 0, rule.NoncurrentDays)
	assert.False(t, rule.DeleteMarkerExpiry)
}

func TestDefaultLifecycleRule_VariousExpirations(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expirationDays int
	}{
		{"short expiration", "short", 1},
		{"medium expiration", "medium", 30},
		{"long expiration", "long", 365},
		{"zero expiration", "zero", 0},
		{"negative expiration", "negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := DefaultLifecycleRule(tt.id, tt.expirationDays)
			assert.Equal(t, tt.id, rule.ID)
			assert.Equal(t, tt.expirationDays, rule.ExpirationDays)
			assert.True(t, rule.Enabled)
		})
	}
}

func TestLifecycleRuleChaining(t *testing.T) {
	rule := DefaultLifecycleRule("expire-logs", 30).
		WithPrefix("logs/").
		WithNoncurrentExpiry(7)

	assert.Equal(t, "expire-logs", rule.ID)
	assert.Equal(t, "logs/", rule.Prefix)
	assert.Equal(t, 30, rule.ExpirationDays)
	assert.Equal(t, 7, rule.NoncurrentDays)
}

func TestLifecycleRule_WithPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"empty prefix", ""},
		{"simple prefix", "logs/"},
		{"nested prefix", "app/logs/2024/"},
		{"with special chars", "logs-\u4e2d\u6587/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := DefaultLifecycleRule("test", 30).WithPrefix(tt.prefix)
			assert.Equal(t, tt.prefix, rule.Prefix)
		})
	}
}

func TestLifecycleRule_WithNoncurrentExpiry(t *testing.T) {
	tests := []struct {
		name string
		days int
	}{
		{"zero days", 0},
		{"one day", 1},
		{"week", 7},
		{"month", 30},
		{"negative days", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := DefaultLifecycleRule("test", 30).WithNoncurrentExpiry(tt.days)
			assert.Equal(t, tt.days, rule.NoncurrentDays)
		})
	}
}

func TestLifecycleRule_ChainingReturnsPointer(t *testing.T) {
	rule := DefaultLifecycleRule("test", 30)

	result1 := rule.WithPrefix("prefix/")
	assert.Same(t, rule, result1)

	result2 := rule.WithNoncurrentExpiry(7)
	assert.Same(t, rule, result2)
}

func TestLifecycleRule_FullChaining(t *testing.T) {
	rule := DefaultLifecycleRule("full-test", 90).
		WithPrefix("archive/").
		WithNoncurrentExpiry(30)

	assert.Equal(t, "full-test", rule.ID)
	assert.Equal(t, "archive/", rule.Prefix)
	assert.True(t, rule.Enabled)
	assert.Equal(t, 90, rule.ExpirationDays)
	assert.Equal(t, 30, rule.NoncurrentDays)
	assert.False(t, rule.DeleteMarkerExpiry)
}

func TestLifecycleRule_DirectConstruction(t *testing.T) {
	// Test direct construction without helper
	rule := &LifecycleRule{
		ID:                 "direct-rule",
		Prefix:             "custom/",
		Enabled:            false,
		ExpirationDays:     45,
		NoncurrentDays:     15,
		DeleteMarkerExpiry: true,
	}

	assert.Equal(t, "direct-rule", rule.ID)
	assert.Equal(t, "custom/", rule.Prefix)
	assert.False(t, rule.Enabled)
	assert.Equal(t, 45, rule.ExpirationDays)
	assert.Equal(t, 15, rule.NoncurrentDays)
	assert.True(t, rule.DeleteMarkerExpiry)
}

func TestConfig_StructFields(t *testing.T) {
	// Test that all struct fields are accessible and have correct types
	config := &Config{
		Endpoint:            "localhost:9000",
		AccessKey:           "access",
		SecretKey:           "secret",
		UseSSL:              true,
		Region:              "us-west-2",
		ConnectTimeout:      45 * time.Second,
		RequestTimeout:      90 * time.Second,
		MaxRetries:          5,
		PartSize:            32 * 1024 * 1024,
		ConcurrentUploads:   8,
		HealthCheckInterval: 60 * time.Second,
	}

	assert.Equal(t, "localhost:9000", config.Endpoint)
	assert.Equal(t, "access", config.AccessKey)
	assert.Equal(t, "secret", config.SecretKey)
	assert.True(t, config.UseSSL)
	assert.Equal(t, "us-west-2", config.Region)
	assert.Equal(t, 45*time.Second, config.ConnectTimeout)
	assert.Equal(t, 90*time.Second, config.RequestTimeout)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, int64(32*1024*1024), config.PartSize)
	assert.Equal(t, 8, config.ConcurrentUploads)
	assert.Equal(t, 60*time.Second, config.HealthCheckInterval)
}

func TestBucketConfig_StructFields(t *testing.T) {
	config := &BucketConfig{
		Name:          "my-bucket",
		RetentionDays: 90,
		Versioning:    true,
		ObjectLocking: true,
		Public:        true,
	}

	assert.Equal(t, "my-bucket", config.Name)
	assert.Equal(t, 90, config.RetentionDays)
	assert.True(t, config.Versioning)
	assert.True(t, config.ObjectLocking)
	assert.True(t, config.Public)
}

func TestLifecycleRule_StructFields(t *testing.T) {
	rule := &LifecycleRule{
		ID:                 "test-rule",
		Prefix:             "test/",
		Enabled:            true,
		ExpirationDays:     30,
		NoncurrentDays:     7,
		DeleteMarkerExpiry: true,
	}

	assert.Equal(t, "test-rule", rule.ID)
	assert.Equal(t, "test/", rule.Prefix)
	assert.True(t, rule.Enabled)
	assert.Equal(t, 30, rule.ExpirationDays)
	assert.Equal(t, 7, rule.NoncurrentDays)
	assert.True(t, rule.DeleteMarkerExpiry)
}
