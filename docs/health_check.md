# Health Check

This feature provides a simple health check endpoint for monitoring the application status and its dependencies.

## API Endpoint

### Health

```bash
GET /api/health
```

Checks the health of the application and its dependencies.

#### Response

```json
{
  "status": "success",
  "data": {
    "time": "2025-05-15T12:00:00Z"
  },
  "message": "API is healthy"
}
```

## Business Logic

The health check endpoint performs the following validations:

1. Verifies the application is running
2. Tests the database connection by establishing a connection and using a ping query
3. Returns the current timestamp in RFC3339 format

## Use Cases

The health check endpoint is used for:

1. **Monitoring**: External monitoring tools can ping this endpoint to check if the service is alive
2. **Load Balancers**: Load balancers can use this endpoint to determine if the service should receive traffic
3. **Docker Health Checks**: Docker containers use this endpoint to determine container health
4. **Kubernetes Probes**: Kubernetes uses this endpoint for liveness and readiness probes
5. **Deployment Verification**: CI/CD pipelines can verify successful deployments by checking this endpoint

## Docker Integration

The health check is integrated with Docker Compose:

```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--spider", "http://localhost:${API_PORT}/api/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

This allows Docker to automatically monitor the health of the service and restart it if necessary.
