package toon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNativeEncoder(t *testing.T) {
	t.Run("with nil options", func(t *testing.T) {
		enc := NewNativeEncoder(nil)
		assert.NotNil(t, enc)
		assert.NotNil(t, enc.Options)
	})

	t.Run("with custom options", func(t *testing.T) {
		opts := &NativeEncoderOptions{
			FieldDelimiter:     ",",
			ArrayDelimiter:     "|",
			KeyValueDelimiter:  "=",
			CompressionLevel:   2,
			AbbreviateKeys:     true,
			OmitTypeIndicators: true,
		}
		enc := NewNativeEncoder(opts)
		assert.Equal(t, opts, enc.Options)
	})
}

func TestNativeEncoder_Encode_Primitives(t *testing.T) {
	enc := NewNativeEncoder(nil)

	t.Run("nil", func(t *testing.T) {
		result, err := enc.Encode(nil)
		require.NoError(t, err)
		assert.Equal(t, "_", result)
	})

	t.Run("bool true", func(t *testing.T) {
		result, err := enc.Encode(true)
		require.NoError(t, err)
		assert.Equal(t, "b=1", result)
	})

	t.Run("bool false", func(t *testing.T) {
		result, err := enc.Encode(false)
		require.NoError(t, err)
		assert.Equal(t, "b=0", result)
	})

	t.Run("int", func(t *testing.T) {
		result, err := enc.Encode(42)
		require.NoError(t, err)
		assert.Equal(t, "n=42", result)
	})

	t.Run("int64", func(t *testing.T) {
		result, err := enc.Encode(int64(9223372036854775807))
		require.NoError(t, err)
		assert.Equal(t, "n=9223372036854775807", result)
	})

	t.Run("float64", func(t *testing.T) {
		result, err := enc.Encode(3.14159)
		require.NoError(t, err)
		assert.Contains(t, result, "n=3.14")
	})

	t.Run("float that is integer", func(t *testing.T) {
		result, err := enc.Encode(42.0)
		require.NoError(t, err)
		assert.Equal(t, "n=42", result)
	})

	t.Run("string", func(t *testing.T) {
		result, err := enc.Encode("hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", result) // compression level 1 omits s= for simple strings
	})

	t.Run("string with special characters", func(t *testing.T) {
		result, err := enc.Encode("hello|world")
		require.NoError(t, err)
		assert.Equal(t, "s=hello\\|world", result) // needs s= because of special chars
	})
}

func TestNativeEncoder_Encode_Pointers(t *testing.T) {
	enc := NewNativeEncoder(nil)

	t.Run("nil pointer", func(t *testing.T) {
		var ptr *string
		result, err := enc.Encode(ptr)
		require.NoError(t, err)
		assert.Equal(t, "_", result)
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		str := "hello"
		result, err := enc.Encode(&str)
		require.NoError(t, err)
		assert.Equal(t, "hello", result) // compression level 1 omits s= for simple strings
	})
}

