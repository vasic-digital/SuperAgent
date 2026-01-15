package skills

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Tracker tracks skill usage across requests and sessions.
type Tracker struct {
	mu            sync.RWMutex
	activeUsages  map[string]*SkillUsage // requestID -> usage
	history       []SkillUsage           // Historical usage records
	historyLimit  int
	stats         *UsageStats
	log           *logrus.Logger
}

// UsageStats provides aggregate usage statistics.
type UsageStats struct {
	TotalInvocations    int64                     `json:"total_invocations"`
	SuccessfulCount     int64                     `json:"successful_count"`
	FailedCount         int64                     `json:"failed_count"`
	BySkill             map[string]*SkillStats    `json:"by_skill"`
	ByCategory          map[string]*CategoryStats `json:"by_category"`
	ByMatchType         map[MatchType]int64       `json:"by_match_type"`
	AverageConfidence   float64                   `json:"average_confidence"`
	AverageDuration     time.Duration             `json:"average_duration"`
	LastUpdated         time.Time                 `json:"last_updated"`
}

// SkillStats provides per-skill statistics.
type SkillStats struct {
	Name             string        `json:"name"`
	Category         string        `json:"category"`
	InvocationCount  int64         `json:"invocation_count"`
	SuccessCount     int64         `json:"success_count"`
	FailureCount     int64         `json:"failure_count"`
	AverageConfidence float64      `json:"average_confidence"`
	AverageDuration  time.Duration `json:"average_duration"`
	TotalDuration    time.Duration `json:"total_duration"`
	LastUsed         time.Time     `json:"last_used"`
	TriggersCounted  map[string]int64 `json:"triggers_counted"`
}

// CategoryStats provides per-category statistics.
type CategoryStats struct {
	Category        string        `json:"category"`
	InvocationCount int64         `json:"invocation_count"`
	SuccessCount    int64         `json:"success_count"`
	UniqueSkills    int           `json:"unique_skills"`
}

// NewTracker creates a new skill usage tracker.
func NewTracker() *Tracker {
	return &Tracker{
		activeUsages: make(map[string]*SkillUsage),
		history:      make([]SkillUsage, 0, 1000),
		historyLimit: 10000,
		stats: &UsageStats{
			BySkill:     make(map[string]*SkillStats),
			ByCategory:  make(map[string]*CategoryStats),
			ByMatchType: make(map[MatchType]int64),
			LastUpdated: time.Now(),
		},
		log: logrus.New(),
	}
}

// SetLogger sets the logger for the tracker.
func (t *Tracker) SetLogger(log *logrus.Logger) {
	t.log = log
}

// StartTracking begins tracking usage for a skill.
func (t *Tracker) StartTracking(requestID string, skill *Skill, match *SkillMatch) *SkillUsage {
	t.mu.Lock()
	defer t.mu.Unlock()

	usage := &SkillUsage{
		SkillName:    skill.Name,
		Category:     skill.Category,
		TriggerUsed:  match.MatchedTrigger,
		MatchType:    match.MatchType,
		Confidence:   match.Confidence,
		ToolsInvoked: make([]string, 0),
		StartedAt:    time.Now(),
	}

	t.activeUsages[requestID] = usage

	t.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"skill":      skill.Name,
		"trigger":    match.MatchedTrigger,
		"confidence": match.Confidence,
	}).Debug("Started tracking skill usage")

	return usage
}

// RecordToolUse records a tool being used by a skill.
func (t *Tracker) RecordToolUse(requestID, toolName string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if usage, ok := t.activeUsages[requestID]; ok {
		usage.ToolsInvoked = append(usage.ToolsInvoked, toolName)
	}
}

// CompleteTracking marks skill usage as complete.
func (t *Tracker) CompleteTracking(requestID string, success bool, err string) *SkillUsage {
	t.mu.Lock()
	defer t.mu.Unlock()

	usage, ok := t.activeUsages[requestID]
	if !ok {
		return nil
	}

	usage.CompletedAt = time.Now()
	usage.Success = success
	usage.Error = err

	// Update statistics
	t.updateStats(usage)

	// Add to history
	t.addToHistory(*usage)

	// Remove from active
	delete(t.activeUsages, requestID)

	t.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"skill":      usage.SkillName,
		"success":    success,
		"duration":   usage.CompletedAt.Sub(usage.StartedAt),
	}).Debug("Completed tracking skill usage")

	return usage
}

// GetActiveUsage returns the active usage for a request.
func (t *Tracker) GetActiveUsage(requestID string) *SkillUsage {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.activeUsages[requestID]
}

// GetActiveUsages returns all currently active usages.
func (t *Tracker) GetActiveUsages() []*SkillUsage {
	t.mu.RLock()
	defer t.mu.RUnlock()

	usages := make([]*SkillUsage, 0, len(t.activeUsages))
	for _, usage := range t.activeUsages {
		usages = append(usages, usage)
	}
	return usages
}

