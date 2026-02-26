// Package governance provides the SEMAP Protocol for AI agent governance.
// SEMAP (Semantic Agent Protocol) implements Design-by-Contract for agent interactions.
package governance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SEMAPConfig configures the SEMAP Protocol.
type SEMAPConfig struct {
	// EnableStrictMode enforces all contracts strictly
	EnableStrictMode bool `json:"enable_strict_mode"`
	// EnableAuditLog logs all contract checks
	EnableAuditLog bool `json:"enable_audit_log"`
	// MaxContractDepth is the maximum depth for nested contracts
	MaxContractDepth int `json:"max_contract_depth"`
	// TimeoutDefault is the default timeout for contract checks
	TimeoutDefault time.Duration `json:"timeout_default"`
	// EnableAutoRemediation allows automatic remediation of violations
	EnableAutoRemediation bool `json:"enable_auto_remediation"`
	// MaxViolationsBeforeHalt is the max violations before halting execution
	MaxViolationsBeforeHalt int `json:"max_violations_before_halt"`
}

// DefaultSEMAPConfig returns a default configuration.
func DefaultSEMAPConfig() SEMAPConfig {
	return SEMAPConfig{
		EnableStrictMode:        true,
		EnableAuditLog:          true,
		MaxContractDepth:        10,
		TimeoutDefault:          30 * time.Second,
		EnableAutoRemediation:   true,
		MaxViolationsBeforeHalt: 3,
	}
}

// ContractType represents the type of a contract.
type ContractType string

const (
	ContractTypePrecondition  ContractType = "precondition"
	ContractTypePostcondition ContractType = "postcondition"
	ContractTypeInvariant     ContractType = "invariant"
	ContractTypeGuardRail     ContractType = "guardrail"
	ContractTypeAssertion     ContractType = "assertion"
)

// ViolationSeverity represents the severity of a contract violation.
type ViolationSeverity string

const (
	ViolationSeverityInfo     ViolationSeverity = "info"
	ViolationSeverityWarning  ViolationSeverity = "warning"
	ViolationSeverityError    ViolationSeverity = "error"
	ViolationSeverityCritical ViolationSeverity = "critical"
)

// Contract represents a contract specification.
type Contract struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        ContractType      `json:"type"`
	Severity    ViolationSeverity `json:"severity"`
	Conditions  []Condition       `json:"conditions"`
	Actions     []ContractAction  `json:"actions,omitempty"`
	Metadata    ContractMetadata  `json:"metadata"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Condition represents a condition to be evaluated.
type Condition struct {
	ID          string                 `json:"id"`
	Expression  string                 `json:"expression"`
	Evaluator   ConditionType          `json:"evaluator"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Message     string                 `json:"message"`
	FailMessage string                 `json:"fail_message,omitempty"`
	// Negate inverts the condition result (true becomes false and vice versa)
	// Used for blocklist patterns where matching = violation
	Negate bool `json:"negate,omitempty"`
}

// ConditionType represents the type of condition evaluator.
type ConditionType string

const (
	ConditionTypeRegex      ConditionType = "regex"
	ConditionTypeJSONSchema ConditionType = "json_schema"
	ConditionTypeScript     ConditionType = "script"
	ConditionTypeRange      ConditionType = "range"
	ConditionTypeEnum       ConditionType = "enum"
	ConditionTypeCustom     ConditionType = "custom"
	ConditionTypeLLM        ConditionType = "llm"
)

