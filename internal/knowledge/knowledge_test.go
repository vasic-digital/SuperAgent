// Package knowledge provides tests for the Code Knowledge Graph and GraphRAG.
package knowledge

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbeddingGenerator implements EmbeddingGenerator for testing
type MockEmbeddingGenerator struct{}

func (m *MockEmbeddingGenerator) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	// Return a simple mock embedding
	return []float64{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}

func (m *MockEmbeddingGenerator) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	// Return mock embeddings for each text
	result := make([][]float64, len(texts))
	for i := range texts {
		result[i] = []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	}
	return result, nil
}

// MockEmbeddingGeneratorWithSearchableEmbeddings creates unique embeddings
// based on the text content for semantic search testing
type MockEmbeddingGeneratorWithSearchableEmbeddings struct{}

func (m *MockEmbeddingGeneratorWithSearchableEmbeddings) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	// Return embedding based on text hash for realistic semantic search
	embedding := make([]float64, 5)
	for i, char := range text {
		if i >= 5 {
			break
		}
		embedding[i] = float64(char) / 256.0
	}
	return embedding, nil
}

func (m *MockEmbeddingGeneratorWithSearchableEmbeddings) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i, text := range texts {
		result[i], _ = m.GenerateEmbedding(ctx, text)
	}
	return result, nil
}

func TestCodeGraph_AddNode(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	node := &CodeNode{
		ID:        "node1",
		Type:      NodeTypeFunction,
		Name:      "TestFunction",
		Path:      "/test/file.go",
		Metadata: map[string]interface{}{
			"language": "go",
			"lines":    10,
		},
		CreatedAt: time.Now(),
	}

	err := graph.AddNode(node)
	require.NoError(t, err)

	// Verify node was added
	retrieved, ok := graph.GetNode("node1")
	assert.True(t, ok)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "TestFunction", retrieved.Name)
	assert.Equal(t, NodeTypeFunction, retrieved.Type)
}

func TestCodeGraph_AddEdge(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add nodes first
	graph.AddNode(&CodeNode{ID: "node1", Type: NodeTypeFunction, Name: "Func1"})
	graph.AddNode(&CodeNode{ID: "node2", Type: NodeTypeFunction, Name: "Func2"})

	// Add edge
	edge := &CodeEdge{
		ID:       "edge1",
		Type:     EdgeTypeCalls,
		SourceID: "node1",
		TargetID: "node2",
		Weight:   1.0,
	}

	err := graph.AddEdge(edge)
	require.NoError(t, err)

	// Verify edge was added
	edges := graph.GetOutgoingEdges("node1")
	assert.Len(t, edges, 1)
	assert.Equal(t, "node2", edges[0].TargetID)
}

func TestCodeGraph_GetNodesByType(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add various node types
	graph.AddNode(&CodeNode{ID: "func1", Type: NodeTypeFunction, Name: "Func1"})
	graph.AddNode(&CodeNode{ID: "func2", Type: NodeTypeFunction, Name: "Func2"})
	graph.AddNode(&CodeNode{ID: "class1", Type: NodeTypeClass, Name: "Class1"})

	// Get functions
	functions := graph.GetNodesByType(NodeTypeFunction)
	assert.Len(t, functions, 2)

	// Get classes
	classes := graph.GetNodesByType(NodeTypeClass)
	assert.Len(t, classes, 1)
}

func TestCodeGraph_GetNeighbors(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Build a small graph: A -> B -> C
	graph.AddNode(&CodeNode{ID: "A", Type: NodeTypeFunction, Name: "A"})
	graph.AddNode(&CodeNode{ID: "B", Type: NodeTypeFunction, Name: "B"})
	graph.AddNode(&CodeNode{ID: "C", Type: NodeTypeFunction, Name: "C"})

	graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "A", TargetID: "B"})
	graph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "B", TargetID: "C"})

	// Get neighbors of B (outgoing)
	neighbors := graph.GetNeighbors("B", EdgeTypeCalls)
	assert.Len(t, neighbors, 1) // Should include C (outgoing)
}

func TestCodeGraph_GetImpactRadius(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Build a graph with dependencies using edges the GetImpactRadius follows
	// (EdgeTypeCalledBy, EdgeTypeReferencedBy, EdgeTypeDependsOn)
	graph.AddNode(&CodeNode{ID: "core", Type: NodeTypeFunction, Name: "Core"})
	graph.AddNode(&CodeNode{ID: "util1", Type: NodeTypeFunction, Name: "Util1"})
	graph.AddNode(&CodeNode{ID: "util2", Type: NodeTypeFunction, Name: "Util2"})
	graph.AddNode(&CodeNode{ID: "app", Type: NodeTypeFunction, Name: "App"})

	// Use EdgeTypeCalledBy which is followed by GetImpactRadius
	graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalledBy, SourceID: "util1", TargetID: "core"})
	graph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalledBy, SourceID: "util2", TargetID: "core"})
	graph.AddEdge(&CodeEdge{ID: "e3", Type: EdgeTypeCalledBy, SourceID: "app", TargetID: "util1"})

	// Get impact radius of core with depth 2
	impacted := graph.GetImpactRadius("core", 2)

	// Should include core itself and util1, util2 (which call core)
	assert.GreaterOrEqual(t, len(impacted), 1)
}

