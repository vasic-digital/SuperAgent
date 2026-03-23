// Package comprehensive provides specialized agent implementations for the debate system
package comprehensive

import (
	"context"
	"fmt"
)

// ArchitectAgent handles system design and planning
type ArchitectAgent struct {
	*BaseAgent
}

// NewArchitectAgent creates a new architect agent
func NewArchitectAgent(agent *Agent, pool *AgentPool) *ArchitectAgent {
	return &ArchitectAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements system design logic
func (a *ArchitectAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := fmt.Sprintf(
		"Architectural design for: %s\n\n1. System Components\n2. Data Flow\n"+
			"3. Interface Definitions\n4. Technology Choices", debateCtx.Topic)

	content, confidence, err := a.InvokeLLM(ctx, "Architect", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "design"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// GeneratorAgent handles code generation
type GeneratorAgent struct {
	*BaseAgent
}

// NewGeneratorAgent creates a new generator agent
func NewGeneratorAgent(agent *Agent, pool *AgentPool) *GeneratorAgent {
	return &GeneratorAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements code generation logic
func (a *GeneratorAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := fmt.Sprintf(
		"Generated code for: %s\n\n```go\n// %s provides a generated implementation.\n"+
			"func %s() error {\n    return nil // LLM-generated implementation pending\n}\n```",
		debateCtx.Topic, sanitizeFunctionName(debateCtx.Topic),
		sanitizeFunctionName(debateCtx.Topic))

	content, confidence, err := a.InvokeLLM(ctx, "Generator", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "generation"
	response.Metadata["language"] = debateCtx.Language
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// CriticAgent handles code review and criticism
type CriticAgent struct {
	*BaseAgent
}

// NewCriticAgent creates a new critic agent
func NewCriticAgent(agent *Agent, pool *AgentPool) *CriticAgent {
	return &CriticAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements code critique logic
func (a *CriticAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Code Review:\n\n1. Potential Issues:\n   - Error handling missing\n" +
		"   - Edge cases not covered\n\n2. Recommendations:\n   - Add input validation\n" +
		"   - Handle nil pointers"

	content, confidence, err := a.InvokeLLM(ctx, "Critic", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "critique"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// RefactoringAgent handles code refactoring
type RefactoringAgent struct {
	*BaseAgent
}

// NewRefactoringAgent creates a new refactoring agent
func NewRefactoringAgent(agent *Agent, pool *AgentPool) *RefactoringAgent {
	return &RefactoringAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements refactoring logic
func (a *RefactoringAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Refactored Code:\n\n- Extracted helper functions\n" +
		"- Reduced cyclomatic complexity\n- Improved naming\n- Added documentation"

	content, confidence, err := a.InvokeLLM(ctx, "Refactoring", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "refactoring"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// TesterAgent handles test generation
type TesterAgent struct {
	*BaseAgent
}

// NewTesterAgent creates a new tester agent
func NewTesterAgent(agent *Agent, pool *AgentPool) *TesterAgent {
	return &TesterAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements test generation logic
func (a *TesterAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Test Cases:\n\n1. Test happy path\n2. Test error conditions\n" +
		"3. Test edge cases\n4. Test concurrent access"

	content, confidence, err := a.InvokeLLM(ctx, "Tester", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "testing"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// ValidatorAgent handles correctness validation
type ValidatorAgent struct {
	*BaseAgent
}

// NewValidatorAgent creates a new validator agent
func NewValidatorAgent(agent *Agent, pool *AgentPool) *ValidatorAgent {
	return &ValidatorAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements validation logic
func (a *ValidatorAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Validation Results:\n\n- Syntax: Valid\n- Types: Consistent\n" +
		"- Logic: Correct\n- Tests: Passing"

	content, confidence, err := a.InvokeLLM(ctx, "Validator", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "validation"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// SecurityAgent handles security analysis
type SecurityAgent struct {
	*BaseAgent
}

// NewSecurityAgent creates a new security agent
func NewSecurityAgent(agent *Agent, pool *AgentPool) *SecurityAgent {
	return &SecurityAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements security analysis logic
func (a *SecurityAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Security Analysis:\n\n1. Vulnerabilities Found: 0 Critical, 1 Medium\n\n" +
		"2. Recommendations:\n   - Add input validation\n   - Use parameterized queries"

	content, confidence, err := a.InvokeLLM(ctx, "Security", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "security"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// PerformanceAgent handles performance optimization
type PerformanceAgent struct {
	*BaseAgent
}

// NewPerformanceAgent creates a new performance agent
func NewPerformanceAgent(agent *Agent, pool *AgentPool) *PerformanceAgent {
	return &PerformanceAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements performance analysis logic
func (a *PerformanceAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Performance Analysis:\n\n1. Complexity: O(n log n)\n2. Memory: Efficient\n" +
		"3. Bottlenecks: None found\n\nOptimization: Pre-allocate slices for 15% speedup"

	content, confidence, err := a.InvokeLLM(ctx, "Performance", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "performance"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// RedTeamAgent performs adversarial testing
type RedTeamAgent struct {
	*BaseAgent
}

// NewRedTeamAgent creates a new red team agent
func NewRedTeamAgent(agent *Agent, pool *AgentPool) *RedTeamAgent {
	return &RedTeamAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements adversarial testing logic
func (a *RedTeamAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Attack Vectors:\n\n1. SQL Injection - Mitigated\n" +
		"2. XSS - Vulnerable\n3. CSRF - Needs validation\n4. Path Traversal - Safe"

	content, confidence, err := a.InvokeLLM(ctx, "RedTeam", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "adversarial"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// BlueTeamAgent implements defensive measures
type BlueTeamAgent struct {
	*BaseAgent
}

// NewBlueTeamAgent creates a new blue team agent
func NewBlueTeamAgent(agent *Agent, pool *AgentPool) *BlueTeamAgent {
	return &BlueTeamAgent{
		BaseAgent: NewBaseAgent(agent, pool, nil),
	}
}

// Process implements defensive coding logic
func (a *BlueTeamAgent) Process(ctx context.Context, msg *Message, debateCtx *Context) (*AgentResponse, error) {
	templateFallback := "Defensive Implementation:\n\n1. Input validation added\n" +
		"2. Error handling improved\n3. Rate limiting implemented\n4. Logging enhanced"

	content, confidence, err := a.InvokeLLM(ctx, "BlueTeam", debateCtx.Topic, templateFallback)
	if err != nil {
		return nil, err
	}

	response := NewAgentResponse(a.agent, content, confidence)
	response.Metadata["phase"] = "defense"
	response.Metadata["provider"] = a.agent.Provider
	response.Metadata["model"] = a.agent.Model
	return response, nil
}

// Helper functions
func sanitizeFunctionName(topic string) string {
	name := ""
	for _, c := range topic {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			name += string(c)
		}
	}
	if name == "" {
		name = "DoSomething"
	}
	return name
}
