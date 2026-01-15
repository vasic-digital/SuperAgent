// Package skills provides a comprehensive skill management system for HelixAgent.
// Skills are structured instruction sets that teach the AI when and how to perform
// specific tasks, auto-activating based on conversation triggers.
package skills

import (
	"time"
)

// Skill represents a parsed SKILL.md file with all its metadata and content.
type Skill struct {
	// Metadata from YAML frontmatter
	Name          string   `json:"name" yaml:"name"`
	Description   string   `json:"description" yaml:"description"`
	AllowedTools  string   `json:"allowed_tools" yaml:"allowed-tools"`
	Version       string   `json:"version" yaml:"version"`
	License       string   `json:"license" yaml:"license"`
	Author        string   `json:"author" yaml:"author"`

	// Additional metadata
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	TriggerPhrases []string `json:"trigger_phrases"`

	// Content sections
	Overview      string            `json:"overview"`
	WhenToUse     string            `json:"when_to_use"`
	Instructions  string            `json:"instructions"`
	Examples      []SkillExample    `json:"examples"`
	Prerequisites []string          `json:"prerequisites"`
	Outputs       []string          `json:"outputs"`
	ErrorHandling []SkillError      `json:"error_handling"`
	Resources     []string          `json:"resources"`
	RelatedSkills []string          `json:"related_skills"`

	// Raw content
	RawContent    string            `json:"raw_content"`
	FilePath      string            `json:"file_path"`

	// Timestamps
	LoadedAt      time.Time         `json:"loaded_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// SkillExample represents a usage example within a skill.
type SkillExample struct {
	Title   string `json:"title"`
	Request string `json:"request"`
	Result  string `json:"result"`
}

// SkillError represents an error handling entry in a skill.
type SkillError struct {
	Error    string `json:"error"`
	Cause    string `json:"cause"`
	Solution string `json:"solution"`
}

// SkillCategory represents a category of skills.
type SkillCategory struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	Tags        []string `json:"tags"`
}

// SkillMatch represents a potential skill match from user input.
type SkillMatch struct {
	Skill       *Skill  `json:"skill"`
	Confidence  float64 `json:"confidence"`
	MatchedTrigger string `json:"matched_trigger"`
	MatchType   MatchType `json:"match_type"`
}

// MatchType indicates how a skill was matched.
type MatchType string

const (
	MatchTypeExact    MatchType = "exact"
	MatchTypePartial  MatchType = "partial"
	MatchTypeSemantic MatchType = "semantic"
	MatchTypeFuzzy    MatchType = "fuzzy"
)

// SkillUsage tracks the usage of a skill in a response.
type SkillUsage struct {
	SkillName    string    `json:"skill_name"`
	Category     string    `json:"category"`
	TriggerUsed  string    `json:"trigger_used"`
	MatchType    MatchType `json:"match_type"`
	Confidence   float64   `json:"confidence"`
	ToolsInvoked []string  `json:"tools_invoked"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
}

// SkillResponse contains the response from skill execution with usage tracking.
type SkillResponse struct {
	Content      string       `json:"content"`
	SkillsUsed   []SkillUsage `json:"skills_used"`
	TotalSkills  int          `json:"total_skills"`
	ProviderUsed string       `json:"provider_used,omitempty"`
	ModelUsed    string       `json:"model_used,omitempty"`
	Protocol     string       `json:"protocol,omitempty"`
}

// RegistryStats provides statistics about the skill registry.
type RegistryStats struct {
	TotalSkills      int            `json:"total_skills"`
	SkillsByCategory map[string]int `json:"skills_by_category"`
	TotalTriggers    int            `json:"total_triggers"`
	LoadedAt         time.Time      `json:"loaded_at"`
	LastUpdated      time.Time      `json:"last_updated"`
}

// SkillConfig configures the skill system.
type SkillConfig struct {
	// SkillsDirectory is the path to the skills directory
	SkillsDirectory string `json:"skills_directory"`

	// Categories to load (empty means all)
	Categories []string `json:"categories"`

	// EnableSemanticMatching uses LLM for semantic trigger matching
	EnableSemanticMatching bool `json:"enable_semantic_matching"`

	// MinConfidence is the minimum confidence for a skill match
	MinConfidence float64 `json:"min_confidence"`

	// MaxConcurrentSkills is the max skills that can execute in parallel
	MaxConcurrentSkills int `json:"max_concurrent_skills"`

	// TrackUsage enables skill usage tracking in responses
	TrackUsage bool `json:"track_usage"`

	// HotReload enables automatic skill reloading on file changes
	HotReload bool `json:"hot_reload"`

	// HotReloadInterval is the interval for checking skill file changes
	HotReloadInterval time.Duration `json:"hot_reload_interval"`
}

// DefaultSkillConfig returns default skill configuration.
func DefaultSkillConfig() *SkillConfig {
	return &SkillConfig{
		SkillsDirectory:        "skills",
		Categories:             nil, // Load all
		EnableSemanticMatching: true,
		MinConfidence:          0.7,
		MaxConcurrentSkills:    5,
		TrackUsage:             true,
		HotReload:              true,
		HotReloadInterval:      30 * time.Second,
	}
}

// AllowedTool represents a tool that a skill is allowed to use.
type AllowedTool struct {
	Name        string            `json:"name"`
	Constraints map[string]string `json:"constraints,omitempty"`
}

// ParseAllowedTools parses the allowed-tools string into structured tools.
func ParseAllowedTools(toolsStr string) []AllowedTool {
	// TODO: Implement parsing of tool strings like "Read, Write, Edit, Bash(cmd:*)"
	return nil
}