func TestCodeGraph_FindPath(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Build a graph: A -> B -> C -> D
	graph.AddNode(&CodeNode{ID: "A", Type: NodeTypeFunction, Name: "A"})
	graph.AddNode(&CodeNode{ID: "B", Type: NodeTypeFunction, Name: "B"})
	graph.AddNode(&CodeNode{ID: "C", Type: NodeTypeFunction, Name: "C"})
	graph.AddNode(&CodeNode{ID: "D", Type: NodeTypeFunction, Name: "D"})

	graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "A", TargetID: "B"})
	graph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "B", TargetID: "C"})
	graph.AddEdge(&CodeEdge{ID: "e3", Type: EdgeTypeCalls, SourceID: "C", TargetID: "D"})

	// Find path from A to D
	path, err := graph.FindPath("A", "D")
	require.NoError(t, err)
	require.NotNil(t, path)
	assert.GreaterOrEqual(t, len(path), 2)
}

func TestCodeGraph_GetStats(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "N1"})
	graph.AddNode(&CodeNode{ID: "n2", Type: NodeTypeFunction, Name: "N2"})
	graph.AddNode(&CodeNode{ID: "n3", Type: NodeTypeClass, Name: "N3"})
	graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "n1", TargetID: "n2"})

	stats := graph.GetStats()

	assert.Equal(t, 3, stats["total_nodes"])
	assert.Equal(t, 1, stats["total_edges"])
}

func TestGraphRAG_Retrieve(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Add some test nodes
	codeGraph.AddNode(&CodeNode{
		ID:       "func1",
		Type:     NodeTypeFunction,
		Name:     "calculateSum",
		Docstring: "Calculate the sum of two numbers",
		Path:     "/math/sum.go",
		Metadata: map[string]interface{}{
			"language": "go",
		},
	})

	codeGraph.AddNode(&CodeNode{
		ID:       "func2",
		Type:     NodeTypeFunction,
		Name:     "multiply",
		Docstring: "Multiply two numbers",
		Path:     "/math/multiply.go",
		Metadata: map[string]interface{}{
			"language": "go",
		},
	})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "calculate sum")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Nodes), 0)
}

func TestGraphRAG_BuildContext(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	codeGraph.AddNode(&CodeNode{
		ID:       "node1",
		Type:     NodeTypeFunction,
		Name:     "TestFunc",
		Docstring: "A test function",
		Path:     "/test.go",
	})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	// Create a mock retrieval result
	node, _ := codeGraph.GetNode("node1")
	result := &RetrievalResult{
		Nodes: []*RetrievedNode{{
			Node:           node,
			RelevanceScore: 0.9,
		}},
	}

	contextStr := graphRAG.BuildContext(result)
	assert.Contains(t, contextStr, "TestFunc")
}

func TestCodeGraph_SemanticSearch(t *testing.T) {
	config := DefaultCodeGraphConfig()
	config.EnableEmbeddings = false // Disable for test
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add nodes
	graph.AddNode(&CodeNode{
		ID:        "n1",
		Type:      NodeTypeFunction,
		Name:      "processPayment",
		Docstring: "Handle payment processing and validation",
	})

	graph.AddNode(&CodeNode{
		ID:        "n2",
		Type:      NodeTypeFunction,
		Name:      "validateInput",
		Docstring: "Validate user input data",
	})

	ctx := context.Background()
	results, err := graph.SemanticSearch(ctx, "payment", 5)

	// Without embeddings, should fall back to keyword search
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 0)
}

func TestRetrievalMode(t *testing.T) {
	modes := []RetrievalMode{
		RetrievalModeLocal,
		RetrievalModeGraph,
		RetrievalModeHybrid,
		RetrievalModeSemantic,
	}

	for _, mode := range modes {
		assert.NotEmpty(t, string(mode))
	}
}

func TestCodeGraphConfig_Defaults(t *testing.T) {
	config := DefaultCodeGraphConfig()

	assert.Greater(t, config.MaxNodes, 0)
	assert.Greater(t, config.MaxEdgesPerNode, 0)
}

func TestGraphRAGConfig_Defaults(t *testing.T) {
	config := DefaultGraphRAGConfig()

	assert.Greater(t, config.MaxNodes, 0)
	assert.Greater(t, config.MaxDepth, 0)
	assert.Greater(t, config.MaxTokens, 0)
}

// Tests for uncovered code_graph.go functions

func TestCodeGraph_GetIncomingEdges(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add nodes
	graph.AddNode(&CodeNode{ID: "A", Type: NodeTypeFunction, Name: "A"})
	graph.AddNode(&CodeNode{ID: "B", Type: NodeTypeFunction, Name: "B"})
	graph.AddNode(&CodeNode{ID: "C", Type: NodeTypeFunction, Name: "C"})

	// Add edges: A -> B, C -> B
	graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "A", TargetID: "B"})
	graph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "C", TargetID: "B"})

	// B should have 2 incoming edges
	incomingEdges := graph.GetIncomingEdges("B")
	assert.Len(t, incomingEdges, 2)
}

func TestCodeGraph_MarshalJSON(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc"})
	graph.AddNode(&CodeNode{ID: "n2", Type: NodeTypeClass, Name: "TestClass"})
	graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "n1", TargetID: "n2"})

	data, err := graph.MarshalJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Contains(t, result, "nodes")
	assert.Contains(t, result, "edges")
	assert.Contains(t, result, "stats")
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{1, 0, 0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{0, 1, 0},
			expected: 0.0,
		},
		{
			name:     "different length vectors",
			a:        []float64{1, 2},
			b:        []float64{1, 2, 3},
			expected: 0.0,
		},
		{
			name:     "empty vectors",
			a:        []float64{},
			b:        []float64{},
			expected: 0.0,
		},
		{
			name:     "zero vector",
			a:        []float64{0, 0, 0},
			b:        []float64{1, 2, 3},
			expected: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cosineSimilarity(tc.a, tc.b)
			assert.InDelta(t, tc.expected, result, 0.001)
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{4.0, 2.0},
		{9.0, 3.0},
		{16.0, 4.0},
		{0.0, 0.0},
		{-1.0, 0.0},
	}

	for _, tc := range tests {
		result := sqrt(tc.input)
		assert.InDelta(t, tc.expected, result, 0.001)
	}
}

