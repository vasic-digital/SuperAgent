# MCP Servers Container
# Builds and runs all Model Context Protocol servers for HelixAgent
#
# This container includes:
# - Active MCP servers from modelcontextprotocol/servers (Node.js + Python)
# - Archived MCP servers from modelcontextprotocol/servers-archived (Node.js + Python)

FROM node:22-alpine AS builder

WORKDIR /build

# Install build dependencies including Python
RUN apk add --no-cache git python3 py3-pip make g++ chromium

# Copy server sources
COPY servers/ /build/servers/
COPY servers-archived/ /build/servers-archived/

# Build active Node.js servers
WORKDIR /build/servers
RUN npm ci && npm run build || true

# Build archived Node.js servers (each has its own package.json)
WORKDIR /build/servers-archived/src
RUN for dir in */; do \
    if [ -f "$dir/package.json" ]; then \
        echo "Installing dependencies for $dir" && \
        cd "$dir" && npm ci && npm run build 2>/dev/null || true && cd ..; \
    fi; \
done

# Production image
FROM node:22-alpine

WORKDIR /app

# Install runtime dependencies including Python and pip
RUN apk add --no-cache chromium postgresql-client redis git curl python3 py3-pip

# Set Puppeteer to use system Chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium-browser

# Copy built servers from builder
COPY --from=builder /build/servers /app/servers
COPY --from=builder /build/servers-archived /app/servers-archived

# Install Python dependencies for Python-based servers
WORKDIR /app/servers/src
RUN for dir in */; do \
    if [ -f "$dir/pyproject.toml" ]; then \
        echo "Installing Python dependencies for $dir" && \
        cd "$dir" && pip install --break-system-packages -e . 2>/dev/null || true && cd ..; \
    fi; \
done

WORKDIR /app/servers-archived/src
RUN for dir in */; do \
    if [ -f "$dir/pyproject.toml" ]; then \
        echo "Installing Python dependencies for $dir" && \
        cd "$dir" && pip install --break-system-packages -e . 2>/dev/null || true && cd ..; \
    fi; \
done

WORKDIR /app

# Copy startup scripts
COPY scripts/ /app/scripts/
RUN chmod +x /app/scripts/*.sh

# Expose ports for MCP servers (SSE mode)
# Each server can run on its own port in SSE mode
EXPOSE 3001-3020

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3001/health || exit 1

# Default command - starts all servers
CMD ["/app/scripts/start-all.sh"]
