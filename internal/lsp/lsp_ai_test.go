// Package lsp provides tests for the LSP-AI integration.
package lsp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockAICompletionProvider implements AICompletionProvider for testing
type MockAICompletionProvider struct {
	generateFunc        func(ctx context.Context, request *AICompletionRequest) ([]CompletionItem, error)
	explainFunc         func(ctx context.Context, uri string, code string, position Position) (string, error)
	suggestRefactorFunc func(ctx context.Context, uri string, code string, selection Range) ([]CodeAction, error)
	analyzeDiagFunc     func(ctx context.Context, diagnostic Diagnostic, code string) (*DiagnosticAnalysis, error)
}

func (m *MockAICompletionProvider) GenerateCompletion(ctx context.Context, request *AICompletionRequest) ([]CompletionItem, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, request)
	}
	return []CompletionItem{
		{
			Label:      "testCompletion",
			Kind:       CompletionItemKindFunction,
			Detail:     "Test completion",
			Confidence: 0.9,
		},
	}, nil
}

func (m *MockAICompletionProvider) ExplainCode(ctx context.Context, uri string, code string, position Position) (string, error) {
	if m.explainFunc != nil {
		return m.explainFunc(ctx, uri, code, position)
	}
	return "This is a test explanation of the code.", nil
}

func (m *MockAICompletionProvider) SuggestRefactoring(ctx context.Context, uri string, code string, selection Range) ([]CodeAction, error) {
	if m.suggestRefactorFunc != nil {
		return m.suggestRefactorFunc(ctx, uri, code, selection)
	}
	return []CodeAction{
		{
			Title:       "Extract method",
			Kind:        CodeActionKindRefactorExtract,
			AIGenerated: true,
			Confidence:  0.85,
		},
	}, nil
}

func (m *MockAICompletionProvider) AnalyzeDiagnostic(ctx context.Context, diagnostic Diagnostic, code string) (*DiagnosticAnalysis, error) {
	if m.analyzeDiagFunc != nil {
		return m.analyzeDiagFunc(ctx, diagnostic, code)
	}
	return &DiagnosticAnalysis{
		Explanation: "This diagnostic indicates a potential issue.",
		RootCause:   "The root cause is...",
		Confidence:  0.8,
	}, nil
}

// TestDefaultLSPAIConfig tests default configuration
func TestDefaultLSPAIConfig(t *testing.T) {
	config := DefaultLSPAIConfig()

	assert.Equal(t, 8192, config.MaxContextWindow)
	assert.True(t, config.EnableSemanticCompletion)
	assert.True(t, config.EnableCodeActions)
	assert.True(t, config.EnableDiagnostics)
	assert.True(t, config.EnableHover)
	assert.True(t, config.EnableRefactoring)
	assert.Equal(t, 5*time.Second, config.CompletionTimeout)
	assert.Equal(t, 10, config.MaxCompletions)
	assert.Equal(t, 0.5, config.MinConfidence)
	assert.True(t, config.FIMEnabled)
	assert.Equal(t, 5, config.ContextChunks)
}

// TestNewLSPAI tests LSPAI creation
func TestNewLSPAI(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}

	lspai := NewLSPAI(config, provider)

	assert.NotNil(t, lspai)
	assert.NotNil(t, lspai.documentStore)
	assert.NotNil(t, lspai.symbolIndex)
	assert.NotNil(t, lspai.completionCache)
}

// TestLSPAI_Initialize tests initialization
func TestLSPAI_Initialize(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	err := lspai.Initialize(ctx)

	assert.NoError(t, err)
}

