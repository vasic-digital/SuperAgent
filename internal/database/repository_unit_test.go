package database

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Additional Comprehensive Unit Tests for Repository Data Models
// These tests run without database access and test model serialization,
// field validation, and helper functions
// =============================================================================

// -----------------------------------------------------------------------------
// User Model Tests
// -----------------------------------------------------------------------------

func TestUser_JSONSerialization_Extended(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	user := &User{
		ID:           "user-123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed",
		APIKey:       "sk-test-key",
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var decoded User
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, user.ID, decoded.ID)
	assert.Equal(t, user.Username, decoded.Username)
	assert.Equal(t, user.Email, decoded.Email)
	assert.Equal(t, user.Role, decoded.Role)
}

func TestUser_DefaultValues_Extended(t *testing.T) {
	user := &User{}

	assert.Empty(t, user.ID)
	assert.Empty(t, user.Username)
	assert.Empty(t, user.Email)
	assert.Empty(t, user.PasswordHash)
	assert.Empty(t, user.APIKey)
	assert.Empty(t, user.Role)
	assert.True(t, user.CreatedAt.IsZero())
	assert.True(t, user.UpdatedAt.IsZero())
}

func TestUser_Roles_Extended(t *testing.T) {
	roles := []string{"admin", "user", "guest", "service", "api"}
	for _, role := range roles {
		user := &User{Role: role}
		assert.Equal(t, role, user.Role)
	}
}

func TestUser_JSONKeys_Extended(t *testing.T) {
	user := &User{
		ID:       "id",
		Username: "user",
		Email:    "email@test.com",
		Role:     "admin",
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	jsonStr := string(data)
	expectedKeys := []string{`"id":`, `"username":`, `"email":`, `"role":`}
	for _, key := range expectedKeys {
		assert.Contains(t, jsonStr, key)
	}
}

func TestNewUserRepository_Extended(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		repo := NewUserRepository(nil, logger)
		assert.NotNil(t, repo)
	})

	t.Run("WithNilLogger", func(t *testing.T) {
		repo := NewUserRepository(nil, nil)
		assert.NotNil(t, repo)
	})
}

// -----------------------------------------------------------------------------
// LLMProvider Model Tests
// -----------------------------------------------------------------------------

func TestLLMProvider_JSONSerialization_Extended(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	provider := &LLMProvider{
		ID:           "provider-123",
		Name:         "OpenAI",
		Type:         "openai",
		APIKey:       "sk-test",
		BaseURL:      "https://api.openai.com",
		Model:        "gpt-4",
		Weight:       1.5,
		Enabled:      true,
		Config:       map[string]interface{}{"temperature": 0.7},
		HealthStatus: "healthy",
		ResponseTime: 150,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, err := json.Marshal(provider)
	require.NoError(t, err)

	var decoded LLMProvider
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, provider.ID, decoded.ID)
	assert.Equal(t, provider.Name, decoded.Name)
	assert.Equal(t, provider.Type, decoded.Type)
	assert.Equal(t, provider.Weight, decoded.Weight)
	assert.Equal(t, provider.Enabled, decoded.Enabled)
	assert.Equal(t, "healthy", decoded.HealthStatus)
}

func TestLLMProvider_DefaultValues_Extended(t *testing.T) {
	provider := &LLMProvider{}

	assert.Empty(t, provider.ID)
	assert.Empty(t, provider.Name)
	assert.Empty(t, provider.Type)
	assert.Empty(t, provider.APIKey)
	assert.Empty(t, provider.BaseURL)
	assert.Empty(t, provider.Model)
	assert.Equal(t, float64(0), provider.Weight)
	assert.False(t, provider.Enabled)
	assert.Nil(t, provider.Config)
	assert.Empty(t, provider.HealthStatus)
	assert.Equal(t, int64(0), provider.ResponseTime)
}

func TestLLMProvider_HealthStatuses_Extended(t *testing.T) {
	statuses := []string{"healthy", "unhealthy", "unknown", "degraded", "error"}
	for _, status := range statuses {
		provider := &LLMProvider{HealthStatus: status}
		assert.Equal(t, status, provider.HealthStatus)
	}
}

func TestLLMProvider_WeightRange_Extended(t *testing.T) {
	testCases := []float64{0.0, 0.5, 1.0, 1.5, 2.0, 5.0, 10.0}
	for _, weight := range testCases {
		provider := &LLMProvider{Weight: weight}
		assert.Equal(t, weight, provider.Weight)
	}
}

func TestNewProviderRepository_Extended(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		repo := NewProviderRepository(nil, logger)
		assert.NotNil(t, repo)
	})

	t.Run("WithNilLogger", func(t *testing.T) {
		repo := NewProviderRepository(nil, nil)
		assert.NotNil(t, repo)
	})
}

// -----------------------------------------------------------------------------
// LLMRequest Model Tests
// -----------------------------------------------------------------------------

func TestLLMRequest_JSONSerialization_Extended(t *testing.T) {
	sessionID := "session-123"
	userID := "user-456"
	startedAt := time.Now()
	completedAt := time.Now().Add(time.Second)

	request := &LLMRequest{
		ID:             "request-123",
		SessionID:      &sessionID,
		UserID:         &userID,
		Prompt:         "Hello, world!",
		Messages:       []map[string]string{{"role": "user", "content": "Hi"}},
		ModelParams:    map[string]interface{}{"temperature": 0.7},
		EnsembleConfig: map[string]interface{}{"strategy": "vote"},
		MemoryEnhanced: true,
		Memory:         map[string]interface{}{"context": "test"},
		Status:         "completed",
		RequestType:    "chat",
		CreatedAt:      time.Now(),
		StartedAt:      &startedAt,
		CompletedAt:    &completedAt,
	}

	data, err := json.Marshal(request)
	require.NoError(t, err)

	var decoded LLMRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, request.ID, decoded.ID)
	assert.Equal(t, request.Prompt, decoded.Prompt)
	assert.Equal(t, request.Status, decoded.Status)
	assert.Equal(t, request.MemoryEnhanced, decoded.MemoryEnhanced)
}