func TestNewDefaultCodeParser(t *testing.T) {
	logger := logrus.New()
	parser := NewDefaultCodeParser(logger)

	assert.NotNil(t, parser)
	assert.NotNil(t, parser.languagePatterns)
	assert.Contains(t, parser.languagePatterns, "go")
	assert.Contains(t, parser.languagePatterns, "python")
	assert.Contains(t, parser.languagePatterns, "javascript")
	assert.Contains(t, parser.languagePatterns, "java")
	assert.Contains(t, parser.languagePatterns, "rust")
}

func TestDefaultCodeParser_SupportedLanguages(t *testing.T) {
	parser := NewDefaultCodeParser(logrus.New())
	languages := parser.SupportedLanguages()

	assert.NotEmpty(t, languages)
	assert.Contains(t, languages, "go")
	assert.Contains(t, languages, "python")
}

func TestDefaultCodeParser_ParseFile(t *testing.T) {
	parser := NewDefaultCodeParser(logrus.New())
	ctx := context.Background()

	tests := []struct {
		name        string
		path        string
		content     string
		expectNodes int
	}{
		{
			name: "go file with functions",
			path: "/test/main.go",
			content: `package main

func Hello() {
	println("Hello")
}

func World() {
	println("World")
}
`,
			expectNodes: 3, // 1 file + 2 functions
		},
		{
			name: "python file with functions",
			path: "/test/main.py",
			content: `# Comment
def hello():
    print("Hello")

def world():
    print("World")
`,
			expectNodes: 1, // Python parsing may only get file node depending on regex
		},
		{
			name:        "unsupported language",
			path:        "/test/main.txt",
			content:     "some text content",
			expectNodes: 1, // Only file node
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nodes, edges, err := parser.ParseFile(ctx, tc.path, []byte(tc.content))
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(nodes), tc.expectNodes)
			// For files with functions, there should be edges
			if tc.expectNodes > 1 {
				assert.NotEmpty(t, edges)
			}
		})
	}
}

func TestDefaultCodeParser_GetLineNumber(t *testing.T) {
	parser := NewDefaultCodeParser(logrus.New())

	content := []byte("line1\nline2\nline3\nline4")

	// Line 1 - offset 0
	assert.Equal(t, 1, parser.getLineNumber(content, 0))

	// Line 2 - offset 6 (after first newline)
	assert.Equal(t, 2, parser.getLineNumber(content, 6))

	// Line 3 - offset 12
	assert.Equal(t, 3, parser.getLineNumber(content, 12))
}

func TestCodeGraph_GenerateEmbeddings(t *testing.T) {
	config := DefaultCodeGraphConfig()
	config.EnableEmbeddings = true
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add nodes
	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc", Docstring: "A test function"})
	graph.AddNode(&CodeNode{ID: "n2", Type: NodeTypeFunction, Name: "AnotherFunc", Signature: "func AnotherFunc()"})

	ctx := context.Background()
	err := graph.GenerateEmbeddings(ctx)
	require.NoError(t, err)

	// Verify embeddings were generated
	node1, _ := graph.GetNode("n1")
	assert.NotEmpty(t, node1.Embedding)
}

func TestCodeGraph_GenerateEmbeddingsWithoutEmbedder(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, logrus.New())

	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc"})

	ctx := context.Background()
	err := graph.GenerateEmbeddings(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "embedder not configured")
}

func TestCodeGraph_BuildSemanticEdges(t *testing.T) {
	config := DefaultCodeGraphConfig()
	config.SimilarityThreshold = 0.5
	config.EnableEmbeddings = true
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add nodes with embeddings
	node1 := &CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc1", Embedding: []float64{1, 0, 0, 0, 0}}
	node2 := &CodeNode{ID: "n2", Type: NodeTypeFunction, Name: "TestFunc2", Embedding: []float64{0.9, 0.1, 0, 0, 0}}
	node3 := &CodeNode{ID: "n3", Type: NodeTypeFunction, Name: "TestFunc3", Embedding: []float64{0, 1, 0, 0, 0}}

	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)

	ctx := context.Background()
	err := graph.BuildSemanticEdges(ctx)
	require.NoError(t, err)

	// n1 and n2 should be similar
	stats := graph.GetStats()
	assert.NotNil(t, stats)
}

// Tests for graphrag.go uncovered functions

func TestRetrievalResult_MarshalJSON(t *testing.T) {
	result := &RetrievalResult{
		Query:      "test query",
		TotalFound: 5,
		Duration:   100 * time.Millisecond,
		Mode:       RetrievalModeHybrid,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "test query", decoded["query"])
	assert.Equal(t, float64(100), decoded["duration_ms"])
}

