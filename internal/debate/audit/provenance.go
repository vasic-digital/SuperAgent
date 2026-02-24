// Package audit provides provenance tracking and audit trail capabilities
// for debate reproducibility. It records all events that occur during debate
// sessions, enabling full replay and analysis of debate execution.
package audit

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// EventType identifies the type of audit event.
type EventType string

const (
	EventPromptSent          EventType = "prompt_sent"
	EventResponseReceived    EventType = "response_received"
	EventToolCalled          EventType = "tool_called"
	EventVoteCast            EventType = "vote_cast"
	EventGateDecision        EventType = "gate_decision"
	EventReflectionGenerated EventType = "reflection_generated"
	EventPhaseStarted        EventType = "phase_started"
	EventPhaseCompleted      EventType = "phase_completed"
	EventRoundStarted        EventType = "round_started"
	EventRoundCompleted      EventType = "round_completed"
	EventDebateStarted       EventType = "debate_started"
	EventDebateCompleted     EventType = "debate_completed"
	EventErrorOccurred       EventType = "error_occurred"
	EventConfigChanged       EventType = "config_changed"
)

// AuditEntry represents a single audit event in the provenance trail.
type AuditEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType EventType              `json:"event_type"`
	SessionID string                 `json:"session_id"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Phase     string                 `json:"phase,omitempty"`
	Round     int                    `json:"round,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// AuditSummary provides a high-level overview of a debate session.
