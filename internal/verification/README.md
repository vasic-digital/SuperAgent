# Verification Package

The verification package provides formal verification capabilities for AI-generated code and plans in HelixAgent.

## Overview

This package implements:

- **Formal Verification**: Z3 and Dafny-based program verification
- **LTL Verification**: Linear Temporal Logic for plan verification
- **Specification Generation**: Automatic spec generation from code
- **Proof Checking**: Verification of mathematical proofs

## Key Components

### Formal Verifier

Main interface for code verification:

```go
verifier := verification.NewFormalVerifier(config, logger)

result, err := verifier.Verify(ctx, &verification.VerificationRequest{
    Code:     code,
    Language: "go",
    Specs:    specifications,
})
```

### Z3 Prover

SMT solver integration for constraint verification:

```go
prover := verification.NewZ3Prover(config)

result, err := prover.Prove(ctx, &verification.ProofRequest{
    Formula:     formula,
    Constraints: constraints,
})
```

### Dafny Verifier

Dafny language integration for verified code:

```go
dafny := verification.NewDafnyVerifier(config)

result, err := dafny.Verify(ctx, &verification.DafnyRequest{
    Source:  dafnyCode,
    Methods: []string{"MethodToVerify"},
})
```

### VeriPlan

Plan verification using temporal logic:

```go
veriplan := verification.NewVeriPlan(config)

result, err := veriplan.VerifyPlan(ctx, &verification.PlanRequest{
    Plan:       plan,
    Properties: []string{"safety", "liveness"},
})
```

## Verification Types

| Type | Description |
|------|-------------|
| **Type Safety** | Ensures type correctness |
| **Memory Safety** | Detects memory issues |
| **Bounds Checking** | Array/slice bounds verification |
| **Invariant Checking** | Loop and class invariants |
| **Pre/Post Conditions** | Function contracts |
| **Temporal Properties** | LTL/CTL model checking |

## LTL Formula Support

```go
// Create LTL formula
formula := verification.NewLTLFormula()
formula.AddGlobally("safe_state")          // G(safe)
formula.AddEventually("goal_reached")       // F(goal)
formula.AddUntil("precond", "postcond")    // precond U postcond

// Verify against plan
result, err := veriplan.CheckLTL(ctx, plan, formula)
```

## Specification Language

```go
spec := &verification.Specification{
    Name: "BinarySearch",
    Preconditions: []string{
        "sorted(array)",
        "len(array) > 0",
    },
    Postconditions: []string{
        "result == -1 || array[result] == target",
    },
    Invariants: []string{
        "low <= high",
        "low >= 0 && high < len(array)",
    },
}
```

## Configuration

```go
type VerifierConfig struct {
    Z3Path         string        // Path to Z3 binary
    DafnyPath      string        // Path to Dafny binary
    Timeout        time.Duration // Verification timeout
    MaxProofDepth  int           // Maximum proof depth
    EnableCaching  bool          // Cache verification results
    ParallelProofs int           // Parallel proof attempts
}
```

## Usage Examples

### Verify Function Contract

```go
result, err := verifier.VerifyContract(ctx, &verification.ContractRequest{
    Function: "Sort",
    Code:     sortCode,
    Pre:      "len(arr) > 0",
    Post:     "sorted(arr)",
})

if result.Verified {
    fmt.Println("Contract verified!")
} else {
    fmt.Printf("Counterexample: %v\n", result.Counterexample)
}
```

### Verify Plan Safety

```go
plan := &verification.Plan{
    Steps: []verification.Step{
        {Action: "move", Args: []string{"robot", "roomA"}},
        {Action: "pickup", Args: []string{"item"}},
        {Action: "move", Args: []string{"robot", "roomB"}},
    },
}

result, err := veriplan.VerifyPlan(ctx, plan, []verification.Property{
    verification.Safety("no_collision"),
    verification.Liveness("goal_reached"),
})
```

## Error Handling

```go
result, err := verifier.Verify(ctx, request)
if err != nil {
    switch err.(type) {
    case *verification.TimeoutError:
        // Handle timeout
    case *verification.ProverError:
        // Handle prover failure
    default:
        // Handle other errors
    }
}
```

## Testing

```bash
# Run all verification tests
go test -v ./internal/verification/...

# Test Z3 integration (requires Z3)
Z3_PATH=/usr/bin/z3 go test -v -run TestZ3 ./internal/verification/

# Test Dafny integration (requires Dafny)
DAFNY_PATH=/usr/bin/dafny go test -v -run TestDafny ./internal/verification/
```

## Dependencies

- **Z3**: SMT solver (optional)
- **Dafny**: Verified programming language (optional)

## See Also

- [Z3 Theorem Prover](https://github.com/Z3Prover/z3)
- [Dafny Language](https://github.com/dafny-lang/dafny)
- `internal/planning/` - Plan generation
