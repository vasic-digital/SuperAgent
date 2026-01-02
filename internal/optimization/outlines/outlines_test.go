package outlines

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaBuilder(t *testing.T) {
	schema := NewSchemaBuilder().
		Object().
		Property("name", StringSchema()).
		Property("age", IntegerSchema()).
		Property("email", PatternString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)).
		Required("name", "age").
		Build()

	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 3)
	assert.Contains(t, schema.Required, "name")
	assert.Contains(t, schema.Required, "age")
}

func TestSchemaBuilder_Array(t *testing.T) {
	schema := NewSchemaBuilder().
		Array().
		Items(StringSchema()).
		MinItems(1).
		MaxItems(10).
		Build()

	assert.Equal(t, "array", schema.Type)
	assert.NotNil(t, schema.Items)
	assert.Equal(t, 1, *schema.MinItems)
	assert.Equal(t, 10, *schema.MaxItems)
}

func TestSchemaBuilder_String(t *testing.T) {
	schema := NewSchemaBuilder().
		String().
		MinLength(5).
		MaxLength(100).
		Pattern(`^[A-Z]`).
		Build()

	assert.Equal(t, "string", schema.Type)
	assert.Equal(t, 5, *schema.MinLength)
	assert.Equal(t, 100, *schema.MaxLength)
	assert.Equal(t, `^[A-Z]`, schema.Pattern)
}

func TestSchemaBuilder_Number(t *testing.T) {
	schema := NewSchemaBuilder().
		Number().
		Minimum(0).
		Maximum(100).
		Build()

	assert.Equal(t, "number", schema.Type)
	assert.Equal(t, float64(0), *schema.Minimum)
	assert.Equal(t, float64(100), *schema.Maximum)
}