// ContractAction represents an action to take on violation.
type ContractAction struct {
	Type       ActionType             `json:"type"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ActionType represents the type of action.
type ActionType string

const (
	ActionTypeLog       ActionType = "log"
	ActionTypeAlert     ActionType = "alert"
	ActionTypeBlock     ActionType = "block"
	ActionTypeRemediate ActionType = "remediate"
	ActionTypeRollback  ActionType = "rollback"
	ActionTypeNotify    ActionType = "notify"
	ActionTypeEscalate  ActionType = "escalate"
)

// ContractMetadata contains metadata about a contract.
type ContractMetadata struct {
	Author     string   `json:"author"`
	Version    string   `json:"version"`
	Tags       []string `json:"tags,omitempty"`
	Category   string   `json:"category,omitempty"`
	References []string `json:"references,omitempty"`
	Deprecated bool     `json:"deprecated,omitempty"`
	Priority   int      `json:"priority"`
}

// AgentCapability represents a capability that an agent can have.
type AgentCapability string

const (
	CapabilityRead         AgentCapability = "read"
	CapabilityWrite        AgentCapability = "write"
	CapabilityExecute      AgentCapability = "execute"
	CapabilityNetwork      AgentCapability = "network"
	CapabilityFileSystem   AgentCapability = "filesystem"
	CapabilityProcess      AgentCapability = "process"
	CapabilityDatabase     AgentCapability = "database"
	CapabilitySecrets      AgentCapability = "secrets"
	CapabilityAuthenticate AgentCapability = "authenticate"
	CapabilityAuthorize    AgentCapability = "authorize"
)

// AgentProfile defines the permissions and constraints for an agent.
type AgentProfile struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Capabilities   []AgentCapability `json:"capabilities"`
	Constraints    []Constraint      `json:"constraints"`
	RateLimits     []RateLimit       `json:"rate_limits,omitempty"`
	ResourceLimits *ResourceLimits   `json:"resource_limits,omitempty"`
	Policies       []string          `json:"policies"` // Policy IDs
	TrustLevel     TrustLevel        `json:"trust_level"`
	CreatedAt      time.Time         `json:"created_at"`
	ExpiresAt      *time.Time        `json:"expires_at,omitempty"`
}

// Constraint defines a constraint on agent behavior.
type Constraint struct {
	Type       ConstraintType         `json:"type"`
	Target     string                 `json:"target"` // What the constraint applies to
	Allowed    []string               `json:"allowed,omitempty"`
	Denied     []string               `json:"denied,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ConstraintType represents the type of constraint.
type ConstraintType string

const (
	ConstraintTypePathPattern        ConstraintType = "path_pattern"
	ConstraintTypeCommandWhitelist   ConstraintType = "command_whitelist"
	ConstraintTypeCommandBlacklist   ConstraintType = "command_blacklist"
	ConstraintTypeResourceQuota      ConstraintType = "resource_quota"
	ConstraintTypeTimeWindow         ConstraintType = "time_window"
	ConstraintTypeIPRange            ConstraintType = "ip_range"
	ConstraintTypeDataClassification ConstraintType = "data_classification"
)

// RateLimit defines rate limiting for an agent.
type RateLimit struct {
	Resource   string        `json:"resource"`
	Limit      int           `json:"limit"`
	Window     time.Duration `json:"window"`
	BurstLimit int           `json:"burst_limit,omitempty"`
}

// ResourceLimits defines resource limits for an agent.
type ResourceLimits struct {
	MaxMemoryMB      int           `json:"max_memory_mb"`
	MaxCPUPercent    int           `json:"max_cpu_percent"`
	MaxDiskWriteMB   int           `json:"max_disk_write_mb"`
	MaxNetworkMBps   int           `json:"max_network_mbps"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
}

// TrustLevel represents the trust level of an agent.
type TrustLevel string

const (
	TrustLevelUntrusted TrustLevel = "untrusted"
	TrustLevelLow       TrustLevel = "low"
	TrustLevelMedium    TrustLevel = "medium"
	TrustLevelHigh      TrustLevel = "high"
	TrustLevelTrusted   TrustLevel = "trusted"
)

// Policy represents a governance policy.
type Policy struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Contracts   []string    `json:"contracts"` // Contract IDs
	Scope       PolicyScope `json:"scope"`
	Priority    int         `json:"priority"`
	Enabled     bool        `json:"enabled"`
	EnforcedAt  time.Time   `json:"enforced_at"`
}

// PolicyScope defines the scope of a policy.
type PolicyScope struct {
	Agents       []string `json:"agents,omitempty"`       // Agent IDs or patterns
	Actions      []string `json:"actions,omitempty"`      // Action types
	Resources    []string `json:"resources,omitempty"`    // Resource patterns
	Environments []string `json:"environments,omitempty"` // Environment names
}

// Violation represents a contract violation.
type Violation struct {
	ID           string            `json:"id"`
	ContractID   string            `json:"contract_id"`
	ContractName string            `json:"contract_name"`
	Type         ContractType      `json:"type"`
	Severity     ViolationSeverity `json:"severity"`
	Message      string            `json:"message"`
	Context      ViolationContext  `json:"context"`
	Remediation  *Remediation      `json:"remediation,omitempty"`
	OccurredAt   time.Time         `json:"occurred_at"`
	ResolvedAt   *time.Time        `json:"resolved_at,omitempty"`
}

// ViolationContext provides context about a violation.
type ViolationContext struct {
	AgentID     string                 `json:"agent_id"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource,omitempty"`
	Input       interface{}            `json:"input,omitempty"`
	Output      interface{}            `json:"output,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Remediation represents a remediation action taken.
type Remediation struct {
	Type      ActionType `json:"type"`
	Status    string     `json:"status"` // "pending", "applied", "failed"
	Details   string     `json:"details"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
}

// SEMAP is the main SEMAP Protocol implementation.
type SEMAP struct {
	config     SEMAPConfig
	contracts  map[string]*Contract
	policies   map[string]*Policy
	profiles   map[string]*AgentProfile
	violations []*Violation
	evaluators map[ConditionType]ConditionEvaluator
	actions    map[ActionType]ActionHandler
	auditLog   *AuditLog
	mu         sync.RWMutex
}

// ConditionEvaluator evaluates a condition.
type ConditionEvaluator interface {
	Evaluate(ctx context.Context, condition *Condition, value interface{}) (bool, error)
}

// ActionHandler handles a contract action.
type ActionHandler interface {
	Execute(ctx context.Context, action *ContractAction, violation *Violation) error
}

// NewSEMAP creates a new SEMAP Protocol instance.
func NewSEMAP(config SEMAPConfig) *SEMAP {
	s := &SEMAP{
		config:     config,
		contracts:  make(map[string]*Contract),
		policies:   make(map[string]*Policy),
		profiles:   make(map[string]*AgentProfile),
		violations: make([]*Violation, 0),
		evaluators: make(map[ConditionType]ConditionEvaluator),
		actions:    make(map[ActionType]ActionHandler),
	}

	if config.EnableAuditLog {
		s.auditLog = NewAuditLog()
	}

	// Register default evaluators
	s.registerDefaultEvaluators()
	s.registerDefaultActions()

	return s
}

// RegisterContract registers a contract.
func (s *SEMAP) RegisterContract(contract *Contract) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if contract.ID == "" {
		contract.ID = s.generateContractID(contract)
	}

	contract.CreatedAt = time.Now()
	contract.UpdatedAt = contract.CreatedAt
	contract.Enabled = true

	s.contracts[contract.ID] = contract

	if s.auditLog != nil {
		s.auditLog.Log(AuditEntry{
			Type:       "contract_registered",
			ContractID: contract.ID,
			Timestamp:  time.Now(),
		})
	}

	return nil
}

