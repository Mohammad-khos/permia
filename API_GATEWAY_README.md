# Permia API Gateway with Traefik

## Overview

The Permia API Gateway is a complete reverse proxy and routing solution built with **Go + Gin** and integrated with **Traefik v3** for intelligent routing, load balancing, and service discovery.

### Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌──────────────────────────────────────┐
│  Traefik (Port 80/443)               │
│  - Request routing & load balancing  │
│  - Docker provider for auto-discovery│
│  - Dashboard on :8080                │
└────────────┬─────────────────────────┘
             │
       ┌─────┴──────────────────┬──────────────────┐
       │                        │                  │
       ▼                        ▼                  ▼
┌─────────────────┐    ┌─────────────────┐   ┌──────────────┐
│ API Gateway     │    │ Core Service    │   │ Bot Service  │
│ :8000           │    │ :8080           │   │ :8081        │
│ (Gin Router)    │    │ (Go/Gin)        │   │ (Telebot)    │
└─────────────────┘    └─────────────────┘   └──────────────┘
       │                        │                  │
       └────────────────────────┼──────────────────┘
                                │
                        ┌───────▼────────┐
                        │  PostgreSQL    │
                        │  :5432         │
                        └────────────────┘
```

## Components

### 1. **Traefik**

- **Image**: `traefik:v3.0`
- **Port**: 80 (HTTP), 443 (HTTPS), 8080 (Dashboard)
- **Provider**: Docker (automatic service discovery)
- **Features**:
  - Auto-discovery of services via Docker labels
  - Health checks
  - Middleware chains (CORS, auth, logging)
  - Dashboard at `http://localhost:8080`

### 2. **API Gateway** (New)

- **Service**: `api-gateway` / Container: `permia_gateway`
- **Port**: 8000 (internal)
- **Built with**: Go 1.25, Gin, Resty
- **Features**:
  - CORS middleware (configurable origins)
  - Authentication middleware
  - Request/response logging with Zap
  - Dynamic routing to backend services
  - Health checks on all endpoints

### 3. **Core Service**

- **Service**: `core-service` / Container: `permia_core`
- **Port**: 8080 (internal)
- **Routes Through Gateway**: `/api/v1/*`
- **Features**: User management, products, orders, payments

### 4. **Bot Service**

- **Service**: `bot-service` / Container: `permia_bot`
- **Port**: 8081 (internal)
- **Routes Through Gateway**: `/api/v1/bot/*`
- **Features**: Telegram bot integration, menu handling

## Environment Setup

### Configuration Variables

```env
# Gateway
GATEWAY_PORT=8000
APP_ENV=production
APP_NAME=Permia API Gateway

# Backend Services
CORE_API_URL=http://core-service:8080/api/v1
BOT_API_URL=http://bot-service:8081/api/v1

# Traefik
TRAEFIK_ENTRYPOINT=web
TRAEFIK_DASHBOARD=true

# CORS
ALLOW_ORIGINS=*

# Logging
LOG_LEVEL=info
```

## API Endpoints

### Through Traefik (Port 80)

```bash
# Health check (Gateway)
GET http://localhost/health

# Core Service Routes
GET http://localhost/api/v1/health
GET http://localhost/api/v1/products
POST http://localhost/api/v1/auth/login
GET http://localhost/api/v1/profile
POST http://localhost/api/v1/orders
POST http://localhost/api/v1/payments

# Admin Routes
POST http://localhost/api/v1/admin/*

# PgAdmin
GET http://localhost/pgadmin

# Traefik Dashboard
GET http://localhost:8080/dashboard/
```

## Traefik Routing Rules

### Router Priorities (Higher = Processed First)

1. **core-api** (Priority: 100)

   - Rule: `PathPrefix(/api/v1)` → routes to `core-service:8080`

2. **bot-api** (Priority: 200)

   - Rule: `PathPrefix(/api/v1/bot)` → routes to `bot-service:8081`

3. **gateway** (Priority: 10)

   - Rule: `PathPrefix(/)` → routes to `api-gateway:8000`

