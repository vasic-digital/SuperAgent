# HelixAgent MCP Server Dockerfile
# Provides MCP protocol access to all HelixAgent features for CLI agents

FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json tsconfig.json ./
COPY src ./src

# Install dependencies and build
RUN npm install && npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Install runtime dependencies only
COPY package.json ./
RUN npm install --production

# Copy built files
COPY --from=builder /app/dist ./dist

# Environment variables
ENV NODE_ENV=production
ENV MCP_PORT=9100
ENV HELIXAGENT_URL=http://helixagent:7061

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD wget -qO- http://localhost:${MCP_PORT}/health || exit 1

# Expose port
EXPOSE 9100

# Start server
CMD ["node", "dist/index.js"]
