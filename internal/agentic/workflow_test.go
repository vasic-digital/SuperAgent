package agentic

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultWorkflowConfig(t *testing.T) {
	config := DefaultWorkflowConfig()

	assert.Equal(t, 100, config.MaxIterations)
	assert.Equal(t, 30*time.Minute, config.Timeout)
	assert.True(t, config.EnableCheckpoints)
	assert.Equal(t, 5, config.CheckpointInterval)
	assert.True(t, config.EnableSelfCorrection)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
}

func TestNewWorkflow(t *testing.T) {
	t.Run("WithNilConfig", func(t *testing.T) {
		w := NewWorkflow("test", "test workflow", nil, nil)

		assert.NotEmpty(t, w.ID)
		assert.Equal(t, "test", w.Name)
		assert.Equal(t, "test workflow", w.Description)
		assert.NotNil(t, w.Config)
		assert.NotNil(t, w.Logger)
		assert.NotNil(t, w.Graph)
		assert.NotNil(t, w.Graph.Nodes)
		assert.NotNil(t, w.Graph.Edges)
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		config := &WorkflowConfig{
			MaxIterations: 50,
			Timeout:       5 * time.Minute,
		}
		logger := logrus.New()

		w := NewWorkflow("custom", "custom workflow", config, logger)

		assert.Equal(t, 50, w.Config.MaxIterations)
		assert.Equal(t, 5*time.Minute, w.Config.Timeout)
		assert.Equal(t, logger, w.Logger)
	})
}

func TestWorkflow_AddNode(t *testing.T) {
	w := NewWorkflow("test", "test", nil, nil)

	t.Run("WithID", func(t *testing.T) {
		node := &Node{
			ID:   "node1",
			Name: "Test Node",
			Type: NodeTypeAgent,
		}

		err := w.AddNode(node)
		require.NoError(t, err)

		assert.Equal(t, node, w.Graph.Nodes["node1"])
	})

	t.Run("WithoutID", func(t *testing.T) {
		node := &Node{
			Name: "Auto ID Node",
			Type: NodeTypeTool,
		}

		err := w.AddNode(node)
		require.NoError(t, err)

		assert.NotEmpty(t, node.ID)
		assert.Equal(t, node, w.Graph.Nodes[node.ID])
	})
}

