package clis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CLIAgentType represents the type of CLI agent
type CLIAgentType string

const (
	TypeAider       CLIAgentType = "aider"
	TypeClaudeCode  CLIAgentType = "claude_code"
	TypeCodex       CLIAgentType = "codex"
	TypeCline       CLIAgentType = "cline"
	TypeOpenHands   CLIAgentType = "openhands"
	TypeKiro        CLIAgentType = "kiro"
	TypeContinue    CLIAgentType = "continue"
	TypeGoose       CLIAgentType = "goose"
	TypeForge       CLIAgentType = "forge"
	TypePlandex     CLIAgentType = "plandex"
)

// InstanceStatus represents the status of a CLI agent instance
type InstanceStatus string

const (
	StatusIdle      InstanceStatus = "idle"
	StatusRunning   InstanceStatus = "running"
	StatusPaused    InstanceStatus = "paused"
	StatusError     InstanceStatus = "error"
	StatusStopped   InstanceStatus = "stopped"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskRunning    TaskStatus = "running"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
	TaskCancelled  TaskStatus = "cancelled"
	TaskRetrying   TaskStatus = "retrying"
)

// CLIAgentInstance represents a CLI agent instance
type CLIAgentInstance struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	Type           CLIAgentType    `json:"type" db:"type"`
	Status         InstanceStatus  `json:"status" db:"status"`
	Config         InstanceConfig  `json:"config" db:"config"`
	SessionID      *uuid.UUID      `json:"session_id,omitempty" db:"session_id"`
	UserID         *uuid.UUID      `json:"user_id,omitempty" db:"user_id"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	LastHeartbeat  *time.Time      `json:"last_heartbeat,omitempty" db:"last_heartbeat"`
	Metadata       json.RawMessage `json:"metadata,omitempty" db:"metadata"`
}

// InstanceConfig holds configuration for a CLI agent instance
type InstanceConfig struct {
	// Common settings
	WorkingDir     string            `json:"working_dir,omitempty"`
	Environment    map[string]string `json:"environment,omitempty"`
	Timeout        time.Duration     `json:"timeout,omitempty"`
	MaxRetries     int               `json:"max_retries,omitempty"`
	
	// Provider settings
	ProviderType   string            `json:"provider_type,omitempty"`
	Model          string            `json:"model,omitempty"`
	APIKey         string            `json:"-"` // Never serialize
	
	// Feature toggles
	EnableGit      bool              `json:"enable_git,omitempty"`
	EnableBrowser  bool              `json:"enable_browser,omitempty"`
	EnableSandbox  bool              `json:"enable_sandbox,omitempty"`
	AutoApprove    bool              `json:"auto_approve,omitempty"`
	
	// Type-specific settings
	AiderConfig    *AiderConfig      `json:"aider_config,omitempty"`
	ClaudeConfig   *ClaudeCodeConfig `json:"claude_config,omitempty"`
	CodexConfig    *CodexConfig      `json:"codex_config,omitempty"`
	ClineConfig    *ClineConfig      `json:"cline_config,omitempty"`
}

// AiderConfig holds Aider-specific configuration
type AiderConfig struct {
	MapTokens          int      `json:"map_tokens,omitempty"`
	EnableRepoMap      bool     `json:"enable_repo_map,omitempty"`
	EnableDiffEditing  bool     `json:"enable_diff_editing,omitempty"`
	Commit Attribution string   `json:"commit_attribution,omitempty"`
	IgnoredFiles       []string `json:"ignored_files,omitempty"`
}

// ClaudeCodeConfig holds Claude Code-specific configuration
type ClaudeCodeConfig struct {
	EnableToolUse      bool     `json:"enable_tool_use,omitempty"`
	EnableAutoMode     bool     `json:"enable_auto_mode,omitempty"`
	ApprovalRequired   []string `json:"approval_required,omitempty"`
	MaxToolCalls       int      `json:"max_tool_calls,omitempty"`
}

// CodexConfig holds Codex-specific configuration
type CodexConfig struct {
	EnableInterpreter  bool              `json:"enable_interpreter,omitempty"`
	EnableReasoning    bool              `json:"enable_reasoning,omitempty"`
	ReasoningEffort    string            `json:"reasoning_effort,omitempty"`
	SandboxConfig      map[string]string `json:"sandbox_config,omitempty"`
}

// ClineConfig holds Cline-specific configuration
type ClineConfig struct {
	EnableBrowser      bool     `json:"enable_browser,omitempty"`
	EnableComputerUse  bool     `json:"enable_computer_use,omitempty"`
	BrowserViewport    string   `json:"browser_viewport,omitempty"`
	MaxAutonomySteps   int      `json:"max_autonomy_steps,omitempty"`
}

// CLIAgentTask represents a task for a CLI agent
type CLIAgentTask struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	InstanceID     uuid.UUID       `json:"instance_id" db:"instance_id"`
	ParentTaskID   *uuid.UUID      `json:"parent_task_id,omitempty" db:"parent_task_id"`
	Type           string          `json:"type" db:"type"`
	Status         TaskStatus      `json:"status" db:"status"`
	Priority       int             `json:"priority" db:"priority"`
	Input          json.RawMessage `json:"input" db:"input"`
	Output         json.RawMessage `json:"output,omitempty" db:"output"`
	Error          *string         `json:"error,omitempty" db:"error"`
	StartedAt      *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	DurationMs     *int            `json:"duration_ms,omitempty" db:"duration_ms"`
	RetryCount     int             `json:"retry_count" db:"retry_count"`
	MaxRetries     int             `json:"max_retries" db:"max_retries"`
}

// TaskInput is a generic interface for task inputs
type TaskInput interface {
	Validate() error
}

// TaskOutput is a generic interface for task outputs
type TaskOutput interface {
	IsSuccess() bool
}

// RepoMapInput for repo map generation tasks
type RepoMapInput struct {
	Query          string   `json:"query"`
	MentionedFiles []string `json:"mentioned_files,omitempty"`
	MapTokens      int      `json:"map_tokens,omitempty"`
}

func (r RepoMapInput) Validate() error {
	if r.Query == "" {
		return fmt.Errorf("query is required")
	}
	return nil
}

// RepoMapOutput for repo map generation results
type RepoMapOutput struct {
	Symbols        []Symbol      `json:"symbols"`
	Files          []FileInfo    `json:"files"`
	TokenCount     int           `json:"token_count"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
}

