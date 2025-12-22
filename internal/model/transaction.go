package model

import (
	"time"

	"github.com/google/uuid"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
)

// TransactionStatus represents the current status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID             uuid.UUID         `json:"id"`
	IdempotencyKey string            `json:"idempotency_key"`
	Type           TransactionType   `json:"type"`
	Status         TransactionStatus `json:"status"`
	Reference      string            `json:"reference,omitempty"`
	InitiatedAt    time.Time         `json:"initiated_at"`
	ProcessedAt    *time.Time        `json:"processed_at,omitempty"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
	// New fields from migration 000002
	Amount        string     `json:"amount,omitempty"`
	Currency      string     `json:"currency,omitempty"`
	FromAccountID *uuid.UUID `json:"from_account_id,omitempty"`
	ToAccountID   *uuid.UUID `json:"to_account_id,omitempty"`
}

// TransactionParty represents a participant in a transaction
type TransactionParty struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	AccountID     uuid.UUID `json:"account_id"`
	Role          string    `json:"role"` // "source" or "destination"
}

// LedgerEntryType represents the type of ledger entry
type LedgerEntryType string

const (
	LedgerEntryTypeDebit  LedgerEntryType = "debit"
	LedgerEntryTypeCredit LedgerEntryType = "credit"
)

// LedgerEntry represents a single entry in the ledger (double-entry bookkeeping)
type LedgerEntry struct {
	ID            uuid.UUID       `json:"id"`
	TransactionID uuid.UUID       `json:"transaction_id"`
	AccountID     uuid.UUID       `json:"account_id"`
	Amount        string          `json:"amount"` // Positive = credit, negative = debit
	EntryType     LedgerEntryType `json:"entry_type"`
	CreatedAt     time.Time       `json:"created_at"`
}

// CreateTransferRequest is the payload for creating a new transfer
type CreateTransferRequest struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        string    `json:"amount"`
	Currency      string    `json:"currency"`
	Reference     string    `json:"reference,omitempty"`
}

// Validate checks if the transfer request is valid
func (r CreateTransferRequest) Validate() error {
	if r.FromAccountID == uuid.Nil {
		return ErrInvalidFromAccount
	}
	if r.ToAccountID == uuid.Nil {
		return ErrInvalidToAccount
	}
	if r.FromAccountID == r.ToAccountID {
		return ErrSameAccount
	}
	if r.Amount == "" {
		return ErrInvalidAmount
	}
	if len(r.Currency) != 3 {
		return ErrInvalidCurrency
	}
	return nil
}

// TransferResponse is the response after creating a transfer
type TransferResponse struct {
	TransactionID uuid.UUID         `json:"transaction_id"`
	Status        TransactionStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
}

// TransactionDetail provides full details of a transaction including parties
type TransactionDetail struct {
	Transaction
	FromAccountID *uuid.UUID `json:"from_account_id,omitempty"`
	ToAccountID   *uuid.UUID `json:"to_account_id,omitempty"`
	Amount        string     `json:"amount,omitempty"`
	Currency      string     `json:"currency,omitempty"`
}
