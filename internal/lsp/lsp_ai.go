// Package lsp provides LSP-AI integration for AI-powered code intelligence.
// Implements context-aware completions, semantic analysis, and intelligent refactoring.
package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// LSPAIConfig configures the LSP-AI integration.
type LSPAIConfig struct {
	// MaxContextWindow is the maximum context window size in tokens
	MaxContextWindow int `json:"max_context_window"`
	// EnableSemanticCompletion enables AI-powered semantic completions
	EnableSemanticCompletion bool `json:"enable_semantic_completion"`
	// EnableCodeActions enables AI-powered code actions
	EnableCodeActions bool `json:"enable_code_actions"`
	// EnableDiagnostics enables AI-powered diagnostics
	EnableDiagnostics bool `json:"enable_diagnostics"`
	// EnableHover enables AI-powered hover information
	EnableHover bool `json:"enable_hover"`
	// EnableRefactoring enables AI-powered refactoring suggestions
	EnableRefactoring bool `json:"enable_refactoring"`
	// CompletionTimeout is the timeout for completion requests
	CompletionTimeout time.Duration `json:"completion_timeout"`
	// MaxCompletions is the maximum number of completions to return
	MaxCompletions int `json:"max_completions"`
	// MinConfidence is the minimum confidence threshold for suggestions
	MinConfidence float64 `json:"min_confidence"`
	// FIMEnabled enables Fill-in-the-Middle completion mode
	FIMEnabled bool `json:"fim_enabled"`
	// ContextChunks is the number of context chunks to use
	ContextChunks int `json:"context_chunks"`
}

// DefaultLSPAIConfig returns a default LSP-AI configuration.
func DefaultLSPAIConfig() LSPAIConfig {
	return LSPAIConfig{
		MaxContextWindow:         8192,
		EnableSemanticCompletion: true,
		EnableCodeActions:        true,
		EnableDiagnostics:        true,
		EnableHover:              true,
		EnableRefactoring:        true,
		CompletionTimeout:        5 * time.Second,
		MaxCompletions:           10,
		MinConfidence:            0.5,
		FIMEnabled:               true,
		ContextChunks:            5,
	}
}

// Position represents a position in a text document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextEdit represents a text edit operation.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// DocumentSymbol represents a symbol in a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Detail         string           `json:"detail,omitempty"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolKind represents the kind of a symbol.
type SymbolKind int

const (
	SymbolKindFile SymbolKind = iota + 1
	SymbolKindModule
	SymbolKindNamespace
	SymbolKindPackage
	SymbolKindClass
	SymbolKindMethod
	SymbolKindProperty
	SymbolKindField
	SymbolKindConstructor
	SymbolKindEnum
	SymbolKindInterface
	SymbolKindFunction
	SymbolKindVariable
	SymbolKindConstant
	SymbolKindString
	SymbolKindNumber
	SymbolKindBoolean
	SymbolKindArray
	SymbolKindObject
	SymbolKindKey
	SymbolKindNull
	SymbolKindEnumMember
	SymbolKindStruct
	SymbolKindEvent
	SymbolKindOperator
	SymbolKindTypeParameter
)

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label            string               `json:"label"`
	Kind             CompletionItemKind   `json:"kind"`
	Detail           string               `json:"detail,omitempty"`
	Documentation    string               `json:"documentation,omitempty"`
	SortText         string               `json:"sortText,omitempty"`
	FilterText       string               `json:"filterText,omitempty"`
	InsertText       string               `json:"insertText,omitempty"`
	InsertTextFormat InsertTextFormat     `json:"insertTextFormat,omitempty"`
	TextEdit         *TextEdit            `json:"textEdit,omitempty"`
	AdditionalEdits  []TextEdit           `json:"additionalTextEdits,omitempty"`
	Confidence       float64              `json:"confidence,omitempty"`
	Source           CompletionSource     `json:"source,omitempty"`
	Context          *CompletionContext   `json:"context,omitempty"`
}

// CompletionItemKind represents the kind of a completion item.
type CompletionItemKind int

const (
	CompletionItemKindText CompletionItemKind = iota + 1
	CompletionItemKindMethod
	CompletionItemKindFunction
	CompletionItemKindConstructor
	CompletionItemKindField
	CompletionItemKindVariable
	CompletionItemKindClass
	CompletionItemKindInterface
	CompletionItemKindModule
	CompletionItemKindProperty
	CompletionItemKindUnit
	CompletionItemKindValue
	CompletionItemKindEnum
	CompletionItemKindKeyword
	CompletionItemKindSnippet
	CompletionItemKindColor
	CompletionItemKindFile
	CompletionItemKindReference
	CompletionItemKindFolder
	CompletionItemKindEnumMember
	CompletionItemKindConstant
	CompletionItemKindStruct
	CompletionItemKindEvent
	CompletionItemKindOperator
	CompletionItemKindTypeParameter
)

