package skills

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Registry manages the collection of available skills.
type Registry struct {
	mu            sync.RWMutex
	skills        map[string]*Skill        // name -> skill
	byCategory    map[string][]*Skill      // category -> skills
	byTrigger     map[string][]*Skill      // trigger -> skills
	categories    map[string]*SkillCategory
	parser        *Parser
	config        *SkillConfig
	stats         *RegistryStats
	watcher       *DirectoryWatcher
	log           *logrus.Logger
}

// DirectoryWatcher watches for skill file changes.
type DirectoryWatcher struct {
	dir      string
	interval time.Duration
	stop     chan struct{}
	callback func()
}

// NewRegistry creates a new skill registry.
func NewRegistry(config *SkillConfig) *Registry {
	if config == nil {
		config = DefaultSkillConfig()
	}

	return &Registry{
		skills:     make(map[string]*Skill),
		byCategory: make(map[string][]*Skill),
		byTrigger:  make(map[string][]*Skill),
		categories: make(map[string]*SkillCategory),
		parser:     NewParser(),
		config:     config,
		stats:      &RegistryStats{SkillsByCategory: make(map[string]int)},
		log:        logrus.New(),
	}
}

// SetLogger sets the logger for the registry.
func (r *Registry) SetLogger(log *logrus.Logger) {
	r.log = log
}

// Load loads all skills from the configured directory.
func (r *Registry) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.log.WithField("directory", r.config.SkillsDirectory).Info("Loading skills")

	skills, err := r.parser.ParseDirectory(r.config.SkillsDirectory)
	if err != nil {
		return fmt.Errorf("failed to parse skills directory: %w", err)
	}

	// Clear existing data
	r.skills = make(map[string]*Skill)
	r.byCategory = make(map[string][]*Skill)
	r.byTrigger = make(map[string][]*Skill)

	// Register each skill
	for _, skill := range skills {
		r.registerSkill(skill)
	}

	// Update stats
	r.updateStats()

	r.log.WithFields(logrus.Fields{
		"total":      len(r.skills),
		"categories": len(r.byCategory),
		"triggers":   len(r.byTrigger),
	}).Info("Skills loaded successfully")

	return nil
}

// LoadFromPath loads skills from a specific path.
func (r *Registry) LoadFromPath(ctx context.Context, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	skills, err := r.parser.ParseDirectory(path)
	if err != nil {
		return fmt.Errorf("failed to parse skills from path %s: %w", path, err)
	}

	for _, skill := range skills {
		r.registerSkill(skill)
	}

	r.updateStats()
	return nil
}

// RegisterSkill registers a single skill.
func (r *Registry) RegisterSkill(skill *Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registerSkill(skill)
	r.updateStats()
}

// registerSkill internal registration (caller must hold lock).
func (r *Registry) registerSkill(skill *Skill) {
	if skill.Name == "" {
		return
	}

	// Register by name
	r.skills[skill.Name] = skill

	// Register by category
	if skill.Category != "" {
		r.byCategory[skill.Category] = append(r.byCategory[skill.Category], skill)
	}

	// Register by triggers
	for _, trigger := range skill.TriggerPhrases {
		r.byTrigger[trigger] = append(r.byTrigger[trigger], skill)
	}

	r.log.WithFields(logrus.Fields{
		"skill":    skill.Name,
		"category": skill.Category,
		"triggers": len(skill.TriggerPhrases),
	}).Debug("Skill registered")
}

// Get retrieves a skill by name.
func (r *Registry) Get(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, ok := r.skills[name]
	return skill, ok
}

// GetByCategory retrieves all skills in a category.
func (r *Registry) GetByCategory(category string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills, ok := r.byCategory[category]
	if !ok {
		return nil
	}

	// Return a copy
	result := make([]*Skill, len(skills))
	copy(result, skills)
	return result
}

// GetByTrigger retrieves skills that match a trigger phrase.
func (r *Registry) GetByTrigger(trigger string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills, ok := r.byTrigger[trigger]
	if !ok {
		return nil
	}

	result := make([]*Skill, len(skills))
	copy(result, skills)
	return result
}

// GetAll returns all registered skills.
func (r *Registry) GetAll() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		result = append(result, skill)
	}
	return result
}

