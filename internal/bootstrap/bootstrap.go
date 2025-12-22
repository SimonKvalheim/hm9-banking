package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

// Initialize ensures all required system accounts exist
// This should be called on server startup after database connection is established
func Initialize(ctx context.Context, db *pgxpool.Pool) error {
	if err := ensureEquityAccount(ctx, db); err != nil {
		return fmt.Errorf("failed to ensure equity account: %w", err)
	}

	return nil
}

// ensureEquityAccount creates the bank equity account if it doesn't exist
func ensureEquityAccount(ctx context.Context, db *pgxpool.Pool) error {
	// Check if equity account already exists
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM accounts WHERE account_number = $1)
	`, model.BankEquityAccountNumber).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check for equity account: %w", err)
	}

	if exists {
		log.Printf("Bank equity account %s already exists", model.BankEquityAccountNumber)
		return nil
	}

	// Create the equity account
	now := time.Now()
	_, err = db.Exec(ctx, `
		INSERT INTO accounts (id, account_number, account_type, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		uuid.New(),
		model.BankEquityAccountNumber,
		model.AccountTypeEquity,
		"NOK", // Default currency for the bank
		model.AccountStatusActive,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create equity account: %w", err)
	}

	log.Printf("Created bank equity account: %s", model.BankEquityAccountNumber)
	return nil
}

// GetEquityAccount retrieves the bank equity account
func GetEquityAccount(ctx context.Context, db *pgxpool.Pool) (*model.Account, error) {
	query := `
		SELECT id, account_number, account_type, currency, status, created_at, updated_at
		FROM accounts
		WHERE account_number = $1
	`

	account := &model.Account{}
	err := db.QueryRow(ctx, query, model.BankEquityAccountNumber).Scan(
		&account.ID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("bank equity account not found - run Initialize first")
	}

	return account, nil
}