// InsertTextFormat specifies the format of the insert text.
type InsertTextFormat int

const (
	InsertTextFormatPlainText InsertTextFormat = 1
	InsertTextFormatSnippet   InsertTextFormat = 2
)

// CompletionSource indicates the source of a completion.
type CompletionSource string

const (
	CompletionSourceLSP        CompletionSource = "lsp"
	CompletionSourceAI         CompletionSource = "ai"
	CompletionSourceFIM        CompletionSource = "fim"
	CompletionSourceSemantic   CompletionSource = "semantic"
	CompletionSourceSnippet    CompletionSource = "snippet"
	CompletionSourceHistory    CompletionSource = "history"
)

// CompletionContext provides context for a completion request.
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
	Prefix           string                `json:"prefix,omitempty"`
	Suffix           string                `json:"suffix,omitempty"`
	CurrentLine      string                `json:"currentLine,omitempty"`
	SurroundingCode  string                `json:"surroundingCode,omitempty"`
	Symbols          []DocumentSymbol      `json:"symbols,omitempty"`
}

// CompletionTriggerKind indicates the trigger for a completion request.
type CompletionTriggerKind int

const (
	CompletionTriggerInvoked          CompletionTriggerKind = 1
	CompletionTriggerCharacter        CompletionTriggerKind = 2
	CompletionTriggerIncomplete       CompletionTriggerKind = 3
)

// Diagnostic represents a diagnostic message.
type Diagnostic struct {
	Range           Range            `json:"range"`
	Severity        DiagnosticSeverity `json:"severity"`
	Code            string           `json:"code,omitempty"`
	Source          string           `json:"source,omitempty"`
	Message         string           `json:"message"`
	RelatedInfo     []DiagnosticRelatedInfo `json:"relatedInformation,omitempty"`
	Tags            []DiagnosticTag  `json:"tags,omitempty"`
	AIExplanation   string           `json:"aiExplanation,omitempty"`
	SuggestedFix    *TextEdit        `json:"suggestedFix,omitempty"`
}

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// DiagnosticTag represents additional metadata about a diagnostic.
type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// DiagnosticRelatedInfo represents related information for a diagnostic.
type DiagnosticRelatedInfo struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// CodeAction represents a code action (quick fix, refactoring, etc.).
type CodeAction struct {
	Title       string            `json:"title"`
	Kind        CodeActionKind    `json:"kind"`
	Diagnostics []Diagnostic      `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit    `json:"edit,omitempty"`
	Command     *Command          `json:"command,omitempty"`
	IsPreferred bool              `json:"isPreferred,omitempty"`
	AIGenerated bool              `json:"aiGenerated,omitempty"`
	Confidence  float64           `json:"confidence,omitempty"`
}

// CodeActionKind represents the kind of a code action.
type CodeActionKind string

const (
	CodeActionKindQuickFix              CodeActionKind = "quickfix"
	CodeActionKindRefactor              CodeActionKind = "refactor"
	CodeActionKindRefactorExtract       CodeActionKind = "refactor.extract"
	CodeActionKindRefactorInline        CodeActionKind = "refactor.inline"
	CodeActionKindRefactorRewrite       CodeActionKind = "refactor.rewrite"
	CodeActionKindSource                CodeActionKind = "source"
	CodeActionKindSourceOrganizeImports CodeActionKind = "source.organizeImports"
	CodeActionKindSourceFixAll          CodeActionKind = "source.fixAll"
)

// WorkspaceEdit represents a workspace edit.
type WorkspaceEdit struct {
	Changes         map[string][]TextEdit `json:"changes,omitempty"`
	DocumentChanges []TextDocumentEdit    `json:"documentChanges,omitempty"`
}

// TextDocumentEdit represents an edit to a text document.
type TextDocumentEdit struct {
	TextDocument VersionedTextDocumentIdentifier `json:"textDocument"`
	Edits        []TextEdit                      `json:"edits"`
}

// VersionedTextDocumentIdentifier identifies a versioned text document.
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// Command represents a command.
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// Hover represents hover information.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents markup content.
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// MarkupKind represents the kind of markup.
type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

// AICompletionProvider provides AI-powered completions.
type AICompletionProvider interface {
	// GenerateCompletion generates completions for the given context
	GenerateCompletion(ctx context.Context, request *AICompletionRequest) ([]CompletionItem, error)
	// ExplainCode provides an explanation for code at a given position
	ExplainCode(ctx context.Context, uri string, code string, position Position) (string, error)
	// SuggestRefactoring suggests refactoring options for the given code
	SuggestRefactoring(ctx context.Context, uri string, code string, selection Range) ([]CodeAction, error)
	// AnalyzeDiagnostic provides AI analysis of a diagnostic
	AnalyzeDiagnostic(ctx context.Context, diagnostic Diagnostic, code string) (*DiagnosticAnalysis, error)
}

// AICompletionRequest represents a request for AI completions.
type AICompletionRequest struct {
	URI         string            `json:"uri"`
	Position    Position          `json:"position"`
	Prefix      string            `json:"prefix"`
	Suffix      string            `json:"suffix"`
	Language    string            `json:"language"`
	Context     *CompletionContext `json:"context,omitempty"`
	MaxTokens   int               `json:"maxTokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
}

