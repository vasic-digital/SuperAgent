# Implementation Specification: Context Templates

**Document ID:** IMPL-002  
**Feature:** Context Templates System  
**Priority:** CRITICAL  
**Phase:** 1  
**Estimated Effort:** 2 weeks  
**Source:** Claude Code, Cline, GPTMe

---

## Overview

Implement a context template system that allows users to save, load, and share predefined context configurations for common development workflows.

## Use Cases

1. **Onboarding Template** - Include README, architecture docs, setup guides
2. **Bug Fix Template** - Include relevant test files, error logs, related code
3. **Feature Development Template** - Include feature specs, API docs, related modules
4. **Code Review Template** - Include changed files, test results, lint output
5. **Debugging Template** - Include logs, traces, configuration files

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Context Template System                          │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │   Template   │  │   Context    │  │  Template    │              │
│  │   Manager    │  │   Resolver   │  │  Marketplace │              │
│  │              │  │              │  │              │              │
│  │ - CRUD ops   │  │ - File glob  │  │ - Publish    │              │
│  │ - Validation │  │ - Git diff   │  │ - Download   │              │
│  │ - Versioning │  │ - Dynamic    │  │ - Ratings    │              │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘              │
│         │                 │                 │                       │
│         └─────────────────┴─────────────────┘                       │
│                           │                                         │
│                           ▼                                         │
│              ┌──────────────────────┐                               │
│              │   Template Store     │                               │
│              │   (File + Database)  │                               │
│              └──────────────────────┘                               │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Model

### Template Schema

```yaml
# Template file format (template.yaml)
api_version: "v1"
kind: "ContextTemplate"

metadata:
  id: "feature-development"
  name: "Feature Development"
  description: "Context for developing new features"
  author: "helixagent"
  version: "1.0.0"
  tags: ["development", "feature", "planning"]
  created_at: "2026-04-04T10:00:00Z"
  updated_at: "2026-04-04T10:00:00Z"
  
spec:
  # Static file patterns
  files:
    include:
      - "README.md"
      - "docs/ARCHITECTURE.md"
      - "docs/API.md"
      - "src/**/*.go"
    exclude:
      - "*_test.go"
      - "vendor/**"
      
  # Dynamic context based on git state
  git_context:
    # Include files changed in current branch
    branch_diff:
      enabled: true
      base: "main"
      max_files: 20
    
    # Include recent commits for context
    recent_commits:
      enabled: true
      count: 5
      
    # Include related files (files that change together)
    related_files:
      enabled: true
      max_files: 10
      
  # Context7 documentation integration
  documentation:
    enabled: true
    sources:
      - type: "mcp"
        server: "context7"
        query: "{{feature_name}}"
        
  # Custom instructions for this template
  instructions: |
    You are helping develop a new feature. Consider:
    1. API compatibility
    2. Test coverage
    3. Documentation updates
    4. Performance implications
    
  # Variable substitution
  variables:
    - name: "feature_name"
      description: "Name of the feature being developed"
      required: true
      default: ""
    - name: "priority"
      description: "Feature priority"
      required: false
      default: "medium"
      options: ["low", "medium", "high", "critical"]
      
  # Predefined prompts for this template
  prompts:
    - name: "plan"
      description: "Create implementation plan"
      template: |
        Create a detailed implementation plan for {{feature_name}} 
        with {{priority}} priority.
        
    - name: "review"
      description: "Review implementation"
      template: |
        Review the implementation of {{feature_name}} for:
        - Code quality
        - Test coverage
        - Documentation
        - Performance
```

### Go Types