func TestLLMRequest_DefaultValues_Extended(t *testing.T) {
	request := &LLMRequest{}

	assert.Empty(t, request.ID)
	assert.Nil(t, request.SessionID)
	assert.Nil(t, request.UserID)
	assert.Empty(t, request.Prompt)
	assert.Nil(t, request.Messages)
	assert.Nil(t, request.ModelParams)
	assert.Nil(t, request.EnsembleConfig)
	assert.False(t, request.MemoryEnhanced)
	assert.Nil(t, request.Memory)
	assert.Empty(t, request.Status)
	assert.Empty(t, request.RequestType)
	assert.Nil(t, request.StartedAt)
	assert.Nil(t, request.CompletedAt)
}

func TestLLMRequest_StatusValues_Extended(t *testing.T) {
	statuses := []string{"pending", "processing", "completed", "failed", "cancelled", "timeout"}
	for _, status := range statuses {
		request := &LLMRequest{Status: status}
		assert.Equal(t, status, request.Status)
	}
}

func TestLLMRequest_RequestTypes_Extended(t *testing.T) {
	types := []string{"chat", "completion", "embedding", "vision", "debate", "streaming"}
	for _, reqType := range types {
		request := &LLMRequest{RequestType: reqType}
		assert.Equal(t, reqType, request.RequestType)
	}
}

func TestNewRequestRepository_Extended(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		repo := NewRequestRepository(nil, logger)
		assert.NotNil(t, repo)
	})
}

// -----------------------------------------------------------------------------
// LLMResponse Model Tests
// -----------------------------------------------------------------------------

func TestLLMResponse_JSONSerialization_Extended(t *testing.T) {
	providerID := "provider-123"
	response := &LLMResponse{
		ID:             "response-123",
		RequestID:      "request-456",
		ProviderID:     &providerID,
		ProviderName:   "OpenAI",
		Content:        "Hello! How can I help you?",
		Confidence:     0.95,
		TokensUsed:     50,
		ResponseTime:   200,
		FinishReason:   "stop",
		Metadata:       map[string]interface{}{"model": "gpt-4"},
		Selected:       true,
		SelectionScore: 0.98,
		CreatedAt:      time.Now(),
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded LLMResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.ID, decoded.ID)
	assert.Equal(t, response.RequestID, decoded.RequestID)
	assert.Equal(t, response.Content, decoded.Content)
	assert.Equal(t, response.Confidence, decoded.Confidence)
	assert.Equal(t, response.Selected, decoded.Selected)
}

func TestLLMResponse_DefaultValues_Extended(t *testing.T) {
	response := &LLMResponse{}

	assert.Empty(t, response.ID)
	assert.Empty(t, response.RequestID)
	assert.Nil(t, response.ProviderID)
	assert.Empty(t, response.ProviderName)
	assert.Empty(t, response.Content)
	assert.Equal(t, float64(0), response.Confidence)
	assert.Equal(t, 0, response.TokensUsed)
	assert.Equal(t, int64(0), response.ResponseTime)
	assert.Empty(t, response.FinishReason)
	assert.Nil(t, response.Metadata)
	assert.False(t, response.Selected)
	assert.Equal(t, float64(0), response.SelectionScore)
}

func TestLLMResponse_FinishReasons_Extended(t *testing.T) {
	reasons := []string{"stop", "length", "content_filter", "function_call", "tool_calls", "error"}
	for _, reason := range reasons {
		response := &LLMResponse{FinishReason: reason}
		assert.Equal(t, reason, response.FinishReason)
	}
}

func TestLLMResponse_ConfidenceRange_Extended(t *testing.T) {
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 0.9, 0.95, 1.0}
	for _, confidence := range testCases {
		response := &LLMResponse{Confidence: confidence}
		assert.Equal(t, confidence, response.Confidence)
	}
}

func TestNewResponseRepository_Extended(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		repo := NewResponseRepository(nil, logger)
		assert.NotNil(t, repo)
	})
}

// -----------------------------------------------------------------------------
// UserSession Model Tests
// -----------------------------------------------------------------------------

func TestUserSession_JSONSerialization_Extended(t *testing.T) {
	memoryID := "memory-123"
	now := time.Now().Truncate(time.Second)

	session := &UserSession{
		ID:           "session-123",
		UserID:       "user-456",
		SessionToken: "token-789",
		Context:      map[string]interface{}{"theme": "dark"},
		MemoryID:     &memoryID,
		Status:       "active",
		RequestCount: 10,
		LastActivity: now,
		ExpiresAt:    now.Add(24 * time.Hour),
		CreatedAt:    now,
	}

	data, err := json.Marshal(session)
	require.NoError(t, err)

	var decoded UserSession
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, session.ID, decoded.ID)
	assert.Equal(t, session.UserID, decoded.UserID)
	assert.Equal(t, session.SessionToken, decoded.SessionToken)
	assert.Equal(t, session.Status, decoded.Status)
	assert.Equal(t, session.RequestCount, decoded.RequestCount)
}

func TestUserSession_StatusValues_Extended(t *testing.T) {
	statuses := []string{"active", "suspended", "terminated", "expired", "locked"}
	for _, status := range statuses {
		session := &UserSession{Status: status}
		assert.Equal(t, status, session.Status)
	}
}

