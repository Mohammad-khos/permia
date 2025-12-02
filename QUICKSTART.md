# API Gateway Quick Start Guide

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose installed
- Port 80, 443, 8000, 8080 available (or configured in `.env`)

### Start the Stack

```bash
cd backend/deployment
docker-compose up -d
```

✅ Services will be available:

- **Gateway**: http://localhost/
- **Core API**: http://localhost/api/v1/
- **Traefik Dashboard**: http://localhost:8080/dashboard/

### Verify Everything is Working

```bash
# Test gateway
curl http://localhost/health

# Test core service through gateway
curl http://localhost/api/v1/health

# Get products
curl http://localhost/api/v1/products
```

## 📊 Traefik Routing

All traffic flows through Traefik on port 80:

```
┌─ Request ─┐
│ Port 80   │
└─────┬─────┘
      │
      ▼
   Traefik
      │
      ├─ /api/v1/* ──────────────────────> Core Service :8080
      │
      ├─ /api/v1/bot/* ─────────────────> Bot Service :8081
      │
      ├─ /pgadmin/* ──────────────────────> PgAdmin :80
      │
      └─ /* ──────────────────────────────> API Gateway :8000
```

## 🛑 Stop the Stack

```bash
docker-compose down
```

## 📝 Common Commands

```bash
# View logs
docker-compose logs -f api-gateway

# Rebuild a service
docker-compose build --no-cache api-gateway

# Restart a service
docker-compose restart api-gateway

# Access database
docker exec -it permia_db psql -U permia_admin -d permia_db

# Check Traefik routers
curl http://localhost:8080/api/http/routers

# Check Traefik services
curl http://localhost:8080/api/http/services
```

## 🔧 Configuration

Edit `.env` file in deployment folder:

```env
GATEWAY_PORT=8000
CORE_API_URL=http://core-service:8080/api/v1
BOT_API_URL=http://bot-service:8081/api/v1
LOG_LEVEL=info
ALLOW_ORIGINS=*
```

Then rebuild:

```bash
docker-compose build --no-cache api-gateway
docker-compose up -d api-gateway
```

## 🧪 Test Endpoints

### Health Checks

```bash
curl http://localhost/health
curl http://localhost/api/v1/health
```

### Products

```bash
curl http://localhost/api/v1/products
```

### Authentication

```bash
curl -H "Authorization: Bearer TOKEN" http://localhost/api/v1/profile
```

### CORS Test

```bash
curl -X OPTIONS http://localhost/api/v1/products \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: GET" -v
```

## 📊 Monitoring

- **Traefik Dashboard**: http://localhost:8080/dashboard/
- **PgAdmin**: http://localhost/pgadmin/
- **Logs**: `docker-compose logs -f [service]`

## 🐛 Troubleshooting

### Gateway not responding?

```bash
docker logs permia_gateway
docker exec permia_gateway curl http://core-service:8080/api/v1/health
```

### Traefik not routing?

```bash
curl http://localhost:8080/api/http/routers
```

### Database connection issues?

```bash
docker logs permia_core | tail -20
```

## 📖 Full Documentation

See `API_GATEWAY_README.md` for detailed documentation.

---

**Status**: ✅ Production Ready
