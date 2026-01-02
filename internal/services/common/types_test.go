package common

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Recovery and Resilience Type Tests
func TestRecoveryStrategy_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	strategy := RecoveryStrategy{
		ID:          "recovery-001",
		Name:        "Failover Strategy",
		Description: "Automatic failover to backup service",
		Strategy:    "failover",
		Parameters: map[string]any{
			"timeout":     30,
			"maxRetries":  3,
			"backupHosts": []string{"backup1.example.com", "backup2.example.com"},
		},
		Priority:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Marshal to JSON
	data, err := json.Marshal(strategy)
	require.NoError(t, err)
	assert.Contains(t, string(data), "recovery-001")
	assert.Contains(t, string(data), "Failover Strategy")

	// Unmarshal back
	var unmarshaled RecoveryStrategy
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, strategy.ID, unmarshaled.ID)
	assert.Equal(t, strategy.Name, unmarshaled.Name)
	assert.Equal(t, strategy.Strategy, unmarshaled.Strategy)
	assert.Equal(t, strategy.Priority, unmarshaled.Priority)
}

func TestRecoveryProcedure_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	procedure := RecoveryProcedure{
		ID:          "proc-001",
		Name:        "Database Recovery",
		Description: "Procedure to recover database",
		Steps: []RecoveryStep{
			{
				ID:          "step-001",
				Name:        "Stop Service",
				Description: "Stop the database service",
				Action:      "stop_service",
				Parameters:  map[string]any{"service": "postgres"},
				Timeout:     30 * time.Second,
				Retryable:   true,
			},
			{
				ID:          "step-002",
				Name:        "Restore Backup",
				Description: "Restore from latest backup",
				Action:      "restore_backup",
				Parameters:  map[string]any{"backup_id": "latest"},
				Timeout:     5 * time.Minute,
				Retryable:   false,
			},
		},
		Parameters: map[string]any{"notify": true},
		RetryCount: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	data, err := json.Marshal(procedure)
	require.NoError(t, err)
	assert.Contains(t, string(data), "proc-001")
	assert.Contains(t, string(data), "Database Recovery")

	var unmarshaled RecoveryProcedure
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, procedure.ID, unmarshaled.ID)
	assert.Len(t, unmarshaled.Steps, 2)
	assert.Equal(t, "step-001", unmarshaled.Steps[0].ID)
}

func TestRecoveryStep_JSONMarshal(t *testing.T) {
	step := RecoveryStep{
		ID:          "step-001",
		Name:        "Restart Service",
		Description: "Restart the application service",
		Action:      "restart",
		Parameters:  map[string]any{"service": "api", "wait": true},
		Timeout:     60 * time.Second,
		Retryable:   true,
	}

	data, err := json.Marshal(step)
	require.NoError(t, err)

	var unmarshaled RecoveryStep
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, step.ID, unmarshaled.ID)
	assert.Equal(t, step.Action, unmarshaled.Action)
	assert.True(t, unmarshaled.Retryable)
}

// Validation Type Tests
func TestValidationRule_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	rule := ValidationRule{
		ID:          "rule-001",
		Name:        "Email Format",
		Description: "Validates email format",
		Type:        "regex",
		Parameters: map[string]any{
			"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
		},
		Severity:  "error",
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(rule)
	require.NoError(t, err)

	var unmarshaled ValidationRule
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, rule.ID, unmarshaled.ID)
	assert.Equal(t, "regex", unmarshaled.Type)
	assert.True(t, unmarshaled.Enabled)
}

func TestIntegrityCheck_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	check := IntegrityCheck{
		ID:          "check-001",
		Name:        "File Hash Verification",
		Description: "Verifies file integrity using SHA-256",
		Type:        "hash",
		Algorithm:   "sha256",
		Parameters: map[string]any{
			"path":     "/data/important.db",
			"expected": "abc123...",
		},
		Schedule:  "0 0 * * *",
		LastRun:   now.Add(-24 * time.Hour),
		Status:    "passed",
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(check)
	require.NoError(t, err)

	var unmarshaled IntegrityCheck
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, check.ID, unmarshaled.ID)
	assert.Equal(t, "sha256", unmarshaled.Algorithm)
	assert.Equal(t, "passed", unmarshaled.Status)
}