func TestUserSession_ContextTypes_Extended(t *testing.T) {
	t.Run("EmptyContext", func(t *testing.T) {
		session := &UserSession{Context: map[string]interface{}{}}
		assert.NotNil(t, session.Context)
		assert.Len(t, session.Context, 0)
	})

	t.Run("SimpleContext", func(t *testing.T) {
		session := &UserSession{
			Context: map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
		}
		assert.Equal(t, "dark", session.Context["theme"])
		assert.Equal(t, "en", session.Context["language"])
	})

	t.Run("NestedContext", func(t *testing.T) {
		session := &UserSession{
			Context: map[string]interface{}{
				"preferences": map[string]interface{}{
					"theme":    "dark",
					"fontSize": 14,
				},
			},
		}
		assert.NotNil(t, session.Context["preferences"])
	})
}

func TestNewSessionRepository_Extended(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		repo := NewSessionRepository(nil, logger)
		assert.NotNil(t, repo)
	})
}

// -----------------------------------------------------------------------------
// Edge Cases and Boundary Tests
// -----------------------------------------------------------------------------

func TestUser_LongUsername(t *testing.T) {
	longUsername := ""
	for i := 0; i < 256; i++ {
		longUsername += "a"
	}
	user := &User{Username: longUsername}
	assert.Len(t, user.Username, 256)
}

func TestLLMProvider_EmptyConfig(t *testing.T) {
	provider := &LLMProvider{Config: map[string]interface{}{}}
	assert.NotNil(t, provider.Config)
	assert.Len(t, provider.Config, 0)
}

func TestLLMRequest_EmptyMessages(t *testing.T) {
	request := &LLMRequest{Messages: []map[string]string{}}
	assert.NotNil(t, request.Messages)
	assert.Len(t, request.Messages, 0)
}

func TestLLMResponse_ZeroTokens(t *testing.T) {
	response := &LLMResponse{TokensUsed: 0}
	assert.Equal(t, 0, response.TokensUsed)
}

func TestLLMResponse_LargeTokens(t *testing.T) {
	response := &LLMResponse{TokensUsed: 128000}
	assert.Equal(t, 128000, response.TokensUsed)
}

func TestUserSession_NilMemoryID(t *testing.T) {
	session := &UserSession{MemoryID: nil}
	assert.Nil(t, session.MemoryID)
}

func TestUserSession_ZeroRequestCount(t *testing.T) {
	session := &UserSession{RequestCount: 0}
	assert.Equal(t, 0, session.RequestCount)
}

func TestUserSession_HighRequestCount(t *testing.T) {
	session := &UserSession{RequestCount: 1000000}
	assert.Equal(t, 1000000, session.RequestCount)
}

func TestLLMRequest_NilPointers(t *testing.T) {
	request := &LLMRequest{
		SessionID:   nil,
		UserID:      nil,
		StartedAt:   nil,
		CompletedAt: nil,
	}

	assert.Nil(t, request.SessionID)
	assert.Nil(t, request.UserID)
	assert.Nil(t, request.StartedAt)
	assert.Nil(t, request.CompletedAt)
}

func TestLLMResponse_NilProviderID(t *testing.T) {
	response := &LLMResponse{ProviderID: nil}
	assert.Nil(t, response.ProviderID)
}

// -----------------------------------------------------------------------------
// JSON Round-Trip Tests
// -----------------------------------------------------------------------------

func TestUser_JSONRoundTrip_Extended(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &User{
		ID:           "user-round-trip",
		Username:     "roundtrip",
		Email:        "roundtrip@test.com",
		PasswordHash: "hash123",
		APIKey:       "sk-roundtrip",
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded User
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Username, decoded.Username)
	assert.Equal(t, original.Email, decoded.Email)
	assert.Equal(t, original.Role, decoded.Role)
}

func TestLLMProvider_JSONRoundTrip_Extended(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &LLMProvider{
		ID:           "provider-round-trip",
		Name:         "RoundTrip Provider",
		Type:         "openai",
		APIKey:       "sk-test",
		BaseURL:      "https://api.test.com",
		Model:        "gpt-4",
		Weight:       1.5,
		Enabled:      true,
		Config:       map[string]interface{}{"key": "value"},
		HealthStatus: "healthy",
		ResponseTime: 100,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded LLMProvider
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Weight, decoded.Weight)
	assert.Equal(t, original.Enabled, decoded.Enabled)
}

func TestLLMRequest_JSONRoundTrip_Extended(t *testing.T) {
	sessionID := "session-rt"
	userID := "user-rt"
	startedAt := time.Now().Truncate(time.Second)

	original := &LLMRequest{
		ID:             "request-rt",
		SessionID:      &sessionID,
		UserID:         &userID,
		Prompt:         "Test prompt",
		Messages:       []map[string]string{{"role": "user", "content": "test"}},
		ModelParams:    map[string]interface{}{"temp": 0.5},
		EnsembleConfig: map[string]interface{}{"strategy": "vote"},
		MemoryEnhanced: true,
		Memory:         map[string]interface{}{"key": "val"},
		Status:         "completed",
		RequestType:    "chat",
		CreatedAt:      startedAt,
		StartedAt:      &startedAt,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded LLMRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Prompt, decoded.Prompt)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.MemoryEnhanced, decoded.MemoryEnhanced)
}

func TestLLMResponse_JSONRoundTrip_Extended(t *testing.T) {
	providerID := "provider-rt"
	now := time.Now().Truncate(time.Second)

	original := &LLMResponse{
		ID:             "response-rt",
		RequestID:      "request-rt",
		ProviderID:     &providerID,
		ProviderName:   "TestProvider",
		Content:        "Test response content",
		Confidence:     0.92,
		TokensUsed:     75,
		ResponseTime:   250,
		FinishReason:   "stop",
		Metadata:       map[string]interface{}{"model": "test"},
		Selected:       true,
		SelectionScore: 0.95,
		CreatedAt:      now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded LLMResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Content, decoded.Content)
	assert.Equal(t, original.Confidence, decoded.Confidence)
	assert.Equal(t, original.Selected, decoded.Selected)
}

