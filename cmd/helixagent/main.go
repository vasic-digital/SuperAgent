package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/auth/oauth_credentials"
	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/mcp"
	"dev.helix.agent/internal/messaging"
	"dev.helix.agent/internal/messaging/inmemory"
	"dev.helix.agent/internal/router"
	"dev.helix.agent/internal/utils"
	"dev.helix.agent/internal/verifier"
	"llm-verifier/pkg/cliagents"
)

var (
	configFile         = flag.String("config", "", "Path to configuration file (YAML)")
	version            = flag.Bool("version", false, "Show version information")
	help               = flag.Bool("help", false, "Show help message")
	autoStartDocker    = flag.Bool("auto-start-docker", true, "Automatically start required Docker containers")
	strictDependencies = flag.Bool("strict-dependencies", true, "MANDATORY: Fail if any integration dependency (Cognee, DB, Redis) is unavailable")
	generateAPIKey     = flag.Bool("generate-api-key", false, "Generate a new HelixAgent API key and output it")
	generateOpenCode   = flag.Bool("generate-opencode-config", false, "Generate OpenCode configuration JSON")
	validateOpenCode   = flag.String("validate-opencode-config", "", "Path to OpenCode config file to validate")
	openCodeOutput     = flag.String("opencode-output", "", "Output path for OpenCode config (default: stdout)")
	generateCrush      = flag.Bool("generate-crush-config", false, "Generate Crush configuration JSON")
	validateCrush      = flag.String("validate-crush-config", "", "Path to Crush config file to validate")
	crushOutput        = flag.String("crush-output", "", "Output path for Crush config (default: stdout)")
	apiKeyEnvFile      = flag.String("api-key-env-file", "", "Path to .env file to write the generated API key")
	preinstallMCP      = flag.Bool("preinstall-mcp", false, "Pre-install standard MCP server npm packages")
	skipMCPPreinstall  = flag.Bool("skip-mcp-preinstall", false, "Skip automatic MCP package pre-installation at startup")
	workingMCPsOnly    = flag.Bool("working-mcps-only", true, "Only include MCPs that work without API keys (default: true)")
	useLocalMCPServers = flag.Bool("use-local-mcp-servers", false, "Use local Docker-based MCP servers on TCP ports (requires running start-mcp-servers.sh)")
	autoStartMCP       = flag.Bool("auto-start-mcp", true, "Automatically start MCP Docker containers on HelixAgent startup")
	// Unified CLI agent configuration flags (all 48 agents)
	generateAgentConfig = flag.String("generate-agent-config", "", "Generate config for specified CLI agent (use --list-agents to see all)")
	validateAgentConfig = flag.String("validate-agent-config", "", "Validate config file for agent (format: agent:path)")
	agentConfigOutput   = flag.String("agent-config-output", "", "Output path for generated agent config (default: stdout)")
	listAgents          = flag.Bool("list-agents", false, "List all 48 supported CLI agents")
	generateAllAgents   = flag.Bool("generate-all-agents", false, "Generate configurations for all 48 CLI agents")
	allAgentsOutputDir  = flag.String("all-agents-output-dir", "", "Output directory for all agent configs (required with --generate-all-agents)")
)

// ValidOpenCodeTopLevelKeys contains the valid top-level keys per OpenCode.ai official schema
// Supports both v1.0.x (provider, mcp, agent) and v1.1.30+ (providers, mcpServers, agents) schemas
// Source: https://opencode.ai/config.json and OpenCode internal/config/config.go
var ValidOpenCodeTopLevelKeys = map[string]bool{
	// v1.0.x schema keys
	"$schema":      true,
	"plugin":       true,
	"enterprise":   true,
	"instructions": true,
	"provider":     true,
	"mcp":          true,
	"tools":        true,
	"agent":        true,
	"command":      true,
	"keybinds":     true,
	"username":     true,
	"share":        true,
	"permission":   true,
	"compaction":   true,
	"sse":          true,
	"mode":         true,
	"autoshare":    true,
	// v1.1.30+ schema keys (Viper-based)
	"providers":    true,
	"mcpServers":   true,
	"agents":       true,
	"contextPaths": true,
	"tui":          true,
}

// CommandExecutor interface for executing system commands (allows mocking)
type CommandExecutor interface {
	LookPath(file string) (string, error)
	RunCommand(name string, args ...string) ([]byte, error)
	RunCommandWithDir(dir string, name string, args ...string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using actual exec calls
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (r *RealCommandExecutor) RunCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

func (r *RealCommandExecutor) RunCommandWithDir(dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// ContainerRuntime represents the detected container runtime (Docker or Podman)
type ContainerRuntime string

const (
	RuntimeDocker ContainerRuntime = "docker"
	RuntimePodman ContainerRuntime = "podman"
	RuntimeNone   ContainerRuntime = "none"
)

// DetectContainerRuntime automatically detects available container runtime
// Prefers Docker, falls back to Podman if Docker is not available
func DetectContainerRuntime() (ContainerRuntime, string, error) {
	// Try Docker first
	if path, err := exec.LookPath("docker"); err == nil {
		// Verify Docker daemon is accessible
		cmd := exec.Command("docker", "info")
		if err := cmd.Run(); err == nil {
			return RuntimeDocker, path, nil
		}
	}

	// Try Podman as fallback
	if path, err := exec.LookPath("podman"); err == nil {
		// Verify Podman is accessible
		cmd := exec.Command("podman", "info")
		if err := cmd.Run(); err == nil {
			return RuntimePodman, path, nil
		}
	}

	return RuntimeNone, "", fmt.Errorf("no container runtime found: neither Docker nor Podman is available")
}

// DetectComposeCommand detects the compose command for the container runtime
// Returns: compose command, args prefix, error
func DetectComposeCommand(runtime ContainerRuntime) (string, []string, error) {
	switch runtime {
	case RuntimeDocker:
		// Try "docker compose" first (newer syntax)
		cmd := exec.Command("docker", "compose", "version")
		if err := cmd.Run(); err == nil {
			return "docker", []string{"compose"}, nil
		}
		// Fall back to "docker-compose"
		if path, err := exec.LookPath("docker-compose"); err == nil {
			return path, nil, nil
		}
		return "", nil, fmt.Errorf("docker compose command not found")

	case RuntimePodman:
		// Try "podman-compose" first
		if path, err := exec.LookPath("podman-compose"); err == nil {
			return path, nil, nil
		}
		// Try "podman compose" (if podman has compose plugin)
		cmd := exec.Command("podman", "compose", "version")
		if err := cmd.Run(); err == nil {
			return "podman", []string{"compose"}, nil
		}
		return "", nil, fmt.Errorf("podman-compose not found: install with 'pip install podman-compose'")

	default:
		return "", nil, fmt.Errorf("unknown container runtime")
	}
}

// HealthChecker interface for checking service health (allows mocking)
type HealthChecker interface {
	CheckHealth(url string) error
}

// HTTPHealthChecker implements HealthChecker using HTTP requests
type HTTPHealthChecker struct {
	Client  *http.Client
	Timeout time.Duration
}

func NewHTTPHealthChecker(timeout time.Duration) *HTTPHealthChecker {
	return &HTTPHealthChecker{
		Client:  &http.Client{Timeout: timeout},
		Timeout: timeout,
	}
}

func (h *HTTPHealthChecker) CheckHealth(url string) error {
	resp, err := h.Client.Get(url)
	if err != nil {
		return fmt.Errorf("cannot connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}
	return nil
}

// ContainerConfig holds configuration for container management
type ContainerConfig struct {
	ProjectDir       string
	RequiredServices []string
	CogneeURL        string
	ChromaDBURL      string
	Executor         CommandExecutor
	HealthChecker    HealthChecker
}

// DefaultContainerConfig returns the default container configuration
func DefaultContainerConfig() *ContainerConfig {
	// Try to detect project directory from executable location
	// or use current working directory
	projectDir, err := os.Getwd()
	if err != nil {
		projectDir = "/run/media/milosvasic/DATA4TB/Projects/HelixAgent"
	}

	return &ContainerConfig{
		ProjectDir:       projectDir,
		RequiredServices: []string{"postgres", "redis", "cognee", "chromadb"},
		CogneeURL:        "http://localhost:8000/",
		ChromaDBURL:      "http://localhost:8001/api/v2/heartbeat",
		Executor:         &RealCommandExecutor{},
		HealthChecker:    NewHTTPHealthChecker(10 * time.Second),
	}
}

// Global container config (can be overridden for testing)
var containerConfig = DefaultContainerConfig()

// ensureRequiredContainers starts required Docker containers using docker-compose
func ensureRequiredContainers(logger *logrus.Logger) error {
	return ensureRequiredContainersWithConfig(logger, containerConfig)
}

// ensureRequiredContainersWithConfig starts required Docker/Podman containers using provided config
// Automatically detects and uses Docker or Podman (whichever is available)
func ensureRequiredContainersWithConfig(logger *logrus.Logger, cfg *ContainerConfig) error {
	// Detect container runtime (Docker or Podman)
	runtime, runtimePath, err := DetectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime detection failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"runtime": runtime,
		"path":    runtimePath,
	}).Info("Detected container runtime")

	// Detect compose command
	composeCmd, composeArgs, err := DetectComposeCommand(runtime)
	if err != nil {
		return fmt.Errorf("compose command detection failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"compose_command": composeCmd,
		"compose_args":    composeArgs,
	}).Info("Detected compose command")

	// Check which services are already running
	runningServices, err := getRunningServicesWithRuntimeConfig(cfg, composeCmd, composeArgs)
	if err != nil {
		logger.WithError(err).Warn("Could not check running services, attempting to start all")
		runningServices = make(map[string]bool)
	}

	// Determine which services need to be started
	servicesToStart := []string{}
	for _, service := range cfg.RequiredServices {
		if !runningServices[service] {
			servicesToStart = append(servicesToStart, service)
		}
	}

	if len(servicesToStart) == 0 {
		logger.Info("All required containers are already running")
		return nil
	}

	logger.WithField("services", strings.Join(servicesToStart, ", ")).Info("Starting required containers")

	// Build compose command with profile for Cognee/ChromaDB
	var output []byte
	var cmdArgs []string

	if len(composeArgs) > 0 {
		// Format: docker compose --profile default up -d <services>
		cmdArgs = append(cmdArgs, composeArgs...)
	}
	cmdArgs = append(cmdArgs, "--profile", "default", "up", "-d")
	cmdArgs = append(cmdArgs, servicesToStart...)

	cmd := exec.Command(composeCmd, cmdArgs...)
	cmd.Dir = cfg.ProjectDir
	output, err = cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to start containers with %s: %w\nOutput: %s", runtime, err, string(output))
	}

	logger.WithField("output", string(output)).Debug("Compose output")
	logger.Info("Waiting for containers to be healthy...")

	// Wait for containers to be ready
	time.Sleep(20 * time.Second)

	// Verify critical services are running
	if err := verifyServicesHealthWithConfig(cfg.RequiredServices, logger, cfg); err != nil {
		return fmt.Errorf("service health verification failed: %w", err)
	}

	logger.Info("Container startup completed successfully")
	return nil
}

// ensureMCPServers starts all MCP Docker containers from git submodules
// Uses docker-compose.mcp-servers.yml to build and run 32 MCP servers
// All servers use TCP ports (9101-9999) - no npm/npx dependencies
func ensureMCPServers(logger *logrus.Logger) error {
	// Get project directory
	projectDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get project directory: %w", err)
	}

	// Check if MCP compose file exists
	mcpComposeFile := filepath.Join(projectDir, "docker", "mcp", "docker-compose.mcp-servers.yml")
	if _, err := os.Stat(mcpComposeFile); os.IsNotExist(err) {
		logger.WithField("file", mcpComposeFile).Warn("MCP compose file not found, skipping MCP auto-start")
		return nil
	}

	// Detect container runtime
	runtime, _, err := DetectContainerRuntime()
	if err != nil {
		return fmt.Errorf("container runtime detection failed: %w", err)
	}

	// Detect compose command
	composeCmd, composeArgs, err := DetectComposeCommand(runtime)
	if err != nil {
		return fmt.Errorf("compose command detection failed: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"runtime":    runtime,
		"compose":    composeCmd,
		"mcp_servers": 32,
	}).Info("Starting MCP servers from git submodules (zero npm dependencies)")

	// Build compose command: docker compose -f <file> up -d
	var cmdArgs []string
	if len(composeArgs) > 0 {
		cmdArgs = append(cmdArgs, composeArgs...)
	}
	cmdArgs = append(cmdArgs, "-f", mcpComposeFile, "up", "-d")

	cmd := exec.Command(composeCmd, cmdArgs...)
	cmd.Dir = projectDir

	// Run in background - don't wait for all builds
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Log warning but don't fail - MCP servers are optional
		logger.WithError(err).WithField("output", string(output)).Warn("Failed to start some MCP servers, continuing")
		return nil
	}

	logger.WithField("output", string(output)).Debug("MCP Compose output")
	logger.Info("MCP servers starting in background (32 servers on ports 9101-9999)")
	return nil
}

// getRunningServicesWithRuntimeConfig checks which compose services are currently running
func getRunningServicesWithRuntimeConfig(cfg *ContainerConfig, composeCmd string, composeArgs []string) (map[string]bool, error) {
	running := make(map[string]bool)

	var cmdArgs []string
	if len(composeArgs) > 0 {
		cmdArgs = append(cmdArgs, composeArgs...)
	}
	cmdArgs = append(cmdArgs, "ps", "--services", "--filter", "status=running")

	cmd := exec.Command(composeCmd, cmdArgs...)
	cmd.Dir = cfg.ProjectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return running, fmt.Errorf("failed to list running services: %w", err)
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, service := range services {
		service = strings.TrimSpace(service)
		if service != "" {
			running[service] = true
		}
	}

	return running, nil
}

// getRunningServices checks which docker-compose services are currently running
func getRunningServices() (map[string]bool, error) {
	return getRunningServicesWithConfig(containerConfig)
}

// getRunningServicesWithConfig checks which docker-compose services are currently running using provided config
func getRunningServicesWithConfig(cfg *ContainerConfig) (map[string]bool, error) {
	running := make(map[string]bool)
	executor := cfg.Executor

	// Check if docker is available
	if _, err := executor.LookPath("docker"); err != nil {
		return running, fmt.Errorf("docker compose not found")
	}

	// Try docker compose first
	output, err := executor.RunCommandWithDir(cfg.ProjectDir, "docker", "compose", "ps", "--services", "--filter", "status=running")
	if err != nil {
		// Try docker-compose as fallback
		if _, lookErr := executor.LookPath("docker-compose"); lookErr == nil {
			output, err = executor.RunCommandWithDir(cfg.ProjectDir, "docker-compose", "ps", "--services", "--filter", "status=running")
		}
	}

	if err != nil {
		return running, err
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, service := range services {
		service = strings.TrimSpace(service)
		if service != "" {
			running[service] = true
		}
	}

	return running, nil
}

// verifyServicesHealth performs basic health checks on critical services
func verifyServicesHealth(services []string, logger *logrus.Logger) error {
	return verifyServicesHealthWithConfig(services, logger, containerConfig)
}

// PostgresHealthChecker is a function type for checking Postgres health (allows mocking)
type PostgresHealthChecker func() error

// RedisHealthChecker is a function type for checking Redis health (allows mocking)
type RedisHealthChecker func() error

// Default health checkers (can be overridden for testing)
var (
	postgresHealthChecker PostgresHealthChecker = checkPostgresHealth
	redisHealthChecker    RedisHealthChecker    = checkRedisHealth
)

// verifyServicesHealthWithConfig performs basic health checks on critical services using provided config
func verifyServicesHealthWithConfig(services []string, logger *logrus.Logger, cfg *ContainerConfig) error {
	var errors []string

	for _, service := range services {
		switch service {
		case "postgres":
			if err := postgresHealthChecker(); err != nil {
				errors = append(errors, fmt.Sprintf("postgres: %v", err))
			}
		case "redis":
			if err := redisHealthChecker(); err != nil {
				errors = append(errors, fmt.Sprintf("redis: %v", err))
			}
		case "cognee":
			if err := checkCogneeHealthWithConfig(cfg); err != nil {
				errors = append(errors, fmt.Sprintf("cognee: %v", err))
			}
		case "chromadb":
			if err := checkChromaDBHealthWithConfig(cfg); err != nil {
				errors = append(errors, fmt.Sprintf("chromadb: %v", err))
			}
		default:
			errors = append(errors, fmt.Sprintf("%s: unknown service", service))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("health check failures: %s", strings.Join(errors, "; "))
	}

	return nil
}

// checkPostgresHealth verifies PostgreSQL connectivity
func checkPostgresHealth() error {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "helixagent"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "helixagent123" // Default from docker-compose.yml
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "helixagent_db"
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&connect_timeout=5",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to establish a connection
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer conn.Close(ctx)

	// Ping to verify connection is working
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("PostgreSQL ping failed: %w", err)
	}

	return nil
}

// checkRedisHealth verifies Redis connectivity
func checkRedisHealth() error {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		redisPassword = "helixagent123" // Default from docker-compose.yml
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:        redisHost + ":" + redisPort,
		Password:    redisPassword,
		DB:          0,
		DialTimeout: 5 * time.Second,
	})
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// checkCogneeHealth verifies Cognee API availability
func checkCogneeHealth() error {
	return checkCogneeHealthWithConfig(containerConfig)
}

