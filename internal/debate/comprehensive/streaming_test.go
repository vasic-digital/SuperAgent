package comprehensive

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamOrchestrator(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("NewStreamOrchestrator", func(t *testing.T) {
		config := DefaultStreamConfig()
		var receivedEvents []*StreamEvent

		handler := func(event *StreamEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		orchestrator := NewStreamOrchestrator(config, handler, "test-debate", 10)
		require.NotNil(t, orchestrator)
		assert.Equal(t, "test-debate", orchestrator.debateID)
		assert.Equal(t, 10, orchestrator.totalAgents)
		assert.Equal(t, 0, orchestrator.completedAgents)
	})

	t.Run("EmitDebateStart", func(t *testing.T) {
		var receivedEvents []*StreamEvent
		handler := func(event *StreamEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		orchestrator := NewStreamOrchestrator(nil, handler, "debate-1", 5)
		err := orchestrator.Emit(StreamEventDebateStart, nil, "Starting debate", nil)
		require.NoError(t, err)
		require.Len(t, receivedEvents, 1)
		assert.Equal(t, StreamEventDebateStart, receivedEvents[0].Type)
		assert.Equal(t, "debate-1", receivedEvents[0].DebateID)
		assert.Equal(t, "Starting debate", receivedEvents[0].Content)
	})

	t.Run("EmitAgentResponse", func(t *testing.T) {
		var receivedEvents []*StreamEvent
		handler := func(event *StreamEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		orchestrator := NewStreamOrchestrator(nil, handler, "debate-1", 5)
		orchestrator.SetPhase("generation")
		orchestrator.SetTeam(TeamImplementation)

		agent := &Agent{
			ID:       "agent-1",
			Name:     "Test Agent",
			Role:     RoleGenerator,
			Provider: "openai",
			Model:    "gpt-4",
		}

		err := orchestrator.Emit(StreamEventAgentResponse, agent, "Generated code", map[string]interface{}{
			"confidence": 0.95,
		})
		require.NoError(t, err)
		require.Len(t, receivedEvents, 1)

		event := receivedEvents[0]
		assert.Equal(t, StreamEventAgentResponse, event.Type)
		assert.Equal(t, "generation", event.Phase)
		assert.Equal(t, TeamImplementation, event.Team)
		assert.Equal(t, "Generated code", event.Content)
		assert.NotNil(t, event.Agent)
		assert.Equal(t, "agent-1", event.Agent.ID)
		assert.Equal(t, RoleGenerator, event.Agent.Role)
	})

	t.Run("EmitError", func(t *testing.T) {
		var receivedEvents []*StreamEvent
		handler := func(event *StreamEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		orchestrator := NewStreamOrchestrator(nil, handler, "debate-1", 5)
		testErr := assert.AnError

		err := orchestrator.EmitError(nil, testErr)
		require.NoError(t, err)
		require.Len(t, receivedEvents, 1)
		assert.Equal(t, StreamEventDebateError, receivedEvents[0].Type)
		assert.Equal(t, testErr.Error(), receivedEvents[0].Error)
	})

	t.Run("ProgressCalculation", func(t *testing.T) {
		orchestrator := NewStreamOrchestrator(nil, nil, "debate-1", 10)
		assert.Equal(t, float64(0), orchestrator.calculateProgress())

		orchestrator.IncrementCompleted()
		assert.Equal(t, float64(10), orchestrator.calculateProgress())

		orchestrator.IncrementCompleted()
		assert.Equal(t, float64(20), orchestrator.calculateProgress())
	})

	t.Run("GetDuration", func(t *testing.T) {
		orchestrator := NewStreamOrchestrator(nil, nil, "debate-1", 5)
		time.Sleep(10 * time.Millisecond)
		duration := orchestrator.GetDuration()
		assert.True(t, duration > 0)
	})
}

func TestStreamEventTypes(t *testing.T) {
	eventTypes := []StreamEventType{
		StreamEventDebateStart,
		StreamEventDebateComplete,
		StreamEventDebateError,
		StreamEventPhaseStart,
		StreamEventPhaseComplete,
		StreamEventTeamStart,
		StreamEventTeamComplete,
		StreamEventAgentStart,
		StreamEventAgentResponse,
		StreamEventAgentComplete,
		StreamEventToolCall,
		StreamEventToolResult,
		StreamEventConsensusUpdate,
		StreamEventConsensusReached,
	}

	for _, et := range eventTypes {
		t.Run(string(et), func(t *testing.T) {
			assert.NotEmpty(t, et)
		})
	}
}

func TestTeamAssignments(t *testing.T) {
	t.Run("GetTeamForRole", func(t *testing.T) {
		tests := []struct {
			role     Role
			expected Team
		}{
			{RoleArchitect, TeamDesign},
			{RoleModerator, TeamDesign},
			{RoleGenerator, TeamImplementation},
			{RoleBlueTeam, TeamImplementation},
			{RoleCritic, TeamQuality},
			{RoleTester, TeamQuality},
			{RoleValidator, TeamQuality},
			{RoleSecurity, TeamQuality},
			{RolePerformance, TeamQuality},
			{RoleRedTeam, TeamRedTeam},
			{RoleRefactoring, TeamRefactoring},
		}

		for _, tt := range tests {
			t.Run(string(tt.role), func(t *testing.T) {
				team := GetTeamForRole(tt.role)
				assert.Equal(t, tt.expected, team)
			})
		}
	})

	t.Run("GetRolesForTeam", func(t *testing.T) {
		tests := []struct {
			team         Team
			expectedLen  int
			expectedRole Role
		}{
			{TeamDesign, 2, RoleArchitect},
			{TeamImplementation, 2, RoleGenerator},
			{TeamQuality, 5, RoleCritic},
			{TeamRedTeam, 1, RoleRedTeam},
			{TeamRefactoring, 1, RoleRefactoring},
		}

		for _, tt := range tests {
			t.Run(string(tt.team), func(t *testing.T) {
				roles := GetRolesForTeam(tt.team)
				assert.Len(t, roles, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.Contains(t, roles, tt.expectedRole)
				}
			})
		}
	})

	t.Run("AssignTeamsToAgents", func(t *testing.T) {
		agents := []*Agent{
			NewAgent(RoleArchitect, "openai", "gpt-4", 8.5),
			NewAgent(RoleGenerator, "anthropic", "claude", 8.5),
			NewAgent(RoleCritic, "openai", "gpt-4", 8.0),
		}

		assignments := AssignTeamsToAgents(agents)
		require.Len(t, assignments, 3)

		assert.Equal(t, TeamDesign, assignments[0].Team)
		assert.Equal(t, TeamImplementation, assignments[1].Team)
		assert.Equal(t, TeamQuality, assignments[2].Team)
	})

	t.Run("GetTeamSummary", func(t *testing.T) {
		agents := []*Agent{
			NewAgent(RoleArchitect, "openai", "gpt-4", 8.5),
			NewAgent(RoleGenerator, "anthropic", "claude", 8.5),
			NewAgent(RoleCritic, "openai", "gpt-4", 8.0),
			NewAgent(RoleTester, "anthropic", "claude", 8.5),
		}

		assignments := AssignTeamsToAgents(agents)
		summary := GetTeamSummary(assignments)

		assert.Len(t, summary[TeamDesign], 1)
		assert.Len(t, summary[TeamImplementation], 1)
		assert.Len(t, summary[TeamQuality], 2)
	})
}

func TestDebateStreamRequest(t *testing.T) {
	req := &DebateStreamRequest{
		DebateRequest: &DebateRequest{
			ID:        "test-id",
			Topic:     "test topic",
			Context:   "test context",
			Language:  "go",
			MaxRounds: 5,
		},
		Stream: true,
	}

	assert.Equal(t, "test-id", req.ID)
	assert.Equal(t, "test topic", req.Topic)
	assert.True(t, req.Stream)
}

func TestDefaultStreamConfig(t *testing.T) {
	config := DefaultStreamConfig()
	require.NotNil(t, config)

	assert.Equal(t, 100, config.BufferSize)
	assert.Equal(t, 100*time.Millisecond, config.FlushInterval)
	assert.True(t, config.EnableProgress)
	assert.True(t, config.EnableTeams)
	assert.True(t, config.EnablePhases)
}

func TestAllTeams(t *testing.T) {
	teams := AllTeams()
	require.Len(t, teams, 5)
	assert.Contains(t, teams, TeamDesign)
	assert.Contains(t, teams, TeamImplementation)
	assert.Contains(t, teams, TeamQuality)
	assert.Contains(t, teams, TeamRedTeam)
	assert.Contains(t, teams, TeamRefactoring)
}

func TestIntegrationManagerStreamDebate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr, err := NewIntegrationManager(DefaultConfig(), logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = mgr.Initialize(ctx)
	require.NoError(t, err)

	t.Run("StreamDebateSuccess", func(t *testing.T) {
		var receivedEvents []*StreamEvent
		streamHandler := func(event *StreamEvent) error {
			receivedEvents = append(receivedEvents, event)
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "stream-test-1",
				Topic:     "Create a REST API",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		resp, err := mgr.StreamDebate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify we received events
		assert.Greater(t, len(receivedEvents), 0)

		// Verify debate completed successfully
		assert.True(t, resp.Success)
		assert.Greater(t, len(resp.Participants), 0)
		assert.NotNil(t, resp.Consensus)

		// Check for expected event types
		hasDebateStart := false
		hasDebateComplete := false
		hasAgentEvents := false
		hasTeamEvents := false

		for _, event := range receivedEvents {
			switch event.Type {
			case StreamEventDebateStart:
				hasDebateStart = true
			case StreamEventDebateComplete:
				hasDebateComplete = true
			case StreamEventAgentStart, StreamEventAgentResponse, StreamEventAgentComplete:
				hasAgentEvents = true
			case StreamEventTeamStart, StreamEventTeamComplete:
				hasTeamEvents = true
			}
		}

		assert.True(t, hasDebateStart, "Should have debate start event")
		assert.True(t, hasDebateComplete, "Should have debate complete event")
		assert.True(t, hasAgentEvents, "Should have agent events")
		assert.True(t, hasTeamEvents, "Should have team events")
	})

	t.Run("StreamDebateWithAllTeams", func(t *testing.T) {
		var teamsSeen []Team
		streamHandler := func(event *StreamEvent) error {
			if event.Type == StreamEventTeamStart {
				teamsSeen = append(teamsSeen, event.Team)
			}
			return nil
		}

		req := &DebateStreamRequest{
			DebateRequest: &DebateRequest{
				ID:        "stream-test-teams",
				Topic:     "Analyze code quality",
				Language:  "go",
				MaxRounds: 1,
			},
			Stream:        true,
			StreamHandler: streamHandler,
		}

		resp, err := mgr.StreamDebate(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify we saw teams
		assert.Greater(t, len(teamsSeen), 0)
	})
}
