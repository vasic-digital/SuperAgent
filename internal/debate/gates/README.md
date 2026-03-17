# debate/gates - Human-in-the-Loop Approval Gates

Provides configurable human approval gates for the debate system with REST API endpoints for approve/reject decisions.

## Purpose

The gates package implements human-in-the-loop checkpoints where debate progress can be paused for human review before continuing to the next phase. Gates are disabled by default and can be configured per debate session.

## Key Types

### ApprovalGate

Manages gate state, pending approvals, and REST API handlers.

```go
gate := gates.NewApprovalGate(config, logger)
gate.RequestApproval(ctx, sessionID, phaseID, proposal)

// REST API handlers
gate.HandleApprove(ctx, sessionID, phaseID)
gate.HandleReject(ctx, sessionID, phaseID, reason)
gate.HandleListGates(ctx)
```

### Gate Configuration

```go
type GateConfig struct {
    Enabled       bool          // Enable gates (default: false)
    Timeout       time.Duration // Auto-approve after timeout
    RequiredGates []string      // Which phases require approval
    WebhookURL    string        // Notify on pending approval
}
```

### Gate States

| State | Description |
|-------|-------------|
| `Pending` | Awaiting human decision |
| `Approved` | Human approved, debate continues |
| `Rejected` | Human rejected, debate halts or revises |
| `TimedOut` | Auto-approved after timeout |

## REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/debates/:id/gates/:phase/approve` | Approve a gate |
| `POST` | `/v1/debates/:id/gates/:phase/reject` | Reject a gate |
| `GET` | `/v1/debates/:id/gates` | List all gates for a session |

## Usage within Debate System

Gates are checked between debate phases by the orchestrator. When a gate is configured for a phase, the orchestrator pauses execution and waits for human approval via the REST API or until the timeout expires.

## Files

- `approval_gate.go` -- ApprovalGate, configuration, REST handlers, state management
- `approval_gate_test.go` -- Unit tests
