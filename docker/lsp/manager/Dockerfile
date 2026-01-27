# docker/lsp/manager/Dockerfile
# LSP Manager - Orchestrates all LSP servers
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Initialize module
RUN go mod init lsp-manager

# Create main.go
COPY main.go .

RUN CGO_ENABLED=0 go build -o lsp-manager .

FROM alpine:3.19

RUN apk add --no-cache curl ca-certificates

COPY --from=builder /build/lsp-manager /usr/local/bin/lsp-manager

# Create config directory
RUN mkdir -p /config

EXPOSE 5100

CMD ["lsp-manager"]
