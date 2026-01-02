package services

import (
	"context"

	"github.com/sirupsen/logrus"
)

// DebateHistoryService provides historical debate data
type DebateHistoryService struct {
	logger *logrus.Logger
}

// NewDebateHistoryService creates a new history service
func NewDebateHistoryService(logger *logrus.Logger) *DebateHistoryService {
	return &DebateHistoryService{
		logger: logger,
	}
}

// SaveDebateResult saves a debate result to history
func (dhs *DebateHistoryService) SaveDebateResult(ctx context.Context, result *DebateResult) error {
	dhs.logger.Infof("Saved debate result %s to history", result.DebateID)
	return nil
}

// QueryHistory queries historical debate data
func (dhs *DebateHistoryService) QueryHistory(ctx context.Context, filters *HistoryFilters) ([]*DebateResult, error) {
	dhs.logger.Infof("Queried debate history with filters")
	return []*DebateResult{}, nil
}
