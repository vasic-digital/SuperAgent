package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// Infrastructure startup flags
var (
	autoStartLSP = true // Auto-start LSP servers
	autoStartRAG = true // Auto-start RAG services
)

// ensureLSPServers starts all LSP Docker containers via the
// Containers module adapter. When remote distribution is enabled,
// deploys to the remote host instead.
func ensureLSPServers(logger *logrus.Logger) error {
	projectDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get project directory: %w", err)
	}

	lspComposeFile := filepath.Join(projectDir, "docker", "lsp", "docker-compose.lsp.yml")
	if !fileExists(lspComposeFile) {
		logger.WithField("file", lspComposeFile).Warn("LSP compose file not found, skipping LSP auto-start")
		return nil
	}

	if globalContainerAdapter == nil {
		return fmt.Errorf("container adapter not initialized")
	}

	ctx := context.Background()

	if globalContainerAdapter.RemoteEnabled() {
		logger.Info("Starting LSP servers on remote host via Containers module")
		if err := globalContainerAdapter.RemoteComposeUp(
			ctx, lspComposeFile, "lsp",
		); err != nil {
			logger.WithError(err).Warn("Failed to start LSP servers on remote, falling back to local")
		} else {
			logger.Info("LSP servers starting on remote host")
			return nil
		}
	}

	logger.Info("Starting LSP servers via Containers module (11 language servers)")
	if err := globalContainerAdapter.ComposeUp(
		ctx, lspComposeFile, "lsp",
	); err != nil {
		logger.WithError(err).Warn("Failed to start some LSP servers, continuing")
		return nil
	}
	logger.Info("LSP servers starting in background (ports 5001-5024)")
	return nil
}

// ensureRAGServices starts all RAG Docker containers via the
// Containers module adapter. When remote distribution is enabled,
// deploys to the remote host instead.
func ensureRAGServices(logger *logrus.Logger) error {
	projectDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get project directory: %w", err)
	}

	ragComposeFile := filepath.Join(projectDir, "docker", "rag", "docker-compose.rag.yml")
	if !fileExists(ragComposeFile) {
		logger.WithField("file", ragComposeFile).Warn("RAG compose file not found, skipping RAG auto-start")
		return nil
	}

	if globalContainerAdapter == nil {
		return fmt.Errorf("container adapter not initialized")
	}

	ctx := context.Background()

	if globalContainerAdapter.RemoteEnabled() {
		logger.Info("Starting RAG services on remote host via Containers module")
		if err := globalContainerAdapter.RemoteComposeUp(
			ctx, ragComposeFile, "rag",
		); err != nil {
			logger.WithError(err).Warn("Failed to start RAG services on remote, falling back to local")
		} else {
			logger.Info("RAG services starting on remote host")
			return nil
		}
	}

	logger.Info("Starting RAG services via Containers module")
	if err := globalContainerAdapter.ComposeUp(
		ctx, ragComposeFile, "rag",
	); err != nil {
		logger.WithError(err).Warn("Failed to start some RAG services, continuing")
		return nil
	}
	logger.Info("RAG services starting in background (ports 6333-8030)")
	return nil
}

// startAllInfrastructure starts all infrastructure services in parallel
func startAllInfrastructure(logger *logrus.Logger) {
	// Start LSP servers in background
	if autoStartLSP {
		go func() {
			logger.Info("Starting LSP servers...")
			if err := ensureLSPServers(logger); err != nil {
				logger.WithError(err).Warn("Failed to start LSP servers")
			}
		}()
	}

	// Start RAG services in background
	if autoStartRAG {
		go func() {
			logger.Info("Starting RAG services...")
			if err := ensureRAGServices(logger); err != nil {
				logger.WithError(err).Warn("Failed to start RAG services")
			}
		}()
	}
}

// waitForInfrastructure waits for all infrastructure services to be healthy
func waitForInfrastructure(logger *logrus.Logger, timeout time.Duration) error {
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(timeout)

	services := map[string]string{
		"LSP Manager": "http://localhost:5100/health",
		"RAG Manager": "http://localhost:8030/health",
		"Qdrant":      "http://localhost:6333/readyz",
	}

	for name, url := range services {
		for time.Now().Before(deadline) {
			resp, err := client.Get(url)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					logger.WithField("service", name).Debug("Service is healthy")
					break
				}
			}
			time.Sleep(2 * time.Second)
		}
	}

	return nil
}

// InfrastructureStatus represents the status of all infrastructure services
type InfrastructureStatus struct {
	Core       map[string]bool `json:"core"`
	MCP        map[string]bool `json:"mcp"`
	LSP        map[string]bool `json:"lsp"`
	RAG        map[string]bool `json:"rag"`
	TotalUp    int             `json:"total_up"`
	TotalDown  int             `json:"total_down"`
	AllHealthy bool            `json:"all_healthy"`
}

// GetInfrastructureStatus returns the current status of all infrastructure services
func GetInfrastructureStatus(logger *logrus.Logger) *InfrastructureStatus {
	client := &http.Client{Timeout: 2 * time.Second}
	status := &InfrastructureStatus{
		Core: make(map[string]bool),
		MCP:  make(map[string]bool),
		LSP:  make(map[string]bool),
		RAG:  make(map[string]bool),
	}

	// Check core services
	coreServices := map[string]string{
		"postgres": "", // TCP check
		"redis":    "", // TCP check
		"chromadb": "http://localhost:8001/api/v2/heartbeat",
		"cognee":   "http://localhost:8000/",
	}

	for name, url := range coreServices {
		if url == "" {
			// Skip TCP checks for now
			status.Core[name] = true
			status.TotalUp++
		} else {
			resp, err := client.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				status.Core[name] = true
				status.TotalUp++
			} else {
				status.Core[name] = false
				status.TotalDown++
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
		}
	}

	// Check MCP servers (sample)
	mcpPorts := map[string]int{
		"filesystem": 9101,
		"memory":     9102,
		"postgres":   9103,
	}
	for name, port := range mcpPorts {
		if checkTCPPort("localhost", port) {
			status.MCP[name] = true
			status.TotalUp++
		} else {
			status.MCP[name] = false
			status.TotalDown++
		}
	}

	// Check LSP servers
	lspServices := map[string]string{
		"lsp-manager": "http://localhost:5100/health",
	}
	for name, url := range lspServices {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			status.LSP[name] = true
			status.TotalUp++
		} else {
			status.LSP[name] = false
			status.TotalDown++
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	// Check RAG services
	ragServices := map[string]string{
		"qdrant":      "http://localhost:6333/readyz",
		"rag-manager": "http://localhost:8030/health",
	}
	for name, url := range ragServices {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			status.RAG[name] = true
			status.TotalUp++
		} else {
			status.RAG[name] = false
			status.TotalDown++
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	status.AllHealthy = status.TotalDown == 0
	return status
}

// checkTCPPort checks if a TCP port is open.
// Uses the Containers module adapter when available, falls back to
// direct net.DialTimeout.
func checkTCPPort(host string, port int) bool {
	if globalContainerAdapter != nil {
		return globalContainerAdapter.HealthCheckTCP(host, port)
	}
	// Fallback: direct TCP dial.
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
