# docker/lsp/clangd/Dockerfile
# C/C++ Language Server (clangd) - Standalone
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    clangd-15 \
    socat \
    curl \
    && ln -s /usr/bin/clangd-15 /usr/bin/clangd \
    && rm -rf /var/lib/apt/lists/*

# Create workspace directory
RUN mkdir -p /workspace

EXPOSE 5005

CMD ["socat", "TCP-LISTEN:5005,reuseaddr,fork", "EXEC:/usr/bin/clangd"]