// GetCategories returns all category names.
func (r *Registry) GetCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]string, 0, len(r.byCategory))
	for cat := range r.byCategory {
		categories = append(categories, cat)
	}
	return categories
}

// GetTriggers returns all trigger phrases.
func (r *Registry) GetTriggers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	triggers := make([]string, 0, len(r.byTrigger))
	for trigger := range r.byTrigger {
		triggers = append(triggers, trigger)
	}
	return triggers
}

// Remove removes a skill from the registry.
func (r *Registry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	skill, ok := r.skills[name]
	if !ok {
		return false
	}

	// Remove from main map
	delete(r.skills, name)

	// Remove from category index
	if skill.Category != "" {
		skills := r.byCategory[skill.Category]
		for i, s := range skills {
			if s.Name == name {
				r.byCategory[skill.Category] = append(skills[:i], skills[i+1:]...)
				break
			}
		}
	}

	// Remove from trigger index
	for _, trigger := range skill.TriggerPhrases {
		skills := r.byTrigger[trigger]
		for i, s := range skills {
			if s.Name == name {
				r.byTrigger[trigger] = append(skills[:i], skills[i+1:]...)
				break
			}
		}
	}

	r.updateStats()
	return true
}

// Stats returns registry statistics.
func (r *Registry) Stats() *RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	stats := &RegistryStats{
		TotalSkills:      r.stats.TotalSkills,
		SkillsByCategory: make(map[string]int),
		TotalTriggers:    r.stats.TotalTriggers,
		LoadedAt:         r.stats.LoadedAt,
		LastUpdated:      r.stats.LastUpdated,
	}
	for k, v := range r.stats.SkillsByCategory {
		stats.SkillsByCategory[k] = v
	}
	return stats
}

// updateStats recalculates registry statistics.
func (r *Registry) updateStats() {
	r.stats.TotalSkills = len(r.skills)
	r.stats.TotalTriggers = len(r.byTrigger)
	r.stats.LastUpdated = time.Now()

	if r.stats.LoadedAt.IsZero() {
		r.stats.LoadedAt = time.Now()
	}

	r.stats.SkillsByCategory = make(map[string]int)
	for cat, skills := range r.byCategory {
		r.stats.SkillsByCategory[cat] = len(skills)
	}
}

// EnableHotReload enables automatic reloading of skills on file changes.
func (r *Registry) EnableHotReload(ctx context.Context) error {
	if !r.config.HotReload {
		return nil
	}

	r.watcher = &DirectoryWatcher{
		dir:      r.config.SkillsDirectory,
		interval: r.config.HotReloadInterval,
		stop:     make(chan struct{}),
		callback: func() {
			if err := r.Load(ctx); err != nil {
				r.log.WithError(err).Error("Failed to reload skills")
			}
		},
	}

	go r.watcher.start()
	r.log.Info("Hot reload enabled for skills")
	return nil
}

// DisableHotReload stops the hot reload watcher.
func (r *Registry) DisableHotReload() {
	if r.watcher != nil {
		close(r.watcher.stop)
		r.watcher = nil
	}
}

// start begins watching for file changes.
func (w *DirectoryWatcher) start() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			// Simple implementation: just reload periodically
			// A more sophisticated version would use fsnotify
			w.callback()
		}
	}
}

// Search searches for skills matching a query.
func (r *Registry) Search(query string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = normalizeQuery(query)
	matches := make([]*Skill, 0)
	seen := make(map[string]bool)

	// Search by trigger
	for trigger, skills := range r.byTrigger {
		if containsIgnoreCase(trigger, query) || containsIgnoreCase(query, trigger) {
			for _, skill := range skills {
				if !seen[skill.Name] {
					matches = append(matches, skill)
					seen[skill.Name] = true
				}
			}
		}
	}

	// Search by name
	for name, skill := range r.skills {
		if !seen[name] && containsIgnoreCase(name, query) {
			matches = append(matches, skill)
			seen[name] = true
		}
	}

	// Search by description
	for _, skill := range r.skills {
		if !seen[skill.Name] && containsIgnoreCase(skill.Description, query) {
			matches = append(matches, skill)
			seen[skill.Name] = true
		}
	}

	return matches
}
