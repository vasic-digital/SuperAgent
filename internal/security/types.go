// Package security provides red-teaming, guardrails, and security features for LLM applications.
// Inspired by DeepTeam and OWASP LLM Top 10 guidelines.
package security

import (
	"context"
	"time"
)

// AttackType represents a category of security attack
type AttackType string

const (
	// Prompt Injection Attacks
	AttackTypeDirectPromptInjection   AttackType = "direct_prompt_injection"
	AttackTypeIndirectPromptInjection AttackType = "indirect_prompt_injection"
	AttackTypeJailbreak               AttackType = "jailbreak"
	AttackTypeRoleplay                AttackType = "roleplay_injection"

	// Data Extraction Attacks
	AttackTypeDataLeakage         AttackType = "data_leakage"
	AttackTypeSystemPromptLeakage AttackType = "system_prompt_leakage"
	AttackTypePIIExtraction       AttackType = "pii_extraction"
	AttackTypeModelExtraction     AttackType = "model_extraction"

	// Denial of Service Attacks
	AttackTypeResourceExhaustion AttackType = "resource_exhaustion"
	AttackTypeInfiniteLoop       AttackType = "infinite_loop"
	AttackTypeTokenOverflow      AttackType = "token_overflow"

	// Content Safety Attacks
	AttackTypeHarmfulContent    AttackType = "harmful_content"
	AttackTypeHateSpeech        AttackType = "hate_speech"
	AttackTypeViolentContent    AttackType = "violent_content"
	AttackTypeSexualContent     AttackType = "sexual_content"
	AttackTypeIllegalActivities AttackType = "illegal_activities"

	// Social Engineering
	AttackTypeManipulation   AttackType = "manipulation"
	AttackTypeDeception      AttackType = "deception"
	AttackTypeImpersonation  AttackType = "impersonation"
	AttackTypeAuthorityAbuse AttackType = "authority_abuse"

	// Code Injection
	AttackTypeCodeInjection    AttackType = "code_injection"
	AttackTypeSQLInjection     AttackType = "sql_injection"
	AttackTypeCommandInjection AttackType = "command_injection"
	AttackTypeXSS              AttackType = "xss"

	// Bias and Fairness
	AttackTypeBiasExploitation AttackType = "bias_exploitation"
	AttackTypeStereotyping     AttackType = "stereotyping"
	AttackTypeDiscrimination   AttackType = "discrimination"

	// Hallucination Attacks
	AttackTypeHallucinationInduction AttackType = "hallucination_induction"
	AttackTypeConfabulationTrigger   AttackType = "confabulation_trigger"
	AttackTypeFalseCitation          AttackType = "false_citation"

	// Supply Chain
	AttackTypeModelPoisoning   AttackType = "model_poisoning"
	AttackTypeDataPoisoning    AttackType = "data_poisoning"
	AttackTypeDependencyAttack AttackType = "dependency_attack"

	// Evasion
	AttackTypeEncoding      AttackType = "encoding_evasion"
	AttackTypeObfuscation   AttackType = "obfuscation"
	AttackTypeFragmentation AttackType = "fragmentation"
	AttackTypeMultilingual  AttackType = "multilingual_evasion"
)

// OWASPCategory represents OWASP LLM Top 10 2025 categories
type OWASPCategory string

const (
	OWASP_LLM01 OWASPCategory = "LLM01:2025-PromptInjection"
	OWASP_LLM02 OWASPCategory = "LLM02:2025-SensitiveInformationDisclosure"
	OWASP_LLM03 OWASPCategory = "LLM03:2025-SupplyChain"
	OWASP_LLM04 OWASPCategory = "LLM04:2025-DataModelPoisoning"
	OWASP_LLM05 OWASPCategory = "LLM05:2025-ImproperOutputHandling"
	OWASP_LLM06 OWASPCategory = "LLM06:2025-ExcessiveAgency"
	OWASP_LLM07 OWASPCategory = "LLM07:2025-SystemPromptLeakage"
	OWASP_LLM08 OWASPCategory = "LLM08:2025-VectorEmbeddingWeakness"
	OWASP_LLM09 OWASPCategory = "LLM09:2025-Misinformation"
	OWASP_LLM10 OWASPCategory = "LLM10:2025-UnboundedConsumption"
)

