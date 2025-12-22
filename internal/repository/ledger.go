package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

// LedgerRepository handles database operations for ledger entries
type LedgerRepository struct {
	db *pgxpool.Pool
}

// NewLedgerRepository creates a new LedgerRepository
func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{db: db}
}

// CreateEntries inserts multiple ledger entries within a database transaction
// This is used by the worker to create balanced entries atomically
func (r *LedgerRepository) CreateEntries(ctx context.Context, dbTx pgx.Tx, entries []model.LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries (id, transaction_id, account_id, amount, entry_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, entry := range entries {
		_, err := dbTx.Exec(ctx, query,
			entry.ID,
			entry.TransactionID,
			entry.AccountID,
			entry.Amount,
			entry.EntryType,
			entry.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}
	}

	return nil
}

// CreateEntriesStandalone inserts ledger entries using the pool connection
// Use this when you don't have an existing transaction
func (r *LedgerRepository) CreateEntriesStandalone(ctx context.Context, entries []model.LedgerEntry) error {
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)

	if err := r.CreateEntries(ctx, dbTx, entries); err != nil {
		return err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit ledger entries: %w", err)
	}

	return nil
}

// GetByTransactionID retrieves all ledger entries for a transaction
func (r *LedgerRepository) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]model.LedgerEntry, error) {
	query := `
		SELECT id, transaction_id, account_id, amount, entry_type, created_at
		FROM ledger_entries
		WHERE transaction_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entries: %w", err)
	}
	defer rows.Close()

	var entries []model.LedgerEntry
	for rows.Next() {
		var entry model.LedgerEntry
		err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.Amount,
			&entry.EntryType,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetByAccountID retrieves all ledger entries for an account
func (r *LedgerRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit int) ([]model.LedgerEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := `
		SELECT id, transaction_id, account_id, amount, entry_type, created_at
		FROM ledger_entries
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entries for account: %w", err)
	}
	defer rows.Close()

	var entries []model.LedgerEntry
	for rows.Next() {
		var entry model.LedgerEntry
		err := rows.Scan(
			&entry.ID,
			&entry.TransactionID,
			&entry.AccountID,
			&entry.Amount,
			&entry.EntryType,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetBalanceAtTime calculates the balance for an account at a specific point in time
func (r *LedgerRepository) GetBalanceAtTime(ctx context.Context, accountID uuid.UUID, asOf time.Time) (string, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) AS balance
		FROM ledger_entries
		WHERE account_id = $1
		  AND created_at <= $2
	`

	var balance string
	err := r.db.QueryRow(ctx, query, accountID, asOf).Scan(&balance)
	if err != nil {
		return "", fmt.Errorf("failed to get balance at time: %w", err)
	}

	return balance, nil
}

// VerifyTransactionBalance checks that a transaction's ledger entries sum to zero
func (r *LedgerRepository) VerifyTransactionBalance(ctx context.Context, transactionID uuid.UUID) (bool, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) = 0 AS is_balanced
		FROM ledger_entries
		WHERE transaction_id = $1
	`

	var isBalanced bool
	err := r.db.QueryRow(ctx, query, transactionID).Scan(&isBalanced)
	if err != nil {
		return false, fmt.Errorf("failed to verify transaction balance: %w", err)
	}

	return isBalanced, nil
}

// BuildTransferEntries creates the ledger entries for a transfer
// Returns two entries that sum to zero (double-entry bookkeeping)
func BuildTransferEntries(transactionID, fromAccountID, toAccountID uuid.UUID, amount string) []model.LedgerEntry {
	now := time.Now()

	return []model.LedgerEntry{
		{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     fromAccountID,
			Amount:        "-" + amount, // Debit (negative)
			EntryType:     model.LedgerEntryTypeDebit,
			CreatedAt:     now,
		},
		{
			ID:            uuid.New(),
			TransactionID: transactionID,
			AccountID:     toAccountID,
			Amount:        amount, // Credit (positive)
			EntryType:     model.LedgerEntryTypeCredit,
			CreatedAt:     now,
		},
	}
}
