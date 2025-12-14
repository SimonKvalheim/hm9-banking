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

// AccountRepository handles database operations for accounts
type AccountRepository struct {
	db *pgxpool.Pool
}

// NewAccountRepository creates a new AccountRepository
func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create inserts a new account into the database
func (r *AccountRepository) Create(ctx context.Context, req model.CreateAccountRequest) (*model.Account, error) {
	account := &model.Account{
		ID:            uuid.New(),
		AccountNumber: generateAccountNumber(),
		AccountType:   req.AccountType,
		Currency:      req.Currency,
		Status:        model.AccountStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO accounts (id, account_number, account_type, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		account.ID,
		account.AccountNumber,
		account.AccountType,
		account.Currency,
		account.Status,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// GetByID retrieves an account by its ID
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	query := `
		SELECT id, account_number, account_type, currency, status, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	account := &model.Account{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

// List retrieves all accounts (with a limit for safety)
func (r *AccountRepository) List(ctx context.Context, limit int) ([]model.Account, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, account_number, account_type, currency, status, created_at, updated_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var account model.Account
		err := rows.Scan(
			&account.ID,
			&account.AccountNumber,
			&account.AccountType,
			&account.Currency,
			&account.Status,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetBalance calculates the current balance for an account from ledger entries
func (r *AccountRepository) GetBalance(ctx context.Context, id uuid.UUID) (*model.AccountBalance, error) {
	// First check the account exists and get its currency
	account, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Sum all ledger entries for this account
	query := `
		SELECT COALESCE(SUM(amount), 0) AS balance
		FROM ledger_entries
		WHERE account_id = $1
	`

	var balance string
	err = r.db.QueryRow(ctx, query, id).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &model.AccountBalance{
		AccountID: id,
		Balance:   balance,
		Currency:  account.Currency,
		AsOf:      time.Now(),
	}, nil
}

// generateAccountNumber creates a simple account number
// In production, this would follow a specific format (e.g., IBAN)
func generateAccountNumber() string {
	// Format: NO + 2 check digits + 10 digit account number
	// This is a simplified version - real IBAN/BBAN has specific check digit algorithms
	return fmt.Sprintf("NO%02d%010d", 
		time.Now().UnixNano()%100,
		time.Now().UnixNano()%10000000000,
	)
}