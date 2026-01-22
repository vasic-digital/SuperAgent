// Package structured provides constrained decoding and structured output generation
// for LLM responses, inspired by XGrammar and similar approaches.
package structured

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// OutputFormat represents the desired output format
type OutputFormat string

const (
	OutputFormatJSON      OutputFormat = "json"
	OutputFormatJSONLines OutputFormat = "jsonl"
	OutputFormatYAML      OutputFormat = "yaml"
	OutputFormatXML       OutputFormat = "xml"
	OutputFormatMarkdown  OutputFormat = "markdown"
	OutputFormatCSV       OutputFormat = "csv"
)

// Schema represents a JSON Schema for structured output
type Schema struct {
	Type        string             `json:"type"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Enum        []interface{}      `json:"enum,omitempty"`
	Pattern     string             `json:"pattern,omitempty"`
	MinLength   *int               `json:"minLength,omitempty"`
	MaxLength   *int               `json:"maxLength,omitempty"`
	Minimum     *float64           `json:"minimum,omitempty"`
	Maximum     *float64           `json:"maximum,omitempty"`
	MinItems    *int               `json:"minItems,omitempty"`
	MaxItems    *int               `json:"maxItems,omitempty"`
	Description string             `json:"description,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Format      string             `json:"format,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	OneOf       []*Schema          `json:"oneOf,omitempty"`
	AnyOf       []*Schema          `json:"anyOf,omitempty"`
	AllOf       []*Schema          `json:"allOf,omitempty"`
	Definitions map[string]*Schema `json:"definitions,omitempty"`
}

// SchemaFromType generates a JSON Schema from a Go type
func SchemaFromType(t interface{}) (*Schema, error) {
	val := reflect.TypeOf(t)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return schemaFromReflectType(val, make(map[string]bool))
}

func schemaFromReflectType(t reflect.Type, visited map[string]bool) (*Schema, error) {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for circular references
	typeName := t.String()
	if visited[typeName] {
		return &Schema{Ref: "#/definitions/" + typeName}, nil
	}
	visited[typeName] = true
	defer delete(visited, typeName)

	schema := &Schema{}

	switch t.Kind() {
	case reflect.String:
		schema.Type = "string"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"

	case reflect.Float32, reflect.Float64:
		schema.Type = "number"

	case reflect.Bool:
		schema.Type = "boolean"

	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		itemSchema, err := schemaFromReflectType(t.Elem(), visited)
		if err != nil {
			return nil, err
		}
		schema.Items = itemSchema

	case reflect.Map:
		schema.Type = "object"
		// For maps, we can't easily determine property schemas

	case reflect.Struct:
		schema.Type = "object"
		schema.Properties = make(map[string]*Schema)
		schema.Required = make([]string, 0)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			// Get JSON tag
			jsonTag := field.Tag.Get("json")
			fieldName := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "-" {
					fieldName = parts[0]
				} else {
					continue
				}
			}

			propSchema, err := schemaFromReflectType(field.Type, visited)
			if err != nil {
				return nil, err
			}

			// Add description from struct tag
			if desc := field.Tag.Get("description"); desc != "" {
				propSchema.Description = desc
			}

			schema.Properties[fieldName] = propSchema

			// Check if field is required
			if jsonTag != "" && !strings.Contains(jsonTag, "omitempty") {
				schema.Required = append(schema.Required, fieldName)
			}
		}

	case reflect.Interface:
		// Interface types can be anything
		schema.Type = "object"

	default:
		return nil, fmt.Errorf("unsupported type: %v", t.Kind())
	}

	return schema, nil
}

// Grammar represents a context-free grammar for constrained decoding
type Grammar struct {
	Rules     map[string]*GrammarRule `json:"rules"`
	StartRule string                  `json:"start_rule"`
}

// GrammarRule represents a rule in the grammar
type GrammarRule struct {
	Name         string     `json:"name"`
	Alternatives [][]string `json:"alternatives"`
	Terminal     bool       `json:"terminal"`
	Pattern      string     `json:"pattern,omitempty"`
}

// NewGrammar creates a new grammar with basic rules
func NewGrammar() *Grammar {
	g := &Grammar{
		Rules: make(map[string]*GrammarRule),
	}

	// Add basic JSON rules
	g.addRule("json", [][]string{
		{"object"},
		{"array"},
		{"string"},
		{"number"},
		{"boolean"},
		{"null"},
	}, false, "")

	g.addRule("object", [][]string{
		{`{`, "members", `}`},
		{`{`, `}`},
	}, false, "")

	g.addRule("members", [][]string{
		{"pair", `,`, "members"},
		{"pair"},
	}, false, "")

	g.addRule("pair", [][]string{
		{"string", `:`, "json"},
	}, false, "")

	g.addRule("array", [][]string{
		{`[`, "elements", `]`},
		{`[`, `]`},
	}, false, "")

	g.addRule("elements", [][]string{
		{"json", `,`, "elements"},
		{"json"},
	}, false, "")

	g.addRule("string", nil, true, `"[^"\\]*(?:\\.[^"\\]*)*"`)
	g.addRule("number", nil, true, `-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`)
	g.addRule("boolean", nil, true, `true|false`)
	g.addRule("null", nil, true, `null`)

	g.StartRule = "json"

	return g
}

