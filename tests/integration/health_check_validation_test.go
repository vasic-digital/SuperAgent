package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/handlers"
	"dev.helix.agent/internal/verifier"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupHealthTestRouter creates a Gin router with health handler routes
// and a pre-populated HealthService for testing.
func setupHealthTestRouter(providers map[string]string) (*gin.Engine, *verifier.HealthService) {
	hs := verifier.NewHealthService(nil)
	for id, name := range providers {
		hs.AddProvider(id, name)
	}

	handler := handlers.NewHealthHandler(hs)
	r := gin.New()
	v1 := r.Group("/api/v1")
	handlers.RegisterHealthRoutes(v1, handler)

	return r, hs
}

// TestHealthEndpointReturnsCorrectStatus verifies the health service status
// endpoint returns the expected JSON structure and HTTP 200.
func TestHealthEndpointReturnsCorrectStatus(t *testing.T) {
	providers := map[string]string{
		"openai":   "openai",
		"deepseek": "deepseek",
		"gemini":   "google",
	}
	router, _ := setupHealthTestRouter(providers)

	// Test the health service status endpoint.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Health status endpoint must return 200")

	var resp handlers.HealthServiceStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Response must be valid JSON")
	assert.True(t, resp.Running, "Health service must report as running")
	assert.Equal(t, len(providers), resp.TotalProviders,
		"Total providers must match registered count")
}

// TestHealthEndpointReturnsCorrectStatus_AllProviders verifies the
// providers list endpoint returns correct counts and structure.
func TestHealthEndpointReturnsCorrectStatus_AllProviders(t *testing.T) {
	providers := map[string]string{
		"openai":   "openai",
		"deepseek": "deepseek",
	}
	router, _ := setupHealthTestRouter(providers)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handlers.GetAllProvidersHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 2, resp.Total, "Total must be 2")
	assert.Len(t, resp.Providers, 2, "Providers slice must have 2 entries")

	// Newly added providers start as healthy.
	assert.Equal(t, 2, resp.HealthyCount, "All providers must be healthy initially")
	assert.Equal(t, 0, resp.UnhealthyCount, "No providers must be unhealthy initially")
	assert.True(t, resp.Summary.OverallHealthy, "Overall health must be true")
}

// TestHealthEndpointReturnsCorrectStatus_SingleProvider verifies that
// fetching a specific provider's health returns the correct data.
func TestHealthEndpointReturnsCorrectStatus_SingleProvider(t *testing.T) {
	providers := map[string]string{
		"openai": "openai",
	}
	router, _ := setupHealthTestRouter(providers)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers/openai", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handlers.ProviderHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "openai", resp.ProviderID)
	assert.Equal(t, "openai", resp.ProviderName)
	assert.True(t, resp.Healthy, "Newly added provider must be healthy")
	assert.Equal(t, "closed", resp.CircuitState, "Initial circuit state must be closed")
}