// TestLSPAI_OpenDocument tests document opening
func TestLSPAI_OpenDocument(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func main() {
	fmt.Println("Hello")
}
`
	err := lspai.OpenDocument(ctx, uri, content, "go", 1)

	assert.NoError(t, err)

	// Verify document is stored
	doc := lspai.documentStore.Get(uri)
	assert.NotNil(t, doc)
	assert.Equal(t, content, doc.Content)
	assert.Equal(t, "go", doc.LanguageID)
}

// TestLSPAI_CloseDocument tests document closing
func TestLSPAI_CloseDocument(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"

	lspai.OpenDocument(ctx, uri, "content", "go", 1)
	lspai.CloseDocument(uri)

	doc := lspai.documentStore.Get(uri)
	assert.Nil(t, doc)
}

// TestLSPAI_UpdateDocument tests document updating
func TestLSPAI_UpdateDocument(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"

	lspai.OpenDocument(ctx, uri, "old content", "go", 1)

	changes := []TextDocumentContentChangeEvent{
		{Text: "new content"},
	}
	err := lspai.UpdateDocument(ctx, uri, changes, 2)

	assert.NoError(t, err)

	doc := lspai.documentStore.Get(uri)
	assert.Equal(t, "new content", doc.Content)
	assert.Equal(t, 2, doc.Version)
}

// TestLSPAI_UpdateDocument_NotFound tests updating non-existent document
func TestLSPAI_UpdateDocument_NotFound(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()

	err := lspai.UpdateDocument(ctx, "non-existent", nil, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
}

// TestLSPAI_GetCompletion tests completion generation
func TestLSPAI_GetCompletion(t *testing.T) {
	config := DefaultLSPAIConfig()
	config.EnableSemanticCompletion = true
	config.MinConfidence = 0.5

	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func main() {
	fmt.Print
}
`
	lspai.OpenDocument(ctx, uri, content, "go", 1)

	completions, err := lspai.GetCompletion(ctx, uri, Position{Line: 3, Character: 12}, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, completions)
}

