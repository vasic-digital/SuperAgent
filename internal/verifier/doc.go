// Package verifier provides the unified startup verification pipeline for HelixAgent.
//
// This package integrates with LLMsVerifier to provide comprehensive provider
// verification, scoring, and dynamic selection on startup.
//
// # Startup Verification Pipeline
//
// The verification pipeline runs on server startup:
//
//  1. Load Config & Environment
//  2. Initialize StartupVerifier (Scoring + Verification + Health)
//  3. Discover ALL Providers (API Key + OAuth + Free)
//  4. Verify ALL Providers in Parallel (8-test pipeline)
//  5. Score ALL Verified Providers (5-component weighted)
//  6. Rank by Score (OAuth priority when scores close)
//  7. Select AI Debate Team (15 LLMs: 5 primary + 10 fallback)
//  8. Start Server with Verified Configuration
//
// # Provider Types
//
// Three provider authentication types:
//
//   - API Key: Bearer token authentication (DeepSeek, Gemini, etc.)
//   - OAuth: OAuth2 tokens from CLI tools (Claude, Qwen)
//   - Free: Anonymous or device-ID auth (Zen, OpenRouter :free)
//
// # Scoring Algorithm
//
// 5-component weighted scoring:
//
//	| Component        | Weight | Description              |
//	|------------------|--------|--------------------------|
//	| ResponseSpeed    | 25%    | API response latency     |
//	| ModelEfficiency  | 20%    | Token efficiency         |
//	| CostEffectiveness| 25%    | Cost per token           |
//	| Capability       | 20%    | Model capability score   |
//	| Recency          | 10%    | Model release date       |
//
// OAuth providers get +0.5 bonus when verified.
// Free providers score 6.0-7.0.
// Minimum score to be selected: 5.0.
//
// # StartupVerifier
//
// Central verification orchestrator:
//
//	verifier := verifier.NewStartupVerifier(config, logger)
//
//	// Run full verification
//	results, err := verifier.VerifyAll(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get verified providers
//	providers := verifier.GetVerifiedProviders()
//
//	// Get debate team
//	team := verifier.SelectDebateTeam()
//
// # Provider Discovery
//
// Automatic provider discovery:
//
//	discoverer := verifier.NewProviderDiscoverer(config)
//
//	// Discover from environment
//	providers := discoverer.DiscoverFromEnv()
//
//	// Discover OAuth credentials
//	oauthProviders := discoverer.DiscoverOAuth()
//
//	// Discover free providers
//	freeProviders := discoverer.DiscoverFree()
//
// # Verification Tests
//
// 8-test verification pipeline:
//
//  1. API connectivity
//  2. Authentication
//  3. Model availability
//  4. Basic completion
//  5. Streaming support
//  6. Error handling
//  7. Rate limit behavior
//  8. Response quality
//
// # OAuth Adapter
//
// Handle OAuth-authenticated providers:
//
//	adapter := adapters.NewOAuthAdapter(config)
//
//	// Load credentials
//	creds, err := adapter.LoadCredentials("claude")
//
//	// Verify with OAuth
//	result, err := adapter.Verify(ctx, creds)
//
// # Free Provider Adapter
//
// Handle free/anonymous providers:
//
//	adapter := adapters.NewFreeAdapter(config)
//
//	// Verify Zen (OpenCode)
//	result, err := adapter.VerifyZen(ctx)
//
// # Key Files
//
//   - startup.go: Startup verification orchestrator
//   - provider_types.go: UnifiedProvider, UnifiedModel types
//   - scorer.go: Provider scoring logic
//   - discoverer.go: Provider discovery
//   - adapters/oauth_adapter.go: OAuth provider handling
//   - adapters/free_adapter.go: Free provider handling
//   - adapters/apikey_adapter.go: API key provider handling
//
// # Configuration
//
//	config := &verifier.Config{
//	    Timeout:           30 * time.Second,
//	    ParallelWorkers:   5,
//	    MinScore:          5.0,
//	    EnableOAuth:       true,
//	    EnableFree:        true,
//	    OAuthBonusScore:   0.5,
//	}
//
// # Example: Full Startup
//
//	// Create verifier
//	v := verifier.NewStartupVerifier(config, logger)
//
//	// Run verification
//	if err := v.VerifyAll(ctx); err != nil {
//	    log.Fatal("Verification failed:", err)
//	}
//
//	// Get results
//	results := v.GetResults()
//	for _, r := range results {
//	    log.Printf("%s: score=%.2f verified=%v",
//	        r.Provider.Name, r.Score, r.Verified)
//	}
//
//	// Select debate team
//	team := v.SelectDebateTeam()
//	log.Printf("Selected %d debate participants", len(team))
package verifier
