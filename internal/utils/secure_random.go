// Package utils provides utility functions for the HelixAgent application.
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
)

// SecureRandomString generates a cryptographically secure random string of the specified length.
// Uses crypto/rand for security-sensitive operations.
func SecureRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

// SecureRandomBytes generates cryptographically secure random bytes.
func SecureRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// SecureRandomHex generates a cryptographically secure random hex string.
func SecureRandomHex(byteLength int) (string, error) {
	bytes, err := SecureRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SecureRandomInt generates a cryptographically secure random integer in [0, max).
func SecureRandomInt(max int64) (int64, error) {
	if max <= 0 {
		return 0, nil
	}
	num, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return num.Int64(), nil
}

// SecureRandomID generates a unique ID suitable for API responses.
// Format: timestamp-based prefix + secure random suffix
func SecureRandomID(prefix string) string {
	randomPart, err := SecureRandomHex(8)
	if err != nil {
		// Fallback to a less secure but functional approach
		randomPart = "00000000"
	}
	if prefix == "" {
		return randomPart
	}
	return prefix + "-" + randomPart
}