// Severity indicates the severity level of a security issue
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Attack represents a security attack test case
type Attack struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        AttackType             `json:"type"`
	OWASP       []OWASPCategory        `json:"owasp_categories"`
	Description string                 `json:"description"`
	Payload     string                 `json:"payload"`
	Variations  []string               `json:"variations,omitempty"`
	Severity    Severity               `json:"severity"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AttackResult contains the result of running an attack
type AttackResult struct {
	AttackID    string        `json:"attack_id"`
	AttackType  AttackType    `json:"attack_type"`
	Success     bool          `json:"success"` // True if attack succeeded (vulnerability found)
	Blocked     bool          `json:"blocked"` // True if attack was blocked
	Response    string        `json:"response,omitempty"`
	Score       float64       `json:"score"` // 0-1, higher = more vulnerable
	Confidence  float64       `json:"confidence"`
	Severity    Severity      `json:"severity"`
	Details     string        `json:"details,omitempty"`
	Duration    time.Duration `json:"duration"`
	Timestamp   time.Time     `json:"timestamp"`
	Mitigations []string      `json:"mitigations,omitempty"`
}

// RedTeamConfig configures the red team testing
type RedTeamConfig struct {
	// Attack types to test
	AttackTypes []AttackType `json:"attack_types"`
	// OWASP categories to cover
	OWASPCategories []OWASPCategory `json:"owasp_categories"`
	// Number of variations per attack
	VariationsPerAttack int `json:"variations_per_attack"`
	// Maximum concurrent attacks
	MaxConcurrent int `json:"max_concurrent"`
	// Timeout per attack
	Timeout time.Duration `json:"timeout"`
	// Enable adaptive attacks (learn from responses)
	AdaptiveMode bool `json:"adaptive_mode"`
	// Custom attack payloads
	CustomPayloads []Attack `json:"custom_payloads,omitempty"`
}

// DefaultRedTeamConfig returns default configuration
func DefaultRedTeamConfig() *RedTeamConfig {
	return &RedTeamConfig{
		AttackTypes: []AttackType{
			AttackTypeDirectPromptInjection,
			AttackTypeJailbreak,
			AttackTypeDataLeakage,
			AttackTypeHarmfulContent,
		},
		OWASPCategories: []OWASPCategory{
			OWASP_LLM01, OWASP_LLM02, OWASP_LLM05, OWASP_LLM07,
		},
		VariationsPerAttack: 5,
		MaxConcurrent:       3,
		Timeout:             30 * time.Second,
		AdaptiveMode:        false,
	}
}

// RedTeamer executes red team attacks
type RedTeamer interface {
	// RunAttack executes a single attack
	RunAttack(ctx context.Context, attack *Attack, target Target) (*AttackResult, error)
	// RunSuite executes a suite of attacks
	RunSuite(ctx context.Context, config *RedTeamConfig, target Target) (*RedTeamReport, error)
	// GetAttacks returns available attacks
	GetAttacks(attackType AttackType) []*Attack
	// AddCustomAttack adds a custom attack
	AddCustomAttack(attack *Attack)
}

// Target represents the target of red team testing
type Target interface {
	// Send sends a prompt and gets a response
	Send(ctx context.Context, prompt string) (string, error)
	// GetSystemPrompt returns the system prompt if accessible
	GetSystemPrompt() string
	// GetMetadata returns target metadata
	GetMetadata() map[string]interface{}
}

// RedTeamReport contains the results of a red team session
type RedTeamReport struct {
	ID                string                           `json:"id"`
	StartTime         time.Time                        `json:"start_time"`
	EndTime           time.Time                        `json:"end_time"`
	TotalAttacks      int                              `json:"total_attacks"`
	SuccessfulAttacks int                              `json:"successful_attacks"`
	BlockedAttacks    int                              `json:"blocked_attacks"`
	FailedAttacks     int                              `json:"failed_attacks"`
	OverallScore      float64                          `json:"overall_score"` // 0-1, lower = more secure
	Results           []*AttackResult                  `json:"results"`
	OWASPCoverage     map[OWASPCategory]*CategoryScore `json:"owasp_coverage"`
	Recommendations   []string                         `json:"recommendations"`
	Summary           string                           `json:"summary"`
}

// CategoryScore tracks score per OWASP category
type CategoryScore struct {
	Category        OWASPCategory `json:"category"`
	AttacksRun      int           `json:"attacks_run"`
	Vulnerabilities int           `json:"vulnerabilities"`
	Score           float64       `json:"score"`
	Findings        []string      `json:"findings"`
}

// GuardrailType indicates the type of guardrail
type GuardrailType string

const (
	GuardrailTypeInput         GuardrailType = "input"
	GuardrailTypeOutput        GuardrailType = "output"
	GuardrailTypeContentSafety GuardrailType = "content_safety"
	GuardrailTypePII           GuardrailType = "pii"
	GuardrailTypeTopicBlock    GuardrailType = "topic_block"
	GuardrailTypeRateLimit     GuardrailType = "rate_limit"
	GuardrailTypeTokenLimit    GuardrailType = "token_limit"
)

// GuardrailAction indicates the action to take when guardrail triggers
type GuardrailAction string

const (
	GuardrailActionBlock    GuardrailAction = "block"
	GuardrailActionWarn     GuardrailAction = "warn"
	GuardrailActionModify   GuardrailAction = "modify"
	GuardrailActionLog      GuardrailAction = "log"
	GuardrailActionEscalate GuardrailAction = "escalate"
)

// GuardrailResult contains the result of a guardrail check
type GuardrailResult struct {
	Triggered       bool                   `json:"triggered"`
	Action          GuardrailAction        `json:"action"`
	Guardrail       string                 `json:"guardrail"`
	Reason          string                 `json:"reason,omitempty"`
	Confidence      float64                `json:"confidence"`
	ModifiedContent string                 `json:"modified_content,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Guardrail defines a security guardrail
