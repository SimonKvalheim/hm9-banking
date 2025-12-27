package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

// CustomerRepository handles database operations for customers
type CustomerRepository struct {
	db *pgxpool.Pool
}

// NewCustomerRepository creates a new CustomerRepository
func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Create inserts a new customer into the database
func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	query := `
		INSERT INTO customers (
			id, email, password_hash, first_name, last_name,
			phone, phone_verified, email_verified,
			address_line1, address_line2, city, postal_code, country,
			date_of_birth, national_id_number,
			status, failed_login_attempts, locked_until, last_login_at, password_changed_at,
			preferred_language, timezone,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15,
			$16, $17, $18, $19, $20,
			$21, $22,
			$23, $24
		)
	`

	_, err := r.db.Exec(ctx, query,
		customer.ID,
		customer.Email,
		customer.PasswordHash,
		customer.FirstName,
		customer.LastName,
		customer.Phone,
		customer.PhoneVerified,
		customer.EmailVerified,
		customer.AddressLine1,
		customer.AddressLine2,
		customer.City,
		customer.PostalCode,
		customer.Country,
		customer.DateOfBirth,
		customer.NationalIDNumber,
		customer.Status,
		customer.FailedLoginAttempts,
		customer.LockedUntil,
		customer.LastLoginAt,
		customer.PasswordChangedAt,
		customer.PreferredLanguage,
		customer.Timezone,
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation on email
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.Message, "email") {
				return nil, model.ErrEmailAlreadyExists
			}
		}
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetByID retrieves a customer by their ID
func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Customer, error) {
	query := `
		SELECT
			id, email, password_hash, first_name, last_name,
			phone, phone_verified, email_verified,
			address_line1, address_line2, city, postal_code, country,
			date_of_birth, national_id_number,
			status, failed_login_attempts, locked_until, last_login_at, password_changed_at,
			preferred_language, timezone,
			created_at, updated_at
		FROM customers
		WHERE id = $1
	`

	customer := &model.Customer{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&customer.ID,
		&customer.Email,
		&customer.PasswordHash,
		&customer.FirstName,
		&customer.LastName,
		&customer.Phone,
		&customer.PhoneVerified,
		&customer.EmailVerified,
		&customer.AddressLine1,
		&customer.AddressLine2,
		&customer.City,
		&customer.PostalCode,
		&customer.Country,
		&customer.DateOfBirth,
		&customer.NationalIDNumber,
		&customer.Status,
		&customer.FailedLoginAttempts,
		&customer.LockedUntil,
		&customer.LastLoginAt,
		&customer.PasswordChangedAt,
		&customer.PreferredLanguage,
		&customer.Timezone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetByEmail retrieves a customer by their email address (case-insensitive)
func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*model.Customer, error) {
	query := `
		SELECT
			id, email, password_hash, first_name, last_name,
			phone, phone_verified, email_verified,
			address_line1, address_line2, city, postal_code, country,
			date_of_birth, national_id_number,
			status, failed_login_attempts, locked_until, last_login_at, password_changed_at,
			preferred_language, timezone,
			created_at, updated_at
		FROM customers
		WHERE LOWER(email) = LOWER($1)
	`

	customer := &model.Customer{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&customer.ID,
		&customer.Email,
		&customer.PasswordHash,
		&customer.FirstName,
		&customer.LastName,
		&customer.Phone,
		&customer.PhoneVerified,
		&customer.EmailVerified,
		&customer.AddressLine1,
		&customer.AddressLine2,
		&customer.City,
		&customer.PostalCode,
		&customer.Country,
		&customer.DateOfBirth,
		&customer.NationalIDNumber,
		&customer.Status,
		&customer.FailedLoginAttempts,
		&customer.LockedUntil,
		&customer.LastLoginAt,
		&customer.PasswordChangedAt,
		&customer.PreferredLanguage,
		&customer.Timezone,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return customer, nil
}

// UpdateLastLogin updates the last login timestamp for a customer
func (r *CustomerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE customers
		SET last_login_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return model.ErrCustomerNotFound
	}

	return nil
}

// IncrementFailedAttempts increments the failed login attempts counter
// Uses SELECT FOR UPDATE to prevent race conditions
func (r *CustomerRepository) IncrementFailedAttempts(ctx context.Context, id uuid.UUID) (int, error) {
	// Start a transaction for SELECT FOR UPDATE
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the row and get current attempts
	var currentAttempts int
	err = tx.QueryRow(ctx, `
		SELECT failed_login_attempts
		FROM customers
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&currentAttempts)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, model.ErrCustomerNotFound
		}
		return 0, fmt.Errorf("failed to get customer for update: %w", err)
	}

	// Increment the counter
	newAttempts := currentAttempts + 1
	_, err = tx.Exec(ctx, `
		UPDATE customers
		SET failed_login_attempts = $1, updated_at = NOW()
		WHERE id = $2
	`, newAttempts, id)
	if err != nil {
		return 0, fmt.Errorf("failed to increment failed attempts: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newAttempts, nil
}

// ResetFailedAttempts resets the failed login attempts counter to zero
func (r *CustomerRepository) ResetFailedAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE customers
		SET failed_login_attempts = 0, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}

	if result.RowsAffected() == 0 {
		return model.ErrCustomerNotFound
	}

	return nil
}

// LockAccount locks a customer account until the specified time
func (r *CustomerRepository) LockAccount(ctx context.Context, id uuid.UUID, until time.Time) error {
	query := `
		UPDATE customers
		SET locked_until = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, until, id)
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return model.ErrCustomerNotFound
	}

	return nil
}
