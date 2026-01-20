// Package debate provides the Lesson Banking system for reusable knowledge from AI debates.
// Implements lesson capture, categorization, retrieval, and application.
package debate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// LessonBankConfig configures the Lesson Banking system.
type LessonBankConfig struct {
	// MaxLessons is the maximum number of lessons to store
	MaxLessons int `json:"max_lessons"`
	// MinConfidence is the minimum confidence for a lesson to be stored
	MinConfidence float64 `json:"min_confidence"`
	// EnableSemanticSearch enables semantic search for lessons
	EnableSemanticSearch bool `json:"enable_semantic_search"`
	// SimilarityThreshold is the threshold for semantic similarity
	SimilarityThreshold float64 `json:"similarity_threshold"`
	// ExpirationDays is the number of days before a lesson expires
	ExpirationDays int `json:"expiration_days"`
	// EnableAutoPromotion enables automatic promotion of successful lessons
	EnableAutoPromotion bool `json:"enable_auto_promotion"`
	// PromotionThreshold is the success rate threshold for promotion
	PromotionThreshold float64 `json:"promotion_threshold"`
}

// DefaultLessonBankConfig returns a default configuration.
func DefaultLessonBankConfig() LessonBankConfig {
	return LessonBankConfig{
		MaxLessons:           10000,
		MinConfidence:        0.7,
		EnableSemanticSearch: true,
		SimilarityThreshold:  0.8,
		ExpirationDays:       90,
		EnableAutoPromotion:  true,
		PromotionThreshold:   0.8,
	}
}

// LessonCategory represents the category of a lesson.
type LessonCategory string

const (
	LessonCategoryPattern       LessonCategory = "pattern"       // Design patterns
	LessonCategoryAntiPattern   LessonCategory = "anti_pattern"  // Anti-patterns to avoid
	LessonCategoryBestPractice  LessonCategory = "best_practice" // Best practices
	LessonCategoryOptimization  LessonCategory = "optimization"  // Performance optimizations
	LessonCategorySecurity      LessonCategory = "security"      // Security considerations
	LessonCategoryRefactoring   LessonCategory = "refactoring"   // Refactoring strategies
	LessonCategoryDebugging     LessonCategory = "debugging"     // Debugging techniques
	LessonCategoryArchitecture  LessonCategory = "architecture"  // Architectural decisions
	LessonCategoryTesting       LessonCategory = "testing"       // Testing strategies
	LessonCategoryDocumentation LessonCategory = "documentation" // Documentation practices
)

// LessonTier represents the tier/importance of a lesson.
type LessonTier int

const (
	LessonTierBronze   LessonTier = iota // Basic lessons
	LessonTierSilver                     // Intermediate lessons
	LessonTierGold                       // Advanced lessons
	LessonTierPlatinum                   // Expert-level lessons
)

