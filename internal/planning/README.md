# Planning Package

The planning package provides AI-powered task planning and execution for HelixAgent.

## Overview

Implements:
- Task decomposition
- Execution planning
- Dependency resolution
- Progress tracking

## Key Components

```go
planner := planning.NewPlanner(config, llmClient)

plan, err := planner.CreatePlan(ctx, &planning.PlanRequest{
    Goal:        "Implement user authentication",
    Constraints: []string{"Use JWT", "Support OAuth2"},
})
```

## See Also

- `internal/verification/` - Plan verification
- `internal/background/` - Task execution
