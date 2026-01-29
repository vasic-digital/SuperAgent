# Formatters Directory

This directory contains all code formatters integrated into HelixAgent as Git submodules.

## Overview

- **Total Formatters**: 118+
- **Native Binaries**: 60+
- **Service-Based**: 20+
- **Built-in**: 15+
- **Unified**: 3

## Directory Structure

```
formatters/
├── README.md                    # This file
├── VERSIONS.yaml                # Pinned versions for all formatters
├── scripts/                     # Management scripts
│   ├── init-submodules.sh       # Initialize all submodules
│   ├── update-all.sh            # Update all submodules
│   ├── build-all.sh             # Build all native binaries
│   ├── pin-versions.sh          # Pin submodules to specific versions
│   └── health-check-all.sh      # Health check all formatters
│
└── <118 formatter submodules>   # Git submodules
    ├── clang-format/
    ├── rustfmt/
    ├── prettier/
    ├── black/
    ├── ruff/
    └── ... (118 total)
```

## Quick Start

### 1. Initialize All Submodules

```bash
./formatters/scripts/init-submodules.sh
```

This will:
- Add all 118 formatters as Git submodules
- Initialize and update all submodules recursively
- Clone all formatter repositories

**Note**: This operation requires internet access and may take 30-60 minutes depending on network speed.

### 2. Pin Versions

```bash
./formatters/scripts/pin-versions.sh
```

This will:
- Checkout each submodule to its pinned version (from VERSIONS.yaml)
- Ensure reproducible builds

### 3. Build Native Binaries

```bash
./formatters/scripts/build-all.sh
```

This will:
- Build all native binary formatters
- Install binaries to `bin/formatters/`
- Verify installations

**Note**: Requires build tools (cargo, go, npm, etc.) to be installed.

### 4. Health Check

```bash
./formatters/scripts/health-check-all.sh
```

This will:
- Verify all formatters are installed correctly
- Check versions
- Test basic functionality

## Management Scripts

### init-submodules.sh

Initialize all formatter Git submodules.

```bash
./formatters/scripts/init-submodules.sh
```

### update-all.sh

Update all submodules to latest versions.

```bash
./formatters/scripts/update-all.sh
```

### build-all.sh

Build all native binary formatters.

```bash
# Build all
./formatters/scripts/build-all.sh

# Build specific category
./formatters/scripts/build-all.sh --category native

# Build specific formatter
./formatters/scripts/build-all.sh --formatter rustfmt
```

### pin-versions.sh

Pin all submodules to versions specified in VERSIONS.yaml.

```bash
./formatters/scripts/pin-versions.sh
```

### health-check-all.sh

Perform health checks on all formatters.

```bash
# Check all
./formatters/scripts/health-check-all.sh

# Check specific formatter
./formatters/scripts/health-check-all.sh rustfmt
```

## Formatter Categories

### Native Binary Formatters (60+)

Standalone compiled binaries that run directly on the system.

**Systems Languages**:
- clang-format (C/C++/ObjectiveC)
- rustfmt (Rust)
- gofmt, goimports (Go)
- zig fmt (Zig)
- nimpretty (Nim)

**JVM Languages**:
- google-java-format (Java)
- ktlint, ktfmt (Kotlin)
- scalafmt (Scala)
- spotless (Multi-language)
- cljfmt, zprint (Clojure)

**Web Languages**:
- prettier (JS/TS/HTML/CSS/etc.)
- biome (JS/TS - 35x faster than Prettier)
- dprint (Pluggable)
- black, ruff (Python - Ruff 30x faster than Black)

**Mobile**:
- swift-format, SwiftFormat (Swift)
- dart format (Dart/Flutter)

**Scripting**:
- shfmt (Bash/Shell)
- stylua (Lua)
- yamlfmt (YAML)
- taplo (TOML)
- buf (Protobuf)

### Service-Based Formatters (20+)

Formatters that run as HTTP/RPC services in Docker containers.

