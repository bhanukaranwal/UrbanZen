# UrbanZen Deployment Guide

## Overview

This guide provides step-by-step instructions for deploying the UrbanZen IoT Smart City Management Platform in different environments.

## Prerequisites

### Development Environment
- Go 1.21+
- Python 3.11+
- Node.js 18+
- Docker & Docker Compose
- Git

### Production Environment
- Kubernetes cluster (1.25+)
- Helm 3.0+
- PostgreSQL 15+
- TimescaleDB
- MongoDB 6.0+
- Redis 7+
- InfluxDB 2.0+
- Apache Kafka
- MQTT Broker (Mosquitto)

## Quick Start (Development)

### 1. Clone the Repository
```bash
git clone https://github.com/bhanukaranwal/UrbanZen.git
cd UrbanZen
```

### 2. Start Infrastructure Services
```bash
# Start databases and message queues
docker-compose up -d postgres timescaledb mongodb redis influxdb kafka mosquitto

# Wait for services to be ready
sleep 30
```

### 3. Build and Run Services
```bash
# Build all services
make build

# Run backend services
make run-services

# In another terminal, run frontend applications
make run-frontend
```

### 4. Access Applications
- **Admin Dashboard**: http://localhost:3001
- **Public Dashboard**: http://localhost:3002
- **API Gateway**: http://localhost:8080
- **API Documentation**: http://localhost:8080/swagger/index.html

## Production Deployment

### 1. Prepare Kubernetes Cluster

#### Create Namespaces
```bash
kubectl apply -f infrastructure/kubernetes/namespaces/namespaces.yaml
```

#### Set up Secrets
```bash
# Database secrets
kubectl create secret generic database-secrets \
  --from-literal=postgres-url="postgresql://user:pass@postgres:5432/urbanzen" \
  --from-literal=timescaledb-url="postgresql://user:pass@timescaledb:5432/urbanzen_timeseries" \
  --from-literal=mongodb-url="mongodb://user:pass@mongodb:27017/urbanzen" \
  -n urbanzen-prod

# Cache secrets
kubectl create secret generic cache-secrets \
  --from-literal=redis-url="redis://redis:6379/0" \
  --from-literal=redis-password="secure_password" \
  -n urbanzen-prod

# Authentication secrets
kubectl create secret generic auth-secrets \
  --from-literal=jwt-secret="your_jwt_secret_key" \
  --from-literal=api-key="your_api_key" \
  -n urbanzen-prod
```

### 2. Deploy Infrastructure

#### Database Deployment
```bash
# Deploy PostgreSQL
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install postgres bitnami/postgresql \
  --set auth.postgresPassword=urbanzen_secure_password \
  --set auth.database=urbanzen \
  -n urbanzen-prod

# Deploy TimescaleDB
helm install timescaledb timescale/timescaledb-single \
  --set credentials.postgres.password=urbanzen_secure_password \
  -n urbanzen-prod

# Deploy MongoDB
helm install mongodb bitnami/mongodb \
  --set auth.rootPassword=urbanzen_secure_password \
  --set auth.database=urbanzen \
  -n urbanzen-prod

# Deploy Redis
helm install redis bitnami/redis \
  --set auth.password=urbanzen_secure_password \
  -n urbanzen-prod

# Deploy Kafka
helm install kafka bitnami/kafka \
  --set zookeeper.enabled=true \
  -n urbanzen-prod
```

#### Monitoring Stack
```bash
# Deploy Prometheus
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n urbanzen-prod

# Deploy Grafana (if not included in prometheus stack)
helm install grafana grafana/grafana \
  --set adminPassword=urbanzen_admin_password \
  -n urbanzen-prod

# Deploy Elasticsearch
helm repo add elastic https://helm.elastic.co
helm install elasticsearch elastic/elasticsearch \
  -n urbanzen-prod
```

### 3. Deploy Application Services

#### Build and Push Docker Images
```bash
# Build all service images
make docker-build

# Tag and push to registry
docker tag urbanzen/api-gateway:latest your-registry/urbanzen/api-gateway:v1.0.0
docker push your-registry/urbanzen/api-gateway:v1.0.0

# Repeat for all services...
```

#### Deploy Services
```bash
# Deploy backend services
kubectl apply -f infrastructure/kubernetes/services/backend-services.yaml

# Deploy frontend applications
kubectl apply -f infrastructure/kubernetes/services/frontend-services.yaml

# Deploy ingress
kubectl apply -f infrastructure/kubernetes/ingress/ingress.yaml
```

### 4. Configure Ingress and SSL

#### Install Nginx Ingress Controller
```bash
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --set controller.service.type=LoadBalancer \
  -n urbanzen-prod
```

