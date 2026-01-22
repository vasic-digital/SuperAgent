// Package toon provides Token-Optimized Object Notation (TOON) encoding and decoding.
// TOON is a compact data format designed to minimize token usage when transmitting
// structured data to Large Language Models (LLMs).
//
// This package provides two encoding approaches:
//
//  1. JSON-based compression (Encoder/Decoder): Uses JSON with key/value abbreviation
//     and optional gzip compression. Best for compatibility with existing JSON parsers.
//
//  2. Native TOON format (NativeEncoder/NativeDecoder): Uses a custom text format with
//     minimal delimiters for maximum token savings.
//
// Native TOON Format Specification:
//   - Objects use | as field delimiter: key1:value1|key2:value2
//   - Arrays use ; as element delimiter: [value1;value2;value3]
//   - Nested structures are supported with parentheses for grouping
//   - Type indicators: s= (string), n= (number), b= (bool), _ (null)
//   - Keys are abbreviated when possible to save tokens
//
// Example:
//
//	JSON:  {"name": "John", "age": 30, "active": true}
//	TOON:  name:s=John|age:n=30|active:b=1
package toon

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// TOONValue is the interface implemented by all TOON value types.
// These types are used for the native TOON format encoding/decoding.
type TOONValue interface {
	// ToTOON returns the TOON-formatted string representation.
	ToTOON(opts *NativeEncoderOptions) string
	// Type returns the type indicator for this value.
	Type() string
	// GoValue returns the underlying Go value.
	GoValue() interface{}
}

// TOONObject represents a TOON object (key-value pairs).
type TOONObject struct {
	// Fields holds the key-value pairs in insertion order.
	Fields []TOONField
	// fieldIndex maps keys to their index for fast lookup.
	fieldIndex map[string]int
}

// TOONField represents a single field in a TOON object.
type TOONField struct {
	Key   string
	Value TOONValue
}

// NewTOONObject creates a new empty TOON object.
func NewTOONObject() *TOONObject {
	return &TOONObject{
		Fields:     make([]TOONField, 0),
		fieldIndex: make(map[string]int),
	}
}

// Set adds or updates a field in the object.
func (o *TOONObject) Set(key string, value TOONValue) {
	if idx, exists := o.fieldIndex[key]; exists {
		o.Fields[idx].Value = value
		return
	}
	o.fieldIndex[key] = len(o.Fields)
	o.Fields = append(o.Fields, TOONField{Key: key, Value: value})
}

// Get retrieves a field value by key.
func (o *TOONObject) Get(key string) (TOONValue, bool) {
	if idx, exists := o.fieldIndex[key]; exists {
		return o.Fields[idx].Value, true
	}
	return nil, false
}

// ToTOON converts the object to TOON format.
func (o *TOONObject) ToTOON(opts *NativeEncoderOptions) string {
	if opts == nil {
		opts = DefaultNativeEncoderOptions()
	}

	if len(o.Fields) == 0 {
		return "{}"
	}

	var buf bytes.Buffer
	for i, field := range o.Fields {
		if i > 0 {
			buf.WriteString(opts.FieldDelimiter)
		}

		key := field.Key
		if opts.AbbreviateKeys {
			key = abbreviateKey(key)
		}

		buf.WriteString(key)
		buf.WriteString(opts.KeyValueDelimiter)

		if field.Value != nil {
			buf.WriteString(field.Value.ToTOON(opts))
		} else {
			buf.WriteString("_")
		}
	}
	return buf.String()
}

// Type returns the type indicator for objects.
func (o *TOONObject) Type() string {
	return "o"
}

// GoValue returns the object as a map[string]interface{}.
func (o *TOONObject) GoValue() interface{} {
	result := make(map[string]interface{})
	for _, field := range o.Fields {
		if field.Value != nil {
			result[field.Key] = field.Value.GoValue()
		} else {
			result[field.Key] = nil
		}
	}
	return result
}

// TOONArray represents a TOON array.
type TOONArray struct {
	Elements []TOONValue
}

// NewTOONArray creates a new TOON array with the given elements.
func NewTOONArray(elements ...TOONValue) *TOONArray {
	return &TOONArray{Elements: elements}
}

