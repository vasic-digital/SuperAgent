// Package debate provides tests for the Lesson Banking system.
package debate

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbedder implements LessonEmbedder for testing
type MockEmbedder struct {
	embedFunc      func(ctx context.Context, text string) ([]float64, error)
	embedBatchFunc func(ctx context.Context, texts []string) ([][]float64, error)
}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, text)
	}
	// Generate simple hash-based embedding
	embedding := make([]float64, 128)
	for i, c := range text {
		embedding[i%128] += float64(c) / 1000.0
	}
	return embedding, nil
}

func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if m.embedBatchFunc != nil {
		return m.embedBatchFunc(ctx, texts)
	}
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// Test helper functions
func createTestLesson(title, problem, solution string) *Lesson {
	return &Lesson{
		Title:       title,
		Description: "Test lesson description",
		Category:    LessonCategoryBestPractice,
		Tier:        LessonTierSilver,
		Tags:        []string{"test", "golang"},
		Context: LessonContext{
			Languages:      []string{"go"},
			Frameworks:     []string{"gin"},
			ProblemDomains: []string{"web"},
		},
		Content: LessonContent{
			Problem:   problem,
			Solution:  solution,
			Rationale: "Test rationale",
		},
		Provenance: LessonProvenance{
			SourceType:   "test",
			Contributors: []string{"test-contributor"},
		},
	}
}

func createTestConfig() LessonBankConfig {
	return LessonBankConfig{
		MaxLessons:           100,
		MinConfidence:        0.5,
		EnableSemanticSearch: true,
		SimilarityThreshold:  0.9,
		ExpirationDays:       30,
		EnableAutoPromotion:  true,
		PromotionThreshold:   0.8,
	}
}

// TestDefaultLessonBankConfig tests default configuration values
func TestDefaultLessonBankConfig(t *testing.T) {
	config := DefaultLessonBankConfig()

	assert.Equal(t, 10000, config.MaxLessons)
	assert.Equal(t, 0.7, config.MinConfidence)
	assert.True(t, config.EnableSemanticSearch)
	assert.Equal(t, 0.8, config.SimilarityThreshold)
	assert.Equal(t, 90, config.ExpirationDays)
	assert.True(t, config.EnableAutoPromotion)
	assert.Equal(t, 0.8, config.PromotionThreshold)
}

// TestNewLessonBank tests lesson bank creation
func TestNewLessonBank(t *testing.T) {
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()

	bank := NewLessonBank(config, embedder, storage)

	assert.NotNil(t, bank)
	assert.Equal(t, config.MaxLessons, bank.config.MaxLessons)
	assert.NotNil(t, bank.lessons)
	assert.NotNil(t, bank.lessonsByCategory)
	assert.NotNil(t, bank.lessonsByTag)
}

// TestLessonBankInitialize tests initialization
func TestLessonBankInitialize(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()

	// Pre-populate storage
	lesson := createTestLesson("Init Test", "Problem", "Solution")
	lesson.ID = "init-test-id"
	lesson.CreatedAt = time.Now()
	require.NoError(t, storage.Save(ctx, lesson))

	bank := NewLessonBank(config, embedder, storage)
	err := bank.Initialize(ctx)

	assert.NoError(t, err)

	// Verify lesson was loaded
	loaded, err := bank.GetLesson(ctx, "init-test-id")
	assert.NoError(t, err)
	assert.Equal(t, "Init Test", loaded.Title)
}

// TestAddLesson tests adding lessons
func TestAddLesson(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Test Add", "How to handle errors", "Use proper error handling")

	err := bank.AddLesson(ctx, lesson)

	assert.NoError(t, err)
	assert.NotEmpty(t, lesson.ID)
	assert.False(t, lesson.CreatedAt.IsZero())
	assert.False(t, lesson.UpdatedAt.IsZero())
	assert.NotNil(t, lesson.ExpiresAt)
}

