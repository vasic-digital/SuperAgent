// Package toon provides Token-Optimized Object Notation (TOON) encoding
// for efficient data transport to AI systems via MCP.
package toon

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// CompressionLevel represents the TOON compression level.
type CompressionLevel int

const (
	// CompressionNone disables compression.
	CompressionNone CompressionLevel = iota
	// CompressionMinimal applies key shortening only.
	CompressionMinimal
	// CompressionStandard applies key shortening and value abbreviation.
	CompressionStandard
	// CompressionAggressive applies maximum compression including gzip.
	CompressionAggressive
)

// Encoder encodes data in TOON format for token-efficient AI consumption.
type Encoder struct {
	compression   CompressionLevel
	keyMapping    map[string]string
	reverseKeyMap map[string]string
	mu            sync.RWMutex
}

// EncoderOptions configures the TOON encoder.
type EncoderOptions struct {
	Compression CompressionLevel
	KeyMapping  map[string]string // Custom key mappings
}

// DefaultEncoderOptions returns default encoder options.
func DefaultEncoderOptions() *EncoderOptions {
	return &EncoderOptions{
		Compression: CompressionStandard,
		KeyMapping:  DefaultKeyMapping(),
	}
}

// DefaultKeyMapping returns the default key abbreviation mapping.
func DefaultKeyMapping() map[string]string {
	return map[string]string{
		// Common fields
		"id":          "i",
		"name":        "n",
		"type":        "t",
		"status":      "s",
		"created_at":  "ca",
		"updated_at":  "ua",
		"timestamp":   "ts",
		"message":     "m",
		"content":     "c",
		"error":       "e",
		"result":      "r",
		"data":        "d",
		"value":       "v",
		"key":         "k",
		"description": "ds",

		// Provider fields
		"provider_id":   "pi",
		"provider_name": "pn",
		"model_id":      "mi",
		"model_name":    "mn",

		// Score fields
		"score":             "sc",
		"overall_score":     "os",
		"response_speed":    "rs",
		"model_efficiency":  "me",
		"cost_effectiveness": "ce",
		"capability":        "cp",
		"recency":           "rc",
		"confidence":        "cf",

		// Debate fields
		"debate_id":      "di",
		"topic":          "tp",
		"round_number":   "rn",
		"participants":   "pt",
		"conclusion":     "cn",
		"responses":      "rp",
		"participant_id": "pti",
		"position":       "ps",
		"role":           "rl",

		// Task fields
		"task_id":      "ti",
		"priority":     "pr",
		"progress":     "pg",
		"started_at":   "sa",
		"completed_at": "cpa",

		// Health fields
		"health_status": "hs",
		"latency_ms":    "lm",
		"last_check":    "lc",
		"error_message": "em",

		// Capability fields
		"context_window":    "cw",
		"max_tokens":        "mt",
		"supports_tools":    "st",
		"supports_vision":   "sv",
		"supports_streaming": "ss",
		"function_calling":  "fc",
		"embeddings":        "eb",
		"completions":       "cm",
		"chat":              "ch",
		"vision":            "vs",
		"tool_use":          "tu",
		"streaming":         "sm",

		// Verification fields
		"total_providers":    "tpr",
		"verified_providers": "vpr",
		"total_models":       "tmd",
		"verified_models":    "vmd",
		"last_verified":      "lv",

		// Token stream fields
		"request_id":  "ri",
		"token":       "tk",
		"is_complete": "ic",
		"token_count": "tc",
	}
}

// NewEncoder creates a new TOON encoder.
func NewEncoder(opts *EncoderOptions) *Encoder {
	if opts == nil {
		opts = DefaultEncoderOptions()
	}

	keyMap := opts.KeyMapping
	if keyMap == nil {
		keyMap = DefaultKeyMapping()
	}

	// Build reverse mapping
	reverseMap := make(map[string]string)
	for k, v := range keyMap {
		reverseMap[v] = k
	}

	return &Encoder{
		compression:   opts.Compression,
		keyMapping:    keyMap,
		reverseKeyMap: reverseMap,
	}
}

// Encode encodes data in TOON format.
func (e *Encoder) Encode(data interface{}) ([]byte, error) {
	// First marshal to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	if e.compression == CompressionNone {
		return jsonData, nil
	}

	// Apply key compression
	compressed, err := e.compressKeys(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress keys: %w", err)
	}

	// Apply gzip for aggressive compression
	if e.compression == CompressionAggressive {
		return e.gzipCompress(compressed)
	}

	return compressed, nil
}

