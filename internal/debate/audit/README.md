# debate/audit - Provenance and Audit Trail

Provides full reproducibility tracking for AI debate sessions with event logging, session summaries, and JSON export capabilities.

## Purpose

The audit package records every significant event during a debate session, enabling post-hoc analysis, compliance auditing, and reproducibility. It tracks 14 event types covering the entire debate lifecycle from initialization through convergence.

## Key Types

### ProvenanceTracker

The main audit trail recorder that captures debate events chronologically.

```go
tracker := audit.NewProvenanceTracker(sessionID, logger)
tracker.RecordEvent(audit.EventProposal, participantID, data)
summary := tracker.GetSessionSummary()
exported, _ := tracker.ExportJSON()
```

### Event Types

| Event Type | Description |
|------------|-------------|
| `EventInitialization` | Debate session started |
| `EventProposal` | Participant submitted a proposal |
| `EventCritique` | Participant critiqued another's proposal |
| `EventReview` | Review phase completed |
| `EventOptimization` | Optimization pass applied |
| `EventAdversarial` | Red/blue team attack-defend cycle |
| `EventConvergence` | Consensus reached |
| `EventVote` | Voting round completed |
| `EventDehallucination` | Hallucination detection pass |
| `EventSelfEvolvement` | Self-improvement iteration |
| `EventGateApproval` | Human approval gate decision |
| `EventGateRejection` | Human rejection gate decision |
| `EventError` | Error during debate |
| `EventCompletion` | Debate session completed |

### SessionSummary

Aggregated summary with participant statistics, event counts, timeline, and outcomes.

## Usage within Debate System

The provenance tracker is initialized by the debate orchestrator at session start and receives events from each phase of the 8-phase protocol. The exported JSON can be stored in PostgreSQL via the debate persistence layer.

## Files

- `provenance.go` -- ProvenanceTracker, event types, session summary, JSON export
- `provenance_test.go` -- Unit tests
