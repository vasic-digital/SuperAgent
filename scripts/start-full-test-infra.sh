#!/bin/bash
#
# start-full-test-infra.sh - Starts ALL test infrastructure
#
# This script starts all required services for comprehensive testing:
# - PostgreSQL (database)
# - Redis (caching)
# - Mock LLM Server (API testing)
# - Kafka + Zookeeper (messaging)
# - RabbitMQ (messaging)
# - MinIO (object storage)
# - Iceberg REST (data lakehouse)
# - Qdrant (vector database)
#
# Supports both Docker and Podman
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     HelixAgent Full Test Infrastructure Startup           ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Detect container runtime (Docker or Podman)
detect_runtime() {
    if command -v docker &> /dev/null && docker info &> /dev/null 2>&1; then
        RUNTIME="docker"
        COMPOSE_CMD="docker compose"
        # Check if docker compose v2 is available
        if ! docker compose version &> /dev/null 2>&1; then
            if command -v docker-compose &> /dev/null; then
                COMPOSE_CMD="docker-compose"
            else
                echo -e "${RED}Error: Neither 'docker compose' nor 'docker-compose' found${NC}"
                exit 1
            fi
        fi
    elif command -v podman &> /dev/null; then
        RUNTIME="podman"
        if command -v podman-compose &> /dev/null; then
            COMPOSE_CMD="podman-compose"
        else
            echo -e "${RED}Error: podman-compose not found. Install it with: pip install podman-compose${NC}"
            exit 1
        fi
    else
        echo -e "${RED}Error: Neither Docker nor Podman found. Please install one of them.${NC}"
        exit 1
    fi

    echo -e "${BLUE}Detected runtime: ${RUNTIME}${NC}"
    echo -e "${BLUE}Compose command: ${COMPOSE_CMD}${NC}"
    echo ""
}

# Create network if it doesn't exist
create_network() {
    echo -e "${YELLOW}Checking network...${NC}"
    if ! $RUNTIME network inspect helixagent-network &> /dev/null 2>&1; then
        echo "Creating helixagent-network..."
        $RUNTIME network create helixagent-network
        echo -e "${GREEN}Network created!${NC}"
    else
        echo -e "${GREEN}Network already exists${NC}"
    fi
}

# Wait for a service to be ready
wait_for_service() {
    local service=$1
    local check_cmd=$2
    local max_attempts=${3:-60}
    local attempt=1

    echo -n "  Waiting for $service"
    while [ $attempt -le $max_attempts ]; do
        if eval "$check_cmd" > /dev/null 2>&1; then
            echo -e " ${GREEN}✓${NC}"
            return 0
        fi
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done

    echo -e " ${RED}✗ (timeout after $max_attempts attempts)${NC}"
    return 1
}

# Start basic services (PostgreSQL, Redis, Mock LLM)
start_basic_services() {
    echo -e "${YELLOW}Starting basic services (PostgreSQL, Redis, Mock LLM)...${NC}"
    cd "$PROJECT_ROOT"

    # Build mock LLM if needed
    $COMPOSE_CMD -f docker-compose.test.yml build mock-llm 2>/dev/null || true

    # Start basic services
    $COMPOSE_CMD -f docker-compose.test.yml up -d postgres redis mock-llm

    # Wait for services
    wait_for_service "PostgreSQL" "$COMPOSE_CMD -f docker-compose.test.yml exec -T postgres pg_isready -U helixagent -d helixagent_db"
    wait_for_service "Redis" "$COMPOSE_CMD -f docker-compose.test.yml exec -T redis redis-cli ping"
    wait_for_service "Mock LLM" "curl -sf http://localhost:18081/health"

    echo ""
}

# Start messaging services (Kafka, RabbitMQ)
start_messaging_services() {
    echo -e "${YELLOW}Starting messaging services (Kafka, RabbitMQ)...${NC}"
    cd "$PROJECT_ROOT"

    # Start with messaging profile
    $COMPOSE_CMD -f docker-compose.test.yml -f docker-compose.messaging.yml --profile messaging up -d

    # Wait for services
    wait_for_service "Zookeeper" "nc -z localhost 2181" 30
    wait_for_service "Kafka" "nc -z localhost 9092" 60
    wait_for_service "RabbitMQ" "curl -sf http://localhost:15672" 60

    echo ""
}

# Start bigdata services (MinIO, Iceberg, Qdrant)
start_bigdata_services() {
    echo -e "${YELLOW}Starting bigdata services (MinIO, Iceberg, Qdrant)...${NC}"
    cd "$PROJECT_ROOT"

    # Start with bigdata profile
    $COMPOSE_CMD -f docker-compose.test.yml -f docker-compose.messaging.yml -f docker-compose.bigdata.yml --profile bigdata up -d

    # Wait for services
    wait_for_service "MinIO" "curl -sf http://localhost:9000/minio/health/live" 60
    wait_for_service "Iceberg REST" "curl -sf http://localhost:8181/v1/config" 60
    wait_for_service "Qdrant" "curl -sf http://localhost:6333/health" 30

    echo ""
}

