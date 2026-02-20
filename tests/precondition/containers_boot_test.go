package precondition

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

func TestPreconditionContainersBoot(t *testing.T) {
	t.Log("╔══════════════════════════════════════════════════════════════════╗")
	t.Log("║     PRECONDITION TEST: Container Boot Verification              ║")
	t.Log("║     This test MUST pass before any other test can execute        ║")
	t.Log("╚══════════════════════════════════════════════════════════════════╝")

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	containersEnvPath := filepath.Join(projectRoot, "Containers", ".env")

	remoteConfig := parseContainersEnv(containersEnvPath)
	if remoteConfig.Enabled {
		t.Logf("Remote distribution enabled: %s", remoteConfig.HostsSummary())
		t.Log("Verifying remote containers are accessible...")
		if err := verifyRemoteContainers(t, remoteConfig); err != nil {
			t.Fatalf("Remote container verification failed: %v", err)
		}
	} else {
		t.Log("No remote distribution configured - verifying local containers...")
		if err := verifyLocalContainers(t, projectRoot); err != nil {
			t.Fatalf("Local container verification failed: %v", err)
		}
	}

	// Check for multiple service instances (priority conflict detection)
	t.Log("\nChecking for multiple service instance conflicts...")
	if err := detectMultipleServiceInstances(t, remoteConfig); err != nil {
		t.Fatalf("Multiple service instance conflict detected: %v", err)
	}

	t.Log("╔══════════════════════════════════════════════════════════════════╗")
	t.Log("║     PRECONDITION PASSED: All containers verified                 ║")
	t.Log("╚══════════════════════════════════════════════════════════════════╝")
}

// detectMultipleServiceInstances checks for multiple instances of the same service
// at the same priority level, which would cause boot failures
func detectMultipleServiceInstances(t *testing.T, rc *RemoteConfig) error {
	// Service priorities: cloud (highest) > local network > localhost (lowest)
	// Fail if multiple instances at same priority level

	servicePorts := map[string]int{
		"postgres":   5432,
		"redis":      6379,
		"chromadb":   8000,
		"helixagent": 7061,
	}

	// Check local Docker/Podman containers
	containerCmd, containerArgs := detectContainerRuntime()
	if containerCmd != "" {
		t.Log("Checking for duplicate container instances...")

		for svcName := range servicePorts {
			instances := findContainerInstances(containerCmd, containerArgs, svcName)
			if len(instances) > 1 {
				return fmt.Errorf("MULTIPLE_INSTANCES: Found %d instances of '%s' service running locally: %v. Only one instance per service allowed.",
					len(instances), svcName, instances)
			}
			if len(instances) == 1 {
				t.Logf("  ✓ Single instance of %s found: %s", svcName, instances[0])
			}
		}
	}

	// Check for duplicate host names in remote config
	if rc.Enabled && len(rc.Hosts) > 0 {
		seenNames := make(map[string]string)
		seenAddresses := make(map[string]string)

		for _, host := range rc.Hosts {
			if existingAddr, exists := seenNames[host.Name]; exists {
				return fmt.Errorf("DUPLICATE_HOST: Host name '%s' defined multiple times (addresses: %s, %s). Each host must have a unique name.",
					host.Name, existingAddr, host.Address)
			}
			if existingName, exists := seenAddresses[host.Address]; exists {
				return fmt.Errorf("DUPLICATE_ADDRESS: Address '%s' used by multiple hosts (%s, %s). Each address must be unique.",
					host.Address, existingName, host.Name)
			}
			seenNames[host.Name] = host.Address
			seenAddresses[host.Address] = host.Name
		}
		t.Log("  ✓ No duplicate remote hosts configured")
	}

	// Check for port conflicts on localhost
	t.Log("Checking for local port conflicts...")
	for svcName, port := range servicePorts {
		testPorts := []int{port, port + 10000} // Check default and test ports

		var foundPorts []int
		for _, p := range testPorts {
			if checkTCPPort(p) {
				foundPorts = append(foundPorts, p)
			}
		}

		if len(foundPorts) > 1 {
			return fmt.Errorf("PORT_CONFLICT: Service '%s' appears to be running on multiple ports: %v. Cannot determine which instance to use.",
				svcName, foundPorts)
		}
	}

	t.Log("  ✓ No multiple service instance conflicts detected")
	return nil
}

