package model

import "errors"

var (
	// Account errors
	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidAccountType = errors.New("invalid account type: must be checking, savings, or loan")
	ErrInvalidCurrency    = errors.New("invalid currency: must be 3-letter ISO code")

	// Transaction errors
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrTransactionExists = errors.New("transaction with this idempotency key already exists")
)