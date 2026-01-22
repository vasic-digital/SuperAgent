// Package skills provides integration between the Skills system and HelixAgent handlers.
package skills

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

// Integration provides integration between Skills and the rest of HelixAgent.
type Integration struct {
	service *Service
	log     *logrus.Logger
	mu      sync.RWMutex
}

// NewIntegration creates a new Skills integration.
func NewIntegration(service *Service) *Integration {
	return &Integration{
		service: service,
		log:     logrus.New(),
	}
}

// SetLogger sets the logger.
func (i *Integration) SetLogger(log *logrus.Logger) {
	i.log = log
}

// ProcessRequest processes a request and finds matching skills.
// Returns the matched skills and their instructions to be included in the prompt.
func (i *Integration) ProcessRequest(ctx context.Context, requestID, userInput string) (*RequestContext, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Find matching skills
	matches, err := i.service.FindSkills(ctx, userInput)
	if err != nil {
		return nil, err
	}

	reqCtx := &RequestContext{
		RequestID:     requestID,
		MatchedSkills: matches,
		SkillsToApply: make([]*Skill, 0),
		Instructions:  make([]string, 0),
		ActiveUsages:  make([]*SkillUsage, 0),
	}

	// Filter by confidence threshold
	minConfidence := i.service.GetConfig().MinConfidence
	for _, match := range matches {
		if match.Confidence >= minConfidence {
			reqCtx.SkillsToApply = append(reqCtx.SkillsToApply, match.Skill)
			if match.Skill.Instructions != "" {
				reqCtx.Instructions = append(reqCtx.Instructions, match.Skill.Instructions)
			}
			// Start tracking this skill usage
			usage := i.service.StartSkillExecution(requestID, match.Skill, &match)
			reqCtx.ActiveUsages = append(reqCtx.ActiveUsages, usage)
		}
	}

	i.log.WithFields(logrus.Fields{
		"request_id":     requestID,
		"matched_skills": len(matches),
		"applied_skills": len(reqCtx.SkillsToApply),
	}).Debug("Processed request for skills")

	return reqCtx, nil
}

// CompleteRequest marks all skill executions for a request as complete.
func (i *Integration) CompleteRequest(requestID string, success bool, errorMsg string) []SkillUsage {
	completedUsages := make([]SkillUsage, 0)

	usage := i.service.CompleteSkillExecution(requestID, success, errorMsg)
	if usage != nil {
		completedUsages = append(completedUsages, *usage)
	}

	return completedUsages
}

// RecordToolUse records that a tool was used within a skill execution.
func (i *Integration) RecordToolUse(requestID, toolName string) {
	i.service.RecordToolUse(requestID, toolName)
}

// BuildSkillsUsedSection builds the "Skills Used" section for response metadata.
func (i *Integration) BuildSkillsUsedSection(usages []SkillUsage) *SkillsUsedMetadata {
	if len(usages) == 0 {
		return nil
	}

	metadata := &SkillsUsedMetadata{
		TotalSkills: len(usages),
		Skills:      make([]SkillUsedInfo, 0, len(usages)),
	}

	for _, usage := range usages {
		info := SkillUsedInfo{
			Name:       usage.SkillName,
			Category:   usage.Category,
			Trigger:    usage.TriggerUsed,
			MatchType:  string(usage.MatchType),
			Confidence: usage.Confidence,
			ToolsUsed:  usage.ToolsInvoked,
			Success:    usage.Success,
		}
		if usage.Error != "" {
			info.Error = usage.Error
		}
		metadata.Skills = append(metadata.Skills, info)
	}

	return metadata
}

// EnhancePromptWithSkills enhances a prompt with matched skill instructions.
func (i *Integration) EnhancePromptWithSkills(originalPrompt string, reqCtx *RequestContext) string {
	if reqCtx == nil || len(reqCtx.Instructions) == 0 {
		return originalPrompt
	}

	// Build enhanced prompt with skill instructions
	enhanced := originalPrompt + "\n\n---\n"
	enhanced += "## Active Skills\n\n"

	for idx, skill := range reqCtx.SkillsToApply {
		enhanced += "### " + skill.Name + "\n"
		if skill.Description != "" {
			enhanced += skill.Description + "\n\n"
		}
		if idx < len(reqCtx.Instructions) && reqCtx.Instructions[idx] != "" {
			enhanced += "**Instructions:**\n" + reqCtx.Instructions[idx] + "\n\n"
		}
	}

	return enhanced
}

// GetService returns the underlying Skills service.
func (i *Integration) GetService() *Service {
	return i.service
}

// RequestContext holds context for a request with skill matching.
type RequestContext struct {
	RequestID     string
	MatchedSkills []SkillMatch
	SkillsToApply []*Skill
	Instructions  []string
	ActiveUsages  []*SkillUsage
}

