package profiling

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type DeadlockDetector struct {
	mu sync.RWMutex

	locks map[uintptr]*LockInfo
	order map[string][]string

	timeout time.Duration
	alerts  chan DeadlockAlert

	running bool
	stopCh  chan struct{}
}

type LockInfo struct {
	ID        uintptr
	Name      string
	Goroutine int
	HeldAt    time.Time
	Stack     []byte
}

type DeadlockAlert struct {
	Type      string
	Timestamp time.Time
	Message   string
	LockName  string
	HeldFor   time.Duration
	Goroutine int
	Stack     []byte
}

type mutexWrapper struct {
	mu       sync.Mutex
	detector *DeadlockDetector
	name     string
	id       uintptr
}

func NewDeadlockDetector(timeout time.Duration) *DeadlockDetector {
	return &DeadlockDetector{
		locks:   make(map[uintptr]*LockInfo),
		order:   make(map[string][]string),
		timeout: timeout,
		alerts:  make(chan DeadlockAlert, 100),
		stopCh:  make(chan struct{}),
	}
}

func (d *DeadlockDetector) Start(ctx context.Context) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("detector already running")
	}
	d.running = true
	d.mu.Unlock()

	go d.monitorLoop(ctx)
	return nil
}

func (d *DeadlockDetector) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return
	}

	close(d.stopCh)
	d.running = false
}

func (d *DeadlockDetector) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.checkDeadlocks()
		}
	}
}

func (d *DeadlockDetector) checkDeadlocks() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	now := time.Now()
	for _, info := range d.locks {
		heldFor := now.Sub(info.HeldAt)
		if heldFor > d.timeout {
			d.sendAlert(DeadlockAlert{
				Type:      "potential_deadlock",
				Timestamp: now,
				Message:   fmt.Sprintf("Lock held for %v exceeds timeout %v", heldFor, d.timeout),
				LockName:  info.Name,
				HeldFor:   heldFor,
				Goroutine: info.Goroutine,
				Stack:     info.Stack,
			})
		}
	}
}

func (d *DeadlockDetector) sendAlert(alert DeadlockAlert) {
	select {
	case d.alerts <- alert:
	default:
	}
}

func (d *DeadlockDetector) Alerts() <-chan DeadlockAlert {
	return d.alerts
}

func (d *DeadlockDetector) RecordLockAcquire(name string) func() {
	id := time.Now().UnixNano()
	stack := make([]byte, 4096)
	n := runtime.Stack(stack, false)
	stack = stack[:n]

	info := &LockInfo{
		ID:        uintptr(id),
		Name:      name,
		Goroutine: getGoroutineID(),
		HeldAt:    time.Now(),
		Stack:     stack,
	}

	d.mu.Lock()
	d.locks[info.ID] = info
	d.mu.Unlock()

	return func() {
		d.mu.Lock()
		delete(d.locks, info.ID)
		d.mu.Unlock()
	}
}

func (d *DeadlockDetector) WrapMutex(mu *sync.Mutex, name string) *MutexWrapper {
	return &MutexWrapper{
		mu:       mu,
		detector: d,
		name:     name,
	}
}

type MutexWrapper struct {
	mu       *sync.Mutex
	detector *DeadlockDetector
	name     string
	release  func()
}

func (w *MutexWrapper) Lock() {
	w.release = w.detector.RecordLockAcquire(w.name)
	w.mu.Lock()
}

func (w *MutexWrapper) Unlock() {
	if w.release != nil {
		w.release()
	}
	w.mu.Unlock()
}

func getGoroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	id := 0
	fmt.Sscanf(string(buf[:n]), "goroutine %d ", &id) //nolint:errcheck
	return id
}

func (d *DeadlockDetector) GetLockInfo() []LockInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]LockInfo, 0, len(d.locks))
	for _, info := range d.locks {
		result = append(result, *info)
	}
	return result
}

func (d *DeadlockDetector) RecordLockOrder(acquired, acquiring string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.order[acquired] = append(d.order[acquired], acquiring)

	if d.hasCycle(acquired) {
		d.sendAlert(DeadlockAlert{
			Type:      "lock_order_violation",
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Potential deadlock: lock order cycle detected involving %s", acquired),
			LockName:  acquired,
		})
	}
}

func (d *DeadlockDetector) hasCycle(start string) bool {
	visited := make(map[string]bool)
	return d.dfs(start, start, visited)
}

func (d *DeadlockDetector) dfs(current, target string, visited map[string]bool) bool {
	if visited[current] {
		return current == target
	}

	visited[current] = true

	for _, next := range d.order[current] {
		if d.dfs(next, target, visited) {
			return true
		}
	}

	return false
}

func (d *DeadlockDetector) DumpGoroutines() []byte {
	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, true)
	return buf[:n]
}

func (d *DeadlockDetector) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}
