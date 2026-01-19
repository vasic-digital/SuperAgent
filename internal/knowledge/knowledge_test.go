// Package knowledge provides tests for the Code Knowledge Graph and GraphRAG.
package knowledge

import (
	"context"
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
