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

// LazyFormatterFunc is a function that creates a Formatter on demand.
// It is called at most once per formatter, on first access.
type LazyFormatterFunc func() (Formatter, error)

// lazyFormatter holds a factory for deferred formatter initialization.
type lazyFormatter struct {
	factory  LazyFormatterFunc
	once     sync.Once
	result   Formatter
	initErr  error
}

// get returns the initialized formatter, calling the factory on first access.
func (lf *lazyFormatter) get() (Formatter, error) {
	lf.once.Do(func() {
		lf.result, lf.initErr = lf.factory()
	})
	return lf.result, lf.initErr
}

// FormatterRegistry manages all available formatters
type FormatterRegistry struct {
	mu             sync.RWMutex
	formatters     map[string]Formatter          // name -> formatter (eager)
	lazyFormatters map[string]*lazyFormatter     // name -> lazy formatter
	byLanguage     map[string][]Formatter        // language -> formatters (eager only)
	lazyByLanguage map[string][]string           // language -> lazy formatter names
	metadata       map[string]*FormatterMetadata // name -> metadata
	config         *RegistryConfig
	logger         *logrus.Logger
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
		formatters:     make(map[string]Formatter),
		lazyFormatters: make(map[string]*lazyFormatter),
		byLanguage:     make(map[string][]Formatter),
		lazyByLanguage: make(map[string][]string),
		metadata:       make(map[string]*FormatterMetadata),
		config:         config,
		logger:         logger,
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

// RegisterLazy registers a formatter factory for deferred initialization.
// The factory is called at most once, when the formatter is first accessed
// via Get or GetByLanguage. The metadata must include Languages so that
// language-based lookups can discover the lazy formatter.
func (r *FormatterRegistry) RegisterLazy(factory LazyFormatterFunc, metadata *FormatterMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := metadata.Name

	// Check for duplicate in both eager and lazy registrations
	if _, exists := r.formatters[name]; exists {
		return fmt.Errorf("formatter %s already registered", name)
	}
	if _, exists := r.lazyFormatters[name]; exists {
		return fmt.Errorf("formatter %s already registered (lazy)", name)
	}

	r.lazyFormatters[name] = &lazyFormatter{factory: factory}
	r.metadata[name] = metadata

	// Register by language for lazy lookup
	for _, lang := range metadata.Languages {
		langLower := strings.ToLower(lang)
		r.lazyByLanguage[langLower] = append(r.lazyByLanguage[langLower], name)
	}

	r.logger.Infof("Registered lazy formatter: %s (v%s) for languages: %v",
		name, metadata.Version, metadata.Languages)

	return nil
}

// Unregister removes a formatter from the registry
func (r *FormatterRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	formatter, eagerExists := r.formatters[name]
	_, lazyExists := r.lazyFormatters[name]

	if !eagerExists && !lazyExists {
		return fmt.Errorf("formatter %s not found", name)
	}

	// Remove eager language mappings
	if eagerExists {
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
	}

	// Remove lazy language mappings
	if lazyExists {
		meta := r.metadata[name]
		if meta != nil {
			for _, lang := range meta.Languages {
				langLower := strings.ToLower(lang)
				lazyNames := r.lazyByLanguage[langLower]
				for i, n := range lazyNames {
					if n == name {
						r.lazyByLanguage[langLower] = append(lazyNames[:i], lazyNames[i+1:]...)
						break
					}
				}
			}
		}
	}

	// Remove from all registries
	delete(r.formatters, name)
	delete(r.lazyFormatters, name)
	delete(r.metadata, name)

	r.logger.Infof("Unregistered formatter: %s", name)

	return nil
}

// Get retrieves a formatter by name. For lazily registered formatters,
// this triggers initialization on the first call.
func (r *FormatterRegistry) Get(name string) (Formatter, error) {
	r.mu.RLock()
	// Check eagerly registered formatters first
	if formatter, exists := r.formatters[name]; exists {
		r.mu.RUnlock()
		return formatter, nil
	}
	// Check lazily registered formatters
	lf, lazyExists := r.lazyFormatters[name]
	r.mu.RUnlock()

	if !lazyExists {
		return nil, fmt.Errorf("formatter %s not found", name)
	}

	formatter, err := lf.get()
	if err != nil {
		return nil, fmt.Errorf("lazy initialization of formatter %s failed: %w", name, err)
	}
	return formatter, nil
}

// GetByLanguage retrieves all formatters for a language.
// Lazy formatters are initialized on access.
func (r *FormatterRegistry) GetByLanguage(language string) []Formatter {
	r.mu.RLock()
	langLower := strings.ToLower(language)
	eagerFormatters := r.byLanguage[langLower]
	lazyNames := r.lazyByLanguage[langLower]
	r.mu.RUnlock()

	// Start with eager formatters
	result := make([]Formatter, 0, len(eagerFormatters)+len(lazyNames))
	result = append(result, eagerFormatters...)

	// Initialize and append lazy formatters
	for _, name := range lazyNames {
		r.mu.RLock()
		lf, ok := r.lazyFormatters[name]
		r.mu.RUnlock()
		if !ok {
			continue
		}
		formatter, err := lf.get()
		if err != nil {
			r.logger.Warnf("Lazy initialization of formatter %s failed: %v", name, err)
			continue
		}
		result = append(result, formatter)
	}

	return result
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

// List returns all registered formatter names (both eager and lazy)
func (r *FormatterRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]struct{}, len(r.formatters)+len(r.lazyFormatters))
	names := make([]string, 0, len(r.formatters)+len(r.lazyFormatters))
	for name := range r.formatters {
		names = append(names, name)
		seen[name] = struct{}{}
	}
	for name := range r.lazyFormatters {
		if _, ok := seen[name]; !ok {
			names = append(names, name)
		}
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

// Count returns the number of registered formatters (both eager and lazy)
func (r *FormatterRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Count unique names across both maps
	seen := make(map[string]struct{}, len(r.formatters)+len(r.lazyFormatters))
	for name := range r.formatters {
		seen[name] = struct{}{}
	}
	for name := range r.lazyFormatters {
		seen[name] = struct{}{}
	}
	return len(seen)
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