type AuditSummary struct {
	SessionID        string        `json:"session_id"`
	TotalPrompts     int           `json:"total_prompts"`
	TotalResponses   int           `json:"total_responses"`
	TotalToolCalls   int           `json:"total_tool_calls"`
	TotalVotes       int           `json:"total_votes"`
	TotalReflections int           `json:"total_reflections"`
	TotalErrors      int           `json:"total_errors"`
	ModelsUsed       []string      `json:"models_used"`
	ProvidersUsed    []string      `json:"providers_used"`
	PhasesExecuted   []string      `json:"phases_executed"`
	Duration         time.Duration `json:"duration"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
}

// AuditTrail represents the complete audit trail for a session.
type AuditTrail struct {
	SessionID string        `json:"session_id"`
	Entries   []*AuditEntry `json:"entries"`
	Summary   *AuditSummary `json:"summary"`
}

// ProvenanceTracker tracks all events in debate sessions for audit
// and reproducibility. It is safe for concurrent use.
type ProvenanceTracker struct {
	entries map[string][]*AuditEntry // keyed by session ID
	counter int64
	mu      sync.RWMutex
}

// NewProvenanceTracker creates a new ProvenanceTracker ready for use.
func NewProvenanceTracker() *ProvenanceTracker {
	return &ProvenanceTracker{
		entries: make(map[string][]*AuditEntry),
	}
}

// Record adds an audit entry to the specified session. If the entry's ID
// is empty, an auto-generated ID is assigned. If the entry's Timestamp is
// zero, the current time is used.
func (t *ProvenanceTracker) Record(sessionID string, entry *AuditEntry) {
	if entry == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if entry.ID == "" {
		t.counter++
		entry.ID = fmt.Sprintf("audit-%d", t.counter)
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	entry.SessionID = sessionID
	t.entries[sessionID] = append(t.entries[sessionID], entry)
}

// GetAuditTrail returns the complete audit trail for a session, including
// entries sorted by timestamp and a computed summary. Returns nil if the
// session does not exist.
func (t *ProvenanceTracker) GetAuditTrail(sessionID string) *AuditTrail {
	entries := t.GetEntries(sessionID)
	if entries == nil {
		return nil
	}

	return &AuditTrail{
		SessionID: sessionID,
		Entries:   entries,
		Summary:   t.GetSummary(sessionID),
	}
}

// GetEntries returns all audit entries for a session, sorted by timestamp.
// Returns nil if the session does not exist.
func (t *ProvenanceTracker) GetEntries(sessionID string) []*AuditEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	raw, ok := t.entries[sessionID]
	if !ok {
		return nil
	}

	// Return a copy sorted by timestamp
	result := make([]*AuditEntry, len(raw))
	copy(result, raw)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result
}

// GetEntriesByType returns all audit entries of the specified type for a
// session, sorted by timestamp. Returns nil if the session does not exist.
func (t *ProvenanceTracker) GetEntriesByType(
	sessionID string,
	eventType EventType,
) []*AuditEntry {
	entries := t.GetEntries(sessionID)
	if entries == nil {
		return nil
	}

	var filtered []*AuditEntry
	for _, e := range entries {
		if e.EventType == eventType {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

// GetEntriesByAgent returns all audit entries for a specific agent in
// a session, sorted by timestamp. Returns nil if the session does not exist.
func (t *ProvenanceTracker) GetEntriesByAgent(
	sessionID string,
	agentID string,
) []*AuditEntry {
	entries := t.GetEntries(sessionID)
	if entries == nil {
		return nil
	}

	var filtered []*AuditEntry
	for _, e := range entries {
		if e.AgentID == agentID {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

// GetSummary computes and returns an AuditSummary for the specified session.
// It counts events by type, extracts unique models, providers, and phases
// from entry data, and calculates timing information. Returns nil if the
// session does not exist.
func (t *ProvenanceTracker) GetSummary(sessionID string) *AuditSummary {
	entries := t.GetEntries(sessionID)
	if entries == nil {
		return nil
	}

	summary := &AuditSummary{
		SessionID: sessionID,
	}

	modelsSet := make(map[string]struct{})
	providersSet := make(map[string]struct{})
	phasesSet := make(map[string]struct{})

	var earliest, latest time.Time

	for _, e := range entries {
		// Track timing
		if earliest.IsZero() || e.Timestamp.Before(earliest) {
			earliest = e.Timestamp
		}
		if latest.IsZero() || e.Timestamp.After(latest) {
			latest = e.Timestamp
		}

		// Count by type and extract metadata
		switch e.EventType {
		case EventPromptSent:
			summary.TotalPrompts++
			extractStringField(e.Data, "model", modelsSet)
			extractStringField(e.Data, "provider", providersSet)
		case EventResponseReceived:
			summary.TotalResponses++
		case EventToolCalled:
			summary.TotalToolCalls++
		case EventVoteCast:
			summary.TotalVotes++
		case EventReflectionGenerated:
			summary.TotalReflections++
		case EventErrorOccurred:
			summary.TotalErrors++
		case EventPhaseStarted, EventPhaseCompleted:
			if e.Phase != "" {
				phasesSet[e.Phase] = struct{}{}
			}
			extractStringField(e.Data, "phase", phasesSet)
		}
	}

	summary.StartTime = earliest
	summary.EndTime = latest
	if !earliest.IsZero() && !latest.IsZero() {
		summary.Duration = latest.Sub(earliest)
	}

	summary.ModelsUsed = sortedKeys(modelsSet)
	summary.ProvidersUsed = sortedKeys(providersSet)
	summary.PhasesExecuted = sortedKeys(phasesSet)

	return summary
}

// MarshalSessionJSON creates an AuditTrail for the specified session and
// marshals it to JSON. Returns an error if the session does not exist or
// if marshaling fails.
func (t *ProvenanceTracker) MarshalSessionJSON(
	sessionID string,
) ([]byte, error) {
	trail := t.GetAuditTrail(sessionID)
	if trail == nil {
		return nil, fmt.Errorf("session %q not found", sessionID)
	}

	return json.Marshal(trail)
}

// Clear removes all audit entries for the specified session.
func (t *ProvenanceTracker) Clear(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.entries, sessionID)
}

// GetSessionIDs returns a sorted list of all tracked session IDs.
func (t *ProvenanceTracker) GetSessionIDs() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ids := make([]string, 0, len(t.entries))
	for id := range t.entries {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	return ids
}

// extractStringField extracts a string value from data by key and adds it
// to the target set if present and non-empty.
func extractStringField(
	data map[string]interface{},
	key string,
	target map[string]struct{},
) {
	if data == nil {
		return
	}

	if val, ok := data[key]; ok {
		if s, ok := val.(string); ok && s != "" {
			target[s] = struct{}{}
		}
	}
}

// sortedKeys returns the keys of a set as a sorted string slice.
func sortedKeys(m map[string]struct{}) []string {
	if len(m) == 0 {
		return []string{}
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
