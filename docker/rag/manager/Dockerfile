# docker/rag/manager/Dockerfile
# RAG Pipeline Manager - Orchestrates all RAG services
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Initialize module
RUN go mod init rag-manager

# Create main.go
COPY main.go .

RUN CGO_ENABLED=0 go build -o rag-manager .

FROM alpine:3.19

RUN apk add --no-cache curl ca-certificates

COPY --from=builder /build/rag-manager /usr/local/bin/rag-manager

# Create config directory
RUN mkdir -p /config

EXPOSE 8030

CMD ["rag-manager"]
