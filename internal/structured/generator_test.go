package structured

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidator_ValidateJSON(t *testing.T) {
	validator := NewSchemaValidator(false)

	t.Run("Valid object", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name"},
		}

		result, err := validator.ValidateJSON(`{"name": "John", "age": 30}`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Missing required field", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name"},
		}

		result, err := validator.ValidateJSON(`{"age": 30}`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("Invalid type", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"age": {Type: "integer"},
			},
		}

		result, err := validator.ValidateJSON(`{"age": "thirty"}`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
	})

	t.Run("Valid array", func(t *testing.T) {
		schema := &Schema{
			Type:  "array",
			Items: &Schema{Type: "string"},
		}

		result, err := validator.ValidateJSON(`["a", "b", "c"]`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("String with enum", func(t *testing.T) {
		schema := &Schema{
			Type: "string",
			Enum: []interface{}{"red", "green", "blue"},
		}

		result, err := validator.ValidateJSON(`"red"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)

		result, err = validator.ValidateJSON(`"yellow"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
	})

	t.Run("String length constraints", func(t *testing.T) {
		minLen := 2
		maxLen := 5
		schema := &Schema{
			Type:      "string",
			MinLength: &minLen,
			MaxLength: &maxLen,
		}

		result, err := validator.ValidateJSON(`"abc"`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)

		result, err = validator.ValidateJSON(`"a"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)

		result, err = validator.ValidateJSON(`"toolong"`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
	})

	t.Run("Number range constraints", func(t *testing.T) {
		min := 0.0
		max := 100.0
		schema := &Schema{
			Type:    "integer",
			Minimum: &min,
			Maximum: &max,
		}

		result, err := validator.ValidateJSON(`50`, schema)
		require.NoError(t, err)
		assert.True(t, result.Valid)

		result, err = validator.ValidateJSON(`-1`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)

		result, err = validator.ValidateJSON(`101`, schema)
		require.NoError(t, err)
		assert.False(t, result.Valid)
	})
}

func TestSchemaValidator_Repair(t *testing.T) {
	validator := NewSchemaValidator(false)

	t.Run("Extract from markdown code block", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
			},
		}

		input := "Here's the response:\n```json\n{\"name\": \"John\"}\n```"
		repaired, err := validator.Repair(input, schema)
		require.NoError(t, err)
		assert.Contains(t, repaired, "John")
	})

	t.Run("Fix trailing comma", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"name": {Type: "string"},
			},
		}

		input := `{"name": "John",}`
		repaired, err := validator.Repair(input, schema)
		require.NoError(t, err)
		assert.NotContains(t, repaired, ",}")
	})
}

func TestSchemaFromType(t *testing.T) {
	type Person struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email,omitempty"`
	}

	schema, err := SchemaFromType(Person{})
	require.NoError(t, err)

	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "age")
	assert.Contains(t, schema.Properties, "email")

	assert.Equal(t, "string", schema.Properties["name"].Type)
	assert.Equal(t, "integer", schema.Properties["age"].Type)

	// Check required fields (non-omitempty)
	assert.Contains(t, schema.Required, "name")
	assert.Contains(t, schema.Required, "age")
	assert.NotContains(t, schema.Required, "email")
}

func TestSchemaFromType_Nested(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	schema, err := SchemaFromType(Person{})
	require.NoError(t, err)

	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "address")
	assert.Equal(t, "object", schema.Properties["address"].Type)
	assert.Contains(t, schema.Properties["address"].Properties, "street")
	assert.Contains(t, schema.Properties["address"].Properties, "city")
}

func TestSchemaFromType_Array(t *testing.T) {
	type Container struct {
		Items []string `json:"items"`
	}

	schema, err := SchemaFromType(Container{})
	require.NoError(t, err)

	assert.Contains(t, schema.Properties, "items")
	assert.Equal(t, "array", schema.Properties["items"].Type)
	assert.NotNil(t, schema.Properties["items"].Items)
	assert.Equal(t, "string", schema.Properties["items"].Items.Type)
}

func TestConstrainedGenerator(t *testing.T) {
	generator := NewConstrainedGenerator(nil, nil)

	t.Run("Valid JSON passes", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"result": {Type: "string"},
			},
		}

		req := &GenerationRequest{
			Schema:   schema,
			Response: `{"result": "success"}`,
		}

		result, err := generator.Generate(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, result.Validation.Valid)
		assert.False(t, result.Repaired)
	})

	t.Run("Invalid JSON triggers repair", func(t *testing.T) {
		schema := &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"result": {Type: "string"},
			},
		}

		req := &GenerationRequest{
			Schema:   schema,
			Response: "```json\n{\"result\": \"success\",}\n```",
		}

		result, err := generator.Generate(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, result.Validation.Valid)
		assert.True(t, result.Repaired)
	})

	t.Run("No schema returns response as-is", func(t *testing.T) {
		req := &GenerationRequest{
			Response: "Just some text",
		}

		result, err := generator.Generate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "Just some text", result.Output)
	})
}

func TestOutputFormatter(t *testing.T) {
	formatter := NewOutputFormatter(nil)

	t.Run("Format JSON", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		output, err := formatter.FormatJSON(data)
		require.NoError(t, err)
		assert.Contains(t, output, "John")
		assert.Contains(t, output, "30")
	})

	t.Run("Format JSON Lines", func(t *testing.T) {
		data := []interface{}{
			map[string]interface{}{"id": 1},
			map[string]interface{}{"id": 2},
		}

		output, err := formatter.FormatJSONLines(data)
		require.NoError(t, err)
		lines := splitLines(output)
		assert.Equal(t, 2, len(lines))
	})

	t.Run("Format Markdown", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		output, err := formatter.FormatMarkdown(data)
		require.NoError(t, err)
		assert.Contains(t, output, "**name**")
		assert.Contains(t, output, "John")
	})

	t.Run("Format CSV", func(t *testing.T) {
		data := []map[string]interface{}{
			{"name": "John", "age": "30"},
			{"name": "Jane", "age": "25"},
		}

		output, err := formatter.FormatCSV(data)
		require.NoError(t, err)
		assert.Contains(t, output, "name")
		assert.Contains(t, output, "John")
		assert.Contains(t, output, "Jane")
	})
}

func TestParseFunctionCall(t *testing.T) {
	generator := NewConstrainedGenerator(nil, nil)

	t.Run("Parse valid function call", func(t *testing.T) {
		output := `{"function": "search", "arguments": {"query": "test"}}`
		call, err := generator.ParseFunctionCall(output)
		require.NoError(t, err)
		assert.Equal(t, "search", call.Function)
		assert.Equal(t, "test", call.Arguments["query"])
	})

	t.Run("Parse from code block", func(t *testing.T) {
		output := "```json\n{\"function\": \"search\", \"arguments\": {}}\n```"
		call, err := generator.ParseFunctionCall(output)
		require.NoError(t, err)
		assert.Equal(t, "search", call.Function)
	})
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			if i > start {
				lines = append(lines, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
