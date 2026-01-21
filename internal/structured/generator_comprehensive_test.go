package structured

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// JSON Schema Validation Tests
// ============================================================

func TestSchemaValidator_ValidateJSON_Comprehensive(t *testing.T) {
	validator := NewSchemaValidator(false)

	t.Run("Invalid JSON syntax", func(t *testing.T) {
		schema := &Schema{Type: "object"}
		result, err := validator.ValidateJSON(`{invalid json}`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)
		assert.Contains(t, result.Errors[0].Message, "Invalid JSON")
	})

	t.Run("Null schema passes validation", func(t *testing.T) {
		result, err := validator.ValidateJSON(`{"any": "value"}`, nil)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Boolean validation - valid true", func(t *testing.T) {
		schema := &Schema{Type: "boolean"}
		result, err := validator.ValidateJSON(`true`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Boolean validation - valid false", func(t *testing.T) {
		schema := &Schema{Type: "boolean"}
		result, err := validator.ValidateJSON(`false`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Boolean validation - invalid type", func(t *testing.T) {
		schema := &Schema{Type: "boolean"}
		result, err := validator.ValidateJSON(`"true"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Expected boolean")
	})

	t.Run("Number validation - valid float", func(t *testing.T) {
		schema := &Schema{Type: "number"}
		result, err := validator.ValidateJSON(`3.14159`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Number validation - invalid type", func(t *testing.T) {
		schema := &Schema{Type: "number"}
		result, err := validator.ValidateJSON(`"not a number"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Expected number")
	})

	t.Run("Integer validation - decimal number is invalid", func(t *testing.T) {
		schema := &Schema{Type: "integer"}
		result, err := validator.ValidateJSON(`3.5`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Expected integer, got float")
	})

	t.Run("Integer validation - wrong type", func(t *testing.T) {
		schema := &Schema{Type: "integer"}
		result, err := validator.ValidateJSON(`"42"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Expected integer")
	})

	t.Run("Array validation - wrong type", func(t *testing.T) {
		schema := &Schema{Type: "array"}
		result, err := validator.ValidateJSON(`{"not": "array"}`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Expected array")
	})

	t.Run("Array with minItems constraint", func(t *testing.T) {
		minItems := 3
		schema := &Schema{
			Type:     "array",
			MinItems: &minItems,
			Items:    &Schema{Type: "string"},
		}
		result, err := validator.ValidateJSON(`["a", "b"]`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Array too short")
	})

	t.Run("Array with maxItems constraint", func(t *testing.T) {
		maxItems := 2
		schema := &Schema{
			Type:     "array",
			MaxItems: &maxItems,
			Items:    &Schema{Type: "string"},
		}
		result, err := validator.ValidateJSON(`["a", "b", "c"]`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Array too long")
	})

	t.Run("Array with valid item constraints", func(t *testing.T) {
		minItems := 1
		maxItems := 5
		schema := &Schema{
			Type:     "array",
			MinItems: &minItems,
			MaxItems: &maxItems,
			Items:    &Schema{Type: "integer"},
		}
		result, err := validator.ValidateJSON(`[1, 2, 3]`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Array items validation - invalid items", func(t *testing.T) {
		schema := &Schema{
			Type:  "array",
			Items: &Schema{Type: "integer"},
		}
		result, err := validator.ValidateJSON(`[1, "not an int", 3]`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Path, "[1]")
	})

	t.Run("Object validation - wrong type", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
			},
		}
		result, err := validator.ValidateJSON(`["not", "object"]`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Expected object")
	})

	t.Run("Object with multiple required fields missing", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name":  {Type: "string"},
				"email": {Type: "string"},
				"age":   {Type: "integer"},
			},
			Required: []string{"name", "email"},
		}
		result, err := validator.ValidateJSON(`{"age": 25}`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 2)
	})

	t.Run("Object with nested validation", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"user": {
					Type: "object",
					Properties: map[string]*Schema{
						"name": {Type: "string"},
						"age":  {Type: "integer"},
					},
					Required: []string{"name"},
				},
			},
		}
		result, err := validator.ValidateJSON(`{"user": {"age": "invalid"}}`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
	})

	t.Run("String with wrong type", func(t *testing.T) {
		schema := &Schema{Type: "string"}
		result, err := validator.ValidateJSON(`123`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "Expected string")
	})
}

// ============================================================
// Regex Constraint Tests
// ============================================================

func TestSchemaValidator_RegexConstraints(t *testing.T) {
	validator := NewSchemaValidator(false)

	t.Run("Valid email pattern", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		}
		result, err := validator.ValidateJSON(`"test@example.com"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Invalid email pattern", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		}
		result, err := validator.ValidateJSON(`"not-an-email"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors[0].Message, "does not match pattern")
	})

	t.Run("Valid phone pattern", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^\+?[0-9]{10,14}$`,
		}
		result, err := validator.ValidateJSON(`"+12345678901"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Invalid phone pattern", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^\+?[0-9]{10,14}$`,
		}
		result, err := validator.ValidateJSON(`"123-456-7890"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
	})

	t.Run("UUID pattern", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
		}
		result, err := validator.ValidateJSON(`"550e8400-e29b-41d4-a716-446655440000"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Alphanumeric pattern", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^[a-zA-Z0-9]+$`,
		}
		result, err := validator.ValidateJSON(`"abc123XYZ"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Pattern with special chars fails on alphanumeric", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^[a-zA-Z0-9]+$`,
		}
		result, err := validator.ValidateJSON(`"hello@world"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
	})

	t.Run("Date pattern ISO 8601", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `^\d{4}-\d{2}-\d{2}$`,
		}
		result, err := validator.ValidateJSON(`"2024-01-21"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})
}

// ============================================================
// Grammar Constraint Tests
// ============================================================

func TestNewGrammar(t *testing.T) {
	grammar := NewGrammar()

	t.Run("Grammar has start rule", func(t *testing.T) {
		assert.Equal(t, "json", grammar.StartRule)
	})

	t.Run("Grammar has basic JSON rules", func(t *testing.T) {
		expectedRules := []string{"json", "object", "members", "pair", "array", "elements", "string", "number", "boolean", "null"}
		for _, rule := range expectedRules {
			assert.Contains(t, grammar.Rules, rule, "Expected rule %s not found", rule)
		}
	})

	t.Run("String rule is terminal", func(t *testing.T) {
		assert.True(t, grammar.Rules["string"].Terminal)
		assert.NotEmpty(t, grammar.Rules["string"].Pattern)
	})

	t.Run("Number rule is terminal", func(t *testing.T) {
		assert.True(t, grammar.Rules["number"].Terminal)
		assert.NotEmpty(t, grammar.Rules["number"].Pattern)
	})

	t.Run("Boolean rule is terminal", func(t *testing.T) {
		assert.True(t, grammar.Rules["boolean"].Terminal)
		assert.Equal(t, "true|false", grammar.Rules["boolean"].Pattern)
	})

	t.Run("Null rule is terminal", func(t *testing.T) {
		assert.True(t, grammar.Rules["null"].Terminal)
		assert.Equal(t, "null", grammar.Rules["null"].Pattern)
	})

	t.Run("Object rule is non-terminal", func(t *testing.T) {
		assert.False(t, grammar.Rules["object"].Terminal)
		assert.NotEmpty(t, grammar.Rules["object"].Alternatives)
	})

	t.Run("Array rule is non-terminal", func(t *testing.T) {
		assert.False(t, grammar.Rules["array"].Terminal)
		assert.NotEmpty(t, grammar.Rules["array"].Alternatives)
	})

	t.Run("JSON rule has multiple alternatives", func(t *testing.T) {
		assert.Len(t, grammar.Rules["json"].Alternatives, 6)
	})
}

func TestGrammarFromSchema(t *testing.T) {
	t.Run("Simple string schema", func(t *testing.T) {
		schema := &Schema{Type: "string"}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.Equal(t, "root", grammar.StartRule)
		assert.Contains(t, grammar.Rules, "root")
		assert.True(t, grammar.Rules["root"].Terminal)
	})

	t.Run("String schema with pattern", func(t *testing.T) {
		schema := &Schema{
			Type:    "string",
			Pattern: `[a-z]+`,
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.Contains(t, grammar.Rules["root"].Pattern, "[a-z]+")
	})

	t.Run("String schema with enum", func(t *testing.T) {
		schema := &Schema{
			Type: "string",
			Enum: []interface{}{"red", "green", "blue"},
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.Contains(t, grammar.Rules["root"].Pattern, `"red"`)
		assert.Contains(t, grammar.Rules["root"].Pattern, `"green"`)
		assert.Contains(t, grammar.Rules["root"].Pattern, `"blue"`)
	})

	t.Run("Integer schema", func(t *testing.T) {
		schema := &Schema{Type: "integer"}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.True(t, grammar.Rules["root"].Terminal)
		assert.NotEmpty(t, grammar.Rules["root"].Pattern)
	})

	t.Run("Number schema", func(t *testing.T) {
		schema := &Schema{Type: "number"}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.True(t, grammar.Rules["root"].Terminal)
	})

	t.Run("Boolean schema", func(t *testing.T) {
		schema := &Schema{Type: "boolean"}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.Equal(t, "true|false", grammar.Rules["root"].Pattern)
	})

	t.Run("Null schema", func(t *testing.T) {
		schema := &Schema{Type: "null"}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.Equal(t, "null", grammar.Rules["root"].Pattern)
	})

	t.Run("Array schema", func(t *testing.T) {
		schema := &Schema{
			Type:  "array",
			Items: &Schema{Type: "string"},
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.False(t, grammar.Rules["root"].Terminal)
		assert.Contains(t, grammar.Rules, "root_item")
		assert.Contains(t, grammar.Rules, "root_elements")
	})

	t.Run("Array schema without items", func(t *testing.T) {
		schema := &Schema{
			Type: "array",
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.NotNil(t, grammar.Rules["root"])
	})

	t.Run("Object schema with properties", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.Contains(t, grammar.Rules, "root")
		// Should have property rules
		foundNameRule := false
		foundAgeRule := false
		for ruleName := range grammar.Rules {
			if ruleName == "root_name" {
				foundNameRule = true
			}
			if ruleName == "root_age" {
				foundAgeRule = true
			}
		}
		assert.True(t, foundNameRule || foundAgeRule, "Expected property rules to be created")
	})

	t.Run("Empty object schema", func(t *testing.T) {
		schema := &Schema{
			Type:       "object",
			Properties: map[string]*Schema{},
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.NotNil(t, grammar.Rules["root"])
	})

	t.Run("Object schema without properties", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.NotNil(t, grammar.Rules["root"])
	})

	t.Run("Unsupported schema type", func(t *testing.T) {
		schema := &Schema{Type: "unsupported"}
		_, err := GrammarFromSchema(schema)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported schema type")
	})

	t.Run("Nested array in object", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"items": {
					Type:  "array",
					Items: &Schema{Type: "integer"},
				},
			},
		}
		grammar, err := GrammarFromSchema(schema)
		require.NoError(t, err)
		assert.NotNil(t, grammar)
	})
}

// ============================================================
// Type Coercion Tests (SchemaFromType)
// ============================================================

func TestSchemaFromType_Comprehensive(t *testing.T) {
	t.Run("Pointer to struct", func(t *testing.T) {
		type Simple struct {
			Name string `json:"name"`
		}
		schema, err := SchemaFromType(&Simple{})
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Type)
		assert.Contains(t, schema.Properties, "name")
	})

	t.Run("Struct with all basic types", func(t *testing.T) {
		type AllTypes struct {
			String  string  `json:"string"`
			Int     int     `json:"int"`
			Int8    int8    `json:"int8"`
			Int16   int16   `json:"int16"`
			Int32   int32   `json:"int32"`
			Int64   int64   `json:"int64"`
			Uint    uint    `json:"uint"`
			Uint8   uint8   `json:"uint8"`
			Uint16  uint16  `json:"uint16"`
			Uint32  uint32  `json:"uint32"`
			Uint64  uint64  `json:"uint64"`
			Float32 float32 `json:"float32"`
			Float64 float64 `json:"float64"`
			Bool    bool    `json:"bool"`
		}

		schema, err := SchemaFromType(AllTypes{})
		require.NoError(t, err)

		assert.Equal(t, "string", schema.Properties["string"].Type)
		assert.Equal(t, "integer", schema.Properties["int"].Type)
		assert.Equal(t, "integer", schema.Properties["int8"].Type)
		assert.Equal(t, "integer", schema.Properties["int16"].Type)
		assert.Equal(t, "integer", schema.Properties["int32"].Type)
		assert.Equal(t, "integer", schema.Properties["int64"].Type)
		assert.Equal(t, "integer", schema.Properties["uint"].Type)
		assert.Equal(t, "integer", schema.Properties["uint8"].Type)
		assert.Equal(t, "integer", schema.Properties["uint16"].Type)
		assert.Equal(t, "integer", schema.Properties["uint32"].Type)
		assert.Equal(t, "integer", schema.Properties["uint64"].Type)
		assert.Equal(t, "number", schema.Properties["float32"].Type)
		assert.Equal(t, "number", schema.Properties["float64"].Type)
		assert.Equal(t, "boolean", schema.Properties["bool"].Type)
	})

	t.Run("Struct with slice", func(t *testing.T) {
		type WithSlice struct {
			Items []int `json:"items"`
		}
		schema, err := SchemaFromType(WithSlice{})
		require.NoError(t, err)
		assert.Equal(t, "array", schema.Properties["items"].Type)
		assert.Equal(t, "integer", schema.Properties["items"].Items.Type)
	})

	t.Run("Struct with array", func(t *testing.T) {
		type WithArray struct {
			Values [5]string `json:"values"`
		}
		schema, err := SchemaFromType(WithArray{})
		require.NoError(t, err)
		assert.Equal(t, "array", schema.Properties["values"].Type)
		assert.Equal(t, "string", schema.Properties["values"].Items.Type)
	})

	t.Run("Struct with map", func(t *testing.T) {
		type WithMap struct {
			Data map[string]int `json:"data"`
		}
		schema, err := SchemaFromType(WithMap{})
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Properties["data"].Type)
	})

	t.Run("Struct with unexported field", func(t *testing.T) {
		type WithUnexported struct {
			Name    string `json:"name"`
			private string //nolint:unused
		}
		schema, err := SchemaFromType(WithUnexported{})
		require.NoError(t, err)
		assert.Contains(t, schema.Properties, "name")
		assert.NotContains(t, schema.Properties, "private")
	})

	t.Run("Struct with json:- tag", func(t *testing.T) {
		type WithIgnored struct {
			Name    string `json:"name"`
			Ignored string `json:"-"`
		}
		schema, err := SchemaFromType(WithIgnored{})
		require.NoError(t, err)
		assert.Contains(t, schema.Properties, "name")
		assert.NotContains(t, schema.Properties, "Ignored")
	})

	t.Run("Struct with description tag", func(t *testing.T) {
		type WithDescription struct {
			Name string `json:"name" description:"The user's name"`
		}
		schema, err := SchemaFromType(WithDescription{})
		require.NoError(t, err)
		assert.Equal(t, "The user's name", schema.Properties["name"].Description)
	})

	t.Run("Struct with interface field", func(t *testing.T) {
		type WithInterface struct {
			Data interface{} `json:"data"`
		}
		schema, err := SchemaFromType(WithInterface{})
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Properties["data"].Type)
	})

	t.Run("Struct with pointer field", func(t *testing.T) {
		type Inner struct {
			Value string `json:"value"`
		}
		type WithPointer struct {
			Inner *Inner `json:"inner"`
		}
		schema, err := SchemaFromType(WithPointer{})
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Properties["inner"].Type)
	})

	t.Run("Deeply nested struct", func(t *testing.T) {
		type Level3 struct {
			Value string `json:"value"`
		}
		type Level2 struct {
			Level3 Level3 `json:"level3"`
		}
		type Level1 struct {
			Level2 Level2 `json:"level2"`
		}
		schema, err := SchemaFromType(Level1{})
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Type)
		assert.Equal(t, "object", schema.Properties["level2"].Type)
		assert.Equal(t, "object", schema.Properties["level2"].Properties["level3"].Type)
		assert.Equal(t, "string", schema.Properties["level2"].Properties["level3"].Properties["value"].Type)
	})

	t.Run("Struct with omitempty", func(t *testing.T) {
		type WithOmitempty struct {
			Required string `json:"required"`
			Optional string `json:"optional,omitempty"`
		}
		schema, err := SchemaFromType(WithOmitempty{})
		require.NoError(t, err)
		assert.Contains(t, schema.Required, "required")
		assert.NotContains(t, schema.Required, "optional")
	})

	t.Run("Struct without json tags", func(t *testing.T) {
		type NoTags struct {
			Name string
			Age  int
		}
		schema, err := SchemaFromType(NoTags{})
		require.NoError(t, err)
		assert.Contains(t, schema.Properties, "Name")
		assert.Contains(t, schema.Properties, "Age")
	})
}

// ============================================================
// Generator Tests
// ============================================================

func TestConstrainedGenerator_Comprehensive(t *testing.T) {
	t.Run("Generator with custom config", func(t *testing.T) {
		config := &GeneratorConfig{
			EnableValidation:  true,
			EnableRepair:      false,
			MaxRepairAttempts: 1,
			StrictMode:        true,
			EnableCaching:     true,
			DefaultFormat:     OutputFormatJSON,
		}
		logger := logrus.New()
		generator := NewConstrainedGenerator(config, logger)
		assert.NotNil(t, generator)
	})

	t.Run("Generate with target unmarshaling", func(t *testing.T) {
		generator := NewConstrainedGenerator(nil, nil)
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
		}

		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		var target Person

		req := &GenerationRequest{
			Schema:   schema,
			Response: `{"name": "John", "age": 30}`,
			Target:   &target,
		}

		result, err := generator.Generate(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, result.Validation.Valid)
		assert.Equal(t, "John", target.Name)
		assert.Equal(t, 30, target.Age)
	})

	t.Run("Generate with validation disabled", func(t *testing.T) {
		config := &GeneratorConfig{
			EnableValidation: false,
		}
		generator := NewConstrainedGenerator(config, nil)

		req := &GenerationRequest{
			Schema:   &Schema{Type: "object"},
			Response: `{invalid json}`,
		}

		result, err := generator.Generate(context.Background(), req)
		require.NoError(t, err)
		assert.Nil(t, result.Validation)
	})

	t.Run("Generate with repair disabled", func(t *testing.T) {
		config := &GeneratorConfig{
			EnableValidation: true,
			EnableRepair:     false,
		}
		generator := NewConstrainedGenerator(config, nil)

		req := &GenerationRequest{
			Schema:   &Schema{Type: "object"},
			Response: `{invalid`,
		}

		result, err := generator.Generate(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, result.Validation.Valid)
		assert.False(t, result.Repaired)
	})

	t.Run("Generate with unrepairable output", func(t *testing.T) {
		config := DefaultGeneratorConfig()
		config.MaxRepairAttempts = 1
		generator := NewConstrainedGenerator(config, nil)

		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"count": {Type: "integer"},
			},
			Required: []string{"count"},
		}

		req := &GenerationRequest{
			Schema:   schema,
			Response: `{"count": "not_an_integer"}`,
		}

		result, err := generator.Generate(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, result.Validation.Valid)
	})
}