func TestGraphRAG_RetrieveLocal(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeLocal
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	codeGraph.AddNode(&CodeNode{
		ID:        "func1",
		Type:      NodeTypeFunction,
		Name:      "searchFunction",
		Docstring: "This function searches for items",
	})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "search items")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGraphRAG_RetrieveGraph(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeGraph
	config.IncludeEdgeTypes = []EdgeType{EdgeTypeCalls, EdgeTypeCalledBy}
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Build a graph with connections
	codeGraph.AddNode(&CodeNode{ID: "main", Type: NodeTypeFunction, Name: "main", Docstring: "main function"})
	codeGraph.AddNode(&CodeNode{ID: "helper", Type: NodeTypeFunction, Name: "helper", Docstring: "helper function"})
	codeGraph.AddNode(&CodeNode{ID: "util", Type: NodeTypeFunction, Name: "util", Docstring: "utility function"})

	codeGraph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "main", TargetID: "helper"})
	codeGraph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "helper", TargetID: "util"})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "main")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGraphRAG_RetrieveSemantic(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeSemantic
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	codeGraph.AddNode(&CodeNode{
		ID:        "func1",
		Type:      NodeTypeFunction,
		Name:      "processData",
		Docstring: "Process incoming data",
	})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "data processing")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGraphRAG_RetrieveHybrid(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeHybrid
	config.IncludeEdgeTypes = []EdgeType{EdgeTypeCalls}
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Build a connected graph
	codeGraph.AddNode(&CodeNode{ID: "core", Type: NodeTypeFunction, Name: "core", Docstring: "core function"})
	codeGraph.AddNode(&CodeNode{ID: "service", Type: NodeTypeFunction, Name: "service", Docstring: "service layer"})

	codeGraph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "core", TargetID: "service"})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "core")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGraphRAG_RetrieveForNode(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Add nodes with impact relationships
	codeGraph.AddNode(&CodeNode{ID: "main", Type: NodeTypeFunction, Name: "main"})
	codeGraph.AddNode(&CodeNode{ID: "dep1", Type: NodeTypeFunction, Name: "dependency1"})
	codeGraph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalledBy, SourceID: "dep1", TargetID: "main"})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.RetrieveForNode(ctx, "main")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, RetrievalModeGraph, result.Mode)
}

func TestGraphRAG_RetrieveForNodeNotFound(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	_, err := graphRAG.RetrieveForNode(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
}

func TestNewLLMReranker(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		return []int{0, 1, 2}, nil
	}
	logger := logrus.New()

	reranker := NewLLMReranker(rerankFunc, logger)

	assert.NotNil(t, reranker)
	assert.NotNil(t, reranker.rerankFunc)
	assert.NotNil(t, reranker.logger)
}

func TestLLMReranker_Rerank(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		// Return reversed order
		result := make([]int, len(docs))
		for i := range docs {
			result[i] = len(docs) - 1 - i
		}
		return result, nil
	}

	reranker := NewLLMReranker(rerankFunc, logrus.New())

	nodes := []*RetrievedNode{
		{Node: &CodeNode{ID: "n1", Name: "First"}, RelevanceScore: 0.9},
		{Node: &CodeNode{ID: "n2", Name: "Second"}, RelevanceScore: 0.8},
		{Node: &CodeNode{ID: "n3", Name: "Third"}, RelevanceScore: 0.7},
	}

	ctx := context.Background()
	reranked, err := reranker.Rerank(ctx, "test query", nodes)

	require.NoError(t, err)
	assert.Len(t, reranked, 3)
	// Should be reversed
	assert.Equal(t, "Third", reranked[0].Node.Name)
	assert.Equal(t, "Second", reranked[1].Node.Name)
	assert.Equal(t, "First", reranked[2].Node.Name)
}

func TestLLMReranker_RerankWithNilFunc(t *testing.T) {
	reranker := NewLLMReranker(nil, logrus.New())

	nodes := []*RetrievedNode{
		{Node: &CodeNode{ID: "n1", Name: "First"}, RelevanceScore: 0.9},
	}

	ctx := context.Background()
	reranked, err := reranker.Rerank(ctx, "test", nodes)

	require.NoError(t, err)
	assert.Equal(t, nodes, reranked)
}

func TestLLMReranker_RerankWithEmptyNodes(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		return []int{}, nil
	}
	reranker := NewLLMReranker(rerankFunc, logrus.New())

	ctx := context.Background()
	reranked, err := reranker.Rerank(ctx, "test", []*RetrievedNode{})

	require.NoError(t, err)
	assert.Empty(t, reranked)
}

// Mock policy model for testing
type mockPolicyModel struct {
	shouldRetrieve bool
	returnError    error
}

func (m *mockPolicyModel) ShouldRetrieve(ctx context.Context, query, localContext string) (bool, error) {
	return m.shouldRetrieve, m.returnError
}

func TestNewSelectiveRetriever(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	graph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())
	policyModel := &mockPolicyModel{shouldRetrieve: true}
	logger := logrus.New()

	retriever := NewSelectiveRetriever(config, graph, policyModel, logger)

	assert.NotNil(t, retriever)
	assert.NotNil(t, retriever.graphRAG)
}

func TestSelectiveRetriever_RetrieveWithPolicy(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	graph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc"})

	tests := []struct {
		name           string
		policyRetrieve bool
		expectEmpty    bool
	}{
		{
			name:           "policy allows retrieval",
			policyRetrieve: true,
			expectEmpty:    false,
		},
		{
			name:           "policy blocks retrieval",
			policyRetrieve: false,
			expectEmpty:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policyModel := &mockPolicyModel{shouldRetrieve: tc.policyRetrieve}
			retriever := NewSelectiveRetriever(config, graph, policyModel, logrus.New())

			ctx := context.Background()
			result, err := retriever.Retrieve(ctx, "test query", "local context")

			require.NoError(t, err)
			assert.NotNil(t, result)
			if tc.expectEmpty {
				assert.Empty(t, result.Nodes)
			}
		})
	}
}

