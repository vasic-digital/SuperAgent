FROM docker.io/golang:1.23-alpine AS builder

WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o mock-llm-server main.go

FROM docker.io/alpine:latest
WORKDIR /app
COPY --from=builder /app/mock-llm-server .

ENV PORT=8090
EXPOSE 8090

CMD ["./mock-llm-server"]
