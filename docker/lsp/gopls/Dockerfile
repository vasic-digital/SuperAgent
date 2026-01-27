# docker/lsp/gopls/Dockerfile
# Go Language Server (gopls) - Standalone
FROM golang:1.24-alpine

RUN apk add --no-cache socat curl

# Install gopls
RUN go install golang.org/x/tools/gopls@latest

# Create workspace directory
RUN mkdir -p /workspace

EXPOSE 5001

CMD ["socat", "TCP-LISTEN:5001,reuseaddr,fork", "EXEC:/go/bin/gopls serve"]