// ============================================================
// Caching Tests
// ============================================================

func TestConstrainedGenerator_Caching(t *testing.T) {
	t.Run("Cache and retrieve schema", func(t *testing.T) {
		generator := NewConstrainedGenerator(nil, nil)
		schema := &Schema{Type: "string"}

		generator.CacheSchema("test-schema", schema)
		retrieved := generator.GetCachedSchema("test-schema")

		assert.Equal(t, schema, retrieved)
	})

	t.Run("Get non-existent cached schema", func(t *testing.T) {
		generator := NewConstrainedGenerator(nil, nil)
		retrieved := generator.GetCachedSchema("non-existent")
		assert.Nil(t, retrieved)
	})

	t.Run("Cache and retrieve grammar", func(t *testing.T) {
		generator := NewConstrainedGenerator(nil, nil)
		grammar := NewGrammar()

		generator.CacheGrammar("test-grammar", grammar)
		retrieved := generator.GetCachedGrammar("test-grammar")

		assert.Equal(t, grammar, retrieved)
	})

	t.Run("Get non-existent cached grammar", func(t *testing.T) {
		generator := NewConstrainedGenerator(nil, nil)
		retrieved := generator.GetCachedGrammar("non-existent")
		assert.Nil(t, retrieved)
	})

	t.Run("Caching disabled does not cache schema", func(t *testing.T) {
		config := &GeneratorConfig{EnableCaching: false}
		generator := NewConstrainedGenerator(config, nil)
		schema := &Schema{Type: "string"}

		generator.CacheSchema("test-schema", schema)
		retrieved := generator.GetCachedSchema("test-schema")

		assert.Nil(t, retrieved)
	})

	t.Run("Caching disabled does not cache grammar", func(t *testing.T) {
		config := &GeneratorConfig{EnableCaching: false}
		generator := NewConstrainedGenerator(config, nil)
		grammar := NewGrammar()

		generator.CacheGrammar("test-grammar", grammar)
		retrieved := generator.GetCachedGrammar("test-grammar")

		assert.Nil(t, retrieved)
	})
}

