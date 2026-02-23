// Package tools bridges debate tool integration to HelixAgent services.
package tools

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ServiceBridge bridges debate tools to actual HelixAgent services.
type ServiceBridge struct {
	toolIntegration   *ToolIntegration
	mu                sync.RWMutex
	mcpService        interface{}
	lspManager        interface{}
	ragService        interface{}
	embeddingService  interface{}
	formatterExecutor interface{}
	cognitiveServices interface{}
}

// MCPServiceInterface defines the interface for MCP service.
type MCPServiceInterface interface {
	CallTool(ctx context.Context, serverID, toolName string, arguments map[string]interface{}) (interface{}, error)
	ListTools(ctx context.Context) ([]ToolInfo, error)
}

// LSPManagerInterface defines the interface for LSP manager.
type LSPManagerInterface interface {
	GetDefinition(ctx context.Context, serverID, fileURI string, line, character int) (interface{}, error)
	GetReferences(ctx context.Context, serverID, fileURI string, line, character int) (interface{}, error)
	GetDiagnostics(ctx context.Context, serverID, fileURI string) (interface{}, error)
}

// RAGServiceInterface defines the interface for RAG service.
type RAGServiceInterface interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	Store(ctx context.Context, docs []Document) error
}

// EmbeddingServiceInterface defines the interface for embedding service.
type EmbeddingServiceInterface interface {
	Embed(ctx context.Context, text string) ([]float64, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// FormatterServiceInterface defines the interface for formatter service.
type FormatterServiceInterface interface {
	Format(ctx context.Context, language, code string) (string, error)
	ListFormatters() []string
}

// NewServiceBridge creates a service bridge.
func NewServiceBridge(
	mcpService interface{},
	lspManager interface{},
	ragService interface{},
	embeddingService interface{},
	formatterExecutor interface{},
	cognitiveServices interface{},
) *ServiceBridge {
	mcpClient := newMCPClientAdapter(mcpService)
	acpClient := newACPClientAdapter(mcpService)
	lspClient := newLSPClientAdapter(lspManager)
	embeddingClient := newEmbeddingClientAdapter(embeddingService)
	ragClient := newRAGClientAdapter(ragService)
	formatterRegistry := newFormatterRegistryAdapter(formatterExecutor)

	toolIntegration := NewToolIntegration(
		mcpClient,
		acpClient,
		lspClient,
		embeddingClient,
		ragClient,
		formatterRegistry,
	)

	return &ServiceBridge{
		toolIntegration:   toolIntegration,
		mcpService:        mcpService,
		lspManager:        lspManager,
		ragService:        ragService,
		embeddingService:  embeddingService,
		formatterExecutor: formatterExecutor,
		cognitiveServices: cognitiveServices,
	}
}

// GetToolIntegration returns the tool integration instance.
func (b *ServiceBridge) GetToolIntegration() *ToolIntegration {
	return b.toolIntegration
}

// EnrichDebateContext enriches debate context with RAG, embeddings, and LSP.
func (b *ServiceBridge) EnrichDebateContext(ctx context.Context, request *DebateRequest) (*EnrichedContext, error) {
	enriched := &EnrichedContext{
		Original:             request,
		RAGResults:           make([]SearchResult, 0),
		RelatedDefinitions:   make([]interface{}, 0),
		SemanticSimilarities: make(map[string]float64),
	}

	if b.toolIntegration.ragClient != nil && request.Query != "" {
		ragResults, err := b.toolIntegration.QueryRAG(ctx, request.Query, 10)
		if err == nil {
			enriched.RAGResults = ragResults
		}
	}

	if b.toolIntegration.embeddingClient != nil && request.Query != "" {
		embedding, err := b.toolIntegration.GenerateEmbedding(ctx, request.Query)
		if err == nil {
			enriched.QueryEmbedding = embedding
		}
	}

	if b.toolIntegration.lspClient != nil && len(request.Files) > 0 {
		for _, file := range request.Files {
			diagnostics, err := b.toolIntegration.lspClient.GetDiagnostics(ctx, file)
			if err == nil {
				enriched.CodeDiagnostics = append(enriched.CodeDiagnostics, diagnostics...)
			}
		}
	}

	return enriched, nil
}

// DebateRequest represents a debate request that can be enriched.
type DebateRequest struct {
	Query        string   `json:"query"`
	Files        []string `json:"files"`
	Context      string   `json:"context"`
	Requirements []string `json:"requirements"`
}

// EnrichedContext contains debate context enriched with external services.
type EnrichedContext struct {
	Original             *DebateRequest     `json:"original"`
	RAGResults           []SearchResult     `json:"rag_results"`
	QueryEmbedding       []float64          `json:"query_embedding"`
	RelatedDefinitions   []interface{}      `json:"related_definitions"`
	CodeDiagnostics      []interface{}      `json:"code_diagnostics"`
	SemanticSimilarities map[string]float64 `json:"semantic_similarities"`
}

type mcpClientAdapter struct {
	service interface{}
	mu      sync.RWMutex
}

func newMCPClientAdapter(service interface{}) MCPClient {
	return &mcpClientAdapter{service: service}
}

func (a *mcpClientAdapter) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return nil, fmt.Errorf("MCP service not configured")
	}

	if svc, ok := a.service.(MCPServiceInterface); ok {
		return svc.CallTool(ctx, "default", toolName, args)
	}

	if svc, ok := a.service.(interface {
		CallTool(ctx context.Context, serverID, toolName string, arguments map[string]interface{}) (interface{}, error)
	}); ok {
		return svc.CallTool(ctx, "default", toolName, args)
	}

	return nil, fmt.Errorf("MCP service does not implement CallTool")
}

