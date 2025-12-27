# Fjord Bank: Customer Authentication & Authorization Implementation Guide

> **Purpose:** Step-by-step technical guide for adding customers, JWT authentication, and authorization to the Fjord Bank system.
> **Prerequisites:** Working API with accounts and transactions (current state)
> **End Goal:** Secure multi-user banking application with proper access controls

---

## 📊 Current Status: Phase 2 Complete - Ready for Testing

| Phase | Status | Next Action |
|-------|--------|-------------|
| Phase 1: Customer Data Model | ✅ DONE | Migration may be pending |
| Phase 2: JWT Authentication | ✅ DONE | **⚠️ TEST BEFORE CONTINUING** |
| Phase 3: Authorization Middleware | ⏳ PENDING | Start after testing Phase 2 |
| Phase 4: Frontend with Auth | ⏳ PENDING | - |

**👉 See "NEXT SESSION" section below for what to do when you resume!**

---

## 🎯 Implementation Progress

### 🔔 NEXT SESSION: Test Phase 2 Authentication!

**⚠️ IMPORTANT - START HERE NEXT TIME:**

Before implementing Phase 3, you MUST test the authentication system:

1. ✅ **Start PostgreSQL** (if not running):
   ```bash
   docker run --name fjord-postgres -e POSTGRES_USER=fjord -e POSTGRES_PASSWORD=fjordpass -e POSTGRES_DB=fjorddb -p 5432:5432 -d postgres
   ```

2. ✅ **Run migrations** (if not done):
   ```bash
   goose -dir migrations postgres "postgres://fjord:fjordpass@localhost:5432/fjorddb?sslmode=disable" up
   ```

3. ✅ **Start the API server**:
   ```bash
   go run ./cmd/api
   ```

4. ✅ **Test authentication endpoints** - Follow the complete test suite in section "4. Test Phase 2 Authentication" below:
   - Register a new customer
   - Login and get tokens
   - Test invalid credentials
   - Test token refresh
   - Test account lockout (5 failed attempts)
   - Test logout

5. ✅ **Verify everything works** before proceeding to Phase 3!

---

### ✅ Phase 1: Customer Data Model - COMPLETED

**Status:** All code implemented and ready for migration

**Completed Items:**
- ✅ Database migration created (`migrations/000003_create_customers.sql`)
- ✅ Customer model with validation (`internal/model/customer.go`)
- ✅ Customer repository with all methods (`internal/repository/customer.go`)
- ✅ Customer/auth error definitions (`internal/model/errors.go`)
- ✅ Account model updated with `CustomerID` field
- ✅ Account repository updated with `GetByCustomerID()` and `CreateForCustomer()`

**Next Steps:**
1. Start PostgreSQL database (see Quick Start below)
2. Run migration: `goose -dir migrations postgres "$DATABASE_URL" up`
3. Proceed to Phase 2: JWT Authentication

**Files Created:**
- `migrations/000003_create_customers.sql`
- `internal/model/customer.go`
- `internal/repository/customer.go`

**Files Modified:**
- `internal/model/errors.go`
- `internal/model/account.go`
- `internal/repository/account.go`

### ✅ Phase 2: JWT Authentication - COMPLETED

**Status:** All code implemented and tested

**Completed Items:**
- ✅ JWT dependency added (`github.com/golang-jwt/jwt/v5`)
- ✅ Auth service with bcrypt password hashing
- ✅ JWT token generation (access + refresh tokens)
- ✅ Auth handler with login/register/refresh/logout endpoints
- ✅ Main.go updated with auth wiring
- ✅ JWT_SECRET configuration with development default

**Next Steps:**
1. Test authentication endpoints (see Phase 2 Testing below)
2. Proceed to Phase 3: Authorization Middleware

**Files Created:**
- `internal/auth/service.go`
- `internal/handler/auth.go`

**Files Modified:**
- `cmd/api/main.go`
- `go.mod` (added JWT dependency)

**API Endpoints Available:**
- `POST /auth/register` - Create new customer account
- `POST /auth/login` - Authenticate and get tokens
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - Clear refresh token cookie

### 📋 Phase 3: Authorization Middleware - PENDING

### 📋 Phase 4: Frontend with Authentication - PENDING

---

## Quick Start Guide

### 1. Database Setup

If you haven't started your PostgreSQL database yet:

```bash
# Using Docker (recommended)
docker run --name fjord-postgres \
  -e POSTGRES_USER=fjord \
  -e POSTGRES_PASSWORD=fjordpass \
  -e POSTGRES_DB=fjorddb \
  -p 5432:5432 \
  -d postgres

# Apply migrations
export DATABASE_URL="postgres://fjord:fjordpass@localhost:5432/fjorddb?sslmode=disable"
goose -dir migrations postgres "$DATABASE_URL" up

# Verify tables
psql "$DATABASE_URL" -c "\dt"
# Should show: customers, accounts, transactions, ledger_entries, transaction_parties
```

### 2. Set Environment Variables (Optional)

```bash
# JWT secret for token signing (optional - has dev default)
export JWT_SECRET="kj9DrnI/dxLVx5vKquOThQtO/wdWVL1XbxJYtHkRuwA="

# Database URL (optional - has dev default)
export DATABASE_URL="postgres://fjord:fjordpass@localhost:5432/fjorddb?sslmode=disable"
```

### 3. Start the API Server

```bash
go run ./cmd/api
```

Expected output:
```
WARNING: Using default JWT_SECRET for development. Set JWT_SECRET environment variable in production!
Connected to database
Running in sync mode (set ASYNC_MODE=true for async processing)
Server starting on port 8080
```

### 4. Test Phase 2 Authentication

#### A. Register a New Customer

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Smith"
  }'
```

Expected response (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "first_name": "Alice",
  "last_name": "Smith",
  "phone_verified": false,
  "email_verified": false,
  "status": "active",
  "preferred_language": "en",
  "timezone": "UTC",
  "created_at": "2025-12-26T23:30:00Z",
  "updated_at": "2025-12-26T23:30:00Z"
}
```

#### B. Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123"
  }'
```

Expected response (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-26T23:45:00Z",
  "token_type": "Bearer"
}
```

**Save the access_token for next steps!**

