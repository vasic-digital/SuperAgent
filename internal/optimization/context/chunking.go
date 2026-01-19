// Package context provides context window management for LLM interactions.
package context

import (
	"strings"
	"unicode"
)

// DocumentChunker chunks documents for context injection.
type DocumentChunker struct {
	config *ChunkConfig
}

// ChunkConfig holds configuration for document chunking.
type ChunkConfig struct {
	// ChunkSize is the target chunk size in tokens.
	ChunkSize int `json:"chunk_size"`
	// ChunkOverlap is the overlap between chunks in tokens.
	ChunkOverlap int `json:"chunk_overlap"`
	// Separator is the separator for splitting (e.g., "\n\n").
	Separator string `json:"separator"`
	// SecondarySeperators are fallback separators.
	SecondarySeparators []string `json:"secondary_separators"`
	// TrimWhitespace trims whitespace from chunks.
	TrimWhitespace bool `json:"trim_whitespace"`
	// PreserveNewlines preserves newline structure.
	PreserveNewlines bool `json:"preserve_newlines"`
}

// DefaultChunkConfig returns a default chunk configuration.
func DefaultChunkConfig() *ChunkConfig {
	return &ChunkConfig{
		ChunkSize:           512,
		ChunkOverlap:        50,
		Separator:           "\n\n",
		SecondarySeparators: []string{"\n", ". ", " "},
		TrimWhitespace:      true,
		PreserveNewlines:    true,
	}
}

// NewDocumentChunker creates a new document chunker.
func NewDocumentChunker(config *ChunkConfig) *DocumentChunker {
	if config == nil {
		config = DefaultChunkConfig()
	}
	return &DocumentChunker{config: config}
}

// Chunk breaks a document into chunks.
func (c *DocumentChunker) Chunk(document string) []Chunk {
	if document == "" {
		return nil
	}

	// First, split by primary separator
	parts := strings.Split(document, c.config.Separator)

	var chunks []Chunk
	var currentChunk strings.Builder
	currentTokens := 0
	chunkIndex := 0

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		partTokens := estimateTokens(part)

		// If part alone exceeds chunk size, split further
		if partTokens > c.config.ChunkSize {
			// Flush current chunk if non-empty
			if currentChunk.Len() > 0 {
				chunks = append(chunks, c.createChunk(currentChunk.String(), chunkIndex))
				chunkIndex++
				currentChunk.Reset()
				currentTokens = 0
			}

			// Split large part using secondary separators
			subChunks := c.splitLargePart(part)
			for _, sub := range subChunks {
				chunks = append(chunks, c.createChunk(sub, chunkIndex))
				chunkIndex++
			}
			continue
		}

		// Check if adding this part would exceed chunk size
		if currentTokens+partTokens > c.config.ChunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, c.createChunk(currentChunk.String(), chunkIndex))
			chunkIndex++

			// Handle overlap
			if c.config.ChunkOverlap > 0 {
				overlap := c.getOverlap(currentChunk.String(), c.config.ChunkOverlap)
				currentChunk.Reset()
				currentChunk.WriteString(overlap)
				currentTokens = estimateTokens(overlap)
			} else {
				currentChunk.Reset()
				currentTokens = 0
			}
		}

		// Add separator between parts
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(c.config.Separator)
		}
		currentChunk.WriteString(part)
		currentTokens += partTokens
	}

	// Flush remaining content
	if currentChunk.Len() > 0 {
		chunks = append(chunks, c.createChunk(currentChunk.String(), chunkIndex))
	}

	return chunks
}

// ChunkWithMetadata chunks a document and includes metadata.
func (c *DocumentChunker) ChunkWithMetadata(document string, metadata map[string]interface{}) []Chunk {
	chunks := c.Chunk(document)
	for i := range chunks {
		if chunks[i].Metadata == nil {
			chunks[i].Metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			chunks[i].Metadata[k] = v
		}
	}
	return chunks
}

