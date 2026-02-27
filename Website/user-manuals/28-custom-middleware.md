# User Manual 28: Custom Middleware

## Overview
Developing custom middleware for HelixAgent.

## Middleware Interface
```go
type Middleware func(http.Handler) http.Handler

func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before
        start := time.Now()
        
        next.ServeHTTP(w, r)
        
        // After
        duration := time.Since(start)
        log.Printf("Request took %v", duration)
    })
}
```

## Registration
```go
router.Use(MyMiddleware)
```

## Common Patterns
- Authentication
- Logging
- Metrics
- Recovery
- CORS

## Testing
```go
func TestMiddleware(t *testing.T) {
    handler := MyMiddleware(testHandler)
    // Test logic
}
```
