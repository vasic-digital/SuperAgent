#!/usr/bin/env bash
set -e

# HelixAgent Diagram Generation Script
# Converts .mmd (Mermaid) and .puml (PlantUML) source files to SVG, PNG, and PDF.
#
# Usage:
#   ./scripts/generate-diagrams.sh
#
# Requirements (optional - script degrades gracefully):
#   - mmdc (mermaid-cli): npm install -g @mermaid-js/mermaid-cli
#   - plantuml: https://plantuml.com/download (requires Java)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

SRC_DIR="${PROJECT_ROOT}/docs/diagrams/src"
OUT_SVG="${PROJECT_ROOT}/docs/diagrams/output/svg"
OUT_PNG="${PROJECT_ROOT}/docs/diagrams/output/png"
OUT_PDF="${PROJECT_ROOT}/docs/diagrams/output/pdf"

# Counters
TOTAL_GENERATED=0
TOTAL_SKIPPED=0
TOTAL_FAILED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info()  { echo -e "${BLUE}[INFO]${NC}  $*"; }
log_ok()    { echo -e "${GREEN}[OK]${NC}    $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# Ensure output directories exist
mkdir -p "${OUT_SVG}" "${OUT_PNG}" "${OUT_PDF}"

# Check for source directory
if [ ! -d "${SRC_DIR}" ]; then
    log_error "Source directory not found: ${SRC_DIR}"
    exit 1
fi

# Detect available tools
HAS_MMDC=false
HAS_PLANTUML=false
PLANTUML_CMD=""

if command -v mmdc &>/dev/null; then
    HAS_MMDC=true
    log_info "Found mmdc (mermaid-cli): $(mmdc --version 2>/dev/null || echo 'version unknown')"
else
    log_warn "mmdc (mermaid-cli) not found. Mermaid diagrams will be skipped."
    log_warn "Install with: npm install -g @mermaid-js/mermaid-cli"
fi

if command -v plantuml &>/dev/null; then
    HAS_PLANTUML=true
    PLANTUML_CMD="plantuml"
    log_info "Found plantuml command"
elif [ -n "${PLANTUML_JAR:-}" ] && [ -f "${PLANTUML_JAR}" ]; then
    HAS_PLANTUML=true
    PLANTUML_CMD="java -jar ${PLANTUML_JAR}"
    log_info "Found PlantUML JAR: ${PLANTUML_JAR}"
else
    log_warn "plantuml not found. PlantUML diagrams will be skipped."
    log_warn "Install with your package manager or set PLANTUML_JAR=/path/to/plantuml.jar"
fi

echo ""
log_info "Source directory: ${SRC_DIR}"
log_info "Output directory: ${PROJECT_ROOT}/docs/diagrams/output/"
echo ""

# ============================================================
# Process Mermaid (.mmd) files
# ============================================================
MMD_FILES=()
while IFS= read -r -d '' f; do
    MMD_FILES+=("$f")
done < <(find "${SRC_DIR}" -name '*.mmd' -type f -print0 2>/dev/null)

if [ ${#MMD_FILES[@]} -gt 0 ]; then
    log_info "Found ${#MMD_FILES[@]} Mermaid file(s)"
    if [ "${HAS_MMDC}" = true ]; then
        for src_file in "${MMD_FILES[@]}"; do
            basename="$(basename "${src_file}" .mmd)"
            log_info "Processing Mermaid: ${basename}.mmd"

            # SVG
            if mmdc -i "${src_file}" -o "${OUT_SVG}/${basename}.svg" -t default 2>/dev/null; then
                log_ok "  -> ${OUT_SVG}/${basename}.svg"
                TOTAL_GENERATED=$((TOTAL_GENERATED + 1))
            else
                log_error "  Failed to generate SVG for ${basename}"
                TOTAL_FAILED=$((TOTAL_FAILED + 1))
            fi

            # PNG
            if mmdc -i "${src_file}" -o "${OUT_PNG}/${basename}.png" -t default -s 2 2>/dev/null; then
                log_ok "  -> ${OUT_PNG}/${basename}.png"
                TOTAL_GENERATED=$((TOTAL_GENERATED + 1))
            else
                log_error "  Failed to generate PNG for ${basename}"
                TOTAL_FAILED=$((TOTAL_FAILED + 1))
            fi

            # PDF
            if mmdc -i "${src_file}" -o "${OUT_PDF}/${basename}.pdf" -t default 2>/dev/null; then
                log_ok "  -> ${OUT_PDF}/${basename}.pdf"
                TOTAL_GENERATED=$((TOTAL_GENERATED + 1))
            else
                log_error "  Failed to generate PDF for ${basename}"
                TOTAL_FAILED=$((TOTAL_FAILED + 1))
            fi
        done
    else
        TOTAL_SKIPPED=$((TOTAL_SKIPPED + ${#MMD_FILES[@]}))
        log_warn "Skipping ${#MMD_FILES[@]} Mermaid file(s) (mmdc not available)"
    fi
else
    log_info "No Mermaid (.mmd) files found"
fi

echo ""

# ============================================================
# Process PlantUML (.puml) files
# ============================================================
PUML_FILES=()
while IFS= read -r -d '' f; do
    PUML_FILES+=("$f")
done < <(find "${SRC_DIR}" -name '*.puml' -type f -print0 2>/dev/null)

if [ ${#PUML_FILES[@]} -gt 0 ]; then
    log_info "Found ${#PUML_FILES[@]} PlantUML file(s)"
    if [ "${HAS_PLANTUML}" = true ]; then
        for src_file in "${PUML_FILES[@]}"; do
            basename="$(basename "${src_file}" .puml)"
            log_info "Processing PlantUML: ${basename}.puml"

            # SVG
            if ${PLANTUML_CMD} -tsvg -o "${OUT_SVG}" "${src_file}" 2>/dev/null; then
                log_ok "  -> ${OUT_SVG}/${basename}.svg"
                TOTAL_GENERATED=$((TOTAL_GENERATED + 1))
            else
                log_error "  Failed to generate SVG for ${basename}"
                TOTAL_FAILED=$((TOTAL_FAILED + 1))
            fi

            # PNG
            if ${PLANTUML_CMD} -tpng -o "${OUT_PNG}" "${src_file}" 2>/dev/null; then
                log_ok "  -> ${OUT_PNG}/${basename}.png"
                TOTAL_GENERATED=$((TOTAL_GENERATED + 1))
            else
                log_error "  Failed to generate PNG for ${basename}"
                TOTAL_FAILED=$((TOTAL_FAILED + 1))
            fi

            # PDF
            if ${PLANTUML_CMD} -tpdf -o "${OUT_PDF}" "${src_file}" 2>/dev/null; then
                log_ok "  -> ${OUT_PDF}/${basename}.pdf"
                TOTAL_GENERATED=$((TOTAL_GENERATED + 1))
            else
                log_error "  Failed to generate PDF for ${basename}"
                TOTAL_FAILED=$((TOTAL_FAILED + 1))
            fi
        done
    else
        TOTAL_SKIPPED=$((TOTAL_SKIPPED + ${#PUML_FILES[@]}))
        log_warn "Skipping ${#PUML_FILES[@]} PlantUML file(s) (plantuml not available)"
    fi
else
    log_info "No PlantUML (.puml) files found"
fi

# ============================================================
# Summary
# ============================================================
echo ""
echo "========================================"
echo "  Diagram Generation Summary"
echo "========================================"
echo -e "  Source files (Mermaid) : ${#MMD_FILES[@]}"
echo -e "  Source files (PlantUML): ${#PUML_FILES[@]}"
echo -e "  ${GREEN}Generated${NC}             : ${TOTAL_GENERATED}"
echo -e "  ${YELLOW}Skipped${NC}               : ${TOTAL_SKIPPED}"
echo -e "  ${RED}Failed${NC}                : ${TOTAL_FAILED}"
echo "========================================"
echo ""

if [ "${TOTAL_GENERATED}" -gt 0 ]; then
    log_info "Output files:"
    for dir in "${OUT_SVG}" "${OUT_PNG}" "${OUT_PDF}"; do
        count=$(find "${dir}" -type f 2>/dev/null | wc -l)
        if [ "${count}" -gt 0 ]; then
            log_info "  ${dir}/ (${count} files)"
        fi
    done
fi

if [ "${TOTAL_FAILED}" -gt 0 ]; then
    exit 1
fi

exit 0
