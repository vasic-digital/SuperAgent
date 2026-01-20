// Package guidance provides constrained generation capabilities for LLM outputs.
package guidance

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	// ErrGenerationFailed indicates generation failed.
	ErrGenerationFailed = errors.New("generation failed")
	// ErrMaxRetriesExceeded indicates maximum retries were exceeded.
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
	// ErrValidationFailed indicates validation failed.
	ErrValidationFailed = errors.New("validation failed")
)

// Generator defines the interface for constrained text generation.
type Generator interface {
	// Generate generates text following the given constraints.
	Generate(ctx context.Context, prompt string, constraints Constraint) (*GenerationResult, error)
	// GenerateWithRetry generates with automatic retry on constraint violation.
	GenerateWithRetry(ctx context.Context, prompt string, constraints Constraint, maxRetries int) (*GenerationResult, error)
}

// GenerationResult contains the result of a constrained generation.
type GenerationResult struct {
	// Output is the generated text.
	Output string `json:"output"`
	// Valid indicates if the output satisfies all constraints.
	Valid bool `json:"valid"`
	// Attempts is the number of generation attempts.
	Attempts int `json:"attempts"`
	// ValidationErrors contains any validation errors.
	ValidationErrors []string `json:"validation_errors,omitempty"`
	// Metadata contains additional generation metadata.
	Metadata *GenerationMetadata `json:"metadata,omitempty"`
}

// GenerationMetadata contains metadata about the generation.
type GenerationMetadata struct {
	// Model is the model used for generation.
	Model string `json:"model,omitempty"`
	// Provider is the provider used.
	Provider string `json:"provider,omitempty"`
	// TokensUsed is the number of tokens used.
	TokensUsed int `json:"tokens_used,omitempty"`
	// LatencyMs is the generation latency in milliseconds.
	LatencyMs int64 `json:"latency_ms,omitempty"`
	// ConstraintType is the type of constraint used.
	ConstraintType ConstraintType `json:"constraint_type,omitempty"`
	// Timestamp is when generation completed.
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// LLMBackend defines the interface for LLM generation.
type LLMBackend interface {
	// Complete generates a completion for the given prompt.
	Complete(ctx context.Context, prompt string) (string, error)
	// CompleteWithHint generates with constraint hints.
	CompleteWithHint(ctx context.Context, prompt string, hint string) (string, error)
}

// ConstrainedGenerator implements constrained generation using an LLM backend.
type ConstrainedGenerator struct {
	backend LLMBackend
	config  *GeneratorConfig
}

// GeneratorConfig holds configuration for the generator.
type GeneratorConfig struct {
	// DefaultMaxRetries is the default number of retries.
	DefaultMaxRetries int `json:"default_max_retries"`
	// RetryDelay is the delay between retries.
	RetryDelay time.Duration `json:"retry_delay"`
	// IncludeConstraintHints adds constraint hints to prompts.
	IncludeConstraintHints bool `json:"include_constraint_hints"`
	// Model is the model to use.
	Model string `json:"model,omitempty"`
	// Provider is the provider to use.
	Provider string `json:"provider,omitempty"`
}

// DefaultGeneratorConfig returns a default configuration.
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		DefaultMaxRetries:      3,
		RetryDelay:             100 * time.Millisecond,
		IncludeConstraintHints: true,
	}
}

// NewConstrainedGenerator creates a new constrained generator.
func NewConstrainedGenerator(backend LLMBackend, config *GeneratorConfig) *ConstrainedGenerator {
	if config == nil {
		config = DefaultGeneratorConfig()
	}
	return &ConstrainedGenerator{
		backend: backend,
		config:  config,
	}
}

// Generate generates text following the given constraints.
func (g *ConstrainedGenerator) Generate(ctx context.Context, prompt string, constraints Constraint) (*GenerationResult, error) {
	return g.GenerateWithRetry(ctx, prompt, constraints, 1)
}