// DiagnosticAnalysis represents AI analysis of a diagnostic.
type DiagnosticAnalysis struct {
	Explanation   string       `json:"explanation"`
	RootCause     string       `json:"rootCause,omitempty"`
	SuggestedFix  *TextEdit    `json:"suggestedFix,omitempty"`
	RelatedDocs   []string     `json:"relatedDocs,omitempty"`
	Confidence    float64      `json:"confidence"`
}

// LSPAI integrates AI capabilities with LSP.
type LSPAI struct {
	config              LSPAIConfig
	aiProvider          AICompletionProvider
	documentStore       *DocumentStore
	symbolIndex         *SymbolIndex
	completionCache     *CompletionCache
	diagnosticsCache    map[string][]Diagnostic
	mu                  sync.RWMutex
}

// NewLSPAI creates a new LSP-AI integration.
func NewLSPAI(config LSPAIConfig, aiProvider AICompletionProvider) *LSPAI {
	return &LSPAI{
		config:           config,
		aiProvider:       aiProvider,
		documentStore:    NewDocumentStore(),
		symbolIndex:      NewSymbolIndex(),
		completionCache:  NewCompletionCache(1000, 5*time.Minute),
		diagnosticsCache: make(map[string][]Diagnostic),
	}
}

// Initialize initializes the LSP-AI integration.
func (l *LSPAI) Initialize(ctx context.Context) error {
	// Initialize symbol index
	if err := l.symbolIndex.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize symbol index: %w", err)
	}

	return nil
}

// OpenDocument handles document open events.
func (l *LSPAI) OpenDocument(ctx context.Context, uri string, content string, languageID string, version int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	doc := &Document{
		URI:        uri,
		Content:    content,
		LanguageID: languageID,
		Version:    version,
	}

	l.documentStore.Open(doc)

	// Index document symbols
	symbols, err := l.extractSymbols(content, languageID)
	if err == nil {
		l.symbolIndex.UpdateDocument(uri, symbols)
	}

	return nil
}

// CloseDocument handles document close events.
func (l *LSPAI) CloseDocument(uri string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.documentStore.Close(uri)
	l.symbolIndex.RemoveDocument(uri)
	delete(l.diagnosticsCache, uri)
}

// UpdateDocument handles document change events.
func (l *LSPAI) UpdateDocument(ctx context.Context, uri string, changes []TextDocumentContentChangeEvent, version int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	doc := l.documentStore.Get(uri)
	if doc == nil {
		return fmt.Errorf("document not found: %s", uri)
	}

	// Apply changes
	for _, change := range changes {
		if change.Range != nil {
			doc.Content = applyTextEdit(doc.Content, TextEdit{
				Range:   *change.Range,
				NewText: change.Text,
			})
		} else {
			doc.Content = change.Text
		}
	}
	doc.Version = version

	// Re-index symbols
	symbols, err := l.extractSymbols(doc.Content, doc.LanguageID)
	if err == nil {
		l.symbolIndex.UpdateDocument(uri, symbols)
	}

	// Invalidate completion cache for this document
	l.completionCache.InvalidateDocument(uri)

	return nil
}

