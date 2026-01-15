package toon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEncoder(t *testing.T) {
	enc := NewEncoder(nil)
	assert.NotNil(t, enc)
	assert.Equal(t, CompressionStandard, enc.compression)
	assert.NotNil(t, enc.keyMapping)
	assert.NotNil(t, enc.reverseKeyMap)
}

func TestEncoder_WithCustomOptions(t *testing.T) {
	opts := &EncoderOptions{
		Compression: CompressionAggressive,
		KeyMapping: map[string]string{
			"custom_field": "cf",
		},
	}
	enc := NewEncoder(opts)
	assert.Equal(t, CompressionAggressive, enc.compression)
	assert.Equal(t, "cf", enc.keyMapping["custom_field"])
}

func TestEncoder_Encode_Simple(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionNone})

	data := map[string]interface{}{
		"id":   "123",
		"name": "test",
	}

	encoded, err := enc.Encode(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Should be valid JSON
	var decoded map[string]interface{}
	err = json.Unmarshal(encoded, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "123", decoded["id"])
	assert.Equal(t, "test", decoded["name"])
}

func TestEncoder_Encode_WithKeyCompression(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionMinimal})

	data := map[string]interface{}{
		"id":         "123",
		"name":       "test",
		"created_at": "2024-01-01",
	}

	encoded, err := enc.Encode(data)
	assert.NoError(t, err)

	// Check that keys are compressed
	var decoded map[string]interface{}
	err = json.Unmarshal(encoded, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "123", decoded["i"])       // id -> i
	assert.Equal(t, "test", decoded["n"])      // name -> n
	assert.Equal(t, "2024-01-01", decoded["ca"]) // created_at -> ca
}

func TestEncoder_Encode_WithValueCompression(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionStandard})

	data := map[string]interface{}{
		"status": "healthy",
	}

	encoded, err := enc.Encode(data)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(encoded, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "H", decoded["s"]) // status -> s, healthy -> H
}

func TestEncoder_Encode_NestedObject(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionMinimal})

	data := map[string]interface{}{
		"id": "123",
		"provider": map[string]interface{}{
			"name":   "test",
			"status": "active",
		},
	}

	encoded, err := enc.Encode(data)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(encoded, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "123", decoded["i"])

	provider := decoded["provider"].(map[string]interface{})
	assert.Equal(t, "test", provider["n"])
	assert.Equal(t, "active", provider["s"])
}

func TestEncoder_Encode_Array(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionMinimal})

	data := map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": "1", "name": "first"},
			{"id": "2", "name": "second"},
		},
	}

	encoded, err := enc.Encode(data)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(encoded, &decoded)
	assert.NoError(t, err)

	items := decoded["items"].([]interface{})
	assert.Len(t, items, 2)

	item1 := items[0].(map[string]interface{})
	assert.Equal(t, "1", item1["i"])
	assert.Equal(t, "first", item1["n"])
}

func TestEncoder_EncodeToString(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionNone})

	data := map[string]string{"test": "value"}
	encoded, err := enc.EncodeToString(data)

	assert.NoError(t, err)
	assert.Contains(t, encoded, "test")
	assert.Contains(t, encoded, "value")
}

func TestEncoder_TokenCount(t *testing.T) {
	enc := NewEncoder(nil)

	data := []byte("this is a test string with some content")
	count := enc.TokenCount(data)

	// ~4 chars per token
	assert.Greater(t, count, 0)
	assert.LessOrEqual(t, count, len(data)/4+1)
}

func TestEncoder_CompressionRatio(t *testing.T) {
	enc := NewEncoder(nil)

	original := []byte("this is a long string that should be compressed")
	compressed := []byte("shorter")

	ratio := enc.CompressionRatio(original, compressed)
	assert.Less(t, ratio, 1.0)

	// Test with empty original
	ratio = enc.CompressionRatio([]byte{}, compressed)
	assert.Equal(t, float64(0), ratio)
}

func TestEncoder_AddKeyMapping(t *testing.T) {
	enc := NewEncoder(nil)

	enc.AddKeyMapping("custom_field", "cf")
	assert.Equal(t, "cf", enc.keyMapping["custom_field"])
	assert.Equal(t, "custom_field", enc.reverseKeyMap["cf"])
}

func TestEncoder_SetCompression(t *testing.T) {
	enc := NewEncoder(nil)
	enc.SetCompression(CompressionAggressive)
	assert.Equal(t, CompressionAggressive, enc.GetCompressionLevel())
}

func TestDecoder_Decode_Simple(t *testing.T) {
	dec := NewDecoder(nil)

	data := []byte(`{"i":"123","n":"test"}`)
	var result map[string]interface{}

	err := dec.Decode(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "test", result["name"])
}

