#!/usr/bin/env bash
set -euo pipefail

# CI Entrypoint - Resource detection and environment setup
# Reads CI_RESOURCE_LIMIT (low|medium|high) and configures limits accordingly

CI_RESOURCE_LIMIT="${CI_RESOURCE_LIMIT:-low}"

# Detect host resources
TOTAL_CPUS=$(nproc 2>/dev/null || echo 4)
TOTAL_MEM_KB=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}' || echo 8388608)
TOTAL_MEM_MB=$((TOTAL_MEM_KB / 1024))

case "${CI_RESOURCE_LIMIT}" in
  low)
    CPU_PERCENT=30
    MEM_PERCENT=30
    export GOMAXPROCS=2
    NICE_LEVEL=19
    IONICE_CLASS=3
    ;;
  medium)
    CPU_PERCENT=50
    MEM_PERCENT=50
    export GOMAXPROCS=4
    NICE_LEVEL=10
    IONICE_CLASS=2
    ;;
  high)
    CPU_PERCENT=70
    MEM_PERCENT=70
    export GOMAXPROCS=6
    NICE_LEVEL=5
    IONICE_CLASS=2
    ;;
  *)
    echo "ERROR: Invalid CI_RESOURCE_LIMIT: ${CI_RESOURCE_LIMIT} (must be low|medium|high)"
    exit 1
    ;;
esac

ALLOWED_CPUS=$(( TOTAL_CPUS * CPU_PERCENT / 100 ))
[ "${ALLOWED_CPUS}" -lt 1 ] && ALLOWED_CPUS=1
ALLOWED_MEM_MB=$(( TOTAL_MEM_MB * MEM_PERCENT / 100 ))

export CI_ALLOWED_CPUS="${ALLOWED_CPUS}"
export CI_ALLOWED_MEM_MB="${ALLOWED_MEM_MB}"
export CI_NICE_LEVEL="${NICE_LEVEL}"
export CI_IONICE_CLASS="${IONICE_CLASS}"
export GOFLAGS="${GOFLAGS:-} -p=${ALLOWED_CPUS}"

echo "========================================"
echo "CI Resource Configuration"
echo "========================================"
echo "Resource Limit:  ${CI_RESOURCE_LIMIT}"
echo "Host CPUs:       ${TOTAL_CPUS}"
echo "Host Memory:     ${TOTAL_MEM_MB} MB"
echo "Allowed CPUs:    ${ALLOWED_CPUS} (${CPU_PERCENT}%)"
echo "Allowed Memory:  ${ALLOWED_MEM_MB} MB (${MEM_PERCENT}%)"
echo "GOMAXPROCS:      ${GOMAXPROCS}"
echo "Nice Level:      ${NICE_LEVEL}"
echo "IOnice Class:    ${IONICE_CLASS}"
echo "========================================"

# Execute the phase script passed as argument
if [ $# -gt 0 ]; then
  exec nice -n "${NICE_LEVEL}" ionice -c "${IONICE_CLASS}" "$@"
else
  echo "ERROR: No phase script specified"
  exit 1
fi