#### Set up SSL with Cert-Manager
```bash
# Install cert-manager
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.11.0/cert-manager.yaml

# Create ClusterIssuer for Let's Encrypt
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@urbanzen.gov.in
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

## Environment Configuration

### Development (.env.development)
```env
ENVIRONMENT=development
LOG_LEVEL=debug
POSTGRES_URL=postgres://urbanzen:urbanzen_secure_password@localhost:5432/urbanzen?sslmode=disable
TIMESCALEDB_URL=postgres://urbanzen:urbanzen_secure_password@localhost:5433/urbanzen_timeseries?sslmode=disable
REDIS_URL=redis://localhost:6379/0
MONGODB_URL=mongodb://urbanzen:urbanzen_secure_password@localhost:27017/urbanzen
KAFKA_BROKERS=localhost:9092
MQTT_BROKER=tcp://localhost:1883
JWT_SECRET=development_jwt_secret_key
```

### Production (.env.production)
```env
ENVIRONMENT=production
LOG_LEVEL=info
POSTGRES_URL=${DATABASE_SECRET_POSTGRES_URL}
TIMESCALEDB_URL=${DATABASE_SECRET_TIMESCALEDB_URL}
REDIS_URL=${CACHE_SECRET_REDIS_URL}
MONGODB_URL=${DATABASE_SECRET_MONGODB_URL}
KAFKA_BROKERS=kafka-service:9092
MQTT_BROKER=tcp://mosquitto-service:1883
JWT_SECRET=${AUTH_SECRET_JWT_SECRET}
```

## Monitoring and Observability

### Prometheus Metrics
- Service health and performance metrics
- Custom business metrics
- Infrastructure metrics

### Grafana Dashboards
- System overview dashboard
- Service-specific dashboards
- Business intelligence dashboards

### Logging with ELK Stack
- Centralized logging for all services
- Log aggregation and analysis
- Real-time log monitoring

### Alerting
- Configure AlertManager for critical alerts
- Set up PagerDuty/Slack integrations
- Define escalation policies

## Backup and Disaster Recovery

### Database Backups
```bash
# PostgreSQL backup
kubectl create cronjob postgres-backup \
  --image=postgres:15 \
  --schedule="0 2 * * *" \
  -- /bin/sh -c "pg_dump $POSTGRES_URL > /backup/postgres-$(date +%Y%m%d).sql"

# TimescaleDB backup
kubectl create cronjob timescaledb-backup \
  --image=timescale/timescaledb:latest-pg15 \
  --schedule="0 3 * * *" \
  -- /bin/sh -c "pg_dump $TIMESCALEDB_URL > /backup/timescaledb-$(date +%Y%m%d).sql"
```

### Application Data Backup
- Regular snapshots of critical data
- Cross-region replication for high availability
- Automated recovery procedures

## Security Best Practices

### Network Security
- Use private networks for internal communication
- Implement network policies in Kubernetes
- Set up WAF (Web Application Firewall)

### Data Security
- Encrypt data at rest and in transit
- Implement proper RBAC
- Regular security audits

### API Security
- Rate limiting and throttling
- Input validation and sanitization
- OAuth2 and JWT for authentication

## Scaling Guidelines

### Horizontal Scaling
- Configure HPA (Horizontal Pod Autoscaler)
- Set up cluster autoscaling
- Use load balancers effectively

### Database Scaling
- Implement read replicas
- Set up database sharding
- Use connection pooling

### Caching Strategy
- Implement multi-level caching
- Use CDN for static content
- Cache frequently accessed data

## Troubleshooting

### Common Issues

#### Service Discovery Problems
```bash
# Check service endpoints
kubectl get endpoints -n urbanzen-prod

# Check DNS resolution
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup api-gateway-service
```

#### Database Connection Issues
```bash
# Check database pods
kubectl get pods -l app=postgres -n urbanzen-prod

# Check connection from service pod
kubectl exec -it deployment/api-gateway -- telnet postgres-service 5432
```

#### Performance Issues
```bash
# Check resource usage
kubectl top pods -n urbanzen-prod

# Check application logs
kubectl logs -f deployment/api-gateway -n urbanzen-prod
```

## Maintenance

### Regular Maintenance Tasks
1. Update dependencies and security patches
2. Monitor resource usage and optimize
3. Review and update backup procedures
4. Conduct disaster recovery drills
5. Update documentation

### Health Checks
- Implement comprehensive health check endpoints
- Monitor service dependencies
- Set up synthetic monitoring

## Support and Documentation

### Additional Resources
- **API Documentation**: Available at `/swagger/index.html`
- **Architecture Documentation**: `docs/ARCHITECTURE.md`
- **Security Documentation**: `docs/SECURITY.md`
- **Contributing Guidelines**: `CONTRIBUTING.md`

### Getting Help
- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check the `docs/` directory
- **Community Support**: Join our community channels

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.