.PHONY: help build test deploy clean docker-build

PROJECT_NAME=urbanzen
DOCKER_REGISTRY=ghcr.io/bhanukaranwal
VERSION=$(shell git describe --tags --always --dirty)
SERVICES=api-gateway device-service notification-service billing-service

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all services
	@echo "Building Go services..."
	@for service in $(SERVICES); do \
		if [ -f "cmd/$$service/main.go" ]; then \
			echo "Building $$service..."; \
			go build -o bin/$$service ./cmd/$$service; \
		fi; \
	done

test: ## Run all tests
	@echo "Running Go tests..."
	@go test -v -race -coverprofile=coverage.out ./...

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		if [ -f "cmd/$$service/Dockerfile" ]; then \
			echo "Building $$service Docker image..."; \
			docker build -t $(DOCKER_REGISTRY)/$(PROJECT_NAME)-$$service:$(VERSION) -f cmd/$$service/Dockerfile .; \
		fi; \
	done

run-dev: ## Run development environment
	@echo "Starting development environment..."
	@docker-compose up -d

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@docker system prune -f

migrate-up: ## Run database migrations up
	@go run cmd/migrate/main.go up

migrate-down: ## Run database migrations down
	@go run cmd/migrate/main.go down