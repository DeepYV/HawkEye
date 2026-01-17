#!/bin/bash

# Monitoring Setup Script
# Sets up Prometheus metrics and monitoring

set -e

echo "=== Monitoring Setup ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check if Prometheus is available
echo -e "${BLUE}=== Prometheus Configuration ===${NC}"

# Create Prometheus config
cat > prometheus.yml <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'event-ingestion'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'

  - job_name: 'session-manager'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'

  - job_name: 'ufse'
    static_configs:
      - targets: ['localhost:8082']
    metrics_path: '/metrics'

  - job_name: 'incident-store'
    static_configs:
      - targets: ['localhost:8084']
    metrics_path: '/metrics'

  - job_name: 'ticket-exporter'
    static_configs:
      - targets: ['localhost:8085']
    metrics_path: '/metrics'
EOF

echo -e "${GREEN}✅ Prometheus configuration created${NC}"

# Metrics to monitor
echo ""
echo -e "${BLUE}=== Key Metrics to Monitor ===${NC}"
echo "Performance Metrics:"
echo "  - Request latency (p50, p95, p99)"
echo "  - Throughput (requests/second)"
echo "  - Connection pool utilization"
echo "  - Cache hit rate"
echo ""
echo "Reliability Metrics:"
echo "  - Error rate"
echo "  - Retry success rate"
echo "  - Circuit breaker state changes"
echo "  - Service availability"
echo ""
echo "Business Metrics:"
echo "  - Events processed per second"
echo "  - Sessions processed per second"
echo "  - Incidents detected"
echo "  - False positive rate"

echo ""
echo -e "${GREEN}=== Monitoring Setup Complete ===${NC}"
echo "Configuration saved to: prometheus.yml"
