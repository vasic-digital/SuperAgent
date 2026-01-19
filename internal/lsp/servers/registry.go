// Package servers provides an LSP server registry for multiple language servers.
package servers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LSPServerDefinition defines an LSP server configuration.
type LSPServerDefinition struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Language     string            `json:"language"`
	FilePatterns []string          `json:"file_patterns"` // *.go, *.rs, *.py, etc.
	Command      string            `json:"command"`
	Args         []string          `json:"args"`
	InitOptions  map[string]interface{} `json:"init_options,omitempty"`
	Capabilities LSPCapabilities   `json:"capabilities"`
	Priority     int               `json:"priority"` // Higher priority is preferred
	Enabled      bool              `json:"enabled"`
	Binary       string            `json:"binary,omitempty"` // Path to binary if found
}

// LSPCapabilities defines the capabilities of an LSP server.
type LSPCapabilities struct {
	Completion    bool `json:"completion"`
	Hover         bool `json:"hover"`
	Definition    bool `json:"definition"`
	References    bool `json:"references"`
	Diagnostics   bool `json:"diagnostics"`
	Rename        bool `json:"rename"`
	CodeAction    bool `json:"code_action"`
	Formatting    bool `json:"formatting"`
	SignatureHelp bool `json:"signature_help"`
}

// LSPServerRegistry manages LSP server configurations.
type LSPServerRegistry struct {
	servers      map[string]*LSPServerDefinition
	byLanguage   map[string][]*LSPServerDefinition
	mu           sync.RWMutex
	logger       *logrus.Logger
	searchPaths  []string
}

// RegistryConfig holds configuration for the LSP registry.
type RegistryConfig struct {
	Logger      *logrus.Logger
	SearchPaths []string
}

// NewLSPServerRegistry creates a new LSP server registry.
func NewLSPServerRegistry(config RegistryConfig) *LSPServerRegistry {
	if config.Logger == nil {
		config.Logger = logrus.New()
	}
	if len(config.SearchPaths) == 0 {
		config.SearchPaths = []string{"/usr/bin", "/usr/local/bin", "/opt/bin", "/home/*/go/bin"}
	}

	registry := &LSPServerRegistry{
		servers:     make(map[string]*LSPServerDefinition),
		byLanguage:  make(map[string][]*LSPServerDefinition),
		logger:      config.Logger,
		searchPaths: config.SearchPaths,
	}

	// Load default server definitions
	registry.loadDefaultServers()

	return registry
}