func TestUserSession_JSONRoundTrip_Extended(t *testing.T) {
	memoryID := "memory-rt"
	now := time.Now().Truncate(time.Second)

	original := &UserSession{
		ID:           "session-rt",
		UserID:       "user-rt",
		SessionToken: "token-rt",
		Context:      map[string]interface{}{"key": "value"},
		MemoryID:     &memoryID,
		Status:       "active",
		RequestCount: 5,
		LastActivity: now,
		ExpiresAt:    now.Add(24 * time.Hour),
		CreatedAt:    now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded UserSession
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.UserID, decoded.UserID)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.RequestCount, decoded.RequestCount)
}

// -----------------------------------------------------------------------------
// DebateLogEntry Model Tests
// -----------------------------------------------------------------------------

func TestDebateLogEntry_JSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	expiresAt := now.Add(5 * 24 * time.Hour)

	entry := &DebateLogEntry{
		ID:                    123,
		DebateID:              "debate-123",
		SessionID:             "session-456",
		ParticipantID:         "part-789",
		ParticipantIdentifier: "DeepSeek-1",
		ParticipantName:       "DeepSeek Participant",
		Role:                  "debater",
		Provider:              "deepseek",
		Model:                 "deepseek-chat",
		Round:                 2,
		Action:                "complete",
		ResponseTimeMs:        1500,
		QualityScore:          0.92,
		TokensUsed:            250,
		ContentLength:         1200,
		ErrorMessage:          "",
		Metadata:              `{"key":"value"}`,
		CreatedAt:             now,
		ExpiresAt:             &expiresAt,
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var decoded DebateLogEntry
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, entry.ID, decoded.ID)
	assert.Equal(t, entry.DebateID, decoded.DebateID)
	assert.Equal(t, entry.ParticipantIdentifier, decoded.ParticipantIdentifier)
	assert.Equal(t, entry.QualityScore, decoded.QualityScore)
	assert.Equal(t, entry.Round, decoded.Round)
}

func TestDebateLogEntry_DefaultValues(t *testing.T) {
	entry := &DebateLogEntry{}

	assert.Equal(t, int64(0), entry.ID)
	assert.Empty(t, entry.DebateID)
	assert.Empty(t, entry.SessionID)
	assert.Empty(t, entry.ParticipantID)
	assert.Empty(t, entry.ParticipantIdentifier)
	assert.Empty(t, entry.Provider)
	assert.Empty(t, entry.Model)
	assert.Equal(t, 0, entry.Round)
	assert.Empty(t, entry.Action)
	assert.Equal(t, int64(0), entry.ResponseTimeMs)
	assert.Equal(t, float64(0), entry.QualityScore)
	assert.Equal(t, 0, entry.TokensUsed)
	assert.Equal(t, 0, entry.ContentLength)
	assert.Nil(t, entry.ExpiresAt)
}

func TestDebateLogEntry_Actions(t *testing.T) {
	actions := []string{"start", "complete", "error", "timeout", "retry", "skip"}
	for _, action := range actions {
		entry := &DebateLogEntry{Action: action}
		assert.Equal(t, action, entry.Action)
	}
}

func TestDebateLogEntry_Roles(t *testing.T) {
	roles := []string{"debater", "moderator", "judge", "observer", "participant"}
	for _, role := range roles {
		entry := &DebateLogEntry{Role: role}
		assert.Equal(t, role, entry.Role)
	}
}

func TestDebateLogEntry_QualityScoreRange(t *testing.T) {
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 0.85, 0.95, 1.0}
	for _, score := range testCases {
		entry := &DebateLogEntry{QualityScore: score}
		assert.Equal(t, score, entry.QualityScore)
	}
}

// -----------------------------------------------------------------------------
// LogRetentionPolicy Tests
// -----------------------------------------------------------------------------

func TestDefaultRetentionPolicy(t *testing.T) {
	policy := DefaultRetentionPolicy()

	assert.Equal(t, 5, policy.RetentionDays)
	assert.False(t, policy.NoExpiration)
	assert.Equal(t, time.Duration(0), policy.RetentionTime)
}

func TestNoExpirationPolicy(t *testing.T) {
	policy := NoExpirationPolicy()

	assert.True(t, policy.NoExpiration)
	assert.Equal(t, 0, policy.RetentionDays)
}

func TestLogRetentionPolicy_CustomDays(t *testing.T) {
	policy := LogRetentionPolicy{
		RetentionDays: 30,
		NoExpiration:  false,
	}

	assert.Equal(t, 30, policy.RetentionDays)
	assert.False(t, policy.NoExpiration)
}

func TestLogRetentionPolicy_CustomDuration(t *testing.T) {
	policy := LogRetentionPolicy{
		RetentionTime: 7 * 24 * time.Hour,
		NoExpiration:  false,
	}

	assert.Equal(t, 7*24*time.Hour, policy.RetentionTime)
	assert.False(t, policy.NoExpiration)
}

// -----------------------------------------------------------------------------
// LogStats Model Tests
// -----------------------------------------------------------------------------

func TestLogStats_JSONSerialization(t *testing.T) {
	now := time.Now()
	stats := &LogStats{
		TotalLogs:          1000,
		PermanentLogs:      100,
		ExpiredLogs:        50,
		UniqueDebates:      25,
		UniqueParticipants: 10,
		OldestLog:          &now,
		NewestLog:          &now,
	}

	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var decoded LogStats
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, stats.TotalLogs, decoded.TotalLogs)
	assert.Equal(t, stats.PermanentLogs, decoded.PermanentLogs)
	assert.Equal(t, stats.ExpiredLogs, decoded.ExpiredLogs)
	assert.Equal(t, stats.UniqueDebates, decoded.UniqueDebates)
	assert.Equal(t, stats.UniqueParticipants, decoded.UniqueParticipants)
}

