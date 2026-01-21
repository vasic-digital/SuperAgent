# Governance Package

The governance package implements the **SEMAP Protocol** (Semantic Agent Protocol) for Design-by-Contract governance of AI agents in HelixAgent.

## Overview

SEMAP provides a formal framework for defining, enforcing, and monitoring behavioral contracts for AI agents. It enables safety guarantees, compliance validation, and remediation strategies when agents violate defined constraints.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SEMAP Governance Engine                   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   Contract Registry                    │  │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────────────┐ │  │
│  │  │Precondition│ │Postcondition│ │     Invariant      │ │  │
│  │  └────────────┘ └────────────┘ └────────────────────┘ │  │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────────────┐ │  │
│  │  │ GuardRail  │ │ Assertion  │ │  RateLimiter       │ │  │
│  │  └────────────┘ └────────────┘ └────────────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                  │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │                 Agent Profiles                         │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  │  │
│  │  │ Claude  │  │DeepSeek │  │ Gemini  │  │ Custom  │  │  │
│  │  │ Profile │  │ Profile │  │ Profile │  │ Profile │  │  │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                  │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │              Violation Handler                         │  │
│  │  Check → Detect → Log → Remediate → Report            │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### Contract

Defines a behavioral constraint for an agent.

```go
type Contract struct {
    ID          string            // Unique contract identifier
    Name        string            // Human-readable name
    Description string            // What this contract enforces
    Type        ContractType      // Precondition, Postcondition, etc.
    Condition   ContractCondition // The check function
    Severity    Severity          // Low, Medium, High, Critical
    Remediation RemediationType   // How to handle violations
    Enabled     bool              // Is contract active
    Metadata    map[string]string // Additional metadata
}

type ContractType string
const (
    ContractTypePrecondition  ContractType = "precondition"
    ContractTypePostcondition ContractType = "postcondition"
    ContractTypeInvariant     ContractType = "invariant"
    ContractTypeGuardRail     ContractType = "guardrail"
    ContractTypeAssertion     ContractType = "assertion"
)
```

### Policy

Groups related contracts into a coherent policy.

```go
type Policy struct {
    ID          string     // Policy identifier
    Name        string     // Policy name
    Description string     // What this policy enforces
    Contracts   []Contract // Associated contracts
    Enabled     bool       // Is policy active
    Priority    int        // Evaluation order
}
```

### AgentProfile

Defines an agent's identity and associated policies.

```go
type AgentProfile struct {
    ID           string            // Agent identifier
    Name         string            // Agent name
    Provider     string            // LLM provider
    Model        string            // Model identifier
    Policies     []string          // Applied policy IDs
    Capabilities []string          // Agent capabilities
    Constraints  map[string]string // Agent-specific constraints
    TrustLevel   TrustLevel        // Low, Medium, High, System
}

type TrustLevel string
const (
    TrustLevelLow    TrustLevel = "low"
    TrustLevelMedium TrustLevel = "medium"
    TrustLevelHigh   TrustLevel = "high"
    TrustLevelSystem TrustLevel = "system"
)
```

### Violation

Records a contract violation.

```go
type Violation struct {
    ID          string       // Violation ID
    ContractID  string       // Which contract was violated
    AgentID     string       // Which agent violated
    Timestamp   time.Time    // When violation occurred
    Severity    Severity     // Violation severity
    Description string       // What happened
    Context     interface{}  // Request/response context
    Remediation string       // Action taken
    Resolved    bool         // Has been handled
}
```

### CheckResult

Result of a contract evaluation.

```go
type CheckResult struct {
    ContractID string        // Contract checked
    Passed     bool          // Did it pass
    Message    string        // Result message
    Duration   time.Duration // Check duration
    Metadata   map[string]interface{}
}
```

## Predefined Contracts

HelixAgent includes these built-in security contracts:

| Contract | Type | Description |
|----------|------|-------------|
| `no_sql_injection` | GuardRail | Prevents SQL injection patterns |
| `no_path_traversal` | GuardRail | Prevents path traversal attacks |
| `no_command_injection` | GuardRail | Prevents shell command injection |
| `no_xss` | GuardRail | Prevents XSS attack patterns |
| `max_response_length` | Postcondition | Limits response size |
| `max_tokens` | Precondition | Limits token usage |
| `no_pii_exposure` | Postcondition | Prevents PII in responses |
| `no_secrets_exposure` | Postcondition | Prevents secret leakage |
| `rate_limit` | Invariant | Enforces rate limits |
| `content_policy` | GuardRail | Content safety checks |

## Usage Examples

### Create Governance Engine

```go
import "dev.helix.agent/internal/governance"

// Create engine
engine := governance.NewSEMAPEngine(governance.EngineConfig{
    EnableAudit:    true,
    StrictMode:     false,
    MaxViolations:  100,
})
```

### Define Custom Contract

```go
// Create a custom contract
contract := governance.Contract{
    ID:          "custom_length_limit",
    Name:        "Response Length Limit",
    Description: "Ensures responses don't exceed 10000 characters",
    Type:        governance.ContractTypePostcondition,
    Severity:    governance.SeverityMedium,
    Condition: func(ctx *governance.CheckContext) (bool, string) {
        response := ctx.Response.(string)
        if len(response) > 10000 {
            return false, fmt.Sprintf("Response too long: %d chars", len(response))
        }
        return true, "Response within limits"
    },
    Remediation: governance.RemediationTruncate,
    Enabled:     true,
}

engine.RegisterContract(contract)
```