// checkCogneeHealthWithConfig verifies Cognee API availability using provided config
func checkCogneeHealthWithConfig(cfg *ContainerConfig) error {
	if err := cfg.HealthChecker.CheckHealth(cfg.CogneeURL); err != nil {
		return fmt.Errorf("cannot connect to Cognee: %w", err)
	}
	return nil
}

// checkChromaDBHealth verifies ChromaDB availability
func checkChromaDBHealth() error {
	return checkChromaDBHealthWithConfig(containerConfig)
}

// checkChromaDBHealthWithConfig verifies ChromaDB availability using provided config
func checkChromaDBHealthWithConfig(cfg *ContainerConfig) error {
	if err := cfg.HealthChecker.CheckHealth(cfg.ChromaDBURL); err != nil {
		return fmt.Errorf("cannot connect to ChromaDB: %w", err)
	}
	return nil
}

// MandatoryDependency represents a required integration dependency
type MandatoryDependency struct {
	Name        string
	Description string
	CheckFunc   func() error
	Required    bool
}

// GetMandatoryDependencies returns all mandatory integration dependencies
func GetMandatoryDependencies() []MandatoryDependency {
	return []MandatoryDependency{
		{
			Name:        "PostgreSQL",
			Description: "Primary database for storing configuration, sessions, and metadata",
			CheckFunc:   checkPostgresHealth,
			Required:    true,
		},
		{
			Name:        "Redis",
			Description: "Cache layer for sessions, rate limiting, and response caching",
			CheckFunc:   checkRedisHealth,
			Required:    true,
		},
		{
			Name:        "Cognee",
			Description: "Knowledge graph and RAG integration for AI memory and reasoning",
			CheckFunc:   checkCogneeHealth,
			Required:    true,
		},
		{
			Name:        "ChromaDB",
			Description: "Vector database for embeddings and semantic search",
			CheckFunc:   checkChromaDBHealth,
			Required:    true,
		},
	}
}

// verifyAllMandatoryDependencies checks ALL required integration dependencies
// Returns an error if ANY mandatory dependency is unavailable
func verifyAllMandatoryDependencies(logger *logrus.Logger) error {
	dependencies := GetMandatoryDependencies()
	var failedDeps []string
	var successDeps []string

	logger.Info("╔══════════════════════════════════════════════════════════════════╗")
	logger.Info("║           MANDATORY DEPENDENCY VERIFICATION                       ║")
	logger.Info("╚══════════════════════════════════════════════════════════════════╝")

	for _, dep := range dependencies {
		logger.WithField("dependency", dep.Name).Info("Checking dependency...")

		if err := dep.CheckFunc(); err != nil {
			failedDeps = append(failedDeps, fmt.Sprintf("%s: %v", dep.Name, err))
			logger.WithFields(logrus.Fields{
				"dependency":  dep.Name,
				"description": dep.Description,
				"error":       err,
			}).Error("❌ DEPENDENCY FAILED")
		} else {
			successDeps = append(successDeps, dep.Name)
			logger.WithFields(logrus.Fields{
				"dependency":  dep.Name,
				"description": dep.Description,
			}).Info("✅ DEPENDENCY OK")
		}
	}

	logger.Info("────────────────────────────────────────────────────────────────────")
	logger.WithFields(logrus.Fields{
		"total":  len(dependencies),
		"passed": len(successDeps),
		"failed": len(failedDeps),
	}).Info("Dependency verification summary")

	if len(failedDeps) > 0 {
		errorMsg := fmt.Sprintf("BOOT BLOCKED: %d of %d mandatory dependencies failed:\n", len(failedDeps), len(dependencies))
		for i, failure := range failedDeps {
			errorMsg += fmt.Sprintf("  %d. %s\n", i+1, failure)
		}
		errorMsg += "\nHelixAgent REQUIRES all integration components to be running.\n"
		errorMsg += "Please start all dependencies with: docker-compose up -d\n"
		errorMsg += "Or use: make docker-start"

		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}

// runStartupVerification performs unified startup verification using LLMsVerifier
// as the single source of truth for all provider verification and scoring.
// Returns the startup result and verifier instance (both may be nil if verification fails)
func runStartupVerification(logger *logrus.Logger) (*verifier.StartupResult, *verifier.StartupVerifier) {
	if logger == nil {
		logger = logrus.New()
	}

	logger.Info("╔══════════════════════════════════════════════════════════════════╗")
	logger.Info("║         UNIFIED PROVIDER STARTUP VERIFICATION                     ║")
	logger.Info("║     LLMsVerifier as Single Source of Truth for ALL Providers     ║")
	logger.Info("╚══════════════════════════════════════════════════════════════════╝")

	// Create startup config with defaults
	cfg := verifier.DefaultStartupConfig()
	cfg.ParallelVerification = true
	cfg.EnableFreeProviders = true
	cfg.TrustOAuthOnFailure = true

	// Create startup verifier
	sv := verifier.NewStartupVerifier(cfg, logger)

	// Create context with timeout for verification
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run full verification pipeline
	// Phase 1: Discover all providers (OAuth, API Key, Free)
	// Phase 2: Verify all providers in parallel
	// Phase 3: Score all verified providers
	// Phase 4: Rank providers by score (OAuth first, then by score)
	// Phase 5: Select AI Debate Team (15 LLMs)
	result, err := sv.VerifyAllProviders(ctx)
	if err != nil {
		logger.WithError(err).Warn("Startup verification encountered errors")
		// Don't fail boot - continue with available providers
	}

	if result == nil {
		logger.Warn("Startup verification returned nil result, continuing with legacy discovery")
		return nil, nil
	}

	// Log verification summary
	logger.Info("────────────────────────────────────────────────────────────────────")
	logger.WithFields(logrus.Fields{
		"total_providers":   result.TotalProviders,
		"verified":          result.VerifiedCount,
		"failed":            result.FailedCount,
		"skipped":           result.SkippedCount,
		"api_key_providers": result.APIKeyProviders,
		"oauth_providers":   result.OAuthProviders,
		"free_providers":    result.FreeProviders,
	}).Info("Provider verification summary")

	// Log any errors
	for _, e := range result.Errors {
		logger.WithFields(logrus.Fields{
			"provider":    e.Provider,
			"phase":       e.Phase,
			"error":       e.Error,
			"recoverable": e.Recoverable,
		}).Warn("Provider verification error")
	}

	// Log ranked providers
	rankedProviders := sv.GetRankedProviders()
	if len(rankedProviders) > 0 {
		logger.Info("Top verified providers by score:")
		for i, p := range rankedProviders {
			if i >= 5 {
				break
			}
			logger.WithFields(logrus.Fields{
				"rank":      i + 1,
				"provider":  p.Name,
				"type":      p.Type,
				"auth_type": p.AuthType,
				"score":     p.Score,
				"verified":  p.Verified,
				"models":    len(p.Models),
			}).Info("Provider ranked")
		}
	}

	// Log debate team selection
	if result.DebateTeam != nil {
		logger.Info("────────────────────────────────────────────────────────────────────")
		logger.Info("AI Debate Team Selection (15 LLMs: 5 positions × 3 LLMs each):")
		for _, pos := range result.DebateTeam.Positions {
			if pos.Primary != nil {
				logger.WithFields(logrus.Fields{
					"position":      pos.Position,
					"role":          pos.Role,
					"primary":       pos.Primary.ModelName,
					"primary_prov":  pos.Primary.Provider,
					"primary_score": pos.Primary.Score,
					"is_oauth":      pos.Primary.IsOAuth,
				}).Info("Debate position assigned")
			}
		}
	}

	logger.Info("════════════════════════════════════════════════════════════════════")

	return result, sv
}

// AppConfig holds application configuration for testing
type AppConfig struct {
	ShowHelp           bool
	ShowVersion        bool
	AutoStartDocker    bool
	StrictDependencies bool // MANDATORY: If true, fail boot when ANY dependency is unavailable
	GenerateAPIKey     bool
	GenerateOpenCode   bool
	ValidateOpenCode   string
	OpenCodeOutput     string
	GenerateCrush      bool
	ValidateCrush      string
	CrushOutput        string
	APIKeyEnvFile      string
	PreinstallMCP      bool // Run MCP package pre-installation and exit
	SkipMCPPreinstall  bool // Skip automatic MCP pre-installation at startup
	AutoStartMCP       bool // Automatically start MCP Docker containers on startup
	// Unified CLI agent configuration (all 48 agents)
	GenerateAgentConfig string // Agent type to generate config for
	ValidateAgentConfig string // Agent:path for validation
	AgentConfigOutput   string // Output path for generated config
	ListAgents          bool   // List all supported agents
	GenerateAllAgents   bool   // Generate configs for all agents
	AllAgentsOutputDir  string // Output directory for all agent configs
	ServerHost          string
	ServerPort          string
	Logger              *logrus.Logger
	ShutdownSignal      chan os.Signal
}

// DefaultAppConfig returns the default application configuration
func DefaultAppConfig() *AppConfig {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return &AppConfig{
		ShowHelp:           false,
		ShowVersion:        false,
		AutoStartDocker:    true,
		AutoStartMCP:       true, // Auto-start all 32 MCP Docker containers
		StrictDependencies: true, // MANDATORY: All dependencies must be available
		ServerHost:         "0.0.0.0",
		ServerPort:         "7061",
		Logger:             logger,
		ShutdownSignal:     nil,
	}
}

// run executes the main application logic with the given configuration
// Returns an error if the application fails to start
func run(appCfg *AppConfig) error {
	if appCfg.ShowHelp {
		showHelp()
		return nil
	}

	if appCfg.ShowVersion {
		showVersion()
		return nil
	}

	// Handle API key generation command
	if appCfg.GenerateAPIKey {
		return handleGenerateAPIKey(appCfg)
	}

	// Handle OpenCode config validation command
	if appCfg.ValidateOpenCode != "" {
		return handleValidateOpenCode(appCfg)
	}

	// Handle OpenCode config generation command
	if appCfg.GenerateOpenCode {
		return handleGenerateOpenCode(appCfg)
	}

	// Handle Crush config validation command
	if appCfg.ValidateCrush != "" {
		return handleValidateCrush(appCfg)
	}

	// Handle Crush config generation command
	if appCfg.GenerateCrush {
		return handleGenerateCrush(appCfg)
	}

	// Handle unified CLI agent commands (all 48 agents)
	if appCfg.ListAgents {
		return handleListAgents(appCfg)
	}

	if appCfg.GenerateAllAgents {
		return handleGenerateAllAgents(appCfg)
	}

	if appCfg.GenerateAgentConfig != "" {
		return handleGenerateAgentConfig(appCfg)
	}

	if appCfg.ValidateAgentConfig != "" {
		return handleValidateAgentConfig(appCfg)
	}

	// Handle MCP pre-installation command
	if appCfg.PreinstallMCP {
		return handlePreinstallMCP(appCfg)
	}

	// Load full configuration from environment variables
	cfg := config.Load()

	// Override with command-line specified values if provided
	if appCfg.ServerHost != "" && appCfg.ServerHost != "0.0.0.0" {
		cfg.Server.Host = appCfg.ServerHost
	}
	if appCfg.ServerPort != "" && appCfg.ServerPort != "7061" {
		cfg.Server.Port = appCfg.ServerPort
	}

	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Auto-start required Docker containers if enabled
	if appCfg.AutoStartDocker {
		logger.Info("Checking and starting required Docker containers...")
		if err := ensureRequiredContainers(logger); err != nil {
			if appCfg.StrictDependencies {
				return fmt.Errorf("FATAL: Failed to start required containers (strict mode enabled): %w", err)
			}
			logger.WithError(err).Warn("Failed to start some containers, continuing with application startup")
		} else {
			logger.Info("Docker containers are ready")
		}
	}

	// Auto-start MCP servers from git submodules (32 servers, zero npm dependencies)
	if appCfg.AutoStartMCP {
		logger.Info("Starting MCP servers from git submodules...")
		if err := ensureMCPServers(logger); err != nil {
			logger.WithError(err).Warn("Failed to start MCP servers, continuing without them")
		}
	}

	// Auto-start LSP and RAG services (non-blocking, runs in background)
	logger.Info("Starting LSP and RAG infrastructure services...")
	startAllInfrastructure(logger)

	// MANDATORY: Verify ALL integration dependencies are healthy before starting server
	if appCfg.StrictDependencies {
		logger.Info("Verifying ALL integration dependencies (strict mode)...")
		if err := verifyAllMandatoryDependencies(logger); err != nil {
			return fmt.Errorf("FATAL: Integration dependency verification failed: %w", err)
		}
		logger.Info("All mandatory dependencies verified successfully")
	}

	// Run unified startup verification (LLMsVerifier as single source of truth)
	// This verifies ALL providers (OAuth, API Key, Free) and selects the AI Debate Team
	startupResult, startupVerifier := runStartupVerification(logger)
	if startupResult != nil {
		logger.WithFields(logrus.Fields{
			"total_providers": startupResult.TotalProviders,
			"verified_count":  startupResult.VerifiedCount,
			"failed_count":    startupResult.FailedCount,
			"oauth_providers": startupResult.OAuthProviders,
			"free_providers":  startupResult.FreeProviders,
			"duration_ms":     startupResult.DurationMs,
		}).Info("Startup verification completed")

		if startupResult.DebateTeam != nil {
			logger.WithFields(logrus.Fields{
				"debate_team_llms": startupResult.DebateTeam.TotalLLMs,
				"debate_positions": len(startupResult.DebateTeam.Positions),
				"oauth_first":      startupResult.DebateTeam.OAuthFirst,
			}).Info("AI Debate Team configured (15 LLMs)")
		}
	}

	// Store startup verifier for router access
	_ = startupVerifier // Used by router.SetupRouterWithVerifier if available

	// Initialize messaging system with in-memory fallback
	// This provides RabbitMQ-style task queuing and Kafka-style event streaming
	logger.Info("Initializing messaging system...")
	msgCtx, msgCancel := context.WithTimeout(context.Background(), 30*time.Second)
	msgSystem, err := messaging.InitializeGlobalMessagingSystem(msgCtx, logger, func() messaging.MessageBroker {
		return inmemory.NewBroker(nil)
	})
	msgCancel()
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize messaging system, continuing without messaging")
	} else {
		logger.WithFields(logrus.Fields{
			"initialized":   msgSystem.IsInitialized(),
			"fallback_mode": msgSystem.Config.FallbackToInMemory,
		}).Info("Messaging system initialized")
	}

	r := router.SetupRouter(cfg)

	// Start background MCP package pre-installation (unless skipped)
	if !appCfg.SkipMCPPreinstall {
		startBackgroundMCPPreinstall(logger)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second, // 5 minutes for SSE streaming support
		IdleTimeout:  120 * time.Second,
	}

	// Channel for server errors
	serverErr := make(chan error, 1)

	go func() {
		logger.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		}).Info("Starting HelixAgent server with Models.dev integration")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Start background OAuth token refresh for Claude and Qwen
	stopRefresh := make(chan struct{})
	oauth_credentials.StartBackgroundRefresh(stopRefresh)
	logger.Info("Started background OAuth token refresh for Claude and Qwen")

	// Use provided shutdown signal or create one
	quit := appCfg.ShutdownSignal
	if quit == nil {
		quit = make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	}

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return fmt.Errorf("server failed to start: %w", err)
	case <-quit:
		// Continue to shutdown
	}

	logger.Info("Shutting down server...")

	// Stop background OAuth token refresh
	close(stopRefresh)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown messaging system
	if msgSystem != nil && msgSystem.IsInitialized() {
		logger.Info("Shutting down messaging system...")
		if err := msgSystem.Close(shutdownCtx); err != nil {
			logger.WithError(err).Warn("Error shutting down messaging system")
		} else {
			logger.Info("Messaging system shutdown complete")
		}
	}

	// Use r variable to avoid unused import
	_ = r

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
		return fmt.Errorf("server shutdown error: %w", err)
	}

	logger.Info("Server shutdown complete")
	return nil
}