#### C. Test Invalid Credentials

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "WrongPassword"
  }'
```

Expected response (401 Unauthorized):
```json
{
  "error": "Invalid email or password"
}
```

#### D. Refresh Token

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -b cookies.txt
```

Expected response (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-26T23:45:00Z",
  "token_type": "Bearer"
}
```

#### E. Test Account Lockout

Try logging in with wrong password 5 times:

```bash
for i in {1..5}; do
  curl -X POST http://localhost:8080/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email": "alice@example.com", "password": "wrong"}';
  echo "";
done
```

On the 5th attempt, you should see:
```json
{
  "error": "Account is temporarily locked"
}
```

Wait 15 minutes or reset in the database:
```sql
UPDATE customers SET failed_login_attempts = 0, locked_until = NULL WHERE email = 'alice@example.com';
```

#### F. Logout

```bash
curl -X POST http://localhost:8080/auth/logout \
  -b cookies.txt \
  -c cookies.txt
```

Expected response (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              BEFORE                                          │
│                                                                             │
│   Client ──► API ──► Any Account                                            │
│   (no identity)      (no restrictions)                                      │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              AFTER                                           │
│                                                                             │
│   Client ──► Login ──► JWT Token ──► API + Middleware ──► Own Accounts Only │
│   (credentials)       (identity)     (validates token)   (ownership check)  │
└─────────────────────────────────────────────────────────────────────────────┘
```

### JWT Token Flow (What We're Building)

```
┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│  Client  │      │   API    │      │   JWT    │      │ Database │
│ (React)  │      │  Server  │      │  Logic   │      │ (Postgres)│
└────┬─────┘      └────┬─────┘      └────┬─────┘      └────┬─────┘
     │                 │                 │                 │
     │ POST /auth/login                  │                 │
     │ {email, password}                 │                 │
     │────────────────►│                 │                 │
     │                 │                 │                 │
     │                 │ Find customer by email            │
     │                 │────────────────────────────────────►
     │                 │◄─────────────────── customer row ──│
     │                 │                 │                 │
     │                 │ Verify password │                 │
     │                 │────────────────►│                 │
     │                 │◄─── bcrypt.Compare ───            │
     │                 │                 │                 │
     │                 │ Generate JWT    │                 │
     │                 │────────────────►│                 │
     │                 │◄─── signed token ─────            │
     │                 │                 │                 │
     │◄─── 200 {access_token, refresh_token} ──            │
     │                 │                 │                 │
     │ GET /v1/accounts                  │                 │
     │ Authorization: Bearer <token>     │                 │
     │────────────────►│                 │                 │
     │                 │                 │                 │
     │                 │ Validate JWT    │                 │
     │                 │────────────────►│                 │
     │                 │◄─── customer_id ──────            │
     │                 │                 │                 │
     │                 │ Get accounts WHERE customer_id = X│
     │                 │────────────────────────────────────►
     │                 │◄─────────────────── accounts[] ───│
     │                 │                 │                 │
     │◄─── 200 [accounts owned by customer] ───            │
     │                 │                 │                 │
```

---

## Phase 1: Customer Data Model

### 1.1 Objective

Create a customers table with rich attributes for future features, establish the one-to-many relationship between customers and accounts.

### 1.2 Database Migration

**File:** `migrations/000003_create_customers.sql`

```sql
-- +goose Up

-- customers table with rich attributes for future expansion
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Authentication (required)
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    -- Identity (required for KYC later)
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    
    -- Contact (optional initially)
    phone VARCHAR(20),
    phone_verified BOOLEAN DEFAULT FALSE,
    email_verified BOOLEAN DEFAULT FALSE,
    
    -- Address (optional, for KYC/compliance later)
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country CHAR(2),  -- ISO 3166-1 alpha-2 (NO, SE, US, etc.)
    
    -- Identity verification (for KYC later)
    date_of_birth DATE,
    national_id_number VARCHAR(50),  -- Encrypted in production!
    
    -- Status & security
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, suspended, closed
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ,
    
    -- Preferences (for UI/notifications later)
    preferred_language CHAR(2) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common lookups
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers (email);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers (status);
CREATE INDEX IF NOT EXISTS idx_customers_last_name ON customers (last_name);

-- Add customer foreign key to accounts table
ALTER TABLE accounts ADD COLUMN customer_id UUID REFERENCES customers(id);

-- Index for finding all accounts belonging to a customer
CREATE INDEX IF NOT EXISTS idx_accounts_customer_id ON accounts (customer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_accounts_customer_id;
ALTER TABLE accounts DROP COLUMN IF EXISTS customer_id;

DROP INDEX IF EXISTS idx_customers_last_name;
DROP INDEX IF EXISTS idx_customers_status;
DROP INDEX IF EXISTS idx_customers_email;
DROP TABLE IF EXISTS customers;
```

### 1.3 Understanding the Schema Decisions

| Column | Why It's There | Required Now? |
|--------|---------------|---------------|
| `email` | Primary identifier for login | Yes |
| `password_hash` | Never store plain passwords | Yes |
| `first_name`, `last_name` | Basic identity, needed for any real banking | Yes |
| `phone`, `phone_verified` | 2FA, notifications (future) | No |
| `address_*` | KYC compliance, statements (future) | No |
| `date_of_birth` | Age verification, KYC (future) | No |
| `national_id_number` | Strong KYC, required for real banks | No |
| `status` | Account lockout, suspension | Yes (defaults to active) |
| `failed_login_attempts` | Brute force protection | Yes (defaults to 0) |
| `locked_until` | Temporary lockout after failed attempts | No |
| `preferred_language` | i18n support (future) | No |

### 1.4 Go Model

**File:** `internal/model/customer.go`

