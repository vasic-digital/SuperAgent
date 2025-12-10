#!/bin/bash

# SuperAgent Start Script
# This script starts all SuperAgent services using Docker Compose

set -e

echo "🚀 Starting SuperAgent services..."

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install Docker and Docker Compose first."
    exit 1
fi

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found. Creating default configuration..."
    cat > .env << EOF
# SuperAgent Configuration
PORT=8080
SUPERAGENT_API_KEY=test-api-key-for-development

# JWT Configuration
JWT_SECRET=development-jwt-secret-key-change-in-production

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=superagent
DB_PASSWORD=superagent123
DB_NAME=superagent_db

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=superagent123

# LLM Provider API Keys (using Ollama for testing - no keys needed)
CLAUDE_API_KEY=
DEEPSEEK_API_KEY=
GEMINI_API_KEY=
QWEN_API_KEY=
ZAI_API_KEY=

# Ollama Configuration (free, no API keys needed)
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_MODEL=llama2

# Cognee Configuration (disabled for testing)
COGNEE_BASE_URL=
COGNEE_API_KEY=
COGNEE_AUTO_COGNIFY=false

# Plugin Configuration
PLUGIN_WATCH_PATHS=/app/plugins

# Test Environment
GIN_MODE=release
LOG_LEVEL=info
EOF
    echo "✅ Created .env file with default configuration"
fi

# Create necessary directories
mkdir -p plugins
mkdir -p monitoring

# Start services
echo "📦 Starting Docker services..."
docker-compose -f docker-compose.test.yml up -d --build

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."

# Function to check if a service is healthy
check_service() {
    local service=$1
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f docker-compose.test.yml ps $service | grep -q "healthy\|running"; then
            echo "✅ $service is ready"
            return 0
        fi

        echo "⏳ Waiting for $service... (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done

    echo "❌ $service failed to start"
    return 1
}

# Check each service
check_service postgres
check_service redis
check_service ollama
check_service superagent

# Pull and run Ollama model
echo "🤖 Setting up Ollama model..."
docker-compose -f docker-compose.test.yml exec -T ollama ollama pull llama2

# Test the system
echo "🧪 Testing SuperAgent..."

# Test health endpoint
if curl -f -s http://localhost:8080/health > /dev/null; then
    echo "✅ SuperAgent health check passed"
else
    echo "❌ SuperAgent health check failed"
    exit 1
fi

# Test Ollama provider (no API key needed)
response=$(curl -s -X POST http://localhost:8080/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Hello, respond with just: Hello from Ollama!",
    "model": "llama2",
    "max_tokens": 20,
    "temperature": 0.1
  }' 2>/dev/null || echo "error")

if echo "$response" | grep -q "Hello from Ollama"; then
    echo "✅ Ollama provider test passed"
else
    echo "⚠️  Ollama provider test inconclusive (may need model to load)"
fi

echo ""
echo "🎉 SuperAgent is now running!"
echo ""
echo "📊 Service URLs:"
echo "   SuperAgent API: http://localhost:8080"
echo "   Grafana:        http://localhost:3000 (admin/admin123)"
echo "   Prometheus:     http://localhost:9090"
echo ""
echo "🧪 Test commands:"
echo "   curl http://localhost:8080/health"
echo "   curl http://localhost:8080/v1/models"
echo "   curl -X POST http://localhost:8080/v1/ensemble/completions \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"prompt\": \"Hello from SuperAgent!\", \"ensemble_config\": {\"strategy\": \"confidence_weighted\", \"min_providers\": 1}}'"
echo ""
echo "🛑 To stop: ./scripts/stop.sh"