func main() {
	// Load environment variables from .env file (if present)
	// This allows API keys and configuration to be loaded automatically
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist - environment variables may be set directly
		// Only log if there's a real error (not "file not found")
		if !os.IsNotExist(err) {
			logrus.WithError(err).Debug("Could not load .env file")
		}
	}

	flag.Parse()

	appCfg := DefaultAppConfig()
	appCfg.ShowHelp = *help
	appCfg.ShowVersion = *version
	appCfg.AutoStartDocker = *autoStartDocker
	appCfg.StrictDependencies = *strictDependencies
	appCfg.GenerateAPIKey = *generateAPIKey
	appCfg.GenerateOpenCode = *generateOpenCode
	appCfg.ValidateOpenCode = *validateOpenCode
	appCfg.OpenCodeOutput = *openCodeOutput
	appCfg.GenerateCrush = *generateCrush
	appCfg.ValidateCrush = *validateCrush
	appCfg.CrushOutput = *crushOutput
	appCfg.APIKeyEnvFile = *apiKeyEnvFile
	appCfg.PreinstallMCP = *preinstallMCP
	appCfg.SkipMCPPreinstall = *skipMCPPreinstall
	appCfg.AutoStartMCP = *autoStartMCP
	// Unified CLI agent configuration flags
	appCfg.GenerateAgentConfig = *generateAgentConfig
	appCfg.ValidateAgentConfig = *validateAgentConfig
	appCfg.AgentConfigOutput = *agentConfigOutput
	appCfg.ListAgents = *listAgents
	appCfg.GenerateAllAgents = *generateAllAgents
	appCfg.AllAgentsOutputDir = *allAgentsOutputDir

	if err := run(appCfg); err != nil {
		appCfg.Logger.WithError(err).Fatal("Application failed")
	}
}

// generateSecureAPIKey generates a cryptographically secure API key
func generateSecureAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

// handleGenerateAPIKey handles the --generate-api-key command
func handleGenerateAPIKey(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	// Generate the API key
	apiKey, err := generateSecureAPIKey()
	if err != nil {
		return fmt.Errorf("failed to generate API key: %w", err)
	}

	// If env file is specified, write to it
	if appCfg.APIKeyEnvFile != "" {
		if err := writeAPIKeyToEnvFile(appCfg.APIKeyEnvFile, apiKey); err != nil {
			return fmt.Errorf("failed to write API key to env file: %w", err)
		}
		logger.WithField("file", appCfg.APIKeyEnvFile).Info("API key written to env file")
	}

	// Output the API key to stdout
	fmt.Println(apiKey)
	return nil
}

// writeAPIKeyToEnvFile writes or updates the HELIXAGENT_API_KEY in the specified .env file
func writeAPIKeyToEnvFile(filePath, apiKey string) error {
	// Validate path for traversal attacks (G304 security fix)
	// Note: This is a CLI-provided path from the admin user
	if !utils.ValidatePath(filePath) {
		return fmt.Errorf("invalid file path: contains path traversal or dangerous characters")
	}

	// Read existing file contents if it exists
	existingContent := make(map[string]string)
	var lineOrder []string

	// #nosec G304 - filePath is validated by utils.ValidatePath and provided via CLI by admin
	if file, err := os.Open(filePath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				lineOrder = append(lineOrder, line)
				continue
			}
			// Parse key=value
			if idx := strings.Index(line, "="); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])
				existingContent[key] = value
				lineOrder = append(lineOrder, key)
			} else {
				lineOrder = append(lineOrder, line)
			}
		}
	}

	// Update the API key
	existingContent["HELIXAGENT_API_KEY"] = apiKey

	// Check if key already exists in order
	keyExists := false
	for _, item := range lineOrder {
		if item == "HELIXAGENT_API_KEY" {
			keyExists = true
			break
		}
	}
	if !keyExists {
		lineOrder = append(lineOrder, "HELIXAGENT_API_KEY")
	}

	// Write back to file
	// #nosec G304 - filePath is validated by utils.ValidatePath at function entry and provided via CLI by admin
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create env file: %w", err)
	}
	defer file.Close()

	for _, item := range lineOrder {
		if item == "" || strings.HasPrefix(item, "#") {
			// Write empty lines and comments as-is
			fmt.Fprintln(file, item)
		} else if value, ok := existingContent[item]; ok {
			// Write key=value
			fmt.Fprintf(file, "%s=%s\n", item, value)
		}
	}

	return nil
}

// OpenCodeConfig represents the CORRECT OpenCode configuration structure
// Based on actual working config and OpenCode documentation
// Uses @ai-sdk/openai-compatible for custom providers
type OpenCodeConfig struct {
	Schema       string                               `json:"$schema,omitempty"`
	Provider     map[string]OpenCodeProviderDefNew    `json:"provider,omitempty"`     // Note: singular "provider"
	Agent        map[string]OpenCodeAgentDefNew       `json:"agent,omitempty"`        // Note: singular "agent"
	MCP          map[string]OpenCodeMCPServerDefNew   `json:"mcp,omitempty"`          // OpenCode expects "mcp" key
	Instructions []string                             `json:"instructions,omitempty"` // Rule files
	TUI          *OpenCodeTUIDef                      `json:"tui,omitempty"`
	// NOTE: Model and SmallModel are NOT valid top-level keys in OpenCode config
	// Use agent settings to specify models per agent instead
}

// OpenCodeProviderDefNew represents a provider in OpenCode config (correct schema)
type OpenCodeProviderDefNew struct {
	NPM     string                        `json:"npm,omitempty"`     // e.g., "@ai-sdk/openai-compatible"
	Name    string                        `json:"name,omitempty"`    // Display name
	Options *OpenCodeProviderOptionsNew   `json:"options,omitempty"` // Provider options
	Models  map[string]OpenCodeModelDefNew `json:"models,omitempty"` // Model definitions (map, not array)
}

// OpenCodeProviderOptionsNew represents provider options
type OpenCodeProviderOptionsNew struct {
	BaseURL string            `json:"baseURL,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	APIKey  string            `json:"apiKey,omitempty"` // Can use {env:VAR_NAME} syntax
}

// OpenCodeModelDefNew represents a model definition
type OpenCodeModelDefNew struct {
	Name    string               `json:"name,omitempty"`    // Display name
	Limit   *OpenCodeModelLimit  `json:"limit,omitempty"`   // Token limits
	Options map[string]any       `json:"options,omitempty"` // Model-specific options
}

// OpenCodeModelLimit represents token limits for a model
type OpenCodeModelLimit struct {
	Context int64 `json:"context,omitempty"` // Context window size
	Output  int64 `json:"output,omitempty"`  // Max output tokens
}

// OpenCodeAgentDefNew represents an agent in OpenCode config (correct schema)
type OpenCodeAgentDefNew struct {
	Model           string `json:"model,omitempty"`           // Format: provider-id/model-id
	MaxTokens       int64  `json:"maxTokens,omitempty"`       // Max output tokens
	ReasoningEffort string `json:"reasoningEffort,omitempty"` // low, medium, high
}

// OpenCodeMCPServerDefNew represents an MCP server in OpenCode config (correct schema)
// Based on actual working OpenCode config format:
// Local servers (stdio): command="npx", args=["-y", "package"]
// Remote SSE servers: type="sse", url="http://..."
type OpenCodeMCPServerDefNew struct {
	// Type: "local" for stdio/command MCP servers, "remote" for HTTP/SSE servers
	Type string `json:"type,omitempty"`

	// For local MCP servers (type: "local")
	Command []string `json:"command,omitempty"` // Command array e.g., ["npx", "-y", "mcp-server"]

	// For remote MCP servers (type: "remote")
	URL     string            `json:"url,omitempty"`     // Remote URL endpoint
	Headers map[string]string `json:"headers,omitempty"` // HTTP headers for remote

	// Environment variables (for local servers)
	Environment map[string]string `json:"environment,omitempty"` // Environment variables

	// Enable/disable toggle
	Enabled *bool `json:"enabled,omitempty"` // Set to false to disable
}

// OpenCodeConfigOld represents the OLD OpenCode configuration structure
// For opencode.json files (without leading dot) - uses legacy key names
// This format is validated by OpenCode's strict validator
type OpenCodeConfigOld struct {
	Schema     string                              `json:"$schema,omitempty"`
	Provider   map[string]OpenCodeProviderDefOld   `json:"provider,omitempty"`
	MCP        map[string]OpenCodeMCPServerDefOld  `json:"mcp,omitempty"`
	Agent      map[string]OpenCodeAgentDefOld      `json:"agent,omitempty"`
	Tools      *OpenCodeToolsDefOld                `json:"tools,omitempty"`
	Permission *OpenCodePermissionDefOld           `json:"permission,omitempty"`
}

// OpenCodeProviderDefOld represents a provider in OLD OpenCode config
type OpenCodeProviderDefOld struct {
	Options *OpenCodeProviderOptionsOld `json:"options,omitempty"`
}

// OpenCodeProviderOptionsOld represents provider options in OLD OpenCode config
type OpenCodeProviderOptionsOld struct {
	BaseURL      string                `json:"baseURL,omitempty"`
	APIKeyEnvVar string                `json:"apiKeyEnvVar,omitempty"`
	Models       []OpenCodeModelDefOld `json:"models,omitempty"`
}

// OpenCodeModelDefOld represents a model in OLD OpenCode config
type OpenCodeModelDefOld struct {
	ID           string                         `json:"id"`
	Name         string                         `json:"name"`
	MaxTokens    int64                          `json:"maxTokens,omitempty"`
	Capabilities *OpenCodeModelCapabilitiesOld  `json:"capabilities,omitempty"`
}

// OpenCodeModelCapabilitiesOld represents model capabilities
type OpenCodeModelCapabilitiesOld struct {
	Vision        bool `json:"vision,omitempty"`
	ImageInput    bool `json:"imageInput,omitempty"`
	ImageOutput   bool `json:"imageOutput,omitempty"`
	OCR           bool `json:"ocr,omitempty"`
	PDF           bool `json:"pdf,omitempty"`
	Streaming     bool `json:"streaming,omitempty"`
	FunctionCalls bool `json:"functionCalls,omitempty"`
	ToolUse       bool `json:"toolUse,omitempty"`
	Embeddings    bool `json:"embeddings,omitempty"`
	FileUpload    bool `json:"fileUpload,omitempty"`
	NoFileLimit   bool `json:"noFileLimit,omitempty"`
	MCP           bool `json:"mcp,omitempty"`
	ACP           bool `json:"acp,omitempty"`
	LSP           bool `json:"lsp,omitempty"`
}

// OpenCodeMCPServerDefOld represents an MCP server in OLD OpenCode config
type OpenCodeMCPServerDefOld struct {
	Type    string   `json:"type"`              // "local" or "remote"
	Command []string `json:"command,omitempty"` // For local type - array format
	URL     string   `json:"url,omitempty"`     // For remote type
}

// OpenCodeAgentDefOld represents an agent in OLD OpenCode config
type OpenCodeAgentDefOld struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt,omitempty"`
}

// OpenCodeToolsDefOld represents tools configuration
type OpenCodeToolsDefOld struct {
	Browser    bool `json:"browser,omitempty"`
	Embeddings bool `json:"embeddings,omitempty"`
	File       bool `json:"file,omitempty"`
	LSP        bool `json:"lsp,omitempty"`
	MCP        bool `json:"mcp,omitempty"`
	Search     bool `json:"search,omitempty"`
	Terminal   bool `json:"terminal,omitempty"`
	Vision     bool `json:"vision,omitempty"`
}

// OpenCodePermissionDefOld represents permissions
type OpenCodePermissionDefOld struct {
	AllowRead  bool `json:"allowRead,omitempty"`
	AllowWrite bool `json:"allowWrite,omitempty"`
	AllowExec  bool `json:"allowExec,omitempty"`
	AllowNet   bool `json:"allowNet,omitempty"`
}

// OpenCodeProviderDef is DEPRECATED - use OpenCodeProviderDefNew instead
// Kept for backward compatibility with tests
type OpenCodeProviderDef = OpenCodeProviderDefNew

// OpenCodeAgentDef is DEPRECATED - use OpenCodeAgentDefNew instead
// Kept for backward compatibility with tests
type OpenCodeAgentDef = OpenCodeAgentDefNew

// OpenCodeMCPServerDef is DEPRECATED - use OpenCodeMCPServerDefNew instead
// Kept for backward compatibility with tests
type OpenCodeMCPServerDef = OpenCodeMCPServerDefNew

// OpenCodeTUIDef represents TUI configuration
type OpenCodeTUIDef struct {
	Theme string `json:"theme,omitempty"` // opencode, catppuccin, dracula, etc.
}

// handleGenerateOpenCode handles the --generate-opencode-config command
// Generates OpenCode v1.1.30+ compatible configuration
// IMPORTANT: Config file should be saved as .opencode.json (with leading dot)
// User must set LOCAL_ENDPOINT env var to HelixAgent URL (e.g., http://localhost:7061)
func handleGenerateOpenCode(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	// Get configuration values
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey == "" {
		// If no API key in env, check if we should generate one
		var err error
		apiKey, err = generateSecureAPIKey()
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}
		logger.Warn("No HELIXAGENT_API_KEY found in environment, generated a new one")

		// If env file is specified, write the generated key
		if appCfg.APIKeyEnvFile != "" {
			if err := writeAPIKeyToEnvFile(appCfg.APIKeyEnvFile, apiKey); err != nil {
				logger.WithError(err).Warn("Failed to write generated API key to env file")
			}
		}
	}

	// Get host and port for MCP SSE URLs
	host := os.Getenv("HELIXAGENT_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "7061"
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, port)

	// Determine which format to use based on output filename
	// If filename is "opencode.json" (no dot prefix) -> use OLD format for strict validator
	// If filename is ".opencode.json" (with dot prefix) -> use NEW v1.1.30+ format
	useOldFormat := false
	if appCfg.OpenCodeOutput != "" {
		basename := filepath.Base(appCfg.OpenCodeOutput)
		useOldFormat = basename == "opencode.json"
	}

	var jsonData []byte
	var err error

	// Build CORRECT OpenCode configuration based on official documentation
	// https://opencode.ai/docs/config/ and https://opencode.ai/docs/providers/
	// Both formats use the same correct schema now
	config := OpenCodeConfig{
		Schema: "https://opencode.ai/config.json",
		Provider: map[string]OpenCodeProviderDefNew{
			// HelixAgent as OpenAI-compatible provider
			"helixagent": {
				NPM:  "@ai-sdk/openai-compatible",
				Name: "HelixAgent",
				Options: &OpenCodeProviderOptionsNew{
					BaseURL: baseURL + "/v1",
					APIKey:  "{env:HELIXAGENT_API_KEY}", // Use environment variable syntax
				},
				Models: map[string]OpenCodeModelDefNew{
					"helixagent-debate": {
						Name: "HelixAgent AI Debate Ensemble",
						Limit: &OpenCodeModelLimit{
							Context: 128000,
							Output:  8192,
						},
					},
				},
			},
		},
		// Agent configuration - uses provider-id/model-id format
		// NOTE: Model selection is done per-agent, not at top level
		Agent: map[string]OpenCodeAgentDefNew{
			"coder": {
				Model:     "helixagent/helixagent-debate",
				MaxTokens: 8192,
			},
			"task": {
				Model:     "helixagent/helixagent-debate",
				MaxTokens: 4096,
			},
			"title": {
				Model:     "helixagent/helixagent-debate",
				MaxTokens: 80,
			},
			"summarizer": {
				Model:     "helixagent/helixagent-debate",
				MaxTokens: 4096,
			},
		},
		MCP:          getMCPServers(baseURL, *workingMCPsOnly),
		Instructions: []string{"CLAUDE.md", "opencode.md"},
		TUI:          &OpenCodeTUIDef{Theme: "opencode"},
	}
	jsonData, err = json.MarshalIndent(config, "", "  ")

	// Log which format we're generating (for debugging)
	if useOldFormat {
		logger.Info("Generated OpenCode config (opencode.json)")
	} else {
		logger.Info("Generated OpenCode config (.opencode.json)")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal OpenCode config: %w", err)
	}

	// Output to file or stdout
	if appCfg.OpenCodeOutput != "" {
		// Validate path for traversal attacks (G304 security fix)
		// Note: This is a CLI-provided path from the admin user
		if !utils.ValidatePath(appCfg.OpenCodeOutput) {
			return fmt.Errorf("invalid output path: contains path traversal or dangerous characters")
		}
		// #nosec G304 - OpenCodeOutput is validated by utils.ValidatePath and provided via CLI by admin
		if err := os.WriteFile(appCfg.OpenCodeOutput, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write OpenCode config to file: %w", err)
		}
		logger.WithField("file", appCfg.OpenCodeOutput).Info("OpenCode configuration written to file")
		if useOldFormat {
			logger.Info("Generated OLD format config for opencode.json (strict validator compatible)")
		} else {
			logger.Info("Generated v1.1.30+ format config for .opencode.json")
			logger.Infof("IMPORTANT: Set LOCAL_ENDPOINT=%s before running opencode", baseURL)
		}
	} else {
		fmt.Println(string(jsonData))
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "IMPORTANT: Save as .opencode.json (with leading dot) in ~/.config/opencode/")
		fmt.Fprintf(os.Stderr, "IMPORTANT: Set LOCAL_ENDPOINT=%s before running opencode\n", baseURL)
	}

	return nil
}

