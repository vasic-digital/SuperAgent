// Package context provides context window management for LLM interactions.
package context

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	// ErrContextOverflow indicates the context window limit was exceeded.
	ErrContextOverflow = errors.New("context window overflow")
	// ErrInvalidTokenLimit indicates an invalid token limit.
	ErrInvalidTokenLimit = errors.New("invalid token limit")
	// ErrEmptyContext indicates the context is empty.
	ErrEmptyContext = errors.New("context is empty")
)

// ContextWindow manages the context window for LLM interactions.
type ContextWindow struct {
	mu           sync.RWMutex
	entries      []ContextEntry
	config       *WindowConfig
	tokenCount   int
	lastAccess   time.Time
	eventHandler WindowEventHandler
}

// ContextEntry represents an entry in the context window.
type ContextEntry struct {
	// ID is the unique identifier for this entry.
	ID string `json:"id"`
	// Role is the message role (system, user, assistant, tool).
	Role string `json:"role"`
	// Content is the entry content.
	Content string `json:"content"`
	// TokenCount is the number of tokens in this entry.
	TokenCount int `json:"token_count"`
	// Timestamp is when this entry was added.
	Timestamp time.Time `json:"timestamp"`
	// Priority determines importance for eviction.
	Priority Priority `json:"priority"`
	// Metadata contains additional metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Pinned indicates the entry should not be evicted.
	Pinned bool `json:"pinned"`
}

// Priority levels for context entries.
type Priority int

const (
	// PriorityLow is low priority (evicted first).
	PriorityLow Priority = 0
	// PriorityNormal is normal priority.
	PriorityNormal Priority = 1
	// PriorityHigh is high priority.
	PriorityHigh Priority = 2
	// PriorityCritical is critical priority (evicted last).
	PriorityCritical Priority = 3
)

// WindowConfig holds configuration for the context window.
type WindowConfig struct {
	// MaxTokens is the maximum tokens allowed.
	MaxTokens int `json:"max_tokens"`
	// ReserveTokens is the number of tokens to reserve for output.
	ReserveTokens int `json:"reserve_tokens"`
	// EvictionPolicy determines how to evict entries.
	EvictionPolicy EvictionPolicy `json:"eviction_policy"`
	// EvictionThreshold triggers eviction when usage exceeds this.
	EvictionThreshold float64 `json:"eviction_threshold"`
	// PreserveSystemPrompt keeps the system prompt from eviction.
	PreserveSystemPrompt bool `json:"preserve_system_prompt"`
	// PreserveLastN keeps the last N entries from eviction.
	PreserveLastN int `json:"preserve_last_n"`
}

// EvictionPolicy defines how entries are evicted.
type EvictionPolicy string

const (
	// EvictionPolicyFIFO evicts oldest entries first.
	EvictionPolicyFIFO EvictionPolicy = "fifo"
	// EvictionPolicyLRU evicts least recently used.
	EvictionPolicyLRU EvictionPolicy = "lru"
	// EvictionPolicyPriority evicts lowest priority first.
	EvictionPolicyPriority EvictionPolicy = "priority"
	// EvictionPolicySummarize summarizes older entries.
	EvictionPolicySummarize EvictionPolicy = "summarize"
)

// DefaultWindowConfig returns a default configuration.
func DefaultWindowConfig() *WindowConfig {
	return &WindowConfig{
		MaxTokens:            4096,
		ReserveTokens:        512,
		EvictionPolicy:       EvictionPolicyFIFO,
		EvictionThreshold:    0.9,
		PreserveSystemPrompt: true,
		PreserveLastN:        2,
	}
}

// WindowEventHandler handles window events.
type WindowEventHandler func(event *WindowEvent)

