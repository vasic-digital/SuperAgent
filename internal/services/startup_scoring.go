package services

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// StartupScoringConfig configures the startup scoring behavior
type StartupScoringConfig struct {
	// Enabled controls whether startup scoring runs
	Enabled bool
	// Async runs scoring in background (non-blocking)
	Async bool
	// Timeout for the entire scoring process
	Timeout time.Duration
	// RetryOnFailure retries failed provider scoring
	RetryOnFailure bool
	// MaxRetries maximum retry attempts per provider
	MaxRetries int
	// ConcurrentWorkers number of concurrent scoring workers
	ConcurrentWorkers int
}

// DefaultStartupScoringConfig returns sensible defaults
func DefaultStartupScoringConfig() StartupScoringConfig {
	return StartupScoringConfig{
		Enabled:           true,
		Async:             true, // Non-blocking by default
		Timeout:           5 * time.Minute,
		RetryOnFailure:    true,
		MaxRetries:        2,
		ConcurrentWorkers: 5,
	}
}

// StartupScoringResult contains results of the startup scoring process
type StartupScoringResult struct {
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	Duration         time.Duration          `json:"duration"`
	TotalProviders   int                    `json:"total_providers"`
	ScoredProviders  int                    `json:"scored_providers"`
	FailedProviders  int                    `json:"failed_providers"`
	SkippedProviders int                    `json:"skipped_providers"`
	ProviderScores   map[string]float64     `json:"provider_scores"`
	ProviderStatus   map[string]string      `json:"provider_status"`
	Errors           []string               `json:"errors,omitempty"`
	Success          bool                   `json:"success"`
}

// StartupScoringService handles automatic provider scoring at system startup
type StartupScoringService struct {
	registry      *ProviderRegistry
	config        StartupScoringConfig
	logger        *logrus.Logger
	result        *StartupScoringResult
	mu            sync.RWMutex
	completed     bool
	completedChan chan struct{}
}

// NewStartupScoringService creates a new startup scoring service
func NewStartupScoringService(registry *ProviderRegistry, config StartupScoringConfig, logger *logrus.Logger) *StartupScoringService {
	if logger == nil {
		logger = logrus.New()
	}
	return &StartupScoringService{
		registry:      registry,
		config:        config,
		logger:        logger,
		completedChan: make(chan struct{}),
	}
}

// Run executes the startup scoring process
// If config.Async is true, this returns immediately and scoring runs in background
func (s *StartupScoringService) Run(ctx context.Context) *StartupScoringResult {
	if !s.config.Enabled {
		s.logger.Info("Startup scoring is disabled")
		return &StartupScoringResult{
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Success:   true,
			ProviderStatus: map[string]string{
				"status": "disabled",
			},
		}
	}

	if s.config.Async {
		// Run in background
		go s.runScoring(ctx)
		return &StartupScoringResult{
			StartTime: time.Now(),
			Success:   true,
			ProviderStatus: map[string]string{
				"status": "running_async",
			},
		}
	}

	// Run synchronously
	return s.runScoring(ctx)
}

// runScoring performs the actual scoring work
func (s *StartupScoringService) runScoring(ctx context.Context) *StartupScoringResult {
	result := &StartupScoringResult{
		StartTime:      time.Now(),
		ProviderScores: make(map[string]float64),
		ProviderStatus: make(map[string]string),
		Errors:         []string{},
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	s.logger.WithFields(logrus.Fields{
		"async":   s.config.Async,
		"timeout": s.config.Timeout,
		"workers": s.config.ConcurrentWorkers,
	}).Info("Starting automatic provider scoring")

	// Run discovery and verification
	if s.registry != nil {
		summary := s.registry.DiscoverAndVerifyProviders(timeoutCtx)

		// Extract results from summary
		if providers, ok := summary["verified_providers"].([]interface{}); ok {
			result.TotalProviders = len(providers)
			for _, p := range providers {
				if pMap, ok := p.(map[string]interface{}); ok {
					name, _ := pMap["name"].(string)
					score, _ := pMap["score"].(float64)
					verified, _ := pMap["verified"].(bool)

					if name != "" {
						result.ProviderScores[name] = score
						if verified {
							result.ProviderStatus[name] = "verified"
							result.ScoredProviders++
						} else {
							result.ProviderStatus[name] = "unverified"
							result.FailedProviders++
						}
					}
				}
			}
		}

		// Also get scores from score adapter
		if scoreAdapter := s.registry.GetScoreAdapter(); scoreAdapter != nil {
			allScores := scoreAdapter.GetAllProviderScores()
			for provider, score := range allScores {
				if _, exists := result.ProviderScores[provider]; !exists {
					result.ProviderScores[provider] = score
					result.ProviderStatus[provider] = "scored"
					result.ScoredProviders++
				}
			}
		}

		// Get healthy providers
		healthyProviders := s.registry.GetHealthyProviders()
		for _, name := range healthyProviders {
			if _, exists := result.ProviderStatus[name]; !exists {
				result.ProviderStatus[name] = "healthy"
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = result.FailedProviders == 0 || result.ScoredProviders > 0

	// Store result
	s.mu.Lock()
	s.result = result
	s.completed = true
	s.mu.Unlock()
	close(s.completedChan)

	s.logger.WithFields(logrus.Fields{
		"duration":   result.Duration,
		"total":      result.TotalProviders,
		"scored":     result.ScoredProviders,
		"failed":     result.FailedProviders,
		"success":    result.Success,
	}).Info("Startup scoring completed")

	return result
}

// GetResult returns the scoring result (may be nil if not completed)
func (s *StartupScoringService) GetResult() *StartupScoringResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result
}

// IsCompleted returns whether scoring has finished
func (s *StartupScoringService) IsCompleted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.completed
}

// WaitForCompletion blocks until scoring is complete or context is cancelled
func (s *StartupScoringService) WaitForCompletion(ctx context.Context) *StartupScoringResult {
	select {
	case <-s.completedChan:
		return s.GetResult()
	case <-ctx.Done():
		return nil
	}
}
