# Skills Package

The skills package provides a skill registry system for managing AI agent capabilities in HelixAgent.

## Overview

This package implements a registry for AI agent skills - discrete capabilities that can be discovered, invoked, and composed. It supports skill discovery via trigger phrases, hot-reload of skill definitions, and hierarchical skill organization.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Skills Registry                          │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Skill Index                           ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ ││
│  │  │  By Name    │  │ By Category │  │   By Trigger    │ ││
│  │  │   Index     │  │    Index    │  │     Index       │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────────┘ ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Skill Definitions                      ││
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐││
│  │  │ Code   │ │ Debug  │ │ Search │ │  Git   │ │ Deploy │││
│  │  │ Gen    │ │        │ │        │ │        │ │        │││
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Hot Reload                             ││
│  │  Directory Watcher | YAML Parser | Validation           ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Key Types

### Registry

The main skill registry.

```go
type Registry struct {
    skills        map[string]*Skill
    byCategory    map[string][]string
    byTrigger     map[string]string
    mu            sync.RWMutex
    watcher       *DirectoryWatcher
    config        RegistryConfig
}
```

### Skill

Represents a single skill definition.

```go
type Skill struct {
    ID          string            // Unique identifier
    Name        string            // Display name
    Description string            // What the skill does
    Category    string            // Skill category
    Triggers    []string          // Trigger phrases
    Parameters  []Parameter       // Input parameters
    Handler     SkillHandler      // Execution handler
    Examples    []Example         // Usage examples
    Enabled     bool              // Is skill active
    Priority    int               // Matching priority
    Metadata    map[string]string // Additional metadata
}

type Parameter struct {
    Name        string      // Parameter name
    Type        string      // Type (string, int, bool, etc.)
    Description string      // Parameter description
    Required    bool        // Is required
    Default     interface{} // Default value
}

type Example struct {
    Input    string // Example input
    Output   string // Expected output
    Context  string // Additional context
}
```

### SkillHandler

Function signature for skill execution.

```go
type SkillHandler func(ctx context.Context, params map[string]interface{}) (*SkillResult, error)

type SkillResult struct {
    Success     bool
    Output      interface{}
    Message     string
    Suggestions []string
    Metadata    map[string]interface{}
}
```

### RegistryConfig

Configuration for the registry.

```go
type RegistryConfig struct {
    SkillsDirectory string        // Directory containing skill definitions
    EnableHotReload bool          // Enable hot-reload on file changes
    WatchInterval   time.Duration // File watch interval
    ValidateOnLoad  bool          // Validate skills on load
}
```

## Skill Categories

HelixAgent organizes skills into categories:

| Category | Description | Example Skills |
|----------|-------------|----------------|
| `code` | Code generation and manipulation | generate, refactor, optimize |
| `debug` | Debugging and diagnostics | trace, profile, analyze |
| `search` | Code and documentation search | find, grep, semantic-search |
| `git` | Version control operations | commit, branch, merge |
| `deploy` | Deployment and CI/CD | build, deploy, rollback |
| `docs` | Documentation generation | document, explain, readme |
| `test` | Testing operations | unit-test, integration-test |
| `review` | Code review and quality | review, lint, security-scan |

## Usage Examples

### Initialize Registry

```go
import "dev.helix.agent/internal/skills"

// Create registry
registry := skills.NewRegistry(skills.RegistryConfig{
    SkillsDirectory: "configs/skills",
    EnableHotReload: true,
    WatchInterval:   5 * time.Second,
})

// Load skills
err := registry.LoadSkills()
if err != nil {
    return err
}
```

### Register a Skill

```go
// Register programmatically
skill := &skills.Skill{
    ID:          "code-generate",
    Name:        "Code Generator",
    Description: "Generate code based on natural language description",
    Category:    "code",
    Triggers:    []string{"generate code", "write code", "create function"},
    Parameters: []skills.Parameter{
        {
            Name:        "description",
            Type:        "string",
            Description: "What code to generate",
            Required:    true,
        },
        {
            Name:        "language",
            Type:        "string",
            Description: "Programming language",
            Required:    false,
            Default:     "go",
        },
    },
    Handler: func(ctx context.Context, params map[string]interface{}) (*skills.SkillResult, error) {
        description := params["description"].(string)
        language := params["language"].(string)

        // Generate code...
        code := generateCode(description, language)

        return &skills.SkillResult{
            Success: true,
            Output:  code,
            Message: "Code generated successfully",
        }, nil
    },
    Enabled: true,
}

registry.Register(skill)
```

