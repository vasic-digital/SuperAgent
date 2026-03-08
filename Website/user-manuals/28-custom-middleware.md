# User Manual 28: Custom Middleware

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Middleware Architecture in HelixAgent](#middleware-architecture-in-helixagent)
4. [Gin Middleware Basics](#gin-middleware-basics)
5. [Authentication Middleware](#authentication-middleware)
6. [Logging Middleware](#logging-middleware)
7. [CORS Middleware](#cors-middleware)
8. [Compression Middleware](#compression-middleware)
9. [Request ID Middleware](#request-id-middleware)
10. [Recovery Middleware](#recovery-middleware)
11. [Creating Custom Middleware](#creating-custom-middleware)
12. [Middleware Ordering](#middleware-ordering)
13. [Testing Middleware](#testing-middleware)
14. [Configuration Reference](#configuration-reference)
15. [Troubleshooting](#troubleshooting)
16. [Related Resources](#related-resources)

## Overview

HelixAgent uses the Gin HTTP framework (v1.11.0) and implements a middleware chain that processes every request. Middleware handles cross-cutting concerns: authentication, rate limiting, logging, CORS, compression, request tracing, and error recovery. This manual covers the built-in middleware, how to create custom middleware, and how to integrate it into the HelixAgent pipeline.

## Prerequisites

- Familiarity with the Gin framework (`github.com/gin-gonic/gin`)
- Understanding of HTTP request/response lifecycle
- Go 1.24+ development environment
- HelixAgent source code access (`internal/middleware/`)

## Middleware Architecture in HelixAgent

```
Incoming Request
       |
       v
+------+---------+
| Recovery       |  Catches panics, returns 500
+------+---------+
       |
+------+---------+
| Request ID     |  Assigns unique ID to each request
+------+---------+
       |
+------+---------+
| Logging        |  Logs method, path, status, duration
+------+---------+
       |
+------+---------+
| CORS           |  Sets Access-Control-* headers
+------+---------+
       |
+------+---------+
| Compression    |  Brotli (primary) / gzip (fallback)
+------+---------+
       |
+------+---------+
| Rate Limiter   |  Token bucket / sliding window
+------+---------+
       |
+------+---------+
| Authentication |  JWT / API key validation
+------+---------+
       |
+------+---------+
| Tracing        |  OpenTelemetry span creation
+------+---------+
       |
+------+---------+
| Handler        |  Route-specific business logic
+------+---------+
       |
       v
Outgoing Response
```

## Gin Middleware Basics

Gin middleware is a function that accepts a `*gin.Context` and calls `c.Next()` to proceed to the next handler in the chain.

### Basic Structure

```go
func MyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Before handler: runs on the way in
        start := time.Now()

        c.Next() // Call the next handler

        // After handler: runs on the way out
        duration := time.Since(start)
        status := c.Writer.Status()
        log.Printf("Request completed: status=%d duration=%v", status, duration)
    }
}
```

### Aborting the Chain

```go
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "authentication required",
            })
            return // c.Abort prevents subsequent handlers from running
        }
        c.Next()
    }
}
```

### Registering Middleware

```go
router := gin.New()

// Global middleware (applies to all routes)
router.Use(Recovery())
router.Use(RequestID())
router.Use(Logging())

// Group-specific middleware
api := router.Group("/v1")
api.Use(AuthRequired())
{
    api.POST("/chat/completions", chatHandler)
    api.GET("/models", modelsHandler)
}

// Route-specific middleware
router.GET("/metrics", PrometheusAuth(), metricsHandler)
```

## Authentication Middleware

### JWT Authentication

```go
func JWTAuth(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{
                    "message": "missing Authorization header",
                    "type":    "auth_error",
                },
            })
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := validateJWT(tokenString, secret)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{
                    "message": "invalid or expired token",
                    "type":    "auth_error",
                },
            })
            return
        }

        // Store user info in context for downstream handlers
        c.Set("user_id", claims.Subject)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

### API Key Authentication

```go
func APIKeyAuth(validKeys map[string]string) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.GetHeader("X-API-Key")
        if key == "" {
            key = c.GetHeader("Authorization")
            key = strings.TrimPrefix(key, "Bearer ")
        }

        userID, ok := validKeys[key]
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{
                    "message": "invalid API key",
                    "type":    "auth_error",
                },
            })
            return
        }

        c.Set("user_id", userID)
        c.Set("auth_method", "api_key")
        c.Next()
    }
}
```

### Combined Auth (JWT or API Key)

```go
func CombinedAuth(jwtSecret string, apiKeys map[string]string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")

        // Try API key first
        if userID, ok := apiKeys[strings.TrimPrefix(authHeader, "Bearer ")]; ok {
            c.Set("user_id", userID)
            c.Set("auth_method", "api_key")
            c.Next()
            return
        }

        // Try JWT
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := validateJWT(tokenString, jwtSecret)
        if err == nil {
            c.Set("user_id", claims.Subject)
            c.Set("auth_method", "jwt")
            c.Next()
            return
        }

        c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
            "error": gin.H{"message": "authentication required"},
        })
    }
}
```

## Logging Middleware

```go
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        requestID := c.GetString("request_id")
        path := c.Request.URL.Path
        method := c.Request.Method

        slog.Info("request started",
            slog.String("request_id", requestID),
            slog.String("method", method),
            slog.String("path", path),
            slog.String("client_ip", c.ClientIP()),
        )

        c.Next()

        duration := time.Since(start)
        status := c.Writer.Status()

        logFn := slog.Info
        if status >= 500 {
            logFn = slog.Error
        } else if status >= 400 {
            logFn = slog.Warn
        }

        logFn("request completed",
            slog.String("request_id", requestID),
            slog.String("method", method),
            slog.String("path", path),
            slog.Int("status", status),
            slog.Duration("duration", duration),
            slog.Int("body_size", c.Writer.Size()),
        )
    }
}
```

## CORS Middleware

```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers",
            "Content-Type, Authorization, X-API-Key, X-Request-ID")
        c.Writer.Header().Set("Access-Control-Expose-Headers",
            "X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, X-Request-ID")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
```

### Restricted CORS

```go
func RestrictedCORS(allowedOrigins []string) gin.HandlerFunc {
    originSet := make(map[string]bool)
    for _, o := range allowedOrigins {
        originSet[o] = true
    }

    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        if originSet[origin] {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Vary", "Origin")
        }
        // ... same headers as above ...
        c.Next()
    }
}
```

## Compression Middleware

HelixAgent mandates Brotli as the primary compression with gzip fallback:

```go
func CompressionMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        acceptEncoding := c.GetHeader("Accept-Encoding")

        if strings.Contains(acceptEncoding, "br") {
            c.Writer.Header().Set("Content-Encoding", "br")
            bw := brotli.NewWriter(c.Writer)
            defer bw.Close()
            c.Writer = &compressedWriter{Writer: bw, ResponseWriter: c.Writer}
        } else if strings.Contains(acceptEncoding, "gzip") {
            c.Writer.Header().Set("Content-Encoding", "gzip")
            gz := gzip.NewWriter(c.Writer)
            defer gz.Close()
            c.Writer = &compressedWriter{Writer: gz, ResponseWriter: c.Writer}
        }

        c.Next()
    }
}
```

Note: SSE streaming endpoints (`/v1/chat/completions` with `stream: true`) should bypass compression to avoid buffering delays.

## Request ID Middleware

```go
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Writer.Header().Set("X-Request-ID", requestID)
        c.Next()
    }
}
```

## Recovery Middleware

```go
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                requestID := c.GetString("request_id")
                stack := string(debug.Stack())

                slog.Error("panic recovered",
                    slog.String("request_id", requestID),
                    slog.Any("panic", r),
                    slog.String("stack", stack),
                )

                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "error": gin.H{
                        "message":    "internal server error",
                        "request_id": requestID,
                    },
                })
            }
        }()
        c.Next()
    }
}
```

## Creating Custom Middleware

### Step-by-Step Guide

1. **Create a new file** in `internal/middleware/`:

```go
// internal/middleware/my_middleware.go
package middleware

