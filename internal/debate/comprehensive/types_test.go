package comprehensive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{"architect", RoleArchitect, true},
		{"generator", RoleGenerator, true},
		{"critic", RoleCritic, true},
		{"invalid", Role("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.role.IsValid())
		})
	}
}

func TestAllRoles(t *testing.T) {
	roles := AllRoles()
	assert.Len(t, roles, 11)

	// Check specific roles exist
	found := make(map[Role]bool)
	for _, r := range roles {
		found[r] = true
	}

	assert.True(t, found[RoleArchitect])
	assert.True(t, found[RoleGenerator])
	assert.True(t, found[RoleCritic])
	assert.True(t, found[RoleRedTeam])
}

func TestNewAgent(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)

	assert.NotEmpty(t, agent.ID)
	assert.NotEmpty(t, agent.Name)
	assert.Equal(t, RoleGenerator, agent.Role)
	assert.Equal(t, "openai", agent.Provider)
	assert.Equal(t, "gpt-4", agent.Model)
	assert.Equal(t, 8.5, agent.Score)
	assert.True(t, agent.IsActive)
	assert.NotEmpty(t, agent.Capabilities)
}

func TestAgent_HasCapability(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)

	// Generator should have code generation capability
	assert.True(t, agent.HasCapability(CapabilityCodeGeneration))

	// Should not have security capability
	assert.False(t, agent.HasCapability(CapabilitySecurityAnalysis))
}

func TestDefaultCapabilitiesForRole(t *testing.T) {
	tests := []struct {
		role          Role
		shouldHave    Capability
		shouldNotHave Capability
	}{
		{RoleArchitect, CapabilityArchitecture, CapabilityCodeGeneration},
		{RoleGenerator, CapabilityCodeGeneration, CapabilitySecurityAnalysis},
		{RoleCritic, CapabilityCodeReview, CapabilityCodeGeneration},
		{RoleSecurity, CapabilitySecurityAnalysis, CapabilityCodeGeneration},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			caps := DefaultCapabilitiesForRole(tt.role)

			// Check it has expected capability
			hasExpected := false
			for _, c := range caps {
				if c == tt.shouldHave {
					hasExpected = true
					break
				}
			}
			assert.True(t, hasExpected, "Should have capability %s", tt.shouldHave)
		})
	}
}

func TestNewAgentResponse(t *testing.T) {
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	resp := NewAgentResponse(agent, "Generated code", 0.95)

	assert.Equal(t, agent.ID, resp.AgentID)
	assert.Equal(t, agent.Role, resp.AgentRole)
	assert.Equal(t, "Generated code", resp.Content)
	assert.Equal(t, 0.95, resp.Confidence)
	assert.NotNil(t, resp.ToolsUsed)
	assert.NotNil(t, resp.Metadata)
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage("agent-123", MessageTypeProposal, "Here is my proposal")

	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "agent-123", msg.FromAgentID)
	assert.Equal(t, MessageTypeProposal, msg.Type)
	assert.Equal(t, "Here is my proposal", msg.Content)
	assert.NotNil(t, msg.Context)
}

func TestNewConsensusResult(t *testing.T) {
	consensus := NewConsensusResult()

	assert.False(t, consensus.Reached)
	assert.Equal(t, 0.0, consensus.Confidence)
	assert.NotNil(t, consensus.KeyPoints)
	assert.NotNil(t, consensus.Dissents)
	assert.NotNil(t, consensus.Votes)
}

func TestConsensusResult_AddVote(t *testing.T) {
	consensus := NewConsensusResult()
	consensus.AddVote("agent-1", 0.9)
	consensus.AddVote("agent-2", 0.8)

	assert.Equal(t, 0.9, consensus.Votes["agent-1"])
	assert.Equal(t, 0.8, consensus.Votes["agent-2"])
}

func TestConsensusResult_CalculateConfidence(t *testing.T) {
	consensus := NewConsensusResult()
	assert.Equal(t, 0.0, consensus.CalculateConfidence())

	consensus.AddVote("agent-1", 0.9)
	consensus.AddVote("agent-2", 0.7)

	// Average of 0.9 and 0.7
	assert.Equal(t, 0.8, consensus.CalculateConfidence())
}

