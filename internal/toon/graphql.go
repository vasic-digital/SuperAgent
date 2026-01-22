package toon

import (
	"encoding/json"
	"fmt"

	gqltypes "dev.helix.agent/internal/graphql/types"
)

// GraphQLConverter provides conversion between GraphQL responses and TOON format.
type GraphQLConverter struct {
	encoder       *Encoder
	nativeEncoder *NativeEncoder
	decoder       *Decoder
	nativeDecoder *NativeDecoder
}

// GraphQLConverterOptions configures the GraphQL converter.
type GraphQLConverterOptions struct {
	// UseNativeFormat uses the native TOON format instead of JSON-compressed format
	UseNativeFormat bool
	// Compression level for JSON-based format
	Compression CompressionLevel
	// NativeOptions for native TOON format
	NativeOptions *NativeEncoderOptions
}

// DefaultGraphQLConverterOptions returns default converter options.
func DefaultGraphQLConverterOptions() *GraphQLConverterOptions {
	return &GraphQLConverterOptions{
		UseNativeFormat: false,
		Compression:     CompressionStandard,
		NativeOptions:   DefaultNativeEncoderOptions(),
	}
}

// NewGraphQLConverter creates a new GraphQL to TOON converter.
func NewGraphQLConverter(opts *GraphQLConverterOptions) *GraphQLConverter {
	if opts == nil {
		opts = DefaultGraphQLConverterOptions()
	}

	return &GraphQLConverter{
		encoder: NewEncoder(&EncoderOptions{
			Compression: opts.Compression,
		}),
		nativeEncoder: NewNativeEncoder(opts.NativeOptions),
		decoder:       NewDecoder(nil),
		nativeDecoder: NewNativeDecoder(opts.NativeOptions),
	}
}

// EncodeGraphQLResponse encodes a GraphQL response to TOON format.
func (c *GraphQLConverter) EncodeGraphQLResponse(data interface{}) ([]byte, error) {
	return c.encoder.Encode(data)
}

// EncodeGraphQLResponseNative encodes a GraphQL response to native TOON format.
func (c *GraphQLConverter) EncodeGraphQLResponseNative(data interface{}) (string, error) {
	return c.nativeEncoder.Encode(data)
}

// DecodeGraphQLResponse decodes a TOON response back to Go value.
func (c *GraphQLConverter) DecodeGraphQLResponse(data []byte, v interface{}) error {
	return c.decoder.Decode(data, v)
}

// DecodeGraphQLResponseNative decodes a native TOON response back to Go value.
func (c *GraphQLConverter) DecodeGraphQLResponseNative(toonStr string) (interface{}, error) {
	return c.nativeDecoder.DecodeToGo(toonStr)
}

// Provider-specific converters

// EncodeProvider encodes a Provider to TOON format.
func (c *GraphQLConverter) EncodeProvider(provider *gqltypes.Provider) ([]byte, error) {
	return c.encoder.Encode(provider)
}

// EncodeProviderNative encodes a Provider to native TOON format.
func (c *GraphQLConverter) EncodeProviderNative(provider *gqltypes.Provider) (string, error) {
	return c.nativeEncoder.Encode(provider)
}

// EncodeProviders encodes multiple Providers to TOON format.
func (c *GraphQLConverter) EncodeProviders(providers []gqltypes.Provider) ([]byte, error) {
	return c.encoder.Encode(providers)
}

// EncodeProvidersNative encodes multiple Providers to native TOON format.
func (c *GraphQLConverter) EncodeProvidersNative(providers []gqltypes.Provider) (string, error) {
	return c.nativeEncoder.Encode(providers)
}

// EncodeDebate encodes a Debate to TOON format.
func (c *GraphQLConverter) EncodeDebate(debate *gqltypes.Debate) ([]byte, error) {
	return c.encoder.Encode(debate)
}

// EncodeDebateNative encodes a Debate to native TOON format.
func (c *GraphQLConverter) EncodeDebateNative(debate *gqltypes.Debate) (string, error) {
	return c.nativeEncoder.Encode(debate)
}

// EncodeDebates encodes multiple Debates to TOON format.
func (c *GraphQLConverter) EncodeDebates(debates []gqltypes.Debate) ([]byte, error) {
	return c.encoder.Encode(debates)
}

// EncodeDebatesNative encodes multiple Debates to native TOON format.
func (c *GraphQLConverter) EncodeDebatesNative(debates []gqltypes.Debate) (string, error) {
	return c.nativeEncoder.Encode(debates)
}

// EncodeTask encodes a Task to TOON format.
func (c *GraphQLConverter) EncodeTask(task *gqltypes.Task) ([]byte, error) {
	return c.encoder.Encode(task)
}

// EncodeTaskNative encodes a Task to native TOON format.
func (c *GraphQLConverter) EncodeTaskNative(task *gqltypes.Task) (string, error) {
	return c.nativeEncoder.Encode(task)
}

// EncodeTasks encodes multiple Tasks to TOON format.
func (c *GraphQLConverter) EncodeTasks(tasks []gqltypes.Task) ([]byte, error) {
	return c.encoder.Encode(tasks)
}

