# HTTP Handlers

## Purpose

Translate HTTP requests into service/repository calls and return JSON responses. Handlers own request validation, authorization checks, and response formatting—but not business logic.

## Architecture

```
handler/
  ├── account.go   → Account CRUD, balance queries
  ├── transfer.go  → Transfer creation, transaction status
  └── auth.go      → Register, login, refresh, logout
```

Each handler:
1. Extracts customer ID from context (set by auth middleware)
2. Parses and validates request
3. Checks authorization (ownership)
4. Calls repository/service
5. Formats and returns response

## Handlers

### AccountHandler
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/accounts` | POST | Create account for authenticated customer |
| `/accounts` | GET | List customer's accounts only |
| `/accounts/{id}` | GET | Get account (must own it) |
| `/accounts/{id}/balance` | GET | Get balance, optional `?as_of=` for point-in-time |

### TransferHandler
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/transfers` | POST | Create transfer (requires `Idempotency-Key` header) |
| `/transactions/{id}` | GET | Get transaction status |

### AuthHandler
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/register` | POST | Create customer account |
| `/auth/login` | POST | Get access + refresh tokens |
| `/auth/refresh` | POST | Renew access token via cookie |
| `/auth/logout` | POST | Clear refresh token cookie |

## Authorization Rules

| Action | Rule |
|--------|------|
| List accounts | Only customer's own accounts |
| Get account | Must own the account |
| Create transfer | Source account must be owned by customer |
| View transaction | Must involve customer's account (source or destination) |

Unauthorized access returns 403 Forbidden.

## Idempotency

Transfer creation requires an `Idempotency-Key` header. If a request is retried with the same key:
- Returns existing transaction if found
- Handles race conditions (duplicate key error → fetch existing)

This prevents duplicate transfers from network retries or client bugs.

## Design Decisions

**Why 202 Accepted for transfers:** Transfers may process asynchronously. 202 signals "accepted for processing" regardless of sync/async mode. Client should poll `/transactions/{id}` for final status.

**Why handlers don't contain business logic:** Separation of concerns. Handlers deal with HTTP; repositories deal with data; services deal with business rules. Makes testing easier.

**Why check authorization after fetching:** Need the resource to check ownership. Could optimize with a single query, but clarity is preferred for learning.
