# debate/tools - Debate Tooling

Provides CI/CD hooks, Git worktree integration, service bridge, and tool integration for the debate system.

## Purpose

The tools package connects the debate system with external development tools, enabling debates to trigger CI/CD pipelines, manage Git worktrees for isolated code changes, and bridge with external services.

## Key Components

### CICDHook

Configurable CI/CD validation pipelines that run between debate phases.

```go
hook := tools.NewCICDHook(config, logger)
result, err := hook.RunValidation(ctx, phase, codeOutput)
```

Supported validations:
- Unit tests
- Linting (golangci-lint)
- Static analysis (go vet, gosec)
- Security scanning
- Build verification

### GitTool

Git worktree management for isolated debate session version control.

```go
gitTool := tools.NewGitTool(repoPath, logger)
worktree, err := gitTool.CreateWorktree(ctx, sessionID)
commitHash, err := gitTool.SnapshotCommit(ctx, worktree, "debate round 3")
diff, err := gitTool.GetDiff(ctx, worktree, baseCommit)
```

Features:
- Creates isolated Git worktrees per debate session
- Snapshot commits for each debate round
- Diff generation between rounds
- Cleanup on session completion

### ServiceBridge

Connects the debate system with external services (formatters, MCP servers, etc.).

```go
bridge := tools.NewServiceBridge(serviceRegistry, logger)
formatted, err := bridge.FormatCode(ctx, code, language)
mcpResult, err := bridge.CallMCPTool(ctx, serverID, toolName, params)
```

### ToolIntegration

Unified tool orchestrator that coordinates all debate tools.

```go
integration := tools.NewToolIntegration(cicd, git, bridge, logger)
integration.OnPhaseComplete(ctx, phase, output)
```

## Key Types

- **CICDConfig** -- Pipeline configuration: which validations to run per phase
- **CICDResult** -- Validation result with pass/fail, logs, and duration
- **WorktreeInfo** -- Git worktree metadata: path, branch, session ID
- **ToolResult** -- Generic tool execution result

## Usage within Debate System

Tools are invoked by the debate orchestrator at phase boundaries. CI/CD hooks validate code quality after the Optimization phase. Git worktrees provide isolation for code-generating debates. The service bridge enables debates to use formatters and MCP tools during the generation process.

## Files

- `cicd_hook.go` -- CI/CD pipeline integration
- `cicd_hook_test.go` -- Unit tests
- `git_tool.go` -- Git worktree management
- `git_tool_test.go` -- Unit tests
- `service_bridge.go` -- External service bridge
- `tool_integration.go` -- Unified tool orchestrator
