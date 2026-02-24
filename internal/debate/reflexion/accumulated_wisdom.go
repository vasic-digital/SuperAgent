package reflexion

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Wisdom represents a generalized insight extracted from episodes.
type Wisdom struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`      // generalized insight
	Source      string    `json:"source"`        // extracted from which episodes
	Frequency   int       `json:"frequency"`     // how often this pattern appeared
	Impact      float64   `json:"impact"`        // how much it improved outcomes (0-1)
	Domain      string    `json:"domain"`        // code, testing, architecture, etc.
	Tags        []string  `json:"tags"`          // searchable tags
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	UseCount    int       `json:"use_count"`
	SuccessRate float64   `json:"success_rate"` // how often applying this wisdom led to success
}

// AccumulatedWisdom manages cross-session learning insights.
type AccumulatedWisdom struct {
	insights []*Wisdom
	byDomain map[string][]*Wisdom
	mu       sync.RWMutex
}

// NewAccumulatedWisdom creates a new AccumulatedWisdom store.
func NewAccumulatedWisdom() *AccumulatedWisdom {
	return &AccumulatedWisdom{
		insights: make([]*Wisdom, 0),
		byDomain: make(map[string][]*Wisdom),
	}
}

// ExtractFromEpisodes analyzes episodes and extracts reusable patterns.
// Groups episodes by similar RootCause, generalizes into patterns,
// weights by frequency and impact.
func (w *AccumulatedWisdom) ExtractFromEpisodes(
	episodes []*Episode,
) ([]*Wisdom, error) {
	if len(episodes) == 0 {
		return []*Wisdom{}, nil
	}

	// Step 1: Group episodes by RootCause (from their Reflection field).
	groups := make(map[string][]*Episode)
	for _, ep := range episodes {
		if ep.Reflection == nil {
			continue
		}
		cause := strings.TrimSpace(ep.Reflection.RootCause)
		if cause == "" {
			continue
		}
		groups[cause] = append(groups[cause], ep)
	}

	var extracted []*Wisdom

	// Step 2: For each group with >= 2 episodes, create a Wisdom entry.
	for cause, group := range groups {
		if len(group) < 2 {
			continue
		}

		// 2a. Pattern = the common RootCause.
		// 2b. Frequency = count of episodes in group.
		frequency := len(group)

		// 2c. Impact = average improvement (compare confidence of later
		//     attempts vs earlier in the same session).
		impact := computeAverageImprovement(group)

		// 2d. Domain = infer from episode context.
		domain := inferDomain(group)

		// 2e. Tags = extract keywords from the pattern.
		tags := extractTags(cause)

		// 2f. Source = comma-separated list of episode IDs.
		ids := make([]string, 0, len(group))
		for _, ep := range group {
			ids = append(ids, ep.ID)
		}
		source := strings.Join(ids, ",")

		wisdom := &Wisdom{
			ID:          generateWisdomID(),
			Pattern:     cause,
			Source:      source,
			Frequency:   frequency,
			Impact:      impact,
			Domain:      domain,
			Tags:        tags,
			CreatedAt:   time.Now(),
			LastUsedAt:  time.Time{},
			UseCount:    0,
			SuccessRate: 0.0,
		}

		if err := w.Store(wisdom); err != nil {
			return nil, fmt.Errorf(
				"store extracted wisdom: %w", err,
			)
		}
		extracted = append(extracted, wisdom)
	}

	if extracted == nil {
		extracted = make([]*Wisdom, 0)
	}
	return extracted, nil
}

// GetRelevant returns wisdom entries relevant to the given task description.
// Uses keyword matching against patterns and tags.
func (w *AccumulatedWisdom) GetRelevant(
	taskDescription string,
	limit int,
) []*Wisdom {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if limit <= 0 || len(w.insights) == 0 {
		return []*Wisdom{}
	}

	queryWords := tokenize(taskDescription)
	if len(queryWords) == 0 {
		return []*Wisdom{}
	}

	type scored struct {
		wisdom *Wisdom
		score  int
	}

	results := make([]scored, 0, len(w.insights))
	for _, insight := range w.insights {
		s := wisdomOverlapScore(queryWords, insight)
		if s > 0 {
			results = append(results, scored{
				wisdom: insight,
				score:  s,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		// Tie-break: prefer higher impact.
		return results[i].wisdom.Impact > results[j].wisdom.Impact
	})

	if limit > len(results) {
		limit = len(results)
	}
	out := make([]*Wisdom, limit)
	for i := 0; i < limit; i++ {
		out[i] = results[i].wisdom
	}
	return out
}

// Store adds a wisdom entry.
func (w *AccumulatedWisdom) Store(wisdom *Wisdom) error {
	if wisdom == nil {
		return fmt.Errorf("wisdom must not be nil")
	}
	if wisdom.ID == "" {
		wisdom.ID = generateWisdomID()
	}
	if wisdom.CreatedAt.IsZero() {
		wisdom.CreatedAt = time.Now()
	}
	if wisdom.Tags == nil {
		wisdom.Tags = []string{}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.insights = append(w.insights, wisdom)
	if wisdom.Domain != "" {
		w.byDomain[wisdom.Domain] = append(
			w.byDomain[wisdom.Domain], wisdom,
		)
	}
	return nil
}

// RecordUsage tracks whether applying wisdom was successful.
func (w *AccumulatedWisdom) RecordUsage(
	wisdomID string,
	success bool,
) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, insight := range w.insights {
		if insight.ID == wisdomID {
			insight.UseCount++
			insight.LastUsedAt = time.Now()

			// Recalculate success rate as a running average.
			if success {
				insight.SuccessRate =
					(insight.SuccessRate*float64(insight.UseCount-1) + 1.0) /
						float64(insight.UseCount)
			} else {
				insight.SuccessRate =
					(insight.SuccessRate * float64(insight.UseCount-1)) /
						float64(insight.UseCount)
			}
			return nil
		}
	}
	return fmt.Errorf("wisdom with ID %q not found", wisdomID)
}

// GetAll returns all wisdom entries.
func (w *AccumulatedWisdom) GetAll() []*Wisdom {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make([]*Wisdom, len(w.insights))
	copy(out, w.insights)
	return out
}

// GetByDomain returns wisdom entries for a specific domain.
func (w *AccumulatedWisdom) GetByDomain(domain string) []*Wisdom {
	w.mu.RLock()
	defer w.mu.RUnlock()

	src := w.byDomain[domain]
	out := make([]*Wisdom, len(src))
	copy(out, src)
	return out
}

// Size returns the number of wisdom entries.
func (w *AccumulatedWisdom) Size() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.insights)
}

// MarshalJSON serializes for persistence.
func (w *AccumulatedWisdom) MarshalJSON() ([]byte, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	payload := struct {
		Insights []*Wisdom `json:"insights"`
	}{
		Insights: w.insights,
	}
	return json.Marshal(payload)
}

// UnmarshalJSON deserializes and rebuilds indexes.
func (w *AccumulatedWisdom) UnmarshalJSON(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	payload := struct {
		Insights []*Wisdom `json:"insights"`
	}{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("unmarshal accumulated wisdom: %w", err)
	}

	w.insights = payload.Insights
	if w.insights == nil {
		w.insights = make([]*Wisdom, 0)
	}

	// Rebuild domain index.
	w.byDomain = make(map[string][]*Wisdom)
	for _, insight := range w.insights {
		if insight.Tags == nil {
			insight.Tags = []string{}
		}
		if insight.Domain != "" {
			w.byDomain[insight.Domain] = append(
				w.byDomain[insight.Domain], insight,
			)
		}
	}
	return nil
}

// computeAverageImprovement calculates the average confidence improvement
// across episodes in the group. For episodes sharing a session, it compares
// later attempts against earlier ones. The result is clamped to [0, 1].
func computeAverageImprovement(episodes []*Episode) float64 {
	// Group by session to compare within-session progression.
	bySess := make(map[string][]*Episode)
	for _, ep := range episodes {
		key := ep.SessionID
		if key == "" {
			key = ep.AgentID
		}
		bySess[key] = append(bySess[key], ep)
	}

	var totalImprovement float64
	var comparisons int

	for _, sessEps := range bySess {
		if len(sessEps) < 2 {
			continue
		}
		// Sort by attempt number to compare progression.
		sorted := make([]*Episode, len(sessEps))
		copy(sorted, sessEps)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].AttemptNumber < sorted[j].AttemptNumber
		})

		for i := 1; i < len(sorted); i++ {
			diff := sorted[i].Confidence - sorted[i-1].Confidence
			if diff > 0 {
				totalImprovement += diff
			}
			comparisons++
		}
	}

	if comparisons == 0 {
		return 0.0
	}

	avg := totalImprovement / float64(comparisons)
	if avg > 1.0 {
		avg = 1.0
	}
	if avg < 0.0 {
		avg = 0.0
	}
	return avg
}

// inferDomain determines the domain based on episode content.
func inferDomain(episodes []*Episode) string {
	var (
		codeSignals         int
		testSignals         int
		architectureSignals int
		concurrencySignals  int
		performanceSignals  int
	)

	for _, ep := range episodes {
		content := strings.ToLower(ep.Code + " " + ep.FailureAnalysis)

		if containsAny(content,
			"func ", "return ", "package ", "import ",
			"class ", "def ", "var ", "const ") {
			codeSignals++
		}
		if containsAny(content,
			"test", "assert", "expect", "mock",
			"coverage", "fail") {
			testSignals++
		}
		if containsAny(content,
			"interface", "module", "service", "layer",
			"pattern", "decouple", "dependency") {
			architectureSignals++
		}
		if containsAny(content,
			"goroutine", "mutex", "channel", "deadlock",
			"race", "concurrent", "sync") {
			concurrencySignals++
		}
		if containsAny(content,
			"latency", "throughput", "memory", "cpu",
			"benchmark", "optimize", "cache") {
			performanceSignals++
		}
	}

	// Pick the domain with the strongest signal.
	max := codeSignals
	domain := "code"

	if testSignals > max {
		max = testSignals
		domain = "testing"
	}
	if architectureSignals > max {
		max = architectureSignals
		domain = "architecture"
	}
	if concurrencySignals > max {
		max = concurrencySignals
		domain = "concurrency"
	}
	if performanceSignals > max {
		domain = "performance"
	}

	return domain
}

// extractTags splits a pattern string into meaningful keyword tags.
func extractTags(pattern string) []string {
	words := strings.Fields(strings.ToLower(pattern))
	seen := make(map[string]struct{})
	tags := make([]string, 0, len(words))

	for _, w := range words {
		w = strings.Trim(w, ".,;:!?\"'()[]{}")
		if len(w) <= 2 {
			continue
		}
		// Skip common stop words.
		if isStopWord(w) {
			continue
		}
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		tags = append(tags, w)
	}
	return tags
}

// isStopWord returns true for common English stop words that do not add
// value as tags.
func isStopWord(w string) bool {
	switch w {
	case "the", "and", "for", "with", "that", "this",
		"from", "are", "was", "were", "been", "being",
		"have", "has", "had", "does", "did", "but",
		"not", "you", "all", "can", "her", "his",
		"its", "may", "our", "out", "who", "how",
		"than", "too", "very", "just", "about", "into",
		"over", "such", "also", "which", "each", "other":
		return true
	}
	return false
}

// wisdomOverlapScore computes a relevance score for a wisdom entry
// against the tokenized query words. It checks both pattern text and tags.
func wisdomOverlapScore(
	queryWords map[string]struct{},
	insight *Wisdom,
) int {
	score := 0

	// Score from pattern text.
	patternWords := tokenize(insight.Pattern)
	for w := range queryWords {
		if _, ok := patternWords[w]; ok {
			score++
		}
	}

	// Score from tags.
	for _, tag := range insight.Tags {
		if _, ok := queryWords[tag]; ok {
			score++
		}
	}

	// Score from domain match.
	if _, ok := queryWords[strings.ToLower(insight.Domain)]; ok {
		score++
	}

	return score
}

// wisdomIDCounter provides unique IDs for wisdom entries.
var wisdomIDCounter uint64
var wisdomIDMu sync.Mutex

// generateWisdomID produces a unique identifier for a wisdom entry.
func generateWisdomID() string {
	wisdomIDMu.Lock()
	defer wisdomIDMu.Unlock()
	wisdomIDCounter++
	return fmt.Sprintf(
		"wis-%d-%04d",
		time.Now().UnixNano(),
		wisdomIDCounter,
	)
}
