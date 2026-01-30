package streaming

import (
	"regexp"
	"strings"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EntityExtractor extracts entities from text using pattern matching and NLP
type EntityExtractor struct {
	logger *zap.Logger

	// Pattern-based extractors
	emailRegex    *regexp.Regexp
	urlRegex      *regexp.Regexp
	codeBlockRegex *regexp.Regexp
	mentionRegex  *regexp.Regexp

	// Cache for extracted entities
	cache map[string][]EntityData
	mu    sync.RWMutex
}

// NewEntityExtractor creates a new entity extractor
func NewEntityExtractor(logger *zap.Logger) *EntityExtractor {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &EntityExtractor{
		logger: logger,
		emailRegex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		urlRegex: regexp.MustCompile(`https?://[^\s]+`),
		codeBlockRegex: regexp.MustCompile("```([a-z]+)?\\n([\\s\\S]*?)```"),
		mentionRegex: regexp.MustCompile(`@([a-zA-Z0-9_]+)`),
		cache: make(map[string][]EntityData),
	}
}

// Extract extracts entities from text
func (ee *EntityExtractor) Extract(text string) []EntityData {
	// Check cache first
	ee.mu.RLock()
	if cached, exists := ee.cache[text]; exists {
		ee.mu.RUnlock()
		return cached
	}
	ee.mu.RUnlock()

	entities := make([]EntityData, 0)

	// Extract emails
	emails := ee.extractEmails(text)
	entities = append(entities, emails...)

	// Extract URLs
	urls := ee.extractURLs(text)
	entities = append(entities, urls...)

	// Extract code blocks
	codeBlocks := ee.extractCodeBlocks(text)
	entities = append(entities, codeBlocks...)

	// Extract mentions
	mentions := ee.extractMentions(text)
	entities = append(entities, mentions...)

	// Extract programming terms
	programmingTerms := ee.extractProgrammingTerms(text)
	entities = append(entities, programmingTerms...)

	// Cache results
	ee.mu.Lock()
	ee.cache[text] = entities
	ee.mu.Unlock()

	ee.logger.Debug("Extracted entities from text",
		zap.Int("total_entities", len(entities)),
		zap.Int("text_length", len(text)))

	return entities
}

// extractEmails extracts email addresses
func (ee *EntityExtractor) extractEmails(text string) []EntityData {
	matches := ee.emailRegex.FindAllString(text, -1)
	entities := make([]EntityData, 0, len(matches))

	seen := make(map[string]bool)
	for _, email := range matches {
		if seen[email] {
			continue
		}
		seen[email] = true

		entities = append(entities, EntityData{
			EntityID:   uuid.New().String(),
			Name:       email,
			Type:       "email",
			Properties: map[string]interface{}{
				"value": email,
			},
			Importance: 0.6,
		})
	}

	return entities
}

// extractURLs extracts URLs
func (ee *EntityExtractor) extractURLs(text string) []EntityData {
	matches := ee.urlRegex.FindAllString(text, -1)
	entities := make([]EntityData, 0, len(matches))

	seen := make(map[string]bool)
	for _, url := range matches {
		if seen[url] {
			continue
		}
		seen[url] = true

		entities = append(entities, EntityData{
			EntityID:   uuid.New().String(),
			Name:       url,
			Type:       "url",
			Properties: map[string]interface{}{
				"value": url,
			},
			Importance: 0.5,
		})
	}

	return entities
}

// extractCodeBlocks extracts code blocks with language info
func (ee *EntityExtractor) extractCodeBlocks(text string) []EntityData {
	matches := ee.codeBlockRegex.FindAllStringSubmatch(text, -1)
	entities := make([]EntityData, 0, len(matches))

	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		language := match[1]
		code := match[2]

		if language == "" {
			language = "unknown"
		}

		entities = append(entities, EntityData{
			EntityID:   uuid.New().String(),
			Name:       "code_block_" + string(rune(i+1)),
			Type:       "code_block",
			Properties: map[string]interface{}{
				"language":    language,
				"code":        code,
				"line_count":  strings.Count(code, "\n") + 1,
			},
			Importance: 0.8,
		})
	}

	return entities
}

// extractMentions extracts @mentions
func (ee *EntityExtractor) extractMentions(text string) []EntityData {
	matches := ee.mentionRegex.FindAllStringSubmatch(text, -1)
	entities := make([]EntityData, 0, len(matches))

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		username := match[1]
		if seen[username] {
			continue
		}
		seen[username] = true

		entities = append(entities, EntityData{
			EntityID:   uuid.New().String(),
			Name:       username,
			Type:       "mention",
			Properties: map[string]interface{}{
				"username": username,
			},
			Importance: 0.4,
		})
	}

	return entities
}

// extractProgrammingTerms extracts common programming-related terms
func (ee *EntityExtractor) extractProgrammingTerms(text string) []EntityData {
	// Common programming keywords and frameworks
	keywords := []string{
		"Kafka", "Redis", "PostgreSQL", "Docker", "Kubernetes",
		"Python", "Go", "JavaScript", "TypeScript", "Rust",
		"API", "REST", "GraphQL", "gRPC", "WebSocket",
		"database", "cache", "queue", "stream", "event",
	}

	entities := make([]EntityData, 0)
	seen := make(map[string]bool)

	lowerText := strings.ToLower(text)

	for _, keyword := range keywords {
		lowerKeyword := strings.ToLower(keyword)
		if strings.Contains(lowerText, lowerKeyword) && !seen[keyword] {
			seen[keyword] = true

			entities = append(entities, EntityData{
				EntityID:   uuid.New().String(),
				Name:       keyword,
				Type:       "technology",
				Properties: map[string]interface{}{
					"category": ee.categorizeTechnology(keyword),
				},
				Importance: 0.7,
			})
		}
	}

	return entities
}

// categorizeTechnology categorizes a technology keyword
func (ee *EntityExtractor) categorizeTechnology(keyword string) string {
	databases := []string{"Kafka", "Redis", "PostgreSQL", "database"}
	languages := []string{"Python", "Go", "JavaScript", "TypeScript", "Rust"}
	infrastructure := []string{"Docker", "Kubernetes"}
	protocols := []string{"REST", "GraphQL", "gRPC", "WebSocket"}

	for _, db := range databases {
		if strings.EqualFold(keyword, db) {
			return "database"
		}
	}

	for _, lang := range languages {
		if strings.EqualFold(keyword, lang) {
			return "programming_language"
		}
	}

	for _, infra := range infrastructure {
		if strings.EqualFold(keyword, infra) {
			return "infrastructure"
		}
	}

	for _, protocol := range protocols {
		if strings.EqualFold(keyword, protocol) {
			return "protocol"
		}
	}

	return "general"
}

// ExtractBatch extracts entities from multiple texts in parallel
func (ee *EntityExtractor) ExtractBatch(texts []string) [][]EntityData {
	results := make([][]EntityData, len(texts))
	var wg sync.WaitGroup

	for i, text := range texts {
		wg.Add(1)
		go func(index int, t string) {
			defer wg.Done()
			results[index] = ee.Extract(t)
		}(i, text)
	}

	wg.Wait()
	return results
}

// ClearCache clears the entity extraction cache
func (ee *EntityExtractor) ClearCache() {
	ee.mu.Lock()
	defer ee.mu.Unlock()
	ee.cache = make(map[string][]EntityData)
}

// GetCacheSize returns the current cache size
func (ee *EntityExtractor) GetCacheSize() int {
	ee.mu.RLock()
	defer ee.mu.RUnlock()
	return len(ee.cache)
}