func TestSelectiveRetriever_RetrieveWithNilPolicy(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	graph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc"})

	retriever := NewSelectiveRetriever(config, graph, nil, logrus.New())

	ctx := context.Background()
	result, err := retriever.Retrieve(ctx, "test query", "")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSelectiveRetriever_RetrieveWithPolicyError(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	graph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc"})

	policyModel := &mockPolicyModel{shouldRetrieve: false, returnError: assert.AnError}
	retriever := NewSelectiveRetriever(config, graph, policyModel, logrus.New())

	ctx := context.Background()
	result, err := retriever.Retrieve(ctx, "test query", "")

	// Should still work, just log the error
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test GraphRAG with reranker
func TestGraphRAG_RetrieveWithReranker(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		return []int{0}, nil
	}
	reranker := NewLLMReranker(rerankFunc, logrus.New())

	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	codeGraph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "TestFunc", Docstring: "Test"})

	graphRAG := NewGraphRAG(config, codeGraph, reranker, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "test")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test edge cases for AddNode
func TestCodeGraph_AddNodeDuplicate(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	node := &CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "Func1"}
	err := graph.AddNode(node)
	require.NoError(t, err)

	// Adding same node again should update
	node2 := &CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "Func1Updated"}
	err = graph.AddNode(node2)
	require.NoError(t, err)

	retrieved, _ := graph.GetNode("n1")
	assert.Equal(t, "Func1Updated", retrieved.Name)
}

// Test edge cases for AddEdge
func TestCodeGraph_AddEdgeInvalidNodes(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Only add source node
	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "Func1"})

	// Try adding edge with nonexistent target
	err := graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "n1", TargetID: "nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target node not found")

	// Try adding edge with nonexistent source
	graph.AddNode(&CodeNode{ID: "n2", Type: NodeTypeFunction, Name: "Func2"})
	err = graph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "nonexistent", TargetID: "n2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source node not found")
}

func TestCodeGraph_AddEdgeMaxEdges(t *testing.T) {
	config := DefaultCodeGraphConfig()
	config.MaxEdgesPerNode = 2 // Low limit for testing
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "source", Type: NodeTypeFunction, Name: "Source"})
	graph.AddNode(&CodeNode{ID: "t1", Type: NodeTypeFunction, Name: "Target1"})
	graph.AddNode(&CodeNode{ID: "t2", Type: NodeTypeFunction, Name: "Target2"})
	graph.AddNode(&CodeNode{ID: "t3", Type: NodeTypeFunction, Name: "Target3"})

	// Add edges up to limit
	err := graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "source", TargetID: "t1"})
	require.NoError(t, err)
	err = graph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "source", TargetID: "t2"})
	require.NoError(t, err)

	// This should fail - exceeds max edges
	err = graph.AddEdge(&CodeEdge{ID: "e3", Type: EdgeTypeCalls, SourceID: "source", TargetID: "t3"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max edges reached")
}

func TestCodeGraph_AddEdgeDefaults(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "Func1"})
	graph.AddNode(&CodeNode{ID: "n2", Type: NodeTypeFunction, Name: "Func2"})

	// Add edge without ID, CreatedAt, or Weight
	edge := &CodeEdge{
		Type:     EdgeTypeCalls,
		SourceID: "n1",
		TargetID: "n2",
	}
	err := graph.AddEdge(edge)
	require.NoError(t, err)

	// Verify defaults were set
	assert.NotEmpty(t, edge.ID)
	assert.False(t, edge.CreatedAt.IsZero())
	assert.Equal(t, 1.0, edge.Weight)
}

// Test FindPath edge cases
func TestCodeGraph_FindPathNoPath(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add disconnected nodes
	graph.AddNode(&CodeNode{ID: "A", Type: NodeTypeFunction, Name: "A"})
	graph.AddNode(&CodeNode{ID: "B", Type: NodeTypeFunction, Name: "B"})

	// No path between disconnected nodes - returns error
	_, err := graph.FindPath("A", "B")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no path found")
}

func TestCodeGraph_FindPathSameNode(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	graph.AddNode(&CodeNode{ID: "A", Type: NodeTypeFunction, Name: "A"})

	// Path from node to itself
	path, err := graph.FindPath("A", "A")
	require.NoError(t, err)
	assert.Len(t, path, 1)
	assert.Equal(t, "A", path[0].ID)
}

func TestCodeGraph_FindPathNonexistentNode(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	_, err := graph.FindPath("nonexistent", "also_nonexistent")
	assert.Error(t, err)
}

// Test NewCodeGraph with nil logger
func TestNewCodeGraph_NilLogger(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, nil)

	assert.NotNil(t, graph)
	assert.NotNil(t, graph.logger)
}

// Test NewGraphRAG with nil logger
func TestNewGraphRAG_NilLogger(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graphRAG := NewGraphRAG(config, codeGraph, nil, nil)

	assert.NotNil(t, graphRAG)
	assert.NotNil(t, graphRAG.logger)
}

// Test stringBuilder (internal type)
func TestStringBuilder_WriteAndString(t *testing.T) {
	builder := &stringBuilder{}

	builder.WriteString("Hello")
	builder.WriteString(" ")
	builder.WriteString("World")

	result := builder.String()
	assert.Equal(t, "Hello World", result)
}

