package minio

import (
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
)

// BenchmarkDefaultConfig benchmarks creating default configuration
func BenchmarkDefaultConfig(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

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

	b.Run("InvalidConfig_EmptyEndpoint", func(b *testing.B) {
		config := &Config{
			Endpoint:  "",
			AccessKey: "test",
			SecretKey: "test",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Validate()
		}
	})

	b.Run("InvalidConfig_SmallPartSize", func(b *testing.B) {
		config := DefaultConfig()
		config.PartSize = 1024 // Less than 5MB
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Validate()
		}
	})
}

// BenchmarkDefaultBucketConfig benchmarks creating default bucket configuration
func BenchmarkDefaultBucketConfig(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultBucketConfig("test-bucket")
	}
}

// BenchmarkBucketConfigChaining benchmarks the fluent bucket configuration API
func BenchmarkBucketConfigChaining(b *testing.B) {
	b.Run("SingleOption", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultBucketConfig("test-bucket").WithVersioning()
		}
	})

	b.Run("AllOptions", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultBucketConfig("test-bucket").
				WithRetention(30).
				WithVersioning().
				WithObjectLocking().
				WithPublicAccess()
		}
	})
}

// BenchmarkDefaultLifecycleRule benchmarks creating default lifecycle rules
func BenchmarkDefaultLifecycleRule(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultLifecycleRule("test-rule", 30)
	}
}

// BenchmarkLifecycleRuleChaining benchmarks the fluent lifecycle rule API
func BenchmarkLifecycleRuleChaining(b *testing.B) {
	b.Run("SingleOption", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultLifecycleRule("test-rule", 30).WithPrefix("logs/")
		}
	})

	b.Run("AllOptions", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = DefaultLifecycleRule("test-rule", 30).
				WithPrefix("logs/").
				WithNoncurrentExpiry(7)
		}
	})
}

// BenchmarkPutOptionCreation benchmarks creating put options
func BenchmarkPutOptionCreation(b *testing.B) {
	b.Run("WithContentType", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = WithContentType("application/json")
		}
	})

	b.Run("WithMetadata_Small", func(b *testing.B) {
		metadata := map[string]string{
			"key1": "value1",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = WithMetadata(metadata)
		}
	})

	b.Run("WithMetadata_Large", func(b *testing.B) {
		metadata := map[string]string{
			"key1":  "value1",
			"key2":  "value2",
			"key3":  "value3",
			"key4":  "value4",
			"key5":  "value5",
			"key6":  "value6",
			"key7":  "value7",
			"key8":  "value8",
			"key9":  "value9",
			"key10": "value10",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = WithMetadata(metadata)
		}
	})
}

// BenchmarkPutOptionApplication benchmarks applying put options
func BenchmarkPutOptionApplication(b *testing.B) {
	b.Run("SingleOption", func(b *testing.B) {
		opt := WithContentType("application/json")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opts := minio.PutObjectOptions{}
			opt(&opts)
		}
	})

	b.Run("MultipleOptions", func(b *testing.B) {
		contentTypeOpt := WithContentType("application/json")
		metadataOpt := WithMetadata(map[string]string{
			"key1": "value1",
			"key2": "value2",
		})
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opts := minio.PutObjectOptions{}
			contentTypeOpt(&opts)
			metadataOpt(&opts)
		}
	})

	b.Run("OptionsSlice", func(b *testing.B) {
		putOpts := []PutOption{
			WithContentType("application/json"),
			WithMetadata(map[string]string{"key": "value"}),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			opts := minio.PutObjectOptions{}
			for _, opt := range putOpts {
				opt(&opts)
			}
		}
	})
}

// BenchmarkObjectInfoCreation benchmarks creating ObjectInfo structures
func BenchmarkObjectInfoCreation(b *testing.B) {
	b.Run("Minimal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ObjectInfo{
				Key:  "test-key",
				Size: 1024,
			}
		}
	})

	b.Run("Full", func(b *testing.B) {
		now := time.Now()
		metadata := map[string]string{
			"custom-key": "custom-value",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ObjectInfo{
				Key:          "path/to/object.json",
				Size:         1024 * 1024,
				LastModified: now,
				ContentType:  "application/json",
				ETag:         "abc123def456",
				Metadata:     metadata,
			}
		}
	})
}

// BenchmarkBucketInfoCreation benchmarks creating BucketInfo structures
func BenchmarkBucketInfoCreation(b *testing.B) {
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BucketInfo{
			Name:         "test-bucket",
			CreationDate: now,
		}
	}
}

// BenchmarkConfigFieldAccess benchmarks accessing config fields
func BenchmarkConfigFieldAccess(b *testing.B) {
	config := DefaultConfig()

	b.Run("SingleField", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Endpoint
		}
	})

	b.Run("MultipleFields", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.Endpoint
			_ = config.AccessKey
			_ = config.SecretKey
			_ = config.UseSSL
			_ = config.Region
		}
	})
}

// BenchmarkBucketConfigFieldAccess benchmarks accessing bucket config fields
func BenchmarkBucketConfigFieldAccess(b *testing.B) {
	config := DefaultBucketConfig("test-bucket").
		WithRetention(30).
		WithVersioning().
		WithObjectLocking()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Name
		_ = config.RetentionDays
		_ = config.Versioning
		_ = config.ObjectLocking
		_ = config.Public
	}
}