func TestWorkflow_AddEdge(t *testing.T) {
	w := NewWorkflow("test", "test", nil, nil)
	w.AddNode(&Node{ID: "node1", Name: "Node 1"})
	w.AddNode(&Node{ID: "node2", Name: "Node 2"})

	t.Run("ValidEdge", func(t *testing.T) {
		err := w.AddEdge("node1", "node2", nil, "edge1")
		require.NoError(t, err)

		assert.Len(t, w.Graph.Edges, 1)
		assert.Equal(t, "node1", w.Graph.Edges[0].From)
		assert.Equal(t, "node2", w.Graph.Edges[0].To)
		assert.Equal(t, "edge1", w.Graph.Edges[0].Label)
	})

	t.Run("InvalidSourceNode", func(t *testing.T) {
		err := w.AddEdge("invalid", "node2", nil, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source node not found")
	})

	t.Run("InvalidTargetNode", func(t *testing.T) {
		err := w.AddEdge("node1", "invalid", nil, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target node not found")
	})

	t.Run("WithCondition", func(t *testing.T) {
		condition := func(state *WorkflowState) bool {
			return state.Variables["proceed"].(bool)
		}
		err := w.AddEdge("node1", "node2", condition, "conditional")
		require.NoError(t, err)
	})
}

func TestWorkflow_SetEntryPoint(t *testing.T) {
	w := NewWorkflow("test", "test", nil, nil)
	w.AddNode(&Node{ID: "node1", Name: "Node 1"})

	t.Run("ValidNode", func(t *testing.T) {
		err := w.SetEntryPoint("node1")
		require.NoError(t, err)
		assert.Equal(t, "node1", w.Graph.EntryPoint)
	})

	t.Run("InvalidNode", func(t *testing.T) {
		err := w.SetEntryPoint("invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})
}

func TestWorkflow_AddEndNode(t *testing.T) {
	w := NewWorkflow("test", "test", nil, nil)
	w.AddNode(&Node{ID: "node1", Name: "Node 1"})

	t.Run("ValidNode", func(t *testing.T) {
		err := w.AddEndNode("node1")
		require.NoError(t, err)
		assert.Contains(t, w.Graph.EndNodes, "node1")
	})

	t.Run("InvalidNode", func(t *testing.T) {
		err := w.AddEndNode("invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "node not found")
	})
}

func TestWorkflow_Execute(t *testing.T) {
	t.Run("NoEntryPoint", func(t *testing.T) {
		w := NewWorkflow("test", "test", nil, nil)
		w.AddNode(&Node{ID: "node1", Name: "Node 1"})

		state, err := w.Execute(context.Background(), nil)
		require.Error(t, err)
		assert.Nil(t, state)
		assert.Contains(t, err.Error(), "no entry point defined")
	})

	t.Run("SimpleWorkflow", func(t *testing.T) {
		w := NewWorkflow("simple", "simple workflow", nil, nil)

		var callCount int32
		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			atomic.AddInt32(&callCount, 1)
			return &NodeOutput{
				Result:    "success",
				ShouldEnd: true,
			}, nil
		}

		w.AddNode(&Node{ID: "start", Name: "Start", Handler: handler})
		w.SetEntryPoint("start")
		w.AddEndNode("start")

		state, err := w.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, StatusCompleted, state.Status)
		assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
		assert.Len(t, state.History, 1)
	})

	t.Run("MultiNodeWorkflow", func(t *testing.T) {
		w := NewWorkflow("multi", "multi-node workflow", nil, nil)

		var order []string
		handler1 := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			order = append(order, "node1")
			return &NodeOutput{}, nil
		}
		handler2 := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			order = append(order, "node2")
			return &NodeOutput{ShouldEnd: true}, nil
		}

		w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: handler1})
		w.AddNode(&Node{ID: "node2", Name: "Node 2", Handler: handler2})
		w.AddEdge("node1", "node2", nil, "")
		w.SetEntryPoint("node1")
		w.AddEndNode("node2")

		state, err := w.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, StatusCompleted, state.Status)
		assert.Equal(t, []string{"node1", "node2"}, order)
	})

	t.Run("ConditionalBranching", func(t *testing.T) {
		w := NewWorkflow("conditional", "conditional workflow", nil, nil)

		handler1 := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			state.Variables["branch"] = "A"
			return &NodeOutput{}, nil
		}
		handlerA := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			state.Variables["result"] = "branch_a"
			return &NodeOutput{ShouldEnd: true}, nil
		}
		handlerB := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			state.Variables["result"] = "branch_b"
			return &NodeOutput{ShouldEnd: true}, nil
		}

		w.AddNode(&Node{ID: "start", Name: "Start", Handler: handler1})
		w.AddNode(&Node{ID: "branch_a", Name: "Branch A", Handler: handlerA})
		w.AddNode(&Node{ID: "branch_b", Name: "Branch B", Handler: handlerB})

		conditionA := func(state *WorkflowState) bool {
			return state.Variables["branch"] == "A"
		}
		conditionB := func(state *WorkflowState) bool {
			return state.Variables["branch"] == "B"
		}

		w.AddEdge("start", "branch_a", conditionA, "if A")
		w.AddEdge("start", "branch_b", conditionB, "if B")
		w.SetEntryPoint("start")

		state, err := w.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, "branch_a", state.Variables["result"])
	})

	t.Run("HandlerError", func(t *testing.T) {
		config := &WorkflowConfig{
			MaxIterations: 10,
			Timeout:       10 * time.Second,
			MaxRetries:    0, // No retries
		}
		w := NewWorkflow("error", "error workflow", config, nil)

		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			return nil, errors.New("handler error")
		}

		w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: handler})
		w.SetEntryPoint("node1")

		state, err := w.Execute(context.Background(), nil)
		require.Error(t, err)
		assert.Equal(t, StatusFailed, state.Status)
		assert.Contains(t, err.Error(), "handler error")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		w := NewWorkflow("cancel", "cancel workflow", nil, nil)

		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			time.Sleep(100 * time.Millisecond)
			return &NodeOutput{}, nil
		}

		w.AddNode(&Node{ID: "slow", Name: "Slow Node", Handler: handler})
		w.AddEdge("slow", "slow", nil, "loop")
		w.SetEntryPoint("slow")

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		state, err := w.Execute(ctx, nil)
		require.Error(t, err)
		assert.Equal(t, StatusFailed, state.Status)
	})

	t.Run("MaxIterations", func(t *testing.T) {
		config := &WorkflowConfig{
			MaxIterations: 3,
			Timeout:       10 * time.Second,
		}
		w := NewWorkflow("loop", "loop workflow", config, nil)

		var count int32
		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			atomic.AddInt32(&count, 1)
			return &NodeOutput{}, nil
		}

		w.AddNode(&Node{ID: "loop", Name: "Loop Node", Handler: handler})
		w.AddEdge("loop", "loop", nil, "loop")
		w.SetEntryPoint("loop")

		state, err := w.Execute(context.Background(), nil)
		require.Error(t, err)
		assert.Equal(t, StatusFailed, state.Status)
		assert.Contains(t, err.Error(), "max iterations reached")
		assert.Equal(t, int32(3), atomic.LoadInt32(&count))
	})

	t.Run("NilHandler", func(t *testing.T) {
		w := NewWorkflow("nil", "nil handler workflow", nil, nil)

		w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: nil})
		w.SetEntryPoint("node1")
		w.AddEndNode("node1")

		state, err := w.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, StatusCompleted, state.Status)
	})

	t.Run("WithInputMessages", func(t *testing.T) {
		w := NewWorkflow("input", "input workflow", nil, nil)

		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			return &NodeOutput{ShouldEnd: true}, nil
		}

		w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: handler})
		w.SetEntryPoint("node1")

		input := &NodeInput{
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
		}

		state, err := w.Execute(context.Background(), input)
		require.NoError(t, err)
		assert.Len(t, state.Messages, 1)
		assert.Equal(t, "Hello", state.Messages[0].Content)
	})

	t.Run("EndNodeTermination", func(t *testing.T) {
		w := NewWorkflow("endnode", "end node workflow", nil, nil)

		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			return &NodeOutput{}, nil
		}

		w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: handler})
		w.SetEntryPoint("node1")
		w.AddEndNode("node1")

		state, err := w.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, StatusCompleted, state.Status)
	})
}

