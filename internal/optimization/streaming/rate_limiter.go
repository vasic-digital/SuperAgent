package streaming

import (
	"context"
	"sync"
	"time"
)

// RateLimiter limits the rate of token/chunk output.
type RateLimiter struct {
	mu              sync.Mutex
	tokensPerSecond float64
	delay           time.Duration
	lastEmit        time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(tokensPerSecond float64) *RateLimiter {
	if tokensPerSecond <= 0 {
		tokensPerSecond = 100 // Default to 100 tokens/second
	}

	return &RateLimiter{
		tokensPerSecond: tokensPerSecond,
		delay:           time.Duration(float64(time.Second) / tokensPerSecond),
	}
}

// Limit rate-limits a string channel.
func (r *RateLimiter) Limit(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case token, ok := <-in:
				if !ok {
					return
				}

				r.wait(ctx)

				select {
				case out <- token:
					r.mu.Lock()
					r.lastEmit = time.Now()
					r.mu.Unlock()
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

// LimitChunks rate-limits a StreamChunk channel.
func (r *RateLimiter) LimitChunks(ctx context.Context, in <-chan *StreamChunk) <-chan *StreamChunk {
	out := make(chan *StreamChunk)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-in:
				if !ok {
					return
				}

				r.wait(ctx)

				select {
				case out <- chunk:
					r.mu.Lock()
					r.lastEmit = time.Now()
					r.mu.Unlock()
				case <-ctx.Done():
					return
				}

				if chunk.Done {
					return
				}
			}
		}
	}()

	return out
}

func (r *RateLimiter) wait(ctx context.Context) {
	r.mu.Lock()
	if r.lastEmit.IsZero() {
		r.lastEmit = time.Now()
		r.mu.Unlock()
		return
	}

	elapsed := time.Since(r.lastEmit)
	delay := r.delay
	r.mu.Unlock()

	if elapsed < delay {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay - elapsed):
		}
	}
}

// Reset resets the rate limiter state.
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastEmit = time.Time{}
}

// SetRate updates the rate limit.
func (r *RateLimiter) SetRate(tokensPerSecond float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tokensPerSecond <= 0 {
		tokensPerSecond = 100
	}
	r.tokensPerSecond = tokensPerSecond
	r.delay = time.Duration(float64(time.Second) / tokensPerSecond)
}

// BurstRateLimiter allows bursts followed by rate limiting.
type BurstRateLimiter struct {
	mu              sync.Mutex
	tokensPerSecond float64
	burstSize       int
	tokens          int
	lastRefill      time.Time
}

// NewBurstRateLimiter creates a rate limiter with burst capability.
func NewBurstRateLimiter(tokensPerSecond float64, burstSize int) *BurstRateLimiter {
	if tokensPerSecond <= 0 {
		tokensPerSecond = 100
	}
	if burstSize <= 0 {
		burstSize = 10
	}

	return &BurstRateLimiter{
		tokensPerSecond: tokensPerSecond,
		burstSize:       burstSize,
		tokens:          burstSize,
		lastRefill:      time.Now(),
	}
}

// Limit rate-limits with burst support.
func (r *BurstRateLimiter) Limit(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case token, ok := <-in:
				if !ok {
					return
				}

				r.waitForToken(ctx)

				select {
				case out <- token:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

func (r *BurstRateLimiter) waitForToken(ctx context.Context) {
	r.mu.Lock()
	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	refill := int(elapsed * r.tokensPerSecond)

	if refill > 0 {
		r.tokens += refill
		if r.tokens > r.burstSize {
			r.tokens = r.burstSize
		}
		r.lastRefill = now
	}

	// Wait if no tokens available
	for r.tokens <= 0 {
		waitTime := time.Duration(float64(time.Second) / r.tokensPerSecond)
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitTime):
		}
		r.mu.Lock()
		r.tokens++
		r.lastRefill = time.Now()
	}

	r.tokens--
	r.mu.Unlock()
}

// Reset resets the rate limiter.
func (r *BurstRateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens = r.burstSize
	r.lastRefill = time.Now()
}
