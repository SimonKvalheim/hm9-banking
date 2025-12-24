package model

import "errors"

var (
	// Account errors
	ErrAccountNotFound      = errors.New("account not found")
	ErrInvalidAccountType   = errors.New("invalid account type: must be checking, savings, or loan")
	ErrSystemAccountType    = errors.New("cannot create system account type via API")
	ErrInvalidCurrency      = errors.New("invalid currency: must be 3-letter ISO code")

	// Transaction errors
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrTransactionExists       = errors.New("transaction with this idempotency key already exists")
	ErrTransactionNotFound     = errors.New("transaction not found")
	ErrInvalidTransactionState = errors.New("invalid transaction state transition")
	ErrInvalidFromAccount      = errors.New("invalid source account")
	ErrInvalidToAccount        = errors.New("invalid destination account")
	ErrSameAccount             = errors.New("source and destination accounts must be different")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrCurrencyMismatch        = errors.New("currency mismatch between accounts")
	ErrAccountNotActive        = errors.New("account is not active")
)