func TestWorkflow_ExecuteWithRetry(t *testing.T) {
	t.Run("RetryOnFailure", func(t *testing.T) {
		config := &WorkflowConfig{
			MaxIterations: 10,
			Timeout:       10 * time.Second,
			MaxRetries:    2,
			RetryDelay:    10 * time.Millisecond,
		}
		w := NewWorkflow("retry", "retry workflow", config, logrus.New())

		var attempts int32
		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			count := atomic.AddInt32(&attempts, 1)
			if count < 3 {
				return nil, errors.New("temporary error")
			}
			return &NodeOutput{ShouldEnd: true}, nil
		}

		w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: handler})
		w.SetEntryPoint("node1")

		state, err := w.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, StatusCompleted, state.Status)
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})

	t.Run("RetryWithNodePolicy", func(t *testing.T) {
		config := &WorkflowConfig{
			MaxIterations: 10,
			Timeout:       10 * time.Second,
			MaxRetries:    1,
			RetryDelay:    10 * time.Millisecond,
		}
		w := NewWorkflow("retry", "retry workflow", config, logrus.New())

		var attempts int32
		handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
			count := atomic.AddInt32(&attempts, 1)
			if count < 4 {
				return nil, errors.New("temporary error")
			}
			return &NodeOutput{ShouldEnd: true}, nil
		}

		w.AddNode(&Node{
			ID:      "node1",
			Name:    "Node 1",
			Handler: handler,
			RetryPolicy: &RetryPolicy{
				MaxRetries: 3,
				Delay:      5 * time.Millisecond,
				Backoff:    1.5,
			},
		})
		w.SetEntryPoint("node1")

		state, err := w.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, StatusCompleted, state.Status)
	})
}