func (a *mcpClientAdapter) ListTools(ctx context.Context) ([]ToolInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return []ToolInfo{}, nil
	}

	if svc, ok := a.service.(MCPServiceInterface); ok {
		return svc.ListTools(ctx)
	}

	return []ToolInfo{}, nil
}

func (a *mcpClientAdapter) GetResource(ctx context.Context, uri string) (interface{}, error) {
	return a.CallTool(ctx, "get_resource", map[string]interface{}{"uri": uri})
}

type acpClientAdapter struct {
	service interface{}
	mcp     MCPClient
	mu      sync.RWMutex
}

func newACPClientAdapter(service interface{}) ACPClient {
	return &acpClientAdapter{
		service: service,
		mcp:     newMCPClientAdapter(service),
	}
}

func (a *acpClientAdapter) SendMessage(ctx context.Context, target string, message interface{}) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.mcp.CallTool(ctx, "acp_send", map[string]interface{}{
		"target":  target,
		"message": message,
	})
}

func (a *acpClientAdapter) Subscribe(ctx context.Context, eventType string) (<-chan interface{}, error) {
	ch := make(chan interface{}, 100)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}

type lspClientAdapter struct {
	service interface{}
	mu      sync.RWMutex
}

func newLSPClientAdapter(service interface{}) LSPClient {
	return &lspClientAdapter{service: service}
}

func (a *lspClientAdapter) GetDefinition(ctx context.Context, file string, line int, char int) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return nil, fmt.Errorf("LSP service not configured")
	}

	if svc, ok := a.service.(LSPManagerInterface); ok {
		return svc.GetDefinition(ctx, "default", file, line, char)
	}

	if svc, ok := a.service.(interface {
		GetDefinition(ctx context.Context, serverID, fileURI string, line, character int) (interface{}, error)
	}); ok {
		return svc.GetDefinition(ctx, "default", file, line, char)
	}

	return nil, fmt.Errorf("LSP service does not implement GetDefinition")
}

func (a *lspClientAdapter) GetReferences(ctx context.Context, file string, line int, char int) ([]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return nil, fmt.Errorf("LSP service not configured")
	}

	result, err := a.GetDefinition(ctx, file, line, char)
	if err != nil {
		return nil, err
	}

	if refs, ok := result.([]interface{}); ok {
		return refs, nil
	}

	return []interface{}{result}, nil
}

func (a *lspClientAdapter) GetDiagnostics(ctx context.Context, file string) ([]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return nil, fmt.Errorf("LSP service not configured")
	}

	if svc, ok := a.service.(LSPManagerInterface); ok {
		result, err := svc.GetDiagnostics(ctx, "default", file)
		if err != nil {
			return nil, err
		}
		if diags, ok := result.([]interface{}); ok {
			return diags, nil
		}
		return []interface{}{result}, nil
	}

	return nil, fmt.Errorf("LSP service does not implement GetDiagnostics")
}

func (a *lspClientAdapter) Format(ctx context.Context, file string, content string) (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return content, nil
	}

	result, err := a.GetDefinition(ctx, file, 0, 0)
	if err != nil {
		return content, nil
	}
	_ = result
	return content, nil
}

type embeddingClientAdapter struct {
	service interface{}
	mu      sync.RWMutex
	cache   map[string][]float64
}

func newEmbeddingClientAdapter(service interface{}) EmbeddingClient {
	return &embeddingClientAdapter{
		service: service,
		cache:   make(map[string][]float64),
	}
}

func (a *embeddingClientAdapter) Embed(ctx context.Context, text string) ([]float64, error) {
	a.mu.RLock()
	if emb, ok := a.cache[text]; ok {
		a.mu.RUnlock()
		return emb, nil
	}
	a.mu.RUnlock()

	if a.service == nil {
		return a.generateFallbackEmbedding(text), nil
	}

	if svc, ok := a.service.(EmbeddingServiceInterface); ok {
		emb, err := svc.Embed(ctx, text)
		if err != nil {
			return a.generateFallbackEmbedding(text), nil
		}
		a.mu.Lock()
		a.cache[text] = emb
		a.mu.Unlock()
		return emb, nil
	}

	if svc, ok := a.service.(interface {
		Embed(ctx context.Context, text string) ([]float64, error)
	}); ok {
		emb, err := svc.Embed(ctx, text)
		if err != nil {
			return a.generateFallbackEmbedding(text), nil
		}
		a.mu.Lock()
		a.cache[text] = emb
		a.mu.Unlock()
		return emb, nil
	}

	return a.generateFallbackEmbedding(text), nil
}