// TextDocumentContentChangeEvent represents a change to a text document.
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// GetCompletion returns completion items for the given position.
func (l *LSPAI) GetCompletion(ctx context.Context, uri string, position Position, completionCtx *CompletionContext) ([]CompletionItem, error) {
	l.mu.RLock()
	doc := l.documentStore.Get(uri)
	l.mu.RUnlock()

	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Check cache
	cacheKey := l.completionCacheKey(uri, position)
	if cached := l.completionCache.Get(cacheKey); cached != nil {
		return cached, nil
	}

	// Extract context
	prefix, suffix := l.extractPrefixSuffix(doc.Content, position)
	currentLine := l.extractCurrentLine(doc.Content, position)

	if completionCtx == nil {
		completionCtx = &CompletionContext{}
	}
	completionCtx.Prefix = prefix
	completionCtx.Suffix = suffix
	completionCtx.CurrentLine = currentLine
	completionCtx.SurroundingCode = l.extractSurroundingCode(doc.Content, position, l.config.ContextChunks)

	// Get symbols in scope
	completionCtx.Symbols = l.symbolIndex.GetSymbolsInScope(uri, position)

	var completions []CompletionItem

	// Get AI completions if enabled
	if l.config.EnableSemanticCompletion && l.aiProvider != nil {
		aiCtx, cancel := context.WithTimeout(ctx, l.config.CompletionTimeout)
		defer cancel()

		aiRequest := &AICompletionRequest{
			URI:      uri,
			Position: position,
			Prefix:   prefix,
			Suffix:   suffix,
			Language: doc.LanguageID,
			Context:  completionCtx,
		}

		aiCompletions, err := l.aiProvider.GenerateCompletion(aiCtx, aiRequest)
		if err == nil {
			for i := range aiCompletions {
				aiCompletions[i].Source = CompletionSourceAI
			}
			completions = append(completions, aiCompletions...)
		}
	}

	// Get FIM completions if enabled
	if l.config.FIMEnabled && l.aiProvider != nil {
		fimCompletions := l.getFIMCompletions(ctx, doc, position, prefix, suffix)
		completions = append(completions, fimCompletions...)
	}

	// Add symbol completions
	symbolCompletions := l.getSymbolCompletions(uri, position, completionCtx)
	completions = append(completions, symbolCompletions...)

	// Filter by confidence
	var filtered []CompletionItem
	for _, c := range completions {
		if c.Confidence >= l.config.MinConfidence {
			filtered = append(filtered, c)
		}
	}

	// Sort and limit
	filtered = l.sortCompletions(filtered)
	if len(filtered) > l.config.MaxCompletions {
		filtered = filtered[:l.config.MaxCompletions]
	}

	// Cache results
	l.completionCache.Set(cacheKey, filtered)

	return filtered, nil
}

// GetHover returns hover information for the given position.
func (l *LSPAI) GetHover(ctx context.Context, uri string, position Position) (*Hover, error) {
	if !l.config.EnableHover {
		return nil, nil
	}

	l.mu.RLock()
	doc := l.documentStore.Get(uri)
	l.mu.RUnlock()

	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Get word at position
	word, wordRange := l.getWordAtPosition(doc.Content, position)
	if word == "" {
		return nil, nil
	}

	// Check symbol index first
	symbol := l.symbolIndex.GetSymbolAtPosition(uri, position)
	if symbol != nil {
		return &Hover{
			Contents: MarkupContent{
				Kind:  MarkupKindMarkdown,
				Value: l.formatSymbolHover(symbol),
			},
			Range: &wordRange,
		}, nil
	}

	// Use AI for explanation if available
	if l.aiProvider != nil {
		explanation, err := l.aiProvider.ExplainCode(ctx, uri, doc.Content, position)
		if err == nil && explanation != "" {
			return &Hover{
				Contents: MarkupContent{
					Kind:  MarkupKindMarkdown,
					Value: explanation,
				},
				Range: &wordRange,
			}, nil
		}
	}

	return nil, nil
}

// GetCodeActions returns code actions for the given range.
func (l *LSPAI) GetCodeActions(ctx context.Context, uri string, rangeVal Range, diagnostics []Diagnostic) ([]CodeAction, error) {
	if !l.config.EnableCodeActions {
		return nil, nil
	}

	l.mu.RLock()
	doc := l.documentStore.Get(uri)
	l.mu.RUnlock()

	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	var actions []CodeAction

	// Get AI-suggested quick fixes for diagnostics
	for _, diag := range diagnostics {
		if l.aiProvider != nil {
			analysis, err := l.aiProvider.AnalyzeDiagnostic(ctx, diag, doc.Content)
			if err == nil && analysis.SuggestedFix != nil {
				actions = append(actions, CodeAction{
					Title:       fmt.Sprintf("AI Fix: %s", diag.Message),
					Kind:        CodeActionKindQuickFix,
					Diagnostics: []Diagnostic{diag},
					Edit: &WorkspaceEdit{
						Changes: map[string][]TextEdit{
							uri: {*analysis.SuggestedFix},
						},
					},
					IsPreferred: true,
					AIGenerated: true,
					Confidence:  analysis.Confidence,
				})
			}
		}
	}

	// Get AI-suggested refactorings if enabled
	if l.config.EnableRefactoring && l.aiProvider != nil {
		refactorings, err := l.aiProvider.SuggestRefactoring(ctx, uri, doc.Content, rangeVal)
		if err == nil {
			actions = append(actions, refactorings...)
		}
	}

	return actions, nil
}