func (c *DocumentChunker) splitLargePart(part string) []string {
	var results []string

	// Try each secondary separator
	for _, sep := range c.config.SecondarySeparators {
		subparts := strings.Split(part, sep)
		if len(subparts) <= 1 {
			continue
		}

		var current strings.Builder
		currentTokens := 0

		for _, subpart := range subparts {
			subpart = strings.TrimSpace(subpart)
			if subpart == "" {
				continue
			}

			subTokens := estimateTokens(subpart)

			if currentTokens+subTokens > c.config.ChunkSize && current.Len() > 0 {
				results = append(results, current.String())
				current.Reset()
				currentTokens = 0
			}

			if current.Len() > 0 {
				current.WriteString(sep)
			}
			current.WriteString(subpart)
			currentTokens += subTokens
		}

		if current.Len() > 0 {
			results = append(results, current.String())
		}

		if len(results) > 0 {
			return results
		}
	}

	// Last resort: split by words
	return c.splitByWords(part)
}

func (c *DocumentChunker) splitByWords(text string) []string {
	words := strings.Fields(text)
	var results []string
	var current []string
	currentTokens := 0

	for _, word := range words {
		wordTokens := estimateTokens(word)

		if currentTokens+wordTokens > c.config.ChunkSize && len(current) > 0 {
			results = append(results, strings.Join(current, " "))
			current = nil
			currentTokens = 0
		}

		current = append(current, word)
		currentTokens += wordTokens
	}

	if len(current) > 0 {
		results = append(results, strings.Join(current, " "))
	}

	return results
}

func (c *DocumentChunker) getOverlap(text string, overlapTokens int) string {
	words := strings.Fields(text)
	if len(words) <= overlapTokens {
		return text
	}
	return strings.Join(words[len(words)-overlapTokens:], " ")
}

func (c *DocumentChunker) createChunk(content string, index int) Chunk {
	if c.config.TrimWhitespace {
		content = strings.TrimSpace(content)
	}
	return Chunk{
		Content:    content,
		Index:      index,
		TokenCount: estimateTokens(content),
	}
}