// WindowEvent represents an event in the context window.
type WindowEvent struct {
	// Type is the event type.
	Type WindowEventType `json:"type"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// Entry is the affected entry (if applicable).
	Entry *ContextEntry `json:"entry,omitempty"`
	// TokenCount is the current token count.
	TokenCount int `json:"token_count"`
	// Metadata contains additional metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// WindowEventType defines types of window events.
type WindowEventType string

const (
	// EventTypeEntryAdded indicates an entry was added.
	EventTypeEntryAdded WindowEventType = "entry_added"
	// EventTypeEntryEvicted indicates an entry was evicted.
	EventTypeEntryEvicted WindowEventType = "entry_evicted"
	// EventTypeEntryUpdated indicates an entry was updated.
	EventTypeEntryUpdated WindowEventType = "entry_updated"
	// EventTypeOverflow indicates context overflow.
	EventTypeOverflow WindowEventType = "overflow"
	// EventTypeSummarized indicates context was summarized.
	EventTypeSummarized WindowEventType = "summarized"
)

// NewContextWindow creates a new context window.
func NewContextWindow(config *WindowConfig) *ContextWindow {
	if config == nil {
		config = DefaultWindowConfig()
	}
	return &ContextWindow{
		entries:    make([]ContextEntry, 0),
		config:     config,
		lastAccess: time.Now(),
	}
}

// SetEventHandler sets the event handler.
func (w *ContextWindow) SetEventHandler(handler WindowEventHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.eventHandler = handler
}

// Add adds an entry to the context window.
func (w *ContextWindow) Add(entry ContextEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if entry.TokenCount == 0 {
		entry.TokenCount = estimateTokens(entry.Content)
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.ID == "" {
		entry.ID = generateID()
	}

	// Check if adding would exceed limit
	availableTokens := w.config.MaxTokens - w.config.ReserveTokens
	if w.tokenCount+entry.TokenCount > availableTokens {
		// Try to evict entries
		needed := w.tokenCount + entry.TokenCount - availableTokens
		if err := w.evictTokens(needed); err != nil {
			w.emitEvent(EventTypeOverflow, nil)
			return ErrContextOverflow
		}
	}

	w.entries = append(w.entries, entry)
	w.tokenCount += entry.TokenCount
	w.lastAccess = time.Now()

	w.emitEvent(EventTypeEntryAdded, &entry)
	return nil
}

// AddMessage adds a message to the context window.
func (w *ContextWindow) AddMessage(role, content string) error {
	return w.Add(ContextEntry{
		Role:     role,
		Content:  content,
		Priority: PriorityNormal,
	})
}

// AddSystemPrompt adds a system prompt (pinned).
func (w *ContextWindow) AddSystemPrompt(content string) error {
	return w.Add(ContextEntry{
		Role:     "system",
		Content:  content,
		Priority: PriorityCritical,
		Pinned:   w.config.PreserveSystemPrompt,
	})
}

// Get returns all entries in the context window.
func (w *ContextWindow) Get() []ContextEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]ContextEntry, len(w.entries))
	copy(result, w.entries)
	return result
}

// GetMessages returns entries formatted as messages.
func (w *ContextWindow) GetMessages() []map[string]string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	messages := make([]map[string]string, len(w.entries))
	for i, entry := range w.entries {
		messages[i] = map[string]string{
			"role":    entry.Role,
			"content": entry.Content,
		}
	}
	return messages
}

// TokenCount returns the current token count.
func (w *ContextWindow) TokenCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.tokenCount
}

// AvailableTokens returns the number of tokens available.
func (w *ContextWindow) AvailableTokens() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config.MaxTokens - w.config.ReserveTokens - w.tokenCount
}

// UsageRatio returns the context window usage ratio (0-1).
func (w *ContextWindow) UsageRatio() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	maxUsable := w.config.MaxTokens - w.config.ReserveTokens
	if maxUsable <= 0 {
		return 1.0
	}
	return float64(w.tokenCount) / float64(maxUsable)
}

// Clear clears all entries from the context window.
func (w *ContextWindow) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = make([]ContextEntry, 0)
	w.tokenCount = 0
}

// ClearExceptPinned clears all except pinned entries.
func (w *ContextWindow) ClearExceptPinned() {
	w.mu.Lock()
	defer w.mu.Unlock()

	var preserved []ContextEntry
	preservedTokens := 0

	for _, entry := range w.entries {
		if entry.Pinned {
			preserved = append(preserved, entry)
			preservedTokens += entry.TokenCount
		}
	}

	w.entries = preserved
	w.tokenCount = preservedTokens
}

// RemoveEntry removes an entry by ID.
func (w *ContextWindow) RemoveEntry(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, entry := range w.entries {
		if entry.ID == id {
			w.entries = append(w.entries[:i], w.entries[i+1:]...)
			w.tokenCount -= entry.TokenCount
			w.emitEvent(EventTypeEntryEvicted, &entry)
			return true
		}
	}
	return false
}

// UpdateEntry updates an existing entry.
func (w *ContextWindow) UpdateEntry(id string, newContent string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, entry := range w.entries {
		if entry.ID == id {
			oldTokens := entry.TokenCount
			newTokens := estimateTokens(newContent)

			// Check if update would exceed limit
			tokenDiff := newTokens - oldTokens
			if tokenDiff > 0 && w.tokenCount+tokenDiff > w.config.MaxTokens-w.config.ReserveTokens {
				return ErrContextOverflow
			}

			w.entries[i].Content = newContent
			w.entries[i].TokenCount = newTokens
			w.tokenCount += tokenDiff

			w.emitEvent(EventTypeEntryUpdated, &w.entries[i])
			return nil
		}
	}
	return errors.New("entry not found")
}

// evictTokens evicts entries to free at least the specified number of tokens.
func (w *ContextWindow) evictTokens(tokensNeeded int) error {
	evicted := 0

	switch w.config.EvictionPolicy {
	case EvictionPolicyFIFO:
		evicted = w.evictFIFO(tokensNeeded)
	case EvictionPolicyLRU:
		evicted = w.evictLRU(tokensNeeded)
	case EvictionPolicyPriority:
		evicted = w.evictByPriority(tokensNeeded)
	default:
		evicted = w.evictFIFO(tokensNeeded)
	}

	if evicted < tokensNeeded {
		return ErrContextOverflow
	}
	return nil
}

func (w *ContextWindow) evictFIFO(tokensNeeded int) int {
	evicted := 0
	preserveFrom := len(w.entries) - w.config.PreserveLastN

	var remaining []ContextEntry
	for i, entry := range w.entries {
		// Preserve pinned, system prompts (if configured), and last N
		if entry.Pinned || (w.config.PreserveSystemPrompt && entry.Role == "system") || i >= preserveFrom {
			remaining = append(remaining, entry)
			continue
		}

		if evicted >= tokensNeeded {
			remaining = append(remaining, entry)
			continue
		}

		evicted += entry.TokenCount
		w.emitEvent(EventTypeEntryEvicted, &entry)
	}

	w.entries = remaining
	w.tokenCount -= evicted
	return evicted
}

func (w *ContextWindow) evictLRU(tokensNeeded int) int {
	// For LRU, we use timestamp as a proxy for "recently used"
	// In a more complete implementation, we'd track access times
	return w.evictFIFO(tokensNeeded)
}

func (w *ContextWindow) evictByPriority(tokensNeeded int) int {
	evicted := 0
	preserveFrom := len(w.entries) - w.config.PreserveLastN

	// Sort by priority (lowest first) while preserving order for same priority
	// We'll do multiple passes, evicting lowest priority first
	for priority := PriorityLow; priority <= PriorityCritical && evicted < tokensNeeded; priority++ {
		var remaining []ContextEntry
		for i, entry := range w.entries {
			if entry.Priority != priority || entry.Pinned || i >= preserveFrom {
				remaining = append(remaining, entry)
				continue
			}

			if evicted >= tokensNeeded {
				remaining = append(remaining, entry)
				continue
			}

			evicted += entry.TokenCount
			w.emitEvent(EventTypeEntryEvicted, &entry)
		}
		w.entries = remaining
	}

	w.tokenCount -= evicted
	return evicted
}

func (w *ContextWindow) emitEvent(eventType WindowEventType, entry *ContextEntry) {
	if w.eventHandler == nil {
		return
	}
	w.eventHandler(&WindowEvent{
		Type:       eventType,
		Timestamp:  time.Now(),
		Entry:      entry,
		TokenCount: w.tokenCount,
	})
}

// Snapshot returns a snapshot of the context window.
func (w *ContextWindow) Snapshot() *WindowSnapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	entries := make([]ContextEntry, len(w.entries))
	copy(entries, w.entries)

	return &WindowSnapshot{
		Entries:    entries,
		TokenCount: w.tokenCount,
		Timestamp:  time.Now(),
		Config:     *w.config,
	}
}

// WindowSnapshot represents a point-in-time snapshot of the context window.
type WindowSnapshot struct {
	// Entries are the context entries.
	Entries []ContextEntry `json:"entries"`
	// TokenCount is the total token count.
	TokenCount int `json:"token_count"`
	// Timestamp is when the snapshot was taken.
	Timestamp time.Time `json:"timestamp"`
	// Config is the window configuration.
	Config WindowConfig `json:"config"`
}

// RestoreFromSnapshot restores the context window from a snapshot.
func (w *ContextWindow) RestoreFromSnapshot(snapshot *WindowSnapshot) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entries = make([]ContextEntry, len(snapshot.Entries))
	copy(w.entries, snapshot.Entries)
	w.tokenCount = snapshot.TokenCount
}

// Helper functions

func estimateTokens(text string) int {
	// Simple approximation: ~4 characters per token
	return len(text) / 4
}

var idCounter int64
var idMu sync.Mutex

func generateID() string {
	idMu.Lock()
	defer idMu.Unlock()
	idCounter++
	return strings.ReplaceAll(time.Now().Format("20060102150405"), ".", "") + "_" + string(rune(idCounter))
}

// WindowStats contains statistics about the context window.
type WindowStats struct {
	// TotalEntries is the total number of entries.
	TotalEntries int `json:"total_entries"`
	// TotalTokens is the total token count.
	TotalTokens int `json:"total_tokens"`
	// AvailableTokens is the available token count.
	AvailableTokens int `json:"available_tokens"`
	// UsageRatio is the usage ratio (0-1).
	UsageRatio float64 `json:"usage_ratio"`
	// PinnedEntries is the number of pinned entries.
	PinnedEntries int `json:"pinned_entries"`
	// MessagesByRole counts messages by role.
	MessagesByRole map[string]int `json:"messages_by_role"`
	// AverageEntrySize is the average entry token count.
	AverageEntrySize float64 `json:"average_entry_size"`
}

// Stats returns statistics about the context window.
func (w *ContextWindow) Stats() *WindowStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	stats := &WindowStats{
		TotalEntries:    len(w.entries),
		TotalTokens:     w.tokenCount,
		AvailableTokens: w.config.MaxTokens - w.config.ReserveTokens - w.tokenCount,
		MessagesByRole:  make(map[string]int),
	}

	if len(w.entries) > 0 {
		stats.AverageEntrySize = float64(w.tokenCount) / float64(len(w.entries))
	}

	maxUsable := w.config.MaxTokens - w.config.ReserveTokens
	if maxUsable > 0 {
		stats.UsageRatio = float64(w.tokenCount) / float64(maxUsable)
	}

	for _, entry := range w.entries {
		stats.MessagesByRole[entry.Role]++
		if entry.Pinned {
			stats.PinnedEntries++
		}
	}

	return stats
}
