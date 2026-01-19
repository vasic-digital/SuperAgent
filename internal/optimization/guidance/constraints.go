// Package guidance provides constrained generation capabilities for LLM outputs.
package guidance

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrInvalidConstraint indicates the constraint is invalid.
	ErrInvalidConstraint = errors.New("invalid constraint")
	// ErrConstraintViolation indicates the output violates the constraint.
	ErrConstraintViolation = errors.New("constraint violation")
	// ErrUnsupportedConstraintType indicates the constraint type is not supported.
	ErrUnsupportedConstraintType = errors.New("unsupported constraint type")
)

// ConstraintType defines the type of constraint.
type ConstraintType string

const (
	// ConstraintTypeRegex uses regular expressions.
	ConstraintTypeRegex ConstraintType = "regex"
	// ConstraintTypeGrammar uses context-free grammars.
	ConstraintTypeGrammar ConstraintType = "grammar"
	// ConstraintTypeSchema uses JSON schema.
	ConstraintTypeSchema ConstraintType = "schema"
	// ConstraintTypeChoice restricts to specific options.
	ConstraintTypeChoice ConstraintType = "choice"
	// ConstraintTypeRange restricts to a numeric range.
	ConstraintTypeRange ConstraintType = "range"
	// ConstraintTypeLength restricts output length.
	ConstraintTypeLength ConstraintType = "length"
	// ConstraintTypeFormat restricts to specific formats.
	ConstraintTypeFormat ConstraintType = "format"
	// ConstraintTypeComposite combines multiple constraints.
	ConstraintTypeComposite ConstraintType = "composite"
)

// Constraint defines the interface for output constraints.
type Constraint interface {
	// Type returns the constraint type.
	Type() ConstraintType
	// Validate checks if the output satisfies the constraint.
	Validate(output string) error
	// Description returns a human-readable description.
	Description() string
	// Hint returns a hint for the LLM to follow the constraint.
	Hint() string
}

// RegexConstraint constrains output to match a regular expression.
type RegexConstraint struct {
	// Pattern is the regex pattern.
	Pattern string `json:"pattern"`
	// Name is an optional name for the constraint.
	Name string `json:"name,omitempty"`
	// Invert inverts the match (output must NOT match).
	Invert bool `json:"invert,omitempty"`

	compiled *regexp.Regexp
}

// NewRegexConstraint creates a new regex constraint.
func NewRegexConstraint(pattern string) (*RegexConstraint, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid regex pattern: %v", ErrInvalidConstraint, err)
	}
	return &RegexConstraint{
		Pattern:  pattern,
		compiled: compiled,
	}, nil
}

// Type returns the constraint type.
func (c *RegexConstraint) Type() ConstraintType {
	return ConstraintTypeRegex
}

// Validate checks if the output matches the pattern.
func (c *RegexConstraint) Validate(output string) error {
	if c.compiled == nil {
		compiled, err := regexp.Compile(c.Pattern)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidConstraint, err)
		}
		c.compiled = compiled
	}

	matches := c.compiled.MatchString(output)
	if c.Invert {
		if matches {
			return fmt.Errorf("%w: output matches excluded pattern", ErrConstraintViolation)
		}
	} else {
		if !matches {
			return fmt.Errorf("%w: output does not match pattern %s", ErrConstraintViolation, c.Pattern)
		}
	}
	return nil
}

// Description returns a human-readable description.
func (c *RegexConstraint) Description() string {
	if c.Name != "" {
		return c.Name
	}
	if c.Invert {
		return fmt.Sprintf("Must not match pattern: %s", c.Pattern)
	}
	return fmt.Sprintf("Must match pattern: %s", c.Pattern)
}

// Hint returns a hint for the LLM.
func (c *RegexConstraint) Hint() string {
	if c.Invert {
		return fmt.Sprintf("Output must NOT match the pattern: %s", c.Pattern)
	}
	return fmt.Sprintf("Output must match the pattern: %s", c.Pattern)
}