func TestLogStats_DefaultValues(t *testing.T) {
	stats := &LogStats{}

	assert.Equal(t, int64(0), stats.TotalLogs)
	assert.Equal(t, int64(0), stats.PermanentLogs)
	assert.Equal(t, int64(0), stats.ExpiredLogs)
	assert.Equal(t, int64(0), stats.UniqueDebates)
	assert.Equal(t, int64(0), stats.UniqueParticipants)
	assert.Nil(t, stats.OldestLog)
	assert.Nil(t, stats.NewestLog)
}

func TestNewDebateLogRepository(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		policy := DefaultRetentionPolicy()
		repo := NewDebateLogRepository(nil, logger, policy)
		assert.NotNil(t, repo)
	})

	t.Run("WithNilLogger", func(t *testing.T) {
		policy := DefaultRetentionPolicy()
		repo := NewDebateLogRepository(nil, nil, policy)
		assert.NotNil(t, repo)
	})

	t.Run("WithCustomPolicy", func(t *testing.T) {
		policy := LogRetentionPolicy{
			RetentionDays: 30,
			NoExpiration:  false,
		}
		repo := NewDebateLogRepository(nil, nil, policy)
		assert.NotNil(t, repo)
		assert.Equal(t, 30, repo.GetRetentionPolicy().RetentionDays)
	})
}

func TestDebateLogRepository_SetRetentionPolicy_Extended(t *testing.T) {
	repo := NewDebateLogRepository(nil, nil, DefaultRetentionPolicy())

	assert.Equal(t, 5, repo.GetRetentionPolicy().RetentionDays)

	newPolicy := LogRetentionPolicy{
		RetentionDays: 14,
		NoExpiration:  false,
	}
	repo.SetRetentionPolicy(newPolicy)

	assert.Equal(t, 14, repo.GetRetentionPolicy().RetentionDays)
}

func TestDebateLogRepository_GetRetentionPolicy(t *testing.T) {
	customPolicy := LogRetentionPolicy{
		RetentionDays: 10,
		NoExpiration:  false,
	}
	repo := NewDebateLogRepository(nil, nil, customPolicy)

	policy := repo.GetRetentionPolicy()

	assert.Equal(t, 10, policy.RetentionDays)
	assert.False(t, policy.NoExpiration)
}

// -----------------------------------------------------------------------------
// NewBackgroundTaskRepository Tests
// -----------------------------------------------------------------------------

func TestNewBackgroundTaskRepository_Extended(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		repo := NewBackgroundTaskRepository(nil, logger)
		assert.NotNil(t, repo)
	})

	t.Run("WithNilLogger", func(t *testing.T) {
		repo := NewBackgroundTaskRepository(nil, nil)
		assert.NotNil(t, repo)
	})
}

// -----------------------------------------------------------------------------
// JSON Tag Tests for Models
// -----------------------------------------------------------------------------

func TestUser_JSONTags(t *testing.T) {
	user := &User{
		ID:           "id-test",
		Username:     "user-test",
		Email:        "email@test.com",
		PasswordHash: "hash-test",
		APIKey:       "api-key-test",
		Role:         "role-test",
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	jsonStr := string(data)
	// Verify snake_case JSON keys
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"username"`)
	assert.Contains(t, jsonStr, `"email"`)
	// Note: password_hash is intentionally excluded from JSON (json:"-") for security
	assert.NotContains(t, jsonStr, `"password_hash"`)
	assert.Contains(t, jsonStr, `"api_key"`)
	assert.Contains(t, jsonStr, `"role"`)
}

func TestLLMProvider_JSONTags(t *testing.T) {
	provider := &LLMProvider{
		ID:           "id-test",
		Name:         "name-test",
		Type:         "type-test",
		APIKey:       "key-test",
		BaseURL:      "url-test",
		Model:        "model-test",
		Weight:       1.0,
		Enabled:      true,
		HealthStatus: "healthy",
		ResponseTime: 100,
	}

	data, err := json.Marshal(provider)
	require.NoError(t, err)

	jsonStr := string(data)
	// Verify snake_case JSON keys
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"name"`)
	assert.Contains(t, jsonStr, `"type"`)
	assert.Contains(t, jsonStr, `"base_url"`)
	assert.Contains(t, jsonStr, `"model"`)
	assert.Contains(t, jsonStr, `"weight"`)
	assert.Contains(t, jsonStr, `"enabled"`)
	assert.Contains(t, jsonStr, `"health_status"`)
	assert.Contains(t, jsonStr, `"response_time"`)
}

func TestLLMRequest_JSONTags(t *testing.T) {
	request := &LLMRequest{
		ID:             "id-test",
		Prompt:         "prompt-test",
		Status:         "status-test",
		RequestType:    "type-test",
		MemoryEnhanced: true,
	}

	data, err := json.Marshal(request)
	require.NoError(t, err)

	jsonStr := string(data)
	// Verify snake_case JSON keys
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"prompt"`)
	assert.Contains(t, jsonStr, `"status"`)
	assert.Contains(t, jsonStr, `"request_type"`)
	assert.Contains(t, jsonStr, `"memory_enhanced"`)
}

func TestLLMResponse_JSONTags(t *testing.T) {
	response := &LLMResponse{
		ID:             "id-test",
		RequestID:      "req-test",
		ProviderName:   "provider-test",
		Content:        "content-test",
		Confidence:     0.9,
		TokensUsed:     100,
		ResponseTime:   200,
		FinishReason:   "stop",
		Selected:       true,
		SelectionScore: 0.95,
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	jsonStr := string(data)
	// Verify snake_case JSON keys
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"request_id"`)
	assert.Contains(t, jsonStr, `"provider_name"`)
	assert.Contains(t, jsonStr, `"content"`)
	assert.Contains(t, jsonStr, `"confidence"`)
	assert.Contains(t, jsonStr, `"tokens_used"`)
	assert.Contains(t, jsonStr, `"response_time"`)
	assert.Contains(t, jsonStr, `"finish_reason"`)
	assert.Contains(t, jsonStr, `"selected"`)
	assert.Contains(t, jsonStr, `"selection_score"`)
}

