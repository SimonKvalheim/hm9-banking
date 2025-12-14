package model

import (
	"time"

	"github.com/google/uuid"
)

// AccountType represents the type of bank account
type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeLoan     AccountType = "loan"
)

// AccountStatus represents the current status of an account
type AccountStatus string

const (
	AccountStatusActive AccountStatus = "active"
	AccountStatusFrozen AccountStatus = "frozen"
	AccountStatusClosed AccountStatus = "closed"
)

// Account represents a bank account
type Account struct {
	ID            uuid.UUID     `json:"id"`
	AccountNumber string        `json:"account_number"`
	AccountType   AccountType   `json:"account_type"`
	Currency      string        `json:"currency"`
	Status        AccountStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// CreateAccountRequest is the payload for creating a new account
type CreateAccountRequest struct {
	AccountType AccountType `json:"account_type"`
	Currency    string      `json:"currency"`
}

// Validate checks if the create request is valid
func (r CreateAccountRequest) Validate() error {
	if r.AccountType != AccountTypeChecking &&
		r.AccountType != AccountTypeSavings &&
		r.AccountType != AccountTypeLoan {
		return ErrInvalidAccountType
	}

	if len(r.Currency) != 3 {
		return ErrInvalidCurrency
	}

	return nil
}

// AccountBalance represents an account's current balance
type AccountBalance struct {
	AccountID uuid.UUID `json:"account_id"`
	Balance   string    `json:"balance"`
	Currency  string    `json:"currency"`
	AsOf      time.Time `json:"as_of"`
}