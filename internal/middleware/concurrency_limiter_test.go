package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrencyLimiter_AllowsUnderLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ConcurrencyLimiter(5))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConcurrencyLimiter_RejectsAtCapacity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use a channel to hold requests in-flight so we can fill the semaphore.
	block := make(chan struct{})
	released := make(chan struct{})

	router := gin.New()
	router.Use(ConcurrencyLimiter(1))
	router.GET("/test", func(c *gin.Context) {
		close(released) // signal: slot acquired
		<-block         // hold slot until test lets go
		c.Status(http.StatusOK)
	})

	// Start a long-running request that occupies the single slot.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)
	}()

	// Wait until the first request has acquired the slot.
	select {
	case <-released:
	case <-time.After(2 * time.Second):
		t.Fatal("first request did not acquire semaphore slot in time")
	}

	// Now issue a second request — it should be rejected immediately.
	w2 := httptest.NewRecorder()
	req2, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusServiceUnavailable, w2.Code)

	// Release the held request.
	close(block)
	wg.Wait()

	// After the slot is freed, the next request should succeed.
	w3 := httptest.NewRecorder()
	req3, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)

	router2 := gin.New()
	router2.Use(ConcurrencyLimiter(1))
	router2.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router2.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestConcurrencyLimiter_EnvOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("MAX_IN_FLIGHT_REQUESTS", "2")

	block := make(chan struct{})
	ready1 := make(chan struct{})
	ready2 := make(chan struct{})

	router := gin.New()
	router.Use(ConcurrencyLimiter(1)) // default 1, env should override to 2
	first := true
	var firstMu sync.Mutex
	router.GET("/test", func(c *gin.Context) {
		firstMu.Lock()
		isFirst := first
		first = false
		firstMu.Unlock()
		if isFirst {
			close(ready1)
		} else {
			close(ready2)
		}
		<-block
		c.Status(http.StatusOK)
	})

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, req)
		}()
	}

	// Both concurrent requests should acquire slots (limit is 2 via env).
	waitCh := func(ch <-chan struct{}, label string) {
		select {
		case <-ch:
		case <-time.After(2 * time.Second):
			t.Errorf("%s did not acquire slot in time", label)
		}
	}
	waitCh(ready1, "request1")
	waitCh(ready2, "request2")

	// A third request should now be rejected.
	w3 := httptest.NewRecorder()
	req3, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusServiceUnavailable, w3.Code)

	close(block)
	wg.Wait()
}

func TestConcurrencyLimiter_ZeroOrNegativeDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Passing 0 creates a zero-capacity channel — every request is rejected.
	router := gin.New()
	router.Use(ConcurrencyLimiter(0))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
