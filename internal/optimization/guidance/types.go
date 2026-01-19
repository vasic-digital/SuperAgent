// Package guidance provides constrained generation capabilities for LLM outputs.
package guidance

import (
	"time"
)

// PromptTemplate represents a template for guided generation.
type PromptTemplate struct {
	// Name is the template name.
	Name string `json:"name"`
	// Description describes the template.
	Description string `json:"description,omitempty"`
	// Template is the template text with placeholders.
	Template string `json:"template"`
	// Variables define the template variables.
	Variables []TemplateVariable `json:"variables"`
	// OutputConstraint is the constraint for the generated output.
	OutputConstraint Constraint `json:"output_constraint,omitempty"`
	// Metadata contains additional template metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TemplateVariable defines a variable in a prompt template.
type TemplateVariable struct {
	// Name is the variable name (used in template as {{name}}).
	Name string `json:"name"`
	// Description describes the variable.
	Description string `json:"description,omitempty"`
	// Required indicates if the variable must be provided.
	Required bool `json:"required"`
	// Default is the default value if not provided.
	Default interface{} `json:"default,omitempty"`
	// Type is the variable type.
	Type VariableType `json:"type"`
	// Constraint is an optional constraint for the variable value.
	Constraint Constraint `json:"constraint,omitempty"`
}

// VariableType defines types of template variables.
type VariableType string

const (
	// VariableTypeString is a string variable.
	VariableTypeString VariableType = "string"
	// VariableTypeNumber is a numeric variable.
	VariableTypeNumber VariableType = "number"
	// VariableTypeBoolean is a boolean variable.
	VariableTypeBoolean VariableType = "boolean"
	// VariableTypeList is a list variable.
	VariableTypeList VariableType = "list"
	// VariableTypeObject is an object/map variable.
	VariableTypeObject VariableType = "object"
)

// GenerationConfig holds configuration for a generation request.
type GenerationConfig struct {
	// Model is the model to use.
	Model string `json:"model,omitempty"`
	// Provider is the provider to use.
	Provider string `json:"provider,omitempty"`
	// Temperature controls randomness (0-2).
	Temperature float64 `json:"temperature,omitempty"`
	// MaxTokens is the maximum tokens to generate.
	MaxTokens int `json:"max_tokens,omitempty"`
	// TopP is nucleus sampling probability.
	TopP float64 `json:"top_p,omitempty"`
	// TopK is the number of top tokens to consider.
	TopK int `json:"top_k,omitempty"`
	// StopSequences are sequences that stop generation.
	StopSequences []string `json:"stop_sequences,omitempty"`
	// FrequencyPenalty penalizes frequent tokens.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	// PresencePenalty penalizes tokens that have appeared.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
	// Seed for reproducibility.
	Seed int64 `json:"seed,omitempty"`
}

// DefaultGenerationConfig returns a default generation configuration.
func DefaultGenerationConfig() *GenerationConfig {
	return &GenerationConfig{
		Temperature: 0.7,
		MaxTokens:   500,
		TopP:        1.0,
	}
}

// ValidationResult contains the result of constraint validation.
type ValidationResult struct {
	// Valid indicates if validation passed.
	Valid bool `json:"valid"`
	// Errors contains validation error messages.
	Errors []string `json:"errors,omitempty"`
	// Warnings contains non-fatal warnings.
	Warnings []string `json:"warnings,omitempty"`
	// Details contains detailed validation information.
	Details map[string]interface{} `json:"details,omitempty"`
}

// GuidanceSession represents a generation session.
type GuidanceSession struct {
	// ID is the session ID.
	ID string `json:"id"`
	// StartedAt is when the session started.
	StartedAt time.Time `json:"started_at"`
	// Config is the session configuration.
	Config *GenerationConfig `json:"config"`
	// History contains the generation history.
	History []GenerationRecord `json:"history"`
	// Context contains session context.
	Context map[string]interface{} `json:"context,omitempty"`
}

// GenerationRecord records a single generation in a session.
type GenerationRecord struct {
	// Timestamp is when the generation occurred.
	Timestamp time.Time `json:"timestamp"`
	// Prompt is the input prompt.
	Prompt string `json:"prompt"`
	// Output is the generated output.
	Output string `json:"output"`
	// Constraints are the constraints used.
	Constraints []string `json:"constraints,omitempty"`
	// Valid indicates if the output was valid.
	Valid bool `json:"valid"`
	// Attempts is the number of attempts.
	Attempts int `json:"attempts"`
	// LatencyMs is the generation latency.
	LatencyMs int64 `json:"latency_ms"`
}

// OutputMode defines how output should be structured.
type OutputMode string

const (
	// OutputModeText generates plain text.
	OutputModeText OutputMode = "text"
	// OutputModeJSON generates JSON.
	OutputModeJSON OutputMode = "json"
	// OutputModeXML generates XML.
	OutputModeXML OutputMode = "xml"
	// OutputModeYAML generates YAML.
	OutputModeYAML OutputMode = "yaml"
	// OutputModeMarkdown generates Markdown.
	OutputModeMarkdown OutputMode = "markdown"
	// OutputModeCode generates code.
	OutputModeCode OutputMode = "code"
)

// OutputSpec specifies the expected output structure.
type OutputSpec struct {
	// Mode is the output mode.
	Mode OutputMode `json:"mode"`
	// Schema is the schema for structured output (JSON/XML/YAML).
	Schema map[string]interface{} `json:"schema,omitempty"`
	// Language is the language for code output.
	Language string `json:"language,omitempty"`
	// Examples are example outputs.
	Examples []string `json:"examples,omitempty"`
}

// GuidanceError represents an error during guided generation.
type GuidanceError struct {
	// Code is the error code.
	Code ErrorCode `json:"code"`
	// Message is the error message.
	Message string `json:"message"`
	// Details contains additional error details.
	Details map[string]interface{} `json:"details,omitempty"`
	// Recoverable indicates if the error can be recovered from.
	Recoverable bool `json:"recoverable"`
}

// Error implements the error interface.
func (e *GuidanceError) Error() string {
	return e.Message
}

// ErrorCode defines guidance error codes.
type ErrorCode string

const (
	// ErrorCodeConstraintViolation indicates a constraint was violated.
	ErrorCodeConstraintViolation ErrorCode = "constraint_violation"
	// ErrorCodeInvalidInput indicates invalid input.
	ErrorCodeInvalidInput ErrorCode = "invalid_input"
	// ErrorCodeGenerationFailed indicates generation failed.
	ErrorCodeGenerationFailed ErrorCode = "generation_failed"
	// ErrorCodeTimeout indicates a timeout.
	ErrorCodeTimeout ErrorCode = "timeout"
	// ErrorCodeRetryExhausted indicates retries were exhausted.
	ErrorCodeRetryExhausted ErrorCode = "retry_exhausted"
	// ErrorCodeBackendError indicates a backend error.
	ErrorCodeBackendError ErrorCode = "backend_error"
)

// GuidanceMetrics contains metrics for guidance operations.
type GuidanceMetrics struct {
	// TotalGenerations is the total number of generations.
	TotalGenerations int64 `json:"total_generations"`
	// SuccessfulGenerations is the number of successful generations.
	SuccessfulGenerations int64 `json:"successful_generations"`
	// FailedGenerations is the number of failed generations.
	FailedGenerations int64 `json:"failed_generations"`
	// TotalRetries is the total number of retries.
	TotalRetries int64 `json:"total_retries"`
	// AverageAttempts is the average attempts per generation.
	AverageAttempts float64 `json:"average_attempts"`
	// AverageLatencyMs is the average generation latency.
	AverageLatencyMs float64 `json:"average_latency_ms"`
	// ConstraintViolationRate is the rate of constraint violations.
	ConstraintViolationRate float64 `json:"constraint_violation_rate"`
	// LastUpdated is when metrics were last updated.
	LastUpdated time.Time `json:"last_updated"`
}

// ConstraintSet is a named set of constraints.
type ConstraintSet struct {
	// Name is the set name.
	Name string `json:"name"`
	// Description describes the constraint set.
	Description string `json:"description,omitempty"`
	// Constraints are the constraints in the set.
	Constraints []Constraint `json:"constraints"`
	// Mode is how constraints are combined.
	Mode CompositeMode `json:"mode"`
}

// NewConstraintSet creates a new constraint set.
func NewConstraintSet(name string, mode CompositeMode, constraints ...Constraint) *ConstraintSet {
	return &ConstraintSet{
		Name:        name,
		Constraints: constraints,
		Mode:        mode,
	}
}

// ToConstraint converts the set to a single composite constraint.
func (s *ConstraintSet) ToConstraint() Constraint {
	return NewCompositeConstraint(s.Mode, s.Constraints...)
}

// ValidationContext provides context for validation.
type ValidationContext struct {
	// Strict enables strict validation.
	Strict bool `json:"strict"`
	// AllowPartial allows partial matches.
	AllowPartial bool `json:"allow_partial"`
	// TrimWhitespace trims whitespace before validation.
	TrimWhitespace bool `json:"trim_whitespace"`
	// CaseInsensitive enables case-insensitive validation where applicable.
	CaseInsensitive bool `json:"case_insensitive"`
}

// DefaultValidationContext returns a default validation context.
func DefaultValidationContext() *ValidationContext {
	return &ValidationContext{
		Strict:         true,
		TrimWhitespace: true,
	}
}

// RetryStrategy defines how retries are handled.
type RetryStrategy struct {
	// MaxRetries is the maximum number of retries.
	MaxRetries int `json:"max_retries"`
	// InitialDelay is the initial delay between retries.
	InitialDelay time.Duration `json:"initial_delay"`
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration `json:"max_delay"`
	// BackoffMultiplier multiplies delay on each retry.
	BackoffMultiplier float64 `json:"backoff_multiplier"`
	// JitterFactor adds randomness to delays (0-1).
	JitterFactor float64 `json:"jitter_factor"`
}

// DefaultRetryStrategy returns a default retry strategy.
func DefaultRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		BackoffMultiplier: 2.0,
		JitterFactor:      0.1,
	}
}

