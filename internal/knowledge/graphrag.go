// Package knowledge provides GraphRAG (Graph-based Retrieval Augmented Generation)
// for context-aware code retrieval using the code knowledge graph.
package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RetrievalMode specifies how to retrieve context
type RetrievalMode string

const (
	// RetrievalModeLocal retrieves only directly related nodes
	RetrievalModeLocal RetrievalMode = "local"
	// RetrievalModeGraph traverses the graph for context
	RetrievalModeGraph RetrievalMode = "graph"
	// RetrievalModeHybrid combines vector search with graph traversal
	RetrievalModeHybrid RetrievalMode = "hybrid"
	// RetrievalModeSemantic uses only semantic similarity
	RetrievalModeSemantic RetrievalMode = "semantic"
)

// GraphRAGConfig holds configuration for GraphRAG
type GraphRAGConfig struct {
	// RetrievalMode specifies the retrieval strategy
	RetrievalMode RetrievalMode `json:"retrieval_mode"`
	// MaxNodes is the maximum nodes to retrieve
	MaxNodes int `json:"max_nodes"`
	// MaxDepth is the maximum graph traversal depth
	MaxDepth int `json:"max_depth"`
	// VectorWeight is the weight for vector similarity
	VectorWeight float64 `json:"vector_weight"`
	// GraphWeight is the weight for graph proximity
	GraphWeight float64 `json:"graph_weight"`
	// MinRelevanceScore is the minimum relevance score
	MinRelevanceScore float64 `json:"min_relevance_score"`
	// IncludeEdgeTypes specifies which edges to follow
	IncludeEdgeTypes []EdgeType `json:"include_edge_types"`
	// EnableReranking enables LLM-based reranking
	EnableReranking bool `json:"enable_reranking"`
	// MaxTokens limits context token count
	MaxTokens int `json:"max_tokens"`
}

// DefaultGraphRAGConfig returns default configuration
func DefaultGraphRAGConfig() GraphRAGConfig {
	return GraphRAGConfig{
		RetrievalMode:     RetrievalModeHybrid,
		MaxNodes:          20,
		MaxDepth:          3,
		VectorWeight:      0.6,
		GraphWeight:       0.4,
		MinRelevanceScore: 0.3,
		IncludeEdgeTypes: []EdgeType{
			EdgeTypeCalls,
			EdgeTypeCalledBy,
			EdgeTypeContains,
			EdgeTypeImports,
			EdgeTypeExtends,
			EdgeTypeImplements,
			EdgeTypeReferences,
		},
		EnableReranking: true,
		MaxTokens:       8000,
	}
}

// RetrievedNode represents a retrieved node with relevance info
type RetrievedNode struct {
	Node           *CodeNode `json:"node"`
	RelevanceScore float64   `json:"relevance_score"`
	VectorScore    float64   `json:"vector_score"`
	GraphScore     float64   `json:"graph_score"`
	Depth          int       `json:"depth"`
	Path           []string  `json:"path,omitempty"`
}

// RetrievalResult holds the result of a retrieval operation
type RetrievalResult struct {
	Query      string           `json:"query"`
	Nodes      []*RetrievedNode `json:"nodes"`
	TotalFound int              `json:"total_found"`
	Duration   time.Duration    `json:"duration"`
	Mode       RetrievalMode    `json:"mode"`
	TokenCount int              `json:"token_count,omitempty"`
}

// MarshalJSON implements custom JSON marshaling
func (r *RetrievalResult) MarshalJSON() ([]byte, error) {
	type Alias RetrievalResult
	return json.Marshal(&struct {
		*Alias
		DurationMs int64 `json:"duration_ms"`
	}{
		Alias:      (*Alias)(r),
		DurationMs: r.Duration.Milliseconds(),
	})
}

// GraphRAG implements graph-based retrieval augmented generation
type GraphRAG struct {
	config   GraphRAGConfig
	graph    *CodeGraph
	reranker Reranker
	mu       sync.RWMutex
	logger   *logrus.Logger
}

// Reranker reranks retrieved nodes based on relevance
type Reranker interface {
	// Rerank reranks nodes based on query relevance
	Rerank(ctx context.Context, query string, nodes []*RetrievedNode) ([]*RetrievedNode, error)
}

// NewGraphRAG creates a new GraphRAG instance
func NewGraphRAG(config GraphRAGConfig, graph *CodeGraph, reranker Reranker, logger *logrus.Logger) *GraphRAG {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	return &GraphRAG{
		config:   config,
		graph:    graph,
		reranker: reranker,
		logger:   logger,
	}
}

