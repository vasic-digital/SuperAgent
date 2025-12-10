#!/bin/bash

# SuperAgent Stop Script
# This script stops all SuperAgent services and cleans up

set -e

echo "🛑 Stopping SuperAgent services..."

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed."
    exit 1
fi

# Stop services
echo "📦 Stopping Docker services..."
docker-compose -f docker-compose.test.yml down

# Optional: Clean up volumes (uncomment if you want to remove all data)
# echo "🧹 Cleaning up volumes..."
# docker-compose -f docker-compose.test.yml down -v

# Remove orphaned containers
echo "🧽 Cleaning up orphaned containers..."
docker system prune -f

echo "✅ SuperAgent services stopped successfully"
echo ""
echo "💡 To restart: ./scripts/start.sh"
echo "💡 To view logs: docker-compose -f docker-compose.test.yml logs"
echo "💡 To restart specific service: docker-compose -f docker-compose.test.yml restart <service>"