package formatters

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// FormatterRegistry manages all available formatters
type FormatterRegistry struct {
	mu         sync.RWMutex
	formatters map[string]Formatter          // name -> formatter
	byLanguage map[string][]Formatter        // language -> formatters
	metadata   map[string]*FormatterMetadata // name -> metadata
	config     *RegistryConfig
	logger     *logrus.Logger
}

// RegistryConfig configures the formatter registry
type RegistryConfig struct {
	// Paths
	SubmodulesPath string // Path to formatters/ directory
	BinariesPath   string // Path to compiled binaries
	ConfigsPath    string // Path to config files

	// Services
	ServicesComposeFile string // docker-compose.formatters.yml path
	ServicesEnabled     bool   // Enable service-based formatters

	// Behavior
	EnableCaching  bool
	CacheTTL       time.Duration
	DefaultTimeout time.Duration
	MaxConcurrent  int

	// Features
	EnableHotReload bool
	EnableMetrics   bool
	EnableTracing   bool
}

// NewFormatterRegistry creates a new formatter registry
func NewFormatterRegistry(config *RegistryConfig, logger *logrus.Logger) *FormatterRegistry {
	return &FormatterRegistry{
		formatters: make(map[string]Formatter),
		byLanguage: make(map[string][]Formatter),
		metadata:   make(map[string]*FormatterMetadata),
		config:     config,
		logger:     logger,
	}
}

// Register registers a formatter with metadata
func (r *FormatterRegistry) Register(formatter Formatter, metadata *FormatterMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := formatter.Name()

	// Check for duplicate
	if _, exists := r.formatters[name]; exists {
		return fmt.Errorf("formatter %s already registered", name)
	}

	// Register formatter
	r.formatters[name] = formatter
	r.metadata[name] = metadata

	// Register by language
	for _, lang := range formatter.Languages() {
		langLower := strings.ToLower(lang)
		r.byLanguage[langLower] = append(r.byLanguage[langLower], formatter)
	}

	r.logger.Infof("Registered formatter: %s (v%s) for languages: %v", name, formatter.Version(), formatter.Languages())

	return nil
}

// Unregister removes a formatter from the registry
func (r *FormatterRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	formatter, exists := r.formatters[name]
	if !exists {
		return fmt.Errorf("formatter %s not found", name)
	}

	// Remove from language mappings
	for _, lang := range formatter.Languages() {
		langLower := strings.ToLower(lang)
		formatters := r.byLanguage[langLower]
		for i, f := range formatters {
			if f.Name() == name {
				r.byLanguage[langLower] = append(formatters[:i], formatters[i+1:]...)
				break
			}
		}
	}

	// Remove from main registry
	delete(r.formatters, name)
	delete(r.metadata, name)

	r.logger.Infof("Unregistered formatter: %s", name)

	return nil
}

// Get retrieves a formatter by name
func (r *FormatterRegistry) Get(name string) (Formatter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	formatter, exists := r.formatters[name]
	if !exists {
		return nil, fmt.Errorf("formatter %s not found", name)
	}

	return formatter, nil
}

// GetByLanguage retrieves all formatters for a language
func (r *FormatterRegistry) GetByLanguage(language string) []Formatter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	langLower := strings.ToLower(language)
	return r.byLanguage[langLower]
}

// GetMetadata retrieves formatter metadata
func (r *FormatterRegistry) GetMetadata(name string) (*FormatterMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, exists := r.metadata[name]
	if !exists {
		return nil, fmt.Errorf("formatter %s not found", name)
	}

	return metadata, nil
}

// List returns all registered formatter names
func (r *FormatterRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.formatters))
	for name := range r.formatters {
		names = append(names, name)
	}

	return names
}

// ListByType returns all formatters of a specific type
func (r *FormatterRegistry) ListByType(ftype FormatterType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0)
	for name, metadata := range r.metadata {
		if metadata.Type == ftype {
			names = append(names, name)
		}
	}

	return names
}

// DetectFormatter detects the appropriate formatter for a file
func (r *FormatterRegistry) DetectFormatter(filePath string, content string) (Formatter, error) {
	// Detect language from file extension
	language := r.DetectLanguageFromPath(filePath)
	if language == "" {
		return nil, fmt.Errorf("unable to detect language from file path: %s", filePath)
	}

	// Get formatters for language
	formatters := r.GetByLanguage(language)
	if len(formatters) == 0 {
		return nil, fmt.Errorf("no formatters available for language: %s", language)
	}

	// Return the first (highest priority) formatter
	return formatters[0], nil
}

// DetectLanguage detects the language from file path and content
func (r *FormatterRegistry) DetectLanguage(filePath string, content string) (string, error) {
	language := r.DetectLanguageFromPath(filePath)
	if language == "" {
		return "", fmt.Errorf("unable to detect language from file path: %s", filePath)
	}

	return language, nil
}

