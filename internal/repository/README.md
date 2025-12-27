# Database Repositories

## Purpose

Data access layer abstracting PostgreSQL operations. Repositories handle SQL queries, row scanning, and transaction management—keeping database details out of handlers and services.

## Architecture

```
repository/
  ├── account.go      → Account CRUD, balance calculation
  ├── customer.go     → Customer CRUD, login tracking
  ├── transaction.go  → Transaction lifecycle, idempotency
  └── ledger.go       → Double-entry ledger operations
```

All repositories receive a `*pgxpool.Pool` connection pool and use parameterized queries.

## Repositories

### AccountRepository
| Method | Description |
|--------|-------------|
| `Create` | Insert new account |
| `CreateForCustomer` | Insert account linked to customer |
| `GetByID` | Fetch single account |
| `GetByCustomerID` | Fetch all accounts for customer |
| `GetBalanceAtTime` | Calculate balance from ledger entries |

### CustomerRepository
| Method | Description |
|--------|-------------|
| `Create` | Insert new customer |
| `GetByID` | Fetch by UUID |
| `GetByEmail` | Fetch for login |
| `UpdateLastLogin` | Track login timestamp |
| `IncrementFailedAttempts` | Brute force tracking |
| `ResetFailedAttempts` | Clear on successful login |
| `LockAccount` | Set locked_until timestamp |

### TransactionRepository
| Method | Description |
|--------|-------------|
| `Create` | Insert transaction + parties atomically |
| `GetByID` | Fetch transaction |
| `GetByIdempotencyKey` | Check for duplicate |
| `UpdateStatus` | Transition state machine |

### LedgerRepository
| Method | Description |
|--------|-------------|
| `CreateEntries` | Insert ledger entries (within transaction) |
| `GetByTransactionID` | Fetch entries for a transaction |
| `GetBalanceAtTime` | Sum entries up to timestamp |
| `VerifyTransactionBalance` | Check entries sum to zero |

## Double-Entry Bookkeeping

Every transfer creates two ledger entries that sum to zero:

```
Transfer: A → B, 100 NOK

Ledger entries:
  Account A: -100 (debit)
  Account B: +100 (credit)
  ─────────────────
  Sum:         0
```

Balance is never stored—always calculated:
```sql
SELECT COALESCE(SUM(amount), 0) FROM ledger_entries WHERE account_id = $1
```

## Balance Calculation

Point-in-time balances are supported:
```sql
SELECT COALESCE(SUM(amount), 0)
FROM ledger_entries
WHERE account_id = $1 AND created_at <= $2
```

This enables historical balance queries and audit trails.

## Design Decisions

**Why calculated balances:** Storing balance creates sync issues—balance could diverge from ledger. Calculating from entries is authoritative and auditable. Trade-off: slightly slower reads, guaranteed consistency.

**Why pgx over database/sql:** pgx is PostgreSQL-native with better performance, COPY support, and cleaner API. No need for generic database abstraction in this project.

**Why repository pattern:** Testability (can mock repositories), separation of concerns (SQL stays in one place), single responsibility.

**Why pass pgx.Tx to CreateEntries:** Allows caller to control transaction scope. Transfer processing needs to update transaction status AND create ledger entries atomically—both must be in same database transaction.

**Why idempotency key unique constraint:** Database enforces idempotency. Race conditions between duplicate requests are handled by unique constraint violation, not application logic.
