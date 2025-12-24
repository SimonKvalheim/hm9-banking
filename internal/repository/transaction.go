package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository struct {
	db *pgxpool.Pool
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a new transaction and its parties into the database
func (r *TransactionRepository) Create(ctx context.Context, tx model.Transaction, parties []model.TransactionParty) (*model.Transaction, error) {
	// Start a database transaction
	dbTx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)

	// Insert the transaction
	query := `
		INSERT INTO transactions (id, idempotency_key, type, status, reference, initiated_at, metadata, amount, currency, from_account_id, to_account_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	metadata := tx.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	_, err = dbTx.Exec(ctx, query,
		tx.ID,
		tx.IdempotencyKey,
		tx.Type,
		tx.Status,
		tx.Reference,
		tx.InitiatedAt,
		metadata,
		tx.Amount,
		tx.Currency,
		tx.FromAccountID,
		tx.ToAccountID,
	)
	if err != nil {
		// Check for unique constraint violation on idempotency_key
		if isUniqueViolation(err) {
			return nil, model.ErrTransactionExists
		}
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Insert transaction parties
	partyQuery := `
		INSERT INTO transaction_parties (id, transaction_id, account_id, role)
		VALUES ($1, $2, $3, $4)
	`
	for _, party := range parties {
		_, err = dbTx.Exec(ctx, partyQuery,
			party.ID,
			party.TransactionID,
			party.AccountID,
			party.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create transaction party: %w", err)
		}
	}

	if err = dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &tx, nil
}

// GetByID retrieves a transaction by its ID
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	query := `
		SELECT id, idempotency_key, type, status, reference, initiated_at, processed_at, completed_at, error_message, metadata, amount, currency, from_account_id, to_account_id
		FROM transactions
		WHERE id = $1
	`

	tx := &model.Transaction{}
	var reference, errorMessage, amount, currency *string
	err := r.db.QueryRow(ctx, query, id).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
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

// GetByIdempotencyKey retrieves a transaction by its idempotency key
func (r *TransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	query := `
		SELECT id, idempotency_key, type, status, reference, initiated_at, processed_at, completed_at, error_message, metadata, amount, currency, from_account_id, to_account_id
		FROM transactions
		WHERE idempotency_key = $1
	`

	tx := &model.Transaction{}
	var reference, errorMessage, amount, currency *string
	err := r.db.QueryRow(ctx, query, key).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get transaction by idempotency key: %w", err)
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

// GetParties retrieves all parties for a transaction
func (r *TransactionRepository) GetParties(ctx context.Context, transactionID uuid.UUID) ([]model.TransactionParty, error) {
	query := `
		SELECT id, transaction_id, account_id, role
		FROM transaction_parties
		WHERE transaction_id = $1
	`

	rows, err := r.db.Query(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction parties: %w", err)
	}
	defer rows.Close()

	var parties []model.TransactionParty
	for rows.Next() {
		var party model.TransactionParty
		err := rows.Scan(
			&party.ID,
			&party.TransactionID,
			&party.AccountID,
			&party.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction party: %w", err)
		}
		parties = append(parties, party)
	}

	return parties, nil
}

// UpdateStatus updates the status of a transaction with appropriate timestamp
func (r *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TransactionStatus, errorMessage string) error {
	now := time.Now()

	var query string
	var args []any

	switch status {
	case model.TransactionStatusProcessing:
		query = `
			UPDATE transactions
			SET status = $1, processed_at = $2
			WHERE id = $3 AND status = $4
		`
		args = []any{status, now, id, model.TransactionStatusPending}

	case model.TransactionStatusCompleted:
		query = `
			UPDATE transactions
			SET status = $1, completed_at = $2
			WHERE id = $3 AND status = $4
		`
		args = []any{status, now, id, model.TransactionStatusProcessing}

	case model.TransactionStatusFailed:
		query = `
			UPDATE transactions
			SET status = $1, completed_at = $2, error_message = $3
			WHERE id = $4 AND status = $5
		`
		args = []any{status, now, errorMessage, id, model.TransactionStatusProcessing}

	default:
		return model.ErrInvalidTransactionState
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return model.ErrInvalidTransactionState
	}

	return nil
}

// ClaimForProcessing atomically claims a pending transaction for processing
// Returns the transaction if successfully claimed, nil if already claimed/processed
func (r *TransactionRepository) ClaimForProcessing(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	now := time.Now()

	query := `
		UPDATE transactions
		SET status = $1, processed_at = $2
		WHERE id = $3 AND status = $4
		RETURNING id, idempotency_key, type, status, reference, initiated_at, processed_at, completed_at, error_message, metadata, amount, currency, from_account_id, to_account_id
	`

	tx := &model.Transaction{}
	var reference, errorMessage, amount, currency *string
	err := r.db.QueryRow(ctx, query,
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
		if errors.Is(err, pgx.ErrNoRows) {
			// Transaction doesn't exist or is not in pending state
			return nil, nil
		}
		return nil, fmt.Errorf("failed to claim transaction: %w", err)
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

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	// PostgreSQL error code for unique_violation is 23505
	return err != nil && (errors.Is(err, pgx.ErrNoRows) == false) &&
		(err.Error() != "" && (contains(err.Error(), "23505") || contains(err.Error(), "unique") || contains(err.Error(), "duplicate")))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
