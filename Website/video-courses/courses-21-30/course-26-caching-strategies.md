# Course-26: Caching Strategies

## Course Information
- **Duration:** 45 minutes
- **Level:** Intermediate
- **Prerequisites:** Course-25

## Module 1: Caching Fundamentals (10 min)

**Why Cache?**
- Reduce latency
- Decrease load
- Improve throughput
- Save costs

**Cache Types:**
- In-memory
- Distributed
- CDN
- Client-side

## Module 2: Cache Patterns (15 min)

**Pattern 1: Cache-Aside**
```go
func (s *Service) GetData(key string) (Data, error) {
    // Check cache
    if val, ok := cache.Get(key); ok {
        return val, nil
    }
    
    // Load from source
    data, err := s.loadFromDB(key)
    if err != nil {
        return Data{}, err
    }
    
    // Store in cache
    cache.Set(key, data, ttl)
    return data, nil
}
```

**Pattern 2: Write-Through**
```go
func (s *Service) UpdateData(key string, data Data) error {
    // Update DB
    if err := s.db.Update(key, data); err != nil {
        return err
    }
    
    // Update cache
    cache.Set(key, data, ttl)
    return nil
}
```

**Pattern 3: Write-Behind**
```go
func (s *Service) UpdateDataAsync(key string, data Data) {
    cache.Set(key, data, ttl)
    
    // Async DB update
    go func() {
        s.db.Update(key, data)
    }()
}
```

## Module 3: Cache Invalidation (12 min)

**TTL-Based:**
```go
cache.Set(key, value, 5*time.Minute)
```

**Explicit:**
```go
cache.Delete(key)
cache.Flush()
```

**Event-Driven:**
```go
eventBus.Subscribe("data.updated", func(key string) {
    cache.Delete(key)
})
```

## Module 4: Advanced Topics (8 min)

**Cache Warming:**
```go
func (s *Service) WarmCache() {
    keys := s.getPopularKeys()
    for _, key := range keys {
        data, _ := s.loadFromDB(key)
        cache.Set(key, data, ttl)
    }
}
```

**Multi-Level Cache:**
```go
func (s *Service) Get(key string) (Data, error) {
    // L1: In-memory
    if val, ok := localCache.Get(key); ok {
        return val, nil
    }
    
    // L2: Redis
    if val, ok := redisCache.Get(key); ok {
        localCache.Set(key, val, shortTTL)
        return val, nil
    }
    
    // Source
    data, _ := s.loadFromSource(key)
    redisCache.Set(key, data, longTTL)
    localCache.Set(key, data, shortTTL)
    return data, nil
}
```

## Assessment

**Lab:** Implement multi-level caching system.