// GetDiagnostics returns AI-enhanced diagnostics for the document.
func (l *LSPAI) GetDiagnostics(ctx context.Context, uri string) ([]Diagnostic, error) {
	if !l.config.EnableDiagnostics {
		return nil, nil
	}

	l.mu.RLock()
	doc := l.documentStore.Get(uri)
	cached := l.diagnosticsCache[uri]
	l.mu.RUnlock()

	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Return cached diagnostics if available
	if cached != nil {
		return cached, nil
	}

	var diagnostics []Diagnostic

	// Analyze code with AI
	if l.aiProvider != nil {
		// Run AI analysis on the document
		// This would typically involve checking for common issues, security vulnerabilities, etc.
		analysis, err := l.analyzeDocument(ctx, doc)
		if err == nil {
			diagnostics = append(diagnostics, analysis...)
		}
	}

	// Cache diagnostics
	l.mu.Lock()
	l.diagnosticsCache[uri] = diagnostics
	l.mu.Unlock()

	return diagnostics, nil
}

// GetDefinition returns the definition location for the symbol at the given position.
func (l *LSPAI) GetDefinition(ctx context.Context, uri string, position Position) ([]Location, error) {
	symbol := l.symbolIndex.GetSymbolAtPosition(uri, position)
	if symbol == nil {
		return nil, nil
	}

	// Look up definition in symbol index
	definitions := l.symbolIndex.FindDefinitions(symbol.Name, symbol.Kind)

	var locations []Location
	for _, def := range definitions {
		locations = append(locations, Location{
			URI:   def.URI,
			Range: def.Range,
		})
	}

	return locations, nil
}

// GetReferences returns all references to the symbol at the given position.
func (l *LSPAI) GetReferences(ctx context.Context, uri string, position Position, includeDeclaration bool) ([]Location, error) {
	symbol := l.symbolIndex.GetSymbolAtPosition(uri, position)
	if symbol == nil {
		return nil, nil
	}

	references := l.symbolIndex.FindReferences(symbol.Name)

	var locations []Location
	for _, ref := range references {
		if !includeDeclaration && ref.IsDeclaration {
			continue
		}
		locations = append(locations, Location{
			URI:   ref.URI,
			Range: ref.Range,
		})
	}

	return locations, nil
}

// FormatDocument formats the entire document.
func (l *LSPAI) FormatDocument(ctx context.Context, uri string) ([]TextEdit, error) {
	l.mu.RLock()
	doc := l.documentStore.Get(uri)
	l.mu.RUnlock()

	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", uri)
	}

	// Use language-specific formatter
	formatted, err := l.formatCode(doc.Content, doc.LanguageID)
	if err != nil {
		return nil, err
	}

	if formatted == doc.Content {
		return nil, nil
	}

	// Calculate diff and return as text edits
	return l.computeTextEdits(doc.Content, formatted), nil
}

// Helper methods

func (l *LSPAI) completionCacheKey(uri string, position Position) string {
	return fmt.Sprintf("%s:%d:%d", uri, position.Line, position.Character)
}

func (l *LSPAI) extractPrefixSuffix(content string, position Position) (string, string) {
	lines := strings.Split(content, "\n")
	if position.Line >= len(lines) {
		return content, ""
	}

	var prefixBuilder strings.Builder
	for i := 0; i < position.Line; i++ {
		prefixBuilder.WriteString(lines[i])
		prefixBuilder.WriteString("\n")
	}

	if position.Character <= len(lines[position.Line]) {
		prefixBuilder.WriteString(lines[position.Line][:position.Character])
	}

	prefix := prefixBuilder.String()

	var suffixBuilder strings.Builder
	if position.Character < len(lines[position.Line]) {
		suffixBuilder.WriteString(lines[position.Line][position.Character:])
	}
	suffixBuilder.WriteString("\n")

	for i := position.Line + 1; i < len(lines); i++ {
		suffixBuilder.WriteString(lines[i])
		if i < len(lines)-1 {
			suffixBuilder.WriteString("\n")
		}
	}

	return prefix, suffixBuilder.String()
}