// Lesson represents a reusable piece of knowledge from debates.
type Lesson struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Category      LessonCategory    `json:"category"`
	Tier          LessonTier        `json:"tier"`
	Tags          []string          `json:"tags"`
	Context       LessonContext     `json:"context"`
	Content       LessonContent     `json:"content"`
	Provenance    LessonProvenance  `json:"provenance"`
	Statistics    LessonStatistics  `json:"statistics"`
	Embedding     []float64         `json:"embedding,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
}

// LessonContext describes when a lesson is applicable.
type LessonContext struct {
	Languages      []string `json:"languages,omitempty"`      // Applicable programming languages
	Frameworks     []string `json:"frameworks,omitempty"`     // Applicable frameworks
	ProblemDomains []string `json:"problem_domains,omitempty"` // Problem domains
	Preconditions  []string `json:"preconditions,omitempty"`  // Required preconditions
	Contraindications []string `json:"contraindications,omitempty"` // When NOT to apply
}

// LessonContent contains the actual lesson knowledge.
type LessonContent struct {
	Problem      string             `json:"problem"`      // The problem being addressed
	Solution     string             `json:"solution"`     // The solution approach
	Rationale    string             `json:"rationale"`    // Why this solution works
	CodeExamples []CodeExample      `json:"code_examples,omitempty"`
	TradeOffs    []TradeOff         `json:"trade_offs,omitempty"`
	References   []string           `json:"references,omitempty"`
	RelatedLessons []string         `json:"related_lessons,omitempty"`
}

// CodeExample represents a code example within a lesson.
type CodeExample struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Before      string `json:"before,omitempty"`
	After       string `json:"after"`
}

// TradeOff represents a trade-off consideration.
type TradeOff struct {
	Aspect  string `json:"aspect"`  // What aspect is being traded off
	Pros    []string `json:"pros"`  // Advantages
	Cons    []string `json:"cons"`  // Disadvantages
}

// LessonProvenance tracks the origin of a lesson.
type LessonProvenance struct {
	SourceDebateID   string    `json:"source_debate_id"`
	SourceType       string    `json:"source_type"`       // "debate", "manual", "import"
	Contributors     []string  `json:"contributors"`      // LLMs/humans that contributed
	ConsensusLevel   float64   `json:"consensus_level"`   // How much agreement there was
	ExtractedAt      time.Time `json:"extracted_at"`
	VerifiedBy       []string  `json:"verified_by,omitempty"`
}

// LessonStatistics tracks lesson usage statistics.
type LessonStatistics struct {
	ViewCount       int       `json:"view_count"`
	ApplyCount      int       `json:"apply_count"`       // Times applied
	SuccessCount    int       `json:"success_count"`     // Successful applications
	FailureCount    int       `json:"failure_count"`     // Failed applications
	FeedbackScore   float64   `json:"feedback_score"`    // Average user feedback
	FeedbackCount   int       `json:"feedback_count"`
	LastApplied     *time.Time `json:"last_applied,omitempty"`
	LastViewed      *time.Time `json:"last_viewed,omitempty"`
}

// SuccessRate calculates the success rate of the lesson.
func (s *LessonStatistics) SuccessRate() float64 {
	total := s.SuccessCount + s.FailureCount
	if total == 0 {
		return 0
	}
	return float64(s.SuccessCount) / float64(total)
}

// LessonBank manages the storage and retrieval of lessons.
type LessonBank struct {
	config           LessonBankConfig
	lessons          map[string]*Lesson
	lessonsByCategory map[LessonCategory][]*Lesson
	lessonsByTag     map[string][]*Lesson
	embedder         LessonEmbedder
	storage          LessonStorage
	mu               sync.RWMutex
}

// LessonEmbedder generates embeddings for lessons.
type LessonEmbedder interface {
	// Embed generates an embedding for the given text
	Embed(ctx context.Context, text string) ([]float64, error)
	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// LessonStorage persists lessons.
type LessonStorage interface {
	// Save saves a lesson
	Save(ctx context.Context, lesson *Lesson) error
	// Load loads a lesson by ID
	Load(ctx context.Context, id string) (*Lesson, error)
	// LoadAll loads all lessons
	LoadAll(ctx context.Context) ([]*Lesson, error)
	// Delete deletes a lesson
	Delete(ctx context.Context, id string) error
	// Query queries lessons
	Query(ctx context.Context, query LessonQuery) ([]*Lesson, error)
}

// LessonQuery represents a query for lessons.
type LessonQuery struct {
	Categories []LessonCategory `json:"categories,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
	MinTier    *LessonTier      `json:"min_tier,omitempty"`
	Languages  []string         `json:"languages,omitempty"`
	TextQuery  string           `json:"text_query,omitempty"`
	MinScore   *float64         `json:"min_score,omitempty"`
	Limit      int              `json:"limit,omitempty"`
	Offset     int              `json:"offset,omitempty"`
}

// NewLessonBank creates a new LessonBank.
func NewLessonBank(config LessonBankConfig, embedder LessonEmbedder, storage LessonStorage) *LessonBank {
	return &LessonBank{
		config:            config,
		lessons:           make(map[string]*Lesson),
		lessonsByCategory: make(map[LessonCategory][]*Lesson),
		lessonsByTag:      make(map[string][]*Lesson),
		embedder:          embedder,
		storage:           storage,
	}
}