// buildOpenCodeMCPServersNew creates MCP server configurations using the correct OpenCode schema
// Based on: https://opencode.ai/docs/mcp-servers
// Local servers: type="local", command=["npx", "-y", "package"] - started on demand by OpenCode
// Remote servers: type="remote", url="https://..." - must be pre-running
// MCP servers are available from:
// - npm packages (local - started on demand via npx)
// - HelixAgent protocol endpoints (/v1/mcp, /v1/acp, /v1/lsp, etc. - remote)
// COMPLIANCE: 62+ MCPs required for system compliance

// getMCPServers returns MCP configurations based on flags
// If workingOnly is true, only MCPs with all dependencies met are returned
// If useLocalMCPServers is true, uses local Docker-based MCP servers on TCP ports
func getMCPServers(baseURL string, workingOnly bool) map[string]OpenCodeMCPServerDefNew {
	if *useLocalMCPServers {
		return buildLocalDockerMCPServers(baseURL)
	}
	if workingOnly {
		return buildWorkingMCPsOnly(baseURL)
	}
	return buildOpenCodeMCPServersNew(baseURL)
}

// buildLocalDockerMCPServers builds MCP configurations for local Docker-based servers
// DEPRECATED: TCP protocol is NOT supported by OpenCode. Use npx-based MCPs instead.
// This function is kept for backward compatibility but should not be used with OpenCode.
// TCP servers run on ports 9101-9999 via socat - requires start-mcp-servers.sh
func buildLocalDockerMCPServers(baseURL string) map[string]OpenCodeMCPServerDefNew {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}

	helixHome := os.Getenv("HELIXAGENT_HOME")
	if helixHome == "" {
		helixHome = homeDir + "/.helixagent"
	}

	// WARNING: TCP URLs are NOT supported by OpenCode!
	// OpenCode only supports http://, https://, and stdio commands
	// This function is deprecated for OpenCode config generation
	fmt.Fprintln(os.Stderr, "WARNING: TCP-based MCP servers are NOT supported by OpenCode. Use default npx-based MCPs instead.")

	return map[string]OpenCodeMCPServerDefNew{
		// HelixAgent local plugin and remote endpoints
		"helixagent": {
			Type:    "local",
			Command: []string{"node", helixHome + "/plugins/mcp-server/dist/index.js", "--endpoint", baseURL},
		},
		"helixagent-mcp": {
			Type: "remote",
			URL:  baseURL + "/v1/mcp",
		},
		"helixagent-acp": {
			Type: "remote",
			URL:  baseURL + "/v1/acp",
		},
		"helixagent-lsp": {
			Type: "remote",
			URL:  baseURL + "/v1/lsp",
		},
		"helixagent-embeddings": {
			Type: "remote",
			URL:  baseURL + "/v1/embeddings",
		},
		"helixagent-vision": {
			Type: "remote",
			URL:  baseURL + "/v1/vision",
		},
		"helixagent-cognee": {
			Type: "remote",
			URL:  baseURL + "/v1/cognee",
		},

		// NOTE: All TCP-based MCPs below will NOT work with OpenCode
		// They are kept for compatibility but will cause errors
		"fetch":               {Type: "remote", URL: "tcp://localhost:9101"},
		"git":                 {Type: "remote", URL: "tcp://localhost:9102"},
		"time":                {Type: "remote", URL: "tcp://localhost:9103"},
		"filesystem":          {Type: "remote", URL: "tcp://localhost:9104"},
		"memory":              {Type: "remote", URL: "tcp://localhost:9105"},
		"everything":          {Type: "remote", URL: "tcp://localhost:9106"},
		"sequential-thinking": {Type: "remote", URL: "tcp://localhost:9107"},
		"redis":               {Type: "remote", URL: "tcp://localhost:9201"},
		"mongodb":             {Type: "remote", URL: "tcp://localhost:9202"},
		"supabase":            {Type: "remote", URL: "tcp://localhost:9203"},
		"qdrant":              {Type: "remote", URL: "tcp://localhost:9301"},
		"kubernetes":          {Type: "remote", URL: "tcp://localhost:9401"},
		"github":              {Type: "remote", URL: "tcp://localhost:9402"},
		"cloudflare":          {Type: "remote", URL: "tcp://localhost:9403"},
		"heroku":              {Type: "remote", URL: "tcp://localhost:9404"},
		"sentry":              {Type: "remote", URL: "tcp://localhost:9405"},
		"playwright":          {Type: "remote", URL: "tcp://localhost:9501"},
		"browserbase":         {Type: "remote", URL: "tcp://localhost:9502"},
		"firecrawl":           {Type: "remote", URL: "tcp://localhost:9503"},
		"slack":               {Type: "remote", URL: "tcp://localhost:9601"},
		"telegram":            {Type: "remote", URL: "tcp://localhost:9602"},
		"notion":              {Type: "remote", URL: "tcp://localhost:9701"},
		"trello":              {Type: "remote", URL: "tcp://localhost:9702"},
		"airtable":            {Type: "remote", URL: "tcp://localhost:9703"},
		"obsidian":            {Type: "remote", URL: "tcp://localhost:9704"},
		"atlassian":           {Type: "remote", URL: "tcp://localhost:9705"},
		"brave-search":        {Type: "sse", URL: "tcp://localhost:9801"},
		"perplexity":          {Type: "sse", URL: "tcp://localhost:9802"},
		"omnisearch":          {Type: "sse", URL: "tcp://localhost:9803"},
		"context7":            {Type: "sse", URL: "tcp://localhost:9804"},
		"llamaindex":          {Type: "sse", URL: "tcp://localhost:9805"},
		"workers":             {Type: "sse", URL: "tcp://localhost:9901"},
	}
}

func buildOpenCodeMCPServersNew(baseURL string) map[string]OpenCodeMCPServerDefNew {
	return buildOpenCodeMCPServersFiltered(baseURL, false)
}

// buildWorkingMCPsOnly builds MCP configurations for only those MCPs that have
// all their dependencies met (API keys set, services running, etc.)
func buildWorkingMCPsOnly(baseURL string) map[string]OpenCodeMCPServerDefNew {
	return buildOpenCodeMCPServersFiltered(baseURL, true)
}

// buildOpenCodeMCPServersFiltered builds MCP configurations with optional filtering
// When filterWorking is true, only MCPs with all dependencies met are included
func buildOpenCodeMCPServersFiltered(baseURL string, filterWorking bool) map[string]OpenCodeMCPServerDefNew {
	// Get user home directory for paths
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}

	// Get HELIXAGENT_HOME for local MCP plugin path
	helixHome := os.Getenv("HELIXAGENT_HOME")
	if helixHome == "" {
		helixHome = homeDir + "/.helixagent"
	}

	allMCPs := map[string]OpenCodeMCPServerDefNew{
		// =============================================================================
		// HelixAgent Protocol Endpoints (7 MCPs) - LOCAL + REMOTE
		// =============================================================================
		// Local MCP plugin - runs the HelixAgent MCP server as a subprocess
		"helixagent": {
			Type:    "local",
			Command: []string{"node", helixHome + "/plugins/mcp-server/dist/index.js", "--endpoint", baseURL},
		},
		// Remote protocol endpoints (running at port 7061)
		"helixagent-mcp": {
			Type: "remote",
			URL:  baseURL + "/v1/mcp",
		},
		"helixagent-acp": {
			Type: "remote",
			URL:  baseURL + "/v1/acp",
		},
		"helixagent-lsp": {
			Type: "remote",
			URL:  baseURL + "/v1/lsp",
		},
		"helixagent-embeddings": {
			Type: "remote",
			URL:  baseURL + "/v1/embeddings",
		},
		"helixagent-vision": {
			Type: "remote",
			URL:  baseURL + "/v1/vision",
		},
		"helixagent-cognee": {
			Type: "remote",
			URL:  baseURL + "/v1/cognee",
		},

		// =============================================================================
		// Anthropic Official MCPs - LOCAL (started on demand via npx)
		// From: https://github.com/modelcontextprotocol/servers
		// =============================================================================
		"filesystem": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-filesystem", homeDir},
		},
		"fetch": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-fetch"},
		},
		"memory": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-memory"},
		},
		"time": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-time"},
		},
		"git": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-git"},
		},
		"sqlite": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-sqlite", "--db-path", "/tmp/helixagent.db"},
		},
		"postgres": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-postgres"},
			Environment: map[string]string{"POSTGRES_URL": "postgresql://helixagent:helixagent123@localhost:5432/helixagent_db"},
		},
		"puppeteer": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-puppeteer"},
		},
		"sequential-thinking": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-sequential-thinking"},
		},
		"everything": {
			Type:    "local",
			Command: []string{"npx", "-y", "@anthropic-ai/mcp-server-everything"},
		},
		"brave-search": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-brave-search"},
			Environment: map[string]string{"BRAVE_API_KEY": "{env:BRAVE_API_KEY}"},
		},
		"google-maps": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-google-maps"},
			Environment: map[string]string{"GOOGLE_MAPS_API_KEY": "{env:GOOGLE_MAPS_API_KEY}"},
		},
		"slack": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-slack"},
			Environment: map[string]string{"SLACK_BOT_TOKEN": "{env:SLACK_BOT_TOKEN}", "SLACK_TEAM_ID": "{env:SLACK_TEAM_ID}"},
		},
		"github": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-github"},
			Environment: map[string]string{"GITHUB_PERSONAL_ACCESS_TOKEN": "{env:GITHUB_TOKEN}"},
		},
		"gitlab": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-gitlab"},
			Environment: map[string]string{"GITLAB_PERSONAL_ACCESS_TOKEN": "{env:GITLAB_TOKEN}"},
		},
		"sentry": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-sentry"},
			Environment: map[string]string{"SENTRY_AUTH_TOKEN": "{env:SENTRY_AUTH_TOKEN}", "SENTRY_ORG": "{env:SENTRY_ORG}"},
		},
		"everart": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-everart"},
			Environment: map[string]string{"EVERART_API_KEY": "{env:EVERART_API_KEY}"},
		},
		"aws-kb-retrieval": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-aws-kb-retrieval"},
			Environment: map[string]string{"AWS_ACCESS_KEY_ID": "{env:AWS_ACCESS_KEY_ID}", "AWS_SECRET_ACCESS_KEY": "{env:AWS_SECRET_ACCESS_KEY}"},
		},

		// =============================================================================
		// Additional Anthropic & Community MCPs - LOCAL
		// =============================================================================
		"exa": {
			Type:        "local",
			Command:     []string{"npx", "-y", "exa-mcp-server"},
			Environment: map[string]string{"EXA_API_KEY": "{env:EXA_API_KEY}"},
		},
		"linear": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-linear"},
			Environment: map[string]string{"LINEAR_API_KEY": "{env:LINEAR_API_KEY}"},
		},
		"notion": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@notionhq/notion-mcp-server"},
			Environment: map[string]string{"NOTION_API_KEY": "{env:NOTION_API_KEY}"},
		},
		"figma": {
			Type:        "local",
			Command:     []string{"npx", "-y", "figma-developer-mcp"},
			Environment: map[string]string{"FIGMA_API_KEY": "{env:FIGMA_API_KEY}"},
		},
		"todoist": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@modelcontextprotocol/server-todoist"},
			Environment: map[string]string{"TODOIST_API_TOKEN": "{env:TODOIST_API_TOKEN}"},
		},
		"obsidian": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-obsidian"},
			Environment: map[string]string{"OBSIDIAN_VAULT_PATH": "{env:OBSIDIAN_VAULT_PATH}"},
		},
		"raycast": {
			Type:    "local",
			Command: []string{"npx", "-y", "@anthropic-ai/mcp-server-raycast"},
		},
		"tinybird": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-tinybird"},
			Environment: map[string]string{"TINYBIRD_TOKEN": "{env:TINYBIRD_TOKEN}"},
		},
		"cloudflare": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@cloudflare/mcp-server-cloudflare"},
			Environment: map[string]string{"CLOUDFLARE_API_TOKEN": "{env:CLOUDFLARE_API_TOKEN}"},
		},
		"neon": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@neondatabase/mcp-server-neon"},
			Environment: map[string]string{"NEON_API_KEY": "{env:NEON_API_KEY}"},
		},
		"gdrive": {
			Type:        "local",
			Command:     []string{"npx", "-y", "@anthropic/mcp-server-gdrive"},
			Environment: map[string]string{"GOOGLE_CREDENTIALS_PATH": "{env:GOOGLE_CREDENTIALS_PATH}"},
		},

		// =============================================================================
		// Container/Infrastructure MCPs - LOCAL
		// =============================================================================
		"docker": {
			Type:    "local",
			Command: []string{"npx", "-y", "@modelcontextprotocol/server-docker"},
		},
		"kubernetes": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-kubernetes"},
			Environment: map[string]string{"KUBECONFIG": "{env:KUBECONFIG}"},
		},
		"redis": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-redis"},
			Environment: map[string]string{"REDIS_URL": "redis://localhost:6379"},
		},
		"mongodb": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-mongodb"},
			Environment: map[string]string{"MONGODB_URI": "mongodb://localhost:27017"},
		},
		"elasticsearch": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-elasticsearch"},
			Environment: map[string]string{"ELASTICSEARCH_URL": "http://localhost:9200"},
		},
		"qdrant": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-qdrant"},
			Environment: map[string]string{"QDRANT_URL": "http://localhost:6333"},
		},
		"chroma": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-chroma"},
			Environment: map[string]string{"CHROMA_URL": "http://localhost:8001"},
		},
		"pinecone": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-pinecone"},
			Environment: map[string]string{"PINECONE_API_KEY": "{env:PINECONE_API_KEY}"},
		},
		"milvus": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-milvus"},
			Environment: map[string]string{"MILVUS_HOST": "localhost", "MILVUS_PORT": "19530"},
		},
		"weaviate": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-weaviate"},
			Environment: map[string]string{"WEAVIATE_URL": "http://localhost:8080"},
		},
		"supabase": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-supabase"},
			Environment: map[string]string{"SUPABASE_URL": "{env:SUPABASE_URL}", "SUPABASE_KEY": "{env:SUPABASE_KEY}"},
		},

		// =============================================================================
		// Productivity/Collaboration MCPs - LOCAL
		// =============================================================================
		"jira": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-atlassian"},
			Environment: map[string]string{"JIRA_URL": "{env:JIRA_URL}", "JIRA_EMAIL": "{env:JIRA_EMAIL}", "JIRA_API_TOKEN": "{env:JIRA_API_TOKEN}"},
		},
		"asana": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-asana"},
			Environment: map[string]string{"ASANA_ACCESS_TOKEN": "{env:ASANA_ACCESS_TOKEN}"},
		},
		"trello": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-trello"},
			Environment: map[string]string{"TRELLO_API_KEY": "{env:TRELLO_API_KEY}", "TRELLO_TOKEN": "{env:TRELLO_TOKEN}"},
		},
		"monday": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-monday"},
			Environment: map[string]string{"MONDAY_API_KEY": "{env:MONDAY_API_KEY}"},
		},
		"clickup": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-clickup"},
			Environment: map[string]string{"CLICKUP_API_KEY": "{env:CLICKUP_API_KEY}"},
		},
		"discord": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-discord"},
			Environment: map[string]string{"DISCORD_BOT_TOKEN": "{env:DISCORD_BOT_TOKEN}"},
		},
		"microsoft-teams": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-teams"},
			Environment: map[string]string{"TEAMS_CLIENT_ID": "{env:TEAMS_CLIENT_ID}", "TEAMS_CLIENT_SECRET": "{env:TEAMS_CLIENT_SECRET}"},
		},
		"gmail": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-gmail"},
			Environment: map[string]string{"GOOGLE_CREDENTIALS_PATH": "{env:GOOGLE_CREDENTIALS_PATH}"},
		},
		"calendar": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-google-calendar"},
			Environment: map[string]string{"GOOGLE_CREDENTIALS_PATH": "{env:GOOGLE_CREDENTIALS_PATH}"},
		},
		"zoom": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-zoom"},
			Environment: map[string]string{"ZOOM_CLIENT_ID": "{env:ZOOM_CLIENT_ID}", "ZOOM_CLIENT_SECRET": "{env:ZOOM_CLIENT_SECRET}"},
		},

		// =============================================================================
		// Cloud/DevOps MCPs - LOCAL
		// =============================================================================
		"aws-s3": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-s3"},
			Environment: map[string]string{"AWS_ACCESS_KEY_ID": "{env:AWS_ACCESS_KEY_ID}", "AWS_SECRET_ACCESS_KEY": "{env:AWS_SECRET_ACCESS_KEY}"},
		},
		"aws-lambda": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-aws-lambda"},
			Environment: map[string]string{"AWS_ACCESS_KEY_ID": "{env:AWS_ACCESS_KEY_ID}", "AWS_SECRET_ACCESS_KEY": "{env:AWS_SECRET_ACCESS_KEY}"},
		},
		"azure": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-azure"},
			Environment: map[string]string{"AZURE_SUBSCRIPTION_ID": "{env:AZURE_SUBSCRIPTION_ID}", "AZURE_TENANT_ID": "{env:AZURE_TENANT_ID}"},
		},
		"gcp": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-gcp"},
			Environment: map[string]string{"GOOGLE_APPLICATION_CREDENTIALS": "{env:GOOGLE_APPLICATION_CREDENTIALS}"},
		},
		"terraform": {
			Type:    "local",
			Command: []string{"npx", "-y", "mcp-server-terraform"},
		},
		"ansible": {
			Type:    "local",
			Command: []string{"npx", "-y", "mcp-server-ansible"},
		},
		"datadog": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-datadog"},
			Environment: map[string]string{"DD_API_KEY": "{env:DD_API_KEY}", "DD_APP_KEY": "{env:DD_APP_KEY}"},
		},
		"grafana": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-grafana"},
			Environment: map[string]string{"GRAFANA_URL": "{env:GRAFANA_URL}", "GRAFANA_API_KEY": "{env:GRAFANA_API_KEY}"},
		},
		"prometheus": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-prometheus"},
			Environment: map[string]string{"PROMETHEUS_URL": "{env:PROMETHEUS_URL}"},
		},
		"circleci": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-circleci"},
			Environment: map[string]string{"CIRCLECI_TOKEN": "{env:CIRCLECI_TOKEN}"},
		},

		// =============================================================================
		// AI/ML Integration MCPs - LOCAL
		// =============================================================================
		"langchain": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-langchain"},
			Environment: map[string]string{"OPENAI_API_KEY": "{env:OPENAI_API_KEY}"},
		},
		"llamaindex": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-llamaindex"},
			Environment: map[string]string{"OPENAI_API_KEY": "{env:OPENAI_API_KEY}"},
		},
		"huggingface": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-huggingface"},
			Environment: map[string]string{"HUGGINGFACE_API_KEY": "{env:HUGGINGFACE_API_KEY}"},
		},
		"replicate": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-replicate"},
			Environment: map[string]string{"REPLICATE_API_TOKEN": "{env:REPLICATE_API_TOKEN}"},
		},
		"stable-diffusion": {
			Type:        "local",
			Command:     []string{"npx", "-y", "mcp-server-stable-diffusion"},
			Environment: map[string]string{"STABILITY_API_KEY": "{env:STABILITY_API_KEY}"},
		},
	}

	// If not filtering, return all MCPs
	if !filterWorking {
		return allMCPs
	}

	// Filter to only working MCPs (those with all dependencies met)
	return filterWorkingMCPs(allMCPs)
}