func TestUserSession_JSONTags(t *testing.T) {
	session := &UserSession{
		ID:           "id-test",
		UserID:       "user-test",
		SessionToken: "token-test",
		Status:       "status-test",
		RequestCount: 5,
	}

	data, err := json.Marshal(session)
	require.NoError(t, err)

	jsonStr := string(data)
	// Verify snake_case JSON keys
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"user_id"`)
	assert.Contains(t, jsonStr, `"session_token"`)
	assert.Contains(t, jsonStr, `"status"`)
	assert.Contains(t, jsonStr, `"request_count"`)
}

// -----------------------------------------------------------------------------
// VectorDocument Model Tests
// -----------------------------------------------------------------------------

func TestVectorDocument_JSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	embeddingID := "emb-123"
	doc := &VectorDocument{
		ID:                "doc-123",
		Title:             "Test Document",
		Content:           "This is test content for embedding",
		Metadata:          json.RawMessage(`{"author":"test","tags":["go","test"]}`),
		EmbeddingID:       &embeddingID,
		EmbeddingProvider: "openai",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	var decoded VectorDocument
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, doc.ID, decoded.ID)
	assert.Equal(t, doc.Title, decoded.Title)
	assert.Equal(t, doc.Content, decoded.Content)
	assert.Equal(t, doc.EmbeddingProvider, decoded.EmbeddingProvider)
	assert.NotNil(t, decoded.EmbeddingID)
	assert.Equal(t, *doc.EmbeddingID, *decoded.EmbeddingID)
}

func TestVectorDocument_DefaultValues(t *testing.T) {
	doc := &VectorDocument{}

	assert.Empty(t, doc.ID)
	assert.Empty(t, doc.Title)
	assert.Empty(t, doc.Content)
	assert.Nil(t, doc.Metadata)
	assert.Nil(t, doc.EmbeddingID)
	assert.Empty(t, doc.EmbeddingProvider)
	assert.True(t, doc.CreatedAt.IsZero())
	assert.True(t, doc.UpdatedAt.IsZero())
}

func TestVectorDocument_JSONTags(t *testing.T) {
	doc := &VectorDocument{
		ID:                "id-test",
		Title:             "title-test",
		Content:           "content-test",
		EmbeddingProvider: "provider-test",
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"title"`)
	assert.Contains(t, jsonStr, `"content"`)
	assert.Contains(t, jsonStr, `"embedding_provider"`)
}

func TestVectorDocument_WithNilEmbeddingID(t *testing.T) {
	doc := &VectorDocument{
		ID:          "doc-456",
		Title:       "No Embedding",
		EmbeddingID: nil,
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	// embedding_id should be omitted when nil (omitempty)
	jsonStr := string(data)
	assert.NotContains(t, jsonStr, `"embedding_id":null`)
}

func TestVectorSearchResult_Fields_Extended(t *testing.T) {
	now := time.Now()
	result := VectorSearchResult{
		Document: VectorDocument{
			ID:                "doc-789",
			Title:             "Search Result",
			Content:           "Matching content",
			EmbeddingProvider: "openai",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		Similarity: 0.95,
	}

	assert.Equal(t, "doc-789", result.Document.ID)
	assert.Equal(t, "Search Result", result.Document.Title)
	assert.InDelta(t, 0.95, result.Similarity, 0.01)
}

func TestVectorDocumentFilter_Fields(t *testing.T) {
	filter := VectorDocumentFilter{
		Provider:  "openai",
		TitleLike: "test",
		Limit:     10,
		Offset:    5,
	}

	assert.Equal(t, "openai", filter.Provider)
	assert.Equal(t, "test", filter.TitleLike)
	assert.Equal(t, 10, filter.Limit)
	assert.Equal(t, 5, filter.Offset)
}

func TestNewVectorDocumentRepository_Extended(t *testing.T) {
	repo := NewVectorDocumentRepository(nil)
	assert.NotNil(t, repo)
}

// -----------------------------------------------------------------------------
// WebhookDelivery Model Tests
// -----------------------------------------------------------------------------

func TestWebhookDeliveryStatus_Constants_Extended(t *testing.T) {
	assert.Equal(t, WebhookDeliveryStatus("pending"), WebhookStatusPending)
	assert.Equal(t, WebhookDeliveryStatus("delivered"), WebhookStatusDelivered)
	assert.Equal(t, WebhookDeliveryStatus("failed"), WebhookStatusFailed)
	assert.Equal(t, WebhookDeliveryStatus("retrying"), WebhookStatusRetrying)
}

func TestWebhookDelivery_JSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	taskID := "task-123"
	lastErr := "timeout error"
	respCode := 500

	delivery := &WebhookDelivery{
		ID:            "delivery-123",
		TaskID:        &taskID,
		WebhookURL:    "https://example.com/webhook",
		EventType:     "task.completed",
		Payload:       json.RawMessage(`{"task_id":"task-123","status":"completed"}`),
		Status:        WebhookStatusDelivered,
		Attempts:      3,
		LastAttemptAt: &now,
		LastError:     &lastErr,
		ResponseCode:  &respCode,
		CreatedAt:     now,
		DeliveredAt:   &now,
	}

	data, err := json.Marshal(delivery)
	require.NoError(t, err)

	var decoded WebhookDelivery
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, delivery.ID, decoded.ID)
	assert.Equal(t, delivery.WebhookURL, decoded.WebhookURL)
	assert.Equal(t, delivery.EventType, decoded.EventType)
	assert.Equal(t, delivery.Status, decoded.Status)
	assert.Equal(t, delivery.Attempts, decoded.Attempts)
}

