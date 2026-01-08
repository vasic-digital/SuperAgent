#!/bin/bash

# HelixAgent Production Deployment Script
# This script sets up HelixAgent for production use

set -e

echo "🚀 HelixAgent Production Deployment Script"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root (not recommended for production)
if [[ $EUID -eq 0 ]]; then
    print_warning "Running as root is not recommended for production deployments."
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check system requirements
print_status "Checking system requirements..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    print_status "Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not available. Please install Docker Compose."
    print_status "Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

print_success "System requirements met!"

# Create deployment directory
DEPLOY_DIR="./helixagent-deployment"
if [ -d "$DEPLOY_DIR" ]; then
    print_warning "Deployment directory already exists: $DEPLOY_DIR"
    read -p "Remove and recreate? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$DEPLOY_DIR"
    else
        print_error "Aborting deployment to avoid overwriting existing files."
        exit 1
    fi
fi

print_status "Creating deployment directory..."
mkdir -p "$DEPLOY_DIR"
cd "$DEPLOY_DIR"

# Clone the repository
print_status "Cloning HelixAgent repository..."
if ! git clone https://github.com/helixagent/helixagent.git .; then
    print_error "Failed to clone repository. Please check your internet connection."
    exit 1
fi

print_success "Repository cloned successfully!"

# Create environment configuration
print_status "Setting up environment configuration..."

cat > .env << EOF
# HelixAgent Production Environment Configuration
# Generated on: $(date)

# ===========================================
# SERVER CONFIGURATION
# ===========================================
PORT=8080
GIN_MODE=release
LOG_LEVEL=info
REQUEST_TIMEOUT=30

# ===========================================
# DATABASE CONFIGURATION
# ===========================================
DB_HOST=postgres
DB_PORT=5432
DB_USER=helixagent_prod
DB_PASSWORD=CHANGE_THIS_STRONG_PASSWORD
DB_NAME=helixagent_prod
DB_MAX_CONNECTIONS=20
DB_SSL_MODE=disable

# ===========================================
# REDIS CONFIGURATION
# ===========================================
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=CHANGE_THIS_STRONG_REDIS_PASSWORD
REDIS_DB=0
CACHE_TTL=3600

# ===========================================
# AI PROVIDER API KEYS
# ===========================================
# REQUIRED: Configure at least one provider below

# Anthropic Claude
CLAUDE_API_KEY=your_claude_api_key_here

# OpenRouter (Multi-provider)
OPENROUTER_API_KEY=your_openrouter_api_key_here

# Optional: Additional providers
# DEEPSEEK_API_KEY=your_deepseek_api_key_here
# GEMINI_API_KEY=your_gemini_api_key_here
# QWEN_API_KEY=your_qwen_api_key_here

# ===========================================
# SECURITY CONFIGURATION
# ===========================================
JWT_SECRET=$(openssl rand -hex 32)
API_KEY_SECRET=$(openssl rand -hex 32)
CORS_ALLOWED_ORIGINS=https://yourdomain.com,http://localhost:3000

# ===========================================
# MONITORING CONFIGURATION
# ===========================================
PROMETHEUS_ENABLED=true
GRAFANA_ENABLED=true
METRICS_ENABLED=true

# ===========================================
# AI DEBATE CONFIGURATION
# ===========================================
ENABLE_AI_DEBATE=true
MAX_DEBATE_ROUNDS=10
DEBATE_TIMEOUT=300000

# ===========================================
# MCP (Model Context Protocol)
# ===========================================
MCP_ENABLED=true
MCP_UNIFIED_TOOL_NAMESPACE=true

# ===========================================
# PERFORMANCE TUNING
# ===========================================
MAX_WORKERS=10
CONNECTION_POOL_SIZE=20
RATE_LIMIT_REQUESTS_PER_MINUTE=60
EOF

print_success "Environment configuration created!"
print_warning "IMPORTANT: Edit .env file and configure your API keys before starting!"

# Make the script executable and create helper scripts
print_status "Creating helper scripts..."

cat > start.sh << 'EOF'
#!/bin/bash
echo "Starting HelixAgent in production mode..."
docker-compose --profile prod up -d
echo "HelixAgent is starting up..."
echo "API will be available at: http://localhost:8080"
echo "Health check: http://localhost:8080/health"
echo "Grafana dashboards: http://localhost:3000 (admin/admin123)"
echo "Prometheus metrics: http://localhost:9090"
echo ""
echo "To view logs: docker-compose logs -f helixagent"
echo "To stop: docker-compose down"
EOF

cat > stop.sh << 'EOF'
#!/bin/bash
echo "Stopping HelixAgent..."
docker-compose down
echo "HelixAgent stopped."
EOF

cat > logs.sh << 'EOF'
#!/bin/bash
echo "Showing HelixAgent logs..."
docker-compose logs -f helixagent
EOF

cat > status.sh << 'EOF'
#!/bin/bash
echo "HelixAgent Status:"
echo "=================="
docker-compose ps

echo ""
echo "Health Check:"
curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/health || echo "API not responding"

echo ""
echo "Service URLs:"
echo "- API: http://localhost:8080"
echo "- Health: http://localhost:8080/health"
echo "- Metrics: http://localhost:9090"
echo "- Grafana: http://localhost:3000 (admin/admin123)"
EOF

chmod +x *.sh

print_success "Helper scripts created!"

# Create production docker-compose override
print_status "Setting up production configuration..."

cat > docker-compose.prod.yml << EOF
version: '3.8'

services:
  # Override for production settings
  helixagent:
    environment:
      - GIN_MODE=release
      - LOG_LEVEL=info
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
        reservations:
          memory: 512M
          cpus: '0.5'
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  postgres:
    environment:
      POSTGRES_PASSWORD: \${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U \${DB_USER} -d \${DB_NAME}"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    command: redis-server --requirepass \${REDIS_PASSWORD} --maxmemory 512mb --maxmemory-policy allkeys-lru
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  prometheus:
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    restart: unless-stopped

  grafana:
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin123
      GF_USERS_ALLOW_SIGN_UP: false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:
EOF

print_success "Production configuration created!"

# Final instructions
cat << EOF

🎉 **HelixAgent Production Deployment Ready!**

Your HelixAgent instance is now configured for production deployment.

**Next Steps:**

1. **Configure API Keys:**
   \`\`\`bash
   nano .env
   \`\`\`
   Edit the .env file and add your AI provider API keys (at minimum, configure CLAUDE_API_KEY or OPENROUTER_API_KEY)

2. **Start HelixAgent:**
   \`\`\`bash
   ./start.sh
   \`\`\`

3. **Check Status:**
   \`\`\`bash
   ./status.sh
   \`\`\`

4. **View Logs:**
   \`\`\`bash
   ./logs.sh
   \`\`\`

**Access Points:**
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Grafana Dashboards**: http://localhost:3000 (admin/admin123)
- **Prometheus Metrics**: http://localhost:9090

**Management Commands:**
- Start: \`./start.sh\`
- Stop: \`./stop.sh\`
- Logs: \`./logs.sh\`
- Status: \`./status.sh\`

**Production Security Notes:**
- Change default passwords in .env
- Configure SSL/TLS for production traffic
- Set up proper firewall rules
- Enable log aggregation and monitoring
- Configure backup strategies

**Need Help?**
- Documentation: Check the \`docs/\` directory
- Troubleshooting: See \`docs/troubleshooting-guide.md\`
- API Reference: See \`docs/api-documentation.md\`

**HelixAgent is now ready for production! 🚀**

EOF

print_success "HelixAgent production deployment setup complete!"
print_warning "Remember to configure your API keys in .env before starting!"