func (l *LSPAI) extractCurrentLine(content string, position Position) string {
	lines := strings.Split(content, "\n")
	if position.Line >= len(lines) {
		return ""
	}
	return lines[position.Line]
}

func (l *LSPAI) extractSurroundingCode(content string, position Position, chunks int) string {
	lines := strings.Split(content, "\n")
	startLine := max(0, position.Line-chunks*10)
	endLine := min(len(lines), position.Line+chunks*10)

	return strings.Join(lines[startLine:endLine], "\n")
}

func (l *LSPAI) getWordAtPosition(content string, position Position) (string, Range) {
	lines := strings.Split(content, "\n")
	if position.Line >= len(lines) {
		return "", Range{}
	}

	line := lines[position.Line]
	if position.Character >= len(line) {
		return "", Range{}
	}

	// Find word boundaries
	start := position.Character
	end := position.Character

	for start > 0 && isWordChar(line[start-1]) {
		start--
	}

	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return "", Range{}
	}

	return line[start:end], Range{
		Start: Position{Line: position.Line, Character: start},
		End:   Position{Line: position.Line, Character: end},
	}
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func (l *LSPAI) formatSymbolHover(symbol *DocumentSymbol) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s** (%s)\n\n", symbol.Name, symbolKindString(symbol.Kind)))
	if symbol.Detail != "" {
		sb.WriteString(symbol.Detail)
	}
	return sb.String()
}

func symbolKindString(kind SymbolKind) string {
	switch kind {
	case SymbolKindFile:
		return "file"
	case SymbolKindModule:
		return "module"
	case SymbolKindNamespace:
		return "namespace"
	case SymbolKindPackage:
		return "package"
	case SymbolKindClass:
		return "class"
	case SymbolKindMethod:
		return "method"
	case SymbolKindProperty:
		return "property"
	case SymbolKindField:
		return "field"
	case SymbolKindConstructor:
		return "constructor"
	case SymbolKindEnum:
		return "enum"
	case SymbolKindInterface:
		return "interface"
	case SymbolKindFunction:
		return "function"
	case SymbolKindVariable:
		return "variable"
	case SymbolKindConstant:
		return "constant"
	case SymbolKindStruct:
		return "struct"
	default:
		return "symbol"
	}
}

func (l *LSPAI) extractSymbols(content string, languageID string) ([]DocumentSymbol, error) {
	// Simple symbol extraction - in production would use tree-sitter or language server
	var symbols []DocumentSymbol
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		// Look for function definitions
		if idx := strings.Index(line, "func "); idx >= 0 {
			name := extractFunctionName(line[idx+5:])
			if name != "" {
				symbols = append(symbols, DocumentSymbol{
					Name: name,
					Kind: SymbolKindFunction,
					Range: Range{
						Start: Position{Line: lineNum, Character: idx},
						End:   Position{Line: lineNum, Character: len(line)},
					},
					SelectionRange: Range{
						Start: Position{Line: lineNum, Character: idx + 5},
						End:   Position{Line: lineNum, Character: idx + 5 + len(name)},
					},
				})
			}
		}

		// Look for type definitions
		if idx := strings.Index(line, "type "); idx >= 0 {
			name := extractTypeName(line[idx+5:])
			if name != "" {
				kind := SymbolKindStruct
				if strings.Contains(line, "interface") {
					kind = SymbolKindInterface
				}
				symbols = append(symbols, DocumentSymbol{
					Name: name,
					Kind: kind,
					Range: Range{
						Start: Position{Line: lineNum, Character: idx},
						End:   Position{Line: lineNum, Character: len(line)},
					},
					SelectionRange: Range{
						Start: Position{Line: lineNum, Character: idx + 5},
						End:   Position{Line: lineNum, Character: idx + 5 + len(name)},
					},
				})
			}
		}
	}

	return symbols, nil
}

func extractFunctionName(s string) string {
	// Skip receiver
	if s[0] == '(' {
		idx := strings.Index(s, ")")
		if idx < 0 {
			return ""
		}
		s = strings.TrimSpace(s[idx+1:])
	}

	// Extract name
	end := strings.IndexAny(s, "([")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(s[:end])
}