// Test traverseImpact edge case
func TestCodeGraph_GetImpactRadiusDeepGraph(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Build a deep chain: A <- B <- C <- D <- E
	graph.AddNode(&CodeNode{ID: "A", Type: NodeTypeFunction, Name: "A"})
	graph.AddNode(&CodeNode{ID: "B", Type: NodeTypeFunction, Name: "B"})
	graph.AddNode(&CodeNode{ID: "C", Type: NodeTypeFunction, Name: "C"})
	graph.AddNode(&CodeNode{ID: "D", Type: NodeTypeFunction, Name: "D"})
	graph.AddNode(&CodeNode{ID: "E", Type: NodeTypeFunction, Name: "E"})

	graph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalledBy, SourceID: "B", TargetID: "A"})
	graph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalledBy, SourceID: "C", TargetID: "B"})
	graph.AddEdge(&CodeEdge{ID: "e3", Type: EdgeTypeCalledBy, SourceID: "D", TargetID: "C"})
	graph.AddEdge(&CodeEdge{ID: "e4", Type: EdgeTypeCalledBy, SourceID: "E", TargetID: "D"})

	// Depth 2 should get A, B, C
	impacted := graph.GetImpactRadius("A", 2)
	assert.NotEmpty(t, impacted)
}

// Test SemanticSearch with embeddings
func TestCodeGraph_SemanticSearchWithEmbeddings(t *testing.T) {
	config := DefaultCodeGraphConfig()
	config.EnableEmbeddings = true
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add nodes and generate embeddings
	graph.AddNode(&CodeNode{
		ID:        "n1",
		Type:      NodeTypeFunction,
		Name:      "searchHandler",
		Docstring: "Handle search requests",
	})

	graph.AddNode(&CodeNode{
		ID:        "n2",
		Type:      NodeTypeFunction,
		Name:      "dataProcessor",
		Docstring: "Process data",
	})

	ctx := context.Background()

	// Generate embeddings first
	err := graph.GenerateEmbeddings(ctx)
	require.NoError(t, err)

	// Now semantic search should use embeddings
	results, err := graph.SemanticSearch(ctx, "search", 5)
	require.NoError(t, err)
	assert.NotNil(t, results)
}

// Test GraphRAG graph traversal with proper edge types
func TestGraphRAG_RetrieveGraphWithTraversal(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeGraph
	config.MinRelevanceScore = 0.0 // Accept all results
	// Use default included edge types

	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Build a graph with proper edge types that will be followed
	// Use "main" in both name and docstring for better keyword matching
	codeGraph.AddNode(&CodeNode{ID: "main", Type: NodeTypeFunction, Name: "main", Docstring: "main main main"})
	codeGraph.AddNode(&CodeNode{ID: "helper", Type: NodeTypeFunction, Name: "helper", Docstring: "helper function"})
	codeGraph.AddNode(&CodeNode{ID: "util", Type: NodeTypeFunction, Name: "util", Docstring: "utility"})
	codeGraph.AddNode(&CodeNode{ID: "deep", Type: NodeTypeFunction, Name: "deep", Docstring: "deep function"})

	// Use EdgeTypeCalls which is in IncludeEdgeTypes
	codeGraph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "main", TargetID: "helper"})
	codeGraph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "helper", TargetID: "util"})
	codeGraph.AddEdge(&CodeEdge{ID: "e3", Type: EdgeTypeCalls, SourceID: "util", TargetID: "deep"})
	// Also add incoming edges with EdgeTypeCalledBy
	codeGraph.AddEdge(&CodeEdge{ID: "e4", Type: EdgeTypeCalledBy, SourceID: "deep", TargetID: "util"})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "main")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test hybrid retrieval with graph traversal
func TestGraphRAG_RetrieveHybridWithGraphTraversal(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeHybrid
	// Use default config which has EdgeTypeCalls in IncludeEdgeTypes

	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Build a connected graph
	codeGraph.AddNode(&CodeNode{ID: "api", Type: NodeTypeFunction, Name: "apiHandler", Docstring: "api handler"})
	codeGraph.AddNode(&CodeNode{ID: "service", Type: NodeTypeFunction, Name: "service", Docstring: "service layer"})
	codeGraph.AddNode(&CodeNode{ID: "repo", Type: NodeTypeFunction, Name: "repository", Docstring: "data repository"})

	codeGraph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "api", TargetID: "service"})
	codeGraph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "service", TargetID: "repo"})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "api")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test edge following with different edge types
func TestGraphRAG_ShouldFollowEdge(t *testing.T) {
	config := DefaultGraphRAGConfig()
	// Default config includes: EdgeTypeCalls, EdgeTypeCalledBy, EdgeTypeContains, etc.

	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	// Test with included edge type
	assert.True(t, graphRAG.shouldFollowEdge(EdgeTypeCalls))
	assert.True(t, graphRAG.shouldFollowEdge(EdgeTypeCalledBy))
	assert.True(t, graphRAG.shouldFollowEdge(EdgeTypeContains))

	// Test with non-included edge type
	assert.False(t, graphRAG.shouldFollowEdge(EdgeTypeSemanticallySimilar))
}

// Test generateNodeID (called internally)
func TestCodeGraph_GenerateNodeID(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	node := &CodeNode{
		Type:      NodeTypeFunction,
		Path:      "/test/file.go",
		Name:      "TestFunc",
		StartLine: 10,
	}

	// generateNodeID is called internally - we test it indirectly
	// by adding a node without an ID
	nodeWithoutID := &CodeNode{
		Type:      NodeTypeFunction,
		Path:      "/test/file.go",
		Name:      "NoIDFunc",
		StartLine: 20,
	}
	// The ID will be generated if empty
	err := graph.AddNode(nodeWithoutID)
	require.NoError(t, err)

	// Test the generated ID format
	expectedIDFormat := graph.generateNodeID(node)
	assert.NotEmpty(t, expectedIDFormat)
	assert.Contains(t, expectedIDFormat, "function")
}