**Languages**:
- sqlfluff (SQL multi-dialect)
- rubocop, standardrb (Ruby)
- php-cs-fixer, laravel-pint (PHP)
- npm-groovy-lint (Groovy)
- autopep8, yapf (Python)
- styler, air (R)
- psscriptanalyzer (PowerShell)
- perltidy (Perl)
- fantomas (F#)
- ocamlformat (OCaml)

**Services run on ports 9201-9215** (see `docker/formatters/docker-compose.formatters.yml`).

### Built-in Formatters (15+)

Formatters built into language toolchains.

- gofmt, goimports (Go)
- zig fmt (Zig)
- dart format (Dart)
- mix format (Elixir)
- terraform fmt (Terraform)
- rustfmt (Rust)
- nimpretty (Nim)
- swift-format (Swift 6+)

### Unified Formatters (3)

Multi-language formatters.

- prettier (10+ languages)
- dprint (pluggable platform)
- editorconfig (editor settings, not a formatter)

## Version Management

All formatters are pinned to specific versions in `VERSIONS.yaml`. This ensures:
- **Reproducible builds**: Same versions across all environments
- **Stability**: No unexpected behavior from updates
- **Security**: Known versions with no surprises

### Updating Versions

1. Edit `VERSIONS.yaml`
2. Run `./formatters/scripts/pin-versions.sh`
3. Commit changes

## Build Requirements

### Native Binaries

- **Rust**: cargo (for rustfmt, ruff, biome, dprint, taplo, stylua, air)
- **Go**: go 1.24+ (for gofmt, goimports, shfmt, yamlfmt, buf)
- **Node.js**: npm 20+ (for prettier)
- **Python**: pip (for black, autopep8, yapf, sqlfluff)
- **C/C++**: gcc/clang, cmake (for clang-format, uncrustify)
- **Haskell**: stack/cabal (for ormolu, fourmolu)
- **OCaml**: opam (for ocamlformat)
- **F#**: dotnet (for fantomas)
- **Scala**: sbt (for scalafmt)
- **Java**: maven/gradle (for google-java-format, spotless, ktlint, ktfmt)

### Service-Based

- **Docker/Podman**: For running service containers
- **docker-compose**: For orchestrating services

## Integration

Formatters are integrated into HelixAgent via:

1. **FormatterRegistry** (`internal/formatters/registry.go`)
2. **Provider Implementations** (`internal/formatters/providers/`)
3. **API Endpoints** (`/v1/format`, `/v1/formatters`, etc.)
4. **CLI Agent Configs** (OpenCode, Crush, etc.)
5. **AI Debate System** (Auto-formatting code outputs)

## Testing

Run formatter tests:

```bash
# Unit tests
go test -v ./internal/formatters/...

# Integration tests
go test -v ./tests/integration/formatters_test.go

# Challenge scripts
./challenges/scripts/formatters_native_challenge.sh
./challenges/scripts/formatters_service_challenge.sh
./challenges/scripts/cli_agents_formatters_challenge.sh
```

## Documentation

- **Architecture**: `docs/architecture/FORMATTERS_ARCHITECTURE.md`
- **Catalog**: `docs/CODE_FORMATTERS_CATALOG.md`
- **Progress**: `docs/FORMATTERS_PROGRESS.md`
- **API Reference**: `docs/api/FORMATTERS_API.md` (coming soon)

## Troubleshooting

### Submodule Initialization Fails

```bash
# Reset and retry
git submodule deinit -f .
git submodule update --init --recursive
```

### Build Fails

```bash
# Check build requirements
./formatters/scripts/health-check-all.sh

# Build specific formatter
cd formatters/<formatter-name>
# Follow formatter-specific build instructions
```

### Service Not Starting

```bash
# Check Docker
docker ps

# Restart services
docker-compose -f docker/formatters/docker-compose.formatters.yml restart
```

## Contributing

When adding a new formatter:

1. Add Git submodule: `git submodule add <url> formatters/<name>`
2. Add version to `VERSIONS.yaml`
3. Update `init-submodules.sh`
4. Implement provider in `internal/formatters/providers/`
5. Add tests
6. Update documentation

## License

Each formatter has its own license. See individual formatter repositories for details.

---

**Last Updated**: 2026-01-29
**Total Formatters**: 118+
**Status**: Infrastructure Complete, Submodules Pending
