#!/bin/bash
# Benchmark runner for all providers
# Usage: ./scripts/benchmark_providers.sh

set -e

cd "$(dirname "$0")/.."

echo "=== HelixAgent Provider Benchmark Runner ==="
echo

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "This will run benchmarks against all configured providers."
echo "Benchmarks make real API calls and may incur costs."
echo "Press Ctrl+C to cancel, or wait 3 seconds..."
sleep 3

echo -e "${GREEN}Running benchmarks...${NC}"
echo

# Run benchmarks
go test -bench=. -benchmem ./tests/benchmarks/... -timeout 600s | tee benchmark_results.txt

echo
echo -e "${GREEN}=== Benchmarks Complete ===${NC}"
echo "Results saved to: benchmark_results.txt"