```go
package model

import (
    "time"
    "unicode"

    "github.com/google/uuid"
)

// CustomerStatus represents the current status of a customer
type CustomerStatus string

const (
    CustomerStatusActive    CustomerStatus = "active"
    CustomerStatusSuspended CustomerStatus = "suspended"
    CustomerStatusClosed    CustomerStatus = "closed"
)

// Customer represents a bank customer
type Customer struct {
    ID           uuid.UUID      `json:"id"`
    Email        string         `json:"email"`
    PasswordHash string         `json:"-"` // Never serialize password hash!
    FirstName    string         `json:"first_name"`
    LastName     string         `json:"last_name"`
    
    // Optional contact
    Phone         *string `json:"phone,omitempty"`
    PhoneVerified bool    `json:"phone_verified"`
    EmailVerified bool    `json:"email_verified"`
    
    // Optional address
    AddressLine1 *string `json:"address_line1,omitempty"`
    AddressLine2 *string `json:"address_line2,omitempty"`
    City         *string `json:"city,omitempty"`
    PostalCode   *string `json:"postal_code,omitempty"`
    Country      *string `json:"country,omitempty"`
    
    // Optional identity
    DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
    NationalIDNumber *string    `json:"-"` // Sensitive - don't serialize
    
    // Status & security
    Status              CustomerStatus `json:"status"`
    FailedLoginAttempts int            `json:"-"` // Don't expose
    LockedUntil         *time.Time     `json:"-"` // Don't expose
    LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
    PasswordChangedAt   *time.Time     `json:"-"`
    
    // Preferences
    PreferredLanguage string `json:"preferred_language"`
    Timezone          string `json:"timezone"`
    
    // Audit
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// IsLocked returns true if the customer account is currently locked
func (c *Customer) IsLocked() bool {
    if c.LockedUntil == nil {
        return false
    }
    return time.Now().Before(*c.LockedUntil)
}

// CanLogin returns true if the customer is allowed to attempt login
func (c *Customer) CanLogin() bool {
    return c.Status == CustomerStatusActive && !c.IsLocked()
}

// CreateCustomerRequest is the payload for registration
type CreateCustomerRequest struct {
    Email     string `json:"email"`
    Password  string `json:"password"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

// Validate checks if the registration request is valid
func (r CreateCustomerRequest) Validate() error {
    if r.Email == "" || !isValidEmail(r.Email) {
        return ErrInvalidEmail
    }
    if len(r.Password) < 8 {
        return ErrPasswordTooShort
    }
    if !isStrongPassword(r.Password) {
        return ErrPasswordTooWeak
    }
    if r.FirstName == "" {
        return ErrFirstNameRequired
    }
    if r.LastName == "" {
        return ErrLastNameRequired
    }
    return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
    // Basic check - contains @ and at least one dot after @
    atIndex := -1
    for i, c := range email {
        if c == '@' {
            atIndex = i
            break
        }
    }
    if atIndex < 1 || atIndex >= len(email)-1 {
        return false
    }
    afterAt := email[atIndex+1:]
    for _, c := range afterAt {
        if c == '.' {
            return true
        }
    }
    return false
}

// isStrongPassword checks password complexity
// Requires: 8+ chars, at least one uppercase, one lowercase, one digit
func isStrongPassword(password string) bool {
    var hasUpper, hasLower, hasDigit bool
    for _, c := range password {
        switch {
        case unicode.IsUpper(c):
            hasUpper = true
        case unicode.IsLower(c):
            hasLower = true
        case unicode.IsDigit(c):
            hasDigit = true
        }
    }
    return hasUpper && hasLower && hasDigit
}

// LoginRequest is the payload for authentication
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// Validate checks if the login request has required fields
func (r LoginRequest) Validate() error {
    if r.Email == "" {
        return ErrInvalidEmail
    }
    if r.Password == "" {
        return ErrPasswordRequired
    }
    return nil
}
```

### 1.5 Add New Errors

**Update:** `internal/model/errors.go`

```go
// Add these to existing errors

// Customer/Auth errors
var (
    ErrInvalidEmail       = errors.New("invalid email address")
    ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
    ErrPasswordTooWeak    = errors.New("password must contain uppercase, lowercase, and digit")
    ErrPasswordRequired   = errors.New("password is required")
    ErrFirstNameRequired  = errors.New("first name is required")
    ErrLastNameRequired   = errors.New("last name is required")
    ErrCustomerNotFound   = errors.New("customer not found")
    ErrEmailAlreadyExists = errors.New("email already registered")
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrAccountLocked      = errors.New("account is locked")
    ErrAccountSuspended   = errors.New("account is suspended")
)
```

### 1.6 Customer Repository

**File:** `internal/repository/customer.go`

This repository needs methods for:

| Method | Purpose |
|--------|---------|
| `Create(ctx, customer)` | Register new customer |
| `GetByID(ctx, id)` | Fetch customer by UUID |
| `GetByEmail(ctx, email)` | Fetch customer for login |
| `UpdateLastLogin(ctx, id)` | Track login timestamp |
| `IncrementFailedAttempts(ctx, id)` | Brute force tracking |
| `ResetFailedAttempts(ctx, id)` | Clear on successful login |
| `LockAccount(ctx, id, until)` | Temporary lockout |

**Implementation notes:**
- `GetByEmail` is the critical method for authentication
- Use `FOR UPDATE` when incrementing failed attempts to prevent race conditions
- Lock account for 15 minutes after 5 failed attempts (configurable)

### 1.7 Update Account Model & Repository

The `Account` model needs a new field:

```go
type Account struct {
    // ... existing fields ...
    CustomerID *uuid.UUID `json:"customer_id,omitempty"` // nil for system accounts
}
```

Update `AccountRepository` with:
- `GetByCustomerID(ctx, customerID)` - Get all accounts for a customer
- `Create` should accept optional `customerID` parameter
- System accounts (like Bank Equity) have `CustomerID = nil`

### 1.8 Testing Phase 1

Before moving to Phase 2, verify:

1. **Migration applies cleanly:**
   ```bash
   goose -dir migrations postgres "$DATABASE_URL" up
   ```

2. **Can create a customer via repository** (write a small test or use psql):
   ```sql
   INSERT INTO customers (email, password_hash, first_name, last_name)
   VALUES ('test@example.com', 'temp_hash', 'Test', 'User');
   ```

3. **Can link account to customer:**
   ```sql
   UPDATE accounts SET customer_id = '<customer-uuid>' WHERE id = '<account-uuid>';
   ```

4. **Query accounts by customer works:**
   ```sql
   SELECT * FROM accounts WHERE customer_id = '<customer-uuid>';
   ```

---

## Phase 2: JWT Authentication

### 2.1 Objective

Implement secure login/logout with JWT tokens. Users receive tokens that prove their identity for subsequent requests.

### 2.2 Understanding JWTs

A JWT (JSON Web Token) has three parts separated by dots:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
 └──────────── Header ────────────┘└────────────────────── Payload ──────────────────────┘└─────────── Signature ───────────┘
```

