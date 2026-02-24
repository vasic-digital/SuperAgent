# Debate Approval Gates

## Overview

Approval gates provide a configurable **human-in-the-loop** mechanism for the HelixAgent debate system. When enabled, they pause the debate at specified phase boundaries and wait for external approval before continuing. When disabled (the default), all gates auto-approve instantly, adding zero overhead to the debate pipeline.

The implementation lives in `internal/debate/gates/approval_gate.go`.

## Core Concepts

### Gate Points

A gate point is a debate phase boundary where the protocol will pause and request approval. Any of the 8 debate phases can be designated as a gate point:

| Phase              | Constant                    | Description                           |
|--------------------|-----------------------------|---------------------------------------|
| Dehallucination    | `PhaseDehallucination`      | After clarification is complete       |
| Self-Evolvement    | `PhaseSelfEvolvement`       | After self-testing is complete        |
| Proposal           | `PhaseProposal`             | After initial proposals are generated |
| Critique           | `PhaseCritique`             | After critiques are gathered          |
| Review             | `PhaseReview`               | After reviews are collected           |
| Optimization       | `PhaseOptimization`         | After optimization is complete        |
| Adversarial        | `PhaseAdversarial`          | After Red/Blue Team cycle finishes    |
| Convergence        | `PhaseConvergence`          | Before final consensus is reported    |

### Gate Request Lifecycle

Each gate request transitions through these statuses:

```
pending -> approved    (external approval)
pending -> rejected    (external rejection)
pending -> timed_out   (timeout expired)
```

## Configuration

The `GateConfig` struct controls all gate behavior:

| Field                  | Type               | Default     | Description                                |
|------------------------|--------------------|-------------|--------------------------------------------|
| `Enabled`              | `bool`             | `false`     | Master switch for all gates                |
| `GatePoints`           | `[]DebatePhase`    | `nil`       | Phases where gates are checked             |
| `Timeout`              | `time.Duration`    | 30 minutes  | How long to wait before auto-timeout       |
| `NotificationChannels` | `[]string`         | `nil`       | Channels to notify when a gate is pending  |

### Default configuration

```go
config := gates.DefaultGateConfig()
// config.Enabled = false
// config.Timeout = 30 * time.Minute
```

### Enabling gates for specific phases

```go
config := gates.GateConfig{
    Enabled: true,
    GatePoints: []topology.DebatePhase{
        topology.PhaseProposal,
        topology.PhaseAdversarial,
        topology.PhaseConvergence,
    },
    Timeout: 15 * time.Minute,
}
gate := gates.NewApprovalGate(config)
```

## Auto-Approve Behavior

When gates are disabled (`Enabled: false`) or a phase is not in the `GatePoints` list, `CheckGate()` returns immediately with an auto-approved decision:

```go
decision := &GateDecision{
    RequestID: "",
    Decision:  GateStatusApproved,
    Reviewer:  "auto",
    Reason:    "gate not enabled for this phase",
    DecidedAt: time.Now(),
}
```

This ensures zero overhead when gates are not configured.

## Gate Check Flow

The `CheckGate()` method is the primary entry point, called by the debate protocol at each phase boundary:

```go
func (g *ApprovalGate) CheckGate(
    ctx       context.Context,
    debateID  string,
    sessionID string,
    phase     topology.DebatePhase,
    summary   string,
    artifacts map[string]interface{},
) (*GateDecision, error)
```

### Algorithm

```
1. IF gates disabled OR phase not in GatePoints:
   RETURN auto-approved decision

2. Create GateRequest with:
   - Unique ID: "gate-{debateID}-{phase}-{timestamp}"
   - Status: pending
   - Summary and artifacts from the phase

3. Create buffered decision channel
4. Store request and channel in internal maps

5. SELECT (blocking wait):
   a. Decision received on channel -> RETURN decision
   b. Timeout expired -> Mark request as timed_out, RETURN timeout decision
   c. Context cancelled -> Clean up, RETURN error
```

### Thread safety

All internal state is protected by `sync.RWMutex`. The decision channel is buffered (capacity 1) to prevent goroutine leaks when approval arrives after timeout.

## REST API Endpoints

### Approve a pending gate

```
POST /v1/debate/{debate_id}/gates/{request_id}/approve
```

Calls `gate.Approve(requestID, reviewer, reason)`.

### Reject a pending gate

```
POST /v1/debate/{debate_id}/gates/{request_id}/reject
```

Calls `gate.Reject(requestID, reviewer, reason)`.

### List pending gates for a debate

```
GET /v1/debate/{debate_id}/gates
```

Calls `gate.GetPendingRequests(debateID)` and returns all requests with status `pending`.

### Get a specific gate request

```
GET /v1/debate/{debate_id}/gates/{request_id}
```

Calls `gate.GetRequest(requestID)`.

## Timeout Handling

When the configured `Timeout` expires before an external decision arrives:

1. The gate request status is set to `timed_out`.
2. The decision channel is removed from the internal map.
3. A `GateDecision` with `Decision: GateStatusTimedOut` is returned.
4. The debate protocol can then decide how to handle the timeout (e.g., auto-approve, fail the debate, or retry).

## Gate Request Structure

```go
type GateRequest struct {
    ID          string                 // Unique request identifier
    DebateID    string                 // Parent debate
    SessionID   string                 // Parent session
    Phase       topology.DebatePhase   // Which phase triggered the gate
    Summary     string                 // Phase summary for reviewer
    Artifacts   map[string]interface{} // Phase outputs for inspection
    RequestedAt time.Time              // When the gate was triggered
    Status      GateRequestStatus      // pending, approved, rejected, timed_out
}
```

## Gate Decision Structure

```go
type GateDecision struct {
    RequestID string            // Which request this decides
    Decision  GateRequestStatus // approved, rejected, timed_out
    Reviewer  string            // Who made the decision (or "auto")
    Reason    string            // Why this decision was made
    DecidedAt time.Time         // When the decision was recorded
}
```

## Error Handling

- **Request not found:** `Approve()` and `Reject()` return an error if the `requestID` does not exist in the internal map.
- **Not pending:** If a request is no longer in `pending` status (e.g., already approved, rejected, or timed out), the call returns an error with the current status.
- **Context cancellation:** If the parent context is cancelled while waiting, the decision channel is cleaned up and an error wrapping `ctx.Err()` is returned.

## Database Support

Approval gate configuration is stored in the `config` JSONB column of the `debate_sessions` table. The session status supports `paused` to represent a debate waiting at a gate, enabling persistence across server restarts.

## Related Files

- `internal/debate/gates/approval_gate.go` -- Core approval gate implementation
- `internal/debate/gates/approval_gate_test.go` -- Unit tests
- `internal/debate/topology/topology.go` -- `DebatePhase` type and constants
- `sql/schema/debate_sessions.sql` -- Session status supports `paused` for gate waits
