# docker/lsp/rust-analyzer/Dockerfile
# Rust Language Server (rust-analyzer) - Standalone
FROM rust:1.75-slim-bookworm

RUN apt-get update && apt-get install -y \
    socat \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install rust-analyzer
RUN rustup component add rust-analyzer

# Create workspace directory
RUN mkdir -p /workspace

EXPOSE 5002

CMD ["socat", "TCP-LISTEN:5002,reuseaddr,fork", "EXEC:/root/.rustup/toolchains/1.75.0-x86_64-unknown-linux-gnu/bin/rust-analyzer"]
