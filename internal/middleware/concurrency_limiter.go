package middleware

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ConcurrencyLimiter returns a Gin middleware that caps in-flight HTTP requests.
// If the in-flight count reaches maxInFlight the request is rejected immediately
// with 503 Service Unavailable instead of queuing, providing hard backpressure.
// The limit can be overridden at runtime via the MAX_IN_FLIGHT_REQUESTS env var.
func ConcurrencyLimiter(maxInFlight int) gin.HandlerFunc {
	if envMax := os.Getenv("MAX_IN_FLIGHT_REQUESTS"); envMax != "" {
		if v, err := strconv.Atoi(envMax); err == nil && v > 0 {
			maxInFlight = v
		}
	}
	sem := make(chan struct{}, maxInFlight)
	return func(c *gin.Context) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusServiceUnavailable,
				gin.H{"error": "server at capacity", "retry_after": "1"})
		}
	}
}
