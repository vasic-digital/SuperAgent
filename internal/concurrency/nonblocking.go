package concurrency

import (
	"context"
	"sync"
	"time"
)

type NonBlockingChan struct {
	ch     chan interface{}
	buffer []interface{}
	mu     sync.Mutex
	size   int
}

func NewNonBlockingChan(size int) *NonBlockingChan {
	return &NonBlockingChan{
		ch:     make(chan interface{}, size),
		buffer: make([]interface{}, 0, size),
		size:   size,
	}
}

func (nbc *NonBlockingChan) Send(item interface{}) bool {
	select {
	case nbc.ch <- item:
		return true
	default:
		nbc.mu.Lock()
		defer nbc.mu.Unlock()

		if len(nbc.buffer) < nbc.size {
			nbc.buffer = append(nbc.buffer, item)
			return true
		}
		return false
	}
}

func (nbc *NonBlockingChan) Receive() (interface{}, bool) {
	select {
	case item := <-nbc.ch:
		return item, true
	default:
		nbc.mu.Lock()
		if len(nbc.buffer) > 0 {
			item := nbc.buffer[0]
			nbc.buffer = nbc.buffer[1:]
			nbc.mu.Unlock()
			return item, true
		}
		nbc.mu.Unlock()
		return nil, false
	}
}

func (nbc *NonBlockingChan) Len() int {
	nbc.mu.Lock()
	defer nbc.mu.Unlock()
	return len(nbc.buffer) + len(nbc.ch)
}

type AsyncProcessor struct {
	workers int
	queue   chan func()
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewAsyncProcessor(workers int, queueSize int) *AsyncProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	ap := &AsyncProcessor{
		workers: workers,
		queue:   make(chan func(), queueSize),
		ctx:     ctx,
		cancel:  cancel,
	}

	for i := 0; i < workers; i++ {
		ap.wg.Add(1)
		go ap.worker()
	}

	return ap
}

func (ap *AsyncProcessor) worker() {
	defer ap.wg.Done()

	for {
		select {
		case task := <-ap.queue:
			if task != nil {
				task()
			}
		case <-ap.ctx.Done():
			return
		}
	}
}

func (ap *AsyncProcessor) Submit(task func()) bool {
	select {
	case ap.queue <- task:
		return true
	default:
		return false
	}
}

func (ap *AsyncProcessor) SubmitWithTimeout(task func(), timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case ap.queue <- task:
		return true
	case <-ctx.Done():
		return false
	}
}

func (ap *AsyncProcessor) Stop() {
	ap.cancel()
	ap.wg.Wait()
}

type LazyLoader struct {
	loader func() (interface{}, error)
	value  interface{}
	err    error
	once   sync.Once
	mu     sync.RWMutex
	loaded bool
}

func NewLazyLoader(loader func() (interface{}, error)) *LazyLoader {
	return &LazyLoader{
		loader: loader,
	}
}

func (ll *LazyLoader) Get() (interface{}, error) {
	ll.once.Do(func() {
		ll.value, ll.err = ll.loader()
		ll.mu.Lock()
		ll.loaded = true
		ll.mu.Unlock()
	})

	return ll.value, ll.err
}

func (ll *LazyLoader) IsLoaded() bool {
	ll.mu.RLock()
	defer ll.mu.RUnlock()
	return ll.loaded
}

func (ll *LazyLoader) GetOrDefault(defaultValue interface{}) interface{} {
	if ll.IsLoaded() {
		return ll.value
	}
	return defaultValue
}

type NonBlockingCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
	ttl  time.Duration
}

func NewNonBlockingCache(ttl time.Duration) *NonBlockingCache {
	return &NonBlockingCache{
		data: make(map[string]interface{}),
		ttl:  ttl,
	}
}

func (nbc *NonBlockingCache) Get(key string) (interface{}, bool) {
	nbc.mu.RLock()
	defer nbc.mu.RUnlock()

	val, ok := nbc.data[key]
	return val, ok
}

func (nbc *NonBlockingCache) Set(key string, value interface{}) {
	nbc.mu.Lock()
	defer nbc.mu.Unlock()

	nbc.data[key] = value
}

func (nbc *NonBlockingCache) Delete(key string) {
	nbc.mu.Lock()
	defer nbc.mu.Unlock()

	delete(nbc.data, key)
}

func (nbc *NonBlockingCache) Len() int {
	nbc.mu.RLock()
	defer nbc.mu.RUnlock()
	return len(nbc.data)
}

type BackgroundTask struct {
	fn     func(context.Context)
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func NewBackgroundTask(fn func(context.Context)) *BackgroundTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundTask{
		fn:     fn,
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
}

func (bt *BackgroundTask) Start() {
	go func() {
		defer close(bt.done)
		bt.fn(bt.ctx)
	}()
}

func (bt *BackgroundTask) Stop() {
	bt.cancel()
	<-bt.done
}

func (bt *BackgroundTask) Done() <-chan struct{} {
	return bt.done
}