```go
package templates

// ContextTemplate represents a context template
type ContextTemplate struct {
    APIVersion string            `json:"api_version" yaml:"api_version"`
    Kind       string            `json:"kind" yaml:"kind"`
    Metadata   TemplateMetadata  `json:"metadata" yaml:"metadata"`
    Spec       TemplateSpec      `json:"spec" yaml:"spec"`
}

type TemplateMetadata struct {
    ID          string            `json:"id" yaml:"id"`
    Name        string            `json:"name" yaml:"name"`
    Description string            `json:"description" yaml:"description"`
    Author      string            `json:"author" yaml:"author"`
    Version     string            `json:"version" yaml:"version"`
    Tags        []string          `json:"tags" yaml:"tags"`
    CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}

type TemplateSpec struct {
    Files          FileSpec         `json:"files" yaml:"files"`
    GitContext     GitContextSpec   `json:"git_context" yaml:"git_context"`
    Documentation  DocumentationSpec `json:"documentation" yaml:"documentation"`
    Instructions   string           `json:"instructions" yaml:"instructions"`
    Variables      []VariableDef    `json:"variables" yaml:"variables"`
    Prompts        []PromptDef      `json:"prompts" yaml:"prompts"`
}

type FileSpec struct {
    Include []string `json:"include" yaml:"include"`
    Exclude []string `json:"exclude" yaml:"exclude"`
}

type GitContextSpec struct {
    BranchDiff    BranchDiffSpec    `json:"branch_diff" yaml:"branch_diff"`
    RecentCommits RecentCommitsSpec `json:"recent_commits" yaml:"recent_commits"`
    RelatedFiles  RelatedFilesSpec  `json:"related_files" yaml:"related_files"`
}

type VariableDef struct {
    Name        string   `json:"name" yaml:"name"`
    Description string   `json:"description" yaml:"description"`
    Required    bool     `json:"required" yaml:"required"`
    Default     string   `json:"default" yaml:"default"`
    Options     []string `json:"options,omitempty" yaml:"options,omitempty"`
}

type PromptDef struct {
    Name        string `json:"name" yaml:"name"`
    Description string `json:"description" yaml:"description"`
    Template    string `json:"template" yaml:"template"`
}
```

## Components

### 1. Template Manager (`internal/templates/manager.go`)

```go
package templates

// Manager handles template CRUD operations
type Manager struct {
    store       TemplateStore
    validator   *Validator
    marketplace MarketplaceClient
    config      ManagerConfig
}

type ManagerConfig struct {
    TemplatesDir     string
    MaxTemplates     int
    MaxTemplateSize  int64
    AllowCustom      bool
    MarketplaceURL   string
}

// CRUD Operations
func (m *Manager) Create(ctx context.Context, template *ContextTemplate) error
func (m *Manager) Get(ctx context.Context, id string) (*ContextTemplate, error)
func (m *Manager) Update(ctx context.Context, template *ContextTemplate) error
func (m *Manager) Delete(ctx context.Context, id string) error
func (m *Manager) List(ctx context.Context, filter ListFilter) ([]*ContextTemplate, error)

// Template Resolution
func (m *Manager) Resolve(ctx context.Context, id string, vars map[string]string) (*ResolvedContext, error)

// Validation
func (m *Manager) Validate(template *ContextTemplate) error
```

### 2. Context Resolver (`internal/templates/resolver.go`)

```go
package templates

// Resolver expands templates into actual context
type Resolver struct {
    gitResolver    *GitResolver
    fileResolver   *FileResolver
    docResolver    *DocumentationResolver
}

// ResolvedContext contains the expanded context
type ResolvedContext struct {
    Files           []ContextFile
    GitInfo         *GitContext
    Documentation   []Document
    Instructions    string
    Variables       map[string]string
    TotalTokens     int
}

type ContextFile struct {
    Path        string
    Content     string
    TokenCount  int
    Relevance   float64
}

func (r *Resolver) Resolve(ctx context.Context, template *ContextTemplate, vars map[string]string) (*ResolvedContext, error) {
    // 1. Validate variables
    // 2. Resolve file patterns
    // 3. Resolve git context
    // 4. Resolve documentation
    // 5. Substitute variables in instructions
    // 6. Calculate token counts
}
```

### 3. Git Resolver (`internal/templates/git_resolver.go`)

```go
package templates

type GitResolver struct {
    repo *git.Repository
}

func (r *GitResolver) GetBranchDiff(base string) ([]FileChange, error)
func (r *GitResolver) GetRecentCommits(count int) ([]Commit, error)
func (r *GitResolver) GetRelatedFiles(files []string, maxFiles int) ([]string, error)
func (r *GitResolver) GetStagedFiles() ([]string, error)
```

## Built-in Templates

Create default templates in `configs/templates/`:

```yaml
# onboarding.yaml
api_version: "v1"
kind: "ContextTemplate"
metadata:
  id: "onboarding"
  name: "Project Onboarding"
  description: "Get up to speed with a new project"
  author: "helixagent"
  version: "1.0.0"
  tags: ["onboarding", "documentation"]
  
spec:
  files:
    include:
      - "README.md"
      - "CONTRIBUTING.md"
      - "docs/**/*.md"
      - "Makefile"
      - "package.json"
      - "go.mod"
      - "pyproject.toml"
      - ".github/workflows/*.yml"
    exclude:
      - "docs/node_modules/**"
      
  instructions: |
    Help me understand this codebase. Focus on:
    1. Project structure and architecture
    2. Key technologies and dependencies
    3. Development workflow
    4. Testing approach
    5. Important conventions
```