func TestParseSchema(t *testing.T) {
	jsonSchema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		},
		"required": ["name"]
	}`

	schema, err := ParseSchema([]byte(jsonSchema))
	require.NoError(t, err)

	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 2)
	assert.True(t, schema.IsRequired("name"))
	assert.False(t, schema.IsRequired("age"))
}

func TestSchemaValidator_ValidObject(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name", "age")

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	result := validator.Validate(`{"name": "John", "age": 30}`)

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestSchemaValidator_MissingRequired(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name", "age")

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	result := validator.Validate(`{"name": "John"}`)

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "required property missing")
}

func TestSchemaValidator_InvalidType(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	})

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	result := validator.Validate(`{"name": "John", "age": "thirty"}`)

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "expected integer")
}

func TestSchemaValidator_StringConstraints(t *testing.T) {
	minLen := 5
	maxLen := 10
	schema := &JSONSchema{
		Type:      "string",
		MinLength: &minLen,
		MaxLength: &maxLen,
	}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Too short
	result := validator.Validate(`"abc"`)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0].Message, "at least 5 characters")

	// Too long
	result = validator.Validate(`"abcdefghijk"`)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0].Message, "at most 10 characters")

	// Valid
	result = validator.Validate(`"abcdef"`)
	assert.True(t, result.Valid)
}

func TestSchemaValidator_NumberConstraints(t *testing.T) {
	min := float64(0)
	max := float64(100)
	schema := &JSONSchema{
		Type:    "number",
		Minimum: &min,
		Maximum: &max,
	}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Too low
	result := validator.Validate(`-1`)
	assert.False(t, result.Valid)

	// Too high
	result = validator.Validate(`101`)
	assert.False(t, result.Valid)

	// Valid
	result = validator.Validate(`50`)
	assert.True(t, result.Valid)
}

func TestSchemaValidator_Pattern(t *testing.T) {
	schema := PatternString(`^[A-Z][a-z]+$`)

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Doesn't match
	result := validator.Validate(`"john"`)
	assert.False(t, result.Valid)

	// Matches
	result = validator.Validate(`"John"`)
	assert.True(t, result.Valid)
}

func TestSchemaValidator_Enum(t *testing.T) {
	schema := EnumSchema("red", "green", "blue")

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Invalid value
	result := validator.Validate(`"yellow"`)
	assert.False(t, result.Valid)

	// Valid value
	result = validator.Validate(`"red"`)
	assert.True(t, result.Valid)
}

func TestSchemaValidator_Array(t *testing.T) {
	minItems := 1
	maxItems := 3
	schema := &JSONSchema{
		Type:     "array",
		Items:    IntegerSchema(),
		MinItems: &minItems,
		MaxItems: &maxItems,
	}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Empty array (too few)
	result := validator.Validate(`[]`)
	assert.False(t, result.Valid)

	// Too many items
	result = validator.Validate(`[1, 2, 3, 4]`)
	assert.False(t, result.Valid)

	// Invalid item type
	result = validator.Validate(`[1, "two", 3]`)
	assert.False(t, result.Valid)

	// Valid
	result = validator.Validate(`[1, 2, 3]`)
	assert.True(t, result.Valid)
}

func TestSchemaValidator_NestedObject(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"user": ObjectSchema(map[string]*JSONSchema{
			"name":  StringSchema(),
			"email": StringSchema(),
		}, "name"),
		"active": BooleanSchema(),
	}, "user")

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Valid nested
	result := validator.Validate(`{"user": {"name": "John", "email": "john@example.com"}, "active": true}`)
	assert.True(t, result.Valid)

	// Missing nested required
	result = validator.Validate(`{"user": {"email": "john@example.com"}}`)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0].Path, "user.name")
}

func TestSchemaValidator_Formats(t *testing.T) {
	tests := []struct {
		format string
		valid  string
		invalid string
	}{
		{"email", `"test@example.com"`, `"not-an-email"`},
		{"uri", `"https://example.com"`, `"not-a-uri"`},
		{"date", `"2024-01-15"`, `"15-01-2024"`},
		{"uuid", `"550e8400-e29b-41d4-a716-446655440000"`, `"not-a-uuid"`},
		{"ipv4", `"192.168.1.1"`, `"999.999.999.999"`},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			schema := &JSONSchema{Type: "string", Format: tt.format}
			validator, err := NewSchemaValidator(schema)
			require.NoError(t, err)

			result := validator.Validate(tt.valid)
			assert.True(t, result.Valid, "Expected %s to be valid for format %s", tt.valid, tt.format)

			result = validator.Validate(tt.invalid)
			assert.False(t, result.Valid, "Expected %s to be invalid for format %s", tt.invalid, tt.format)
		})
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain json",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "json with text",
			input:    `Here is the result: {"key": "value"} and some more text`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "json in markdown",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "json array",
			input:    `[1, 2, 3]`,
			expected: `[1, 2, 3]`,
		},
		{
			name:     "nested json",
			input:    `{"outer": {"inner": "value"}}`,
			expected: `{"outer": {"inner": "value"}}`,
		},
		{
			name:     "no json",
			input:    `Just plain text`,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindMatchingBrace(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{`{}`, 1},
		{`{"key": "value"}`, 15},
		{`{"nested": {"key": "value"}}`, 27},
		{`{"array": [1, 2, 3]}`, 19},
		{`{`, -1},
		{`{"key": "value with { brace"}`, 28},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := findMatchingBrace(tt.input, '{', '}')
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock LLM provider for testing
type mockProvider struct {
	responses []string
	callCount int
}

func (m *mockProvider) Complete(ctx context.Context, prompt string) (string, error) {
	if m.callCount < len(m.responses) {
		response := m.responses[m.callCount]
		m.callCount++
		return response, nil
	}
	return "", nil
}

func TestStructuredGenerator_ValidResponse(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name", "age")

	provider := &mockProvider{
		responses: []string{`{"name": "John", "age": 30}`},
	}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate a person")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, 0, result.Retries)

	data := result.ParsedData.(map[string]interface{})
	assert.Equal(t, "John", data["name"])
}

func TestStructuredGenerator_RetryOnInvalid(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name", "age")

	provider := &mockProvider{
		responses: []string{
			`{"name": "John"}`,                    // Missing age
			`{"name": "John", "age": "thirty"}`,   // Wrong type
			`{"name": "John", "age": 30}`,         // Valid
		},
	}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate a person")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, 2, result.Retries)
}

func TestStructuredGenerator_StrictMode(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
	}, "name")

	config := DefaultGeneratorConfig()
	config.StrictMode = true

	provider := &mockProvider{
		responses: []string{`not valid json`},
	}

	generator, err := NewStructuredGenerator(provider, schema, config)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate")
	require.NoError(t, err)

	assert.False(t, result.Valid)
	assert.Equal(t, 0, result.Retries)
}

func TestRegexGenerator(t *testing.T) {
	provider := &mockProvider{
		responses: []string{`ABC123`},
	}

	generator, err := NewRegexGenerator(provider, `^[A-Z]{3}[0-9]{3}$`, nil)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate a code")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, "ABC123", result.Content)
}

func TestChoiceGenerator(t *testing.T) {
	provider := &mockProvider{
		responses: []string{`red`},
	}

	generator := NewChoiceGenerator(provider, []string{"red", "green", "blue"}, nil)

	result, err := generator.Generate(context.Background(), "Pick a color")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, "red", result.Content)
}

func TestChoiceGenerator_CaseInsensitive(t *testing.T) {
	provider := &mockProvider{
		responses: []string{`RED`},
	}

	generator := NewChoiceGenerator(provider, []string{"red", "green", "blue"}, nil)

	result, err := generator.Generate(context.Background(), "Pick a color")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, "red", result.Content) // Returns canonical form
}

func TestCompiledPattern(t *testing.T) {
	pattern, err := CompilePattern(`^[A-Z][a-z]+$`)
	require.NoError(t, err)

	assert.True(t, pattern.Match("John"))
	assert.False(t, pattern.Match("john"))
	assert.False(t, pattern.Match("JOHN"))
}

func TestValidate_Convenience(t *testing.T) {
	schema := StringSchema()

	result := Validate(`"hello"`, schema)
	assert.True(t, result.Valid)

	result = Validate(`123`, schema)
	assert.False(t, result.Valid)
}

func TestSchemaValidator_OneOf(t *testing.T) {
	schema := &JSONSchema{
		OneOf: []*JSONSchema{
			StringSchema(),
			IntegerSchema(),
		},
	}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// String is valid
	result := validator.Validate(`"hello"`)
	assert.True(t, result.Valid)

	// Integer is valid
	result = validator.Validate(`42`)
	assert.True(t, result.Valid)

	// Array is not valid
	result = validator.Validate(`[1, 2, 3]`)
	assert.False(t, result.Valid)
}

func TestSchemaValidator_AnyOf(t *testing.T) {
	schema := &JSONSchema{
		AnyOf: []*JSONSchema{
			{Type: "string", MinLength: intPtr(5)},
			{Type: "integer"},
		},
	}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Long string is valid
	result := validator.Validate(`"hello world"`)
	assert.True(t, result.Valid)

	// Short string is invalid
	result = validator.Validate(`"hi"`)
	assert.False(t, result.Valid)

	// Integer is valid
	result = validator.Validate(`42`)
	assert.True(t, result.Valid)
}

func intPtr(i int) *int {
	return &i
}

func BenchmarkSchemaValidator(b *testing.B) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name":  StringSchema(),
		"age":   IntegerSchema(),
		"email": PatternString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		"tags":  ArraySchema(StringSchema()),
	}, "name", "age")

	validator, _ := NewSchemaValidator(schema)
	jsonStr := `{"name": "John Doe", "age": 30, "email": "john@example.com", "tags": ["developer", "go"]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(jsonStr)
	}
}