// GetDelay calculates the delay for a given attempt.
func (s *RetryStrategy) GetDelay(attempt int) time.Duration {
	delay := float64(s.InitialDelay) * pow(s.BackoffMultiplier, float64(attempt-1))
	if time.Duration(delay) > s.MaxDelay {
		delay = float64(s.MaxDelay)
	}
	return time.Duration(delay)
}

func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

// GuidanceCapabilities describes the capabilities of a guidance implementation.
type GuidanceCapabilities struct {
	// SupportsRegex indicates regex constraints are supported.
	SupportsRegex bool `json:"supports_regex"`
	// SupportsGrammar indicates grammar constraints are supported.
	SupportsGrammar bool `json:"supports_grammar"`
	// SupportsSchema indicates JSON schema constraints are supported.
	SupportsSchema bool `json:"supports_schema"`
	// SupportsChoice indicates choice constraints are supported.
	SupportsChoice bool `json:"supports_choice"`
	// SupportsStreaming indicates streaming generation is supported.
	SupportsStreaming bool `json:"supports_streaming"`
	// MaxRetries is the maximum retries supported.
	MaxRetries int `json:"max_retries"`
	// SupportedFormats are the supported output formats.
	SupportedFormats []OutputFormat `json:"supported_formats"`
}

// DefaultCapabilities returns default guidance capabilities.
func DefaultCapabilities() *GuidanceCapabilities {
	return &GuidanceCapabilities{
		SupportsRegex:   true,
		SupportsGrammar: false,
		SupportsSchema:  true,
		SupportsChoice:  true,
		MaxRetries:      10,
		SupportedFormats: []OutputFormat{
			FormatJSON,
			FormatEmail,
			FormatURL,
			FormatUUID,
		},
	}
}

