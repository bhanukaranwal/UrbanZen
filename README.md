# UrbanZen: Government-Grade IoT Smart City Management Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Python Version](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://python.org/)
[![React Version](https://img.shields.io/badge/React-18+-blue.svg)](https://reactjs.org/)
[![Flutter Version](https://img.shields.io/badge/Flutter-3.0+-blue.svg)](https://flutter.dev/)

## Overview

UrbanZen is a comprehensive, secure, and fault-tolerant IoT ecosystem designed to transform city infrastructure management in India. The platform supports multi-utility services (water, electricity, transport) with cutting-edge hardware, AI-powered analytics, secure cloud architecture, and intuitive interfaces.

## Architecture

### Microservices Architecture
- **API Gateway** - Central entry point with authentication and routing
- **Device Management** - IoT device registration and monitoring
- **Data Ingestion** - Real-time data processing with Kafka
- **Analytics** - AI/ML models for predictions and insights
- **Notification** - Multi-channel notification system
- **User Management** - Authentication and RBAC
- **Billing** - Utility billing and payment processing
- **Reporting** - Government compliance reports

### Frontend Applications
- **Admin Dashboard** (React + TypeScript) - Real-time monitoring
- **Citizen Mobile App** (Flutter) - Cross-platform mobile app
- **Field Officer App** (React Native) - Maintenance teams
- **Public Dashboard** (Next.js) - Public transparency

### Database Architecture
- **TimescaleDB** - Time-series sensor data
- **PostgreSQL** - Relational data with PostGIS
- **MongoDB** - Unstructured data
- **Redis** - Caching and sessions
- **InfluxDB** - High-frequency metrics

## Quick Start

### Prerequisites
- Go 1.21+
- Python 3.11+
- Node.js 18+
- Docker & Docker Compose
- Kubernetes (optional)

### Automated Setup
```bash
# Clone the repository
git clone https://github.com/bhanukaranwal/UrbanZen.git
cd UrbanZen

# Run the automated setup script
./scripts/setup.sh
```

### Manual Setup
```bash
# Start infrastructure services
docker-compose up -d

# Build and run microservices
make build
make run-services

# Start frontend applications
make run-frontend
```

### Access Applications
- **Admin Dashboard**: http://localhost:3001
- **Public Dashboard**: http://localhost:3002  
- **API Gateway**: http://localhost:8080
- **API Documentation**: http://localhost:8080/swagger/index.html
- **Grafana Monitoring**: http://localhost:3000 (admin/admin)
- **Prometheus Metrics**: http://localhost:9090

## Project Structure

```
UrbanZen/
├── services/              # Backend microservices
│   ├── api-gateway/      # API Gateway service (Go)
│   ├── device-mgmt/      # Device Management service (Go)
│   ├── data-ingestion/   # Data Ingestion service (Go)
│   ├── analytics/        # Analytics service (Python)
│   ├── notification/     # Notification service (Go)
│   ├── user-mgmt/        # User Management service (Go)
│   ├── billing/          # Billing service (Go)
│   └── reporting/        # Reporting service (Go)
├── frontend/             # Frontend applications
│   ├── admin-dashboard/  # React + TypeScript admin app
│   ├── citizen-app/      # Flutter mobile app
│   ├── field-officer/    # React Native maintenance app
│   └── public-dashboard/ # Next.js public dashboard
├── infrastructure/       # Infrastructure configurations
│   ├── databases/        # Database schemas and migrations
│   ├── kubernetes/       # K8s deployment manifests
│   ├── monitoring/       # Prometheus, Grafana configs
│   └── docker/          # Docker configurations
├── iot/                 # IoT device simulators and configs
├── ml-models/           # AI/ML models and training pipelines
├── docs/                # Comprehensive documentation
└── scripts/             # Automation and deployment scripts
```

## Features

### Water Utility Management
- Real-time flow and quality monitoring
- AI-powered leak detection
- Automated valve control
- Smart billing with tamper detection
- Water quality alerts
- Predictive maintenance

### Electricity Distribution
- Real-time grid monitoring
- Demand response automation
- Power quality monitoring
- Outage management
- Energy theft detection
- Renewable energy integration

### Smart Transport
- AI traffic signal optimization
- Real-time incident detection
- Public transport tracking
- Parking management
- Air quality monitoring
- Emergency vehicle priority

### Citizen Services
- Multi-channel complaint management
- Real-time service status
- Utility consumption analytics
- Payment processing
- Emergency alerts
- Feedback tracking

## Security & Compliance

- End-to-end encryption (TLS 1.3, AES-256)
- Multi-factor authentication
- Role-based access control
- OAuth2 API security
- Blockchain audit trail
- CERT-IN compliance
- Data privacy (GDPR, Indian Data Protection Act)

## Government Integration

- e-Gov API integration
- Payment gateway integration
- SMS gateway for notifications
- GIS integration with Survey of India
- Compliance reporting
- Public grievance system

## Implementation Status

### ✅ Completed Components
- **Project Structure**: Complete microservices architecture
- **API Gateway**: Go-based gateway with JWT authentication
- **Device Management**: IoT device registration and monitoring
- **Database Layer**: PostgreSQL + TimescaleDB + schema design
- **Admin Dashboard**: React + TypeScript with Material-UI
- **Citizen Mobile App**: Flutter app structure
- **Data Ingestion**: Kafka-based real-time data processing
- **Analytics Service**: Python + FastAPI framework
- **Docker Configuration**: Complete containerization
- **Kubernetes Deployment**: Production-ready manifests
- **CI/CD Pipeline**: GitHub Actions automation
- **Monitoring**: Prometheus + Grafana setup
- **MQTT Broker**: Mosquitto configuration
- **IoT Simulators**: Water meter simulator
- **ML Models**: Anomaly detection implementation
- **Documentation**: Comprehensive API and deployment docs
- **Setup Automation**: One-click development setup

### 🚧 In Progress
- Additional microservices (Notification, User Management, Billing)
- Complete frontend implementations
- Advanced ML models
- Government integration APIs

### 📋 Planned Features
- Advanced analytics dashboards
- Mobile app completion
- Security compliance (CERT-In)
- Performance optimization
- Multi-language support

## Deployment

### Docker Compose (Development)
```bash
docker-compose up -d
```

### Kubernetes (Production)
```bash
helm install urbanzen ./infrastructure/kubernetes/helm-chart
```

### CI/CD Pipeline
GitHub Actions pipeline automatically builds, tests, and deploys the application.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Email: support@urbanzen.gov.in
- Documentation: [docs/](docs/)
- Issues: [GitHub Issues](https://github.com/bhanukaranwal/UrbanZen/issues)

## Government Compliance

- **Security Standards**: ISO 27001, CERT-In guidelines
- **Data Localization**: All data stored in India
- **Accessibility**: WCAG 2.1 AA compliance
- **Multi-language**: Hindi, English, regional languages
- **Audit Trail**: Immutable transaction logs
- **Disaster Recovery**: RPO < 1 hour, RTO < 4 hours