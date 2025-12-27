# Authentication Service

## Purpose

Handles customer authentication: registration, login, token generation, and validation. Stateless JWT-based auth with dual token strategy.

## Architecture

```
service.go
  ├── Register()        → Validate input, hash password, create customer
  ├── Login()           → Verify credentials, generate token pair
  ├── RefreshTokens()   → Validate refresh token, issue new pair
  ├── ValidateToken()   → Parse and verify JWT signature/expiry
  └── handleFailedLogin() → Track attempts, lock account if needed
```

**Dependencies:**
- `CustomerRepository` for customer data access
- `bcrypt` for password hashing
- `golang-jwt/jwt` for token operations

## Token Strategy

| Token | Lifetime | Storage | Purpose |
|-------|----------|---------|---------|
| Access | 15 min | Client memory | API authorization |
| Refresh | 7 days | HttpOnly cookie | Silent token renewal |

**Flow:**
1. Login returns access token in body, refresh token as cookie
2. Client uses access token in `Authorization: Bearer <token>` header
3. On 401, client calls `/auth/refresh` (cookie sent automatically)
4. Server validates refresh token, returns new access token

## Security Measures

**Password handling:**
- bcrypt with cost factor 10 (2^10 iterations)
- Constant-time comparison prevents timing attacks
- Hash never exposed via JSON serialization

**Brute force protection:**
- Track failed login attempts per customer
- Lock account for 15 minutes after 5 failures
- Generic "invalid credentials" message hides whether email exists

**Token validation:**
- Verify HMAC-SHA256 signature
- Check expiration timestamp
- Validate token type (prevent refresh token used as access)

## Design Decisions

**Why JWT over sessions:** Stateless tokens scale horizontally without shared session storage. Good for learning distributed system patterns.

**Why dual tokens:** Short-lived access tokens limit damage if stolen. Long-lived refresh tokens in HttpOnly cookies provide persistence without exposing tokens to JavaScript (XSS protection).

**Why bcrypt over argon2:** bcrypt is battle-tested and sufficient for this use case. Argon2 is newer and arguably better, but adds complexity without meaningful security benefit here.

**Why generic error messages:** Revealing "email not found" vs "wrong password" helps attackers enumerate valid accounts.