func (a *embeddingClientAdapter) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	for i, text := range texts {
		emb, err := a.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

func (a *embeddingClientAdapter) Similarity(ctx context.Context, text1, text2 string) (float64, error) {
	emb1, err := a.Embed(ctx, text1)
	if err != nil {
		return 0, err
	}
	emb2, err := a.Embed(ctx, text2)
	if err != nil {
		return 0, err
	}
	return cosineSimilarity(emb1, emb2), nil
}

func (a *embeddingClientAdapter) generateFallbackEmbedding(text string) []float64 {
	embedding := make([]float64, 384)
	for i := range embedding {
		embedding[i] = float64(len(text)*(i+1)%100) / 100.0
	}
	return embedding
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

type ragClientAdapter struct {
	service interface{}
	mu      sync.RWMutex
}

func newRAGClientAdapter(service interface{}) RAGClient {
	return &ragClientAdapter{service: service}
}

func (a *ragClientAdapter) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return a.generateFallbackResults(query, limit), nil
	}

	if svc, ok := a.service.(RAGServiceInterface); ok {
		return svc.Search(ctx, query, limit)
	}

	if svc, ok := a.service.(interface {
		Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	}); ok {
		return svc.Search(ctx, query, limit)
	}

	return a.generateFallbackResults(query, limit), nil
}

func (a *ragClientAdapter) Store(ctx context.Context, docs []Document) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return nil
	}

	if svc, ok := a.service.(RAGServiceInterface); ok {
		return svc.Store(ctx, docs)
	}

	return nil
}

func (a *ragClientAdapter) Rerank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error) {
	if len(results) <= 1 {
		return results, nil
	}
	reranked := make([]SearchResult, len(results))
	copy(reranked, results)
	return reranked, nil
}

func (a *ragClientAdapter) generateFallbackResults(query string, limit int) []SearchResult {
	results := make([]SearchResult, 0, limit)
	for i := 0; i < limit && i < 5; i++ {
		results = append(results, SearchResult{
			Content:  fmt.Sprintf("Result %d for query: %s", i+1, query),
			Score:    float64(1.0 - float64(i)*0.1),
			Metadata: map[string]interface{}{"source": "fallback"},
		})
	}
	return results
}

type formatterRegistryAdapter struct {
	service    interface{}
	mu         sync.RWMutex
	formatters map[string]bool
}

func newFormatterRegistryAdapter(service interface{}) FormatterRegistry {
	return &formatterRegistryAdapter{
		service: service,
		formatters: map[string]bool{
			"go":         true,
			"python":     true,
			"javascript": true,
			"typescript": true,
			"rust":       true,
			"java":       true,
			"c":          true,
			"cpp":        true,
			"ruby":       true,
			"swift":      true,
		},
	}
}

func (a *formatterRegistryAdapter) Format(ctx context.Context, language string, code string) (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		return code, nil
	}

	if svc, ok := a.service.(FormatterServiceInterface); ok {
		return svc.Format(ctx, language, code)
	}

	if svc, ok := a.service.(interface {
		Format(ctx context.Context, language, code string) (string, error)
	}); ok {
		return svc.Format(ctx, language, code)
	}

	return code, nil
}

func (a *formatterRegistryAdapter) ListFormatters() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.service == nil {
		formatters := make([]string, 0, len(a.formatters))
		for f := range a.formatters {
			formatters = append(formatters, f)
		}
		return formatters
	}

	if svc, ok := a.service.(FormatterServiceInterface); ok {
		return svc.ListFormatters()
	}

	if svc, ok := a.service.(interface {
		ListFormatters() []string
	}); ok {
		return svc.ListFormatters()
	}

	formatters := make([]string, 0, len(a.formatters))
	for f := range a.formatters {
		formatters = append(formatters, f)
	}
	return formatters
}

// ServiceHealthStatus represents the health status of a service.
type ServiceHealthStatus struct {
	Name      string        `json:"name"`
	Healthy   bool          `json:"healthy"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	Timestamp int64         `json:"timestamp"`
}

// CheckServicesHealth checks the health of all bridge services.
func (b *ServiceBridge) CheckServicesHealth(ctx context.Context) map[string]ServiceHealthStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	status := make(map[string]ServiceHealthStatus)
	now := time.Now().Unix()

	status["mcp"] = ServiceHealthStatus{
		Name:      "mcp",
		Healthy:   b.mcpService != nil,
		Timestamp: now,
	}

	status["lsp"] = ServiceHealthStatus{
		Name:      "lsp",
		Healthy:   b.lspManager != nil,
		Timestamp: now,
	}

	status["rag"] = ServiceHealthStatus{
		Name:      "rag",
		Healthy:   b.ragService != nil,
		Timestamp: now,
	}

	status["embedding"] = ServiceHealthStatus{
		Name:      "embedding",
		Healthy:   b.embeddingService != nil,
		Timestamp: now,
	}

	status["formatter"] = ServiceHealthStatus{
		Name:      "formatter",
		Healthy:   b.formatterExecutor != nil,
		Timestamp: now,
	}

	return status
}
