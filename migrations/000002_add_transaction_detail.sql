-- +goose Up
-- Add transfer details to transactions table for idempotency validation
ALTER TABLE transactions ADD COLUMN amount DECIMAL(19,4);
ALTER TABLE transactions ADD COLUMN currency CHAR(3);
ALTER TABLE transactions ADD COLUMN from_account_id UUID REFERENCES accounts(id);
ALTER TABLE transactions ADD COLUMN to_account_id UUID REFERENCES accounts(id);

-- Create indexes for common lookups
CREATE INDEX IF NOT EXISTS idx_transactions_from_account_id ON transactions (from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account_id ON transactions (to_account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_to_account_id;
DROP INDEX IF EXISTS idx_transactions_from_account_id;

ALTER TABLE transactions DROP COLUMN IF EXISTS to_account_id;
ALTER TABLE transactions DROP COLUMN IF EXISTS from_account_id;
ALTER TABLE transactions DROP COLUMN IF EXISTS currency;
ALTER TABLE transactions DROP COLUMN IF EXISTS amount;