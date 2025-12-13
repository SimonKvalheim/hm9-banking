-- +goose Up
-- Create extension for UUID functions (pgcrypto)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- accounts table
CREATE TABLE IF NOT EXISTS accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_number VARCHAR(34) UNIQUE NOT NULL,
  account_type VARCHAR(20) NOT NULL,
  currency CHAR(3) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_account_number ON accounts (account_number);

-- transactions table
CREATE TABLE IF NOT EXISTS transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  idempotency_key VARCHAR(64) UNIQUE NOT NULL,
  "type" VARCHAR(20) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  reference VARCHAR(255),
  initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  processed_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  error_message VARCHAR(500),
  metadata JSONB DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_transactions_idempotency_key ON transactions (idempotency_key);

-- ledger_entries table
CREATE TABLE IF NOT EXISTS ledger_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
  account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  amount DECIMAL(19,4) NOT NULL,
  entry_type VARCHAR(10) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (amount <> 0)
);

CREATE INDEX IF NOT EXISTS idx_ledger_entries_account_id ON ledger_entries (account_id);
CREATE INDEX IF NOT EXISTS idx_ledger_entries_transaction_id ON ledger_entries (transaction_id);

-- transaction_parties linking table
CREATE TABLE IF NOT EXISTS transaction_parties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
  account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  role VARCHAR(20) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transaction_parties_transaction_id ON transaction_parties (transaction_id);

-- +goose Down
DROP INDEX IF EXISTS idx_transaction_parties_transaction_id;
DROP TABLE IF EXISTS transaction_parties;

DROP INDEX IF EXISTS idx_ledger_entries_account_id;
DROP INDEX IF EXISTS idx_ledger_entries_transaction_id;
DROP TABLE IF EXISTS ledger_entries;

DROP INDEX IF EXISTS idx_transactions_idempotency_key;
DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS idx_accounts_account_number;
DROP TABLE IF EXISTS accounts;

-- We don't drop the extension to avoid impacting other db objects in environments where pgcrypto may be shared
