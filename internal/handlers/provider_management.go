package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/services"
)

// ProviderManagementHandler handles provider CRUD operations
type ProviderManagementHandler struct {
	providerRegistry *services.ProviderRegistry
	log              *logrus.Logger
}

// NewProviderManagementHandler creates a new provider management handler
func NewProviderManagementHandler(providerRegistry *services.ProviderRegistry, log *logrus.Logger) *ProviderManagementHandler {
	return &ProviderManagementHandler{
		providerRegistry: providerRegistry,
		log:              log,
	}
}

// AddProviderRequest represents a request to add a new provider
type AddProviderRequest struct {
	Name    string                 `json:"name" binding:"required"`
	Type    string                 `json:"type" binding:"required"`
	APIKey  string                 `json:"api_key" binding:"required"`
	BaseURL string                 `json:"base_url" binding:"required"`
	Model   string                 `json:"model" binding:"required"`
	Weight  float64                `json:"weight"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

// UpdateProviderRequest represents a request to update a provider
type UpdateProviderRequest struct {
	Name    string                 `json:"name"`
	APIKey  string                 `json:"api_key"`
	BaseURL string                 `json:"base_url"`
	Model   string                 `json:"model"`
	Weight  float64                `json:"weight"`
	Enabled *bool                  `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

// ProviderResponse represents a provider response
type ProviderResponse struct {
	Success  bool                `json:"success"`
	Message  string              `json:"message"`
	Provider *models.LLMProvider `json:"provider,omitempty"`
}

// AddProvider handles POST /v1/providers
func (h *ProviderManagementHandler) AddProvider(c *gin.Context) {
	var req AddProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Failed to bind add provider request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate provider type - DYNAMIC: get valid types from provider registry/discovery
	// This ensures new provider types are automatically supported
	validTypes := h.getValidProviderTypes()
	isValidType := false
	for _, t := range validTypes {
		if req.Type == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid provider type",
			"valid_types": validTypes,
			"hint":        "Provider types are dynamically discovered from environment variables and registered providers",
		})
		return
	}

	// Check if provider already exists
	existingProviders := h.providerRegistry.ListProviders()
	for _, name := range existingProviders {
		if name == req.Name {
			c.JSON(http.StatusConflict, gin.H{
				"error": "provider already exists",
				"name":  req.Name,
			})
			return
		}
	}

	// Set default weight if not provided
	if req.Weight == 0 {
		req.Weight = 1.0
	}

	// Create provider configuration
	providerConfig := services.ProviderConfig{
		Name:    req.Name,
		Type:    req.Type,
		APIKey:  req.APIKey,
		BaseURL: req.BaseURL,
		Weight:  req.Weight,
		Enabled: req.Enabled,
		Timeout: 30 * time.Second,
		Models: []services.ModelConfig{{
			ID:      req.Model,
			Name:    req.Model,
			Enabled: true,
			Weight:  1.0,
		}},
	}

	// Register the provider
	if err := h.providerRegistry.RegisterProviderFromConfig(providerConfig); err != nil {
		h.log.WithError(err).Error("Failed to register provider")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	provider := &models.LLMProvider{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Type:         req.Type,
		BaseURL:      req.BaseURL,
		Model:        req.Model,
		Weight:       req.Weight,
		Enabled:      req.Enabled,
		Config:       req.Config,
		HealthStatus: "unknown",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	h.log.WithFields(logrus.Fields{
		"name": req.Name,
		"type": req.Type,
	}).Info("Provider added successfully")

	c.JSON(http.StatusCreated, ProviderResponse{
		Success:  true,
		Message:  "Provider added successfully",
		Provider: provider,
	})
}

// GetProvider handles GET /v1/providers/:id
func (h *ProviderManagementHandler) GetProvider(c *gin.Context) {
	providerID := c.Param("id")

	provider, err := h.providerRegistry.GetProvider(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "provider not found",
			"provider_id": providerID,
		})
		return
	}

	capabilities := provider.GetCapabilities()

	c.JSON(http.StatusOK, gin.H{
		"id":                        providerID,
		"name":                      providerID,
		"supported_models":          capabilities.SupportedModels,
		"supported_features":        capabilities.SupportedFeatures,
		"supports_streaming":        capabilities.SupportsStreaming,
		"supports_function_calling": capabilities.SupportsFunctionCalling,
		"supports_vision":           capabilities.SupportsVision,
		"limits":                    capabilities.Limits,
		"metadata":                  capabilities.Metadata,
	})
}