func extractTypeName(s string) string {
	end := strings.IndexAny(s, " \t")
	if end < 0 {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(s[:end])
}

func (l *LSPAI) getFIMCompletions(ctx context.Context, doc *Document, position Position, prefix, suffix string) []CompletionItem {
	// Fill-in-the-Middle completion
	// This would call the AI provider with FIM format
	return nil
}

func (l *LSPAI) getSymbolCompletions(uri string, position Position, context *CompletionContext) []CompletionItem {
	var completions []CompletionItem

	// Get symbols in scope
	symbols := l.symbolIndex.GetSymbolsInScope(uri, position)
	for _, symbol := range symbols {
		completions = append(completions, CompletionItem{
			Label:      symbol.Name,
			Kind:       symbolKindToCompletionKind(symbol.Kind),
			Detail:     symbol.Detail,
			Source:     CompletionSourceLSP,
			Confidence: 0.8,
		})
	}

	return completions
}

func symbolKindToCompletionKind(kind SymbolKind) CompletionItemKind {
	switch kind {
	case SymbolKindClass:
		return CompletionItemKindClass
	case SymbolKindInterface:
		return CompletionItemKindInterface
	case SymbolKindFunction:
		return CompletionItemKindFunction
	case SymbolKindMethod:
		return CompletionItemKindMethod
	case SymbolKindVariable:
		return CompletionItemKindVariable
	case SymbolKindConstant:
		return CompletionItemKindConstant
	case SymbolKindField:
		return CompletionItemKindField
	case SymbolKindProperty:
		return CompletionItemKindProperty
	case SymbolKindStruct:
		return CompletionItemKindStruct
	default:
		return CompletionItemKindText
	}
}

func (l *LSPAI) sortCompletions(completions []CompletionItem) []CompletionItem {
	// Sort by confidence descending, then by source priority
	// AI completions get priority, then semantic, then LSP
	// In a real implementation, would use a proper sorting algorithm
	return completions
}

func (l *LSPAI) analyzeDocument(ctx context.Context, doc *Document) ([]Diagnostic, error) {
	// AI-based code analysis
	// Would call AI provider to analyze code for issues
	return nil, nil
}

func (l *LSPAI) formatCode(content string, languageID string) (string, error) {
	// Would call language-specific formatter
	return content, nil
}

func (l *LSPAI) computeTextEdits(original, formatted string) []TextEdit {
	// Would compute minimal text edits to transform original to formatted
	if original == formatted {
		return nil
	}

	// Simple implementation: replace all
	lines := strings.Split(original, "\n")
	return []TextEdit{{
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: len(lines) - 1, Character: len(lines[len(lines)-1])},
		},
		NewText: formatted,
	}}
}

func applyTextEdit(content string, edit TextEdit) string {
	lines := strings.Split(content, "\n")

	// Get text before and after the edit range
	var before strings.Builder
	for i := 0; i < edit.Range.Start.Line; i++ {
		before.WriteString(lines[i])
		before.WriteString("\n")
	}
	if edit.Range.Start.Line < len(lines) {
		before.WriteString(lines[edit.Range.Start.Line][:min(edit.Range.Start.Character, len(lines[edit.Range.Start.Line]))])
	}

	var after strings.Builder
	if edit.Range.End.Line < len(lines) {
		if edit.Range.End.Character < len(lines[edit.Range.End.Line]) {
			after.WriteString(lines[edit.Range.End.Line][edit.Range.End.Character:])
		}
		after.WriteString("\n")
		for i := edit.Range.End.Line + 1; i < len(lines); i++ {
			after.WriteString(lines[i])
			if i < len(lines)-1 {
				after.WriteString("\n")
			}
		}
	}

	return before.String() + edit.NewText + after.String()
}

// Document represents an open document.
type Document struct {
	URI        string
	Content    string
	LanguageID string
	Version    int
}

// DocumentStore stores open documents.
type DocumentStore struct {
	documents map[string]*Document
	mu        sync.RWMutex
}

// NewDocumentStore creates a new document store.
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: make(map[string]*Document),
	}
}

// Open adds a document to the store.
func (s *DocumentStore) Open(doc *Document) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.documents[doc.URI] = doc
}

// Close removes a document from the store.
func (s *DocumentStore) Close(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.documents, uri)
}

// Get retrieves a document from the store.
func (s *DocumentStore) Get(uri string) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.documents[uri]
}

// SymbolIndex indexes document symbols.
type SymbolIndex struct {
	symbolsByURI map[string][]DocumentSymbol
	symbolsByName map[string][]SymbolEntry
	mu           sync.RWMutex
}

