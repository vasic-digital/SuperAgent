package services

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// AdvancedDebateService provides advanced debate management capabilities
type AdvancedDebateService struct {
	debateService      *DebateService
	monitoringService  *DebateMonitoringService
	performanceService *DebatePerformanceService
	historyService     *DebateHistoryService
	resilienceService  *DebateResilienceService
	reportingService   *DebateReportingService
	securityService    *DebateSecurityService
	logger             *logrus.Logger
}

// NewAdvancedDebateService creates a new advanced debate service
func NewAdvancedDebateService(
	debateService *DebateService,
	monitoringService *DebateMonitoringService,
	performanceService *DebatePerformanceService,
	historyService *DebateHistoryService,
	resilienceService *DebateResilienceService,
	reportingService *DebateReportingService,
	securityService *DebateSecurityService,
	logger *logrus.Logger,
) *AdvancedDebateService {
	return &AdvancedDebateService{
		debateService:      debateService,
		monitoringService:  monitoringService,
		performanceService: performanceService,
		historyService:     historyService,
		resilienceService:  resilienceService,
		reportingService:   reportingService,
		securityService:    securityService,
		logger:             logger,
	}
}

// ConductAdvancedDebate conducts a debate with advanced features
func (ads *AdvancedDebateService) ConductAdvancedDebate(
	ctx context.Context,
	config *DebateConfig,
) (*DebateResult, error) {
	// Security check
	if err := ads.securityService.ValidateDebateRequest(ctx, config); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	// Start monitoring
	monitoringID, err := ads.monitoringService.StartMonitoring(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start monitoring: %w", err)
	}
	defer func() { _ = ads.monitoringService.StopMonitoring(ctx, monitoringID) }()

	// Conduct debate
	result, err := ads.debateService.ConductDebate(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("debate failed: %w", err)
	}

	// Record performance metrics
	metrics := ads.performanceService.CalculateMetrics(result)
	if err := ads.performanceService.RecordMetrics(ctx, result.DebateID, metrics); err != nil {
		ads.logger.Warnf("Failed to record performance metrics: %v", err)
	}

	// Save to history
	if err := ads.historyService.SaveDebateResult(ctx, result); err != nil {
		ads.logger.Warnf("Failed to save debate to history: %v", err)
	}

	// Generate report
	report, err := ads.reportingService.GenerateReport(ctx, result)
	if err != nil {
		ads.logger.Warnf("Failed to generate report: %v", err)
	} else {
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["report"] = report
	}

	return result, nil
}

// GetDebateStatus retrieves the current status of a debate
func (ads *AdvancedDebateService) GetDebateStatus(
	ctx context.Context,
	debateID string,
) (*DebateStatus, error) {
	return ads.monitoringService.GetStatus(ctx, debateID)
}

// GetDebateHistory retrieves historical debate data
func (ads *AdvancedDebateService) GetDebateHistory(
	ctx context.Context,
	filters *HistoryFilters,
) ([]*DebateResult, error) {
	return ads.historyService.QueryHistory(ctx, filters)
}

// GetPerformanceMetrics retrieves performance metrics
func (ads *AdvancedDebateService) GetPerformanceMetrics(
	ctx context.Context,
	timeRange TimeRange,
) (*PerformanceMetrics, error) {
	return ads.performanceService.GetMetrics(ctx, timeRange)
}
