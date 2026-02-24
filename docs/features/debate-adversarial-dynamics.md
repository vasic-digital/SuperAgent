# Debate Adversarial Dynamics

## Overview

The adversarial dynamics subsystem implements a **Red Team / Blue Team** attack-defend protocol within the HelixAgent debate system. A Red Team agent probes the current solution for vulnerabilities, edge cases, and stress scenarios, then a Blue Team agent patches the identified issues. This cycle repeats until the solution is hardened or termination conditions are met.

The implementation lives in `internal/debate/agents/adversarial.go` and integrates with the main debate protocol as Phase 7 (Adversarial) of the 8-phase pipeline.

## Core Types

### Vulnerability

Each finding discovered by the Red Team is represented as a `Vulnerability`:

| Field         | Type     | Description                                                    |
|---------------|----------|----------------------------------------------------------------|
| `ID`          | `string` | Unique identifier (e.g., `VULN-001`)                          |
| `Category`    | `string` | One of: `injection`, `overflow`, `race_condition`, `logic_error`, `auth`, `xss`, `other` |
| `Severity`    | `string` | One of: `critical`, `high`, `medium`, `low`                   |
| `Description` | `string` | What the vulnerability is                                     |
| `Evidence`    | `string` | Code evidence supporting the finding                          |
| `Exploit`     | `string` | How the vulnerability can be exploited                        |

### Attack Report

The Red Team produces an `AttackReport` per round:

| Field               | Type               | Description                           |
|---------------------|--------------------|---------------------------------------|
| `Vulnerabilities`   | `[]Vulnerability`  | Security and correctness issues found |
| `EdgeCases`         | `[]EdgeCase`       | Boundary conditions identified        |
| `StressScenarios`   | `[]StressScenario` | High-load / resource exhaustion tests |
| `OverallRisk`       | `float64`          | Aggregate risk score (0.0 - 1.0)     |
| `Round`             | `int`              | Which adversarial round               |

### Defense Report

The Blue Team responds with a `DefenseReport`:

| Field                    | Type                | Description                               |
|--------------------------|---------------------|-------------------------------------------|
| `PatchedVulnerabilities` | `[]string`          | IDs of vulnerabilities addressed           |
| `Patches`                | `map[string]string` | Vulnerability ID to fix description        |
| `RemainingRisks`         | `[]string`          | Issues that could not be fully resolved    |
| `ConfidenceInDefense`    | `float64`           | Blue Team confidence (0.0 - 1.0)          |
| `PatchedCode`            | `string`            | The improved source code                   |
| `Round`                  | `int`               | Which adversarial round                    |

## Protocol Configuration

The `AdversarialConfig` controls the protocol behavior:

| Parameter            | Default   | Description                                              |
|----------------------|-----------|----------------------------------------------------------|
| `MaxRounds`          | 3         | Maximum attack-defend cycles                             |
| `MinVulnerabilities` | 1         | Minimum new vulnerabilities to continue the cycle        |
| `RiskThreshold`      | 0.2       | Stop if overall risk drops below this value              |
| `Timeout`            | 5 minutes | Total time budget for the adversarial protocol           |

## Multi-Round Cycle

The `AdversarialProtocol.Execute(ctx, solution, language)` method runs the full loop:

```
FOR round = 1 to MaxRounds:
    1. Check timeout / context cancellation
    2. Red Team ATTACKS the current code
       -> Produces AttackReport (vulnerabilities, edge cases, stress scenarios)
    3. Check termination:
       a. If vulnerabilities found < MinVulnerabilities -> STOP (solution is hardened)
       b. If OverallRisk < RiskThreshold -> STOP (risk is acceptable)
    4. Blue Team DEFENDS against the attack report
       -> Produces DefenseReport (patches, remaining risks, patched code)
    5. Update current code with patched version
    6. Feed defense report into next round's attack context
RETURN AdversarialResult
```

## Termination Conditions

The adversarial loop terminates when any of these conditions is met:

