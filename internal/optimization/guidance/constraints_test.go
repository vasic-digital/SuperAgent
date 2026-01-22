package guidance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexConstraint_Valid(t *testing.T) {
	constraint, err := NewRegexConstraint(`^[a-zA-Z]+$`)
	require.NoError(t, err)

	assert.Equal(t, ConstraintTypeRegex, constraint.Type())
	assert.NoError(t, constraint.Validate("Hello"))
	assert.Error(t, constraint.Validate("Hello123"))
	assert.Error(t, constraint.Validate(""))
}

func TestRegexConstraint_Inverted(t *testing.T) {
	constraint, err := NewRegexConstraint(`\d+`)
	require.NoError(t, err)
	constraint.Invert = true

	assert.NoError(t, constraint.Validate("Hello"))
	assert.Error(t, constraint.Validate("Hello123"))
}

func TestRegexConstraint_InvalidPattern(t *testing.T) {
	_, err := NewRegexConstraint(`[invalid`)
	assert.Error(t, err)
}

func TestRegexConstraint_Description(t *testing.T) {
	constraint, _ := NewRegexConstraint(`^\d+$`)
	assert.Contains(t, constraint.Description(), "pattern")

	constraint.Name = "Numbers only"
	assert.Equal(t, "Numbers only", constraint.Description())
}

func TestChoiceConstraint(t *testing.T) {
	constraint := NewChoiceConstraint([]string{"yes", "no", "maybe"})

	assert.Equal(t, ConstraintTypeChoice, constraint.Type())
	assert.NoError(t, constraint.Validate("yes"))
	assert.NoError(t, constraint.Validate("no"))
	assert.NoError(t, constraint.Validate("maybe"))
	assert.Error(t, constraint.Validate("unknown"))
}

func TestChoiceConstraint_CaseInsensitive(t *testing.T) {
	constraint := NewChoiceConstraint([]string{"Yes", "No"})
	constraint.CaseSensitive = false

	assert.NoError(t, constraint.Validate("yes"))
	assert.NoError(t, constraint.Validate("YES"))
	assert.NoError(t, constraint.Validate("Yes"))
}

func TestChoiceConstraint_Multiple(t *testing.T) {
	constraint := NewChoiceConstraint([]string{"red", "green", "blue"})
	constraint.AllowMultiple = true
	constraint.Separator = ","

	assert.NoError(t, constraint.Validate("red,blue"))
	assert.NoError(t, constraint.Validate("red, green"))
	assert.Error(t, constraint.Validate("red,yellow"))
}

func TestLengthConstraint_Characters(t *testing.T) {
	constraint := NewLengthConstraint(5, 20, LengthUnitCharacters)

	assert.Equal(t, ConstraintTypeLength, constraint.Type())
	assert.NoError(t, constraint.Validate("Hello World"))
	assert.Error(t, constraint.Validate("Hi"))
	assert.Error(t, constraint.Validate("This is a very long string that exceeds the maximum"))
}

func TestLengthConstraint_Words(t *testing.T) {
	constraint := NewLengthConstraint(2, 5, LengthUnitWords)

	assert.NoError(t, constraint.Validate("Hello World"))
	assert.NoError(t, constraint.Validate("One two three four five"))
	assert.Error(t, constraint.Validate("One"))
	assert.Error(t, constraint.Validate("One two three four five six seven"))
}

func TestLengthConstraint_Sentences(t *testing.T) {
	constraint := NewLengthConstraint(1, 3, LengthUnitSentences)

	assert.NoError(t, constraint.Validate("Hello. World!"))
	assert.Error(t, constraint.Validate("One. Two. Three. Four."))
}

func TestLengthConstraint_NoMin(t *testing.T) {
	constraint := NewLengthConstraint(0, 10, LengthUnitWords)

	assert.NoError(t, constraint.Validate("Hello"))
	assert.NoError(t, constraint.Validate(""))
}

func TestLengthConstraint_NoMax(t *testing.T) {
	constraint := NewLengthConstraint(2, 0, LengthUnitWords)

	assert.NoError(t, constraint.Validate("Hello World"))
	assert.NoError(t, constraint.Validate("One two three four five six seven eight nine ten"))
}

func TestRangeConstraint(t *testing.T) {
	constraint := NewRangeConstraint(0, 100)

	assert.Equal(t, ConstraintTypeRange, constraint.Type())
	assert.NoError(t, constraint.Validate("50"))
	assert.NoError(t, constraint.Validate("0"))
	assert.NoError(t, constraint.Validate("100"))
	assert.Error(t, constraint.Validate("-1"))
	assert.Error(t, constraint.Validate("101"))
	assert.Error(t, constraint.Validate("not a number"))
}

func TestRangeConstraint_IntegerOnly(t *testing.T) {
	constraint := NewRangeConstraint(0, 100)
	constraint.IntegerOnly = true

	assert.NoError(t, constraint.Validate("50"))
	assert.Error(t, constraint.Validate("50.5"))
}

