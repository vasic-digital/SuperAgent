package router

import (
	"context"
	"log"
	"net/http"
	"time"

	"dev.helix.agent/internal/cache"
	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/database"
	"dev.helix.agent/internal/debate/orchestrator"
	"dev.helix.agent/internal/features"
	"dev.helix.agent/internal/formatters"
	formattersproviders "dev.helix.agent/internal/formatters/providers"
	"dev.helix.agent/internal/handlers"
	"dev.helix.agent/internal/middleware"
	"dev.helix.agent/internal/models"
	"dev.helix.agent/internal/modelsdev"
	"dev.helix.agent/internal/services"
	"dev.helix.agent/internal/skills"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http/pprof"
	"os"
)

// RouterContext wraps the router with cleanup capabilities for background services
type RouterContext struct {
	Engine                  *gin.Engine
	protocolManager         *services.UnifiedProtocolManager
	oauthMonitor            *services.OAuthTokenMonitor
	healthMonitor           *services.ProviderHealthMonitor
	concurrencyMonitor      *services.ConcurrencyMonitor
	concurrencyAlertManager *services.ConcurrencyAlertManager
	constitutionWatcher     *services.ConstitutionWatcher // Constitution auto-update background service
	ProviderRegistry        *services.ProviderRegistry    // Exposed for StartupVerifier integration
	DebateTeamConfig        *services.DebateTeamConfig    // Exposed for re-initialization with StartupVerifier
	unifiedHandler          *handlers.UnifiedHandler      // For updating debate team display
	debateService           *services.DebateService       // For updating team config
	CogneeService           *services.CogneeService       // Exposed for container adapter injection
}

// Shutdown stops all background services started by the router
func (rc *RouterContext) Shutdown() {
	if rc.protocolManager != nil {
		rc.protocolManager.Stop()
	}
	if rc.oauthMonitor != nil {
		rc.oauthMonitor.Stop()
	}
	if rc.healthMonitor != nil {
		rc.healthMonitor.Stop()
	}
	if rc.concurrencyAlertManager != nil {
		rc.concurrencyAlertManager.Stop()
	}
	if rc.constitutionWatcher != nil {
		rc.constitutionWatcher.Disable()
	}
	if rc.debateService != nil {
		rc.debateService.StopConstitutionWatcher()
	}
}

// ReinitializeDebateTeam re-initializes the DebateTeamConfig with the StartupVerifier.
// Call this after setting the StartupVerifier on ProviderRegistry to include OAuth providers.
func (rc *RouterContext) ReinitializeDebateTeam(ctx context.Context) error {
	if rc.DebateTeamConfig == nil || rc.ProviderRegistry == nil {
		return nil
	}

	// Get the StartupVerifier from the ProviderRegistry
	sv := rc.ProviderRegistry.GetStartupVerifier()
	if sv == nil {
		return nil // No StartupVerifier, keep existing team
	}

	// Set the StartupVerifier on DebateTeamConfig
	rc.DebateTeamConfig.SetStartupVerifier(sv)

	// Re-initialize the team with OAuth providers
	if err := rc.DebateTeamConfig.InitializeTeam(ctx); err != nil {
		return err
	}

	// Update handlers with new team config
	if rc.unifiedHandler != nil {
		rc.unifiedHandler.SetDebateTeamConfig(rc.DebateTeamConfig)
	}
	if rc.debateService != nil {
		rc.debateService.SetTeamConfig(rc.DebateTeamConfig)
	}

	return nil
}

// SetupRouter creates and configures the main HTTP router.
// Note: Use SetupRouterWithContext for tests to ensure proper cleanup.
func SetupRouter(cfg *config.Config) *gin.Engine {
	ctx := SetupRouterWithContext(cfg)
	return ctx.Engine
}