### Find Skill by Trigger

```go
// Match user input to skill
userInput := "generate code for user authentication"

skill, params := registry.MatchTrigger(userInput)
if skill != nil {
    fmt.Printf("Matched skill: %s\n", skill.Name)
    fmt.Printf("Extracted params: %v\n", params)
}
```

### Execute a Skill

```go
// Execute by ID
result, err := registry.Execute(ctx, "code-generate", map[string]interface{}{
    "description": "user authentication handler",
    "language":    "go",
})
if err != nil {
    return err
}

if result.Success {
    fmt.Printf("Output: %v\n", result.Output)
}
```

### List Skills

```go
// Get all skills
allSkills := registry.ListSkills()

// Get skills by category
codeSkills := registry.GetByCategory("code")

// Search skills
matches := registry.Search("generate")
for _, skill := range matches {
    fmt.Printf("- %s: %s\n", skill.Name, skill.Description)
}
```

### YAML Skill Definition

Skills can be defined in YAML files:

```yaml
# configs/skills/code-generate.yaml
id: code-generate
name: Code Generator
description: Generate code based on natural language description
category: code
enabled: true
priority: 10

triggers:
  - generate code
  - write code
  - create function
  - implement

parameters:
  - name: description
    type: string
    description: What code to generate
    required: true
  - name: language
    type: string
    description: Programming language
    required: false
    default: go

examples:
  - input: "generate code for user authentication"
    output: "func Authenticate(username, password string) error { ... }"
    context: "Go authentication handler"
```

### Hot Reload

```go
// Enable hot reload
registry := skills.NewRegistry(skills.RegistryConfig{
    SkillsDirectory: "configs/skills",
    EnableHotReload: true,
})

// Skills automatically reload when files change
registry.OnReload(func(skillID string, action string) {
    log.Printf("Skill %s %s", skillID, action) // "added", "updated", "removed"
})
```

## Integration with HelixAgent

Skills are used in the agent workflow:

```go
// In intent classifier
func (c *Classifier) ClassifyIntent(input string) (*Intent, error) {
    // Check if input matches a skill trigger
    skill, params := c.skillRegistry.MatchTrigger(input)
    if skill != nil {
        return &Intent{
            Type:       IntentSkill,
            SkillID:    skill.ID,
            Parameters: params,
            Confidence: 0.9,
        }, nil
    }

    // Continue with LLM classification...
}

// In debate service
func (s *DebateService) ExecuteSkill(ctx context.Context, skillID string, params map[string]interface{}) (*SkillResult, error) {
    return s.skillRegistry.Execute(ctx, skillID, params)
}
```

## Trigger Matching

The registry uses fuzzy matching for triggers:

```go
// Exact match
"generate code" → matches "generate code"

// Partial match
"can you generate some code" → matches "generate code"

// Semantic similarity (with embedding)
"write a function" → matches "generate code"
```

## Testing

```bash
go test -v ./internal/skills/...
```

### Testing Skills

```go
func TestCodeGenerateSkill(t *testing.T) {
    registry := skills.NewTestRegistry()

    // Register test skill
    registry.Register(&skills.Skill{
        ID:       "test-skill",
        Triggers: []string{"test trigger"},
        Handler: func(ctx context.Context, params map[string]interface{}) (*skills.SkillResult, error) {
            return &skills.SkillResult{
                Success: true,
                Output:  "test output",
            }, nil
        },
    })

    // Test execution
    result, err := registry.Execute(ctx, "test-skill", nil)
    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, "test output", result.Output)
}
```

## Metrics

```go
metrics := registry.GetMetrics()
fmt.Printf("Registered skills: %d\n", metrics.SkillCount)
fmt.Printf("Skill executions: %d\n", metrics.ExecutionCount)
fmt.Printf("Average execution time: %v\n", metrics.AverageExecutionTime)
fmt.Printf("Most used skill: %s\n", metrics.MostUsedSkill)
```

## Best Practices

1. **Use descriptive triggers**: Make triggers natural and varied
2. **Validate parameters**: Check required params before execution
3. **Provide examples**: Help users understand skill usage
4. **Enable hot-reload in dev**: For faster iteration
5. **Organize by category**: Keep skills well-organized
6. **Set appropriate priority**: Higher priority skills match first
