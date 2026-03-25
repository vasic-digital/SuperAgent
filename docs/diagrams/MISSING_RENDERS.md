# Missing Rendered Diagrams

The following PlantUML sources in `docs/diagrams/src/` do not have a
corresponding rendered SVG in `docs/diagrams/output/svg/`.

PlantUML rendering requires a running PlantUML server or Docker container
(`plantuml/plantuml-server`). Renders cannot be produced without Docker.

## Missing

| Source file | Expected output |
|-------------|-----------------|
| `src/architecture.puml` | `output/svg/architecture.svg` |
| `src/debate-orchestration-flow.puml` | `output/svg/debate-orchestration-flow.svg` |
| `src/goroutine-lifecycle.puml` | `output/svg/goroutine-lifecycle.svg` |
| `src/lazy-loading-architecture.puml` | `output/svg/lazy-loading-architecture.svg` |
| `src/module-dependency-graph.puml` | `output/svg/module-dependency-graph.svg` |
| `src/security-scanning-pipeline.puml` | `output/svg/security-scanning-pipeline.svg` |
| `src/test-pyramid.puml` | `output/svg/test-pyramid.svg` |

## Already Rendered

| Source file | Output |
|-------------|--------|
| `src/boot-sequence.puml` | `output/svg/boot-sequence.svg` |
| `src/database-er.puml` | `output/svg/database-er.svg` |

Note: `output/svg/` also contains `architecture-overview.svg`, `data-flow.svg`,
`debate-system.svg`, `service-dependencies.svg`, and `shutdown-sequence.svg`
which have no corresponding `.puml` source (likely rendered from external tools
or earlier source files that have since been renamed).

## How to Render

```bash
# Using Docker (requires Docker/Podman running)
docker run --rm -v "$(pwd)/docs/diagrams:/diagrams" plantuml/plantuml-server \
  -tsvg /diagrams/src/*.puml -o /diagrams/output/svg/
```
