package toon

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// NativeEncoder encodes Go values to native TOON format.
type NativeEncoder struct {
	Options *NativeEncoderOptions
}

// NewNativeEncoder creates a new native TOON encoder with the given options.
func NewNativeEncoder(opts *NativeEncoderOptions) *NativeEncoder {
	if opts == nil {
		opts = DefaultNativeEncoderOptions()
	}
	return &NativeEncoder{Options: opts}
}

// Encode converts a Go value to native TOON format string.
func (e *NativeEncoder) Encode(v interface{}) (string, error) {
	toonValue, err := e.valueToTOON(v)
	if err != nil {
		return "", err
	}
	return toonValue.ToTOON(e.Options), nil
}

// EncodeToValue converts a Go value to a TOONValue.
func (e *NativeEncoder) EncodeToValue(v interface{}) (TOONValue, error) {
	return e.valueToTOON(v)
}

// valueToTOON converts a Go value to a TOONValue.
func (e *NativeEncoder) valueToTOON(v interface{}) (TOONValue, error) {
	if v == nil {
		return NewTOONNull(), nil
	}

	// Handle TOONValue directly
	if tv, ok := v.(TOONValue); ok {
		return tv, nil
	}

	val := reflect.ValueOf(v)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return NewTOONNull(), nil
		}
		return e.valueToTOON(val.Elem().Interface())
	}

	switch val.Kind() {
	case reflect.Bool:
		return NewTOONBool(val.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewTOONInt(val.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Note: Large uint64 values (> int64 max) will overflow. This is acceptable for
		// typical use cases where values are within int64 range.
		return NewTOONInt(int64(val.Uint())), nil // #nosec G115 - TOON format uses int64, overflow is acceptable for edge cases

	case reflect.Float32, reflect.Float64:
		f := val.Float()
		// Check if it's actually an integer
		if f == float64(int64(f)) {
			return NewTOONInt(int64(f)), nil
		}
		return NewTOONFloat(f), nil

	case reflect.String:
		return NewTOONString(val.String()), nil

	case reflect.Slice, reflect.Array:
		return e.sliceToTOON(val)

	case reflect.Map:
		return e.mapToTOON(val)

	case reflect.Struct:
		return e.structToTOON(val)

	case reflect.Interface:
		if val.IsNil() {
			return NewTOONNull(), nil
		}
		return e.valueToTOON(val.Elem().Interface())

	default:
		return nil, fmt.Errorf("unsupported type: %v", val.Kind())
	}
}

// sliceToTOON converts a slice to a TOONArray.
func (e *NativeEncoder) sliceToTOON(val reflect.Value) (TOONValue, error) {
	// Handle []byte specially as a string
	if val.Type().Elem().Kind() == reflect.Uint8 {
		return NewTOONString(string(val.Bytes())), nil
	}

	arr := NewTOONArray()
	for i := 0; i < val.Len(); i++ {
		elem, err := e.valueToTOON(val.Index(i).Interface())
		if err != nil {
			return nil, fmt.Errorf("error encoding array element %d: %w", i, err)
		}
		arr.Append(elem)
	}
	return arr, nil
}

// mapToTOON converts a map to a TOONObject.
func (e *NativeEncoder) mapToTOON(val reflect.Value) (TOONValue, error) {
	obj := NewTOONObject()

	iter := val.MapRange()
	for iter.Next() {
		key := iter.Key()
		keyStr, ok := key.Interface().(string)
		if !ok {
			keyStr = fmt.Sprintf("%v", key.Interface())
		}

		value, err := e.valueToTOON(iter.Value().Interface())
		if err != nil {
			return nil, fmt.Errorf("error encoding map value for key %q: %w", keyStr, err)
		}

		// Skip null values if configured
		if e.Options.OmitNullValues {
			if _, isNull := value.(*TOONNull); isNull {
				continue
			}
		}

		obj.Set(keyStr, value)
	}
	return obj, nil
}

// structToTOON converts a struct to a TOONObject.
func (e *NativeEncoder) structToTOON(val reflect.Value) (TOONValue, error) {
	// Handle time.Time specially
	if t, ok := val.Interface().(time.Time); ok {
		if t.IsZero() {
			return NewTOONNull(), nil
		}
		// Use RFC3339 format for compact representation
		return NewTOONString(t.Format(time.RFC3339)), nil
	}

	obj := NewTOONObject()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag if present
		fieldName := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			parts := splitStructTag(tag)
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				fieldName = parts[0]
			}

			// Check for omitempty
			if len(parts) > 1 && parts[1] == "omitempty" && isEmptyReflectValue(val.Field(i)) {
				continue
			}
		}

		// Get TOON-specific tag if present
		if tag := field.Tag.Get("toon"); tag != "" {
			parts := splitStructTag(tag)
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		value, err := e.valueToTOON(val.Field(i).Interface())
		if err != nil {
			return nil, fmt.Errorf("error encoding field %q: %w", fieldName, err)
		}

		// Skip null values if configured
		if e.Options.OmitNullValues {
			if _, isNull := value.(*TOONNull); isNull {
				continue
			}
		}

		obj.Set(fieldName, value)
	}
	return obj, nil
}

