#!/bin/bash
# Generate Software Bill of Materials (SBOM) for HelixAgent
# Supports CycloneDX and SPDX formats via syft or cyclonedx-gomod

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="${PROJECT_ROOT}/reports/sbom"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "$REPORTS_DIR"

# Color output helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Detect available SBOM tools
SYFT_CMD=""
CYCLONEDX_CMD=""

if command -v syft &>/dev/null; then
    SYFT_CMD="syft"
    info "Found syft: $(syft version 2>/dev/null | head -1)"
fi

if command -v cyclonedx-gomod &>/dev/null; then
    CYCLONEDX_CMD="cyclonedx-gomod"
    info "Found cyclonedx-gomod"
fi

if [ -z "$SYFT_CMD" ] && [ -z "$CYCLONEDX_CMD" ]; then
    warn "No SBOM tool found. Generating go mod-based SBOM."
    info "For full SBOM, install: go install github.com/anchore/syft/cmd/syft@latest"
    info "  or: go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest"

    # Fallback: generate dependency list from go mod
    info "Generating dependency list from go.mod..."
    cd "$PROJECT_ROOT"
    go list -m -json all > "${REPORTS_DIR}/go-modules-${TIMESTAMP}.json" 2>/dev/null || true
    go mod graph > "${REPORTS_DIR}/go-mod-graph-${TIMESTAMP}.txt" 2>/dev/null || true
    info "Dependency list saved to ${REPORTS_DIR}/"
    exit 0
fi

cd "$PROJECT_ROOT"

# Generate SBOM with syft (preferred)
if [ -n "$SYFT_CMD" ]; then
    info "Generating CycloneDX SBOM with syft..."
    $SYFT_CMD dir:. -o cyclonedx-json > "${REPORTS_DIR}/sbom-cyclonedx-${TIMESTAMP}.json" 2>/dev/null
    info "CycloneDX SBOM: ${REPORTS_DIR}/sbom-cyclonedx-${TIMESTAMP}.json"

    info "Generating SPDX SBOM with syft..."
    $SYFT_CMD dir:. -o spdx-json > "${REPORTS_DIR}/sbom-spdx-${TIMESTAMP}.json" 2>/dev/null
    info "SPDX SBOM: ${REPORTS_DIR}/sbom-spdx-${TIMESTAMP}.json"

    # Create latest symlinks
    ln -sf "sbom-cyclonedx-${TIMESTAMP}.json" "${REPORTS_DIR}/sbom-cyclonedx-latest.json"
    ln -sf "sbom-spdx-${TIMESTAMP}.json" "${REPORTS_DIR}/sbom-spdx-latest.json"
fi

# Generate SBOM with cyclonedx-gomod
if [ -n "$CYCLONEDX_CMD" ]; then
    info "Generating CycloneDX SBOM with cyclonedx-gomod..."
    $CYCLONEDX_CMD mod -json -output "${REPORTS_DIR}/sbom-gomod-cyclonedx-${TIMESTAMP}.json" 2>/dev/null
    info "CycloneDX Go SBOM: ${REPORTS_DIR}/sbom-gomod-cyclonedx-${TIMESTAMP}.json"
    ln -sf "sbom-gomod-cyclonedx-${TIMESTAMP}.json" "${REPORTS_DIR}/sbom-gomod-cyclonedx-latest.json"
fi

# Summary
info ""
info "SBOM generation complete. Reports in: ${REPORTS_DIR}/"
ls -la "${REPORTS_DIR}/"*"${TIMESTAMP}"* 2>/dev/null || true
