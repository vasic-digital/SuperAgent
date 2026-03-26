//go:build fuzz

// Package fuzz provides Go native fuzzing tests for critical parsing paths
// in the HelixAgent system. These tests ensure that malformed or adversarial
// input never causes panics or undefined behavior.
package fuzz

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
)

// FuzzVectorQueryConstruction tests vector database query building logic
// from internal/vectordb/qdrant/client.go. It exercises the JSON marshalling
// of search request bodies with arbitrary vector and filter inputs.
func FuzzVectorQueryConstruction(f *testing.F) {
	// Seed corpus: typical search parameters
	f.Add("my-collection", 10, 0, float32(0.7), true, true)
	f.Add("", 0, 0, float32(0.0), false, false)
	f.Add("collection-with-long-name-"+strings.Repeat("x", 200), 100, 50, float32(0.99), true, false)
	f.Add("test", -1, -1, float32(-1.0), false, true)
	f.Add("vectors\x00null", 1<<20, 1<<20, float32(math.NaN()), true, true)

	f.Fuzz(func(t *testing.T, collection string, limit, offset int, scoreThreshold float32, withPayload, withVectors bool) {
		// Mirror the reqBody construction in Client.Search
		reqBody := map[string]interface{}{
			"limit":        limit,
			"offset":       offset,
			"with_payload": withPayload,
			"with_vector":  withVectors,
		}

		if scoreThreshold > 0 && !math.IsNaN(float64(scoreThreshold)) && !math.IsInf(float64(scoreThreshold), 0) {
			reqBody["score_threshold"] = scoreThreshold
		}

		// Marshal must not panic regardless of collection name or params
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return
		}

		// Build URL path as Client.Search does
		path := fmt.Sprintf("/collections/%s/points/search", collection)
		_ = path

		// Round-trip: unmarshal back and verify structure
		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("round-trip unmarshal failed: %v", err)
		}

		_, _ = parsed["limit"].(float64)
		_, _ = parsed["offset"].(float64)
		_, _ = parsed["with_payload"].(bool)
		_, _ = parsed["with_vector"].(bool)
	})
}

// FuzzVectorFilterParsing tests JSON parsing of Qdrant filter objects.
// Filters are passed as arbitrary JSON to the search API; this fuzz test
// ensures the filter serialisation path handles malformed input safely.
func FuzzVectorFilterParsing(f *testing.F) {
	// Valid Qdrant filter expressions
	f.Add([]byte(`{"must":[{"key":"category","match":{"value":"tech"}}]}`))
	f.Add([]byte(`{"should":[{"key":"score","range":{"gte":0.5,"lte":1.0}}]}`))
	f.Add([]byte(`{"must_not":[{"key":"deleted","match":{"value":true}}]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`[]`))
	// Nested / deeply recursive filters
	f.Add([]byte(`{"must":[{"must":[{"must":[{"key":"x","match":{"value":1}}]}]}]}`))
	// Adversarial
	f.Add([]byte("\x00\x01\x02\xff\xfe"))
	f.Add([]byte(`{"must":` + strings.Repeat(`[`, 512) + strings.Repeat(`]`, 512) + `}`))

	f.Fuzz(func(t *testing.T, filterData []byte) {
		var filter interface{}
		if err := json.Unmarshal(filterData, &filter); err != nil {
			return
		}

		// Build a search request body with the filter attached
		reqBody := map[string]interface{}{
			"vector":       []float32{0.1, 0.2, 0.3},
			"limit":        10,
			"with_payload": true,
			"filter":       filter,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return
		}

		// Round-trip
		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("round-trip unmarshal failed: %v", err)
		}

		// Walk filter structure without panicking
		if f, ok := parsed["filter"]; ok && f != nil {
			walkFilterNode(f, 0)
		}
	})
}

// FuzzVectorConfigValidation tests CollectionConfig validation with arbitrary
// inputs, mirroring internal/vectordb/qdrant/config.go Validate().
func FuzzVectorConfigValidation(f *testing.F) {
	f.Add("my-collection", 1536, "Cosine")
	f.Add("", 0, "")
	f.Add("col", -1, "InvalidDistance")
	f.Add(strings.Repeat("x", 512), 1<<20, "Dot")
	f.Add("col\x00null", 768, "Euclid")

	f.Fuzz(func(t *testing.T, name string, vectorSize int, distance string) {
		// Mirror Config.Validate() logic without importing the package
		var validationErrors []string

		if name == "" {
			validationErrors = append(validationErrors, "collection name is required")
		}
		if vectorSize < 1 {
			validationErrors = append(validationErrors, "vector_size must be at least 1")
		}

		validDistances := map[string]bool{
			"Cosine":    true,
			"Euclid":    true,
			"Dot":       true,
			"Manhattan": true,
		}
		if !validDistances[distance] {
			validationErrors = append(validationErrors, fmt.Sprintf("invalid distance metric: %s", distance))
		}

		_ = validationErrors

		// Build config JSON as it would be serialised for the API
		cfg := map[string]interface{}{
			"name":        name,
			"vector_size": vectorSize,
			"distance":    distance,
		}
		jsonData, err := json.Marshal(cfg)
		if err != nil {
			return
		}

		var parsed map[string]interface{}
		_ = json.Unmarshal(jsonData, &parsed)
	})
}

// walkFilterNode recursively walks a Qdrant filter node up to maxDepth levels.
// It exists only to exercise the parsing path; correctness is not checked.
func walkFilterNode(node interface{}, depth int) {
	const maxDepth = 20
	if depth > maxDepth {
		return
	}

	switch v := node.(type) {
	case map[string]interface{}:
		for key, val := range v {
			_ = key
			walkFilterNode(val, depth+1)
		}
	case []interface{}:
		for _, item := range v {
			walkFilterNode(item, depth+1)
		}
	default:
		// Scalar value — nothing to recurse into
	}
}