// EncodeTasksNative encodes multiple Tasks to native TOON format.
func (c *GraphQLConverter) EncodeTasksNative(tasks []gqltypes.Task) (string, error) {
	return c.nativeEncoder.Encode(tasks)
}

// EncodeVerificationResults encodes VerificationResults to TOON format.
func (c *GraphQLConverter) EncodeVerificationResults(results *gqltypes.VerificationResults) ([]byte, error) {
	return c.encoder.Encode(results)
}

// EncodeVerificationResultsNative encodes VerificationResults to native TOON format.
func (c *GraphQLConverter) EncodeVerificationResultsNative(results *gqltypes.VerificationResults) (string, error) {
	return c.nativeEncoder.Encode(results)
}

// EncodeProviderScores encodes ProviderScores to TOON format.
func (c *GraphQLConverter) EncodeProviderScores(scores []gqltypes.ProviderScore) ([]byte, error) {
	return c.encoder.Encode(scores)
}

// EncodeProviderScoresNative encodes ProviderScores to native TOON format.
func (c *GraphQLConverter) EncodeProviderScoresNative(scores []gqltypes.ProviderScore) (string, error) {
	return c.nativeEncoder.Encode(scores)
}

// GraphQL Response Wrapper

// GraphQLResponse represents a standard GraphQL response.
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents an error location.
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// EncodeGraphQLFullResponse encodes a full GraphQL response (with data and errors).
func (c *GraphQLConverter) EncodeGraphQLFullResponse(resp *GraphQLResponse) ([]byte, error) {
	return c.encoder.Encode(resp)
}

// EncodeGraphQLFullResponseNative encodes a full GraphQL response to native TOON format.
func (c *GraphQLConverter) EncodeGraphQLFullResponseNative(resp *GraphQLResponse) (string, error) {
	return c.nativeEncoder.Encode(resp)
}

// Convenience functions

// GraphQLToTOON converts a GraphQL JSON response to TOON format.
func GraphQLToTOON(jsonResponse []byte) ([]byte, error) {
	return GraphQLToTOONWithOptions(jsonResponse, nil)
}

// GraphQLToTOONWithOptions converts a GraphQL JSON response to TOON format with options.
func GraphQLToTOONWithOptions(jsonResponse []byte, opts *GraphQLConverterOptions) ([]byte, error) {
	converter := NewGraphQLConverter(opts)

	var data interface{}
	if err := json.Unmarshal(jsonResponse, &data); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return converter.EncodeGraphQLResponse(data)
}

// GraphQLToNativeTOON converts a GraphQL JSON response to native TOON format.
func GraphQLToNativeTOON(jsonResponse []byte) (string, error) {
	return GraphQLToNativeTOONWithOptions(jsonResponse, nil)
}

// GraphQLToNativeTOONWithOptions converts a GraphQL JSON response to native TOON format with options.
func GraphQLToNativeTOONWithOptions(jsonResponse []byte, opts *GraphQLConverterOptions) (string, error) {
	converter := NewGraphQLConverter(opts)

	var data interface{}
	if err := json.Unmarshal(jsonResponse, &data); err != nil {
		return "", fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return converter.EncodeGraphQLResponseNative(data)
}

// TOONToGraphQL converts a TOON response back to GraphQL JSON format.
func TOONToGraphQL(toonResponse []byte) ([]byte, error) {
	decoder := NewDecoder(nil)
	var data interface{}
	if err := decoder.Decode(toonResponse, &data); err != nil {
		return nil, fmt.Errorf("failed to decode TOON response: %w", err)
	}
	return json.Marshal(data)
}

// NativeTOONToGraphQL converts a native TOON response back to GraphQL JSON format.
func NativeTOONToGraphQL(toonResponse string) ([]byte, error) {
	decoder := NewNativeDecoder(nil)
	data, err := decoder.DecodeToGo(toonResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode native TOON response: %w", err)
	}
	return json.Marshal(data)
}

// EstimateTokenSavings estimates the token savings for a GraphQL response.
func EstimateTokenSavings(jsonResponse []byte) (float64, error) {
	toonResponse, err := GraphQLToTOON(jsonResponse)
	if err != nil {
		return 0, err
	}

	jsonTokens := len(jsonResponse) / 4 // Rough estimate
	toonTokens := len(toonResponse) / 4

	if jsonTokens == 0 {
		return 0, nil
	}

	return float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100, nil
}

// EstimateNativeTokenSavings estimates the token savings for native TOON format.
func EstimateNativeTokenSavings(jsonResponse []byte) (float64, error) {
	toonResponse, err := GraphQLToNativeTOON(jsonResponse)
	if err != nil {
		return 0, err
	}

	jsonTokens := len(jsonResponse) / 4 // Rough estimate
	toonTokens := len(toonResponse) / 4

	if jsonTokens == 0 {
		return 0, nil
	}

	return float64(jsonTokens-toonTokens) / float64(jsonTokens) * 100, nil
}
