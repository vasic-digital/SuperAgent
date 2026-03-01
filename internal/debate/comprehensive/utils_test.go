package comprehensive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptBuilder_New(t *testing.T) {
	pb := NewPromptBuilder("You are an expert")
	assert.NotNil(t, pb)
}

func TestPromptBuilder_Build(t *testing.T) {
	pb := NewPromptBuilder("System prompt")
	pb.AddContext("Context info")
	pb.AddInstruction("Do something")
	pb.AddExample("input", "output")
	pb.AddConstraint("Be careful")

	prompt := pb.Build()
	assert.Contains(t, prompt, "System prompt")
	assert.Contains(t, prompt, "Context info")
	assert.Contains(t, prompt, "Do something")
	assert.Contains(t, prompt, "Input: input")
	assert.Contains(t, prompt, "Be careful")
}

func TestPromptBuilder_Build_Empty(t *testing.T) {
	pb := NewPromptBuilder("Only system")
	prompt := pb.Build()
	assert.Equal(t, "# System\nOnly system", prompt)
}

func TestRolePrompts_Architect(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Architect()
	assert.Contains(t, prompt, "architect")
	assert.Contains(t, prompt, "scalable")
}

func TestRolePrompts_Generator(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Generator()
	assert.Contains(t, prompt, "developer")
	assert.Contains(t, prompt, "clean")
}

func TestRolePrompts_Critic(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Critic()
	assert.Contains(t, prompt, "reviewer")
	assert.Contains(t, prompt, "bugs")
}

func TestRolePrompts_Refactoring(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Refactoring()
	assert.Contains(t, prompt, "refactoring")
}

func TestRolePrompts_Tester(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Tester()
	assert.Contains(t, prompt, "testing")
}

func TestRolePrompts_Validator(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Validator()
	assert.Contains(t, prompt, "correctness")
}

func TestRolePrompts_Security(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Security()
	assert.Contains(t, prompt, "security")
}

func TestRolePrompts_Performance(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Performance()
	assert.Contains(t, prompt, "performance")
}

func TestRolePrompts_Moderator(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.Moderator()
	assert.Contains(t, prompt, "moderator")
}

func TestRolePrompts_RedTeam(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.RedTeam()
	assert.Contains(t, prompt, "adversar")
	assert.Contains(t, prompt, "attack")
}

func TestRolePrompts_BlueTeam(t *testing.T) {
	rp := RolePrompts{}
	prompt := rp.BlueTeam()
	assert.Contains(t, prompt, "defensive")
	assert.Contains(t, prompt, "error")
}

func TestParser_ParseCodeBlocks(t *testing.T) {
	parser := Parser{}
	content := `
Some text
` + "```go\npackage main\n\nfunc main() {}\n```" + `
More text
` + "```python\nprint('hello')\n```"

	blocks := parser.ParseCodeBlocks(content)
	assert.Len(t, blocks, 2)
	assert.Equal(t, "go", blocks[0].Language)
	assert.Equal(t, "python", blocks[1].Language)
}

func TestParser_ParseCodeBlocks_Empty(t *testing.T) {
	parser := Parser{}
	blocks := parser.ParseCodeBlocks("No code here")
	assert.Empty(t, blocks)
}

func TestParser_ExtractThoughts(t *testing.T) {
	parser := Parser{}
	content := `Thinking: This is my reasoning

Some other text

Reasoning: Another thought`

	thoughts := parser.ExtractThoughts(content)
	assert.NotEmpty(t, thoughts)
}

