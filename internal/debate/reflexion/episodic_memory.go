package reflexion

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Episode represents a single attempt at solving a task within the Reflexion framework.
// It captures the full context of what was tried, what happened, and what was learned.
type Episode struct {
	ID              string                 `json:"id"`
	SessionID       string                 `json:"session_id"`
	TurnID          string                 `json:"turn_id"`
	AgentID         string                 `json:"agent_id"`
	TaskDescription string                 `json:"task_description"`
	AttemptNumber   int                    `json:"attempt_number"`
	Code            string                 `json:"code"`
	TestResults     map[string]interface{} `json:"test_results"`
	FailureAnalysis string                 `json:"failure_analysis"`
	Reflection      *Reflection            `json:"reflection"`
	Improvement     string                 `json:"improvement"`
	Confidence      float64                `json:"confidence"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// Reflection captures the self-assessment generated after an episode completes.
type Reflection struct {
	RootCause        string    `json:"root_cause"`
	WhatWentWrong    string    `json:"what_went_wrong"`
	WhatToChangeNext string    `json:"what_to_change_next"`
	ConfidenceInFix  float64   `json:"confidence_in_fix"`
	GeneratedAt      time.Time `json:"generated_at"`
}

// EpisodicMemoryBuffer maintains an ordered buffer of episodes with indexed
// lookups by agent and session. It enforces a maximum size via FIFO eviction
// and is safe for concurrent use.
type EpisodicMemoryBuffer struct {
	episodes  []*Episode
	byAgent   map[string][]*Episode
	bySession map[string][]*Episode
	maxSize   int
	mu        sync.RWMutex
}

// NewEpisodicMemoryBuffer creates a new buffer. If maxSize is <= 0 it defaults to 1000.
func NewEpisodicMemoryBuffer(maxSize int) *EpisodicMemoryBuffer {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &EpisodicMemoryBuffer{
		episodes:  make([]*Episode, 0),
		byAgent:   make(map[string][]*Episode),
		bySession: make(map[string][]*Episode),
		maxSize:   maxSize,
	}
}

// Store adds an episode to the buffer. AgentID must not be empty. If ID is
// empty a unique identifier is generated. When the buffer is at capacity the
// oldest episode is evicted before the new one is stored.
func (b *EpisodicMemoryBuffer) Store(episode *Episode) error {
	if episode == nil {
		return fmt.Errorf("episode must not be nil")
	}
	if episode.AgentID == "" {
		return fmt.Errorf("episode AgentID must not be empty")
	}
	if episode.ID == "" {
		episode.ID = generateID()
	}
	if episode.Timestamp.IsZero() {
		episode.Timestamp = time.Now()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// FIFO eviction when at capacity.
	if len(b.episodes) >= b.maxSize {
		b.evictOldest()
	}

	b.episodes = append(b.episodes, episode)
	b.byAgent[episode.AgentID] = append(b.byAgent[episode.AgentID], episode)
	if episode.SessionID != "" {
		b.bySession[episode.SessionID] = append(b.bySession[episode.SessionID], episode)
	}
	return nil
}

// GetByAgent returns a copy of all episodes for the given agent.
func (b *EpisodicMemoryBuffer) GetByAgent(agentID string) []*Episode {
	b.mu.RLock()
	defer b.mu.RUnlock()

	src := b.byAgent[agentID]
	out := make([]*Episode, len(src))
	copy(out, src)
	return out
}

// GetBySession returns a copy of all episodes for the given session.
func (b *EpisodicMemoryBuffer) GetBySession(sessionID string) []*Episode {
	b.mu.RLock()
	defer b.mu.RUnlock()

	src := b.bySession[sessionID]
	out := make([]*Episode, len(src))
	copy(out, src)
	return out
}

// GetRecent returns the last n episodes ordered most-recent first.
func (b *EpisodicMemoryBuffer) GetRecent(n int) []*Episode {
	b.mu.RLock()
	defer b.mu.RUnlock()

	total := len(b.episodes)
	if n <= 0 || total == 0 {
		return []*Episode{}
	}
	if n > total {
		n = total
	}
	out := make([]*Episode, n)
	for i := 0; i < n; i++ {
		out[i] = b.episodes[total-1-i]
	}
	return out
}

// GetRelevant returns episodes whose task descriptions share the most keyword
// overlap with the provided taskDescription, up to limit results.
func (b *EpisodicMemoryBuffer) GetRelevant(
	taskDescription string,
	limit int,
) []*Episode {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if limit <= 0 || len(b.episodes) == 0 {
		return []*Episode{}
	}

	queryWords := tokenize(taskDescription)
	if len(queryWords) == 0 {
		return []*Episode{}
	}

	type scored struct {
		episode *Episode
		score   int
	}

	scored1 := make([]scored, 0, len(b.episodes))
	for _, ep := range b.episodes {
		s := overlapScore(queryWords, tokenize(ep.TaskDescription))
		if s > 0 {
			scored1 = append(scored1, scored{episode: ep, score: s})
		}
	}

	sort.Slice(scored1, func(i, j int) bool {
		return scored1[i].score > scored1[j].score
	})

	if limit > len(scored1) {
		limit = len(scored1)
	}
	out := make([]*Episode, limit)
	for i := 0; i < limit; i++ {
		out[i] = scored1[i].episode
	}
	return out
}

// Size returns the current number of episodes in the buffer.
func (b *EpisodicMemoryBuffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.episodes)
}

// Clear removes all episodes and resets the indexes.
func (b *EpisodicMemoryBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.episodes = make([]*Episode, 0)
	b.byAgent = make(map[string][]*Episode)
	b.bySession = make(map[string][]*Episode)
}

// GetAll returns a copy of every episode in insertion order.
func (b *EpisodicMemoryBuffer) GetAll() []*Episode {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]*Episode, len(b.episodes))
	copy(out, b.episodes)
	return out
}

// MarshalJSON serializes the buffer for database persistence.
func (b *EpisodicMemoryBuffer) MarshalJSON() ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	payload := struct {
		Episodes []*Episode `json:"episodes"`
		MaxSize  int        `json:"max_size"`
	}{
		Episodes: b.episodes,
		MaxSize:  b.maxSize,
	}
	return json.Marshal(payload)
}

// UnmarshalJSON deserializes data into the buffer and rebuilds all indexes.
func (b *EpisodicMemoryBuffer) UnmarshalJSON(data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	payload := struct {
		Episodes []*Episode `json:"episodes"`
		MaxSize  int        `json:"max_size"`
	}{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("unmarshal episodic memory: %w", err)
	}

	b.maxSize = payload.MaxSize
	if b.maxSize <= 0 {
		b.maxSize = 1000
	}
	b.episodes = payload.Episodes
	if b.episodes == nil {
		b.episodes = make([]*Episode, 0)
	}

	// Rebuild indexes.
	b.byAgent = make(map[string][]*Episode)
	b.bySession = make(map[string][]*Episode)
	for _, ep := range b.episodes {
		if ep.AgentID != "" {
			b.byAgent[ep.AgentID] = append(b.byAgent[ep.AgentID], ep)
		}
		if ep.SessionID != "" {
			b.bySession[ep.SessionID] = append(
				b.bySession[ep.SessionID], ep,
			)
		}
	}
	return nil
}

// evictOldest removes the oldest episode (index 0) and cleans up index maps.
// Caller must hold b.mu write lock.
func (b *EpisodicMemoryBuffer) evictOldest() {
	if len(b.episodes) == 0 {
		return
	}
	oldest := b.episodes[0]
	b.episodes = b.episodes[1:]

	// Clean agent index.
	if agents, ok := b.byAgent[oldest.AgentID]; ok {
		for i, ep := range agents {
			if ep == oldest {
				b.byAgent[oldest.AgentID] = append(agents[:i], agents[i+1:]...)
				break
			}
		}
		if len(b.byAgent[oldest.AgentID]) == 0 {
			delete(b.byAgent, oldest.AgentID)
		}
	}

	// Clean session index.
	if oldest.SessionID != "" {
		if sessions, ok := b.bySession[oldest.SessionID]; ok {
			for i, ep := range sessions {
				if ep == oldest {
					b.bySession[oldest.SessionID] = append(
						sessions[:i], sessions[i+1:]...,
					)
					break
				}
			}
			if len(b.bySession[oldest.SessionID]) == 0 {
				delete(b.bySession, oldest.SessionID)
			}
		}
	}
}

// generateID produces a unique identifier using timestamp and a counter.
var idCounter uint64
var idMu sync.Mutex

func generateID() string {
	idMu.Lock()
	defer idMu.Unlock()
	idCounter++
	return fmt.Sprintf(
		"ep-%d-%04d",
		time.Now().UnixNano(),
		idCounter,
	)
}

// tokenize splits text into lowercase words, filtering out short tokens.
func tokenize(text string) map[string]struct{} {
	words := strings.Fields(strings.ToLower(text))
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?\"'()[]{}")
		if len(w) > 2 {
			set[w] = struct{}{}
		}
	}
	return set
}

// overlapScore counts how many words from the query appear in the candidate.
func overlapScore(
	query map[string]struct{},
	candidate map[string]struct{},
) int {
	count := 0
	for w := range query {
		if _, ok := candidate[w]; ok {
			count++
		}
	}
	return count
}
