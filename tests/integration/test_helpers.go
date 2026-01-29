package integration

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// TestConfig holds common test configuration
type TestConfig struct {
	HelixAgentURL     string
	HelixAgentHost    string
	HelixAgentPort    string
	HelixAgentAPIKey  string
	BaseURL           string
	BinaryPath        string
	TempDir           string
	PostgresHost      string
	PostgresPort      string
	PostgresDB        string
	PostgresUser      string
	PostgresPass      string
	RedisHost         string
	RedisPort         string
	RedisPass         string
	ChromaDBURL       string
	CogneeURL         string
}

// MCPServerConfig holds MCP server configuration for testing
type MCPServerConfig struct {
	Name        string
	Type        string // "local" or "remote"
	Command     []string
	Args        []string
	URL         string
	PackageName string
	Env         map[string]string
}

// getEnv retrieves an environment variable with a fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isServerRunning checks if a server is running on the given host:port
func isServerRunning(host, port string) bool {
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// waitForHTTP waits for an HTTP server to be ready
func waitForHTTP(url string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}