// Append adds an element to the array.
func (a *TOONArray) Append(value TOONValue) {
	a.Elements = append(a.Elements, value)
}

// ToTOON converts the array to TOON format.
func (a *TOONArray) ToTOON(opts *NativeEncoderOptions) string {
	if opts == nil {
		opts = DefaultNativeEncoderOptions()
	}

	if len(a.Elements) == 0 {
		return "[]"
	}

	var buf bytes.Buffer
	buf.WriteString("[")
	for i, elem := range a.Elements {
		if i > 0 {
			buf.WriteString(opts.ArrayDelimiter)
		}
		if elem != nil {
			buf.WriteString(elem.ToTOON(opts))
		} else {
			buf.WriteString("_")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

// Type returns the type indicator for arrays.
func (a *TOONArray) Type() string {
	return "a"
}

// GoValue returns the array as []interface{}.
func (a *TOONArray) GoValue() interface{} {
	result := make([]interface{}, len(a.Elements))
	for i, elem := range a.Elements {
		if elem != nil {
			result[i] = elem.GoValue()
		} else {
			result[i] = nil
		}
	}
	return result
}

// TOONString represents a TOON string value.
type TOONString struct {
	Value string
}

// NewTOONString creates a new TOON string.
func NewTOONString(value string) *TOONString {
	return &TOONString{Value: value}
}

// ToTOON converts the string to TOON format.
func (s *TOONString) ToTOON(opts *NativeEncoderOptions) string {
	if opts == nil {
		opts = DefaultNativeEncoderOptions()
	}

	// Check if the string needs escaping
	escaped := escapeString(s.Value, opts)

	if opts.OmitTypeIndicators {
		return escaped
	}

	// For short strings without special chars, omit type indicator
	if opts.CompressionLevel > 0 && !needsTypeIndicator(s.Value, opts) {
		return escaped
	}

	return "s=" + escaped
}

// Type returns the type indicator for strings.
func (s *TOONString) Type() string {
	return "s"
}

// GoValue returns the underlying string.
func (s *TOONString) GoValue() interface{} {
	return s.Value
}

// TOONNumber represents a TOON numeric value.
type TOONNumber struct {
	IntValue   int64
	FloatValue float64
	IsFloat    bool
}

// NewTOONInt creates a new TOON integer.
func NewTOONInt(value int64) *TOONNumber {
	return &TOONNumber{IntValue: value, IsFloat: false}
}

// NewTOONFloat creates a new TOON float.
func NewTOONFloat(value float64) *TOONNumber {
	return &TOONNumber{FloatValue: value, IsFloat: true}
}

// ToTOON converts the number to TOON format.
func (n *TOONNumber) ToTOON(opts *NativeEncoderOptions) string {
	if opts == nil {
		opts = DefaultNativeEncoderOptions()
	}

	var numStr string
	if n.IsFloat {
		numStr = strconv.FormatFloat(n.FloatValue, 'f', -1, 64)
		// Remove trailing zeros after decimal point
		if strings.Contains(numStr, ".") {
			numStr = strings.TrimRight(numStr, "0")
			numStr = strings.TrimRight(numStr, ".")
		}
	} else {
		numStr = strconv.FormatInt(n.IntValue, 10)
	}

	if opts.OmitTypeIndicators {
		return numStr
	}

	// Numbers are self-evident, can omit type indicator in high compression
	if opts.CompressionLevel > 1 {
		return numStr
	}

	return "n=" + numStr
}

// Type returns the type indicator for numbers.
func (n *TOONNumber) Type() string {
	return "n"
}

// GoValue returns the underlying numeric value.
func (n *TOONNumber) GoValue() interface{} {
	if n.IsFloat {
		return n.FloatValue
	}
	return n.IntValue
}

// TOONBool represents a TOON boolean value.
type TOONBool struct {
	Value bool
}

// NewTOONBool creates a new TOON boolean.
func NewTOONBool(value bool) *TOONBool {
	return &TOONBool{Value: value}
}

// ToTOON converts the boolean to TOON format.
func (b *TOONBool) ToTOON(opts *NativeEncoderOptions) string {
	if opts == nil {
		opts = DefaultNativeEncoderOptions()
	}

	var boolStr string
	if b.Value {
		boolStr = "1"
	} else {
		boolStr = "0"
	}

	if opts.OmitTypeIndicators {
		return boolStr
	}

	if opts.CompressionLevel > 1 {
		return boolStr
	}

	return "b=" + boolStr
}

// Type returns the type indicator for booleans.
func (b *TOONBool) Type() string {
	return "b"
}

// GoValue returns the underlying boolean.
func (b *TOONBool) GoValue() interface{} {
	return b.Value
}

// TOONNull represents a TOON null value.
type TOONNull struct{}

// NewTOONNull creates a new TOON null.
func NewTOONNull() *TOONNull {
	return &TOONNull{}
}

// ToTOON converts null to TOON format.
func (n *TOONNull) ToTOON(opts *NativeEncoderOptions) string {
	return "_"
}

// Type returns the type indicator for null.
func (n *TOONNull) Type() string {
	return "_"
}

// GoValue returns nil.
func (n *TOONNull) GoValue() interface{} {
	return nil
}

// NativeEncoderOptions configures the native TOON encoder behavior.
// This is separate from the JSON-based EncoderOptions to avoid conflicts.
type NativeEncoderOptions struct {
	// FieldDelimiter separates fields in objects (default: |)
	FieldDelimiter string
	// ArrayDelimiter separates elements in arrays (default: ;)
	ArrayDelimiter string
	// KeyValueDelimiter separates keys from values (default: :)
	KeyValueDelimiter string
	// CompressionLevel controls token optimization (0-3)
	// 0: No compression, full type indicators
	// 1: Omit obvious type indicators
	// 2: Maximum compression, minimal indicators
	// 3: Aggressive compression with key abbreviation
	CompressionLevel int
	// AbbreviateKeys abbreviates common key names
	AbbreviateKeys bool
	// OmitTypeIndicators removes all type indicators
	OmitTypeIndicators bool
	// OmitNullValues skips null values entirely
	OmitNullValues bool
}

// DefaultNativeEncoderOptions returns the default native encoder options.
func DefaultNativeEncoderOptions() *NativeEncoderOptions {
	return &NativeEncoderOptions{
		FieldDelimiter:     "|",
		ArrayDelimiter:     ";",
		KeyValueDelimiter:  ":",
		CompressionLevel:   1,
		AbbreviateKeys:     false,
		OmitTypeIndicators: false,
		OmitNullValues:     false,
	}
}

// HighCompressionNativeOptions returns options for maximum token savings.
func HighCompressionNativeOptions() *NativeEncoderOptions {
	return &NativeEncoderOptions{
		FieldDelimiter:     "|",
		ArrayDelimiter:     ";",
		KeyValueDelimiter:  ":",
		CompressionLevel:   3,
		AbbreviateKeys:     true,
		OmitTypeIndicators: true,
		OmitNullValues:     true,
	}
}

// Helper functions

// abbreviateKey abbreviates common keys to save tokens.
func abbreviateKey(key string) string {
	abbreviations := map[string]string{
		"id":           "i",
		"name":         "n",
		"type":         "t",
		"value":        "v",
		"status":       "st",
		"score":        "sc",
		"created_at":   "ca",
		"updated_at":   "ua",
		"description":  "d",
		"content":      "c",
		"message":      "m",
		"error":        "e",
		"result":       "r",
		"data":         "da",
		"items":        "it",
		"count":        "ct",
		"total":        "to",
		"page":         "pg",
		"limit":        "lm",
		"offset":       "of",
		"provider_id":  "pi",
		"model_id":     "mi",
		"confidence":   "cf",
		"timestamp":    "ts",
		"latency":      "la",
		"token_count":  "tc",
		"max_tokens":   "mt",
		"temperature":  "tp",
		"participant":  "pt",
		"response":     "rs",
		"request":      "rq",
		"enabled":      "en",
		"disabled":     "di",
		"active":       "ac",
		"pending":      "pe",
		"completed":    "cp",
		"failed":       "fl",
		"priority":     "pr",
		"progress":     "pg",
		"version":      "vr",
		"context":      "cx",
		"capabilities": "cap",
	}

	if abbr, exists := abbreviations[key]; exists {
		return abbr
	}

	// For unknown keys, use first 3 chars if key is long enough
	if len(key) > 5 {
		return key[:3]
	}
	return key
}

// escapeString escapes special characters in a string.
func escapeString(s string, opts *NativeEncoderOptions) string {
	// Characters that need escaping based on delimiters
	needsEscape := false
	for _, c := range s {
		if string(c) == opts.FieldDelimiter ||
			string(c) == opts.ArrayDelimiter ||
			string(c) == opts.KeyValueDelimiter ||
			c == '[' || c == ']' || c == '(' || c == ')' ||
			c == '\\' || c == '"' {
			needsEscape = true
			break
		}
	}

	if !needsEscape {
		return s
	}

	var buf bytes.Buffer
	for _, c := range s {
		switch {
		case string(c) == opts.FieldDelimiter:
			buf.WriteString("\\|")
		case string(c) == opts.ArrayDelimiter:
			buf.WriteString("\\;")
		case string(c) == opts.KeyValueDelimiter:
			buf.WriteString("\\:")
		case c == '[':
			buf.WriteString("\\[")
		case c == ']':
			buf.WriteString("\\]")
		case c == '(':
			buf.WriteString("\\(")
		case c == ')':
			buf.WriteString("\\)")
		case c == '\\':
			buf.WriteString("\\\\")
		case c == '"':
			buf.WriteString("\\\"")
		default:
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

// unescapeString unescapes a TOON string.
func unescapeString(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}

	var buf bytes.Buffer
	escaped := false
	for _, c := range s {
		if escaped {
			switch c {
			case '|':
				buf.WriteRune('|')
			case ';':
				buf.WriteRune(';')
			case ':':
				buf.WriteRune(':')
			case '[':
				buf.WriteRune('[')
			case ']':
				buf.WriteRune(']')
			case '(':
				buf.WriteRune('(')
			case ')':
				buf.WriteRune(')')
			case '\\':
				buf.WriteRune('\\')
			case '"':
				buf.WriteRune('"')
			default:
				buf.WriteRune('\\')
				buf.WriteRune(c)
			}
			escaped = false
		} else if c == '\\' {
			escaped = true
		} else {
			buf.WriteRune(c)
		}
	}
	if escaped {
		buf.WriteRune('\\')
	}
	return buf.String()
}

// needsTypeIndicator determines if a string value needs a type indicator.
func needsTypeIndicator(s string, opts *NativeEncoderOptions) bool {
	// If string could be confused with a number or boolean, it needs indicator
	if s == "1" || s == "0" || s == "_" {
		return true
	}

	// Check if it looks like a number
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}

	// Check for special characters that might cause parsing issues
	if strings.ContainsAny(s, opts.FieldDelimiter+opts.ArrayDelimiter+opts.KeyValueDelimiter+"[]()") {
		return true
	}

	return false
}

// String returns a human-readable string representation.
func (o *TOONObject) String() string {
	return fmt.Sprintf("TOONObject{%d fields}", len(o.Fields))
}

// String returns a human-readable string representation.
func (a *TOONArray) String() string {
	return fmt.Sprintf("TOONArray{%d elements}", len(a.Elements))
}

// String returns a human-readable string representation.
func (s *TOONString) String() string {
	return fmt.Sprintf("TOONString{%q}", s.Value)
}

// String returns a human-readable string representation.
func (n *TOONNumber) String() string {
	if n.IsFloat {
		return fmt.Sprintf("TOONNumber{%f}", n.FloatValue)
	}
	return fmt.Sprintf("TOONNumber{%d}", n.IntValue)
}

// String returns a human-readable string representation.
func (b *TOONBool) String() string {
	return fmt.Sprintf("TOONBool{%v}", b.Value)
}

// String returns a human-readable string representation.
func (n *TOONNull) String() string {
	return "TOONNull{}"
}
