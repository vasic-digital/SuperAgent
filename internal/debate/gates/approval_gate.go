// Package gates provides configurable human-in-the-loop approval gates for
// the debate system. Gates are disabled by default and auto-approve. When
// enabled, they pause the debate at configured phase points and wait for
// external approval via Approve/Reject calls.
package gates

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dev.helix.agent/internal/debate/topology"
)

// GateRequestStatus represents the status of an approval gate request.
type GateRequestStatus string

const (
	GateStatusPending  GateRequestStatus = "pending"
	GateStatusApproved GateRequestStatus = "approved"
	GateStatusRejected GateRequestStatus = "rejected"
	GateStatusTimedOut GateRequestStatus = "timed_out"
)

// GateConfig configures approval gate behavior.
type GateConfig struct {
	Enabled              bool
	GatePoints           []topology.DebatePhase
	Timeout              time.Duration
	NotificationChannels []string
}

// GateRequest represents a pending approval request.
type GateRequest struct {
	ID          string                 `json:"id"`
	DebateID    string                 `json:"debate_id"`
	SessionID   string                 `json:"session_id"`
	Phase       topology.DebatePhase   `json:"phase"`
	Summary     string                 `json:"summary"`
	Artifacts   map[string]interface{} `json:"artifacts,omitempty"`
	RequestedAt time.Time              `json:"requested_at"`
	Status      GateRequestStatus      `json:"status"`
}

// GateDecision represents a decision on an approval request.
type GateDecision struct {
	RequestID string            `json:"request_id"`
	Decision  GateRequestStatus `json:"decision"`
	Reviewer  string            `json:"reviewer"`
	Reason    string            `json:"reason"`
	DecidedAt time.Time         `json:"decided_at"`
}

// ApprovalGate manages approval gates for debate phases.
type ApprovalGate struct {
	config    GateConfig
	requests  map[string]*GateRequest
	decisions map[string]chan *GateDecision
	mu        sync.RWMutex
}

// DefaultGateConfig returns a GateConfig with gates disabled, no gate
// points, and a 30-minute timeout.
func DefaultGateConfig() GateConfig {
	return GateConfig{
		Enabled:              false,
		GatePoints:           nil,
		Timeout:              30 * time.Minute,
		NotificationChannels: nil,
	}
}

// NewApprovalGate creates a new ApprovalGate with the given configuration.
func NewApprovalGate(config GateConfig) *ApprovalGate {
	return &ApprovalGate{
		config:    config,
		requests:  make(map[string]*GateRequest),
		decisions: make(map[string]chan *GateDecision),
	}
}

// CheckGate evaluates whether a gate applies at the given phase. If gates
// are disabled or the phase is not a configured gate point, an auto-approved
// decision is returned immediately. Otherwise the call blocks until an
// external Approve/Reject is received, the configured timeout elapses, or
// the context is cancelled.
func (g *ApprovalGate) CheckGate(
	ctx context.Context,
	debateID, sessionID string,
	phase topology.DebatePhase,
	summary string,
	artifacts map[string]interface{},
) (*GateDecision, error) {
	// Auto-approve when gates are disabled or this phase is not a gate point.
	if !g.config.Enabled || !g.isGatePoint(phase) {
		return &GateDecision{
			RequestID: "",
			Decision:  GateStatusApproved,
			Reviewer:  "auto",
			Reason:    "gate not enabled for this phase",
			DecidedAt: time.Now(),
		}, nil
	}

	requestID := fmt.Sprintf(
		"gate-%s-%s-%d", debateID, phase, time.Now().UnixNano(),
	)

	request := &GateRequest{
		ID:          requestID,
		DebateID:    debateID,
		SessionID:   sessionID,
		Phase:       phase,
		Summary:     summary,
		Artifacts:   artifacts,
		RequestedAt: time.Now(),
		Status:      GateStatusPending,
	}

	decisionCh := make(chan *GateDecision, 1)

	g.mu.Lock()
	g.requests[requestID] = request
	g.decisions[requestID] = decisionCh
	g.mu.Unlock()

	// Wait for a decision, timeout, or context cancellation.
	select {
	case decision := <-decisionCh:
		return decision, nil

	case <-time.After(g.config.Timeout):
		g.mu.Lock()
		if req, ok := g.requests[requestID]; ok {
			req.Status = GateStatusTimedOut
		}
		delete(g.decisions, requestID)
		g.mu.Unlock()

		return &GateDecision{
			RequestID: requestID,
			Decision:  GateStatusTimedOut,
			Reviewer:  "",
			Reason:    "approval gate timed out",
			DecidedAt: time.Now(),
		}, nil

	case <-ctx.Done():
		g.mu.Lock()
		delete(g.decisions, requestID)
		g.mu.Unlock()

		return nil, fmt.Errorf(
			"approval gate cancelled for request %s: %w",
			requestID, ctx.Err(),
		)
	}
}

// Approve marks a pending request as approved and unblocks the waiting
// CheckGate call. Returns an error if the request does not exist or is
// not in pending status.
func (g *ApprovalGate) Approve(requestID, reviewer, reason string) error {
	return g.submitDecision(requestID, GateStatusApproved, reviewer, reason)
}

// Reject marks a pending request as rejected and unblocks the waiting
// CheckGate call. Returns an error if the request does not exist or is
// not in pending status.
func (g *ApprovalGate) Reject(requestID, reviewer, reason string) error {
	return g.submitDecision(requestID, GateStatusRejected, reviewer, reason)
}

// GetPendingRequests returns all pending requests for the given debate.
func (g *ApprovalGate) GetPendingRequests(debateID string) []*GateRequest {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var pending []*GateRequest
	for _, req := range g.requests {
		if req.DebateID == debateID && req.Status == GateStatusPending {
			pending = append(pending, req)
		}
	}
	return pending
}

// GetRequest returns a specific request by ID along with a boolean
// indicating whether it was found.
func (g *ApprovalGate) GetRequest(requestID string) (*GateRequest, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	req, ok := g.requests[requestID]
	return req, ok
}

// IsEnabled reports whether the approval gate is enabled.
func (g *ApprovalGate) IsEnabled() bool {
	return g.config.Enabled
}

// isGatePoint checks whether the given phase is in the configured list of
// gate points.
func (g *ApprovalGate) isGatePoint(phase topology.DebatePhase) bool {
	for _, gp := range g.config.GatePoints {
		if gp == phase {
			return true
		}
	}
	return false
}

// submitDecision is the shared implementation for Approve and Reject.
func (g *ApprovalGate) submitDecision(
	requestID string,
	status GateRequestStatus,
	reviewer, reason string,
) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	req, ok := g.requests[requestID]
	if !ok {
		return fmt.Errorf("approval gate request %s not found", requestID)
	}
	if req.Status != GateStatusPending {
		return fmt.Errorf(
			"approval gate request %s is not pending (status: %s)",
			requestID, req.Status,
		)
	}

	decision := &GateDecision{
		RequestID: requestID,
		Decision:  status,
		Reviewer:  reviewer,
		Reason:    reason,
		DecidedAt: time.Now(),
	}

	ch, hasCh := g.decisions[requestID]
	if hasCh {
		ch <- decision
		delete(g.decisions, requestID)
	}

	req.Status = status
	return nil
}