// Initialize initializes the lesson bank.
func (lb *LessonBank) Initialize(ctx context.Context) error {
	if lb.storage == nil {
		return nil
	}

	lessons, err := lb.storage.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load lessons: %w", err)
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, lesson := range lessons {
		lb.indexLesson(lesson)
	}

	return nil
}

// AddLesson adds a new lesson to the bank.
func (lb *LessonBank) AddLesson(ctx context.Context, lesson *Lesson) error {
	if lesson.ID == "" {
		lesson.ID = lb.generateLessonID(lesson)
	}

	// Generate embedding if enabled (do this BEFORE acquiring lock to avoid blocking)
	if lb.config.EnableSemanticSearch && lb.embedder != nil {
		embedding, err := lb.embedder.Embed(ctx, lb.lessonText(lesson))
		if err == nil {
			lesson.Embedding = embedding
		}
	}

	// Set timestamps
	now := time.Now()
	lesson.CreatedAt = now
	lesson.UpdatedAt = now

	// Set expiration
	if lb.config.ExpirationDays > 0 {
		expires := now.AddDate(0, 0, lb.config.ExpirationDays)
		lesson.ExpiresAt = &expires
	}

	// Acquire lock BEFORE checking for duplicates to prevent race condition
	lb.mu.Lock()

	// Check for duplicates (now protected by lock)
	if lb.isDuplicateLocked(ctx, lesson) {
		lb.mu.Unlock()
		return fmt.Errorf("duplicate lesson detected")
	}

	// Store lesson (still holding lock)
	lb.indexLesson(lesson)
	lb.mu.Unlock()

	// Persist
	if lb.storage != nil {
		if err := lb.storage.Save(ctx, lesson); err != nil {
			return fmt.Errorf("failed to save lesson: %w", err)
		}
	}

	// Enforce max lessons
	if err := lb.enforceMaxLessons(ctx); err != nil {
		return err
	}

	return nil
}

// GetLesson retrieves a lesson by ID.
func (lb *LessonBank) GetLesson(ctx context.Context, id string) (*Lesson, error) {
	lb.mu.RLock()
	lesson := lb.lessons[id]
	lb.mu.RUnlock()

	if lesson == nil {
		return nil, fmt.Errorf("lesson not found: %s", id)
	}

	// Update view statistics
	lb.mu.Lock()
	lesson.Statistics.ViewCount++
	now := time.Now()
	lesson.Statistics.LastViewed = &now
	lb.mu.Unlock()

	return lesson, nil
}