// Authentication and Authorization Type Tests
func TestAuthenticationManager_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	manager := AuthenticationManager{
		ID:   "auth-mgr-001",
		Name: "Main Auth Manager",
		AuthMethods: map[string]any{
			"jwt":   true,
			"oauth": true,
		},
		AuthProviders: map[string]any{
			"google":   map[string]any{"enabled": true},
			"github":   map[string]any{"enabled": true},
			"internal": map[string]any{"enabled": true},
		},
		AuthValidators: map[string]any{
			"token_validator": map[string]any{"type": "jwt"},
		},
		CredentialManagers: map[string]any{
			"password": map[string]any{"hash": "bcrypt"},
		},
		AuthPolicies:    []any{"require_mfa", "session_timeout"},
		SessionPolicies: []any{"max_sessions_5", "idle_timeout_30m"},
		MFAProviders:    []any{"totp", "sms", "email"},
		Configuration:   map[string]any{"default_timeout": 3600},
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	data, err := json.Marshal(manager)
	require.NoError(t, err)

	var unmarshaled AuthenticationManager
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, manager.ID, unmarshaled.ID)
	assert.Equal(t, "active", unmarshaled.Status)
}

func TestAuthenticationProvider_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	provider := AuthenticationProvider{
		ID:   "provider-001",
		Name: "OAuth Provider",
		Type: "oauth2",
		Providers: map[string]any{
			"google": map[string]any{"client_id": "xxx"},
		},
		Protocols: map[string]any{
			"oauth2": map[string]any{"version": "2.0"},
		},
		TokenManagers: map[string]any{
			"jwt": map[string]any{"secret": "hidden"},
		},
		IdentityProviders: map[string]any{
			"internal": map[string]any{"type": "database"},
		},
		AuthStrategies:    []any{"oauth2", "api_key"},
		ValidationMethods: []any{"signature", "expiry"},
		Configuration:     map[string]any{"strict_mode": true},
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	data, err := json.Marshal(provider)
	require.NoError(t, err)

	var unmarshaled AuthenticationProvider
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, provider.ID, unmarshaled.ID)
	assert.Equal(t, "oauth2", unmarshaled.Type)
}

func TestAccessController_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	controller := AccessController{
		ID:   "ac-001",
		Name: "Main Access Controller",
		AccessRules: []any{
			map[string]any{"resource": "/api/*", "action": "allow"},
		},
		PermissionMatrix: map[string]any{
			"admin": []string{"read", "write", "delete"},
			"user":  []string{"read"},
		},
		RoleHierarchy: map[string]any{
			"super_admin": []string{"admin"},
			"admin":       []string{"user"},
		},
		ResourceHierarchy: map[string]any{
			"/api": []string{"/api/v1", "/api/v2"},
		},
		AccessPolicies:        []any{"default_deny", "log_access"},
		EnforcementStrategies: []any{"strict", "audit"},
		Configuration:         map[string]any{"cache_ttl": 300},
		Status:                "enforcing",
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	data, err := json.Marshal(controller)
	require.NoError(t, err)

	var unmarshaled AccessController
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, controller.ID, unmarshaled.ID)
	assert.Equal(t, "enforcing", unmarshaled.Status)
}

func TestPermissionManager_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	manager := PermissionManager{
		ID:   "pm-001",
		Name: "Permission Manager",
		Permissions: map[string]any{
			"read":   map[string]any{"description": "Read access"},
			"write":  map[string]any{"description": "Write access"},
			"delete": map[string]any{"description": "Delete access"},
		},
		Roles: map[string]any{
			"admin": map[string]any{"permissions": []string{"read", "write", "delete"}},
			"user":  map[string]any{"permissions": []string{"read"}},
		},
		Policies: []any{"least_privilege", "separation_of_duties"},
		PermissionMatrix: map[string]any{
			"resource_x": map[string]any{"admin": "full", "user": "read_only"},
		},
		RoleMappings:  map[string]any{"default": "user"},
		Configuration: map[string]any{"inherit_permissions": true},
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	data, err := json.Marshal(manager)
	require.NoError(t, err)

	var unmarshaled PermissionManager
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, manager.ID, unmarshaled.ID)
	assert.Equal(t, "active", unmarshaled.Status)
}

// Audit and Logging Type Tests
func TestAuditTrail_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	audit := AuditTrail{
		ID:        "audit-001",
		Timestamp: now,
		Action:    "user.login",
		Actor:     "user@example.com",
		Resource:  "/api/v1/auth/login",
		Outcome:   "success",
		Details: map[string]any{
			"method": "password",
			"mfa":    true,
		},
		Metadata: map[string]any{
			"browser": "Chrome",
			"os":      "Linux",
		},
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64)",
		SessionID: "sess-123456",
		RequestID: "req-789012",
		TraceID:   "trace-345678",
	}

	data, err := json.Marshal(audit)
	require.NoError(t, err)

	var unmarshaled AuditTrail
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, audit.ID, unmarshaled.ID)
	assert.Equal(t, "user.login", unmarshaled.Action)
	assert.Equal(t, "success", unmarshaled.Outcome)
	assert.Equal(t, "192.168.1.100", unmarshaled.IPAddress)
}

