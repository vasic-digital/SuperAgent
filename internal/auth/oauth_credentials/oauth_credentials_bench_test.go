package oauth_credentials

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkNeedsRefresh benchmarks the token refresh check function
func BenchmarkNeedsRefresh(b *testing.B) {
	testCases := []struct {
		name      string
		expiresAt int64
	}{
		{"NoExpiration", 0},
		{"FutureExpiration", time.Now().Add(2 * time.Hour).UnixMilli()},
		{"NeedsRefresh", time.Now().Add(5 * time.Minute).UnixMilli()},
		{"AlreadyExpired", time.Now().Add(-1 * time.Hour).UnixMilli()},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = NeedsRefresh(tc.expiresAt)
			}
		})
	}
}

// BenchmarkIsExpired benchmarks the token expiration check function
func BenchmarkIsExpired(b *testing.B) {
	testCases := []struct {
		name      string
		expiresAt int64
	}{
		{"NoExpiration", 0},
		{"FutureExpiration", time.Now().Add(2 * time.Hour).UnixMilli()},
		{"JustExpired", time.Now().Add(-1 * time.Millisecond).UnixMilli()},
		{"LongExpired", time.Now().Add(-24 * time.Hour).UnixMilli()},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = IsExpired(tc.expiresAt)
			}
		})
	}
}

// BenchmarkOAuthCredentialReaderCreation benchmarks creating a new credential reader
func BenchmarkOAuthCredentialReaderCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOAuthCredentialReader()
	}
}

// BenchmarkGetGlobalReader benchmarks getting the global reader singleton
func BenchmarkGetGlobalReader(b *testing.B) {
	// Ensure singleton is initialized
	_ = GetGlobalReader()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetGlobalReader()
	}
}

// BenchmarkGetClaudeCredentialsPath benchmarks the path generation function
func BenchmarkGetClaudeCredentialsPath(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetClaudeCredentialsPath()
	}
}

// BenchmarkGetQwenCredentialsPath benchmarks the path generation function
func BenchmarkGetQwenCredentialsPath(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetQwenCredentialsPath()
	}
}

// BenchmarkClaudeCredentialsJSONParsing benchmarks JSON parsing of Claude credentials
func BenchmarkClaudeCredentialsJSONParsing(b *testing.B) {
	testData := []byte(`{
		"claudeAiOauth": {
			"accessToken": "test-access-token-12345678901234567890",
			"refreshToken": "test-refresh-token-12345678901234567890",
			"expiresAt": 1709251200000,
			"scopes": ["read", "write", "admin"],
			"subscriptionType": "pro",
			"rateLimitTier": "tier2"
		}
	}`)

	b.Run("Unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var creds ClaudeOAuthCredentials
			_ = json.Unmarshal(testData, &creds)
		}
	})

	b.Run("Marshal", func(b *testing.B) {
		creds := ClaudeOAuthCredentials{
			ClaudeAiOauth: &ClaudeAiOauth{
				AccessToken:      "test-access-token-12345678901234567890",
				RefreshToken:     "test-refresh-token-12345678901234567890",
				ExpiresAt:        1709251200000,
				Scopes:           []string{"read", "write", "admin"},
				SubscriptionType: "pro",
				RateLimitTier:    "tier2",
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(creds)
		}
	})
}

// BenchmarkQwenCredentialsJSONParsing benchmarks JSON parsing of Qwen credentials
func BenchmarkQwenCredentialsJSONParsing(b *testing.B) {
	testData := []byte(`{
		"access_token": "test-access-token-12345678901234567890",
		"refresh_token": "test-refresh-token-12345678901234567890",
		"id_token": "test-id-token-12345678901234567890",
		"expiry_date": 1709251200000,
		"token_type": "Bearer",
		"resource_url": "https://dashscope.aliyuncs.com"
	}`)

	b.Run("Unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var creds QwenOAuthCredentials
			_ = json.Unmarshal(testData, &creds)
		}
	})

	b.Run("Marshal", func(b *testing.B) {
		creds := QwenOAuthCredentials{
			AccessToken:  "test-access-token-12345678901234567890",
			RefreshToken: "test-refresh-token-12345678901234567890",
			IDToken:      "test-id-token-12345678901234567890",
			ExpiryDate:   1709251200000,
			TokenType:    "Bearer",
			ResourceURL:  "https://dashscope.aliyuncs.com",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(creds)
		}
	})
}

// BenchmarkTokenRefresherCreation benchmarks creating a new token refresher
func BenchmarkTokenRefresherCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTokenRefresher()
	}
}

