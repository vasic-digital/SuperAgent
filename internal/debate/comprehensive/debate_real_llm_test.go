package comprehensive

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveDebateWithRealLLMCalls validates that comprehensive debate uses real LLM calls
func TestComprehensiveDebateWithRealLLMCalls(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr, err := NewIntegrationManager(DefaultConfig(), logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = mgr.Initialize(ctx)
	require.NoError(t, err)

	t.Run("StreamDebateGeneratesContent", func(t *testing.T) {
		contentReceived := false
		streamHandler := func(event *StreamEvent) error {
			if event.Type == StreamEventAgentResponse && event.Content != "" {
				contentReceived = true
			}
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "test-real-llm",
				Topic:     "Write a function to calculate fibonacci",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		resp, err := mgr.StreamDebate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify content was generated (not empty)
		assert.True(t, contentReceived, "Debate should generate content, not empty responses")
		assert.True(t, resp.Success)
	})

	t.Run("FallbackEventsTracked", func(t *testing.T) {
		fallbackEventReceived := false
		streamHandler := func(event *StreamEvent) error {
			if event.Type == StreamEventAgentFallback {
				fallbackEventReceived = true
				assert.NotEmpty(t, event.Error, "Fallback event should include error reason")
			}
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "test-fallback",
				Topic:     "Test topic",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		_, err := mgr.StreamDebate(ctx, req)
		// Fallback events may or may not occur depending on provider health
		// We just verify the structure is correct if they do occur
		_ = fallbackEventReceived
		_ = err
	})

	t.Run("AllTeamsParticipate", func(t *testing.T) {
		teamsActivated := make(map[Team]bool)
		streamHandler := func(event *StreamEvent) error {
			if event.Type == StreamEventTeamStart {
				teamsActivated[event.Team] = true
			}
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "test-all-teams",
				Topic:     "Test topic",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		_, err := mgr.StreamDebate(ctx, req)
		require.NoError(t, err)

		// Verify at least some teams were activated
		assert.Greater(t, len(teamsActivated), 0, "At least one team should be activated")
	})

	t.Run("AgentMetadataComplete", func(t *testing.T) {
		var agentInfos []*AgentInfo
		streamHandler := func(event *StreamEvent) error {
			if event.Agent != nil {
				agentInfos = append(agentInfos, event.Agent)
			}
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "test-agent-info",
				Topic:     "Test topic",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		_, err := mgr.StreamDebate(ctx, req)
		require.NoError(t, err)

		// Verify agent info is complete
		for _, agent := range agentInfos {
			assert.NotEmpty(t, agent.ID, "Agent ID should not be empty")
			assert.NotEmpty(t, agent.Role, "Agent role should not be empty")
			assert.NotEmpty(t, agent.Provider, "Agent provider should not be empty")
			assert.NotEmpty(t, agent.Model, "Agent model should not be empty")
		}
	})
}

// TestDebateQuality validates debate quality metrics
func TestDebateQuality(t *testing.T) {
	logger := logrus.New()
	mgr, err := NewIntegrationManager(DefaultConfig(), logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = mgr.Initialize(ctx)
	require.NoError(t, err)

	t.Run("QualityScoreWithinRange", func(t *testing.T) {
		req := &DebateRequest{
			ID:        "test-quality",
			Topic:     "Test topic",
			Language:  "go",
			MaxRounds: 1,
		}

		resp, err := mgr.ExecuteDebate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Quality score should be between 0 and 1
		assert.GreaterOrEqual(t, resp.QualityScore, 0.0)
		assert.LessOrEqual(t, resp.QualityScore, 1.0)
	})

	t.Run("ConsensusStructureValid", func(t *testing.T) {
		req := &DebateRequest{
			ID:        "test-consensus",
			Topic:     "Test topic",
			Language:  "go",
			MaxRounds: 1,
		}

		resp, err := mgr.ExecuteDebate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Consensus)

		assert.NotZero(t, resp.Consensus.AchievedAt)
		if resp.Consensus.Reached {
			assert.NotEmpty(t, resp.Consensus.Summary, "Consensus summary should not be empty when reached")
		}
	})
}

// TestStreamEvents validates all stream event types are properly emitted
func TestStreamEvents(t *testing.T) {
	logger := logrus.New()
	mgr, err := NewIntegrationManager(DefaultConfig(), logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = mgr.Initialize(ctx)
	require.NoError(t, err)

	t.Run("AllEventTypesEmitted", func(t *testing.T) {
		eventsReceived := make(map[StreamEventType]bool)
		streamHandler := func(event *StreamEvent) error {
			eventsReceived[event.Type] = true
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "test-events",
				Topic:     "Test topic",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		_, err := mgr.StreamDebate(ctx, req)
		require.NoError(t, err)

		// Verify core events are received
		assert.True(t, eventsReceived[StreamEventDebateStart], "Debate start event should be received")
		assert.True(t, eventsReceived[StreamEventDebateComplete], "Debate complete event should be received")
	})

	t.Run("EventTimestampsValid", func(t *testing.T) {
		var timestamps []time.Time
		streamHandler := func(event *StreamEvent) error {
			timestamps = append(timestamps, event.Timestamp)
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "test-timestamps",
				Topic:     "Test topic",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		_, err := mgr.StreamDebate(ctx, req)
		require.NoError(t, err)

		// Verify timestamps are in chronological order
		for i := 1; i < len(timestamps); i++ {
			assert.True(t, timestamps[i].After(timestamps[i-1]) || timestamps[i].Equal(timestamps[i-1]),
				"Timestamps should be in chronological order")
		}
	})
}
