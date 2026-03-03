package comprehensive

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MemoryType represents the type of memory
type MemoryType string

const (
	MemoryTypeShortTerm MemoryType = "short_term" // Current conversation
	MemoryTypeLongTerm  MemoryType = "long_term"  // Lessons learned
	MemoryTypeEpisodic  MemoryType = "episodic"   // Reflections (Reflexion)
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	ID          string                 `json:"id"`
	Type        MemoryType             `json:"type"`
	AgentID     string                 `json:"agent_id"`
	Content     string                 `json:"content"`
	Context     map[string]interface{} `json:"context"`
	Importance  float64                `json:"importance"` // 0.0 - 1.0
	AccessCount int                    `json:"access_count"`
	CreatedAt   time.Time              `json:"created_at"`
	AccessedAt  time.Time              `json:"accessed_at"`
}

// NewMemoryEntry creates a new memory entry
func NewMemoryEntry(memType MemoryType, agentID, content string, importance float64) *MemoryEntry {
	now := time.Now()
	return &MemoryEntry{
		ID:          uuid.New().String(),
		Type:        memType,
		AgentID:     agentID,
		Content:     content,
		Context:     make(map[string]interface{}),
		Importance:  importance,
		AccessCount: 0,
		CreatedAt:   now,
		AccessedAt:  now,
	}
}

// Access marks the memory as accessed
func (m *MemoryEntry) Access() {
	m.AccessCount++
	m.AccessedAt = time.Now()
}

// RelevanceScore calculates a relevance score based on importance and recency
func (m *MemoryEntry) RelevanceScore() float64 {
	// Base importance
	score := m.Importance

	// Decay with time (older = less relevant)
	age := time.Since(m.CreatedAt).Hours()
	if age < 24 {
		score *= 1.0 // Full relevance within 24 hours
	} else if age < 168 {
		score *= 0.8 // 80% within a week
	} else {
		score *= 0.5 // 50% after a week
	}

	// Boost for frequently accessed
	accessBonus := float64(m.AccessCount) * 0.05
	if accessBonus > 0.2 {
		accessBonus = 0.2
	}

	return score + accessBonus
}

// ShortTermMemory stores recent conversation context
type ShortTermMemory struct {
	entries []*MemoryEntry
	maxSize int
	agentID string
}

// NewShortTermMemory creates a new short-term memory buffer
func NewShortTermMemory(agentID string, maxSize int) *ShortTermMemory {
	if maxSize <= 0 {
		maxSize = 10
	}

	return &ShortTermMemory{
		entries: make([]*MemoryEntry, 0),
		maxSize: maxSize,
		agentID: agentID,
	}
}

// Add adds an entry to short-term memory
func (s *ShortTermMemory) Add(content string, context map[string]interface{}) *MemoryEntry {
	entry := NewMemoryEntry(MemoryTypeShortTerm, s.agentID, content, 1.0)
	if context != nil {
		// Copy context map to avoid mutating the caller's map
		entry.Context = make(map[string]interface{}, len(context))
		for k, v := range context {
			entry.Context[k] = v
		}
	}

	s.entries = append(s.entries, entry)

	// Trim if exceeds max size
	if len(s.entries) > s.maxSize {
		s.entries = s.entries[len(s.entries)-s.maxSize:]
	}

	return entry
}

// GetRecent returns the N most recent entries
func (s *ShortTermMemory) GetRecent(n int) []*MemoryEntry {
	if n > len(s.entries) {
		n = len(s.entries)
	}

	result := make([]*MemoryEntry, n)
	copy(result, s.entries[len(s.entries)-n:])

	// Mark as accessed
	for _, entry := range result {
		entry.Access()
	}

	return result
}