// ============================================================
// Prompt Creation Tests
// ============================================================

func TestCreatePromptWithSchema(t *testing.T) {
	generator := NewConstrainedGenerator(nil, nil)

	t.Run("Creates prompt with schema", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
			},
		}

		prompt := generator.CreatePromptWithSchema("Generate a user", schema)

		assert.Contains(t, prompt, "Generate a user")
		assert.Contains(t, prompt, "```json")
		assert.Contains(t, prompt, "Please respond with a JSON object")
		assert.Contains(t, prompt, "strictly adheres to the schema")
	})

	t.Run("Schema is included in prompt", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"email": {Type: "string", Pattern: "^[a-z]+@[a-z]+$"},
			},
		}

		prompt := generator.CreatePromptWithSchema("Create email", schema)
		assert.Contains(t, prompt, "email")
	})
}

func TestCreateFunctionCallingPrompt(t *testing.T) {
	generator := NewConstrainedGenerator(nil, nil)

	t.Run("Creates function calling prompt", func(t *testing.T) {
		functions := []FunctionDef{
			{
				Name:        "search",
				Description: "Search for items",
				Parameters: &Schema{
					Type: "object",
					Properties: map[string]*Schema{
						"query": {Type: "string"},
					},
				},
			},
		}

		prompt := generator.CreateFunctionCallingPrompt("Find items", functions)

		assert.Contains(t, prompt, "Find items")
		assert.Contains(t, prompt, "### search")
		assert.Contains(t, prompt, "Search for items")
		assert.Contains(t, prompt, "function_name")
	})

	t.Run("Multiple functions", func(t *testing.T) {
		functions := []FunctionDef{
			{Name: "create", Description: "Create item"},
			{Name: "delete", Description: "Delete item"},
			{Name: "update", Description: "Update item"},
		}

		prompt := generator.CreateFunctionCallingPrompt("Manage items", functions)

		assert.Contains(t, prompt, "### create")
		assert.Contains(t, prompt, "### delete")
		assert.Contains(t, prompt, "### update")
	})
}