// RegisterPolicy registers a policy.
func (s *SEMAP) RegisterPolicy(policy *Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if policy.ID == "" {
		policy.ID = s.generatePolicyID(policy)
	}

	policy.EnforcedAt = time.Now()
	policy.Enabled = true

	s.policies[policy.ID] = policy

	return nil
}

// RegisterAgentProfile registers an agent profile.
func (s *SEMAP) RegisterAgentProfile(profile *AgentProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile.CreatedAt = time.Now()
	s.profiles[profile.ID] = profile

	return nil
}

// CheckPreconditions checks preconditions before an action.
func (s *SEMAP) CheckPreconditions(ctx context.Context, agentID string, action string, input interface{}) (*CheckResult, error) {
	return s.checkContracts(ctx, agentID, action, input, nil, ContractTypePrecondition)
}

// CheckPostconditions checks postconditions after an action.
func (s *SEMAP) CheckPostconditions(ctx context.Context, agentID string, action string, input interface{}, output interface{}) (*CheckResult, error) {
	return s.checkContracts(ctx, agentID, action, input, output, ContractTypePostcondition)
}

// CheckInvariants checks invariants during execution.
func (s *SEMAP) CheckInvariants(ctx context.Context, agentID string, state interface{}) (*CheckResult, error) {
	return s.checkContracts(ctx, agentID, "invariant_check", state, nil, ContractTypeInvariant)
}