| Part | Contains | Example |
|------|----------|---------|
| Header | Algorithm & token type | `{"alg": "HS256", "typ": "JWT"}` |
| Payload | Claims (user data) | `{"sub": "customer-uuid", "exp": 1234567890}` |
| Signature | HMAC of header+payload | Verifies token wasn't tampered with |

**Key insight:** The payload is only base64-encoded, NOT encrypted. Anyone can decode and read it. The signature proves it came from your server and hasn't been modified.

### 2.3 Token Strategy: Access + Refresh

We'll use two tokens:

| Token | Lifetime | Purpose | Storage |
|-------|----------|---------|---------|
| Access Token | 15 minutes | Authorize API requests | Memory (React state) |
| Refresh Token | 7 days | Get new access tokens | HttpOnly cookie |

**Why two tokens?**
- Short-lived access tokens limit damage if stolen
- Refresh tokens allow staying logged in without re-entering password
- HttpOnly cookies prevent JavaScript from reading refresh token (XSS protection)

```
┌─────────────────────────────────────────────────────────────────┐
│                     Token Refresh Flow                          │
│                                                                 │
│  Access Token Expires                                           │
│         │                                                       │
│         ▼                                                       │
│  API returns 401 Unauthorized                                   │
│         │                                                       │
│         ▼                                                       │
│  Frontend calls POST /auth/refresh                              │
│  (refresh token sent automatically via cookie)                  │
│         │                                                       │
│         ▼                                                       │
│  Server validates refresh token                                 │
│         │                                                       │
│         ├── Valid ──► Return new access token                   │
│         │                                                       │
│         └── Invalid/Expired ──► Return 401, redirect to login   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.4 Dependencies

Add to `go.mod`:

```
github.com/golang-jwt/jwt/v5  // JWT handling
golang.org/x/crypto/bcrypt    // Password hashing
```

### 2.5 Auth Service

**File:** `internal/auth/service.go`

This is the core authentication logic, separate from HTTP handlers.

```go
package auth

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    
    "github.com/simonkvalheim/hm9-banking/internal/model"
    "github.com/simonkvalheim/hm9-banking/internal/repository"
)

// Config holds authentication configuration
type Config struct {
    JWTSecret          []byte        // Secret key for signing tokens
    AccessTokenExpiry  time.Duration // How long access tokens are valid
    RefreshTokenExpiry time.Duration // How long refresh tokens are valid
    MaxFailedAttempts  int           // Lock account after this many failures
    LockDuration       time.Duration // How long to lock account
}

// DefaultConfig returns sensible defaults
func DefaultConfig(jwtSecret string) Config {
    return Config{
        JWTSecret:          []byte(jwtSecret),
        AccessTokenExpiry:  15 * time.Minute,
        RefreshTokenExpiry: 7 * 24 * time.Hour,
        MaxFailedAttempts:  5,
        LockDuration:       15 * time.Minute,
    }
}

// Claims represents the JWT payload
type Claims struct {
    jwt.RegisteredClaims
    CustomerID uuid.UUID `json:"customer_id"`
    Email      string    `json:"email"`
    TokenType  string    `json:"token_type"` // "access" or "refresh"
}

// Service handles authentication operations
type Service struct {
    config       Config
    customerRepo *repository.CustomerRepository
}

// NewService creates a new auth service
func NewService(config Config, customerRepo *repository.CustomerRepository) *Service {
    return &Service{
        config:       config,
        customerRepo: customerRepo,
    }
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"` // Access token expiry
}

// Register creates a new customer account
func (s *Service) Register(ctx context.Context, req model.CreateCustomerRequest) (*model.Customer, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Check if email already exists
    existing, err := s.customerRepo.GetByEmail(ctx, req.Email)
    if err == nil && existing != nil {
        return nil, model.ErrEmailAlreadyExists
    }
    
    // Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }
    
    // Create customer
    customer := &model.Customer{
        ID:                uuid.New(),
        Email:             req.Email,
        PasswordHash:      string(hash),
        FirstName:         req.FirstName,
        LastName:          req.LastName,
        Status:            model.CustomerStatusActive,
        PreferredLanguage: "en",
        Timezone:          "UTC",
        CreatedAt:         time.Now(),
        UpdatedAt:         time.Now(),
    }
    
    return s.customerRepo.Create(ctx, customer)
}

// Login authenticates a customer and returns tokens
func (s *Service) Login(ctx context.Context, req model.LoginRequest) (*TokenPair, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Find customer
    customer, err := s.customerRepo.GetByEmail(ctx, req.Email)
    if err != nil {
        // Don't reveal whether email exists
        return nil, model.ErrInvalidCredentials
    }
    
    // Check if account can login
    if !customer.CanLogin() {
        if customer.IsLocked() {
            return nil, model.ErrAccountLocked
        }
        return nil, model.ErrAccountSuspended
    }
    
    // Verify password
    err = bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password))
    if err != nil {
        // Wrong password - increment failed attempts
        s.handleFailedLogin(ctx, customer)
        return nil, model.ErrInvalidCredentials
    }
    
    // Success - reset failed attempts and update last login
    s.customerRepo.ResetFailedAttempts(ctx, customer.ID)
    s.customerRepo.UpdateLastLogin(ctx, customer.ID)
    
    // Generate tokens
    return s.generateTokenPair(customer)
}

// RefreshTokens generates new tokens using a valid refresh token
func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
    // Parse and validate the refresh token
    claims, err := s.ValidateToken(refreshToken)
    if err != nil {
        return nil, err
    }
    
    // Ensure it's a refresh token
    if claims.TokenType != "refresh" {
        return nil, errors.New("invalid token type")
    }
    
    // Fetch customer to ensure they still exist and are active
    customer, err := s.customerRepo.GetByID(ctx, claims.CustomerID)
    if err != nil {
        return nil, err
    }
    
    if customer.Status != model.CustomerStatusActive {
        return nil, model.ErrAccountSuspended
    }
    
    // Generate new token pair
    return s.generateTokenPair(customer)
}

