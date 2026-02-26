package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MonitoringConfig holds configuration for monitoring
type MonitoringConfig struct {
	CheckInterval     time.Duration
	AlertThreshold    int
	HealthCheckPeriod time.Duration
}

// ExtendedDebateStatus extends DebateStatus with monitoring-specific fields
type ExtendedDebateStatus struct {
	DebateStatus
	LastUpdateTime time.Time `json:"last_update_time"`
	HealthScore    float64   `json:"health_score"`
	ErrorCount     int       `json:"error_count"`
	WarningCount   int       `json:"warning_count"`
}

// MonitoringSession represents an active monitoring session
type MonitoringSession struct {
	ID            string
	DebateID      string
	Config        *DebateConfig
	Status        *ExtendedDebateStatus
	StartTime     time.Time
	LastCheck     time.Time
	Active        bool
	Alerts        []MonitoringAlert
	CancelFunc    context.CancelFunc
	monitoringCtx context.Context
}

// MonitoringAlert represents an alert generated during monitoring
type MonitoringAlert struct {
	ID         string    `json:"id"`
	DebateID   string    `json:"debate_id"`
	Level      string    `json:"level"` // info, warning, error, critical
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	Resolved   bool      `json:"resolved"`
	ResolvedAt time.Time `json:"resolved_at,omitempty"`
}

// DebateMonitoringService provides monitoring capabilities
type DebateMonitoringService struct {
	logger     *logrus.Logger
	sessions   map[string]*MonitoringSession
	sessionsMu sync.RWMutex
	config     *MonitoringConfig
}

// NewDebateMonitoringService creates a new monitoring service
func NewDebateMonitoringService(logger *logrus.Logger) *DebateMonitoringService {
	return &DebateMonitoringService{
		logger:   logger,
		sessions: make(map[string]*MonitoringSession),
		config: &MonitoringConfig{
			CheckInterval:     time.Second * 5,
			AlertThreshold:    3,
			HealthCheckPeriod: time.Second * 30,
		},
	}
}

// NewDebateMonitoringServiceWithConfig creates a monitoring service with custom config
func NewDebateMonitoringServiceWithConfig(logger *logrus.Logger, config *MonitoringConfig) *DebateMonitoringService {
	if config == nil {
		config = &MonitoringConfig{
			CheckInterval:     time.Second * 5,
			AlertThreshold:    3,
			HealthCheckPeriod: time.Second * 30,
		}
	}
	return &DebateMonitoringService{
		logger:   logger,
		sessions: make(map[string]*MonitoringSession),
		config:   config,
	}
}

// StartMonitoring starts monitoring for a debate
func (dms *DebateMonitoringService) StartMonitoring(ctx context.Context, config *DebateConfig) (string, error) {
	if config == nil {
		return "", fmt.Errorf("debate config is required")
	}

	monitoringID := "mon-" + uuid.New().String()[:8]

	// Create cancellable context for this monitoring session
	monitoringCtx, cancelFunc := context.WithCancel(ctx)

	session := &MonitoringSession{
		ID:       monitoringID,
		DebateID: config.DebateID,
		Config:   config,
		Status: &ExtendedDebateStatus{
			DebateStatus: DebateStatus{
				DebateID:     config.DebateID,
				Status:       "pending",
				CurrentRound: 0,
				TotalRounds:  config.MaxRounds,
				StartTime:    time.Now(),
				Participants: make([]ParticipantStatus, 0),
			},
			HealthScore: 100.0,
		},
		StartTime:     time.Now(),
		LastCheck:     time.Now(),
		Active:        true,
		Alerts:        make([]MonitoringAlert, 0),
		CancelFunc:    cancelFunc,
		monitoringCtx: monitoringCtx,
	}

	// Initialize participant status
	for _, participant := range config.Participants {
		session.Status.Participants = append(session.Status.Participants, ParticipantStatus{
			ParticipantID:   participant.ParticipantID,
			ParticipantName: participant.Name,
			Status:          "pending",
		})
	}

	dms.sessionsMu.Lock()
	dms.sessions[monitoringID] = session
	dms.sessionsMu.Unlock()

	// Start background monitoring goroutine
	go dms.runMonitoringLoop(session)

	dms.logger.WithFields(logrus.Fields{
		"monitoring_id": monitoringID,
		"debate_id":     config.DebateID,
	}).Info("Started monitoring for debate")

	return monitoringID, nil
}