func TestParseFunctionCall_Comprehensive(t *testing.T) {
	generator := NewConstrainedGenerator(nil, nil)

	t.Run("Parse with nested arguments", func(t *testing.T) {
		output := `{"function": "createUser", "arguments": {"user": {"name": "John", "age": 30}}}`
		call, err := generator.ParseFunctionCall(output)
		require.NoError(t, err)
		assert.Equal(t, "createUser", call.Function)
		user := call.Arguments["user"].(map[string]interface{})
		assert.Equal(t, "John", user["name"])
	})

	t.Run("Parse invalid JSON", func(t *testing.T) {
		output := `not valid json`
		_, err := generator.ParseFunctionCall(output)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse function call")
	})

	t.Run("Parse with code block markers", func(t *testing.T) {
		output := "```\n{\"function\": \"test\", \"arguments\": {}}\n```"
		call, err := generator.ParseFunctionCall(output)
		require.NoError(t, err)
		assert.Equal(t, "test", call.Function)
	})

	t.Run("Parse with array arguments", func(t *testing.T) {
		output := `{"function": "process", "arguments": {"ids": [1, 2, 3]}}`
		call, err := generator.ParseFunctionCall(output)
		require.NoError(t, err)
		ids := call.Arguments["ids"].([]interface{})
		assert.Len(t, ids, 3)
	})
}

