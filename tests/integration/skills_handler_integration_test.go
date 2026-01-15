package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/handlers"
	"dev.helix.agent/internal/skills"
)

func setupSkillsService(t *testing.T) (*skills.Service, *skills.Integration) {
	config := skills.DefaultSkillConfig()
	config.MinConfidence = 0.5
	service := skills.NewService(config)

	// Start service in manual mode (without loading from disk)
	service.Start()

	// Register test skills
	dockerSkill := &skills.Skill{
		Name:           "docker-deploy",
		Description:    "Deploy applications with Docker",
		Category:       "devops",
		TriggerPhrases: []string{"deploy docker", "docker deployment", "container deploy"},
		Instructions:   "Use Docker commands to deploy containers...",
	}
	service.RegisterSkill(dockerSkill)

	kubernetesSkill := &skills.Skill{
		Name:           "kubernetes-config",
		Description:    "Configure Kubernetes clusters",
		Category:       "devops",
		TriggerPhrases: []string{"kubernetes", "k8s", "kubectl"},
		Instructions:   "Use kubectl commands to manage clusters...",
	}
	service.RegisterSkill(kubernetesSkill)

	pythonSkill := &skills.Skill{
		Name:           "python-debug",
		Description:    "Debug Python applications",
		Category:       "development",
		TriggerPhrases: []string{"python debug", "debug python", "pdb"},
		Instructions:   "Use Python debugger and profiling tools...",
	}
	service.RegisterSkill(pythonSkill)

	integration := skills.NewIntegration(service)
	return service, integration
}

func TestCompletionHandler_WithSkillsIntegration_Complete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, skillsIntegration := setupSkillsService(t)

	// Test that handler can be created with skills integration
	// Note: We can't fully test the completion flow without a proper RequestService
	// but we can verify the handler accepts skills integration
	handler := handlers.NewCompletionHandler(nil)
	handler.SetSkillsIntegration(skillsIntegration)

	// Verify the integration was set (handler won't panic on setting)
	assert.NotNil(t, handler)
}

func TestCompletionHandler_WithSkillsIntegration_Chat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, skillsIntegration := setupSkillsService(t)

	// Test using NewCompletionHandlerWithSkills constructor
	handler := handlers.NewCompletionHandlerWithSkills(nil, skillsIntegration)
	assert.NotNil(t, handler)
}

func TestSkillsIntegration_ProcessRequest(t *testing.T) {
	service, integration := setupSkillsService(t)
	require.NotNil(t, service)
	require.NotNil(t, integration)

	ctx := context.Background()

	// Test Docker skill matching
	reqCtx, err := integration.ProcessRequest(ctx, "test-req-1", "deploy docker container")
	require.NoError(t, err)
	require.NotNil(t, reqCtx)
	assert.Equal(t, "test-req-1", reqCtx.RequestID)
	assert.NotEmpty(t, reqCtx.MatchedSkills)
}

func TestSkillsIntegration_EnhancePrompt(t *testing.T) {
	_, integration := setupSkillsService(t)

	ctx := context.Background()

	// Get request context with matched skills
	reqCtx, err := integration.ProcessRequest(ctx, "test-req-2", "deploy docker")
	require.NoError(t, err)

	// Enhance prompt
	original := "Help me deploy my app"
	enhanced := integration.EnhancePromptWithSkills(original, reqCtx)

	// Should contain original prompt and skill information
	assert.Contains(t, enhanced, original)
	if len(reqCtx.Instructions) > 0 {
		assert.Contains(t, enhanced, "Active Skills")
	}
}

func TestSkillsIntegration_CompleteRequest(t *testing.T) {
	service, integration := setupSkillsService(t)
	require.NotNil(t, service)

	ctx := context.Background()

	// Process request
	_, err := integration.ProcessRequest(ctx, "test-req-3", "kubernetes deployment")
	require.NoError(t, err)

	// Complete request
	usages := integration.CompleteRequest("test-req-3", true, "")
	// May or may not have usages depending on matching
	_ = usages
}