func TestParser_ParseConfidence(t *testing.T) {
	parser := Parser{}

	tests := []struct {
		content  string
		expected float64
	}{
		{"Confidence: 95%", 0.95},
		{"confidence: 0.8", 0.8},
		{"I am 90% confident", 0.9},
		{"No confidence here", 0.5}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			result := parser.ParseConfidence(tt.content)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestParser_ExtractKeyPoints(t *testing.T) {
	parser := Parser{}
	content := `
- First point
- Second point
* Third point
1. Numbered point
`

	points := parser.ExtractKeyPoints(content)
	assert.Len(t, points, 4)
}

func TestValidator_ValidateCode_Go(t *testing.T) {
	validator := Validator{}

	// Valid Go code
	code := `package main

func main() {
	println("hello")
}`

	errors := validator.ValidateCode(code, "go")
	assert.Empty(t, errors)
}

func TestValidator_ValidateCode_Empty(t *testing.T) {
	validator := Validator{}

	errors := validator.ValidateCode("", "go")
	assert.NotEmpty(t, errors)
	assert.Equal(t, "empty", errors[0].Type)
}

func TestValidator_ValidateCode_MissingPackage(t *testing.T) {
	validator := Validator{}

	code := `func main() {}`
	errors := validator.ValidateCode(code, "go")

	found := false
	for _, e := range errors {
		if e.Type == "missing_package" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidator_ValidateCode_UnbalancedBraces(t *testing.T) {
	validator := Validator{}

	code := `package main
func main() {`
	errors := validator.ValidateCode(code, "go")

	found := false
	for _, e := range errors {
		if e.Type == "unbalanced_braces" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidator_ValidateAgentResponse(t *testing.T) {
	validator := Validator{}

	resp := &AgentResponse{
		Content:    "Valid response",
		Confidence: 0.9,
	}

	errors := validator.ValidateAgentResponse(resp)
	assert.Empty(t, errors)
}

func TestValidator_ValidateAgentResponse_Empty(t *testing.T) {
	validator := Validator{}

	resp := &AgentResponse{
		Content:    "",
		Confidence: 0.5,
	}

	errors := validator.ValidateAgentResponse(resp)
	assert.NotEmpty(t, errors)
	assert.Equal(t, "empty_response", errors[0].Type)
}

func TestValidator_ValidateAgentResponse_InvalidConfidence(t *testing.T) {
	validator := Validator{}

	resp := &AgentResponse{
		Content:    "Test",
		Confidence: 1.5,
	}

	errors := validator.ValidateAgentResponse(resp)
	assert.NotEmpty(t, errors)
	assert.Equal(t, "invalid_confidence", errors[0].Type)
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		result := TruncateString(tt.input, tt.maxLen)
		assert.Equal(t, tt.expected, result)
	}
}

func TestCleanWhitespace(t *testing.T) {
	input := "  multiple   spaces   and\ttabs  "
	expected := "multiple spaces and tabs"

	result := CleanWhitespace(input)
	assert.Equal(t, expected, result)
}

func TestCountTokens(t *testing.T) {
	// Rough approximation: ~4 chars per token
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"test", 1},
		{"this is a longer sentence", 6},
	}

	for _, tt := range tests {
		result := CountTokens(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestConsensusAlgorithm_Calculate(t *testing.T) {
	alg := NewConsensusAlgorithm(0.8)

	responses := []*AgentResponse{
		{AgentID: "1", Confidence: 0.9, Content: "Agree"},
		{AgentID: "2", Confidence: 0.85, Content: "Agree"},
		{AgentID: "3", Confidence: 0.9, Content: "Agree"},
	}

	consensus, err := alg.Calculate(responses)
	assert.NoError(t, err)
	assert.True(t, consensus.Reached)
	assert.Greater(t, consensus.Confidence, 0.8)
}

func TestConsensusAlgorithm_Calculate_NoResponses(t *testing.T) {
	alg := NewConsensusAlgorithm(0.8)

	_, err := alg.Calculate([]*AgentResponse{})
	assert.Error(t, err)
}

func TestVoteAggregator_MajorityVote(t *testing.T) {
	agg := NewVoteAggregator(VotingMethodMajority)

	votes := map[string]float64{
		"option_a": 3,
		"option_b": 1,
		"option_c": 1,
	}

	winner, confidence, err := agg.Aggregate(votes)
	assert.NoError(t, err)
	assert.Equal(t, "option_a", winner)
	assert.Greater(t, confidence, 0.5)
}

func TestVoteAggregator_UnanimousVote(t *testing.T) {
	agg := NewVoteAggregator(VotingMethodUnanimous)

	// Unanimous
	votes := map[string]float64{
		"option_a": 1,
	}

	_, confidence, err := agg.Aggregate(votes)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, confidence)
}

func TestVoteAggregator_Unanimous_NotUnanimous(t *testing.T) {
	agg := NewVoteAggregator(VotingMethodUnanimous)

	votes := map[string]float64{
		"option_a": 1,
		"option_b": 1,
	}

	_, _, err := agg.Aggregate(votes)
	assert.Error(t, err)
}

func TestVoteAggregator_NoVotes(t *testing.T) {
	agg := NewVoteAggregator(VotingMethodMajority)

	_, _, err := agg.Aggregate(map[string]float64{})
	assert.Error(t, err)
}

func TestConvergenceDetector_Check(t *testing.T) {
	detector := NewConvergenceDetector(3, 0.9)

	// Should converge on high confidence
	assert.True(t, detector.Check(5, 3, 0.95))

	// Should not converge yet
	assert.False(t, detector.Check(2, 0, 0.5))

	// Should converge on too many unchanged rounds
	assert.True(t, detector.Check(10, 6, 0.5))
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 10, config.MaxRounds)
	assert.Equal(t, 0.8, config.MinConsensus)
	assert.Equal(t, 0.95, config.QualityThreshold)
	assert.Equal(t, 0.95, config.TestPassRate)

	// All agents should be enabled
	assert.True(t, config.EnableArchitect)
	assert.True(t, config.EnableGenerator)
	assert.True(t, config.EnableCritic)
}