func BenchmarkExtractJSON(b *testing.B) {
	input := `Here is your result: {"name": "John", "age": 30, "items": [1, 2, 3]} and some more text after`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractJSON(input)
	}
}

func TestDefaultGeneratorConfig(t *testing.T) {
	config := DefaultGeneratorConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.False(t, config.StrictMode)
	assert.True(t, config.IncludeSchemaInPrompt)
	assert.True(t, config.IncludeErrorFeedback)
}

func TestStructuredGenerator_GenerateWithType(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name", "age")

	provider := &mockProvider{
		responses: []string{`{"name": "Alice", "age": 25}`},
	}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var person Person
	err = generator.GenerateWithType(context.Background(), "Generate a person", &person)
	require.NoError(t, err)

	assert.Equal(t, "Alice", person.Name)
	assert.Equal(t, 25, person.Age)
}

func TestStructuredGenerator_GenerateWithType_ValidationFails(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name", "age")

	config := DefaultGeneratorConfig()
	config.MaxRetries = 0
	config.StrictMode = true

	provider := &mockProvider{
		responses: []string{`{"name": "John"}`}, // Missing age
	}

	generator, err := NewStructuredGenerator(provider, schema, config)
	require.NoError(t, err)

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var person Person
	err = generator.GenerateWithType(context.Background(), "Generate a person", &person)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

type errorProvider struct {
	err error
}

func (e *errorProvider) Complete(ctx context.Context, prompt string) (string, error) {
	return "", e.err
}

func TestStructuredGenerator_ProviderError(t *testing.T) {
	schema := StringSchema()

	provider := &errorProvider{err: context.DeadlineExceeded}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)

	_, err = generator.Generate(context.Background(), "Test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM completion failed")
}

