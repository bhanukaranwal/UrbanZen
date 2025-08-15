#!/bin/bash

# UrbanZen Platform Setup Script
# This script automates the setup of the development environment

set -e

echo "🏙️  UrbanZen Smart City Platform Setup"
echo "======================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    local missing_deps=()
    
    if ! command_exists docker; then
        missing_deps+=("docker")
    fi
    
    if ! command_exists docker-compose; then
        missing_deps+=("docker-compose")
    fi
    
    if ! command_exists go; then
        missing_deps+=("go")
    fi
    
    if ! command_exists python3; then
        missing_deps+=("python3")
    fi
    
    if ! command_exists node; then
        missing_deps+=("node")
    fi
    
    if ! command_exists npm; then
        missing_deps+=("npm")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        print_error "Please install the missing dependencies and run this script again."
        exit 1
    fi
    
    print_success "All prerequisites are installed"
}

# Check Docker daemon
check_docker() {
    print_status "Checking Docker daemon..."
    
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker daemon is not running. Please start Docker and try again."
        exit 1
    fi
    
    print_success "Docker daemon is running"
}

# Setup environment variables
setup_environment() {
    print_status "Setting up environment variables..."
    
    if [ ! -f .env ]; then
        print_status "Creating .env file from template..."
        cat > .env << EOF
# UrbanZen Development Environment Configuration

# General
ENVIRONMENT=development
LOG_LEVEL=info

# Database URLs
POSTGRES_URL=postgres://urbanzen:urbanzen_secure_password@localhost:5432/urbanzen?sslmode=disable
TIMESCALEDB_URL=postgres://urbanzen:urbanzen_secure_password@localhost:5433/urbanzen_timeseries?sslmode=disable
MONGODB_URL=mongodb://urbanzen:urbanzen_secure_password@localhost:27017/urbanzen
REDIS_URL=redis://localhost:6379/0
INFLUXDB_URL=http://localhost:8086

# Message Queue
KAFKA_BROKERS=localhost:9092

# MQTT
MQTT_BROKER=tcp://localhost:1883

# Security
JWT_SECRET=urbanzen_jwt_secret_key_very_secure_for_development
API_KEY=urbanzen_api_key_for_development

# External APIs
API_GATEWAY_URL=http://localhost:8080

# Frontend URLs
REACT_APP_API_URL=http://localhost:8080
NEXT_PUBLIC_API_URL=http://localhost:8080
EOF
        print_success "Created .env file"
    else
        print_warning ".env file already exists, skipping creation"
    fi
}

# Start infrastructure services
start_infrastructure() {
    print_status "Starting infrastructure services with Docker Compose..."
    
    # Pull latest images
    print_status "Pulling Docker images..."
    docker-compose pull
    
    # Start services
    print_status "Starting services..."
    docker-compose up -d postgres timescaledb mongodb redis influxdb kafka zookeeper mosquitto elasticsearch kibana prometheus grafana
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 30
    
    # Check service health
    print_status "Checking service health..."
    
    local services=(
        "postgres:5432"
        "timescaledb:5433" 
        "mongodb:27017"
        "redis:6379"
        "influxdb:8086"
        "kafka:9092"
        "mosquitto:1883"
        "elasticsearch:9200"
        "prometheus:9090"
        "grafana:3000"
    )
    
    for service in "${services[@]}"; do
        local host=$(echo $service | cut -d':' -f1)
        local port=$(echo $service | cut -d':' -f2)
        
        if nc -z localhost $port 2>/dev/null; then
            print_success "$host is ready on port $port"
        else
            print_warning "$host is not ready on port $port"
        fi
    done
}

# Initialize databases
init_databases() {
    print_status "Initializing databases..."
    
    # Wait a bit more for databases to be fully ready
    sleep 15
    
    # Check if PostgreSQL is ready and run initial schema
    print_status "Setting up PostgreSQL schema..."
    until docker-compose exec -T postgres pg_isready -U urbanzen; do
        print_status "Waiting for PostgreSQL to be ready..."
        sleep 5
    done
    
    # Apply PostgreSQL schema
    if docker-compose exec -T postgres psql -U urbanzen -d urbanzen -f /docker-entrypoint-initdb.d/01_schema.sql; then
        print_success "PostgreSQL schema applied successfully"
    else
        print_warning "PostgreSQL schema may have already been applied"
    fi
    
    # Apply TimescaleDB schema
    print_status "Setting up TimescaleDB schema..."
    until docker-compose exec -T timescaledb pg_isready -U urbanzen; do
        print_status "Waiting for TimescaleDB to be ready..."
        sleep 5
    done
    
    if docker-compose exec -T timescaledb psql -U urbanzen -d urbanzen_timeseries -f /docker-entrypoint-initdb.d/01_timeseries_schema.sql; then
        print_success "TimescaleDB schema applied successfully"
    else
        print_warning "TimescaleDB schema may have already been applied"
    fi
}