// Retrieve retrieves relevant context for a query
func (g *GraphRAG) Retrieve(ctx context.Context, query string) (*RetrievalResult, error) {
	startTime := time.Now()

	var nodes []*RetrievedNode
	var err error

	switch g.config.RetrievalMode {
	case RetrievalModeLocal:
		nodes, err = g.retrieveLocal(ctx, query)
	case RetrievalModeGraph:
		nodes, err = g.retrieveGraph(ctx, query)
	case RetrievalModeSemantic:
		nodes, err = g.retrieveSemantic(ctx, query)
	case RetrievalModeHybrid:
		nodes, err = g.retrieveHybrid(ctx, query)
	default:
		nodes, err = g.retrieveHybrid(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	// Filter by minimum score
	filteredNodes := make([]*RetrievedNode, 0)
	for _, node := range nodes {
		if node.RelevanceScore >= g.config.MinRelevanceScore {
			filteredNodes = append(filteredNodes, node)
		}
	}

	// Rerank if enabled
	if g.config.EnableReranking && g.reranker != nil && len(filteredNodes) > 0 {
		filteredNodes, err = g.reranker.Rerank(ctx, query, filteredNodes)
		if err != nil {
			g.logger.Warnf("Reranking failed: %v", err)
		}
	}

	// Limit to max nodes
	if len(filteredNodes) > g.config.MaxNodes {
		filteredNodes = filteredNodes[:g.config.MaxNodes]
	}

	result := &RetrievalResult{
		Query:      query,
		Nodes:      filteredNodes,
		TotalFound: len(nodes),
		Duration:   time.Since(startTime),
		Mode:       g.config.RetrievalMode,
	}

	return result, nil
}

// retrieveLocal retrieves nodes matching the query directly
func (g *GraphRAG) retrieveLocal(ctx context.Context, query string) ([]*RetrievedNode, error) {
	// Use semantic search for local retrieval
	semanticNodes, err := g.graph.SemanticSearch(ctx, query, g.config.MaxNodes*2)
	if err != nil {
		return nil, err
	}

	nodes := make([]*RetrievedNode, 0, len(semanticNodes))
	for i, node := range semanticNodes {
		score := 1.0 - float64(i)/float64(len(semanticNodes))
		nodes = append(nodes, &RetrievedNode{
			Node:           node,
			RelevanceScore: score,
			VectorScore:    score,
			Depth:          0,
		})
	}

	return nodes, nil
}

// retrieveSemantic retrieves nodes using only semantic similarity
func (g *GraphRAG) retrieveSemantic(ctx context.Context, query string) ([]*RetrievedNode, error) {
	return g.retrieveLocal(ctx, query)
}

// retrieveGraph retrieves nodes by traversing the graph
func (g *GraphRAG) retrieveGraph(ctx context.Context, query string) ([]*RetrievedNode, error) {
	// First find seed nodes
	seedNodes, err := g.graph.SemanticSearch(ctx, query, 5)
	if err != nil {
		return nil, err
	}

	if len(seedNodes) == 0 {
		return []*RetrievedNode{}, nil
	}

	// Traverse from seed nodes
	visited := make(map[string]bool)
	retrieved := make(map[string]*RetrievedNode)

	for seedIdx, seedNode := range seedNodes {
		seedScore := 1.0 - float64(seedIdx)/float64(len(seedNodes))
		g.traverseGraph(seedNode.ID, 0, seedScore, []string{}, visited, retrieved)
	}

	// Convert to slice and sort
	nodes := make([]*RetrievedNode, 0, len(retrieved))
	for _, node := range retrieved {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].RelevanceScore > nodes[j].RelevanceScore
	})

	return nodes, nil
}

// traverseGraph traverses the graph from a node
func (g *GraphRAG) traverseGraph(nodeID string, depth int, parentScore float64, path []string, visited map[string]bool, retrieved map[string]*RetrievedNode) {
	if depth > g.config.MaxDepth || visited[nodeID] {
		return
	}

	visited[nodeID] = true
	currentPath := append(path, nodeID)

	node, exists := g.graph.GetNode(nodeID)
	if !exists {
		return
	}

	// Calculate graph score (decays with depth)
	graphScore := parentScore * (1.0 - float64(depth)*0.2)
	if graphScore < 0.1 {
		graphScore = 0.1
	}

	retrieved[nodeID] = &RetrievedNode{
		Node:           node,
		RelevanceScore: graphScore,
		GraphScore:     graphScore,
		Depth:          depth,
		Path:           currentPath,
	}

	// Traverse neighbors
	edges := g.graph.GetOutgoingEdges(nodeID)
	for _, edge := range edges {
		if g.shouldFollowEdge(edge.Type) {
			g.traverseGraph(edge.TargetID, depth+1, graphScore, currentPath, visited, retrieved)
		}
	}

	// Also traverse incoming edges for called_by, etc.
	incomingEdges := g.graph.GetIncomingEdges(nodeID)
	for _, edge := range incomingEdges {
		if g.shouldFollowEdge(edge.Type) {
			g.traverseGraph(edge.SourceID, depth+1, graphScore, currentPath, visited, retrieved)
		}
	}
}