func TestRegexGenerator_ProviderError(t *testing.T) {
	provider := &errorProvider{err: context.DeadlineExceeded}

	generator, err := NewRegexGenerator(provider, `^test$`, nil)
	require.NoError(t, err)

	_, err = generator.Generate(context.Background(), "Test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM completion failed")
}

func TestChoiceGenerator_ProviderError(t *testing.T) {
	provider := &errorProvider{err: context.DeadlineExceeded}

	generator := NewChoiceGenerator(provider, []string{"a", "b"}, nil)

	_, err := generator.Generate(context.Background(), "Test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM completion failed")
}

func TestRegexGenerator_Retry(t *testing.T) {
	provider := &mockProvider{
		responses: []string{
			"not matching",
			"still not matching",
			"ABC123", // Matches
		},
	}

	generator, err := NewRegexGenerator(provider, `^[A-Z]{3}[0-9]{3}$`, nil)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate a code")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, 2, result.Retries)
}

func TestRegexGenerator_StrictMode(t *testing.T) {
	config := DefaultGeneratorConfig()
	config.StrictMode = true

	provider := &mockProvider{
		responses: []string{"not matching"},
	}

	generator, err := NewRegexGenerator(provider, `^[A-Z]{3}[0-9]{3}$`, config)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate a code")
	require.NoError(t, err)

	assert.False(t, result.Valid)
	assert.Equal(t, 0, result.Retries)
	assert.Contains(t, result.Errors[0], "does not match pattern")
}

func TestRegexGenerator_MaxRetries(t *testing.T) {
	config := DefaultGeneratorConfig()
	config.MaxRetries = 2

	provider := &mockProvider{
		responses: []string{"bad1", "bad2", "bad3"}, // All bad
	}

	generator, err := NewRegexGenerator(provider, `^[A-Z]{3}[0-9]{3}$`, config)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate a code")
	require.NoError(t, err)

	assert.False(t, result.Valid)
	assert.Equal(t, 2, result.Retries)
}

func TestChoiceGenerator_ContainsMatch(t *testing.T) {
	provider := &mockProvider{
		responses: []string{"I think the answer is red."},
	}

	generator := NewChoiceGenerator(provider, []string{"red", "green", "blue"}, nil)

	result, err := generator.Generate(context.Background(), "Pick a color")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, "red", result.Content)
}

func TestChoiceGenerator_Retry(t *testing.T) {
	provider := &mockProvider{
		responses: []string{
			"yellow",
			"purple",
			"green",
		},
	}

	generator := NewChoiceGenerator(provider, []string{"red", "green", "blue"}, nil)

	result, err := generator.Generate(context.Background(), "Pick a color")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, "green", result.Content)
	assert.Equal(t, 2, result.Retries)
}

func TestChoiceGenerator_StrictMode(t *testing.T) {
	config := DefaultGeneratorConfig()
	config.StrictMode = true

	provider := &mockProvider{
		responses: []string{"yellow"},
	}

	generator := NewChoiceGenerator(provider, []string{"red", "green", "blue"}, config)

	result, err := generator.Generate(context.Background(), "Pick a color")
	require.NoError(t, err)

	assert.False(t, result.Valid)
	assert.Equal(t, 0, result.Retries)
}

func TestChoiceGenerator_MaxRetries(t *testing.T) {
	config := DefaultGeneratorConfig()
	config.MaxRetries = 1

	provider := &mockProvider{
		responses: []string{"yellow", "purple"},
	}

	generator := NewChoiceGenerator(provider, []string{"red", "green", "blue"}, config)

	result, err := generator.Generate(context.Background(), "Pick a color")
	require.NoError(t, err)

	assert.False(t, result.Valid)
	assert.Equal(t, 1, result.Retries)
}

func TestNewRegexGenerator_InvalidPattern(t *testing.T) {
	provider := &mockProvider{}

	_, err := NewRegexGenerator(provider, `[invalid`, nil)

	assert.Error(t, err)
}

func TestNewStructuredGenerator_NilConfig(t *testing.T) {
	schema := StringSchema()
	provider := &mockProvider{responses: []string{`"test"`}}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)
	assert.NotNil(t, generator)
}