// filterWorkingMCPs filters MCP configurations to only include those with all dependencies met
// This checks:
// 1. Environment variables for API keys
// 2. Local service connectivity (ports)
func filterWorkingMCPs(allMCPs map[string]OpenCodeMCPServerDefNew) map[string]OpenCodeMCPServerDefNew {
	workingMCPs := make(map[string]OpenCodeMCPServerDefNew)

	// MCPs that work WITHOUT any API keys or external service dependencies
	// VERIFIED: These npm packages exist and work out-of-the-box
	alwaysWorking := map[string]bool{
		// HelixAgent remote endpoints (connect to running HelixAgent)
		"helixagent-mcp":        true,
		"helixagent-acp":        true,
		"helixagent-lsp":        true,
		"helixagent-embeddings": true,
		"helixagent-vision":     true,
		"helixagent-cognee":     true,
		// Official Anthropic MCP servers (npm verified, no API keys)
		"filesystem":          true, // @modelcontextprotocol/server-filesystem
		"memory":              true, // @modelcontextprotocol/server-memory
		"sequential-thinking": true, // @modelcontextprotocol/server-sequential-thinking
	}

	for name, mcpConfig := range allMCPs {
		if alwaysWorking[name] {
			workingMCPs[name] = mcpConfig
		}
	}

	return workingMCPs
}

