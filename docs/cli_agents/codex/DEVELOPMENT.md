# OpenAI Codex - Development Guide

## Building from Source

### Prerequisites

- Rust 1.70+ (for codex-rs)
- Node.js 16+ (for codex-cli)
- Bazel (optional, for full build)
- Just (task runner)

### Build Rust Implementation

```bash
cd codex-rs
cargo build --release
```

### Build with Bazel

```bash
# Full workspace
bazel build //...

# Specific target
bazel build //codex-rs:codex
```

## Development Workflow

### Rust Conventions

- Use `just fmt` before committing
- Run `cargo test -p <crate>` for specific tests
- Use `just fix -p <crate>` for linting
- Keep modules under 500 LoC
- Add snapshot tests for UI changes

### Testing

```bash
# Unit tests
cargo test

# With nextest
cargo nextest run

# Snapshot tests
cargo insta review
```

## Contributing

1. Create topic branch: `git checkout -b feat/feature-name`
2. Make focused changes
3. Run tests: `cargo test`
4. Sign CLA in PR
5. Submit for review

---

*See [API Reference](./API.md) for usage*
