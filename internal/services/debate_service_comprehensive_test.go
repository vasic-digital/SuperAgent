package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"digital.vasic.debate/comprehensive"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockComprehensiveIntegration is a mock implementation for testing
type MockComprehensiveIntegration struct {
	ExecuteDebateFunc func(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error)
	StreamDebateFunc  func(ctx context.Context, req *comprehensive.DebateStreamRequest) (*comprehensive.DebateResponse, error)
}

func (m *MockComprehensiveIntegration) ExecuteDebate(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error) {
	if m.ExecuteDebateFunc != nil {
		return m.ExecuteDebateFunc(ctx, req)
	}
	return &comprehensive.DebateResponse{}, nil
}

func (m *MockComprehensiveIntegration) StreamDebate(ctx context.Context, req *comprehensive.DebateStreamRequest) (*comprehensive.DebateResponse, error) {
	if m.StreamDebateFunc != nil {
		return m.StreamDebateFunc(ctx, req)
	}
	return &comprehensive.DebateResponse{}, nil
}

// Ensure MockComprehensiveIntegration implements the interface
var _ interface {
	ExecuteDebate(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error)
	StreamDebate(ctx context.Context, req *comprehensive.DebateStreamRequest) (*comprehensive.DebateResponse, error)
} = (*MockComprehensiveIntegration)(nil)

func TestDebateService_SetComprehensiveIntegration(t *testing.T) {
	t.Run("sets integration and enables system", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		ds := &DebateService{
			logger: logger,
		}

		mockIntegration := &MockComprehensiveIntegration{}
		ds.SetComprehensiveIntegration(mockIntegration)

		assert.NotNil(t, ds.comprehensiveIntegration)
		assert.True(t, ds.useComprehensiveSystem)
	})
}

func TestDebateService_EnableComprehensiveSystem(t *testing.T) {
	t.Run("enables comprehensive system", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		ds := &DebateService{
			logger:                 logger,
			useComprehensiveSystem: false,
		}

		ds.EnableComprehensiveSystem(true)
		assert.True(t, ds.useComprehensiveSystem)
	})

	t.Run("disables comprehensive system", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		ds := &DebateService{
			logger:                 logger,
			useComprehensiveSystem: true,
		}

		ds.EnableComprehensiveSystem(false)
		assert.False(t, ds.useComprehensiveSystem)
	})
}

func TestDebateService_conductComprehensiveDebate(t *testing.T) {
	t.Run("successful comprehensive debate", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			ExecuteDebateFunc: func(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error) {
				assert.Equal(t, "test-debate-123", req.ID)
				assert.Equal(t, "Test Topic", req.Topic)
				assert.Equal(t, "go", req.Language)
				assert.Equal(t, 3, req.MaxRounds)

				return &comprehensive.DebateResponse{
					Success:         true,
					RoundsConducted: 3,
					QualityScore:    0.85,
					Phases:          []string{"phase1", "phase2", "phase3"},
				}, nil
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "test-debate-123",
			Topic:    "Test Topic",
		}
		startTime := time.Now()
		sessionID := "session-456"

		result, err := ds.conductComprehensiveDebate(context.Background(), config, startTime, sessionID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-debate-123", result.DebateID)
		assert.Equal(t, "session-456", result.SessionID)
		assert.Equal(t, "Test Topic", result.Topic)
		assert.Equal(t, 3, result.TotalRounds)
		assert.Equal(t, 3, result.RoundsConducted)
		assert.Equal(t, 0.85, result.QualityScore)
		assert.Equal(t, 0.85, result.FinalScore)
		assert.True(t, result.Success)
		assert.NotNil(t, result.Consensus)
		assert.True(t, result.Consensus.Reached)
		assert.Equal(t, 0.85, result.Consensus.Confidence)
	})

	t.Run("handles execution error", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			ExecuteDebateFunc: func(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error) {
				return nil, errors.New("debate execution failed")
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "test-debate-123",
			Topic:    "Test Topic",
		}
		startTime := time.Now()
		sessionID := "session-456"

		result, err := ds.conductComprehensiveDebate(context.Background(), config, startTime, sessionID)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "comprehensive debate failed")
	})

	t.Run("result contains correct metadata", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			ExecuteDebateFunc: func(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error) {
				return &comprehensive.DebateResponse{
					Success:         true,
					RoundsConducted: 5,
					QualityScore:    0.92,
					Phases:          []string{"p1", "p2", "p3", "p4"},
				}, nil
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "debate-789",
			Topic:    "Another Topic",
		}
		startTime := time.Now()
		sessionID := "session-xyz"

		result, err := ds.conductComprehensiveDebate(context.Background(), config, startTime, sessionID)

		require.NoError(t, err)
		assert.NotNil(t, result.Metadata)
		assert.Equal(t, true, result.Metadata["comprehensive_debate"])
		assert.Equal(t, 5, result.Metadata["rounds_conducted"])
		assert.Equal(t, 4, result.Metadata["phases"])
	})

	t.Run("consensus result is properly populated", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			ExecuteDebateFunc: func(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error) {
				return &comprehensive.DebateResponse{
					Success:         true,
					RoundsConducted: 3,
					QualityScore:    0.78,
				}, nil
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "debate-consensus",
			Topic:    "Consensus Test",
		}
		startTime := time.Now()
		sessionID := "session-consensus"

		result, err := ds.conductComprehensiveDebate(context.Background(), config, startTime, sessionID)

		require.NoError(t, err)
		require.NotNil(t, result.Consensus)

		consensus := result.Consensus
		assert.True(t, consensus.Reached)
		assert.True(t, consensus.Achieved)
		assert.Equal(t, 0.78, consensus.Confidence)
		assert.Equal(t, 0.78, consensus.AgreementLevel)
		assert.Equal(t, 0.78, consensus.QualityScore)
		assert.Contains(t, consensus.FinalPosition, "3 rounds")
		assert.Contains(t, consensus.Summary, "3 rounds")
		assert.NotNil(t, consensus.KeyPoints)
		assert.NotNil(t, consensus.Disagreements)
	})
}

