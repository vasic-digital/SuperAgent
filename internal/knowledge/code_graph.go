// Package knowledge provides Code Knowledge Graph implementation for
// repository-scale code intelligence and semantic understanding.
package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// NodeType represents the type of a node in the code graph
type NodeType string

const (
	NodeTypeFile      NodeType = "file"
	NodeTypeClass     NodeType = "class"
	NodeTypeInterface NodeType = "interface"
	NodeTypeStruct    NodeType = "struct"
	NodeTypeFunction  NodeType = "function"
	NodeTypeMethod    NodeType = "method"
	NodeTypeVariable  NodeType = "variable"
	NodeTypeConstant  NodeType = "constant"
	NodeTypePackage   NodeType = "package"
	NodeTypeModule    NodeType = "module"
	NodeTypeImport    NodeType = "import"
	NodeTypeConcept   NodeType = "concept"
)

// EdgeType represents the type of an edge in the code graph
type EdgeType string

const (
	EdgeTypeContains            EdgeType = "contains"
	EdgeTypeImports             EdgeType = "imports"
	EdgeTypeExtends             EdgeType = "extends"
	EdgeTypeImplements          EdgeType = "implements"
	EdgeTypeCalls               EdgeType = "calls"
	EdgeTypeCalledBy            EdgeType = "called_by"
	EdgeTypeReferences          EdgeType = "references"
	EdgeTypeReferencedBy        EdgeType = "referenced_by"
	EdgeTypeOverrides           EdgeType = "overrides"
	EdgeTypeUses                EdgeType = "uses"
	EdgeTypeModifies            EdgeType = "modifies"
	EdgeTypeReturns             EdgeType = "returns"
	EdgeTypeSemanticallySimilar EdgeType = "semantically_similar"
	EdgeTypeDependsOn           EdgeType = "depends_on"
)

// CodeNode represents a node in the code knowledge graph
type CodeNode struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      NodeType               `json:"type"`
	Path      string                 `json:"path"`
	StartLine int                    `json:"start_line,omitempty"`
	EndLine   int                    `json:"end_line,omitempty"`
	Signature string                 `json:"signature,omitempty"`
	Docstring string                 `json:"docstring,omitempty"`
	Embedding []float64              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// CodeEdge represents an edge in the code knowledge graph
