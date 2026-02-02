package database

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// ModelMetadataRepository Extended Tests
// Tests for all CRUD methods using nil pool + panic recovery.
// =============================================================================

// -----------------------------------------------------------------------------
// CreateModelMetadata Tests
// -----------------------------------------------------------------------------

func TestModelMetadataRepository_CreateModelMetadata_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	metadata := &ModelMetadata{
		ModelID:      "gpt-4",
		ModelName:    "GPT-4",
		ProviderID:   "openai",
		ProviderName: "OpenAI",
		Description:  "Latest GPT model",
		Tags:         []string{"chat", "code"},
		RawMetadata:  map[string]interface{}{"version": "latest"},
	}

	err := safeCallError(func() error {
		return repo.CreateModelMetadata(context.Background(), metadata)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_CreateModelMetadata_NilTags(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	metadata := &ModelMetadata{
		ModelID:      "gpt-4",
		ModelName:    "GPT-4",
		ProviderID:   "openai",
		ProviderName: "OpenAI",
		Tags:         nil,
		RawMetadata:  nil,
	}

	err := safeCallError(func() error {
		return repo.CreateModelMetadata(context.Background(), metadata)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// GetModelMetadata Tests
// -----------------------------------------------------------------------------

func TestModelMetadataRepository_GetModelMetadata_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, err := safeCallResult(func() (*ModelMetadata, error) {
		return repo.GetModelMetadata(context.Background(), "gpt-4")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// ListModels Tests
// -----------------------------------------------------------------------------

func TestModelMetadataRepository_ListModels_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, _, err := safeCallListResult(func() ([]*ModelMetadata, int, error) {
		return repo.ListModels(context.Background(), "", "", 10, 0)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_ListModels_WithFilters_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, _, err := safeCallListResult(func() ([]*ModelMetadata, int, error) {
		return repo.ListModels(context.Background(), "openai", "chat", 20, 5)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// SearchModels Tests
// -----------------------------------------------------------------------------

func TestModelMetadataRepository_SearchModels_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, _, err := safeCallListResult(func() ([]*ModelMetadata, int, error) {
		return repo.SearchModels(context.Background(), "gpt", 10, 0)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Benchmark Tests
// -----------------------------------------------------------------------------

func TestModelMetadataRepository_CreateBenchmark_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	score := 95.5
	benchmark := &ModelBenchmark{
		ModelID:       "gpt-4",
		BenchmarkName: "MMLU",
		Score:         &score,
		Metadata:      map[string]interface{}{"version": "2024"},
	}

	err := safeCallError(func() error {
		return repo.CreateBenchmark(context.Background(), benchmark)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_GetBenchmarks_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, err := safeCallResult(func() ([]*ModelBenchmark, error) {
		return repo.GetBenchmarks(context.Background(), "gpt-4")
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_GetBenchmarkByID_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, err := safeCallResult(func() (*ModelBenchmark, error) {
		return repo.GetBenchmarkByID(context.Background(), "bench-1")
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_UpdateBenchmark_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	benchmark := &ModelBenchmark{
		ID:       "bench-1",
		Metadata: map[string]interface{}{"updated": true},
	}

	err := safeCallError(func() error {
		return repo.UpdateBenchmark(context.Background(), benchmark)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_DeleteBenchmark_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.DeleteBenchmark(context.Background(), "bench-1")
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_DeleteBenchmarksByModelID_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, err := safeCallResult(func() (int64, error) {
		return repo.DeleteBenchmarksByModelID(context.Background(), "gpt-4")
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_ListAllBenchmarks_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, _, err := safeCallListBenchResult(func() ([]*ModelBenchmark, int, error) {
		return repo.ListAllBenchmarks(context.Background(), "", 10, 0)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_ListAllBenchmarks_WithType_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, _, err := safeCallListBenchResult(func() ([]*ModelBenchmark, int, error) {
		return repo.ListAllBenchmarks(context.Background(), "accuracy", 20, 5)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_GetTopBenchmarksByName_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, err := safeCallResult(func() ([]*ModelBenchmark, error) {
		return repo.GetTopBenchmarksByName(context.Background(), "MMLU", 10)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_CountBenchmarks_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, err := safeCallResult(func() (int64, error) {
		return repo.CountBenchmarks(context.Background())
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Refresh History Tests
// -----------------------------------------------------------------------------

func TestModelMetadataRepository_CreateRefreshHistory_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	history := &ModelsRefreshHistory{
		RefreshType:     "full",
		Status:          "completed",
		ModelsRefreshed: 100,
		ModelsFailed:    2,
		StartedAt:       time.Now(),
		Metadata:        map[string]interface{}{"source": "modelsdev"},
	}

	err := safeCallError(func() error {
		return repo.CreateRefreshHistory(context.Background(), history)
	})
	assert.Error(t, err)
}

func TestModelMetadataRepository_GetLatestRefreshHistory_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	_, err := safeCallResult(func() ([]*ModelsRefreshHistory, error) {
		return repo.GetLatestRefreshHistory(context.Background(), 10)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// UpdateProviderSyncInfo Tests
// -----------------------------------------------------------------------------

func TestModelMetadataRepository_UpdateProviderSyncInfo_NilPool(t *testing.T) {
	repo := NewModelMetadataRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.UpdateProviderSyncInfo(context.Background(), "openai", 50, 45)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Helper functions for 3-return-value functions
// -----------------------------------------------------------------------------

func safeCallListResult(fn func() ([]*ModelMetadata, int, error)) ([]*ModelMetadata, int, error) {
	var result []*ModelMetadata
	var count int
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = assert.AnError
			}
		}()
		result, count, err = fn()
	}()
	return result, count, err
}

func safeCallListBenchResult(fn func() ([]*ModelBenchmark, int, error)) ([]*ModelBenchmark, int, error) {
	var result []*ModelBenchmark
	var count int
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = assert.AnError
			}
		}()
		result, count, err = fn()
	}()
	return result, count, err
}
