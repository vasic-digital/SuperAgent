package services

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/superagent/superagent/internal/database"
)

func TestModelMetadataCache(t *testing.T) {
	t.Run("SetAndGet", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		metadata := createTestMetadata()

		cache.Set(metadata.ModelID, metadata)
		result, exists := cache.Get(metadata.ModelID)

		assert.True(t, exists)
		assert.NotNil(t, result)
		assert.Equal(t, metadata.ModelID, result.ModelID)
		assert.Equal(t, metadata.ModelName, result.ModelName)
	})

	t.Run("CacheMiss", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)

		_, exists := cache.Get("non-existent")
		assert.False(t, exists)
	})

	t.Run("Delete", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		metadata := createTestMetadata()

		cache.Set(metadata.ModelID, metadata)
		cache.Delete(metadata.ModelID)

		_, exists := cache.Get(metadata.ModelID)
		assert.False(t, exists)
	})

	t.Run("Clear", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		metadata1 := createTestMetadata()
		metadata2 := createTestMetadata()
		metadata2.ModelID = "model-2"

		cache.Set(metadata1.ModelID, metadata1)
		cache.Set(metadata2.ModelID, metadata2)
		cache.Clear()

		_, exists1 := cache.Get(metadata1.ModelID)
		_, exists2 := cache.Get(metadata2.ModelID)
		assert.False(t, exists1)
		assert.False(t, exists2)
	})

	t.Run("Size", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		metadata1 := createTestMetadata()
		metadata2 := createTestMetadata()
		metadata2.ModelID = "model-2"

		cache.Set(metadata1.ModelID, metadata1)
		cache.Set(metadata2.ModelID, metadata2)

		size := cache.Size()
		assert.Equal(t, 2, size)
	})

	t.Run("Expiry", func(t *testing.T) {
		cache := NewModelMetadataCache(100 * time.Millisecond)
		metadata := createTestMetadata()

		cache.Set(metadata.ModelID, metadata)

		time.Sleep(150 * time.Millisecond)

		_, exists := cache.Get(metadata.ModelID)
		assert.False(t, exists)
	})

	t.Run("Overwrite", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		metadata1 := createTestMetadata()
		metadata2 := createTestMetadata()
		metadata2.ModelName = "Updated Name"

		cache.Set(metadata1.ModelID, metadata1)
		cache.Set(metadata2.ModelID, metadata2)

		result, exists := cache.Get(metadata1.ModelID)
		assert.True(t, exists)
		assert.Equal(t, "Updated Name", result.ModelName)
	})

	t.Run("MultipleEntries", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)

		for i := 0; i < 100; i++ {
			metadata := createTestMetadata()
			metadata.ModelID = "model-" + string(rune('0'+i%100%10)) + "-" + string(rune('0'+i/10))
			cache.Set(metadata.ModelID, metadata)
		}

		assert.Equal(t, 100, cache.Size())
	})
}

func TestModelMetadataConfig(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		config := &ModelMetadataConfig{}

		assert.NotNil(t, config)
	})

	t.Run("CustomValues", func(t *testing.T) {
		config := &ModelMetadataConfig{
			RefreshInterval:   24 * time.Hour,
			CacheTTL:          1 * time.Hour,
			DefaultBatchSize:  100,
			MaxRetries:        3,
			EnableAutoRefresh: true,
		}

		assert.Equal(t, 24*time.Hour, config.RefreshInterval)
		assert.Equal(t, 1*time.Hour, config.CacheTTL)
		assert.Equal(t, 100, config.DefaultBatchSize)
		assert.Equal(t, 3, config.MaxRetries)
		assert.True(t, config.EnableAutoRefresh)
	})

	t.Run("ZeroValues", func(t *testing.T) {
		config := &ModelMetadataConfig{
			RefreshInterval:   0,
			CacheTTL:          0,
			DefaultBatchSize:  0,
			MaxRetries:        0,
			EnableAutoRefresh: false,
		}

		assert.Equal(t, time.Duration(0), config.RefreshInterval)
		assert.Equal(t, time.Duration(0), config.CacheTTL)
		assert.Equal(t, 0, config.DefaultBatchSize)
		assert.Equal(t, 0, config.MaxRetries)
		assert.False(t, config.EnableAutoRefresh)
	})
}

