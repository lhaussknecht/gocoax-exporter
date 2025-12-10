.PHONY: all build test clean run docker-build docker-run help

# Variables
BINARY_NAME=gocoax-exporter
DOCKER_IMAGE=gocoax-exporter
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the exporter locally
run:
	@echo "Running $(BINARY_NAME)..."
	go run . -config config.yaml

# Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(VERSION)..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@echo "Docker build complete"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 9090:9090 -v $(PWD)/config.yaml:/etc/gocoax-exporter/config.yaml $(DOCKER_IMAGE):latest

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run || go vet ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-race       - Run tests with race detection"
	@echo "  clean           - Remove build artifacts"
	@echo "  run             - Run the exporter locally"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter"
	@echo "  deps            - Install dependencies"
	@echo "  help            - Show this help message"