# Build Go services
build_services() {
    print_status "Building Go services..."
    
    local services=(
        "api-gateway"
        "device-mgmt"
        "data-ingestion"
        "notification"
        "user-mgmt"
        "billing"
        "reporting"
    )
    
    for service in "${services[@]}"; do
        if [ -d "services/$service" ]; then
            print_status "Building $service..."
            cd "services/$service"
            go mod tidy
            go build -o bin/$service ./cmd/main.go
            cd ../..
            print_success "$service built successfully"
        else
            print_warning "Service directory services/$service not found, skipping"
        fi
    done
}

# Setup Python analytics service
setup_analytics() {
    print_status "Setting up Analytics service..."
    
    if [ -d "services/analytics" ]; then
        cd services/analytics
        
        # Create virtual environment if it doesn't exist
        if [ ! -d "venv" ]; then
            print_status "Creating Python virtual environment..."
            python3 -m venv venv
        fi
        
        # Activate virtual environment and install dependencies
        print_status "Installing Python dependencies..."
        source venv/bin/activate
        pip install -r requirements.txt
        deactivate
        
        cd ../..
        print_success "Analytics service setup completed"
    else
        print_warning "Analytics service directory not found, skipping"
    fi
}

# Setup frontend applications
setup_frontend() {
    print_status "Setting up frontend applications..."
    
    # Admin Dashboard
    if [ -d "frontend/admin-dashboard" ]; then
        print_status "Setting up Admin Dashboard..."
        cd frontend/admin-dashboard
        npm install
        cd ../..
        print_success "Admin Dashboard setup completed"
    fi
    
    # Public Dashboard
    if [ -d "frontend/public-dashboard" ]; then
        print_status "Setting up Public Dashboard..."
        cd frontend/public-dashboard
        if [ -f "package.json" ]; then
            npm install
            print_success "Public Dashboard setup completed"
        else
            print_warning "Public Dashboard package.json not found, skipping"
        fi
        cd ../..
    fi
    
    # Citizen App (Flutter)
    if [ -d "frontend/citizen-app" ] && command_exists flutter; then
        print_status "Setting up Citizen App..."
        cd frontend/citizen-app
        flutter pub get
        cd ../..
        print_success "Citizen App setup completed"
    elif [ -d "frontend/citizen-app" ]; then
        print_warning "Flutter not installed, skipping Citizen App setup"
    fi
}

# Generate sample data
generate_sample_data() {
    print_status "Generating sample data..."
    
    # This is a placeholder - implement actual data generation
    print_status "Sample data generation is not yet implemented"
    print_status "You can manually run IoT simulators from the iot/simulators directory"
}

# Display next steps
show_next_steps() {
    echo ""
    print_success "🎉 UrbanZen setup completed successfully!"
    echo ""
    echo "Next steps:"
    echo "==========="
    echo ""
    echo "1. Start the backend services:"
    echo "   make run-services"
    echo ""
    echo "2. Start the frontend applications:"
    echo "   make run-frontend"
    echo ""
    echo "3. Access the applications:"
    echo "   - Admin Dashboard: http://localhost:3001"
    echo "   - Public Dashboard: http://localhost:3002"
    echo "   - API Gateway: http://localhost:8080"
    echo "   - API Documentation: http://localhost:8080/swagger/index.html"
    echo "   - Grafana: http://localhost:3000 (admin/admin)"
    echo "   - Prometheus: http://localhost:9090"
    echo "   - Kibana: http://localhost:5601"
    echo ""
    echo "4. Run IoT simulators (optional):"
    echo "   cd iot/simulators"
    echo "   python3 water_meter_simulator.py"
    echo ""
    echo "5. Check the documentation:"
    echo "   - docs/API.md - API documentation"
    echo "   - docs/DEPLOYMENT.md - Deployment guide"
    echo "   - README.md - Project overview"
    echo ""
    print_success "Happy coding! 🚀"
}

# Main execution
main() {
    check_prerequisites
    check_docker
    setup_environment
    start_infrastructure
    init_databases
    build_services
    setup_analytics
    setup_frontend
    generate_sample_data
    show_next_steps
}

# Run main function
main "$@"