func TestDecoder_Decode_WithValueExpansion(t *testing.T) {
	dec := NewDecoder(nil)

	data := []byte(`{"s":"H"}`)
	var result map[string]interface{}

	err := dec.Decode(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", result["status"])
}

func TestDecoder_Decode_NestedObject(t *testing.T) {
	dec := NewDecoder(nil)

	data := []byte(`{"i":"123","d":{"n":"nested","s":"A"}}`)
	var result map[string]interface{}

	err := dec.Decode(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "123", result["id"])

	nested := result["data"].(map[string]interface{})
	assert.Equal(t, "nested", nested["name"])
	assert.Equal(t, "active", nested["status"])
}

func TestDecoder_Decode_Array(t *testing.T) {
	dec := NewDecoder(nil)

	data := []byte(`{"items":[{"i":"1","s":"H"},{"i":"2","s":"D"}]}`)
	var result map[string]interface{}

	err := dec.Decode(data, &result)
	assert.NoError(t, err)

	items := result["items"].([]interface{})
	assert.Len(t, items, 2)

	item1 := items[0].(map[string]interface{})
	assert.Equal(t, "1", item1["id"])
	assert.Equal(t, "healthy", item1["status"])

	item2 := items[1].(map[string]interface{})
	assert.Equal(t, "2", item2["id"])
	assert.Equal(t, "degraded", item2["status"])
}

func TestDecoder_DecodeToMap(t *testing.T) {
	dec := NewDecoder(nil)

	data := []byte(`{"i":"123","n":"test"}`)
	result, err := dec.DecodeToMap(data)

	assert.NoError(t, err)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "test", result["name"])
}

func TestEncoder_Decoder_RoundTrip(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionStandard})
	dec := NewDecoder(nil)

	original := map[string]interface{}{
		"id":     "123",
		"name":   "test",
		"status": "healthy",
		"nested": map[string]interface{}{
			"score":      8.5,
			"created_at": "2024-01-01",
		},
		"items": []interface{}{
			map[string]interface{}{"id": "1", "status": "completed"},
			map[string]interface{}{"id": "2", "status": "pending"},
		},
	}

	// Encode
	encoded, err := enc.Encode(original)
	assert.NoError(t, err)

	// Decode
	var decoded map[string]interface{}
	err = dec.Decode(encoded, &decoded)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, "123", decoded["id"])
	assert.Equal(t, "test", decoded["name"])
	assert.Equal(t, "healthy", decoded["status"])

	nested := decoded["nested"].(map[string]interface{})
	assert.Equal(t, 8.5, nested["score"])

	items := decoded["items"].([]interface{})
	assert.Len(t, items, 2)
}

func TestEncoder_GzipCompression(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionAggressive})
	dec := NewDecoder(nil)

	// Create a large enough payload to benefit from gzip
	data := map[string]interface{}{
		"id":          "123",
		"name":        "test",
		"description": "This is a long description that should compress well when using gzip compression algorithm",
		"items":       make([]map[string]interface{}, 10),
	}
	for i := 0; i < 10; i++ {
		data["items"].([]map[string]interface{})[i] = map[string]interface{}{
			"id":     i,
			"name":   "item",
			"status": "pending",
		}
	}

	encoded, err := enc.Encode(data)
	assert.NoError(t, err)

	// Decode should handle gzip automatically
	var decoded map[string]interface{}
	err = dec.Decode(encoded, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, "123", decoded["id"])
}

func TestValueAbbreviations(t *testing.T) {
	enc := NewEncoder(&EncoderOptions{Compression: CompressionStandard})

	tests := []struct {
		input    string
		expected string
	}{
		{"healthy", "H"},
		{"degraded", "D"},
		{"unhealthy", "U"},
		{"pending", "P"},
		{"running", "R"},
		{"completed", "C"},
		{"failed", "F"},
		{"active", "A"},
		{"inactive", "I"},
		{"queued", "Q"},
		{"cancelled", "X"},
		{"unknown", "unknown"}, // Unchanged
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, enc.abbreviateValue(tt.input))
		})
	}
}

func TestValueExpansions(t *testing.T) {
	dec := NewDecoder(nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"H", "healthy"},
		{"D", "degraded"},
		{"U", "unhealthy"},
		{"P", "pending"},
		{"R", "running"},
		{"C", "completed"},
		{"F", "failed"},
		{"A", "active"},
		{"I", "inactive"},
		{"Q", "queued"},
		{"X", "cancelled"},
		{"unknown", "unknown"}, // Unchanged
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, dec.expandValue(tt.input))
		})
	}
}

func TestDefaultKeyMapping(t *testing.T) {
	mapping := DefaultKeyMapping()

	// Check some key mappings exist
	assert.Equal(t, "i", mapping["id"])
	assert.Equal(t, "n", mapping["name"])
	assert.Equal(t, "s", mapping["status"])
	assert.Equal(t, "ca", mapping["created_at"])
	assert.Equal(t, "sc", mapping["score"])
	assert.Equal(t, "pi", mapping["provider_id"])
}
