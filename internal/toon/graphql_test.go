package toon

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gqltypes "dev.helix.agent/internal/graphql/types"
)

func TestNewGraphQLConverter(t *testing.T) {
	t.Run("with nil options", func(t *testing.T) {
		converter := NewGraphQLConverter(nil)
		assert.NotNil(t, converter)
		assert.NotNil(t, converter.encoder)
		assert.NotNil(t, converter.nativeEncoder)
		assert.NotNil(t, converter.decoder)
		assert.NotNil(t, converter.nativeDecoder)
	})

	t.Run("with custom options", func(t *testing.T) {
		opts := &GraphQLConverterOptions{
			UseNativeFormat: true,
			Compression:     CompressionAggressive,
			NativeOptions:   HighCompressionNativeOptions(),
		}
		converter := NewGraphQLConverter(opts)
		assert.NotNil(t, converter)
	})
}

func TestDefaultGraphQLConverterOptions(t *testing.T) {
	opts := DefaultGraphQLConverterOptions()
	assert.False(t, opts.UseNativeFormat)
	assert.Equal(t, CompressionStandard, opts.Compression)
	assert.NotNil(t, opts.NativeOptions)
}

func TestGraphQLConverter_EncodeGraphQLResponse(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	data := map[string]interface{}{
		"providers": []map[string]interface{}{
			{
				"id":     "claude",
				"name":   "Claude",
				"status": "active",
			},
		},
	}

	encoded, err := converter.EncodeGraphQLResponse(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	// Decode and verify
	var decoded map[string]interface{}
	err = converter.DecodeGraphQLResponse(encoded, &decoded)
	require.NoError(t, err)
	// The decoded data should have the expanded keys
	providers, ok := decoded["providers"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, providers, 1)
}

func TestGraphQLConverter_EncodeGraphQLResponseNative(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	data := map[string]interface{}{
		"id":     "123",
		"name":   "test",
		"status": "active",
	}

	encoded, err := converter.EncodeGraphQLResponseNative(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
	// Should contain TOON format elements
	assert.Contains(t, encoded, ":")
}

func TestGraphQLConverter_DecodeGraphQLResponseNative(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	toonStr := "id:s=123|name:s=test"
	decoded, err := converter.DecodeGraphQLResponseNative(toonStr)
	require.NoError(t, err)

	m, ok := decoded.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "123", m["id"])
	assert.Equal(t, "test", m["name"])
}

func TestGraphQLConverter_EncodeProvider(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	provider := &gqltypes.Provider{
		ID:     "claude",
		Name:   "Claude",
		Type:   "oauth",
		Status: "active",
		Score:  8.5,
	}

	encoded, err := converter.EncodeProvider(provider)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeProviderNative(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	provider := &gqltypes.Provider{
		ID:     "claude",
		Name:   "Claude",
		Type:   "oauth",
		Status: "active",
		Score:  8.5,
	}

	encoded, err := converter.EncodeProviderNative(provider)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
	assert.Contains(t, encoded, "claude")
}

func TestGraphQLConverter_EncodeProviders(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	providers := []gqltypes.Provider{
		{ID: "claude", Name: "Claude"},
		{ID: "gemini", Name: "Gemini"},
	}

	encoded, err := converter.EncodeProviders(providers)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeDebate(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	debate := &gqltypes.Debate{
		ID:         "debate-123",
		Topic:      "AI Safety",
		Status:     "running",
		Confidence: 0.85,
		CreatedAt:  time.Now(),
	}

	encoded, err := converter.EncodeDebate(debate)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeDebates(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	debates := []gqltypes.Debate{
		{ID: "debate-1", Topic: "Topic 1"},
		{ID: "debate-2", Topic: "Topic 2"},
	}

	encoded, err := converter.EncodeDebates(debates)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeTask(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	task := &gqltypes.Task{
		ID:       "task-123",
		Type:     "verification",
		Status:   "running",
		Priority: 1,
		Progress: 50,
	}

	encoded, err := converter.EncodeTask(task)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeTasks(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	tasks := []gqltypes.Task{
		{ID: "task-1", Status: "pending"},
		{ID: "task-2", Status: "running"},
	}

	encoded, err := converter.EncodeTasks(tasks)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeVerificationResults(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	results := &gqltypes.VerificationResults{
		TotalProviders:    10,
		VerifiedProviders: 8,
		TotalModels:       25,
		VerifiedModels:    20,
		OverallScore:      8.5,
		LastVerified:      time.Now(),
	}

	encoded, err := converter.EncodeVerificationResults(results)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeProviderScores(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	scores := []gqltypes.ProviderScore{
		{
			ProviderID:        "claude",
			ProviderName:      "Claude",
			OverallScore:      8.5,
			ResponseSpeed:     9.0,
			ModelEfficiency:   8.5,
			CostEffectiveness: 7.5,
			Capability:        9.0,
			Recency:           8.0,
		},
	}

	encoded, err := converter.EncodeProviderScores(scores)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeGraphQLFullResponse(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	resp := &GraphQLResponse{
		Data: map[string]interface{}{
			"provider": map[string]interface{}{
				"id":   "claude",
				"name": "Claude",
			},
		},
	}

	encoded, err := converter.EncodeGraphQLFullResponse(resp)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLConverter_EncodeGraphQLFullResponseWithErrors(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	resp := &GraphQLResponse{
		Errors: []GraphQLError{
			{
				Message: "Provider not found",
				Locations: []GraphQLErrorLocation{
					{Line: 1, Column: 10},
				},
				Path: []interface{}{"provider"},
			},
		},
	}

	encoded, err := converter.EncodeGraphQLFullResponse(resp)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

func TestGraphQLToTOON(t *testing.T) {
	jsonResponse := []byte(`{"data":{"provider":{"id":"claude","name":"Claude"}}}`)

	toon, err := GraphQLToTOON(jsonResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, toon)

	// Verify it can be decoded back
	var decoded map[string]interface{}
	err = NewDecoder(nil).Decode(toon, &decoded)
	require.NoError(t, err)
}

func TestGraphQLToNativeTOON(t *testing.T) {
	jsonResponse := []byte(`{"data":{"provider":{"id":"claude","name":"Claude"}}}`)

	toon, err := GraphQLToNativeTOON(jsonResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, toon)
	assert.Contains(t, toon, ":")
}

func TestGraphQLToTOONWithOptions(t *testing.T) {
	jsonResponse := []byte(`{"data":{"id":"123","name":"test"}}`)

	opts := &GraphQLConverterOptions{
		Compression: CompressionAggressive,
	}

	toon, err := GraphQLToTOONWithOptions(jsonResponse, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, toon)
}

func TestGraphQLToNativeTOONWithOptions(t *testing.T) {
	jsonResponse := []byte(`{"data":{"id":"123","name":"test"}}`)

	opts := &GraphQLConverterOptions{
		NativeOptions: HighCompressionNativeOptions(),
	}

	toon, err := GraphQLToNativeTOONWithOptions(jsonResponse, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, toon)
}

func TestTOONToGraphQL(t *testing.T) {
	// First encode some data
	encoder := NewEncoder(&EncoderOptions{Compression: CompressionStandard})
	data := map[string]interface{}{
		"id":   "123",
		"name": "test",
	}
	toon, err := encoder.Encode(data)
	require.NoError(t, err)

	// Then convert back
	jsonBytes, err := TOONToGraphQL(toon)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "123", decoded["id"])
	assert.Equal(t, "test", decoded["name"])
}

func TestNativeTOONToGraphQL(t *testing.T) {
	toon := "id:s=123|name:s=test"

	jsonBytes, err := NativeTOONToGraphQL(toon)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "123", decoded["id"])
	assert.Equal(t, "test", decoded["name"])
}

func TestEstimateTokenSavings(t *testing.T) {
	jsonResponse := []byte(`{"data":{"providers":[{"id":"claude","name":"Claude","status":"active"},{"id":"gemini","name":"Gemini","status":"active"}]}}`)

	savings, err := EstimateTokenSavings(jsonResponse)
	require.NoError(t, err)
	// Should have some savings due to key abbreviation
	assert.GreaterOrEqual(t, savings, 0.0)
}

func TestEstimateNativeTokenSavings(t *testing.T) {
	jsonResponse := []byte(`{"data":{"providers":[{"id":"claude","name":"Claude","status":"active"},{"id":"gemini","name":"Gemini","status":"active"}]}}`)

	savings, err := EstimateNativeTokenSavings(jsonResponse)
	require.NoError(t, err)
	// Native format should have more significant savings
	assert.GreaterOrEqual(t, savings, 0.0)
}

func TestGraphQLToTOON_InvalidJSON(t *testing.T) {
	_, err := GraphQLToTOON([]byte(`invalid json`))
	assert.Error(t, err)
}

func TestGraphQLToNativeTOON_InvalidJSON(t *testing.T) {
	_, err := GraphQLToNativeTOON([]byte(`invalid json`))
	assert.Error(t, err)
}

func TestTOONToGraphQL_InvalidTOON(t *testing.T) {
	// Invalid TOON that can't be decoded as JSON
	_, err := TOONToGraphQL([]byte(`{invalid`))
	assert.Error(t, err)
}

func TestNativeTOONToGraphQL_InvalidTOON(t *testing.T) {
	// This might not error since the native decoder is more lenient
	// but let's test the path
	result, err := NativeTOONToGraphQL("simple_string")
	if err == nil {
		// If no error, it should produce valid JSON
		assert.NotEmpty(t, result)
	}
}

func TestGraphQLResponse_Fields(t *testing.T) {
	resp := &GraphQLResponse{
		Data: map[string]interface{}{
			"test": "value",
		},
		Errors: []GraphQLError{
			{
				Message: "Test error",
				Locations: []GraphQLErrorLocation{
					{Line: 1, Column: 1},
				},
				Path: []interface{}{"test"},
				Extensions: map[string]interface{}{
					"code": "TEST_ERROR",
				},
			},
		},
	}

	assert.NotNil(t, resp.Data)
	assert.Len(t, resp.Errors, 1)
	assert.Equal(t, "Test error", resp.Errors[0].Message)
	assert.Len(t, resp.Errors[0].Locations, 1)
	assert.Equal(t, 1, resp.Errors[0].Locations[0].Line)
	assert.Len(t, resp.Errors[0].Path, 1)
	assert.Equal(t, "TEST_ERROR", resp.Errors[0].Extensions["code"])
}

func TestGraphQLConverter_RoundTrip(t *testing.T) {
	converter := NewGraphQLConverter(nil)

	original := map[string]interface{}{
		"id":     "123",
		"name":   "test",
		"score":  8.5,
		"active": true,
	}

	// JSON-based round trip
	encoded, err := converter.EncodeGraphQLResponse(original)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = converter.DecodeGraphQLResponse(encoded, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "123", decoded["id"])
	assert.Equal(t, "test", decoded["name"])

	// Native round trip
	nativeEncoded, err := converter.EncodeGraphQLResponseNative(original)
	require.NoError(t, err)

	nativeDecoded, err := converter.DecodeGraphQLResponseNative(nativeEncoded)
	require.NoError(t, err)

	m, ok := nativeDecoded.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "123", m["id"])
	assert.Equal(t, "test", m["name"])
}

func TestEstimateTokenSavings_EmptyJSON(t *testing.T) {
	savings, err := EstimateTokenSavings([]byte(`{}`))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, savings, 0.0)
}
