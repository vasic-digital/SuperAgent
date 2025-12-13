// Package response provides utilities for parsing API responses.
package response

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// JSONParser provides JSON response parsing utilities.
type JSONParser struct{}

// ParseJSON parses a JSON response into the given result.
func (p *JSONParser) ParseJSON(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}

// ParseJSONFromBytes parses JSON from a byte slice.
func (p *JSONParser) ParseJSONFromBytes(data []byte, result interface{}) error {
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}

// StreamingParser handles streaming responses.
type StreamingParser struct {
	onData  func([]byte) error
	onError func(error)
}

// NewStreamingParser creates a new streaming parser.
func NewStreamingParser(onData func([]byte) error, onError func(error)) *StreamingParser {
	return &StreamingParser{
		onData:  onData,
		onError: onError,
	}
}

// ParseStream parses a streaming response.
func (p *StreamingParser) ParseStream(resp *http.Response) error {
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle SSE format (data: ...)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			if p.onData != nil {
				if err := p.onData([]byte(data)); err != nil {
					if p.onError != nil {
						p.onError(err)
					}
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if p.onError != nil {
			p.onError(err)
		}
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// ErrorDetector detects errors in API responses.
type ErrorDetector struct{}

// DetectError checks if a response contains an error.
func (d *ErrorDetector) DetectError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	// Try to parse as JSON error
	var errorResp struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
		return fmt.Errorf("API error [%s]: %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Fallback to generic error
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}

// ResponseValidator validates response structure.
type ResponseValidator struct {
	requiredFields []string
}

// NewResponseValidator creates a new response validator.
func NewResponseValidator(requiredFields ...string) *ResponseValidator {
	return &ResponseValidator{requiredFields: requiredFields}
}

// Validate validates a response map.
func (v *ResponseValidator) Validate(response map[string]interface{}) error {
	for _, field := range v.requiredFields {
		if _, exists := response[field]; !exists {
			return fmt.Errorf("required field '%s' is missing from response", field)
		}
	}
	return nil
}

// PaginationParser handles paginated responses.
type PaginationParser struct {
	hasNextPage func(map[string]interface{}) bool
	getNextURL  func(map[string]interface{}) string
}

// NewPaginationParser creates a new pagination parser.
func NewPaginationParser(hasNextPage func(map[string]interface{}) bool, getNextURL func(map[string]interface{}) string) *PaginationParser {
	return &PaginationParser{
		hasNextPage: hasNextPage,
		getNextURL:  getNextURL,
	}
}

// ParsePaginated parses a paginated response.
func (p *PaginationParser) ParsePaginated(resp *http.Response, result interface{}) (hasNext bool, nextURL string, err error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return false, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract data
	if data, exists := response["data"]; exists {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return false, "", fmt.Errorf("failed to marshal data: %w", err)
		}

		if err := json.Unmarshal(dataBytes, result); err != nil {
			return false, "", fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	// Check pagination
	hasNext = p.hasNextPage(response)
	if hasNext {
		nextURL = p.getNextURL(response)
	}

	return hasNext, nextURL, nil
}

// ChunkedParser handles chunked responses.
type ChunkedParser struct {
	chunkSize int
	onChunk   func([]byte) error
}

// NewChunkedParser creates a new chunked parser.
func NewChunkedParser(chunkSize int, onChunk func([]byte) error) *ChunkedParser {
	return &ChunkedParser{
		chunkSize: chunkSize,
		onChunk:   onChunk,
	}
}

// ParseChunked parses a response in chunks.
func (p *ChunkedParser) ParseChunked(resp *http.Response) error {
	defer resp.Body.Close()

	buffer := make([]byte, p.chunkSize)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if p.onChunk != nil {
				if err := p.onChunk(buffer[:n]); err != nil {
					return err
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading chunk: %w", err)
		}
	}

	return nil
}

// ResponseBuilder builds responses from raw data.
type ResponseBuilder struct{}

// BuildChatResponse builds a chat response from raw data.
func (b *ResponseBuilder) BuildChatResponse(data map[string]interface{}) (interface{}, error) {
	// This would be implemented based on the specific response format
	// For now, return the data as-is
	return data, nil
}

// BuildEmbeddingResponse builds an embedding response from raw data.
func (b *ResponseBuilder) BuildEmbeddingResponse(data map[string]interface{}) (interface{}, error) {
	// This would be implemented based on the specific response format
	return data, nil
}

// SanitizeResponse sanitizes response data by removing sensitive information.
func (b *ResponseBuilder) SanitizeResponse(response interface{}) interface{} {
	// Implementation would depend on what needs to be sanitized
	// For now, return as-is
	return response
}