// ============================================================
// Output Formatter Tests
// ============================================================

func TestOutputFormatter_Comprehensive(t *testing.T) {
	t.Run("Format JSON without indent", func(t *testing.T) {
		config := &FormatterConfig{IndentJSON: false}
		formatter := NewOutputFormatter(config)

		data := map[string]interface{}{"name": "John"}
		output, err := formatter.FormatJSON(data)
		require.NoError(t, err)
		assert.Equal(t, `{"name":"John"}`, output)
	})

	t.Run("Format JSON with custom indent", func(t *testing.T) {
		config := &FormatterConfig{IndentJSON: true, IndentSize: 4}
		formatter := NewOutputFormatter(config)

		data := map[string]interface{}{"name": "John"}
		output, err := formatter.FormatJSON(data)
		require.NoError(t, err)
		assert.Contains(t, output, "    ") // 4 spaces
	})

	t.Run("Format JSON error", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		// Create an unmarshallable value
		data := make(chan int)
		_, err := formatter.FormatJSON(data)
		assert.Error(t, err)
	})

	t.Run("FormatJSONLines error", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := []interface{}{make(chan int)}
		_, err := formatter.FormatJSONLines(data)
		assert.Error(t, err)
	})

	t.Run("Format Markdown with array", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		}

		output, err := formatter.FormatMarkdown(data)
		require.NoError(t, err)
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "Bob")
	})

	t.Run("Format Markdown with primitive", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		output, err := formatter.FormatMarkdown("simple string")
		require.NoError(t, err)
		assert.Equal(t, "simple string", output)
	})

	t.Run("Format Markdown with nested map", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "John",
				"profile": map[string]interface{}{
					"age": 30,
				},
			},
		}

		output, err := formatter.FormatMarkdown(data)
		require.NoError(t, err)
		assert.Contains(t, output, "**user**")
		assert.Contains(t, output, "**name**")
	})

	t.Run("Format Markdown with array of primitives", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := []interface{}{"one", "two", "three"}

		output, err := formatter.FormatMarkdown(data)
		require.NoError(t, err)
		assert.Contains(t, output, "- one")
		assert.Contains(t, output, "- two")
	})

	t.Run("Format Markdown with nested arrays", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"id": 1},
				map[string]interface{}{"id": 2},
			},
		}

		output, err := formatter.FormatMarkdown(data)
		require.NoError(t, err)
		assert.Contains(t, output, "**items**")
	})

	t.Run("Format CSV with empty data", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		output, err := formatter.FormatCSV([]map[string]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, "", output)
	})

	t.Run("Format CSV with special characters", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := []map[string]interface{}{
			{"name": "John, Jr.", "description": "A \"test\" value"},
		}

		output, err := formatter.FormatCSV(data)
		require.NoError(t, err)
		assert.Contains(t, output, `"John, Jr."`)
		assert.Contains(t, output, `""test""`)
	})

	t.Run("Format CSV with newlines", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := []map[string]interface{}{
			{"text": "line1\nline2"},
		}

		output, err := formatter.FormatCSV(data)
		require.NoError(t, err)
		assert.Contains(t, output, `"line1`)
	})

	t.Run("Format CSV with missing values", func(t *testing.T) {
		formatter := NewOutputFormatter(nil)
		data := []map[string]interface{}{
			{"name": "John", "age": "30"},
			{"name": "Jane"}, // missing age
		}

		output, err := formatter.FormatCSV(data)
		require.NoError(t, err)
		assert.Contains(t, output, "John")
		assert.Contains(t, output, "Jane")
	})
}

