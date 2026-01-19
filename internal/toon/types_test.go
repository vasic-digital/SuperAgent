package toon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTOONObject_NewTOONObject(t *testing.T) {
	obj := NewTOONObject()
	assert.NotNil(t, obj)
	assert.Empty(t, obj.Fields)
	assert.NotNil(t, obj.fieldIndex)
}

func TestTOONObject_Set_And_Get(t *testing.T) {
	obj := NewTOONObject()

	obj.Set("name", NewTOONString("John"))
	obj.Set("age", NewTOONInt(30))

	val, ok := obj.Get("name")
	assert.True(t, ok)
	assert.Equal(t, "John", val.(*TOONString).Value)

	val, ok = obj.Get("age")
	assert.True(t, ok)
	assert.Equal(t, int64(30), val.(*TOONNumber).IntValue)

	_, ok = obj.Get("nonexistent")
	assert.False(t, ok)
}

func TestTOONObject_Set_Update(t *testing.T) {
	obj := NewTOONObject()

	obj.Set("name", NewTOONString("John"))
	obj.Set("name", NewTOONString("Jane"))

	val, ok := obj.Get("name")
	assert.True(t, ok)
	assert.Equal(t, "Jane", val.(*TOONString).Value)
	assert.Len(t, obj.Fields, 1)
}

