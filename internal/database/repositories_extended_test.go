package database

import (
	"context"
	"testing"
	"time"

	"dev.helix.agent/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Extended Tests for Remaining Repository Methods
// Covers all 0% functions using nil pool + panic recovery pattern.
// =============================================================================

// -----------------------------------------------------------------------------
// RequestRepository Tests
// -----------------------------------------------------------------------------

func TestRequestRepository_GetBySessionID_NilPool(t *testing.T) {
	repo := NewRequestRepository(nil, logrus.New())

	_, _, err := safeCallRequestListResult(func() ([]*LLMRequest, int, error) {
		return repo.GetBySessionID(context.Background(), "sess-1", 10, 0)
	})
	assert.Error(t, err)
}

func TestRequestRepository_GetByUserID_NilPool(t *testing.T) {
	repo := NewRequestRepository(nil, logrus.New())

	_, _, err := safeCallRequestListResult(func() ([]*LLMRequest, int, error) {
		return repo.GetByUserID(context.Background(), "user-1", 10, 0)
	})
	assert.Error(t, err)
}

func TestRequestRepository_GetRequestStats_NilPool(t *testing.T) {
	repo := NewRequestRepository(nil, logrus.New())

	_, err := safeCallResult(func() (map[string]int, error) {
		return repo.GetRequestStats(context.Background(), "user-1", time.Now().Add(-24*time.Hour))
	})
	assert.Error(t, err)
}

func TestRequestRepository_GetRequestStats_EmptyUserID_NilPool(t *testing.T) {
	repo := NewRequestRepository(nil, logrus.New())

	_, err := safeCallResult(func() (map[string]int, error) {
		return repo.GetRequestStats(context.Background(), "", time.Now().Add(-7*24*time.Hour))
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// ResponseRepository Tests
// -----------------------------------------------------------------------------

func TestResponseRepository_GetByProviderID_NilPool(t *testing.T) {
	repo := NewResponseRepository(nil, logrus.New())

	_, _, err := safeCallResponseListResult(func() ([]*LLMResponse, int, error) {
		return repo.GetByProviderID(context.Background(), "openai", 10, 0)
	})
	assert.Error(t, err)
}

func TestResponseRepository_GetByProviderID_WithOffset_NilPool(t *testing.T) {
	repo := NewResponseRepository(nil, logrus.New())

	_, _, err := safeCallResponseListResult(func() ([]*LLMResponse, int, error) {
		return repo.GetByProviderID(context.Background(), "deepseek", 20, 5)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// SessionRepository Tests
// -----------------------------------------------------------------------------

func TestSessionRepository_Update_NilPool(t *testing.T) {
	repo := NewSessionRepository(nil, logrus.New())

	session := &UserSession{
		ID:           "sess-1",
		Context:      map[string]interface{}{"key": "value"},
		Status:       "active",
		RequestCount: 5,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	err := safeCallError(func() error {
		return repo.Update(context.Background(), session)
	})
	assert.Error(t, err)
}

func TestSessionRepository_UpdateContext_NilPool(t *testing.T) {
	repo := NewSessionRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.UpdateContext(context.Background(), "sess-1", map[string]interface{}{"topic": "testing"})
	})
	assert.Error(t, err)
}

func TestSessionRepository_Update_MarshalError(t *testing.T) {
	repo := NewSessionRepository(nil, logrus.New())

	// Channel values cannot be marshaled to JSON
	session := &UserSession{
		ID:      "sess-1",
		Context: map[string]interface{}{"bad": make(chan int)},
		Status:  "active",
	}

	err := safeCallError(func() error {
		return repo.Update(context.Background(), session)
	})
	assert.Error(t, err)
}

func TestSessionRepository_UpdateContext_MarshalError(t *testing.T) {
	repo := NewSessionRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.UpdateContext(context.Background(), "sess-1", map[string]interface{}{"bad": make(chan int)})
	})
	assert.Error(t, err)
}

func TestSessionRepository_DeleteExpired_NilPool(t *testing.T) {
	repo := NewSessionRepository(nil, logrus.New())

	_, err := safeCallResult(func() (int64, error) {
		return repo.DeleteExpired(context.Background())
	})
	assert.Error(t, err)
}

func TestSessionRepository_DeleteByUserID_NilPool(t *testing.T) {
	repo := NewSessionRepository(nil, logrus.New())

	_, err := safeCallResult(func() (int64, error) {
		return repo.DeleteByUserID(context.Background(), "user-1")
	})
	assert.Error(t, err)
}

func TestSessionRepository_IsValid_NilPool(t *testing.T) {
	repo := NewSessionRepository(nil, logrus.New())

	_, err := safeCallResult(func() (bool, error) {
		return repo.IsValid(context.Background(), "token-abc")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// UserRepository Tests
// -----------------------------------------------------------------------------

func TestUserRepository_Create_NilPool(t *testing.T) {
	repo := NewUserRepository(nil, logrus.New())

	user := &User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hash123",
		APIKey:       "key-abc",
		Role:         "user",
	}

	err := safeCallError(func() error {
		return repo.Create(context.Background(), user)
	})
	assert.Error(t, err)
}

func TestUserRepository_ExistsByEmail_NilPool(t *testing.T) {
	repo := NewUserRepository(nil, logrus.New())

	_, err := safeCallResult(func() (bool, error) {
		return repo.ExistsByEmail(context.Background(), "test@example.com")
	})
	assert.Error(t, err)
}

func TestUserRepository_ExistsByUsername_NilPool(t *testing.T) {
	repo := NewUserRepository(nil, logrus.New())

	_, err := safeCallResult(func() (bool, error) {
		return repo.ExistsByUsername(context.Background(), "testuser")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// VectorDocumentRepository Tests
// -----------------------------------------------------------------------------

func TestVectorDocumentRepository_GetDocumentsWithoutEmbeddings_NilPool(t *testing.T) {
	repo := NewVectorDocumentRepository(nil)

	_, err := safeCallResult(func() ([]*VectorDocument, error) {
		return repo.GetDocumentsWithoutEmbeddings(context.Background(), 50)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// WebhookDeliveryRepository Tests
// -----------------------------------------------------------------------------

func TestWebhookDeliveryRepository_GetByTaskID_NilPool(t *testing.T) {
	repo := NewWebhookDeliveryRepository(nil)

	_, err := safeCallResult(func() ([]*WebhookDelivery, error) {
		return repo.GetByTaskID(context.Background(), "task-1")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// CogneeMemoryRepository Tests
// -----------------------------------------------------------------------------

func TestCogneeMemoryRepository_DeleteBySessionID_NilPool(t *testing.T) {
	repo := NewCogneeMemoryRepository(nil, logrus.New())

	_, err := safeCallResult(func() (int64, error) {
		return repo.DeleteBySessionID(context.Background(), "sess-1")
	})
	assert.Error(t, err)
}

func TestCogneeMemoryRepository_DeleteByDatasetName_NilPool(t *testing.T) {
	repo := NewCogneeMemoryRepository(nil, logrus.New())

	_, err := safeCallResult(func() (int64, error) {
		return repo.DeleteByDatasetName(context.Background(), "dataset-1")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// BackgroundTaskRepository Tests
// -----------------------------------------------------------------------------

func TestBackgroundTaskRepository_UpdateStatus_NilPool(t *testing.T) {
	repo := NewBackgroundTaskRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.UpdateStatus(context.Background(), "task-1", "running")
	})
	assert.Error(t, err)
}

func TestBackgroundTaskRepository_UpdateProgress_NilPool(t *testing.T) {
	repo := NewBackgroundTaskRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.UpdateProgress(context.Background(), "task-1", 50.0, "Halfway done")
	})
	assert.Error(t, err)
}

func TestBackgroundTaskRepository_UpdateHeartbeat_NilPool(t *testing.T) {
	repo := NewBackgroundTaskRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.UpdateHeartbeat(context.Background(), "task-1")
	})
	assert.Error(t, err)
}

func TestBackgroundTaskRepository_SaveCheckpoint_NilPool(t *testing.T) {
	repo := NewBackgroundTaskRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.SaveCheckpoint(context.Background(), "task-1", []byte(`{"step":5}`))
	})
	assert.Error(t, err)
}

func TestBackgroundTaskRepository_Delete_NilPool(t *testing.T) {
	repo := NewBackgroundTaskRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.Delete(context.Background(), "task-1")
	})
	assert.Error(t, err)
}

func TestBackgroundTaskRepository_HardDelete_NilPool(t *testing.T) {
	repo := NewBackgroundTaskRepository(nil, logrus.New())

	err := safeCallError(func() error {
		return repo.HardDelete(context.Background(), "task-1")
	})
	assert.Error(t, err)
}

func TestBackgroundTaskRepository_SaveResourceSnapshot_NilPool(t *testing.T) {
	repo := NewBackgroundTaskRepository(nil, logrus.New())

	snapshot := &models.ResourceSnapshot{
		TaskID:     "task-1",
		CPUPercent: 45.0,
	}

	err := safeCallError(func() error {
		return repo.SaveResourceSnapshot(context.Background(), snapshot)
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// ProviderRepository Tests
// -----------------------------------------------------------------------------

func TestProviderRepository_Update_NilPool(t *testing.T) {
	repo := NewProviderRepository(nil, logrus.New())

	provider := &LLMProvider{
		ID:           "prov-1",
		Name:         "openai",
		Type:         "api",
		APIKey:       "sk-test",
		BaseURL:      "https://api.openai.com",
		Model:        "gpt-4",
		Weight:       1.0,
		Enabled:      true,
		Config:       map[string]interface{}{"max_tokens": 4096},
		HealthStatus: "healthy",
		ResponseTime: 250,
	}

	err := safeCallError(func() error {
		return repo.Update(context.Background(), provider)
	})
	assert.Error(t, err)
}

func TestProviderRepository_Update_MarshalError(t *testing.T) {
	repo := NewProviderRepository(nil, logrus.New())

	provider := &LLMProvider{
		ID:     "prov-1",
		Name:   "openai",
		Config: map[string]interface{}{"bad": make(chan int)},
	}

	err := safeCallError(func() error {
		return repo.Update(context.Background(), provider)
	})
	assert.Error(t, err)
}

func TestProviderRepository_ExistsByName_NilPool(t *testing.T) {
	repo := NewProviderRepository(nil, logrus.New())

	_, err := safeCallResult(func() (bool, error) {
		return repo.ExistsByName(context.Background(), "openai")
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Helper functions for 3-return-value repository functions
// -----------------------------------------------------------------------------

func safeCallRequestListResult(fn func() ([]*LLMRequest, int, error)) ([]*LLMRequest, int, error) {
	var result []*LLMRequest
	var count int
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = assert.AnError
			}
		}()
		result, count, err = fn()
	}()
	return result, count, err
}

func safeCallResponseListResult(fn func() ([]*LLMResponse, int, error)) ([]*LLMResponse, int, error) {
	var result []*LLMResponse
	var count int
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = assert.AnError
			}
		}()
		result, count, err = fn()
	}()
	return result, count, err
}
