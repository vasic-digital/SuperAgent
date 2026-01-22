# Common Services Package

This package provides shared type definitions used across the services layer for recovery, validation, authentication, and audit functionality.

## Overview

The Common package defines reusable types for cross-cutting concerns like error recovery, validation, authentication management, and audit trails that are shared across multiple services.

## Types

### Recovery Types

```go
type RecoveryStrategy struct {
    ID          string         // Unique identifier
    Name        string         // Strategy name
    Description string         // Human-readable description
    Strategy    string         // Strategy type (retry, fallback, circuit-breaker)
    Parameters  map[string]any // Strategy-specific parameters
    Priority    int            // Execution priority
}

type RecoveryProcedure struct {
    ID         string         // Unique identifier
    Name       string         // Procedure name
    Steps      []RecoveryStep // Ordered recovery steps
    RetryCount int            // Maximum retry attempts
}

type RecoveryStep struct {
    ID        string        // Step identifier
    Action    string        // Action to execute
    Timeout   time.Duration // Step timeout
    Retryable bool          // Whether step can be retried
}
```

### Validation Types

```go
type ValidationRule struct {
    ID       string // Rule identifier
    Name     string // Rule name
    Type     string // Validation type (schema, business, format)
    Severity string // Error severity (error, warning, info)
    Enabled  bool   // Whether rule is active
}

type IntegrityCheck struct {
    ID        string    // Check identifier
    Algorithm string    // Hash/check algorithm
    Schedule  string    // Cron schedule expression
    LastRun   time.Time // Last execution time
    Status    string    // Check status
}
```

### Authentication Types

```go
type AuthenticationManager struct {
    AuthMethods        map[string]any // Supported auth methods
    AuthProviders      map[string]any // Identity providers
    CredentialManagers map[string]any // Credential handlers
    MFAProviders       []any          // MFA providers
}

type AuthenticationProvider struct {
    Type              string         // Provider type (oauth2, saml, ldap)
    TokenManagers     map[string]any // Token handlers
    IdentityProviders map[string]any // Identity sources
}

type AccessController struct {
    AccessRules       []any          // Access control rules
    PermissionMatrix  map[string]any // Permission mappings
    RoleHierarchy     map[string]any // Role inheritance
    ResourceHierarchy map[string]any // Resource tree
}

type PermissionManager struct {
    Permissions      map[string]any // Permission definitions
    Roles            map[string]any // Role definitions
    PermissionMatrix map[string]any // Role-permission mappings
}
```

### Audit Types

```go
type AuditTrail struct {
    ID        string    // Audit entry ID
    Timestamp time.Time // Event timestamp
    Action    string    // Action performed
    Actor     string    // User/system performing action
    Resource  string    // Affected resource
    Outcome   string    // Success/failure
    IPAddress string    // Client IP
    SessionID string    // Session identifier
    TraceID   string    // Distributed trace ID
}
```

### Utility Types

```go
type DateRange struct {
    StartDate time.Time
    EndDate   time.Time
    Timezone  string
    Inclusive bool
}

type ValidationResult struct {
    Valid    bool
    RuleID   string
    Message  string
    Severity string
}

type RecoveryResult struct {
    Success     bool
    ProcedureID string
    Duration    time.Duration
}

type AuthenticationResult struct {
    Success     bool
    Token       string
    ExpiresAt   time.Time
    Roles       []string
    Permissions []string
}
```

## Usage

### Defining Recovery Strategy

```go
import "dev.helix.agent/internal/services/common"

strategy := &common.RecoveryStrategy{
    ID:       "retry-with-backoff",
    Name:     "Exponential Backoff Retry",
    Strategy: "retry",
    Parameters: map[string]any{
        "max_attempts": 3,
        "base_delay":   "100ms",
        "max_delay":    "10s",
    },
    Priority: 1,
}
```

### Creating Audit Trail

```go
audit := &common.AuditTrail{
    ID:        uuid.New().String(),
    Timestamp: time.Now(),
    Action:    "user.login",
    Actor:     "user@example.com",
    Outcome:   "success",
    IPAddress: clientIP,
    SessionID: sessionID,
}
```

### Validation Rules

```go
rule := &common.ValidationRule{
    ID:       "email-format",
    Name:     "Email Format Validation",
    Type:     "format",
    Severity: "error",
    Enabled:  true,
}
```

## Testing

```bash
go test -v ./internal/services/common/...
```

## Files

- `types.go` - Type definitions for recovery, validation, auth, and audit