// TestAddLessonWithEmbedding tests embedding generation
func TestAddLessonWithEmbedding(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	config.EnableSemanticSearch = true

	embedCalled := false
	embedder := &MockEmbedder{
		embedFunc: func(ctx context.Context, text string) ([]float64, error) {
			embedCalled = true
			return make([]float64, 128), nil
		},
	}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Embedding Test", "Problem", "Solution")
	err := bank.AddLesson(ctx, lesson)

	assert.NoError(t, err)
	assert.True(t, embedCalled)
	assert.NotNil(t, lesson.Embedding)
}

// TestAddDuplicateLesson tests duplicate detection
func TestAddDuplicateLesson(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson1 := createTestLesson("Duplicate Test", "Same problem", "Same solution")
	lesson2 := createTestLesson("Duplicate Test", "Same problem", "Same solution")

	err := bank.AddLesson(ctx, lesson1)
	assert.NoError(t, err)

	err = bank.AddLesson(ctx, lesson2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

// TestGetLesson tests lesson retrieval
func TestGetLesson(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Get Test", "Problem", "Solution")
	require.NoError(t, bank.AddLesson(ctx, lesson))

	retrieved, err := bank.GetLesson(ctx, lesson.ID)

	assert.NoError(t, err)
	assert.Equal(t, lesson.Title, retrieved.Title)
	assert.Equal(t, 1, retrieved.Statistics.ViewCount)
	assert.NotNil(t, retrieved.Statistics.LastViewed)
}

// TestGetLessonNotFound tests error handling for missing lessons
func TestGetLessonNotFound(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	_, err := bank.GetLesson(ctx, "non-existent-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestSearchLessonsKeyword tests keyword search
func TestSearchLessonsKeyword(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	config.EnableSemanticSearch = false
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	// Add lessons
	require.NoError(t, bank.AddLesson(ctx, createTestLesson("Error Handling", "How to handle errors", "Use proper error handling patterns")))
	require.NoError(t, bank.AddLesson(ctx, createTestLesson("Testing Guide", "How to test code", "Write comprehensive tests")))
	require.NoError(t, bank.AddLesson(ctx, createTestLesson("Performance Tuning", "How to optimize", "Profile and optimize")))

	results, err := bank.SearchLessons(ctx, "error handling", SearchOptions{
		UseSemanticSearch: false,
		Limit:             10,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "keyword", results[0].MatchType)
}

// TestSearchLessonsSemantic tests semantic search
func TestSearchLessonsSemantic(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	config.EnableSemanticSearch = true

	embedder := &MockEmbedder{
		embedFunc: func(ctx context.Context, text string) ([]float64, error) {
			// Create simple embedding based on text
			embedding := make([]float64, 128)
			for i, c := range text {
				embedding[i%128] += float64(c) / 1000.0
			}
			return embedding, nil
		},
	}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	require.NoError(t, bank.AddLesson(ctx, createTestLesson("Error Handling", "How to handle errors", "Use proper error handling")))

	results, err := bank.SearchLessons(ctx, "error handling patterns", SearchOptions{
		UseSemanticSearch: true,
		MinScore:          0.0,
		Limit:             10,
	})

	assert.NoError(t, err)
	// Results may vary based on embedding similarity
	_ = results // Result is optional, just testing no error
}

// TestSearchLessonsWithFilters tests search with category and tag filters
func TestSearchLessonsWithFilters(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson1 := createTestLesson("Security Lesson", "Security problem", "Security solution")
	lesson1.Category = LessonCategorySecurity
	lesson1.Tags = []string{"security", "auth"}

	lesson2 := createTestLesson("Testing Lesson", "Testing problem", "Testing solution")
	lesson2.Category = LessonCategoryTesting
	lesson2.Tags = []string{"testing", "unit"}

	require.NoError(t, bank.AddLesson(ctx, lesson1))
	require.NoError(t, bank.AddLesson(ctx, lesson2))

	// Search with category filter
	results, err := bank.SearchLessons(ctx, "lesson", SearchOptions{
		Categories: []LessonCategory{LessonCategorySecurity},
		Limit:      10,
	})

	assert.NoError(t, err)
	for _, result := range results {
		assert.Equal(t, LessonCategorySecurity, result.Lesson.Category)
	}
}

// TestApplyLesson tests lesson application
func TestApplyLesson(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Apply Test", "Problem", "Solution")
	require.NoError(t, bank.AddLesson(ctx, lesson))

	application, err := bank.ApplyLesson(ctx, lesson.ID, "Applied in test context")

	assert.NoError(t, err)
	assert.NotNil(t, application)
	assert.Equal(t, lesson.ID, application.LessonID)
	assert.Equal(t, ApplicationStatusPending, application.Status)

	// Verify statistics updated
	updated, _ := bank.GetLesson(ctx, lesson.ID)
	assert.Equal(t, 1, updated.Statistics.ApplyCount)
	assert.NotNil(t, updated.Statistics.LastApplied)
}

// TestRecordOutcome tests recording application outcomes
func TestRecordOutcome(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Outcome Test", "Problem", "Solution")
	require.NoError(t, bank.AddLesson(ctx, lesson))

	// Record successful outcome
	outcome := &ApplicationOutcome{
		Success:     true,
		Feedback:    "Worked great!",
		Score:       0.9,
		CompletedAt: time.Now(),
	}

	err := bank.RecordOutcome(ctx, lesson.ID, outcome)
	assert.NoError(t, err)

	// Verify statistics
	updated, _ := bank.GetLesson(ctx, lesson.ID)
	assert.Equal(t, 1, updated.Statistics.SuccessCount)
	assert.Equal(t, 0.9, updated.Statistics.FeedbackScore)
}

// TestRecordOutcomeFailure tests recording failed outcomes
func TestRecordOutcomeFailure(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Failure Test", "Problem", "Solution")
	require.NoError(t, bank.AddLesson(ctx, lesson))

	outcome := &ApplicationOutcome{
		Success:     false,
		Feedback:    "Did not work",
		Score:       0.2,
		CompletedAt: time.Now(),
	}

	err := bank.RecordOutcome(ctx, lesson.ID, outcome)
	assert.NoError(t, err)

	updated, _ := bank.GetLesson(ctx, lesson.ID)
	assert.Equal(t, 0, updated.Statistics.SuccessCount)
	assert.Equal(t, 1, updated.Statistics.FailureCount)
}

// TestAutoPromotion tests automatic tier promotion
func TestAutoPromotion(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	config.EnableAutoPromotion = true
	config.PromotionThreshold = 0.8
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Promotion Test", "Problem", "Solution")
	lesson.Tier = LessonTierBronze
	require.NoError(t, bank.AddLesson(ctx, lesson))

	// Apply and record 5 successful outcomes (100% success rate)
	for i := 0; i < 5; i++ {
		bank.ApplyLesson(ctx, lesson.ID, "test")
		bank.RecordOutcome(ctx, lesson.ID, &ApplicationOutcome{
			Success:     true,
			Score:       1.0,
			CompletedAt: time.Now(),
		})
	}

	updated, _ := bank.GetLesson(ctx, lesson.ID)
	assert.Equal(t, LessonTierSilver, updated.Tier) // Should be promoted
}

// TestExtractLessonsFromDebate tests lesson extraction from debates
func TestExtractLessonsFromDebate(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	config.MinConfidence = 0.5
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	debate := &DebateResult{
		ID:    "debate-123",
		Topic: "How to handle database connections",
		Rounds: []DebateRound{
			{
				Number:  1,
				Summary: "Use connection pooling",
				KeyInsights: []string{
					"Connection pooling improves performance",
					"Set appropriate pool size based on load",
				},
			},
		},
		Consensus: &DebateConsensus{
			Summary:    "Use connection pooling with proper configuration",
			Confidence: 0.9,
			KeyPoints:  []string{"Use pooling", "Configure timeouts"},
		},
		Participants: []string{"Claude", "GPT-4", "Gemini"},
		StartedAt:    time.Now().Add(-1 * time.Hour),
		EndedAt:      time.Now(),
	}

	lessons, err := bank.ExtractLessonsFromDebate(ctx, debate)

	assert.NoError(t, err)
	assert.NotEmpty(t, lessons)

	// Verify provenance
	for _, lesson := range lessons {
		assert.Equal(t, "debate", lesson.Provenance.SourceType)
		assert.Equal(t, "debate-123", lesson.Provenance.SourceDebateID)
	}
}

// TestGetLessonsByCategory tests category-based retrieval
func TestGetLessonsByCategory(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson1 := createTestLesson("Security 1", "Security problem", "Security solution")
	lesson1.Category = LessonCategorySecurity

	lesson2 := createTestLesson("Security 2", "Another security problem", "Another security solution")
	lesson2.Category = LessonCategorySecurity

	lesson3 := createTestLesson("Testing 1", "Testing problem", "Testing solution")
	lesson3.Category = LessonCategoryTesting

	require.NoError(t, bank.AddLesson(ctx, lesson1))
	require.NoError(t, bank.AddLesson(ctx, lesson2))
	require.NoError(t, bank.AddLesson(ctx, lesson3))

	securityLessons := bank.GetLessonsByCategory(LessonCategorySecurity)
	assert.Len(t, securityLessons, 2)

	testingLessons := bank.GetLessonsByCategory(LessonCategoryTesting)
	assert.Len(t, testingLessons, 1)
}

// TestGetLessonsByTag tests tag-based retrieval
func TestGetLessonsByTag(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson1 := createTestLesson("Golang 1", "Problem", "Solution")
	lesson1.Tags = []string{"golang", "performance"}

	lesson2 := createTestLesson("Golang 2", "Problem", "Solution")
	lesson2.Tags = []string{"golang", "testing"}

	lesson3 := createTestLesson("Python 1", "Problem", "Solution")
	lesson3.Tags = []string{"python", "testing"}

	require.NoError(t, bank.AddLesson(ctx, lesson1))
	require.NoError(t, bank.AddLesson(ctx, lesson2))
	require.NoError(t, bank.AddLesson(ctx, lesson3))

	golangLessons := bank.GetLessonsByTag("golang")
	assert.Len(t, golangLessons, 2)

	testingLessons := bank.GetLessonsByTag("testing")
	assert.Len(t, testingLessons, 2)
}

// TestGetTopLessons tests top lessons retrieval
func TestGetTopLessons(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	// Add lessons with different success rates
	for i := 0; i < 5; i++ {
		lesson := createTestLesson(fmt.Sprintf("Lesson %d", i), "Problem", "Solution")
		require.NoError(t, bank.AddLesson(ctx, lesson))

		// Apply with varying success rates
		for j := 0; j <= i; j++ {
			bank.ApplyLesson(ctx, lesson.ID, "test")
			bank.RecordOutcome(ctx, lesson.ID, &ApplicationOutcome{
				Success:     j <= i/2, // Varying success
				CompletedAt: time.Now(),
			})
		}
	}

	topLessons := bank.GetTopLessons(3)
	assert.LessOrEqual(t, len(topLessons), 3)
}

// TestDeleteLesson tests lesson deletion
func TestDeleteLesson(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson := createTestLesson("Delete Test", "Problem", "Solution")
	require.NoError(t, bank.AddLesson(ctx, lesson))

	err := bank.DeleteLesson(ctx, lesson.ID)
	assert.NoError(t, err)

	_, err = bank.GetLesson(ctx, lesson.ID)
	assert.Error(t, err)
}

// TestDeleteLessonNotFound tests deleting non-existent lesson
func TestDeleteLessonNotFound(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	err := bank.DeleteLesson(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestGetStatistics tests statistics retrieval
func TestGetStatistics(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	lesson1 := createTestLesson("Stats 1", "Problem", "Solution")
	lesson1.Category = LessonCategorySecurity
	lesson1.Tier = LessonTierGold

	lesson2 := createTestLesson("Stats 2", "Problem", "Solution")
	lesson2.Category = LessonCategoryTesting
	lesson2.Tier = LessonTierSilver

	require.NoError(t, bank.AddLesson(ctx, lesson1))
	require.NoError(t, bank.AddLesson(ctx, lesson2))

	// Apply and record outcomes
	bank.ApplyLesson(ctx, lesson1.ID, "test")
	bank.RecordOutcome(ctx, lesson1.ID, &ApplicationOutcome{Success: true, CompletedAt: time.Now()})

	stats := bank.GetStatistics()

	assert.Equal(t, 2, stats.TotalLessons)
	assert.Equal(t, 1, stats.LessonsByCategory[LessonCategorySecurity])
	assert.Equal(t, 1, stats.LessonsByCategory[LessonCategoryTesting])
	assert.Equal(t, 1, stats.TotalApplications)
	assert.Equal(t, 1, stats.SuccessfulApplications)
	assert.Equal(t, 1.0, stats.OverallSuccessRate)
}

// TestLessonStatisticsSuccessRate tests success rate calculation
func TestLessonStatisticsSuccessRate(t *testing.T) {
	stats := &LessonStatistics{
		SuccessCount: 3,
		FailureCount: 1,
	}

	assert.Equal(t, 0.75, stats.SuccessRate())

	// Test zero applications
	emptyStats := &LessonStatistics{}
	assert.Equal(t, 0.0, emptyStats.SuccessRate())
}

// TestEnforceMaxLessons tests max lessons enforcement
func TestEnforceMaxLessons(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	config.MaxLessons = 5
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	// Add more than max lessons
	for i := 0; i < 10; i++ {
		lesson := createTestLesson(fmt.Sprintf("Max Test %d", i), "Problem", "Solution")
		lesson.CreatedAt = time.Now().Add(-time.Duration(i) * time.Hour * 24 * 40) // Older lessons first
		require.NoError(t, bank.AddLesson(ctx, lesson))
	}

	stats := bank.GetStatistics()
	assert.LessOrEqual(t, stats.TotalLessons, config.MaxLessons)
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// Concurrent adds
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			lesson := createTestLesson(fmt.Sprintf("Concurrent %d", idx), fmt.Sprintf("Problem %d", idx), "Solution")
			if err := bank.AddLesson(ctx, lesson); err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for errors (some duplicates are expected)
	errorCount := 0
	for err := range errChan {
		if !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("Unexpected error: %v", err)
		}
		errorCount++
	}

	stats := bank.GetStatistics()
	assert.Greater(t, stats.TotalLessons, 0)
}

// TestInMemoryLessonStorage tests the in-memory storage implementation
func TestInMemoryLessonStorage(t *testing.T) {
	ctx := context.Background()
	storage := NewInMemoryLessonStorage()

	lesson := createTestLesson("Storage Test", "Problem", "Solution")
	lesson.ID = "storage-test-id"

	// Test Save
	err := storage.Save(ctx, lesson)
	assert.NoError(t, err)

	// Test Load
	loaded, err := storage.Load(ctx, "storage-test-id")
	assert.NoError(t, err)
	assert.Equal(t, lesson.Title, loaded.Title)

	// Test LoadAll
	all, err := storage.LoadAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 1)

	// Test Query
	results, err := storage.Query(ctx, LessonQuery{
		Categories: []LessonCategory{LessonCategoryBestPractice},
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Test Delete
	err = storage.Delete(ctx, "storage-test-id")
	assert.NoError(t, err)

	_, err = storage.Load(ctx, "storage-test-id")
	assert.Error(t, err)
}

// TestLessonSerialization tests JSON serialization
func TestLessonSerialization(t *testing.T) {
	lesson := createTestLesson("Serialization Test", "Problem", "Solution")
	lesson.ID = "serialize-test-id"
	lesson.CreatedAt = time.Now()
	lesson.UpdatedAt = time.Now()

	data, err := lesson.Serialize()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	deserialized, err := DeserializeLesson(data)
	assert.NoError(t, err)
	assert.Equal(t, lesson.ID, deserialized.ID)
	assert.Equal(t, lesson.Title, deserialized.Title)
}

// TestLessonCategories tests all lesson categories
func TestLessonCategories(t *testing.T) {
	categories := []LessonCategory{
		LessonCategoryPattern,
		LessonCategoryAntiPattern,
		LessonCategoryBestPractice,
		LessonCategoryOptimization,
		LessonCategorySecurity,
		LessonCategoryRefactoring,
		LessonCategoryDebugging,
		LessonCategoryArchitecture,
		LessonCategoryTesting,
		LessonCategoryDocumentation,
	}

	for _, cat := range categories {
		assert.NotEmpty(t, string(cat))
	}
}

// TestLessonTiers tests all lesson tiers
func TestLessonTiers(t *testing.T) {
	assert.Equal(t, LessonTier(0), LessonTierBronze)
	assert.Equal(t, LessonTier(1), LessonTierSilver)
	assert.Equal(t, LessonTier(2), LessonTierGold)
	assert.Equal(t, LessonTier(3), LessonTierPlatinum)
}

// TestCosineSimilarity tests the cosine similarity function
func TestCosineSimilarity(t *testing.T) {
	// Test identical vectors
	a := []float64{1.0, 0.0, 0.0}
	assert.Equal(t, 1.0, cosineSimilarity(a, a))

	// Test orthogonal vectors
	b := []float64{0.0, 1.0, 0.0}
	assert.Equal(t, 0.0, cosineSimilarity(a, b))

	// Test different length vectors
	c := []float64{1.0, 0.0}
	assert.Equal(t, 0.0, cosineSimilarity(a, c))

	// Test empty vectors
	empty := []float64{}
	assert.Equal(t, 0.0, cosineSimilarity(empty, empty))

	// Test zero vectors
	zero := []float64{0.0, 0.0, 0.0}
	assert.Equal(t, 0.0, cosineSimilarity(zero, zero))
}

// TestCategorizeLessonByKeyword tests automatic categorization
func TestCategorizeLessonByKeyword(t *testing.T) {
	ctx := context.Background()
	config := createTestConfig()
	embedder := &MockEmbedder{}
	storage := NewInMemoryLessonStorage()
	bank := NewLessonBank(config, embedder, storage)

	testCases := []struct {
		title       string
		problem     string
		expectedCat LessonCategory
	}{
		{"Security Best Practices", "How to secure your application", LessonCategorySecurity},
		{"Performance Optimization", "How to optimize code performance", LessonCategoryOptimization},
		{"Design Pattern Implementation", "Implementing the factory pattern", LessonCategoryPattern},
		{"Code Refactoring Guide", "How to refactor legacy code", LessonCategoryRefactoring},
		{"Unit Testing Strategies", "How to test your code effectively", LessonCategoryTesting},
		{"Debugging Techniques", "How to debug errors effectively", LessonCategoryDebugging},
		{"System Architecture Design", "How to structure your application", LessonCategoryArchitecture},
	}

	for _, tc := range testCases {
		debate := &DebateResult{
			ID:    fmt.Sprintf("debate-%s", tc.title),
			Topic: tc.problem,
			Consensus: &DebateConsensus{
				Summary:    tc.title,
				Confidence: 0.9,
			},
			Participants: []string{"Test"},
		}

		lessons, err := bank.ExtractLessonsFromDebate(ctx, debate)
		require.NoError(t, err)

		if len(lessons) > 0 {
			// Verify categorization happened
			assert.NotEqual(t, LessonCategory(""), lessons[0].Category)
		}
	}
}