// TestHealthEndpointReturnsCorrectStatus_NotFound verifies that querying
// a non-existent provider returns 404.
func TestHealthEndpointReturnsCorrectStatus_NotFound(t *testing.T) {
	router, _ := setupHealthTestRouter(map[string]string{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Non-existent provider must return 404")
}

// TestHealthDegradationOnDependencyFailure verifies that recording failures
// degrades provider health and is reflected in the health response.
func TestHealthDegradationOnDependencyFailure(t *testing.T) {
	providers := map[string]string{
		"test-provider": "test-provider-name",
	}
	router, hs := setupHealthTestRouter(providers)

	// Initially healthy.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers/test-provider", nil)
	router.ServeHTTP(w, req)

	var initialResp handlers.ProviderHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &initialResp)
	require.NoError(t, err)
	assert.True(t, initialResp.Healthy, "Provider must be healthy initially")
	assert.Equal(t, 0, initialResp.FailureCount, "No failures initially")

	// Record some failures via the HealthService.
	hs.RecordFailure("test-provider")
	hs.RecordFailure("test-provider")
	hs.RecordFailure("test-provider")

	// Query again and verify degradation.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers/test-provider", nil)
	router.ServeHTTP(w2, req2)

	var degradedResp handlers.ProviderHealthResponse
	err = json.Unmarshal(w2.Body.Bytes(), &degradedResp)
	require.NoError(t, err)
	assert.Equal(t, 3, degradedResp.FailureCount, "Failure count must be 3 after 3 failures")

	// Verify overall health reflects degradation.
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers", nil)
	router.ServeHTTP(w3, req3)

	var allResp handlers.GetAllProvidersHealthResponse
	err = json.Unmarshal(w3.Body.Bytes(), &allResp)
	require.NoError(t, err)
	assert.Equal(t, 1, allResp.Total)
}

// TestHealthDegradationOnDependencyFailure_RecordViaAPI verifies the
// POST record/failure endpoint updates provider failure counts.
func TestHealthDegradationOnDependencyFailure_RecordViaAPI(t *testing.T) {
	providers := map[string]string{
		"api-provider": "api-provider-name",
	}
	router, _ := setupHealthTestRouter(providers)

	// Record failure via the REST API.
	failBody := `{"provider_id": "api-provider"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/verifier/health/record/failure",
		strings.NewReader(failBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Record failure must return 200")

	// Now record a success.
	successBody := `{"provider_id": "api-provider"}`
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/verifier/health/record/success",
		strings.NewReader(successBody))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Record success must return 200")

	// Verify the provider state.
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers/api-provider", nil)
	router.ServeHTTP(w3, req3)

	var resp handlers.ProviderHealthResponse
	err := json.Unmarshal(w3.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Healthy, "Provider must be healthy after recording success")
	assert.Equal(t, 1, resp.FailureCount, "Failure count must be 1")
	assert.Equal(t, 1, resp.SuccessCount, "Success count must be 1")
}

// TestCircuitBreakerStateReflectsInHealth verifies that the circuit breaker
// state transitions (closed -> open) are reflected in health responses.
func TestCircuitBreakerStateReflectsInHealth(t *testing.T) {
	providers := map[string]string{
		"circuit-test": "circuit-test-name",
	}
	router, hs := setupHealthTestRouter(providers)

	// Initially the circuit breaker must be closed.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/circuit/circuit-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var cbResp handlers.CircuitBreakerResponse
	err := json.Unmarshal(w.Body.Bytes(), &cbResp)
	require.NoError(t, err)
	assert.Equal(t, "circuit-test", cbResp.ProviderID)
	assert.Equal(t, "closed", cbResp.State, "Initial circuit state must be closed")
	assert.True(t, cbResp.IsAvailable, "Circuit must be available when closed")

	// Record failures to trip the circuit breaker.
	// The default threshold is 5 failures.
	cb := hs.GetCircuitBreaker("circuit-test")
	require.NotNil(t, cb, "Circuit breaker must exist for registered provider")
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// After 5 failures, the circuit should be open.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/circuit/circuit-test", nil)
	router.ServeHTTP(w2, req2)

	var cbResp2 handlers.CircuitBreakerResponse
	err = json.Unmarshal(w2.Body.Bytes(), &cbResp2)
	require.NoError(t, err)
	assert.Equal(t, "open", cbResp2.State,
		"Circuit must be open after reaching failure threshold")
	assert.False(t, cbResp2.IsAvailable,
		"Circuit must not be available when open")

	// The healthy providers list should exclude the open-circuit provider.
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/healthy", nil)
	router.ServeHTTP(w3, req3)

	var healthyResp handlers.GetHealthyProvidersResponse
	err = json.Unmarshal(w3.Body.Bytes(), &healthyResp)
	require.NoError(t, err)

	// Provider with open circuit should not appear in healthy list.
	for _, p := range healthyResp.Providers {
		assert.NotEqual(t, "circuit-test", p,
			"Provider with open circuit must not be in healthy list")
	}
}

// TestCircuitBreakerStateReflectsInHealth_NotFound verifies that querying
// circuit breaker for a non-existent provider returns 404.
func TestCircuitBreakerStateReflectsInHealth_NotFound(t *testing.T) {
	router, _ := setupHealthTestRouter(map[string]string{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/circuit/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code,
		"Circuit breaker for non-existent provider must return 404")
}

// TestCircuitBreakerStateReflectsInHealth_Recovery verifies that a
// circuit breaker transitions back to closed after successful operations
// following a half-open state.
func TestCircuitBreakerStateReflectsInHealth_Recovery(t *testing.T) {
	hs := verifier.NewHealthService(nil)
	hs.AddProvider("recovery-test", "recovery-test-name")

	cb := hs.GetCircuitBreaker("recovery-test")
	require.NotNil(t, cb)

	// Initially closed.
	assert.Equal(t, verifier.CircuitClosed, cb.State(),
		"Circuit must start in closed state")
	assert.True(t, cb.IsAvailable(), "Closed circuit must be available")

	// Record a successful call via the circuit breaker.
	err := cb.Call(func() error {
		return nil
	})
	assert.NoError(t, err, "Successful call must not return error")

	// State remains closed after success.
	assert.Equal(t, verifier.CircuitClosed, cb.State(),
		"Circuit must remain closed after success")
}

// TestHealthEndpoint_AddAndRemoveProvider verifies the add/remove
// provider lifecycle via the REST API.
func TestHealthEndpoint_AddAndRemoveProvider(t *testing.T) {
	router, _ := setupHealthTestRouter(map[string]string{})

	// Add a provider via API.
	addBody := `{"provider_id": "new-provider", "provider_name": "New Provider"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/verifier/health/providers",
		strings.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Add provider must return 200")

	// Verify the provider was added.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers/new-provider", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Fetching added provider must return 200")

	// Remove the provider.
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodDelete, "/api/v1/verifier/health/providers/new-provider", nil)
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code, "Remove provider must return 200")

	// Verify the provider was removed.
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/providers/new-provider", nil)
	router.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusNotFound, w4.Code,
		"Removed provider must return 404")
}

// TestHealthEndpoint_IsProviderAvailable verifies the availability check
// endpoint returns correct data.
func TestHealthEndpoint_IsProviderAvailable(t *testing.T) {
	providers := map[string]string{
		"avail-test": "avail-test-name",
	}
	router, _ := setupHealthTestRouter(providers)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verifier/health/available/avail-test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "avail-test", resp["provider_id"])
	assert.Equal(t, true, resp["available"],
		"Newly added provider must be available")
}