func TestStructuredGenerator_MaxRetriesExhausted(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
	}, "name")

	config := DefaultGeneratorConfig()
	config.MaxRetries = 2

	provider := &mockProvider{
		responses: []string{
			`{"wrong": "field"}`,
			`{"wrong": "field"}`,
			`{"wrong": "field"}`,
		},
	}

	generator, err := NewStructuredGenerator(provider, schema, config)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate")
	require.NoError(t, err)

	assert.False(t, result.Valid)
	assert.Equal(t, 2, result.Retries)
	assert.Contains(t, result.Errors[0], "failed to generate valid output")
}

func TestStructuredGenerator_NoJSONRetry(t *testing.T) {
	schema := StringSchema()

	provider := &mockProvider{
		responses: []string{
			"Just plain text",
			"Still plain text",
			`"valid json string"`,
		},
	}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate")
	require.NoError(t, err)

	assert.True(t, result.Valid)
	assert.Equal(t, 2, result.Retries)
}

func TestExtractJSON_MarkdownCodeBlock(t *testing.T) {
	input := "Here's the result:\n```\n{\"key\": \"value\"}\n```\nDone."

	result := extractJSON(input)
	assert.Equal(t, `{"key": "value"}`, result)
}

func TestExtractJSON_ArrayWithText(t *testing.T) {
	input := "The array is: [1, 2, 3, 4] and that's it."

	result := extractJSON(input)
	assert.Equal(t, `[1, 2, 3, 4]`, result)
}

func TestSchemaValidator_Boolean(t *testing.T) {
	schema := BooleanSchema()

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	result := validator.Validate(`true`)
	assert.True(t, result.Valid)

	result = validator.Validate(`false`)
	assert.True(t, result.Valid)

	result = validator.Validate(`"true"`)
	assert.False(t, result.Valid)
}

func TestSchemaValidator_Null(t *testing.T) {
	schema := &JSONSchema{Type: "null"}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	result := validator.Validate(`null`)
	assert.True(t, result.Valid)

	result = validator.Validate(`"null"`)
	assert.False(t, result.Valid)
}

func TestSchemaBuilder_Integer(t *testing.T) {
	schema := NewSchemaBuilder().
		Integer().
		Minimum(0).
		Maximum(100).
		Build()

	assert.Equal(t, "integer", schema.Type)
	assert.Equal(t, float64(0), *schema.Minimum)
	assert.Equal(t, float64(100), *schema.Maximum)
}

func TestSchemaBuilder_Boolean(t *testing.T) {
	schema := NewSchemaBuilder().Boolean().Build()

	assert.Equal(t, "boolean", schema.Type)
}

func TestSchemaBuilder_Enum(t *testing.T) {
	schema := NewSchemaBuilder().
		Enum("a", "b", "c").
		Build()

	assert.Len(t, schema.Enum, 3)
}

func TestValidationResult_ErrorMessages(t *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []*ValidationError{
			{Path: "field1", Message: "error 1"},
			{Path: "field2", Message: "error 2"},
		},
	}

	messages := result.ErrorMessages()

	assert.Len(t, messages, 2)
	assert.Contains(t, messages[0], "error 1")
	assert.Contains(t, messages[1], "error 2")
}

func TestArraySchema(t *testing.T) {
	schema := ArraySchema(StringSchema())

	assert.Equal(t, "array", schema.Type)
	assert.NotNil(t, schema.Items)
	assert.Equal(t, "string", schema.Items.Type)
}

func TestParseSchemaFromMap(t *testing.T) {
	schemaMap := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
			"age":  map[string]interface{}{"type": "integer"},
		},
		"required": []interface{}{"name"},
	}

	schema, err := ParseSchemaFromMap(schemaMap)
	require.NoError(t, err)

	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 2)
	assert.Contains(t, schema.Required, "name")
}

func TestJSONSchema_String(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name")

	str := schema.String()

	assert.Contains(t, str, "object")
	assert.Contains(t, str, "name")
}

func TestJSONSchema_GetPropertySchema(t *testing.T) {
	nameSchema := StringSchema()
	ageSchema := IntegerSchema()

	schema := ObjectSchema(map[string]*JSONSchema{
		"name": nameSchema,
		"age":  ageSchema,
	}, "name")

	prop := schema.GetPropertySchema("name")
	assert.NotNil(t, prop)
	assert.Equal(t, "string", prop.Type)

	prop = schema.GetPropertySchema("age")
	assert.NotNil(t, prop)
	assert.Equal(t, "integer", prop.Type)

	prop = schema.GetPropertySchema("nonexistent")
	assert.Nil(t, prop)
}