func TestNativeEncoder_Encode_Slice(t *testing.T) {
	enc := NewNativeEncoder(nil)

	t.Run("empty slice", func(t *testing.T) {
		result, err := enc.Encode([]int{})
		require.NoError(t, err)
		assert.Equal(t, "[]", result)
	})

	t.Run("slice of ints", func(t *testing.T) {
		result, err := enc.Encode([]int{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, "[n=1;n=2;n=3]", result)
	})

	t.Run("slice of strings", func(t *testing.T) {
		result, err := enc.Encode([]string{"a", "b", "c"})
		require.NoError(t, err)
		assert.Equal(t, "[a;b;c]", result) // compression level 1 omits s= for simple strings
	})

	t.Run("slice of mixed interface", func(t *testing.T) {
		result, err := enc.Encode([]interface{}{"hello", 42, true})
		require.NoError(t, err)
		assert.Equal(t, "[hello;n=42;b=1]", result) // compression level 1 omits s= for simple strings
	})

	t.Run("byte slice as string", func(t *testing.T) {
		result, err := enc.Encode([]byte("hello"))
		require.NoError(t, err)
		assert.Equal(t, "hello", result) // compression level 1 omits s= for simple strings
	})
}

func TestNativeEncoder_Encode_Map(t *testing.T) {
	enc := NewNativeEncoder(nil)

	t.Run("empty map", func(t *testing.T) {
		result, err := enc.Encode(map[string]string{})
		require.NoError(t, err)
		assert.Equal(t, "{}", result)
	})

	t.Run("simple map", func(t *testing.T) {
		// Maps don't have guaranteed order, so we just check contains
		result, err := enc.Encode(map[string]string{"name": "John"})
		require.NoError(t, err)
		assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
	})

	t.Run("map with int keys", func(t *testing.T) {
		result, err := enc.Encode(map[int]string{1: "one"})
		require.NoError(t, err)
		assert.Contains(t, result, "1:one") // compression level 1 omits s= for simple strings
	})
}

func TestNativeEncoder_Encode_Struct(t *testing.T) {
	enc := NewNativeEncoder(nil)

	t.Run("simple struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		p := Person{Name: "John", Age: 30}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "Name:John") // compression level 1 omits s= for simple strings
		assert.Contains(t, result, "Age:n=30")
	})

	t.Run("struct with json tags", func(t *testing.T) {
		type Person struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		p := Person{FirstName: "John", LastName: "Doe"}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "first_name:John") // compression level 1 omits s= for simple strings
		assert.Contains(t, result, "last_name:Doe")
	})

	t.Run("struct with omitempty", func(t *testing.T) {
		type Person struct {
			Name  string `json:"name"`
			Email string `json:"email,omitempty"`
		}
		p := Person{Name: "John"}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
		assert.NotContains(t, result, "email")
	})

	t.Run("struct with toon tags", func(t *testing.T) {
		type Person struct {
			Name string `json:"name" toon:"n"`
		}
		p := Person{Name: "John"}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "n:John") // compression level 1 omits s= for simple strings
	})

	t.Run("struct with ignored field", func(t *testing.T) {
		type Person struct {
			Name   string `json:"name"`
			Secret string `json:"-"`
		}
		p := Person{Name: "John", Secret: "hidden"}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
		assert.NotContains(t, result, "Secret")
		assert.NotContains(t, result, "hidden")
	})

	t.Run("struct with unexported field", func(t *testing.T) {
		type Person struct {
			Name     string
			internal string
		}
		p := Person{Name: "John", internal: "hidden"}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "Name:John") // compression level 1 omits s= for simple strings
		assert.NotContains(t, result, "internal")
	})
}

func TestNativeEncoder_Encode_Time(t *testing.T) {
	enc := NewNativeEncoder(nil)

	t.Run("non-zero time", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result, err := enc.Encode(tm)
		require.NoError(t, err)
		assert.Contains(t, result, "2024-01-15")
	})

	t.Run("zero time", func(t *testing.T) {
		result, err := enc.Encode(time.Time{})
		require.NoError(t, err)
		assert.Equal(t, "_", result)
	})
}

func TestNativeEncoder_Encode_NestedStructures(t *testing.T) {
	enc := NewNativeEncoder(nil)

	t.Run("nested struct", func(t *testing.T) {
		type Address struct {
			City string `json:"city"`
			Zip  int    `json:"zip"`
		}
		type Person struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}
		p := Person{
			Name:    "John",
			Address: Address{City: "NYC", Zip: 10001},
		}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
		assert.Contains(t, result, "address:")
		assert.Contains(t, result, "city:NYC") // compression level 1 omits s= for simple strings
		assert.Contains(t, result, "zip:n=10001")
	})

	t.Run("struct with slice", func(t *testing.T) {
		type Person struct {
			Name   string `json:"name"`
			Scores []int  `json:"scores"`
		}
		p := Person{
			Name:   "John",
			Scores: []int{85, 90, 95},
		}
		result, err := enc.Encode(p)
		require.NoError(t, err)
		assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
		assert.Contains(t, result, "scores:[n=85;n=90;n=95]")
	})
}

