package toon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNativeDecoder(t *testing.T) {
	t.Run("with nil options", func(t *testing.T) {
		dec := NewNativeDecoder(nil)
		assert.NotNil(t, dec)
		assert.NotNil(t, dec.Options)
	})

	t.Run("with custom options", func(t *testing.T) {
		opts := &NativeEncoderOptions{
			FieldDelimiter:    ",",
			ArrayDelimiter:    "|",
			KeyValueDelimiter: "=",
		}
		dec := NewNativeDecoder(opts)
		assert.Equal(t, opts, dec.Options)
	})
}

func TestNativeDecoder_Decode_Primitives(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("null", func(t *testing.T) {
		value, err := dec.Decode("_")
		require.NoError(t, err)
		_, ok := value.(*TOONNull)
		assert.True(t, ok)
	})

	t.Run("empty string", func(t *testing.T) {
		value, err := dec.Decode("")
		require.NoError(t, err)
		_, ok := value.(*TOONNull)
		assert.True(t, ok)
	})

	t.Run("string with type indicator", func(t *testing.T) {
		value, err := dec.Decode("s=hello")
		require.NoError(t, err)
		str, ok := value.(*TOONString)
		assert.True(t, ok)
		assert.Equal(t, "hello", str.Value)
	})

	t.Run("number with type indicator", func(t *testing.T) {
		value, err := dec.Decode("n=42")
		require.NoError(t, err)
		num, ok := value.(*TOONNumber)
		assert.True(t, ok)
		assert.Equal(t, int64(42), num.IntValue)
	})

	t.Run("float with type indicator", func(t *testing.T) {
		value, err := dec.Decode("n=3.14")
		require.NoError(t, err)
		num, ok := value.(*TOONNumber)
		assert.True(t, ok)
		assert.True(t, num.IsFloat)
		assert.InDelta(t, 3.14, num.FloatValue, 0.001)
	})

	t.Run("bool true with type indicator", func(t *testing.T) {
		value, err := dec.Decode("b=1")
		require.NoError(t, err)
		b, ok := value.(*TOONBool)
		assert.True(t, ok)
		assert.True(t, b.Value)
	})

	t.Run("bool false with type indicator", func(t *testing.T) {
		value, err := dec.Decode("b=0")
		require.NoError(t, err)
		b, ok := value.(*TOONBool)
		assert.True(t, ok)
		assert.False(t, b.Value)
	})
}

func TestNativeDecoder_Decode_InferredTypes(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("inferred true", func(t *testing.T) {
		value, err := dec.Decode("true")
		require.NoError(t, err)
		b, ok := value.(*TOONBool)
		assert.True(t, ok)
		assert.True(t, b.Value)
	})

	t.Run("inferred false", func(t *testing.T) {
		value, err := dec.Decode("false")
		require.NoError(t, err)
		b, ok := value.(*TOONBool)
		assert.True(t, ok)
		assert.False(t, b.Value)
	})

	t.Run("inferred number", func(t *testing.T) {
		value, err := dec.Decode("123")
		require.NoError(t, err)
		num, ok := value.(*TOONNumber)
		assert.True(t, ok)
		assert.Equal(t, int64(123), num.IntValue)
	})

	t.Run("inferred string", func(t *testing.T) {
		value, err := dec.Decode("hello")
		require.NoError(t, err)
		str, ok := value.(*TOONString)
		assert.True(t, ok)
		assert.Equal(t, "hello", str.Value)
	})
}