func (r RepoMapOutput) IsSuccess() bool {
	return r.Success
}

// Symbol represents a code symbol
type Symbol struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"` // function, class, method, variable, etc.
	FilePath       string   `json:"file_path"`
	LineStart      int      `json:"line_start"`
	LineEnd        int      `json:"line_end"`
	Signature      string   `json:"signature,omitempty"`
	Documentation  string   `json:"documentation,omitempty"`
	RelevanceScore float64  `json:"relevance_score"`
}

// FileInfo represents file information
type FileInfo struct {
	Path           string `json:"path"`
	Language       string `json:"language"`
	Size           int64  `json:"size"`
	LastModified   int64  `json:"last_modified"`
}

// DiffApplyInput for diff application tasks
type DiffApplyInput struct {
	FilePath       string `json:"file_path"`
	SearchBlock    string `json:"search_block"`
	ReplaceBlock   string `json:"replace_block"`
	CreateIfMissing bool  `json:"create_if_missing,omitempty"`
}

func (d DiffApplyInput) Validate() error {
	if d.FilePath == "" {
		return fmt.Errorf("file_path is required")
	}
	if d.SearchBlock == "" && !d.CreateIfMissing {
		return fmt.Errorf("search_block is required unless create_if_missing is true")
	}
	return nil
}

// DiffApplyOutput for diff application results
type DiffApplyOutput struct {
	Success        bool   `json:"success"`
	FilePath       string `json:"file_path"`
	OriginalContent string `json:"original_content,omitempty"`
	NewContent     string `json:"new_content,omitempty"`
	DiffContent    string `json:"diff_content,omitempty"`
	Error          string `json:"error,omitempty"`
}

func (d DiffApplyOutput) IsSuccess() bool {
	return d.Success
}

// GitCommitInput for git commit tasks
type GitCommitInput struct {
	Message        string   `json:"message"`
	Files          []string `json:"files,omitempty"`
	Attribution    string   `json:"attribution,omitempty"`
	AllowEmpty     bool     `json:"allow_empty,omitempty"`
}

func (g GitCommitInput) Validate() error {
	if g.Message == "" {
		return fmt.Errorf("commit message is required")
	}
	return nil
}

// GitCommitOutput for git commit results
type GitCommitOutput struct {
	Success        bool     `json:"success"`
	CommitHash     string   `json:"commit_hash,omitempty"`
	FilesChanged   []string `json:"files_changed,omitempty"`
	Error          string   `json:"error,omitempty"`
}

func (g GitCommitOutput) IsSuccess() bool {
	return g.Success
}

// ToolUseInput for tool use tasks
type ToolUseInput struct {
	ToolName       string                 `json:"tool_name"`
	Arguments      map[string]interface{} `json:"arguments"`
	RequiresApproval bool                 `json:"requires_approval"`
}

func (t ToolUseInput) Validate() error {
	if t.ToolName == "" {
		return fmt.Errorf("tool_name is required")
	}
	return nil
}

// ToolUseOutput for tool use results
type ToolUseOutput struct {
	Success        bool                   `json:"success"`
	ToolName       string                 `json:"tool_name"`
	Result         map[string]interface{} `json:"result,omitempty"`
	Error          string                 `json:"error,omitempty"`
	DurationMs     int                    `json:"duration_ms"`
}

