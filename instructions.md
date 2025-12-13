# Fjord Bank – Transaction System Design

> **Status:** Draft  
> **Last Updated:** December 2024  
> **Purpose:** Technical reference for the core transaction processing system

---

## 1. Overview

Fjord Bank is a learning project simulating a banking system with production-grade patterns. This document describes the design of the core transaction system, which handles money movement between accounts.

### Design Principles

- **Double-entry bookkeeping:** Every money movement creates balanced ledger entries that sum to zero.
- **Immutable audit trail:** Ledger entries are append-only; corrections are made by adding reversing entries.
- **Asynchronous processing:** Transactions are queued and processed by workers, not inline with API requests.
- **Idempotency:** Every operation can be safely retried without risk of duplication.
- **Explicit state management:** Transactions move through well-defined states with clear transition rules.

---

## 2. Core Concepts
- Allows reconstruction of balance at any point in time
- Prevents hidden balance mutations

**Example:** A €100 transfer from Account A to Account B creates two ledger entries:

| Account | Amount  | Effect |
|---------|---------|--------|
| A       | -100.00 | Debit  |
| B       | +100.00 | Credit |
| **Sum** | **0.00**|        |

### 2.2 Transaction States

Transactions progress through a state machine:

```
PENDING ──────► PROCESSING ──────► COMPLETED
                    │
                    └──────────► FAILED
```

**State definitions:**

| State | Description | Ledger Entries Exist? |
|-------|-------------|----------------------|
| `pending` | Transaction created and validated, awaiting processing | No |
| `processing` | Worker has claimed the transaction, currently executing | No (in progress) |
| `completed` | Successfully processed, all entries written | Yes |
| `failed` | Processing failed, no money moved | No |

**Transition rules:**
- Only `pending` → `processing` and `processing` → `completed`/`failed` are valid
- Once `completed` or `failed`, a transaction cannot change state
- Reversals are handled by creating a new, linked transaction (not by modifying state)

### 2.3 Idempotency

Every transaction request includes a client-generated `idempotency_key`. The system guarantees:

1. If a transaction with the given key exists, return the existing transaction
2. If not, create a new transaction
3. This check and create is atomic

This allows clients to safely retry requests on timeout or network failure.

---

## 3. Data Model

### 3.1 Entity Relationship

```
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│   accounts    │       │ transactions  │       │ ledger_entries│
├───────────────┤       ├───────────────┤       ├───────────────┤
│ id (PK)       │◄──────│ id (PK)       │◄──────│ id (PK)       │
│ account_number│       │ idempotency_  │       │ transaction_id│
│ account_type  │       │   key (UNIQUE)│       │ account_id    │
│ currency      │       │ type          │       │ amount        │
│ status        │       │ status        │       │ entry_type    │
│ created_at    │       │ reference     │       │ created_at    │
│ updated_at    │       │ initiated_at  │       └───────────────┘
└───────────────┘       │ completed_at  │
                        │ error_message │
                        │ metadata      │
                        └───────────────┘
```

### 3.2 Table Definitions

#### accounts
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Internal identifier |
| account_number | VARCHAR(34) | UNIQUE, NOT NULL | Human-readable account number |
| account_type | VARCHAR(20) | NOT NULL | checking, savings, loan |
| currency | CHAR(3) | NOT NULL | ISO 4217 code (NOK, EUR, USD) |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'active' | active, frozen, closed |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |

#### transactions
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Internal identifier |
| idempotency_key | VARCHAR(64) | UNIQUE, NOT NULL | Client-provided deduplication key |
| type | VARCHAR(20) | NOT NULL | transfer, deposit, withdrawal |
| status | VARCHAR(20) | NOT NULL, DEFAULT 'pending' | pending, processing, completed, failed |
| reference | VARCHAR(255) | | Human-readable description |
| initiated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | When the request was received |
| processed_at | TIMESTAMPTZ | | When processing started |
| completed_at | TIMESTAMPTZ | | When terminal state was reached |
| error_message | VARCHAR(500) | | Populated on failure |
| metadata | JSONB | DEFAULT '{}' | Flexible additional data |

#### ledger_entries
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Internal identifier |
| transaction_id | UUID | FK → transactions, NOT NULL | Parent transaction |
| account_id | UUID | FK → accounts, NOT NULL | Affected account |
| amount | DECIMAL(19,4) | NOT NULL | Positive = credit, negative = debit |
| entry_type | VARCHAR(10) | NOT NULL | debit, credit (for readability) |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |

