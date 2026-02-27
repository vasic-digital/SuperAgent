# User Manual 27: API Rate Limiting

## Overview
Configuring and managing API rate limits.

## Default Limits
```yaml
rate_limits:
  anonymous:
    requests_per_minute: 60
    requests_per_hour: 1000
  authenticated:
    requests_per_minute: 600
    requests_per_hour: 10000
```

## Headers
```http
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 599
X-RateLimit-Reset: 1645000000
```

## Custom Limits
```go
limiter := rate.NewLimiter(rate.Limit(10), 100)

func handler(w http.ResponseWriter, r *http.Request) {
    if !limiter.Allow() {
        http.Error(w, "Rate limit exceeded", 429)
        return
    }
    // Handle request
}
```

## Redis Backend
```yaml
rate_limiter:
  backend: redis
  redis_url: redis://localhost:6379
```