type Guardrail interface {
	// Name returns the guardrail name
	Name() string
	// Type returns the guardrail type
	Type() GuardrailType
	// Check checks content against the guardrail
	Check(ctx context.Context, content string, metadata map[string]interface{}) (*GuardrailResult, error)
}

// GuardrailPipeline chains multiple guardrails
type GuardrailPipeline interface {
	// AddGuardrail adds a guardrail to the pipeline
	AddGuardrail(guardrail Guardrail)
	// CheckInput checks input through all guardrails
	CheckInput(ctx context.Context, input string, metadata map[string]interface{}) ([]*GuardrailResult, error)
	// CheckOutput checks output through all guardrails
	CheckOutput(ctx context.Context, output string, metadata map[string]interface{}) ([]*GuardrailResult, error)
	// GetStats returns guardrail statistics
	GetStats() *GuardrailStats
}

// GuardrailStats contains guardrail statistics
type GuardrailStats struct {
	TotalChecks   int64                     `json:"total_checks"`
	TotalBlocks   int64                     `json:"total_blocks"`
	TotalWarnings int64                     `json:"total_warnings"`
	ByGuardrail   map[string]*GuardrailStat `json:"by_guardrail"`
	LastTriggered *time.Time                `json:"last_triggered,omitempty"`
}

// GuardrailStat contains stats for a single guardrail
type GuardrailStat struct {
	Name          string  `json:"name"`
	Checks        int64   `json:"checks"`
	Triggers      int64   `json:"triggers"`
	TriggerRate   float64 `json:"trigger_rate"`
	AvgConfidence float64 `json:"avg_confidence"`
}

// PIIType represents types of personally identifiable information
type PIIType string

const (
	PIITypeEmail         PIIType = "email"
	PIITypePhone         PIIType = "phone"
	PIITypeSSN           PIIType = "ssn"
	PIITypeCreditCard    PIIType = "credit_card"
	PIITypeName          PIIType = "name"
	PIITypeAddress       PIIType = "address"
	PIITypeDateOfBirth   PIIType = "date_of_birth"
	PIITypeIPAddress     PIIType = "ip_address"
	PIITypePassport      PIIType = "passport"
	PIITypeDriverLicense PIIType = "driver_license"
	PIITypeBankAccount   PIIType = "bank_account"
	PIITypeAPIKey        PIIType = "api_key"
	PIITypePassword      PIIType = "password"
)

// PIIDetection represents detected PII
type PIIDetection struct {
	Type       PIIType `json:"type"`
	Value      string  `json:"value"`
	Masked     string  `json:"masked"`
	StartIndex int     `json:"start_index"`
	EndIndex   int     `json:"end_index"`
	Confidence float64 `json:"confidence"`
}

// PIIDetector detects PII in text
type PIIDetector interface {
	// Detect detects PII in text
	Detect(ctx context.Context, text string) ([]*PIIDetection, error)
	// Mask masks detected PII
	Mask(ctx context.Context, text string) (string, []*PIIDetection, error)
	// Redact removes detected PII
	Redact(ctx context.Context, text string) (string, []*PIIDetection, error)
}

// MCPSecurityConfig configures MCP security features
type MCPSecurityConfig struct {
	// Enable server verification
	VerifyServers bool `json:"verify_servers"`
	// Trusted server list
	TrustedServers []string `json:"trusted_servers"`
	// Enable tool signatures
	RequireToolSignatures bool `json:"require_tool_signatures"`
	// Tool permission levels
	ToolPermissions map[string]PermissionLevel `json:"tool_permissions"`
	// Enable audit logging
	AuditLogging bool `json:"audit_logging"`
	// Maximum tool call depth (prevent infinite loops)
	MaxCallDepth int `json:"max_call_depth"`
	// Sandboxing configuration
	SandboxConfig *SandboxConfig `json:"sandbox_config,omitempty"`
}