// updateStats updates aggregate statistics.
func (t *Tracker) updateStats(usage *SkillUsage) {
	t.stats.TotalInvocations++

	if usage.Success {
		t.stats.SuccessfulCount++
	} else {
		t.stats.FailedCount++
	}

	// Update by match type
	t.stats.ByMatchType[usage.MatchType]++

	// Update by skill
	skillStats, ok := t.stats.BySkill[usage.SkillName]
	if !ok {
		skillStats = &SkillStats{
			Name:            usage.SkillName,
			Category:        usage.Category,
			TriggersCounted: make(map[string]int64),
		}
		t.stats.BySkill[usage.SkillName] = skillStats
	}

	skillStats.InvocationCount++
	if usage.Success {
		skillStats.SuccessCount++
	} else {
		skillStats.FailureCount++
	}

	duration := usage.CompletedAt.Sub(usage.StartedAt)
	skillStats.TotalDuration += duration
	skillStats.AverageDuration = skillStats.TotalDuration / time.Duration(skillStats.InvocationCount)
	skillStats.LastUsed = usage.CompletedAt
	skillStats.TriggersCounted[usage.TriggerUsed]++

	// Update average confidence
	oldConfidence := skillStats.AverageConfidence
	skillStats.AverageConfidence = ((oldConfidence * float64(skillStats.InvocationCount-1)) + usage.Confidence) / float64(skillStats.InvocationCount)

	// Update by category
	catStats, ok := t.stats.ByCategory[usage.Category]
	if !ok {
		catStats = &CategoryStats{
			Category: usage.Category,
		}
		t.stats.ByCategory[usage.Category] = catStats
	}

	catStats.InvocationCount++
	if usage.Success {
		catStats.SuccessCount++
	}

	// Calculate global average confidence
	totalConfidence := float64(0)
	for _, ss := range t.stats.BySkill {
		totalConfidence += ss.AverageConfidence * float64(ss.InvocationCount)
	}
	t.stats.AverageConfidence = totalConfidence / float64(t.stats.TotalInvocations)

	// Calculate global average duration
	totalDuration := time.Duration(0)
	for _, ss := range t.stats.BySkill {
		totalDuration += ss.TotalDuration
	}
	t.stats.AverageDuration = totalDuration / time.Duration(t.stats.TotalInvocations)

	t.stats.LastUpdated = time.Now()
}

// addToHistory adds a usage record to history.
func (t *Tracker) addToHistory(usage SkillUsage) {
	t.history = append(t.history, usage)

	// Trim if over limit
	if len(t.history) > t.historyLimit {
		t.history = t.history[len(t.history)-t.historyLimit:]
	}
}

// GetHistory returns usage history.
func (t *Tracker) GetHistory(limit int) []SkillUsage {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if limit <= 0 || limit > len(t.history) {
		limit = len(t.history)
	}

	// Return most recent
	start := len(t.history) - limit
	if start < 0 {
		start = 0
	}

	result := make([]SkillUsage, limit)
	copy(result, t.history[start:])
	return result
}

// GetStats returns aggregate statistics.
func (t *Tracker) GetStats() *UsageStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy
	stats := &UsageStats{
		TotalInvocations:  t.stats.TotalInvocations,
		SuccessfulCount:   t.stats.SuccessfulCount,
		FailedCount:       t.stats.FailedCount,
		BySkill:           make(map[string]*SkillStats),
		ByCategory:        make(map[string]*CategoryStats),
		ByMatchType:       make(map[MatchType]int64),
		AverageConfidence: t.stats.AverageConfidence,
		AverageDuration:   t.stats.AverageDuration,
		LastUpdated:       t.stats.LastUpdated,
	}

	for k, v := range t.stats.BySkill {
		stats.BySkill[k] = v
	}
	for k, v := range t.stats.ByCategory {
		stats.ByCategory[k] = v
	}
	for k, v := range t.stats.ByMatchType {
		stats.ByMatchType[k] = v
	}

	return stats
}

// GetSkillStats returns statistics for a specific skill.
func (t *Tracker) GetSkillStats(skillName string) *SkillStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.stats.BySkill[skillName]
}

// GetTopSkills returns the most frequently used skills.
func (t *Tracker) GetTopSkills(limit int) []*SkillStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Convert to slice and sort
	skills := make([]*SkillStats, 0, len(t.stats.BySkill))
	for _, s := range t.stats.BySkill {
		skills = append(skills, s)
	}

	// Sort by invocation count
	for i := 0; i < len(skills)-1; i++ {
		for j := i + 1; j < len(skills); j++ {
			if skills[j].InvocationCount > skills[i].InvocationCount {
				skills[i], skills[j] = skills[j], skills[i]
			}
		}
	}

	if limit > 0 && limit < len(skills) {
		skills = skills[:limit]
	}

	return skills
}

// Reset clears all tracking data.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.activeUsages = make(map[string]*SkillUsage)
	t.history = make([]SkillUsage, 0, 1000)
	t.stats = &UsageStats{
		BySkill:     make(map[string]*SkillStats),
		ByCategory:  make(map[string]*CategoryStats),
		ByMatchType: make(map[MatchType]int64),
		LastUpdated: time.Now(),
	}
}