// Date and Time Type Tests
func TestDateRange_JSONMarshal(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	dateRange := DateRange{
		StartDate: start,
		EndDate:   end,
		Timezone:  "UTC",
		Inclusive: true,
	}

	data, err := json.Marshal(dateRange)
	require.NoError(t, err)

	var unmarshaled DateRange
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, "UTC", unmarshaled.Timezone)
	assert.True(t, unmarshaled.Inclusive)
}

// Result Type Tests
func TestValidationResult_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	result := ValidationResult{
		Valid:   false,
		RuleID:  "rule-001",
		Message: "Email format is invalid",
		Details: map[string]any{
			"field":   "email",
			"pattern": "expected format: user@domain.com",
		},
		Severity:  "error",
		Timestamp: now,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled ValidationResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.False(t, unmarshaled.Valid)
	assert.Equal(t, "rule-001", unmarshaled.RuleID)
	assert.Equal(t, "error", unmarshaled.Severity)
}

func TestRecoveryResult_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	result := RecoveryResult{
		Success:     true,
		ProcedureID: "proc-001",
		StepID:      "step-003",
		Message:     "Recovery completed successfully",
		Details: map[string]any{
			"restored_rows":    1000,
			"backup_timestamp": "2025-01-01T00:00:00Z",
		},
		Duration:  2 * time.Minute,
		Timestamp: now,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled RecoveryResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.True(t, unmarshaled.Success)
	assert.Equal(t, "proc-001", unmarshaled.ProcedureID)
}

func TestAuthenticationResult_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	expiry := now.Add(24 * time.Hour)

	result := AuthenticationResult{
		Success:     true,
		Token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		ExpiresAt:   expiry,
		UserID:      "user-001",
		Roles:       []string{"admin", "user"},
		Permissions: []string{"read", "write", "delete"},
		Message:     "Authentication successful",
		Details: map[string]any{
			"login_method": "password",
			"mfa_used":     true,
		},
		Timestamp: now,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled AuthenticationResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.True(t, unmarshaled.Success)
	assert.Equal(t, "user-001", unmarshaled.UserID)
	assert.Len(t, unmarshaled.Roles, 2)
	assert.Contains(t, unmarshaled.Roles, "admin")
}

func TestAuthenticationResult_FailedAuth(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	result := AuthenticationResult{
		Success: false,
		Message: "Invalid credentials",
		Details: map[string]any{
			"reason": "password_mismatch",
		},
		Timestamp: now,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled AuthenticationResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.False(t, unmarshaled.Success)
	assert.Empty(t, unmarshaled.Token)
	assert.Empty(t, unmarshaled.UserID)
}

// Zero value tests
func TestZeroValues(t *testing.T) {
	t.Run("RecoveryStrategy zero value", func(t *testing.T) {
		var strategy RecoveryStrategy
		data, err := json.Marshal(strategy)
		require.NoError(t, err)

		var unmarshaled RecoveryStrategy
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		assert.Empty(t, unmarshaled.ID)
	})

	t.Run("ValidationRule zero value", func(t *testing.T) {
		var rule ValidationRule
		data, err := json.Marshal(rule)
		require.NoError(t, err)

		var unmarshaled ValidationRule
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		assert.False(t, unmarshaled.Enabled)
	})

	t.Run("AuditTrail zero value", func(t *testing.T) {
		var audit AuditTrail
		data, err := json.Marshal(audit)
		require.NoError(t, err)

		var unmarshaled AuditTrail
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		assert.Empty(t, unmarshaled.Action)
	})
}

// Benchmark tests
func BenchmarkRecoveryStrategy_Marshal(b *testing.B) {
	strategy := RecoveryStrategy{
		ID:          "recovery-001",
		Name:        "Failover Strategy",
		Description: "Automatic failover to backup service",
		Strategy:    "failover",
		Parameters:  map[string]any{"timeout": 30},
		Priority:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(strategy)
	}
}

func BenchmarkAuditTrail_Marshal(b *testing.B) {
	audit := AuditTrail{
		ID:        "audit-001",
		Timestamp: time.Now(),
		Action:    "user.login",
		Actor:     "user@example.com",
		Resource:  "/api/v1/auth/login",
		Outcome:   "success",
		Details:   map[string]any{"method": "password"},
		Metadata:  map[string]any{"browser": "Chrome"},
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0",
		SessionID: "sess-123456",
		RequestID: "req-789012",
		TraceID:   "trace-345678",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(audit)
	}
}