4. **pgadmin** (Priority: 300)
   - Rule: `PathPrefix(/pgadmin)` → routes to `pgadmin:80`

## Docker Compose Services

```yaml
traefik:
  - Reverse proxy and load balancer
  - Auto-discovery of services
  - Health checks and middleware

postgres:
  - Database for core-service
  - Persistent volume: pg_data

core-service:
  - Backend API service
  - Connected to postgres

bot-service:
  - Telegram bot service
  - Connected to core-service

api-gateway:
  - Reverse proxy router
  - Routes to core and bot services
  - CORS and auth middleware

pgadmin:
  - Database management UI
  - Accessible via Traefik
```

## Deployment Commands

### Build All Services

```bash
cd deployment
docker-compose build --no-cache
```

### Bring Up Stack

```bash
docker-compose up -d
```

### Check Service Status

```bash
docker-compose ps
```

### View Logs

```bash
# Gateway logs
docker-compose logs -f api-gateway

# Core service logs
docker-compose logs -f core-service

# Bot service logs
docker-compose logs -f bot-service

# Traefik logs
docker-compose logs -f traefik
```

### Stop Stack

```bash
docker-compose down
```

### Stop and Remove Volumes

```bash
docker-compose down -v
```

## Testing the Setup

### Test Gateway Health

```bash
curl http://localhost/health
```

**Response:**

```json
{ "status": "healthy", "service": "Permia API Gateway" }
```

### Test Core Service Through Gateway

```bash
curl http://localhost/api/v1/health
```

**Response:**

```json
{ "success": true, "message": "Core Service is healthy 🚀", "data": "OK" }
```

### Test Products Endpoint

```bash
curl http://localhost/api/v1/products
```

### Test with Authorization Header

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost/api/v1/profile
```

### Test CORS Preflight

```bash
curl -X OPTIONS http://localhost/api/v1/health \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: GET"
```

## Gateway Features

### 1. CORS Middleware

- Handles CORS preflight requests
- Configurable allowed origins (default: `*`)
- Supports methods: GET, POST, PUT, DELETE, PATCH, OPTIONS
- Allowed headers: Content-Type, Authorization, X-Admin-Token

### 2. Logging Middleware

- Logs all incoming requests
- Records method, path, status code, and latency
- Integrated with Uber Zap for structured logging

### 3. Authentication Middleware

- Optional token validation
- Public routes bypass authentication
- Skips auth for: `/health`, `/api/v1/products`, `/api/v1/auth/login`

### 4. Proxy Service

- Routes requests to backend services
- Preserves headers
- Handles request/response bodies
- Retry logic with configurable timeouts

### 5. Error Handling

- Graceful error responses
- Status code forwarding
- Detailed error logging

## Gateway Configuration

### File: `services/api-gateway/internal/config/config.go`

```go
type Config struct {
    Port                string  // Server port (8000)
    AppEnv              string  // Environment (development/production)
    AppName             string  // Application name
    CoreServiceURL      string  // Core service backend URL
    BotServiceURL       string  // Bot service backend URL
    TraefikEntrypoint   string  // Traefik entrypoint (web)
    TraefikDashboard    string  // Enable Traefik dashboard
    AllowOrigins        string  // CORS allowed origins
    LogLevel            string  // Logging level (info/debug)
}
```

## Middleware Chain

Requests flow through middleware in this order:

1. **LoggingMiddleware** - Logs request details
2. **CORSMiddleware** - Handles CORS
3. **AuthenticationMiddleware** - Validates auth tokens
4. **TraefikLabelMiddleware** - Adds gateway headers
5. **Recovery** - Handles panics

## Health Checks

All services include health check endpoints:

```bash
# Gateway
GET http://localhost/health
GET http://localhost/api/v1/health

# Core service (through gateway)
GET http://localhost/api/v1/health

# Traefik dashboard
GET http://localhost:8080/ping
```

## Monitoring

### Traefik Dashboard

Access at: `http://localhost:8080/dashboard/`

