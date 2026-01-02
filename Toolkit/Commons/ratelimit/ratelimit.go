// Package ratelimit provides rate limiting functionality using token bucket algorithm.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// TokenBucketConfig holds configuration for a token bucket rate limiter.
type TokenBucketConfig struct {
	Capacity   int           // Maximum number of tokens
	RefillRate time.Duration // Time to refill one token
}

// TokenBucket implements a token bucket rate limiter.
type TokenBucket struct {
	capacity   int
	tokens     int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter.
func NewTokenBucket(config TokenBucketConfig) *TokenBucket {
	return &TokenBucket{
		capacity:   config.Capacity,
		tokens:     config.Capacity,
		refillRate: config.RefillRate,
		lastRefill: time.Now(),
	}
}

// Wait waits for a token to be available.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.takeToken() {
			return nil
		}

		select {
		case <-time.After(tb.refillRate):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// takeToken attempts to take a token from the bucket.
func (tb *TokenBucket) takeToken() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed / tb.refillRate)

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}
