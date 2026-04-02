package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev.helix.agent/internal/services"
	"dev.helix.agent/internal/services/debate_integration"
	"dev.helix.agent/internal/skills"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDebateTestRouter(handler *DebateHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/v1/debates", handler.CreateDebate)
	r.GET("/v1/debates/:id", handler.GetDebate)
	r.GET("/v1/debates/:id/status", handler.GetDebateStatus)
	r.GET("/v1/debates/:id/results", handler.GetDebateResults)

	return r
}

func TestNewDebateHandler(t *testing.T) {
	t.Run("creates handler with services", func(t *testing.T) {
		logger := logrus.New()
		debateService := &services.DebateService{}
		advancedDebate := &services.AdvancedDebateService{}

		handler := NewDebateHandler(debateService, advancedDebate, logger)

		assert.NotNil(t, handler)
		assert.Equal(t, debateService, handler.debateService)
		assert.Equal(t, advancedDebate, handler.advancedDebate)
		assert.Equal(t, logger, handler.logger)
		assert.NotNil(t, handler.activeDebates)
	})
}

func TestNewDebateHandlerWithSkills(t *testing.T) {
	t.Run("creates handler with skills integration", func(t *testing.T) {
		logger := logrus.New()
		debateService := &services.DebateService{}
		advancedDebate := &services.AdvancedDebateService{}
		skillsIntegration := &skills.Integration{}

		handler := NewDebateHandlerWithSkills(debateService, advancedDebate, skillsIntegration, logger)

		assert.NotNil(t, handler)
		assert.Equal(t, skillsIntegration, handler.skillsIntegration)
	})
}

func TestDebateHandler_SetSkillsIntegration(t *testing.T) {
	t.Run("sets skills integration", func(t *testing.T) {
		logger := logrus.New()
		handler := NewDebateHandler(nil, nil, logger)
		skillsIntegration := &skills.Integration{}

		handler.SetSkillsIntegration(skillsIntegration)

		assert.Equal(t, skillsIntegration, handler.skillsIntegration)
	})
}

func TestDebateHandler_SetOrchestratorIntegration(t *testing.T) {
	t.Run("sets orchestrator integration", func(t *testing.T) {
		logger := logrus.New()
		handler := NewDebateHandler(nil, nil, logger)
		orchestratorIntegration := &debate_integration.ServiceIntegration{}

		handler.SetOrchestratorIntegration(orchestratorIntegration)

		assert.Equal(t, orchestratorIntegration, handler.orchestratorIntegration)
	})
}

func TestDebateHandler_CreateDebate(t *testing.T) {
	t.Run("returns error for invalid request body", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj, ok := response["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Invalid request body", errorObj["message"])
	})

	t.Run("returns error for missing required fields", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		body := map[string]interface{}{
			"topic": "Test topic",
			// Missing participants
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("creates debate with valid request", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		body := CreateDebateRequest{
			Topic: "Test debate topic",
			Participants: []ParticipantConfigRequest{
				{Name: "Participant 1"},
				{Name: "Participant 2"},
			},
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["debate_id"])
		assert.Equal(t, "pending", response["status"])
		assert.Equal(t, "Test debate topic", response["topic"])
		assert.Equal(t, float64(2), response["participants"])
	})

	t.Run("uses provided debate ID", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		body := CreateDebateRequest{
			DebateID: "custom-debate-id",
			Topic:    "Test debate topic",
			Participants: []ParticipantConfigRequest{
				{Name: "Participant 1"},
				{Name: "Participant 2"},
			},
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "custom-debate-id", response["debate_id"])
	})

	t.Run("applies default values", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		body := CreateDebateRequest{
			Topic: "Test debate topic",
			Participants: []ParticipantConfigRequest{
				{Name: "Participant 1"},
				{Name: "Participant 2"},
			},
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check defaults applied
		assert.Equal(t, float64(3), response["max_rounds"])
		assert.Equal(t, float64(300), response["timeout"]) // 5 minutes
	})

	t.Run("processes skills integration when available", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		// Set up mock skills integration that returns error
		skillsIntegration := &skills.Integration{}
		handler.SetSkillsIntegration(skillsIntegration)

		router := setupDebateTestRouter(handler)

		body := CreateDebateRequest{
			Topic: "Test debate topic",
			Participants: []ParticipantConfigRequest{
				{Name: "Participant 1"},
				{Name: "Participant 2"},
			},
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		// Should still succeed even if skills integration fails
		assert.Equal(t, http.StatusAccepted, w.Code)
	})
}

