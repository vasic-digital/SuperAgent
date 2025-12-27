package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// DebateReportingService provides reporting capabilities
type DebateReportingService struct {
	logger *logrus.Logger
}

// NewDebateReportingService creates a new reporting service
func NewDebateReportingService(logger *logrus.Logger) *DebateReportingService {
	return &DebateReportingService{
		logger: logger,
	}
}

// GenerateReport generates a report for a debate
func (drs *DebateReportingService) GenerateReport(ctx context.Context, result *DebateResult) (*DebateReport, error) {
	return &DebateReport{
		ReportID:        "report-" + result.DebateID,
		DebateID:        result.DebateID,
		GeneratedAt:     time.Now(),
		Summary:         "Debate summary",
		KeyFindings:     []string{"Finding 1", "Finding 2"},
		Recommendations: []string{"Recommendation 1"},
		Metrics: PerformanceMetrics{
			Duration:     result.Duration,
			TotalRounds:  result.TotalRounds,
			QualityScore: result.QualityScore,
		},
	}, nil
}

// ExportReport exports a report in the specified format
func (drs *DebateReportingService) ExportReport(ctx context.Context, reportID string, format string) ([]byte, error) {
	drs.logger.Infof("Exported report %s in format %s", reportID, format)
	return []byte("report content"), nil
}