Features:

- Live routing rules
- Service status
- Middleware configuration
- Request history

### Logs

All services log structured output (JSON format with Zap logger):

```json
{
  "level": "info",
  "ts": 1764684092.245,
  "caller": "cmd/main.go:30",
  "msg": "🌉 API Gateway starting"
}
```

## Performance Tuning

### Connection Timeouts

- Gateway: 30 seconds default
- Retry count: 1
- Retry wait time: 1 second

### Rate Limiting

- Currently disabled (placeholder middleware)
- Can be enhanced with rate limiting library

### Resource Limits

Can be configured in `docker-compose.yml`:

```yaml
services:
  api-gateway:
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
```

## Troubleshooting

### Gateway Not Responding

```bash
# Check gateway logs
docker logs permia_gateway

# Check if gateway container is running
docker ps | grep permia_gateway
```

### Routes Not Working

```bash
# Verify Traefik configuration
curl http://localhost:8080/api/http/routers

# Check service labels in docker-compose
docker-compose config | grep -A 10 "labels:"
```

### CORS Issues

```bash
# Check CORS headers in response
curl -v http://localhost/api/v1/health

# Should show: Access-Control-Allow-* headers
```

### Backend Service Unavailable

```bash
# Check core-service
docker logs permia_core

# Check connectivity between containers
docker exec permia_gateway curl http://core-service:8080/api/v1/health
```

## Development

### Adding New Endpoints

1. Update `internal/handler/handler.go` with new handler function
2. Register route in `cmd/main.go`:

```go
coreGroup.GET("/new-endpoint", apiHandler.ProxyCoreAPI)
```

3. Rebuild gateway:

```bash
docker-compose build --no-cache api-gateway
docker-compose up -d api-gateway
```

### Modifying Middleware

Edit `internal/middleware/middleware.go` and rebuild.

### Changing Router Priority

Update Traefik labels in `docker-compose.yml`:

```yaml
- "traefik.http.routers.gateway.priority=10"
```

Lower numbers = lower priority (processed after higher priority routes).

## Security Considerations

1. **CORS**: Set specific origins instead of `*` in production
2. **Authentication**: Implement strict token validation
3. **Rate Limiting**: Enable in production
4. **HTTPS**: Configure SSL/TLS with certificates
5. **Admin Routes**: Protect with authentication middleware
6. **Secrets**: Store tokens/credentials in `.env` (not in compose file)

## Production Checklist

- [ ] Set `APP_ENV=production` in `.env`
- [ ] Configure specific `ALLOW_ORIGINS` (not `*`)
- [ ] Enable HTTPS with SSL certificates
- [ ] Enable rate limiting
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Enable Traefik access logs
- [ ] Set resource limits in compose
- [ ] Use secrets management (Docker Secrets or HashiCorp Vault)
- [ ] Implement authentication on all protected routes
- [ ] Set up backup strategy for PostgreSQL

## Files Modified/Created

### New Files

- `services/api-gateway/cmd/main.go` - Main entry point
- `services/api-gateway/internal/config/config.go` - Configuration loader
- `services/api-gateway/internal/handler/handler.go` - HTTP request handlers
- `services/api-gateway/internal/middleware/middleware.go` - Middleware functions
- `services/api-gateway/internal/service/proxy.go` - Proxy routing logic
- `services/api-gateway/internal/domain/backend.go` - Domain models
- `deployment/gateway.Dockerfile` - Multi-stage build for gateway

### Modified Files

- `deployment/docker-compose.yml` - Added Traefik, updated service labels
- `services/api-gateway/go.mod` - Added dependencies

## Support & Debugging

For debugging, enable debug logging:

```bash
# Update environment variable
GATEWAY_LOG_LEVEL=debug

# Rebuild and restart
docker-compose up -d --build api-gateway
```

Check logs with:

```bash
docker logs -f permia_gateway --tail=100
```

---

**Last Updated**: December 2, 2025  
**Status**: ✅ Fully Implemented & Tested