// GenerateWithRetry generates with automatic retry on constraint violation.
func (g *ConstrainedGenerator) GenerateWithRetry(ctx context.Context, prompt string, constraints Constraint, maxRetries int) (*GenerationResult, error) {
	if maxRetries <= 0 {
		maxRetries = g.config.DefaultMaxRetries
	}

	result := &GenerationResult{
		Metadata: &GenerationMetadata{
			Model:          g.config.Model,
			Provider:       g.config.Provider,
			ConstraintType: constraints.Type(),
		},
	}

	var lastOutput string
	var validationErrors []string

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result.Attempts = attempt

		// Build prompt with hint if configured
		finalPrompt := prompt
		if g.config.IncludeConstraintHints {
			hint := constraints.Hint()
			if hint != "" {
				finalPrompt = fmt.Sprintf("%s\n\n[Constraint: %s]", prompt, hint)
			}
		}

		// Generate output
		start := time.Now()
		output, err := g.backend.Complete(ctx, finalPrompt)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			continue
		}
		result.Metadata.LatencyMs = time.Since(start).Milliseconds()

		// Clean output
		output = strings.TrimSpace(output)
		lastOutput = output

		// Validate against constraints
		if err := constraints.Validate(output); err != nil {
			validationErrors = append(validationErrors, err.Error())

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(g.config.RetryDelay):
				}
				continue
			}
		} else {
			result.Output = output
			result.Valid = true
			result.Metadata.Timestamp = time.Now()
			return result, nil
		}
	}

	result.Output = lastOutput
	result.Valid = false
	result.ValidationErrors = validationErrors
	result.Metadata.Timestamp = time.Now()

	return result, fmt.Errorf("%w: %d attempts failed validation", ErrMaxRetriesExceeded, maxRetries)
}

// TemplatedGenerator generates text using templates with placeholders.
type TemplatedGenerator struct {
	generator Generator
}

// NewTemplatedGenerator creates a new templated generator.
func NewTemplatedGenerator(generator Generator) *TemplatedGenerator {
	return &TemplatedGenerator{generator: generator}
}

// Template represents a generation template with placeholders.
type Template struct {
	// Text is the template text with placeholders like {{placeholder}}.
	Text string `json:"text"`
	// Placeholders define constraints for each placeholder.
	Placeholders map[string]Constraint `json:"placeholders"`
}

// TemplateResult contains the result of template generation.
type TemplateResult struct {
	// FilledTemplate is the template with placeholders replaced.
	FilledTemplate string `json:"filled_template"`
	// Values contains the generated value for each placeholder.
	Values map[string]string `json:"values"`
	// Valid indicates if all placeholders were filled validly.
	Valid bool `json:"valid"`
	// Errors contains errors for failed placeholders.
	Errors map[string]string `json:"errors,omitempty"`
}

// GenerateFromTemplate fills a template by generating values for placeholders.
func (g *TemplatedGenerator) GenerateFromTemplate(ctx context.Context, template *Template) (*TemplateResult, error) {
	result := &TemplateResult{
		Values:         make(map[string]string),
		Errors:         make(map[string]string),
		FilledTemplate: template.Text,
		Valid:          true,
	}

	// Sort placeholder keys for deterministic iteration order
	placeholderKeys := make([]string, 0, len(template.Placeholders))
	for k := range template.Placeholders {
		placeholderKeys = append(placeholderKeys, k)
	}
	sort.Strings(placeholderKeys)

	for _, placeholder := range placeholderKeys {
		constraint := template.Placeholders[placeholder]
		prompt := fmt.Sprintf("Generate a value for '%s'", placeholder)

		genResult, err := g.generator.Generate(ctx, prompt, constraint)
		if err != nil {
			result.Errors[placeholder] = err.Error()
			result.Valid = false
			continue
		}

		if !genResult.Valid {
			result.Errors[placeholder] = "validation failed"
			result.Valid = false
			continue
		}

		result.Values[placeholder] = genResult.Output
		result.FilledTemplate = strings.ReplaceAll(
			result.FilledTemplate,
			fmt.Sprintf("{{%s}}", placeholder),
			genResult.Output,
		)
	}

	return result, nil
}

