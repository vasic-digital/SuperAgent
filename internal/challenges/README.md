# internal/challenges

## Overview

The challenges package contains HelixAgent-specific challenge implementations
that validate real-life system behavior. These are Go-native challenge
definitions that complement the shell-based challenge scripts in
`challenges/scripts/`.

## Package Structure

| File | Purpose |
|------|---------|
| `types.go` | Core types: ChallengeResult, ChallengeConfig, status enums |
| `orchestrator.go` | Challenge orchestrator: runs challenges in dependency order |
| `plugin.go` | Plugin system challenge: validates plugin loading and lifecycle |
| `infra_provider.go` | Infrastructure bridge challenge: container health and boot flow |
| `shell_adapter.go` | Shell adapter: bridges Go challenges to shell script execution |
| `reporter.go` | Challenge reporter: generates structured pass/fail reports |
| `env_loader.go` | Environment loader: reads `.env` files for challenge configuration |
| `stall_config.go` | Stall detection config: timeout and retry settings |
| `api_quality.go` | API quality challenge: validates endpoint responses |
| `debate_formation.go` | Debate formation challenge: validates team selection |
| `provider_verification.go` | Provider verification challenge: validates LLM provider access |

## Userflow Challenges

The `userflow/` subdirectory contains 22 Go-native userflow challenges
with dependency graph orchestration:

| File | Purpose |
|------|---------|
| `orchestrator.go` | Userflow orchestrator: manages flow execution and dependencies |
| `flows.go` | Flow definitions: 22 userflow challenge implementations |
| `results/` | Result types and aggregation for userflow reporting |

Userflow challenges are run with `--run-challenges=userflow` and cover
browser, API, gRPC, WebSocket, and build verification scenarios.

## Running Challenges

```bash
# Run Go-native challenges
go test ./internal/challenges/... -run TestChallenge

# Run userflow challenges
./bin/helixagent --run-challenges=userflow

# Run shell-based challenges (separate scripts)
./challenges/scripts/run_all_challenges.sh
```

## Design

All challenges follow the project mandate of validating actual behavior,
not just return codes. Each challenge:

1. Performs real operations (HTTP requests, container health checks, etc.)
2. Validates the response content matches expected behavior
3. Reports structured pass/fail results with diagnostic messages
4. Respects resource limits (GOMAXPROCS=2, nice -n 19)
