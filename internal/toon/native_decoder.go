package toon

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// NativeDecoder decodes native TOON format strings back to Go values.
type NativeDecoder struct {
	Options *NativeEncoderOptions
}

// NewNativeDecoder creates a new native TOON decoder.
func NewNativeDecoder(opts *NativeEncoderOptions) *NativeDecoder {
	if opts == nil {
		opts = DefaultNativeEncoderOptions()
	}
	return &NativeDecoder{Options: opts}
}

// Decode decodes a TOON format string to a TOONValue.
func (d *NativeDecoder) Decode(toonStr string) (TOONValue, error) {
	if toonStr == "" {
		return NewTOONNull(), nil
	}

	toonStr = strings.TrimSpace(toonStr)

	// Handle special cases
	if toonStr == "_" {
		return NewTOONNull(), nil
	}
	if toonStr == "[]" {
		return NewTOONArray(), nil
	}
	if toonStr == "{}" {
		return NewTOONObject(), nil
	}

	// Check for array
	if strings.HasPrefix(toonStr, "[") && strings.HasSuffix(toonStr, "]") {
		return d.decodeArray(toonStr[1 : len(toonStr)-1])
	}

	// Check for type indicators FIRST before checking for object structure
	// This ensures strings like "s=hello\:world" are correctly decoded as strings
	if strings.HasPrefix(toonStr, "s=") {
		return NewTOONString(unescapeString(toonStr[2:])), nil
	}
	if strings.HasPrefix(toonStr, "n=") {
		return d.decodeNumber(toonStr[2:])
	}
	if strings.HasPrefix(toonStr, "b=") {
		return d.decodeBool(toonStr[2:])
	}

	// Check if it looks like an object (has key:value pattern)
	if strings.Contains(toonStr, d.Options.KeyValueDelimiter) {
		return d.decodeObject(toonStr)
	}

	// Try to decode as a primitive value
	return d.decodeValue(toonStr)
}

// DecodeToGo decodes a TOON format string to a Go interface{}.
func (d *NativeDecoder) DecodeToGo(toonStr string) (interface{}, error) {
	value, err := d.Decode(toonStr)
	if err != nil {
		return nil, err
	}
	return value.GoValue(), nil
}

// DecodeToJSON decodes a TOON format string to JSON.
func (d *NativeDecoder) DecodeToJSON(toonStr string) ([]byte, error) {
	value, err := d.DecodeToGo(toonStr)
	if err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

// DecodeToJSONString decodes a TOON format string to a JSON string.
func (d *NativeDecoder) DecodeToJSONString(toonStr string) (string, error) {
	jsonBytes, err := d.DecodeToJSON(toonStr)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// decodeObject decodes a TOON object string.
func (d *NativeDecoder) decodeObject(s string) (*TOONObject, error) {
	obj := NewTOONObject()

	fields := d.splitFields(s)
	for _, field := range fields {
		if field == "" {
			continue
		}

		keyValue := d.splitKeyValue(field)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid field format: %q", field)
		}

		key := keyValue[0]
		valueStr := keyValue[1]

		// Expand abbreviated key if possible
		key = d.expandKey(key)

		value, err := d.decodeValue(valueStr)
		if err != nil {
			return nil, fmt.Errorf("error decoding value for key %q: %w", key, err)
		}

		obj.Set(key, value)
	}

	return obj, nil
}

// decodeArray decodes a TOON array string (without brackets).
func (d *NativeDecoder) decodeArray(s string) (*TOONArray, error) {
	arr := NewTOONArray()

	if s == "" {
		return arr, nil
	}

	elements := d.splitArrayElements(s)
	for _, elem := range elements {
		value, err := d.decodeValue(elem)
		if err != nil {
			return nil, fmt.Errorf("error decoding array element: %w", err)
		}
		arr.Append(value)
	}

	return arr, nil
}

// decodeValue decodes a TOON value string.
func (d *NativeDecoder) decodeValue(s string) (TOONValue, error) {
	s = strings.TrimSpace(s)

	if s == "" || s == "_" {
		return NewTOONNull(), nil
	}

	// Check for type indicators
	if strings.HasPrefix(s, "s=") {
		return NewTOONString(unescapeString(s[2:])), nil
	}
	if strings.HasPrefix(s, "n=") {
		return d.decodeNumber(s[2:])
	}
	if strings.HasPrefix(s, "b=") {
		return d.decodeBool(s[2:])
	}

	// Check for nested array
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		return d.decodeArray(s[1 : len(s)-1])
	}

	// Check for nested object (has key:value pattern)
	if strings.Contains(s, d.Options.KeyValueDelimiter) && d.looksLikeObject(s) {
		return d.decodeObject(s)
	}

	// Try to infer type
	return d.inferValue(s)
}

