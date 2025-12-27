# Domain Models

## Purpose

Core data structures and validation logic for the banking domain. Models define the shape of data and business rules independent of storage or transport.

## Architecture

```
model/
  ├── account.go      → Account, AccountBalance, CreateAccountRequest
  ├── customer.go     → Customer, CreateCustomerRequest, LoginRequest
  ├── transaction.go  → Transaction, LedgerEntry, TransactionParty
  └── errors.go       → Domain-specific error definitions
```

## Models

### Customer
User identity and authentication state.

| Field | Type | Description |
|-------|------|-------------|
| ID | UUID | Primary key |
| Email | string | Login identifier |
| PasswordHash | string | bcrypt hash (never serialized) |
| Status | CustomerStatus | active, suspended, closed |
| FailedLoginAttempts | int | Brute force tracking |
| LockedUntil | *time.Time | Temporary lockout |

### Account
Bank account with customer ownership.

| Field | Type | Description |
|-------|------|-------------|
| ID | UUID | Primary key |
| AccountNumber | string | Human-readable identifier |
| AccountType | AccountType | checking, savings, loan, equity |
| Currency | string | 3-letter ISO code (NOK, USD) |
| Status | AccountStatus | active, frozen, closed |
| CustomerID | *UUID | Owner (nil for system accounts) |

### Transaction
Money movement record with state machine.

| Field | Type | Description |
|-------|------|-------------|
| ID | UUID | Primary key |
| IdempotencyKey | string | Duplicate prevention |
| Type | TransactionType | transfer, deposit, withdrawal |
| Status | TransactionStatus | pending → processing → completed/failed |
| FromAccountID | *UUID | Source account |
| ToAccountID | *UUID | Destination account |
| Amount | string | Decimal as string |
| Currency | string | 3-letter ISO code |

### LedgerEntry
Double-entry bookkeeping record.

| Field | Type | Description |
|-------|------|-------------|
| AccountID | UUID | Account affected |
| TransactionID | UUID | Parent transaction |
| Amount | string | Positive = credit, negative = debit |
| EntryType | LedgerEntryType | debit, credit |

## Transaction State Machine

```
PENDING ──► PROCESSING ──► COMPLETED
                  │
                  └──► FAILED
```

- **Pending:** Created, waiting for processing
- **Processing:** Actively being executed
- **Completed:** Successfully finished
- **Failed:** Error occurred, includes error_message

## Validation

Request structs have `Validate()` methods:
- `CreateAccountRequest.Validate()` - Rejects equity type, validates currency
- `CreateTransferRequest.Validate()` - Checks UUIDs, prevents same-account transfer
- `CreateCustomerRequest.Validate()` - Email format, password strength
- `LoginRequest.Validate()` - Required fields

## Design Decisions

**Why string for amounts:** Floats cause precision errors with decimals. Strings preserve exact values; conversion to `decimal.Decimal` happens at calculation time.

**Why pointer fields for optional data:** Go's zero values (empty string, 0) can be valid data. Pointers distinguish "not set" (nil) from "set to zero value."

**Why `json:"-"` on sensitive fields:** Prevents accidental serialization of password hashes, failed attempt counts, and other internal state.

**Why typed enums (AccountType, TransactionStatus):** Type safety at compile time. Prevents passing wrong status strings. Self-documenting code.

**Why CustomerID is nullable on Account:** System accounts (like Bank Equity) aren't owned by customers. Nil CustomerID indicates a system account.
