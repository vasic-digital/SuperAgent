package qdrant

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"
)

// BenchmarkConfigValidation benchmarks configuration validation
func BenchmarkConfigValidation(b *testing.B) {
	b.Run("ValidConfig", func(b *testing.B) {
		config := DefaultConfig()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Validate()
		}
	})

	b.Run("InvalidConfig", func(b *testing.B) {
		config := &Config{
			Host:     "",
			HTTPPort: 0,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Validate()
		}
	})
}

// BenchmarkDefaultConfig benchmarks creating default configuration
func BenchmarkDefaultConfig(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

// BenchmarkGetHTTPURL benchmarks URL generation
func BenchmarkGetHTTPURL(b *testing.B) {
	config := DefaultConfig()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetHTTPURL()
	}
}

// BenchmarkGetGRPCAddress benchmarks gRPC address generation
func BenchmarkGetGRPCAddress(b *testing.B) {
	config := DefaultConfig()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetGRPCAddress()
	}
}

// BenchmarkCollectionConfigValidation benchmarks collection config validation
func BenchmarkCollectionConfigValidation(b *testing.B) {
	b.Run("ValidConfig", func(b *testing.B) {
		config := DefaultCollectionConfig("test", 1536)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Validate()
		}
	})

	b.Run("InvalidConfig_NoName", func(b *testing.B) {
		config := &CollectionConfig{
			Name:       "",
			VectorSize: 1536,
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Validate()
		}
	})
}

// BenchmarkDefaultCollectionConfig benchmarks creating default collection configuration
func BenchmarkDefaultCollectionConfig(b *testing.B) {
	vectorSizes := []int{384, 768, 1536, 3072}

	for _, size := range vectorSizes {
		b.Run("VectorSize_"+string(rune(size)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = DefaultCollectionConfig("test", size)
			}
		})
	}
}

// BenchmarkCollectionConfigChaining benchmarks the fluent configuration API
func BenchmarkCollectionConfigChaining(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultCollectionConfig("test", 1536).
			WithDistance(DistanceCosine).
			WithOnDiskPayload().
			WithIndexingThreshold(10000).
			WithShards(2).
			WithReplication(3)
	}
}

// BenchmarkSearchOptionsCreation benchmarks creating search options
func BenchmarkSearchOptionsCreation(b *testing.B) {
	b.Run("DefaultSearchOptions", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultSearchOptions()
		}
	})

	b.Run("SearchOptionsChaining", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultSearchOptions().
				WithLimit(20).
				WithOffset(10).
				WithScoreThreshold(0.7).
				WithVectorsEnabled()
		}
	})

	b.Run("SearchOptionsWithFilter", func(b *testing.B) {
		filter := map[string]interface{}{
			"must": []map[string]interface{}{
				{"key": "category", "match": map[string]string{"value": "test"}},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultSearchOptions().WithFilter(filter)
		}
	})
}

// BenchmarkPointCreation benchmarks creating point structures
func BenchmarkPointCreation(b *testing.B) {
	vectorSizes := []int{384, 768, 1536, 3072}

	for _, size := range vectorSizes {
		b.Run("VectorSize_"+string(rune(size)), func(b *testing.B) {
			vector := generateRandomVector(size)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Point{
					ID:     "test-id",
					Vector: vector,
					Payload: map[string]interface{}{
						"text": "test content",
						"meta": "metadata",
					},
				}
			}
		})
	}
}

// BenchmarkPointJSONSerialization benchmarks JSON serialization of points
func BenchmarkPointJSONSerialization(b *testing.B) {
	vectorSizes := []int{384, 768, 1536}

	for _, size := range vectorSizes {
		point := Point{
			ID:     "test-point-id",
			Vector: generateRandomVector(size),
			Payload: map[string]interface{}{
				"text":     "This is a sample text for testing",
				"category": "test",
				"score":    0.95,
			},
		}

		b.Run("Marshal_VectorSize_"+string(rune(size)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = json.Marshal(point)
			}
		})
	}
}

// BenchmarkPointJSONDeserialization benchmarks JSON deserialization of points
func BenchmarkPointJSONDeserialization(b *testing.B) {
	vectorSizes := []int{384, 768, 1536}

	for _, size := range vectorSizes {
		point := Point{
			ID:     "test-point-id",
			Vector: generateRandomVector(size),
			Payload: map[string]interface{}{
				"text":     "This is a sample text for testing",
				"category": "test",
				"score":    0.95,
			},
		}
		data, _ := json.Marshal(point)

		b.Run("Unmarshal_VectorSize_"+string(rune(size)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var p Point
				_ = json.Unmarshal(data, &p)
			}
		})
	}
}

// BenchmarkBatchPointCreation benchmarks creating batches of points
func BenchmarkBatchPointCreation(b *testing.B) {
	batchSizes := []int{10, 100, 1000}
	vectorSize := 1536

	for _, batchSize := range batchSizes {
		b.Run("BatchSize_"+string(rune(batchSize)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				points := make([]Point, batchSize)
				for j := 0; j < batchSize; j++ {
					points[j] = Point{
						ID:     "test-id-" + string(rune(j)),
						Vector: generateRandomVector(vectorSize),
						Payload: map[string]interface{}{
							"text": "content",
						},
					}
				}
			}
		})
	}
}

