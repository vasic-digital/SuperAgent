package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRepo *ModelMetadataRepository

func setupModelMetadataTest(t *testing.T) *ModelMetadataRepository {
	config := NewTestDBConfig()
	pool, err := NewTestDBPool(config)
	require.NoError(t, err, "Failed to create test database pool")

	repo := NewModelMetadataRepository(pool, testLogger)
	testRepo = repo

	return repo
}

func teardownModelMetadataTest(t *testing.T) {
	if testRepo != nil && testRepo.pool != nil {
		testRepo.pool.Close()
	}
}

func TestMain(m *testing.M) {
	setupTestDB()
	code := m.Run()
	teardownTestDB()
	os.Exit(code)
}

func TestModelMetadataRepository_CreateModelMetadata(t *testing.T) {
	repo := setupModelMetadataTest(t)
	defer teardownModelMetadataTest(t)

	ctx := context.Background()
	ctx := context.Background()

	metadata := &ModelMetadata{
		ModelID:                 "test-model-1",
		ModelName:               "Test Model",
		ProviderID:              "test-provider",
		ProviderName:            "Test Provider",
		Description:             "A test model for unit testing",
		ContextWindow:           intPtr(128000),
		MaxTokens:               intPtr(4096),
		PricingInput:            float64Ptr(0.00001),
		PricingOutput:           float64Ptr(0.00002),
		PricingCurrency:         "USD",
		SupportsVision:          true,
		SupportsFunctionCalling: true,
		SupportsStreaming:       true,
		BenchmarkScore:          float64Ptr(95.5),
		PopularityScore:         intPtr(1000),
		ReliabilityScore:        float64Ptr(98.2),
		ModelType:               stringPtr("text-generation"),
		ModelFamily:             stringPtr("gpt"),
		Version:                 stringPtr("v1.0"),
		Tags:                    []string{"chat", "code", "fast"},
		ModelsDevURL:            stringPtr("https://models.dev/test-model-1"),
		ModelsDevID:             stringPtr("test-model-1"),
		RawMetadata:             map[string]interface{}{"test": "data"},
		LastRefreshedAt:         time.Now(),
	}

	err := testRepo.CreateModelMetadata(ctx, metadata)
	require.NoError(t, err)
	assert.NotEmpty(t, metadata.ID)
}

func TestModelMetadataRepository_GetModelMetadata(t *testing.T) {
	ctx := context.Background()

	metadata := &ModelMetadata{
		ModelID:                 "test-model-2",
		ModelName:               "Get Test Model",
		ProviderID:              "test-provider",
		ProviderName:            "Test Provider",
		Description:             "Model for testing GetModelMetadata",
		ContextWindow:           intPtr(100000),
		MaxTokens:               intPtr(2048),
		PricingInput:            float64Ptr(0.000005),
		PricingOutput:           float64Ptr(0.000015),
		PricingCurrency:         "USD",
		SupportsVision:          false,
		SupportsFunctionCalling: true,
		Tags:                    []string{"text"},
		LastRefreshedAt:         time.Now(),
	}

	err := testRepo.CreateModelMetadata(ctx, metadata)
	require.NoError(t, err)

	retrieved, err := testRepo.GetModelMetadata(ctx, "test-model-2")
	require.NoError(t, err)
	assert.Equal(t, "test-model-2", retrieved.ModelID)
	assert.Equal(t, "Get Test Model", retrieved.ModelName)
	assert.Equal(t, "test-provider", retrieved.ProviderID)
	assert.Equal(t, "Model for testing GetModelMetadata", retrieved.Description)
	assert.Equal(t, 100000, *retrieved.ContextWindow)
	assert.Equal(t, 2048, *retrieved.MaxTokens)
	assert.Equal(t, 0.000005, *retrieved.PricingInput)
	assert.Equal(t, 0.000015, *retrieved.PricingOutput)
	assert.Equal(t, 1, len(retrieved.Tags))
	assert.Equal(t, "text", retrieved.Tags[0])
}