// runMonitoringLoop runs the background monitoring loop
func (dms *DebateMonitoringService) runMonitoringLoop(session *MonitoringSession) {
	ticker := time.NewTicker(dms.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-session.monitoringCtx.Done():
			dms.logger.Infof("Monitoring loop stopped for %s", session.ID)
			return
		case <-ticker.C:
			dms.performHealthCheck(session)
		}
	}
}

// performHealthCheck performs a health check on the monitoring session
func (dms *DebateMonitoringService) performHealthCheck(session *MonitoringSession) {
	dms.sessionsMu.Lock()
	defer dms.sessionsMu.Unlock()

	if !session.Active {
		return
	}

	session.LastCheck = time.Now()

	// Update health score based on status
	healthScore := 100.0

	// Reduce health for errors
	if session.Status.ErrorCount > 0 {
		healthScore -= float64(session.Status.ErrorCount) * 10
	}

	// Reduce health for warnings
	if session.Status.WarningCount > 0 {
		healthScore -= float64(session.Status.WarningCount) * 5
	}

	// Check for stale participants (based on status)
	for i := range session.Status.Participants {
		p := &session.Status.Participants[i]
		if p.Status == "active" && p.ResponseTime == 0 {
			// Mark as potentially stale if no response recorded
			healthScore -= 5
		}
	}

	// Ensure health score is within bounds
	if healthScore < 0 {
		healthScore = 0
	}
	session.Status.HealthScore = healthScore

	// Generate critical alert if health is too low
	if healthScore < 50 && session.Status.Status != "failed" {
		dms.addAlert(session, "critical", "Debate health is critically low")
	}
}

// addAlert adds an alert to the monitoring session
func (dms *DebateMonitoringService) addAlert(session *MonitoringSession, level, message string) {
	alert := MonitoringAlert{
		ID:        "alert-" + uuid.New().String()[:8],
		DebateID:  session.DebateID,
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
	}

	session.Alerts = append(session.Alerts, alert)

	dms.logger.WithFields(logrus.Fields{
		"alert_id":  alert.ID,
		"debate_id": alert.DebateID,
		"level":     level,
		"message":   message,
	}).Warn("Monitoring alert generated")
}

// StopMonitoring stops monitoring for a debate
func (dms *DebateMonitoringService) StopMonitoring(ctx context.Context, monitoringID string) error {
	dms.sessionsMu.Lock()
	defer dms.sessionsMu.Unlock()

	session, exists := dms.sessions[monitoringID]
	if !exists {
		return fmt.Errorf("monitoring session not found: %s", monitoringID)
	}

	session.Active = false
	if session.CancelFunc != nil {
		session.CancelFunc()
	}

	dms.logger.Infof("Stopped monitoring %s for debate %s", monitoringID, session.DebateID)
	return nil
}

// GetStatus retrieves the current status of a debate
func (dms *DebateMonitoringService) GetStatus(ctx context.Context, debateID string) (*DebateStatus, error) {
	dms.sessionsMu.RLock()
	defer dms.sessionsMu.RUnlock()

	// Find session by debate ID
	for _, session := range dms.sessions {
		if session.DebateID == debateID {
			// Return a copy of the embedded DebateStatus
			statusCopy := session.Status.DebateStatus
			return &statusCopy, nil
		}
	}

	return nil, fmt.Errorf("no monitoring session found for debate: %s", debateID)
}

// GetStatusByMonitoringID retrieves status by monitoring ID
func (dms *DebateMonitoringService) GetStatusByMonitoringID(ctx context.Context, monitoringID string) (*DebateStatus, error) {
	dms.sessionsMu.RLock()
	defer dms.sessionsMu.RUnlock()

	session, exists := dms.sessions[monitoringID]
	if !exists {
		return nil, fmt.Errorf("monitoring session not found: %s", monitoringID)
	}

	// Return a copy of the embedded DebateStatus
	statusCopy := session.Status.DebateStatus
	return &statusCopy, nil
}

// GetExtendedStatus retrieves the full extended status including health metrics
func (dms *DebateMonitoringService) GetExtendedStatus(ctx context.Context, monitoringID string) (*ExtendedDebateStatus, error) {
	dms.sessionsMu.RLock()
	defer dms.sessionsMu.RUnlock()

	session, exists := dms.sessions[monitoringID]
	if !exists {
		return nil, fmt.Errorf("monitoring session not found: %s", monitoringID)
	}

	// Return a copy
	statusCopy := *session.Status
	return &statusCopy, nil
}