func TestWebhookDelivery_DefaultValues(t *testing.T) {
	delivery := &WebhookDelivery{}

	assert.Empty(t, delivery.ID)
	assert.Nil(t, delivery.TaskID)
	assert.Empty(t, delivery.WebhookURL)
	assert.Empty(t, delivery.EventType)
	assert.Nil(t, delivery.Payload)
	assert.Empty(t, delivery.Status)
	assert.Equal(t, 0, delivery.Attempts)
	assert.Nil(t, delivery.LastAttemptAt)
	assert.Nil(t, delivery.LastError)
	assert.Nil(t, delivery.ResponseCode)
	assert.True(t, delivery.CreatedAt.IsZero())
	assert.Nil(t, delivery.DeliveredAt)
}

func TestWebhookDelivery_JSONTags(t *testing.T) {
	taskID := "task-test"
	delivery := &WebhookDelivery{
		ID:         "id-test",
		TaskID:     &taskID,
		WebhookURL: "url-test",
		EventType:  "event-test",
		Status:     WebhookStatusPending,
		Attempts:   1,
	}

	data, err := json.Marshal(delivery)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"task_id"`)
	assert.Contains(t, jsonStr, `"webhook_url"`)
	assert.Contains(t, jsonStr, `"event_type"`)
	assert.Contains(t, jsonStr, `"status"`)
	assert.Contains(t, jsonStr, `"attempts"`)
}

func TestWebhookDelivery_StatusTransitions(t *testing.T) {
	testCases := []struct {
		name   string
		status WebhookDeliveryStatus
	}{
		{"Pending", WebhookStatusPending},
		{"Delivered", WebhookStatusDelivered},
		{"Failed", WebhookStatusFailed},
		{"Retrying", WebhookStatusRetrying},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			delivery := &WebhookDelivery{Status: tc.status}
			assert.Equal(t, tc.status, delivery.Status)
		})
	}
}

func TestWebhookDeliveryFilter_Fields(t *testing.T) {
	filter := WebhookDeliveryFilter{
		TaskID:    "task-123",
		Status:    WebhookStatusPending,
		EventType: "task.completed",
		Limit:     20,
		Offset:    10,
	}

	assert.Equal(t, "task-123", filter.TaskID)
	assert.Equal(t, WebhookStatusPending, filter.Status)
	assert.Equal(t, "task.completed", filter.EventType)
	assert.Equal(t, 20, filter.Limit)
	assert.Equal(t, 10, filter.Offset)
}

func TestNewWebhookDeliveryRepository_Extended(t *testing.T) {
	repo := NewWebhookDeliveryRepository(nil)
	assert.NotNil(t, repo)
}

// -----------------------------------------------------------------------------
// CogneeMemory Model Tests
// -----------------------------------------------------------------------------

func TestCogneeMemory_JSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	vectorID := "vec-123"
	searchKey := "test-search-key"

	memory := &CogneeMemory{
		ID:          "mem-123",
		SessionID:   "session-456",
		DatasetName: "test-dataset",
		ContentType: "text/plain",
		Content:     "This is test content for Cognee memory",
		VectorID:    &vectorID,
		GraphNodes: map[string]interface{}{
			"entity1": "value1",
			"entity2": 123,
		},
		SearchKey: &searchKey,
		CreatedAt: now,
	}

	data, err := json.Marshal(memory)
	require.NoError(t, err)

	var decoded CogneeMemory
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, memory.ID, decoded.ID)
	assert.Equal(t, memory.SessionID, decoded.SessionID)
	assert.Equal(t, memory.DatasetName, decoded.DatasetName)
	assert.Equal(t, memory.ContentType, decoded.ContentType)
	assert.Equal(t, memory.Content, decoded.Content)
}

func TestCogneeMemory_DefaultValues(t *testing.T) {
	memory := &CogneeMemory{}

	assert.Empty(t, memory.ID)
	assert.Empty(t, memory.SessionID)
	assert.Empty(t, memory.DatasetName)
	assert.Empty(t, memory.ContentType)
	assert.Empty(t, memory.Content)
	assert.Nil(t, memory.VectorID)
	assert.Nil(t, memory.GraphNodes)
	assert.Nil(t, memory.SearchKey)
	assert.True(t, memory.CreatedAt.IsZero())
}

func TestCogneeMemory_JSONTags(t *testing.T) {
	vectorID := "vec-test"
	searchKey := "key-test"
	memory := &CogneeMemory{
		ID:          "id-test",
		SessionID:   "session-test",
		DatasetName: "dataset-test",
		ContentType: "text/plain",
		Content:     "content-test",
		VectorID:    &vectorID,
		SearchKey:   &searchKey,
	}

	data, err := json.Marshal(memory)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"id"`)
	assert.Contains(t, jsonStr, `"session_id"`)
	assert.Contains(t, jsonStr, `"dataset_name"`)
	assert.Contains(t, jsonStr, `"content_type"`)
	assert.Contains(t, jsonStr, `"content"`)
	assert.Contains(t, jsonStr, `"vector_id"`)
	assert.Contains(t, jsonStr, `"search_key"`)
}

func TestCogneeMemory_ContentTypes(t *testing.T) {
	contentTypes := []string{"text/plain", "application/json", "text/markdown", "application/xml"}
	for _, contentType := range contentTypes {
		memory := &CogneeMemory{ContentType: contentType}
		assert.Equal(t, contentType, memory.ContentType)
	}
}

