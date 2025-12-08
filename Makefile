.PHONY: all build test run fmt lint security-scan

all: build

build:
	@echo "Building..."
	go build ./...

test:
	@echo "Running tests..."
	go test ./...

run:
	@echo "Running server..."
	go run ./cmd/superagent/main.go

fmt:
	@echo "Formatting..."
	go fmt ./...

lint:
	@echo "Linting... (offline placeholder)"

security-scan:
	@echo "Security scan: (offline placeholder)"