func TestNativeDecoder_Decode_Array(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("empty array", func(t *testing.T) {
		value, err := dec.Decode("[]")
		require.NoError(t, err)
		arr, ok := value.(*TOONArray)
		assert.True(t, ok)
		assert.Empty(t, arr.Elements)
	})

	t.Run("array of strings", func(t *testing.T) {
		value, err := dec.Decode("[s=a;s=b;s=c]")
		require.NoError(t, err)
		arr, ok := value.(*TOONArray)
		assert.True(t, ok)
		assert.Len(t, arr.Elements, 3)
		assert.Equal(t, "a", arr.Elements[0].(*TOONString).Value)
		assert.Equal(t, "b", arr.Elements[1].(*TOONString).Value)
		assert.Equal(t, "c", arr.Elements[2].(*TOONString).Value)
	})

	t.Run("array of numbers", func(t *testing.T) {
		value, err := dec.Decode("[n=1;n=2;n=3]")
		require.NoError(t, err)
		arr, ok := value.(*TOONArray)
		assert.True(t, ok)
		assert.Len(t, arr.Elements, 3)
		assert.Equal(t, int64(1), arr.Elements[0].(*TOONNumber).IntValue)
		assert.Equal(t, int64(2), arr.Elements[1].(*TOONNumber).IntValue)
		assert.Equal(t, int64(3), arr.Elements[2].(*TOONNumber).IntValue)
	})

	t.Run("array of mixed types", func(t *testing.T) {
		value, err := dec.Decode("[s=hello;n=42;b=1;_]")
		require.NoError(t, err)
		arr, ok := value.(*TOONArray)
		assert.True(t, ok)
		assert.Len(t, arr.Elements, 4)
		assert.Equal(t, "hello", arr.Elements[0].(*TOONString).Value)
		assert.Equal(t, int64(42), arr.Elements[1].(*TOONNumber).IntValue)
		assert.True(t, arr.Elements[2].(*TOONBool).Value)
		_, isNull := arr.Elements[3].(*TOONNull)
		assert.True(t, isNull)
	})
}

func TestNativeDecoder_Decode_Object(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("empty object", func(t *testing.T) {
		value, err := dec.Decode("{}")
		require.NoError(t, err)
		obj, ok := value.(*TOONObject)
		assert.True(t, ok)
		assert.Empty(t, obj.Fields)
	})

	t.Run("simple object", func(t *testing.T) {
		value, err := dec.Decode("name:s=John")
		require.NoError(t, err)
		obj, ok := value.(*TOONObject)
		assert.True(t, ok)
		val, exists := obj.Get("name")
		assert.True(t, exists)
		assert.Equal(t, "John", val.(*TOONString).Value)
	})

	t.Run("object with multiple fields", func(t *testing.T) {
		value, err := dec.Decode("name:s=John|age:n=30")
		require.NoError(t, err)
		obj, ok := value.(*TOONObject)
		assert.True(t, ok)

		name, exists := obj.Get("name")
		assert.True(t, exists)
		assert.Equal(t, "John", name.(*TOONString).Value)

		age, exists := obj.Get("age")
		assert.True(t, exists)
		assert.Equal(t, int64(30), age.(*TOONNumber).IntValue)
	})

	t.Run("object with abbreviated keys", func(t *testing.T) {
		value, err := dec.Decode("i:s=123|n:s=test")
		require.NoError(t, err)
		obj, ok := value.(*TOONObject)
		assert.True(t, ok)

		// Abbreviated keys should be expanded
		id, exists := obj.Get("id")
		assert.True(t, exists)
		assert.Equal(t, "123", id.(*TOONString).Value)

		name, exists := obj.Get("name")
		assert.True(t, exists)
		assert.Equal(t, "test", name.(*TOONString).Value)
	})
}

func TestNativeDecoder_Decode_NestedStructures(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("object with array", func(t *testing.T) {
		value, err := dec.Decode("name:s=John|scores:[n=1;n=2;n=3]")
		require.NoError(t, err)
		obj, ok := value.(*TOONObject)
		assert.True(t, ok)

		scores, exists := obj.Get("scores")
		assert.True(t, exists)
		arr, ok := scores.(*TOONArray)
		assert.True(t, ok)
		assert.Len(t, arr.Elements, 3)
	})

	t.Run("nested array", func(t *testing.T) {
		value, err := dec.Decode("[[n=1;n=2];[n=3;n=4]]")
		require.NoError(t, err)
		arr, ok := value.(*TOONArray)
		assert.True(t, ok)
		assert.Len(t, arr.Elements, 2)

		inner1, ok := arr.Elements[0].(*TOONArray)
		assert.True(t, ok)
		assert.Len(t, inner1.Elements, 2)
	})
}

