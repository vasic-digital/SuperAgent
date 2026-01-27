package cli

import (
	"time"

	"dev.helix.agent/internal/models"
)

// RenderStyle defines the output style for CLI rendering
type RenderStyle string

const (
	RenderStyleTheater    RenderStyle = "theater"    // Theatrical presentation with boxes
	RenderStyleNovel      RenderStyle = "novel"      // Novel-style prose narration
	RenderStyleScreenplay RenderStyle = "screenplay" // Screenplay/script format
	RenderStyleMinimal    RenderStyle = "minimal"    // Minimal formatting
	RenderStylePlain      RenderStyle = "plain"      // Plain text, no formatting
)

// CLIClient represents the type of CLI client
type CLIClient string

const (
	CLIClientOpenCode  CLIClient = "opencode"
	CLIClientCrush     CLIClient = "crush"
	CLIClientHelixCode CLIClient = "helixcode"
	CLIClientKiloCode  CLIClient = "kilocode"
	CLIClientUnknown   CLIClient = "unknown"
)

// ProgressBarStyle defines the style of progress bar
type ProgressBarStyle string

const (
	ProgressBarStyleASCII   ProgressBarStyle = "ascii"   // [=====     ]
	ProgressBarStyleUnicode ProgressBarStyle = "unicode" // [█████░░░░░]
	ProgressBarStyleBlock   ProgressBarStyle = "block"   // ▓▓▓▓▓░░░░░
	ProgressBarStyleDots    ProgressBarStyle = "dots"    // ●●●●●○○○○○
)

// ColorScheme defines color support levels
type ColorScheme string

const (
	ColorSchemeNone      ColorScheme = "none" // No colors
	ColorScheme8         ColorScheme = "8"    // 8 basic colors
	ColorScheme256       ColorScheme = "256"  // 256 colors
	ColorSchemeTrueColor ColorScheme = "true" // 24-bit true color
)

// RenderConfig holds configuration for CLI rendering
type RenderConfig struct {
	Style         RenderStyle      `yaml:"style"`
	ProgressStyle ProgressBarStyle `yaml:"progress_style"`
	ColorScheme   ColorScheme      `yaml:"color_scheme"`
	ShowResources bool             `yaml:"show_resources"`
	ShowLogs      bool             `yaml:"show_logs"`
	LogLines      int              `yaml:"log_lines"`
	Width         int              `yaml:"width"`
	Animate       bool             `yaml:"animate"`
	RefreshRate   time.Duration    `yaml:"refresh_rate"`
}

// DefaultRenderConfig returns sensible defaults
func DefaultRenderConfig() *RenderConfig {
	return &RenderConfig{
		Style:         RenderStyleTheater,
		ProgressStyle: ProgressBarStyleUnicode,
		ColorScheme:   ColorScheme256,
		ShowResources: true,
		ShowLogs:      true,
		LogLines:      5,
		Width:         80,
		Animate:       true,
		RefreshRate:   100 * time.Millisecond,
	}
}

// ProgressBarContent represents progress bar data
type ProgressBarContent struct {
	TaskID      string         `json:"task_id"`
	TaskName    string         `json:"task_name"`
	TaskType    string         `json:"task_type"`
	Progress    float64        `json:"progress"` // 0-100
	Message     string         `json:"message"`
	Status      string         `json:"status"`
	StartedAt   time.Time      `json:"started_at"`
	ETA         *time.Duration `json:"eta,omitempty"`
	CurrentStep int            `json:"current_step,omitempty"`
	TotalSteps  int            `json:"total_steps,omitempty"`
}

// StatusTableContent represents status table data
type StatusTableContent struct {
	Tasks      []TaskStatusRow `json:"tasks"`
	TotalCount int             `json:"total_count"`
	Timestamp  time.Time       `json:"timestamp"`
}

// TaskStatusRow represents a row in the status table
type TaskStatusRow struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Status   string        `json:"status"`
	Progress float64       `json:"progress"`
	Duration time.Duration `json:"duration"`
	WorkerID string        `json:"worker_id,omitempty"`
	Message  string        `json:"message,omitempty"`
}