func TestModelMetadataService_Constructor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("ValidConstruction", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		config := &ModelMetadataConfig{
			CacheTTL:          1 * time.Hour,
			EnableAutoRefresh: false,
		}

		service := NewModelMetadataService(nil, nil, cache, config, logger)

		assert.NotNil(t, service)
	})

	t.Run("NilCache", func(t *testing.T) {
		config := &ModelMetadataConfig{
			CacheTTL:          1 * time.Hour,
			EnableAutoRefresh: false,
		}

		service := NewModelMetadataService(nil, nil, nil, config, logger)

		assert.NotNil(t, service)
	})

	t.Run("NilConfig", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)

		service := NewModelMetadataService(nil, nil, cache, nil, logger)

		assert.NotNil(t, service)
	})
}

func TestModelMetadataCache_Concurrency(t *testing.T) {
	t.Run("ConcurrentReads", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		metadata := createTestMetadata()
		cache.Set(metadata.ModelID, metadata)

		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				_, exists := cache.Get(metadata.ModelID)
				assert.True(t, exists)
				done <- true
			}()
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})

	t.Run("ConcurrentWrites", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)

		done := make(chan bool)
		for i := 0; i < 50; i++ {
			go func(i int) {
				metadata := createTestMetadata()
				metadata.ModelID = "model-" + string(rune('0'+i%100%10)) + "-" + string(rune('0'+i/10))
				cache.Set(metadata.ModelID, metadata)
				done <- true
			}(i)
		}

		for i := 0; i < 50; i++ {
			<-done
		}

		assert.Equal(t, 50, cache.Size())
	})

	t.Run("ConcurrentReadWrite", func(t *testing.T) {
		cache := NewModelMetadataCache(1 * time.Hour)
		metadata := createTestMetadata()
		cache.Set(metadata.ModelID, metadata)

		done := make(chan bool)
		for i := 0; i < 25; i++ {
			go func() {
				_, exists := cache.Get(metadata.ModelID)
				assert.True(t, exists)
				done <- true
			}()
		}
		for i := 0; i < 25; i++ {
			go func(i int) {
				metadata := createTestMetadata()
				metadata.ModelID = "new-model-" + string(rune('0'+i%100%10)) + "-" + string(rune('0'+i/10))
				cache.Set(metadata.ModelID, metadata)
				done <- true
			}(i)
		}

		for i := 0; i < 50; i++ {
			<-done
		}
	})
}

func createTestMetadata() *database.ModelMetadata {
	ctx := 128000
	maxTokens := 4096
	pricingInput := 3.0
	pricingOutput := 15.0
	benchmarkScore := 95.5
	popularityScore := 100
	reliabilityScore := 0.95
	modelType := "chat"
	modelFamily := "claude"
	version := "20240229"

	return &database.ModelMetadata{
		ModelID:                 "claude-3-sonnet-20240229",
		ModelName:               "Claude 3 Sonnet",
		ProviderID:              "anthropic",
		ProviderName:            "Anthropic",
		Description:             "Claude 3 Sonnet is a balanced model",
		ContextWindow:           &ctx,
		MaxTokens:               &maxTokens,
		PricingInput:            &pricingInput,
		PricingOutput:           &pricingOutput,
		PricingCurrency:         "USD",
		SupportsVision:          true,
		SupportsFunctionCalling: true,
		SupportsStreaming:       true,
		SupportsJSONMode:        true,
		SupportsImageGeneration: false,
		SupportsAudio:           false,
		SupportsCodeGeneration:  true,
		SupportsReasoning:       true,
		BenchmarkScore:          &benchmarkScore,
		PopularityScore:         &popularityScore,
		ReliabilityScore:        &reliabilityScore,
		ModelType:               &modelType,
		ModelFamily:             &modelFamily,
		Version:                 &version,
		Tags:                    []string{"vision", "function-calling"},
		LastRefreshedAt:         time.Now(),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
}