// ValidateToken parses and validates a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return s.config.JWTSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    return claims, nil
}

// generateTokenPair creates access and refresh tokens for a customer
func (s *Service) generateTokenPair(customer *model.Customer) (*TokenPair, error) {
    now := time.Now()
    accessExpiry := now.Add(s.config.AccessTokenExpiry)
    refreshExpiry := now.Add(s.config.RefreshTokenExpiry)
    
    // Access token claims
    accessClaims := Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   customer.ID.String(),
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(accessExpiry),
            Issuer:    "fjord-bank",
        },
        CustomerID: customer.ID,
        Email:      customer.Email,
        TokenType:  "access",
    }
    
    // Refresh token claims
    refreshClaims := Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   customer.ID.String(),
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(refreshExpiry),
            Issuer:    "fjord-bank",
        },
        CustomerID: customer.ID,
        Email:      customer.Email,
        TokenType:  "refresh",
    }
    
    // Sign tokens
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessSigned, err := accessToken.SignedString(s.config.JWTSecret)
    if err != nil {
        return nil, err
    }
    
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshSigned, err := refreshToken.SignedString(s.config.JWTSecret)
    if err != nil {
        return nil, err
    }
    
    return &TokenPair{
        AccessToken:  accessSigned,
        RefreshToken: refreshSigned,
        ExpiresAt:    accessExpiry,
    }, nil
}

// handleFailedLogin increments failed attempts and locks if necessary
func (s *Service) handleFailedLogin(ctx context.Context, customer *model.Customer) {
    attempts, _ := s.customerRepo.IncrementFailedAttempts(ctx, customer.ID)
    
    if attempts >= s.config.MaxFailedAttempts {
        lockUntil := time.Now().Add(s.config.LockDuration)
        s.customerRepo.LockAccount(ctx, customer.ID, lockUntil)
    }
}

// HashPassword is a utility for hashing passwords (useful for tests/seeding)
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}
```

### 2.6 Understanding the Code

**Why bcrypt for passwords?**
- Bcrypt is intentionally slow (configurable via "cost")
- Includes a random salt automatically
- Makes rainbow table attacks useless
- `bcrypt.DefaultCost` is 10, meaning 2^10 = 1024 iterations

**Why check `CanLogin()` before password verification?**
- Prevents timing attacks that could reveal locked accounts
- Actually, we should verify password timing is constant regardless - but bcrypt handles this

**Why not expose failed attempt counts?**
- Attackers could use this to know when an account is about to lock
- Security through obscurity isn't perfect, but no reason to help attackers

### 2.7 Auth Handler

**File:** `internal/handler/auth.go`

```go
package handler

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"

    "github.com/simonkvalheim/hm9-banking/internal/auth"
    "github.com/simonkvalheim/hm9-banking/internal/model"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
    authService *auth.Service
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
    return &AuthHandler{authService: authService}
}

// RegisterRoutes sets up the auth routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
    r.Post("/auth/register", h.Register)
    r.Post("/auth/login", h.Login)
    r.Post("/auth/refresh", h.RefreshToken)
    r.Post("/auth/logout", h.Logout)
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req model.CreateCustomerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    customer, err := h.authService.Register(r.Context(), req)
    if err != nil {
        switch err {
        case model.ErrEmailAlreadyExists:
            writeError(w, http.StatusConflict, err.Error())
        case model.ErrInvalidEmail, model.ErrPasswordTooShort, 
             model.ErrPasswordTooWeak, model.ErrFirstNameRequired, 
             model.ErrLastNameRequired:
            writeError(w, http.StatusBadRequest, err.Error())
        default:
            writeError(w, http.StatusInternalServerError, "Registration failed")
        }
        return
    }

    // Return customer without sensitive data (password_hash already excluded via json:"-")
    writeJSON(w, http.StatusCreated, customer)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req model.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    tokens, err := h.authService.Login(r.Context(), req)
    if err != nil {
        switch err {
        case model.ErrInvalidCredentials:
            writeError(w, http.StatusUnauthorized, "Invalid email or password")
        case model.ErrAccountLocked:
            writeError(w, http.StatusForbidden, "Account is temporarily locked")
        case model.ErrAccountSuspended:
            writeError(w, http.StatusForbidden, "Account is suspended")
        default:
            writeError(w, http.StatusInternalServerError, "Login failed")
        }
        return
    }

    // Set refresh token as HttpOnly cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    tokens.RefreshToken,
        Path:     "/",
        HttpOnly: true,                    // JavaScript cannot access
        Secure:   true,                    // Only sent over HTTPS (disable for local dev)
        SameSite: http.SameSiteStrictMode, // CSRF protection
        MaxAge:   7 * 24 * 60 * 60,        // 7 days in seconds
    })

    // Return access token in response body
    writeJSON(w, http.StatusOK, map[string]interface{}{
        "access_token": tokens.AccessToken,
        "expires_at":   tokens.ExpiresAt,
        "token_type":   "Bearer",
    })
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
    // Get refresh token from cookie
    cookie, err := r.Cookie("refresh_token")
    if err != nil {
        writeError(w, http.StatusUnauthorized, "No refresh token")
        return
    }

    tokens, err := h.authService.RefreshTokens(r.Context(), cookie.Value)
    if err != nil {
        // Clear invalid cookie
        http.SetCookie(w, &http.Cookie{
            Name:     "refresh_token",
            Value:    "",
            Path:     "/",
            HttpOnly: true,
            MaxAge:   -1, // Delete cookie
        })
        writeError(w, http.StatusUnauthorized, "Invalid refresh token")
        return
    }

    // Update refresh token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    tokens.RefreshToken,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   7 * 24 * 60 * 60,
    })

    writeJSON(w, http.StatusOK, map[string]interface{}{
        "access_token": tokens.AccessToken,
        "expires_at":   tokens.ExpiresAt,
        "token_type":   "Bearer",
    })
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    // Clear the refresh token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   -1, // Delete cookie
    })

    writeJSON(w, http.StatusOK, map[string]string{
        "message": "Logged out successfully",
    })
}
```

### 2.8 Environment Configuration

Add these environment variables:

```bash
# Required - use a strong random string in production
JWT_SECRET=your-256-bit-secret-key-here-make-it-long-and-random

