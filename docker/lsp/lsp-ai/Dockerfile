# docker/lsp/lsp-ai/Dockerfile
# LSP-AI: AI-powered Language Server
FROM rust:1.75-slim-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Clone and build lsp-ai
WORKDIR /build
RUN git clone --depth 1 https://github.com/SilasMarvin/lsp-ai.git && \
    cd lsp-ai && \
    cargo build --release

FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

# Copy the built binary
COPY --from=builder /build/lsp-ai/target/release/lsp-ai /usr/local/bin/lsp-ai

# Create config directory
RUN mkdir -p /root/.config/lsp-ai

# Health check script
RUN echo '#!/bin/bash\ncurl -sf http://localhost:${LSP_AI_PORT:-5000}/health || exit 1' > /healthcheck.sh && \
    chmod +x /healthcheck.sh

EXPOSE 5000

# Entry point that runs lsp-ai
CMD ["lsp-ai", "--port", "5000"]
