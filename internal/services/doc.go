// Package services implements the core business logic layer for HelixAgent.
//
// This package provides the central services that power HelixAgent's functionality,
// including provider management, AI debate orchestration, intent classification,
// and context management.
//
// # Core Services
//
//   - ProviderRegistry: Unified interface for managing multiple LLM providers
//   - EnsembleService: Orchestrates multi-provider responses
//   - DebateService: Manages AI debate sessions for consensus building
//   - IntentClassifier: LLM-based semantic intent detection
//   - ContextManager: Manages conversation and session context
//
// # Provider Registry
//
// The ProviderRegistry provides a unified interface for provider management:
//
//	registry := services.NewProviderRegistry(config)
//	registry.RegisterProvider("claude", claudeProvider)
//	registry.RegisterProvider("gemini", geminiProvider)
//
//	// Get provider by name
//	provider, err := registry.GetProvider("claude")
//
//	// Get strongest provider by verification score
//	provider = registry.GetStrongestProvider()
//
// # AI Debate System
//
// The debate system enables multi-round discussions between providers:
//
//	debateService := services.NewDebateService(registry, config)
//
//	result, err := debateService.RunDebate(ctx, topic, participants)
//	fmt.Printf("Consensus: %s (confidence: %.2f)\n", result.Consensus, result.Confidence)
//
// The debate system supports:
//   - Multi-pass validation for response refinement
//   - Dynamic team selection based on LLMsVerifier scores
//   - 5 positions × 3 LLMs = 15 total debate participants
//
// # Intent Classification
//
// The LLM-based intent classifier provides semantic understanding:
//
//	classifier := services.NewLLMIntentClassifier(provider)
//	intent, err := classifier.Classify(ctx, "Yes, please proceed!")
//
//	// Intent types: confirmation, refusal, question, request, clarification, unclear
//
// The classifier uses zero hardcoding - all intent detection is semantic.
//
// # Debate Team Configuration
//
// The debate team is dynamically selected on startup:
//
//  1. OAuth2 providers first (Claude, Qwen) if verified
//  2. Then LLMsVerifier-scored providers by score
//  3. 5 positions × 3 LLMs (1 primary + 2 fallbacks)
//
// # Multi-Pass Validation
//
// The debate system includes multi-pass validation phases:
//
//  1. INITIAL RESPONSE: Each AI provides perspective
//  2. VALIDATION: Cross-validation for accuracy
//  3. POLISH & IMPROVE: Refinement based on feedback
//  4. FINAL CONCLUSION: Synthesized consensus
//
// # Key Files
//
//   - provider_registry.go: Provider management
//   - ensemble.go: Ensemble orchestration
//   - debate_service.go: AI debate orchestration
//   - debate_team_config.go: Team configuration
//   - debate_dialogue.go: Dialogue formatting
//   - debate_multipass_validation.go: Multi-pass validation
//   - llm_intent_classifier.go: LLM-based intent classification
//   - intent_classifier.go: Fallback pattern-based classification
//   - context_manager.go: Context management
//
// # Example: Running a Debate
//
//	config := &services.DebateConfig{
//	    EnableMultiPassValidation: true,
//	    ValidationConfig: &ValidationConfig{
//	        EnableValidation: true,
//	        EnablePolish:     true,
//	        MaxValidationRounds: 3,
//	    },
//	}
//
//	debateService := services.NewDebateService(registry, config)
//	result, err := debateService.RunDebate(ctx, "Should AI have consciousness?", nil)
package services
