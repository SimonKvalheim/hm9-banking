# Middleware

## Purpose

Cross-cutting concerns applied to HTTP requests before reaching handlers. Middleware wraps handlers to add functionality like authentication, logging, and CORS.

## Architecture

```
Middleware chain (applied in order):

  Request → CORS → Logger → Recoverer → Auth → Handler → Response
            │       │         │          │
            │       │         │          └── Validates JWT, injects customer ID
            │       │         └── Catches panics, returns 500
            │       └── Logs request/response (chi built-in)
            └── Sets CORS headers for frontend

Files:
  cors.go  → CORS configuration and middleware
  auth.go  → JWT validation and context injection
```

## Auth Middleware

**What it does:**
1. Extracts `Authorization: Bearer <token>` header
2. Validates JWT signature and expiration via auth service
3. Checks token type is "access" (not refresh)
4. Injects `customer_id` and `customer_email` into request context

**Helper functions for handlers:**
- `GetCustomerID(ctx)` → Returns authenticated customer's UUID
- `GetCustomerEmail(ctx)` → Returns authenticated customer's email

**Responses:**
- Missing header → 401 "Missing authorization header"
- Invalid format → 401 "Invalid authorization header format"
- Invalid/expired token → 401 "Invalid or expired token"
- Wrong token type → 401 "Invalid token type"

## CORS Middleware

**What it does:**
1. Checks if request origin is in allowed list
2. Sets appropriate CORS headers
3. Handles OPTIONS preflight requests

**Configuration:**
- Allowed origins: `localhost:5173`, `localhost:3000`
- Credentials: enabled (for refresh token cookies)
- Allowed headers: `Content-Type`, `Authorization`, `Idempotency-Key`
- Max age: 24 hours (preflight cache)

## Design Decisions

**Why context for customer identity:** Go idiom for request-scoped values. Context flows through the call chain without modifying function signatures. Type-safe keys prevent collisions.

**Why middleware chain order matters:** CORS must run first to handle preflight requests before auth rejects them. Logger wraps everything for complete request logging. Recoverer catches panics from any subsequent middleware or handler.

**Why custom CORS over library:** Explicit control over allowed origins. In production, change from localhost to actual domain—never use `*` with credentials enabled.

**Why typed context keys:** Using `ContextKey` type instead of raw strings prevents accidental key collisions with other packages.
