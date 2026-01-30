package streaming

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEntityExtractor_New(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())
	require.NotNil(t, extractor)
	assert.NotNil(t, extractor.emailRegex)
	assert.NotNil(t, extractor.urlRegex)
	assert.NotNil(t, extractor.codeBlockRegex)
	assert.NotNil(t, extractor.mentionRegex)
}

func TestEntityExtractor_ExtractEmails(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "single email",
			text:     "Contact me at john@example.com for more info.",
			expected: 1,
		},
		{
			name:     "multiple emails",
			text:     "Email john@example.com or jane@company.org",
			expected: 2,
		},
		{
			name:     "no emails",
			text:     "No email addresses here",
			expected: 0,
		},
		{
			name:     "duplicate emails",
			text:     "john@example.com and john@example.com again",
			expected: 1, // Should deduplicate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := extractor.Extract(tt.text)
			emailCount := 0
			for _, e := range entities {
				if e.Type == "email" {
					emailCount++
				}
			}
			assert.Equal(t, tt.expected, emailCount)
		})
	}
}

func TestEntityExtractor_ExtractURLs(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "single URL",
			text:     "Visit https://example.com for details",
			expected: 1,
		},
		{
			name:     "multiple URLs",
			text:     "Check https://example.com and http://test.org",
			expected: 2,
		},
		{
			name:     "no URLs",
			text:     "No links in this text",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := extractor.Extract(tt.text)
			urlCount := 0
			for _, e := range entities {
				if e.Type == "url" {
					urlCount++
				}
			}
			assert.Equal(t, tt.expected, urlCount)
		})
	}
}

func TestEntityExtractor_ExtractCodeBlocks(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	tests := []struct {
		name     string
		text     string
		expected int
		language string
	}{
		{
			name: "Python code block",
			text: "Here's some code:\n```python\ndef hello():\n    print('Hello')\n```",
			expected: 1,
			language: "python",
		},
		{
			name: "Go code block",
			text: "Example:\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			expected: 1,
			language: "go",
		},
		{
			name: "Code block without language",
			text: "```\nsome code\n```",
			expected: 1,
			language: "unknown",
		},
		{
			name:     "No code blocks",
			text:     "Plain text without code",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := extractor.Extract(tt.text)
			codeBlockCount := 0
			for _, e := range entities {
				if e.Type == "code_block" {
					codeBlockCount++
					if tt.language != "" && e.Properties != nil {
						assert.Equal(t, tt.language, e.Properties["language"])
					}
				}
			}
			assert.Equal(t, tt.expected, codeBlockCount)
		})
	}
}

func TestEntityExtractor_ExtractMentions(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "single mention",
			text:     "Thanks @alice for the help",
			expected: 1,
		},
		{
			name:     "multiple mentions",
			text:     "@bob and @charlie please review",
			expected: 2,
		},
		{
			name:     "duplicate mentions",
			text:     "@alice and @alice again",
			expected: 1, // Should deduplicate
		},
		{
			name:     "no mentions",
			text:     "No mentions here",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := extractor.Extract(tt.text)
			mentionCount := 0
			for _, e := range entities {
				if e.Type == "mention" {
					mentionCount++
				}
			}
			assert.Equal(t, tt.expected, mentionCount)
		})
	}
}

func TestEntityExtractor_ExtractProgrammingTerms(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "database technologies",
			text:     "We use Kafka and PostgreSQL for our backend",
			expected: []string{"Kafka", "PostgreSQL"},
		},
		{
			name:     "programming languages",
			text:     "Our stack includes Python, Go, and TypeScript",
			expected: []string{"Python", "Go", "TypeScript"},
		},
		{
			name:     "infrastructure",
			text:     "Deploy with Docker and Kubernetes",
			expected: []string{"Docker", "Kubernetes"},
		},
		{
			name:     "protocols",
			text:     "API supports REST, GraphQL, and WebSocket",
			expected: []string{"API", "REST", "GraphQL", "WebSocket"},
		},
		{
			name:     "mixed technologies",
			text:     "Build a REST API with PostgreSQL database using Docker containers",
			expected: []string{"REST", "API", "PostgreSQL", "database", "Docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := extractor.Extract(tt.text)
			techEntities := make(map[string]bool)
			for _, e := range entities {
				if e.Type == "technology" {
					techEntities[e.Name] = true
				}
			}

			for _, expectedTerm := range tt.expected {
				assert.True(t, techEntities[expectedTerm],
					"Expected to find technology term: %s", expectedTerm)
			}
		})
	}
}

