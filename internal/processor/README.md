# Transfer Processor

## Purpose

Core transaction processing logic implementing double-entry bookkeeping. Handles the atomic execution of transfers: balance checks, ledger entry creation, and state transitions.

## Architecture

```
processor/
  └── transfer.go  → TransferProcessor.Process()

Process flow:
  1. Claim transaction (pending → processing)
  2. Get transaction parties (source + destination)
  3. Validate amount
  4. Check sufficient funds
  5. Create ledger entries
  6. Complete transaction (processing → completed)
```

All steps execute within a single database transaction for atomicity.

## Processing Steps

### 1. Claim Transaction
```sql
UPDATE transactions
SET status = 'processing', processed_at = NOW()
WHERE id = $1 AND status = 'pending'
```
Only one processor can claim a transaction. If already claimed, returns nil (idempotent).

### 2. Get Parties
Fetch source and destination account IDs from `transaction_parties` table.

### 3. Balance Check
```sql
SELECT COALESCE(SUM(amount), 0) FROM ledger_entries WHERE account_id = $1
```
Compares current balance against transfer amount. If insufficient, marks transaction as failed.

### 4. Create Ledger Entries
Two entries that sum to zero:
- Source account: `-amount` (debit)
- Destination account: `+amount` (credit)

### 5. Complete Transaction
```sql
UPDATE transactions SET status = 'completed', completed_at = NOW()
```

## Failure Handling

If any step fails:
- Database transaction rolls back
- Transaction remains in previous state (pending or processing)
- Can be retried later

If business rule fails (insufficient funds):
- Transaction marked as `failed` with error message
- Database transaction commits (failure is recorded)
- Not retriable

## Design Decisions

**Why single database transaction:** All-or-nothing execution. Either the transfer completes fully (state updated + ledger entries created) or nothing happens. Prevents partial state.

**Why claim pattern:** Prevents double-processing. Multiple workers or retry attempts will see transaction already claimed and skip.

**Why WHERE status = 'pending':** State machine enforcement at database level. Can only transition from pending → processing, not from completed → processing.

**Why balance check without FOR UPDATE:** Reading ledger entries to sum balance. Since we're in a transaction with SERIALIZABLE isolation level, phantom reads are prevented.

**Why commit on failure:** Recording that a transaction failed is important for debugging and user feedback. Failure state is committed; business operation is not.