// UpdateProvider handles PUT /v1/providers/:id
func (h *ProviderManagementHandler) UpdateProvider(c *gin.Context) {
	providerID := c.Param("id")

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Error("Failed to bind update provider request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if provider exists
	_, err := h.providerRegistry.GetProvider(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "provider not found",
			"provider_id": providerID,
		})
		return
	}

	// Build update config
	updateConfig := services.ProviderConfig{
		Name:    providerID,
		APIKey:  req.APIKey,
		BaseURL: req.BaseURL,
		Weight:  req.Weight,
	}

	if req.Model != "" {
		updateConfig.Models = []services.ModelConfig{{
			ID:      req.Model,
			Name:    req.Model,
			Enabled: true,
			Weight:  1.0,
		}}
	}

	if req.Enabled != nil {
		updateConfig.Enabled = *req.Enabled
	}

	// Update the provider
	if err := h.providerRegistry.UpdateProvider(providerID, updateConfig); err != nil {
		h.log.WithError(err).Error("Failed to update provider")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.log.WithField("provider_id", providerID).Info("Provider updated successfully")

	c.JSON(http.StatusOK, ProviderResponse{
		Success: true,
		Message: "Provider updated successfully",
	})
}

// DeleteProvider handles DELETE /v1/providers/:id
func (h *ProviderManagementHandler) DeleteProvider(c *gin.Context) {
	providerID := c.Param("id")

	// Check if provider exists
	_, err := h.providerRegistry.GetProvider(providerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "provider not found",
			"provider_id": providerID,
		})
		return
	}

	// Check force parameter
	force := c.Query("force") == "true"

	// Remove the provider
	if err := h.providerRegistry.RemoveProvider(providerID, force); err != nil {
		h.log.WithError(err).Error("Failed to remove provider")
		c.JSON(http.StatusConflict, gin.H{
			"error":       err.Error(),
			"provider_id": providerID,
			"hint":        "Use ?force=true to force removal",
		})
		return
	}

	h.log.WithField("provider_id", providerID).Info("Provider removed successfully")

	c.JSON(http.StatusOK, ProviderResponse{
		Success: true,
		Message: "Provider removed successfully",
	})
}

// VerifyProvider handles POST /v1/providers/:id/verify
// This endpoint triggers verification of a specific provider with an actual API call
func (h *ProviderManagementHandler) VerifyProvider(c *gin.Context) {
	providerID := c.Param("id")

	h.log.WithField("provider_id", providerID).Info("Starting provider verification")

	// Verify the provider with actual API call
	result := h.providerRegistry.VerifyProvider(c.Request.Context(), providerID)

	// Build response based on verification result
	response := gin.H{
		"provider":         result.Provider,
		"status":           result.Status,
		"verified":         result.Verified,
		"response_time_ms": result.ResponseTime.Milliseconds(),
		"tested_at":        result.TestedAt,
	}

	if result.Error != "" {
		response["error"] = result.Error
	}

	// Determine HTTP status code based on verification result
	statusCode := http.StatusOK
	if !result.Verified {
		switch result.Status {
		case services.ProviderStatusRateLimited:
			statusCode = http.StatusTooManyRequests
		case services.ProviderStatusAuthFailed:
			statusCode = http.StatusUnauthorized
		case services.ProviderStatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusServiceUnavailable
		}
	}

	h.log.WithFields(logrus.Fields{
		"provider_id": providerID,
		"status":      result.Status,
		"verified":    result.Verified,
	}).Info("Provider verification completed")

	c.JSON(statusCode, response)
}

// VerifyAllProviders handles POST /v1/providers/verify
// This endpoint triggers verification of all registered providers
func (h *ProviderManagementHandler) VerifyAllProviders(c *gin.Context) {
	h.log.Info("Starting verification of all providers")

	// Verify all providers concurrently
	results := h.providerRegistry.VerifyAllProviders(c.Request.Context())

	// Build response with statistics
	var healthy, rateLimited, authFailed, unhealthy int
	providerResults := make([]gin.H, 0, len(results))

	for _, result := range results {
		switch result.Status {
		case services.ProviderStatusHealthy:
			healthy++
		case services.ProviderStatusRateLimited:
			rateLimited++
		case services.ProviderStatusAuthFailed:
			authFailed++
		default:
			unhealthy++
		}

		providerResult := gin.H{
			"provider":         result.Provider,
			"status":           result.Status,
			"verified":         result.Verified,
			"response_time_ms": result.ResponseTime.Milliseconds(),
			"tested_at":        result.TestedAt,
		}
		if result.Error != "" {
			providerResult["error"] = result.Error
		}
		providerResults = append(providerResults, providerResult)
	}

	// Determine overall ensemble operational status
	ensembleOperational := healthy > 0

	h.log.WithFields(logrus.Fields{
		"total":               len(results),
		"healthy":             healthy,
		"rate_limited":        rateLimited,
		"auth_failed":         authFailed,
		"unhealthy":           unhealthy,
		"ensemble_operational": ensembleOperational,
	}).Info("All providers verification completed")

	c.JSON(http.StatusOK, gin.H{
		"providers": providerResults,
		"summary": gin.H{
			"total":        len(results),
			"healthy":      healthy,
			"rate_limited": rateLimited,
			"auth_failed":  authFailed,
			"unhealthy":    unhealthy,
		},
		"ensemble_operational": ensembleOperational,
		"tested_at":            time.Now(),
	})
}