func (g *Grammar) addRule(name string, alternatives [][]string, terminal bool, pattern string) {
	g.Rules[name] = &GrammarRule{
		Name:         name,
		Alternatives: alternatives,
		Terminal:     terminal,
		Pattern:      pattern,
	}
}

// GrammarFromSchema generates a grammar from a JSON Schema
func GrammarFromSchema(schema *Schema) (*Grammar, error) {
	g := NewGrammar()

	// Override the json rule with schema-specific rule
	rootRule, err := g.ruleFromSchema("root", schema)
	if err != nil {
		return nil, err
	}

	g.Rules["root"] = rootRule
	g.StartRule = "root"

	return g, nil
}

func (g *Grammar) ruleFromSchema(name string, schema *Schema) (*GrammarRule, error) {
	rule := &GrammarRule{Name: name}

	switch schema.Type {
	case "string":
		rule.Terminal = true
		if schema.Pattern != "" {
			rule.Pattern = `"` + schema.Pattern + `"`
		} else if len(schema.Enum) > 0 {
			patterns := make([]string, len(schema.Enum))
			for i, e := range schema.Enum {
				patterns[i] = fmt.Sprintf(`"%v"`, e)
			}
			rule.Pattern = strings.Join(patterns, "|")
		} else {
			rule.Pattern = `"[^"\\]*(?:\\.[^"\\]*)*"`
		}

	case "integer":
		rule.Terminal = true
		rule.Pattern = `-?(?:0|[1-9]\d*)`

	case "number":
		rule.Terminal = true
		rule.Pattern = `-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`

	case "boolean":
		rule.Terminal = true
		rule.Pattern = `true|false`

	case "null":
		rule.Terminal = true
		rule.Pattern = `null`

	case "array":
		rule.Alternatives = [][]string{
			{`[`, name + "_elements", `]`},
			{`[`, `]`},
		}
		if schema.Items != nil {
			itemRule, err := g.ruleFromSchema(name+"_item", schema.Items)
			if err != nil {
				return nil, err
			}
			g.Rules[name+"_item"] = itemRule
			g.addRule(name+"_elements", [][]string{
				{name + "_item", `,`, name + "_elements"},
				{name + "_item"},
			}, false, "")
		}

	case "object":
		if len(schema.Properties) == 0 {
			rule.Alternatives = [][]string{
				{`{`, "members", `}`},
				{`{`, `}`},
			}
		} else {
			// Generate specific object format
			var parts []string
			parts = append(parts, `{`)
			first := true
			for propName, propSchema := range schema.Properties {
				if !first {
					parts = append(parts, `,`)
				}
				first = false

				propRuleName := name + "_" + propName
				propRule, err := g.ruleFromSchema(propRuleName, propSchema)
				if err != nil {
					return nil, err
				}
				g.Rules[propRuleName] = propRule

				parts = append(parts, fmt.Sprintf(`"%s"`, propName), `:`, propRuleName)
			}
			parts = append(parts, `}`)
			rule.Alternatives = [][]string{parts}
		}

	default:
		return nil, fmt.Errorf("unsupported schema type: %s", schema.Type)
	}

	return rule, nil
}

