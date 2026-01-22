// ACP Manager - Agent Communication Protocol Manager
// Handles agent discovery, routing, and message passing
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Agent represents a registered agent
type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Capabilities []string          `json:"capabilities"`
	Endpoint     string            `json:"endpoint"`
	Status       string            `json:"status"`
	LastSeen     time.Time         `json:"last_seen"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Message represents an inter-agent message
type Message struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// AgentRegistry holds registered agents
type AgentRegistry struct {
	agents map[string]Agent
	mu     sync.RWMutex
}

var registry = &AgentRegistry{
	agents: make(map[string]Agent),
}

// Pre-registered CLI agents
var preregisteredAgents = []Agent{
	{ID: "claude-code", Name: "Claude Code", Type: "cli-agent", Capabilities: []string{"code", "chat", "tools"}, Endpoint: "stdio"},
	{ID: "opencode", Name: "OpenCode", Type: "cli-agent", Capabilities: []string{"code", "chat", "mcp"}, Endpoint: "stdio"},
	{ID: "cline", Name: "Cline", Type: "cli-agent", Capabilities: []string{"code", "chat", "hooks"}, Endpoint: "stdio"},
	{ID: "kilo-code", Name: "Kilo Code", Type: "cli-agent", Capabilities: []string{"code", "chat", "multi-platform"}, Endpoint: "stdio"},
	{ID: "aider", Name: "Aider", Type: "cli-agent", Capabilities: []string{"code", "chat", "git"}, Endpoint: "stdio"},
	{ID: "goose", Name: "Codename Goose", Type: "cli-agent", Capabilities: []string{"code", "chat", "mcp"}, Endpoint: "stdio"},
	{ID: "amazon-q", Name: "Amazon Q", Type: "cli-agent", Capabilities: []string{"code", "chat", "aws"}, Endpoint: "stdio"},
	{ID: "kiro", Name: "Kiro", Type: "cli-agent", Capabilities: []string{"code", "chat"}, Endpoint: "stdio"},
	{ID: "gemini-cli", Name: "Gemini CLI", Type: "cli-agent", Capabilities: []string{"code", "chat"}, Endpoint: "stdio"},
	{ID: "deepseek-cli", Name: "DeepSeek CLI", Type: "cli-agent", Capabilities: []string{"code", "chat"}, Endpoint: "stdio"},
	// HelixAgent internal agents
	{ID: "helixagent-debate", Name: "HelixAgent Debate", Type: "internal", Capabilities: []string{"debate", "ensemble"}, Endpoint: "http://helixagent:7061/v1/debates"},
	{ID: "helixagent-rag", Name: "HelixAgent RAG", Type: "internal", Capabilities: []string{"rag", "retrieval"}, Endpoint: "http://helixagent:7061/v1/rag"},
	{ID: "helixagent-memory", Name: "HelixAgent Memory", Type: "internal", Capabilities: []string{"memory", "storage"}, Endpoint: "http://helixagent:7061/v1/memory"},
}

func init() {
	for _, agent := range preregisteredAgents {
		agent.Status = "registered"
		agent.LastSeen = time.Now()
		registry.agents[agent.ID] = agent
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"agents":    len(registry.agents),
	})
}

func handleListAgents(w http.ResponseWriter, r *http.Request) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	agents := make([]Agent, 0, len(registry.agents))
	for _, agent := range registry.agents {
		agents = append(agents, agent)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":     len(agents),
		"agents":    agents,
		"timestamp": time.Now(),
	})
}

func handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var agent Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	agent.Status = "registered"
	agent.LastSeen = time.Now()

	registry.mu.Lock()
	registry.agents[agent.ID] = agent
	registry.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"agent":   agent,
	})
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg.Timestamp = time.Now()

	// In a real implementation, route the message to the target agent
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": msg,
	})
}

func main() {
	port := os.Getenv("ACP_PORT")
	if port == "" {
		port = "9200"
	}

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/v1/agents", handleListAgents)
	http.HandleFunc("/v1/agents/register", handleRegisterAgent)
	http.HandleFunc("/v1/messages", handleSendMessage)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":    "HelixAgent ACP Manager",
			"version": "1.0.0",
			"endpoints": []string{
				"/health",
				"/v1/agents",
				"/v1/agents/register",
				"/v1/messages",
			},
		})
	})

	log.Printf("ACP Manager starting on port %s", port)
	log.Printf("Pre-registered %d agents", len(preregisteredAgents))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