func TestNativeEncoder_EncodeWithOptions(t *testing.T) {
	t.Run("with key abbreviation", func(t *testing.T) {
		opts := &NativeEncoderOptions{
			FieldDelimiter:    "|",
			ArrayDelimiter:    ";",
			KeyValueDelimiter: ":",
			AbbreviateKeys:    true,
		}
		enc := NewNativeEncoder(opts)

		result, err := enc.Encode(map[string]string{
			"id":         "123",
			"name":       "test",
			"created_at": "2024-01-01",
		})
		require.NoError(t, err)
		assert.Contains(t, result, "i:s=123")
		assert.Contains(t, result, "n:s=test")
		assert.Contains(t, result, "ca:s=2024-01-01")
	})

	t.Run("with omit type indicators", func(t *testing.T) {
		opts := &NativeEncoderOptions{
			FieldDelimiter:     "|",
			ArrayDelimiter:     ";",
			KeyValueDelimiter:  ":",
			OmitTypeIndicators: true,
		}
		enc := NewNativeEncoder(opts)

		result, err := enc.Encode(map[string]interface{}{
			"name": "John",
			"age":  30,
		})
		require.NoError(t, err)
		// Without type indicators
		assert.Contains(t, result, "name:John")
		assert.Contains(t, result, "age:30")
	})

	t.Run("with omit null values", func(t *testing.T) {
		opts := &NativeEncoderOptions{
			FieldDelimiter:    "|",
			ArrayDelimiter:    ";",
			KeyValueDelimiter: ":",
			OmitNullValues:    true,
		}
		enc := NewNativeEncoder(opts)

		result, err := enc.Encode(map[string]interface{}{
			"name":   "John",
			"middle": nil,
		})
		require.NoError(t, err)
		assert.Contains(t, result, "name:")
		assert.NotContains(t, result, "middle")
	})
}

func TestNativeEncode_ConvenienceFunction(t *testing.T) {
	result, err := NativeEncode(map[string]string{"name": "John"})
	require.NoError(t, err)
	assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
}

func TestNativeEncodeWithOptions_ConvenienceFunction(t *testing.T) {
	opts := HighCompressionNativeOptions()
	result, err := NativeEncodeWithOptions(map[string]string{"id": "123"}, opts)
	require.NoError(t, err)
	assert.Contains(t, result, "i:123")
}

func TestNativeEncodeJSON(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		result, err := NativeEncodeJSON(`{"name":"John","age":30}`)
		require.NoError(t, err)
		assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
		assert.Contains(t, result, "age:n=30")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := NativeEncodeJSON(`invalid`)
		assert.Error(t, err)
	})
}

func TestNativeEncodeJSONBytes(t *testing.T) {
	result, err := NativeEncodeJSONBytes([]byte(`{"name":"John"}`))
	require.NoError(t, err)
	assert.Contains(t, result, "name:John") // compression level 1 omits s= for simple strings
}

func TestMarshalNativeTOON(t *testing.T) {
	data, err := MarshalNativeTOON(map[string]string{"name": "John"})
	require.NoError(t, err)
	assert.Contains(t, string(data), "name:John") // compression level 1 omits s= for simple strings
}

func TestNativeTokenEstimate(t *testing.T) {
	// 20 characters should be about 5 tokens
	tokens := NativeTokenEstimate("12345678901234567890")
	assert.GreaterOrEqual(t, tokens, 4)
	assert.LessOrEqual(t, tokens, 6)
}

func TestNativeTokenSavings(t *testing.T) {
	json := `{"name":"John","age":30,"active":true}`
	toon := "name:John|age:30|active:1"

	savings := NativeTokenSavings(json, toon)
	assert.Greater(t, savings, 0.0)
}

func TestNativeEncoder_EncodeToValue(t *testing.T) {
	enc := NewNativeEncoder(nil)

	value, err := enc.EncodeToValue(map[string]interface{}{
		"name": "John",
		"age":  30,
	})
	require.NoError(t, err)

	obj, ok := value.(*TOONObject)
	assert.True(t, ok)
	assert.Len(t, obj.Fields, 2)
}

func TestNativeEncoder_TOONValue_DirectPass(t *testing.T) {
	enc := NewNativeEncoder(nil)

	toonObj := NewTOONObject()
	toonObj.Set("test", NewTOONString("value"))

	result, err := enc.Encode(toonObj)
	require.NoError(t, err)
	assert.Contains(t, result, "test:value") // compression level 1 omits s= for simple strings
}