#### transaction_parties (linking table for transfer participants)
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | |
| transaction_id | UUID | FK → transactions, NOT NULL | |
| account_id | UUID | FK → accounts, NOT NULL | |
| role | VARCHAR(20) | NOT NULL | source, destination |

### 3.3 Derived Calculations

**Account balance:**
```sql
SELECT COALESCE(SUM(amount), 0) AS balance
FROM ledger_entries
WHERE account_id = :account_id;
```

**Account balance at a point in time:**
```sql
SELECT COALESCE(SUM(amount), 0) AS balance
FROM ledger_entries
WHERE account_id = :account_id
  AND created_at <= :as_of_timestamp;
```

**Verify transaction is balanced:**
```sql
SELECT SUM(amount) = 0 AS is_balanced
FROM ledger_entries
WHERE transaction_id = :transaction_id;
```

---

## 4. System Architecture

### 4.1 Component Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Clients                                    │
│                    (Web App, Mobile, API)                           │
└─────────────────────────────────┬───────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         API Gateway                                  │
│                     (Authentication, Routing)                        │
└─────────────────────────────────┬───────────────────────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    ▼                           ▼
          ┌─────────────────┐         ┌─────────────────┐
          │ Account Service │         │ Transaction     │
          │   (Go )         │         │      (Go)       │
          └────────┬────────┘         └────────┬────────┘
                   │                            │
                   │                            ▼
                   │                   ┌─────────────────┐
                   │                   │  Message Queue  │
                   │                   │  (Redis)       
                   │                   │                 │
                   │                   └────────┬────────┘
                   │                            │
                   │                            ▼
                   │                   ┌─────────────────┐
                   │                   │ Transaction     │
                   │                   │ Worker          │
                   │                   │ (Go     )       │
                   │                   └────────┬────────┘
                   │                            │
                   └─────────────┬──────────────┘
                                 ▼
                        ┌─────────────────┐
                        │   PostgreSQL    │
                        │   (Primary DB)  │
                        └─────────────────┘
```

### 4.2 Service Responsibilities

**Account Service (Python/FastAPI):**
- Account CRUD operations
- Balance inquiries (read from ledger_entries)
- Account status management

**Transaction Service (Go or Rust):**
- Receives transfer requests
- Validates requests (accounts exist, sufficient balance, etc.)
- Creates transaction record with `pending` status
- Publishes message to queue
- Returns transaction ID to client

**Transaction Worker (Go or Rust):**
- Consumes messages from queue
- Claims transaction (sets `processing` status)
- Executes the transfer within a database transaction
- Creates ledger entries
- Updates transaction to `completed` or `failed`

### 4.3 Message Queue Design

**Queue:** `transactions.pending`

**Message schema:**
```json
{
  "transaction_id": "uuid",
  "type": "transfer",
  "published_at": "iso8601 timestamp"
}
```

The message is intentionally minimal—the worker fetches full details from the database. This ensures the database is the source of truth and avoids message/database state divergence.

---

## 5. Transaction Processing Flow

### 5.1 Internal Transfer (Happy Path)

```
Client                API Service              Queue              Worker              Database
  │                        │                     │                   │                    │
  │ POST /transfers        │                     │                   │                    │
  │ {idempotency_key,      │                     │                   │                    │
  │  from, to, amount}     │                     │                   │                    │
  │───────────────────────►│                     │                   │                    │
  │                        │                     │                   │                    │
  │                        │ Check idempotency_key exists?           │                    │
  │                        │─────────────────────────────────────────────────────────────►│
  │                        │◄────────────────────────────────────────────────────── No ──│
  │                        │                     │                   │                    │
  │                        │ Validate accounts, check balance        │                    │
  │                        │─────────────────────────────────────────────────────────────►│
  │                        │◄─────────────────────────────────────────────────────── OK ──│
  │                        │                     │                   │                    │
  │                        │ INSERT transaction (status=pending)     │                    │
  │                        │─────────────────────────────────────────────────────────────►│
  │                        │                     │                   │                    │
  │                        │ Publish message     │                   │                    │
  │                        │────────────────────►│                   │                    │
  │                        │                     │                   │                    │
  │ 202 Accepted           │                     │                   │                    │
  │ {transaction_id,       │                     │                   │                    │
  │  status: pending}      │                     │                   │                    │
  │◄───────────────────────│                     │                   │                    │
  │                        │                     │                   │                    │
  │                        │                     │ Consume message   │                    │
  │                        │                     │──────────────────►│                    │
  │                        │                     │                   │                    │
  │                        │                     │                   │ UPDATE status =    │
  │                        │                     │                   │ processing         │
  │                        │                     │                   │───────────────────►│
  │                        │                     │                   │                    │
  │                        │                     │                   │ BEGIN TRANSACTION  │
  │                        │                     │                   │───────────────────►│
  │                        │                     │                   │                    │
  │                        │                     │                   │ Re-validate balance│
  │                        │                     │                   │───────────────────►│
  │                        │                     │                   │                    │
  │                        │                     │                   │ INSERT ledger      │
  │                        │                     │                   │ entries (2 rows)   │
  │                        │                     │                   │───────────────────►│
  │                        │                     │                   │                    │
  │                        │                     │                   │ UPDATE status =    │
  │                        │                     │                   │ completed          │
  │                        │                     │                   │───────────────────►│
  │                        │                     │                   │                    │
  │                        │                     │                   │ COMMIT             │
  │                        │                     │                   │───────────────────►│
  │                        │                     │                   │                    │
  │                        │                     │◄── ACK ───────────│                    │