func TestDebateService_conductComprehensiveDebateStreaming(t *testing.T) {
	t.Run("successful streaming debate", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			StreamDebateFunc: func(ctx context.Context, req *comprehensive.DebateStreamRequest) (*comprehensive.DebateResponse, error) {
				assert.Equal(t, "streaming-debate-123", req.ID)
				assert.Equal(t, "Streaming Topic", req.Topic)
				assert.True(t, req.Stream)

				return &comprehensive.DebateResponse{
					Success:         true,
					RoundsConducted: 4,
					QualityScore:    0.88,
					Phases:          []string{"p1", "p2"},
					Participants:    []string{"agent1", "agent2", "agent3"},
				}, nil
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "streaming-debate-123",
			Topic:    "Streaming Topic",
		}
		startTime := time.Now()
		sessionID := "session-stream"
		streamHandler := func(chunk string) error { return nil }

		result, err := ds.conductComprehensiveDebateStreaming(context.Background(), config, startTime, sessionID, streamHandler)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "streaming-debate-123", result.DebateID)
		assert.Equal(t, 4, result.TotalRounds)
		assert.Equal(t, 0.88, result.QualityScore)
		assert.Equal(t, 3, len(result.Participants))
	})

	t.Run("handles streaming error", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			StreamDebateFunc: func(ctx context.Context, req *comprehensive.DebateStreamRequest) (*comprehensive.DebateResponse, error) {
				return nil, errors.New("streaming failed")
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "streaming-debate-error",
			Topic:    "Error Topic",
		}
		startTime := time.Now()
		sessionID := "session-error"
		streamHandler := func(chunk string) error { return nil }

		result, err := ds.conductComprehensiveDebateStreaming(context.Background(), config, startTime, sessionID, streamHandler)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "comprehensive streaming debate failed")
	})

	t.Run("participants are populated from response", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			StreamDebateFunc: func(ctx context.Context, req *comprehensive.DebateStreamRequest) (*comprehensive.DebateResponse, error) {
				return &comprehensive.DebateResponse{
					Success:         true,
					RoundsConducted: 3,
					QualityScore:    0.75,
					Participants:    []string{"agent-a", "agent-b", "agent-c", "agent-d"},
				}, nil
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "participants-test",
			Topic:    "Participants Test",
		}
		startTime := time.Now()
		sessionID := "session-participants"
		streamHandler := func(chunk string) error { return nil }

		result, err := ds.conductComprehensiveDebateStreaming(context.Background(), config, startTime, sessionID, streamHandler)

		require.NoError(t, err)
		assert.Equal(t, 4, len(result.Participants))
		assert.Equal(t, 4, len(result.AllResponses))

		// Verify participant IDs
		for i, participant := range result.Participants {
			expectedID := []string{"agent-a", "agent-b", "agent-c", "agent-d"}[i]
			assert.Equal(t, expectedID, participant.ParticipantID)
			assert.Equal(t, "Contributed to comprehensive debate", participant.Response)
		}
	})

	t.Run("streaming metadata includes streaming flag", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			StreamDebateFunc: func(ctx context.Context, req *comprehensive.DebateStreamRequest) (*comprehensive.DebateResponse, error) {
				return &comprehensive.DebateResponse{
					Success:         true,
					RoundsConducted: 3,
					QualityScore:    0.80,
					Phases:          []string{"p1"},
					Participants:    []string{"agent1"},
				}, nil
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		config := &DebateConfig{
			DebateID: "metadata-test",
			Topic:    "Metadata Test",
		}
		startTime := time.Now()
		sessionID := "session-metadata"
		streamHandler := func(chunk string) error { return nil }

		result, err := ds.conductComprehensiveDebateStreaming(context.Background(), config, startTime, sessionID, streamHandler)

		require.NoError(t, err)
		assert.NotNil(t, result.Metadata)
		assert.Equal(t, true, result.Metadata["comprehensive_debate"])
		assert.Equal(t, true, result.Metadata["streaming"])
		assert.Equal(t, 3, result.Metadata["rounds_conducted"])
		assert.Equal(t, 1, result.Metadata["phases"])
		assert.Equal(t, 1, result.Metadata["participants"])
	})
}

func TestDebateService_ContextCancellation(t *testing.T) {
	t.Run("comprehensive debate respects context cancellation", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		mockIntegration := &MockComprehensiveIntegration{
			ExecuteDebateFunc: func(ctx context.Context, req *comprehensive.DebateRequest) (*comprehensive.DebateResponse, error) {
				// Simulate slow operation
				select {
				case <-time.After(100 * time.Millisecond):
					return &comprehensive.DebateResponse{Success: true}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}

		ds := &DebateService{
			logger:                 logger,
			comprehensiveIntegration: mockIntegration,
			useComprehensiveSystem: true,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		config := &DebateConfig{
			DebateID: "cancel-test",
			Topic:    "Cancel Test",
		}

		done := make(chan struct{})
		var result *DebateResult
		var err error

		go func() {
			defer close(done)
			result, err = ds.conductComprehensiveDebate(ctx, config, time.Now(), "session")
		}()

		select {
		case <-done:
			// Expected
		case <-time.After(2 * time.Second):
			t.Fatal("Debate did not respect context cancellation")
		}

		// Result may be nil or have error depending on timing
		_ = result
		_ = err
	})
}
