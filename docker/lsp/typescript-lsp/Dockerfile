# docker/lsp/typescript-lsp/Dockerfile
# TypeScript Language Server - Standalone
FROM node:20-slim

RUN apt-get update && apt-get install -y \
    socat \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install TypeScript and language server
RUN npm install -g \
    typescript \
    typescript-language-server \
    @vscode/vscode-languageserver-node

# Create workspace directory
RUN mkdir -p /workspace

EXPOSE 5004

CMD ["socat", "TCP-LISTEN:5004,reuseaddr,fork", "EXEC:typescript-language-server --stdio"]
