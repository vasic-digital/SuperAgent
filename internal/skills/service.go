package skills

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Service is the main entry point for the skill system.
// It coordinates the registry, matcher, and tracker to provide
// comprehensive skill management.
type Service struct {
	registry *Registry
	matcher  *Matcher
	tracker  *Tracker
	config   *SkillConfig
	log      *logrus.Logger
	mu       sync.RWMutex
	running  bool
}

// NewService creates a new skill service.
func NewService(config *SkillConfig) *Service {
	if config == nil {
		config = DefaultSkillConfig()
	}

	registry := NewRegistry(config)
	matcher := NewMatcher(registry, config)
	tracker := NewTracker()

	return &Service{
		registry: registry,
		matcher:  matcher,
		tracker:  tracker,
		config:   config,
		log:      logrus.New(),
	}
}

// SetLogger sets the logger for all components.
func (s *Service) SetLogger(log *logrus.Logger) {
	s.log = log
	s.registry.SetLogger(log)
	s.matcher.SetLogger(log)
	s.tracker.SetLogger(log)
}

// Initialize loads skills and starts the service.
func (s *Service) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.log.Info("Initializing skill service")

	// Load skills from directory
	if err := s.registry.Load(ctx); err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	// Enable hot reload if configured
	if s.config.HotReload {
		if err := s.registry.EnableHotReload(ctx); err != nil {
			s.log.WithError(err).Warn("Failed to enable hot reload")
		}
	}

	s.running = true
	s.log.WithField("skills_loaded", len(s.registry.GetAll())).Info("Skill service initialized")

	return nil
}

// Start starts the service without loading skills from disk.
// Useful for testing or when skills are registered manually.
func (s *Service) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = true
	s.log.Info("Skill service started (manual mode)")
}

// Shutdown stops the service.
func (s *Service) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.log.Info("Shutting down skill service")
	s.registry.DisableHotReload()
	s.running = false

	return nil
}

// FindSkills finds skills matching the input.
func (s *Service) FindSkills(ctx context.Context, input string) ([]SkillMatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return nil, fmt.Errorf("skill service not running")
	}

	return s.matcher.Match(ctx, input)
}

// FindBestSkill finds the single best matching skill.
func (s *Service) FindBestSkill(ctx context.Context, input string) (*SkillMatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return nil, fmt.Errorf("skill service not running")
	}

	return s.matcher.MatchBest(ctx, input)
}

// GetSkill retrieves a skill by name.
func (s *Service) GetSkill(name string) (*Skill, bool) {
	return s.registry.Get(name)
}

// GetSkillsByCategory retrieves skills in a category.
func (s *Service) GetSkillsByCategory(category string) []*Skill {
	return s.registry.GetByCategory(category)
}

// GetAllSkills returns all registered skills.
func (s *Service) GetAllSkills() []*Skill {
	return s.registry.GetAll()
}

// GetCategories returns all skill categories.
func (s *Service) GetCategories() []string {
	return s.registry.GetCategories()
}

// SearchSkills searches for skills matching a query.
func (s *Service) SearchSkills(query string) []*Skill {
	return s.registry.Search(query)
}

// StartSkillExecution begins tracking a skill execution.
func (s *Service) StartSkillExecution(requestID string, skill *Skill, match *SkillMatch) *SkillUsage {
	return s.tracker.StartTracking(requestID, skill, match)
}

// RecordToolUse records tool usage within a skill execution.
func (s *Service) RecordToolUse(requestID, toolName string) {
	s.tracker.RecordToolUse(requestID, toolName)
}

// CompleteSkillExecution marks skill execution as complete.
func (s *Service) CompleteSkillExecution(requestID string, success bool, errorMsg string) *SkillUsage {
	return s.tracker.CompleteTracking(requestID, success, errorMsg)
}

// GetActiveExecutions returns currently active skill executions.
func (s *Service) GetActiveExecutions() []*SkillUsage {
	return s.tracker.GetActiveUsages()
}

// GetUsageStats returns skill usage statistics.
func (s *Service) GetUsageStats() *UsageStats {
	return s.tracker.GetStats()
}

// GetSkillStats returns statistics for a specific skill.
func (s *Service) GetSkillStats(skillName string) *SkillStats {
	return s.tracker.GetSkillStats(skillName)
}

// GetTopSkills returns the most used skills.
func (s *Service) GetTopSkills(limit int) []*SkillStats {
	return s.tracker.GetTopSkills(limit)
}

// GetUsageHistory returns recent skill usage history.
func (s *Service) GetUsageHistory(limit int) []SkillUsage {
	return s.tracker.GetHistory(limit)
}

// GetRegistryStats returns registry statistics.
func (s *Service) GetRegistryStats() *RegistryStats {
	return s.registry.Stats()
}

// LoadSkillsFromPath loads additional skills from a path.
func (s *Service) LoadSkillsFromPath(ctx context.Context, path string) error {
	return s.registry.LoadFromPath(ctx, path)
}

// RegisterSkill manually registers a skill.
func (s *Service) RegisterSkill(skill *Skill) {
	s.registry.RegisterSkill(skill)
}

// RemoveSkill removes a skill from the registry.
func (s *Service) RemoveSkill(name string) bool {
	return s.registry.Remove(name)
}

// SetSemanticMatcher sets the semantic matcher (LLM-based).
func (s *Service) SetSemanticMatcher(sm SemanticMatcher) {
	s.matcher.SetSemanticMatcher(sm)
}

// CreateResponse creates a skill response with usage tracking.
func (s *Service) CreateResponse(content string, usages []SkillUsage, provider, model, protocol string) *SkillResponse {
	return &SkillResponse{
		Content:      content,
		SkillsUsed:   usages,
		TotalSkills:  len(usages),
		ProviderUsed: provider,
		ModelUsed:    model,
		Protocol:     protocol,
	}
}

// IsRunning returns whether the service is running.
func (s *Service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// HealthCheck performs a health check on the skill service.
func (s *Service) HealthCheck(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return fmt.Errorf("skill service not running")
	}

	stats := s.registry.Stats()
	if stats.TotalSkills == 0 {
		return fmt.Errorf("no skills loaded")
	}

	return nil
}

// SkillExecutionContext holds context for skill execution.
type SkillExecutionContext struct {
	RequestID string
	Skill     *Skill
	Match     *SkillMatch
	StartTime time.Time
	Service   *Service
}

// ExecuteWithTracking executes a function with skill tracking.
func (s *Service) ExecuteWithTracking(ctx context.Context, requestID string, skill *Skill, match *SkillMatch, fn func(ctx context.Context) error) *SkillUsage {
	// Start tracking
	_ = s.StartSkillExecution(requestID, skill, match)

	// Execute the function
	err := fn(ctx)

	// Complete tracking
	success := err == nil
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return s.CompleteSkillExecution(requestID, success, errMsg)
}

// GetConfig returns the skill configuration.
func (s *Service) GetConfig() *SkillConfig {
	return s.config
}

// UpdateConfig updates the skill configuration.
func (s *Service) UpdateConfig(config *SkillConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	s.matcher.config = config
}