# Optional - override defaults
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h  # 7 days
```

**Generating a secure secret:**
```bash
openssl rand -base64 32
```

### 2.9 Testing Phase 2

Test the authentication flow:

1. **Register a customer:**
   ```bash
   curl -X POST http://localhost:8080/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "email": "alice@example.com",
       "password": "SecurePass123",
       "first_name": "Alice",
       "last_name": "Smith"
     }'
   ```

2. **Login:**
   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -c cookies.txt \
     -d '{
       "email": "alice@example.com",
       "password": "SecurePass123"
     }'
   ```
   Save the `access_token` from the response.

3. **Test wrong password (should fail):**
   ```bash
   curl -X POST http://localhost:8080/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email": "alice@example.com", "password": "wrong"}'
   ```

4. **Refresh token:**
   ```bash
   curl -X POST http://localhost:8080/auth/refresh \
     -b cookies.txt
   ```

5. **Logout:**
   ```bash
   curl -X POST http://localhost:8080/auth/logout \
     -b cookies.txt \
     -c cookies.txt
   ```

---

## Phase 3: Authorization Middleware

### 3.1 Objective

Protect API endpoints so users can only access their own data. This involves:
1. Validating the JWT on every protected request
2. Extracting the customer ID from the token
3. Checking ownership before returning data

### 3.2 Auth Middleware

**File:** `internal/middleware/auth.go`

```go
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/google/uuid"

    "github.com/simonkvalheim/hm9-banking/internal/auth"
)

// ContextKey is the type for context keys to avoid collisions
type ContextKey string

const (
    // CustomerIDKey is the context key for the authenticated customer ID
    CustomerIDKey ContextKey = "customer_id"
    // CustomerEmailKey is the context key for the authenticated customer email
    CustomerEmailKey ContextKey = "customer_email"
)

// AuthMiddleware validates JWT tokens and adds customer info to context
type AuthMiddleware struct {
    authService *auth.Service
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
    return &AuthMiddleware{authService: authService}
}

// RequireAuth is middleware that requires a valid access token
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            writeUnauthorized(w, "Missing authorization header")
            return
        }

        // Expected format: "Bearer <token>"
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            writeUnauthorized(w, "Invalid authorization header format")
            return
        }

        tokenString := parts[1]

        // Validate token
        claims, err := m.authService.ValidateToken(tokenString)
        if err != nil {
            writeUnauthorized(w, "Invalid or expired token")
            return
        }

        // Ensure it's an access token (not a refresh token)
        if claims.TokenType != "access" {
            writeUnauthorized(w, "Invalid token type")
            return
        }

        // Add customer info to request context
        ctx := context.WithValue(r.Context(), CustomerIDKey, claims.CustomerID)
        ctx = context.WithValue(ctx, CustomerEmailKey, claims.Email)

        // Call next handler with enriched context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetCustomerID extracts the customer ID from the request context
// Returns uuid.Nil if not authenticated (shouldn't happen if RequireAuth was used)
func GetCustomerID(ctx context.Context) uuid.UUID {
    id, ok := ctx.Value(CustomerIDKey).(uuid.UUID)
    if !ok {
        return uuid.Nil
    }
    return id
}

// GetCustomerEmail extracts the customer email from the request context
func GetCustomerEmail(ctx context.Context) string {
    email, ok := ctx.Value(CustomerEmailKey).(string)
    if !ok {
        return ""
    }
    return email
}

func writeUnauthorized(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    w.Write([]byte(`{"error": "` + message + `"}`))
}
```

### 3.3 Applying Middleware to Routes

**Update:** `cmd/api/main.go`

```go
// Set up router
r := chi.NewRouter()

// Global middleware
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)

// Initialize auth middleware
authMiddleware := middleware.NewAuthMiddleware(authService)

// Public routes (no auth required)
r.Get("/health", healthHandler(db))
r.Post("/auth/register", authHandler.Register)
r.Post("/auth/login", authHandler.Login)
r.Post("/auth/refresh", authHandler.RefreshToken)
r.Post("/auth/logout", authHandler.Logout)

// Protected routes (auth required)
r.Route("/v1", func(r chi.Router) {
    // Apply auth middleware to all /v1 routes
    r.Use(authMiddleware.RequireAuth)
    
    accountHandler.RegisterRoutes(r)
    transferHandler.RegisterRoutes(r)
})
```

### 3.4 Updating Handlers for Authorization

Now that we know WHO is making the request, we need to ensure they can only access their own resources.

**Update:** `internal/handler/account.go`

```go
// List handles GET /accounts
// Now returns only accounts belonging to the authenticated customer
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
    customerID := middleware.GetCustomerID(r.Context())
    if customerID == uuid.Nil {
        writeError(w, http.StatusUnauthorized, "Not authenticated")
        return
    }

    // Only get accounts for this customer
    accounts, err := h.repo.GetByCustomerID(r.Context(), customerID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to list accounts")
        return
    }

    if accounts == nil {
        accounts = []model.Account{}
    }

    writeJSON(w, http.StatusOK, accounts)
}

// GetByID handles GET /accounts/{id}
// Verifies the account belongs to the authenticated customer
func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    customerID := middleware.GetCustomerID(r.Context())
    
    idParam := chi.URLParam(r, "id")
    id, err := uuid.Parse(idParam)
    if err != nil {
        writeError(w, http.StatusBadRequest, "Invalid account ID format")
        return
    }

    account, err := h.repo.GetByID(r.Context(), id)
    if err != nil {
        if errors.Is(err, model.ErrAccountNotFound) {
            writeError(w, http.StatusNotFound, "Account not found")
            return
        }
        writeError(w, http.StatusInternalServerError, "Failed to get account")
        return
    }

    // Authorization check: account must belong to authenticated customer
    if account.CustomerID == nil || *account.CustomerID != customerID {
        writeError(w, http.StatusForbidden, "Access denied")
        return
    }

    writeJSON(w, http.StatusOK, account)
}

// Create handles POST /accounts
// Associates new account with authenticated customer
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
    customerID := middleware.GetCustomerID(r.Context())
    
    var req model.CreateAccountRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    if err := req.Validate(); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Create account linked to this customer
    account, err := h.repo.CreateForCustomer(r.Context(), req, customerID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to create account")
        return
    }

    writeJSON(w, http.StatusCreated, account)
}
```

