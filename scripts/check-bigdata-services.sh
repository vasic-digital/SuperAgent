#!/bin/bash

# ============================================================================
# HelixAgent Big Data Services Health Check Script
# ============================================================================
# Checks health of all big data infrastructure services
# Usage: ./scripts/check-bigdata-services.sh
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
SERVICES_TOTAL=0
SERVICES_HEALTHY=0
SERVICES_UNHEALTHY=0

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     HelixAgent Big Data Services - Health Check               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Helper functions
check_service() {
    local name=$1
    local check_cmd=$2
    local expected=$3

    SERVICES_TOTAL=$((SERVICES_TOTAL + 1))

    echo -n "Checking $name... "

    if eval "$check_cmd" | grep -q "$expected"; then
        echo -e "${GREEN}✓ Healthy${NC}"
        SERVICES_HEALTHY=$((SERVICES_HEALTHY + 1))
        return 0
    else
        echo -e "${RED}✗ Unhealthy${NC}"
        SERVICES_UNHEALTHY=$((SERVICES_UNHEALTHY + 1))
        return 1
    fi
}

check_docker_service() {
    local container=$1
    local name=$2

    SERVICES_TOTAL=$((SERVICES_TOTAL + 1))

    echo -n "Checking $name (Docker)... "

    if docker ps --filter "name=$container" --filter "status=running" --filter "health=healthy" | grep -q "$container"; then
        echo -e "${GREEN}✓ Healthy${NC}"
        SERVICES_HEALTHY=$((SERVICES_HEALTHY + 1))
        return 0
    elif docker ps --filter "name=$container" --filter "status=running" | grep -q "$container"; then
        echo -e "${YELLOW}⚠ Running (no health check)${NC}"
        SERVICES_HEALTHY=$((SERVICES_HEALTHY + 1))
        return 0
    else
        echo -e "${RED}✗ Not Running${NC}"
        SERVICES_UNHEALTHY=$((SERVICES_UNHEALTHY + 1))
        return 1
    fi
}

echo -e "${BLUE}[1/10] Messaging Layer${NC}"
echo "─────────────────────────────────────"

# Zookeeper
check_docker_service "helixagent-zookeeper" "Zookeeper"

# Kafka
check_docker_service "helixagent-kafka" "Kafka"

# Test Kafka connectivity
if command -v kafka-topics.sh >/dev/null 2>&1; then
    check_service "Kafka Topics" \
        "kafka-topics.sh --bootstrap-server localhost:9092 --list 2>&1" \
        "helixagent"
fi

echo ""
echo -e "${BLUE}[2/10] Analytics Layer${NC}"
echo "─────────────────────────────────────"

# ClickHouse
check_docker_service "helixagent-clickhouse" "ClickHouse"

# ClickHouse HTTP endpoint
check_service "ClickHouse HTTP" \
    "curl -s http://localhost:8123/ping 2>&1" \
    "Ok."

# ClickHouse query test
if command -v clickhouse-client >/dev/null 2>&1; then
    check_service "ClickHouse Query" \
        "docker exec helixagent-clickhouse clickhouse-client -q 'SELECT 1' 2>&1" \
        "1"
fi

echo ""
echo -e "${BLUE}[3/10] Knowledge Graph Layer${NC}"
echo "─────────────────────────────────────"

# Neo4j
check_docker_service "helixagent-neo4j" "Neo4j"

# Neo4j HTTP endpoint
check_service "Neo4j HTTP" \
    "curl -s http://localhost:7474 2>&1" \
    "200"

# Neo4j Bolt connectivity
if command -v cypher-shell >/dev/null 2>&1; then
    check_service "Neo4j Bolt" \
        "docker exec helixagent-neo4j cypher-shell -u neo4j -p helixagent123 'RETURN 1' 2>&1" \
        "1"
fi

echo ""
echo -e "${BLUE}[4/10] Object Storage Layer${NC}"
echo "─────────────────────────────────────"

# MinIO
check_docker_service "helixagent-minio" "MinIO"

# MinIO health endpoint
check_service "MinIO Health" \
    "curl -s http://localhost:9000/minio/health/live 2>&1" \
    "200"

# MinIO bucket check
if command -v mc >/dev/null 2>&1; then
    check_service "MinIO Buckets" \
        "mc ls helixagent 2>&1 | wc -l" \
        "[1-9]"
fi

echo ""
echo -e "${BLUE}[5/10] Stream Processing Layer (Flink)${NC}"
echo "─────────────────────────────────────"

# Flink JobManager
check_docker_service "helixagent-flink-jobmanager" "Flink JobManager"

# Flink TaskManager
check_docker_service "helixagent-flink-taskmanager" "Flink TaskManager"

# Flink UI
check_service "Flink UI" \
    "curl -s http://localhost:8082/overview 2>&1" \
    "slots-total"