func TestNativeDecoder_Decode_EscapedStrings(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("escaped pipe", func(t *testing.T) {
		value, err := dec.Decode("s=hello\\|world")
		require.NoError(t, err)
		str, ok := value.(*TOONString)
		assert.True(t, ok)
		assert.Equal(t, "hello|world", str.Value)
	})

	t.Run("escaped semicolon", func(t *testing.T) {
		value, err := dec.Decode("s=hello\\;world")
		require.NoError(t, err)
		str, ok := value.(*TOONString)
		assert.True(t, ok)
		assert.Equal(t, "hello;world", str.Value)
	})

	t.Run("escaped colon", func(t *testing.T) {
		value, err := dec.Decode("s=hello\\:world")
		require.NoError(t, err)
		str, ok := value.(*TOONString)
		assert.True(t, ok)
		assert.Equal(t, "hello:world", str.Value)
	})
}

func TestNativeDecoder_DecodeToGo(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("object to map", func(t *testing.T) {
		value, err := dec.DecodeToGo("name:s=John|age:n=30")
		require.NoError(t, err)
		m, ok := value.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "John", m["name"])
		assert.Equal(t, int64(30), m["age"])
	})

	t.Run("array to slice", func(t *testing.T) {
		value, err := dec.DecodeToGo("[n=1;n=2;n=3]")
		require.NoError(t, err)
		arr, ok := value.([]interface{})
		assert.True(t, ok)
		assert.Len(t, arr, 3)
	})
}

func TestNativeDecoder_DecodeToJSON(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("object to JSON", func(t *testing.T) {
		jsonBytes, err := dec.DecodeToJSON("name:s=John|age:n=30")
		require.NoError(t, err)
		jsonStr := string(jsonBytes)
		assert.Contains(t, jsonStr, `"name":"John"`)
		assert.Contains(t, jsonStr, `"age":30`)
	})

	t.Run("array to JSON", func(t *testing.T) {
		jsonBytes, err := dec.DecodeToJSON("[n=1;n=2;n=3]")
		require.NoError(t, err)
		jsonStr := string(jsonBytes)
		assert.Equal(t, "[1,2,3]", jsonStr)
	})
}

func TestNativeDecoder_DecodeToJSONString(t *testing.T) {
	dec := NewNativeDecoder(nil)

	jsonStr, err := dec.DecodeToJSONString("name:s=John")
	require.NoError(t, err)
	assert.Contains(t, jsonStr, `"name":"John"`)
}

func TestNativeDecode_ConvenienceFunction(t *testing.T) {
	value, err := NativeDecode("name:s=John|age:n=30")
	require.NoError(t, err)
	m, ok := value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "John", m["name"])
}

func TestNativeDecodeWithOptions_ConvenienceFunction(t *testing.T) {
	opts := &NativeEncoderOptions{
		FieldDelimiter:    "|",
		ArrayDelimiter:    ";",
		KeyValueDelimiter: ":",
	}
	value, err := NativeDecodeWithOptions("name:s=John", opts)
	require.NoError(t, err)
	m, ok := value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "John", m["name"])
}

func TestNativeDecodeToJSON_ConvenienceFunction(t *testing.T) {
	jsonStr, err := NativeDecodeToJSON("name:s=John")
	require.NoError(t, err)
	assert.Contains(t, jsonStr, `"name":"John"`)
}

func TestNativeDecoder_RoundTrip(t *testing.T) {
	enc := NewNativeEncoder(nil)
	dec := NewNativeDecoder(nil)

	original := map[string]interface{}{
		"name":   "John",
		"age":    int64(30),
		"active": true,
		"scores": []interface{}{int64(85), int64(90), int64(95)},
	}

	// Encode
	toon, err := enc.Encode(original)
	require.NoError(t, err)

	// Decode
	decoded, err := dec.DecodeToGo(toon)
	require.NoError(t, err)

	m, ok := decoded.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "John", m["name"])
	assert.Equal(t, int64(30), m["age"])
	assert.Equal(t, true, m["active"])

	scores, ok := m["scores"].([]interface{})
	require.True(t, ok)
	assert.Len(t, scores, 3)
}

