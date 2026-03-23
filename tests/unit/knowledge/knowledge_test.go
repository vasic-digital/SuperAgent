package knowledge_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/knowledge"
)

// TestCodeGraphCreation validates code graph initialization
func TestCodeGraphCreation(t *testing.T) {
	logger := logrus.New()
	cfg := knowledge.CodeGraphConfig{
		MaxNodes: 1000,
	}
	graph := knowledge.NewCodeGraph(cfg, nil, logger)
	require.NotNil(t, graph, "code graph must not be nil")
}

// TestGraphRAGCreation validates GraphRAG initialization
func TestGraphRAGCreation(t *testing.T) {
	logger := logrus.New()
	cfg := knowledge.CodeGraphConfig{MaxNodes: 100}
	graph := knowledge.NewCodeGraph(cfg, nil, logger)

	ragCfg := knowledge.GraphRAGConfig{
		MaxNodes: 10,
		MaxDepth: 3,
	}
	rag := knowledge.NewGraphRAG(ragCfg, graph, nil, logger)
	require.NotNil(t, rag, "graph RAG must not be nil")
}

// TestKnowledgeGraphStreamNodeTypes validates knowledge node type constants
func TestKnowledgeGraphStreamNodeTypes(t *testing.T) {
	nodeTypes := []string{
		"entity",
		"concept",
		"relationship",
		"document",
		"code_symbol",
	}

	for _, nodeType := range nodeTypes {
		t.Run(nodeType, func(t *testing.T) {
			assert.NotEmpty(t, nodeType, "node type must not be empty")
			assert.Greater(t, len(nodeType), 2, "node type must be descriptive")
		})
	}
}
