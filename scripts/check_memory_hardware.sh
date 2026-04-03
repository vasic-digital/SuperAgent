#!/bin/bash
# Hardware detection script for HelixMemory containers
# Determines if local Cognee, Mem0, and Letta can run or if cloud fallback is needed

set -e

echo "=== HelixMemory Hardware Detection ==="
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Minimum requirements
MIN_CPU_CORES=4
MIN_MEMORY_GB=8
MIN_DISK_GB=20

# Recommended for all 3 services
REC_CPU_CORES=8
REC_MEMORY_GB=16
REC_DISK_GB=50

# Detection results
CAN_RUN_LOCAL=true
RECOMMEND_CLOUD=false

# Function to check command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if running in container (we can't run nested containers easily)
if [ -f /.dockerenv ] || grep -q docker /proc/1/cgroup 2>/dev/null; then
    echo -e "${YELLOW}⚠ Running inside a container - local memory services not recommended${NC}"
    CAN_RUN_LOCAL=false
    RECOMMEND_CLOUD=true
fi

# Detect OS
OS=$(uname -s)
ARCH=$(uname -m)
echo -e "${BLUE}OS:${NC} $OS ($ARCH)"

# Detect CPU
echo
echo "=== CPU Detection ==="

if command_exists nproc; then
    CPU_CORES=$(nproc)
elif command_exists sysctl; then
    CPU_CORES=$(sysctl -n hw.ncpu 2>/dev/null || echo "unknown")
else
    CPU_CORES=$(grep -c ^processor /proc/cpuinfo 2>/dev/null || echo "unknown")
fi

echo -e "${BLUE}CPU Cores:${NC} $CPU_CORES"

if [ "$CPU_CORES" != "unknown" ]; then
    if [ "$CPU_CORES" -lt "$MIN_CPU_CORES" ]; then
        echo -e "${RED}✗ Insufficient CPU cores (minimum: $MIN_CPU_CORES)${NC}"
        CAN_RUN_LOCAL=false
        RECOMMEND_CLOUD=true
    elif [ "$CPU_CORES" -lt "$REC_CPU_CORES" ]; then
        echo -e "${YELLOW}⚠ Below recommended CPU cores (recommended: $REC_CPU_CORES)${NC}"
        RECOMMEND_CLOUD=true
    else
        echo -e "${GREEN}✓ Sufficient CPU cores${NC}"
    fi
fi

# Detect Memory
echo
echo "=== Memory Detection ==="

if command_exists free; then
    # Linux
    MEMORY_KB=$(free -k | awk '/^Mem:/ {print $2}')
    MEMORY_GB=$((MEMORY_KB / 1024 / 1024))
elif command_exists vm_stat; then
    # macOS
    MEMORY_BYTES=$(vm_stat | grep "Pages free" | awk '{print $3}' | sed 's/\.//')
    # Rough estimation
    MEMORY_GB=$(sysctl -n hw.memsize 2>/dev/null | awk '{print int($1/1024/1024/1024)}')
else
    MEMORY_GB="unknown"
fi

echo -e "${BLUE}Memory:${NC} ${MEMORY_GB}GB"

if [ "$MEMORY_GB" != "unknown" ]; then
    if [ "$MEMORY_GB" -lt "$MIN_MEMORY_GB" ]; then
        echo -e "${RED}✗ Insufficient memory (minimum: ${MIN_MEMORY_GB}GB)${NC}"
        CAN_RUN_LOCAL=false
        RECOMMEND_CLOUD=true
    elif [ "$MEMORY_GB" -lt "$REC_MEMORY_GB" ]; then
        echo -e "${YELLOW}⚠ Below recommended memory (recommended: ${REC_MEMORY_GB}GB)${NC}"
        RECOMMEND_CLOUD=true
    else
        echo -e "${GREEN}✓ Sufficient memory${NC}"
    fi
fi

# Detect Disk Space
echo
echo "=== Disk Space Detection ==="

if command_exists df; then
    DISK_KB=$(df -k . | tail -1 | awk '{print $4}')
    DISK_GB=$((DISK_KB / 1024 / 1024))
else
    DISK_GB="unknown"
fi

echo -e "${BLUE}Available Disk:${NC} ${DISK_GB}GB"

