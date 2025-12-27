# Fjord Bank - Docker Setup Guide

This guide explains how to run the full Fjord Bank stack using Docker and Docker Compose.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                     Development Setup                                 │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────┐                         ┌───────────────┐         │
│  │   Frontend   │────────┐                │   PostgreSQL  │         │
│  │  React:3000  │        │                │   DB:5432     │         │
│  │  (Docker)    │        │                │   (Docker)    │         │
│  └──────────────┘        │                └───────────────┘         │
│         │                │                      ▲       ▲            │
│         │                │                      │       │            │
│    Hot Reload            ▼                      │       │            │
│                   ┌─────────────┐               │       │            │
│                   │     API     │───────────────┘       │            │
│                   │   Go:8080   │                       │            │
│                   │   (Local)   │               ┌───────────────┐   │
│                   └─────────────┘──────────────►│     Redis     │   │
│                          │                      │     :6379     │   │
│                   Native Binary                 │   (Docker)    │   │
│                   Fast Performance              └───────────────┘   │
│                                                          ▲            │
│                   ┌─────────────┐                       │            │
│                   │   Worker    │───────────────────────┘            │
│                   │   (Local)   │  Async Job Processing              │
│                   │  Optional   │                                    │
│                   └─────────────┘                                    │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘

Why Go services (API + Worker) run locally:
✅ Native performance (no Docker overhead)
✅ Easier debugging (IDE breakpoints work)
✅ Faster compilation (instant hot reload)
✅ Direct access to host resources
✅ Production-like behavior (Go deploys as binaries)

Why infrastructure (Postgres + Redis + Frontend) runs in Docker:
✅ Consistent versions across team
✅ No local installation required
✅ Easy cleanup and reset
✅ Matches production setup
```

## Prerequisites

- Docker Desktop installed and running
- Docker Compose (included with Docker Desktop)
- Node.js 20+ (for local development without Docker)
- Go 1.21+ (for local development without Docker)

## Quick Start (First Time Setup)

### 1. Initialize React Frontend

```bash
make frontend-init
```

This will:
- Create a new React app with TypeScript in `frontend/`
- Install React Router and other dependencies
- Set up the project structure

### 2. Start Infrastructure (Frontend + DB + Redis)

```bash
make dev
```

This command starts:
- PostgreSQL database on port 5432
- Redis on port 6379
- React frontend on port 3000 (in Docker)

### 3. Start Go Services (Locally)

**Terminal 2 - API Server (Required):**

```bash
make api
# OR
go run ./cmd/api
```

**Terminal 3 - Worker (Optional, for async mode):**

```bash
make worker
# OR
go run ./cmd/worker
```

The worker processes background jobs from Redis queue. Only needed if you set `ASYNC_MODE=true`.

### 4. Run Database Migrations

```bash
make migrate
```

### 5. Access the Application

- **Frontend:** http://localhost:3000
- **API:** http://localhost:8080
- **API Health Check:** http://localhost:8080/health

## Why Run Go Services Locally?

Go applications (API + Worker) are designed to compile to native binaries:

1. **Performance:** No Docker overhead, native CPU instructions
2. **Debugging:** Use your IDE debugger directly on both services
3. **Hot Reload:** `go run` or Air with instant recompilation
4. **Development Speed:** Skip Docker build steps
5. **Production Similarity:** Production Go apps run as binaries, not containers
6. **Your Code:** You're actively developing these, unlike Redis/Postgres which are dependencies

## Common Commands

### Starting & Stopping

```bash
# Terminal 1: Start infrastructure (Frontend + DB + Redis)
make dev

# Terminal 2: Start API server (required)
make api

# Terminal 3: Start worker (optional, for async mode)
make worker

# Stop all Docker containers
make dev-down

# Rebuild frontend container
make dev-build

# Full restart
make dev-down && make dev
# Then start API and optionally worker in separate terminals
```

### Viewing Logs

```bash
# View logs from all Docker services
make dev-logs

# View frontend logs only
make dev-logs-fe

# View logs from specific service
docker-compose logs -f postgres
docker-compose logs -f redis

# Go service logs are in the terminals where you ran them
# Terminal 2: API logs
# Terminal 3: Worker logs
```

### Database Access

```bash
# Open PostgreSQL shell
make psql

