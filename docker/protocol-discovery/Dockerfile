# Protocol Discovery Service Dockerfile
# Central registry for all MCP, LSP, ACP, Embedding servers
# Provides discovery API for CLI agents to find available servers

FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /discovery-server ./main.go

# Production stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=builder /discovery-server /app/discovery-server

ENV DISCOVERY_PORT=9300
ENV ENABLE_AUTO_DISCOVERY=true

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:${DISCOVERY_PORT}/health || exit 1

EXPOSE 9300

CMD ["/app/discovery-server"]
