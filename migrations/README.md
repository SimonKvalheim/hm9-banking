# Database Migrations

## Purpose

Version-controlled database schema evolution using goose. Each migration file modifies the schema in a numbered sequence.

## Running Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://fjord:fjordpass@localhost:5432/fjorddb?sslmode=disable"

# Apply all migrations
goose -dir migrations postgres "$DATABASE_URL" up

# Roll back one migration
goose -dir migrations postgres "$DATABASE_URL" down

# Check current version
goose -dir migrations postgres "$DATABASE_URL" status
```

## Schema Overview

```
customers ──────────┐
                    │ 1:N
                    ▼
accounts ◄──────── ledger_entries
    │                    │
    │ N:M                │ N:1
    ▼                    ▼
transaction_parties ─► transactions
```

## Tables

| Table | Purpose |
|-------|---------|
| `customers` | User accounts with auth data |
| `accounts` | Bank accounts (checking, savings, etc.) |
| `transactions` | Money movement records |
| `ledger_entries` | Double-entry bookkeeping |
| `transaction_parties` | Links transactions to accounts |

## Key Columns

### accounts
- `id` UUID - Primary key
- `account_number` - Human-readable identifier
- `customer_id` - Owner (nullable for system accounts)
- `currency` - 3-letter ISO code
- `status` - active, frozen, closed

### transactions
- `idempotency_key` - Unique, prevents duplicates
- `status` - pending, processing, completed, failed
- `from_account_id`, `to_account_id` - Transfer endpoints
- `amount`, `currency` - Transfer details

### ledger_entries
- `amount` - DECIMAL(19,4), positive or negative
- `entry_type` - debit or credit
- CHECK constraint: amount != 0

### customers
- `email` - Unique login identifier
- `password_hash` - bcrypt hash
- `failed_login_attempts` - Brute force tracking
- `locked_until` - Temporary lockout timestamp

## Migration Files

| File | Description |
|------|-------------|
| `000001_create_schema.sql` | Core tables: accounts, transactions, ledger_entries, transaction_parties |
| `000002_add_transaction_detail.sql` | Add amount/currency/account IDs directly to transactions |
| `000003_create_customers.sql` | Customer table + accounts.customer_id foreign key |

## Design Decisions

**Why UUID primary keys:** Unpredictable (security), distributed-safe (no sequence coordination), can be generated client-side.

**Why DECIMAL(19,4) for amounts:** 19 digits handle large values, 4 decimal places handle sub-cent precision. Avoids float rounding errors.

**Why idempotency_key unique constraint:** Database enforces exactly-once processing. Duplicate requests fail on constraint, not application logic.

**Why customer_id nullable:** System accounts (bank equity) have no owner. Null indicates internal account.

**Why ON DELETE RESTRICT for accounts:** Prevent deleting accounts that have ledger entries. Preserves audit trail.

**Why ON DELETE CASCADE for transactions:** Deleting a transaction removes its parties and entries. Used for cleanup, not normal operation.