// ValidationResult contains the result of output validation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
	Data     interface{}       `json:"data,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// Validator validates structured output
type Validator interface {
	// Validate validates output against a schema
	Validate(output string, schema *Schema) (*ValidationResult, error)
	// ValidateJSON validates JSON output
	ValidateJSON(output string, schema *Schema) (*ValidationResult, error)
	// Repair attempts to repair invalid output
	Repair(output string, schema *Schema) (string, error)
}

// SchemaValidator validates output against JSON schemas
type SchemaValidator struct {
	strictMode bool
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(strictMode bool) *SchemaValidator {
	return &SchemaValidator{strictMode: strictMode}
}

// Validate validates output against a schema
func (v *SchemaValidator) Validate(output string, schema *Schema) (*ValidationResult, error) {
	// Try to parse as JSON
	return v.ValidateJSON(output, schema)
}

// ValidateJSON validates JSON output against a schema
func (v *SchemaValidator) ValidateJSON(output string, schema *Schema) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	// Parse JSON
	var data interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:    "$",
			Message: fmt.Sprintf("Invalid JSON: %v", err),
			Value:   truncate(output, 100),
		})
		return result, nil
	}

	result.Data = data

	// Validate against schema
	errors := v.validateValue(data, schema, "$")
	if len(errors) > 0 {
		result.Valid = false
		result.Errors = errors
	}

	return result, nil
}

func (v *SchemaValidator) validateValue(value interface{}, schema *Schema, path string) []ValidationError {
	var errors []ValidationError

	if schema == nil {
		return errors
	}

	switch schema.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "Expected string",
				Value:   fmt.Sprintf("%T", value),
			})
			return errors
		}

		if schema.MinLength != nil && len(str) < *schema.MinLength {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("String too short (min: %d)", *schema.MinLength),
			})
		}
		if schema.MaxLength != nil && len(str) > *schema.MaxLength {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("String too long (max: %d)", *schema.MaxLength),
			})
		}
		if schema.Pattern != "" {
			matched, _ := regexp.MatchString(schema.Pattern, str)
			if !matched {
				errors = append(errors, ValidationError{
					Path:    path,
					Message: fmt.Sprintf("String does not match pattern: %s", schema.Pattern),
				})
			}
		}
		if len(schema.Enum) > 0 {
			found := false
			for _, e := range schema.Enum {
				if e == str {
					found = true
					break
				}
			}
			if !found {
				errors = append(errors, ValidationError{
					Path:    path,
					Message: fmt.Sprintf("Value not in enum: %v", schema.Enum),
				})
			}
		}

	case "integer":
		num, ok := value.(float64)
		if !ok {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "Expected integer",
				Value:   fmt.Sprintf("%T", value),
			})
			return errors
		}
		if num != float64(int64(num)) {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "Expected integer, got float",
			})
		}
		if schema.Minimum != nil && num < *schema.Minimum {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("Value below minimum (%v)", *schema.Minimum),
			})
		}
		if schema.Maximum != nil && num > *schema.Maximum {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("Value above maximum (%v)", *schema.Maximum),
			})
		}

	case "number":
		_, ok := value.(float64)
		if !ok {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "Expected number",
				Value:   fmt.Sprintf("%T", value),
			})
		}

	case "boolean":
		_, ok := value.(bool)
		if !ok {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "Expected boolean",
				Value:   fmt.Sprintf("%T", value),
			})
		}

	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "Expected array",
				Value:   fmt.Sprintf("%T", value),
			})
			return errors
		}

		if schema.MinItems != nil && len(arr) < *schema.MinItems {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("Array too short (min: %d)", *schema.MinItems),
			})
		}
		if schema.MaxItems != nil && len(arr) > *schema.MaxItems {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: fmt.Sprintf("Array too long (max: %d)", *schema.MaxItems),
			})
		}

		if schema.Items != nil {
			for i, item := range arr {
				itemPath := fmt.Sprintf("%s[%d]", path, i)
				errors = append(errors, v.validateValue(item, schema.Items, itemPath)...)
			}
		}

	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			errors = append(errors, ValidationError{
				Path:    path,
				Message: "Expected object",
				Value:   fmt.Sprintf("%T", value),
			})
			return errors
		}

		// Check required properties
		for _, req := range schema.Required {
			if _, exists := obj[req]; !exists {
				errors = append(errors, ValidationError{
					Path:    path + "." + req,
					Message: "Required property missing",
				})
			}
		}

		// Validate properties
		for propName, propSchema := range schema.Properties {
			if propValue, exists := obj[propName]; exists {
				propPath := path + "." + propName
				errors = append(errors, v.validateValue(propValue, propSchema, propPath)...)
			}
		}
	}

	return errors
}

// Repair attempts to repair invalid output
func (v *SchemaValidator) Repair(output string, schema *Schema) (string, error) {
	// Try to extract JSON from markdown code blocks
	if strings.Contains(output, "```json") {
		re := regexp.MustCompile("```json\\s*\\n([\\s\\S]*?)\\n```")
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			output = strings.TrimSpace(matches[1])
		}
	} else if strings.Contains(output, "```") {
		re := regexp.MustCompile("```\\s*\\n([\\s\\S]*?)\\n```")
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			output = strings.TrimSpace(matches[1])
		}
	}

	// Try to fix common JSON issues
	output = strings.TrimSpace(output)

	// Remove trailing commas
	output = regexp.MustCompile(`,\s*([\]}])`).ReplaceAllString(output, "$1")

	// Add missing quotes to keys
	output = regexp.MustCompile(`(\{|,)\s*(\w+)\s*:`).ReplaceAllString(output, `$1"$2":`)

	// Validate the repaired output
	result, err := v.ValidateJSON(output, schema)
	if err != nil {
		return "", err
	}

	if !result.Valid {
		return "", fmt.Errorf("could not repair output: %v", result.Errors)
	}

	// Re-serialize to ensure proper formatting
	data, err := json.MarshalIndent(result.Data, "", "  ")
	if err != nil {
		return output, nil
	}

	return string(data), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
