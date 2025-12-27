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
	AccountTypeEquity   AccountType = "equity"
)

// BankEquityAccountNumber is the well-known account number for the bank's equity account
const BankEquityAccountNumber = "BANK-EQUITY-001"

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
	CustomerID    *uuid.UUID    `json:"customer_id,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// IsSystemAccount returns true if this is a system account (e.g., bank equity)
// System accounts bypass certain validations like insufficient funds checks
func (a *Account) IsSystemAccount() bool {
	return a.AccountType == AccountTypeEquity
}

// CreateAccountRequest is the payload for creating a new account
type CreateAccountRequest struct {
	AccountType AccountType `json:"account_type"`
	Currency    string      `json:"currency"`
}

// Validate checks if the create request is valid
func (r CreateAccountRequest) Validate() error {
	// Reject system account types - these can only be created internally
	if r.AccountType == AccountTypeEquity {
		return ErrSystemAccountType
	}

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