// splitStructTag splits a struct tag value.
func splitStructTag(tag string) []string {
	result := make([]string, 0, 2)
	current := ""
	for _, c := range tag {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

// isEmptyReflectValue checks if a reflect.Value is empty.
func isEmptyReflectValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// Special handling for time.Time
		if t, ok := v.Interface().(time.Time); ok {
			return t.IsZero()
		}
		return false
	}
	return false
}

// NativeEncode is a convenience function that encodes a Go value to native TOON format.
func NativeEncode(v interface{}) (string, error) {
	return NewNativeEncoder(nil).Encode(v)
}

// NativeEncodeWithOptions encodes a Go value to native TOON format with custom options.
func NativeEncodeWithOptions(v interface{}, opts *NativeEncoderOptions) (string, error) {
	return NewNativeEncoder(opts).Encode(v)
}

// NativeEncodeJSON converts a JSON string to native TOON format.
func NativeEncodeJSON(jsonStr string) (string, error) {
	return NativeEncodeJSONWithOptions(jsonStr, nil)
}

// NativeEncodeJSONWithOptions converts a JSON string to native TOON format with custom options.
func NativeEncodeJSONWithOptions(jsonStr string, opts *NativeEncoderOptions) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	return NewNativeEncoder(opts).Encode(data)
}

// NativeEncodeJSONBytes converts JSON bytes to native TOON format.
func NativeEncodeJSONBytes(jsonBytes []byte) (string, error) {
	return NativeEncodeJSONBytesWithOptions(jsonBytes, nil)
}

// NativeEncodeJSONBytesWithOptions converts JSON bytes to native TOON format with custom options.
func NativeEncodeJSONBytesWithOptions(jsonBytes []byte, opts *NativeEncoderOptions) (string, error) {
	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	return NewNativeEncoder(opts).Encode(data)
}

// MarshalNativeTOON converts a Go value to native TOON format bytes.
func MarshalNativeTOON(v interface{}) ([]byte, error) {
	s, err := NativeEncode(v)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// MarshalNativeTOONWithOptions converts a Go value to native TOON format bytes with custom options.
func MarshalNativeTOONWithOptions(v interface{}, opts *NativeEncoderOptions) ([]byte, error) {
	s, err := NativeEncodeWithOptions(v, opts)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// NativeTokenEstimate estimates the number of LLM tokens in a TOON string.
// This uses a simple heuristic: approximately 4 characters per token.
func NativeTokenEstimate(toonStr string) int {
	// Most tokenizers use roughly 4 characters per token on average
	// This is a rough estimate and will vary by tokenizer
	return (len(toonStr) + 3) / 4
}

// NativeTokenSavings calculates the estimated token savings between JSON and TOON.
func NativeTokenSavings(jsonStr, toonStr string) float64 {
	jsonTokens := NativeTokenEstimate(jsonStr)
	toonTokens := NativeTokenEstimate(toonStr)

	if jsonTokens == 0 {
		return 0
	}

	return float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100
}
