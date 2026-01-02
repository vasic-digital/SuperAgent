package modelsdev

import (
	"context"
	"sync"
	"time"
)

type RateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	interval  time.Duration
	lastReset time.Time
}

func NewRateLimiter(maxTokens int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		interval:  interval,
		lastReset: time.Now(),
	}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastReset) >= r.interval {
		r.tokens = r.maxTokens
		r.lastReset = now
	}

	if r.tokens <= 0 {
		waitTime := r.interval - now.Sub(r.lastReset)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			r.tokens = r.maxTokens
			r.lastReset = time.Now()
		}
	}

	r.tokens--
	return nil
}
