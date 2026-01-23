// Package middleware provides HTTP middleware for HelixAgent's Gin-based API.
//
// This package implements common middleware functions for authentication,
// authorization, rate limiting, logging, and request processing.
//
// # Available Middleware
//
//   - Authentication: JWT and API key validation
//   - Authorization: Role-based access control
//   - RateLimiting: Request rate limiting
//   - CORS: Cross-origin resource sharing
//   - Logging: Request/response logging
//   - Validation: Request validation
//   - Recovery: Panic recovery
//   - Timeout: Request timeout handling
//   - Compression: Response compression
//
// # Authentication Middleware
//
// Support for multiple authentication methods:
//
//	// JWT authentication
//	router.Use(middleware.JWTAuth(jwtSecret))
//
//	// API key authentication
//	router.Use(middleware.APIKeyAuth(keyValidator))
//
//	// Combined (either works)
//	router.Use(middleware.Auth(authConfig))
//
// # Rate Limiting
//
// Configurable rate limiting:
//
//	limiter := middleware.NewRateLimiter(&RateLimitConfig{
//	    RequestsPerMinute: 60,
//	    BurstSize:         10,
//	    KeyFunc:           middleware.IPKey,
//	})
//
//	router.Use(limiter.Middleware())
//
// Rate limit key functions:
//   - IPKey: Limit by IP address
//   - APIKeyKey: Limit by API key
//   - UserKey: Limit by user ID
//   - Custom: Custom key extraction
//
// # CORS Configuration
//
// Cross-origin resource sharing:
//
//	corsConfig := middleware.CORSConfig{
//	    AllowOrigins:     []string{"https://example.com"},
//	    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
//	    AllowHeaders:     []string{"Authorization", "Content-Type"},
//	    ExposeHeaders:    []string{"X-Request-ID"},
//	    AllowCredentials: true,
//	    MaxAge:           12 * time.Hour,
//	}
//
//	router.Use(middleware.CORS(corsConfig))
//
// # Request Logging
//
// Structured request logging:
//
//	logger := middleware.NewRequestLogger(&LogConfig{
//	    Logger:         log,
//	    SkipPaths:      []string{"/health"},
//	    LogRequestBody: true,
//	    LogResponseBody: false,
//	})
//
//	router.Use(logger.Middleware())
//
// # Request Validation
//
// Automatic request validation:
//
//	validator := middleware.NewValidator()
//	router.Use(validator.Middleware())
//
//	// Validates against struct tags
//	type CreateRequest struct {
//	    Name  string `json:"name" binding:"required,min=3"`
//	    Email string `json:"email" binding:"required,email"`
//	}
//
// # Timeout Handling
//
// Request timeout middleware:
//
//	router.Use(middleware.Timeout(30 * time.Second))
//
// # Panic Recovery
//
// Graceful panic recovery:
//
//	router.Use(middleware.Recovery(logger))
//
// Logs stack trace and returns 500 error.
//
// # Context Enrichment
//
// Add context values to requests:
//
//	router.Use(middleware.RequestID())     // Add X-Request-ID
//	router.Use(middleware.UserContext())   // Add user info
//	router.Use(middleware.ProviderContext()) // Add provider info
//
// # Key Files
//
//   - auth.go: Authentication middleware
//   - ratelimit.go: Rate limiting
//   - cors.go: CORS handling
//   - logging.go: Request logging
//   - validation.go: Request validation
//   - timeout.go: Timeout handling
//   - recovery.go: Panic recovery
//   - context.go: Context enrichment
//
// # Middleware Chain
//
// Typical middleware chain:
//
//	router := gin.New()
//
//	// Global middleware
//	router.Use(middleware.Recovery(logger))
//	router.Use(middleware.RequestID())
//	router.Use(middleware.CORS(corsConfig))
//	router.Use(middleware.RequestLogger(logConfig))
//
//	// API routes with authentication
//	api := router.Group("/v1")
//	api.Use(middleware.Auth(authConfig))
//	api.Use(middleware.RateLimit(rateLimitConfig))
//
// # Custom Middleware
//
// Create custom middleware:
//
//	func CustomMiddleware() gin.HandlerFunc {
//	    return func(c *gin.Context) {
//	        // Before request
//	        start := time.Now()
//
//	        c.Next()  // Process request
//
//	        // After request
//	        duration := time.Since(start)
//	        log.Printf("Request took %v", duration)
//	    }
//	}
package middleware