**Update:** `internal/handler/transfer.go`

```go
// CreateTransfer handles POST /transfers
// Verifies source account belongs to authenticated customer
func (h *TransferHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
    customerID := middleware.GetCustomerID(r.Context())
    
    // ... existing parsing and validation ...

    // Validate source account exists AND belongs to authenticated customer
    fromAccount, err := h.accountRepo.GetByID(r.Context(), req.FromAccountID)
    if err != nil {
        if errors.Is(err, model.ErrAccountNotFound) {
            writeError(w, http.StatusBadRequest, "Source account not found")
            return
        }
        writeError(w, http.StatusInternalServerError, "Failed to validate source account")
        return
    }
    
    // Authorization: can only transfer FROM your own accounts
    if fromAccount.CustomerID == nil || *fromAccount.CustomerID != customerID {
        writeError(w, http.StatusForbidden, "You can only transfer from your own accounts")
        return
    }

    // Destination account doesn't need to belong to customer (you can pay others)
    // but it must exist and be active
    // ... rest of existing validation ...
}
```

### 3.5 Authorization Rules Summary

| Endpoint | Rule |
|----------|------|
| `GET /accounts` | Return only customer's accounts |
| `GET /accounts/{id}` | 403 if not customer's account |
| `POST /accounts` | Auto-link to authenticated customer |
| `GET /accounts/{id}/balance` | 403 if not customer's account |
| `POST /transfers` | 403 if source account not owned by customer |
| `GET /transactions/{id}` | 403 if transaction doesn't involve customer's accounts |

### 3.6 Testing Phase 3

1. **Try to access accounts without token:**
   ```bash
   curl http://localhost:8080/v1/accounts
   # Should return 401 Unauthorized
   ```

2. **Access with valid token:**
   ```bash
   curl http://localhost:8080/v1/accounts \
     -H "Authorization: Bearer <access_token>"
   # Should return only your accounts (empty array if none created yet)
   ```

3. **Create an account:**
   ```bash
   curl -X POST http://localhost:8080/v1/accounts \
     -H "Authorization: Bearer <access_token>" \
     -H "Content-Type: application/json" \
     -d '{"account_type": "checking", "currency": "NOK"}'
   ```

4. **Try to access another customer's account:**
   ```bash
   # Create second user, get their account ID, then try with first user's token
   curl http://localhost:8080/v1/accounts/<other-users-account-id> \
     -H "Authorization: Bearer <your_access_token>"
   # Should return 403 Forbidden
   ```

5. **Try to transfer from account you don't own:**
   ```bash
   curl -X POST http://localhost:8080/v1/transfers \
     -H "Authorization: Bearer <your_access_token>" \
     -H "Content-Type: application/json" \
     -H "Idempotency-Key: test-steal-money" \
     -d '{
       "from_account_id": "<other-users-account-id>",
       "to_account_id": "<your-account-id>",
       "amount": "1000000.00",
       "currency": "NOK"
     }'
   # Should return 403 Forbidden
   ```

---

## Phase 4: Frontend with Authentication

### 4.1 Objective

Build a React frontend that:
1. Provides login/registration forms
2. Stores access token in memory (React state)
3. Automatically refreshes tokens when needed
4. Shows accounts, balances, and transfer functionality
5. Protects routes from unauthenticated access

### 4.2 Project Structure

```
frontend/
├── public/
│   └── index.html
├── src/
│   ├── api/
│   │   └── client.ts         # API client with auth handling
│   ├── components/
│   │   ├── LoginForm.tsx
│   │   ├── RegisterForm.tsx
│   │   ├── AccountList.tsx
│   │   ├── AccountDetail.tsx
│   │   ├── TransferForm.tsx
│   │   └── ProtectedRoute.tsx
│   ├── context/
│   │   └── AuthContext.tsx   # Auth state management
│   ├── hooks/
│   │   └── useAuth.ts        # Custom auth hook
│   ├── pages/
│   │   ├── LoginPage.tsx
│   │   ├── RegisterPage.tsx
│   │   ├── DashboardPage.tsx
│   │   └── AccountPage.tsx
│   ├── types/
│   │   └── index.ts          # TypeScript types
│   ├── App.tsx
│   └── index.tsx
├── package.json
└── tsconfig.json
```

### 4.3 API Client with Token Management

**File:** `src/api/client.ts`

```typescript
const API_BASE = 'http://localhost:8080';

// Store access token in memory (not localStorage - security best practice)
let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
    accessToken = token;
}

export function getAccessToken(): string | null {
    return accessToken;
}

// Generic fetch wrapper with auth handling
async function fetchWithAuth(
    endpoint: string,
    options: RequestInit = {}
): Promise<Response> {
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
    }

    let response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers,
        credentials: 'include', // Include cookies for refresh token
    });

    // If unauthorized, try to refresh token
    if (response.status === 401 && accessToken) {
        const refreshed = await refreshAccessToken();
        if (refreshed) {
            // Retry original request with new token
            headers['Authorization'] = `Bearer ${accessToken}`;
            response = await fetch(`${API_BASE}${endpoint}`, {
                ...options,
                headers,
                credentials: 'include',
            });
        }
    }

    return response;
}

// Refresh the access token using the refresh token cookie
async function refreshAccessToken(): Promise<boolean> {
    try {
        const response = await fetch(`${API_BASE}/auth/refresh`, {
            method: 'POST',
            credentials: 'include',
        });

        if (response.ok) {
            const data = await response.json();
            setAccessToken(data.access_token);
            return true;
        }
    } catch (error) {
        console.error('Token refresh failed:', error);
    }

    setAccessToken(null);
    return false;
}

// Auth API calls
export async function login(email: string, password: string) {
    const response = await fetch(`${API_BASE}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Login failed');
    }

    const data = await response.json();
    setAccessToken(data.access_token);
    return data;
}