// detectContainerRuntime detects the available container runtime
func detectContainerRuntime() (string, []string) {
	runtimes := []string{"docker", "podman"}
	for _, runtime := range runtimes {
		checkCmd := exec.Command(runtime, "compose", "version")
		if err := checkCmd.Run(); err == nil {
			return runtime, []string{"compose"}
		}
		standalone := runtime + "-compose"
		if _, err := exec.LookPath(standalone); err == nil {
			return standalone, nil
		}
	}
	return "", nil
}

// findContainerInstances finds running container instances for a service
func findContainerInstances(cmd string, args []string, serviceName string) []string {
	fullArgs := append(args, "ps", "--filter", "name="+serviceName, "--format", "{{.Names}}")
	execCmd := exec.Command(cmd, fullArgs...)
	output, err := execCmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var instances []string
	for _, line := range lines {
		if line != "" {
			instances = append(instances, line)
		}
	}
	return instances
}

type RemoteConfig struct {
	Enabled bool
	Hosts   []RemoteHost
	SSHUser string
	SSHKey  string
}

type RemoteHost struct {
	Name    string
	Address string
	Port    int
	User    string
}

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

func verifyRemoteContainers(t *testing.T, rc *RemoteConfig) error {
	t.Log("Verifying remote container infrastructure...")

	for _, host := range rc.Hosts {
		t.Logf("Checking host: %s (%s:%d)", host.Name, host.Address, host.Port)

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host.Address, host.Port), 10*time.Second)
		if err != nil {
			return fmt.Errorf("cannot connect to host %s at %s:%d: %w", host.Name, host.Address, host.Port, err)
		}
		conn.Close()
		t.Logf("  ✓ Host %s is reachable", host.Name)
	}

	services := []struct {
		name string
		url  string
	}{
		{"HelixAgent", "http://localhost:7061/health"},
	}

	for _, svc := range services {
		if err := checkHTTPHealth(svc.url, 10*time.Second); err != nil {
			t.Logf("  WARNING: %s health check failed: %v", svc.name, err)
		} else {
			t.Logf("  ✓ %s is healthy", svc.name)
		}
	}

	return nil
}

func verifyLocalContainers(t *testing.T, projectRoot string) error {
	t.Log("Verifying local container infrastructure...")

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

	optionalServices := []struct {
		name        string
		port        int
		healthURL   string
		description string
	}{
		{"Mem0", 8080, "http://localhost:8080/health", "Memory service"},
		{"Cognee", 8000, "http://localhost:8000/", "Knowledge graph"},
	}

	allHealthy := true

	t.Log("\nRequired Services:")
	for _, svc := range requiredServices {
		healthy := checkTCPPort(svc.port)
		if svc.healthURL != "" {
			healthy = checkHTTPHealth(svc.healthURL, 5*time.Second) == nil
		}

		if !healthy {
			t.Errorf("  ✗ %s (port %d) - %s - NOT AVAILABLE", svc.name, svc.port, svc.description)
			allHealthy = false
		} else {
			t.Logf("  ✓ %s (port %d) - %s", svc.name, svc.port, svc.description)
		}
	}

	t.Log("\nOptional Services:")
	for _, svc := range optionalServices {
		healthy := checkTCPPort(svc.port)
		if svc.healthURL != "" {
			healthy = checkHTTPHealth(svc.healthURL, 5*time.Second) == nil
		}

		if !healthy {
			t.Logf("  ~ %s (port %d) - %s - not running (optional)", svc.name, svc.port, svc.description)
		} else {
			t.Logf("  ✓ %s (port %d) - %s", svc.name, svc.port, svc.description)
		}
	}

	if !allHealthy {
		return fmt.Errorf("one or more required services are not running - run 'make test-infra-start' to start them")
	}

	t.Log("\nHelixAgent MCP/LSP/ACP/RAG Endpoints:")
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
		if err := checkHTTPHealth(ep.url, 3*time.Second); err != nil {
			t.Logf("  ~ %s endpoint at %s - not responding", ep.name, ep.url)
		} else {
			t.Logf("  ✓ %s endpoint available", ep.name)
		}
	}

	return nil
}

func checkTCPPort(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func checkHTTPHealth(url string, timeout time.Duration) error {
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

func findProjectRoot() (string, error) {
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

	return "", fmt.Errorf("project root not found")
}