func TestSchemaBuilder_Type(t *testing.T) {
	schema := NewSchemaBuilder().
		Type("custom").
		Build()

	assert.Equal(t, "custom", schema.Type)
}

func TestSchemaBuilder_Description(t *testing.T) {
	schema := NewSchemaBuilder().
		String().
		Description("A test description").
		Build()

	assert.Equal(t, "A test description", schema.Description)
}

func TestSchemaBuilder_Default(t *testing.T) {
	schema := NewSchemaBuilder().
		String().
		Default("default value").
		Build()

	assert.Equal(t, "default value", schema.Default)
}

func TestSchemaBuilder_Format(t *testing.T) {
	schema := NewSchemaBuilder().
		String().
		Format("email").
		Build()

	assert.Equal(t, "email", schema.Format)
}

func TestNumberSchema(t *testing.T) {
	schema := NumberSchema()

	assert.Equal(t, "number", schema.Type)
}

func TestSchemaValidator_ValidateData(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name")

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	data := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	result := validator.ValidateData(data)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestSchemaValidator_ValidateData_MissingRequired(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
		"age":  IntegerSchema(),
	}, "name", "age")

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	data := map[string]interface{}{
		"name": "John",
	}

	result := validator.ValidateData(data)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestSchemaValidator_ValidateAllOf(t *testing.T) {
	schema := &JSONSchema{
		AllOf: []*JSONSchema{
			{Type: "object", Properties: map[string]*JSONSchema{"name": StringSchema()}},
			{Type: "object", Properties: map[string]*JSONSchema{"age": IntegerSchema()}},
		},
	}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	result := validator.Validate(`{"name": "John", "age": 30}`)
	assert.True(t, result.Valid)
}

func TestSchemaValidator_DateTimeFormat(t *testing.T) {
	schema := NewSchemaBuilder().
		String().
		Format("date-time").
		Build()

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Valid RFC3339 datetime
	result := validator.Validate(`"2024-01-15T10:30:00Z"`)
	assert.True(t, result.Valid)

	// Invalid datetime
	result = validator.Validate(`"not-a-datetime"`)
	assert.False(t, result.Valid)
}

func TestSchemaValidator_TimeFormat(t *testing.T) {
	schema := NewSchemaBuilder().
		String().
		Format("time").
		Build()

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Valid time
	result := validator.Validate(`"10:30:00"`)
	assert.True(t, result.Valid)

	// Invalid time
	result = validator.Validate(`"not-a-time"`)
	assert.False(t, result.Valid)
}

func TestSchemaValidator_IPv6Format(t *testing.T) {
	schema := NewSchemaBuilder().
		String().
		Format("ipv6").
		Build()

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Valid IPv6
	result := validator.Validate(`"2001:0db8:85a3:0000:0000:8a2e:0370:7334"`)
	assert.True(t, result.Valid)

	// Invalid IPv6
	result = validator.Validate(`"not-an-ipv6"`)
	assert.False(t, result.Valid)
}

func TestExtractJSON_EdgeCases(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"name": StringSchema(),
	}, "name")

	provider := &mockProvider{
		responses: []string{`Here is the JSON: {"name": "John"} hope that helps!`},
	}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate a name")
	require.NoError(t, err)

	assert.True(t, result.Valid)
}

func TestExtractJSON_NestedBraces(t *testing.T) {
	schema := ObjectSchema(map[string]*JSONSchema{
		"data": ObjectSchema(map[string]*JSONSchema{
			"nested": StringSchema(),
		}),
	}, "data")

	provider := &mockProvider{
		responses: []string{`{"data": {"nested": "value"}}`},
	}

	generator, err := NewStructuredGenerator(provider, schema, nil)
	require.NoError(t, err)

	result, err := generator.Generate(context.Background(), "Generate nested")
	require.NoError(t, err)

	assert.True(t, result.Valid)
}

func TestSchemaValidator_NumberValidation(t *testing.T) {
	min := float64(0)
	max := float64(100)
	schema := &JSONSchema{
		Type:    "number",
		Minimum: &min,
		Maximum: &max,
	}

	validator, err := NewSchemaValidator(schema)
	require.NoError(t, err)

	// Valid number
	result := validator.Validate(`50.5`)
	assert.True(t, result.Valid)

	// Below minimum
	result = validator.Validate(`-10`)
	assert.False(t, result.Valid)

	// Above maximum
	result = validator.Validate(`150`)
	assert.False(t, result.Valid)
}
