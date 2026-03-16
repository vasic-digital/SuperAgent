package knowledge

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CodeGraph extended tests
// =============================================================================

func TestCodeGraph_AddNode_MaxNodes(t *testing.T) {
	config := DefaultCodeGraphConfig()
	config.MaxNodes = 2

	graph := NewCodeGraph(config, nil, nil)

	assert.NoError(t, graph.AddNode(&CodeNode{ID: "n1", Name: "node1", Type: NodeTypeFunction}))
	assert.NoError(t, graph.AddNode(&CodeNode{ID: "n2", Name: "node2", Type: NodeTypeFunction}))

	err := graph.AddNode(&CodeNode{ID: "n3", Name: "node3", Type: NodeTypeFunction})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum nodes reached")
}

func TestCodeGraph_AddNode_AutoGenerateID(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	node := &CodeNode{Name: "myFunc", Type: NodeTypeFunction, Path: "/main.go", StartLine: 10}
	err := graph.AddNode(node)
	assert.NoError(t, err)
	assert.NotEmpty(t, node.ID)
	assert.Contains(t, node.ID, "function")
}

func TestCodeGraph_AddNode_SetsTimestamps(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	node := &CodeNode{ID: "ts-node", Name: "test", Type: NodeTypeFunction}
	err := graph.AddNode(node)
	assert.NoError(t, err)
	assert.False(t, node.CreatedAt.IsZero())
	assert.False(t, node.UpdatedAt.IsZero())
}

func TestCodeGraph_AddNode_InitializesMetadata(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	node := &CodeNode{ID: "meta-node", Name: "test", Type: NodeTypeFunction}
	assert.Nil(t, node.Metadata)

	err := graph.AddNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, node.Metadata)
}