func TestDebateHandler_GetDebate(t *testing.T) {
	t.Run("returns debate not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/non-existent-id", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj, ok := response["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Debate not found", errorObj["message"])
	})

	t.Run("returns existing debate", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		// Create a debate first
		handler.mu.Lock()
		handler.activeDebates["test-debate-123"] = &debateState{
			Config: &services.DebateConfig{
				DebateID:  "test-debate-123",
				Topic:     "Test Topic",
				MaxRounds: 5,
			},
			Status:    "running",
			StartTime: time.Now(),
		}
		handler.mu.Unlock()

		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/test-debate-123", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "test-debate-123", response["debate_id"])
		assert.Equal(t, "Test Topic", response["topic"])
		assert.Equal(t, "running", response["status"])
		assert.Equal(t, float64(5), response["max_rounds"])
	})

	t.Run("returns completed debate with results", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		endTime := time.Now()
		handler.mu.Lock()
		handler.activeDebates["completed-debate"] = &debateState{
			Config: &services.DebateConfig{
				DebateID:  "completed-debate",
				Topic:     "Completed Topic",
				MaxRounds: 3,
			},
			Status:    "completed",
			StartTime: time.Now().Add(-5 * time.Minute),
			EndTime:   &endTime,
			Result: &services.DebateResult{
				DebateID: "completed-debate",
				Topic:    "Completed Topic",
				Success:  true,
			},
		}
		handler.mu.Unlock()

		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/completed-debate", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotNil(t, response["end_time"])
		assert.NotNil(t, response["duration_seconds"])
		assert.NotNil(t, response["result"])
	})

	t.Run("returns failed debate with error", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		endTime := time.Now()
		handler.mu.Lock()
		handler.activeDebates["failed-debate"] = &debateState{
			Config: &services.DebateConfig{
				DebateID: "failed-debate",
				Topic:    "Failed Topic",
			},
			Status:    "failed",
			StartTime: time.Now().Add(-2 * time.Minute),
			EndTime:   &endTime,
			Error:     "Debate execution failed",
		}
		handler.mu.Unlock()

		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/failed-debate", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "failed", response["status"])
		assert.Equal(t, "Debate execution failed", response["error"])
	})
}

func TestDebateHandler_GetDebateStatus(t *testing.T) {
	t.Run("returns not found for unknown debate", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/unknown/status", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns status for running debate", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		handler.mu.Lock()
		handler.activeDebates["running-debate"] = &debateState{
			Config: &services.DebateConfig{
				DebateID: "running-debate",
				Topic:    "Running Topic",
				MaxRounds: 5,
				Timeout:  10 * time.Minute,
			},
			Status:       "running",
			StartTime:    time.Now(),
			CurrentPhase: string(services.PhaseValidation),
		}
		handler.mu.Unlock()

		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/running-debate/status", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "running-debate", response["debate_id"])
		assert.Equal(t, "running", response["status"])
		assert.Equal(t, string(services.PhaseValidation), response["current_phase"])
		assert.NotNil(t, response["max_rounds"])
		assert.NotNil(t, response["timeout_seconds"])
	})

	t.Run("returns completed status with results", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		endTime := time.Now()
		handler.mu.Lock()
		handler.activeDebates["completed-status-debate"] = &debateState{
			Config: &services.DebateConfig{
				DebateID: "completed-status-debate",
				Topic:    "Status Topic",
			},
			Status:    "completed",
			StartTime: time.Now().Add(-5 * time.Minute),
			EndTime:   &endTime,
			MultiPassResult: &services.MultiPassResult{
				OverallConfidence:  0.85,
				QualityImprovement: 0.15,
				Phases:             []services.ValidationPhase{{}, {}, {}},
			},
		}
		handler.mu.Unlock()

		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/completed-status-debate/status", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(0.85), response["overall_confidence"])
		assert.Equal(t, float64(0.15), response["quality_improvement"])
		assert.Equal(t, float64(3), response["phases_completed"])
	})
}