### Create Policy

```go
// Create a security policy
policy := governance.Policy{
    ID:          "security_policy",
    Name:        "Security Policy",
    Description: "Core security contracts for all agents",
    Contracts: []governance.Contract{
        noSQLInjection,
        noPathTraversal,
        noCommandInjection,
    },
    Enabled:  true,
    Priority: 1,
}

engine.RegisterPolicy(policy)
```

### Create Agent Profile

```go
// Register an agent with policies
profile := governance.AgentProfile{
    ID:       "claude-agent",
    Name:     "Claude Code Assistant",
    Provider: "anthropic",
    Model:    "claude-3-opus",
    Policies: []string{"security_policy", "content_policy"},
    Capabilities: []string{
        "code_generation",
        "code_review",
        "documentation",
    },
    TrustLevel: governance.TrustLevelHigh,
}

engine.RegisterAgent(profile)
```

### Check Contracts

```go
// Before request (preconditions)
ctx := &governance.CheckContext{
    AgentID: "claude-agent",
    Request: userRequest,
    Phase:   governance.PhasePrecondition,
}

results := engine.CheckContracts(ctx)
for _, result := range results {
    if !result.Passed {
        log.Printf("Precondition failed: %s - %s",
            result.ContractID, result.Message)
        return nil, errors.New("request violates policy")
    }
}

// After response (postconditions)
ctx.Response = llmResponse
ctx.Phase = governance.PhasePostcondition

results = engine.CheckContracts(ctx)
// Handle violations...
```

### Handle Violations

```go
// Register violation handler
engine.OnViolation(func(v governance.Violation) {
    // Log violation
    log.Printf("Violation: %s by %s - %s",
        v.ContractID, v.AgentID, v.Description)

    // Alert if critical
    if v.Severity == governance.SeverityCritical {
        alerting.Send("Critical SEMAP violation", v)
    }

    // Record for audit
    audit.RecordViolation(v)
})
```

### Remediation Strategies

```go
type RemediationType string
const (
    // Block the request/response entirely
    RemediationBlock RemediationType = "block"

    // Truncate content to safe limits
    RemediationTruncate RemediationType = "truncate"

    // Sanitize problematic content
    RemediationSanitize RemediationType = "sanitize"

    // Log and allow (soft enforcement)
    RemediationWarn RemediationType = "warn"

    // Retry with modified parameters
    RemediationRetry RemediationType = "retry"

    // Escalate to human review
    RemediationEscalate RemediationType = "escalate"
)
```

### Query Violations

```go
// Get recent violations
violations := engine.GetViolations(governance.ViolationQuery{
    AgentID:   "claude-agent",
    Severity:  governance.SeverityHigh,
    Since:     time.Now().Add(-24 * time.Hour),
    Limit:     100,
})

// Get violation statistics
stats := engine.GetViolationStats()
fmt.Printf("Total: %d, Critical: %d, Resolved: %d\n",
    stats.Total, stats.Critical, stats.Resolved)
```

## Integration with HelixAgent

SEMAP is integrated into the request pipeline:

```
Request → Precondition Check → LLM Call → Postcondition Check → Response
              ↓                               ↓
          Violation?                      Violation?
              ↓                               ↓
          Remediate                       Remediate
```

### Middleware Integration

```go
// Apply SEMAP middleware
router.Use(governance.SEMAPMiddleware(engine))
```

### Debate Integration

```go
// Check debate responses
for _, response := range debateResponses {
    ctx := &governance.CheckContext{
        AgentID:  response.AgentID,
        Response: response.Content,
        Phase:    governance.PhasePostcondition,
    }
    engine.CheckContracts(ctx)
}
```

## Testing

```bash
go test -v ./internal/governance/...
```

### Testing Contracts

```go
func TestSQLInjectionContract(t *testing.T) {
    engine := governance.NewSEMAPEngine(governance.EngineConfig{})
    engine.RegisterContract(governance.PredefinedContracts["no_sql_injection"])

    testCases := []struct {
        input    string
        expected bool
    }{
        {"SELECT * FROM users", false},
        {"Hello, how are you?", true},
        {"'; DROP TABLE users;--", false},
    }

    for _, tc := range testCases {
        ctx := &governance.CheckContext{Request: tc.input}
        results := engine.CheckContracts(ctx)
        assert.Equal(t, tc.expected, results[0].Passed)
    }
}
```

## Audit Logging

SEMAP maintains comprehensive audit logs:

```go
// Enable audit logging
engine := governance.NewSEMAPEngine(governance.EngineConfig{
    EnableAudit: true,
    AuditPath:   "/var/log/helix/semap-audit.log",
})

// Audit entries include:
// - All contract checks (pass/fail)
// - All violations with context
// - Remediation actions taken
// - Agent activity summaries
```

## Performance Considerations

1. **Contract Order**: Evaluate cheaper contracts first
2. **Caching**: Cache contract compilation where possible
3. **Async Checks**: Non-critical checks can run asynchronously
4. **Sampling**: For high-volume, sample-based checking option