func TestCodeGraph_AddEdge_SourceNotFound(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	edge := &CodeEdge{SourceID: "missing", TargetID: "also-missing", Type: EdgeTypeCalls}
	err := graph.AddEdge(edge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source node not found")
}

func TestCodeGraph_AddEdge_TargetNotFound(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	_ = graph.AddNode(&CodeNode{ID: "src", Name: "source", Type: NodeTypeFunction})

	edge := &CodeEdge{SourceID: "src", TargetID: "missing", Type: EdgeTypeCalls}
	err := graph.AddEdge(edge)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target node not found")
}

func TestCodeGraph_AddEdge_MaxEdges(t *testing.T) {
	config := DefaultCodeGraphConfig()
	config.MaxEdgesPerNode = 1

	graph := NewCodeGraph(config, nil, nil)
	_ = graph.AddNode(&CodeNode{ID: "s", Name: "source", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "t1", Name: "target1", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "t2", Name: "target2", Type: NodeTypeFunction})

	assert.NoError(t, graph.AddEdge(&CodeEdge{SourceID: "s", TargetID: "t1", Type: EdgeTypeCalls}))

	err := graph.AddEdge(&CodeEdge{SourceID: "s", TargetID: "t2", Type: EdgeTypeCalls})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max edges reached")
}

func TestCodeGraph_AddEdge_DefaultWeight(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	_ = graph.AddNode(&CodeNode{ID: "a", Name: "a", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "b", Name: "b", Type: NodeTypeFunction})

	edge := &CodeEdge{SourceID: "a", TargetID: "b", Type: EdgeTypeCalls}
	assert.NoError(t, graph.AddEdge(edge))
	assert.Equal(t, 1.0, edge.Weight)
}

func TestCodeGraph_GetNode_NotFound(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	node, exists := graph.GetNode("nonexistent")
	assert.Nil(t, node)
	assert.False(t, exists)
}

func TestCodeGraph_GetNodesByType_Extended(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	_ = graph.AddNode(&CodeNode{ID: "f1", Name: "func1", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "f2", Name: "func2", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "s1", Name: "struct1", Type: NodeTypeStruct})

	funcs := graph.GetNodesByType(NodeTypeFunction)
	assert.Len(t, funcs, 2)

	structs := graph.GetNodesByType(NodeTypeStruct)
	assert.Len(t, structs, 1)
}

func TestCodeGraph_GetNeighbors_AllEdges(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	_ = graph.AddNode(&CodeNode{ID: "a", Name: "a", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "b", Name: "b", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "c", Name: "c", Type: NodeTypeFunction})

	_ = graph.AddEdge(&CodeEdge{SourceID: "a", TargetID: "b", Type: EdgeTypeCalls})
	_ = graph.AddEdge(&CodeEdge{SourceID: "a", TargetID: "c", Type: EdgeTypeImports})

	// Get all neighbors (empty edge type filter)
	neighbors := graph.GetNeighbors("a", "")
	assert.Len(t, neighbors, 2)

	// Get only calls neighbors
	callNeighbors := graph.GetNeighbors("a", EdgeTypeCalls)
	assert.Len(t, callNeighbors, 1)
}

func TestCodeGraph_FindPath_Extended(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	_ = graph.AddNode(&CodeNode{ID: "a", Name: "a", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "b", Name: "b", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "c", Name: "c", Type: NodeTypeFunction})

	_ = graph.AddEdge(&CodeEdge{SourceID: "a", TargetID: "b", Type: EdgeTypeCalls})
	_ = graph.AddEdge(&CodeEdge{SourceID: "b", TargetID: "c", Type: EdgeTypeCalls})

	path, err := graph.FindPath("a", "c")
	assert.NoError(t, err)
	assert.Len(t, path, 3)

	// No path
	_, err = graph.FindPath("c", "a")
	assert.Error(t, err)
}

func TestCodeGraph_FindPath_SourceMissing(t *testing.T) {
	graph := NewCodeGraph(DefaultCodeGraphConfig(), nil, nil)
	_, err := graph.FindPath("missing", "also-missing")
	assert.Error(t, err)
}

func TestCodeGraph_GetImpactRadius_Extended(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	_ = graph.AddNode(&CodeNode{ID: "core", Name: "core", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "dep1", Name: "dep1", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "dep2", Name: "dep2", Type: NodeTypeFunction})

	_ = graph.AddEdge(&CodeEdge{SourceID: "dep1", TargetID: "core", Type: EdgeTypeCalledBy})
	_ = graph.AddEdge(&CodeEdge{SourceID: "dep2", TargetID: "dep1", Type: EdgeTypeDependsOn})

	impacted := graph.GetImpactRadius("core", 2)
	assert.GreaterOrEqual(t, len(impacted), 1) // at least the node itself
}

func TestCodeGraph_SemanticSearch_NoEmbedder(t *testing.T) {
	graph := NewCodeGraph(DefaultCodeGraphConfig(), nil, nil)
	_, err := graph.SemanticSearch(context.Background(), "test", 5)
	assert.Error(t, err)
}

func TestCodeGraph_GetStats_Extended(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)

	_ = graph.AddNode(&CodeNode{ID: "n1", Name: "n1", Type: NodeTypeFunction})
	_ = graph.AddNode(&CodeNode{ID: "n2", Name: "n2", Type: NodeTypeStruct})
	_ = graph.AddEdge(&CodeEdge{SourceID: "n1", TargetID: "n2", Type: EdgeTypeCalls})

	stats := graph.GetStats()
	assert.Equal(t, 2, stats["total_nodes"])
	assert.Equal(t, 1, stats["total_edges"])
}

func TestCodeGraph_MarshalJSON_Extended(t *testing.T) {
	config := DefaultCodeGraphConfig()
	graph := NewCodeGraph(config, nil, nil)
	_ = graph.AddNode(&CodeNode{ID: "n1", Name: "n1", Type: NodeTypeFunction})

	data, err := graph.MarshalJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(data), "n1")
}

// =============================================================================
// Cosine similarity tests
// =============================================================================