func TestNewContext(t *testing.T) {
	ctx := NewContext("Implement auth", "myapp", "go")

	assert.NotEmpty(t, ctx.ID)
	assert.Equal(t, "Implement auth", ctx.Topic)
	assert.Equal(t, "myapp", ctx.Codebase)
	assert.Equal(t, "go", ctx.Language)
	assert.NotNil(t, ctx.Messages)
	assert.NotNil(t, ctx.Responses)
	assert.NotNil(t, ctx.Artifacts)
	assert.NotNil(t, ctx.Metadata)
}

func TestContext_AddMessage(t *testing.T) {
	ctx := NewContext("Test", "app", "go")
	msg := NewMessage("agent-1", MessageTypeProposal, "Proposal")

	ctx.AddMessage(msg)

	assert.Len(t, ctx.Messages, 1)
	assert.Equal(t, msg, ctx.Messages[0])
}

func TestContext_AddResponse(t *testing.T) {
	ctx := NewContext("Test", "app", "go")
	agent := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	resp := NewAgentResponse(agent, "Code", 0.9)

	ctx.AddResponse(resp)

	assert.Len(t, ctx.Responses, 1)
	assert.Equal(t, resp, ctx.Responses[0])
}

func TestContext_GetMessagesByType(t *testing.T) {
	ctx := NewContext("Test", "app", "go")

	ctx.AddMessage(NewMessage("agent-1", MessageTypeProposal, "P1"))
	ctx.AddMessage(NewMessage("agent-2", MessageTypeCritique, "C1"))
	ctx.AddMessage(NewMessage("agent-1", MessageTypeProposal, "P2"))

	proposals := ctx.GetMessagesByType(MessageTypeProposal)
	assert.Len(t, proposals, 2)

	critiques := ctx.GetMessagesByType(MessageTypeCritique)
	assert.Len(t, critiques, 1)
}

func TestContext_GetResponsesByRole(t *testing.T) {
	ctx := NewContext("Test", "app", "go")

	gen := NewAgent(RoleGenerator, "openai", "gpt-4", 8.5)
	critic := NewAgent(RoleCritic, "openai", "gpt-4", 8.5)

	ctx.AddResponse(NewAgentResponse(gen, "Code", 0.9))
	ctx.AddResponse(NewAgentResponse(critic, "Critique", 0.8))

	genResponses := ctx.GetResponsesByRole(RoleGenerator)
	assert.Len(t, genResponses, 1)
}

func TestNewScore(t *testing.T) {
	score := NewScore(85, 100, "correctness", "Code correctness")

	assert.Equal(t, 85.0, score.Value)
	assert.Equal(t, 100.0, score.MaxValue)
	assert.Equal(t, "correctness", score.Category)
	assert.Equal(t, "Code correctness", score.Description)
}

func TestScore_Percentage(t *testing.T) {
	tests := []struct {
		value    float64
		maxValue float64
		expected float64
	}{
		{85, 100, 85.0},
		{50, 100, 50.0},
		{0, 100, 0.0},
		{1, 0, 0.0}, // Edge case
	}

	for _, tt := range tests {
		score := NewScore(tt.value, tt.maxValue, "test", "Test")
		assert.Equal(t, tt.expected, score.Percentage())
	}
}

func TestScore_IsPassing(t *testing.T) {
	score := NewScore(85, 100, "test", "Test")

	assert.True(t, score.IsPassing(80))
	assert.True(t, score.IsPassing(85))
	assert.False(t, score.IsPassing(90))
}

func TestNewArtifact(t *testing.T) {
	artifact := &Artifact{
		ID:      "artifact-1",
		Type:    ArtifactTypeCode,
		Name:    "auth.go",
		Content: "package auth",
		AgentID: "agent-1",
		Version: 1,
	}

	assert.Equal(t, "artifact-1", artifact.ID)
	assert.Equal(t, ArtifactTypeCode, artifact.Type)
	assert.Equal(t, "auth.go", artifact.Name)
}
