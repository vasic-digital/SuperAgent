# docker/lsp/python-lsp/Dockerfile
# Python Language Server (pylsp + pyright) - Standalone
FROM python:3.11-slim-bookworm

RUN apt-get update && apt-get install -y \
    socat \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Python LSP servers
RUN pip install --no-cache-dir \
    python-lsp-server[all] \
    pyright \
    pylsp-mypy \
    python-lsp-black \
    pylsp-rope \
    ruff-lsp

# Create workspace directory
RUN mkdir -p /workspace

EXPOSE 5003

CMD ["socat", "TCP-LISTEN:5003,reuseaddr,fork", "EXEC:pylsp"]