1. **No new vulnerabilities** -- the Red Team finds fewer than `MinVulnerabilities` new issues, indicating the solution has been sufficiently hardened.
2. **Risk below threshold** -- the `OverallRisk` score drops below `RiskThreshold`, indicating acceptable residual risk.
3. **Maximum rounds reached** -- the configured `MaxRounds` limit is hit.
4. **Timeout** -- the context deadline or configured `Timeout` expires.

## Red Team (Attack Phase)

### LLM-based attack

The Red Team agent sends a structured prompt to the LLM containing:
- The current code in a fenced block
- The programming language
- The round number
- If round > 1: the Blue Team's prior patches (so the Red Team can attempt to bypass them)

The prompt requests structured output in three sections: `VULNERABILITIES`, `EDGE_CASES`, `STRESS_SCENARIOS`, each with labelled fields, separated by `---` delimiters, followed by an `OVERALL_RISK` score.

### Deterministic fallback attack

When the LLM is unavailable or returns unparseable output, the Red Team falls back to static pattern analysis of the source code. It scans for:

| Pattern                                    | Category         | Severity |
|--------------------------------------------|------------------|----------|
| String formatting in queries without `prepare`/`parameterize` | `injection`      | critical |
| No `validate`/`sanitize`/`escape` calls    | `logic_error`    | medium   |
| `goroutine`/`thread` without `mutex`/`sync`| `race_condition` | high     |
| Discarded error values (`_ =`)             | `logic_error`    | medium   |
| `unsafe`/`pointer`/`buffer`/`alloc`        | `overflow`       | high     |

Language-specific edge cases are generated (Go: nil pointer, Python: None, JS/TS: undefined). A standard stress scenario (1000 concurrent requests) is always included.

The fallback risk score is computed from vulnerability severity weights: critical=1.0, high=0.7, medium=0.4, low=0.2, averaged across all findings.

## Blue Team (Defense Phase)

### LLM-based defense

The Blue Team receives the attack report and the original code, then produces:
- A list of patched vulnerability IDs
- A description of each patch
- Any remaining risks
- A confidence score
- The fully patched source code

### Deterministic fallback defense

When the LLM is unavailable, the Blue Team generates generic fix descriptions based on vulnerability category:

| Category         | Fix Description                                     |
|------------------|-----------------------------------------------------|
| `injection`      | Use parameterized queries and input sanitization     |
| `overflow`       | Add bounds checking and safe buffer operations       |
| `race_condition` | Add mutex/channel synchronization for shared state   |
| `logic_error`    | Add proper validation and error handling             |
| `auth`           | Enforce authentication and authorization checks      |
| `xss`            | Escape output and apply Content-Security-Policy      |

Edge cases are marked as remaining risks. The fallback defense confidence is set to 0.4 to reflect its limited accuracy.

## Integration with the 8-Phase Protocol

The adversarial dynamics integrate as **Phase 7 (Adversarial)** in the debate protocol defined in `internal/debate/protocol/protocol.go`. The phase configuration specifies:

- **Required roles:** `RoleRedTeam` and `RoleBlueTeam`
- **Max parallelism:** 4 agents
- **Prompt template:** instructs Red Team agents to find vulnerabilities and Blue Team agents to patch them

The phase produces `PhaseResponse` records that flow into the convergence phase (Phase 8) for final voting.

## Result Structure

The `AdversarialResult` returned by `Execute()`:

| Field            | Type               | Description                              |
|------------------|--------------------|------------------------------------------|
| `Rounds`         | `int`              | Number of completed rounds               |
| `FinalCode`      | `string`           | The hardened solution                    |
| `AttackReports`  | `[]*AttackReport`  | All attack findings across rounds        |
| `DefenseReports` | `[]*DefenseReport` | All defense patches across rounds        |
| `AllResolved`    | `bool`             | Whether all vulnerabilities were patched |
| `RemainingRisks` | `[]string`         | Unresolved issues from the last round    |
| `Duration`       | `time.Duration`    | Total protocol execution time            |

## Related Files

- `internal/debate/agents/adversarial.go` -- Core adversarial protocol implementation
- `internal/debate/agents/adversarial_test.go` -- Unit tests
- `internal/debate/protocol/protocol.go` -- Phase 7 configuration
- `internal/debate/topology/topology.go` -- `RoleRedTeam` and `RoleBlueTeam` role definitions
