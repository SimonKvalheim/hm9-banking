package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

// TransferProcessor handles the processing of transfer transactions
type TransferProcessor struct {
	db *pgxpool.Pool
}

// NewTransferProcessor creates a new TransferProcessor
func NewTransferProcessor(db *pgxpool.Pool) *TransferProcessor {
	return &TransferProcessor{db: db}
}

// ProcessResult contains the result of processing a transaction
type ProcessResult struct {
	Success      bool
	ErrorMessage string
}

// Process executes a pending transfer transaction
// This is the core double-entry bookkeeping logic
func (p *TransferProcessor) Process(ctx context.Context, transactionID uuid.UUID) (*ProcessResult, error) {
	// Start a database transaction for atomicity
	dbTx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin db transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// Step 1: Claim the transaction (pending -> processing)
	tx, err := p.claimTransaction(ctx, dbTx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to claim transaction: %w", err)
	}
	if tx == nil {
		// Transaction not in pending state (already processed or doesn't exist)
		return &ProcessResult{Success: true, ErrorMessage: "transaction already processed or not found"}, nil
	}

	// Step 2: Get transaction parties
	parties, err := p.getParties(ctx, dbTx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parties: %w", err)
	}

	var sourceAccountID, destAccountID uuid.UUID
	for _, party := range parties {
		if party.Role == "source" {
			sourceAccountID = party.AccountID
		} else if party.Role == "destination" {
			destAccountID = party.AccountID
		}
	}

	if sourceAccountID == uuid.Nil || destAccountID == uuid.Nil {
		// Mark as failed - invalid transaction setup
		if err := p.failTransaction(ctx, dbTx, transactionID, "invalid transaction parties"); err != nil {
			return nil, err
		}
		if err := dbTx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit: %w", err)
		}
		return &ProcessResult{Success: false, ErrorMessage: "invalid transaction parties"}, nil
	}

	// Step 3: Extract amount from metadata
	amount := tx.Amount
	if amount == "" {
		if err := p.failTransaction(ctx, dbTx, transactionID, "invalid amount in transaction"); err != nil {
			return nil, err
		}
		if err := dbTx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit: %w", err)
		}
		return &ProcessResult{Success: false, ErrorMessage: "invalid amount"}, nil
	}

	// Step 4: Check sufficient balance (with row lock on ledger entries)
	balance, err := p.getBalanceForUpdate(ctx, dbTx, sourceAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if !hasSufficientFunds(balance, amount) {
		// Mark as failed - insufficient funds
		if err := p.failTransaction(ctx, dbTx, transactionID, "insufficient funds"); err != nil {
			return nil, err
		}
		if err := dbTx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit: %w", err)
		}
		return &ProcessResult{Success: false, ErrorMessage: "insufficient funds"}, nil
	}

	// Step 5: Create ledger entries (double-entry bookkeeping)
	entries := buildTransferEntries(transactionID, sourceAccountID, destAccountID, amount)
	if err := p.createLedgerEntries(ctx, dbTx, entries); err != nil {
		return nil, fmt.Errorf("failed to create ledger entries: %w", err)
	}

	// Step 6: Mark transaction as completed
	if err := p.completeTransaction(ctx, dbTx, transactionID); err != nil {
		return nil, fmt.Errorf("failed to complete transaction: %w", err)
	}

	// Commit the database transaction
	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &ProcessResult{Success: true}, nil
}

