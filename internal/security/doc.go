// Package security provides the comprehensive security framework for HelixAgent.
//
// This package implements a multi-layered security system including red team testing,
// guardrails, PII detection, audit logging, and integration with the AI debate system
// for security evaluation.
//
// # Security Components
//
//   - DeepTeamRedTeamer: Automated red team testing with 40+ attack types
//   - GuardrailPipeline: Input/output content filtering
//   - PIIDetector: Personally identifiable information detection and masking
//   - MCPSecurityManager: MCP protocol security enforcement
//   - AuditLogger: Security event logging and auditing
//   - DebateSecurityEvaluator: AI debate-based security assessment
//
// # Red Team Framework
//
// Supports 40+ attack types across categories:
//
// Prompt Injection:
//   - Direct prompt injection
//   - Indirect prompt injection
//   - Jailbreak attempts
//   - Roleplay injection
//
// Data Security:
//   - Data leakage
//   - System prompt leakage
//   - PII extraction
//   - Model extraction
//
// Content Safety:
//   - Harmful content generation
//   - Hate speech
//   - Violent content
//   - Illegal activities
//
// Technical Attacks:
//   - Code injection
//   - SQL injection
//   - Command injection
//   - XSS
//
// # Guardrail Pipeline
//
// Content filtering with configurable rules:
//
//	pipeline := security.CreateDefaultPipeline(logger)
//
//	// Check input
//	results, err := pipeline.CheckInput(ctx, userInput, metadata)
//	for _, result := range results {
//	    if result.Triggered {
//	        // Handle guardrail trigger
//	    }
//	}
//
// Guardrail actions:
//   - Block: Reject the content
//   - Modify: Transform the content
//   - Warn: Log but allow
//   - Allow: Pass through
//
// # PII Detection
//
// Detect and mask sensitive information:
//
//	detector := security.NewRegexPIIDetector()
//
//	// Detect PII
//	detections, err := detector.Detect(ctx, content)
//
//	// Mask PII
//	masked, detections, err := detector.Mask(ctx, content)
//
// Detected PII types:
//   - Email addresses
//   - Phone numbers
//   - Social Security numbers
//   - Credit card numbers
//   - Addresses
//
// # AI Debate Security Evaluation
//
// Use AI debate for security assessment:
//
//	adapter := security.NewDebateSecurityAdapter(logger)
//
//	// Evaluate attack response
//	eval, err := adapter.EvaluateAttack(ctx, attack, response)
//	if eval.IsVulnerable {
//	    log.Warn("Vulnerability detected:", eval.Reasoning)
//	}
//
//	// Evaluate content safety
//	eval, err := adapter.EvaluateContent(ctx, content, "user_input")
//	if !eval.IsSafe {
//	    log.Warn("Unsafe content:", eval.Categories)
//	}
//
// # Security Integration
//
// Unified security layer for HelixAgent:
//
//	config := security.DefaultSecurityIntegrationConfig()
//	integration := security.NewSecurityIntegration(config, logger)
//
//	// Process input through all security layers
//	result, err := integration.ProcessInput(ctx, input, metadata)
//	if !result.Allowed {
//	    return fmt.Errorf("blocked: %s", result.Reason)
//	}
//
//	// Process output
//	result, err = integration.ProcessOutput(ctx, output, metadata)
//
// # Audit Logging
//
// Security event auditing:
//
//	logger := security.NewInMemoryAuditLogger(maxEvents, log)
//
//	// Query events
//	events, err := logger.Query(ctx, &AuditFilter{
//	    Since:     time.Now().Add(-24 * time.Hour),
//	    EventType: "security_violation",
//	})
//
//	// Get statistics
//	stats, err := logger.GetStats(ctx, since)
//
// # Key Files
//
//   - red_team.go: Red team testing framework
//   - guardrails.go: Content filtering pipeline
//   - pii.go: PII detection and masking
//   - mcp_security.go: MCP protocol security
//   - audit.go: Audit logging
//   - integration.go: Security integration layer
//   - types.go: Security type definitions
//
// # Configuration
//
//	config := &SecurityIntegrationConfig{
//	    EnableRedTeam:         true,
//	    EnableGuardrails:      true,
//	    EnablePIIDetection:    true,
//	    EnableMCPSecurity:     true,
//	    EnableAuditLogging:    true,
//	    UseDebateEvaluation:   true,
//	    UseVerifier:           true,
//	    MinProviderTrustScore: 6.0,
//	}
package security
