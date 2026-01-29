package formatters

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete formatters configuration
type Config struct {
	// Global settings
	Enabled          bool `yaml:"enabled"`
	AutoFormat       bool `yaml:"auto_format"`
	FormatOnSave     bool `yaml:"format_on_save"`
	FormatOnDebate   bool `yaml:"format_on_debate"`

	// Paths
	SubmodulesPath string `yaml:"submodules_path"`
	BinariesPath   string `yaml:"binaries_path"`
	ConfigsPath    string `yaml:"configs_path"`

	// Services
	ServicesComposeFile string `yaml:"services_compose_file"`
	ServicesEnabled     bool   `yaml:"services_enabled"`

	// Performance
	CacheEnabled   bool          `yaml:"cache_enabled"`
	CacheTTL       time.Duration `yaml:"cache_ttl"`
	DefaultTimeout time.Duration `yaml:"default_timeout"`
	MaxConcurrent  int           `yaml:"max_concurrent"`

	// Features
	HotReload bool `yaml:"hot_reload"`
	Metrics   bool `yaml:"metrics"`
	Tracing   bool `yaml:"tracing"`

	// Defaults
	DefaultLineLength int  `yaml:"default_line_length"`
	DefaultIndentSize int  `yaml:"default_indent_size"`
	UseTabs           bool `yaml:"use_tabs"`

	// Preferences (language -> formatter name)
	Preferences map[string]string `yaml:"preferences"`

	// Fallback chains (language -> []formatter names)
	Fallback map[string][]string `yaml:"fallback"`

	// Language-specific configs
	LanguageConfigs map[string]LanguageConfig `yaml:"language_configs"`

	// Overrides (pattern-based overrides)
	Overrides []OverrideConfig `yaml:"overrides"`
}

// LanguageConfig represents language-specific configuration
type LanguageConfig struct {
	LineLength               int                    `yaml:"line_length"`
	IndentSize               int                    `yaml:"indent_size"`
	UseTabs                  bool                   `yaml:"use_tabs"`
	FormatterSpecificOptions map[string]interface{} `yaml:"formatter_specific_options"`
}

// OverrideConfig represents pattern-based configuration overrides
type OverrideConfig struct {
	Pattern    string                 `yaml:"pattern"`
	Formatter  string                 `yaml:"formatter"`
	Config     map[string]interface{} `yaml:"config"`
	LineLength int                    `yaml:"line_length"`
}

// DefaultConfig returns the default formatters configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:          true,
		AutoFormat:       true,
		FormatOnSave:     true,
		FormatOnDebate:   true,
		SubmodulesPath:   "./formatters",
		BinariesPath:     "./bin/formatters",
		ConfigsPath:      "./configs/formatters",
		ServicesComposeFile: "./docker/formatters/docker-compose.formatters.yml",
		ServicesEnabled:  true,
		CacheEnabled:     true,
		CacheTTL:         3600 * time.Second,
		DefaultTimeout:   30 * time.Second,
		MaxConcurrent:    10,
		HotReload:        true,
		Metrics:          true,
		Tracing:          true,
		DefaultLineLength: 88,
		DefaultIndentSize: 4,
		UseTabs:          false,
		Preferences: map[string]string{
			"python":     "ruff",
			"javascript": "biome",
			"typescript": "biome",
			"rust":       "rustfmt",
			"go":         "gofmt",
			"c":          "clang-format",
			"cpp":        "clang-format",
			"java":       "google-java-format",
			"kotlin":     "ktlint",
			"scala":      "scalafmt",
			"swift":      "swift-format",
			"dart":       "dart_format",
			"ruby":       "rubocop",
			"php":        "php-cs-fixer",
			"elixir":     "mix_format",
			"haskell":    "ormolu",
			"ocaml":      "ocamlformat",
			"fsharp":     "fantomas",
			"clojure":    "cljfmt",
			"erlang":     "erlfmt",
			"bash":       "shfmt",
			"powershell": "psscriptanalyzer",
			"lua":        "stylua",
			"perl":       "perltidy",
			"r":          "air",
			"sql":        "sqlfluff",
			"yaml":       "yamlfmt",
			"json":       "jq",
			"toml":       "taplo",
			"xml":        "xmllint",
			"html":       "prettier",
			"css":        "prettier",
			"scss":       "prettier",
			"markdown":   "prettier",
			"graphql":    "prettier",
			"protobuf":   "buf",
			"terraform":  "terraform_fmt",
			"dockerfile": "hadolint",
		},
		Fallback: map[string][]string{
			"python":     {"black", "autopep8"},
			"javascript": {"prettier", "dprint"},
			"typescript": {"prettier", "dprint"},
			"ruby":       {"standardrb", "rufo"},
			"java":       {"spotless"},
			"kotlin":     {"ktfmt"},
			"css":        {"stylelint"},
		},
		LanguageConfigs: make(map[string]LanguageConfig),
		Overrides:       make([]OverrideConfig, 0),
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Resolve relative paths
	if err := config.resolvePaths(filepath.Dir(path)); err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// resolvePaths resolves relative paths in the configuration