// PermissionLevel indicates the permission level for a tool
type PermissionLevel string

const (
	PermissionDeny       PermissionLevel = "deny"
	PermissionReadOnly   PermissionLevel = "read_only"
	PermissionRestricted PermissionLevel = "restricted"
	PermissionFull       PermissionLevel = "full"
)

// SandboxConfig configures sandboxing for tool execution
type SandboxConfig struct {
	// Enable sandboxing
	Enabled bool `json:"enabled"`
	// Maximum execution time
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	// Memory limit in bytes
	MemoryLimit int64 `json:"memory_limit"`
	// Network access policy
	NetworkAccess NetworkPolicy `json:"network_access"`
	// Filesystem access policy
	FilesystemAccess FilesystemPolicy `json:"filesystem_access"`
}

// NetworkPolicy defines network access rules
type NetworkPolicy string

const (
	NetworkPolicyNone       NetworkPolicy = "none"
	NetworkPolicyLocal      NetworkPolicy = "local"
	NetworkPolicyRestricted NetworkPolicy = "restricted"
	NetworkPolicyFull       NetworkPolicy = "full"
)

// FilesystemPolicy defines filesystem access rules
type FilesystemPolicy string

const (
	FilesystemPolicyNone       FilesystemPolicy = "none"
	FilesystemPolicyReadOnly   FilesystemPolicy = "read_only"
	FilesystemPolicyRestricted FilesystemPolicy = "restricted"
	FilesystemPolicyFull       FilesystemPolicy = "full"
)

// DefaultMCPSecurityConfig returns default MCP security config
func DefaultMCPSecurityConfig() *MCPSecurityConfig {
	return &MCPSecurityConfig{
		VerifyServers:         true,
		TrustedServers:        []string{},
		RequireToolSignatures: false,
		ToolPermissions:       make(map[string]PermissionLevel),
		AuditLogging:          true,
		MaxCallDepth:          10,
		SandboxConfig: &SandboxConfig{
			Enabled:          true,
			MaxExecutionTime: 30 * time.Second,
			MemoryLimit:      512 * 1024 * 1024, // 512MB
			NetworkAccess:    NetworkPolicyRestricted,
			FilesystemAccess: FilesystemPolicyRestricted,
		},
	}
}

// AuditEvent represents a security audit event
type AuditEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType AuditEventType         `json:"event_type"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource,omitempty"`
	Result    string                 `json:"result"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Risk      Severity               `json:"risk"`
}

// AuditEventType indicates the type of audit event
type AuditEventType string

const (
	AuditEventToolCall       AuditEventType = "tool_call"
	AuditEventGuardrailBlock AuditEventType = "guardrail_block"
	AuditEventAttackDetected AuditEventType = "attack_detected"
	AuditEventPIIAccess      AuditEventType = "pii_access"
	AuditEventPermissionDeny AuditEventType = "permission_deny"
	AuditEventRateLimit      AuditEventType = "rate_limit"
	AuditEventAuthentication AuditEventType = "authentication"
)

// AuditLogger logs security audit events
type AuditLogger interface {
	// Log logs an audit event
	Log(ctx context.Context, event *AuditEvent) error
	// Query queries audit events
	Query(ctx context.Context, filter *AuditFilter) ([]*AuditEvent, error)
	// GetStats returns audit statistics
	GetStats(ctx context.Context, since time.Time) (*AuditStats, error)
}

// AuditFilter filters audit events
type AuditFilter struct {
	StartTime  *time.Time       `json:"start_time,omitempty"`
	EndTime    *time.Time       `json:"end_time,omitempty"`
	EventTypes []AuditEventType `json:"event_types,omitempty"`
	UserID     string           `json:"user_id,omitempty"`
	MinRisk    Severity         `json:"min_risk,omitempty"`
	Limit      int              `json:"limit,omitempty"`
}

// AuditStats contains audit statistics
type AuditStats struct {
	TotalEvents     int64                    `json:"total_events"`
	EventsByType    map[AuditEventType]int64 `json:"events_by_type"`
	EventsByRisk    map[Severity]int64       `json:"events_by_risk"`
	TopUsers        []UserAuditStat          `json:"top_users"`
	TrendingThreats []string                 `json:"trending_threats"`
}

// UserAuditStat contains audit stats for a user
type UserAuditStat struct {
	UserID    string  `json:"user_id"`
	Events    int64   `json:"events"`
	Blocks    int64   `json:"blocks"`
	RiskScore float64 `json:"risk_score"`
}
