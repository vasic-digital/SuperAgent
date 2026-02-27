package config

import (
	"os"
	"testing"
)

// TestContainersRemoteEnabledPrecedence verifies that CONTAINERS_REMOTE_ENABLED
// takes precedence over individual SVC_*_REMOTE environment variables.
// This is a critical fix for remote distribution to work correctly.
func TestContainersRemoteEnabledPrecedence(t *testing.T) {
	// Save and restore original environment
	originalRemoteEnabled := os.Getenv("CONTAINERS_REMOTE_ENABLED")
	originalPgRemote := os.Getenv("SVC_POSTGRESQL_REMOTE")
	originalRedisRemote := os.Getenv("SVC_REDIS_REMOTE")

	defer func() {
		os.Setenv("CONTAINERS_REMOTE_ENABLED", originalRemoteEnabled)
		if originalPgRemote != "" {
			os.Setenv("SVC_POSTGRESQL_REMOTE", originalPgRemote)
		} else {
			os.Unsetenv("SVC_POSTGRESQL_REMOTE")
		}
		if originalRedisRemote != "" {
			os.Setenv("SVC_REDIS_REMOTE", originalRedisRemote)
		} else {
			os.Unsetenv("SVC_REDIS_REMOTE")
		}
	}()

	tests := []struct {
		name                   string
		containersRemote       string
		svcPostgresqlRemote    string
		svcRedisRemote         string
		expectPostgresqlRemote bool
		expectRedisRemote      bool
	}{
		{
			name:                   "Global remote enabled ignores individual false",
			containersRemote:       "true",
			svcPostgresqlRemote:    "false",
			svcRedisRemote:         "false",
			expectPostgresqlRemote: true, // Should be true despite SVC_POSTGRESQL_REMOTE=false
			expectRedisRemote:      true, // Should be true despite SVC_REDIS_REMOTE=false
		},
		{
			name:                   "Global remote disabled allows individual true",
			containersRemote:       "false",
			svcPostgresqlRemote:    "true",
			svcRedisRemote:         "true",
			expectPostgresqlRemote: true, // Should be true because SVC_POSTGRESQL_REMOTE=true
			expectRedisRemote:      true, // Should be true because SVC_REDIS_REMOTE=true
		},
		{
			name:                   "Global remote disabled allows individual false",
			containersRemote:       "false",
			svcPostgresqlRemote:    "false",
			svcRedisRemote:         "false",
			expectPostgresqlRemote: false, // Should be false
			expectRedisRemote:      false, // Should be false
		},
		{
			name:                   "Global remote enabled, no individual overrides",
			containersRemote:       "true",
			svcPostgresqlRemote:    "",
			svcRedisRemote:         "",
			expectPostgresqlRemote: true, // Should be true from DefaultServicesConfig
			expectRedisRemote:      true, // Should be true from DefaultServicesConfig
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("CONTAINERS_REMOTE_ENABLED", tt.containersRemote)
			if tt.svcPostgresqlRemote != "" {
				os.Setenv("SVC_POSTGRESQL_REMOTE", tt.svcPostgresqlRemote)
			} else {
				os.Unsetenv("SVC_POSTGRESQL_REMOTE")
			}
			if tt.svcRedisRemote != "" {
				os.Setenv("SVC_REDIS_REMOTE", tt.svcRedisRemote)
			} else {
				os.Unsetenv("SVC_REDIS_REMOTE")
			}

			// Load configuration
			cfg := Load()

			// Verify PostgreSQL Remote setting
			if cfg.Services.PostgreSQL.Remote != tt.expectPostgresqlRemote {
				t.Errorf("PostgreSQL.Remote = %v, want %v (CONTAINERS_REMOTE_ENABLED=%s, SVC_POSTGRESQL_REMOTE=%s)",
					cfg.Services.PostgreSQL.Remote,
					tt.expectPostgresqlRemote,
					tt.containersRemote,
					tt.svcPostgresqlRemote)
			}

			// Verify Redis Remote setting
			if cfg.Services.Redis.Remote != tt.expectRedisRemote {
				t.Errorf("Redis.Remote = %v, want %v (CONTAINERS_REMOTE_ENABLED=%s, SVC_REDIS_REMOTE=%s)",
					cfg.Services.Redis.Remote,
					tt.expectRedisRemote,
					tt.containersRemote,
					tt.svcRedisRemote)
			}
		})
	}
}

// TestLoadServicesFromEnvPreservesRemote verifies that LoadServicesFromEnv
// correctly preserves Remote=true when CONTAINERS_REMOTE_ENABLED=true
func TestLoadServicesFromEnvPreservesRemote(t *testing.T) {
	// Save and restore
	originalRemoteEnabled := os.Getenv("CONTAINERS_REMOTE_ENABLED")
	defer os.Setenv("CONTAINERS_REMOTE_ENABLED", originalRemoteEnabled)

	// Set global remote enabled
	os.Setenv("CONTAINERS_REMOTE_ENABLED", "true")

	// Create initial config with Remote=true
	cfg := DefaultServicesConfig()

	if !cfg.PostgreSQL.Remote {
		t.Fatal("DefaultServicesConfig should set PostgreSQL.Remote=true when CONTAINERS_REMOTE_ENABLED=true")
	}
	if !cfg.Redis.Remote {
		t.Fatal("DefaultServicesConfig should set Redis.Remote=true when CONTAINERS_REMOTE_ENABLED=true")
	}

	// Now apply LoadServicesFromEnv with conflicting SVC_*_REMOTE=false
	os.Setenv("SVC_POSTGRESQL_REMOTE", "false")
	os.Setenv("SVC_REDIS_REMOTE", "false")

	LoadServicesFromEnv(&cfg)

	// Remote should still be true (global setting takes precedence)
	if !cfg.PostgreSQL.Remote {
		t.Error("LoadServicesFromEnv should preserve PostgreSQL.Remote=true when CONTAINERS_REMOTE_ENABLED=true")
	}
	if !cfg.Redis.Remote {
		t.Error("LoadServicesFromEnv should preserve Redis.Remote=true when CONTAINERS_REMOTE_ENABLED=true")
	}
}