func TestWorkflow_RestoreFromCheckpoint(t *testing.T) {
	state := &WorkflowState{
		CurrentNode: "node1",
		Variables:   map[string]interface{}{"key": "value1"},
		Checkpoints: []Checkpoint{
			{
				ID:        "cp1",
				NodeID:    "node2",
				State:     map[string]interface{}{"key": "value2"},
				Timestamp: time.Now(),
			},
		},
		Status: StatusFailed,
	}

	w := NewWorkflow("test", "test", nil, nil)

	t.Run("ValidCheckpoint", func(t *testing.T) {
		err := w.RestoreFromCheckpoint(state, "cp1")
		require.NoError(t, err)
		assert.Equal(t, "node2", state.CurrentNode)
		assert.Equal(t, "value2", state.Variables["key"])
		assert.Equal(t, StatusRunning, state.Status)
	})

	t.Run("InvalidCheckpoint", func(t *testing.T) {
		err := w.RestoreFromCheckpoint(state, "invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "checkpoint not found")
	})
}

func TestWorkflow_Checkpoints(t *testing.T) {
	config := &WorkflowConfig{
		MaxIterations:      20,
		Timeout:            10 * time.Second,
		EnableCheckpoints:  true,
		CheckpointInterval: 2,
	}
	w := NewWorkflow("checkpoint", "checkpoint workflow", config, nil)

	var count int32
	handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
		c := atomic.AddInt32(&count, 1)
		state.Variables["count"] = c
		if c >= 5 {
			return &NodeOutput{ShouldEnd: true}, nil
		}
		return &NodeOutput{}, nil
	}

	w.AddNode(&Node{ID: "loop", Name: "Loop Node", Handler: handler})
	w.AddEdge("loop", "loop", nil, "loop")
	w.SetEntryPoint("loop")

	state, err := w.Execute(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, state.Status)
	// Should have checkpoints at iterations 2 and 4
	assert.GreaterOrEqual(t, len(state.Checkpoints), 2)
}

func TestWorkflow_NextNodeOverride(t *testing.T) {
	w := NewWorkflow("override", "override workflow", nil, nil)

	handler1 := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
		// Override to skip node2 and go directly to node3
		return &NodeOutput{NextNode: "node3"}, nil
	}
	handler2 := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
		state.Variables["visited_node2"] = true
		return &NodeOutput{ShouldEnd: true}, nil
	}
	handler3 := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
		state.Variables["visited_node3"] = true
		return &NodeOutput{ShouldEnd: true}, nil
	}

	w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: handler1})
	w.AddNode(&Node{ID: "node2", Name: "Node 2", Handler: handler2})
	w.AddNode(&Node{ID: "node3", Name: "Node 3", Handler: handler3})
	w.AddEdge("node1", "node2", nil, "")
	w.SetEntryPoint("node1")

	state, err := w.Execute(context.Background(), nil)
	require.NoError(t, err)
	assert.Nil(t, state.Variables["visited_node2"])
	assert.True(t, state.Variables["visited_node3"].(bool))
}

func TestNodeTypes(t *testing.T) {
	assert.Equal(t, NodeType("agent"), NodeTypeAgent)
	assert.Equal(t, NodeType("tool"), NodeTypeTool)
	assert.Equal(t, NodeType("condition"), NodeTypeCondition)
	assert.Equal(t, NodeType("parallel"), NodeTypeParallel)
	assert.Equal(t, NodeType("human"), NodeTypeHuman)
	assert.Equal(t, NodeType("subgraph"), NodeTypeSubgraph)
}