// SymbolEntry represents a symbol entry in the index.
type SymbolEntry struct {
	URI           string
	Symbol        DocumentSymbol
	Range         Range
	IsDeclaration bool
}

// NewSymbolIndex creates a new symbol index.
func NewSymbolIndex() *SymbolIndex {
	return &SymbolIndex{
		symbolsByURI:  make(map[string][]DocumentSymbol),
		symbolsByName: make(map[string][]SymbolEntry),
	}
}

// Initialize initializes the symbol index.
func (s *SymbolIndex) Initialize(ctx context.Context) error {
	return nil
}

// UpdateDocument updates the symbols for a document.
func (s *SymbolIndex) UpdateDocument(uri string, symbols []DocumentSymbol) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove old entries
	s.removeDocumentSymbols(uri)

	// Add new entries
	s.symbolsByURI[uri] = symbols
	for _, sym := range symbols {
		s.symbolsByName[sym.Name] = append(s.symbolsByName[sym.Name], SymbolEntry{
			URI:           uri,
			Symbol:        sym,
			Range:         sym.Range,
			IsDeclaration: true,
		})
	}
}

// RemoveDocument removes symbols for a document.
func (s *SymbolIndex) RemoveDocument(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeDocumentSymbols(uri)
}

func (s *SymbolIndex) removeDocumentSymbols(uri string) {
	delete(s.symbolsByURI, uri)

	// Clean up symbol name index
	for name, entries := range s.symbolsByName {
		var filtered []SymbolEntry
		for _, entry := range entries {
			if entry.URI != uri {
				filtered = append(filtered, entry)
			}
		}
		if len(filtered) > 0 {
			s.symbolsByName[name] = filtered
		} else {
			delete(s.symbolsByName, name)
		}
	}
}

// GetSymbolAtPosition returns the symbol at the given position.
func (s *SymbolIndex) GetSymbolAtPosition(uri string, position Position) *DocumentSymbol {
	s.mu.RLock()
	defer s.mu.RUnlock()

	symbols := s.symbolsByURI[uri]
	for i := range symbols {
		if positionInRange(position, symbols[i].SelectionRange) {
			return &symbols[i]
		}
	}
	return nil
}

// GetSymbolsInScope returns symbols visible at the given position.
func (s *SymbolIndex) GetSymbolsInScope(uri string, position Position) []DocumentSymbol {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return all symbols in the document and imported symbols
	// In a real implementation, would compute actual scope
	var result []DocumentSymbol
	result = append(result, s.symbolsByURI[uri]...)
	return result
}

// FindDefinitions finds definitions for a symbol name.
func (s *SymbolIndex) FindDefinitions(name string, kind SymbolKind) []SymbolEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []SymbolEntry
	for _, entry := range s.symbolsByName[name] {
		if entry.IsDeclaration && entry.Symbol.Kind == kind {
			result = append(result, entry)
		}
	}
	return result
}

// FindReferences finds all references to a symbol name.
func (s *SymbolIndex) FindReferences(name string) []SymbolEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.symbolsByName[name]
}

func positionInRange(pos Position, r Range) bool {
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}
	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}
	return true
}

// CompletionCache caches completion results.
type CompletionCache struct {
	items    map[string]cacheEntry
	maxSize  int
	ttl      time.Duration
	mu       sync.RWMutex
}

type cacheEntry struct {
	items     []CompletionItem
	timestamp time.Time
}

// NewCompletionCache creates a new completion cache.
func NewCompletionCache(maxSize int, ttl time.Duration) *CompletionCache {
	return &CompletionCache{
		items:   make(map[string]cacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get retrieves items from the cache.
func (c *CompletionCache) Get(key string) []CompletionItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[key]
	if !ok {
		return nil
	}

	if time.Since(entry.timestamp) > c.ttl {
		return nil
	}

	return entry.items
}

// Set stores items in the cache.
func (c *CompletionCache) Set(key string, items []CompletionItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if over capacity
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	c.items[key] = cacheEntry{
		items:     items,
		timestamp: time.Now(),
	}
}

// InvalidateDocument invalidates all cache entries for a document.
func (c *CompletionCache) InvalidateDocument(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if strings.HasPrefix(key, uri+":") {
			delete(c.items, key)
		}
	}
}

func (c *CompletionCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.items {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// LSPAIResult represents the result of an LSP-AI operation.
type LSPAIResult struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
}

// SerializeToJSON serializes the result to JSON.
func (r *LSPAIResult) SerializeToJSON() ([]byte, error) {
	return json.Marshal(r)
}
