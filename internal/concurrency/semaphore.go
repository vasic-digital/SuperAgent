package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Semaphore struct {
	ch      chan struct{}
	mu      sync.Mutex
	max     int
	current int
}

func NewSemaphore(max int) *Semaphore {
	return &Semaphore{
		ch:  make(chan struct{}, max),
		max: max,
	}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		s.mu.Lock()
		s.current++
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) AcquireWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Acquire(ctx)
}

func (s *Semaphore) Release() {
	select {
	case <-s.ch:
		s.mu.Lock()
		if s.current > 0 {
			s.current--
		}
		s.mu.Unlock()
	default:
	}
}

func (s *Semaphore) Current() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

func (s *Semaphore) Available() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.max - s.current
}

func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		s.mu.Lock()
		s.current++
		s.mu.Unlock()
		return true
	default:
		return false
	}
}

func (s *Semaphore) Close() {
	close(s.ch)
}

type RateLimiter struct {
	semaphore *Semaphore
	interval  time.Duration
	ticker    *time.Ticker
	stopCh    chan struct{}
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	rl := &RateLimiter{
		semaphore: NewSemaphore(requestsPerSecond),
		interval:  time.Second / time.Duration(requestsPerSecond),
		stopCh:    make(chan struct{}),
	}

	rl.ticker = time.NewTicker(rl.interval)
	go rl.refill()

	return rl
}

func (rl *RateLimiter) refill() {
	for {
		select {
		case <-rl.ticker.C:
			rl.semaphore.Release()
		case <-rl.stopCh:
			return
		}
	}
}

func (rl *RateLimiter) Acquire(ctx context.Context) error {
	return rl.semaphore.Acquire(ctx)
}

func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
	close(rl.stopCh)
}

type PrioritySemaphore struct {
	highPriority chan struct{}
	lowPriority  chan struct{}
	maxHigh      int
	maxLow       int
	mu           sync.Mutex
}

func NewPrioritySemaphore(maxHigh, maxLow int) *PrioritySemaphore {
	return &PrioritySemaphore{
		highPriority: make(chan struct{}, maxHigh),
		lowPriority:  make(chan struct{}, maxLow),
		maxHigh:      maxHigh,
		maxLow:       maxLow,
	}
}

func (ps *PrioritySemaphore) AcquireHigh(ctx context.Context) error {
	select {
	case ps.highPriority <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ps *PrioritySemaphore) AcquireLow(ctx context.Context) error {
	select {
	case ps.highPriority <- struct{}{}:
		return nil
	case ps.lowPriority <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (ps *PrioritySemaphore) Release() {
	select {
	case <-ps.highPriority:
		return
	case <-ps.lowPriority:
		return
	default:
	}
}

type ResourcePool struct {
	resources chan interface{}
	factory   func() (interface{}, error)
	mu        sync.Mutex
	closed    bool
}

func NewResourcePool(size int, factory func() (interface{}, error)) (*ResourcePool, error) {
	pool := &ResourcePool{
		resources: make(chan interface{}, size),
		factory:   factory,
	}

	for i := 0; i < size; i++ {
		res, err := factory()
		if err != nil {
			return nil, fmt.Errorf("failed to create resource %d: %w", i, err)
		}
		pool.resources <- res
	}

	return pool, nil
}

func (p *ResourcePool) Acquire(ctx context.Context) (interface{}, error) {
	select {
	case res := <-p.resources:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *ResourcePool) Release(res interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("pool is closed")
	}

	select {
	case p.resources <- res:
		return nil
	default:
		return fmt.Errorf("pool is full")
	}
}

func (p *ResourcePool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.closed {
		p.closed = true
		close(p.resources)
	}
}