type CodeEdge struct {
	ID        string                 `json:"id"`
	SourceID  string                 `json:"source_id"`
	TargetID  string                 `json:"target_id"`
	Type      EdgeType               `json:"type"`
	Weight    float64                `json:"weight"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// CodeGraphConfig holds configuration for the code graph
type CodeGraphConfig struct {
	// MaxNodes is the maximum number of nodes
	MaxNodes int `json:"max_nodes"`
	// MaxEdgesPerNode is the max edges per node
	MaxEdgesPerNode int `json:"max_edges_per_node"`
	// EnableEmbeddings enables embedding generation
	EnableEmbeddings bool `json:"enable_embeddings"`
	// EmbeddingDimension is the embedding dimension
	EmbeddingDimension int `json:"embedding_dimension"`
	// SimilarityThreshold for semantic edges
	SimilarityThreshold float64 `json:"similarity_threshold"`
	// IndexLanguages to index
	IndexLanguages []string `json:"index_languages"`
	// ExcludePatterns for files to exclude
	ExcludePatterns []string `json:"exclude_patterns"`
}

// DefaultCodeGraphConfig returns default configuration
func DefaultCodeGraphConfig() CodeGraphConfig {
	return CodeGraphConfig{
		MaxNodes:            100000,
		MaxEdgesPerNode:     100,
		EnableEmbeddings:    true,
		EmbeddingDimension:  768,
		SimilarityThreshold: 0.8,
		IndexLanguages: []string{
			"go", "python", "javascript", "typescript", "java",
			"rust", "c", "cpp", "csharp", "ruby", "php",
		},
		ExcludePatterns: []string{
			"**/node_modules/**",
			"**/vendor/**",
			"**/.git/**",
			"**/dist/**",
			"**/build/**",
		},
	}
}

// CodeGraph implements the code knowledge graph
type CodeGraph struct {
	config        CodeGraphConfig
	nodes         map[string]*CodeNode
	edges         map[string]*CodeEdge
	nodesByType   map[NodeType][]*CodeNode
	edgesBySource map[string][]*CodeEdge
	edgesByTarget map[string][]*CodeEdge
	embedder      EmbeddingGenerator
	mu            sync.RWMutex
	logger        *logrus.Logger
}

// EmbeddingGenerator generates embeddings for code
type EmbeddingGenerator interface {
	// GenerateEmbedding generates an embedding for text
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
	// GenerateBatchEmbeddings generates embeddings for multiple texts
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error)
}

// NewCodeGraph creates a new code knowledge graph
func NewCodeGraph(config CodeGraphConfig, embedder EmbeddingGenerator, logger *logrus.Logger) *CodeGraph {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	return &CodeGraph{
		config:        config,
		nodes:         make(map[string]*CodeNode),
		edges:         make(map[string]*CodeEdge),
		nodesByType:   make(map[NodeType][]*CodeNode),
		edgesBySource: make(map[string][]*CodeEdge),
		edgesByTarget: make(map[string][]*CodeEdge),
		embedder:      embedder,
		logger:        logger,
	}
}

// AddNode adds a node to the graph
func (g *CodeGraph) AddNode(node *CodeNode) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.nodes) >= g.config.MaxNodes {
		return fmt.Errorf("maximum nodes reached: %d", g.config.MaxNodes)
	}

	if node.ID == "" {
		node.ID = g.generateNodeID(node)
	}

	now := time.Now()
	if node.CreatedAt.IsZero() {
		node.CreatedAt = now
	}
	node.UpdatedAt = now

	if node.Metadata == nil {
		node.Metadata = make(map[string]interface{})
	}

	g.nodes[node.ID] = node
	g.nodesByType[node.Type] = append(g.nodesByType[node.Type], node)

	return nil
}

// AddEdge adds an edge to the graph
func (g *CodeGraph) AddEdge(edge *CodeEdge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Verify nodes exist
	if _, exists := g.nodes[edge.SourceID]; !exists {
		return fmt.Errorf("source node not found: %s", edge.SourceID)
	}
	if _, exists := g.nodes[edge.TargetID]; !exists {
		return fmt.Errorf("target node not found: %s", edge.TargetID)
	}

	// Check edge limit
	if len(g.edgesBySource[edge.SourceID]) >= g.config.MaxEdgesPerNode {
		return fmt.Errorf("max edges reached for node: %s", edge.SourceID)
	}

	if edge.ID == "" {
		edge.ID = fmt.Sprintf("%s-%s-%s", edge.SourceID, edge.Type, edge.TargetID)
	}

	if edge.CreatedAt.IsZero() {
		edge.CreatedAt = time.Now()
	}

	if edge.Weight == 0 {
		edge.Weight = 1.0
	}

	g.edges[edge.ID] = edge
	g.edgesBySource[edge.SourceID] = append(g.edgesBySource[edge.SourceID], edge)
	g.edgesByTarget[edge.TargetID] = append(g.edgesByTarget[edge.TargetID], edge)

	return nil
}

// GetNode retrieves a node by ID
func (g *CodeGraph) GetNode(id string) (*CodeNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	node, exists := g.nodes[id]
	return node, exists
}

// GetNodesByType retrieves nodes by type
func (g *CodeGraph) GetNodesByType(nodeType NodeType) []*CodeNode {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.nodesByType[nodeType]
}

// GetOutgoingEdges gets edges from a node
func (g *CodeGraph) GetOutgoingEdges(nodeID string) []*CodeEdge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edgesBySource[nodeID]
}

// GetIncomingEdges gets edges to a node
func (g *CodeGraph) GetIncomingEdges(nodeID string) []*CodeEdge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edgesByTarget[nodeID]
}

// GetNeighbors gets neighbor nodes
func (g *CodeGraph) GetNeighbors(nodeID string, edgeType EdgeType) []*CodeNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	neighbors := make([]*CodeNode, 0)
	edges := g.edgesBySource[nodeID]

	for _, edge := range edges {
		if edgeType == "" || edge.Type == edgeType {
			if node, exists := g.nodes[edge.TargetID]; exists {
				neighbors = append(neighbors, node)
			}
		}
	}

	return neighbors
}

// GetImpactRadius finds all nodes affected by changes to a node
func (g *CodeGraph) GetImpactRadius(nodeID string, maxDepth int) []*CodeNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[string]bool)
	impacted := make([]*CodeNode, 0)

	g.traverseImpact(nodeID, 0, maxDepth, visited, &impacted)

	return impacted
}

func (g *CodeGraph) traverseImpact(nodeID string, depth, maxDepth int, visited map[string]bool, impacted *[]*CodeNode) {
	if depth > maxDepth || visited[nodeID] {
		return
	}

	visited[nodeID] = true

	if node, exists := g.nodes[nodeID]; exists {
		*impacted = append(*impacted, node)
	}

	// Traverse incoming edges (nodes that depend on this one)
	for _, edge := range g.edgesByTarget[nodeID] {
		if edge.Type == EdgeTypeCalledBy || edge.Type == EdgeTypeReferencedBy || edge.Type == EdgeTypeDependsOn {
			g.traverseImpact(edge.SourceID, depth+1, maxDepth, visited, impacted)
		}
	}
}

// FindPath finds a path between two nodes
func (g *CodeGraph) FindPath(sourceID, targetID string) ([]*CodeNode, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[sourceID]; !exists {
		return nil, fmt.Errorf("source node not found: %s", sourceID)
	}
	if _, exists := g.nodes[targetID]; !exists {
		return nil, fmt.Errorf("target node not found: %s", targetID)
	}

	// BFS to find shortest path
	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{sourceID}
	visited[sourceID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == targetID {
			// Reconstruct path
			path := []*CodeNode{}
			for id := targetID; id != ""; id = parent[id] {
				if node, exists := g.nodes[id]; exists {
					path = append([]*CodeNode{node}, path...)
				}
			}
			return path, nil
		}

		for _, edge := range g.edgesBySource[current] {
			if !visited[edge.TargetID] {
				visited[edge.TargetID] = true
				parent[edge.TargetID] = current
				queue = append(queue, edge.TargetID)
			}
		}
	}

	return nil, fmt.Errorf("no path found between %s and %s", sourceID, targetID)
}

// SemanticSearch finds semantically similar nodes
func (g *CodeGraph) SemanticSearch(ctx context.Context, query string, topK int) ([]*CodeNode, error) {
	if g.embedder == nil {
		return nil, fmt.Errorf("embedder not configured")
	}

	queryEmbedding, err := g.embedder.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	type scoredNode struct {
		node  *CodeNode
		score float64
	}

	scoredNodes := make([]scoredNode, 0)

	for _, node := range g.nodes {
		if len(node.Embedding) == 0 {
			continue
		}

		score := cosineSimilarity(queryEmbedding, node.Embedding)
		if score >= g.config.SimilarityThreshold {
			scoredNodes = append(scoredNodes, scoredNode{node, score})
		}
	}

	// Sort by score
	for i := 0; i < len(scoredNodes); i++ {
		for j := i + 1; j < len(scoredNodes); j++ {
			if scoredNodes[j].score > scoredNodes[i].score {
				scoredNodes[i], scoredNodes[j] = scoredNodes[j], scoredNodes[i]
			}
		}
	}

	// Return top K
	results := make([]*CodeNode, 0, topK)
	for i := 0; i < len(scoredNodes) && i < topK; i++ {
		results = append(results, scoredNodes[i].node)
	}

	return results, nil
}

// GenerateEmbeddings generates embeddings for all nodes
func (g *CodeGraph) GenerateEmbeddings(ctx context.Context) error {
	if g.embedder == nil {
		return fmt.Errorf("embedder not configured")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Collect texts for batch embedding
	nodeIDs := make([]string, 0)
	texts := make([]string, 0)

	for id, node := range g.nodes {
		text := g.nodeToText(node)
		if text != "" {
			nodeIDs = append(nodeIDs, id)
			texts = append(texts, text)
		}
	}

	// Generate embeddings in batches
	batchSize := 100
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		embeddings, err := g.embedder.GenerateBatchEmbeddings(ctx, texts[i:end])
		if err != nil {
			g.logger.Warnf("Failed to generate embeddings for batch: %v", err)
			continue
		}

		for j, embedding := range embeddings {
			idx := i + j
			if idx < len(nodeIDs) {
				if node, exists := g.nodes[nodeIDs[idx]]; exists {
					node.Embedding = embedding
				}
			}
		}
	}

	return nil
}

// BuildSemanticEdges creates edges between semantically similar nodes
func (g *CodeGraph) BuildSemanticEdges(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	nodesWithEmbeddings := make([]*CodeNode, 0)
	for _, node := range g.nodes {
		if len(node.Embedding) > 0 {
			nodesWithEmbeddings = append(nodesWithEmbeddings, node)
		}
	}

	// Compare all pairs (O(n²) - consider optimization for large graphs)
	for i := 0; i < len(nodesWithEmbeddings); i++ {
		for j := i + 1; j < len(nodesWithEmbeddings); j++ {
			similarity := cosineSimilarity(nodesWithEmbeddings[i].Embedding, nodesWithEmbeddings[j].Embedding)
			if similarity >= g.config.SimilarityThreshold {
				edge := &CodeEdge{
					SourceID:  nodesWithEmbeddings[i].ID,
					TargetID:  nodesWithEmbeddings[j].ID,
					Type:      EdgeTypeSemanticallySimilar,
					Weight:    similarity,
					CreatedAt: time.Now(),
				}
				edge.ID = fmt.Sprintf("%s-semantic-%s", edge.SourceID, edge.TargetID)
				g.edges[edge.ID] = edge
				g.edgesBySource[edge.SourceID] = append(g.edgesBySource[edge.SourceID], edge)
				g.edgesByTarget[edge.TargetID] = append(g.edgesByTarget[edge.TargetID], edge)
			}
		}
	}

	return nil
}

// nodeToText converts a node to text for embedding
func (g *CodeGraph) nodeToText(node *CodeNode) string {
	parts := []string{node.Name}
	if node.Signature != "" {
		parts = append(parts, node.Signature)
	}
	if node.Docstring != "" {
		parts = append(parts, node.Docstring)
	}
	return strings.Join(parts, " ")
}

// generateNodeID generates a unique node ID
func (g *CodeGraph) generateNodeID(node *CodeNode) string {
	return fmt.Sprintf("%s:%s:%s:%d", node.Type, node.Path, node.Name, node.StartLine)
}

// GetStats returns graph statistics
func (g *CodeGraph) GetStats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodesByType := make(map[string]int)
	for nodeType, nodes := range g.nodesByType {
		nodesByType[string(nodeType)] = len(nodes)
	}

	edgesByType := make(map[string]int)
	for _, edge := range g.edges {
		edgesByType[string(edge.Type)]++
	}

	return map[string]interface{}{
		"total_nodes":   len(g.nodes),
		"total_edges":   len(g.edges),
		"nodes_by_type": nodesByType,
		"edges_by_type": edgesByType,
	}
}

// MarshalJSON implements custom JSON marshaling
func (g *CodeGraph) MarshalJSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]*CodeNode, 0, len(g.nodes))
	for _, node := range g.nodes {
		nodes = append(nodes, node)
	}

	edges := make([]*CodeEdge, 0, len(g.edges))
	for _, edge := range g.edges {
		edges = append(edges, edge)
	}

	return json.Marshal(map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
		"stats": g.GetStats(),
	})
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// CodeParser parses code files and extracts nodes/edges
type CodeParser interface {
	// ParseFile parses a file and returns nodes and edges
	ParseFile(ctx context.Context, path string, content []byte) ([]*CodeNode, []*CodeEdge, error)
	// SupportedLanguages returns supported languages
	SupportedLanguages() []string
}

// DefaultCodeParser provides basic code parsing
type DefaultCodeParser struct {
	languagePatterns map[string]*regexp.Regexp
	logger           *logrus.Logger
}

// NewDefaultCodeParser creates a new default code parser
func NewDefaultCodeParser(logger *logrus.Logger) *DefaultCodeParser {
	return &DefaultCodeParser{
		languagePatterns: map[string]*regexp.Regexp{
			"go":         regexp.MustCompile(`(?m)^func\s+(?:\(.*?\)\s+)?(\w+)\s*\(`),
			"python":     regexp.MustCompile(`(?m)^def\s+(\w+)\s*\(`),
			"javascript": regexp.MustCompile(`(?m)(?:function\s+(\w+)|(\w+)\s*[=:]\s*(?:async\s+)?function)`),
			"java":       regexp.MustCompile(`(?m)(?:public|private|protected)\s+\w+\s+(\w+)\s*\(`),
			"rust":       regexp.MustCompile(`(?m)^(?:pub\s+)?fn\s+(\w+)\s*[<(]`),
		},
		logger: logger,
	}
}

// ParseFile parses a file and extracts code elements
func (p *DefaultCodeParser) ParseFile(ctx context.Context, path string, content []byte) ([]*CodeNode, []*CodeEdge, error) {
	nodes := make([]*CodeNode, 0)
	edges := make([]*CodeEdge, 0)

	// Detect language from extension
	ext := filepath.Ext(path)
	lang := strings.TrimPrefix(ext, ".")

	// Create file node
	fileNode := &CodeNode{
		ID:        fmt.Sprintf("file:%s", path),
		Name:      filepath.Base(path),
		Type:      NodeTypeFile,
		Path:      path,
		CreatedAt: time.Now(),
		Metadata:  map[string]interface{}{"language": lang},
	}
	nodes = append(nodes, fileNode)

	// Parse functions/methods based on language
	pattern, ok := p.languagePatterns[lang]
	if !ok {
		return nodes, edges, nil
	}

	matches := pattern.FindAllSubmatchIndex(content, -1)
	lines := strings.Split(string(content), "\n")

	for i, match := range matches {
		if len(match) < 4 {
			continue
		}

		nameStart := match[2]
		nameEnd := match[3]
		if nameStart < 0 || nameEnd < 0 {
			nameStart = match[4]
			nameEnd = match[5]
		}
		if nameStart < 0 || nameEnd < 0 {
			continue
		}

		funcName := string(content[nameStart:nameEnd])
		lineNum := p.getLineNumber(content, nameStart)

		funcNode := &CodeNode{
			ID:        fmt.Sprintf("func:%s:%s:%d", path, funcName, lineNum),
			Name:      funcName,
			Type:      NodeTypeFunction,
			Path:      path,
			StartLine: lineNum,
			CreatedAt: time.Now(),
			Metadata:  map[string]interface{}{"index": i},
		}

		// Extract docstring if available
		if lineNum > 1 && lineNum <= len(lines) {
			prevLine := lines[lineNum-2]
			if strings.Contains(prevLine, "//") || strings.Contains(prevLine, "#") || strings.Contains(prevLine, "\"\"\"") {
				funcNode.Docstring = strings.TrimSpace(prevLine)
			}
		}

		nodes = append(nodes, funcNode)

		// Create contains edge
		edges = append(edges, &CodeEdge{
			SourceID:  fileNode.ID,
			TargetID:  funcNode.ID,
			Type:      EdgeTypeContains,
			CreatedAt: time.Now(),
		})
	}

	return nodes, edges, nil
}

// getLineNumber gets line number from byte offset
func (p *DefaultCodeParser) getLineNumber(content []byte, offset int) int {
	lineNum := 1
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			lineNum++
		}
	}
	return lineNum
}

// SupportedLanguages returns supported languages
func (p *DefaultCodeParser) SupportedLanguages() []string {
	langs := make([]string, 0, len(p.languagePatterns))
	for lang := range p.languagePatterns {
		langs = append(langs, lang)
	}
	return langs
}