// BenchmarkGetGlobalRefresher benchmarks getting the global refresher singleton
func BenchmarkGetGlobalRefresher(b *testing.B) {
	// Ensure singleton is initialized
	_ = GetGlobalRefresher()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetGlobalRefresher()
	}
}

// BenchmarkClearCache benchmarks clearing the credential cache
func BenchmarkClearCache(b *testing.B) {
	reader := NewOAuthCredentialReader()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.ClearCache()
	}
}

// BenchmarkOAuthEnabledCheck benchmarks the OAuth enabled check functions
func BenchmarkOAuthEnabledCheck(b *testing.B) {
	b.Run("IsClaudeOAuthEnabled", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = IsClaudeOAuthEnabled()
		}
	})
}

// BenchmarkReadCredentialsFromFile benchmarks reading credentials from a temp file
func BenchmarkReadCredentialsFromFile(b *testing.B) {
	// Create a temporary credentials file
	tempDir := b.TempDir()
	credPath := filepath.Join(tempDir, "credentials.json")

	validCreds := ClaudeOAuthCredentials{
		ClaudeAiOauth: &ClaudeAiOauth{
			AccessToken:      "test-access-token",
			RefreshToken:     "test-refresh-token",
			ExpiresAt:        time.Now().Add(2 * time.Hour).UnixMilli(),
			Scopes:           []string{"read", "write"},
			SubscriptionType: "pro",
			RateLimitTier:    "tier2",
		},
	}

	data, _ := json.MarshalIndent(validCreds, "", "  ")
	_ = os.WriteFile(credPath, data, 0600)

	b.Run("FileRead", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = os.ReadFile(credPath)
		}
	})

	b.Run("FileReadAndParse", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data, err := os.ReadFile(credPath)
			if err != nil {
				continue
			}
			var creds ClaudeOAuthCredentials
			_ = json.Unmarshal(data, &creds)
		}
	})
}

// BenchmarkCLIRefresherCreation benchmarks creating a CLI refresher
func BenchmarkCLIRefresherCreation(b *testing.B) {
	b.Run("WithDefaultConfig", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewCLIRefresher(nil)
		}
	})

	b.Run("WithCustomConfig", func(b *testing.B) {
		config := &CLIRefreshConfig{
			QwenCLIPath:        "/usr/local/bin/qwen",
			RefreshTimeout:     30 * time.Second,
			MinRefreshInterval: 30 * time.Second,
			MaxRetries:         5,
			RetryDelay:         2 * time.Second,
			Prompt:             "test",
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewCLIRefresher(config)
		}
	})
}

// BenchmarkDefaultCLIRefreshConfig benchmarks creating default CLI refresh config
func BenchmarkDefaultCLIRefreshConfig(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultCLIRefreshConfig()
	}
}

// BenchmarkCredentialValidation benchmarks validating credential structures
func BenchmarkCredentialValidation(b *testing.B) {
	b.Run("ClaudeCredentials_Valid", func(b *testing.B) {
		creds := &ClaudeOAuthCredentials{
			ClaudeAiOauth: &ClaudeAiOauth{
				AccessToken:  "test-token",
				RefreshToken: "test-refresh",
				ExpiresAt:    time.Now().Add(time.Hour).UnixMilli(),
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Validate by checking fields
			_ = creds.ClaudeAiOauth != nil &&
				creds.ClaudeAiOauth.AccessToken != "" &&
				!IsExpired(creds.ClaudeAiOauth.ExpiresAt)
		}
	})

	b.Run("QwenCredentials_Valid", func(b *testing.B) {
		creds := &QwenOAuthCredentials{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
			ExpiryDate:   time.Now().Add(time.Hour).UnixMilli(),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Validate by checking fields
			_ = creds.AccessToken != "" && !IsExpired(creds.ExpiryDate)
		}
	})
}

// BenchmarkTimeUnixMilli benchmarks time conversion operations used in token validation
func BenchmarkTimeUnixMilli(b *testing.B) {
	b.Run("Now_UnixMilli", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = time.Now().UnixMilli()
		}
	})

	b.Run("UnixMilli_Conversion", func(b *testing.B) {
		timestamp := time.Now().UnixMilli()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = time.UnixMilli(timestamp)
		}
	})

	b.Run("Until_Calculation", func(b *testing.B) {
		futureTime := time.Now().Add(time.Hour)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = time.Until(futureTime)
		}
	})
}

// BenchmarkConcurrentReaderAccess benchmarks concurrent access to the credential reader
func BenchmarkConcurrentReaderAccess(b *testing.B) {
	reader := NewOAuthCredentialReader()

	b.Run("ParallelSingletonAccess", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = GetGlobalReader()
			}
		})
	})

	b.Run("ParallelCacheClear", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				reader.ClearCache()
			}
		})
	})
}