func TestCogneeMemory_GraphNodesComplexStructure(t *testing.T) {
	memory := &CogneeMemory{
		GraphNodes: map[string]interface{}{
			"entities": []string{"entity1", "entity2"},
			"relations": map[string]interface{}{
				"type": "association",
				"weight": 0.85,
			},
			"metadata": map[string]interface{}{
				"source": "test",
				"timestamp": time.Now().Unix(),
			},
		},
	}

	assert.NotNil(t, memory.GraphNodes)
	assert.NotNil(t, memory.GraphNodes["entities"])
	assert.NotNil(t, memory.GraphNodes["relations"])
	assert.NotNil(t, memory.GraphNodes["metadata"])
}

func TestNewCogneeMemoryRepository_Extended(t *testing.T) {
	t.Run("WithNilPool", func(t *testing.T) {
		logger := logrus.New()
		repo := NewCogneeMemoryRepository(nil, logger)
		assert.NotNil(t, repo)
	})

	t.Run("WithNilLogger", func(t *testing.T) {
		repo := NewCogneeMemoryRepository(nil, nil)
		assert.NotNil(t, repo)
	})
}

// -----------------------------------------------------------------------------
// Protocol Repository Model Tests
// -----------------------------------------------------------------------------

func TestNewProtocolRepository_Extended(t *testing.T) {
	repo := NewProtocolRepository(nil)
	assert.NotNil(t, repo)
}

func TestNewModelMetadataRepository_Extended(t *testing.T) {
	logger := logrus.New()
	repo := NewModelMetadataRepository(nil, logger)
	assert.NotNil(t, repo)
}

func TestNewModelMetadataRepository_WithNilLogger(t *testing.T) {
	repo := NewModelMetadataRepository(nil, nil)
	assert.NotNil(t, repo)
}

// -----------------------------------------------------------------------------
// Additional JSON Round-Trip Tests
// -----------------------------------------------------------------------------

func TestVectorDocument_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	embID := "emb-rt"

	original := &VectorDocument{
		ID:                "doc-rt",
		Title:             "Round Trip Document",
		Content:           "Content for round trip test",
		Metadata:          json.RawMessage(`{"key":"value"}`),
		EmbeddingID:       &embID,
		EmbeddingProvider: "openai",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded VectorDocument
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Title, decoded.Title)
	assert.Equal(t, original.Content, decoded.Content)
	assert.Equal(t, original.EmbeddingProvider, decoded.EmbeddingProvider)
}

func TestWebhookDelivery_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	taskID := "task-rt"

	original := &WebhookDelivery{
		ID:         "delivery-rt",
		TaskID:     &taskID,
		WebhookURL: "https://example.com/hook",
		EventType:  "task.done",
		Payload:    json.RawMessage(`{"status":"done"}`),
		Status:     WebhookStatusDelivered,
		Attempts:   2,
		CreatedAt:  now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded WebhookDelivery
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.WebhookURL, decoded.WebhookURL)
	assert.Equal(t, original.EventType, decoded.EventType)
	assert.Equal(t, original.Status, decoded.Status)
}

func TestCogneeMemory_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	vecID := "vec-rt"
	searchKey := "key-rt"

	original := &CogneeMemory{
		ID:          "mem-rt",
		SessionID:   "sess-rt",
		DatasetName: "dataset-rt",
		ContentType: "text/plain",
		Content:     "Round trip content",
		VectorID:    &vecID,
		GraphNodes:  map[string]interface{}{"key": "value"},
		SearchKey:   &searchKey,
		CreatedAt:   now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CogneeMemory
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.SessionID, decoded.SessionID)
	assert.Equal(t, original.DatasetName, decoded.DatasetName)
	assert.Equal(t, original.ContentType, decoded.ContentType)
}

// -----------------------------------------------------------------------------
// Edge Cases for New Models
// -----------------------------------------------------------------------------

func TestVectorDocument_LargeMetadata(t *testing.T) {
	largeMetadata := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largeMetadata[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	metadataBytes, _ := json.Marshal(largeMetadata)

	doc := &VectorDocument{
		ID:       "doc-large",
		Metadata: json.RawMessage(metadataBytes),
	}

	assert.NotNil(t, doc.Metadata)
	assert.Greater(t, len(doc.Metadata), 1000)
}

func TestWebhookDelivery_MaxAttempts(t *testing.T) {
	delivery := &WebhookDelivery{
		Attempts: 999,
	}
	assert.Equal(t, 999, delivery.Attempts)
}

func TestWebhookDelivery_NilOptionalFields_Extended(t *testing.T) {
	delivery := &WebhookDelivery{
		ID:            "delivery-nil",
		TaskID:        nil,
		LastAttemptAt: nil,
		LastError:     nil,
		ResponseCode:  nil,
		DeliveredAt:   nil,
	}

	assert.Nil(t, delivery.TaskID)
	assert.Nil(t, delivery.LastAttemptAt)
	assert.Nil(t, delivery.LastError)
	assert.Nil(t, delivery.ResponseCode)
	assert.Nil(t, delivery.DeliveredAt)
}

func TestCogneeMemory_LargeContent(t *testing.T) {
	largeContent := ""
	for i := 0; i < 10000; i++ {
		largeContent += "a"
	}

	memory := &CogneeMemory{
		Content: largeContent,
	}

	assert.Len(t, memory.Content, 10000)
}

func TestCogneeMemory_EmptyGraphNodes(t *testing.T) {
	memory := &CogneeMemory{
		GraphNodes: map[string]interface{}{},
	}

	assert.NotNil(t, memory.GraphNodes)
	assert.Len(t, memory.GraphNodes, 0)
}