// PredefinedConstraints provides common predefined constraints.
var PredefinedConstraints = struct {
	// Email is a constraint for email addresses.
	Email Constraint
	// URL is a constraint for URLs.
	URL Constraint
	// UUID is a constraint for UUIDs.
	UUID Constraint
	// YesNo is a constraint for yes/no answers.
	YesNo Constraint
	// TrueFalse is a constraint for true/false answers.
	TrueFalse Constraint
	// Numeric is a constraint for numeric values.
	Numeric Constraint
	// AlphaNumeric is a constraint for alphanumeric values.
	AlphaNumeric Constraint
	// SingleWord is a constraint for single word output.
	SingleWord Constraint
	// SingleSentence is a constraint for single sentence output.
	SingleSentence Constraint
}{
	Email:          NewFormatConstraint(FormatEmail),
	URL:            NewFormatConstraint(FormatURL),
	UUID:           NewFormatConstraint(FormatUUID),
	YesNo:          NewChoiceConstraint([]string{"yes", "no"}),
	TrueFalse:      NewChoiceConstraint([]string{"true", "false"}),
	Numeric:        mustRegexConstraint(`^-?\d+(\.\d+)?$`),
	AlphaNumeric:   mustRegexConstraint(`^[a-zA-Z0-9]+$`),
	SingleWord:     mustRegexConstraint(`^\S+$`),
	SingleSentence: NewLengthConstraint(1, 1, LengthUnitSentences),
}

func mustRegexConstraint(pattern string) Constraint {
	c, err := NewRegexConstraint(pattern)
	if err != nil {
		panic(err)
	}
	return c
}
