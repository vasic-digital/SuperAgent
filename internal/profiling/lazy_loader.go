package profiling

import (
	"sync"
	"sync/atomic"
)

type LazyLoader struct {
	mu sync.RWMutex

	instances map[string]interface{}
	factories map[string]func() (interface{}, error)
	loading   map[string]*atomic.Bool
	cond      *sync.Cond
}

func NewLazyLoader() *LazyLoader {
	l := &LazyLoader{
		instances: make(map[string]interface{}),
		factories: make(map[string]func() (interface{}, error)),
		loading:   make(map[string]*atomic.Bool),
	}
	l.cond = sync.NewCond(&l.mu)
	return l
}

func (l *LazyLoader) Register(name string, factory func() (interface{}, error)) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.factories[name] = factory
	l.loading[name] = &atomic.Bool{}
}

func (l *LazyLoader) Get(name string) (interface{}, error) {
	l.mu.RLock()
	if instance, ok := l.instances[name]; ok {
		l.mu.RUnlock()
		return instance, nil
	}
	l.mu.RUnlock()

	return l.load(name)
}

func (l *LazyLoader) load(name string) (interface{}, error) {
	l.mu.Lock()

	loading := l.loading[name]
	if loading == nil {
		l.mu.Unlock()
		return nil, nil
	}

	for loading.Load() {
		l.cond.Wait()
	}

	if instance, ok := l.instances[name]; ok {
		l.mu.Unlock()
		return instance, nil
	}

	loading.Store(true)
	l.mu.Unlock()

	defer func() {
		loading.Store(false)
		l.cond.Broadcast()
	}()

	factory, ok := l.factories[name]
	if !ok {
		return nil, nil
	}

	instance, err := factory()
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	l.instances[name] = instance
	l.mu.Unlock()

	return instance, nil
}

func (l *LazyLoader) IsLoaded(name string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	_, ok := l.instances[name]
	return ok
}

func (l *LazyLoader) Unload(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.instances, name)
}

func (l *LazyLoader) Preload(names ...string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(names))

	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			if _, err := l.Get(n); err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}(name)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *LazyLoader) GetLoaded() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]string, 0, len(l.instances))
	for name := range l.instances {
		result = append(result, name)
	}
	return result
}

func (l *LazyLoader) GetRegistered() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]string, 0, len(l.factories))
	for name := range l.factories {
		result = append(result, name)
	}
	return result
}

func (l *LazyLoader) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.instances = make(map[string]interface{})
}