// CheckGuardRails checks guard rails for an action.
func (s *SEMAP) CheckGuardRails(ctx context.Context, agentID string, action string, input interface{}) (*CheckResult, error) {
	return s.checkContracts(ctx, agentID, action, input, nil, ContractTypeGuardRail)
}

// CheckResult represents the result of a contract check.
type CheckResult struct {
	Passed           bool          `json:"passed"`
	Violations       []*Violation  `json:"violations,omitempty"`
	Warnings         []*Violation  `json:"warnings,omitempty"`
	CheckedAt        time.Time     `json:"checked_at"`
	Duration         time.Duration `json:"duration"`
	ContractsChecked int           `json:"contracts_checked"`
}

func (s *SEMAP) checkContracts(ctx context.Context, agentID string, action string, input interface{}, output interface{}, contractType ContractType) (*CheckResult, error) {
	start := time.Now()

	s.mu.RLock()
	profile := s.profiles[agentID]
	contracts := s.getApplicableContracts(agentID, action, contractType)
	s.mu.RUnlock()

	result := &CheckResult{
		Passed:           true,
		Violations:       make([]*Violation, 0),
		Warnings:         make([]*Violation, 0),
		CheckedAt:        start,
		ContractsChecked: len(contracts),
	}

	// Check profile constraints first
	if profile != nil {
		if err := s.checkProfileConstraints(ctx, profile, action, input); err != nil {
			violation := &Violation{
				ID:           s.generateViolationID(),
				ContractID:   "profile_constraint",
				ContractName: "Profile Constraint",
				Type:         ContractTypeGuardRail,
				Severity:     ViolationSeverityError,
				Message:      err.Error(),
				Context: ViolationContext{
					AgentID: agentID,
					Action:  action,
					Input:   input,
				},
				OccurredAt: time.Now(),
			}
			result.Violations = append(result.Violations, violation)
			result.Passed = false
		}
	}

	// Check each contract
	for _, contract := range contracts {
		if !contract.Enabled {
			continue
		}

		passed, err := s.evaluateContract(ctx, contract, input, output)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate contract %s: %w", contract.ID, err)
		}

		if !passed {
			violation := &Violation{
				ID:           s.generateViolationID(),
				ContractID:   contract.ID,
				ContractName: contract.Name,
				Type:         contract.Type,
				Severity:     contract.Severity,
				Message:      s.formatViolationMessage(contract, input, output),
				Context: ViolationContext{
					AgentID: agentID,
					Action:  action,
					Input:   input,
					Output:  output,
				},
				OccurredAt: time.Now(),
			}

			if contract.Severity == ViolationSeverityWarning || contract.Severity == ViolationSeverityInfo {
				result.Warnings = append(result.Warnings, violation)
			} else {
				result.Violations = append(result.Violations, violation)
				result.Passed = false

				// Execute contract actions
				for _, action := range contract.Actions {
					s.executeAction(ctx, &action, violation)
				}
			}

			// Store violation
			s.mu.Lock()
			s.violations = append(s.violations, violation)
			s.mu.Unlock()

			// Auto-remediation if enabled
			if s.config.EnableAutoRemediation && len(result.Violations) <= s.config.MaxViolationsBeforeHalt {
				s.attemptRemediation(ctx, violation)
			}
		}
	}

	result.Duration = time.Since(start)

	// Audit log
	if s.auditLog != nil {
		s.auditLog.Log(AuditEntry{
			Type:       "contract_check",
			AgentID:    agentID,
			Action:     action,
			Result:     result.Passed,
			Violations: len(result.Violations),
			Timestamp:  time.Now(),
		})
	}

	return result, nil
}

