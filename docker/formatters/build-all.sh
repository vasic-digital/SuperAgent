#!/bin/bash
# Build all formatter service Docker containers

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🚀 Building all formatter service containers..."
echo

FORMATTERS=(
    "autopep8"
    "yapf"
    "sqlfluff"
    "rubocop"
    "standardrb"
    "php-cs-fixer"
    "laravel-pint"
    "perltidy"
    "cljfmt"
    "spotless"
    "groovy-lint"
    "styler"
    "air"
    "psscriptanalyzer"
)

TOTAL=${#FORMATTERS[@]}
SUCCESS=0
FAILED=0

for formatter in "${FORMATTERS[@]}"; do
    echo "📦 Building $formatter..."

    if docker build -t "helixagent/formatter-$formatter:latest" -f "Dockerfile.$formatter" .; then
        echo "✅ $formatter built successfully"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "❌ $formatter build failed"
        FAILED=$((FAILED + 1))
    fi

    echo
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Build Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total formatters: $TOTAL"
echo "✅ Successful:    $SUCCESS"
echo "❌ Failed:        $FAILED"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ "$FAILED" -eq 0 ]; then
    echo
    echo "🎉 All formatter services built successfully!"
    echo
    echo "To start services:"
    echo "  docker-compose -f docker-compose.formatters.yml up -d"
    echo
    echo "To test a formatter:"
    echo "  curl -X POST http://localhost:9211/format \\"
    echo "    -H 'Content-Type: application/json' \\"
    echo "    -d '{\"content\":\"def hello(  x,y ):\\n  return x+y\"}'"
    exit 0
else
    echo
    echo "⚠️  Some builds failed. Check the output above for errors."
    exit 1
fi