// DetectLanguageFromPath detects language from file extension
func (r *FormatterRegistry) DetectLanguageFromPath(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext == "" {
		return ""
	}

	// Remove leading dot
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)

	// Map extensions to languages
	extensionMap := map[string]string{
		"c":          "c",
		"h":          "c",
		"cc":         "cpp",
		"cpp":        "cpp",
		"cxx":        "cpp",
		"hpp":        "cpp",
		"hxx":        "cpp",
		"rs":         "rust",
		"go":         "go",
		"py":         "python",
		"pyw":        "python",
		"js":         "javascript",
		"jsx":        "javascript",
		"ts":         "typescript",
		"tsx":        "typescript",
		"java":       "java",
		"kt":         "kotlin",
		"kts":        "kotlin",
		"scala":      "scala",
		"sc":         "scala",
		"groovy":     "groovy",
		"gvy":        "groovy",
		"gy":         "groovy",
		"gsh":        "groovy",
		"clj":        "clojure",
		"cljs":       "clojure",
		"cljc":       "clojure",
		"rb":         "ruby",
		"php":        "php",
		"swift":      "swift",
		"dart":       "dart",
		"m":          "objectivec",
		"mm":         "objectivec",
		"sh":         "bash",
		"bash":       "bash",
		"ps1":        "powershell",
		"psm1":       "powershell",
		"lua":        "lua",
		"pl":         "perl",
		"pm":         "perl",
		"r":          "r",
		"sql":        "sql",
		"yaml":       "yaml",
		"yml":        "yaml",
		"json":       "json",
		"toml":       "toml",
		"xml":        "xml",
		"html":       "html",
		"htm":        "html",
		"css":        "css",
		"scss":       "scss",
		"sass":       "sass",
		"less":       "less",
		"md":         "markdown",
		"markdown":   "markdown",
		"graphql":    "graphql",
		"gql":        "graphql",
		"proto":      "protobuf",
		"tf":         "terraform",
		"tfvars":     "terraform",
		"dockerfile": "dockerfile",
		"hs":         "haskell",
		"ml":         "ocaml",
		"mli":        "ocaml",
		"fs":         "fsharp",
		"fsx":        "fsharp",
		"ex":         "elixir",
		"exs":        "elixir",
		"erl":        "erlang",
		"hrl":        "erlang",
		"zig":        "zig",
		"nim":        "nim",
	}

	return extensionMap[ext]
}

// maxConcurrentHealthChecks limits the number of parallel health checks
const maxConcurrentHealthChecks = 10

// HealthCheckAll performs health checks on all formatters with bounded concurrency
func (r *FormatterRegistry) HealthCheckAll(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, maxConcurrentHealthChecks)

	for name, formatter := range r.formatters {
		wg.Add(1)
		go func(name string, formatter Formatter) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			err := formatter.HealthCheck(ctx)

			mu.Lock()
			results[name] = err
			mu.Unlock()

			if err != nil {
				r.logger.Warnf("Health check failed for formatter %s: %v", name, err)
			} else {
				r.logger.Debugf("Health check passed for formatter %s", name)
			}
		}(name, formatter)
	}

	wg.Wait()

	return results
}

// Start initializes the registry
func (r *FormatterRegistry) Start(ctx context.Context) error {
	r.logger.Info("Starting formatter registry")

	// Perform health checks
	results := r.HealthCheckAll(ctx)

	// Log summary
	healthy := 0
	unhealthy := 0
	for _, err := range results {
		if err == nil {
			healthy++
		} else {
			unhealthy++
		}
	}

	r.logger.Infof("Formatter registry started: %d healthy, %d unhealthy", healthy, unhealthy)

	return nil
}

// Stop shuts down the registry
func (r *FormatterRegistry) Stop(ctx context.Context) error {
	r.logger.Info("Stopping formatter registry")

	// Clear all formatters
	r.mu.Lock()
	defer r.mu.Unlock()

	r.formatters = make(map[string]Formatter)
	r.byLanguage = make(map[string][]Formatter)
	r.metadata = make(map[string]*FormatterMetadata)

	r.logger.Info("Formatter registry stopped")

	return nil
}

// Count returns the number of registered formatters
func (r *FormatterRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.formatters)
}

// CountByLanguage returns the number of formatters for a language
func (r *FormatterRegistry) CountByLanguage(language string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	langLower := strings.ToLower(language)
	return len(r.byLanguage[langLower])
}

// GetPreferredFormatter returns the preferred formatter for a language
func (r *FormatterRegistry) GetPreferredFormatter(language string, preferences map[string]string) (Formatter, error) {
	// Check if there's a preference
	if preferences != nil {
		if preferred, ok := preferences[strings.ToLower(language)]; ok {
			return r.Get(preferred)
		}
	}

	// Fall back to first formatter for language
	formatters := r.GetByLanguage(language)
	if len(formatters) == 0 {
		return nil, fmt.Errorf("no formatters available for language: %s", language)
	}

	return formatters[0], nil
}