// loadDefaultServers loads the default LSP server definitions.
func (r *LSPServerRegistry) loadDefaultServers() {
	defaultServers := []LSPServerDefinition{
		// Go
		{
			ID:           "gopls",
			Name:         "gopls",
			Language:     "go",
			FilePatterns: []string{"*.go"},
			Command:      "gopls",
			Args:         []string{"serve"},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Rust
		{
			ID:           "rust-analyzer",
			Name:         "rust-analyzer",
			Language:     "rust",
			FilePatterns: []string{"*.rs"},
			Command:      "rust-analyzer",
			Args:         []string{},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Python - pylsp
		{
			ID:           "pylsp",
			Name:         "Python LSP Server",
			Language:     "python",
			FilePatterns: []string{"*.py", "*.pyi"},
			Command:      "pylsp",
			Args:         []string{},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 90,
			Enabled:  true,
		},
		// Python - pyright
		{
			ID:           "pyright",
			Name:         "Pyright",
			Language:     "python",
			FilePatterns: []string{"*.py", "*.pyi"},
			Command:      "pyright-langserver",
			Args:         []string{"--stdio"},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// TypeScript/JavaScript
		{
			ID:           "typescript-language-server",
			Name:         "TypeScript Language Server",
			Language:     "typescript",
			FilePatterns: []string{"*.ts", "*.tsx", "*.js", "*.jsx"},
			Command:      "typescript-language-server",
			Args:         []string{"--stdio"},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// C/C++ - clangd
		{
			ID:           "clangd",
			Name:         "clangd",
			Language:     "c_cpp",
			FilePatterns: []string{"*.c", "*.h", "*.cpp", "*.hpp", "*.cc", "*.cxx"},
			Command:      "clangd",
			Args:         []string{},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Java - eclipse.jdt.ls
		{
			ID:           "jdtls",
			Name:         "Eclipse JDT Language Server",
			Language:     "java",
			FilePatterns: []string{"*.java"},
			Command:      "jdtls",
			Args:         []string{},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// C# - omnisharp
		{
			ID:           "omnisharp",
			Name:         "OmniSharp",
			Language:     "csharp",
			FilePatterns: []string{"*.cs"},
			Command:      "omnisharp",
			Args:         []string{"-lsp"},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// PHP - phpactor
		{
			ID:           "phpactor",
			Name:         "Phpactor",
			Language:     "php",
			FilePatterns: []string{"*.php"},
			Command:      "phpactor",
			Args:         []string{"language-server"},
			Capabilities: LSPCapabilities{
				Completion:  true,
				Hover:       true,
				Definition:  true,
				References:  true,
				Diagnostics: true,
				Rename:      true,
				CodeAction:  true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Ruby - solargraph
		{
			ID:           "solargraph",
			Name:         "Solargraph",
			Language:     "ruby",
			FilePatterns: []string{"*.rb", "*.rake", "Gemfile"},
			Command:      "solargraph",
			Args:         []string{"stdio"},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Elixir
		{
			ID:           "elixir-ls",
			Name:         "Elixir LS",
			Language:     "elixir",
			FilePatterns: []string{"*.ex", "*.exs"},
			Command:      "elixir-ls",
			Args:         []string{},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				CodeAction:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Haskell
		{
			ID:           "haskell-language-server",
			Name:         "Haskell Language Server",
			Language:     "haskell",
			FilePatterns: []string{"*.hs", "*.lhs"},
			Command:      "haskell-language-server-wrapper",
			Args:         []string{"--lsp"},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Bash
		{
			ID:           "bash-language-server",
			Name:         "Bash Language Server",
			Language:     "bash",
			FilePatterns: []string{"*.sh", "*.bash", ".bashrc", ".bash_profile"},
			Command:      "bash-language-server",
			Args:         []string{"start"},
			Capabilities: LSPCapabilities{
				Completion:  true,
				Hover:       true,
				Definition:  true,
				References:  true,
				Diagnostics: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// YAML
		{
			ID:           "yaml-language-server",
			Name:         "YAML Language Server",
			Language:     "yaml",
			FilePatterns: []string{"*.yaml", "*.yml"},
			Command:      "yaml-language-server",
			Args:         []string{"--stdio"},
			Capabilities: LSPCapabilities{
				Completion:  true,
				Hover:       true,
				Diagnostics: true,
				Formatting:  true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Dockerfile
		{
			ID:           "docker-langserver",
			Name:         "Dockerfile Language Server",
			Language:     "dockerfile",
			FilePatterns: []string{"Dockerfile", "Dockerfile.*", "*.dockerfile"},
			Command:      "docker-langserver",
			Args:         []string{"--stdio"},
			Capabilities: LSPCapabilities{
				Completion:  true,
				Hover:       true,
				Diagnostics: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Terraform
		{
			ID:           "terraform-ls",
			Name:         "Terraform Language Server",
			Language:     "terraform",
			FilePatterns: []string{"*.tf", "*.tfvars"},
			Command:      "terraform-ls",
			Args:         []string{"serve"},
			Capabilities: LSPCapabilities{
				Completion:  true,
				Hover:       true,
				Definition:  true,
				References:  true,
				Diagnostics: true,
				Formatting:  true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// Lua
		{
			ID:           "lua-language-server",
			Name:         "Lua Language Server",
			Language:     "lua",
			FilePatterns: []string{"*.lua"},
			Command:      "lua-language-server",
			Args:         []string{},
			Capabilities: LSPCapabilities{
				Completion:    true,
				Hover:         true,
				Definition:    true,
				References:    true,
				Diagnostics:   true,
				Rename:        true,
				CodeAction:    true,
				Formatting:    true,
				SignatureHelp: true,
			},
			Priority: 100,
			Enabled:  true,
		},
		// XML - lemminx
		{
			ID:           "lemminx",
			Name:         "LemMinX",
			Language:     "xml",
			FilePatterns: []string{"*.xml", "*.xsd", "*.xsl"},
			Command:      "lemminx",
			Args:         []string{},
			Capabilities: LSPCapabilities{
				Completion:  true,
				Hover:       true,
				Definition:  true,
				Diagnostics: true,
				Formatting:  true,
			},
			Priority: 100,
			Enabled:  true,
		},
	}

	for _, server := range defaultServers {
		serverCopy := server
		r.registerServer(&serverCopy)
	}
}

// registerServer registers a server definition.
func (r *LSPServerRegistry) registerServer(server *LSPServerDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if binary exists
	server.Binary = r.findBinary(server.Command)
	if server.Binary == "" {
		r.logger.WithFields(logrus.Fields{
			"server":  server.ID,
			"command": server.Command,
		}).Debug("LSP server binary not found")
	}

	r.servers[server.ID] = server

	// Add to language index
	if _, exists := r.byLanguage[server.Language]; !exists {
		r.byLanguage[server.Language] = make([]*LSPServerDefinition, 0)
	}
	r.byLanguage[server.Language] = append(r.byLanguage[server.Language], server)
}

// findBinary searches for a binary in the search paths.
func (r *LSPServerRegistry) findBinary(command string) string {
	// Check if it's an absolute path
	if filepath.IsAbs(command) {
		if _, err := os.Stat(command); err == nil {
			return command
		}
		return ""
	}

	// Try exec.LookPath first
	if path, err := exec.LookPath(command); err == nil {
		return path
	}

	// Search in configured paths
	for _, searchPath := range r.searchPaths {
		// Handle wildcards in path
		matches, _ := filepath.Glob(searchPath)
		for _, dir := range matches {
			candidate := filepath.Join(dir, command)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}

	return ""
}

// Get returns a server definition by ID.
func (r *LSPServerRegistry) Get(id string) (*LSPServerDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, exists := r.servers[id]
	if !exists {
		return nil, fmt.Errorf("LSP server not found: %s", id)
	}
	return server, nil
}

// GetByLanguage returns all server definitions for a language.
func (r *LSPServerRegistry) GetByLanguage(language string) []*LSPServerDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	servers := r.byLanguage[language]
	if servers == nil {
		return []*LSPServerDefinition{}
	}

	// Return only available servers (with binary found)
	available := make([]*LSPServerDefinition, 0)
	for _, s := range servers {
		if s.Binary != "" && s.Enabled {
			available = append(available, s)
		}
	}
	return available
}

// GetPreferredByLanguage returns the preferred (highest priority) server for a language.
func (r *LSPServerRegistry) GetPreferredByLanguage(language string) *LSPServerDefinition {
	servers := r.GetByLanguage(language)
	if len(servers) == 0 {
		return nil
	}

	// Find highest priority
	var preferred *LSPServerDefinition
	for _, s := range servers {
		if preferred == nil || s.Priority > preferred.Priority {
			preferred = s
		}
	}
	return preferred
}

// GetByFilePattern returns server definitions that match a file pattern.
func (r *LSPServerRegistry) GetByFilePattern(filename string) []*LSPServerDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	matches := make([]*LSPServerDefinition, 0)
	for _, server := range r.servers {
		if !server.Enabled || server.Binary == "" {
			continue
		}
		for _, pattern := range server.FilePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(filename)); matched {
				matches = append(matches, server)
				break
			}
		}
	}
	return matches
}

// List returns all registered server IDs.
func (r *LSPServerRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.servers))
	for id := range r.servers {
		ids = append(ids, id)
	}
	return ids
}

// ListAvailable returns all available (binary found) server IDs.
func (r *LSPServerRegistry) ListAvailable() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0)
	for id, server := range r.servers {
		if server.Binary != "" && server.Enabled {
			ids = append(ids, id)
		}
	}
	return ids
}

// ListLanguages returns all supported languages.
func (r *LSPServerRegistry) ListLanguages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	languages := make([]string, 0, len(r.byLanguage))
	for lang := range r.byLanguage {
		languages = append(languages, lang)
	}
	return languages
}

// Refresh re-scans for binaries.
func (r *LSPServerRegistry) Refresh() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, server := range r.servers {
		server.Binary = r.findBinary(server.Command)
	}

	r.logger.Info("LSP server registry refreshed")
}

// Register adds a custom server definition.
func (r *LSPServerRegistry) Register(server *LSPServerDefinition) error {
	if server.ID == "" {
		return fmt.Errorf("server ID is required")
	}
	if server.Command == "" {
		return fmt.Errorf("server command is required")
	}

	r.registerServer(server)
	r.logger.WithField("server", server.ID).Info("Custom LSP server registered")
	return nil
}

// Unregister removes a server definition.
func (r *LSPServerRegistry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	server, exists := r.servers[id]
	if !exists {
		return fmt.Errorf("server not found: %s", id)
	}

	// Remove from language index
	if servers, ok := r.byLanguage[server.Language]; ok {
		newServers := make([]*LSPServerDefinition, 0)
		for _, s := range servers {
			if s.ID != id {
				newServers = append(newServers, s)
			}
		}
		r.byLanguage[server.Language] = newServers
	}

	delete(r.servers, id)
	return nil
}

// DetectLanguage detects the language from a file path.
func DetectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".py", ".pyi":
		return "python"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx":
		return "cpp"
	case ".java":
		return "java"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".rb", ".rake":
		return "ruby"
	case ".ex", ".exs":
		return "elixir"
	case ".hs", ".lhs":
		return "haskell"
	case ".sh", ".bash":
		return "bash"
	case ".yaml", ".yml":
		return "yaml"
	case ".xml", ".xsd", ".xsl":
		return "xml"
	case ".tf", ".tfvars":
		return "terraform"
	case ".lua":
		return "lua"
	default:
		// Check for special filenames
		base := filepath.Base(filePath)
		switch {
		case base == "Dockerfile" || filepath.Ext(base) == ".dockerfile":
			return "dockerfile"
		case base == "Gemfile":
			return "ruby"
		case base == ".bashrc" || base == ".bash_profile":
			return "bash"
		}
		return "unknown"
	}
}

// LSPServerStatus represents the status of an LSP server.
type LSPServerStatus struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Language  string    `json:"language"`
	Available bool      `json:"available"`
	Enabled   bool      `json:"enabled"`
	Binary    string    `json:"binary,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// GetStatuses returns the status of all servers.
func (r *LSPServerRegistry) GetStatuses() []LSPServerStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	statuses := make([]LSPServerStatus, 0, len(r.servers))
	now := time.Now()

	for _, server := range r.servers {
		statuses = append(statuses, LSPServerStatus{
			ID:        server.ID,
			Name:      server.Name,
			Language:  server.Language,
			Available: server.Binary != "",
			Enabled:   server.Enabled,
			Binary:    server.Binary,
			CheckedAt: now,
		})
	}

	return statuses
}

// HealthCheck checks if all enabled servers are available.
func (r *LSPServerRegistry) HealthCheck(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)

	for id, server := range r.servers {
		if !server.Enabled {
			continue
		}

		if server.Binary == "" {
			results[id] = fmt.Errorf("binary not found: %s", server.Command)
			continue
		}

		// Verify binary is still accessible
		if _, err := os.Stat(server.Binary); err != nil {
			results[id] = fmt.Errorf("binary inaccessible: %w", err)
			continue
		}

		results[id] = nil
	}

	return results
}
