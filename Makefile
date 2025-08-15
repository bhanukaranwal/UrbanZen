# Makefile for UrbanZen IoT Smart City Platform

.PHONY: help build test run clean docker-build docker-run k8s-deploy

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build all services"
	@echo "  test           - Run all tests"
	@echo "  run-services   - Run all backend services"
	@echo "  run-frontend   - Run all frontend applications"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  k8s-deploy     - Deploy to Kubernetes"
	@echo "  lint           - Run linters"
	@echo "  format         - Format code"

# Build targets
build: build-services build-frontend

build-services:
	@echo "Building backend services..."
	cd services/api-gateway && go build -o bin/api-gateway ./cmd/main.go
	cd services/device-mgmt && go build -o bin/device-mgmt ./cmd/main.go
	cd services/data-ingestion && go build -o bin/data-ingestion ./cmd/main.go
	cd services/notification && go build -o bin/notification ./cmd/main.go
	cd services/user-mgmt && go build -o bin/user-mgmt ./cmd/main.go
	cd services/billing && go build -o bin/billing ./cmd/main.go
	cd services/reporting && go build -o bin/reporting ./cmd/main.go
	cd services/analytics && pip install -r requirements.txt

build-frontend:
	@echo "Building frontend applications..."
	cd frontend/admin-dashboard && npm install && npm run build
	cd frontend/public-dashboard && npm install && npm run build
	cd frontend/field-officer && npm install
	cd frontend/citizen-app && flutter pub get

# Test targets
test: test-services test-frontend

test-services:
	@echo "Testing backend services..."
	cd services/api-gateway && go test ./...
	cd services/device-mgmt && go test ./...
	cd services/data-ingestion && go test ./...
	cd services/notification && go test ./...
	cd services/user-mgmt && go test ./...
	cd services/billing && go test ./...
	cd services/reporting && go test ./...
	cd services/analytics && python -m pytest

test-frontend:
	@echo "Testing frontend applications..."
	cd frontend/admin-dashboard && npm test
	cd frontend/public-dashboard && npm test
	cd frontend/field-officer && npm test
	cd frontend/citizen-app && flutter test

# Run targets
run-services:
	@echo "Starting backend services..."
	docker-compose up -d postgres timescaledb mongodb redis influxdb kafka
	cd services/api-gateway && go run cmd/main.go &
	cd services/device-mgmt && go run cmd/main.go &
	cd services/data-ingestion && go run cmd/main.go &
	cd services/notification && go run cmd/main.go &
	cd services/user-mgmt && go run cmd/main.go &
	cd services/billing && go run cmd/main.go &
	cd services/reporting && go run cmd/main.go &
	cd services/analytics && python main.py &

run-frontend:
	@echo "Starting frontend applications..."
	cd frontend/admin-dashboard && npm start &
	cd frontend/public-dashboard && npm run dev &
	cd frontend/field-officer && npm start &
	cd frontend/citizen-app && flutter run &

# Docker targets
docker-build:
	@echo "Building Docker images..."
	docker-compose build

docker-run:
	@echo "Running with Docker Compose..."
	docker-compose up -d

# Kubernetes targets
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f infrastructure/kubernetes/namespaces/
	kubectl apply -f infrastructure/kubernetes/databases/
	kubectl apply -f infrastructure/kubernetes/services/
	kubectl apply -f infrastructure/kubernetes/ingress/

# Code quality targets
lint:
	@echo "Running linters..."
	cd services/api-gateway && golangci-lint run
	cd services/device-mgmt && golangci-lint run
	cd services/data-ingestion && golangci-lint run
	cd services/notification && golangci-lint run
	cd services/user-mgmt && golangci-lint run
	cd services/billing && golangci-lint run
	cd services/reporting && golangci-lint run
	cd services/analytics && flake8 . && black --check .
	cd frontend/admin-dashboard && npm run lint
	cd frontend/public-dashboard && npm run lint
	cd frontend/field-officer && npm run lint

format:
	@echo "Formatting code..."
	cd services/api-gateway && go fmt ./...
	cd services/device-mgmt && go fmt ./...
	cd services/data-ingestion && go fmt ./...
	cd services/notification && go fmt ./...
	cd services/user-mgmt && go fmt ./...
	cd services/billing && go fmt ./...
	cd services/reporting && go fmt ./...
	cd services/analytics && black .
	cd frontend/admin-dashboard && npm run format
	cd frontend/public-dashboard && npm run format
	cd frontend/field-officer && npm run format

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	find . -name "bin" -type d -exec rm -rf {} +
	find . -name "node_modules" -type d -exec rm -rf {} +
	find . -name "build" -type d -exec rm -rf {} +
	find . -name "dist" -type d -exec rm -rf {} +
	find . -name "__pycache__" -type d -exec rm -rf {} +
	find . -name "*.pyc" -delete

# Setup targets
setup:
	@echo "Setting up development environment..."
	@echo "Installing Go dependencies..."
	go mod tidy
	@echo "Installing Python dependencies..."
	pip install -r services/analytics/requirements.txt
	@echo "Installing Node.js dependencies..."
	cd frontend/admin-dashboard && npm install
	cd frontend/public-dashboard && npm install
	cd frontend/field-officer && npm install
	@echo "Installing Flutter dependencies..."
	cd frontend/citizen-app && flutter pub get

# Database targets
db-migrate:
	@echo "Running database migrations..."
	cd infrastructure/databases && ./migrate.sh

db-seed:
	@echo "Seeding database with sample data..."
	cd infrastructure/databases && ./seed.sh

# Monitoring targets
monitoring-setup:
	@echo "Setting up monitoring stack..."
	kubectl apply -f infrastructure/monitoring/prometheus/
	kubectl apply -f infrastructure/monitoring/grafana/
	kubectl apply -f infrastructure/monitoring/elasticsearch/

# Security targets
security-scan:
	@echo "Running security scans..."
	gosec ./services/...
	bandit -r services/analytics/
	npm audit --audit-level moderate