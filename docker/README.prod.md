# Production Monitoring Stack

This production-ready monitoring stack uses Cloudflare R2 for object storage and implements security best practices.

## Key Features

- **Cloudflare R2 Storage**: Cost-effective object storage for Loki logs and Tempo traces
- **Authentication**: Basic auth on Prometheus, secure Grafana setup
- **Security**: Non-root users, internal networks, resource limits
- **Monitoring**: Health checks, alerting with Alertmanager
- **Retention**: Configurable data retention policies

## Setup Instructions

### 1. Create Cloudflare R2 Buckets

```bash
# Create buckets in Cloudflare R2
# - grafto-tempo-traces
# - grafto-loki-logs
```

### 2. Configure Environment Variables

```bash
cp .env.prod.example .env.prod
# Edit .env.prod with your actual values
```

### 3. Deploy the Stack

```bash
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

## Security Improvements

### Authentication
- Grafana: Admin user with secure password
- Prometheus: Basic authentication enabled
- Loki: Authentication enabled

### Network Security
- Internal network prevents external access
- Only Grafana exposed externally (port 3001)
- All inter-service communication internal

### Container Security
- Non-root users for all services
- Resource limits to prevent resource exhaustion
- Health checks for service monitoring

## Storage Configuration

### Cloudflare R2 Benefits
- S3-compatible API
- No egress fees
- Cost-effective for long-term storage
- Global edge locations

### Retention Policies
- Tempo traces: 7 days
- Loki logs: 30 days
- Prometheus metrics: 30 days (50GB limit)

## Monitoring & Alerting

### Alertmanager
- Email notifications configured
- Alert grouping and deduplication
- Configurable notification channels

### Health Checks
- All services have health check endpoints
- Automatic restart on failure
- Service dependency management

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `CLOUDFLARE_R2_ENDPOINT` | R2 endpoint URL | `https://account-id.r2.cloudflarestorage.com` |
| `CLOUDFLARE_R2_ACCESS_KEY` | R2 access key | `your-access-key` |
| `CLOUDFLARE_R2_SECRET_KEY` | R2 secret key | `your-secret-key` |
| `TEMPO_R2_BUCKET` | Tempo traces bucket | `grafto-tempo-traces` |
| `LOKI_R2_BUCKET` | Loki logs bucket | `grafto-loki-logs` |
| `GRAFANA_ADMIN_PASSWORD` | Grafana admin password | `secure-password` |
| `GRAFANA_SECRET_KEY` | Grafana secret key | `32-character-key` |
| `PROMETHEUS_ADMIN_USER` | Prometheus username | `prometheus` |
| `PROMETHEUS_ADMIN_PASSWORD` | Prometheus password | `secure-password` |

## Accessing Services

- **Grafana**: https://grafana.yourdomain.com (or localhost:3001)
- **Prometheus**: Internal only (http://prometheus:9090)
- **Alertmanager**: Internal only (http://alertmanager:9093)

## Backup Strategy

### Data Persistence
- Prometheus: Docker managed volume (automatic persistence)
- Grafana: Docker managed volume (automatic persistence)  
- Alertmanager: Docker managed volume (automatic persistence)
- Loki/Tempo: R2 object storage (automatically replicated)

### Backup Recommendations
- Regular backup of Grafana dashboards and configuration
- Export Prometheus rules and alerting configuration
- Monitor R2 storage costs and usage

## Scaling Considerations

### Horizontal Scaling
- Add multiple Prometheus instances with federation
- Use Loki clustering for high availability
- Implement Grafana clustering for redundancy

### Resource Scaling
- Adjust memory/CPU limits based on usage
- Monitor storage growth and adjust retention
- Scale R2 storage as needed (automatic)

## Troubleshooting

### Common Issues
1. **R2 Connection**: Verify credentials and endpoint URL
2. **Authentication**: Check username/password configuration
3. **Storage**: Monitor disk usage on local volumes
4. **Network**: Ensure internal network connectivity

### Logs
```bash
# View service logs
docker-compose -f docker-compose.prod.yml logs -f [service-name]

# Check health status
docker-compose -f docker-compose.prod.yml ps
```