func TestWorkflowStatus(t *testing.T) {
	assert.Equal(t, WorkflowStatus("pending"), StatusPending)
	assert.Equal(t, WorkflowStatus("running"), StatusRunning)
	assert.Equal(t, WorkflowStatus("paused"), StatusPaused)
	assert.Equal(t, WorkflowStatus("completed"), StatusCompleted)
	assert.Equal(t, WorkflowStatus("failed"), StatusFailed)
}

func TestPow(t *testing.T) {
	tests := []struct {
		base     float64
		exp      float64
		expected float64
	}{
		{2, 0, 1},
		{2, 1, 2},
		{2, 3, 8},
		{1.5, 2, 2.25},
	}

	for _, tt := range tests {
		result := pow(tt.base, tt.exp)
		assert.InDelta(t, tt.expected, result, 0.001)
	}
}

func TestWorkflowConcurrency(t *testing.T) {
	w := NewWorkflow("concurrent", "concurrent test", nil, nil)

	var counter int64
	handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
		atomic.AddInt64(&counter, 1)
		return &NodeOutput{ShouldEnd: true}, nil
	}

	w.AddNode(&Node{ID: "node1", Name: "Node 1", Handler: handler})
	w.SetEntryPoint("node1")

	// Run multiple executions concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := w.Execute(context.Background(), nil)
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, int64(10), atomic.LoadInt64(&counter))
}

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "assistant",
		Content: "Hello",
		Name:    "bot",
		ToolCalls: []ToolCall{
			{ID: "1", Name: "test"},
		},
	}

	assert.Equal(t, "assistant", msg.Role)
	assert.Equal(t, "Hello", msg.Content)
	assert.Equal(t, "bot", msg.Name)
	assert.Len(t, msg.ToolCalls, 1)
}

func TestTool(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"param1": "value1",
		},
		Handler: handler,
	}

	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.NotNil(t, tool.Parameters)
	assert.NotNil(t, tool.Handler)

	result, err := tool.Handler(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "result", result)
}

func TestToolCall(t *testing.T) {
	tc := ToolCall{
		ID:   "call1",
		Name: "test_tool",
		Arguments: map[string]interface{}{
			"arg1": "value1",
		},
		Result: "success",
	}

	assert.Equal(t, "call1", tc.ID)
	assert.Equal(t, "test_tool", tc.Name)
	assert.Equal(t, "value1", tc.Arguments["arg1"])
	assert.Equal(t, "success", tc.Result)
}

func TestNodeInput(t *testing.T) {
	input := &NodeInput{
		Query: "test query",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Tools: []Tool{
			{Name: "tool1"},
		},
		Context: map[string]interface{}{
			"key": "value",
		},
		Previous: &NodeOutput{
			Result: "previous result",
		},
	}

	assert.Equal(t, "test query", input.Query)
	assert.Len(t, input.Messages, 1)
	assert.Len(t, input.Tools, 1)
	assert.Equal(t, "value", input.Context["key"])
	assert.Equal(t, "previous result", input.Previous.Result)
}

func TestNodeOutput(t *testing.T) {
	output := &NodeOutput{
		Result: "test result",
		Messages: []Message{
			{Role: "assistant", Content: "Response"},
		},
		ToolCalls: []ToolCall{
			{ID: "1", Name: "tool1"},
		},
		NextNode:  "next",
		ShouldEnd: false,
		Error:     nil,
		Metadata: map[string]interface{}{
			"meta": "data",
		},
	}

	assert.Equal(t, "test result", output.Result)
	assert.Len(t, output.Messages, 1)
	assert.Len(t, output.ToolCalls, 1)
	assert.Equal(t, "next", output.NextNode)
	assert.False(t, output.ShouldEnd)
	assert.Nil(t, output.Error)
	assert.Equal(t, "data", output.Metadata["meta"])
}

