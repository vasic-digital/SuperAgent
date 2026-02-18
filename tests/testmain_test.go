package tests

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestMain is the entry point for all tests in this package.
// It ensures infrastructure is running before any tests execute.
func TestMain(m *testing.M) {
	// Check if we should skip infrastructure setup
	if os.Getenv("SKIP_INFRA_SETUP") == "true" {
		os.Exit(m.Run())
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     AUTO-BOOTING INFRASTRUCTURE FOR TESTS                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")

	// Ensure infrastructure is running
	if err := ensureInfrastructure(); err != nil {
		fmt.Printf("WARNING: Infrastructure setup had issues: %v\n", err)
		// Continue anyway - some tests may still work
	}

	// Wait for services to be ready
	waitForServices()

	// Run tests
	code := m.Run()

	os.Exit(code)
}

// ensureInfrastructure runs the infrastructure startup script
func ensureInfrastructure() error {
	// Find the project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Check if infrastructure script exists
	scriptPath := filepath.Join(projectRoot, "scripts", "ensure-infrastructure.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Println("Infrastructure script not found, attempting direct compose...")
		return startInfrastructureDirect(projectRoot)
	}

	// Run the infrastructure script
	fmt.Println("Running infrastructure startup script...")
	cmd := exec.Command(scriptPath, "start")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("infrastructure script failed: %w", err)
	}

	return nil
}

// startInfrastructureDirect starts infrastructure using docker-compose directly.
// Uses centralized container runtime detection.
func startInfrastructureDirect(projectRoot string) error {
	// Detect compose command via centralized helper.
	composeCmd, composeArgs := detectComposeCommand()
	if composeCmd == "" {
		return fmt.Errorf("no compose command found")
	}

	// Start core services
	args := append(composeArgs, "--profile", "default", "up", "-d", "postgres", "redis", "chromadb", "cognee")
	cmd := exec.Command(composeCmd, args...)
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		// Try without profile
		args = append(composeArgs, "up", "-d", "postgres", "redis", "chromadb")
		cmd = exec.Command(composeCmd, args...)
		cmd.Dir = projectRoot
		cmd.Run() // Ignore error
	}

	return nil
}

// containerRuntimes lists the container runtimes to probe, in preference order.
var containerRuntimes = []string{"docker", "podman"}

// detectComposeCommand returns the compose command and initial args.
// It probes each runtime in containerRuntimes order, checking for the
// compose plugin first (e.g. "docker compose") then the standalone binary
// (e.g. "docker-compose").
func detectComposeCommand() (string, []string) {
	for _, runtime := range containerRuntimes {
		// Try "<runtime> compose" (compose plugin)
		checkCmd := exec.Command(runtime, "compose", "version")
		if err := checkCmd.Run(); err == nil {
			return runtime, []string{"compose"}
		}
		// Try <runtime>-compose standalone binary
		standalone := runtime + "-compose"
		if _, err := exec.LookPath(standalone); err == nil {
			return standalone, nil
		}
	}
	return "", nil
}

// waitForServices waits for critical services to be ready
func waitForServices() {
	fmt.Println("Waiting for services to be ready...")

	services := []struct {
		name    string
		check   func() bool
		timeout time.Duration
	}{
		{"PostgreSQL", checkPostgres, 60 * time.Second},
		{"Redis", checkRedis, 30 * time.Second},
		{"ChromaDB", func() bool { return checkHTTP("http://localhost:8001/api/v2/heartbeat") }, 60 * time.Second},
		{"Cognee", func() bool { return checkHTTP("http://localhost:8000/") }, 90 * time.Second},
		{"HelixAgent", func() bool { return checkHTTP("http://localhost:7061/health") }, 30 * time.Second},
	}

	for _, svc := range services {
		start := time.Now()
		for time.Since(start) < svc.timeout {
			if svc.check() {
				fmt.Printf("  ✓ %s ready\n", svc.name)
				break
			}
			time.Sleep(2 * time.Second)
		}
	}
}

// findProjectRoot finds the project root directory
func findProjectRoot() (string, error) {
	// Try current directory first
	if _, err := os.Stat("go.mod"); err == nil {
		return ".", nil
	}

	// Try parent directories
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Default to known location
	return "/run/media/milosvasic/DATA4TB/Projects/HelixAgent", nil
}

// checkPostgres checks if PostgreSQL is ready
func checkPostgres() bool {
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "15432" // Default port for test infrastructure
	}
	conn, err := net.DialTimeout("tcp", "localhost:"+port, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// checkRedis checks if Redis is ready
func checkRedis() bool {
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "16379" // Default port for test infrastructure
	}
	conn, err := net.DialTimeout("tcp", "localhost:"+port, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// checkHTTP checks if an HTTP endpoint is ready
func checkHTTP(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// checkTCP checks if a TCP port is open
func checkTCP(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