// retrieveHybrid combines vector and graph retrieval
func (g *GraphRAG) retrieveHybrid(ctx context.Context, query string) ([]*RetrievedNode, error) {
	// Get semantic results
	semanticNodes, err := g.graph.SemanticSearch(ctx, query, g.config.MaxNodes*2)
	if err != nil {
		return nil, err
	}

	// Build initial scores from semantic search
	nodeScores := make(map[string]*RetrievedNode)
	for i, node := range semanticNodes {
		vectorScore := 1.0 - float64(i)/float64(len(semanticNodes)+1)
		nodeScores[node.ID] = &RetrievedNode{
			Node:        node,
			VectorScore: vectorScore,
			GraphScore:  0,
			Depth:       0,
		}
	}

	// Traverse graph from top semantic nodes
	if len(semanticNodes) > 0 {
		visited := make(map[string]bool)
		topK := 3
		if topK > len(semanticNodes) {
			topK = len(semanticNodes)
		}

		for i := 0; i < topK; i++ {
			seedNode := semanticNodes[i]
			seedScore := nodeScores[seedNode.ID].VectorScore
			g.traverseGraphHybrid(seedNode.ID, 0, seedScore, visited, nodeScores)
		}
	}

	// Calculate combined scores
	nodes := make([]*RetrievedNode, 0, len(nodeScores))
	for _, rn := range nodeScores {
		rn.RelevanceScore = g.config.VectorWeight*rn.VectorScore + g.config.GraphWeight*rn.GraphScore
		nodes = append(nodes, rn)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].RelevanceScore > nodes[j].RelevanceScore
	})

	return nodes, nil
}

// traverseGraphHybrid traverses graph for hybrid retrieval
func (g *GraphRAG) traverseGraphHybrid(nodeID string, depth int, parentScore float64, visited map[string]bool, nodeScores map[string]*RetrievedNode) {
	if depth > g.config.MaxDepth || visited[nodeID] {
		return
	}

	visited[nodeID] = true

	node, exists := g.graph.GetNode(nodeID)
	if !exists {
		return
	}

	graphScore := parentScore * (1.0 - float64(depth)*0.2)

	if existing, ok := nodeScores[nodeID]; ok {
		if graphScore > existing.GraphScore {
			existing.GraphScore = graphScore
			existing.Depth = depth
		}
	} else {
		nodeScores[nodeID] = &RetrievedNode{
			Node:       node,
			GraphScore: graphScore,
			Depth:      depth,
		}
	}

	// Traverse edges
	edges := g.graph.GetOutgoingEdges(nodeID)
	for _, edge := range edges {
		if g.shouldFollowEdge(edge.Type) {
			g.traverseGraphHybrid(edge.TargetID, depth+1, graphScore, visited, nodeScores)
		}
	}

	incomingEdges := g.graph.GetIncomingEdges(nodeID)
	for _, edge := range incomingEdges {
		if g.shouldFollowEdge(edge.Type) {
			g.traverseGraphHybrid(edge.SourceID, depth+1, graphScore, visited, nodeScores)
		}
	}
}

// shouldFollowEdge checks if an edge type should be followed
func (g *GraphRAG) shouldFollowEdge(edgeType EdgeType) bool {
	for _, t := range g.config.IncludeEdgeTypes {
		if t == edgeType {
			return true
		}
	}
	return false
}