func TestModelMetadataRepository_GetModelMetadata_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := testRepo.GetModelMetadata(ctx, "non-existent-model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model metadata not found")
}

func TestModelMetadataRepository_ListModels(t *testing.T) {
	ctx := context.Background()

	models := []*ModelMetadata{
		{
			ModelID:         "list-test-1",
			ModelName:       "List Test Model 1",
			ProviderID:      "provider-1",
			ProviderName:    "Provider 1",
			Description:     "First list test model",
			ContextWindow:   intPtr(100000),
			ModelType:       stringPtr("text"),
			LastRefreshedAt: time.Now(),
		},
		{
			ModelID:         "list-test-2",
			ModelName:       "List Test Model 2",
			ProviderID:      "provider-1",
			ProviderName:    "Provider 1",
			Description:     "Second list test model",
			ContextWindow:   intPtr(80000),
			ModelType:       stringPtr("code"),
			LastRefreshedAt: time.Now(),
		},
		{
			ModelID:         "list-test-3",
			ModelName:       "List Test Model 3",
			ProviderID:      "provider-2",
			ProviderName:    "Provider 2",
			Description:     "Third list test model",
			ContextWindow:   intPtr(120000),
			ModelType:       stringPtr("text"),
			LastRefreshedAt: time.Now(),
		},
	}

	for _, model := range models {
		err := testRepo.CreateModelMetadata(ctx, model)
		require.NoError(t, err)
	}

	retrieved, total, err := testRepo.ListModels(ctx, "provider-1", "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(retrieved))
	assert.Equal(t, 2, total)

	retrievedAll, totalAll, err := testRepo.ListModels(ctx, "", "", 100, 0)
	require.NoError(t, err)
	assert.Equal(t, 3, len(retrievedAll))
	assert.Equal(t, 3, totalAll)
}

func TestModelMetadataRepository_ListModels_WithPagination(t *testing.T) {
	ctx := context.Background()

	for i := 1; i <= 25; i++ {
		model := &ModelMetadata{
			ModelID:         fmt.Sprintf("paginated-test-%d", i),
			ModelName:       fmt.Sprintf("Paginated Test Model %d", i),
			ProviderID:      "provider-1",
			ProviderName:    "Provider 1",
			Description:     "Paginated test model",
			ModelType:       stringPtr("text"),
			LastRefreshedAt: time.Now(),
		}
		err := testRepo.CreateModelMetadata(ctx, model)
		require.NoError(t, err)
	}

	page1, total1, err := testRepo.ListModels(ctx, "", "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 10, len(page1))
	assert.Equal(t, 25, total1)

	page2, total2, err := testRepo.ListModels(ctx, "", "", 10, 10)
	require.NoError(t, err)
	assert.Equal(t, 10, len(page2))
	assert.Equal(t, 25, total2)
}

func TestModelMetadataRepository_SearchModels(t *testing.T) {
	ctx := context.Background()

	models := []*ModelMetadata{
		{
			ModelID:         "search-test-1",
			ModelName:       "Code Generation Model",
			ProviderID:      "provider-1",
			ProviderName:    "Code Provider",
			Description:     "Best for code generation tasks",
			Tags:            []string{"code", "programming", "dev"},
			LastRefreshedAt: time.Now(),
		},
		{
			ModelID:         "search-test-2",
			ModelName:       "Chat Model",
			ProviderID:      "provider-2",
			ProviderName:    "Chat Provider",
			Description:     "Perfect for chat conversations",
			Tags:            []string{"chat", "conversation"},
			LastRefreshedAt: time.Now(),
		},
		{
			ModelID:         "search-test-3",
			ModelName:       "Code Assistant",
			ProviderID:      "provider-1",
			ProviderName:    "Code Provider",
			Description:     "Assistant for coding",
			Tags:            []string{"code", "assistant"},
			LastRefreshedAt: time.Now(),
		},
	}

	for _, model := range models {
		err := testRepo.CreateModelMetadata(ctx, model)
		require.NoError(t, err)
	}

	results, total, err := testRepo.SearchModels(ctx, "code", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
	assert.GreaterOrEqual(t, total, 2)

	hasCodeGen := false
	hasCodeAssistant := false
	for _, model := range results {
		if model.ModelID == "search-test-1" {
			hasCodeGen = true
		}
		if model.ModelID == "search-test-3" {
			hasCodeAssistant = true
		}
	}
	assert.True(t, hasCodeGen || hasCodeAssistant)
}

func TestModelMetadataRepository_CreateBenchmark(t *testing.T) {
	ctx := context.Background()

	model := &ModelMetadata{
		ModelID:         "benchmark-test-model",
		ModelName:       "Benchmark Test Model",
		ProviderID:      "test-provider",
		ProviderName:    "Test Provider",
		LastRefreshedAt: time.Now(),
	}
	err := testRepo.CreateModelMetadata(ctx, model)
	require.NoError(t, err)

	benchmark := &ModelBenchmark{
		ModelID:         "benchmark-test-model",
		BenchmarkName:   "MMLU",
		BenchmarkType:   stringPtr("general"),
		Score:           float64Ptr(95.5),
		Rank:            intPtr(1),
		NormalizedScore: float64Ptr(100.0),
		BenchmarkDate:   timePtr(time.Now().AddDate(0, 0, -1)),
		Metadata:        map[string]interface{}{"source": "leaderboard"},
	}

	err = testRepo.CreateBenchmark(ctx, benchmark)
	require.NoError(t, err)
	assert.NotEmpty(t, benchmark.ID)
}

func TestModelMetadataRepository_GetBenchmarks(t *testing.T) {
	ctx := context.Background()

	model := &ModelMetadata{
		ModelID:         "benchmarks-get-test",
		ModelName:       "Get Benchmarks Model",
		ProviderID:      "test-provider",
		ProviderName:    "Test Provider",
		LastRefreshedAt: time.Now(),
	}
	err := testRepo.CreateModelMetadata(ctx, model)
	require.NoError(t, err)

	benchmarks := []*ModelBenchmark{
		{
			ModelID:       "benchmarks-get-test",
			BenchmarkName: "MMLU",
			BenchmarkType: stringPtr("general"),
			Score:         float64Ptr(95.5),
			Rank:          intPtr(1),
		},
		{
			ModelID:       "benchmarks-get-test",
			BenchmarkName: "HumanEval",
			BenchmarkType: stringPtr("code"),
			Score:         float64Ptr(88.2),
			Rank:          intPtr(2),
		},
		{
			ModelID:       "benchmarks-get-test",
			BenchmarkName: "GSM8K",
			BenchmarkType: stringPtr("math"),
			Score:         float64Ptr(92.1),
			Rank:          intPtr(1),
		},
	}

	for _, bench := range benchmarks {
		err := testRepo.CreateBenchmark(ctx, bench)
		require.NoError(t, err)
	}

	retrieved, err := testRepo.GetBenchmarks(ctx, "benchmarks-get-test")
	require.NoError(t, err)
	assert.Equal(t, 3, len(retrieved))
}

func TestModelMetadataRepository_CreateRefreshHistory(t *testing.T) {
	ctx := context.Background()

	history := &ModelsRefreshHistory{
		RefreshType:     "full",
		Status:          "completed",
		ModelsRefreshed: 100,
		ModelsFailed:    2,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		Metadata:        map[string]interface{}{"trigger": "manual"},
	}

	err := testRepo.CreateRefreshHistory(ctx, history)
	require.NoError(t, err)
	assert.NotEmpty(t, history.ID)
}

func TestModelMetadataRepository_GetLatestRefreshHistory(t *testing.T) {
	ctx := context.Background()

	histories := []*ModelsRefreshHistory{
		{
			RefreshType:     "provider",
			Status:          "completed",
			ModelsRefreshed: 50,
			ModelsFailed:    0,
			StartedAt:       time.Now().Add(-3 * time.Hour),
		},
		{
			RefreshType:     "full",
			Status:          "completed",
			ModelsRefreshed: 100,
			ModelsFailed:    2,
			StartedAt:       time.Now().Add(-2 * time.Hour),
		},
		{
			RefreshType:     "full",
			Status:          "completed",
			ModelsRefreshed: 105,
			ModelsFailed:    0,
			StartedAt:       time.Now().Add(-1 * time.Hour),
		},
	}

	for _, history := range histories {
		err := testRepo.CreateRefreshHistory(ctx, history)
		require.NoError(t, err)
	}

	retrieved, err := testRepo.GetLatestRefreshHistory(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 3, len(retrieved))

	assert.Equal(t, "full", retrieved[0].RefreshType)
	assert.Equal(t, 105, retrieved[0].ModelsRefreshed)
}

func TestModelMetadataRepository_UpdateProviderSyncInfo(t *testing.T) {
	ctx := context.Background()

	totalModels := 150
	enabledModels := 145

	err := testRepo.UpdateProviderSyncInfo(ctx, "test-provider", totalModels, enabledModels)
	require.NoError(t, err)
}

func TestModelMetadataRepository_UpdateModelMetadata(t *testing.T) {
	ctx := context.Background()

	metadata := &ModelMetadata{
		ModelID:         "update-test-model",
		ModelName:       "Original Name",
		ProviderID:      "test-provider",
		ProviderName:    "Test Provider",
		Description:     "Original description",
		ContextWindow:   intPtr(100000),
		BenchmarkScore:  float64Ptr(90.0),
		LastRefreshedAt: time.Now(),
	}

	err := testRepo.CreateModelMetadata(ctx, metadata)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	metadata.ModelName = "Updated Name"
	metadata.Description = "Updated description"
	metadata.ContextWindow = intPtr(150000)
	metadata.BenchmarkScore = float64Ptr(95.5)

	err = testRepo.CreateModelMetadata(ctx, metadata)
	require.NoError(t, err)

	retrieved, err := testRepo.GetModelMetadata(ctx, "update-test-model")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.ModelName)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, 150000, *retrieved.ContextWindow)
	assert.Equal(t, 95.5, *retrieved.BenchmarkScore)
}

func TestModelMetadataRepository_BenchmarkUpsert(t *testing.T) {
	ctx := context.Background()

	model := &ModelMetadata{
		ModelID:         "upsert-benchmark-test",
		ModelName:       "Upsert Benchmark Model",
		ProviderID:      "test-provider",
		ProviderName:    "Test Provider",
		LastRefreshedAt: time.Now(),
	}
	err := testRepo.CreateModelMetadata(ctx, model)
	require.NoError(t, err)

	benchmark1 := &ModelBenchmark{
		ModelID:       "upsert-benchmark-test",
		BenchmarkName: "MMLU",
		BenchmarkType: stringPtr("general"),
		Score:         float64Ptr(95.5),
		Rank:          intPtr(1),
	}

	err = testRepo.CreateBenchmark(ctx, benchmark1)
	require.NoError(t, err)

	benchmark2 := &ModelBenchmark{
		ModelID:       "upsert-benchmark-test",
		BenchmarkName: "MMLU",
		BenchmarkType: stringPtr("general"),
		Score:         float64Ptr(96.8),
		Rank:          intPtr(1),
	}

	err = testRepo.CreateBenchmark(ctx, benchmark2)
	require.NoError(t, err)

	retrieved, err := testRepo.GetBenchmarks(ctx, "upsert-benchmark-test")
	require.NoError(t, err)
	assert.Equal(t, 1, len(retrieved))
	assert.Equal(t, 96.8, *retrieved[0].Score)
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
