# Generate Constitution

A utility binary that generates and synchronizes the project Constitution from code analysis.

## Purpose

The Constitution is a machine-readable definition of project rules, constraints, and conventions. This tool:

1. **Analyzes** the codebase structure and patterns
2. **Generates** a CONSTITUTION.json with discovered rules
3. **Synchronizes** the Constitution to AGENTS.md and CLAUDE.md
4. **Creates** a validation report

## Installation

```bash
# Build from source
go build -o bin/generate-constitution ./cmd/generate-constitution

# Run directly
go run ./cmd/generate-constitution
```

## Usage

### Basic Usage

```bash
# Generate constitution for current project
./bin/generate-constitution

# Specify project root
PROJECT_ROOT=/path/to/project ./bin/generate-constitution
```

## Output Files

| File | Description |
|------|-------------|
| `CONSTITUTION.json` | Machine-readable constitution with all rules |
| `CONSTITUTION.md` | Human-readable markdown version |
| `CONSTITUTION_REPORT.md` | Validation and analysis report |

## Constitution Structure

```json
{
  "version": "1.0.0",
  "updated": "2026-02-25T12:00:00Z",
  "rules": [
    {
      "id": "test-coverage",
      "category": "testing",
      "description": "All code must have 100% test coverage",
      "mandatory": true,
      "priority": 1
    }
  ]
}
```

## Rule Categories

| Category | Description |
|----------|-------------|
| `quality` | Code quality standards |
| `security` | Security requirements |
| `testing` | Test coverage and types |
| `documentation` | Documentation requirements |
| `architecture` | Architectural patterns |
| `containerization` | Container deployment rules |

## Synchronization

The tool automatically updates:
- `AGENTS.md` - Agent instructions section
- `CLAUDE.md` - Claude Code instructions section

Both files contain a `<!-- BEGIN_CONSTITUTION -->` / `<!-- END_CONSTITUTION -->` block that is replaced.

## Integration

The constitution generator is typically run:
1. During project setup
2. Before releases
3. When adding new modules or features
4. By the Constitution Watcher background service

## Related

- `internal/services/constitution_manager.go` - Core logic
- `internal/services/constitution_watcher.go` - Background watcher
- `internal/services/documentation_sync.go` - Sync to docs