// GetProviderVerification handles GET /v1/providers/:id/verification
// This endpoint returns the last verification result for a specific provider
func (h *ProviderManagementHandler) GetProviderVerification(c *gin.Context) {
	providerID := c.Param("id")

	result := h.providerRegistry.GetProviderHealth(providerID)
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "no verification result found for provider",
			"provider_id": providerID,
			"hint":        "Use POST /v1/providers/" + providerID + "/verify to verify the provider",
		})
		return
	}

	response := gin.H{
		"provider":         result.Provider,
		"status":           result.Status,
		"verified":         result.Verified,
		"response_time_ms": result.ResponseTime.Milliseconds(),
		"tested_at":        result.TestedAt,
	}

	if result.Error != "" {
		response["error"] = result.Error
	}

	c.JSON(http.StatusOK, response)
}

// GetAllProvidersVerification handles GET /v1/providers/verification
// This endpoint returns verification results for all providers
func (h *ProviderManagementHandler) GetAllProvidersVerification(c *gin.Context) {
	results := h.providerRegistry.GetAllProviderHealth()

	// Build response with statistics
	var healthy, rateLimited, authFailed, unhealthy, unknown int
	providerResults := make([]gin.H, 0, len(results))

	for _, result := range results {
		switch result.Status {
		case services.ProviderStatusHealthy:
			healthy++
		case services.ProviderStatusRateLimited:
			rateLimited++
		case services.ProviderStatusAuthFailed:
			authFailed++
		case services.ProviderStatusUnknown:
			unknown++
		default:
			unhealthy++
		}

		providerResult := gin.H{
			"provider":         result.Provider,
			"status":           result.Status,
			"verified":         result.Verified,
			"response_time_ms": result.ResponseTime.Milliseconds(),
			"tested_at":        result.TestedAt,
		}
		if result.Error != "" {
			providerResult["error"] = result.Error
		}
		providerResults = append(providerResults, providerResult)
	}

	// Determine overall ensemble operational status
	ensembleOperational := healthy > 0

	// Get list of healthy providers for the debate group
	healthyProviders := h.providerRegistry.GetHealthyProviders()

	c.JSON(http.StatusOK, gin.H{
		"providers": providerResults,
		"summary": gin.H{
			"total":        len(results),
			"healthy":      healthy,
			"rate_limited": rateLimited,
			"auth_failed":  authFailed,
			"unhealthy":    unhealthy,
			"unknown":      unknown,
		},
		"ensemble_operational": ensembleOperational,
		"healthy_providers":    healthyProviders,
		"debate_group_ready":   len(healthyProviders) >= 2,
	})
}

// GetDiscoverySummary handles GET /v1/providers/discovery
// This endpoint returns a summary of auto-discovered providers and their scores
func (h *ProviderManagementHandler) GetDiscoverySummary(c *gin.Context) {
	discovery := h.providerRegistry.GetDiscovery()
	if discovery == nil {
		c.JSON(http.StatusOK, gin.H{
			"auto_discovery_enabled": false,
			"message":                "Provider auto-discovery is not initialized",
		})
		return
	}

	// Get discovery summary
	summary := discovery.Summary()

	h.log.WithFields(logrus.Fields{
		"total_discovered": summary["total_discovered"],
		"healthy":          summary["healthy"],
	}).Info("Provider discovery summary requested")

	c.JSON(http.StatusOK, summary)
}

// DiscoverAndVerifyProviders handles POST /v1/providers/discover
// This endpoint triggers auto-discovery and verification of all providers from environment
func (h *ProviderManagementHandler) DiscoverAndVerifyProviders(c *gin.Context) {
	h.log.Info("Starting provider auto-discovery and verification")

	// Run discovery and verification
	summary := h.providerRegistry.DiscoverAndVerifyProviders(c.Request.Context())

	// Check if auto-discovery is enabled
	if enabled, ok := summary["auto_discovery_enabled"].(bool); !ok || !enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "auto-discovery not enabled or not initialized",
			"summary": summary,
		})
		return
	}

	h.log.WithFields(logrus.Fields{
		"total_discovered": summary["total_discovered"],
		"healthy":          summary["healthy"],
		"debate_ready":     summary["debate_ready"],
	}).Info("Provider auto-discovery and verification completed")

	c.JSON(http.StatusOK, gin.H{
		"message": "Auto-discovery and verification completed",
		"summary": summary,
	})
}

