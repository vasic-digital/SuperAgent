package skills

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIntegration(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	assert.NotNil(t, integration)
	assert.NotNil(t, integration.service)
	assert.NotNil(t, integration.log)
}

func TestIntegration_ProcessRequest(t *testing.T) {
	// Create service with test skills
	config := DefaultSkillConfig()
	config.MinConfidence = 0.5
	service := NewService(config)

	// Register test skill
	skill := &Skill{
		Name:           "docker-deploy",
		Description:    "Deploy applications with Docker",
		Category:       "devops",
		TriggerPhrases: []string{"deploy docker", "docker deployment"},
		Instructions:   "Use Docker commands to deploy...",
	}
	service.RegisterSkill(skill)

	// Mark service as running for test
	service.mu.Lock()
	service.running = true
	service.mu.Unlock()

	integration := NewIntegration(service)

	// Process request
	ctx := context.Background()
	reqCtx, err := integration.ProcessRequest(ctx, "test-req-1", "deploy docker container")
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
	assert.Equal(t, "test-req-1", reqCtx.RequestID)
}

func TestIntegration_CompleteRequest(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	// Register and start skill
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test skill",
		Category:    "test",
	}
	service.RegisterSkill(skill)

	// Create a match for the skill
	match := &SkillMatch{
		Skill:      skill,
		Confidence: 0.9,
		MatchType:  MatchTypeExact,
	}
	service.StartSkillExecution("req-123", skill, match)

	// Complete request
	usages := integration.CompleteRequest("req-123", true, "")
	assert.NotEmpty(t, usages)
	assert.True(t, usages[0].Success)
}

func TestIntegration_RecordToolUse(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	// Register and start skill
	skill := &Skill{
		Name:        "bash-skill",
		Description: "Bash operations",
		Category:    "devops",
	}
	service.RegisterSkill(skill)

	// Create a match for the skill
	match := &SkillMatch{
		Skill:      skill,
		Confidence: 0.9,
		MatchType:  MatchTypeExact,
	}
	service.StartSkillExecution("req-456", skill, match)

	// Record tool use
	integration.RecordToolUse("req-456", "Bash")
	integration.RecordToolUse("req-456", "Read")

	// Complete and verify
	usages := integration.CompleteRequest("req-456", true, "")
	require.NotEmpty(t, usages)
	assert.Contains(t, usages[0].ToolsInvoked, "Bash")
	assert.Contains(t, usages[0].ToolsInvoked, "Read")
}

func TestIntegration_BuildSkillsUsedSection(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	usages := []SkillUsage{
		{
			SkillName:    "docker-deploy",
			Category:     "devops",
			TriggerUsed:  "deploy docker",
			MatchType:    MatchTypeExact,
			Confidence:   0.95,
			ToolsInvoked: []string{"Bash", "Read"},
			Success:      true,
		},
		{
			SkillName:    "kubernetes-config",
			Category:     "devops",
			TriggerUsed:  "kubernetes",
			MatchType:    MatchTypePartial,
			Confidence:   0.85,
			ToolsInvoked: []string{"Write", "Bash"},
			Success:      true,
		},
	}

	metadata := integration.BuildSkillsUsedSection(usages)
	require.NotNil(t, metadata)
	assert.Equal(t, 2, metadata.TotalSkills)
	assert.Len(t, metadata.Skills, 2)

	// Verify first skill
	assert.Equal(t, "docker-deploy", metadata.Skills[0].Name)
	assert.Equal(t, "devops", metadata.Skills[0].Category)
	assert.Equal(t, 0.95, metadata.Skills[0].Confidence)
	assert.True(t, metadata.Skills[0].Success)
}

func TestIntegration_BuildSkillsUsedSection_Empty(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	metadata := integration.BuildSkillsUsedSection([]SkillUsage{})
	assert.Nil(t, metadata)
}