func (t ToolUseOutput) IsSuccess() bool {
	return t.Success
}

// InstanceManager manages CLI agent instances
type InstanceManager interface {
	Create(ctx context.Context, agentType CLIAgentType, config InstanceConfig) (*CLIAgentInstance, error)
	Get(ctx context.Context, id uuid.UUID) (*CLIAgentInstance, error)
	List(ctx context.Context, filter InstanceFilter) ([]*CLIAgentInstance, error)
	Update(ctx context.Context, id uuid.UUID, config InstanceConfig) (*CLIAgentInstance, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Start(ctx context.Context, id uuid.UUID) error
	Stop(ctx context.Context, id uuid.UUID) error
	Heartbeat(ctx context.Context, id uuid.UUID) error
}

// InstanceFilter for listing instances
type InstanceFilter struct {
	Type      *CLIAgentType
	Status    *InstanceStatus
	UserID    *uuid.UUID
	SessionID *uuid.UUID
}

// TaskManager manages CLI agent tasks
type TaskManager interface {
	Create(ctx context.Context, instanceID uuid.UUID, taskType string, input TaskInput, priority int) (*CLIAgentTask, error)
	Get(ctx context.Context, id uuid.UUID) (*CLIAgentTask, error)
	List(ctx context.Context, filter TaskFilter) ([]*CLIAgentTask, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status TaskStatus, output TaskOutput, err error) error
	Cancel(ctx context.Context, id uuid.UUID) error
	Retry(ctx context.Context, id uuid.UUID) (*CLIAgentTask, error)
}

// TaskFilter for listing tasks
type TaskFilter struct {
	InstanceID *uuid.UUID
	Status     *TaskStatus
	Type       *string
}

// CLIAgent defines the interface that all CLI agent implementations must satisfy
type CLIAgent interface {
	// Lifecycle
	Initialize(ctx context.Context, config InstanceConfig) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Destroy(ctx context.Context) error
	
	// Capabilities
	GetCapabilities() Capabilities
	
	// Task execution
	ExecuteTask(ctx context.Context, task *CLIAgentTask) (TaskOutput, error)
	
	// Health check
	HealthCheck(ctx context.Context) error
}

// Capabilities represents what a CLI agent can do
type Capabilities struct {
	SupportsRepoMap      bool     `json:"supports_repo_map"`
	SupportsGitOps       bool     `json:"supports_git_ops"`
	SupportsDiffEditing  bool     `json:"supports_diff_editing"`
	SupportsToolUse      bool     `json:"supports_tool_use"`
	SupportsBrowser      bool     `json:"supports_browser"`
	SupportsSandbox      bool     `json:"supports_sandbox"`
	SupportsInterpreter  bool     `json:"supports_interpreter"`
	SupportedTools       []string `json:"supported_tools,omitempty"`
}

// DefaultCapabilities returns default capabilities for each agent type
func DefaultCapabilities(agentType CLIAgentType) Capabilities {
	switch agentType {
	case TypeAider:
		return Capabilities{
			SupportsRepoMap:     true,
			SupportsGitOps:      true,
			SupportsDiffEditing: true,
			SupportsToolUse:     false,
			SupportsBrowser:     false,
			SupportsSandbox:     false,
			SupportsInterpreter: false,
		}
	case TypeClaudeCode:
		return Capabilities{
			SupportsRepoMap:     false,
			SupportsGitOps:      true,
			SupportsDiffEditing: false,
			SupportsToolUse:     true,
			SupportsBrowser:     false,
			SupportsSandbox:     false,
			SupportsInterpreter: false,
			SupportedTools:      []string{"Bash", "Read", "Write", "Edit", "Glob", "Grep"},
		}
	case TypeCodex:
		return Capabilities{
			SupportsRepoMap:     false,
			SupportsGitOps:      false,
			SupportsDiffEditing: false,
			SupportsToolUse:     false,
			SupportsBrowser:     false,
			SupportsSandbox:     true,
			SupportsInterpreter: true,
		}
	case TypeCline:
		return Capabilities{
			SupportsRepoMap:     false,
			SupportsGitOps:      false,
			SupportsDiffEditing: false,
			SupportsToolUse:     true,
			SupportsBrowser:     true,
			SupportsSandbox:     false,
			SupportsInterpreter: false,
			SupportedTools:      []string{"Browser", "ComputerUse", "Bash"},
		}
	case TypeOpenHands:
		return Capabilities{
			SupportsRepoMap:     false,
			SupportsGitOps:      true,
			SupportsDiffEditing: false,
			SupportsToolUse:     true,
			SupportsBrowser:     false,
			SupportsSandbox:     true,
			SupportsInterpreter: true,
		}
	default:
		return Capabilities{}
	}
}
