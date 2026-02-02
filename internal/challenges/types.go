package challenges

import "time"

// ProviderInfo represents an LLM provider's metadata.
type ProviderInfo struct {
	Name        string
	DisplayName string
	Type        string // "api_key", "oauth", "free"
	APIKeyEnv   string
	Models      []string
	Score       float64
	Verified    bool
	VerifiedAt  time.Time
}

// ModelScore holds scoring details for a model.
type ModelScore struct {
	ModelID           string
	Provider          string
	ResponseSpeed     float64
	CostEffectiveness float64
	ModelEfficiency   float64
	Capability        float64
	Recency           float64
	TotalScore        float64
}

// DebateGroup represents a group of providers for debate.
type DebateGroup struct {
	Name     string
	Members  []DebateGroupMember
	Strategy string
}

// DebateGroupMember is a provider participating in debate.
type DebateGroupMember struct {
	Provider string
	Model    string
	Role     string
	Score    float64
}

// TestPrompt defines a prompt for API testing.
type TestPrompt struct {
	ID         string
	Category   string
	Prompt     string
	Expected   string
	MaxLatency time.Duration
	MinQuality float64
}

// APITestResult captures results of an API test.
type APITestResult struct {
	PromptID   string
	Response   string
	Latency    time.Duration
	StatusCode int
	Quality    float64
	Passed     bool
	Error      string
}