// BenchmarkScoredPointCreation benchmarks creating scored point structures
func BenchmarkScoredPointCreation(b *testing.B) {
	vector := generateRandomVector(1536)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScoredPoint{
			ID:      "test-id",
			Version: 1,
			Score:   0.95,
			Payload: map[string]interface{}{
				"text": "content",
			},
			Vector: vector,
		}
	}
}

// BenchmarkCollectionInfoJSONParsing benchmarks parsing collection info responses
func BenchmarkCollectionInfoJSONParsing(b *testing.B) {
	response := `{
		"result": {
			"status": "green",
			"vectors_count": 1000000,
			"points_count": 1000000,
			"segments_count": 4
		}
	}`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp struct {
			Result struct {
				Status        string `json:"status"`
				VectorsCount  int64  `json:"vectors_count"`
				PointsCount   int64  `json:"points_count"`
				SegmentsCount int    `json:"segments_count"`
			} `json:"result"`
		}
		_ = json.Unmarshal([]byte(response), &resp)
	}
}

// BenchmarkSearchResponseParsing benchmarks parsing search responses
func BenchmarkSearchResponseParsing(b *testing.B) {
	resultCounts := []int{10, 50, 100}

	for _, count := range resultCounts {
		results := make([]ScoredPoint, count)
		for i := 0; i < count; i++ {
			results[i] = ScoredPoint{
				ID:      "id-" + string(rune(i)),
				Version: 1,
				Score:   float32(0.9) - float32(i)*0.01,
				Payload: map[string]interface{}{
					"text": "sample text content",
				},
			}
		}

		response := struct {
			Result []ScoredPoint `json:"result"`
		}{Result: results}

		data, _ := json.Marshal(response)

		b.Run("ResultCount_"+string(rune(count)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var resp struct {
					Result []ScoredPoint `json:"result"`
				}
				_ = json.Unmarshal(data, &resp)
			}
		})
	}
}

// BenchmarkRequestBodyConstruction benchmarks building request bodies
func BenchmarkRequestBodyConstruction(b *testing.B) {
	vector := generateRandomVector(1536)

	b.Run("SearchRequestBody", func(b *testing.B) {
		opts := DefaultSearchOptions()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = map[string]interface{}{
				"vector":       vector,
				"limit":        opts.Limit,
				"offset":       opts.Offset,
				"with_payload": opts.WithPayload,
				"with_vector":  opts.WithVectors,
			}
		}
	})

	b.Run("SearchRequestBodyWithFilter", func(b *testing.B) {
		opts := DefaultSearchOptions()
		filter := map[string]interface{}{
			"must": []map[string]interface{}{
				{"key": "category", "match": map[string]string{"value": "test"}},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = map[string]interface{}{
				"vector":          vector,
				"limit":           opts.Limit,
				"offset":          opts.Offset,
				"with_payload":    opts.WithPayload,
				"with_vector":     opts.WithVectors,
				"score_threshold": opts.ScoreThreshold,
				"filter":          filter,
			}
		}
	})

	b.Run("UpsertRequestBody", func(b *testing.B) {
		points := make([]Point, 10)
		for j := 0; j < 10; j++ {
			points[j] = Point{
				ID:      "id-" + string(rune(j)),
				Vector:  vector,
				Payload: map[string]interface{}{"text": "content"},
			}
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = map[string]interface{}{
				"points": points,
			}
		}
	})
}

// BenchmarkRequestBodySerialization benchmarks serializing request bodies
func BenchmarkRequestBodySerialization(b *testing.B) {
	vector := generateRandomVector(1536)
	opts := DefaultSearchOptions()

	reqBody := map[string]interface{}{
		"vector":       vector,
		"limit":        opts.Limit,
		"offset":       opts.Offset,
		"with_payload": opts.WithPayload,
		"with_vector":  opts.WithVectors,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(reqBody)
	}
}

// BenchmarkDistanceTypeString benchmarks distance type string conversion
func BenchmarkDistanceTypeString(b *testing.B) {
	distances := []Distance{DistanceCosine, DistanceEuclid, DistanceDot, DistanceManhattan}

	for _, d := range distances {
		b.Run(string(d), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = string(d)
			}
		})
	}
}

// BenchmarkConcurrentConfigAccess benchmarks concurrent access to configurations
func BenchmarkConcurrentConfigAccess(b *testing.B) {
	config := DefaultConfig()

	b.Run("ParallelGetHTTPURL", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = config.GetHTTPURL()
			}
		})
	})

	b.Run("ParallelValidate", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = config.Validate()
			}
		})
	})
}

// generateRandomVector creates a random float32 vector for testing
func generateRandomVector(size int) []float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	vector := make([]float32, size)
	for i := 0; i < size; i++ {
		vector[i] = r.Float32()*2 - 1 // Values between -1 and 1
	}
	return vector
}