func (c *Config) resolvePaths(baseDir string) error {
	// Resolve SubmodulesPath
	if !filepath.IsAbs(c.SubmodulesPath) {
		c.SubmodulesPath = filepath.Join(baseDir, c.SubmodulesPath)
	}

	// Resolve BinariesPath
	if !filepath.IsAbs(c.BinariesPath) {
		c.BinariesPath = filepath.Join(baseDir, c.BinariesPath)
	}

	// Resolve ConfigsPath
	if !filepath.IsAbs(c.ConfigsPath) {
		c.ConfigsPath = filepath.Join(baseDir, c.ConfigsPath)
	}

	// Resolve ServicesComposeFile
	if !filepath.IsAbs(c.ServicesComposeFile) {
		c.ServicesComposeFile = filepath.Join(baseDir, c.ServicesComposeFile)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultTimeout <= 0 {
		return fmt.Errorf("default_timeout must be positive")
	}

	if c.MaxConcurrent <= 0 {
		return fmt.Errorf("max_concurrent must be positive")
	}

	if c.DefaultLineLength <= 0 {
		return fmt.Errorf("default_line_length must be positive")
	}

	if c.DefaultIndentSize <= 0 {
		return fmt.Errorf("default_indent_size must be positive")
	}

	return nil
}

// GetPreferredFormatter returns the preferred formatter for a language
func (c *Config) GetPreferredFormatter(language string) (string, bool) {
	formatter, ok := c.Preferences[language]
	return formatter, ok
}

// GetFallbackChain returns the fallback chain for a language
func (c *Config) GetFallbackChain(language string) []string {
	return c.Fallback[language]
}

// GetLanguageConfig returns the language-specific configuration
func (c *Config) GetLanguageConfig(language string) (LanguageConfig, bool) {
	config, ok := c.LanguageConfigs[language]
	return config, ok
}

// ApplyOverrides applies pattern-based overrides to a request
func (c *Config) ApplyOverrides(req *FormatRequest) {
	for _, override := range c.Overrides {
		if matchesPattern(req.FilePath, override.Pattern) {
			if override.LineLength > 0 {
				req.LineLength = override.LineLength
			}

			if override.Config != nil {
				if req.Config == nil {
					req.Config = make(map[string]interface{})
				}
				for k, v := range override.Config {
					req.Config[k] = v
				}
			}
		}
	}
}

// matchesPattern checks if a file path matches a pattern
func matchesPattern(filePath, pattern string) bool {
	matched, err := filepath.Match(pattern, filepath.Base(filePath))
	if err != nil {
		return false
	}
	return matched
}

// ToRegistryConfig converts to RegistryConfig
func (c *Config) ToRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		SubmodulesPath:      c.SubmodulesPath,
		BinariesPath:        c.BinariesPath,
		ConfigsPath:         c.ConfigsPath,
		ServicesComposeFile: c.ServicesComposeFile,
		ServicesEnabled:     c.ServicesEnabled,
		EnableCaching:       c.CacheEnabled,
		CacheTTL:            c.CacheTTL,
		DefaultTimeout:      c.DefaultTimeout,
		MaxConcurrent:       c.MaxConcurrent,
		EnableHotReload:     c.HotReload,
		EnableMetrics:       c.Metrics,
		EnableTracing:       c.Tracing,
	}
}

// ToExecutorConfig converts to ExecutorConfig
func (c *Config) ToExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		DefaultTimeout: c.DefaultTimeout,
		MaxRetries:     3,
		EnableCache:    c.CacheEnabled,
		EnableMetrics:  c.Metrics,
		EnableTracing:  c.Tracing,
	}
}

// ToCacheConfig converts to CacheConfig
func (c *Config) ToCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL:         c.CacheTTL,
		MaxSize:     10000,
		CleanupFreq: 300 * time.Second,
	}
}
