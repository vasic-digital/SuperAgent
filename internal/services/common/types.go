package common

import (
	"time"
)

// Recovery and Resilience Types
type RecoveryStrategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Strategy    string                 `json:"strategy"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type RecoveryProcedure struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []RecoveryStep         `json:"steps"`
	Parameters  map[string]interface{} `json:"parameters"`
	RetryCount  int                    `json:"retry_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type RecoveryStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration          `json:"timeout"`
	Retryable   bool                   `json:"retryable"`
}

// Validation Types
type ValidationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Severity    string                 `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type IntegrityCheck struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Algorithm   string                 `json:"algorithm"`
	Parameters  map[string]interface{} `json:"parameters"`
	Schedule    string                 `json:"schedule"`
	LastRun     time.Time              `json:"last_run"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Authentication and Authorization Types
type AuthenticationManager struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	AuthMethods        map[string]interface{} `json:"auth_methods"`
	AuthProviders      map[string]interface{} `json:"auth_providers"`
	AuthValidators     map[string]interface{} `json:"auth_validators"`
	CredentialManagers map[string]interface{} `json:"credential_managers"`
	AuthPolicies       []interface{}          `json:"auth_policies"`
	SessionPolicies    []interface{}          `json:"session_policies"`
	MFAProviders       []interface{}          `json:"mfa_providers"`
	Configuration      map[string]interface{} `json:"configuration"`
	Status             string                 `json:"status"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type AuthenticationProvider struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Type              string                 `json:"type"`
	Providers         map[string]interface{} `json:"providers"`
	Protocols         map[string]interface{} `json:"protocols"`
	TokenManagers     map[string]interface{} `json:"token_managers"`
	IdentityProviders map[string]interface{} `json:"identity_providers"`
	AuthStrategies    []interface{}          `json:"auth_strategies"`
	ValidationMethods []interface{}          `json:"validation_methods"`
	Configuration     map[string]interface{} `json:"configuration"`
	Status            string                 `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type AccessController struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	AccessRules           []interface{}          `json:"access_rules"`
	PermissionMatrix      map[string]interface{} `json:"permission_matrix"`
	RoleHierarchy         map[string]interface{} `json:"role_hierarchy"`
	ResourceHierarchy     map[string]interface{} `json:"resource_hierarchy"`
	AccessPolicies        []interface{}          `json:"access_policies"`
	EnforcementStrategies []interface{}          `json:"enforcement_strategies"`
	Configuration         map[string]interface{} `json:"configuration"`
	Status                string                 `json:"status"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

type PermissionManager struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Permissions      map[string]interface{} `json:"permissions"`
	Roles            map[string]interface{} `json:"roles"`
	Policies         []interface{}          `json:"policies"`
	PermissionMatrix map[string]interface{} `json:"permission_matrix"`
	RoleMappings     map[string]interface{} `json:"role_mappings"`
	Configuration    map[string]interface{} `json:"configuration"`
	Status           string                 `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// Audit and Logging Types
type AuditTrail struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"`
	Actor     string                 `json:"actor"`
	Resource  string                 `json:"resource"`
	Outcome   string                 `json:"outcome"`
	Details   map[string]interface{} `json:"details"`
	Metadata  map[string]interface{} `json:"metadata"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	SessionID string                 `json:"session_id"`
	RequestID string                 `json:"request_id"`
	TraceID   string                 `json:"trace_id"`
}

// Date and Time Types
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Timezone  string    `json:"timezone"`
	Inclusive bool      `json:"inclusive"`
}

// Additional shared types
type ValidationResult struct {
	Valid     bool                   `json:"valid"`
	RuleID    string                 `json:"rule_id"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
	Severity  string                 `json:"severity"`
	Timestamp time.Time              `json:"timestamp"`
}

type RecoveryResult struct {
	Success     bool                   `json:"success"`
	ProcedureID string                 `json:"procedure_id"`
	StepID      string                 `json:"step_id"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
}

type AuthenticationResult struct {
	Success     bool                   `json:"success"`
	Token       string                 `json:"token,omitempty"`
	ExpiresAt   time.Time              `json:"expires_at,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
}