// looksLikeObject checks if a string appears to be a TOON object.
func (d *NativeDecoder) looksLikeObject(s string) bool {
	// Simple heuristic: objects have field delimiter or start with key:value
	if strings.Contains(s, d.Options.FieldDelimiter) {
		return true
	}
	// Check if it starts with an alphanumeric key followed by :
	parts := strings.SplitN(s, d.Options.KeyValueDelimiter, 2)
	if len(parts) == 2 && len(parts[0]) > 0 {
		// Key should be alphanumeric/underscore
		for _, c := range parts[0] {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
		return true
	}
	return false
}

// inferValue infers the type of a value without a type indicator.
func (d *NativeDecoder) inferValue(s string) (TOONValue, error) {
	// Boolean check
	if s == "1" || s == "true" || s == "True" || s == "TRUE" {
		return NewTOONBool(true), nil
	}
	if s == "0" || s == "false" || s == "False" || s == "FALSE" {
		return NewTOONBool(false), nil
	}

	// Number check
	if num, err := d.decodeNumber(s); err == nil {
		return num, nil
	}

	// Default to string
	return NewTOONString(unescapeString(s)), nil
}

// decodeNumber decodes a TOON number.
func (d *NativeDecoder) decodeNumber(s string) (*TOONNumber, error) {
	s = strings.TrimSpace(s)

	// Try integer first
	if intVal, err := strconv.ParseInt(s, 10, 64); err == nil {
		return NewTOONInt(intVal), nil
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(s, 64); err == nil {
		return NewTOONFloat(floatVal), nil
	}

	return nil, errors.New("invalid number format")
}

// decodeBool decodes a TOON boolean.
func (d *NativeDecoder) decodeBool(s string) (*TOONBool, error) {
	s = strings.TrimSpace(s)

	switch strings.ToLower(s) {
	case "1", "true", "yes":
		return NewTOONBool(true), nil
	case "0", "false", "no":
		return NewTOONBool(false), nil
	}

	return nil, errors.New("invalid boolean format")
}

// splitFields splits an object string into fields, respecting escape sequences and nesting.
func (d *NativeDecoder) splitFields(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	escaped := false

	for _, c := range s {
		if escaped {
			current.WriteRune(c)
			escaped = false
			continue
		}

		if c == '\\' {
			current.WriteRune(c)
			escaped = true
			continue
		}

		if c == '[' || c == '(' {
			depth++
			current.WriteRune(c)
			continue
		}

		if c == ']' || c == ')' {
			depth--
			current.WriteRune(c)
			continue
		}

		if depth == 0 && string(c) == d.Options.FieldDelimiter {
			result = append(result, current.String())
			current.Reset()
			continue
		}

		current.WriteRune(c)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// splitKeyValue splits a field into key and value, respecting escape sequences.
func (d *NativeDecoder) splitKeyValue(s string) []string {
	var result []string
	var current strings.Builder
	escaped := false
	foundDelimiter := false

	for _, c := range s {
		if escaped {
			current.WriteRune(c)
			escaped = false
			continue
		}

		if c == '\\' {
			current.WriteRune(c)
			escaped = true
			continue
		}

		if !foundDelimiter && string(c) == d.Options.KeyValueDelimiter {
			result = append(result, current.String())
			current.Reset()
			foundDelimiter = true
			continue
		}

		current.WriteRune(c)
	}

	if current.Len() > 0 || foundDelimiter {
		result = append(result, current.String())
	}

	return result
}

// splitArrayElements splits an array string into elements.
func (d *NativeDecoder) splitArrayElements(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0
	escaped := false

	for _, c := range s {
		if escaped {
			current.WriteRune(c)
			escaped = false
			continue
		}

		if c == '\\' {
			current.WriteRune(c)
			escaped = true
			continue
		}

		if c == '[' || c == '(' {
			depth++
			current.WriteRune(c)
			continue
		}

		if c == ']' || c == ')' {
			depth--
			current.WriteRune(c)
			continue
		}

		if depth == 0 && string(c) == d.Options.ArrayDelimiter {
			result = append(result, current.String())
			current.Reset()
			continue
		}

		current.WriteRune(c)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// expandKey expands an abbreviated key to its full form.
func (d *NativeDecoder) expandKey(key string) string {
	expansions := map[string]string{
		"i":   "id",
		"n":   "name",
		"t":   "type",
		"v":   "value",
		"st":  "status",
		"sc":  "score",
		"ca":  "created_at",
		"ua":  "updated_at",
		"d":   "description",
		"c":   "content",
		"m":   "message",
		"e":   "error",
		"r":   "result",
		"da":  "data",
		"it":  "items",
		"ct":  "count",
		"to":  "total",
		"pg":  "page",
		"lm":  "limit",
		"of":  "offset",
		"pi":  "provider_id",
		"mi":  "model_id",
		"cf":  "confidence",
		"ts":  "timestamp",
		"la":  "latency",
		"tc":  "token_count",
		"mt":  "max_tokens",
		"tp":  "temperature",
		"pt":  "participant",
		"rs":  "response",
		"rq":  "request",
		"en":  "enabled",
		"di":  "disabled",
		"ac":  "active",
		"pe":  "pending",
		"cp":  "completed",
		"fl":  "failed",
		"pr":  "priority",
		"vr":  "version",
		"cx":  "context",
		"cap": "capabilities",
	}

	if expanded, exists := expansions[key]; exists {
		return expanded
	}
	return key
}

// NativeDecode is a convenience function that decodes a TOON format string to a Go value.
func NativeDecode(toonStr string) (interface{}, error) {
	return NewNativeDecoder(nil).DecodeToGo(toonStr)
}

// NativeDecodeWithOptions decodes a TOON format string with custom options.
func NativeDecodeWithOptions(toonStr string, opts *NativeEncoderOptions) (interface{}, error) {
	return NewNativeDecoder(opts).DecodeToGo(toonStr)
}

// NativeDecodeToJSON converts a TOON format string to JSON.
func NativeDecodeToJSON(toonStr string) (string, error) {
	return NewNativeDecoder(nil).DecodeToJSONString(toonStr)
}

// NativeDecodeToJSONWithOptions converts a TOON format string to JSON with custom options.
func NativeDecodeToJSONWithOptions(toonStr string, opts *NativeEncoderOptions) (string, error) {
	return NewNativeDecoder(opts).DecodeToJSONString(toonStr)
}