func TestFormatConstraint_JSON(t *testing.T) {
	constraint := NewFormatConstraint(FormatJSON)

	assert.Equal(t, ConstraintTypeFormat, constraint.Type())
	assert.NoError(t, constraint.Validate(`{"name": "test"}`))
	assert.NoError(t, constraint.Validate(`[1, 2, 3]`))
	assert.Error(t, constraint.Validate(`{invalid json}`))
	assert.Error(t, constraint.Validate(`not json`))
}

func TestFormatConstraint_Email(t *testing.T) {
	constraint := NewFormatConstraint(FormatEmail)

	assert.NoError(t, constraint.Validate("test@example.com"))
	assert.NoError(t, constraint.Validate("user.name+tag@example.co.uk"))
	assert.Error(t, constraint.Validate("invalid"))
	assert.Error(t, constraint.Validate("@example.com"))
}

func TestFormatConstraint_URL(t *testing.T) {
	constraint := NewFormatConstraint(FormatURL)

	assert.NoError(t, constraint.Validate("https://example.com"))
	assert.NoError(t, constraint.Validate("http://localhost:8080/path"))
	assert.Error(t, constraint.Validate("not a url"))
	assert.Error(t, constraint.Validate("ftp://invalid"))
}

func TestFormatConstraint_UUID(t *testing.T) {
	constraint := NewFormatConstraint(FormatUUID)

	assert.NoError(t, constraint.Validate("550e8400-e29b-41d4-a716-446655440000"))
	assert.NoError(t, constraint.Validate("550E8400-E29B-41D4-A716-446655440000"))
	assert.Error(t, constraint.Validate("not-a-uuid"))
	assert.Error(t, constraint.Validate("550e8400-e29b-41d4-a716"))
}

func TestFormatConstraint_IPv4(t *testing.T) {
	constraint := NewFormatConstraint(FormatIPv4)

	assert.NoError(t, constraint.Validate("192.168.1.1"))
	assert.NoError(t, constraint.Validate("10.0.0.1"))
	assert.Error(t, constraint.Validate("not.an.ip.address"))
}

func TestFormatConstraint_PhoneNumber(t *testing.T) {
	constraint := NewFormatConstraint(FormatPhoneNumber)

	assert.NoError(t, constraint.Validate("555-123-4567"))
	assert.NoError(t, constraint.Validate("+1 (555) 123-4567"))
	assert.Error(t, constraint.Validate("not a phone"))
}

func TestSchemaConstraint(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
			"age":  map[string]interface{}{"type": "integer"},
		},
		"required": []interface{}{"name"},
	}

	constraint := NewSchemaConstraint(schema)

	assert.Equal(t, ConstraintTypeSchema, constraint.Type())
	assert.NoError(t, constraint.Validate(`{"name": "John", "age": 30}`))
	assert.NoError(t, constraint.Validate(`{"name": "John"}`))
	assert.Error(t, constraint.Validate(`{"age": 30}`)) // Missing required name
	assert.Error(t, constraint.Validate(`not json`))
}

func TestSchemaConstraint_TypeValidation(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":   map[string]interface{}{"type": "string"},
			"count":  map[string]interface{}{"type": "number"},
			"active": map[string]interface{}{"type": "boolean"},
			"items":  map[string]interface{}{"type": "array"},
			"data":   map[string]interface{}{"type": "object"},
		},
	}

	constraint := NewSchemaConstraint(schema)

	// Valid types
	assert.NoError(t, constraint.Validate(`{"name": "test"}`))
	assert.NoError(t, constraint.Validate(`{"count": 42}`))
	assert.NoError(t, constraint.Validate(`{"active": true}`))
	assert.NoError(t, constraint.Validate(`{"items": [1, 2, 3]}`))
	assert.NoError(t, constraint.Validate(`{"data": {"key": "value"}}`))

	// Invalid types
	assert.Error(t, constraint.Validate(`{"name": 123}`))
	assert.Error(t, constraint.Validate(`{"count": "not a number"}`))
	assert.Error(t, constraint.Validate(`{"active": "yes"}`))
	assert.Error(t, constraint.Validate(`{"items": "not array"}`))
	assert.Error(t, constraint.Validate(`{"data": "not object"}`))
}

func TestCompositeConstraint_All(t *testing.T) {
	lengthConstraint := NewLengthConstraint(5, 50, LengthUnitCharacters)
	choiceConstraint := NewChoiceConstraint([]string{"hello", "world", "hello world"})

	composite := NewCompositeConstraint(CompositeModeAll, lengthConstraint, choiceConstraint)

	assert.Equal(t, ConstraintTypeComposite, composite.Type())
	assert.NoError(t, composite.Validate("hello world"))
	assert.Error(t, composite.Validate("hi"))            // Too short and not a choice
	assert.Error(t, composite.Validate("unknown value")) // Not a choice
}