# Export environment variables
export_env_vars() {
    echo -e "${YELLOW}Setting environment variables...${NC}"

    # Database
    export DB_HOST=localhost
    export DB_PORT=15432
    export DB_USER=helixagent
    export DB_PASSWORD=helixagent123
    export DB_NAME=helixagent_db
    export DATABASE_URL="postgres://helixagent:helixagent123@localhost:15432/helixagent_db?sslmode=disable"

    # Redis
    export REDIS_HOST=localhost
    export REDIS_PORT=16379
    export REDIS_PASSWORD=helixagent123
    export REDIS_URL="redis://:helixagent123@localhost:16379"

    # Mock LLM
    export MOCK_LLM_URL=http://localhost:18081
    export MOCK_LLM_ENABLED=true

    # Kafka
    export KAFKA_BROKERS=localhost:9092
    export KAFKA_BROKER=localhost:9092

    # RabbitMQ
    export RABBITMQ_HOST=localhost
    export RABBITMQ_PORT=5672
    export RABBITMQ_USER=helixagent
    export RABBITMQ_PASSWORD=helixagent123
    export RABBITMQ_URL="amqp://helixagent:helixagent123@localhost:5672/"

    # MinIO
    export MINIO_ENDPOINT=localhost:9000
    export MINIO_ACCESS_KEY=minioadmin
    export MINIO_SECRET_KEY=minioadmin123
    export MINIO_USE_SSL=false

    # Iceberg
    export ICEBERG_CATALOG_URI=http://localhost:8181

    # Qdrant
    export QDRANT_HOST=localhost
    export QDRANT_PORT=6333

    # LLM Providers (mock)
    export CLAUDE_API_KEY=mock-api-key
    export CLAUDE_BASE_URL=http://localhost:18081/v1
    export DEEPSEEK_API_KEY=mock-api-key
    export DEEPSEEK_BASE_URL=http://localhost:18081/v1
    export GEMINI_API_KEY=mock-api-key
    export GEMINI_BASE_URL=http://localhost:18081/v1
    export QWEN_API_KEY=mock-api-key
    export QWEN_BASE_URL=http://localhost:18081/v1
    export ZAI_API_KEY=mock-api-key
    export ZAI_BASE_URL=http://localhost:18081/v1
    export OLLAMA_BASE_URL=http://localhost:18081

    # JWT and Server
    export JWT_SECRET=test-jwt-secret-key-for-testing
    export HELIXAGENT_API_KEY=test-api-key
    export GIN_MODE=test

    # Write to env file for other scripts
    cat > "$PROJECT_ROOT/.env.test" << EOF
# Generated by start-full-test-infra.sh
# Database
DB_HOST=localhost
DB_PORT=15432
DB_USER=helixagent
DB_PASSWORD=helixagent123
DB_NAME=helixagent_db
DATABASE_URL=postgres://helixagent:helixagent123@localhost:15432/helixagent_db?sslmode=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=16379
REDIS_PASSWORD=helixagent123
REDIS_URL=redis://:helixagent123@localhost:16379

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_BROKER=localhost:9092

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=helixagent
RABBITMQ_PASSWORD=helixagent123
RABBITMQ_URL=amqp://helixagent:helixagent123@localhost:5672/

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false

# Iceberg
ICEBERG_CATALOG_URI=http://localhost:8181

# Qdrant
QDRANT_HOST=localhost
QDRANT_PORT=6333

# Mock LLM
MOCK_LLM_URL=http://localhost:18081
MOCK_LLM_ENABLED=true

# JWT and Server
JWT_SECRET=test-jwt-secret-key-for-testing
HELIXAGENT_API_KEY=test-api-key
GIN_MODE=test
EOF

    echo -e "${GREEN}Environment variables exported to .env.test${NC}"
    echo ""
}

# Print status
print_status() {
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║           Test Infrastructure Ready!                       ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Services available:${NC}"
    echo "  PostgreSQL:    localhost:15432"
    echo "  Redis:         localhost:16379"
    echo "  Mock LLM:      http://localhost:18081"
    echo "  Kafka:         localhost:9092"
    echo "  RabbitMQ:      localhost:5672 (Management: http://localhost:15672)"
    echo "  MinIO:         http://localhost:9000 (Console: http://localhost:9001)"
    echo "  Iceberg REST:  http://localhost:8181"
    echo "  Qdrant:        http://localhost:6333"
    echo ""
    echo -e "${BLUE}To run tests:${NC}"
    echo "  source .env.test && go test ./... -v"
    echo ""
    echo -e "${BLUE}To stop all services:${NC}"
    echo "  ./scripts/stop-full-test-infra.sh"
    echo ""
}

# Main execution
main() {
    local mode="${1:-all}"

    detect_runtime
    create_network

    case "$mode" in
        basic)
            start_basic_services
            ;;
        messaging)
            start_basic_services
            start_messaging_services
            ;;
        bigdata)
            start_basic_services
            start_bigdata_services
            ;;
        all|*)
            start_basic_services
            start_messaging_services
            start_bigdata_services
            ;;
    esac

    export_env_vars
    print_status
}

# Show help
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo "Usage: $0 [mode]"
    echo ""
    echo "Modes:"
    echo "  basic     - Start PostgreSQL, Redis, Mock LLM only"
    echo "  messaging - Start basic + Kafka, RabbitMQ"
    echo "  bigdata   - Start basic + MinIO, Iceberg, Qdrant"
    echo "  all       - Start all services (default)"
    echo ""
    exit 0
fi

main "$@"