// GetBestProviders handles GET /v1/providers/best
// This endpoint returns the best providers for the debate group based on scores
func (h *ProviderManagementHandler) GetBestProviders(c *gin.Context) {
	// Parse query parameters
	minProviders := 2
	maxProviders := 5

	if min := c.Query("min"); min != "" {
		if _, err := h.parseIntParam(min, &minProviders); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min parameter"})
			return
		}
	}

	if max := c.Query("max"); max != "" {
		if _, err := h.parseIntParam(max, &maxProviders); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid max parameter"})
			return
		}
	}

	// Get best providers for debate group
	bestProviders := h.providerRegistry.GetBestProvidersForDebate(minProviders, maxProviders)

	// Get detailed provider information
	discovery := h.providerRegistry.GetDiscovery()
	providerDetails := make([]gin.H, 0, len(bestProviders))

	for _, name := range bestProviders {
		detail := gin.H{
			"name": name,
		}

		// Add discovery details if available
		if discovery != nil {
			if dp := discovery.GetProviderByName(name); dp != nil {
				detail["type"] = dp.Type
				detail["score"] = dp.Score
				detail["status"] = dp.Status
				detail["verified"] = dp.Verified
				detail["default_model"] = dp.DefaultModel
			}

			if score := discovery.GetProviderScore(name); score != nil {
				detail["score_details"] = gin.H{
					"overall":        score.OverallScore,
					"response_speed": score.ResponseSpeed,
					"reliability":    score.Reliability,
					"capabilities":   score.Capabilities,
					"scored_at":      score.ScoredAt,
				}
			}
		}

		providerDetails = append(providerDetails, detail)
	}

	h.log.WithFields(logrus.Fields{
		"best_providers": len(bestProviders),
		"min_requested":  minProviders,
		"max_requested":  maxProviders,
	}).Info("Best providers requested")

	c.JSON(http.StatusOK, gin.H{
		"best_providers":     providerDetails,
		"count":              len(providerDetails),
		"min_providers":      minProviders,
		"max_providers":      maxProviders,
		"debate_group_ready": len(providerDetails) >= minProviders,
	})
}

// parseIntParam parses a string to int
func (h *ProviderManagementHandler) parseIntParam(s string, result *int) (bool, error) {
	var val int
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		val = val*10 + int(c-'0')
	}
	*result = val
	return true, nil
}

// ReDiscoverProviders handles POST /v1/providers/rediscover
// This endpoint re-runs provider discovery from environment variables
func (h *ProviderManagementHandler) ReDiscoverProviders(c *gin.Context) {
	discovery := h.providerRegistry.GetDiscovery()
	if discovery == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "auto-discovery not initialized",
		})
		return
	}

	h.log.Info("Re-running provider auto-discovery")

	// Re-discover providers
	discovered, err := discovery.DiscoverProviders()
	if err != nil {
		h.log.WithError(err).Error("Provider re-discovery failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Register any newly discovered providers
	existingProviders := h.providerRegistry.ListProviders()
	existingMap := make(map[string]bool)
	for _, name := range existingProviders {
		existingMap[name] = true
	}

	newlyRegistered := 0
	for _, dp := range discovered {
		if !existingMap[dp.Name] && dp.Provider != nil {
			if err := h.providerRegistry.RegisterProvider(dp.Name, dp.Provider); err == nil {
				newlyRegistered++
			}
		}
	}

	h.log.WithFields(logrus.Fields{
		"discovered":       len(discovered),
		"newly_registered": newlyRegistered,
	}).Info("Provider re-discovery completed")

	c.JSON(http.StatusOK, gin.H{
		"message":          "Provider re-discovery completed",
		"discovered":       len(discovered),
		"newly_registered": newlyRegistered,
		"total_providers":  len(h.providerRegistry.ListProviders()),
	})
}

// getValidProviderTypes returns dynamically discovered provider types
// DYNAMIC: No hardcoded list - types come from:
// 1. Provider mappings in discovery (from environment scan)
// 2. Known provider implementations from registry
func (h *ProviderManagementHandler) getValidProviderTypes() []string {
	seen := make(map[string]bool)
	types := make([]string, 0)

	// Get types from discovery (provider mappings scanned from environment)
	discovery := h.providerRegistry.GetDiscovery()
	if discovery != nil {
		allProviders := discovery.GetAllProviders()
		for _, p := range allProviders {
			if p.Type != "" && !seen[p.Type] {
				seen[p.Type] = true
				types = append(types, p.Type)
			}
		}
	}

	// Add known implementation types from the registry
	// These are types we have implementations for in the codebase
	knownTypes := h.providerRegistry.GetKnownProviderTypes()
	for _, t := range knownTypes {
		if !seen[t] {
			seen[t] = true
			types = append(types, t)
		}
	}

	return types
}