// ChoiceConstraint restricts output to specific options.
type ChoiceConstraint struct {
	// Options are the allowed choices.
	Options []string `json:"options"`
	// CaseSensitive indicates if comparison is case-sensitive.
	CaseSensitive bool `json:"case_sensitive,omitempty"`
	// AllowMultiple allows selecting multiple options.
	AllowMultiple bool `json:"allow_multiple,omitempty"`
	// Separator is used when multiple options are allowed.
	Separator string `json:"separator,omitempty"`
}

// NewChoiceConstraint creates a new choice constraint.
func NewChoiceConstraint(options []string) *ChoiceConstraint {
	return &ChoiceConstraint{
		Options:       options,
		CaseSensitive: true,
		Separator:     ",",
	}
}

// Type returns the constraint type.
func (c *ChoiceConstraint) Type() ConstraintType {
	return ConstraintTypeChoice
}

// Validate checks if the output is one of the allowed options.
func (c *ChoiceConstraint) Validate(output string) error {
	output = strings.TrimSpace(output)

	var toCheck []string
	if c.AllowMultiple && c.Separator != "" {
		parts := strings.Split(output, c.Separator)
		for _, p := range parts {
			toCheck = append(toCheck, strings.TrimSpace(p))
		}
	} else {
		toCheck = []string{output}
	}

	for _, check := range toCheck {
		found := false
		for _, opt := range c.Options {
			if c.CaseSensitive {
				if check == opt {
					found = true
					break
				}
			} else {
				if strings.EqualFold(check, opt) {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("%w: '%s' is not one of the allowed options: %v", ErrConstraintViolation, check, c.Options)
		}
	}
	return nil
}

// Description returns a human-readable description.
func (c *ChoiceConstraint) Description() string {
	return fmt.Sprintf("Must be one of: %s", strings.Join(c.Options, ", "))
}

// Hint returns a hint for the LLM.
func (c *ChoiceConstraint) Hint() string {
	return fmt.Sprintf("Choose from: %s", strings.Join(c.Options, ", "))
}

// LengthConstraint restricts output length.
type LengthConstraint struct {
	// MinLength is the minimum length (0 = no minimum).
	MinLength int `json:"min_length,omitempty"`
	// MaxLength is the maximum length (0 = no maximum).
	MaxLength int `json:"max_length,omitempty"`
	// Unit is the unit of measurement (characters, words, sentences, tokens).
	Unit LengthUnit `json:"unit"`
}

// LengthUnit defines the unit for length constraints.
type LengthUnit string

const (
	// LengthUnitCharacters counts characters.
	LengthUnitCharacters LengthUnit = "characters"
	// LengthUnitWords counts words.
	LengthUnitWords LengthUnit = "words"
	// LengthUnitSentences counts sentences.
	LengthUnitSentences LengthUnit = "sentences"
	// LengthUnitTokens counts tokens (approximated by words).
	LengthUnitTokens LengthUnit = "tokens"
)

// NewLengthConstraint creates a new length constraint.
func NewLengthConstraint(minLength, maxLength int, unit LengthUnit) *LengthConstraint {
	return &LengthConstraint{
		MinLength: minLength,
		MaxLength: maxLength,
		Unit:      unit,
	}
}

// Type returns the constraint type.
func (c *LengthConstraint) Type() ConstraintType {
	return ConstraintTypeLength
}

// Validate checks if the output length is within bounds.
func (c *LengthConstraint) Validate(output string) error {
	length := c.countUnits(output)

	if c.MinLength > 0 && length < c.MinLength {
		return fmt.Errorf("%w: output too short (%d %s, minimum %d)", ErrConstraintViolation, length, c.Unit, c.MinLength)
	}
	if c.MaxLength > 0 && length > c.MaxLength {
		return fmt.Errorf("%w: output too long (%d %s, maximum %d)", ErrConstraintViolation, length, c.Unit, c.MaxLength)
	}
	return nil
}

func (c *LengthConstraint) countUnits(text string) int {
	switch c.Unit {
	case LengthUnitCharacters:
		return len(text)
	case LengthUnitWords, LengthUnitTokens:
		return len(strings.Fields(text))
	case LengthUnitSentences:
		return countSentences(text)
	default:
		return len(text)
	}
}

func countSentences(text string) int {
	// Simple sentence counting based on terminal punctuation
	count := 0
	for _, r := range text {
		if r == '.' || r == '!' || r == '?' {
			count++
		}
	}
	return count
}

// Description returns a human-readable description.
func (c *LengthConstraint) Description() string {
	parts := []string{}
	if c.MinLength > 0 {
		parts = append(parts, fmt.Sprintf("at least %d %s", c.MinLength, c.Unit))
	}
	if c.MaxLength > 0 {
		parts = append(parts, fmt.Sprintf("at most %d %s", c.MaxLength, c.Unit))
	}
	return fmt.Sprintf("Length must be %s", strings.Join(parts, " and "))
}

// Hint returns a hint for the LLM.
func (c *LengthConstraint) Hint() string {
	if c.MinLength > 0 && c.MaxLength > 0 {
		return fmt.Sprintf("Keep response between %d and %d %s", c.MinLength, c.MaxLength, c.Unit)
	}
	if c.MinLength > 0 {
		return fmt.Sprintf("Response must be at least %d %s", c.MinLength, c.Unit)
	}
	if c.MaxLength > 0 {
		return fmt.Sprintf("Response must be at most %d %s", c.MaxLength, c.Unit)
	}
	return ""
}

// RangeConstraint restricts numeric output to a range.
type RangeConstraint struct {
	// Min is the minimum value.
	Min float64 `json:"min"`
	// Max is the maximum value.
	Max float64 `json:"max"`
	// IntegerOnly restricts to integers.
	IntegerOnly bool `json:"integer_only,omitempty"`
}

// NewRangeConstraint creates a new range constraint.
func NewRangeConstraint(min, max float64) *RangeConstraint {
	return &RangeConstraint{
		Min: min,
		Max: max,
	}
}

// Type returns the constraint type.
func (c *RangeConstraint) Type() ConstraintType {
	return ConstraintTypeRange
}

// Validate checks if the output is within the range.
func (c *RangeConstraint) Validate(output string) error {
	output = strings.TrimSpace(output)
	var value float64
	if _, err := fmt.Sscanf(output, "%f", &value); err != nil {
		return fmt.Errorf("%w: output is not a valid number", ErrConstraintViolation)
	}

	if c.IntegerOnly {
		if value != float64(int64(value)) {
			return fmt.Errorf("%w: output must be an integer", ErrConstraintViolation)
		}
	}

	if value < c.Min || value > c.Max {
		return fmt.Errorf("%w: value %f is outside range [%f, %f]", ErrConstraintViolation, value, c.Min, c.Max)
	}
	return nil
}

// Description returns a human-readable description.
func (c *RangeConstraint) Description() string {
	if c.IntegerOnly {
		return fmt.Sprintf("Integer between %d and %d", int64(c.Min), int64(c.Max))
	}
	return fmt.Sprintf("Number between %f and %f", c.Min, c.Max)
}

// Hint returns a hint for the LLM.
func (c *RangeConstraint) Hint() string {
	if c.IntegerOnly {
		return fmt.Sprintf("Provide an integer between %d and %d", int64(c.Min), int64(c.Max))
	}
	return fmt.Sprintf("Provide a number between %f and %f", c.Min, c.Max)
}

// FormatConstraint restricts output to specific formats.
type FormatConstraint struct {
	// Format is the required format.
	Format OutputFormat `json:"format"`
}

// OutputFormat defines standard output formats.
type OutputFormat string

const (
	// FormatJSON requires valid JSON output.
	FormatJSON OutputFormat = "json"
	// FormatXML requires valid XML output.
	FormatXML OutputFormat = "xml"
	// FormatYAML requires valid YAML output.
	FormatYAML OutputFormat = "yaml"
	// FormatMarkdown requires valid Markdown output.
	FormatMarkdown OutputFormat = "markdown"
	// FormatEmail requires valid email format.
	FormatEmail OutputFormat = "email"
	// FormatURL requires valid URL format.
	FormatURL OutputFormat = "url"
	// FormatDate requires valid date format.
	FormatDate OutputFormat = "date"
	// FormatTime requires valid time format.
	FormatTime OutputFormat = "time"
	// FormatDateTime requires valid datetime format.
	FormatDateTime OutputFormat = "datetime"
	// FormatPhoneNumber requires valid phone number format.
	FormatPhoneNumber OutputFormat = "phone"
	// FormatIPv4 requires valid IPv4 address.
	FormatIPv4 OutputFormat = "ipv4"
	// FormatIPv6 requires valid IPv6 address.
	FormatIPv6 OutputFormat = "ipv6"
	// FormatUUID requires valid UUID format.
	FormatUUID OutputFormat = "uuid"
)

// NewFormatConstraint creates a new format constraint.
func NewFormatConstraint(format OutputFormat) *FormatConstraint {
	return &FormatConstraint{Format: format}
}

// Type returns the constraint type.
func (c *FormatConstraint) Type() ConstraintType {
	return ConstraintTypeFormat
}

// Validate checks if the output matches the required format.
func (c *FormatConstraint) Validate(output string) error {
	output = strings.TrimSpace(output)

	switch c.Format {
	case FormatJSON:
		var js json.RawMessage
		if err := json.Unmarshal([]byte(output), &js); err != nil {
			return fmt.Errorf("%w: invalid JSON: %v", ErrConstraintViolation, err)
		}
	case FormatEmail:
		if !emailRegex.MatchString(output) {
			return fmt.Errorf("%w: invalid email format", ErrConstraintViolation)
		}
	case FormatURL:
		if !urlRegex.MatchString(output) {
			return fmt.Errorf("%w: invalid URL format", ErrConstraintViolation)
		}
	case FormatUUID:
		if !uuidRegex.MatchString(output) {
			return fmt.Errorf("%w: invalid UUID format", ErrConstraintViolation)
		}
	case FormatIPv4:
		if !ipv4Regex.MatchString(output) {
			return fmt.Errorf("%w: invalid IPv4 format", ErrConstraintViolation)
		}
	case FormatPhoneNumber:
		if !phoneRegex.MatchString(output) {
			return fmt.Errorf("%w: invalid phone number format", ErrConstraintViolation)
		}
	default:
		// For other formats, basic validation
		if len(output) == 0 {
			return fmt.Errorf("%w: output is empty", ErrConstraintViolation)
		}
	}
	return nil
}

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlRegex   = regexp.MustCompile(`^https?://[^\s]+$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	ipv4Regex  = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	phoneRegex = regexp.MustCompile(`^[\d\s\-\+\(\)]{7,20}$`)
)

// Description returns a human-readable description.
func (c *FormatConstraint) Description() string {
	return fmt.Sprintf("Must be valid %s format", c.Format)
}

// Hint returns a hint for the LLM.
func (c *FormatConstraint) Hint() string {
	return fmt.Sprintf("Output must be in %s format", c.Format)
}

// SchemaConstraint constrains output to match a JSON schema.
type SchemaConstraint struct {
	// Schema is the JSON schema.
	Schema map[string]interface{} `json:"schema"`
	// Strict enables strict validation.
	Strict bool `json:"strict,omitempty"`
}

// NewSchemaConstraint creates a new schema constraint.
func NewSchemaConstraint(schema map[string]interface{}) *SchemaConstraint {
	return &SchemaConstraint{
		Schema: schema,
		Strict: true,
	}
}

// Type returns the constraint type.
func (c *SchemaConstraint) Type() ConstraintType {
	return ConstraintTypeSchema
}

// Validate checks if the output matches the JSON schema.
func (c *SchemaConstraint) Validate(output string) error {
	// First check if it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		return fmt.Errorf("%w: output is not valid JSON: %v", ErrConstraintViolation, err)
	}

	// Basic schema validation (type checking for required properties)
	if props, ok := c.Schema["properties"].(map[string]interface{}); ok {
		required := []string{}
		if req, ok := c.Schema["required"].([]interface{}); ok {
			for _, r := range req {
				if s, ok := r.(string); ok {
					required = append(required, s)
				}
			}
		}

		for _, prop := range required {
			if _, exists := parsed[prop]; !exists {
				return fmt.Errorf("%w: missing required property: %s", ErrConstraintViolation, prop)
			}
		}

		for propName, propSchema := range props {
			if propValue, exists := parsed[propName]; exists {
				if err := validateProperty(propName, propValue, propSchema); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateProperty(name string, value interface{}, schema interface{}) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return nil
	}

	expectedType, hasType := schemaMap["type"].(string)
	if !hasType {
		return nil
	}

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%w: property '%s' must be a string", ErrConstraintViolation, name)
		}
	case "number", "integer":
		switch value.(type) {
		case float64, int, int64:
			// Valid
		default:
			return fmt.Errorf("%w: property '%s' must be a number", ErrConstraintViolation, name)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%w: property '%s' must be a boolean", ErrConstraintViolation, name)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("%w: property '%s' must be an array", ErrConstraintViolation, name)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("%w: property '%s' must be an object", ErrConstraintViolation, name)
		}
	}

	return nil
}

// Description returns a human-readable description.
func (c *SchemaConstraint) Description() string {
	return "Must match JSON schema"
}

// Hint returns a hint for the LLM.
func (c *SchemaConstraint) Hint() string {
	schemaJSON, _ := json.MarshalIndent(c.Schema, "", "  ")
	return fmt.Sprintf("Output must be valid JSON matching this schema:\n%s", string(schemaJSON))
}

// CompositeConstraint combines multiple constraints.
type CompositeConstraint struct {
	// Constraints are the individual constraints.
	Constraints []Constraint `json:"constraints"`
	// Mode determines how constraints are combined.
	Mode CompositeMode `json:"mode"`
}

// CompositeMode defines how constraints are combined.
type CompositeMode string

const (
	// CompositeModeAll requires all constraints to pass (AND).
	CompositeModeAll CompositeMode = "all"
	// CompositeModeAny requires at least one constraint to pass (OR).
	CompositeModeAny CompositeMode = "any"
)

// NewCompositeConstraint creates a new composite constraint.
func NewCompositeConstraint(mode CompositeMode, constraints ...Constraint) *CompositeConstraint {
	return &CompositeConstraint{
		Constraints: constraints,
		Mode:        mode,
	}
}

// Type returns the constraint type.
func (c *CompositeConstraint) Type() ConstraintType {
	return ConstraintTypeComposite
}

// Validate checks if the output satisfies the composite constraint.
func (c *CompositeConstraint) Validate(output string) error {
	var errs []error

	for _, constraint := range c.Constraints {
		err := constraint.Validate(output)
		if err != nil {
			errs = append(errs, err)
		}
	}

	switch c.Mode {
	case CompositeModeAll:
		if len(errs) > 0 {
			return fmt.Errorf("%w: %v", ErrConstraintViolation, errs[0])
		}
	case CompositeModeAny:
		if len(errs) == len(c.Constraints) {
			return fmt.Errorf("%w: none of the constraints were satisfied", ErrConstraintViolation)
		}
	}

	return nil
}

// Description returns a human-readable description.
func (c *CompositeConstraint) Description() string {
	var parts []string
	for _, constraint := range c.Constraints {
		parts = append(parts, constraint.Description())
	}
	connector := " AND "
	if c.Mode == CompositeModeAny {
		connector = " OR "
	}
	return strings.Join(parts, connector)
}

// Hint returns a hint for the LLM.
func (c *CompositeConstraint) Hint() string {
	var parts []string
	for _, constraint := range c.Constraints {
		parts = append(parts, constraint.Hint())
	}
	return strings.Join(parts, "\n")
}

// GrammarConstraint constrains output using a context-free grammar.
type GrammarConstraint struct {
	// Grammar is the EBNF or similar grammar definition.
	Grammar string `json:"grammar"`
	// StartSymbol is the start symbol of the grammar.
	StartSymbol string `json:"start_symbol,omitempty"`
}

// NewGrammarConstraint creates a new grammar constraint.
func NewGrammarConstraint(grammar string) *GrammarConstraint {
	return &GrammarConstraint{
		Grammar:     grammar,
		StartSymbol: "start",
	}
}

// Type returns the constraint type.
func (c *GrammarConstraint) Type() ConstraintType {
	return ConstraintTypeGrammar
}

// Validate checks if the output follows the grammar.
// Note: Full grammar validation requires a parser, this is a placeholder.
func (c *GrammarConstraint) Validate(output string) error {
	// Grammar validation would require a full parser implementation
	// This is a placeholder that always passes
	if len(strings.TrimSpace(output)) == 0 {
		return fmt.Errorf("%w: output is empty", ErrConstraintViolation)
	}
	return nil
}

// Description returns a human-readable description.
func (c *GrammarConstraint) Description() string {
	return "Must follow the specified grammar"
}

// Hint returns a hint for the LLM.
func (c *GrammarConstraint) Hint() string {
	return fmt.Sprintf("Output must follow this grammar:\n%s", c.Grammar)
}

// ConstraintBuilder helps build constraints fluently.
type ConstraintBuilder struct {
	constraints []Constraint
}

// NewConstraintBuilder creates a new constraint builder.
func NewConstraintBuilder() *ConstraintBuilder {
	return &ConstraintBuilder{
		constraints: []Constraint{},
	}
}

// WithRegex adds a regex constraint.
func (b *ConstraintBuilder) WithRegex(pattern string) *ConstraintBuilder {
	if c, err := NewRegexConstraint(pattern); err == nil {
		b.constraints = append(b.constraints, c)
	}
	return b
}

// WithChoice adds a choice constraint.
func (b *ConstraintBuilder) WithChoice(options ...string) *ConstraintBuilder {
	b.constraints = append(b.constraints, NewChoiceConstraint(options))
	return b
}

// WithLength adds a length constraint.
func (b *ConstraintBuilder) WithLength(min, max int, unit LengthUnit) *ConstraintBuilder {
	b.constraints = append(b.constraints, NewLengthConstraint(min, max, unit))
	return b
}

// WithRange adds a range constraint.
func (b *ConstraintBuilder) WithRange(min, max float64) *ConstraintBuilder {
	b.constraints = append(b.constraints, NewRangeConstraint(min, max))
	return b
}

// WithFormat adds a format constraint.
func (b *ConstraintBuilder) WithFormat(format OutputFormat) *ConstraintBuilder {
	b.constraints = append(b.constraints, NewFormatConstraint(format))
	return b
}

// WithSchema adds a schema constraint.
func (b *ConstraintBuilder) WithSchema(schema map[string]interface{}) *ConstraintBuilder {
	b.constraints = append(b.constraints, NewSchemaConstraint(schema))
	return b
}

// BuildAll builds a composite constraint requiring all constraints (AND).
func (b *ConstraintBuilder) BuildAll() Constraint {
	if len(b.constraints) == 1 {
		return b.constraints[0]
	}
	return NewCompositeConstraint(CompositeModeAll, b.constraints...)
}

// BuildAny builds a composite constraint requiring any constraint (OR).
func (b *ConstraintBuilder) BuildAny() Constraint {
	if len(b.constraints) == 1 {
		return b.constraints[0]
	}
	return NewCompositeConstraint(CompositeModeAny, b.constraints...)
}