func TestRetryPolicy(t *testing.T) {
	policy := &RetryPolicy{
		MaxRetries: 5,
		Delay:      100 * time.Millisecond,
		Backoff:    2.0,
	}

	assert.Equal(t, 5, policy.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, policy.Delay)
	assert.Equal(t, 2.0, policy.Backoff)
}

func TestCheckpoint(t *testing.T) {
	checkpoint := Checkpoint{
		ID:     "cp1",
		NodeID: "node1",
		State: map[string]interface{}{
			"key": "value",
		},
		Timestamp: time.Now(),
	}

	assert.Equal(t, "cp1", checkpoint.ID)
	assert.Equal(t, "node1", checkpoint.NodeID)
	assert.Equal(t, "value", checkpoint.State["key"])
	assert.False(t, checkpoint.Timestamp.IsZero())
}

func TestNodeExecution(t *testing.T) {
	now := time.Now()
	exec := NodeExecution{
		NodeID:    "node1",
		NodeName:  "Test Node",
		StartTime: now,
		EndTime:   now.Add(time.Second),
		Input: &NodeInput{
			Query: "test",
		},
		Output: &NodeOutput{
			Result: "success",
		},
		Error: nil,
	}

	assert.Equal(t, "node1", exec.NodeID)
	assert.Equal(t, "Test Node", exec.NodeName)
	assert.Equal(t, time.Second, exec.EndTime.Sub(exec.StartTime))
	assert.NotNil(t, exec.Input)
	assert.NotNil(t, exec.Output)
	assert.Nil(t, exec.Error)
}

func TestEdge(t *testing.T) {
	condition := func(state *WorkflowState) bool {
		return true
	}

	edge := &Edge{
		From:      "node1",
		To:        "node2",
		Condition: condition,
		Label:     "test edge",
	}

	assert.Equal(t, "node1", edge.From)
	assert.Equal(t, "node2", edge.To)
	assert.NotNil(t, edge.Condition)
	assert.Equal(t, "test edge", edge.Label)
}

func TestNode(t *testing.T) {
	handler := func(ctx context.Context, state *WorkflowState, input *NodeInput) (*NodeOutput, error) {
		return &NodeOutput{}, nil
	}
	condition := func(state *WorkflowState) bool {
		return true
	}

	node := &Node{
		ID:        "node1",
		Name:      "Test Node",
		Type:      NodeTypeAgent,
		Handler:   handler,
		Condition: condition,
		Config: map[string]interface{}{
			"setting": "value",
		},
		RetryPolicy: &RetryPolicy{MaxRetries: 3},
	}

	assert.Equal(t, "node1", node.ID)
	assert.Equal(t, "Test Node", node.Name)
	assert.Equal(t, NodeTypeAgent, node.Type)
	assert.NotNil(t, node.Handler)
	assert.NotNil(t, node.Condition)
	assert.Equal(t, "value", node.Config["setting"])
	assert.Equal(t, 3, node.RetryPolicy.MaxRetries)
}

func TestWorkflowState(t *testing.T) {
	now := time.Now()
	endTime := now.Add(time.Minute)

	state := &WorkflowState{
		ID:          "state1",
		WorkflowID:  "workflow1",
		CurrentNode: "node1",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Variables: map[string]interface{}{
			"var": "value",
		},
		History: []NodeExecution{
			{NodeID: "node0"},
		},
		Checkpoints: []Checkpoint{
			{ID: "cp1"},
		},
		Status:    StatusRunning,
		StartTime: now,
		EndTime:   &endTime,
		Error:     nil,
	}

	assert.Equal(t, "state1", state.ID)
	assert.Equal(t, "workflow1", state.WorkflowID)
	assert.Equal(t, "node1", state.CurrentNode)
	assert.Len(t, state.Messages, 1)
	assert.Equal(t, "value", state.Variables["var"])
	assert.Len(t, state.History, 1)
	assert.Len(t, state.Checkpoints, 1)
	assert.Equal(t, StatusRunning, state.Status)
	assert.Equal(t, now, state.StartTime)
	assert.NotNil(t, state.EndTime)
	assert.Nil(t, state.Error)
}