func TestCompositeConstraint_Any(t *testing.T) {
	emailConstraint := NewFormatConstraint(FormatEmail)
	urlConstraint := NewFormatConstraint(FormatURL)

	composite := NewCompositeConstraint(CompositeModeAny, emailConstraint, urlConstraint)

	assert.NoError(t, composite.Validate("test@example.com"))
	assert.NoError(t, composite.Validate("https://example.com"))
	assert.Error(t, composite.Validate("not email or url"))
}

func TestGrammarConstraint(t *testing.T) {
	grammar := `
		start: greeting
		greeting: "hello" | "hi"
	`

	constraint := NewGrammarConstraint(grammar)

	assert.Equal(t, ConstraintTypeGrammar, constraint.Type())
	assert.NoError(t, constraint.Validate("hello"))
	assert.Error(t, constraint.Validate(""))
}

func TestConstraintBuilder(t *testing.T) {
	builder := NewConstraintBuilder()

	constraint := builder.
		WithLength(10, 100, LengthUnitCharacters).
		WithFormat(FormatJSON).
		BuildAll()

	assert.NotNil(t, constraint)
	assert.NoError(t, constraint.Validate(`{"test": "value"}`)) // 17 chars, valid JSON
	assert.Error(t, constraint.Validate(`{"a":1}`))             // 7 chars, too short (min 10)
	assert.Error(t, constraint.Validate(`invalid json`))        // Not valid JSON
}

func TestConstraintBuilder_BuildAny(t *testing.T) {
	builder := NewConstraintBuilder()

	constraint := builder.
		WithFormat(FormatEmail).
		WithFormat(FormatURL).
		BuildAny()

	assert.NoError(t, constraint.Validate("test@example.com"))
	assert.NoError(t, constraint.Validate("https://example.com"))
}

func TestConstraintBuilder_SingleConstraint(t *testing.T) {
	builder := NewConstraintBuilder()

	constraint := builder.
		WithLength(5, 100, LengthUnitCharacters).
		BuildAll()

	// Should return the single constraint directly, not a composite
	assert.NotNil(t, constraint)
}

func TestConstraint_Hints(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
		contains   string
	}{
		{
			name:       "regex",
			constraint: func() Constraint { c, _ := NewRegexConstraint(`\d+`); return c }(),
			contains:   "pattern",
		},
		{
			name:       "choice",
			constraint: NewChoiceConstraint([]string{"a", "b", "c"}),
			contains:   "Choose from",
		},
		{
			name:       "length",
			constraint: NewLengthConstraint(5, 10, LengthUnitWords),
			contains:   "between",
		},
		{
			name:       "range",
			constraint: NewRangeConstraint(0, 100),
			contains:   "number",
		},
		{
			name:       "format",
			constraint: NewFormatConstraint(FormatJSON),
			contains:   "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := tt.constraint.Hint()
			assert.Contains(t, hint, tt.contains)
		})
	}
}

func TestConstraint_Descriptions(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
	}{
		{"regex", func() Constraint { c, _ := NewRegexConstraint(`\d+`); return c }()},
		{"choice", NewChoiceConstraint([]string{"a", "b"})},
		{"length", NewLengthConstraint(5, 10, LengthUnitWords)},
		{"range", NewRangeConstraint(0, 100)},
		{"format", NewFormatConstraint(FormatJSON)},
		{"schema", NewSchemaConstraint(map[string]interface{}{"type": "object"})},
		{"grammar", NewGrammarConstraint("start: word")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.constraint.Description()
			assert.NotEmpty(t, desc)
		})
	}
}

func TestLengthConstraint_MinOnly(t *testing.T) {
	constraint := NewLengthConstraint(5, 0, LengthUnitWords)

	desc := constraint.Description()
	assert.Contains(t, desc, "at least")
	assert.NotContains(t, desc, "at most")

	hint := constraint.Hint()
	assert.Contains(t, hint, "at least")
}

func TestLengthConstraint_MaxOnly(t *testing.T) {
	constraint := NewLengthConstraint(0, 10, LengthUnitWords)

	desc := constraint.Description()
	assert.Contains(t, desc, "at most")
	assert.NotContains(t, desc, "at least")

	hint := constraint.Hint()
	assert.Contains(t, hint, "at most")
}

func TestRangeConstraint_IntegerHint(t *testing.T) {
	constraint := NewRangeConstraint(0, 100)
	constraint.IntegerOnly = true

	desc := constraint.Description()
	assert.Contains(t, desc, "Integer")

	hint := constraint.Hint()
	assert.Contains(t, hint, "integer")
}

func TestCompositeConstraint_Description(t *testing.T) {
	c1 := NewLengthConstraint(5, 10, LengthUnitWords)
	c2 := NewFormatConstraint(FormatJSON)

	allComposite := NewCompositeConstraint(CompositeModeAll, c1, c2)
	assert.Contains(t, allComposite.Description(), "AND")

	anyComposite := NewCompositeConstraint(CompositeModeAny, c1, c2)
	assert.Contains(t, anyComposite.Description(), "OR")
}
