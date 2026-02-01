package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
)

// ServiceHealthChecker performs health checks against service endpoints using their configured check type.
type ServiceHealthChecker struct {
	Logger *logrus.Logger
}

// NewServiceHealthChecker creates a new ServiceHealthChecker.
func NewServiceHealthChecker(logger *logrus.Logger) *ServiceHealthChecker {
	return &ServiceHealthChecker{Logger: logger}
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

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Errorf("TCP connection to %s (%s) failed: %w", name, addr, err)
	}
	conn.Close()
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
		// If the URL doesn't have a scheme, add http://
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
