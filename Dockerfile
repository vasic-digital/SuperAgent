# Multi-stage build for optimized production image
FROM docker.io/golang:1.24-alpine AS builder

# Build arguments for version injection
ARG BUILD_VERSION=dev
ARG BUILD_VERSION_CODE=0
ARG BUILD_COMMIT=unknown
ARG BUILD_DATE=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata curl make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy all source code first (needed for local replace directives)
COPY . .

# Download dependencies (skip verify for local replace directives)
RUN go mod download

# Note: Tests are run in CI before Docker build, skip here for faster builds

# Build the application with optimizations and version info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags \"-static\" \
    -X dev.helix.agent/internal/version.Version=${BUILD_VERSION} \
    -X dev.helix.agent/internal/version.VersionCode=${BUILD_VERSION_CODE} \
    -X dev.helix.agent/internal/version.GitCommit=${BUILD_COMMIT} \
    -X dev.helix.agent/internal/version.BuildDate=${BUILD_DATE} \
    -X dev.helix.agent/internal/version.Builder=docker" \
    -a -installsuffix cgo \
    -o helixagent \
    ./cmd/helixagent

# Production stage
FROM docker.io/alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    jq \
    dumb-init \
    && rm -rf /var/cache/apk/*

# Create non-root user with proper UID/GID
RUN addgroup -g 1001 -S helixagent && \
    adduser -u 1001 -S helixagent -G helixagent -h /app -s /bin/sh

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/helixagent .
COPY --from=builder /app/*.md ./

# Create necessary directories with proper permissions
RUN mkdir -p /app/plugins /app/logs /app/config /app/data && \
    chown -R helixagent:helixagent /app && \
    chmod 755 /app

# Switch to non-root user
USER helixagent

# Re-declare ARG after FROM to use in labels
ARG BUILD_VERSION=dev

# Add labels for metadata
LABEL org.opencontainers.image.title="HelixAgent" \
      org.opencontainers.image.description="AI-powered ensemble LLM service" \
      org.opencontainers.image.version="${BUILD_VERSION}" \
      org.opencontainers.image.vendor="HelixAgent" \
      org.opencontainers.image.licenses="MIT"

# Expose port
EXPOSE 7061

# Health check with better reliability
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f -s http://localhost:7061/health > /dev/null || exit 1

# Use dumb-init for proper signal handling
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Run the application
CMD ["./helixagent"]