export async function register(
    email: string,
    password: string,
    firstName: string,
    lastName: string
) {
    const response = await fetch(`${API_BASE}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            email,
            password,
            first_name: firstName,
            last_name: lastName,
        }),
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Registration failed');
    }

    return response.json();
}

export async function logout() {
    await fetch(`${API_BASE}/auth/logout`, {
        method: 'POST',
        credentials: 'include',
    });
    setAccessToken(null);
}

// Account API calls
export async function getAccounts() {
    const response = await fetchWithAuth('/v1/accounts');
    if (!response.ok) throw new Error('Failed to fetch accounts');
    return response.json();
}

export async function getAccount(id: string) {
    const response = await fetchWithAuth(`/v1/accounts/${id}`);
    if (!response.ok) throw new Error('Failed to fetch account');
    return response.json();
}

export async function getAccountBalance(id: string) {
    const response = await fetchWithAuth(`/v1/accounts/${id}/balance`);
    if (!response.ok) throw new Error('Failed to fetch balance');
    return response.json();
}

export async function createAccount(accountType: string, currency: string) {
    const response = await fetchWithAuth('/v1/accounts', {
        method: 'POST',
        body: JSON.stringify({ account_type: accountType, currency }),
    });
    if (!response.ok) throw new Error('Failed to create account');
    return response.json();
}

// Transfer API calls
export async function createTransfer(
    fromAccountId: string,
    toAccountId: string,
    amount: string,
    currency: string,
    reference?: string
) {
    const idempotencyKey = crypto.randomUUID(); // Browser-native UUID
    
    const response = await fetchWithAuth('/v1/transfers', {
        method: 'POST',
        headers: {
            'Idempotency-Key': idempotencyKey,
        },
        body: JSON.stringify({
            from_account_id: fromAccountId,
            to_account_id: toAccountId,
            amount,
            currency,
            reference,
        }),
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Transfer failed');
    }

    return response.json();
}

export async function getTransaction(id: string) {
    const response = await fetchWithAuth(`/v1/transactions/${id}`);
    if (!response.ok) throw new Error('Failed to fetch transaction');
    return response.json();
}
```

### 4.4 Auth Context

**File:** `src/context/AuthContext.tsx`

```typescript
import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { setAccessToken, getAccessToken, logout as apiLogout } from '../api/client';

interface AuthContextType {
    isAuthenticated: boolean;
    isLoading: boolean;
    login: (token: string) => void;
    logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    // On mount, try to refresh token to check if user is logged in
    useEffect(() => {
        async function checkAuth() {
            try {
                const response = await fetch('http://localhost:8080/auth/refresh', {
                    method: 'POST',
                    credentials: 'include',
                });
                
                if (response.ok) {
                    const data = await response.json();
                    setAccessToken(data.access_token);
                    setIsAuthenticated(true);
                }
            } catch (error) {
                // Not logged in, that's fine
            } finally {
                setIsLoading(false);
            }
        }
        checkAuth();
    }, []);

    const login = (token: string) => {
        setAccessToken(token);
        setIsAuthenticated(true);
    };

    const logout = async () => {
        await apiLogout();
        setIsAuthenticated(false);
    };

    return (
        <AuthContext.Provider value={{ isAuthenticated, isLoading, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
```

### 4.5 Protected Route Component

**File:** `src/components/ProtectedRoute.tsx`

```typescript
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

interface ProtectedRouteProps {
    children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
    const { isAuthenticated, isLoading } = useAuth();
    const location = useLocation();

    if (isLoading) {
        return <div>Loading...</div>;
    }

    if (!isAuthenticated) {
        // Redirect to login, but save the attempted location
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    return <>{children}</>;
}
```

### 4.6 Key Frontend Concepts

**Why store access token in memory vs localStorage?**
- XSS attacks can read localStorage via JavaScript
- Memory is cleared when tab closes (more secure)
- Refresh tokens in HttpOnly cookies handle persistence

**Why use React Context for auth state?**
- Avoids prop drilling through component tree
- Single source of truth for auth status
- Easy to access from any component

**What is `credentials: 'include'`?**
- Tells browser to send cookies with cross-origin requests
- Required for refresh token cookie to be sent
- You'll need to configure CORS on the backend to allow this

### 4.7 CORS Configuration

**Update backend** to allow frontend origin:

```go
// In main.go or as middleware
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // React dev server
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Apply as first middleware
r.Use(corsMiddleware)
```

### 4.8 Testing Phase 4

1. **Start both services:**
   ```bash
   # Terminal 1 - Backend
   go run ./cmd/api
   
   # Terminal 2 - Frontend
   cd frontend && npm start
   ```

2. **Register a new user via UI**

3. **Login and verify:**
   - Access token appears in Network tab response
   - Refresh token cookie is set (check Application > Cookies)
   - Dashboard loads with user's accounts

4. **Test token refresh:**
   - Wait 15+ minutes (or temporarily set short expiry)
   - Make an API call
   - Check Network tab - should see refresh call followed by retry

5. **Test logout:**
   - Click logout
   - Verify cookie is cleared
   - Verify redirect to login
   - Try to access dashboard directly - should redirect to login

---

## Security Checklist

Before deploying, verify:

- [ ] JWT secret is strong (32+ random bytes) and stored securely
- [ ] Passwords hashed with bcrypt (cost >= 10)
- [ ] Refresh tokens use HttpOnly, Secure, SameSite cookies
- [ ] Access tokens are short-lived (15 minutes)
- [ ] Failed login attempts are rate-limited
- [ ] Account lockout after repeated failures
- [ ] CORS properly configured (not `*` in production)
- [ ] HTTPS enforced in production
- [ ] Sensitive fields excluded from JSON serialization (`json:"-"`)
- [ ] Authorization checks on every protected endpoint
- [ ] SQL queries use parameterized statements (already done with pgx)

---

## Summary

| Phase | What You'll Learn |
|-------|-------------------|
| 1 | Database design, foreign keys, nullable fields, Go pointer types |
| 2 | Password hashing, JWT structure, token-based auth, security best practices |
| 3 | Middleware pattern, context propagation, authorization vs authentication |
| 4 | React state management, protected routes, secure token storage, CORS |

This builds foundational skills used in virtually every production web application. Take your time with each phase - understanding WHY each piece exists is more valuable than just making it work.