```yaml
# bug-fix.yaml
api_version: "v1"
kind: "ContextTemplate"
metadata:
  id: "bug-fix"
  name: "Bug Fix"
  description: "Context for investigating and fixing bugs"
  author: "helixagent"
  version: "1.0.0"
  tags: ["debugging", "bug-fix"]
  
spec:
  files:
    include:
      - "**/*test*.go"
      - "**/*test*.py"
      - "logs/**"
      - ".github/workflows/**"
      
  git_context:
    recent_commits:
      enabled: true
      count: 10
      
  variables:
    - name: "error_message"
      description: "The error message or exception"
      required: true
    - name: "affected_component"
      description: "Component where bug occurs"
      required: false
      
  instructions: |
    Help me fix this bug. Consider:
    1. Root cause analysis
    2. Impact assessment
    3. Fix implementation
    4. Test coverage
    5. Regression prevention
```

```yaml
# code-review.yaml
api_version: "v1"
kind: "ContextTemplate"
metadata:
  id: "code-review"
  name: "Code Review"
  description: "Context for reviewing code changes"
  author: "helixagent"
  version: "1.0.0"
  tags: ["review", "quality"]
  
spec:
  git_context:
    branch_diff:
      enabled: true
      base: "main"
      max_files: 50
    recent_commits:
      enabled: true
      count: 3
      
  files:
    include:
      - "**/*test*"
      - "CODEOWNERS"
      - ".github/PULL_REQUEST_TEMPLATE.md"
      
  instructions: |
    Review this code for:
    1. Correctness and logic
    2. Code quality and style
    3. Test coverage
    4. Documentation
    5. Performance implications
    6. Security considerations
```

## API Endpoints

```go
// internal/handlers/templates_handler.go

func (h *TemplatesHandler) RegisterRoutes(router *gin.Engine) {
    templates := router.Group("/v1/templates")
    {
        templates.GET("", h.ListTemplates)
        templates.POST("", h.CreateTemplate)
        templates.GET("/:id", h.GetTemplate)
        templates.PUT("/:id", h.UpdateTemplate)
        templates.DELETE("/:id", h.DeleteTemplate)
        
        // Resolution
        templates.POST("/:id/resolve", h.ResolveTemplate)
        
        // Marketplace
        templates.GET("/marketplace", h.ListMarketplace)
        templates.POST("/marketplace/:id/install", h.InstallTemplate)
        templates.POST("/:id/publish", h.PublishTemplate)
        
        // Variables
        templates.GET("/:id/variables", h.GetVariables)
        templates.POST("/:id/validate", h.ValidateVariables)
    }
}
```

## CLI Commands

```bash
# List available templates
helixagent templates list

# Create new template
helixagent templates create --name "my-template" --from-current

# Apply template
helixagent templates apply onboarding

# Apply with variables
helixagent templates apply bug-fix --var error_message="null pointer" --var affected_component="auth"

# Install from marketplace
helixagent templates install helixagent/go-backend

# Publish template
helixagent templates publish my-template
```

## MCP Integration

```go
// Add to MCP tools
{
    "name": "load_context_template",
    "description": "Load a predefined context template",
    "input_schema": {
        "type": "object",
        "properties": {
            "template_id": {
                "type": "string",
                "description": "ID of the template to load"
            },
            "variables": {
                "type": "object",
                "description": "Template variables"
            }
        },
        "required": ["template_id"]
    }
}
```

## Implementation Timeline

**Week 1: Core System**
- [ ] Design template schema
- [ ] Implement TemplateManager
- [ ] Create built-in templates
- [ ] Write validation logic

**Week 2: Resolution & Integration**
- [ ] Implement GitResolver
- [ ] Implement FileResolver
- [ ] Add API endpoints
- [ ] MCP tool integration
- [ ] CLI commands
- [ ] Testing

## Testing

```go
func TestTemplateManager_Create(t *testing.T) {}
func TestResolver_Resolve(t *testing.T) {}
func TestGitResolver_GetBranchDiff(t *testing.T) {}
```
