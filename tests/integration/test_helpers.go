package integration

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// TestConfig holds common test configuration
type TestConfig struct {
	HelixAgentURL    string
	HelixAgentHost   string
	HelixAgentPort   string
	HelixAgentAPIKey string
	BaseURL          string
	BinaryPath       string
	TempDir          string
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresUser     string
	PostgresPass     string
	RedisHost        string
	RedisPort        string
	RedisPass        string
	ChromaDBURL      string
	CogneeURL        string
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

// containerRuntimeOnce caches the detected container runtime name.
var (
	containerRuntimeOnce sync.Once
	containerRuntimeName string // "podman" or "docker"
)

// containerRuntime returns the available container runtime ("podman" or
// "docker"). The result is cached after the first call. Returns empty
// string if neither is available.
func containerRuntime() string {
	containerRuntimeOnce.Do(func() {
		if _, err := exec.LookPath("podman"); err == nil {
			containerRuntimeName = "podman"
		} else if _, err := exec.LookPath("docker"); err == nil {
			containerRuntimeName = "docker"
		}
	})
	return containerRuntimeName
}

// containerExec runs a container command with the detected runtime,
// returning combined output. Falls back from podman to docker
// automatically.
func containerExec(args ...string) ([]byte, error) {
	rt := containerRuntime()
	if rt == "" {
		return nil, fmt.Errorf("no container runtime found")
	}
	cmd := exec.Command(rt, args...)
	return cmd.CombinedOutput()
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
