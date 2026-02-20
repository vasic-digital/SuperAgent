package tests

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	fmt.Println("║     PRECONDITION: Container Boot Verification                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")

	// STEP 1: Run precondition check FIRST - this is mandatory
	if err := runPreconditionCheck(); err != nil {
		fmt.Printf("FATAL: Precondition check failed: %v\n", err)
		fmt.Println("Tests cannot proceed without proper container infrastructure.")
		fmt.Println("Run 'make test-infra-start' or configure Containers/.env for remote distribution.")
		os.Exit(1)
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

// runPreconditionCheck verifies container infrastructure is available
func runPreconditionCheck() error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	containersEnvPath := filepath.Join(projectRoot, "Containers", ".env")
	remoteConfig := parseContainersEnv(containersEnvPath)

	if remoteConfig.Enabled {
		fmt.Printf("Remote distribution enabled: %s\n", remoteConfig.HostsSummary())
		fmt.Println("Verifying remote containers are accessible...")
		return verifyRemoteContainers(remoteConfig)
	}

	fmt.Println("No remote distribution configured - verifying local containers...")
	return verifyLocalContainers(projectRoot)
}

// RemoteConfig holds remote distribution configuration
type RemoteConfig struct {
	Enabled bool
	Hosts   []RemoteHost
	SSHUser string
	SSHKey  string
}

// RemoteHost represents a remote host configuration
type RemoteHost struct {
	Name    string
	Address string
	Port    int
	User    string
}

// HostsSummary returns a summary of configured hosts
func (rc *RemoteConfig) HostsSummary() string {
	if len(rc.Hosts) == 0 {
		return "no hosts configured"
	}
	names := make([]string, len(rc.Hosts))
	for i, h := range rc.Hosts {
		names[i] = fmt.Sprintf("%s (%s)", h.Name, h.Address)
	}
	return strings.Join(names, ", ")
}

// parseContainersEnv parses the Containers/.env file
func parseContainersEnv(path string) *RemoteConfig {
	rc := &RemoteConfig{}

	data, err := os.ReadFile(path)
	if err != nil {
		return rc
	}

	lines := strings.Split(string(data), "\n")
	hostMap := make(map[int]*RemoteHost)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		switch key {
		case "CONTAINERS_REMOTE_ENABLED":
			rc.Enabled = strings.EqualFold(value, "true")
		case "CONTAINERS_REMOTE_DEFAULT_SSH_USER":
			rc.SSHUser = value
		case "CONTAINERS_REMOTE_DEFAULT_SSH_KEY":
			rc.SSHKey = value
		}

		if strings.HasPrefix(key, "CONTAINERS_REMOTE_HOST_") {
			idx, field := parseHostKey(key)
			if idx > 0 {
				if _, exists := hostMap[idx]; !exists {
					hostMap[idx] = &RemoteHost{Port: 22}
				}
				switch field {
				case "NAME":
					hostMap[idx].Name = value
				case "ADDRESS":
					hostMap[idx].Address = value
				case "PORT":
					fmt.Sscanf(value, "%d", &hostMap[idx].Port)
				case "USER":
					hostMap[idx].User = value
				}
			}
		}
	}

	for _, host := range hostMap {
		if host.Name != "" && host.Address != "" {
			if host.User == "" {
				host.User = rc.SSHUser
			}
			rc.Hosts = append(rc.Hosts, *host)
		}
	}

	return rc
}

// parseHostKey parses CONTAINERS_REMOTE_HOST_N_FIELD keys
func parseHostKey(key string) (int, string) {
	parts := strings.Split(key, "_")
	if len(parts) < 5 {
		return 0, ""
	}
	var idx int
	fmt.Sscanf(parts[3], "%d", &idx)
	field := strings.Join(parts[4:], "_")
	return idx, field
}

// verifyRemoteContainers verifies remote container infrastructure
func verifyRemoteContainers(rc *RemoteConfig) error {
	fmt.Println("Verifying remote container infrastructure...")

	for _, host := range rc.Hosts {
		fmt.Printf("Checking host: %s (%s:%d)\n", host.Name, host.Address, host.Port)

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host.Address, host.Port), 10*time.Second)
		if err != nil {
			return fmt.Errorf("cannot connect to host %s at %s:%d: %w", host.Name, host.Address, host.Port, err)
		}
		conn.Close()
		fmt.Printf("  ✓ Host %s is reachable\n", host.Name)
	}

	// Check HelixAgent health endpoint
	if err := checkHTTPWithTimeout("http://localhost:7061/health", 10*time.Second); err != nil {
		fmt.Printf("  WARNING: HelixAgent health check failed: %v\n", err)
	} else {
		fmt.Println("  ✓ HelixAgent is healthy")
	}

	return nil
}

// verifyLocalContainers verifies local container infrastructure
func verifyLocalContainers(projectRoot string) error {
	fmt.Println("Verifying local container infrastructure...")

	requiredServices := []struct {
		name        string
		port        int
		healthURL   string
		description string
	}{
		{"PostgreSQL", 15432, "", "Primary database"},
		{"Redis", 16379, "", "Cache and session store"},
		{"ChromaDB", 8001, "http://localhost:8001/api/v2/heartbeat", "Vector store"},
		{"HelixAgent", 7061, "http://localhost:7061/health", "Main service"},
	}

	allHealthy := true

	fmt.Println("\nRequired Services:")
	for _, svc := range requiredServices {
		healthy := checkTCP("localhost", svc.port)
		if svc.healthURL != "" {
			healthy = checkHTTPWithTimeout(svc.healthURL, 5*time.Second) == nil
		}

		if !healthy {
			fmt.Printf("  ✗ %s (port %d) - %s - NOT AVAILABLE\n", svc.name, svc.port, svc.description)
			allHealthy = false
		} else {
			fmt.Printf("  ✓ %s (port %d) - %s\n", svc.name, svc.port, svc.description)
		}
	}

	if !allHealthy {
		return fmt.Errorf("one or more required services are not running - run 'make test-infra-start' to start them")
	}

	fmt.Println("\nHelixAgent MCP/LSP/ACP/RAG Endpoints:")
	endpoints := []struct {
		name string
		url  string
	}{
		{"MCP", "http://localhost:7061/v1/mcp"},
		{"LSP", "http://localhost:7061/v1/lsp"},
		{"ACP", "http://localhost:7061/v1/acp"},
		{"Embeddings", "http://localhost:7061/v1/embeddings"},
		{"RAG", "http://localhost:7061/v1/rag"},
		{"Formatters", "http://localhost:7061/v1/formatters"},
		{"Vision", "http://localhost:7061/v1/vision"},
		{"Monitoring", "http://localhost:7061/v1/monitoring"},
	}

	for _, ep := range endpoints {
		if err := checkHTTPWithTimeout(ep.url, 3*time.Second); err != nil {
			fmt.Printf("  ~ %s endpoint at %s - not responding\n", ep.name, ep.url)
		} else {
			fmt.Printf("  ✓ %s endpoint available\n", ep.name)
		}
	}

	return nil
}

// checkHTTPWithTimeout checks if an HTTP endpoint is ready
func checkHTTPWithTimeout(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}
	return nil
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