func TestDebateHandler_GetDebateResults(t *testing.T) {
	t.Run("returns not found for unknown debate", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/unknown/results", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns results for completed debate", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		handler.mu.Lock()
		handler.activeDebates["results-debate"] = &debateState{
			Config: &services.DebateConfig{
				DebateID: "results-debate",
				Topic:    "Results Topic",
			},
			Status:    "completed",
			StartTime: time.Now(),
			Result: &services.DebateResult{
				DebateID:     "results-debate",
				Topic:        "Results Topic",
				Success:      true,
				QualityScore: 0.92,
				Consensus: &services.ConsensusResult{
					Reached:    true,
					Confidence: 0.88,
					Summary:    "Consensus reached on topic",
				},
			},
		}
		handler.mu.Unlock()

		router := setupDebateTestRouter(handler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/debates/results-debate/results", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "completed", response["status"])
		assert.NotNil(t, response["result"])
	})
}

func TestDebateHandler_runDebate(t *testing.T) {
	t.Run("handles missing orchestrator", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)

		// Create debate state
		handler.mu.Lock()
		handler.activeDebates["test-debate"] = &debateState{
			Config: &services.DebateConfig{
				DebateID: "test-debate",
				Topic:    "Test Topic",
			},
			Status:    "pending",
			StartTime: time.Now(),
		}
		handler.mu.Unlock()

		config := &services.DebateConfig{
			DebateID: "test-debate",
			Topic:    "Test Topic",
		}

		// Run debate synchronously for test
		handler.runDebate("test-debate", config, nil)

		// Check that debate failed
		handler.mu.RLock()
		state := handler.activeDebates["test-debate"]
		handler.mu.RUnlock()

		assert.Equal(t, "failed", state.Status)
		assert.Contains(t, state.Error, "orchestrator integration not initialized")
	})
}

func TestDebateError(t *testing.T) {
	t.Run("returns error message", func(t *testing.T) {
		err := &debateError{message: "test error message"}
		assert.Equal(t, "test error message", err.Error())
	})
}

func TestDebateHandler_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent debate creation", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				body := CreateDebateRequest{
					DebateID: "concurrent-debate-" + string(rune('a'+idx)),
					Topic:    "Concurrent Topic",
					Participants: []ParticipantConfigRequest{
						{Name: "Participant 1"},
						{Name: "Participant 2"},
					},
				}
				jsonBody, _ := json.Marshal(body)

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusAccepted, w.Code)
			}(i)
		}

		wg.Wait()

		// Verify all debates were created
		handler.mu.RLock()
		count := len(handler.activeDebates)
		handler.mu.RUnlock()

		assert.Equal(t, 5, count)
	})

	t.Run("handles concurrent read and write", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		handler := NewDebateHandler(nil, nil, logger)
		router := setupDebateTestRouter(handler)

		// Create initial debate
		body := CreateDebateRequest{
			DebateID: "concurrent-access-debate",
			Topic:    "Concurrent Access",
			Participants: []ParticipantConfigRequest{
				{Name: "P1"},
				{Name: "P2"},
			},
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/debates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		var wg sync.WaitGroup

		// Concurrent reads
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/debates/concurrent-access-debate", nil)
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
			}()
		}

		// Concurrent status checks
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/v1/debates/concurrent-access-debate/status", nil)
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
			}()
		}

		wg.Wait()
	})
}