// RetrieveForNode retrieves context for a specific node
func (g *GraphRAG) RetrieveForNode(ctx context.Context, nodeID string) (*RetrievalResult, error) {
	startTime := time.Now()

	node, exists := g.graph.GetNode(nodeID)
	if !exists {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	// Get impact radius
	impactNodes := g.graph.GetImpactRadius(nodeID, g.config.MaxDepth)

	retrieved := make([]*RetrievedNode, 0, len(impactNodes))
	for i, impactNode := range impactNodes {
		score := 1.0 - float64(i)/float64(len(impactNodes)+1)
		retrieved = append(retrieved, &RetrievedNode{
			Node:           impactNode,
			RelevanceScore: score,
			GraphScore:     score,
		})
	}

	return &RetrievalResult{
		Query:      fmt.Sprintf("Context for %s", node.Name),
		Nodes:      retrieved,
		TotalFound: len(retrieved),
		Duration:   time.Since(startTime),
		Mode:       RetrievalModeGraph,
	}, nil
}

// BuildContext builds a text context from retrieved nodes
func (g *GraphRAG) BuildContext(result *RetrievalResult) string {
	if result == nil || len(result.Nodes) == 0 {
		return ""
	}

	var builder stringBuilder
	builder.WriteString("## Relevant Code Context\n\n")

	for i, rn := range result.Nodes {
		builder.WriteString(fmt.Sprintf("### %d. %s (%s)\n", i+1, rn.Node.Name, rn.Node.Type))
		builder.WriteString(fmt.Sprintf("Path: %s\n", rn.Node.Path))
		if rn.Node.StartLine > 0 {
			builder.WriteString(fmt.Sprintf("Lines: %d-%d\n", rn.Node.StartLine, rn.Node.EndLine))
		}
		if rn.Node.Signature != "" {
			builder.WriteString(fmt.Sprintf("Signature: %s\n", rn.Node.Signature))
		}
		if rn.Node.Docstring != "" {
			builder.WriteString(fmt.Sprintf("Documentation: %s\n", rn.Node.Docstring))
		}
		builder.WriteString(fmt.Sprintf("Relevance: %.2f\n\n", rn.RelevanceScore))
	}

	return builder.String()
}

// stringBuilder is a simple string builder
type stringBuilder struct {
	data []byte
}

func (b *stringBuilder) WriteString(s string) {
	b.data = append(b.data, s...)
}

func (b *stringBuilder) String() string {
	return string(b.data)
}

// LLMReranker implements Reranker using an LLM
type LLMReranker struct {
	rerankFunc func(ctx context.Context, query string, documents []string) ([]int, error)
	logger     *logrus.Logger
}

// NewLLMReranker creates a new LLM-based reranker
func NewLLMReranker(rerankFunc func(ctx context.Context, query string, documents []string) ([]int, error), logger *logrus.Logger) *LLMReranker {
	return &LLMReranker{
		rerankFunc: rerankFunc,
		logger:     logger,
	}
}

// Rerank reranks nodes using LLM
func (r *LLMReranker) Rerank(ctx context.Context, query string, nodes []*RetrievedNode) ([]*RetrievedNode, error) {
	if r.rerankFunc == nil || len(nodes) == 0 {
		return nodes, nil
	}

	// Convert nodes to documents
	documents := make([]string, len(nodes))
	for i, node := range nodes {
		documents[i] = fmt.Sprintf("%s: %s", node.Node.Name, node.Node.Docstring)
	}

	// Get reranked indices
	indices, err := r.rerankFunc(ctx, query, documents)
	if err != nil {
		return nodes, err
	}

	// Reorder nodes
	reranked := make([]*RetrievedNode, 0, len(nodes))
	for _, idx := range indices {
		if idx >= 0 && idx < len(nodes) {
			reranked = append(reranked, nodes[idx])
		}
	}

	// Add any nodes that weren't in the reranking
	seen := make(map[int]bool)
	for _, idx := range indices {
		seen[idx] = true
	}
	for i, node := range nodes {
		if !seen[i] {
			reranked = append(reranked, node)
		}
	}

	return reranked, nil
}

// SelectiveRetriever implements the Repoformer selective retrieval policy
type SelectiveRetriever struct {
	config      GraphRAGConfig
	graph       *CodeGraph
	policyModel RetrievalPolicyModel
	graphRAG    *GraphRAG
	logger      *logrus.Logger
}

// RetrievalPolicyModel decides whether to retrieve context
type RetrievalPolicyModel interface {
	// ShouldRetrieve returns true if context retrieval is needed
	ShouldRetrieve(ctx context.Context, query string, localContext string) (bool, error)
}

// NewSelectiveRetriever creates a new selective retriever
func NewSelectiveRetriever(config GraphRAGConfig, graph *CodeGraph, policyModel RetrievalPolicyModel, logger *logrus.Logger) *SelectiveRetriever {
	return &SelectiveRetriever{
		config:      config,
		graph:       graph,
		policyModel: policyModel,
		graphRAG:    NewGraphRAG(config, graph, nil, logger),
		logger:      logger,
	}
}

// Retrieve retrieves context selectively based on policy model
func (r *SelectiveRetriever) Retrieve(ctx context.Context, query string, localContext string) (*RetrievalResult, error) {
	// Check if retrieval is needed
	if r.policyModel != nil {
		shouldRetrieve, err := r.policyModel.ShouldRetrieve(ctx, query, localContext)
		if err != nil {
			r.logger.Warnf("Policy model error: %v", err)
		} else if !shouldRetrieve {
			return &RetrievalResult{
				Query:      query,
				Nodes:      []*RetrievedNode{},
				TotalFound: 0,
				Mode:       RetrievalModeLocal,
			}, nil
		}
	}

	// Perform full retrieval
	return r.graphRAG.Retrieve(ctx, query)
}