// BenchmarkLifecycleRuleFieldAccess benchmarks accessing lifecycle rule fields
func BenchmarkLifecycleRuleFieldAccess(b *testing.B) {
	rule := DefaultLifecycleRule("test-rule", 30).
		WithPrefix("logs/").
		WithNoncurrentExpiry(7)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rule.ID
		_ = rule.Prefix
		_ = rule.Enabled
		_ = rule.ExpirationDays
		_ = rule.NoncurrentDays
		_ = rule.DeleteMarkerExpiry
	}
}

// BenchmarkPutObjectOptionsConstruction benchmarks constructing minio PutObjectOptions
func BenchmarkPutObjectOptionsConstruction(b *testing.B) {
	b.Run("Minimal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = minio.PutObjectOptions{}
		}
	})

	b.Run("WithPartSize", func(b *testing.B) {
		partSize := uint64(16 * 1024 * 1024)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = minio.PutObjectOptions{
				PartSize: partSize,
			}
		}
	})

	b.Run("Full", func(b *testing.B) {
		partSize := uint64(16 * 1024 * 1024)
		metadata := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = minio.PutObjectOptions{
				PartSize:     partSize,
				ContentType:  "application/json",
				UserMetadata: metadata,
			}
		}
	})
}

// BenchmarkTimeDurationComparisons benchmarks time duration comparisons used in config
func BenchmarkTimeDurationComparisons(b *testing.B) {
	b.Run("GreaterThanZero", func(b *testing.B) {
		duration := 30 * time.Second
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = duration > 0
		}
	})

	b.Run("LessThanOrEqual", func(b *testing.B) {
		duration := 30 * time.Second
		limit := time.Minute
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = duration <= limit
		}
	})
}

// BenchmarkMapCopy benchmarks copying metadata maps
func BenchmarkMapCopy(b *testing.B) {
	sizes := []int{1, 5, 10, 20}

	for _, size := range sizes {
		original := make(map[string]string, size)
		for j := 0; j < size; j++ {
			original["key"+string(rune(j))] = "value" + string(rune(j))
		}

		b.Run("Size_"+string(rune(size)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				copy := make(map[string]string, len(original))
				for k, v := range original {
					copy[k] = v
				}
			}
		})
	}
}

// BenchmarkConcurrentConfigAccess benchmarks concurrent access to configurations
func BenchmarkConcurrentConfigAccess(b *testing.B) {
	config := DefaultConfig()

	b.Run("ParallelValidate", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = config.Validate()
			}
		})
	})

	b.Run("ParallelFieldAccess", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = config.Endpoint
				_ = config.AccessKey
				_ = config.Region
			}
		})
	})
}

// BenchmarkRetentionDaysConversion benchmarks retention days handling
func BenchmarkRetentionDaysConversion(b *testing.B) {
	config := DefaultBucketConfig("test").WithRetention(30)

	b.Run("CheckPositive", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.RetentionDays > 0
		}
	})

	b.Run("CheckUnlimited", func(b *testing.B) {
		config := DefaultBucketConfig("test") // -1 is unlimited
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.RetentionDays == -1
		}
	})
}

// BenchmarkPartSizeCalculation benchmarks part size calculations
func BenchmarkPartSizeCalculation(b *testing.B) {
	sizes := []int64{
		5 * 1024 * 1024,   // 5MB
		16 * 1024 * 1024,  // 16MB
		64 * 1024 * 1024,  // 64MB
		128 * 1024 * 1024, // 128MB
	}

	for _, size := range sizes {
		b.Run("PartSize_"+string(rune(size/(1024*1024)))+"MB", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = uint64(size) // #nosec G115
			}
		})
	}
}

// BenchmarkConfigCreationWithValues benchmarks creating config with custom values
func BenchmarkConfigCreationWithValues(b *testing.B) {
	b.Run("AllFields", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &Config{
				Endpoint:            "minio.example.com:9000",
				AccessKey:           "AKIAIOSFODNN7EXAMPLE",
				SecretKey:           "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				UseSSL:              true,
				Region:              "us-west-2",
				ConnectTimeout:      30 * time.Second,
				RequestTimeout:      60 * time.Second,
				MaxRetries:          5,
				PartSize:            32 * 1024 * 1024,
				ConcurrentUploads:   8,
				HealthCheckInterval: 60 * time.Second,
			}
		}
	})
}

// BenchmarkObjectInfoConversion benchmarks converting between MinIO and local types
func BenchmarkObjectInfoConversion(b *testing.B) {
	now := time.Now()

	b.Run("FromMinioObjectInfo", func(b *testing.B) {
		minioInfo := minio.ObjectInfo{
			Key:          "path/to/object.json",
			Size:         1024 * 1024,
			LastModified: now,
			ContentType:  "application/json",
			ETag:         "abc123",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ObjectInfo{
				Key:          minioInfo.Key,
				Size:         minioInfo.Size,
				LastModified: minioInfo.LastModified,
				ContentType:  minioInfo.ContentType,
				ETag:         minioInfo.ETag,
			}
		}
	})
}

// BenchmarkSliceAllocation benchmarks slice allocations for list operations
func BenchmarkSliceAllocation(b *testing.B) {
	counts := []int{10, 100, 1000}

	for _, count := range counts {
		b.Run("Count_"+string(rune(count)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = make([]ObjectInfo, count)
			}
		})
	}
}

// BenchmarkBucketInfoSliceAllocation benchmarks slice allocations for bucket lists
func BenchmarkBucketInfoSliceAllocation(b *testing.B) {
	counts := []int{5, 20, 50}

	for _, count := range counts {
		b.Run("Count_"+string(rune(count)), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = make([]BucketInfo, count)
			}
		})
	}
}