// Chunk represents a document chunk.
type Chunk struct {
	// Content is the chunk content.
	Content string `json:"content"`
	// Index is the chunk index.
	Index int `json:"index"`
	// TokenCount is the estimated token count.
	TokenCount int `json:"token_count"`
	// StartOffset is the start offset in the original document.
	StartOffset int `json:"start_offset,omitempty"`
	// EndOffset is the end offset in the original document.
	EndOffset int `json:"end_offset,omitempty"`
	// Metadata contains additional metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RecursiveCharacterTextSplitter implements recursive character-based text splitting.
type RecursiveCharacterTextSplitter struct {
	config *RecursiveSplitConfig
}

// RecursiveSplitConfig configures recursive text splitting.
type RecursiveSplitConfig struct {
	// ChunkSize is the target chunk size.
	ChunkSize int `json:"chunk_size"`
	// ChunkOverlap is the overlap between chunks.
	ChunkOverlap int `json:"chunk_overlap"`
	// Separators are the separators to try in order.
	Separators []string `json:"separators"`
	// LengthFunction calculates content length.
	LengthFunction func(string) int `json:"-"`
}

// DefaultRecursiveSplitConfig returns default config.
func DefaultRecursiveSplitConfig() *RecursiveSplitConfig {
	return &RecursiveSplitConfig{
		ChunkSize:    1000,
		ChunkOverlap: 200,
		Separators:   []string{"\n\n", "\n", " ", ""},
		LengthFunction: func(s string) int {
			return len(s)
		},
	}
}

// NewRecursiveCharacterTextSplitter creates a new recursive splitter.
func NewRecursiveCharacterTextSplitter(config *RecursiveSplitConfig) *RecursiveCharacterTextSplitter {
	if config == nil {
		config = DefaultRecursiveSplitConfig()
	}
	if config.LengthFunction == nil {
		config.LengthFunction = func(s string) int { return len(s) }
	}
	return &RecursiveCharacterTextSplitter{config: config}
}

// SplitText splits text into chunks.
func (s *RecursiveCharacterTextSplitter) SplitText(text string) []string {
	return s.splitText(text, s.config.Separators)
}

func (s *RecursiveCharacterTextSplitter) splitText(text string, separators []string) []string {
	var finalChunks []string

	separator := separators[len(separators)-1]
	newSeparators := separators

	for i, sep := range separators {
		if sep == "" || strings.Contains(text, sep) {
			separator = sep
			newSeparators = separators[i+1:]
			break
		}
	}

	var splits []string
	if separator != "" {
		splits = strings.Split(text, separator)
	} else {
		splits = []string{text}
	}

	var goodSplits []string
	mergeSeparator := separator
	if separator == "" {
		mergeSeparator = ""
	}

	for _, split := range splits {
		if s.config.LengthFunction(split) < s.config.ChunkSize {
			goodSplits = append(goodSplits, split)
		} else {
			if len(goodSplits) > 0 {
				mergedText := s.mergeSplits(goodSplits, mergeSeparator)
				finalChunks = append(finalChunks, mergedText...)
				goodSplits = nil
			}
			if len(newSeparators) > 0 {
				otherChunks := s.splitText(split, newSeparators)
				finalChunks = append(finalChunks, otherChunks...)
			} else {
				finalChunks = append(finalChunks, split)
			}
		}
	}

	if len(goodSplits) > 0 {
		mergedText := s.mergeSplits(goodSplits, mergeSeparator)
		finalChunks = append(finalChunks, mergedText...)
	}

	return finalChunks
}

func (s *RecursiveCharacterTextSplitter) mergeSplits(splits []string, separator string) []string {
	var docs []string
	var currentDoc []string
	total := 0

	for _, split := range splits {
		length := s.config.LengthFunction(split)

		if total+length+(len(currentDoc)*len(separator)) > s.config.ChunkSize {
			if len(currentDoc) > 0 {
				doc := strings.Join(currentDoc, separator)
				if doc != "" {
					docs = append(docs, doc)
				}
				// Keep overlap
				for total > s.config.ChunkOverlap ||
					(total+length+(len(currentDoc)*len(separator)) > s.config.ChunkSize && total > 0) {
					if len(currentDoc) == 0 {
						break
					}
					total -= s.config.LengthFunction(currentDoc[0]) + len(separator)
					currentDoc = currentDoc[1:]
				}
			}
		}
		currentDoc = append(currentDoc, split)
		total += length
	}

	doc := strings.Join(currentDoc, separator)
	if doc != "" {
		docs = append(docs, doc)
	}

	return docs
}

// SentenceSplitter splits text into sentences.
type SentenceSplitter struct {
	abbreviations map[string]bool
}

// NewSentenceSplitter creates a new sentence splitter.
func NewSentenceSplitter() *SentenceSplitter {
	return &SentenceSplitter{
		abbreviations: map[string]bool{
			"mr.": true, "mrs.": true, "ms.": true, "dr.": true,
			"prof.": true, "jr.": true, "sr.": true, "st.": true,
			"inc.": true, "ltd.": true, "corp.": true, "etc.": true,
			"e.g.": true, "i.e.": true, "vs.": true, "fig.": true,
		},
	}
}

// Split splits text into sentences.
func (s *SentenceSplitter) Split(text string) []string {
	var sentences []string
	var current strings.Builder
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		current.WriteRune(r)

		// Check for sentence ending
		if r == '.' || r == '!' || r == '?' {
			// Look ahead to see if this is end of sentence
			if s.isEndOfSentence(runes, i) {
				sentence := strings.TrimSpace(current.String())
				if sentence != "" {
					sentences = append(sentences, sentence)
				}
				current.Reset()
			}
		}
	}

	// Handle remaining text
	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		sentences = append(sentences, remaining)
	}

	return sentences
}

func (s *SentenceSplitter) isEndOfSentence(runes []rune, pos int) bool {
	// Check if followed by whitespace and uppercase or end of text
	if pos >= len(runes)-1 {
		return true
	}

	// Must be followed by whitespace
	if !unicode.IsSpace(runes[pos+1]) {
		return false
	}

	// Find next non-whitespace character
	for i := pos + 2; i < len(runes); i++ {
		if !unicode.IsSpace(runes[i]) {
			// Next word should start with uppercase for sentence boundary
			return unicode.IsUpper(runes[i])
		}
	}

	return true
}

// TokenTextSplitter splits text by token count.
type TokenTextSplitter struct {
	chunkSize    int
	chunkOverlap int
	tokenizer    Tokenizer
}

// Tokenizer interface for token counting.
type Tokenizer interface {
	// Tokenize returns tokens for the text.
	Tokenize(text string) []string
	// CountTokens returns the token count.
	CountTokens(text string) int
}