func TestCosineSimilarity_Extended(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{"identical vectors", []float64{1, 0, 0}, []float64{1, 0, 0}, 1.0},
		{"orthogonal vectors", []float64{1, 0}, []float64{0, 1}, 0.0},
		{"different length", []float64{1}, []float64{1, 2}, 0.0},
		{"empty vectors", []float64{}, []float64{}, 0.0},
		{"zero vectors", []float64{0, 0}, []float64{0, 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

// =============================================================================
// DefaultCodeParser tests
// =============================================================================

func TestDefaultCodeParser_ParseFile_Go(t *testing.T) {
	parser := NewDefaultCodeParser(nil)

	code := []byte(`package main

func main() {
    fmt.Println("hello")
}

func add(a, b int) int {
    return a + b
}
`)

	nodes, edges, err := parser.ParseFile(context.Background(), "/test/main.go", code)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(nodes), 1) // at least the file node
	assert.GreaterOrEqual(t, len(edges), 0)
}

func TestDefaultCodeParser_ParseFile_UnknownLanguage(t *testing.T) {
	parser := NewDefaultCodeParser(nil)

	nodes, edges, err := parser.ParseFile(context.Background(), "/test/file.xyz", []byte("content"))
	assert.NoError(t, err)
	assert.Len(t, nodes, 1) // only file node
	assert.Len(t, edges, 0)
}

func TestDefaultCodeParser_SupportedLanguages_Extended(t *testing.T) {
	parser := NewDefaultCodeParser(nil)
	langs := parser.SupportedLanguages()

	assert.GreaterOrEqual(t, len(langs), 3) // at least go, python, javascript
}

// =============================================================================
// GraphRAG extended tests
// =============================================================================

func TestGraphRAG_BuildContext_NilResult(t *testing.T) {
	rag := NewGraphRAG(DefaultGraphRAGConfig(), nil, nil, nil)
	assert.Empty(t, rag.BuildContext(nil))
}

func TestGraphRAG_BuildContext_EmptyNodes(t *testing.T) {
	rag := NewGraphRAG(DefaultGraphRAGConfig(), nil, nil, nil)
	result := &RetrievalResult{Nodes: []*RetrievedNode{}}
	assert.Empty(t, rag.BuildContext(result))
}

func TestGraphRAG_BuildContext_WithNodes(t *testing.T) {
	rag := NewGraphRAG(DefaultGraphRAGConfig(), nil, nil, nil)
	result := &RetrievalResult{
		Nodes: []*RetrievedNode{
			{
				Node: &CodeNode{
					Name:      "TestFunc",
					Type:      NodeTypeFunction,
					Path:      "/test.go",
					StartLine: 10,
					EndLine:   20,
					Signature: "func TestFunc()",
					Docstring: "TestFunc is a test",
				},
				RelevanceScore: 0.95,
			},
		},
	}

	ctx := rag.BuildContext(result)
	assert.Contains(t, ctx, "TestFunc")
	assert.Contains(t, ctx, "0.95")
}

func TestGraphRAG_RetrieveForNode_NotFound(t *testing.T) {
	graph := NewCodeGraph(DefaultCodeGraphConfig(), nil, nil)
	rag := NewGraphRAG(DefaultGraphRAGConfig(), graph, nil, nil)

	_, err := rag.RetrieveForNode(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestGraphRAG_RetrieveForNode_Found(t *testing.T) {
	graph := NewCodeGraph(DefaultCodeGraphConfig(), nil, nil)
	_ = graph.AddNode(&CodeNode{ID: "target", Name: "Target", Type: NodeTypeFunction})

	rag := NewGraphRAG(DefaultGraphRAGConfig(), graph, nil, nil)
	result, err := rag.RetrieveForNode(context.Background(), "target")

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalFound, 1)
}

func TestRetrievalResult_MarshalJSON_Extended(t *testing.T) {
	result := &RetrievalResult{
		Query:      "test",
		Duration:   100 * time.Millisecond,
		TotalFound: 5,
		Mode:       RetrievalModeHybrid,
	}

	data, err := result.MarshalJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(data), "duration_ms")
	assert.Contains(t, string(data), "100")
}

// =============================================================================
// Helper functions tests
// =============================================================================

func TestGetEntityID(t *testing.T) {
	t.Run("with entity", func(t *testing.T) {
		update := &EntityUpdate{Entity: &GraphEntity{ID: "e1"}}
		assert.Equal(t, "e1", getEntityID(update))
	})

	t.Run("with relationship", func(t *testing.T) {
		update := &EntityUpdate{Relationship: &GraphRelationship{ID: "r1"}}
		assert.Equal(t, "r1", getEntityID(update))
	})

	t.Run("empty", func(t *testing.T) {
		update := &EntityUpdate{}
		assert.Empty(t, getEntityID(update))
	})
}

func TestGraphEntity_Fields(t *testing.T) {
	entity := &GraphEntity{
		ID:         "e1",
		Type:       "concept",
		Name:       "Test Entity",
		Value:      "some value",
		Confidence: 0.9,
		Importance: 0.8,
	}

	assert.Equal(t, "e1", entity.ID)
	assert.Equal(t, "concept", entity.Type)
	assert.Equal(t, 0.9, entity.Confidence)
}

func TestGraphRelationship_Fields(t *testing.T) {
	rel := &GraphRelationship{
		ID:                "r1",
		Type:              "RELATED_TO",
		SourceID:          "e1",
		TargetID:          "e2",
		Strength:          0.7,
		CooccurrenceCount: 5,
	}

	assert.Equal(t, "r1", rel.ID)
	assert.Equal(t, "e1", rel.SourceID)
	assert.Equal(t, 5, rel.CooccurrenceCount)
}

func TestDefaultCodeGraphConfig(t *testing.T) {
	config := DefaultCodeGraphConfig()

	assert.Equal(t, 100000, config.MaxNodes)
	assert.Equal(t, 100, config.MaxEdgesPerNode)
	assert.True(t, config.EnableEmbeddings)
	assert.Equal(t, 768, config.EmbeddingDimension)
	assert.Equal(t, 0.8, config.SimilarityThreshold)
	assert.GreaterOrEqual(t, len(config.IndexLanguages), 5)
	assert.GreaterOrEqual(t, len(config.ExcludePatterns), 3)
}

func TestDefaultGraphRAGConfig(t *testing.T) {
	config := DefaultGraphRAGConfig()

	assert.Equal(t, RetrievalModeHybrid, config.RetrievalMode)
	assert.Equal(t, 20, config.MaxNodes)
	assert.Equal(t, 3, config.MaxDepth)
	assert.Equal(t, 0.6, config.VectorWeight)
	assert.Equal(t, 0.4, config.GraphWeight)
	assert.True(t, config.EnableReranking)
}

func TestGetString(t *testing.T) {
	props := map[string]interface{}{"key": "value", "num": 42}
	assert.Equal(t, "value", getString(props, "key"))
	assert.Empty(t, getString(props, "num"))
	assert.Empty(t, getString(props, "missing"))
}

func TestGetFloat64(t *testing.T) {
	props := map[string]interface{}{"score": 0.95, "name": "test"}
	assert.Equal(t, 0.95, getFloat64(props, "score"))
	assert.Equal(t, 0.0, getFloat64(props, "name"))
	assert.Equal(t, 0.0, getFloat64(props, "missing"))
}

func TestSqrt_Extended(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
		delta    float64
	}{
		{"sqrt(4)", 4.0, 2.0, 0.001},
		{"sqrt(9)", 9.0, 3.0, 0.001},
		{"sqrt(0)", 0.0, 0.0, 0.001},
		{"sqrt(-1)", -1.0, 0.0, 0.001},
		{"sqrt(1)", 1.0, 1.0, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqrt(tt.input)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}