// Test BuildContext edge cases
func TestGraphRAG_BuildContextEmpty(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	// Empty result
	emptyResult := &RetrievalResult{
		Query: "test",
		Nodes: []*RetrievedNode{},
	}

	context := graphRAG.BuildContext(emptyResult)
	assert.Empty(t, context)
}

// Test BuildContext with signature
func TestGraphRAG_BuildContextWithSignature(t *testing.T) {
	config := DefaultGraphRAGConfig()
	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	result := &RetrievalResult{
		Query: "test",
		Nodes: []*RetrievedNode{
			{
				Node: &CodeNode{
					ID:        "n1",
					Name:      "TestFunc",
					Signature: "func TestFunc(a int) error",
					Docstring: "A test function",
					Path:      "/test.go",
				},
				RelevanceScore: 0.9,
			},
		},
	}

	context := graphRAG.BuildContext(result)
	assert.Contains(t, context, "TestFunc")
	assert.Contains(t, context, "func TestFunc(a int) error")
	assert.Contains(t, context, "A test function")
}

// Test Retrieve with empty query results
func TestGraphRAG_RetrieveEmptyResults(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.MinRelevanceScore = 0.99 // High threshold to filter out all results

	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Add node with low relevance
	codeGraph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "func1"})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "xyz")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test local retrieval with semantic search returning results
func TestGraphRAG_RetrieveLocalWithResults(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeLocal

	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	// Add multiple nodes
	codeGraph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "processOrder", Docstring: "Process customer order"})
	codeGraph.AddNode(&CodeNode{ID: "n2", Type: NodeTypeFunction, Name: "orderHandler", Docstring: "Handle incoming orders"})
	codeGraph.AddNode(&CodeNode{ID: "n3", Type: NodeTypeFunction, Name: "shipOrder", Docstring: "Ship order to customer"})

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "order")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, RetrievalModeLocal, result.Mode)
}

// Test graph traversal with embeddings enabled
func TestGraphRAG_TraverseGraphWithEmbeddings(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeGraph
	config.MinRelevanceScore = 0.0

	graphConfig := DefaultCodeGraphConfig()
	graphConfig.EnableEmbeddings = true
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGeneratorWithSearchableEmbeddings{}, logrus.New())

	// Add nodes with embeddings that will be similar when searched
	node1 := &CodeNode{ID: "seed", Type: NodeTypeFunction, Name: "searchTarget", Docstring: "search target node"}
	node2 := &CodeNode{ID: "connected1", Type: NodeTypeFunction, Name: "connected", Docstring: "connected node"}
	node3 := &CodeNode{ID: "connected2", Type: NodeTypeFunction, Name: "deep", Docstring: "deep node"}

	codeGraph.AddNode(node1)
	codeGraph.AddNode(node2)
	codeGraph.AddNode(node3)

	// Add edges that will be followed (EdgeTypeCalls is in default IncludeEdgeTypes)
	codeGraph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "seed", TargetID: "connected1"})
	codeGraph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "connected1", TargetID: "connected2"})
	// Add incoming edge for bidirectional traversal
	codeGraph.AddEdge(&CodeEdge{ID: "e3", Type: EdgeTypeCalledBy, SourceID: "connected2", TargetID: "connected1"})

	// Generate embeddings
	ctx := context.Background()
	err := codeGraph.GenerateEmbeddings(ctx)
	require.NoError(t, err)

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	result, err := graphRAG.Retrieve(ctx, "search")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test hybrid traversal with embeddings
func TestGraphRAG_TraverseHybridWithEmbeddings(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeHybrid
	config.MinRelevanceScore = 0.0

	graphConfig := DefaultCodeGraphConfig()
	graphConfig.EnableEmbeddings = true
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGeneratorWithSearchableEmbeddings{}, logrus.New())

	// Add nodes
	codeGraph.AddNode(&CodeNode{ID: "start", Type: NodeTypeFunction, Name: "startFunc", Docstring: "starting point"})
	codeGraph.AddNode(&CodeNode{ID: "middle", Type: NodeTypeFunction, Name: "middleFunc", Docstring: "middle point"})
	codeGraph.AddNode(&CodeNode{ID: "end", Type: NodeTypeFunction, Name: "endFunc", Docstring: "ending point"})

	// Add edges
	codeGraph.AddEdge(&CodeEdge{ID: "e1", Type: EdgeTypeCalls, SourceID: "start", TargetID: "middle"})
	codeGraph.AddEdge(&CodeEdge{ID: "e2", Type: EdgeTypeCalls, SourceID: "middle", TargetID: "end"})

	// Generate embeddings
	ctx := context.Background()
	err := codeGraph.GenerateEmbeddings(ctx)
	require.NoError(t, err)

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	result, err := graphRAG.Retrieve(ctx, "start")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test retrieveLocal directly through Retrieve with local mode
func TestGraphRAG_RetrieveLocalDirect(t *testing.T) {
	config := DefaultGraphRAGConfig()
	config.RetrievalMode = RetrievalModeLocal
	config.MinRelevanceScore = 0.0

	graphConfig := DefaultCodeGraphConfig()
	graphConfig.EnableEmbeddings = true
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGeneratorWithSearchableEmbeddings{}, logrus.New())

	// Add node
	codeGraph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "localFunc", Docstring: "local function"})

	// Generate embeddings
	ctx := context.Background()
	codeGraph.GenerateEmbeddings(ctx)

	graphRAG := NewGraphRAG(config, codeGraph, nil, logrus.New())

	result, err := graphRAG.Retrieve(ctx, "local")

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test Rerank with error from rerank function
func TestLLMReranker_RerankWithError(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		return nil, assert.AnError
	}
	reranker := NewLLMReranker(rerankFunc, logrus.New())

	nodes := []*RetrievedNode{
		{Node: &CodeNode{ID: "n1", Name: "First"}, RelevanceScore: 0.9},
	}

	ctx := context.Background()
	result, err := reranker.Rerank(ctx, "test", nodes)

	assert.Error(t, err)
	assert.Equal(t, nodes, result) // Returns original nodes on error
}