func (s *SEMAP) getApplicableContracts(agentID string, action string, contractType ContractType) []*Contract {
	var applicable []*Contract

	// Get contracts from policies
	profile := s.profiles[agentID]
	if profile != nil {
		for _, policyID := range profile.Policies {
			if policy, ok := s.policies[policyID]; ok && policy.Enabled {
				for _, contractID := range policy.Contracts {
					if contract, ok := s.contracts[contractID]; ok {
						if contract.Type == contractType {
							applicable = append(applicable, contract)
						}
					}
				}
			}
		}
	}

	// Get global contracts of the specified type
	for _, contract := range s.contracts {
		if contract.Type == contractType && contract.Enabled {
			// Check if already included
			found := false
			for _, c := range applicable {
				if c.ID == contract.ID {
					found = true
					break
				}
			}
			if !found {
				applicable = append(applicable, contract)
			}
		}
	}

	return applicable
}

func (s *SEMAP) evaluateContract(ctx context.Context, contract *Contract, input interface{}, output interface{}) (bool, error) {
	for _, condition := range contract.Conditions {
		evaluator := s.evaluators[condition.Evaluator]
		if evaluator == nil {
			return false, fmt.Errorf("unknown evaluator type: %s", condition.Evaluator)
		}

		// Determine what to evaluate
		var value interface{}
		switch contract.Type {
		case ContractTypePrecondition, ContractTypeGuardRail:
			value = input
		case ContractTypePostcondition:
			value = output
		case ContractTypeInvariant:
			value = input
		default:
			value = input
		}

		passed, err := evaluator.Evaluate(ctx, &condition, value)
		if err != nil {
			return false, err
		}

		// Apply negation if specified (for blocklist patterns)
		if condition.Negate {
			passed = !passed
		}

		if !passed {
			return false, nil
		}
	}

	return true, nil
}

func (s *SEMAP) checkProfileConstraints(ctx context.Context, profile *AgentProfile, action string, input interface{}) error {
	for _, constraint := range profile.Constraints {
		switch constraint.Type {
		case ConstraintTypeCommandWhitelist:
			if !s.isCommandAllowed(action, constraint.Allowed) {
				return fmt.Errorf("action '%s' is not in the allowed list", action)
			}
		case ConstraintTypeCommandBlacklist:
			if s.isCommandAllowed(action, constraint.Denied) {
				return fmt.Errorf("action '%s' is explicitly denied", action)
			}
		case ConstraintTypePathPattern:
			if resource, ok := input.(string); ok {
				if !s.matchesPathPattern(resource, constraint.Allowed) {
					return fmt.Errorf("resource '%s' does not match allowed paths", resource)
				}
			}
		}
	}

	// Check rate limits
	for _, rateLimit := range profile.RateLimits {
		if !s.checkRateLimit(profile.ID, rateLimit) {
			return fmt.Errorf("rate limit exceeded for resource '%s'", rateLimit.Resource)
		}
	}

	return nil
}

func (s *SEMAP) isCommandAllowed(action string, allowed []string) bool {
	for _, a := range allowed {
		if matched, _ := regexp.MatchString(a, action); matched { //nolint:errcheck
			return true
		}
	}
	return false
}

func (s *SEMAP) matchesPathPattern(resource string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, resource); matched { //nolint:errcheck
			return true
		}
	}
	return len(patterns) == 0 // Allow all if no patterns specified
}

func (s *SEMAP) checkRateLimit(agentID string, limit RateLimit) bool {
	// In a real implementation, would track rate limits per agent
	return true
}

func (s *SEMAP) formatViolationMessage(contract *Contract, input interface{}, output interface{}) string {
	for _, condition := range contract.Conditions {
		if condition.FailMessage != "" {
			return condition.FailMessage
		}
	}
	return fmt.Sprintf("Contract '%s' violated", contract.Name)
}

func (s *SEMAP) executeAction(ctx context.Context, action *ContractAction, violation *Violation) {
	handler := s.actions[action.Type]
	if handler != nil {
		_ = handler.Execute(ctx, action, violation) //nolint:errcheck
	}
}