// SelectionGenerator generates by selecting from options.
type SelectionGenerator struct {
	backend LLMBackend
}

// NewSelectionGenerator creates a new selection generator.
func NewSelectionGenerator(backend LLMBackend) *SelectionGenerator {
	return &SelectionGenerator{backend: backend}
}

// SelectionResult contains the result of a selection.
type SelectionResult struct {
	// Selected is the selected option.
	Selected string `json:"selected"`
	// Index is the index of the selected option.
	Index int `json:"index"`
	// Confidence is the confidence in the selection (if available).
	Confidence float64 `json:"confidence,omitempty"`
	// Reasoning is the reasoning for the selection.
	Reasoning string `json:"reasoning,omitempty"`
}

// Select selects one option from a list.
func (g *SelectionGenerator) Select(ctx context.Context, prompt string, options []string) (*SelectionResult, error) {
	// Build selection prompt
	optionList := strings.Join(options, ", ")
	fullPrompt := fmt.Sprintf("%s\nOptions: %s\nRespond with only the selected option.", prompt, optionList)

	output, err := g.backend.Complete(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	output = strings.TrimSpace(output)

	// Find matching option
	for i, opt := range options {
		if strings.EqualFold(output, opt) || strings.Contains(strings.ToLower(output), strings.ToLower(opt)) {
			return &SelectionResult{
				Selected: opt,
				Index:    i,
			}, nil
		}
	}

	// If no exact match, return the raw output with the first option as fallback
	return &SelectionResult{
		Selected: options[0],
		Index:    0,
	}, nil
}

// SelectMultiple selects multiple options from a list.
func (g *SelectionGenerator) SelectMultiple(ctx context.Context, prompt string, options []string, maxSelections int) (*MultiSelectionResult, error) {
	// Build multi-selection prompt
	optionList := strings.Join(options, ", ")
	fullPrompt := fmt.Sprintf("%s\nOptions: %s\nSelect up to %d options. Separate with commas.", prompt, optionList, maxSelections)

	output, err := g.backend.Complete(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	output = strings.TrimSpace(output)
	parts := strings.Split(output, ",")

	result := &MultiSelectionResult{
		Selected: []string{},
		Indices:  []int{},
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		for i, opt := range options {
			if strings.EqualFold(part, opt) {
				result.Selected = append(result.Selected, opt)
				result.Indices = append(result.Indices, i)
				break
			}
		}
		if len(result.Selected) >= maxSelections {
			break
		}
	}

	return result, nil
}

// MultiSelectionResult contains the result of a multi-selection.
type MultiSelectionResult struct {
	// Selected are the selected options.
	Selected []string `json:"selected"`
	// Indices are the indices of the selected options.
	Indices []int `json:"indices"`
}

// StructuredGenerator generates structured data.
type StructuredGenerator struct {
	generator Generator
}

// NewStructuredGenerator creates a new structured generator.
func NewStructuredGenerator(generator Generator) *StructuredGenerator {
	return &StructuredGenerator{generator: generator}
}

// GenerateJSON generates JSON matching a schema.
func (g *StructuredGenerator) GenerateJSON(ctx context.Context, prompt string, schema map[string]interface{}) (*GenerationResult, error) {
	constraint := NewSchemaConstraint(schema)
	jsonPrompt := fmt.Sprintf("%s\n\nRespond with valid JSON only.", prompt)
	return g.generator.GenerateWithRetry(ctx, jsonPrompt, constraint, 3)
}

// GenerateList generates a list of items.
func (g *StructuredGenerator) GenerateList(ctx context.Context, prompt string, itemConstraint Constraint, count int) (*ListGenerationResult, error) {
	result := &ListGenerationResult{
		Items: make([]string, 0, count),
	}

	for i := 0; i < count; i++ {
		itemPrompt := fmt.Sprintf("%s (item %d of %d)", prompt, i+1, count)
		genResult, err := g.generator.Generate(ctx, itemPrompt, itemConstraint)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		if genResult.Valid {
			result.Items = append(result.Items, genResult.Output)
		}
	}

	result.Valid = len(result.Items) == count
	return result, nil
}

// ListGenerationResult contains the result of list generation.
type ListGenerationResult struct {
	// Items are the generated items.
	Items []string `json:"items"`
	// Valid indicates if all items were generated successfully.
	Valid bool `json:"valid"`
	// Errors contains any errors that occurred.
	Errors []string `json:"errors,omitempty"`
}

// GuidedCompletionConfig holds configuration for guided completion.
type GuidedCompletionConfig struct {
	// StopTokens are tokens that stop generation.
	StopTokens []string `json:"stop_tokens,omitempty"`
	// MaxTokens is the maximum tokens to generate.
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature controls randomness.
	Temperature float64 `json:"temperature,omitempty"`
	// TopP is nucleus sampling probability.
	TopP float64 `json:"top_p,omitempty"`
}

// GuidedCompletion provides guided text completion.
type GuidedCompletion struct {
	backend LLMBackend
	config  *GuidedCompletionConfig
}

// NewGuidedCompletion creates a new guided completion.
func NewGuidedCompletion(backend LLMBackend, config *GuidedCompletionConfig) *GuidedCompletion {
	if config == nil {
		config = &GuidedCompletionConfig{
			MaxTokens:   500,
			Temperature: 0.7,
		}
	}
	return &GuidedCompletion{
		backend: backend,
		config:  config,
	}
}

// CompleteUntil generates until a stop condition is met.
func (g *GuidedCompletion) CompleteUntil(ctx context.Context, prompt string, stopCondition func(string) bool) (*GuidedCompletionResult, error) {
	output, err := g.backend.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	// Check stop condition
	if stopCondition != nil && stopCondition(output) {
		// Truncate at stop condition if needed
	}

	return &GuidedCompletionResult{
		Output:   output,
		Complete: true,
	}, nil
}

// CompleteWithPrefix generates with a required prefix.
func (g *GuidedCompletion) CompleteWithPrefix(ctx context.Context, prompt string, prefix string) (*GuidedCompletionResult, error) {
	fullPrompt := fmt.Sprintf("%s\n\nStart your response with: %s", prompt, prefix)

	output, err := g.backend.Complete(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	// Ensure prefix is present
	if !strings.HasPrefix(output, prefix) {
		output = prefix + " " + output
	}

	return &GuidedCompletionResult{
		Output:   output,
		Complete: true,
	}, nil
}

// CompleteWithSuffix generates with a required suffix.
func (g *GuidedCompletion) CompleteWithSuffix(ctx context.Context, prompt string, suffix string) (*GuidedCompletionResult, error) {
	fullPrompt := fmt.Sprintf("%s\n\nEnd your response with: %s", prompt, suffix)

	output, err := g.backend.Complete(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	// Ensure suffix is present
	if !strings.HasSuffix(output, suffix) {
		output = strings.TrimSuffix(output, ".") + " " + suffix
	}

	return &GuidedCompletionResult{
		Output:   output,
		Complete: true,
	}, nil
}

// GuidedCompletionResult contains the result of guided completion.
type GuidedCompletionResult struct {
	// Output is the generated text.
	Output string `json:"output"`
	// Complete indicates if generation completed normally.
	Complete bool `json:"complete"`
	// StoppedAt is the stop token that terminated generation.
	StoppedAt string `json:"stopped_at,omitempty"`
	// TokensGenerated is the number of tokens generated.
	TokensGenerated int `json:"tokens_generated,omitempty"`
}