// EncodeToString encodes data and returns as string.
func (e *Encoder) EncodeToString(data interface{}) (string, error) {
	encoded, err := e.Encode(data)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// compressKeys applies key abbreviation to JSON data.
func (e *Encoder) compressKeys(data []byte) ([]byte, error) {
	// Parse as generic map
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	// Recursively compress keys
	compressed := e.compressObject(obj)

	// Re-marshal
	return json.Marshal(compressed)
}

// compressObject recursively compresses keys in an object.
func (e *Encoder) compressObject(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			newKey := e.abbreviateKey(key)
			result[newKey] = e.compressObject(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = e.compressObject(item)
		}
		return result
	default:
		// Apply value compression for strings if standard or higher
		if e.compression >= CompressionStandard {
			if s, ok := v.(string); ok {
				return e.abbreviateValue(s)
			}
		}
		return v
	}
}

// abbreviateKey returns the abbreviated form of a key.
func (e *Encoder) abbreviateKey(key string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if abbr, ok := e.keyMapping[key]; ok {
		return abbr
	}
	return key
}

// abbreviateValue abbreviates common string values.
func (e *Encoder) abbreviateValue(value string) string {
	// Common status values
	switch strings.ToLower(value) {
	case "healthy":
		return "H"
	case "degraded":
		return "D"
	case "unhealthy":
		return "U"
	case "pending":
		return "P"
	case "running":
		return "R"
	case "completed":
		return "C"
	case "failed":
		return "F"
	case "active":
		return "A"
	case "inactive":
		return "I"
	case "queued":
		return "Q"
	case "cancelled":
		return "X"
	}
	return value
}

// gzipCompress applies gzip compression.
func (e *Encoder) gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// TokenCount estimates the token count for the encoded data.
// Uses a simple heuristic: ~4 characters per token for English text.
func (e *Encoder) TokenCount(data []byte) int {
	return len(data) / 4
}

// CompressionRatio calculates the compression ratio.
func (e *Encoder) CompressionRatio(original, compressed []byte) float64 {
	if len(original) == 0 {
		return 0
	}
	return float64(len(compressed)) / float64(len(original))
}

// AddKeyMapping adds a custom key mapping.
func (e *Encoder) AddKeyMapping(original, abbreviated string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.keyMapping[original] = abbreviated
	e.reverseKeyMap[abbreviated] = original
}

// SetCompression sets the compression level.
func (e *Encoder) SetCompression(level CompressionLevel) {
	e.compression = level
}

// GetCompressionLevel returns the current compression level.
func (e *Encoder) GetCompressionLevel() CompressionLevel {
	return e.compression
}

// Decoder decodes TOON-encoded data.
type Decoder struct {
	keyMapping    map[string]string
	reverseKeyMap map[string]string
	mu            sync.RWMutex
}

// NewDecoder creates a new TOON decoder.
func NewDecoder(opts *EncoderOptions) *Decoder {
	if opts == nil {
		opts = DefaultEncoderOptions()
	}

	keyMap := opts.KeyMapping
	if keyMap == nil {
		keyMap = DefaultKeyMapping()
	}

	// Build reverse mapping
	reverseMap := make(map[string]string)
	for k, v := range keyMap {
		reverseMap[v] = k
	}

	return &Decoder{
		keyMapping:    keyMap,
		reverseKeyMap: reverseMap,
	}
}

// Decode decodes TOON-encoded data.
func (d *Decoder) Decode(data []byte, v interface{}) error {
	// Try gzip decompress first
	decompressed, err := d.gzipDecompress(data)
	if err != nil {
		// Not gzipped, use original data
		decompressed = data
	}

	// Expand keys
	expanded, err := d.expandKeys(decompressed)
	if err != nil {
		return fmt.Errorf("failed to expand keys: %w", err)
	}

	// Unmarshal to target
	return json.Unmarshal(expanded, v)
}

// DecodeToMap decodes TOON data to a map.
func (d *Decoder) DecodeToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := d.Decode(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// expandKeys expands abbreviated keys.
func (d *Decoder) expandKeys(data []byte) ([]byte, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	expanded := d.expandObject(obj)
	return json.Marshal(expanded)
}

// expandObject recursively expands keys in an object.
func (d *Decoder) expandObject(obj interface{}) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			newKey := d.expandKey(key)
			result[newKey] = d.expandObject(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = d.expandObject(item)
		}
		return result
	default:
		// Expand abbreviated values
		if s, ok := v.(string); ok {
			return d.expandValue(s)
		}
		return v
	}
}

// expandKey returns the original form of an abbreviated key.
func (d *Decoder) expandKey(key string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if original, ok := d.reverseKeyMap[key]; ok {
		return original
	}
	return key
}

// expandValue expands abbreviated string values.
func (d *Decoder) expandValue(value string) string {
	switch value {
	case "H":
		return "healthy"
	case "D":
		return "degraded"
	case "U":
		return "unhealthy"
	case "P":
		return "pending"
	case "R":
		return "running"
	case "C":
		return "completed"
	case "F":
		return "failed"
	case "A":
		return "active"
	case "I":
		return "inactive"
	case "Q":
		return "queued"
	case "X":
		return "cancelled"
	}
	return value
}

// gzipDecompress decompresses gzip data.
func (d *Decoder) gzipDecompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