if [ "$DISK_GB" != "unknown" ]; then
    if [ "$DISK_GB" -lt "$MIN_DISK_GB" ]; then
        echo -e "${RED}✗ Insufficient disk space (minimum: ${MIN_DISK_GB}GB)${NC}"
        CAN_RUN_LOCAL=false
        RECOMMEND_CLOUD=true
    elif [ "$DISK_GB" -lt "$REC_DISK_GB" ]; then
        echo -e "${YELLOW}⚠ Below recommended disk space (recommended: ${REC_DISK_GB}GB)${NC}"
        RECOMMEND_CLOUD=true
    else
        echo -e "${GREEN}✓ Sufficient disk space${NC}"
    fi
fi

# Check Docker availability
echo
echo "=== Container Runtime Detection ==="

if command_exists docker; then
    if docker info >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Docker is available${NC}"
        
        # Check Docker Compose
        if command_exists docker-compose || docker compose version >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Docker Compose is available${NC}"
        else
            echo -e "${YELLOW}⚠ Docker Compose not found${NC}"
            CAN_RUN_LOCAL=false
        fi
    else
        echo -e "${YELLOW}⚠ Docker daemon not running${NC}"
        CAN_RUN_LOCAL=false
    fi
else
    echo -e "${YELLOW}⚠ Docker not installed${NC}"
    CAN_RUN_LOCAL=false
fi

# Check for GPU (optional but recommended)
echo
echo "=== GPU Detection (Optional) ==="

if command_exists nvidia-smi; then
    if nvidia-smi >/dev/null 2>&1; then
        GPU_INFO=$(nvidia-smi --query-gpu=name,memory.total --format=csv,noheader 2>/dev/null | head -1)
        echo -e "${GREEN}✓ NVIDIA GPU detected:${NC} $GPU_INFO"
    else
        echo -e "${YELLOW}⚠ NVIDIA driver not available${NC}"
    fi
elif [ -d /dev/dri ]; then
    echo -e "${GREEN}✓ GPU device nodes present${NC}"
else
    echo -e "${BLUE}ℹ No GPU detected (CPU-only mode)${NC}"
fi

# Resource estimation for memory services
echo
echo "=== HelixMemory Resource Requirements ==="
echo "┌─────────────────┬─────────────┬─────────────┐"
echo "│ Service         │ Min Memory  │ Recommended │"
echo "├─────────────────┼─────────────┼─────────────┤"
echo "│ Cognee          │ 2GB         │ 4GB         │"
echo "│ Mem0            │ 1GB         │ 2GB         │"
echo "│ Letta           │ 1GB         │ 2GB         │"
echo "│ Qdrant          │ 1GB         │ 2GB         │"
echo "│ Neo4j           │ 2GB         │ 4GB         │"
echo "│ Redis           │ 256MB       │ 512MB       │"
echo "├─────────────────┼─────────────┼─────────────┤"
echo "│ Total           │ ~6.5GB      │ ~14.5GB     │"
echo "└─────────────────┴─────────────┴─────────────┘"

# Final recommendation
echo
echo "=== Recommendation ==="

if [ "$CAN_RUN_LOCAL" = true ] && [ "$RECOMMEND_CLOUD" = false ]; then
    echo -e "${GREEN}✓ System can run all HelixMemory services locally${NC}"
    echo "  Set HELIX_MEMORY_MODE=local in your .env file"
    exit 0
elif [ "$CAN_RUN_LOCAL" = true ] && [ "$RECOMMEND_CLOUD" = true ]; then
    echo -e "${YELLOW}⚠ System can run locally but performance may be limited${NC}"
    echo "  Options:"
    echo "    1. Run locally: HELIX_MEMORY_MODE=local"
    echo "    2. Use cloud APIs: HELIX_MEMORY_MODE=cloud"
    echo "    3. Hybrid: Run only essential services locally"
    exit 1
else
    echo -e "${RED}✗ System cannot run all services locally${NC}"
    echo "  Set HELIX_MEMORY_MODE=cloud in your .env file"
    echo "  and configure API keys for:"
    echo "    - Cognee: https://platform.cognee.ai"
    echo "    - Mem0: https://app.mem0.ai"
    echo "    - Letta: https://docs.letta.com"
    exit 2
fi