func (s *SEMAP) attemptRemediation(ctx context.Context, violation *Violation) {
	contract := s.contracts[violation.ContractID]
	if contract == nil {
		return
	}

	for _, action := range contract.Actions {
		if action.Type == ActionTypeRemediate {
			handler := s.actions[ActionTypeRemediate]
			if handler != nil {
				err := handler.Execute(ctx, &action, violation)
				if err == nil {
					now := time.Now()
					violation.Remediation = &Remediation{
						Type:      ActionTypeRemediate,
						Status:    "applied",
						Details:   "Auto-remediation applied",
						AppliedAt: &now,
					}
					violation.ResolvedAt = &now
				}
			}
			break
		}
	}
}

func (s *SEMAP) generateContractID(contract *Contract) string {
	hash := sha256.Sum256([]byte(contract.Name + contract.Description))
	return hex.EncodeToString(hash[:8])
}

func (s *SEMAP) generatePolicyID(policy *Policy) string {
	hash := sha256.Sum256([]byte(policy.Name + policy.Description))
	return hex.EncodeToString(hash[:8])
}

func (s *SEMAP) generateViolationID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:8])
}

// RegisterEvaluator registers a custom condition evaluator.
func (s *SEMAP) RegisterEvaluator(condType ConditionType, evaluator ConditionEvaluator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evaluators[condType] = evaluator
}

// RegisterActionHandler registers a custom action handler.
func (s *SEMAP) RegisterActionHandler(actionType ActionType, handler ActionHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions[actionType] = handler
}

func (s *SEMAP) registerDefaultEvaluators() {
	s.evaluators[ConditionTypeRegex] = &RegexEvaluator{}
	s.evaluators[ConditionTypeRange] = &RangeEvaluator{}
	s.evaluators[ConditionTypeEnum] = &EnumEvaluator{}
	s.evaluators[ConditionTypeJSONSchema] = &JSONSchemaEvaluator{}
}

func (s *SEMAP) registerDefaultActions() {
	s.actions[ActionTypeLog] = &LogActionHandler{}
	s.actions[ActionTypeBlock] = &BlockActionHandler{}
	s.actions[ActionTypeAlert] = &AlertActionHandler{}
}

// GetViolations returns all violations.
func (s *SEMAP) GetViolations() []*Violation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Violation, len(s.violations))
	copy(result, s.violations)
	return result
}

// GetViolationsByAgent returns violations for a specific agent.
func (s *SEMAP) GetViolationsByAgent(agentID string) []*Violation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Violation
	for _, v := range s.violations {
		if v.Context.AgentID == agentID {
			result = append(result, v)
		}
	}
	return result
}

// GetContract returns a contract by ID.
func (s *SEMAP) GetContract(id string) *Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.contracts[id]
}

// GetPolicy returns a policy by ID.
func (s *SEMAP) GetPolicy(id string) *Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.policies[id]
}

// GetAgentProfile returns an agent profile by ID.
func (s *SEMAP) GetAgentProfile(id string) *AgentProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profiles[id]
}

// GetStatistics returns statistics about the SEMAP system.
func (s *SEMAP) GetStatistics() *SEMAPStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SEMAPStatistics{
		TotalContracts:       len(s.contracts),
		TotalPolicies:        len(s.policies),
		TotalProfiles:        len(s.profiles),
		TotalViolations:      len(s.violations),
		ViolationsBySeverity: make(map[ViolationSeverity]int),
		ViolationsByType:     make(map[ContractType]int),
	}

	for _, v := range s.violations {
		stats.ViolationsBySeverity[v.Severity]++
		stats.ViolationsByType[v.Type]++
		if v.ResolvedAt == nil {
			stats.UnresolvedViolations++
		}
	}

	return stats
}

// SEMAPStatistics represents statistics about the SEMAP system.
type SEMAPStatistics struct {
	TotalContracts       int                       `json:"total_contracts"`
	TotalPolicies        int                       `json:"total_policies"`
	TotalProfiles        int                       `json:"total_profiles"`
	TotalViolations      int                       `json:"total_violations"`
	UnresolvedViolations int                       `json:"unresolved_violations"`
	ViolationsBySeverity map[ViolationSeverity]int `json:"violations_by_severity"`
	ViolationsByType     map[ContractType]int      `json:"violations_by_type"`
}