// SearchLessons searches for relevant lessons.
func (lb *LessonBank) SearchLessons(ctx context.Context, query string, options SearchOptions) ([]*LessonSearchResult, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var results []*LessonSearchResult

	// If semantic search is enabled and we have an embedder
	if lb.config.EnableSemanticSearch && lb.embedder != nil && options.UseSemanticSearch {
		queryEmbedding, err := lb.embedder.Embed(ctx, query)
		if err == nil {
			results = lb.semanticSearch(queryEmbedding, options)
		}
	}

	// Fall back to keyword search
	if len(results) == 0 {
		results = lb.keywordSearch(query, options)
	}

	// Filter by category and tags
	results = lb.filterResults(results, options)

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	limit := options.Limit
	if limit <= 0 {
		limit = 10
	}
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchOptions configures lesson search.
type SearchOptions struct {
	UseSemanticSearch bool             `json:"use_semantic_search"`
	Categories        []LessonCategory `json:"categories,omitempty"`
	Tags              []string         `json:"tags,omitempty"`
	Languages         []string         `json:"languages,omitempty"`
	MinTier           *LessonTier      `json:"min_tier,omitempty"`
	MinScore          float64          `json:"min_score"`
	Limit             int              `json:"limit"`
}

// LessonSearchResult represents a search result.
type LessonSearchResult struct {
	Lesson      *Lesson `json:"lesson"`
	Score       float64 `json:"score"`
	MatchType   string  `json:"match_type"` // "semantic", "keyword", "tag"
	Highlights  []string `json:"highlights,omitempty"`
}

// ApplyLesson records a lesson application.
func (lb *LessonBank) ApplyLesson(ctx context.Context, lessonID string, context string) (*LessonApplication, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lesson := lb.lessons[lessonID]
	if lesson == nil {
		return nil, fmt.Errorf("lesson not found: %s", lessonID)
	}

	// Update statistics
	lesson.Statistics.ApplyCount++
	now := time.Now()
	lesson.Statistics.LastApplied = &now
	lesson.UpdatedAt = now

	application := &LessonApplication{
		ID:        generateApplicationID(),
		LessonID:  lessonID,
		Context:   context,
		AppliedAt: now,
		Status:    ApplicationStatusPending,
	}

	return application, nil
}

// LessonApplication represents an application of a lesson.
type LessonApplication struct {
	ID         string             `json:"id"`
	LessonID   string             `json:"lesson_id"`
	Context    string             `json:"context"`
	AppliedAt  time.Time          `json:"applied_at"`
	Status     ApplicationStatus  `json:"status"`
	Outcome    *ApplicationOutcome `json:"outcome,omitempty"`
}

// ApplicationStatus represents the status of a lesson application.
type ApplicationStatus string

const (
	ApplicationStatusPending   ApplicationStatus = "pending"
	ApplicationStatusSuccess   ApplicationStatus = "success"
	ApplicationStatusFailure   ApplicationStatus = "failure"
	ApplicationStatusPartial   ApplicationStatus = "partial"
)

// ApplicationOutcome represents the outcome of applying a lesson.
type ApplicationOutcome struct {
	Success     bool      `json:"success"`
	Feedback    string    `json:"feedback"`
	Score       float64   `json:"score"`
	Improvements []string `json:"improvements,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

// RecordOutcome records the outcome of a lesson application.
func (lb *LessonBank) RecordOutcome(ctx context.Context, lessonID string, outcome *ApplicationOutcome) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lesson := lb.lessons[lessonID]
	if lesson == nil {
		return fmt.Errorf("lesson not found: %s", lessonID)
	}

	// Update statistics
	if outcome.Success {
		lesson.Statistics.SuccessCount++
	} else {
		lesson.Statistics.FailureCount++
	}

	// Update feedback
	if outcome.Score > 0 {
		total := lesson.Statistics.FeedbackScore * float64(lesson.Statistics.FeedbackCount)
		lesson.Statistics.FeedbackCount++
		lesson.Statistics.FeedbackScore = (total + outcome.Score) / float64(lesson.Statistics.FeedbackCount)
	}

	lesson.UpdatedAt = time.Now()

	// Check for auto-promotion
	if lb.config.EnableAutoPromotion {
		lb.checkPromotion(lesson)
	}

	// Persist
	if lb.storage != nil {
		if err := lb.storage.Save(ctx, lesson); err != nil {
			return fmt.Errorf("failed to save lesson: %w", err)
		}
	}

	return nil
}

// ExtractLessonsFromDebate extracts lessons from a debate result.
func (lb *LessonBank) ExtractLessonsFromDebate(ctx context.Context, debate *DebateResult) ([]*Lesson, error) {
	var lessons []*Lesson

	// Extract insights from debate rounds
	for _, round := range debate.Rounds {
		lesson := lb.extractLessonFromRound(debate, round)
		if lesson != nil && lesson.Content.Solution != "" {
			// Check confidence threshold
			if debate.Consensus.Confidence >= lb.config.MinConfidence {
				lessons = append(lessons, lesson)
			}
		}
	}

	// Extract overall consensus as a lesson
	if debate.Consensus != nil && debate.Consensus.Confidence >= lb.config.MinConfidence {
		consensusLesson := lb.extractConsensusLesson(debate)
		if consensusLesson != nil {
			lessons = append(lessons, consensusLesson)
		}
	}

	// Categorize lessons
	for _, lesson := range lessons {
		lb.categorizeLesson(lesson)
	}

	// Add lessons to bank
	for _, lesson := range lessons {
		if err := lb.AddLesson(ctx, lesson); err != nil {
			// Log but don't fail on individual lesson errors
			continue
		}
	}

	return lessons, nil
}

// DebateResult represents the result of an AI debate (imported type).
type DebateResult struct {
	ID         string          `json:"id"`
	Topic      string          `json:"topic"`
	Rounds     []DebateRound   `json:"rounds"`
	Consensus  *DebateConsensus `json:"consensus,omitempty"`
	Participants []string      `json:"participants"`
	StartedAt  time.Time       `json:"started_at"`
	EndedAt    time.Time       `json:"ended_at"`
}

// DebateRound represents a round in a debate.
type DebateRound struct {
	Number      int               `json:"number"`
	Responses   []DebateResponse  `json:"responses"`
	Summary     string            `json:"summary,omitempty"`
	KeyInsights []string          `json:"key_insights,omitempty"`
}

// DebateResponse represents a response in a debate round.
type DebateResponse struct {
	Participant string   `json:"participant"`
	Content     string   `json:"content"`
	Confidence  float64  `json:"confidence"`
	Arguments   []string `json:"arguments,omitempty"`
}

// DebateConsensus represents the consensus from a debate.
type DebateConsensus struct {
	Summary     string    `json:"summary"`
	Confidence  float64   `json:"confidence"`
	KeyPoints   []string  `json:"key_points"`
	Dissents    []string  `json:"dissents,omitempty"`
}

// GetLessonsByCategory returns lessons in a category.
func (lb *LessonBank) GetLessonsByCategory(category LessonCategory) []*Lesson {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	lessons := lb.lessonsByCategory[category]
	result := make([]*Lesson, len(lessons))
	copy(result, lessons)
	return result
}

// GetLessonsByTag returns lessons with a tag.
func (lb *LessonBank) GetLessonsByTag(tag string) []*Lesson {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	lessons := lb.lessonsByTag[tag]
	result := make([]*Lesson, len(lessons))
	copy(result, lessons)
	return result
}

// GetTopLessons returns the top lessons by success rate.
func (lb *LessonBank) GetTopLessons(limit int) []*Lesson {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Collect all lessons with applications
	var active []*Lesson
	for _, lesson := range lb.lessons {
		if lesson.Statistics.ApplyCount > 0 {
			active = append(active, lesson)
		}
	}

	// Sort by success rate
	sort.Slice(active, func(i, j int) bool {
		return active[i].Statistics.SuccessRate() > active[j].Statistics.SuccessRate()
	})

	if len(active) > limit {
		active = active[:limit]
	}

	return active
}

// DeleteLesson removes a lesson from the bank.
func (lb *LessonBank) DeleteLesson(ctx context.Context, id string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lesson := lb.lessons[id]
	if lesson == nil {
		return fmt.Errorf("lesson not found: %s", id)
	}

	// Remove from indexes
	delete(lb.lessons, id)
	lb.removeFromCategoryIndex(lesson)
	lb.removeFromTagIndex(lesson)

	// Delete from storage
	if lb.storage != nil {
		if err := lb.storage.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete lesson: %w", err)
		}
	}

	return nil
}

// GetStatistics returns statistics about the lesson bank.
func (lb *LessonBank) GetStatistics() *LessonBankStatistics {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	stats := &LessonBankStatistics{
		TotalLessons:     len(lb.lessons),
		LessonsByCategory: make(map[LessonCategory]int),
		LessonsByTier:    make(map[LessonTier]int),
		TotalApplications: 0,
		SuccessfulApplications: 0,
	}

	for _, lesson := range lb.lessons {
		stats.LessonsByCategory[lesson.Category]++
		stats.LessonsByTier[lesson.Tier]++
		stats.TotalApplications += lesson.Statistics.ApplyCount
		stats.SuccessfulApplications += lesson.Statistics.SuccessCount
	}

	if stats.TotalApplications > 0 {
		stats.OverallSuccessRate = float64(stats.SuccessfulApplications) / float64(stats.TotalApplications)
	}

	return stats
}

// LessonBankStatistics represents statistics about the lesson bank.
type LessonBankStatistics struct {
	TotalLessons           int                       `json:"total_lessons"`
	LessonsByCategory      map[LessonCategory]int    `json:"lessons_by_category"`
	LessonsByTier          map[LessonTier]int        `json:"lessons_by_tier"`
	TotalApplications      int                       `json:"total_applications"`
	SuccessfulApplications int                       `json:"successful_applications"`
	OverallSuccessRate     float64                   `json:"overall_success_rate"`
}

// Helper methods

func (lb *LessonBank) generateLessonID(lesson *Lesson) string {
	hash := sha256.Sum256([]byte(lesson.Title + lesson.Content.Problem + lesson.Content.Solution))
	return hex.EncodeToString(hash[:16])
}

func (lb *LessonBank) lessonText(lesson *Lesson) string {
	var sb strings.Builder
	sb.WriteString(lesson.Title)
	sb.WriteString(" ")
	sb.WriteString(lesson.Description)
	sb.WriteString(" ")
	sb.WriteString(lesson.Content.Problem)
	sb.WriteString(" ")
	sb.WriteString(lesson.Content.Solution)
	return sb.String()
}

func (lb *LessonBank) indexLesson(lesson *Lesson) {
	lb.lessons[lesson.ID] = lesson

	// Index by category
	lb.lessonsByCategory[lesson.Category] = append(lb.lessonsByCategory[lesson.Category], lesson)

	// Index by tags
	for _, tag := range lesson.Tags {
		lb.lessonsByTag[tag] = append(lb.lessonsByTag[tag], lesson)
	}
}

func (lb *LessonBank) removeFromCategoryIndex(lesson *Lesson) {
	lessons := lb.lessonsByCategory[lesson.Category]
	for i, l := range lessons {
		if l.ID == lesson.ID {
			lb.lessonsByCategory[lesson.Category] = append(lessons[:i], lessons[i+1:]...)
			break
		}
	}
}

func (lb *LessonBank) removeFromTagIndex(lesson *Lesson) {
	for _, tag := range lesson.Tags {
		lessons := lb.lessonsByTag[tag]
		for i, l := range lessons {
			if l.ID == lesson.ID {
				lb.lessonsByTag[tag] = append(lessons[:i], lessons[i+1:]...)
				break
			}
		}
	}
}

// isDuplicate checks if a lesson is a duplicate (thread-safe version).
// This method acquires the lock internally for safe standalone use.
func (lb *LessonBank) isDuplicate(ctx context.Context, lesson *Lesson) bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.isDuplicateLocked(ctx, lesson)
}

// isDuplicateLocked checks if a lesson is a duplicate.
// IMPORTANT: Caller MUST hold lb.mu lock (read or write) before calling.
func (lb *LessonBank) isDuplicateLocked(ctx context.Context, lesson *Lesson) bool {
	// Check by ID
	if _, exists := lb.lessons[lesson.ID]; exists {
		return true
	}

	// Check by semantic similarity if enabled
	if lb.config.EnableSemanticSearch && lb.embedder != nil && len(lesson.Embedding) > 0 {
		for _, existing := range lb.lessons {
			if len(existing.Embedding) > 0 {
				similarity := cosineSimilarity(lesson.Embedding, existing.Embedding)
				if similarity >= lb.config.SimilarityThreshold {
					return true
				}
			}
		}
	}

	return false
}

func (lb *LessonBank) semanticSearch(queryEmbedding []float64, options SearchOptions) []*LessonSearchResult {
	var results []*LessonSearchResult

	for _, lesson := range lb.lessons {
		if len(lesson.Embedding) == 0 {
			continue
		}

		similarity := cosineSimilarity(queryEmbedding, lesson.Embedding)
		if similarity >= options.MinScore {
			results = append(results, &LessonSearchResult{
				Lesson:    lesson,
				Score:     similarity,
				MatchType: "semantic",
			})
		}
	}

	return results
}

func (lb *LessonBank) keywordSearch(query string, options SearchOptions) []*LessonSearchResult {
	var results []*LessonSearchResult
	query = strings.ToLower(query)
	keywords := strings.Fields(query)

	for _, lesson := range lb.lessons {
		score := 0.0
		var highlights []string

		text := strings.ToLower(lb.lessonText(lesson))

		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				score += 0.2
				highlights = append(highlights, keyword)
			}
		}

		// Boost for title matches
		if strings.Contains(strings.ToLower(lesson.Title), query) {
			score += 0.5
		}

		if score > 0 {
			results = append(results, &LessonSearchResult{
				Lesson:     lesson,
				Score:      score,
				MatchType:  "keyword",
				Highlights: highlights,
			})
		}
	}

	return results
}

func (lb *LessonBank) filterResults(results []*LessonSearchResult, options SearchOptions) []*LessonSearchResult {
	var filtered []*LessonSearchResult

	for _, result := range results {
		// Filter by category
		if len(options.Categories) > 0 {
			found := false
			for _, cat := range options.Categories {
				if result.Lesson.Category == cat {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by tags
		if len(options.Tags) > 0 {
			found := false
			for _, tag := range options.Tags {
				for _, lt := range result.Lesson.Tags {
					if strings.EqualFold(lt, tag) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by tier
		if options.MinTier != nil && result.Lesson.Tier < *options.MinTier {
			continue
		}

		// Filter by language
		if len(options.Languages) > 0 {
			found := false
			for _, lang := range options.Languages {
				for _, ll := range result.Lesson.Context.Languages {
					if strings.EqualFold(ll, lang) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found && len(result.Lesson.Context.Languages) > 0 {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

func (lb *LessonBank) checkPromotion(lesson *Lesson) {
	if lesson.Statistics.ApplyCount < 5 {
		return // Not enough applications
	}

	successRate := lesson.Statistics.SuccessRate()
	if successRate >= lb.config.PromotionThreshold && lesson.Tier < LessonTierPlatinum {
		lesson.Tier++
	}
}

func (lb *LessonBank) enforceMaxLessons(ctx context.Context) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.lessons) <= lb.config.MaxLessons {
		return nil
	}

	// Find lessons to evict (lowest success rate, oldest, expired)
	var candidates []*Lesson
	now := time.Now()

	for _, lesson := range lb.lessons {
		// Expired lessons are candidates
		if lesson.ExpiresAt != nil && now.After(*lesson.ExpiresAt) {
			candidates = append(candidates, lesson)
			continue
		}

		// Unused lessons are candidates
		if lesson.Statistics.ApplyCount == 0 && time.Since(lesson.CreatedAt) > 30*24*time.Hour {
			candidates = append(candidates, lesson)
		}
	}

	// If not enough candidates, add all remaining lessons to candidates
	excess := len(lb.lessons) - lb.config.MaxLessons
	if len(candidates) < excess {
		// Add all non-candidate lessons to the list
		for _, lesson := range lb.lessons {
			found := false
			for _, c := range candidates {
				if c.ID == lesson.ID {
					found = true
					break
				}
			}
			if !found {
				candidates = append(candidates, lesson)
			}
		}
	}

	// Sort by success rate (lowest first), then by age (oldest first)
	sort.Slice(candidates, func(i, j int) bool {
		rateI := candidates[i].Statistics.SuccessRate()
		rateJ := candidates[j].Statistics.SuccessRate()
		if rateI != rateJ {
			return rateI < rateJ
		}
		return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
	})

	// Delete excess
	for i := 0; i < excess && i < len(candidates); i++ {
		delete(lb.lessons, candidates[i].ID)
		lb.removeFromCategoryIndex(candidates[i])
		lb.removeFromTagIndex(candidates[i])

		if lb.storage != nil {
			lb.storage.Delete(ctx, candidates[i].ID)
		}
	}

	return nil
}

func (lb *LessonBank) extractLessonFromRound(debate *DebateResult, round DebateRound) *Lesson {
	if len(round.KeyInsights) == 0 {
		return nil
	}

	return &Lesson{
		Title:       fmt.Sprintf("Insight from %s (Round %d)", debate.Topic, round.Number),
		Description: round.Summary,
		Content: LessonContent{
			Problem:   debate.Topic,
			Solution:  strings.Join(round.KeyInsights, "\n"),
			Rationale: round.Summary,
		},
		Provenance: LessonProvenance{
			SourceDebateID: debate.ID,
			SourceType:     "debate",
			Contributors:   debate.Participants,
			ExtractedAt:    time.Now(),
		},
	}
}

func (lb *LessonBank) extractConsensusLesson(debate *DebateResult) *Lesson {
	return &Lesson{
		Title:       fmt.Sprintf("Consensus on %s", debate.Topic),
		Description: debate.Consensus.Summary,
		Content: LessonContent{
			Problem:   debate.Topic,
			Solution:  debate.Consensus.Summary,
			Rationale: strings.Join(debate.Consensus.KeyPoints, "\n"),
		},
		Provenance: LessonProvenance{
			SourceDebateID: debate.ID,
			SourceType:     "debate",
			Contributors:   debate.Participants,
			ConsensusLevel: debate.Consensus.Confidence,
			ExtractedAt:    time.Now(),
		},
	}
}

func (lb *LessonBank) categorizeLesson(lesson *Lesson) {
	text := strings.ToLower(lb.lessonText(lesson))

	// Simple keyword-based categorization
	if strings.Contains(text, "security") || strings.Contains(text, "vulnerability") {
		lesson.Category = LessonCategorySecurity
	} else if strings.Contains(text, "performance") || strings.Contains(text, "optimize") {
		lesson.Category = LessonCategoryOptimization
	} else if strings.Contains(text, "pattern") || strings.Contains(text, "design") {
		lesson.Category = LessonCategoryPattern
	} else if strings.Contains(text, "refactor") || strings.Contains(text, "clean") {
		lesson.Category = LessonCategoryRefactoring
	} else if strings.Contains(text, "test") {
		lesson.Category = LessonCategoryTesting
	} else if strings.Contains(text, "debug") || strings.Contains(text, "error") {
		lesson.Category = LessonCategoryDebugging
	} else if strings.Contains(text, "architecture") || strings.Contains(text, "structure") {
		lesson.Category = LessonCategoryArchitecture
	} else {
		lesson.Category = LessonCategoryBestPractice
	}
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

func generateApplicationID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:8])
}

// InMemoryLessonStorage is an in-memory implementation of LessonStorage.
type InMemoryLessonStorage struct {
	lessons map[string]*Lesson
	mu      sync.RWMutex
}

// NewInMemoryLessonStorage creates a new in-memory lesson storage.
func NewInMemoryLessonStorage() *InMemoryLessonStorage {
	return &InMemoryLessonStorage{
		lessons: make(map[string]*Lesson),
	}
}

// Save saves a lesson.
func (s *InMemoryLessonStorage) Save(ctx context.Context, lesson *Lesson) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lessons[lesson.ID] = lesson
	return nil
}

// Load loads a lesson by ID.
func (s *InMemoryLessonStorage) Load(ctx context.Context, id string) (*Lesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lesson, ok := s.lessons[id]
	if !ok {
		return nil, fmt.Errorf("lesson not found: %s", id)
	}
	return lesson, nil
}

// LoadAll loads all lessons.
func (s *InMemoryLessonStorage) LoadAll(ctx context.Context) ([]*Lesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var lessons []*Lesson
	for _, lesson := range s.lessons {
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

// Delete deletes a lesson.
func (s *InMemoryLessonStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lessons, id)
	return nil
}

// Query queries lessons.
func (s *InMemoryLessonStorage) Query(ctx context.Context, query LessonQuery) ([]*Lesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Lesson
	for _, lesson := range s.lessons {
		// Filter by category
		if len(query.Categories) > 0 {
			found := false
			for _, cat := range query.Categories {
				if lesson.Category == cat {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by tags
		if len(query.Tags) > 0 {
			found := false
			for _, tag := range query.Tags {
				for _, lt := range lesson.Tags {
					if strings.EqualFold(lt, tag) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by tier
		if query.MinTier != nil && lesson.Tier < *query.MinTier {
			continue
		}

		results = append(results, lesson)
	}

	// Apply limit/offset
	if query.Offset > 0 && query.Offset < len(results) {
		results = results[query.Offset:]
	}
	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}

	return results, nil
}

// Serialize serializes the lesson to JSON.
func (l *Lesson) Serialize() ([]byte, error) {
	return json.Marshal(l)
}

// DeserializeLesson deserializes a lesson from JSON.
func DeserializeLesson(data []byte) (*Lesson, error) {
	var lesson Lesson
	if err := json.Unmarshal(data, &lesson); err != nil {
		return nil, err
	}
	return &lesson, nil
}