// UpdateParticipantStatus updates the status of a participant
func (dms *DebateMonitoringService) UpdateParticipantStatus(
	ctx context.Context,
	monitoringID string,
	participantID string,
	status string,
	responseTime time.Duration,
) error {
	dms.sessionsMu.Lock()
	defer dms.sessionsMu.Unlock()

	session, exists := dms.sessions[monitoringID]
	if !exists {
		return fmt.Errorf("monitoring session not found: %s", monitoringID)
	}

	for i := range session.Status.Participants {
		p := &session.Status.Participants[i]
		if p.ParticipantID == participantID {
			p.Status = status
			p.ResponseTime = responseTime
			session.Status.LastUpdateTime = time.Now()
			return nil
		}
	}

	return fmt.Errorf("participant not found: %s", participantID)
}

// UpdateRound updates the current round of a debate
func (dms *DebateMonitoringService) UpdateRound(ctx context.Context, monitoringID string, round int) error {
	dms.sessionsMu.Lock()
	defer dms.sessionsMu.Unlock()

	session, exists := dms.sessions[monitoringID]
	if !exists {
		return fmt.Errorf("monitoring session not found: %s", monitoringID)
	}

	session.Status.CurrentRound = round
	session.Status.LastUpdateTime = time.Now()

	// Update debate status based on round
	if round > 0 && session.Status.Status == "pending" {
		session.Status.Status = "active"
	}

	if round >= session.Status.TotalRounds {
		session.Status.Status = "completed"
	}

	return nil
}

// RecordError records an error during debate
func (dms *DebateMonitoringService) RecordError(ctx context.Context, monitoringID string, errMsg string) error {
	dms.sessionsMu.Lock()
	defer dms.sessionsMu.Unlock()

	session, exists := dms.sessions[monitoringID]
	if !exists {
		return fmt.Errorf("monitoring session not found: %s", monitoringID)
	}

	session.Status.ErrorCount++
	dms.addAlert(session, "error", errMsg)

	if session.Status.ErrorCount >= dms.config.AlertThreshold {
		session.Status.Status = "failed"
		dms.addAlert(session, "critical", "Debate failed due to too many errors")
	}

	return nil
}

// GetAlerts retrieves alerts for a monitoring session
func (dms *DebateMonitoringService) GetAlerts(ctx context.Context, monitoringID string) ([]MonitoringAlert, error) {
	dms.sessionsMu.RLock()
	defer dms.sessionsMu.RUnlock()

	session, exists := dms.sessions[monitoringID]
	if !exists {
		return nil, fmt.Errorf("monitoring session not found: %s", monitoringID)
	}

	// Return a copy of alerts
	alerts := make([]MonitoringAlert, len(session.Alerts))
	copy(alerts, session.Alerts)
	return alerts, nil
}

// ListActiveSessions returns all active monitoring session IDs
func (dms *DebateMonitoringService) ListActiveSessions() []string {
	dms.sessionsMu.RLock()
	defer dms.sessionsMu.RUnlock()

	ids := make([]string, 0)
	for id, session := range dms.sessions {
		if session.Active {
			ids = append(ids, id)
		}
	}
	return ids
}

// CleanupInactiveSessions removes inactive sessions older than maxAge
func (dms *DebateMonitoringService) CleanupInactiveSessions(maxAge time.Duration) int {
	dms.sessionsMu.Lock()
	defer dms.sessionsMu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, session := range dms.sessions {
		if !session.Active && session.LastCheck.Before(cutoff) {
			delete(dms.sessions, id)
			removed++
		}
	}

	if removed > 0 {
		dms.logger.Infof("Cleaned up %d inactive monitoring sessions", removed)
	}

	return removed
}

// GetStats returns monitoring service statistics
func (dms *DebateMonitoringService) GetStats() map[string]interface{} {
	dms.sessionsMu.RLock()
	defer dms.sessionsMu.RUnlock()

	stats := map[string]interface{}{
		"total_sessions":  len(dms.sessions),
		"active_sessions": 0,
		"total_alerts":    0,
		"critical_alerts": 0,
		"error_alerts":    0,
		"warning_alerts":  0,
	}

	for _, session := range dms.sessions {
		if session.Active {
			stats["active_sessions"] = stats["active_sessions"].(int) + 1 //nolint:errcheck
		}
		for _, alert := range session.Alerts {
			stats["total_alerts"] = stats["total_alerts"].(int) + 1
			switch alert.Level {
			case "critical":
				stats["critical_alerts"] = stats["critical_alerts"].(int) + 1
			case "error":
				stats["error_alerts"] = stats["error_alerts"].(int) + 1
			case "warning":
				stats["warning_alerts"] = stats["warning_alerts"].(int) + 1
			}
		}
	}

	return stats
}
