package processor

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

func TestHasSufficientFunds(t *testing.T) {
	tests := []struct {
		name    string
		balance float64
		amount  string
		want    bool
	}{
		{
			name:    "sufficient funds - exact",
			balance: 100.00,
			amount:  "100.00",
			want:    true,
		},
		{
			name:    "sufficient funds - more than needed",
			balance: 150.00,
			amount:  "100.00",
			want:    true,
		},
		{
			name:    "insufficient funds",
			balance: 50.00,
			amount:  "100.00",
			want:    false,
		},
		{
			name:    "zero balance",
			balance: 0.00,
			amount:  "100.00",
			want:    false,
		},
		{
			name:    "negative balance",
			balance: -50.00,
			amount:  "100.00",
			want:    false,
		},
		{
			name:    "small amount",
			balance: 0.01,
			amount:  "0.01",
			want:    true,
		},
		{
			name:    "invalid amount format",
			balance: 100.00,
			amount:  "invalid",
			want:    false,
		},
		{
			name:    "empty amount",
			balance: 100.00,
			amount:  "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSufficientFunds(tt.balance, tt.amount)
			if got != tt.want {
				t.Errorf("hasSufficientFunds(%v, %v) = %v, want %v", tt.balance, tt.amount, got, tt.want)
			}
		})
	}
}

func TestBuildTransferEntries(t *testing.T) {
	txID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	amount := "100.00"

	entries := buildTransferEntries(txID, fromID, toID, amount)

	// Should create exactly 2 entries
	if len(entries) != 2 {
		t.Fatalf("buildTransferEntries() returned %d entries, want 2", len(entries))
	}

	// First entry should be debit (negative) from source
	debitEntry := entries[0]
	if debitEntry.TransactionID != txID {
		t.Errorf("debit entry transaction ID = %v, want %v", debitEntry.TransactionID, txID)
	}
	if debitEntry.AccountID != fromID {
		t.Errorf("debit entry account ID = %v, want %v", debitEntry.AccountID, fromID)
	}
	if debitEntry.Amount != "-100.00" {
		t.Errorf("debit entry amount = %v, want -100.00", debitEntry.Amount)
	}
	if debitEntry.EntryType != model.LedgerEntryTypeDebit {
		t.Errorf("debit entry type = %v, want %v", debitEntry.EntryType, model.LedgerEntryTypeDebit)
	}

	// Second entry should be credit (positive) to destination
	creditEntry := entries[1]
	if creditEntry.TransactionID != txID {
		t.Errorf("credit entry transaction ID = %v, want %v", creditEntry.TransactionID, txID)
	}
	if creditEntry.AccountID != toID {
		t.Errorf("credit entry account ID = %v, want %v", creditEntry.AccountID, toID)
	}
	if creditEntry.Amount != "100.00" {
		t.Errorf("credit entry amount = %v, want 100.00", creditEntry.Amount)
	}
	if creditEntry.EntryType != model.LedgerEntryTypeCredit {
		t.Errorf("credit entry type = %v, want %v", creditEntry.EntryType, model.LedgerEntryTypeCredit)
	}

	// Entries should have unique IDs
	if debitEntry.ID == creditEntry.ID {
		t.Error("debit and credit entries have the same ID")
	}

	// Entries should have same timestamp (approximately)
	if debitEntry.CreatedAt != creditEntry.CreatedAt {
		t.Error("debit and credit entries have different timestamps")
	}
}

func TestBuildTransferEntries_DoubleEntryBalance(t *testing.T) {
	// Test that entries sum to zero (core double-entry principle)
	txID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()

	testAmounts := []string{"100.00", "0.01", "999999.99", "1.50"}

	for _, amount := range testAmounts {
		t.Run("amount_"+amount, func(t *testing.T) {
			entries := buildTransferEntries(txID, fromID, toID, amount)

			// Parse amounts and verify they sum to zero
			var sum float64
			for _, entry := range entries {
				var val float64
				_, err := fmt.Sscanf(entry.Amount, "%f", &val)
				if err != nil {
					t.Fatalf("failed to parse amount %s: %v", entry.Amount, err)
				}
				sum += val
			}

			if sum != 0 {
				t.Errorf("entries sum = %v, want 0 (double-entry principle violated)", sum)
			}
		})
	}
}
