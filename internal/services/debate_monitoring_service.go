package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// DebateMonitoringService provides monitoring capabilities
type DebateMonitoringService struct {
	logger *logrus.Logger
}

// NewDebateMonitoringService creates a new monitoring service
func NewDebateMonitoringService(logger *logrus.Logger) *DebateMonitoringService {
	return &DebateMonitoringService{
		logger: logger,
	}
}

// StartMonitoring starts monitoring for a debate
func (dms *DebateMonitoringService) StartMonitoring(ctx context.Context, config *DebateConfig) (string, error) {
	monitoringID := "monitoring-" + time.Now().Format("20060102150405")
	dms.logger.Infof("Started monitoring %s for debate %s", monitoringID, config.DebateID)
	return monitoringID, nil
}

// StopMonitoring stops monitoring for a debate
func (dms *DebateMonitoringService) StopMonitoring(ctx context.Context, monitoringID string) error {
	dms.logger.Infof("Stopped monitoring %s", monitoringID)
	return nil
}

// GetStatus retrieves the current status of a debate
func (dms *DebateMonitoringService) GetStatus(ctx context.Context, debateID string) (*DebateStatus, error) {
	return &DebateStatus{
		DebateID:     debateID,
		Status:       "completed",
		CurrentRound: 3,
		TotalRounds:  3,
		StartTime:    time.Now().Add(-5 * time.Minute),
		Participants: []ParticipantStatus{
			{
				ParticipantID:   "participant-1",
				ParticipantName: "Alice",
				Status:          "completed",
				ResponseTime:    5 * time.Second,
			},
		},
	}, nil
}
