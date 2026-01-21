package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureRandomString(t *testing.T) {
	t.Run("generates string of correct length", func(t *testing.T) {
		for _, length := range []int{0, 1, 10, 32, 64, 128} {
			result, err := SecureRandomString(length)
			require.NoError(t, err)
			assert.Len(t, result, length)
		}
	})

	t.Run("generates unique strings", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			result, err := SecureRandomString(32)
			require.NoError(t, err)
			assert.False(t, seen[result], "should generate unique strings")
			seen[result] = true
		}
	})

	t.Run("uses only alphanumeric characters", func(t *testing.T) {
		result, err := SecureRandomString(1000)
		require.NoError(t, err)
		matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", result)
		require.NoError(t, err)
		assert.True(t, matched, "should only contain alphanumeric characters")
	})
}

func TestSecureRandomBytes(t *testing.T) {
	t.Run("generates bytes of correct length", func(t *testing.T) {
		for _, length := range []int{0, 1, 16, 32, 64} {
			result, err := SecureRandomBytes(length)
			require.NoError(t, err)
			assert.Len(t, result, length)
		}
	})

	t.Run("generates unique bytes", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			result, err := SecureRandomBytes(16)
			require.NoError(t, err)
			key := string(result)
			assert.False(t, seen[key], "should generate unique bytes")
			seen[key] = true
		}
	})
}

func TestSecureRandomHex(t *testing.T) {
	t.Run("generates hex string of correct length", func(t *testing.T) {
		for _, byteLength := range []int{1, 8, 16, 32} {
			result, err := SecureRandomHex(byteLength)
			require.NoError(t, err)
			// Hex encoding doubles the length
			assert.Len(t, result, byteLength*2)
		}
	})

	t.Run("generates valid hex string", func(t *testing.T) {
		result, err := SecureRandomHex(16)
		require.NoError(t, err)
		matched, err := regexp.MatchString("^[0-9a-f]+$", result)
		require.NoError(t, err)
		assert.True(t, matched, "should be valid hex string")
	})
}

func TestSecureRandomInt(t *testing.T) {
	t.Run("returns 0 for max <= 0", func(t *testing.T) {
		result, err := SecureRandomInt(0)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)

		result, err = SecureRandomInt(-1)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns value in range [0, max)", func(t *testing.T) {
		max := int64(100)
		for i := 0; i < 1000; i++ {
			result, err := SecureRandomInt(max)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, result, int64(0))
			assert.Less(t, result, max)
		}
	})

	t.Run("generates varied results", func(t *testing.T) {
		seen := make(map[int64]bool)
		for i := 0; i < 100; i++ {
			result, err := SecureRandomInt(1000)
			require.NoError(t, err)
			seen[result] = true
		}
		// Should have generated at least 50 unique values out of 100
		assert.GreaterOrEqual(t, len(seen), 50, "should generate varied results")
	})
}

func TestSecureRandomID(t *testing.T) {
	t.Run("generates ID with prefix", func(t *testing.T) {
		result := SecureRandomID("chatcmpl")
		assert.Contains(t, result, "chatcmpl-")
		assert.Len(t, result, len("chatcmpl-")+16) // 16 hex chars = 8 bytes
	})

	t.Run("generates ID without prefix", func(t *testing.T) {
		result := SecureRandomID("")
		assert.Len(t, result, 16) // Just the hex part
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			result := SecureRandomID("test")
			assert.False(t, seen[result], "should generate unique IDs")
			seen[result] = true
		}
	})
}

func BenchmarkSecureRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SecureRandomString(32)
	}
}

func BenchmarkSecureRandomBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SecureRandomBytes(32)
	}
}

func BenchmarkSecureRandomHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SecureRandomHex(16)
	}
}

func BenchmarkSecureRandomInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SecureRandomInt(1000)
	}
}

func BenchmarkSecureRandomID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SecureRandomID("test")
	}
}