// ============================================================
// Repair Tests
// ============================================================

func TestSchemaValidator_Repair_Comprehensive(t *testing.T) {
	validator := NewSchemaValidator(false)

	t.Run("Extract from plain code block", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"value": {Type: "string"},
			},
		}

		input := "```\n{\"value\": \"test\"}\n```"
		repaired, err := validator.Repair(input, schema)
		require.NoError(t, err)
		assert.Contains(t, repaired, "test")
	})

	t.Run("Fix unquoted keys", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
			},
		}

		input := `{name: "John"}`
		repaired, err := validator.Repair(input, schema)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(repaired), &result)
		require.NoError(t, err)
		assert.Equal(t, "John", result["name"])
	})

	t.Run("Fix trailing comma in array", func(t *testing.T) {
		schema := &Schema{
			Type:  "array",
			Items: &Schema{Type: "integer"},
		}

		input := `[1, 2, 3,]`
		repaired, err := validator.Repair(input, schema)
		require.NoError(t, err)

		var result []interface{}
		err = json.Unmarshal([]byte(repaired), &result)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("Unrepairable output", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"count": {Type: "integer"},
			},
			Required: []string{"count"},
		}

		input := `totally not json at all`
		_, err := validator.Repair(input, schema)
		assert.Error(t, err)
	})
}