// AuditLog logs audit entries.
type AuditLog struct {
	entries []AuditEntry
	mu      sync.RWMutex
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	Type       string    `json:"type"`
	AgentID    string    `json:"agent_id,omitempty"`
	ContractID string    `json:"contract_id,omitempty"`
	Action     string    `json:"action,omitempty"`
	Result     bool      `json:"result,omitempty"`
	Violations int       `json:"violations,omitempty"`
	Details    string    `json:"details,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewAuditLog creates a new audit log.
func NewAuditLog() *AuditLog {
	return &AuditLog{
		entries: make([]AuditEntry, 0),
	}
}

// Log adds an entry to the audit log.
func (l *AuditLog) Log(entry AuditEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

// GetEntries returns all audit log entries.
func (l *AuditLog) GetEntries() []AuditEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]AuditEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Default evaluators

// RegexEvaluator evaluates regex conditions.
type RegexEvaluator struct{}

// Evaluate evaluates a regex condition.
func (e *RegexEvaluator) Evaluate(ctx context.Context, condition *Condition, value interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("regex evaluator requires string input")
	}

	matched, err := regexp.MatchString(condition.Expression, str)
	if err != nil {
		return false, err
	}

	return matched, nil
}

// RangeEvaluator evaluates range conditions.
type RangeEvaluator struct{}

// Evaluate evaluates a range condition.
func (e *RangeEvaluator) Evaluate(ctx context.Context, condition *Condition, value interface{}) (bool, error) {
	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float64:
		num = v
	default:
		return false, fmt.Errorf("range evaluator requires numeric input")
	}

	minVal, hasMin := condition.Parameters["min"].(float64)
	maxVal, hasMax := condition.Parameters["max"].(float64)

	if hasMin && num < minVal {
		return false, nil
	}
	if hasMax && num > maxVal {
		return false, nil
	}

	return true, nil
}

// EnumEvaluator evaluates enum conditions.
type EnumEvaluator struct{}

// Evaluate evaluates an enum condition.
func (e *EnumEvaluator) Evaluate(ctx context.Context, condition *Condition, value interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("enum evaluator requires string input")
	}

	allowed, ok := condition.Parameters["values"].([]interface{})
	if !ok {
		return false, fmt.Errorf("enum evaluator requires 'values' parameter")
	}

	for _, v := range allowed {
		if vs, ok := v.(string); ok && strings.EqualFold(vs, str) {
			return true, nil
		}
	}

	return false, nil
}

// JSONSchemaEvaluator evaluates JSON schema conditions.
type JSONSchemaEvaluator struct{}

// Evaluate evaluates a JSON schema condition.
func (e *JSONSchemaEvaluator) Evaluate(ctx context.Context, condition *Condition, value interface{}) (bool, error) {
	// In a real implementation, would use a JSON Schema validation library
	return true, nil
}

// Default action handlers

// LogActionHandler logs violations.
type LogActionHandler struct{}

// Execute executes a log action.
func (h *LogActionHandler) Execute(ctx context.Context, action *ContractAction, violation *Violation) error {
	// Log the violation
	return nil
}

// BlockActionHandler blocks operations.
type BlockActionHandler struct{}

// Execute executes a block action.
func (h *BlockActionHandler) Execute(ctx context.Context, action *ContractAction, violation *Violation) error {
	// Block the operation - this is handled by returning violations
	return nil
}

// AlertActionHandler sends alerts.
type AlertActionHandler struct{}

// Execute executes an alert action.
func (h *AlertActionHandler) Execute(ctx context.Context, action *ContractAction, violation *Violation) error {
	// Send alert - would integrate with alerting system
	return nil
}

// Serialize serializes a contract to JSON.
func (c *Contract) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// DeserializeContract deserializes a contract from JSON.
func DeserializeContract(data []byte) (*Contract, error) {
	var contract Contract
	if err := json.Unmarshal(data, &contract); err != nil {
		return nil, err
	}
	return &contract, nil
}

// PredefinedContracts provides common contract templates.
var PredefinedContracts = map[string]*Contract{
	"no_sql_injection": {
		ID:          "no_sql_injection",
		Name:        "No SQL Injection",
		Description: "Prevents SQL injection attacks",
		Type:        ContractTypeGuardRail,
		Severity:    ViolationSeverityCritical,
		Conditions: []Condition{{
			ID:          "sql_injection_check",
			Expression:  `(?i)(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|--|;)`,
			Evaluator:   ConditionTypeRegex,
			Message:     "Check for SQL injection patterns",
			FailMessage: "Potential SQL injection detected",
			Negate:      true, // Match = violation (blocklist pattern)
		}},
		Actions: []ContractAction{{
			Type: ActionTypeBlock,
		}},
		Metadata: ContractMetadata{
			Category: "security",
			Priority: 1,
		},
		Enabled: true,
	},
	"no_path_traversal": {
		ID:          "no_path_traversal",
		Name:        "No Path Traversal",
		Description: "Prevents path traversal attacks",
		Type:        ContractTypeGuardRail,
		Severity:    ViolationSeverityCritical,
		Conditions: []Condition{{
			ID:          "path_traversal_check",
			Expression:  `\.\./|\.\.\\`,
			Evaluator:   ConditionTypeRegex,
			Message:     "Check for path traversal patterns",
			FailMessage: "Path traversal attempt detected",
			Negate:      true, // Match = violation (blocklist pattern)
		}},
		Actions: []ContractAction{{
			Type: ActionTypeBlock,
		}},
		Metadata: ContractMetadata{
			Category: "security",
			Priority: 1,
		},
		Enabled: true,
	},
	"max_response_length": {
		ID:          "max_response_length",
		Name:        "Maximum Response Length",
		Description: "Limits response length to prevent resource exhaustion",
		Type:        ContractTypePostcondition,
		Severity:    ViolationSeverityWarning,
		Conditions: []Condition{{
			ID:        "length_check",
			Evaluator: ConditionTypeRange,
			Parameters: map[string]interface{}{
				"max": float64(1000000),
			},
			Message:     "Check response length",
			FailMessage: "Response exceeds maximum length",
		}},
		Metadata: ContractMetadata{
			Category: "resource",
			Priority: 2,
		},
		Enabled: true,
	},
}

// CreateDefaultProfile creates a default agent profile.
func CreateDefaultProfile(agentID string, trustLevel TrustLevel) *AgentProfile {
	profile := &AgentProfile{
		ID:         agentID,
		Name:       agentID,
		TrustLevel: trustLevel,
		Capabilities: []AgentCapability{
			CapabilityRead,
		},
		CreatedAt: time.Now(),
	}

	switch trustLevel {
	case TrustLevelTrusted:
		profile.Capabilities = []AgentCapability{
			CapabilityRead,
			CapabilityWrite,
			CapabilityExecute,
			CapabilityNetwork,
			CapabilityFileSystem,
			CapabilityDatabase,
		}
	case TrustLevelHigh:
		profile.Capabilities = []AgentCapability{
			CapabilityRead,
			CapabilityWrite,
			CapabilityExecute,
			CapabilityFileSystem,
		}
	case TrustLevelMedium:
		profile.Capabilities = []AgentCapability{
			CapabilityRead,
			CapabilityWrite,
		}
	case TrustLevelLow:
		profile.Capabilities = []AgentCapability{
			CapabilityRead,
		}
		profile.RateLimits = []RateLimit{{
			Resource:   "api_call",
			Limit:      10,
			Window:     time.Minute,
			BurstLimit: 5,
		}}
	case TrustLevelUntrusted:
		profile.Capabilities = []AgentCapability{}
		profile.RateLimits = []RateLimit{{
			Resource:   "api_call",
			Limit:      1,
			Window:     time.Minute,
			BurstLimit: 1,
		}}
		profile.ResourceLimits = &ResourceLimits{
			MaxMemoryMB:      256,
			MaxCPUPercent:    10,
			MaxExecutionTime: 30 * time.Second,
		}
	}

	return profile
}
