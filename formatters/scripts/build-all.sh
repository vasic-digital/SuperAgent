#!/bin/bash
# Build all native binary formatters

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FORMATTERS_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$FORMATTERS_DIR")"
BIN_DIR="$PROJECT_ROOT/bin/formatters"

mkdir -p "$BIN_DIR"

echo "=== Building Native Binary Formatters ==="
echo "Output directory: $BIN_DIR"
echo

BUILT=0
FAILED=0
SKIPPED=0

# Function to build a formatter
build_formatter() {
    local name=$1
    local path=$2
    local build_cmd=$3

    echo "Building $name..."

    if [ ! -d "$path" ]; then
        echo "  ⚠ SKIP: Directory not found: $path"
        SKIPPED=$((SKIPPED+1))
        return
    fi

    cd "$path"

    if eval "$build_cmd" > /dev/null 2>&1; then
        echo "  ✓ SUCCESS"
        BUILT=$((BUILT+1))
    else
        echo "  ✗ FAILED"
        FAILED=$((FAILED+1))
    fi

    cd "$PROJECT_ROOT"
}

# Rust formatters
build_formatter "rustfmt" "$FORMATTERS_DIR/rustfmt" "cargo build --release && cp target/release/rustfmt $BIN_DIR/"
build_formatter "ruff" "$FORMATTERS_DIR/ruff" "cargo build --release && cp target/release/ruff $BIN_DIR/"
build_formatter "biome" "$FORMATTERS_DIR/biome" "cargo build --release && cp target/release/biome $BIN_DIR/"
build_formatter "dprint" "$FORMATTERS_DIR/dprint" "cargo build --release && cp target/release/dprint $BIN_DIR/"
build_formatter "taplo" "$FORMATTERS_DIR/taplo" "cargo build --release && cp target/release/taplo $BIN_DIR/"
build_formatter "stylua" "$FORMATTERS_DIR/stylua" "cargo build --release && cp target/release/stylua $BIN_DIR/"

# Go formatters
build_formatter "shfmt" "$FORMATTERS_DIR/shfmt" "go build -o $BIN_DIR/shfmt ./cmd/shfmt"
build_formatter "yamlfmt" "$FORMATTERS_DIR/yamlfmt" "go build -o $BIN_DIR/yamlfmt ./cmd/yamlfmt"
build_formatter "buf" "$FORMATTERS_DIR/buf" "go build -o $BIN_DIR/buf ./cmd/buf"

# Node.js formatters
build_formatter "prettier" "$FORMATTERS_DIR/prettier" "npm install && npm run build && ln -sf $(pwd)/bin/prettier.cjs $BIN_DIR/prettier"

# Python formatters
build_formatter "black" "$FORMATTERS_DIR/black" "pip install -e . && ln -sf $(which black) $BIN_DIR/black"

# C/C++ formatters
build_formatter "clang-format" "$FORMATTERS_DIR/clang-format" "cmake -B build -DLLVM_ENABLE_PROJECTS=clang && cmake --build build --target clang-format && cp build/bin/clang-format $BIN_DIR/"

# JVM formatters
build_formatter "google-java-format" "$FORMATTERS_DIR/google-java-format" "mvn package && cp core/target/google-java-format-*-all-deps.jar $BIN_DIR/google-java-format.jar"
build_formatter "ktlint" "$FORMATTERS_DIR/ktlint" "./gradlew shadowJarExecutable && cp ktlint/build/run/ktlint $BIN_DIR/"
build_formatter "scalafmt" "$FORMATTERS_DIR/scalafmt" "sbt assembly && cp scalafmt-cli/target/scala-*/scalafmt.jar $BIN_DIR/"

# Haskell formatters
build_formatter "ormolu" "$FORMATTERS_DIR/ormolu" "stack install && cp ~/.local/bin/ormolu $BIN_DIR/"

# OCaml formatters
build_formatter "ocamlformat" "$FORMATTERS_DIR/ocamlformat" "opam install . && ln -sf $(which ocamlformat) $BIN_DIR/ocamlformat"

# F# formatters
build_formatter "fantomas" "$FORMATTERS_DIR/fantomas" "dotnet build && dotnet pack && dotnet tool install -g fantomas && ln -sf ~/.dotnet/tools/fantomas $BIN_DIR/fantomas"

echo
echo "=== Build Summary ==="
echo "Built:   $BUILT"
echo "Failed:  $FAILED"
echo "Skipped: $SKIPPED"
echo "Total:   $((BUILT+FAILED+SKIPPED))"
echo
echo "Binaries installed to: $BIN_DIR"

if [ $FAILED -gt 0 ]; then
    echo "⚠ Some formatters failed to build"
    exit 1
fi

echo "✓ All available formatters built successfully"