func TestSkillsIntegration_BuildSkillsUsedSection(t *testing.T) {
	_, integration := setupSkillsService(t)

	usages := []skills.SkillUsage{
		{
			SkillName:    "docker-deploy",
			Category:     "devops",
			TriggerUsed:  "docker deployment",
			MatchType:    skills.MatchTypeExact,
			Confidence:   0.95,
			ToolsInvoked: []string{"Bash", "Read"},
			Success:      true,
		},
		{
			SkillName:   "kubernetes-config",
			Category:    "devops",
			TriggerUsed: "k8s",
			MatchType:   skills.MatchTypePartial,
			Confidence:  0.85,
			Success:     true,
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

func TestSkillsIntegration_EmptyUsages(t *testing.T) {
	_, integration := setupSkillsService(t)

	metadata := integration.BuildSkillsUsedSection([]skills.SkillUsage{})
	assert.Nil(t, metadata)
}

func TestResponseEnhancer(t *testing.T) {
	_, integration := setupSkillsService(t)
	enhancer := skills.NewResponseEnhancer(integration)

	response := map[string]interface{}{
		"content": "Test response content",
		"model":   "test-model",
	}

	usages := []skills.SkillUsage{
		{
			SkillName:  "docker-deploy",
			Category:   "devops",
			MatchType:  skills.MatchTypeExact,
			Confidence: 0.95,
			Success:    true,
		},
	}

	enhanced := enhancer.EnhanceResponse(response, usages)
	assert.Contains(t, enhanced, "skills_used")

	skillsUsed := enhanced["skills_used"].(*skills.SkillsUsedMetadata)
	assert.Equal(t, 1, skillsUsed.TotalSkills)
}

func TestDebateIntegration(t *testing.T) {
	service, integration := setupSkillsService(t)
	require.NotNil(t, service)

	debateInt := skills.NewDebateIntegration(integration)
	require.NotNil(t, debateInt)

	ctx := context.Background()
	reqCtx, err := debateInt.ProcessDebateRound(ctx, 1, 1, "Should AI be regulated?")
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestMCPIntegration(t *testing.T) {
	service, integration := setupSkillsService(t)
	require.NotNil(t, service)

	mcpInt := skills.NewMCPIntegration(integration)
	require.NotNil(t, mcpInt)

	ctx := context.Background()
	params := map[string]interface{}{
		"content": "deploy docker container",
	}
	reqCtx, err := mcpInt.ProcessMCPRequest(ctx, "mcp-req-1", "completion", params)
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestACPIntegration(t *testing.T) {
	service, integration := setupSkillsService(t)
	require.NotNil(t, service)

	acpInt := skills.NewACPIntegration(integration)
	require.NotNil(t, acpInt)

	ctx := context.Background()
	reqCtx, err := acpInt.ProcessACPMessage(ctx, "acp-req-1", "agent_message", "kubernetes help")
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestLSPIntegration(t *testing.T) {
	service, integration := setupSkillsService(t)
	require.NotNil(t, service)

	lspInt := skills.NewLSPIntegration(integration)
	require.NotNil(t, lspInt)

	ctx := context.Background()
	reqCtx, err := lspInt.ProcessLSPRequest(ctx, "lsp-req-1", "textDocument/completion", "file:///src/main.go")
	require.NoError(t, err)
	assert.NotNil(t, reqCtx)
}

func TestDebateHandler_WithSkillsIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, skillsIntegration := setupSkillsService(t)

	logger := logrus.New()
	handler := handlers.NewDebateHandler(nil, nil, logger)
	handler.SetSkillsIntegration(skillsIntegration)

	router := gin.New()
	handler.RegisterRoutes(router.Group("/v1"))

	// Create debate with topic that matches skills
	reqBody := map[string]interface{}{
		"topic": "How to deploy docker containers efficiently",
		"participants": []map[string]interface{}{
			{"name": "Proposer", "role": "proposer"},
			{"name": "Critic", "role": "critic"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/debates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "debate_id")
}

func TestSkillsServiceLifecycle(t *testing.T) {
	config := skills.DefaultSkillConfig()
	service := skills.NewService(config)

	// Service should not be running initially
	assert.False(t, service.IsRunning())

	// Start service in manual mode
	service.Start()
	assert.True(t, service.IsRunning())

	// Register skill while running
	skill := &skills.Skill{
		Name:        "test-skill",
		Description: "Test skill",
		Category:    "test",
	}
	service.RegisterSkill(skill)

	// Shutdown service
	err := service.Shutdown()
	require.NoError(t, err)
	assert.False(t, service.IsRunning())
}

func TestSkillMatching(t *testing.T) {
	service, _ := setupSkillsService(t)

	tests := []struct {
		name          string
		query         string
		expectedSkill string
		shouldMatch   bool
	}{
		{
			name:          "Exact Docker match",
			query:         "deploy docker",
			expectedSkill: "docker-deploy",
			shouldMatch:   true,
		},
		{
			name:          "Kubernetes match",
			query:         "kubernetes deployment",
			expectedSkill: "kubernetes-config",
			shouldMatch:   true,
		},
		{
			name:          "Python debug match",
			query:         "debug python code",
			expectedSkill: "python-debug",
			shouldMatch:   true,
		},
		{
			name:        "No match",
			query:       "random unrelated query",
			shouldMatch: false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := service.FindSkills(ctx, tt.query)
			require.NoError(t, err)

			if tt.shouldMatch {
				assert.NotEmpty(t, matches, "Expected to find matching skills for: %s", tt.query)
				if len(matches) > 0 {
					found := false
					for _, m := range matches {
						if m.Skill.Name == tt.expectedSkill {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected to find skill %s in matches", tt.expectedSkill)
				}
			}
		})
	}
}
