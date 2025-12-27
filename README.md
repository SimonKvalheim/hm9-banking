# Fjord Bank

A full-stack banking application built for learning Go backend development, React frontend, and financial system patterns.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   React     │────▶│   Go API    │────▶│ PostgreSQL  │
│  Frontend   │     │   (chi)     │     │             │
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                          ▼ (optional)
                   ┌─────────────┐
                   │    Redis    │
                   │   (async)   │
                   └─────────────┘
```

**Key patterns:**
- JWT authentication with access + refresh tokens
- Double-entry bookkeeping for transactions
- Repository pattern for data access
- Middleware chain for cross-cutting concerns

## Quick Start

```bash
# 1. Start PostgreSQL
docker run --name fjord-postgres \
  -e POSTGRES_USER=fjord \
  -e POSTGRES_PASSWORD=fjordpass \
  -e POSTGRES_DB=fjorddb \
  -p 5432:5432 -d postgres

# 2. Run migrations
export DATABASE_URL="postgres://fjord:fjordpass@localhost:5432/fjorddb?sslmode=disable"
goose -dir migrations postgres "$DATABASE_URL" up

# 3. Start API
go run ./cmd/api

# 4. Start frontend (separate terminal)
cd frontend && npm install && npm run dev
```

- **API:** http://localhost:8080
- **Frontend:** http://localhost:5173

## API Endpoints

| Endpoint | Auth | Description |
|----------|------|-------------|
| `POST /auth/register` | No | Create customer account |
| `POST /auth/login` | No | Get access + refresh tokens |
| `POST /auth/refresh` | Cookie | Refresh access token |
| `POST /auth/logout` | Cookie | Clear refresh token |
| `GET /v1/accounts` | JWT | List customer's accounts |
| `POST /v1/accounts` | JWT | Create new account |
| `GET /v1/accounts/{id}` | JWT | Get account details |
| `GET /v1/accounts/{id}/balance` | JWT | Get current balance |
| `POST /v1/transfers` | JWT | Create transfer |
| `GET /v1/transactions/{id}` | JWT | Get transaction status |

## Module Documentation

Each directory contains a README with architecture and design decisions.

> **For AI assistants:** Always read the relevant README files before planning or making changes to a module. Update them after making significant changes.

Modules:

- [cmd/api/](cmd/api/) - API entry point and configuration
- [internal/auth/](internal/auth/) - Authentication service
- [internal/handler/](internal/handler/) - HTTP handlers
- [internal/middleware/](internal/middleware/) - Middleware chain
- [internal/model/](internal/model/) - Domain models
- [internal/repository/](internal/repository/) - Database access
- [internal/processor/](internal/processor/) - Transaction processing
- [migrations/](migrations/) - Database schema
- [frontend/](frontend/) - React application

## Technology Stack

**Backend:** Go 1.21+, chi router, pgx, golang-jwt, bcrypt
**Frontend:** React 18, TypeScript, Vite, Tailwind CSS, Shadcn/ui
**Database:** PostgreSQL 15+
**Optional:** Redis (for async transaction processing)
