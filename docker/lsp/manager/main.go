package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// LSPServer represents an LSP server configuration
type LSPServer struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Language string    `json:"language"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	Status   string    `json:"status"`
	LastPing time.Time `json:"last_ping"`
}

// Manager manages LSP servers
type Manager struct {
	servers map[string]*LSPServer
	mu      sync.RWMutex
}

func NewManager() *Manager {
	m := &Manager{
		servers: make(map[string]*LSPServer),
	}
	m.initializeServers()
	return m
}

func (m *Manager) initializeServers() {
	servers := []LSPServer{
		{ID: "gopls", Name: "Go Language Server", Language: "go", Host: "lsp-multi", Port: 5001},
		{ID: "rust-analyzer", Name: "Rust Analyzer", Language: "rust", Host: "lsp-multi", Port: 5002},
		{ID: "pylsp", Name: "Python LSP", Language: "python", Host: "lsp-multi", Port: 5003},
		{ID: "tsserver", Name: "TypeScript Server", Language: "typescript", Host: "lsp-multi", Port: 5004},
		{ID: "clangd", Name: "Clangd", Language: "cpp", Host: "lsp-multi", Port: 5005},
		{ID: "jdtls", Name: "Eclipse JDT.LS", Language: "java", Host: "lsp-multi", Port: 5006},
		{ID: "bash-lsp", Name: "Bash LSP", Language: "bash", Host: "lsp-devops", Port: 5020},
		{ID: "yaml-lsp", Name: "YAML LSP", Language: "yaml", Host: "lsp-devops", Port: 5021},
		{ID: "docker-lsp", Name: "Dockerfile LSP", Language: "dockerfile", Host: "lsp-devops", Port: 5022},
		{ID: "terraform-lsp", Name: "Terraform LSP", Language: "terraform", Host: "lsp-devops", Port: 5023},
		{ID: "xml-lsp", Name: "XML LSP", Language: "xml", Host: "lsp-devops", Port: 5024},
	}

	for i := range servers {
		servers[i].Status = "unknown"
		m.servers[servers[i].ID] = &servers[i]
	}
}

func (m *Manager) checkServer(server *LSPServer) {
	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		m.mu.Lock()
		server.Status = "offline"
		m.mu.Unlock()
		return
	}
	conn.Close()

	m.mu.Lock()
	server.Status = "online"
	server.LastPing = time.Now()
	m.mu.Unlock()
}

func (m *Manager) startHealthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			m.mu.RLock()
			servers := make([]*LSPServer, 0, len(m.servers))
			for _, s := range m.servers {
				servers = append(servers, s)
			}
			m.mu.RUnlock()

			for _, server := range servers {
				go m.checkServer(server)
			}
		}
	}()

	// Initial check
	m.mu.RLock()
	for _, server := range m.servers {
		go m.checkServer(server)
	}
	m.mu.RUnlock()
}

func (m *Manager) handleHealth(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	online := 0
	for _, s := range m.servers {
		if s.Status == "online" {
			online++
		}
	}
	m.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "lsp-manager",
		"servers":   len(m.servers),
		"online":    online,
		"timestamp": time.Now().Unix(),
	})
}

func (m *Manager) handleServers(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	servers := make([]*LSPServer, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	m.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"servers": servers,
		"count":   len(servers),
	})
}

func (m *Manager) handleServerStatus(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("id")
	if serverID == "" {
		http.Error(w, "server id required", http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	server, exists := m.servers[serverID]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(server)
}

func main() {
	port := os.Getenv("LSP_MANAGER_PORT")
	if port == "" {
		port = "5100"
	}

	manager := NewManager()
	manager.startHealthChecker()

	http.HandleFunc("/health", manager.handleHealth)
	http.HandleFunc("/servers", manager.handleServers)
	http.HandleFunc("/server", manager.handleServerStatus)

	log.Printf("LSP Manager starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
