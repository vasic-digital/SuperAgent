package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// ConstrainedGenerator provides constrained decoding for structured output
// Inspired by XGrammar for high-performance grammar-based generation
type ConstrainedGenerator struct {
	validator    *SchemaValidator
	config       *GeneratorConfig
	schemaCache  map[string]*Schema
	grammarCache map[string]*Grammar
	logger       *logrus.Logger
	mu           sync.RWMutex
}

// GeneratorConfig configures the constrained generator
type GeneratorConfig struct {
	// Enable schema validation
	EnableValidation bool `json:"enable_validation"`
	// Enable automatic repair
	EnableRepair bool `json:"enable_repair"`
	// Maximum repair attempts
	MaxRepairAttempts int `json:"max_repair_attempts"`
	// Strict mode for validation
	StrictMode bool `json:"strict_mode"`
	// Enable caching
	EnableCaching bool `json:"enable_caching"`
	// Default output format
	DefaultFormat OutputFormat `json:"default_format"`
}

// DefaultGeneratorConfig returns default configuration
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		EnableValidation:  true,
		EnableRepair:      true,
		MaxRepairAttempts: 3,
		StrictMode:        false,
		EnableCaching:     true,
		DefaultFormat:     OutputFormatJSON,
	}
}

// NewConstrainedGenerator creates a new constrained generator
func NewConstrainedGenerator(config *GeneratorConfig, logger *logrus.Logger) *ConstrainedGenerator {
	if config == nil {
		config = DefaultGeneratorConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &ConstrainedGenerator{
		validator:    NewSchemaValidator(config.StrictMode),
		config:       config,
		schemaCache:  make(map[string]*Schema),
		grammarCache: make(map[string]*Grammar),
		logger:       logger,
	}
}

// GenerationRequest represents a request for structured generation
type GenerationRequest struct {
	// The schema to generate output for
	Schema *Schema `json:"schema"`
	// The prompt for generation
	Prompt string `json:"prompt"`
	// The raw LLM response
	Response string `json:"response"`
	// Target type for deserialization
	Target interface{} `json:"-"`
	// Output format
	Format OutputFormat `json:"format"`
}

// GenerationResult contains the result of constrained generation
type GenerationResult struct {
	// The validated/repaired output
	Output string `json:"output"`
	// Parsed data (if JSON)
	Data interface{} `json:"data,omitempty"`
	// Validation result
	Validation *ValidationResult `json:"validation"`
	// Whether output was repaired
	Repaired bool `json:"repaired"`
	// Number of repair attempts
	RepairAttempts int `json:"repair_attempts"`
	// Grammar used (if any)
	Grammar *Grammar `json:"grammar,omitempty"`
}

// Generate validates and potentially repairs LLM output
func (g *ConstrainedGenerator) Generate(ctx context.Context, req *GenerationRequest) (*GenerationResult, error) {
	result := &GenerationResult{
		Output: req.Response,
	}

	if req.Schema == nil {
		// No schema, just return the response as-is
		return result, nil
	}

	// Validate the response
	if g.config.EnableValidation {
		validation, err := g.validator.Validate(req.Response, req.Schema)
		if err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
		result.Validation = validation
		result.Data = validation.Data

		if !validation.Valid && g.config.EnableRepair {
			// Attempt to repair
			repaired, attempts, err := g.repair(req.Response, req.Schema)
			if err != nil {
				g.logger.WithError(err).Debug("Repair failed")
			} else {
				result.Output = repaired
				result.Repaired = true
				result.RepairAttempts = attempts

				// Re-validate repaired output
				validation, _ = g.validator.Validate(repaired, req.Schema) //nolint:errcheck
				result.Validation = validation
				result.Data = validation.Data
			}
		}
	}

	// Unmarshal to target if provided
	if req.Target != nil && result.Validation != nil && result.Validation.Valid {
		if err := json.Unmarshal([]byte(result.Output), req.Target); err != nil {
			g.logger.WithError(err).Warn("Failed to unmarshal to target")
		}
	}

	return result, nil
}

// repair attempts to repair invalid output
func (g *ConstrainedGenerator) repair(output string, schema *Schema) (string, int, error) {
	var lastErr error
	for attempt := 1; attempt <= g.config.MaxRepairAttempts; attempt++ {
		repaired, err := g.validator.Repair(output, schema)
		if err == nil {
			return repaired, attempt, nil
		}
		lastErr = err
		output = repaired // Use partially repaired output for next attempt
	}
	return "", g.config.MaxRepairAttempts, lastErr
}

// CreatePromptWithSchema creates a prompt that instructs the LLM to output structured data
func (g *ConstrainedGenerator) CreatePromptWithSchema(basePrompt string, schema *Schema) string {
	var sb strings.Builder

	sb.WriteString(basePrompt)
	sb.WriteString("\n\n")
	sb.WriteString("Please respond with a JSON object that follows this schema:\n\n")
	sb.WriteString("```json\n")

	schemaJSON, _ := json.MarshalIndent(schema, "", "  ") //nolint:errcheck
	sb.WriteString(string(schemaJSON))

	sb.WriteString("\n```\n\n")
	sb.WriteString("Important: Your response must be valid JSON that strictly adheres to the schema above.")

	return sb.String()
}

// CreateFunctionCallingPrompt creates a prompt for function calling style output
func (g *ConstrainedGenerator) CreateFunctionCallingPrompt(basePrompt string, functions []FunctionDef) string {
	var sb strings.Builder

	sb.WriteString(basePrompt)
	sb.WriteString("\n\n")
	sb.WriteString("You have access to the following functions:\n\n")

	for _, fn := range functions {
		sb.WriteString(fmt.Sprintf("### %s\n", fn.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", fn.Description))
		sb.WriteString("Parameters:\n")
		paramJSON, _ := json.MarshalIndent(fn.Parameters, "", "  ") //nolint:errcheck
		sb.WriteString(string(paramJSON))
		sb.WriteString("\n\n")
	}

	sb.WriteString("To call a function, respond with JSON in this format:\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{"function": "function_name", "arguments": {...}}`)
	sb.WriteString("\n```\n")

	return sb.String()
}

// FunctionDef represents a function definition
type FunctionDef struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Parameters  *Schema `json:"parameters"`
}

// ParseFunctionCall parses a function call from LLM output
func (g *ConstrainedGenerator) ParseFunctionCall(output string) (*FunctionCall, error) {
	// Extract JSON from output
	output = strings.TrimSpace(output)

	// Try to extract from code blocks
	if strings.Contains(output, "```") {
		re := strings.NewReplacer("```json\n", "", "```\n", "", "\n```", "")
		output = re.Replace(output)
	}

	var call FunctionCall
	if err := json.Unmarshal([]byte(output), &call); err != nil {
		return nil, fmt.Errorf("failed to parse function call: %w", err)
	}

	return &call, nil
}

// FunctionCall represents a parsed function call
type FunctionCall struct {
	Function  string                 `json:"function"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CacheSchema caches a schema for reuse
func (g *ConstrainedGenerator) CacheSchema(name string, schema *Schema) {
	if !g.config.EnableCaching {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.schemaCache[name] = schema
}

// GetCachedSchema retrieves a cached schema
func (g *ConstrainedGenerator) GetCachedSchema(name string) *Schema {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.schemaCache[name]
}

// CacheGrammar caches a grammar for reuse
func (g *ConstrainedGenerator) CacheGrammar(name string, grammar *Grammar) {
	if !g.config.EnableCaching {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.grammarCache[name] = grammar
}

// GetCachedGrammar retrieves a cached grammar
func (g *ConstrainedGenerator) GetCachedGrammar(name string) *Grammar {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.grammarCache[name]
}

// OutputFormatter formats validated data into different output formats
type OutputFormatter struct {
	config *FormatterConfig
}

// FormatterConfig configures output formatting
type FormatterConfig struct {
	IndentJSON bool `json:"indent_json"`
	IndentSize int  `json:"indent_size"`
	SortKeys   bool `json:"sort_keys"`
	EscapeHTML bool `json:"escape_html"`
}

// DefaultFormatterConfig returns default formatter config
func DefaultFormatterConfig() *FormatterConfig {
	return &FormatterConfig{
		IndentJSON: true,
		IndentSize: 2,
		SortKeys:   false,
		EscapeHTML: false,
	}
}

// NewOutputFormatter creates a new output formatter
func NewOutputFormatter(config *FormatterConfig) *OutputFormatter {
	if config == nil {
		config = DefaultFormatterConfig()
	}
	return &OutputFormatter{config: config}
}

// FormatJSON formats data as JSON
func (f *OutputFormatter) FormatJSON(data interface{}) (string, error) {
	var output []byte
	var err error

	if f.config.IndentJSON {
		indent := strings.Repeat(" ", f.config.IndentSize)
		output, err = json.MarshalIndent(data, "", indent)
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return "", err
	}

	return string(output), nil
}

// FormatJSONLines formats data as JSON Lines
func (f *OutputFormatter) FormatJSONLines(data []interface{}) (string, error) {
	var sb strings.Builder

	for _, item := range data {
		line, err := json.Marshal(item)
		if err != nil {
			return "", err
		}
		sb.WriteString(string(line))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// FormatMarkdown formats data as Markdown
func (f *OutputFormatter) FormatMarkdown(data interface{}) (string, error) {
	var sb strings.Builder

	switch v := data.(type) {
	case map[string]interface{}:
		f.formatMapAsMarkdown(&sb, v, 0)
	case []interface{}:
		f.formatArrayAsMarkdown(&sb, v)
	default:
		sb.WriteString(fmt.Sprintf("%v", v))
	}

	return sb.String(), nil
}

func (f *OutputFormatter) formatMapAsMarkdown(sb *strings.Builder, m map[string]interface{}, depth int) {
	for key, value := range m {
		prefix := strings.Repeat("  ", depth)
		switch v := value.(type) {
		case map[string]interface{}:
			sb.WriteString(fmt.Sprintf("%s- **%s**:\n", prefix, key))
			f.formatMapAsMarkdown(sb, v, depth+1)
		case []interface{}:
			sb.WriteString(fmt.Sprintf("%s- **%s**:\n", prefix, key))
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					f.formatMapAsMarkdown(sb, m, depth+1)
				} else {
					sb.WriteString(fmt.Sprintf("%s  - %v\n", prefix, item))
				}
			}
		default:
			sb.WriteString(fmt.Sprintf("%s- **%s**: %v\n", prefix, key, value))
		}
	}
}

func (f *OutputFormatter) formatArrayAsMarkdown(sb *strings.Builder, arr []interface{}) {
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			f.formatMapAsMarkdown(sb, m, 0)
			sb.WriteString("\n")
		} else {
			sb.WriteString(fmt.Sprintf("- %v\n", item))
		}
	}
}

// FormatCSV formats data as CSV
func (f *OutputFormatter) FormatCSV(data []map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var sb strings.Builder

	// Get headers from first item
	headers := make([]string, 0)
	for key := range data[0] {
		headers = append(headers, key)
	}

	// Write header row
	sb.WriteString(strings.Join(headers, ","))
	sb.WriteString("\n")

	// Write data rows
	for _, row := range data {
		values := make([]string, len(headers))
		for i, header := range headers {
			if val, exists := row[header]; exists {
				values[i] = f.escapeCSVValue(fmt.Sprintf("%v", val))
			}
		}
		sb.WriteString(strings.Join(values, ","))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (f *OutputFormatter) escapeCSVValue(value string) string {
	if strings.ContainsAny(value, ",\"\n") {
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(value, `"`, `""`))
	}
	return value
}