// TestLSPAI_GetCompletion_NotFound tests completion for non-existent document
func TestLSPAI_GetCompletion_NotFound(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()

	_, err := lspai.GetCompletion(ctx, "non-existent", Position{}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
}

// TestLSPAI_GetHover tests hover information
func TestLSPAI_GetHover(t *testing.T) {
	config := DefaultLSPAIConfig()
	config.EnableHover = true

	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func testFunc() {
	x := 1
}
`
	lspai.OpenDocument(ctx, uri, content, "go", 1)

	hover, err := lspai.GetHover(ctx, uri, Position{Line: 2, Character: 7})

	assert.NoError(t, err)
	// Hover may or may not return content depending on symbol detection
	_ = hover // Result is optional, just testing no error
}

// TestLSPAI_GetHover_Disabled tests hover when disabled
func TestLSPAI_GetHover_Disabled(t *testing.T) {
	config := DefaultLSPAIConfig()
	config.EnableHover = false

	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()

	hover, err := lspai.GetHover(ctx, "any", Position{})

	assert.NoError(t, err)
	assert.Nil(t, hover)
}

// TestLSPAI_GetCodeActions tests code action generation
func TestLSPAI_GetCodeActions(t *testing.T) {
	config := DefaultLSPAIConfig()
	config.EnableCodeActions = true
	config.EnableRefactoring = true

	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func main() {
	x := 1
}
`
	lspai.OpenDocument(ctx, uri, content, "go", 1)

	diagnostics := []Diagnostic{
		{
			Range: Range{
				Start: Position{Line: 3, Character: 1},
				End:   Position{Line: 3, Character: 2},
			},
			Message: "unused variable x",
		},
	}

	actions, err := lspai.GetCodeActions(ctx, uri, Range{
		Start: Position{Line: 3, Character: 0},
		End:   Position{Line: 3, Character: 10},
	}, diagnostics)

	assert.NoError(t, err)
	assert.NotEmpty(t, actions)
}

// TestLSPAI_GetCodeActions_Disabled tests code actions when disabled
func TestLSPAI_GetCodeActions_Disabled(t *testing.T) {
	config := DefaultLSPAIConfig()
	config.EnableCodeActions = false

	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()

	actions, err := lspai.GetCodeActions(ctx, "any", Range{}, nil)

	assert.NoError(t, err)
	assert.Nil(t, actions)
}

// TestLSPAI_GetDiagnostics tests diagnostic generation
func TestLSPAI_GetDiagnostics(t *testing.T) {
	config := DefaultLSPAIConfig()
	config.EnableDiagnostics = true

	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func main() {
	x := 1
}
`
	lspai.OpenDocument(ctx, uri, content, "go", 1)

	diagnostics, err := lspai.GetDiagnostics(ctx, uri)

	assert.NoError(t, err)
	// Diagnostics may be empty if no issues found
	_ = diagnostics // Result is optional, just testing no error
}

// TestLSPAI_GetDiagnostics_Disabled tests diagnostics when disabled
func TestLSPAI_GetDiagnostics_Disabled(t *testing.T) {
	config := DefaultLSPAIConfig()
	config.EnableDiagnostics = false

	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()

	diagnostics, err := lspai.GetDiagnostics(ctx, "any")

	assert.NoError(t, err)
	assert.Nil(t, diagnostics)
}

// TestLSPAI_GetDefinition tests definition lookup
func TestLSPAI_GetDefinition(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func testFunc() {
	x := 1
}
`
	lspai.OpenDocument(ctx, uri, content, "go", 1)

	locations, err := lspai.GetDefinition(ctx, uri, Position{Line: 2, Character: 7})

	assert.NoError(t, err)
	// May or may not find definitions
	_ = locations // Result is optional, just testing no error
}

// TestLSPAI_GetReferences tests reference lookup
func TestLSPAI_GetReferences(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func testFunc() {
	x := 1
}
`
	lspai.OpenDocument(ctx, uri, content, "go", 1)

	locations, err := lspai.GetReferences(ctx, uri, Position{Line: 2, Character: 7}, true)

	assert.NoError(t, err)
	// May or may not find references
	_ = locations
}

// TestLSPAI_FormatDocument tests document formatting
func TestLSPAI_FormatDocument(t *testing.T) {
	config := DefaultLSPAIConfig()
	provider := &MockAICompletionProvider{}
	lspai := NewLSPAI(config, provider)

	ctx := context.Background()
	uri := "file:///test/main.go"
	content := `package main

func main() {
	x := 1
}
`
	lspai.OpenDocument(ctx, uri, content, "go", 1)

	edits, err := lspai.FormatDocument(ctx, uri)

	assert.NoError(t, err)
	// May or may not return edits
	_ = edits
}

// TestDocumentStore tests document store operations
func TestDocumentStore(t *testing.T) {
	store := NewDocumentStore()

	assert.NotNil(t, store)

	// Test Open
	doc := &Document{
		URI:        "file:///test.go",
		Content:    "content",
		LanguageID: "go",
		Version:    1,
	}
	store.Open(doc)

	// Test Get
	retrieved := store.Get("file:///test.go")
	assert.NotNil(t, retrieved)
	assert.Equal(t, doc.Content, retrieved.Content)

	// Test Close
	store.Close("file:///test.go")
	retrieved = store.Get("file:///test.go")
	assert.Nil(t, retrieved)
}

// TestSymbolIndex tests symbol index operations
func TestSymbolIndex(t *testing.T) {
	index := NewSymbolIndex()

	assert.NotNil(t, index)

	ctx := context.Background()
	err := index.Initialize(ctx)
	assert.NoError(t, err)
}

// TestSymbolIndex_UpdateDocument tests symbol indexing
func TestSymbolIndex_UpdateDocument(t *testing.T) {
	index := NewSymbolIndex()

	symbols := []DocumentSymbol{
		{
			Name: "testFunc",
			Kind: SymbolKindFunction,
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 5, Character: 0},
			},
		},
	}

	index.UpdateDocument("file:///test.go", symbols)

	// Verify symbol was indexed
	inScope := index.GetSymbolsInScope("file:///test.go", Position{Line: 2, Character: 0})
	assert.NotEmpty(t, inScope)
}

// TestSymbolIndex_RemoveDocument tests symbol removal
func TestSymbolIndex_RemoveDocument(t *testing.T) {
	index := NewSymbolIndex()

	symbols := []DocumentSymbol{
		{Name: "test", Kind: SymbolKindFunction},
	}

	index.UpdateDocument("file:///test.go", symbols)
	index.RemoveDocument("file:///test.go")

	inScope := index.GetSymbolsInScope("file:///test.go", Position{})
	assert.Empty(t, inScope)
}

// TestSymbolIndex_FindDefinitions tests definition finding
func TestSymbolIndex_FindDefinitions(t *testing.T) {
	index := NewSymbolIndex()

	symbols := []DocumentSymbol{
		{
			Name: "testFunc",
			Kind: SymbolKindFunction,
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 5, Character: 0},
			},
		},
	}

	index.UpdateDocument("file:///test.go", symbols)

	definitions := index.FindDefinitions("testFunc", SymbolKindFunction)
	assert.NotEmpty(t, definitions)
}

// TestSymbolIndex_FindReferences tests reference finding
func TestSymbolIndex_FindReferences(t *testing.T) {
	index := NewSymbolIndex()

	symbols := []DocumentSymbol{
		{Name: "testFunc", Kind: SymbolKindFunction},
	}

	index.UpdateDocument("file:///test.go", symbols)

	refs := index.FindReferences("testFunc")
	assert.NotEmpty(t, refs)
}

// TestCompletionCache tests completion cache
func TestCompletionCache(t *testing.T) {
	cache := NewCompletionCache(100, 5*time.Minute)

	assert.NotNil(t, cache)
}

// TestCompletionCache_GetSet tests cache get/set
func TestCompletionCache_GetSet(t *testing.T) {
	cache := NewCompletionCache(100, 5*time.Minute)

	items := []CompletionItem{
		{Label: "test", Kind: CompletionItemKindFunction},
	}

	cache.Set("key", items)

	retrieved := cache.Get("key")
	assert.NotNil(t, retrieved)
	assert.Len(t, retrieved, 1)
}

// TestCompletionCache_GetMiss tests cache miss
func TestCompletionCache_GetMiss(t *testing.T) {
	cache := NewCompletionCache(100, 5*time.Minute)

	retrieved := cache.Get("non-existent")
	assert.Nil(t, retrieved)
}

// TestCompletionCache_InvalidateDocument tests document invalidation
func TestCompletionCache_InvalidateDocument(t *testing.T) {
	cache := NewCompletionCache(100, 5*time.Minute)

	cache.Set("file:///test.go:1:5", []CompletionItem{{Label: "a"}})
	cache.Set("file:///test.go:2:10", []CompletionItem{{Label: "b"}})
	cache.Set("file:///other.go:1:5", []CompletionItem{{Label: "c"}})

	cache.InvalidateDocument("file:///test.go")

	assert.Nil(t, cache.Get("file:///test.go:1:5"))
	assert.Nil(t, cache.Get("file:///test.go:2:10"))
	assert.NotNil(t, cache.Get("file:///other.go:1:5"))
}

// TestPositionInRange tests position range checking
func TestPositionInRange(t *testing.T) {
	r := Range{
		Start: Position{Line: 5, Character: 10},
		End:   Position{Line: 5, Character: 20},
	}

	testCases := []struct {
		pos      Position
		expected bool
	}{
		{Position{Line: 5, Character: 15}, true},
		{Position{Line: 5, Character: 10}, true},
		{Position{Line: 5, Character: 20}, true},
		{Position{Line: 5, Character: 9}, false},
		{Position{Line: 5, Character: 21}, false},
		{Position{Line: 4, Character: 15}, false},
		{Position{Line: 6, Character: 15}, false},
	}

	for _, tc := range testCases {
		result := positionInRange(tc.pos, r)
		assert.Equal(t, tc.expected, result, "Position %v", tc.pos)
	}
}

// TestSymbolKinds tests all symbol kind constants
func TestSymbolKinds(t *testing.T) {
	kinds := []SymbolKind{
		SymbolKindFile,
		SymbolKindModule,
		SymbolKindNamespace,
		SymbolKindPackage,
		SymbolKindClass,
		SymbolKindMethod,
		SymbolKindProperty,
		SymbolKindField,
		SymbolKindConstructor,
		SymbolKindEnum,
		SymbolKindInterface,
		SymbolKindFunction,
		SymbolKindVariable,
		SymbolKindConstant,
	}

	for _, kind := range kinds {
		assert.NotEqual(t, SymbolKind(0), kind)
	}
}

// TestCompletionItemKinds tests all completion item kinds
func TestCompletionItemKinds(t *testing.T) {
	kinds := []CompletionItemKind{
		CompletionItemKindText,
		CompletionItemKindMethod,
		CompletionItemKindFunction,
		CompletionItemKindConstructor,
		CompletionItemKindField,
		CompletionItemKindVariable,
		CompletionItemKindClass,
		CompletionItemKindInterface,
	}

	for _, kind := range kinds {
		assert.NotEqual(t, CompletionItemKind(0), kind)
	}
}

// TestDiagnosticSeverities tests diagnostic severity levels
func TestDiagnosticSeverities(t *testing.T) {
	assert.Equal(t, DiagnosticSeverity(1), DiagnosticSeverityError)
	assert.Equal(t, DiagnosticSeverity(2), DiagnosticSeverityWarning)
	assert.Equal(t, DiagnosticSeverity(3), DiagnosticSeverityInformation)
	assert.Equal(t, DiagnosticSeverity(4), DiagnosticSeverityHint)
}

// TestCodeActionKinds tests code action kind constants
func TestCodeActionKinds(t *testing.T) {
	kinds := []CodeActionKind{
		CodeActionKindQuickFix,
		CodeActionKindRefactor,
		CodeActionKindRefactorExtract,
		CodeActionKindRefactorInline,
		CodeActionKindRefactorRewrite,
		CodeActionKindSource,
		CodeActionKindSourceOrganizeImports,
		CodeActionKindSourceFixAll,
	}

	for _, kind := range kinds {
		assert.NotEmpty(t, string(kind))
	}
}

// TestCompletionSources tests completion source constants
func TestCompletionSources(t *testing.T) {
	sources := []CompletionSource{
		CompletionSourceLSP,
		CompletionSourceAI,
		CompletionSourceFIM,
		CompletionSourceSemantic,
		CompletionSourceSnippet,
		CompletionSourceHistory,
	}

	for _, source := range sources {
		assert.NotEmpty(t, string(source))
	}
}

// TestMarkupKinds tests markup kind constants
func TestMarkupKinds(t *testing.T) {
	assert.Equal(t, MarkupKind("plaintext"), MarkupKindPlainText)
	assert.Equal(t, MarkupKind("markdown"), MarkupKindMarkdown)
}

// TestInsertTextFormats tests insert text format constants
func TestInsertTextFormats(t *testing.T) {
	assert.Equal(t, InsertTextFormat(1), InsertTextFormatPlainText)
	assert.Equal(t, InsertTextFormat(2), InsertTextFormatSnippet)
}

// TestCompletionTriggerKinds tests completion trigger kinds
func TestCompletionTriggerKinds(t *testing.T) {
	assert.Equal(t, CompletionTriggerKind(1), CompletionTriggerInvoked)
	assert.Equal(t, CompletionTriggerKind(2), CompletionTriggerCharacter)
	assert.Equal(t, CompletionTriggerKind(3), CompletionTriggerIncomplete)
}

// TestDiagnosticTags tests diagnostic tag constants
func TestDiagnosticTags(t *testing.T) {
	assert.Equal(t, DiagnosticTag(1), DiagnosticTagUnnecessary)
	assert.Equal(t, DiagnosticTag(2), DiagnosticTagDeprecated)
}

// TestExtractSymbols tests symbol extraction from code
func TestExtractSymbols(t *testing.T) {
	config := DefaultLSPAIConfig()
	lspai := NewLSPAI(config, nil)

	code := `package main

func testFunction() {
	x := 1
}

type TestStruct struct {
	Field int
}

type TestInterface interface {
	Method()
}
`

	symbols, err := lspai.extractSymbols(code, "go")

	assert.NoError(t, err)
	assert.NotEmpty(t, symbols)

	// Should find function
	foundFunc := false
	foundStruct := false
	foundInterface := false

	for _, sym := range symbols {
		if sym.Name == "testFunction" && sym.Kind == SymbolKindFunction {
			foundFunc = true
		}
		if sym.Name == "TestStruct" && sym.Kind == SymbolKindStruct {
			foundStruct = true
		}
		if sym.Name == "TestInterface" && sym.Kind == SymbolKindInterface {
			foundInterface = true
		}
	}

	assert.True(t, foundFunc, "Should find testFunction")
	assert.True(t, foundStruct, "Should find TestStruct")
	assert.True(t, foundInterface, "Should find TestInterface")
}

// TestApplyTextEdit tests text edit application
func TestApplyTextEdit(t *testing.T) {
	content := "line1\nline2\nline3\n"

	edit := TextEdit{
		Range: Range{
			Start: Position{Line: 1, Character: 0},
			End:   Position{Line: 1, Character: 5},
		},
		NewText: "replaced",
	}

	result := applyTextEdit(content, edit)

	assert.Contains(t, result, "replaced")
	assert.Contains(t, result, "line1")
	assert.Contains(t, result, "line3")
}

// TestLSPAIResult tests the result structure
func TestLSPAIResult(t *testing.T) {
	result := &LSPAIResult{
		Success:  true,
		Message:  "Operation completed",
		Data:     map[string]string{"key": "value"},
		Duration: 100 * time.Millisecond,
	}

	data, err := result.SerializeToJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "success")
	assert.Contains(t, string(data), "Operation completed")
}

// TestIsWordChar tests word character detection
func TestIsWordChar(t *testing.T) {
	assert.True(t, isWordChar('a'))
	assert.True(t, isWordChar('Z'))
	assert.True(t, isWordChar('5'))
	assert.True(t, isWordChar('_'))
	assert.False(t, isWordChar(' '))
	assert.False(t, isWordChar('.'))
	assert.False(t, isWordChar('-'))
}

// TestExtractFunctionName tests function name extraction
func TestExtractFunctionName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"main()", "main"},
		{"testFunc(a int)", "testFunc"},
		{"(r *Receiver) Method()", "Method"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := extractFunctionName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestExtractTypeName tests type name extraction
func TestExtractTypeName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"MyStruct struct", "MyStruct"},
		{"MyInterface interface", "MyInterface"},
		{"SimpleType", "SimpleType"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := extractTypeName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSymbolKindToCompletionKind tests conversion between symbol and completion kinds
func TestSymbolKindToCompletionKind(t *testing.T) {
	testCases := []struct {
		symbolKind     SymbolKind
		completionKind CompletionItemKind
	}{
		{SymbolKindClass, CompletionItemKindClass},
		{SymbolKindInterface, CompletionItemKindInterface},
		{SymbolKindFunction, CompletionItemKindFunction},
		{SymbolKindMethod, CompletionItemKindMethod},
		{SymbolKindVariable, CompletionItemKindVariable},
		{SymbolKindConstant, CompletionItemKindConstant},
		{SymbolKindField, CompletionItemKindField},
		{SymbolKindProperty, CompletionItemKindProperty},
		{SymbolKindStruct, CompletionItemKindStruct},
	}

	for _, tc := range testCases {
		result := symbolKindToCompletionKind(tc.symbolKind)
		assert.Equal(t, tc.completionKind, result)
	}
}

// TestSymbolKindString tests symbol kind to string conversion
func TestSymbolKindString(t *testing.T) {
	testCases := []struct {
		kind     SymbolKind
		expected string
	}{
		{SymbolKindFile, "file"},
		{SymbolKindModule, "module"},
		{SymbolKindClass, "class"},
		{SymbolKindFunction, "function"},
		{SymbolKindMethod, "method"},
		{SymbolKindVariable, "variable"},
		{SymbolKindStruct, "struct"},
		{SymbolKind(999), "symbol"}, // Unknown kind
	}

	for _, tc := range testCases {
		result := symbolKindString(tc.kind)
		assert.Equal(t, tc.expected, result)
	}
}
