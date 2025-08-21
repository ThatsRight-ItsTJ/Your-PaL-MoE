.PHONY: build test lint clean docker-build docker-run help

# Variables
BINARY_NAME=intelligent-ai-gateway
DOCKER_IMAGE=intelligent-ai-gateway
DOCKER_TAG=latest

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	cd core && go build -o ../$(BINARY_NAME) .

# Run tests
test:
	@echo "Running tests..."
	cd tests && go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	cd tests && go test -v -coverprofile=coverage.out ./...
	cd tests && go tool cover -html=coverage.out -o coverage.html

# Run load tests
load-test:
	@echo "Running load tests..."
	cd tests && go test -v -run TestHighConcurrency

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	cd tests && go test -bench=. -benchmem

# Lint the code
lint:
	@echo "Running linter..."
	cd core && golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f tests/coverage.out tests/coverage.html

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		--name $(BINARY_NAME) \
		-p 3000:3000 \
		-e ADMIN_KEY=admin-key \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Run with Docker Compose
compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose services
compose-down:
	@echo "Stopping services..."
	docker-compose down

# View logs
logs:
	docker-compose logs -f gateway

# Run development server
dev:
	@echo "Starting development server..."
	cd core && go run . --config ../dev-config.yaml

# Install dependencies
deps:
	@echo "Installing dependencies..."
	cd core && go mod download
	cd tests && go mod download

# Update providers configuration
update-providers:
	@echo "Reloading providers configuration..."
	curl -X POST http://localhost:3000/admin/reload \
		-H "Authorization: Bearer admin-key"

# Health check
health:
	@echo "Checking gateway health..."
	curl -s http://localhost:3000/health || echo "Gateway not responding"

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  load-test      - Run load tests"
	@echo "  benchmark      - Run benchmark tests"
	@echo "  lint           - Run code linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  compose-up     - Start with Docker Compose"
	@echo "  compose-down   - Stop Docker Compose services"
	@echo "  logs           - View application logs"
	@echo "  dev            - Run development server"
	@echo "  deps           - Install dependencies"
	@echo "  update-providers - Reload providers configuration"
	@echo "  health         - Check gateway health"
	@echo "  help           - Show this help"