// GetAll returns all entries
func (s *ShortTermMemory) GetAll() []*MemoryEntry {
	result := make([]*MemoryEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

// Clear clears all entries
func (s *ShortTermMemory) Clear() {
	s.entries = make([]*MemoryEntry, 0)
}

// Size returns the number of entries
func (s *ShortTermMemory) Size() int {
	return len(s.entries)
}

// LongTermMemory stores lessons learned across debates
type LongTermMemory struct {
	entries map[string]*MemoryEntry
	agentID string
	maxSize int
}

// NewLongTermMemory creates a new long-term memory
func NewLongTermMemory(agentID string, maxSize int) *LongTermMemory {
	if maxSize <= 0 {
		maxSize = 100
	}

	return &LongTermMemory{
		entries: make(map[string]*MemoryEntry),
		agentID: agentID,
		maxSize: maxSize,
	}
}

// Store stores a lesson learned
func (l *LongTermMemory) Store(lesson string, importance float64, context map[string]interface{}) *MemoryEntry {
	// Check if similar lesson already exists
	for _, entry := range l.entries {
		if entry.Content == lesson {
			entry.Importance = (entry.Importance + importance) / 2
			if context != nil {
				for k, v := range context {
					entry.Context[k] = v
				}
			}
			entry.Access()
			return entry
		}
	}

	// Create new entry
	entry := NewMemoryEntry(MemoryTypeLongTerm, l.agentID, lesson, importance)
	if context != nil {
		entry.Context = context
	}

	l.entries[entry.ID] = entry

	// Prune if exceeds max size
	if len(l.entries) > l.maxSize {
		l.pruneLeastImportant()
	}

	return entry
}

// Retrieve retrieves lessons relevant to a topic
func (l *LongTermMemory) Retrieve(topic string, limit int) []*MemoryEntry {
	var matches []*MemoryEntry

	for _, entry := range l.entries {
		if containsIgnoreCase(entry.Content, topic) {
			entry.Access()
			matches = append(matches, entry)
		}
	}

	// Sort by relevance
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].RelevanceScore() > matches[i].RelevanceScore() {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	if limit > len(matches) {
		limit = len(matches)
	}

	return matches[:limit]
}

// GetAll returns all entries
func (l *LongTermMemory) GetAll() []*MemoryEntry {
	result := make([]*MemoryEntry, 0, len(l.entries))
	for _, entry := range l.entries {
		result = append(result, entry)
	}
	return result
}

// pruneLeastImportant removes the least important entries
func (l *LongTermMemory) pruneLeastImportant() {
	var leastImportant *MemoryEntry
	for _, entry := range l.entries {
		if leastImportant == nil || entry.RelevanceScore() < leastImportant.RelevanceScore() {
			leastImportant = entry
		}
	}

	if leastImportant != nil {
		delete(l.entries, leastImportant.ID)
	}
}

// EpisodicMemory stores reflections for Reflexion
type EpisodicMemory struct {
	entries     []*MemoryEntry
	agentID     string
	maxEpisodes int
}

// NewEpisodicMemory creates a new episodic memory
func NewEpisodicMemory(agentID string, maxEpisodes int) *EpisodicMemory {
	if maxEpisodes <= 0 {
		maxEpisodes = 50
	}

	return &EpisodicMemory{
		entries:     make([]*MemoryEntry, 0),
		agentID:     agentID,
		maxEpisodes: maxEpisodes,
	}
}

// AddReflection adds a reflection after a failure
func (e *EpisodicMemory) AddReflection(reflection string, failure string, context map[string]interface{}) *MemoryEntry {
	content := fmt.Sprintf("Reflection: %s\nFailure: %s", reflection, failure)
	entry := NewMemoryEntry(MemoryTypeEpisodic, e.agentID, content, 0.9)

	if context != nil {
		// Copy context map to avoid mutating the caller's map
		entry.Context = make(map[string]interface{}, len(context)+2)
		for k, v := range context {
			entry.Context[k] = v
		}
		entry.Context["reflection"] = reflection
		entry.Context["failure"] = failure
	}

	e.entries = append(e.entries, entry)

	// Trim if exceeds max
	if len(e.entries) > e.maxEpisodes {
		e.entries = e.entries[len(e.entries)-e.maxEpisodes:]
	}

	return entry
}

// GetRelevantReflections retrieves reflections relevant to current context
func (e *EpisodicMemory) GetRelevantReflections(context string, limit int) []*MemoryEntry {
	var matches []*MemoryEntry

	for _, entry := range e.entries {
		// Simple relevance check - could be enhanced with embeddings
		if containsIgnoreCase(entry.Content, context) ||
			containsIgnoreCase(entry.Content, "error") ||
			containsIgnoreCase(entry.Content, "failure") {
			entry.Access()
			matches = append(matches, entry)
		}
	}

	// Sort by importance and recency
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].RelevanceScore() > matches[i].RelevanceScore() {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	if limit > len(matches) {
		limit = len(matches)
	}

	return matches[:limit]
}

// GetAll returns all reflections
func (e *EpisodicMemory) GetAll() []*MemoryEntry {
	result := make([]*MemoryEntry, len(e.entries))
	copy(result, e.entries)
	return result
}

// MemoryManager manages all memory types for an agent
type MemoryManager struct {
	ShortTerm *ShortTermMemory
	LongTerm  *LongTermMemory
	Episodic  *EpisodicMemory
	AgentID   string
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(agentID string) *MemoryManager {
	return &MemoryManager{
		ShortTerm: NewShortTermMemory(agentID, 20),
		LongTerm:  NewLongTermMemory(agentID, 100),
		Episodic:  NewEpisodicMemory(agentID, 50),
		AgentID:   agentID,
	}
}

// AddToShortTerm adds to short-term memory
func (m *MemoryManager) AddToShortTerm(content string, context map[string]interface{}) {
	m.ShortTerm.Add(content, context)
}

// StoreLesson stores a lesson in long-term memory
func (m *MemoryManager) StoreLesson(lesson string, importance float64, context map[string]interface{}) {
	m.LongTerm.Store(lesson, importance, context)
}

// AddReflection adds a reflection to episodic memory
func (m *MemoryManager) AddReflection(reflection, failure string, context map[string]interface{}) {
	m.Episodic.AddReflection(reflection, failure, context)
}

// GetContext retrieves relevant context for current task
func (m *MemoryManager) GetContext(topic string) map[string]interface{} {
	context := make(map[string]interface{})

	// Get recent short-term context
	recent := m.ShortTerm.GetRecent(5)
	context["recent_history"] = recent

	// Get relevant lessons
	lessons := m.LongTerm.Retrieve(topic, 3)
	context["relevant_lessons"] = lessons

	// Get relevant reflections
	reflections := m.Episodic.GetRelevantReflections(topic, 3)
	context["relevant_reflections"] = reflections

	return context
}

// Serialize serializes all memories to JSON
func (m *MemoryManager) Serialize() ([]byte, error) {
	data := map[string]interface{}{
		"short_term": m.ShortTerm.GetAll(),
		"long_term":  m.LongTerm.GetAll(),
		"episodic":   m.Episodic.GetAll(),
	}

	return json.MarshalIndent(data, "", "  ")
}

// Helper function
func containsIgnoreCase(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