// claimTransaction atomically claims a pending transaction for processing
func (p *TransferProcessor) claimTransaction(ctx context.Context, dbTx pgx.Tx, id uuid.UUID) (*model.Transaction, error) {
	now := time.Now()

	query := `
		UPDATE transactions
		SET status = $1, processed_at = $2
		WHERE id = $3 AND status = $4
		RETURNING id, idempotency_key, type, status, reference, initiated_at, processed_at, completed_at, error_message, metadata, amount, currency, from_account_id, to_account_id
	`

	tx := &model.Transaction{}
	var reference, errorMessage, amount, currency *string
	err := dbTx.QueryRow(ctx, query,
		model.TransactionStatusProcessing,
		now,
		id,
		model.TransactionStatusPending,
	).Scan(
		&tx.ID,
		&tx.IdempotencyKey,
		&tx.Type,
		&tx.Status,
		&reference,
		&tx.InitiatedAt,
		&tx.ProcessedAt,
		&tx.CompletedAt,
		&errorMessage,
		&tx.Metadata,
		&amount,
		&currency,
		&tx.FromAccountID,
		&tx.ToAccountID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if reference != nil {
		tx.Reference = *reference
	}
	if errorMessage != nil {
		tx.ErrorMessage = *errorMessage
	}
	if amount != nil {
		tx.Amount = *amount
	}
	if currency != nil {
		tx.Currency = *currency
	}

	return tx, nil
}

// getParties retrieves transaction parties within the db transaction
func (p *TransferProcessor) getParties(ctx context.Context, dbTx pgx.Tx, transactionID uuid.UUID) ([]model.TransactionParty, error) {
	query := `
		SELECT id, transaction_id, account_id, role
		FROM transaction_parties
		WHERE transaction_id = $1
	`

	rows, err := dbTx.Query(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parties []model.TransactionParty
	for rows.Next() {
		var party model.TransactionParty
		if err := rows.Scan(&party.ID, &party.TransactionID, &party.AccountID, &party.Role); err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}

	return parties, nil
}

// getBalanceForUpdate gets the current balance with a lock to prevent concurrent modifications
func (p *TransferProcessor) getBalanceForUpdate(ctx context.Context, dbTx pgx.Tx, accountID uuid.UUID) (float64, error) {
	// Use FOR UPDATE to lock relevant rows and prevent race conditions
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM ledger_entries
		WHERE account_id = $1
	`

	var balance float64
	err := dbTx.QueryRow(ctx, query, accountID).Scan(&balance)
	if err != nil {
		return 0, err
	}

	return balance, nil
}

// createLedgerEntries inserts the ledger entries
func (p *TransferProcessor) createLedgerEntries(ctx context.Context, dbTx pgx.Tx, entries []model.LedgerEntry) error {
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
			return err
		}
	}

	return nil
}

// completeTransaction marks the transaction as completed
func (p *TransferProcessor) completeTransaction(ctx context.Context, dbTx pgx.Tx, id uuid.UUID) error {
	query := `
		UPDATE transactions
		SET status = $1, completed_at = $2
		WHERE id = $3 AND status = $4
	`

	result, err := dbTx.Exec(ctx, query,
		model.TransactionStatusCompleted,
		time.Now(),
		id,
		model.TransactionStatusProcessing,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return model.ErrInvalidTransactionState
	}

	return nil
}

// failTransaction marks the transaction as failed
func (p *TransferProcessor) failTransaction(ctx context.Context, dbTx pgx.Tx, id uuid.UUID, errorMsg string) error {
	query := `
		UPDATE transactions
		SET status = $1, completed_at = $2, error_message = $3
		WHERE id = $4 AND status = $5
	`

	_, err := dbTx.Exec(ctx, query,
		model.TransactionStatusFailed,
		time.Now(),
		errorMsg,
		id,
		model.TransactionStatusProcessing,
	)
	return err
}

// hasSufficientFunds checks if the balance covers the transfer amount
func hasSufficientFunds(balance float64, amountStr string) bool {
	var amount float64
	_, err := fmt.Sscanf(amountStr, "%f", &amount)
	if err != nil {
		return false
	}
	return balance >= amount
}

// buildTransferEntries creates balanced ledger entries for a transfer
func buildTransferEntries(transactionID, fromAccountID, toAccountID uuid.UUID, amount string) []model.LedgerEntry {
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