func TestEntityExtractor_CategorizeTechnology(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	tests := []struct {
		keyword  string
		expected string
	}{
		{"Kafka", "database"},
		{"PostgreSQL", "database"},
		{"Redis", "database"},
		{"Python", "programming_language"},
		{"Go", "programming_language"},
		{"JavaScript", "programming_language"},
		{"Docker", "infrastructure"},
		{"Kubernetes", "infrastructure"},
		{"REST", "protocol"},
		{"GraphQL", "protocol"},
		{"WebSocket", "protocol"},
		{"API", "general"},
	}

	for _, tt := range tests {
		t.Run(tt.keyword, func(t *testing.T) {
			category := extractor.categorizeTechnology(tt.keyword)
			assert.Equal(t, tt.expected, category)
		})
	}
}

func TestEntityExtractor_ExtractBatch(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	texts := []string{
		"Email me at test@example.com",
		"Visit https://example.com",
		"We use PostgreSQL database",
	}

	results := extractor.ExtractBatch(texts)
	assert.Len(t, results, 3)

	// First text should have email entities
	assert.NotEmpty(t, results[0])

	// Second text should have URL entities
	assert.NotEmpty(t, results[1])

	// Third text should have technology entities
	assert.NotEmpty(t, results[2])
}

func TestEntityExtractor_Cache(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	text := "We use Kafka and PostgreSQL for streaming analytics"

	// First extraction - should cache
	entities1 := extractor.Extract(text)
	cacheSize1 := extractor.GetCacheSize()
	assert.Equal(t, 1, cacheSize1)

	// Second extraction - should use cache
	entities2 := extractor.Extract(text)
	cacheSize2 := extractor.GetCacheSize()
	assert.Equal(t, 1, cacheSize2) // Cache size should remain the same
	assert.Equal(t, len(entities1), len(entities2))

	// Clear cache
	extractor.ClearCache()
	cacheSize3 := extractor.GetCacheSize()
	assert.Equal(t, 0, cacheSize3)

	// Extract again - should re-extract and cache
	entities3 := extractor.Extract(text)
	cacheSize4 := extractor.GetCacheSize()
	assert.Equal(t, 1, cacheSize4)
	assert.Equal(t, len(entities1), len(entities3))
}

func TestEntityExtractor_ComprehensiveExtraction(t *testing.T) {
	extractor := NewEntityExtractor(zap.NewNop())

	text := `
Hey @alice, check out this code:

` + "```go" + `
func main() {
	fmt.Println("Hello")
}
` + "```" + `

We should use PostgreSQL and Redis for the backend API.
Also, visit https://example.com and email me at test@example.com.

The stack will include Python, Go, and Docker containers.
`

	entities := extractor.Extract(text)

	// Should extract multiple entity types
	hasEmail := false
	hasURL := false
	hasCodeBlock := false
	hasMention := false
	hasTechnology := false

	for _, e := range entities {
		switch e.Type {
		case "email":
			hasEmail = true
		case "url":
			hasURL = true
		case "code_block":
			hasCodeBlock = true
		case "mention":
			hasMention = true
		case "technology":
			hasTechnology = true
		}
	}

	assert.True(t, hasEmail, "Should extract email")
	assert.True(t, hasURL, "Should extract URL")
	assert.True(t, hasCodeBlock, "Should extract code block")
	assert.True(t, hasMention, "Should extract mention")
	assert.True(t, hasTechnology, "Should extract technology terms")
}

func BenchmarkEntityExtractor_Extract(b *testing.B) {
	extractor := NewEntityExtractor(zap.NewNop())
	text := "We need to configure Kafka Streams with PostgreSQL and Redis for the backend API using Docker containers. Email me at test@example.com or visit https://example.com for more details."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.Extract(text)
	}
}

func BenchmarkEntityExtractor_ExtractWithCache(b *testing.B) {
	extractor := NewEntityExtractor(zap.NewNop())
	text := "We need to configure Kafka Streams with PostgreSQL and Redis for the backend API using Docker containers."

	// Pre-populate cache
	extractor.Extract(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.Extract(text)
	}
}