// SimpleTokenizer uses whitespace tokenization.
type SimpleTokenizer struct{}

// Tokenize tokenizes by whitespace.
func (t *SimpleTokenizer) Tokenize(text string) []string {
	return strings.Fields(text)
}

// CountTokens counts whitespace-separated tokens.
func (t *SimpleTokenizer) CountTokens(text string) int {
	return len(strings.Fields(text))
}

// NewTokenTextSplitter creates a new token-based splitter.
func NewTokenTextSplitter(chunkSize, chunkOverlap int, tokenizer Tokenizer) *TokenTextSplitter {
	if tokenizer == nil {
		tokenizer = &SimpleTokenizer{}
	}
	return &TokenTextSplitter{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		tokenizer:    tokenizer,
	}
}

// SplitText splits text into token-bounded chunks.
func (s *TokenTextSplitter) SplitText(text string) []string {
	tokens := s.tokenizer.Tokenize(text)
	if len(tokens) == 0 {
		return nil
	}

	var chunks []string
	start := 0

	for start < len(tokens) {
		end := start + s.chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}

		chunk := strings.Join(tokens[start:end], " ")
		chunks = append(chunks, chunk)

		if end >= len(tokens) {
			break
		}

		// Move start with overlap
		start = end - s.chunkOverlap
		if start < 0 {
			start = 0
		}
		if start >= end {
			start = end
		}
	}

	return chunks
}

// MarkdownHeaderTextSplitter splits Markdown by headers.
type MarkdownHeaderTextSplitter struct {
	headersToSplitOn []struct {
		header string
		name   string
	}
}

// NewMarkdownHeaderTextSplitter creates a Markdown splitter.
func NewMarkdownHeaderTextSplitter() *MarkdownHeaderTextSplitter {
	return &MarkdownHeaderTextSplitter{
		headersToSplitOn: []struct {
			header string
			name   string
		}{
			{"#", "Header 1"},
			{"##", "Header 2"},
			{"###", "Header 3"},
			{"####", "Header 4"},
		},
	}
}

// SplitText splits Markdown text by headers.
func (s *MarkdownHeaderTextSplitter) SplitText(text string) []MarkdownChunk {
	lines := strings.Split(text, "\n")
	var chunks []MarkdownChunk
	var currentChunk strings.Builder
	var currentHeaders map[string]string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this is a header
		isHeader := false
		for _, h := range s.headersToSplitOn {
			if strings.HasPrefix(trimmedLine, h.header+" ") {
				// Save current chunk if non-empty
				if currentChunk.Len() > 0 {
					chunks = append(chunks, MarkdownChunk{
						Content:  strings.TrimSpace(currentChunk.String()),
						Metadata: copyMap(currentHeaders),
					})
					currentChunk.Reset()
				}

				// Update headers
				if currentHeaders == nil {
					currentHeaders = make(map[string]string)
				}
				headerText := strings.TrimPrefix(trimmedLine, h.header+" ")
				currentHeaders[h.name] = headerText

				// Clear lower-level headers
				clearLowerHeaders(currentHeaders, h.header)

				isHeader = true
				break
			}
		}

		if !isHeader {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n")
			}
			currentChunk.WriteString(line)
		}
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, MarkdownChunk{
			Content:  strings.TrimSpace(currentChunk.String()),
			Metadata: copyMap(currentHeaders),
		})
	}

	return chunks
}

// MarkdownChunk represents a Markdown chunk with header metadata.
type MarkdownChunk struct {
	// Content is the chunk content.
	Content string `json:"content"`
	// Metadata contains header information.
	Metadata map[string]string `json:"metadata,omitempty"`
}

func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v
	}
	return result
}

func clearLowerHeaders(headers map[string]string, currentHeader string) {
	headerLevels := []string{"#", "##", "###", "####", "#####", "######"}
	headerNames := []string{"Header 1", "Header 2", "Header 3", "Header 4", "Header 5", "Header 6"}

	currentLevel := -1
	for i, h := range headerLevels {
		if h == currentHeader {
			currentLevel = i
			break
		}
	}

	if currentLevel >= 0 {
		for i := currentLevel + 1; i < len(headerNames); i++ {
			delete(headers, headerNames[i])
		}
	}
}