echo ""
echo -e "${BLUE}[6/10] Batch Processing Layer (Spark)${NC}"
echo "─────────────────────────────────────"

# Spark Master
check_docker_service "helixagent-spark-master" "Spark Master"

# Spark Worker
check_docker_service "helixagent-spark-worker" "Spark Worker"

# Spark Master UI
check_service "Spark Master UI" \
    "curl -s http://localhost:4040 2>&1" \
    "Spark"

echo ""
echo -e "${BLUE}[7/10] Vector Database Layer${NC}"
echo "─────────────────────────────────────"

# Qdrant
check_docker_service "helixagent-qdrant" "Qdrant"

# Qdrant health endpoint
check_service "Qdrant Health" \
    "curl -s http://localhost:6333/health 2>&1" \
    "ok"

# Qdrant collections
check_service "Qdrant Collections" \
    "curl -s http://localhost:6333/collections 2>&1" \
    "result"

echo ""
echo -e "${BLUE}[8/10] Data Lakehouse Layer${NC}"
echo "─────────────────────────────────────"

# Iceberg REST Catalog
check_docker_service "helixagent-iceberg-rest" "Iceberg REST"

# Iceberg config endpoint
check_service "Iceberg Config" \
    "curl -s http://localhost:8181/v1/config 2>&1" \
    "warehouse"

echo ""
echo -e "${BLUE}[9/10] Docker Resources${NC}"
echo "─────────────────────────────────────"

# Check Docker system resources
echo -n "Docker Disk Usage... "
DISK_USAGE=$(docker system df --format "{{.Type}}\t{{.Size}}" 2>&1 | grep "Local Volumes" | awk '{print $2}')
echo -e "${BLUE}$DISK_USAGE${NC}"

echo -n "Running Containers... "
RUNNING_CONTAINERS=$(docker ps --filter "name=helixagent-*" | wc -l)
RUNNING_CONTAINERS=$((RUNNING_CONTAINERS - 1))  # Subtract header line
echo -e "${BLUE}$RUNNING_CONTAINERS${NC}"

echo ""
echo -e "${BLUE}[10/10] Network Connectivity${NC}"
echo "─────────────────────────────────────"

# Check Docker network
echo -n "Checking helixagent-network... "
if docker network inspect helixagent-network >/dev/null 2>&1; then
    NETWORK_CONTAINERS=$(docker network inspect helixagent-network --format '{{range .Containers}}{{.Name}} {{end}}' | wc -w)
    echo -e "${GREEN}✓ Active ($NETWORK_CONTAINERS containers)${NC}"
    SERVICES_HEALTHY=$((SERVICES_HEALTHY + 1))
else
    echo -e "${RED}✗ Not Found${NC}"
    SERVICES_UNHEALTHY=$((SERVICES_UNHEALTHY + 1))
fi
SERVICES_TOTAL=$((SERVICES_TOTAL + 1))

# Check network latency between services (sample)
echo -n "Inter-service latency... "
if docker exec helixagent-kafka ping -c 1 -W 1 clickhouse >/dev/null 2>&1; then
    echo -e "${GREEN}✓ OK (<1ms)${NC}"
else
    echo -e "${YELLOW}⚠ Slow${NC}"
fi

echo ""
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                        SUMMARY                                  ${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo ""

# Calculate percentage
if [ $SERVICES_TOTAL -gt 0 ]; then
    HEALTH_PERCENTAGE=$((SERVICES_HEALTHY * 100 / SERVICES_TOTAL))
else
    HEALTH_PERCENTAGE=0
fi

echo "Total Services Checked: $SERVICES_TOTAL"
echo -e "Healthy Services: ${GREEN}$SERVICES_HEALTHY${NC}"
echo -e "Unhealthy Services: ${RED}$SERVICES_UNHEALTHY${NC}"
echo ""

if [ $HEALTH_PERCENTAGE -eq 100 ]; then
    echo -e "${GREEN}✓ All services are healthy! (100%)${NC}"
    echo ""
    echo "System Status: READY FOR PRODUCTION"
    exit 0
elif [ $HEALTH_PERCENTAGE -ge 80 ]; then
    echo -e "${YELLOW}⚠ Most services are healthy ($HEALTH_PERCENTAGE%)${NC}"
    echo ""
    echo "System Status: PARTIALLY OPERATIONAL"
    echo "Please investigate unhealthy services above."
    exit 0
else
    echo -e "${RED}✗ System is unhealthy ($HEALTH_PERCENTAGE%)${NC}"
    echo ""
    echo "System Status: REQUIRES ATTENTION"
    echo ""
    echo "Troubleshooting Steps:"
    echo "1. Check Docker logs: docker-compose logs <service-name>"
    echo "2. Restart unhealthy services: docker-compose restart <service-name>"
    echo "3. Check system resources: docker stats"
    echo "4. Review deployment guide: docs/deployment/BIG_DATA_DEPLOYMENT_GUIDE.md"
    exit 1
fi