// ResourceGaugeContent represents resource usage gauges
type ResourceGaugeContent struct {
	TaskID        string  `json:"task_id"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryBytes   int64   `json:"memory_bytes"`
	MemoryMax     int64   `json:"memory_max,omitempty"`
	IOReadBytes   int64   `json:"io_read_bytes"`
	IOWriteBytes  int64   `json:"io_write_bytes"`
	NetBytesSent  int64   `json:"net_bytes_sent"`
	NetBytesRecv  int64   `json:"net_bytes_recv"`
	OpenFDs       int     `json:"open_fds"`
	ThreadCount   int     `json:"thread_count"`
}

// LogLineContent represents a log line
type LogLineContent struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// TaskPanelContent represents a complete task panel
type TaskPanelContent struct {
	Task      *models.BackgroundTask `json:"task"`
	Progress  *ProgressBarContent    `json:"progress"`
	Resources *ResourceGaugeContent  `json:"resources,omitempty"`
	Logs      []LogLineContent       `json:"logs,omitempty"`
	ETA       *time.Duration         `json:"eta,omitempty"`
}

// DashboardContent represents the full dashboard
type DashboardContent struct {
	Title       string             `json:"title"`
	Timestamp   time.Time          `json:"timestamp"`
	WorkerStats WorkerStatsContent `json:"worker_stats"`
	QueueStats  QueueStatsContent  `json:"queue_stats"`
	Tasks       []TaskPanelContent `json:"tasks"`
	SystemInfo  *SystemInfoContent `json:"system_info,omitempty"`
}

// WorkerStatsContent represents worker pool statistics
type WorkerStatsContent struct {
	ActiveWorkers  int   `json:"active_workers"`
	IdleWorkers    int   `json:"idle_workers"`
	TotalWorkers   int   `json:"total_workers"`
	TasksCompleted int64 `json:"tasks_completed"`
	TasksFailed    int64 `json:"tasks_failed"`
}

// QueueStatsContent represents queue statistics
type QueueStatsContent struct {
	PendingCount int64            `json:"pending_count"`
	RunningCount int64            `json:"running_count"`
	ByPriority   map[string]int64 `json:"by_priority,omitempty"`
	ByStatus     map[string]int64 `json:"by_status,omitempty"`
}

// SystemInfoContent represents system resource info
type SystemInfoContent struct {
	CPUCores          int     `json:"cpu_cores"`
	CPUUsedPercent    float64 `json:"cpu_used_percent"`
	MemoryTotalMB     int64   `json:"memory_total_mb"`
	MemoryUsedMB      int64   `json:"memory_used_mb"`
	MemoryUsedPercent float64 `json:"memory_used_percent"`
	LoadAvg1          float64 `json:"load_avg_1"`
	LoadAvg5          float64 `json:"load_avg_5"`
	LoadAvg15         float64 `json:"load_avg_15"`
}

// ANSIColor constants
const (
	ColorReset     = "\033[0m"
	ColorBold      = "\033[1m"
	ColorDim       = "\033[2m"
	ColorItalic    = "\033[3m"
	ColorUnderline = "\033[4m"

	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"

	ColorBgBlack   = "\033[40m"
	ColorBgRed     = "\033[41m"
	ColorBgGreen   = "\033[42m"
	ColorBgYellow  = "\033[43m"
	ColorBgBlue    = "\033[44m"
	ColorBgMagenta = "\033[45m"
	ColorBgCyan    = "\033[46m"
	ColorBgWhite   = "\033[47m"

	// Bright colors
	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"
)

// Unicode box drawing characters
const (
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	BoxTopLeft     = "┌"
	BoxTopRight    = "┐"
	BoxBottomLeft  = "└"
	BoxBottomRight = "┘"
	BoxTeeLeft     = "├"
	BoxTeeRight    = "┤"
	BoxTeeTop      = "┬"
	BoxTeeBottom   = "┴"
	BoxCross       = "┼"

	BoxDoubleHorizontal  = "═"
	BoxDoubleVertical    = "║"
	BoxDoubleTopLeft     = "╔"
	BoxDoubleTopRight    = "╗"
	BoxDoubleBottomLeft  = "╚"
	BoxDoubleBottomRight = "╝"
	BoxDoubleTeeLeft     = "╠"
	BoxDoubleTeeRight    = "╣"
)

// Progress bar characters
const (
	ProgressFilled      = "█"
	ProgressEmpty       = "░"
	ProgressPartial1    = "▏"
	ProgressPartial2    = "▎"
	ProgressPartial3    = "▍"
	ProgressPartial4    = "▌"
	ProgressPartial5    = "▋"
	ProgressPartial6    = "▊"
	ProgressPartial7    = "▉"
	ProgressASCIIFilled = "="
	ProgressASCIIEmpty  = " "
	ProgressASCIITip    = ">"
)

// Status icons
const (
	IconPending   = "○"
	IconRunning   = "◉"
	IconCompleted = "✓"
	IconFailed    = "✗"
	IconStuck     = "⚠"
	IconCancelled = "⊘"
	IconPaused    = "⏸"
	IconSpinner1  = "⠋"
	IconSpinner2  = "⠙"
	IconSpinner3  = "⠹"
	IconSpinner4  = "⠸"
	IconSpinner5  = "⠼"
	IconSpinner6  = "⠴"
	IconSpinner7  = "⠦"
	IconSpinner8  = "⠧"
	IconSpinner9  = "⠇"
	IconSpinner10 = "⠏"
)

// Fallback status icons - visual indicators for LLM fallback events
const (
	IconFallbackTriggered = "⚡" // Fallback triggered (lightning)
	IconFallbackSuccess   = "🔄" // Fallback succeeded (cycle arrows)
	IconFallbackFailed    = "⛔" // Fallback failed (no entry)
	IconFallbackExhausted = "💀" // All fallbacks exhausted (skull)
	IconFallbackChain     = "🔗" // Complete chain (chain link)

	// Error category icons
	IconErrorRateLimit   = "🚦"  // Rate limit (traffic light)
	IconErrorTimeout     = "⏱️" // Timeout (timer)
	IconErrorAuth        = "🔑"  // Auth error (key)
	IconErrorQuota       = "📊"  // Quota exceeded (chart)
	IconErrorConnection  = "🔌"  // Connection error (plug)
	IconErrorUnavailable = "🚫"  // Service unavailable (no entry)
	IconErrorOverloaded  = "🔥"  // Service overloaded (fire)
	IconErrorInvalid     = "⚠️" // Invalid request (warning)
	IconErrorEmpty       = "📭"  // Empty response (empty mailbox)
	IconErrorUnknown     = "❓"  // Unknown error (question mark)
)

// FallbackIndicatorContent represents fallback event visual data for CLI rendering
type FallbackIndicatorContent struct {
	// Event type and identity
	EventType     string `json:"event_type"`
	DebateID      string `json:"debate_id"`
	Position      int    `json:"position"`
	Role          string `json:"role"`
	AttemptNumber int    `json:"attempt_number"`
	TotalAttempts int    `json:"total_attempts"`

	// Provider info
	PrimaryProvider  string `json:"primary_provider"`
	PrimaryModel     string `json:"primary_model"`
	FallbackProvider string `json:"fallback_provider,omitempty"`
	FallbackModel    string `json:"fallback_model,omitempty"`

	// Error info
	ErrorCode     string `json:"error_code"`
	ErrorMessage  string `json:"error_message"`
	ErrorCategory string `json:"error_category"`

	// Visual elements
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	Animation string `json:"animation,omitempty"`

	// Timing
	Duration  int64 `json:"duration_ms"`
	Timestamp int64 `json:"timestamp"`
}

// GetFallbackEventIcon returns the icon for a fallback event type
func GetFallbackEventIcon(eventType string) string {
	switch eventType {
	case "fallback.triggered":
		return IconFallbackTriggered
	case "fallback.success":
		return IconFallbackSuccess
	case "fallback.failed":
		return IconFallbackFailed
	case "fallback.exhausted":
		return IconFallbackExhausted
	case "fallback.chain":
		return IconFallbackChain
	default:
		return IconStuck
	}
}

// GetFallbackEventColor returns ANSI color for a fallback event type
func GetFallbackEventColor(eventType string) string {
	switch eventType {
	case "fallback.triggered":
		return ColorBrightYellow // Yellow for warning/attention
	case "fallback.success":
		return ColorBrightGreen // Green for success
	case "fallback.failed":
		return ColorBrightRed // Red for failure
	case "fallback.exhausted":
		return ColorRed + ColorBold // Bold red for critical
	case "fallback.chain":
		return ColorBrightCyan // Cyan for info
	default:
		return ColorYellow
	}
}

// GetErrorCategoryIcon returns the icon for an error category
func GetErrorCategoryIcon(category string) string {
	switch category {
	case "rate_limit":
		return IconErrorRateLimit
	case "timeout":
		return IconErrorTimeout
	case "auth":
		return IconErrorAuth
	case "quota":
		return IconErrorQuota
	case "connection":
		return IconErrorConnection
	case "unavailable":
		return IconErrorUnavailable
	case "overloaded":
		return IconErrorOverloaded
	case "invalid_request":
		return IconErrorInvalid
	case "empty_response":
		return IconErrorEmpty
	default:
		return IconErrorUnknown
	}
}

// FormatFallbackIndicator formats a fallback indicator for CLI display
func FormatFallbackIndicator(content *FallbackIndicatorContent) string {
	icon := GetFallbackEventIcon(content.EventType)
	color := GetFallbackEventColor(content.EventType)
	errIcon := GetErrorCategoryIcon(content.ErrorCategory)

	// Build the formatted string
	result := color + icon + " [" + content.Role + "] "

	switch content.EventType {
	case "fallback.triggered":
		result += "Fallback triggered: " + content.PrimaryProvider + "/" + content.PrimaryModel + " failed"
		if content.ErrorMessage != "" {
			result += "\n   " + errIcon + " " + content.ErrorMessage
		}
		if content.FallbackProvider != "" {
			result += "\n   → Trying: " + content.FallbackProvider + "/" + content.FallbackModel
		}
	case "fallback.success":
		result += "Fallback succeeded: " + content.FallbackProvider + "/" + content.FallbackModel
		result += " (attempt " + string(rune('0'+content.AttemptNumber)) + ")"
	case "fallback.failed":
		result += "Fallback failed: " + content.FallbackProvider + "/" + content.FallbackModel
		if content.ErrorMessage != "" {
			result += "\n   " + errIcon + " " + content.ErrorMessage
		}
	case "fallback.exhausted":
		result += "ALL FALLBACKS EXHAUSTED - No response available"
	case "fallback.chain":
		result += "Fallback chain summary: " + string(rune('0'+content.AttemptNumber)) + " attempts"
	}

	result += ColorReset
	return result
}

// SpinnerFrames for animated spinners
var SpinnerFrames = []string{
	IconSpinner1, IconSpinner2, IconSpinner3, IconSpinner4, IconSpinner5,
	IconSpinner6, IconSpinner7, IconSpinner8, IconSpinner9, IconSpinner10,
}

// GetStatusIcon returns the appropriate icon for a task status
func GetStatusIcon(status models.TaskStatus) string {
	switch status {
	case models.TaskStatusPending, models.TaskStatusQueued:
		return IconPending
	case models.TaskStatusRunning:
		return IconRunning
	case models.TaskStatusCompleted:
		return IconCompleted
	case models.TaskStatusFailed:
		return IconFailed
	case models.TaskStatusStuck:
		return IconStuck
	case models.TaskStatusCancelled:
		return IconCancelled
	case models.TaskStatusPaused:
		return IconPaused
	default:
		return IconPending
	}
}

// GetStatusColor returns the appropriate color for a task status
func GetStatusColor(status models.TaskStatus) string {
	switch status {
	case models.TaskStatusPending, models.TaskStatusQueued:
		return ColorYellow
	case models.TaskStatusRunning:
		return ColorCyan
	case models.TaskStatusCompleted:
		return ColorGreen
	case models.TaskStatusFailed:
		return ColorRed
	case models.TaskStatusStuck:
		return ColorBrightYellow
	case models.TaskStatusCancelled:
		return ColorMagenta
	case models.TaskStatusPaused:
		return ColorBlue
	default:
		return ColorWhite
	}
}