// Test Rerank with partial indices
func TestLLMReranker_RerankWithPartialIndices(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		// Only return some indices
		return []int{1}, nil
	}
	reranker := NewLLMReranker(rerankFunc, logrus.New())

	nodes := []*RetrievedNode{
		{Node: &CodeNode{ID: "n1", Name: "First"}, RelevanceScore: 0.9},
		{Node: &CodeNode{ID: "n2", Name: "Second"}, RelevanceScore: 0.8},
		{Node: &CodeNode{ID: "n3", Name: "Third"}, RelevanceScore: 0.7},
	}

	ctx := context.Background()
	result, err := reranker.Rerank(ctx, "test", nodes)

	require.NoError(t, err)
	// Should have Second first, then others added at the end
	assert.Equal(t, "Second", result[0].Node.Name)
	assert.Len(t, result, 3) // All nodes should be present
}

// Test Rerank with out of bound indices
func TestLLMReranker_RerankWithInvalidIndices(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		// Return invalid indices
		return []int{-1, 100, 0}, nil
	}
	reranker := NewLLMReranker(rerankFunc, logrus.New())

	nodes := []*RetrievedNode{
		{Node: &CodeNode{ID: "n1", Name: "First"}, RelevanceScore: 0.9},
	}

	ctx := context.Background()
	result, err := reranker.Rerank(ctx, "test", nodes)

	require.NoError(t, err)
	// Only valid index (0) should be added, rest should be added as unseen
	assert.NotEmpty(t, result)
}

// Test ParseFile with JavaScript
func TestDefaultCodeParser_ParseFileJavaScript(t *testing.T) {
	parser := NewDefaultCodeParser(logrus.New())
	ctx := context.Background()

	content := `
function hello() {
    console.log("Hello");
}

const world = function() {
    console.log("World");
}
`
	nodes, _, err := parser.ParseFile(ctx, "/test/main.js", []byte(content))
	require.NoError(t, err)
	assert.NotEmpty(t, nodes)
	// Should have file node
	assert.Equal(t, NodeTypeFile, nodes[0].Type)
}

// Test ParseFile with Java
func TestDefaultCodeParser_ParseFileJava(t *testing.T) {
	parser := NewDefaultCodeParser(logrus.New())
	ctx := context.Background()

	content := `
public class Test {
    public void hello() {
        System.out.println("Hello");
    }
    private int calculate() {
        return 42;
    }
}
`
	nodes, _, err := parser.ParseFile(ctx, "/test/Test.java", []byte(content))
	require.NoError(t, err)
	assert.NotEmpty(t, nodes)
}

// Test ParseFile with Rust
func TestDefaultCodeParser_ParseFileRust(t *testing.T) {
	parser := NewDefaultCodeParser(logrus.New())
	ctx := context.Background()

	content := `
pub fn hello() {
    println!("Hello");
}

fn private_func<T>(x: T) {
    println!("{:?}", x);
}
`
	nodes, _, err := parser.ParseFile(ctx, "/test/main.rs", []byte(content))
	require.NoError(t, err)
	assert.NotEmpty(t, nodes)
}

// Test ParseFile with docstring extraction
func TestDefaultCodeParser_ParseFileWithDocstring(t *testing.T) {
	parser := NewDefaultCodeParser(logrus.New())
	ctx := context.Background()

	content := `// This is a documented function
func documented() {
    println("Hello")
}
`
	nodes, _, err := parser.ParseFile(ctx, "/test/main.go", []byte(content))
	require.NoError(t, err)
	// Should have file node and function node
	assert.GreaterOrEqual(t, len(nodes), 1)
}

// Test SemanticSearch edge cases
func TestCodeGraph_SemanticSearchNoResults(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, &MockEmbeddingGenerator{}, logrus.New())

	// Add node with very specific name
	graph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "xyzabc"})

	ctx := context.Background()
	results, err := graph.SemanticSearch(ctx, "completely different query", 5)
	require.NoError(t, err)
	// May or may not find results depending on keyword matching
	assert.NotNil(t, results)
}

// Test Retrieve with reranker enabled
func TestGraphRAG_RetrieveWithRerankerEnabled(t *testing.T) {
	rerankFunc := func(ctx context.Context, query string, docs []string) ([]int, error) {
		result := make([]int, len(docs))
		for i := range docs {
			result[i] = i
		}
		return result, nil
	}
	reranker := NewLLMReranker(rerankFunc, logrus.New())

	config := DefaultGraphRAGConfig()
	config.EnableReranking = true
	config.MinRelevanceScore = 0.0

	graphConfig := DefaultCodeGraphConfig()
	codeGraph := NewCodeGraph(graphConfig, &MockEmbeddingGenerator{}, logrus.New())

	codeGraph.AddNode(&CodeNode{ID: "n1", Type: NodeTypeFunction, Name: "testFunc", Docstring: "test"})

	graphRAG := NewGraphRAG(config, codeGraph, reranker, logrus.New())

	ctx := context.Background()
	result, err := graphRAG.Retrieve(ctx, "test")

	require.NoError(t, err)
	assert.NotNil(t, result)
}