import "github.com/gin-gonic/gin"

// MyCustomMiddleware does X for every request.
func MyCustomMiddleware(config MyConfig) gin.HandlerFunc {
    // Initialize any resources needed (connection pools, caches, etc.)
    // This runs once when the middleware is registered.

    return func(c *gin.Context) {
        // Pre-processing (runs before the handler)

        c.Next()

        // Post-processing (runs after the handler)
    }
}

type MyConfig struct {
    Enabled bool
    // ... configuration fields ...
}
```

2. **Register the middleware** in the router setup:

```go
router.Use(middleware.MyCustomMiddleware(middleware.MyConfig{
    Enabled: true,
}))
```

3. **Write tests** (see next section).

### Example: Request Size Limiter

```go
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.ContentLength > maxBytes {
            c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
                "error": gin.H{
                    "message": fmt.Sprintf("request body too large (max: %d bytes)", maxBytes),
                    "type":    "validation_error",
                },
            })
            return
        }
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
        c.Next()
    }
}
```

## Middleware Ordering

Order matters. The first registered middleware runs first on the way in and last on the way out:

```go
router := gin.New()
router.Use(Recovery())       // 1st: catches panics from all subsequent middleware
router.Use(RequestID())      // 2nd: assigns ID for tracing
router.Use(RequestLogger())  // 3rd: logs with request ID
router.Use(CORSMiddleware()) // 4th: handles preflight
router.Use(Compression())    // 5th: compresses responses
router.Use(RateLimit())      // 6th: rejects excess traffic
router.Use(Auth())           // 7th: authenticates the request
router.Use(Tracing())        // 8th: creates OpenTelemetry span
```

### Execution Order

```
Request  -> Recovery -> RequestID -> Logger -> CORS -> Compression -> RateLimit -> Auth -> Tracing -> Handler
Response <- Recovery <- RequestID <- Logger <- CORS <- Compression <- RateLimit <- Auth <- Tracing <- Handler
```

## Testing Middleware

### Unit Test with httptest

```go
func TestRequestID_GeneratesID(t *testing.T) {
    router := gin.New()
    router.Use(RequestID())
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"request_id": c.GetString("request_id")})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/test", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.NotEmpty(t, w.Header().Get("X-Request-ID"))

    var body map[string]string
    json.Unmarshal(w.Body.Bytes(), &body)
    assert.NotEmpty(t, body["request_id"])
}