// SkillsUsedMetadata represents metadata about skills used in a response.
type SkillsUsedMetadata struct {
	TotalSkills int             `json:"total_skills"`
	Skills      []SkillUsedInfo `json:"skills"`
}

// SkillUsedInfo represents information about a single skill used.
type SkillUsedInfo struct {
	Name       string   `json:"name"`
	Category   string   `json:"category"`
	Trigger    string   `json:"trigger,omitempty"`
	MatchType  string   `json:"match_type"`
	Confidence float64  `json:"confidence"`
	ToolsUsed  []string `json:"tools_used,omitempty"`
	Success    bool     `json:"success"`
	Error      string   `json:"error,omitempty"`
}

// ResponseEnhancer enhances LLM responses with skill usage information.
type ResponseEnhancer struct {
	integration *Integration
}

// NewResponseEnhancer creates a new response enhancer.
func NewResponseEnhancer(integration *Integration) *ResponseEnhancer {
	return &ResponseEnhancer{
		integration: integration,
	}
}

// EnhanceResponse adds skill usage metadata to a response.
func (r *ResponseEnhancer) EnhanceResponse(response map[string]interface{}, usages []SkillUsage) map[string]interface{} {
	if len(usages) == 0 {
		return response
	}

	// Add skills_used metadata
	skillsMetadata := r.integration.BuildSkillsUsedSection(usages)
	if skillsMetadata != nil {
		response["skills_used"] = skillsMetadata
	}

	return response
}

// DebateIntegration provides specific integration for the AI Debate system.
type DebateIntegration struct {
	integration *Integration
}

// NewDebateIntegration creates a new debate integration.
func NewDebateIntegration(integration *Integration) *DebateIntegration {
	return &DebateIntegration{
		integration: integration,
	}
}

// ProcessDebateRound processes a debate round with skill matching.
func (d *DebateIntegration) ProcessDebateRound(ctx context.Context, debateID, roundNum int, topic string) (*RequestContext, error) {
	requestID := generateDebateRequestID(debateID, roundNum)
	return d.integration.ProcessRequest(ctx, requestID, topic)
}

// CompleteDebateRound marks a debate round's skill executions as complete.
func (d *DebateIntegration) CompleteDebateRound(debateID, roundNum int, success bool) []SkillUsage {
	requestID := generateDebateRequestID(debateID, roundNum)
	return d.integration.CompleteRequest(requestID, success, "")
}

// generateDebateRequestID generates a unique request ID for a debate round.
func generateDebateRequestID(debateID, roundNum int) string {
	return "debate-" + string(rune(debateID)) + "-round-" + string(rune(roundNum))
}

// MCPIntegration provides specific integration for MCP protocol.
type MCPIntegration struct {
	integration *Integration
}

// NewMCPIntegration creates a new MCP integration.
func NewMCPIntegration(integration *Integration) *MCPIntegration {
	return &MCPIntegration{
		integration: integration,
	}
}

// ProcessMCPRequest processes an MCP request with skill matching.
func (m *MCPIntegration) ProcessMCPRequest(ctx context.Context, requestID string, method string, params map[string]interface{}) (*RequestContext, error) {
	// Build input from MCP method and params
	input := method
	if content, ok := params["content"].(string); ok {
		input += " " + content
	}
	if query, ok := params["query"].(string); ok {
		input += " " + query
	}

	return m.integration.ProcessRequest(ctx, requestID, input)
}

// ACPIntegration provides specific integration for ACP protocol.
type ACPIntegration struct {
	integration *Integration
}

// NewACPIntegration creates a new ACP integration.
func NewACPIntegration(integration *Integration) *ACPIntegration {
	return &ACPIntegration{
		integration: integration,
	}
}

// ProcessACPMessage processes an ACP message with skill matching.
func (a *ACPIntegration) ProcessACPMessage(ctx context.Context, requestID string, messageType string, content string) (*RequestContext, error) {
	input := messageType + " " + content
	return a.integration.ProcessRequest(ctx, requestID, input)
}

// LSPIntegration provides specific integration for LSP protocol.
type LSPIntegration struct {
	integration *Integration
}

// NewLSPIntegration creates a new LSP integration.
func NewLSPIntegration(integration *Integration) *LSPIntegration {
	return &LSPIntegration{
		integration: integration,
	}
}

// ProcessLSPRequest processes an LSP request with skill matching.
func (l *LSPIntegration) ProcessLSPRequest(ctx context.Context, requestID string, method string, documentURI string) (*RequestContext, error) {
	input := method + " " + documentURI
	return l.integration.ProcessRequest(ctx, requestID, input)
}