func TestTOONObject_ToTOON(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *TOONObject
		opts     *NativeEncoderOptions
		expected string
	}{
		{
			name: "empty object",
			setup: func() *TOONObject {
				return NewTOONObject()
			},
			opts:     nil,
			expected: "{}",
		},
		{
			name: "simple object",
			setup: func() *TOONObject {
				obj := NewTOONObject()
				obj.Set("name", NewTOONString("John"))
				return obj
			},
			opts:     nil,
			expected: "name:John", // compression level 1 omits s= for simple strings
		},
		{
			name: "object with multiple fields",
			setup: func() *TOONObject {
				obj := NewTOONObject()
				obj.Set("name", NewTOONString("John"))
				obj.Set("age", NewTOONInt(30))
				return obj
			},
			opts:     nil,
			expected: "name:John|age:n=30", // compression level 1 omits s= for simple strings
		},
		{
			name: "object with abbreviated keys",
			setup: func() *TOONObject {
				obj := NewTOONObject()
				obj.Set("id", NewTOONString("123"))
				obj.Set("name", NewTOONString("test"))
				return obj
			},
			opts: &NativeEncoderOptions{
				FieldDelimiter:    "|",
				ArrayDelimiter:    ";",
				KeyValueDelimiter: ":",
				AbbreviateKeys:    true,
			},
			expected: "i:s=123|n:s=test", // "123" looks like a number, so needs s=
		},
		{
			name: "object with null value",
			setup: func() *TOONObject {
				obj := NewTOONObject()
				obj.Set("name", NewTOONString("John"))
				obj.Set("middle", NewTOONNull())
				return obj
			},
			opts:     nil,
			expected: "name:John|middle:_", // compression level 1 omits s= for simple strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := tt.setup()
			result := obj.ToTOON(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTOONObject_GoValue(t *testing.T) {
	obj := NewTOONObject()
	obj.Set("name", NewTOONString("John"))
	obj.Set("age", NewTOONInt(30))
	obj.Set("active", NewTOONBool(true))

	goVal := obj.GoValue().(map[string]interface{})
	assert.Equal(t, "John", goVal["name"])
	assert.Equal(t, int64(30), goVal["age"])
	assert.Equal(t, true, goVal["active"])
}

func TestTOONObject_Type(t *testing.T) {
	obj := NewTOONObject()
	assert.Equal(t, "o", obj.Type())
}

func TestTOONObject_String(t *testing.T) {
	obj := NewTOONObject()
	obj.Set("a", NewTOONString("1"))
	obj.Set("b", NewTOONString("2"))
	assert.Equal(t, "TOONObject{2 fields}", obj.String())
}

func TestTOONArray_NewTOONArray(t *testing.T) {
	arr := NewTOONArray()
	assert.NotNil(t, arr)
	assert.Empty(t, arr.Elements)
}

func TestTOONArray_NewTOONArray_WithElements(t *testing.T) {
	arr := NewTOONArray(
		NewTOONString("a"),
		NewTOONInt(1),
		NewTOONBool(true),
	)
	assert.Len(t, arr.Elements, 3)
}

func TestTOONArray_Append(t *testing.T) {
	arr := NewTOONArray()
	arr.Append(NewTOONString("first"))
	arr.Append(NewTOONInt(2))

	assert.Len(t, arr.Elements, 2)
	assert.Equal(t, "first", arr.Elements[0].(*TOONString).Value)
	assert.Equal(t, int64(2), arr.Elements[1].(*TOONNumber).IntValue)
}

func TestTOONArray_ToTOON(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *TOONArray
		opts     *NativeEncoderOptions
		expected string
	}{
		{
			name: "empty array",
			setup: func() *TOONArray {
				return NewTOONArray()
			},
			opts:     nil,
			expected: "[]",
		},
		{
			name: "array with strings",
			setup: func() *TOONArray {
				return NewTOONArray(
					NewTOONString("a"),
					NewTOONString("b"),
					NewTOONString("c"),
				)
			},
			opts:     nil,
			expected: "[a;b;c]", // compression level 1 omits s= for simple strings
		},
		{
			name: "array with numbers",
			setup: func() *TOONArray {
				return NewTOONArray(
					NewTOONInt(1),
					NewTOONInt(2),
					NewTOONInt(3),
				)
			},
			opts:     nil,
			expected: "[n=1;n=2;n=3]",
		},
		{
			name: "array with mixed types",
			setup: func() *TOONArray {
				return NewTOONArray(
					NewTOONString("test"),
					NewTOONInt(42),
					NewTOONBool(true),
					NewTOONNull(),
				)
			},
			opts:     nil,
			expected: "[test;n=42;b=1;_]", // compression level 1 omits s= for simple strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := tt.setup()
			result := arr.ToTOON(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTOONArray_GoValue(t *testing.T) {
	arr := NewTOONArray(
		NewTOONString("test"),
		NewTOONInt(42),
		NewTOONBool(true),
	)

	goVal := arr.GoValue().([]interface{})
	assert.Len(t, goVal, 3)
	assert.Equal(t, "test", goVal[0])
	assert.Equal(t, int64(42), goVal[1])
	assert.Equal(t, true, goVal[2])
}

func TestTOONArray_Type(t *testing.T) {
	arr := NewTOONArray()
	assert.Equal(t, "a", arr.Type())
}

func TestTOONArray_String(t *testing.T) {
	arr := NewTOONArray(NewTOONString("a"), NewTOONString("b"))
	assert.Equal(t, "TOONArray{2 elements}", arr.String())
}

func TestTOONString_NewTOONString(t *testing.T) {
	s := NewTOONString("hello")
	assert.Equal(t, "hello", s.Value)
}

func TestTOONString_ToTOON(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		opts     *NativeEncoderOptions
		expected string
	}{
		{
			name:     "simple string",
			value:    "hello",
			opts:     nil,
			expected: "hello", // compression level 1 omits s= for simple strings that don't need it
		},
		{
			name:  "string without type indicator",
			value: "hello",
			opts: &NativeEncoderOptions{
				OmitTypeIndicators: true,
			},
			expected: "hello",
		},
		{
			name:  "string with high compression",
			value: "hello",
			opts: &NativeEncoderOptions{
				CompressionLevel: 2,
			},
			expected: "hello",
		},
		{
			name:     "string that looks like number",
			value:    "123",
			opts:     DefaultNativeEncoderOptions(),
			expected: "s=123",
		},
		{
			name:     "string with special chars",
			value:    "hello|world",
			opts:     DefaultNativeEncoderOptions(),
			expected: "s=hello\\|world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTOONString(tt.value)
			result := s.ToTOON(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTOONString_GoValue(t *testing.T) {
	s := NewTOONString("hello")
	assert.Equal(t, "hello", s.GoValue())
}

func TestTOONString_Type(t *testing.T) {
	s := NewTOONString("hello")
	assert.Equal(t, "s", s.Type())
}

func TestTOONString_String(t *testing.T) {
	s := NewTOONString("hello")
	assert.Equal(t, `TOONString{"hello"}`, s.String())
}

func TestTOONNumber_NewTOONInt(t *testing.T) {
	n := NewTOONInt(42)
	assert.Equal(t, int64(42), n.IntValue)
	assert.False(t, n.IsFloat)
}

func TestTOONNumber_NewTOONFloat(t *testing.T) {
	n := NewTOONFloat(3.14)
	assert.Equal(t, 3.14, n.FloatValue)
	assert.True(t, n.IsFloat)
}

func TestTOONNumber_ToTOON(t *testing.T) {
	tests := []struct {
		name     string
		number   *TOONNumber
		opts     *NativeEncoderOptions
		expected string
	}{
		{
			name:     "integer",
			number:   NewTOONInt(42),
			opts:     nil,
			expected: "n=42",
		},
		{
			name:     "negative integer",
			number:   NewTOONInt(-42),
			opts:     nil,
			expected: "n=-42",
		},
		{
			name:     "float",
			number:   NewTOONFloat(3.14),
			opts:     nil,
			expected: "n=3.14",
		},
		{
			name:   "integer without type indicator",
			number: NewTOONInt(42),
			opts: &NativeEncoderOptions{
				OmitTypeIndicators: true,
			},
			expected: "42",
		},
		{
			name:   "integer with high compression",
			number: NewTOONInt(42),
			opts: &NativeEncoderOptions{
				CompressionLevel: 2,
			},
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.number.ToTOON(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTOONNumber_GoValue(t *testing.T) {
	intNum := NewTOONInt(42)
	assert.Equal(t, int64(42), intNum.GoValue())

	floatNum := NewTOONFloat(3.14)
	assert.Equal(t, 3.14, floatNum.GoValue())
}

func TestTOONNumber_Type(t *testing.T) {
	n := NewTOONInt(42)
	assert.Equal(t, "n", n.Type())
}

func TestTOONNumber_String(t *testing.T) {
	intNum := NewTOONInt(42)
	assert.Equal(t, "TOONNumber{42}", intNum.String())

	floatNum := NewTOONFloat(3.14)
	assert.Contains(t, floatNum.String(), "3.14")
}

func TestTOONBool_NewTOONBool(t *testing.T) {
	trueVal := NewTOONBool(true)
	assert.True(t, trueVal.Value)

	falseVal := NewTOONBool(false)
	assert.False(t, falseVal.Value)
}

func TestTOONBool_ToTOON(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		opts     *NativeEncoderOptions
		expected string
	}{
		{
			name:     "true",
			value:    true,
			opts:     nil,
			expected: "b=1",
		},
		{
			name:     "false",
			value:    false,
			opts:     nil,
			expected: "b=0",
		},
		{
			name:  "true without type indicator",
			value: true,
			opts: &NativeEncoderOptions{
				OmitTypeIndicators: true,
			},
			expected: "1",
		},
		{
			name:  "true with high compression",
			value: true,
			opts: &NativeEncoderOptions{
				CompressionLevel: 2,
			},
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewTOONBool(tt.value)
			result := b.ToTOON(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTOONBool_GoValue(t *testing.T) {
	trueVal := NewTOONBool(true)
	assert.Equal(t, true, trueVal.GoValue())

	falseVal := NewTOONBool(false)
	assert.Equal(t, false, falseVal.GoValue())
}

func TestTOONBool_Type(t *testing.T) {
	b := NewTOONBool(true)
	assert.Equal(t, "b", b.Type())
}

func TestTOONBool_String(t *testing.T) {
	b := NewTOONBool(true)
	assert.Equal(t, "TOONBool{true}", b.String())
}

func TestTOONNull_NewTOONNull(t *testing.T) {
	n := NewTOONNull()
	assert.NotNil(t, n)
}

func TestTOONNull_ToTOON(t *testing.T) {
	n := NewTOONNull()
	assert.Equal(t, "_", n.ToTOON(nil))
}

func TestTOONNull_GoValue(t *testing.T) {
	n := NewTOONNull()
	assert.Nil(t, n.GoValue())
}

func TestTOONNull_Type(t *testing.T) {
	n := NewTOONNull()
	assert.Equal(t, "_", n.Type())
}

func TestTOONNull_String(t *testing.T) {
	n := NewTOONNull()
	assert.Equal(t, "TOONNull{}", n.String())
}

func TestDefaultNativeEncoderOptions(t *testing.T) {
	opts := DefaultNativeEncoderOptions()

	assert.Equal(t, "|", opts.FieldDelimiter)
	assert.Equal(t, ";", opts.ArrayDelimiter)
	assert.Equal(t, ":", opts.KeyValueDelimiter)
	assert.Equal(t, 1, opts.CompressionLevel)
	assert.False(t, opts.AbbreviateKeys)
	assert.False(t, opts.OmitTypeIndicators)
	assert.False(t, opts.OmitNullValues)
}

func TestHighCompressionNativeOptions(t *testing.T) {
	opts := HighCompressionNativeOptions()

	assert.Equal(t, "|", opts.FieldDelimiter)
	assert.Equal(t, ";", opts.ArrayDelimiter)
	assert.Equal(t, ":", opts.KeyValueDelimiter)
	assert.Equal(t, 3, opts.CompressionLevel)
	assert.True(t, opts.AbbreviateKeys)
	assert.True(t, opts.OmitTypeIndicators)
	assert.True(t, opts.OmitNullValues)
}

func TestAbbreviateKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"id", "i"},
		{"name", "n"},
		{"type", "t"},
		{"value", "v"},
		{"status", "st"},
		{"score", "sc"},
		{"created_at", "ca"},
		{"updated_at", "ua"},
		{"description", "d"},
		{"unknown", "unk"},       // unknown is > 5 chars so gets abbreviated to first 3
		{"very_long_key_name", "ver"},
		{"short", "short"},       // short is exactly 5 chars so stays unchanged
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := abbreviateKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeString(t *testing.T) {
	opts := DefaultNativeEncoderOptions()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello|world", "hello\\|world"},
		{"hello;world", "hello\\;world"},
		{"hello:world", "hello\\:world"},
		{"hello[world]", "hello\\[world\\]"},
		{"hello(world)", "hello\\(world\\)"},
		{"hello\\world", "hello\\\\world"},
		{`hello"world`, `hello\"world`},
		{"a|b;c:d", "a\\|b\\;c\\:d"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeString(tt.input, opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnescapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello\\|world", "hello|world"},
		{"hello\\;world", "hello;world"},
		{"hello\\:world", "hello:world"},
		{"hello\\[world\\]", "hello[world]"},
		{"hello\\(world\\)", "hello(world)"},
		{"hello\\\\world", "hello\\world"},
		{`hello\"world`, `hello"world`},
		{"a\\|b\\;c\\:d", "a|b;c:d"},
		{"trailing\\", "trailing\\"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := unescapeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNeedsTypeIndicator(t *testing.T) {
	opts := DefaultNativeEncoderOptions()

	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", false},
		{"1", true},
		{"0", true},
		{"_", true},
		{"123", true},
		{"123.45", true},
		{"hello|world", true},
		{"hello;world", true},
		{"hello:world", true},
		{"hello[world]", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := needsTypeIndicator(tt.input, opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNestedStructures(t *testing.T) {
	// Create a nested object with array
	obj := NewTOONObject()
	obj.Set("name", NewTOONString("John"))

	arr := NewTOONArray(
		NewTOONInt(1),
		NewTOONInt(2),
		NewTOONInt(3),
	)
	obj.Set("scores", arr)

	nested := NewTOONObject()
	nested.Set("city", NewTOONString("NYC"))
	nested.Set("zip", NewTOONInt(10001))
	obj.Set("address", nested)

	result := obj.ToTOON(nil)
	// With default compression (level 1), string type indicators are omitted for simple strings
	assert.Contains(t, result, "name:John")
	assert.Contains(t, result, "scores:[n=1;n=2;n=3]")
	assert.Contains(t, result, "address:city:NYC")
	assert.Contains(t, result, "zip:n=10001")
}