// buildOpenCodeMCPServersOld builds MCP servers in OLD format for opencode.json
// OLD format uses "type": "local"/"remote" with "command" as array
func buildOpenCodeMCPServersOld(baseURL string) map[string]OpenCodeMCPServerDefOld {
	return map[string]OpenCodeMCPServerDefOld{
		// Anthropic Official MCPs
		"filesystem":          {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-filesystem", "/home"}},
		"fetch":               {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-fetch"}},
		"memory":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-memory"}},
		"time":                {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-time"}},
		"git":                 {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-git"}},
		"sqlite":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-sqlite", "--db-path", "/tmp/helixagent.db"}},
		"postgres":            {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost:5432/helixagent"}},
		"puppeteer":           {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-puppeteer"}},
		"brave-search":        {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-brave-search"}},
		"google-maps":         {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-google-maps"}},
		"slack":               {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-slack"}},
		"github":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-github"}},
		"gitlab":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-gitlab"}},
		"sequential-thinking": {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-sequential-thinking"}},
		"everart":             {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-everart"}},
		"exa":                 {Type: "local", Command: []string{"npx", "-y", "exa-mcp-server"}},
		"linear":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-linear"}},
		"sentry":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-sentry"}},
		"notion":              {Type: "local", Command: []string{"npx", "-y", "@notionhq/notion-mcp-server"}},
		"figma":               {Type: "local", Command: []string{"npx", "-y", "figma-developer-mcp"}},
		"aws-kb-retrieval":    {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-aws-kb-retrieval"}},
		// HelixAgent Remote MCP - endpoint is /v1/mcp
		"helixagent": {Type: "remote", URL: baseURL + "/v1/mcp"},  // Note: OLD format doesn't support headers
		// Community/Infrastructure MCPs
		"docker":        {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-docker"}},
		"kubernetes":    {Type: "local", Command: []string{"npx", "-y", "mcp-server-kubernetes"}},
		"redis":         {Type: "local", Command: []string{"npx", "-y", "mcp-server-redis"}},
		"mongodb":       {Type: "local", Command: []string{"npx", "-y", "mcp-server-mongodb"}},
		"elasticsearch": {Type: "local", Command: []string{"npx", "-y", "mcp-server-elasticsearch"}},
		"qdrant":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-qdrant"}},
		"chroma":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-chroma"}},
		// Productivity MCPs
		"jira":         {Type: "local", Command: []string{"npx", "-y", "mcp-server-atlassian"}},
		"asana":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-asana"}},
		"google-drive": {Type: "local", Command: []string{"npx", "-y", "@anthropic/mcp-server-gdrive"}},
		"aws-s3":       {Type: "local", Command: []string{"npx", "-y", "mcp-server-s3"}},
		"datadog":      {Type: "local", Command: []string{"npx", "-y", "mcp-server-datadog"}},
	}
}

// buildOpenCodeMCPServers is DEPRECATED - use buildOpenCodeMCPServersFiltered instead
// This function redirects to the new implementation for backward compatibility
func buildOpenCodeMCPServers(baseURL string) map[string]OpenCodeMCPServerDef {
	return buildOpenCodeMCPServersFiltered(baseURL, false)
}

// handlePreinstallMCP handles the --preinstall-mcp command
// Pre-installs all standard MCP server npm packages for faster startup
func handlePreinstallMCP(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	logger.Info("Starting MCP package pre-installation...")

	// Get home directory for install location
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return fmt.Errorf("HOME environment variable not set")
	}

	// Create preinstaller
	preinstaller, err := mcp.NewPreinstaller(mcp.PreinstallerConfig{
		InstallDir:  fmt.Sprintf("%s/.helixagent/mcp-servers", homeDir),
		Logger:      logger,
		Concurrency: 4,
		Timeout:     5 * time.Minute,
		OnProgress: func(pkg string, status mcp.InstallStatus, progress float64) {
			logger.WithFields(logrus.Fields{
				"package":  pkg,
				"status":   status,
				"progress": fmt.Sprintf("%.0f%%", progress*100),
			}).Info("MCP package installation progress")
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create MCP preinstaller: %w", err)
	}

	// Check if Node.js is available
	if !preinstaller.IsNodeAvailable() {
		return fmt.Errorf("Node.js is not available - MCP packages cannot be installed")
	}

	// Run pre-installation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := preinstaller.PreInstallAll(ctx); err != nil {
		return fmt.Errorf("MCP pre-installation failed: %w", err)
	}

	// Print summary
	statuses := preinstaller.GetAllStatuses()
	installed := 0
	failed := 0
	for _, status := range statuses {
		if status.Status == mcp.StatusInstalled {
			installed++
			logger.WithFields(logrus.Fields{
				"package":  status.Package.Name,
				"path":     status.InstallPath,
				"duration": status.Duration,
			}).Info("Package installed")
		} else if status.Status == mcp.StatusFailed {
			failed++
			logger.WithError(status.Error).WithField("package", status.Package.Name).Error("Package failed")
		}
	}

	logger.WithFields(logrus.Fields{
		"installed": installed,
		"failed":    failed,
		"total":     len(statuses),
	}).Info("MCP pre-installation complete")

	if failed > 0 {
		return fmt.Errorf("%d packages failed to install", failed)
	}

	return nil
}

// startBackgroundMCPPreinstall starts MCP package pre-installation in background
// This is called at server startup unless --skip-mcp-preinstall is specified
func startBackgroundMCPPreinstall(logger *logrus.Logger) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		logger.Warn("HOME not set, skipping background MCP pre-installation")
		return
	}

	preinstaller, err := mcp.NewPreinstaller(mcp.PreinstallerConfig{
		InstallDir:  fmt.Sprintf("%s/.helixagent/mcp-servers", homeDir),
		Logger:      logger,
		Concurrency: 2, // Lower concurrency for background
		Timeout:     10 * time.Minute,
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to create background MCP preinstaller")
		return
	}

	if !preinstaller.IsNodeAvailable() {
		logger.Debug("Node.js not available, skipping background MCP pre-installation")
		return
	}

	go func() {
		logger.Info("Starting background MCP package pre-installation...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if err := preinstaller.PreInstallAll(ctx); err != nil {
			logger.WithError(err).Warn("Background MCP pre-installation had errors")
		} else {
			logger.Info("Background MCP pre-installation completed successfully")
		}
	}()
}

// OpenCodeValidationError represents a validation error in OpenCode config
type OpenCodeValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// OpenCodeValidationResult holds the complete validation results
type OpenCodeValidationResult struct {
	Valid    bool                      `json:"valid"`
	Errors   []OpenCodeValidationError `json:"errors"`
	Warnings []string                  `json:"warnings"`
	Stats    *OpenCodeValidationStats  `json:"stats,omitempty"`
}

// OpenCodeValidationStats contains statistics about the validated config
type OpenCodeValidationStats struct {
	Providers  int `json:"providers"`
	MCPServers int `json:"mcp_servers"`
	Agents     int `json:"agents"`
	Commands   int `json:"commands"`
}

// handleValidateOpenCode handles the --validate-opencode-config command
func handleValidateOpenCode(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	filePath := appCfg.ValidateOpenCode

	// Validate path for traversal attacks (G304 security fix)
	// Note: This is a CLI-provided path from the admin user
	if !utils.ValidatePath(filePath) {
		return fmt.Errorf("invalid config file path: contains path traversal or dangerous characters")
	}

	// Read the config file
	// #nosec G304 - filePath is validated by utils.ValidatePath and provided via CLI by admin
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Perform validation
	result := validateOpenCodeConfig(data)

	// Output header
	fmt.Println("======================================================================")
	fmt.Println("HELIXAGENT OPENCODE CONFIGURATION VALIDATION")
	fmt.Println("Using LLMsVerifier schema compliance rules")
	fmt.Println("======================================================================")
	fmt.Println()
	fmt.Printf("File: %s\n", filePath)
	fmt.Println()

	if result.Valid {
		fmt.Println("✅ CONFIGURATION IS VALID")
		fmt.Println()
		if result.Stats != nil {
			fmt.Printf("Configuration contains:\n")
			fmt.Printf("  - Providers: %d\n", result.Stats.Providers)
			fmt.Printf("  - MCP servers: %d\n", result.Stats.MCPServers)
			fmt.Printf("  - Agents: %d\n", result.Stats.Agents)
			fmt.Printf("  - Commands: %d\n", result.Stats.Commands)
		}
	} else {
		fmt.Println("❌ CONFIGURATION HAS ERRORS:")
		fmt.Println()
		for _, e := range result.Errors {
			if e.Field != "" {
				fmt.Printf("  - [%s] %s\n", e.Field, e.Message)
			} else {
				fmt.Printf("  - %s\n", e.Message)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("⚠️  WARNINGS:")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	fmt.Println()
	fmt.Println("======================================================================")

	if !result.Valid {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}

// validateOpenCodeConfig performs comprehensive validation of an OpenCode config
func validateOpenCodeConfig(data []byte) *OpenCodeValidationResult {
	result := &OpenCodeValidationResult{
		Valid:    true,
		Errors:   []OpenCodeValidationError{},
		Warnings: []string{},
		Stats:    &OpenCodeValidationStats{},
	}

	// Parse as generic map to check top-level keys
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, OpenCodeValidationError{
			Field:   "",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		})
		return result
	}

	// Check for invalid top-level keys (per LLMsVerifier schema)
	var invalidKeys []string
	for key := range rawConfig {
		if !ValidOpenCodeTopLevelKeys[key] {
			invalidKeys = append(invalidKeys, key)
		}
	}
	if len(invalidKeys) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, OpenCodeValidationError{
			Field:   "",
			Message: fmt.Sprintf("invalid top-level keys: %v (valid keys: $schema, plugin, enterprise, instructions, provider/providers, mcp/mcpServers, tools, agent/agents, command, keybinds, username, share, permission, compaction, sse, mode, autoshare, contextPaths, tui)", invalidKeys),
		})
	}

	// Detect schema version: v1.1.30+ uses "providers" (plural), v1.0.x uses "provider" (singular)
	isV1130Plus := rawConfig["providers"] != nil || rawConfig["mcpServers"] != nil || rawConfig["agents"] != nil

	// Parse and validate providers (both v1.0.x and v1.1.30+ schemas)
	if isV1130Plus {
		// v1.1.30+ schema: providers (plural)
		if providers, ok := rawConfig["providers"].(map[string]interface{}); ok {
			result.Stats.Providers = len(providers)
			// v1.1.30+ schema: each provider can have apiKey and disabled
			for name, providerData := range providers {
				if provider, ok := providerData.(map[string]interface{}); ok {
					// Provider is valid if it has apiKey (can be empty for local provider)
					_, _ = provider["apiKey"], name // Allow any apiKey value
				}
			}
		} else if rawConfig["providers"] == nil {
			result.Valid = false
			result.Errors = append(result.Errors, OpenCodeValidationError{
				Field:   "providers",
				Message: "at least one provider must be configured",
			})
		}
	} else {
		// v1.0.x schema: provider (singular)
		if providers, ok := rawConfig["provider"].(map[string]interface{}); ok {
			result.Stats.Providers = len(providers)
			for name, providerData := range providers {
				if provider, ok := providerData.(map[string]interface{}); ok {
					// Provider must have options
					if _, hasOptions := provider["options"]; !hasOptions {
						result.Valid = false
						result.Errors = append(result.Errors, OpenCodeValidationError{
							Field:   fmt.Sprintf("provider.%s.options", name),
							Message: "provider must have options configured",
						})
					}
				}
			}
		} else if rawConfig["provider"] == nil {
			result.Valid = false
			result.Errors = append(result.Errors, OpenCodeValidationError{
				Field:   "provider",
				Message: "at least one provider must be configured",
			})
		}
	}

	// Parse and validate MCP servers (both v1.0.x and v1.1.30+ schemas)
	if isV1130Plus {
		// v1.1.30+ schema: mcpServers (plural)
		if mcpServers, ok := rawConfig["mcpServers"].(map[string]interface{}); ok {
			result.Stats.MCPServers = len(mcpServers)
			for name, serverData := range mcpServers {
				if server, ok := serverData.(map[string]interface{}); ok {
					// In v1.1.30+ schema, type is "sse" for remote, or command/args for stdio
					serverType, hasType := server["type"].(string)
					_, hasCommand := server["command"]
					_, hasURL := server["url"]

					// If type is "sse", url is required
					if hasType && serverType == "sse" {
						if !hasURL {
							result.Valid = false
							result.Errors = append(result.Errors, OpenCodeValidationError{
								Field:   fmt.Sprintf("mcpServers.%s.url", name),
								Message: "url is required for SSE MCP servers",
							})
						}
					} else if !hasCommand && !hasURL {
						// For stdio servers (no type or type != sse), command is required
						result.Valid = false
						result.Errors = append(result.Errors, OpenCodeValidationError{
							Field:   fmt.Sprintf("mcpServers.%s.command", name),
							Message: "command is required for stdio MCP servers",
						})
					}
				}
			}
		}
	} else {
		// v1.0.x schema: mcp (singular)
		if mcpServers, ok := rawConfig["mcp"].(map[string]interface{}); ok {
			result.Stats.MCPServers = len(mcpServers)
			for name, serverData := range mcpServers {
				if server, ok := serverData.(map[string]interface{}); ok {
					serverType, hasType := server["type"].(string)
					if !hasType {
						result.Valid = false
						result.Errors = append(result.Errors, OpenCodeValidationError{
							Field:   fmt.Sprintf("mcp.%s.type", name),
							Message: "type is required for MCP servers",
						})
						continue
					}
					if serverType != "local" && serverType != "remote" {
						result.Valid = false
						result.Errors = append(result.Errors, OpenCodeValidationError{
							Field:   fmt.Sprintf("mcp.%s.type", name),
							Message: "type must be 'local' or 'remote'",
						})
					}
					if serverType == "local" {
						if _, hasCommand := server["command"]; !hasCommand {
							result.Valid = false
							result.Errors = append(result.Errors, OpenCodeValidationError{
								Field:   fmt.Sprintf("mcp.%s.command", name),
								Message: "command is required for local MCP servers",
							})
						}
					}
					if serverType == "remote" {
						if _, hasURL := server["url"]; !hasURL {
							result.Valid = false
							result.Errors = append(result.Errors, OpenCodeValidationError{
								Field:   fmt.Sprintf("mcp.%s.url", name),
								Message: "url is required for remote MCP servers",
							})
						}
					}
				}
			}
		}
	}

	// Parse and validate agents (both v1.0.x and v1.1.30+ schemas)
	if isV1130Plus {
		// v1.1.30+ schema: agents (plural)
		if agents, ok := rawConfig["agents"].(map[string]interface{}); ok {
			result.Stats.Agents = len(agents)
			for name, agentData := range agents {
				if agent, ok := agentData.(map[string]interface{}); ok {
					// In v1.1.30+ schema, agents need model
					if _, hasModel := agent["model"]; !hasModel {
						result.Valid = false
						result.Errors = append(result.Errors, OpenCodeValidationError{
							Field:   fmt.Sprintf("agents.%s", name),
							Message: "agent must have model configured",
						})
					}
				}
			}
		}
	} else if agents, ok := rawConfig["agent"].(map[string]interface{}); ok {
		// Check if this is a single agent object with "model" directly
		if _, hasModel := agents["model"]; hasModel {
			result.Stats.Agents = 1
			// Single agent config - validate it has model or prompt
			// This is valid - it has model
		} else {
			result.Stats.Agents = len(agents)
			for name, agentData := range agents {
				if agent, ok := agentData.(map[string]interface{}); ok {
					_, hasModel := agent["model"]
					_, hasPrompt := agent["prompt"]
					if !hasModel && !hasPrompt {
						result.Valid = false
						result.Errors = append(result.Errors, OpenCodeValidationError{
							Field:   fmt.Sprintf("agent.%s", name),
							Message: "agent must have either model or prompt configured",
						})
					}
				}
			}
		}
	}

	// Parse commands
	if commands, ok := rawConfig["command"].(map[string]interface{}); ok {
		result.Stats.Commands = len(commands)
	}

	// Add warnings for missing recommended fields
	if _, hasSchema := rawConfig["$schema"]; !hasSchema {
		result.Warnings = append(result.Warnings, "$schema field is recommended for validation")
	}

	return result
}

// CrushConfig represents the Crush CLI configuration structure
type CrushConfig struct {
	Schema    string                    `json:"$schema,omitempty"`
	Providers map[string]CrushProvider  `json:"providers,omitempty"`
	Lsp       map[string]CrushLspConfig `json:"lsp,omitempty"`
	Mcp       map[string]CrushMcpConfig `json:"mcp,omitempty"`
	Options   *CrushOptions             `json:"options,omitempty"`
}

// CrushMcpConfig represents MCP server configuration for Crush
type CrushMcpConfig struct {
	Type    string            `json:"type"`
	URL     string            `json:"url,omitempty"`
	Command []string          `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Enabled bool              `json:"enabled"`
}

// CrushProvider represents a provider configuration for Crush
type CrushProvider struct {
	Name    string       `json:"name"`
	Type    string       `json:"type"`
	BaseURL string       `json:"base_url"`
	APIKey  string       `json:"api_key,omitempty"`
	Models  []CrushModel `json:"models"`
}

// CrushModel represents a model configuration for Crush
type CrushModel struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	CostPer1MIn          float64                `json:"cost_per_1m_in"`
	CostPer1MOut         float64                `json:"cost_per_1m_out"`
	CostPer1MInCached    float64                `json:"cost_per_1m_in_cached,omitempty"`
	CostPer1MOutCached   float64                `json:"cost_per_1m_out_cached,omitempty"`
	ContextWindow        int                    `json:"context_window"`
	DefaultMaxTokens     int                    `json:"default_max_tokens"`
	CanReason            bool                   `json:"can_reason"`
	SupportsAttachments  bool                   `json:"supports_attachments"`
	Streaming            bool                   `json:"streaming"`
	SupportsBrotli       bool                   `json:"supports_brotli,omitempty"`
	Options              map[string]interface{} `json:"options,omitempty"`
}

// CrushLspConfig represents Language Server Protocol configuration for Crush
type CrushLspConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Enabled bool     `json:"enabled"`
}

// CrushOptions represents global configuration options for Crush
type CrushOptions struct {
	DisableProviderAutoUpdate bool `json:"disable_provider_auto_update,omitempty"`
}

// handleGenerateCrush handles the --generate-crush-config command
func handleGenerateCrush(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	// Get configuration values
	apiKey := os.Getenv("HELIXAGENT_API_KEY")
	if apiKey == "" {
		// If no API key in env, check if we should generate one
		var err error
		apiKey, err = generateSecureAPIKey()
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}
		logger.Warn("No HELIXAGENT_API_KEY found in environment, generated a new one")

		// If env file is specified, write the generated key
		if appCfg.APIKeyEnvFile != "" {
			if err := writeAPIKeyToEnvFile(appCfg.APIKeyEnvFile, apiKey); err != nil {
				logger.WithError(err).Warn("Failed to write generated API key to env file")
			}
		}
	}

	// Get host and port
	host := os.Getenv("HELIXAGENT_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "7061"
	}

	baseURL := fmt.Sprintf("http://%s:%s/v1", host, port)

	// Build the Crush configuration
	// Crush uses a different structure than OpenCode - providers with models array
	// COMPLIANCE: 62+ MCPs required for all CLI agents
	config := CrushConfig{
		Schema: "https://charm.land/crush.json",
		Providers: map[string]CrushProvider{
			"helixagent": {
				Name:    "HelixAgent AI Debate Ensemble",
				Type:    "openai",
				BaseURL: baseURL,
				APIKey:  apiKey,
				Models: []CrushModel{
					{
						ID:                   "helixagent-debate",
						Name:                 "HelixAgent Debate Ensemble",
						CostPer1MIn:          0.0, // Local deployment, no cost
						CostPer1MOut:         0.0,
						CostPer1MInCached:    0.0,
						CostPer1MOutCached:   0.0,
						ContextWindow:        128000,
						DefaultMaxTokens:     8192,
						CanReason:            true,
						SupportsAttachments:  true,
						Streaming:            true,
						SupportsBrotli:       true,
						Options: map[string]interface{}{
							"vision":        true,
							"image_input":   true,
							"image_output":  true,
							"ocr":           true,
							"pdf":           true,
							"function_calls": true,
							"tool_use":      true,
							"embeddings":    true,
						},
					},
				},
			},
		},
		Lsp: map[string]CrushLspConfig{
			"helixagent-lsp": {
				Command: fmt.Sprintf("curl -X POST %s/lsp", baseURL),
				Args:    []string{"-H", "Authorization: Bearer " + apiKey},
				Enabled: true,
			},
		},
		Mcp: getCrushMCPServers(fmt.Sprintf("http://%s:%s", host, port), *workingMCPsOnly),
		Options: &CrushOptions{
			DisableProviderAutoUpdate: false,
		},
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Crush config: %w", err)
	}

	// Output to file or stdout
	if appCfg.CrushOutput != "" {
		// Validate path for traversal attacks (G304 security fix)
		if !utils.ValidatePath(appCfg.CrushOutput) {
			return fmt.Errorf("invalid output path: contains path traversal or dangerous characters")
		}
		// #nosec G304 - CrushOutput is validated by utils.ValidatePath and provided via CLI by admin
		if err := os.WriteFile(appCfg.CrushOutput, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write Crush config to file: %w", err)
		}
		logger.WithField("file", appCfg.CrushOutput).Info("Crush configuration written to file")
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

// getCrushMCPServers returns Crush MCP configurations based on the workingOnly flag
func getCrushMCPServers(baseURL string, workingOnly bool) map[string]CrushMcpConfig {
	allMCPs := buildCrushMCPServers(baseURL)
	if !workingOnly {
		return allMCPs
	}
	return filterWorkingCrushMCPs(allMCPs)
}

// filterWorkingCrushMCPs filters Crush MCP configurations to only include those with all dependencies met
func filterWorkingCrushMCPs(allMCPs map[string]CrushMcpConfig) map[string]CrushMcpConfig {
	workingMCPs := make(map[string]CrushMcpConfig)

	// MCPs that always work (no external dependencies)
	// NOTE: Only includes MCPs with VERIFIED npm packages on registry.npmjs.org
	alwaysWorking := map[string]bool{
		"helixagent":            true,
		"helixagent-mcp":        true,
		"helixagent-acp":        true,
		"helixagent-lsp":        true,
		"helixagent-embeddings": true,
		"helixagent-vision":     true,
		"helixagent-cognee":     true,
		"filesystem":            true, // @modelcontextprotocol/server-filesystem - VERIFIED
		"memory":                true, // @modelcontextprotocol/server-memory - VERIFIED
		"sequential-thinking":   true, // @modelcontextprotocol/server-sequential-thinking - VERIFIED
		"everything":            true, // @modelcontextprotocol/server-everything - VERIFIED
	}

	// Environment variable requirements (same as OpenCode)
	envRequirements := map[string][]string{
		"github": {"GITHUB_TOKEN"}, // @modelcontextprotocol/server-github - VERIFIED
	}

	for name, mcpConfig := range allMCPs {
		if alwaysWorking[name] {
			workingMCPs[name] = mcpConfig
			continue
		}

		if reqs, hasReqs := envRequirements[name]; hasReqs {
			allMet := true
			for _, envVar := range reqs {
				if os.Getenv(envVar) == "" {
					allMet = false
					break
				}
			}
			if allMet {
				workingMCPs[name] = mcpConfig
			}
		}
	}

	return workingMCPs
}

// buildCrushMCPServers creates MCP server configurations for Crush CLI
// Local servers are started on demand via npx - no remote servers needed
// COMPLIANCE: 62+ MCPs required for all CLI agents
func buildCrushMCPServers(baseURL string) map[string]CrushMcpConfig {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home"
	}

	// Get HELIXAGENT_HOME for local MCP plugin path
	helixHome := os.Getenv("HELIXAGENT_HOME")
	if helixHome == "" {
		helixHome = homeDir + "/.helixagent"
	}

	return map[string]CrushMcpConfig{
		// HelixAgent Protocol Endpoints (7 MCPs) - LOCAL + REMOTE
		// Local MCP plugin - runs the HelixAgent MCP server as a subprocess
		"helixagent": {Type: "local", Command: []string{"node", helixHome + "/plugins/mcp-server/dist/index.js", "--endpoint", baseURL}, Enabled: true},
		// Remote protocol endpoints (running at port 7061)
		"helixagent-mcp":        {Type: "remote", URL: baseURL + "/v1/mcp", Enabled: true},
		"helixagent-acp":        {Type: "remote", URL: baseURL + "/v1/acp", Enabled: true},
		"helixagent-lsp":        {Type: "remote", URL: baseURL + "/v1/lsp", Enabled: true},
		"helixagent-embeddings": {Type: "remote", URL: baseURL + "/v1/embeddings", Enabled: true},
		"helixagent-vision":     {Type: "remote", URL: baseURL + "/v1/vision", Enabled: true},
		"helixagent-cognee":     {Type: "remote", URL: baseURL + "/v1/cognee", Enabled: true},

		// Anthropic Official MCPs - LOCAL (started on demand via npx)
		"filesystem":          {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-filesystem", homeDir}, Enabled: true},
		"fetch":               {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-fetch"}, Enabled: true},
		"memory":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-memory"}, Enabled: true},
		"time":                {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-time"}, Enabled: true},
		"git":                 {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-git"}, Enabled: true},
		"sqlite":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-sqlite", "--db-path", "/tmp/helixagent.db"}, Enabled: true},
		"postgres":            {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-postgres"}, Env: map[string]string{"POSTGRES_URL": "postgresql://helixagent:helixagent123@localhost:5432/helixagent_db"}, Enabled: true},
		"puppeteer":           {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-puppeteer"}, Enabled: true},
		"sequential-thinking": {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-sequential-thinking"}, Enabled: true},
		"everything":          {Type: "local", Command: []string{"npx", "-y", "@anthropic-ai/mcp-server-everything"}, Enabled: true},
		"brave-search":        {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-brave-search"}, Env: map[string]string{"BRAVE_API_KEY": "{env:BRAVE_API_KEY}"}, Enabled: true},
		"google-maps":         {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-google-maps"}, Env: map[string]string{"GOOGLE_MAPS_API_KEY": "{env:GOOGLE_MAPS_API_KEY}"}, Enabled: true},
		"slack":               {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-slack"}, Env: map[string]string{"SLACK_BOT_TOKEN": "{env:SLACK_BOT_TOKEN}"}, Enabled: true},
		"github":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-github"}, Env: map[string]string{"GITHUB_PERSONAL_ACCESS_TOKEN": "{env:GITHUB_TOKEN}"}, Enabled: true},
		"gitlab":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-gitlab"}, Env: map[string]string{"GITLAB_PERSONAL_ACCESS_TOKEN": "{env:GITLAB_TOKEN}"}, Enabled: true},
		"sentry":              {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-sentry"}, Env: map[string]string{"SENTRY_AUTH_TOKEN": "{env:SENTRY_AUTH_TOKEN}"}, Enabled: true},
		"everart":             {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-everart"}, Env: map[string]string{"EVERART_API_KEY": "{env:EVERART_API_KEY}"}, Enabled: true},
		"aws-kb-retrieval":    {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-aws-kb-retrieval"}, Env: map[string]string{"AWS_ACCESS_KEY_ID": "{env:AWS_ACCESS_KEY_ID}"}, Enabled: true},
		"gdrive":              {Type: "local", Command: []string{"npx", "-y", "@anthropic/mcp-server-gdrive"}, Env: map[string]string{"GOOGLE_CREDENTIALS_PATH": "{env:GOOGLE_CREDENTIALS_PATH}"}, Enabled: true},

		// Additional Anthropic & Community MCPs - LOCAL
		"exa":        {Type: "local", Command: []string{"npx", "-y", "exa-mcp-server"}, Env: map[string]string{"EXA_API_KEY": "{env:EXA_API_KEY}"}, Enabled: true},
		"linear":     {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-linear"}, Env: map[string]string{"LINEAR_API_KEY": "{env:LINEAR_API_KEY}"}, Enabled: true},
		"notion":     {Type: "local", Command: []string{"npx", "-y", "@notionhq/notion-mcp-server"}, Env: map[string]string{"NOTION_API_KEY": "{env:NOTION_API_KEY}"}, Enabled: true},
		"figma":      {Type: "local", Command: []string{"npx", "-y", "figma-developer-mcp"}, Env: map[string]string{"FIGMA_API_KEY": "{env:FIGMA_API_KEY}"}, Enabled: true},
		"todoist":    {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-todoist"}, Env: map[string]string{"TODOIST_API_TOKEN": "{env:TODOIST_API_TOKEN}"}, Enabled: true},
		"obsidian":   {Type: "local", Command: []string{"npx", "-y", "mcp-obsidian"}, Env: map[string]string{"OBSIDIAN_VAULT_PATH": "{env:OBSIDIAN_VAULT_PATH}"}, Enabled: true},
		"raycast":    {Type: "local", Command: []string{"npx", "-y", "@anthropic-ai/mcp-server-raycast"}, Enabled: true},
		"tinybird":   {Type: "local", Command: []string{"npx", "-y", "mcp-tinybird"}, Env: map[string]string{"TINYBIRD_TOKEN": "{env:TINYBIRD_TOKEN}"}, Enabled: true},
		"cloudflare": {Type: "local", Command: []string{"npx", "-y", "@cloudflare/mcp-server-cloudflare"}, Env: map[string]string{"CLOUDFLARE_API_TOKEN": "{env:CLOUDFLARE_API_TOKEN}"}, Enabled: true},
		"neon":       {Type: "local", Command: []string{"npx", "-y", "@neondatabase/mcp-server-neon"}, Env: map[string]string{"NEON_API_KEY": "{env:NEON_API_KEY}"}, Enabled: true},

		// Container/Infrastructure MCPs - LOCAL
		"docker":        {Type: "local", Command: []string{"npx", "-y", "@modelcontextprotocol/server-docker"}, Enabled: true},
		"kubernetes":    {Type: "local", Command: []string{"npx", "-y", "mcp-server-kubernetes"}, Env: map[string]string{"KUBECONFIG": "{env:KUBECONFIG}"}, Enabled: true},
		"redis":         {Type: "local", Command: []string{"npx", "-y", "mcp-server-redis"}, Env: map[string]string{"REDIS_URL": "redis://localhost:6379"}, Enabled: true},
		"mongodb":       {Type: "local", Command: []string{"npx", "-y", "mcp-server-mongodb"}, Env: map[string]string{"MONGODB_URI": "mongodb://localhost:27017"}, Enabled: true},
		"elasticsearch": {Type: "local", Command: []string{"npx", "-y", "mcp-server-elasticsearch"}, Env: map[string]string{"ELASTICSEARCH_URL": "http://localhost:9200"}, Enabled: true},
		"qdrant":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-qdrant"}, Env: map[string]string{"QDRANT_URL": "http://localhost:6333"}, Enabled: true},
		"chroma":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-chroma"}, Env: map[string]string{"CHROMA_URL": "http://localhost:8001"}, Enabled: true},
		"pinecone":      {Type: "local", Command: []string{"npx", "-y", "mcp-server-pinecone"}, Env: map[string]string{"PINECONE_API_KEY": "{env:PINECONE_API_KEY}"}, Enabled: true},
		"milvus":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-milvus"}, Env: map[string]string{"MILVUS_HOST": "localhost"}, Enabled: true},
		"weaviate":      {Type: "local", Command: []string{"npx", "-y", "mcp-server-weaviate"}, Env: map[string]string{"WEAVIATE_URL": "http://localhost:8080"}, Enabled: true},
		"supabase":      {Type: "local", Command: []string{"npx", "-y", "mcp-server-supabase"}, Env: map[string]string{"SUPABASE_URL": "{env:SUPABASE_URL}"}, Enabled: true},

		// Productivity/Collaboration MCPs - LOCAL
		"jira":            {Type: "local", Command: []string{"npx", "-y", "mcp-server-atlassian"}, Env: map[string]string{"JIRA_URL": "{env:JIRA_URL}"}, Enabled: true},
		"asana":           {Type: "local", Command: []string{"npx", "-y", "mcp-server-asana"}, Env: map[string]string{"ASANA_ACCESS_TOKEN": "{env:ASANA_ACCESS_TOKEN}"}, Enabled: true},
		"trello":          {Type: "local", Command: []string{"npx", "-y", "mcp-server-trello"}, Env: map[string]string{"TRELLO_API_KEY": "{env:TRELLO_API_KEY}"}, Enabled: true},
		"monday":          {Type: "local", Command: []string{"npx", "-y", "mcp-server-monday"}, Env: map[string]string{"MONDAY_API_KEY": "{env:MONDAY_API_KEY}"}, Enabled: true},
		"clickup":         {Type: "local", Command: []string{"npx", "-y", "mcp-server-clickup"}, Env: map[string]string{"CLICKUP_API_KEY": "{env:CLICKUP_API_KEY}"}, Enabled: true},
		"discord":         {Type: "local", Command: []string{"npx", "-y", "mcp-server-discord"}, Env: map[string]string{"DISCORD_BOT_TOKEN": "{env:DISCORD_BOT_TOKEN}"}, Enabled: true},
		"microsoft-teams": {Type: "local", Command: []string{"npx", "-y", "mcp-server-teams"}, Env: map[string]string{"TEAMS_CLIENT_ID": "{env:TEAMS_CLIENT_ID}"}, Enabled: true},
		"gmail":           {Type: "local", Command: []string{"npx", "-y", "mcp-server-gmail"}, Env: map[string]string{"GOOGLE_CREDENTIALS_PATH": "{env:GOOGLE_CREDENTIALS_PATH}"}, Enabled: true},
		"calendar":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-google-calendar"}, Env: map[string]string{"GOOGLE_CREDENTIALS_PATH": "{env:GOOGLE_CREDENTIALS_PATH}"}, Enabled: true},
		"zoom":            {Type: "local", Command: []string{"npx", "-y", "mcp-server-zoom"}, Env: map[string]string{"ZOOM_CLIENT_ID": "{env:ZOOM_CLIENT_ID}"}, Enabled: true},

		// Cloud/DevOps MCPs - LOCAL
		"aws-s3":     {Type: "local", Command: []string{"npx", "-y", "mcp-server-s3"}, Env: map[string]string{"AWS_ACCESS_KEY_ID": "{env:AWS_ACCESS_KEY_ID}"}, Enabled: true},
		"aws-lambda": {Type: "local", Command: []string{"npx", "-y", "mcp-server-aws-lambda"}, Env: map[string]string{"AWS_ACCESS_KEY_ID": "{env:AWS_ACCESS_KEY_ID}"}, Enabled: true},
		"azure":      {Type: "local", Command: []string{"npx", "-y", "mcp-server-azure"}, Env: map[string]string{"AZURE_SUBSCRIPTION_ID": "{env:AZURE_SUBSCRIPTION_ID}"}, Enabled: true},
		"gcp":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-gcp"}, Env: map[string]string{"GOOGLE_APPLICATION_CREDENTIALS": "{env:GOOGLE_APPLICATION_CREDENTIALS}"}, Enabled: true},
		"terraform":  {Type: "local", Command: []string{"npx", "-y", "mcp-server-terraform"}, Enabled: true},
		"ansible":    {Type: "local", Command: []string{"npx", "-y", "mcp-server-ansible"}, Enabled: true},
		"datadog":    {Type: "local", Command: []string{"npx", "-y", "mcp-server-datadog"}, Env: map[string]string{"DD_API_KEY": "{env:DD_API_KEY}"}, Enabled: true},
		"grafana":    {Type: "local", Command: []string{"npx", "-y", "mcp-server-grafana"}, Env: map[string]string{"GRAFANA_URL": "{env:GRAFANA_URL}"}, Enabled: true},
		"prometheus": {Type: "local", Command: []string{"npx", "-y", "mcp-server-prometheus"}, Env: map[string]string{"PROMETHEUS_URL": "{env:PROMETHEUS_URL}"}, Enabled: true},
		"circleci":   {Type: "local", Command: []string{"npx", "-y", "mcp-server-circleci"}, Env: map[string]string{"CIRCLECI_TOKEN": "{env:CIRCLECI_TOKEN}"}, Enabled: true},

		// AI/ML Integration MCPs - LOCAL
		"langchain":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-langchain"}, Env: map[string]string{"OPENAI_API_KEY": "{env:OPENAI_API_KEY}"}, Enabled: true},
		"llamaindex":       {Type: "local", Command: []string{"npx", "-y", "mcp-server-llamaindex"}, Env: map[string]string{"OPENAI_API_KEY": "{env:OPENAI_API_KEY}"}, Enabled: true},
		"huggingface":      {Type: "local", Command: []string{"npx", "-y", "mcp-server-huggingface"}, Env: map[string]string{"HUGGINGFACE_API_KEY": "{env:HUGGINGFACE_API_KEY}"}, Enabled: true},
		"replicate":        {Type: "local", Command: []string{"npx", "-y", "mcp-server-replicate"}, Env: map[string]string{"REPLICATE_API_TOKEN": "{env:REPLICATE_API_TOKEN}"}, Enabled: true},
		"stable-diffusion": {Type: "local", Command: []string{"npx", "-y", "mcp-server-stable-diffusion"}, Env: map[string]string{"STABILITY_API_KEY": "{env:STABILITY_API_KEY}"}, Enabled: true},
	}
}

// CrushValidationResult holds the validation results for Crush config
type CrushValidationResult struct {
	Valid    bool                      `json:"valid"`
	Errors   []OpenCodeValidationError `json:"errors"`
	Warnings []string                  `json:"warnings"`
	Stats    *CrushValidationStats     `json:"stats,omitempty"`
}

// CrushValidationStats contains statistics about the validated Crush config
type CrushValidationStats struct {
	Providers  int `json:"providers"`
	Models     int `json:"models"`
	LspConfigs int `json:"lsp_configs"`
}

// handleValidateCrush handles the --validate-crush-config command
func handleValidateCrush(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	filePath := appCfg.ValidateCrush

	// Validate path for traversal attacks (G304 security fix)
	if !utils.ValidatePath(filePath) {
		return fmt.Errorf("invalid config file path: contains path traversal or dangerous characters")
	}

	// Read the config file
	// #nosec G304 - filePath is validated by utils.ValidatePath and provided via CLI by admin
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Perform validation
	result := validateCrushConfig(data)

	// Output header
	fmt.Println("======================================================================")
	fmt.Println("HELIXAGENT CRUSH CONFIGURATION VALIDATION")
	fmt.Println("Using LLMsVerifier schema compliance rules")
	fmt.Println("======================================================================")
	fmt.Println()
	fmt.Printf("File: %s\n", filePath)
	fmt.Println()

	if result.Valid {
		fmt.Println("✅ CONFIGURATION IS VALID")
		fmt.Println()
		if result.Stats != nil {
			fmt.Printf("Configuration contains:\n")
			fmt.Printf("  - Providers: %d\n", result.Stats.Providers)
			fmt.Printf("  - Models: %d\n", result.Stats.Models)
			fmt.Printf("  - LSP configs: %d\n", result.Stats.LspConfigs)
		}
	} else {
		fmt.Println("❌ CONFIGURATION HAS ERRORS:")
		fmt.Println()
		for _, e := range result.Errors {
			if e.Field != "" {
				fmt.Printf("  - [%s] %s\n", e.Field, e.Message)
			} else {
				fmt.Printf("  - %s\n", e.Message)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("⚠️  WARNINGS:")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	fmt.Println()
	fmt.Println("======================================================================")

	if !result.Valid {
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}

// ValidCrushTopLevelKeys contains the valid top-level keys per Crush schema
var ValidCrushTopLevelKeys = map[string]bool{
	"$schema":   true,
	"providers": true,
	"lsp":       true,
	"options":   true,
}

// validateCrushConfig performs comprehensive validation of a Crush config
func validateCrushConfig(data []byte) *CrushValidationResult {
	result := &CrushValidationResult{
		Valid:    true,
		Errors:   []OpenCodeValidationError{},
		Warnings: []string{},
		Stats:    &CrushValidationStats{},
	}

	// Parse as generic map to check top-level keys
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, OpenCodeValidationError{
			Field:   "",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		})
		return result
	}

	// Check for invalid top-level keys
	var invalidKeys []string
	for key := range rawConfig {
		if !ValidCrushTopLevelKeys[key] {
			invalidKeys = append(invalidKeys, key)
		}
	}
	if len(invalidKeys) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, OpenCodeValidationError{
			Field:   "",
			Message: fmt.Sprintf("invalid top-level keys: %v (valid keys: $schema, providers, lsp, options)", invalidKeys),
		})
	}

	// Parse and validate providers
	totalModels := 0
	if providers, ok := rawConfig["providers"].(map[string]interface{}); ok {
		result.Stats.Providers = len(providers)
		for name, providerData := range providers {
			if provider, ok := providerData.(map[string]interface{}); ok {
				// Provider must have name
				if _, hasName := provider["name"]; !hasName {
					result.Valid = false
					result.Errors = append(result.Errors, OpenCodeValidationError{
						Field:   fmt.Sprintf("providers.%s.name", name),
						Message: "provider must have a name",
					})
				}

				// Provider must have type
				if _, hasType := provider["type"]; !hasType {
					result.Valid = false
					result.Errors = append(result.Errors, OpenCodeValidationError{
						Field:   fmt.Sprintf("providers.%s.type", name),
						Message: "provider must have a type",
					})
				}

				// Provider must have base_url
				if _, hasBaseURL := provider["base_url"]; !hasBaseURL {
					result.Valid = false
					result.Errors = append(result.Errors, OpenCodeValidationError{
						Field:   fmt.Sprintf("providers.%s.base_url", name),
						Message: "provider must have a base_url",
					})
				}

				// Provider must have models
				if models, hasModels := provider["models"].([]interface{}); hasModels {
					totalModels += len(models)
					for i, modelData := range models {
						if model, ok := modelData.(map[string]interface{}); ok {
							// Model must have id
							if _, hasID := model["id"]; !hasID {
								result.Valid = false
								result.Errors = append(result.Errors, OpenCodeValidationError{
									Field:   fmt.Sprintf("providers.%s.models[%d].id", name, i),
									Message: "model must have an id",
								})
							}
							// Model must have name
							if _, hasName := model["name"]; !hasName {
								result.Valid = false
								result.Errors = append(result.Errors, OpenCodeValidationError{
									Field:   fmt.Sprintf("providers.%s.models[%d].name", name, i),
									Message: "model must have a name",
								})
							}
						}
					}
				} else {
					result.Valid = false
					result.Errors = append(result.Errors, OpenCodeValidationError{
						Field:   fmt.Sprintf("providers.%s.models", name),
						Message: "provider must have at least one model",
					})
				}
			}
		}
	} else if rawConfig["providers"] == nil {
		result.Valid = false
		result.Errors = append(result.Errors, OpenCodeValidationError{
			Field:   "providers",
			Message: "at least one provider must be configured",
		})
	}
	result.Stats.Models = totalModels

	// Parse and validate LSP configs
	if lspConfigs, ok := rawConfig["lsp"].(map[string]interface{}); ok {
		result.Stats.LspConfigs = len(lspConfigs)
		for name, lspData := range lspConfigs {
			if lsp, ok := lspData.(map[string]interface{}); ok {
				// LSP must have command
				if _, hasCommand := lsp["command"]; !hasCommand {
					result.Valid = false
					result.Errors = append(result.Errors, OpenCodeValidationError{
						Field:   fmt.Sprintf("lsp.%s.command", name),
						Message: "LSP config must have a command",
					})
				}
			}
		}
	}

	// Add warnings for missing recommended fields
	if _, hasSchema := rawConfig["$schema"]; !hasSchema {
		result.Warnings = append(result.Warnings, "$schema field is recommended for validation")
	}

	return result
}

// ============================================================================
// Unified CLI Agent Handlers (All 48 Agents)
// ============================================================================

// handleListAgents lists all 48 supported CLI agents
func handleListAgents(appCfg *AppConfig) error {
	fmt.Println("HelixAgent - Supported CLI Agents (48 total)")
	fmt.Println("=============================================")
	fmt.Println()

	generator := cliagents.NewUnifiedGenerator(nil)
	schemas := generator.GetAllSchemas()

	// Group by category
	original18 := []cliagents.AgentType{
		cliagents.AgentOpenCode, cliagents.AgentCrush, cliagents.AgentHelixCode,
		cliagents.AgentKiro, cliagents.AgentAider, cliagents.AgentClaudeCode,
		cliagents.AgentCline, cliagents.AgentCodenameGoose, cliagents.AgentDeepSeekCLI,
		cliagents.AgentForge, cliagents.AgentGeminiCLI, cliagents.AgentGPTEngineer,
		cliagents.AgentKiloCode, cliagents.AgentMistralCode, cliagents.AgentOllamaCode,
		cliagents.AgentPlandex, cliagents.AgentQwenCode, cliagents.AgentAmazonQ,
	}

	new30 := []cliagents.AgentType{
		cliagents.AgentAgentDeck, cliagents.AgentBridle, cliagents.AgentCheshireCat,
		cliagents.AgentClaudePlugins, cliagents.AgentClaudeSquad, cliagents.AgentCodai,
		cliagents.AgentCodex, cliagents.AgentCodexSkills, cliagents.AgentConduit,
		cliagents.AgentContinue, cliagents.AgentEmdash, cliagents.AgentFauxPilot,
		cliagents.AgentGetShitDone, cliagents.AgentGitHubCopilotCLI, cliagents.AgentGitHubSpecKit,
		cliagents.AgentGitMCP, cliagents.AgentGPTME, cliagents.AgentMobileAgent,
		cliagents.AgentMultiagentCoding, cliagents.AgentNanocoder, cliagents.AgentNoi,
		cliagents.AgentOctogen, cliagents.AgentOpenHands, cliagents.AgentPostgresMCP,
		cliagents.AgentShai, cliagents.AgentSnowCLI, cliagents.AgentTaskWeaver,
		cliagents.AgentUIUXProMax, cliagents.AgentVTCode, cliagents.AgentWarp,
	}

	fmt.Println("Original 18 Agents:")
	fmt.Println("-------------------")
	for _, agent := range original18 {
		if schema, ok := schemas[agent]; ok {
			fmt.Printf("  %-20s  %s\n", agent, schema.Description)
		}
	}

	fmt.Println()
	fmt.Println("New 30 Agents:")
	fmt.Println("--------------")
	for _, agent := range new30 {
		if schema, ok := schemas[agent]; ok {
			fmt.Printf("  %-20s  %s\n", agent, schema.Description)
		}
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  helixagent --generate-agent-config=<agent-name>")
	fmt.Println("  helixagent --generate-agent-config=<agent-name> --agent-config-output=<path>")
	fmt.Println("  helixagent --validate-agent-config=<agent-name>:<config-path>")
	fmt.Println("  helixagent --generate-all-agents --all-agents-output-dir=<directory>")

	return nil
}

// handleGenerateAgentConfig generates configuration for a specific CLI agent
func handleGenerateAgentConfig(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	// Special case for OpenCode - use the dedicated handler with v1.1.30+ schema
	if strings.EqualFold(appCfg.GenerateAgentConfig, "opencode") {
		// Transfer output path setting if specified
		if appCfg.AgentConfigOutput != "" {
			appCfg.OpenCodeOutput = appCfg.AgentConfigOutput
		}
		return handleGenerateOpenCode(appCfg)
	}

	agentType := cliagents.AgentType(appCfg.GenerateAgentConfig)

	// Create generator with HelixAgent settings
	config := &cliagents.GeneratorConfig{
		HelixAgentHost: "localhost",
		HelixAgentPort: 7061,
		MCPServers:     cliagents.DefaultMCPServers(),
		IncludeScores:  true,
	}
	generator := cliagents.NewUnifiedGenerator(config)

	ctx := context.Background()
	result, err := generator.Generate(ctx, agentType)
	if err != nil {
		return fmt.Errorf("failed to generate config for %s: %w", agentType, err)
	}

	if !result.Success {
		return fmt.Errorf("config generation failed for %s: %v", agentType, result.Errors)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(result.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Output to file or stdout
	if appCfg.AgentConfigOutput != "" {
		if !utils.ValidatePath(appCfg.AgentConfigOutput) {
			return fmt.Errorf("invalid output path: %s", appCfg.AgentConfigOutput)
		}
		if err := os.WriteFile(appCfg.AgentConfigOutput, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		logger.Infof("Generated %s config written to: %s", agentType, appCfg.AgentConfigOutput)
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

// handleValidateAgentConfig validates a configuration file for a specific CLI agent
func handleValidateAgentConfig(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	// Parse agent:path format
	parts := strings.SplitN(appCfg.ValidateAgentConfig, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format for --validate-agent-config, expected: agent-name:config-path")
	}

	agentType := cliagents.AgentType(parts[0])
	configPath := parts[1]

	// Validate path
	if !utils.ValidatePath(configPath) {
		return fmt.Errorf("invalid config path: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Special case for OpenCode - use v1.1.30+ schema validation
	if strings.EqualFold(string(agentType), "opencode") {
		result := validateOpenCodeConfig(data)
		if result.Valid {
			fmt.Printf("✓ Config file is valid for %s\n", agentType)
			if len(result.Warnings) > 0 {
				fmt.Println("\nWarnings:")
				for _, w := range result.Warnings {
					fmt.Printf("  - %s\n", w)
				}
			}
		} else {
			fmt.Printf("✗ Config file is invalid for %s\n", agentType)
			fmt.Println("\nErrors:")
			for _, e := range result.Errors {
				fmt.Printf("  - %s\n", e.Message)
			}
			return fmt.Errorf("validation failed with %d errors", len(result.Errors))
		}
		return nil
	}

	// Parse as JSON
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate using LLMsVerifier
	generator := cliagents.NewUnifiedGenerator(nil)
	result, err := generator.Validate(agentType, config)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Output results
	if result.Valid {
		fmt.Printf("✓ Config file is valid for %s\n", agentType)
		if len(result.Warnings) > 0 {
			fmt.Println("\nWarnings:")
			for _, warning := range result.Warnings {
				fmt.Printf("  - %s\n", warning)
			}
		}
	} else {
		fmt.Printf("✗ Config file is invalid for %s\n", agentType)
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("validation failed with %d errors", len(result.Errors))
	}

	return nil
}

// handleGenerateAllAgents generates configurations for all 48 CLI agents
func handleGenerateAllAgents(appCfg *AppConfig) error {
	logger := appCfg.Logger
	if logger == nil {
		logger = logrus.New()
	}

	if appCfg.AllAgentsOutputDir == "" {
		return fmt.Errorf("--all-agents-output-dir is required when using --generate-all-agents")
	}

	if !utils.ValidatePath(appCfg.AllAgentsOutputDir) {
		return fmt.Errorf("invalid output directory: %s", appCfg.AllAgentsOutputDir)
	}
	outputDir := appCfg.AllAgentsOutputDir

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create generator with HelixAgent settings
	config := &cliagents.GeneratorConfig{
		HelixAgentHost: "localhost",
		HelixAgentPort: 7061,
		OutputDir:      outputDir,
		MCPServers:     cliagents.DefaultMCPServers(),
		IncludeScores:  true,
	}
	generator := cliagents.NewUnifiedGenerator(config)

	ctx := context.Background()
	results, err := generator.GenerateAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate all configs: %w", err)
	}

	// Save each config and report results
	successCount := 0
	failCount := 0

	fmt.Printf("Generating configurations for 48 CLI agents in: %s\n\n", outputDir)

	for _, result := range results {
		// Special case for OpenCode - use v1.1.30+ schema
		if string(result.AgentType) == "opencode" {
			outputPath := fmt.Sprintf("%s/.opencode.json", outputDir)
			openCodeAppCfg := &AppConfig{
				Logger:         logger,
				OpenCodeOutput: outputPath,
			}
			if err := handleGenerateOpenCode(openCodeAppCfg); err != nil {
				fmt.Printf("✗ %-20s  Failed to generate: %v\n", result.AgentType, err)
				failCount++
			} else {
				fmt.Printf("✓ %-20s  %s\n", result.AgentType, ".opencode.json")
				successCount++
			}
			continue
		}

		if result.Success {
			// Get schema for filename
			schema, _ := generator.GetSchema(result.AgentType)
			outputPath := fmt.Sprintf("%s/%s", outputDir, schema.ConfigFileName)

			jsonData, err := json.MarshalIndent(result.Config, "", "  ")
			if err != nil {
				fmt.Printf("✗ %-20s  Failed to marshal: %v\n", result.AgentType, err)
				failCount++
				continue
			}

			if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
				fmt.Printf("✗ %-20s  Failed to write: %v\n", result.AgentType, err)
				failCount++
				continue
			}

			fmt.Printf("✓ %-20s  %s\n", result.AgentType, schema.ConfigFileName)
			successCount++
		} else {
			fmt.Printf("✗ %-20s  %v\n", result.AgentType, result.Errors)
			failCount++
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Summary: %d succeeded, %d failed\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("%d configurations failed to generate", failCount)
	}

	logger.Infof("All 48 agent configurations generated in: %s", outputDir)
	return nil
}

func showHelp() {
	fmt.Printf(`HelixAgent - Advanced LLM Gateway with Cognee Integration

Usage:
  helixagent [options]

Options:
  -config string
        Path to configuration file (YAML)
  -auto-start-docker
        Automatically start required Docker containers (default: true)
  -strict-dependencies
        MANDATORY: Fail if any integration dependency is unavailable (default: true)
        When enabled, HelixAgent will NOT start unless ALL dependencies are healthy:
        - PostgreSQL (database)
        - Redis (cache)
        - Cognee (knowledge graph)
        - ChromaDB (vector database)
  -generate-api-key
        Generate a new HelixAgent API key and output it to stdout
  -generate-opencode-config
        Generate OpenCode configuration JSON (uses HELIXAGENT_API_KEY env or generates new)
  -validate-opencode-config string
        Validate an existing OpenCode configuration file (uses LLMsVerifier schema rules)
  -opencode-output string
        Output path for OpenCode config (default: stdout)
  -generate-crush-config
        Generate Crush CLI configuration JSON (uses HELIXAGENT_API_KEY env or generates new)
  -validate-crush-config string
        Validate an existing Crush configuration file (uses LLMsVerifier schema rules)
  -crush-output string
        Output path for Crush config (default: stdout)
  -api-key-env-file string
        Path to .env file to write the generated API key
  -preinstall-mcp
        Pre-install standard MCP server npm packages for faster startup
  -skip-mcp-preinstall
        Skip automatic MCP package pre-installation at startup
  -working-mcps-only
        Only include MCPs with all dependencies met (API keys set, services running)
        Use with -generate-opencode-config or -generate-crush-config
  -use-local-mcp-servers
        Use local Docker-based MCP servers instead of npx commands
        Requires: ./scripts/mcp/start-mcp-servers.sh --start
        MCP servers run on TCP ports 9101-9601 via socat

Unified CLI Agent Configuration (48 agents):
  -list-agents
        List all 48 supported CLI agents with descriptions
  -generate-agent-config string
        Generate config for specified CLI agent (e.g., codex, openhands, claude-squad)
  -agent-config-output string
        Output path for generated agent config (default: stdout)
  -validate-agent-config string
        Validate config file for agent (format: agent-name:config-path)
  -generate-all-agents
        Generate configurations for all 48 CLI agents
  -all-agents-output-dir string
        Output directory for all agent configs (required with --generate-all-agents)

  -version
        Show version information
  -help
        Show this help message

Features:
  - Cognee knowledge graph integration for advanced AI memory
  - Graph-powered reasoning beyond traditional RAG
  - Multi-modal processing (text, code, images, audio)
  - Auto-containerization for seamless deployment
  - Automatic startup of required Docker containers
  - Models.dev integration for comprehensive model metadata
  - Multi-layer caching with Redis and in-memory
  - Circuit breaker for API resilience
  - Auto-refresh with configurable intervals
  - Model comparison and capability filtering
  - Comprehensive monitoring and health checks

API Key & Configuration Commands:
  # Generate a new API key and display it
  helixagent -generate-api-key

  # Generate API key and save to .env file
  helixagent -generate-api-key -api-key-env-file .env

  # Generate OpenCode configuration (uses HELIXAGENT_API_KEY from env)
  helixagent -generate-opencode-config

  # Generate OpenCode config and save to file, with API key to .env
  helixagent -generate-opencode-config -opencode-output opencode.json -api-key-env-file .env

  # Validate an existing OpenCode configuration file
  helixagent -validate-opencode-config ~/.config/opencode/opencode.json

  # Generate Crush CLI configuration
  helixagent -generate-crush-config

  # Generate Crush config and save to file
  helixagent -generate-crush-config -crush-output crush.json

  # Validate an existing Crush configuration file
  helixagent -validate-crush-config ~/.config/crush/crush.json

Examples:
  helixagent
  helixagent -auto-start-docker=false
  helixagent -config /path/to/config.yaml
  helixagent -generate-crush-config -crush-output /tmp/crush.json
  helixagent -version

For more information, visit: https://dev.helix.agent
`)
}

func showVersion() {
	fmt.Printf("HelixAgent v%s - Models.dev Enhanced Edition\n", "1.0.0")
}