# Run SQL query
docker exec -it fjord-pg psql -U fjord -d fjorddb -c "SELECT * FROM customers;"

# Open Redis CLI
make redis-cli
```

### Container Shell Access

```bash
# Frontend container shell
make frontend-shell

# API container shell
make api-shell
```

### Check Status

```bash
# List all Fjord Bank containers
make status

# Or use docker-compose
docker-compose ps
```

## Development Workflow

### Hot Reload

Both frontend and backend support hot reload:

**Frontend (Docker):**
- Edit files in `frontend/src/`
- Changes automatically refresh in browser
- React Fast Refresh enabled
- Volume-mounted for instant updates

**Backend (Local):**
- Changes trigger automatic recompilation with `go run`
- Or use [Air](https://github.com/cosmtrek/air) for even faster hot reload:
  ```bash
  # Install Air (one-time)
  go install github.com/cosmtrek/air@latest

  # Run with Air instead of 'make api'
  air
  ```

### Environment Variables

Edit `docker-compose.yml` to change environment variables:

```yaml
api:
  environment:
    DATABASE_URL: postgres://fjord:fjordpass@postgres:5432/fjorddb?sslmode=disable
    JWT_SECRET: your-secret-here
    ASYNC_MODE: "false"

frontend:
  environment:
    REACT_APP_API_URL: http://localhost:8080
```

## Troubleshooting

### Port Already in Use

If you get "port already in use" errors:

```bash
# Check what's using the port
lsof -i :3000  # Frontend
lsof -i :8080  # API
lsof -i :5432  # PostgreSQL

# Stop conflicting services
make stop-db  # If you have standalone DB running
```

### Database Connection Issues

```bash
# Check if PostgreSQL is healthy
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

### Frontend Not Loading

```bash
# Check frontend logs
make dev-logs-fe

# Rebuild frontend container
docker-compose build frontend
docker-compose up -d frontend
```

### Clean Start

To completely reset everything:

```bash
# Stop and remove all containers and volumes
docker-compose down -v

# Remove node_modules
rm -rf frontend/node_modules

# Start fresh
make frontend-init
make dev
make migrate
```

## Production Build

### Build Production Images

```bash
# Build optimized production images
docker-compose -f docker-compose.prod.yml build
```

### Run Production Stack

```bash
# Start production stack
docker-compose -f docker-compose.prod.yml up -d
```

Production differences:
- Frontend uses nginx instead of dev server
- Smaller image sizes
- Optimized builds
- No hot reload

## Volume Management

### Data Persistence

Docker volumes persist data between container restarts:

- `fjord-pg-data` - PostgreSQL data
- `fjord-redis-data` - Redis data

### Backup Database

```bash
# Backup
docker exec fjord-pg pg_dump -U fjord fjorddb > backup.sql

# Restore
cat backup.sql | docker exec -i fjord-pg psql -U fjord fjorddb
```

### Clear All Data

```bash
# Stop containers and remove volumes
docker-compose down -v

# This will DELETE all data!
```

## Individual Service Mode

If you prefer running services individually (legacy mode):

```bash
# Database only
make create-db
make start-db

# Redis only
make create-redis
make start-redis

# Run API locally
go run ./cmd/api

# Run frontend locally
cd frontend && npm start
```

## Network Configuration

All services run on the `fjord-bank_default` network:

```bash
# Inspect network
docker network inspect fjord-bank_default

# From one container, access another:
# - Frontend can reach: http://api:8080
# - API can reach: postgres:5432, redis:6379
```

## Performance Tips

1. **Increase Docker Resources**
   - Docker Desktop → Settings → Resources
   - Recommended: 4GB RAM, 2 CPUs minimum

2. **Use BuildKit**
   ```bash
   export DOCKER_BUILDKIT=1
   docker-compose build
   ```

3. **Prune Unused Resources**
   ```bash
   docker system prune -a
   ```

## Next Steps

1. Check [fjord_bank_auth_guide.md](fjord_bank_auth_guide.md) for authentication implementation details
2. Explore the API endpoints: http://localhost:8080/health
3. Access the frontend: http://localhost:3000
4. View API documentation (if available): http://localhost:8080/docs

## Support

For issues, check:
- Container logs: `make dev-logs`
- Service status: `make status`
- Docker resources: Docker Desktop → Dashboard