func TestRequestID_UsesProvidedID(t *testing.T) {
    router := gin.New()
    router.Use(RequestID())
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"request_id": c.GetString("request_id")})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/test", nil)
    req.Header.Set("X-Request-ID", "custom-id-123")
    router.ServeHTTP(w, req)

    assert.Equal(t, "custom-id-123", w.Header().Get("X-Request-ID"))
}
```

### Testing Auth Middleware

```go
func TestJWTAuth_RejectsInvalidToken(t *testing.T) {
    router := gin.New()
    router.Use(JWTAuth("test-secret"))
    router.GET("/protected", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    router.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
}

func TestJWTAuth_AcceptsValidToken(t *testing.T) {
    token := generateTestJWT("test-secret", "user-1", "admin")

    router := gin.New()
    router.Use(JWTAuth("test-secret"))
    router.GET("/protected", func(c *gin.Context) {
        c.JSON(200, gin.H{"user_id": c.GetString("user_id")})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
}
```

### Testing Recovery Middleware

```go
func TestRecovery_CatchesPanic(t *testing.T) {
    router := gin.New()
    router.Use(Recovery())
    router.GET("/panic", func(c *gin.Context) {
        panic("test panic")
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/panic", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 500, w.Code)
    assert.Contains(t, w.Body.String(), "internal server error")
}
```

## Configuration Reference

| Setting | Default | Description |
|---|---|---|
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed origins |
| `CORS_MAX_AGE` | `86400` | Preflight cache duration (seconds) |
| `MAX_REQUEST_BODY_BYTES` | `10485760` | Maximum request body size (10 MB) |
| `COMPRESSION_ENABLED` | `true` | Enable response compression |
| `COMPRESSION_PRIMARY` | `brotli` | Primary compression algorithm |
| `JWT_SECRET` | (required) | JWT signing secret |
| `LOG_REQUESTS` | `true` | Enable request logging middleware |
| `GIN_MODE` | `release` | Gin mode (debug, test, release) |

## Troubleshooting

### Middleware Not Executing

**Symptom:** Custom middleware registered but never called.

**Solutions:**
1. Verify `router.Use()` is called before route registration
2. Check that the middleware calls `c.Next()` (or `c.Abort*` for terminating middleware)
3. Ensure the middleware is not registered on a different router group

### CORS Preflight Failing

**Symptom:** Browser reports CORS errors on cross-origin requests.

**Solutions:**
1. Verify OPTIONS requests return 204 with correct headers
2. Check `Access-Control-Allow-Origin` matches the request origin
3. Ensure `Access-Control-Allow-Headers` includes all headers sent by the client
4. Place CORS middleware before authentication middleware

### Response Compression Breaks Streaming

**Symptom:** SSE streaming responses are buffered and delivered all at once.

**Solutions:**
1. Bypass compression for SSE endpoints: check `Accept: text/event-stream` header
2. Set `Content-Type: text/event-stream` before writing
3. Flush after each SSE event: call `c.Writer.Flush()`
4. Disable proxy buffering in reverse proxy: `X-Accel-Buffering: no`

### Panic Recovery Hides Root Cause

**Symptom:** 500 errors without useful information in logs.

**Solutions:**
1. Enable stack trace logging in the Recovery middleware
2. Set `GIN_MODE=debug` during development for verbose error output
3. Check structured logs for the `stack` field in panic recovery entries
4. Run with the race detector to catch concurrent access panics: `go test -race`

## Related Resources

- [User Manual 27: API Rate Limiting](27-api-rate-limiting.md) -- Rate limit middleware configuration
- [User Manual 26: Compliance Guide](26-compliance-guide.md) -- Authentication and access control
- [User Manual 23: Observability Setup](23-observability-setup.md) -- Tracing middleware
- Middleware source: `internal/middleware/`
- Auth module: `Auth/`
- Gin documentation: https://gin-gonic.com/docs/
