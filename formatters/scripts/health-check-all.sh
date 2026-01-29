#!/bin/bash
# Health check all formatters

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname $(dirname "$SCRIPT_DIR"))"

echo "=== Formatter Health Check ==="
echo

HEALTHY=0
UNHEALTHY=0

# Function to check a formatter
check_formatter() {
    local name=$1
    local cmd=$2

    echo -n "Checking $name... "

    if command -v "$cmd" > /dev/null 2>&1; then
        version=$($cmd --version 2>&1 | head -n1 || echo "unknown")
        echo "✓ HEALTHY ($version)"
        HEALTHY=$((HEALTHY+1))
    else
        echo "✗ UNAVAILABLE"
        UNHEALTHY=$((UNHEALTHY+1))
    fi
}

# Native binaries
echo "=== Native Binary Formatters ==="
check_formatter "clang-format" "clang-format"
check_formatter "rustfmt" "rustfmt"
check_formatter "gofmt" "gofmt"
check_formatter "black" "black"
check_formatter "ruff" "ruff"
check_formatter "prettier" "prettier"
check_formatter "biome" "biome"
check_formatter "dprint" "dprint"
check_formatter "google-java-format" "google-java-format"
check_formatter "ktlint" "ktlint"
check_formatter "scalafmt" "scalafmt"
check_formatter "swift-format" "swift-format"
check_formatter "shfmt" "shfmt"
check_formatter "stylua" "stylua"
check_formatter "yamlfmt" "yamlfmt"
check_formatter "taplo" "taplo"
check_formatter "buf" "buf"
check_formatter "ormolu" "ormolu"
check_formatter "ocamlformat" "ocamlformat"
check_formatter "fantomas" "fantomas"
check_formatter "dart" "dart"
check_formatter "zig" "zig"

echo
echo "=== Service-Based Formatters ==="
echo "Checking Docker services..."

if command -v docker > /dev/null 2>&1; then
    # Check if services are running
    services=(
        "helixagent-formatter-sqlfluff:9201"
        "helixagent-formatter-rubocop:9202"
        "helixagent-formatter-spotless:9203"
        "helixagent-formatter-php-cs-fixer:9204"
    )

    for service in "${services[@]}"; do
        name="${service%%:*}"
        port="${service##*:}"

        echo -n "Checking $name... "

        if docker ps --filter "name=$name" --filter "status=running" | grep -q "$name"; then
            if curl -sf "http://localhost:$port/health" > /dev/null 2>&1; then
                echo "✓ HEALTHY (http://localhost:$port)"
                HEALTHY=$((HEALTHY+1))
            else
                echo "✗ RUNNING but unhealthy"
                UNHEALTHY=$((UNHEALTHY+1))
            fi
        else
            echo "✗ NOT RUNNING"
            UNHEALTHY=$((UNHEALTHY+1))
        fi
    done
else
    echo "⚠ Docker not available - skipping service health checks"
fi

echo
echo "=== Built-in Formatters ==="
check_formatter "gofmt" "gofmt"
check_formatter "goimports" "goimports"
check_formatter "zig fmt" "zig"
check_formatter "dart format" "dart"
check_formatter "mix format" "mix"
check_formatter "terraform fmt" "terraform"

echo
echo "=== Health Summary ==="
echo "Healthy:   $HEALTHY"
echo "Unhealthy: $UNHEALTHY"
echo "Total:     $((HEALTHY+UNHEALTHY))"
echo

if [ $UNHEALTHY -eq 0 ]; then
    echo "✓ All formatters are healthy"
    exit 0
else
    echo "⚠ Some formatters are unavailable"
    exit 1
fi
