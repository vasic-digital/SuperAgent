package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"dev.helix.agent/internal/adapters/containers"
	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
)

// DefaultHealthCheckTimeout is the default overall timeout for batch health
// check operations. Individual service timeouts are still respected, but
// the overall batch will not exceed this deadline.
const DefaultHealthCheckTimeout = 30 * time.Second

// DefaultMaxConcurrentChecks limits how many health checks run in parallel
// to avoid overwhelming the system with network connections.
const DefaultMaxConcurrentChecks = 10

// ServiceHealthChecker performs health checks against service endpoints
// using their configured check type. When a ContainerAdapter is set,
// delegates checks through the Containers module.
type ServiceHealthChecker struct {
	Logger              *logrus.Logger
	ContainerAdapter    *containers.Adapter
	BatchTimeout        time.Duration // Overall timeout for CheckAll/batch operations
	MaxConcurrentChecks int           // Max parallel health checks
}

// NewServiceHealthChecker creates a new ServiceHealthChecker.
func NewServiceHealthChecker(logger *logrus.Logger) *ServiceHealthChecker {
	return &ServiceHealthChecker{
		Logger:              logger,
		BatchTimeout:        DefaultHealthCheckTimeout,
		MaxConcurrentChecks: DefaultMaxConcurrentChecks,
	}
}

// Check dispatches to the appropriate health check based on the endpoint's HealthType.
func (hc *ServiceHealthChecker) Check(name string, ep config.ServiceEndpoint) error {
	switch ep.HealthType {
	case "pgx":
		return hc.checkTCP(name, ep)
	case "redis":
		return hc.checkTCP(name, ep)
	case "http":
		return hc.checkHTTP(name, ep)
	case "tcp":
		return hc.checkTCP(name, ep)
	default:
		return hc.checkTCP(name, ep)
	}
}

// CheckWithRetry performs a health check with retries.
func (hc *ServiceHealthChecker) CheckWithRetry(name string, ep config.ServiceEndpoint) error {
	retries := ep.RetryCount
	if retries <= 0 {
		retries = 1
	}

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		lastErr = hc.Check(name, ep)
		if lastErr == nil {
			return nil
		}
		if attempt < retries {
			hc.Logger.WithFields(logrus.Fields{
				"service": name,
				"attempt": attempt,
				"max":     retries,
				"error":   lastErr,
			}).Debug("Health check failed, retrying...")
			time.Sleep(2 * time.Second)
		}
	}
	return fmt.Errorf("service %s health check failed after %d attempts: %w", name, retries, lastErr)
}

func (hc *ServiceHealthChecker) checkTCP(name string, ep config.ServiceEndpoint) error {
	addr := ep.ResolvedURL()
	if addr == "" {
		return fmt.Errorf("no address configured for service %s", name)
	}

	timeout := ep.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Delegate to Containers module adapter when available.
	if hc.ContainerAdapter != nil {
		result, err := hc.ContainerAdapter.HealthCheck(
			context.Background(),
			name, ep.Host, ep.Port, "", "tcp", timeout,
		)
		if err != nil {
			return fmt.Errorf("TCP health check for %s failed: %w", name, err)
		}
		if !result.Healthy {
			return fmt.Errorf("TCP health check for %s failed: %s", name, result.Error)
		}
		return nil
	}

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Errorf("TCP connection to %s (%s) failed: %w", name, addr, err)
	}
	_ = conn.Close()
	return nil
}

func (hc *ServiceHealthChecker) checkHTTP(name string, ep config.ServiceEndpoint) error {
	baseURL := ep.ResolvedURL()
	if baseURL == "" {
		return fmt.Errorf("no address configured for service %s", name)
	}

	// Build the full URL
	url := baseURL
	if ep.HealthPath != "" {
		if len(url) > 0 && url[0] != 'h' {
			url = "http://" + url
		}
		url = url + ep.HealthPath
	} else {
		if len(url) > 0 && url[0] != 'h' {
			url = "http://" + url
		}
	}

	timeout := ep.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Delegate to Containers module adapter when available.
	if hc.ContainerAdapter != nil {
		result, err := hc.ContainerAdapter.HealthCheck(
			context.Background(),
			name, ep.Host, ep.Port, ep.HealthPath, "http",
			timeout,
		)
		if err != nil {
			return fmt.Errorf("HTTP health check for %s failed: %w", name, err)
		}
		if !result.Healthy {
			return fmt.Errorf("HTTP health check for %s failed: %s", name, result.Error)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", name, err)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP health check for %s (%s) failed: %w", name, url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("HTTP health check for %s returned status %d", name, resp.StatusCode)
	}

	return nil
}

// CheckWithContext performs a health check that respects the provided context
// for cancellation and deadline. If the context is cancelled or its deadline
// expires, the check returns immediately with the context error.
func (hc *ServiceHealthChecker) CheckWithContext(ctx context.Context, name string, ep config.ServiceEndpoint) error {
	done := make(chan error, 1)
	go func() {
		done <- hc.Check(name, ep)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("health check for %s cancelled: %w", name, ctx.Err())
	}
}

// HealthCheckResult holds the result of a single service health check.
type HealthCheckResult struct {
	Name     string
	Error    error
	Duration time.Duration
}

// CheckAllNonBlocking runs health checks for all provided endpoints concurrently
// with bounded parallelism and an overall deadline. It does not block indefinitely;
// the overall timeout is controlled by BatchTimeout (or the provided context,
// whichever expires first). Results are returned for all endpoints that completed
// within the deadline. Services that did not complete in time are reported with
// a timeout error.
func (hc *ServiceHealthChecker) CheckAllNonBlocking(
	ctx context.Context,
	endpoints map[string]config.ServiceEndpoint,
) map[string]*HealthCheckResult {
	// Apply overall batch timeout as a deadline
	batchTimeout := hc.BatchTimeout
	if batchTimeout <= 0 {
		batchTimeout = DefaultHealthCheckTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, batchTimeout)
	defer cancel()

	maxConcurrent := hc.MaxConcurrentChecks
	if maxConcurrent <= 0 {
		maxConcurrent = DefaultMaxConcurrentChecks
	}
	sem := make(chan struct{}, maxConcurrent)

	results := make(map[string]*HealthCheckResult, len(endpoints))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, ep := range endpoints {
		wg.Add(1)
		go func(name string, ep config.ServiceEndpoint) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				results[name] = &HealthCheckResult{
					Name:  name,
					Error: fmt.Errorf("health check for %s skipped: %w", name, ctx.Err()),
				}
				mu.Unlock()
				return
			}

			start := time.Now()
			err := hc.CheckWithContext(ctx, name, ep)
			duration := time.Since(start)

			mu.Lock()
			results[name] = &HealthCheckResult{
				Name:     name,
				Error:    err,
				Duration: duration,
			}
			mu.Unlock()

			if err != nil {
				hc.Logger.WithFields(logrus.Fields{
					"service":  name,
					"duration": duration,
					"error":    err,
				}).Debug("Non-blocking health check failed")
			} else {
				hc.Logger.WithFields(logrus.Fields{
					"service":  name,
					"duration": duration,
				}).Debug("Non-blocking health check passed")
			}
		}(name, ep)
	}

	// Wait for all goroutines to complete (they all respect the context deadline)
	wg.Wait()

	return results
}
