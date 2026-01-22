# ACP Manager Dockerfile
# Agent Communication Protocol Manager
# Handles agent discovery, routing, and message passing

FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /acp-manager ./main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=builder /acp-manager /app/acp-manager

ENV ACP_PORT=9200
ENV ENABLE_AGENT_DISCOVERY=true
ENV ENABLE_AGENT_ROUTING=true

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:${ACP_PORT}/health || exit 1

EXPOSE 9200

CMD ["/app/acp-manager"]
