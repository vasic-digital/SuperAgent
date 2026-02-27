# Course-25: Lazy Loading Implementation

## Course Information
- **Duration:** 40 minutes
- **Level:** Intermediate
- **Prerequisites:** Go basics

## Module 1: Introduction to Lazy Loading (8 min)

**What is Lazy Loading?**
- Defer initialization until needed
- Reduce startup time
- Save resources
- Improve responsiveness

**Use Cases:**
- Database connections
- External API clients
- Cache initialization
- Heavy computations

## Module 2: Basic Implementation (12 min)

**Naive Approach:**
```go
type Service struct {
    db *sql.DB  // Initialized at startup
}
```

**Lazy Approach:**
```go
type Service struct {
    dbOnce sync.Once
    db     *sql.DB
}

func (s *Service) DB() *sql.DB {
    s.dbOnce.Do(func() {
        s.db = connectDB()
    })
    return s.db
}
```

## Module 3: Advanced Patterns (15 min)

**Generic Lazy Loader:**
```go
type Lazy[T any] struct {
    once     sync.Once
    value    T
    factory  func() (T, error)
}

func (l *Lazy[T]) Get() (T, error) {
    l.once.Do(func() {
        l.value, l.err = l.factory()
    })
    return l.value, l.err
}
```

**With Context:**
```go
func (l *Lazy[T]) Get(ctx context.Context) (T, error) {
    select {
    case <-l.initDone:
        return l.value, l.err
    default:
        l.initOnce.Do(func() {
            l.value, l.err = l.factory(ctx)
            close(l.initDone)
        })
        return l.value, l.err
    }
}
```

## Module 4: HelixAgent Integration (5 min)

**Usage Example:**
```go
cache := lazy.New(redis.Connect, &lazy.Config{
    TTL: 5 * time.Minute,
})

// First call initializes
client, err := cache.Get(ctx)
```

## Assessment

**Lab:** Implement lazy loading for 3 services.