func TestNativeDecoder_ExpandKey(t *testing.T) {
	dec := NewNativeDecoder(nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"i", "id"},
		{"n", "name"},
		{"t", "type"},
		{"v", "value"},
		{"st", "status"},
		{"sc", "score"},
		{"ca", "created_at"},
		{"ua", "updated_at"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := dec.expandKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNativeDecoder_SplitFields(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("simple fields", func(t *testing.T) {
		fields := dec.splitFields("a:1|b:2|c:3")
		assert.Len(t, fields, 3)
		assert.Equal(t, "a:1", fields[0])
		assert.Equal(t, "b:2", fields[1])
		assert.Equal(t, "c:3", fields[2])
	})

	t.Run("with escaped delimiter", func(t *testing.T) {
		fields := dec.splitFields("a:1\\|2|b:3")
		assert.Len(t, fields, 2)
		assert.Equal(t, "a:1\\|2", fields[0])
		assert.Equal(t, "b:3", fields[1])
	})

	t.Run("with nested array", func(t *testing.T) {
		fields := dec.splitFields("a:[1;2;3]|b:4")
		assert.Len(t, fields, 2)
		assert.Equal(t, "a:[1;2;3]", fields[0])
		assert.Equal(t, "b:4", fields[1])
	})
}

func TestNativeDecoder_SplitArrayElements(t *testing.T) {
	dec := NewNativeDecoder(nil)

	t.Run("simple elements", func(t *testing.T) {
		elements := dec.splitArrayElements("1;2;3")
		assert.Len(t, elements, 3)
		assert.Equal(t, "1", elements[0])
		assert.Equal(t, "2", elements[1])
		assert.Equal(t, "3", elements[2])
	})

	t.Run("with escaped delimiter", func(t *testing.T) {
		elements := dec.splitArrayElements("1\\;2;3")
		assert.Len(t, elements, 2)
		assert.Equal(t, "1\\;2", elements[0])
		assert.Equal(t, "3", elements[1])
	})

	t.Run("with nested array", func(t *testing.T) {
		elements := dec.splitArrayElements("[1;2];[3;4]")
		assert.Len(t, elements, 2)
		assert.Equal(t, "[1;2]", elements[0])
		assert.Equal(t, "[3;4]", elements[1])
	})
}

func TestNativeDecoder_LooksLikeObject(t *testing.T) {
	dec := NewNativeDecoder(nil)

	tests := []struct {
		input    string
		expected bool
	}{
		{"name:value", true},
		{"a:1|b:2", true},
		{"hello", false},
		{"123", false},
		{"[1;2;3]", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := dec.looksLikeObject(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNativeDecoder_BoolVariants(t *testing.T) {
	dec := NewNativeDecoder(nil)

	trueVariants := []string{"b=1", "b=true", "b=yes"}
	for _, v := range trueVariants {
		t.Run("true: "+v, func(t *testing.T) {
			val, err := dec.Decode(v)
			require.NoError(t, err)
			b, ok := val.(*TOONBool)
			assert.True(t, ok)
			assert.True(t, b.Value)
		})
	}

	falseVariants := []string{"b=0", "b=false", "b=no"}
	for _, v := range falseVariants {
		t.Run("false: "+v, func(t *testing.T) {
			val, err := dec.Decode(v)
			require.NoError(t, err)
			b, ok := val.(*TOONBool)
			assert.True(t, ok)
			assert.False(t, b.Value)
		})
	}
}

func TestNativeDecoder_InvalidField(t *testing.T) {
	dec := NewNativeDecoder(nil)

	// A field without proper key:value format should error
	// However, if it's treated as a simple value, it might not error
	// Let's test a malformed object scenario
	_, err := dec.decodeObject("invalid_without_colon")
	assert.Error(t, err)
}