// ============================================================
// Truncate Tests
// ============================================================

func TestTruncate(t *testing.T) {
	t.Run("String shorter than max", func(t *testing.T) {
		result := truncate("short", 10)
		assert.Equal(t, "short", result)
	})

	t.Run("String equal to max", func(t *testing.T) {
		result := truncate("exactly10c", 10)
		assert.Equal(t, "exactly10c", result)
	})

	t.Run("String longer than max", func(t *testing.T) {
		result := truncate("this is a very long string", 10)
		assert.Equal(t, "this is a ...", result)
	})

	t.Run("Empty string", func(t *testing.T) {
		result := truncate("", 10)
		assert.Equal(t, "", result)
	})
}

// ============================================================
// Edge Case Tests
// ============================================================

func TestEdgeCases(t *testing.T) {
	t.Run("Validate empty object against empty schema", func(t *testing.T) {
		validator := NewSchemaValidator(false)
		schema := &Schema{Type: "object"}
		result, err := validator.ValidateJSON(`{}`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Validate empty array against empty schema", func(t *testing.T) {
		validator := NewSchemaValidator(false)
		schema := &Schema{Type: "array"}
		result, err := validator.ValidateJSON(`[]`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Validate with Data field populated", func(t *testing.T) {
		validator := NewSchemaValidator(false)
		schema := &Schema{Type: "object"}
		result, err := validator.ValidateJSON(`{"key": "value"}`, schema)
		require.NoError(t, err)
		assert.NotNil(t, result.Data)
		data := result.Data.(map[string]interface{})
		assert.Equal(t, "value", data["key"])
	})

	t.Run("DefaultFormatterConfig values", func(t *testing.T) {
		config := DefaultFormatterConfig()
		assert.True(t, config.IndentJSON)
		assert.Equal(t, 2, config.IndentSize)
		assert.False(t, config.SortKeys)
		assert.False(t, config.EscapeHTML)
	})

	t.Run("DefaultGeneratorConfig values", func(t *testing.T) {
		config := DefaultGeneratorConfig()
		assert.True(t, config.EnableValidation)
		assert.True(t, config.EnableRepair)
		assert.Equal(t, 3, config.MaxRepairAttempts)
		assert.False(t, config.StrictMode)
		assert.True(t, config.EnableCaching)
		assert.Equal(t, OutputFormatJSON, config.DefaultFormat)
	})

	t.Run("Number minimum/maximum validation", func(t *testing.T) {
		validator := NewSchemaValidator(false)
		min := 5.0
		max := 10.0
		schema := &Schema{
			Type:    "number",
			Minimum: &min,
			Maximum: &max,
		}

		result, err := validator.ValidateJSON(`7.5`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("Generator Validate method", func(t *testing.T) {
		validator := NewSchemaValidator(false)
		schema := &Schema{Type: "string"}
		result, err := validator.Validate(`"test"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})
}

// ============================================================
// Concurrent Access Tests
// ============================================================

func TestConcurrentAccess(t *testing.T) {
	t.Run("Concurrent cache operations", func(t *testing.T) {
		generator := NewConstrainedGenerator(nil, nil)

		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func(idx int) {
				schema := &Schema{Type: "string"}
				key := "test-schema"
				generator.CacheSchema(key, schema)
				_ = generator.GetCachedSchema(key)
				done <- true
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})

	t.Run("Concurrent grammar cache operations", func(t *testing.T) {
		generator := NewConstrainedGenerator(nil, nil)

		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func(idx int) {
				grammar := NewGrammar()
				key := "test-grammar"
				generator.CacheGrammar(key, grammar)
				_ = generator.GetCachedGrammar(key)
				done <- true
			}(i)
		}

		for i := 0; i < 100; i++ {
			<-done
		}
	})
}
