# Deployment Guide

Complete guide for deploying the Frustration Engine to production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Docker Deployment](#docker-deployment)
3. [Kubernetes Deployment](#kubernetes-deployment)
4. [Cloud Provider Deployment](#cloud-provider-deployment)
5. [Environment Configuration](#environment-configuration)
6. [Database Setup](#database-setup)
7. [Monitoring Setup](#monitoring-setup)

---

## Prerequisites

- Docker & Docker Compose (for containerized deployment)
- Kubernetes cluster (for K8s deployment)
- PostgreSQL 15+ database
- ClickHouse database
- Domain name and SSL certificates
- API keys for Jira/Linear (if using ticket export)

---

## Docker Deployment

### 1. Create Docker Compose File

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: frustration_engine
      POSTGRES_USER: ${POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  clickhouse:
    image: clickhouse/clickhouse-server:latest
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    ports:
      - "9000:9000"
      - "8123:8123"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "localhost:8123/ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  event-ingestion:
    build:
      context: .
      dockerfile: Dockerfile
    command: ./bin/event-ingestion
    environment:
      PORT: 8080
      CLICKHOUSE_DSN: clickhouse://clickhouse:9000
      SESSION_MANAGER_URL: http://session-manager:8081
      RATE_LIMIT_RPS: ${RATE_LIMIT_RPS:-1000}
      RATE_LIMIT_BURST: ${RATE_LIMIT_BURST:-2000}
    ports:
      - "8080:8080"
    depends_on:
      clickhouse:
        condition: service_healthy
      session-manager:
        condition: service_started
    restart: unless-stopped

  session-manager:
    build:
      context: .
      dockerfile: Dockerfile
    command: ./bin/session-manager
    environment:
      PORT: 8081
      UFSE_URL: http://ufse:8082
    ports:
      - "8081:8081"
    depends_on:
      ufse:
        condition: service_started
    restart: unless-stopped

  ufse:
    build:
      context: .
      dockerfile: Dockerfile
    command: ./bin/ufse
    environment:
      PORT: 8082
      INCIDENT_STORE_URL: http://incident-store:8084
    ports:
      - "8082:8082"
    depends_on:
      incident-store:
        condition: service_started
    restart: unless-stopped

  incident-store:
    build:
      context: .
      dockerfile: Dockerfile
    command: ./bin/incident-store
    environment:
      PORT: 8084
      DATABASE_URL: postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD}@postgres:5432/frustration_engine?sslmode=disable
    ports:
      - "8084:8084"
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  ticket-exporter:
    build:
      context: .
      dockerfile: Dockerfile
    command: ./bin/ticket-exporter
    environment:
      PORT: 8085
      INCIDENT_STORE_URL: http://incident-store:8084
      JIRA_URL: ${JIRA_URL}
      JIRA_EMAIL: ${JIRA_EMAIL}
      JIRA_API_TOKEN: ${JIRA_API_TOKEN}
      LINEAR_API_KEY: ${LINEAR_API_KEY}
      LINEAR_TEAM_ID: ${LINEAR_TEAM_ID}
    ports:
      - "8085:8085"
    depends_on:
      incident-store:
        condition: service_started
    restart: unless-stopped

volumes:
  postgres_data:
  clickhouse_data:
```

### 2. Create Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/event-ingestion ./cmd/event-ingestion
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/session-manager ./cmd/session-manager
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/ufse ./cmd/ufse
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/incident-store ./cmd/incident-store
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/ticket-exporter ./cmd/ticket-exporter

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin ./bin

EXPOSE 8080 8081 8082 8084 8085

CMD ["./bin/event-ingestion"]
```

### 3. Deploy

```bash
# Build and start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f event-ingestion
```

---

## Kubernetes Deployment

### 1. Create Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: frustration-engine
```

### 2. Create ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: frustration-engine-config
  namespace: frustration-engine
data:
  SESSION_MANAGER_URL: "http://session-manager:8081"
  UFSE_URL: "http://ufse:8082"
  INCIDENT_STORE_URL: "http://incident-store:8084"
  CLICKHOUSE_DSN: "clickhouse://clickhouse:9000"
```

### 3. Create Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: frustration-engine-secrets
  namespace: frustration-engine
type: Opaque
stringData:
  POSTGRES_PASSWORD: "your-password"
  JIRA_API_TOKEN: "your-token"
  LINEAR_API_KEY: "your-key"
```

### 4. Deploy Services

See `k8s/` directory for complete Kubernetes manifests.

---

## Cloud Provider Deployment

### AWS (ECS)

1. **Create ECR repositories** for each service
2. **Build and push images**:

```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

docker build -t frustration-engine-event-ingestion .
docker tag frustration-engine-event-ingestion:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/frustration-engine-event-ingestion:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/frustration-engine-event-ingestion:latest
```

3. **Create ECS task definitions** for each service
4. **Create ECS services** with load balancers
5. **Set up RDS** for PostgreSQL
6. **Set up ClickHouse** on EC2 or use managed service

### GCP (GKE)

1. **Create GCR repositories**
2. **Build and push images**:

```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/frustration-engine-event-ingestion
```

3. **Deploy to GKE** using Kubernetes manifests
4. **Set up Cloud SQL** for PostgreSQL
5. **Set up ClickHouse** on GCE or use managed service

---

## Environment Configuration

### Required Environment Variables

**Event Ingestion:**
```bash
PORT=8080
CLICKHOUSE_DSN=clickhouse://localhost:9000
SESSION_MANAGER_URL=http://session-manager:8081
RATE_LIMIT_RPS=1000
RATE_LIMIT_BURST=2000
```

**Session Manager:**
```bash
PORT=8081
UFSE_URL=http://ufse:8082
```

**UFSE:**
```bash
PORT=8082
INCIDENT_STORE_URL=http://incident-store:8084
```

**Incident Store:**
```bash
PORT=8084
DATABASE_URL=postgres://user:password@localhost:5432/frustration_engine?sslmode=disable
```

**Ticket Exporter:**
```bash
PORT=8085
INCIDENT_STORE_URL=http://incident-store:8084
JIRA_URL=https://your-domain.atlassian.net
JIRA_EMAIL=your-email@example.com
JIRA_API_TOKEN=your-token
LINEAR_API_KEY=your-key
LINEAR_TEAM_ID=your-team-id
```

---

## Database Setup

### PostgreSQL

```sql
-- Create database
CREATE DATABASE frustration_engine;

-- Run migrations (see internal/store/schema.go)
-- The Incident Store will create tables automatically on first run
```

### ClickHouse

```sql
-- Create database
CREATE DATABASE frustration_engine;

-- Tables are created automatically by Event Ingestion service
```

---

## Monitoring Setup

### Prometheus

Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'frustration-engine'
    static_configs:
      - targets:
        - 'event-ingestion:8080'
        - 'session-manager:8081'
        - 'ufse:8082'
        - 'incident-store:8084'
        - 'ticket-exporter:8085'
```

### Grafana

Import dashboards from `grafana/` directory.

### Log Aggregation

Configure log shipping to:
- **AWS**: CloudWatch Logs
- **GCP**: Cloud Logging
- **Azure**: Application Insights
- **Self-hosted**: ELK Stack, Loki

---

## Production Checklist

- [ ] All services are running and healthy
- [ ] Databases are backed up
- [ ] SSL/TLS certificates are configured
- [ ] API keys are stored in secrets manager
- [ ] Rate limiting is configured
- [ ] Monitoring and alerting are set up
- [ ] Log aggregation is configured
- [ ] CORS is properly configured
- [ ] Health checks are working
- [ ] Load balancers are configured
- [ ] Auto-scaling is configured (if needed)
- [ ] Disaster recovery plan is in place

---

## Troubleshooting

### Services Not Starting

1. Check logs: `docker-compose logs <service-name>`
2. Verify environment variables
3. Check database connectivity
4. Verify port availability

### High Latency

1. Check database performance
2. Verify network connectivity between services
3. Check resource limits (CPU/memory)
4. Review rate limiting settings

### Database Connection Issues

1. Verify connection strings
2. Check firewall rules
3. Verify credentials
4. Test connectivity: `psql -h host -U user -d database`

---

## Support

For deployment issues:
1. Check service logs
2. Verify health endpoints
3. Review environment configuration
4. Check database connectivity