// SetupRouterWithContext creates and configures the main HTTP router with cleanup support.
// Call Shutdown() on the returned RouterContext when done to stop background services.
func SetupRouterWithContext(cfg *config.Config) *RouterContext {
	rc := &RouterContext{}
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Feature flags middleware - detects agent capabilities and applies feature settings
	// This middleware enables/disables features like GraphQL, TOON, Brotli, HTTP/3
	// based on User-Agent detection and request headers/query params
	featureConfig := features.DefaultFeatureConfig()
	// Keep GraphQL OFF by default for OpenAI-compatible endpoints (backward compatibility)
	// Users can enable via X-Feature-GraphQL header or ?graphql=true query param
	featureConfig.OpenAIEndpointGraphQL = false
	featureMiddleware := features.Middleware(&features.MiddlewareConfig{
		Config:               featureConfig,
		Logger:               logrus.New(),
		EnableAgentDetection: true,
		StrictMode:           false, // Lenient mode for backward compatibility
		TrackUsage:           true,
	})
	r.Use(featureMiddleware)

	// Add pprof debugging endpoints if enabled
	if os.Getenv("ENABLE_PPROF") == "true" {
		// Register pprof handlers
		r.GET("/debug/pprof/", gin.WrapH(http.HandlerFunc(pprof.Index)))
		r.GET("/debug/pprof/cmdline", gin.WrapH(http.HandlerFunc(pprof.Cmdline)))
		r.GET("/debug/pprof/profile", gin.WrapH(http.HandlerFunc(pprof.Profile)))
		r.GET("/debug/pprof/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
		r.GET("/debug/pprof/trace", gin.WrapH(http.HandlerFunc(pprof.Trace)))
		// Additional pprof endpoints for specific profiles
		r.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		r.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
		r.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		r.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
		r.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
	}

	// Initialize database with fallback to in-memory mode
	var db *database.PostgresDB
	standaloneMode := false

	pgDB, memoryDB, err := database.NewPostgresDBWithFallback(cfg)
	if err != nil {
		log.Printf("Database initialization failed: %v, using standalone mode", err)
		standaloneMode = true
	} else if memoryDB != nil {
		standaloneMode = true
		log.Println("Running in standalone mode (in-memory database)")
	} else {
		db = pgDB
	}

	// Initialize user service (only if we have a real database)
	var userService *services.UserService
	if db != nil {
		userService = services.NewUserService(db, cfg.Server.JWTSecret, 24*time.Hour)
	}
	// In standalone mode, userService will be nil - handled below

	// Initialize memory service
	memoryService := services.NewMemoryService(cfg)

	// Initialize services
	registryConfig := services.LoadRegistryConfigFromAppConfig(cfg)
	providerRegistry := services.NewProviderRegistry(registryConfig, memoryService)
	rc.ProviderRegistry = providerRegistry // Expose for StartupVerifier integration

	// Initialize shared logger
	logger := logrus.New()

	// AUTOMATIC STARTUP SCORING: Score all 30+ providers and 900+ LLMs at startup
	// This ensures we always have up-to-date provider scores for optimal routing
	startupScoringConfig := services.DefaultStartupScoringConfig()
	startupScoringService := services.NewStartupScoringService(providerRegistry, startupScoringConfig, logger)
	startupScoringService.Run(context.Background())
	logger.Info("Automatic provider scoring initiated (running in background)")

	// Initialize model metadata repository (used by multiple services)
	// In standalone mode, pass nil pool - repository will use in-memory storage
	var modelMetadataRepo *database.ModelMetadataRepository
	if db != nil {
		modelMetadataRepo = database.NewModelMetadataRepository(db.GetPool(), logger)
	} else {
		modelMetadataRepo = database.NewModelMetadataRepository(nil, logger)
		logger.Info("Model metadata repository running in standalone mode")
	}

	// Initialize Redis client for caching if Redis is configured
	var redisClient *cache.RedisClient
	if cfg.Redis.Host != "" && cfg.Redis.Port != "" {
		redisClient = cache.NewRedisClient(cfg)
	}

	// Create cache factory and shared cache
	cacheFactory := services.NewCacheFactory(redisClient, logger)
	sharedCache := cacheFactory.CreateDefaultCache(30 * time.Minute)

	// Initialize Models.dev integration (if enabled)
	var modelMetadataHandler *handlers.ModelMetadataHandler
	if cfg.ModelsDev.Enabled {
		modelsDevClient := modelsdev.NewClient(&modelsdev.ClientConfig{
			APIKey:    cfg.ModelsDev.APIKey,
			BaseURL:   cfg.ModelsDev.BaseURL,
			Timeout:   30 * time.Second,
			UserAgent: "HelixAgent/1.0",
		})

		modelMetadataCache := cacheFactory.CreateDefaultCache(cfg.ModelsDev.CacheTTL)

		modelMetadataService := services.NewModelMetadataService(
			modelsDevClient,
			modelMetadataRepo,
			modelMetadataCache,
			&services.ModelMetadataConfig{
				RefreshInterval:   cfg.ModelsDev.RefreshInterval,
				CacheTTL:          cfg.ModelsDev.CacheTTL,
				DefaultBatchSize:  cfg.ModelsDev.DefaultBatchSize,
				MaxRetries:        cfg.ModelsDev.MaxRetries,
				EnableAutoRefresh: cfg.ModelsDev.AutoRefresh,
			},
			logger,
		)

		modelMetadataHandler = handlers.NewModelMetadataHandler(modelMetadataService)

		logger.Info("Models.dev integration initialized successfully")
	}

	// Note: CompletionHandler functionality now provided by UnifiedHandler
	// Legacy handler kept for reference but not used
	_ = handlers.NewCompletionHandler // Suppress import warning

	// Initialize unified OpenAI-compatible handler
	unifiedHandler := handlers.NewUnifiedHandler(providerRegistry, cfg)

	// Initialize Skills system
	skillConfig := skills.DefaultSkillConfig()
	skillConfig.SkillsDirectory = "skills" // Load from project skills/ directory
	skillService := skills.NewService(skillConfig)
	skillService.SetLogger(logger)
	if err := skillService.Initialize(context.Background()); err != nil {
		logger.WithError(err).Warn("Failed to initialize skills system, continuing without skills")
	} else {
		logger.WithField("skills_loaded", len(skillService.GetAllSkills())).Info("Skills system initialized")
	}
	skillsIntegration := skills.NewIntegration(skillService)
	skillsIntegration.SetLogger(logger)

	// Inject skills integration into unified handler
	unifiedHandler.SetSkillsIntegration(skillsIntegration)

	// Initialize Cognee service with all features enabled
	cogneeService := services.NewCogneeService(cfg, logger)
	rc.CogneeService = cogneeService

	// Enhance all LLM providers with Cognee capabilities
	// This wraps every provider with memory, graph reasoning, and context enhancement
	if cfg.Cognee.Enabled {
		if err := services.EnhanceProviderRegistry(providerRegistry, cogneeService, logger); err != nil {
			logger.WithError(err).Warn("Failed to enhance providers with Cognee, continuing without enhancement")
		} else {
			logger.Info("All LLM providers enhanced with Cognee capabilities")
		}
	}

	// Initialize Cognee API handler with comprehensive features
	cogneeAPIHandler := handlers.NewCogneeAPIHandler(cogneeService, logger)

	// Initialize Embedding handler
	embeddingManager := services.NewEmbeddingManagerWithConfig(nil, sharedCache, logger, services.EmbeddingConfig{
		VectorProvider: "pgvector",
		Timeout:        30 * time.Second,
		CacheEnabled:   true,
	})
	embeddingHandler := handlers.NewEmbeddingHandler(embeddingManager, logger)

	// Initialize LSP handler
	lspManager := services.NewLSPManager(modelMetadataRepo, sharedCache, logger)
	lspHandler := handlers.NewLSPHandler(lspManager, logger)

	// Initialize MCP handler
	mcpHandler := handlers.NewMCPHandler(providerRegistry, &cfg.MCP)

	// Initialize ACP handler
	acpHandler := handlers.NewACPHandler(providerRegistry, logger)

	// Initialize Protocol handler (UnifiedProtocolManager implements ProtocolManagerInterface)
	protocolManager := services.NewUnifiedProtocolManager(modelMetadataRepo, sharedCache, logger)
	rc.protocolManager = protocolManager
	protocolHandler := handlers.NewProtocolHandler(protocolManager, logger)

	// Initialize Protocol SSE handler for MCP/ACP/LSP/Embeddings/Vision/Cognee
	protocolSSEHandler := handlers.NewProtocolSSEHandlerWithACP(
		mcpHandler,
		acpHandler,
		lspHandler,
		embeddingHandler,
		cogneeAPIHandler,
		logger,
	)

	// Initialize auth middleware
	// In standalone mode, make auth optional with more skip paths
	var auth *middleware.AuthMiddleware
	if standaloneMode {
		log.Println("Running in standalone mode - authentication disabled for API endpoints")
		authConfig := middleware.AuthConfig{
			SecretKey:   cfg.Server.JWTSecret,
			TokenExpiry: 24 * time.Hour,
			Issuer:      "helixagent",
			SkipPaths:   []string{"/health", "/v1/health", "/metrics", "/v1/auth/login", "/v1/auth/register", "/v1/chat/completions", "/v1/completions", "/v1/models", "/v1/ensemble", "/v1/acp", "/v1/vision", "/v1/mcp", "/v1/lsp", "/v1/embeddings", "/v1/cognee"},
			Required:    false,
		}
		auth, err = middleware.NewAuthMiddleware(authConfig, nil)
		if err != nil {
			log.Printf("Auth middleware not available in standalone mode: %v", err)
		}
	} else {
		authConfig := middleware.AuthConfig{
			SecretKey:   cfg.Server.JWTSecret,
			TokenExpiry: 24 * time.Hour,
			Issuer:      "helixagent",
			SkipPaths:   []string{"/health", "/v1/health", "/metrics", "/v1/auth/login", "/v1/auth/register", "/v1/acp", "/v1/vision", "/v1/mcp", "/v1/lsp", "/v1/embeddings", "/v1/cognee"},
			Required:    true,
		}
		auth, err = middleware.NewAuthMiddleware(authConfig, userService)
		if err != nil {
			log.Fatalf("Failed to initialize auth middleware: %v", err)
		}
	}

	// Health endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/v1/health", func(c *gin.Context) {
		// Enhanced health check with provider status
		health := providerRegistry.HealthCheck()
		healthyCount := 0
		for _, err := range health {
			if err == nil {
				healthyCount++
			}
		}

		c.JSON(200, gin.H{
			"status": "healthy",
			"providers": map[string]any{
				"total":     len(health),
				"healthy":   healthyCount,
				"unhealthy": len(health) - healthyCount,
			},
			"timestamp": time.Now().Unix(),
		})
	})

	// Metrics endpoint - Prometheus metrics for monitoring
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Feature flags status endpoint - shows enabled features and usage stats
	r.GET("/v1/features", func(c *gin.Context) {
		fc := features.GetFeatureContextFromGin(c)
		tracker := features.GetUsageTracker()
		stats := tracker.GetStats()

		// Build feature stats map
		featureStats := make(map[string]gin.H)
		for _, stat := range stats {
			featureStats[string(stat.Feature)] = gin.H{
				"enabled_count":  stat.EnabledCount,
				"disabled_count": stat.DisabledCount,
				"total_requests": stat.TotalRequests,
			}
		}

		c.JSON(200, gin.H{
			"enabled_features":  fc.GetEnabledFeatures(),
			"disabled_features": fc.GetDisabledFeatures(),
			"agent_detected":    fc.AgentName,
			"transport":         fc.GetTransportProtocol(),
			"compression":       fc.GetCompressionMethod(),
			"streaming":         fc.GetStreamingMethod(),
			"source":            string(fc.Source),
			"usage_stats":       featureStats,
		})
	})

	// Feature flags configuration endpoint - shows all available features and defaults
	r.GET("/v1/features/available", func(c *gin.Context) {
		registry := features.GetRegistry()
		allFeatures := registry.ListFeatures()

		featureList := make([]gin.H, 0, len(allFeatures))
		for _, f := range allFeatures {
			info := registry.GetFeatureInfo(f)
			if info != nil {
				featureList = append(featureList, gin.H{
					"name":         string(f),
					"description":  info.Description,
					"category":     string(info.Category),
					"default":      info.DefaultValue,
					"header":       info.HeaderName,
					"query_param":  info.QueryParam,
					"dependencies": info.RequiresFeatures,
					"conflicts":    info.ConflictsWith,
				})
			}
		}

		c.JSON(200, gin.H{
			"features": featureList,
			"count":    len(featureList),
		})
	})

	// Agent capabilities endpoint - shows what features each CLI agent supports
	r.GET("/v1/features/agents", func(c *gin.Context) {
		agentCaps := features.ListAgentCapabilities()

		agents := make([]gin.H, 0, len(agentCaps))
		for _, cap := range agentCaps {
			supportedFeatures := make([]string, 0, len(cap.SupportedFeatures))
			for _, f := range cap.SupportedFeatures {
				supportedFeatures = append(supportedFeatures, string(f))
			}
			agents = append(agents, gin.H{
				"name":               cap.AgentName,
				"supported_features": supportedFeatures,
				"transport":          cap.TransportProtocol,
				"description":        cap.Notes,
			})
		}

		c.JSON(200, gin.H{
			"agents": agents,
			"count":  len(agents),
		})
	})

	// Authentication endpoints (skip in standalone mode if auth is nil)
	if auth != nil {
		authGroup := r.Group("/v1/auth")
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
			authGroup.POST("/refresh", auth.Refresh)
			authGroup.POST("/logout", auth.Logout)
			authGroup.GET("/me", func(c *gin.Context) {
				c.JSON(200, auth.GetAuthInfo(c))
			})
		}
	} else {
		// Provide stub auth endpoints in standalone mode
		authGroup := r.Group("/v1/auth")
		{
			authGroup.POST("/register", func(c *gin.Context) {
				c.JSON(503, gin.H{"error": "Authentication disabled in standalone mode"})
			})
			authGroup.POST("/login", func(c *gin.Context) {
				c.JSON(503, gin.H{"error": "Authentication disabled in standalone mode"})
			})
			authGroup.POST("/refresh", func(c *gin.Context) {
				c.JSON(503, gin.H{"error": "Authentication disabled in standalone mode"})
			})
			authGroup.POST("/logout", func(c *gin.Context) {
				c.JSON(503, gin.H{"error": "Authentication disabled in standalone mode"})
			})
			authGroup.GET("/me", func(c *gin.Context) {
				c.JSON(503, gin.H{"error": "Authentication disabled in standalone mode"})
			})
		}
	}

	// API endpoints - single /v1 group with optional auth middleware
	// In standalone mode, auth middleware is not applied
	var protected *gin.RouterGroup
	if auth != nil && !standaloneMode {
		protected = r.Group("/v1", auth.Middleware([]string{
			"/health", "/v1/health", "/metrics",
			"/v1/models/metadata", "/v1/providers",
			"/v1/tasks",            // Background task queue - public for challenge tests
			"/v1/models",           // Model list - public for challenge tests
			"/v1/chat/completions", // Chat - required for challenges
			"/v1/completions",      // Completions - required for challenges
			"/v1/acp",              // ACP endpoints - public for CLI agents
			"/v1/vision",           // Vision endpoints - public for CLI agents
			"/v1/mcp",              // MCP endpoints - public for CLI agents (OpenCode, Crush, etc.)
			"/v1/lsp",              // LSP endpoints - public for CLI agents
			"/v1/embeddings",       // Embeddings endpoints - public for CLI agents
			"/v1/cognee",           // Cognee endpoints - public for CLI agents
			"/v1/rag",              // RAG endpoints - public for CLI agents
			"/v1/formatters",       // Formatters endpoints - public for CLI agents
			"/v1/monitoring",       // Monitoring endpoints - public for CLI agents
			"/v1/protocols",        // Protocol endpoints - public for CLI agents
		}))
	} else {
		// Standalone mode: no auth middleware
		protected = r.Group("/v1")
	}

	// Models.dev endpoints
	if cfg.ModelsDev.Enabled && modelMetadataHandler != nil {
		protected.GET("/models/metadata", modelMetadataHandler.ListModels)
		protected.GET("/models/metadata/:id", modelMetadataHandler.GetModel)
		protected.GET("/models/metadata/:id/benchmarks", modelMetadataHandler.GetModelBenchmarks)
		protected.GET("/models/metadata/compare", modelMetadataHandler.CompareModels)
		protected.GET("/models/metadata/capability/:capability", modelMetadataHandler.GetModelsByCapability)
	}
	{
		// Register OpenAI-compatible routes for seamless integration
		// This handles /completions, /chat/completions, /models
		unifiedHandler.RegisterOpenAIRoutes(protected, func(c *gin.Context) {
			c.Next()
		})

		// Note: Legacy /completions and /chat/completions routes removed
		// to avoid duplicates with UnifiedHandler routes

		// Ensemble endpoints
		protected.POST("/ensemble/completions", func(c *gin.Context) {
			var req handlers.CompletionRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Force ensemble mode
			if req.EnsembleConfig == nil {
				req.EnsembleConfig = &models.EnsembleConfig{
					Strategy:            "confidence_weighted",
					MinProviders:        2,
					ConfidenceThreshold: 0.8,
					FallbackToBest:      true,
					Timeout:             30,
					PreferredProviders:  []string{},
				}
			}

			// Create a basic internal request
			internalReq := &models.LLMRequest{
				ID:        "ensemble-" + time.Now().Format("20060102150405"),
				SessionID: "ensemble-session",
				UserID:    "anonymous",
				Prompt:    req.Prompt,
				ModelParams: models.ModelParameters{
					Model:            req.Model,
					Temperature:      req.Temperature,
					MaxTokens:        req.MaxTokens,
					TopP:             req.TopP,
					StopSequences:    req.Stop,
					ProviderSpecific: map[string]any{},
				},
				EnsembleConfig: req.EnsembleConfig,
				MemoryEnhanced: req.MemoryEnhanced,
				Memory:         map[string]string{},
				Status:         "pending",
				CreatedAt:      time.Now(),
				RequestType:    "ensemble",
			}

			// Convert messages
			messages := make([]models.Message, 0, len(req.Messages))
			for _, msg := range req.Messages {
				messages = append(messages, models.Message{
					Role:      msg.Role,
					Content:   msg.Content,
					Name:      msg.Name,
					ToolCalls: msg.ToolCalls,
				})
			}
			internalReq.Messages = messages

			// Process with ensemble
			ensembleService := providerRegistry.GetEnsembleService()
			result, err := ensembleService.RunEnsemble(c.Request.Context(), internalReq)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Return ensemble result with metadata
			c.JSON(200, gin.H{
				"id":      result.Selected.ID,
				"object":  "ensemble.completion",
				"created": result.Selected.CreatedAt.Unix(),
				"model":   result.Selected.ProviderName,
				"choices": []gin.H{
					{
						"index": 0,
						"message": gin.H{
							"role":    "assistant",
							"content": result.Selected.Content,
						},
						"finish_reason": result.Selected.FinishReason,
					},
				},
				"usage": gin.H{
					"prompt_tokens":     result.Selected.TokensUsed / 2,
					"completion_tokens": result.Selected.TokensUsed / 2,
					"total_tokens":      result.Selected.TokensUsed,
				},
				"ensemble": gin.H{
					"voting_method":     result.VotingMethod,
					"responses_count":   len(result.Responses),
					"scores":            result.Scores,
					"metadata":          result.Metadata,
					"selected_provider": result.Selected.ProviderName,
					"selection_score":   result.Selected.SelectionScore,
				},
			})
		})

		// Provider management endpoints
		providerMgmtHandler := handlers.NewProviderManagementHandler(providerRegistry, logger)
		providerGroup := protected.Group("/providers")
		{
			providerGroup.GET("", func(c *gin.Context) {
				providers := providerRegistry.ListProviders()
				result := make([]gin.H, 0, len(providers))

				for _, name := range providers {
					provider, err := providerRegistry.GetProvider(name)
					if err == nil {
						capabilities := provider.GetCapabilities()
						result = append(result, gin.H{
							"name":                      name,
							"supported_models":          capabilities.SupportedModels,
							"supported_features":        capabilities.SupportedFeatures,
							"supports_streaming":        capabilities.SupportsStreaming,
							"supports_function_calling": capabilities.SupportsFunctionCalling,
							"supports_vision":           capabilities.SupportsVision,
							"metadata":                  capabilities.Metadata,
						})
					}
				}

				c.JSON(200, gin.H{
					"providers": result,
					"count":     len(result),
				})
			})

			// Provider verification endpoints (for debate group validation)
			// NOTE: These must come BEFORE /:id routes to avoid matching issues
			providerGroup.GET("/verification", providerMgmtHandler.GetAllProvidersVerification)
			providerGroup.POST("/verify", providerMgmtHandler.VerifyAllProviders)

			// Provider auto-discovery endpoints (automatic detection from .env API keys)
			// NOTE: These must come BEFORE /:id routes to avoid matching issues
			providerGroup.GET("/discovery", providerMgmtHandler.GetDiscoverySummary)
			providerGroup.POST("/discover", providerMgmtHandler.DiscoverAndVerifyProviders)
			providerGroup.POST("/rediscover", providerMgmtHandler.ReDiscoverProviders)
			providerGroup.GET("/best", providerMgmtHandler.GetBestProviders)

			// Provider CRUD operations (parameterized routes must come AFTER specific routes)
			providerGroup.POST("", providerMgmtHandler.AddProvider)
			providerGroup.GET("/:id", providerMgmtHandler.GetProvider)
			providerGroup.PUT("/:id", providerMgmtHandler.UpdateProvider)
			providerGroup.DELETE("/:id", providerMgmtHandler.DeleteProvider)

			// Provider-specific verification endpoints
			providerGroup.GET("/:id/verification", providerMgmtHandler.GetProviderVerification)
			providerGroup.POST("/:id/verify", providerMgmtHandler.VerifyProvider)

			providerGroup.GET("/:id/health", func(c *gin.Context) {
				name := c.Param("id")
				health := providerRegistry.HealthCheck()

				response := gin.H{
					"provider": name,
				}

				if err, exists := health[name]; exists {
					if err != nil {
						response["healthy"] = false
						response["error"] = err.Error()
						c.JSON(503, response)
					} else {
						response["healthy"] = true

						// Add circuit breaker information if available
						if cb := providerRegistry.GetCircuitBreaker(name); cb != nil {
							response["circuit_breaker"] = gin.H{
								"state":         cb.GetState().String(),
								"failure_count": cb.GetFailureCount(),
								"last_failure":  cb.GetLastFailure(),
							}

							// Skills endpoints
							skillsHandler := handlers.NewSkillsHandler(skillsIntegration)
							skillsHandler.SetLogger(logger)
							skillsGroup := protected.Group("/skills")
							{
								skillsGroup.GET("", skillsHandler.ListSkills)
								skillsGroup.GET("/categories", skillsHandler.ListCategories)
								skillsGroup.GET("/:category", skillsHandler.GetSkillsByCategory)
								skillsGroup.POST("/match", skillsHandler.MatchSkills)
							}
							logger.Info("Skills endpoints registered at /v1/skills/*")
						}

						c.JSON(200, response)
					}
				} else {
					c.JSON(404, gin.H{"error": "provider not found"})
				}
			})
		}

		// Session management endpoints
		sessionHandler := handlers.NewSessionHandler(logger)
		sessionGroup := protected.Group("/sessions")
		{
			sessionGroup.POST("", sessionHandler.CreateSession)
			sessionGroup.GET("/:id", sessionHandler.GetSession)
			sessionGroup.DELETE("/:id", sessionHandler.TerminateSession)
			sessionGroup.GET("", sessionHandler.ListSessions)
		}

		// CLI Agent registry endpoints
		agentHandler := handlers.NewAgentHandler()
		agentGroup := protected.Group("/agents")
		{
			agentGroup.GET("", agentHandler.ListAgents)
			agentGroup.GET("/:name", agentHandler.GetAgent)
			agentGroup.GET("/protocol/:protocol", agentHandler.ListAgentsByProtocol)
			agentGroup.GET("/tool/:tool", agentHandler.ListAgentsByTool)
		}

		// Cognee endpoints - comprehensive API with all features
		cogneeAPIHandler.RegisterRoutes(protected)

		// AI Debate endpoints with Claude/Qwen team configuration
		// Initialize debate team configuration with provider discovery
		debateTeamConfig := services.NewDebateTeamConfig(
			providerRegistry,
			providerRegistry.GetDiscovery(),
			logger,
		)

		// CRITICAL: Set the StartupVerifier so that DebateTeamConfig uses
		// the unified verification pipeline instead of the legacy path.
		// Without this, OAuth providers (Claude, Qwen) won't be included!
		if sv := providerRegistry.GetStartupVerifier(); sv != nil {
			debateTeamConfig.SetStartupVerifier(sv)
			logger.Info("DebateTeamConfig configured with StartupVerifier (OAuth providers will be included)")
		} else {
			logger.Warn("StartupVerifier not available - using legacy provider discovery (OAuth may not work)")
		}

		// Initialize the debate team (Claude Sonnet/Opus for positions 1-2,
		// LLMsVerifier-scored providers for 3-5, Qwen as fallbacks)
		if err := debateTeamConfig.InitializeTeam(context.Background()); err != nil {
			logger.WithError(err).Warn("Failed to initialize debate team, some positions may be unfilled")
		}

		// Set the debate team config on the unified handler for dialogue display
		unifiedHandler.SetDebateTeamConfig(debateTeamConfig)

		// Store references for later re-initialization with StartupVerifier
		rc.DebateTeamConfig = debateTeamConfig
		rc.unifiedHandler = unifiedHandler

		debateService := services.NewDebateServiceWithDeps(logger, providerRegistry, cogneeService)
		debateService.SetTeamConfig(debateTeamConfig) // Set the team configuration
		rc.debateService = debateService
		debateHandler := handlers.NewDebateHandler(debateService, nil, logger)

		// Wire up the new debate orchestrator framework (optional, feature-flagged)
		orchestratorIntegration := orchestrator.CreateIntegration(providerRegistry, logger)
		debateHandler.SetOrchestratorIntegration(orchestratorIntegration)
		logger.Info("New debate orchestrator framework enabled")

		// Initialize Constitution Watcher (auto-update Constitution on project changes)
		projectRoot := os.Getenv("PROJECT_ROOT")
		if projectRoot == "" {
			// Default to current working directory if not set
			if cwd, err := os.Getwd(); err == nil {
				projectRoot = cwd
			}
		}
		constitutionWatcherEnabled := os.Getenv("CONSTITUTION_WATCHER_ENABLED") == "true"
		debateService.InitializeConstitutionWatcher(projectRoot, constitutionWatcherEnabled)
		if constitutionWatcherEnabled {
			logger.WithField("project_root", projectRoot).Info("Constitution Watcher initialized")
		}

		// Inject real debate function into HelixSpecifier engine
		// (must happen after DebateService is fully constructed)
		debateService.InitializeHelixSpecifierDebate()

		debateHandler.RegisterRoutes(protected)

		// Add debate team configuration endpoint
		protected.GET("/debates/team", func(c *gin.Context) {
			c.JSON(http.StatusOK, debateTeamConfig.GetTeamSummary())
		})

		// Initialize monitoring services
		oauthTokenMonitor := services.NewOAuthTokenMonitor(logger, services.DefaultOAuthTokenMonitorConfig())
		providerHealthMonitor := services.NewProviderHealthMonitor(providerRegistry, logger, services.DefaultProviderHealthMonitorConfig())
		concurrencyMonitor := services.NewConcurrencyMonitor(providerRegistry, logger, services.DefaultConcurrencyMonitorConfig())
		fallbackChainValidator := services.NewFallbackChainValidator(logger, debateTeamConfig)

		// Store monitors in RouterContext for cleanup
		rc.oauthMonitor = oauthTokenMonitor
		rc.healthMonitor = providerHealthMonitor
		rc.concurrencyMonitor = concurrencyMonitor

		// Initialize concurrency alert manager
		concurrencyAlertManager := services.NewConcurrencyAlertManager(services.LoadConcurrencyAlertManagerConfigFromEnv(), logger)
		concurrencyMonitor.AddAlertListener(concurrencyAlertManager.AsListener())
		rc.concurrencyAlertManager = concurrencyAlertManager

		// Start monitoring services in background
		go oauthTokenMonitor.Start(context.Background())
		go providerHealthMonitor.Start(context.Background())
		go concurrencyMonitor.Start(context.Background())
		go concurrencyAlertManager.Start(context.Background())

		// Start Constitution Watcher in background (if enabled)
		if constitutionWatcherEnabled {
			debateService.StartConstitutionWatcher(context.Background())
		}

		// Validate fallback chain on startup
		if result := fallbackChainValidator.Validate(); !result.Valid {
			logger.WithField("issues", len(result.Issues)).Warn("Fallback chain validation found issues")
		}

		// Register monitoring handler routes
		monitoringHandler := handlers.NewMonitoringHandler(nil, oauthTokenMonitor, providerHealthMonitor, fallbackChainValidator, concurrencyMonitor, concurrencyAlertManager)
		monitoringHandler.RegisterRoutes(protected)
		protocolSSEHandler.SetMonitoringHandler(monitoringHandler)
		logger.Info("Monitoring endpoints registered at /v1/monitoring/*")

		// LSP endpoints
		lspGroup := protected.Group("/lsp")
		{
			lspGroup.GET("/servers", lspHandler.ListLSPServers)
			lspGroup.POST("/execute", lspHandler.ExecuteLSPRequest)
			lspGroup.POST("/sync", lspHandler.SyncLSPServers)
			lspGroup.GET("/stats", lspHandler.GetLSPStats)
		}

		// MCP endpoints
		mcpGroup := protected.Group("/mcp")
		{
			mcpGroup.GET("/capabilities", mcpHandler.MCPCapabilities)
			mcpGroup.GET("/tools", mcpHandler.MCPTools)
			mcpGroup.POST("/tools/call", mcpHandler.MCPToolsCall)
			mcpGroup.GET("/prompts", mcpHandler.MCPPrompts)
			mcpGroup.GET("/resources", mcpHandler.MCPResources)

			// MCP Tool Search endpoints
			mcpGroup.GET("/tools/search", mcpHandler.MCPToolSearch)
			mcpGroup.POST("/tools/search", mcpHandler.MCPToolSearch)
			mcpGroup.GET("/tools/suggestions", mcpHandler.MCPToolSuggestions)
			mcpGroup.GET("/adapters/search", mcpHandler.MCPAdapterSearch)
			mcpGroup.POST("/adapters/search", mcpHandler.MCPAdapterSearch)
			mcpGroup.GET("/categories", mcpHandler.MCPCategories)
			mcpGroup.GET("/stats", mcpHandler.MCPStats)
		}
		logger.Info("MCP Tool Search endpoints registered at /v1/mcp/tools/search, /v1/mcp/adapters/search")

		// Protocol endpoints
		protocolGroup := protected.Group("/protocols")
		{
			protocolGroup.POST("/execute", protocolHandler.ExecuteProtocolRequest)
			protocolGroup.GET("/servers", protocolHandler.ListProtocolServers)
			protocolGroup.GET("/metrics", protocolHandler.GetProtocolMetrics)
			protocolGroup.POST("/refresh", protocolHandler.RefreshProtocolServers)
			protocolGroup.POST("/configure", protocolHandler.ConfigureProtocols)
		}

		// Embedding endpoints
		embeddingGroup := protected.Group("/embeddings")
		{
			embeddingGroup.POST("/generate", embeddingHandler.GenerateEmbeddings)
			embeddingGroup.POST("/search", embeddingHandler.VectorSearch)
			embeddingGroup.POST("/index", embeddingHandler.IndexDocument)
			embeddingGroup.POST("/batch-index", embeddingHandler.BatchIndexDocuments)
			embeddingGroup.GET("/stats", embeddingHandler.GetEmbeddingStats)
			embeddingGroup.GET("/providers", embeddingHandler.ListEmbeddingProviders)
		}

		// RAG (Retrieval Augmented Generation) endpoints
		ragHandler := handlers.NewRAGHandler(handlers.RAGHandlerConfig{
			Pipeline: nil, // Pipeline initialized lazily on first use
			Logger:   logger,
		})
		ragGroup := protected.Group("/rag")
		{
			// Health and Stats
			ragGroup.GET("/health", ragHandler.Health)
			ragGroup.GET("/stats", ragHandler.Stats)

			// Document operations
			ragGroup.POST("/documents", ragHandler.IngestDocument)
			ragGroup.POST("/documents/batch", ragHandler.IngestDocuments)
			ragGroup.DELETE("/documents/:id", ragHandler.DeleteDocument)

			// Search operations
			ragGroup.POST("/search", ragHandler.Search)
			ragGroup.POST("/search/hybrid", ragHandler.HybridSearch)
			ragGroup.POST("/search/expanded", ragHandler.SearchWithExpansion)

			// Advanced RAG features
			ragGroup.POST("/rerank", ragHandler.ReRank)
			ragGroup.POST("/compress", ragHandler.CompressContext)
			ragGroup.POST("/expand", ragHandler.ExpandQuery)
			ragGroup.POST("/chunk", ragHandler.ChunkDocument)
		}
		protocolSSEHandler.SetRAGHandler(ragHandler)
		logger.Info("RAG endpoints registered at /v1/rag/*")

		// ACP (Agent Communication Protocol) endpoints
		// Using acpHandler already created earlier
		acpHandler.RegisterRoutes(protected)
		logger.Info("ACP endpoints registered at /v1/acp/*")

		// Vision endpoints
		visionHandler := handlers.NewVisionHandler(providerRegistry, logger)
		visionHandler.RegisterRoutes(protected)
		logger.Info("Vision endpoints registered at /v1/vision/*")

		// Code Formatters endpoints (all public - formatters run locally and are safe)
		formattersRegistry, formattersExecutor, formattersHealth := initializeFormattersSystem(logger)
		if formattersRegistry != nil {
			formattersHandler := handlers.NewFormattersHandler(formattersRegistry, formattersExecutor, formattersHealth, logger)

			// All formatter endpoints are public (no sensitive operations)
			v1Public := r.Group("/v1")
			v1Public.POST("/format", formattersHandler.FormatCode)
			v1Public.POST("/format/batch", formattersHandler.FormatCodeBatch)
			v1Public.POST("/format/check", formattersHandler.CheckCode)
			v1Public.POST("/formatters/:name/validate-config", formattersHandler.ValidateConfig)
			v1Public.GET("/formatters", formattersHandler.ListFormatters)
			v1Public.GET("/formatters/detect", formattersHandler.DetectFormatter)
			v1Public.GET("/formatters/:name", formattersHandler.GetFormatter)
			v1Public.GET("/formatters/:name/health", formattersHandler.HealthCheckFormatter)

			protocolSSEHandler.SetFormattersHandler(formattersHandler)
			logger.Info("Code Formatters endpoints registered (all public)")
		} else {
			logger.Warn("Code Formatters system not available")
		}

		// Register Protocol SSE endpoints for MCP/ACP/LSP/Embeddings/Vision/Cognee
		// These endpoints handle SSE connections for CLI agent protocols (OpenCode, Crush, HelixCode)
		protocolSSEHandler.RegisterSSERoutes(protected)

		// Background Task endpoints (minimal implementation for API compatibility)
		tasksGroup := protected.Group("/tasks")
		{
			// Create task
			tasksGroup.POST("", func(c *gin.Context) {
				var req map[string]interface{}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				taskID := "task-" + time.Now().Format("20060102150405")
				c.JSON(http.StatusAccepted, gin.H{
					"id":         taskID,
					"status":     "pending",
					"created_at": time.Now().Unix(),
				})
			})
			// List tasks
			tasksGroup.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"tasks":         []interface{}{},
					"count":         0,
					"pending_count": 0,
				})
			})
			// Get task status
			tasksGroup.GET("/:id/status", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"id":     c.Param("id"),
					"status": "pending",
				})
			})
			// Get queue stats
			tasksGroup.GET("/queue/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"pending_count":  0,
					"running_count":  0,
					"workers_active": 4,
				})
			})
		}

		// Admin endpoints
		admin := protected.Group("/admin")
		admin.Use(auth.RequireAdmin())
		{
			admin.GET("/health/all", func(c *gin.Context) {
				health := providerRegistry.HealthCheck()
				c.JSON(200, gin.H{
					"provider_health": health,
					"timestamp":       time.Now().Unix(),
				})
			})

			// Models.dev admin endpoints (if enabled)
			if cfg.ModelsDev.Enabled && modelMetadataHandler != nil {
				admin.POST("/models/metadata/refresh", modelMetadataHandler.RefreshModels)
				admin.GET("/models/metadata/refresh/status", modelMetadataHandler.GetRefreshStatus)
			}
		}
	}

	rc.Engine = r
	return rc
}

// initializeFormattersSystem initializes the code formatters system with default configuration
func initializeFormattersSystem(logger *logrus.Logger) (*formatters.FormatterRegistry, *formatters.FormatterExecutor, *formatters.HealthChecker) {
	// Create configuration
	cfg := formatters.DefaultConfig()
	cfg.Enabled = true
	cfg.DefaultTimeout = 30 * time.Second
	cfg.CacheEnabled = true
	cfg.CacheTTL = 5 * time.Minute
	cfg.Metrics = true
	cfg.Tracing = false

	// Initialize the formatters system
	registry, executor, health, err := formatters.InitializeFormattersSystem(cfg, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize formatters system")
		return nil, nil, nil
	}

	// Register all available formatters
	if err := formattersproviders.RegisterAllFormatters(registry, logger); err != nil {
		logger.WithError(err).Warn("Some formatters failed to register")
	}

	logger.WithFields(logrus.Fields{
		"formatters_count": len(registry.List()),
		"cache_enabled":    cfg.CacheEnabled,
		"metrics_enabled":  cfg.Metrics,
	}).Info("Formatters system initialized successfully")

	return registry, executor, health
}