func TestIntegration_EnhancePromptWithSkills(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	reqCtx := &RequestContext{
		RequestID: "test-123",
		SkillsToApply: []*Skill{
			{
				Name:         "docker-deploy",
				Description:  "Deploy with Docker",
				Instructions: "1. Build image\n2. Push to registry\n3. Deploy",
			},
		},
		Instructions: []string{"1. Build image\n2. Push to registry\n3. Deploy"},
	}

	original := "Deploy my application"
	enhanced := integration.EnhancePromptWithSkills(original, reqCtx)

	assert.Contains(t, enhanced, original)
	assert.Contains(t, enhanced, "docker-deploy")
	assert.Contains(t, enhanced, "Deploy with Docker")
	assert.Contains(t, enhanced, "Build image")
}

func TestIntegration_EnhancePromptWithSkills_NoSkills(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	original := "Hello world"
	enhanced := integration.EnhancePromptWithSkills(original, nil)
	assert.Equal(t, original, enhanced)

	reqCtx := &RequestContext{
		Instructions: []string{},
	}
	enhanced = integration.EnhancePromptWithSkills(original, reqCtx)
	assert.Equal(t, original, enhanced)
}

func TestResponseEnhancer(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)
	enhancer := NewResponseEnhancer(integration)

	response := map[string]interface{}{
		"content": "Here is your deployment guide...",
		"model":   "claude-3-opus",
	}

	usages := []SkillUsage{
		{
			SkillName:  "docker-deploy",
			Category:   "devops",
			MatchType:  MatchTypeExact,
			Confidence: 0.95,
			Success:    true,
		},
	}

	enhanced := enhancer.EnhanceResponse(response, usages)
	assert.Contains(t, enhanced, "skills_used")

	skillsUsed := enhanced["skills_used"].(*SkillsUsedMetadata)
	assert.Equal(t, 1, skillsUsed.TotalSkills)
}

func TestDebateIntegration(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)

	// Mark service as running
	service.mu.Lock()
	service.running = true
	service.mu.Unlock()

	integration := NewIntegration(service)
	debateInt := NewDebateIntegration(integration)

	// Register skill
	skill := &Skill{
		Name:           "debate-analysis",
		Description:    "Analyze debate topics",
		Category:       "analysis",
		TriggerPhrases: []string{"analyze", "debate"},
	}
	service.RegisterSkill(skill)

	ctx := context.Background()
	reqCtx, err := debateInt.ProcessDebateRound(ctx, 1, 1, "Should AI be regulated?")
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestMCPIntegration(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)

	// Mark service as running
	service.mu.Lock()
	service.running = true
	service.mu.Unlock()

	integration := NewIntegration(service)
	mcpInt := NewMCPIntegration(integration)

	// Register skill
	skill := &Skill{
		Name:           "code-completion",
		Description:    "Code completion skill",
		Category:       "development",
		TriggerPhrases: []string{"completion", "complete code"},
	}
	service.RegisterSkill(skill)

	ctx := context.Background()
	params := map[string]interface{}{
		"content": "complete code function",
	}
	reqCtx, err := mcpInt.ProcessMCPRequest(ctx, "mcp-req-1", "completion", params)
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestACPIntegration(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)

	// Mark service as running
	service.mu.Lock()
	service.running = true
	service.mu.Unlock()

	integration := NewIntegration(service)
	acpInt := NewACPIntegration(integration)

	ctx := context.Background()
	reqCtx, err := acpInt.ProcessACPMessage(ctx, "acp-req-1", "agent_message", "Help with deployment")
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestLSPIntegration(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)

	// Mark service as running
	service.mu.Lock()
	service.running = true
	service.mu.Unlock()

	integration := NewIntegration(service)
	lspInt := NewLSPIntegration(integration)

	ctx := context.Background()
	reqCtx, err := lspInt.ProcessLSPRequest(ctx, "lsp-req-1", "textDocument/completion", "file:///src/main.go")
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestIntegration_GetService(t *testing.T) {
	config := DefaultSkillConfig()
	service := NewService(config)
	integration := NewIntegration(service)

	assert.Equal(t, service, integration.GetService())
}
