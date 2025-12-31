FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY main.go .

RUN go build -o mock-llm-server main.go

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/mock-llm-server .

EXPOSE 8081

CMD ["./mock-llm-server"]