```

### 5.2 Failure Scenarios

**Insufficient funds (detected at processing time):**
- Worker sets status to `failed`
- Error message populated: "Insufficient funds"
- No ledger entries created
- Message acknowledged (not requeued)

**Worker crashes mid-processing:**
- Database transaction rolls back (no partial writes)
- Message returns to queue (unacknowledged)
- Another worker picks up the message
- Transaction still in `pending` or `processing` state

**Duplicate message delivery:**
- Worker checks current status before processing
- If already `completed` or `failed`, acknowledges without reprocessing
- If `processing` by another worker, skips (or waits with timeout)

---

## 6. API Design

### 6.1 Create Transfer

```
POST /v1/transfers
Content-Type: application/json
Idempotency-Key: <client-generated-uuid>

{
  "from_account_id": "uuid",
  "to_account_id": "uuid",
  "amount": "100.00",
  "currency": "NOK",
  "reference": "Rent payment"
}
```

**Response (202 Accepted):**
```json
{
  "transaction_id": "uuid",
  "status": "pending",
  "created_at": "2024-12-13T10:00:00Z"
}
```

### 6.2 Get Transaction Status

```
GET /v1/transactions/{transaction_id}
```

**Response:**
```json
{
  "transaction_id": "uuid",
  "type": "transfer",
  "status": "completed",
  "from_account_id": "uuid",
  "to_account_id": "uuid",
  "amount": "100.00",
  "currency": "NOK",
  "reference": "Rent payment",
  "initiated_at": "2024-12-13T10:00:00Z",
  "completed_at": "2024-12-13T10:00:01Z"
}
```

### 6.3 Get Account Balance

```
GET /v1/accounts/{account_id}/balance
```

**Response:**
```json
{
  "account_id": "uuid",
  "balance": "4500.00",
  "currency": "NOK",
  "as_of": "2024-12-13T10:00:05Z"
}
```

---

## 7. Future Considerations

### 7.1 External Transfers (Phase 2)
- Introduce a "Mock External Bank" service
- Transactions to external accounts enter a `pending_external` state
- Settlement simulation via message passing between services
- Failure and reconciliation handling

### 7.2 AML Screening (Phase 3)
- Hook into transaction flow before `processing`
- Transactions may enter `pending_review` state
- Alert generation and case management

### 7.3 Performance Optimizations
- Materialized balance views (with careful invalidation)
- Read replicas for balance queries
- Partitioning ledger_entries by date

---

## 8. Glossary

| Term | Definition |
|------|------------|
| **Debit** | An entry that decreases an account balance (negative amount) |
| **Credit** | An entry that increases an account balance (positive amount) |
| **Idempotency** | The property that an operation produces the same result whether executed once or multiple times |
| **Ledger** | The complete record of all financial entries |
| **Settlement** | The actual transfer of funds between institutions |
| **ACID** | Atomicity, Consistency, Isolation, Durability—database transaction guarantees |