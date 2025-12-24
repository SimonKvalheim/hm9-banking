package repository

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

func TestBuildTransferEntries(t *testing.T) {
	txID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	amount := "250.00"

	entries := BuildTransferEntries(txID, fromID, toID, amount)

	// Should create exactly 2 entries
	if len(entries) != 2 {
		t.Fatalf("BuildTransferEntries() returned %d entries, want 2", len(entries))
	}

	// Find debit and credit entries
	var debitEntry, creditEntry model.LedgerEntry
	for _, entry := range entries {
		if entry.EntryType == model.LedgerEntryTypeDebit {
			debitEntry = entry
		} else if entry.EntryType == model.LedgerEntryTypeCredit {
			creditEntry = entry
		}
	}

	// Verify debit entry
	if debitEntry.AccountID != fromID {
		t.Errorf("debit entry account ID = %v, want %v", debitEntry.AccountID, fromID)
	}
	if debitEntry.Amount != "-250.00" {
		t.Errorf("debit entry amount = %v, want -250.00", debitEntry.Amount)
	}

	// Verify credit entry
	if creditEntry.AccountID != toID {
		t.Errorf("credit entry account ID = %v, want %v", creditEntry.AccountID, toID)
	}
	if creditEntry.Amount != "250.00" {
		t.Errorf("credit entry amount = %v, want 250.00", creditEntry.Amount)
	}

	// Both entries should have same transaction ID
	if debitEntry.TransactionID != txID || creditEntry.TransactionID != txID {
		t.Error("entries have incorrect transaction ID")
	}
}

func TestBuildTransferEntries_SumsToZero(t *testing.T) {
	// Core double-entry bookkeeping invariant: all entries must sum to zero
	testCases := []string{
		"1.00",
		"100.00",
		"0.01",
		"12345.67",
		"999999.99",
	}

	for _, amount := range testCases {
		t.Run("amount_"+amount, func(t *testing.T) {
			entries := BuildTransferEntries(uuid.New(), uuid.New(), uuid.New(), amount)

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
				t.Errorf("entries sum = %v, want 0", sum)
			}
		})
	}
}

func TestBuildTransferEntries_UniqueIDs(t *testing.T) {
	entries := BuildTransferEntries(uuid.New(), uuid.New(), uuid.New(), "100.00")

	if entries[0].ID == entries[1].ID {
		t.Error("entries have duplicate IDs")
	}

	if entries[0].ID == uuid.Nil || entries[1].ID == uuid.Nil {
		t.Error("entries have nil IDs")
	}
}

func TestBuildTransferEntries_SameTimestamp(t *testing.T) {
	entries := BuildTransferEntries(uuid.New(), uuid.New(), uuid.New(), "100.00")

	if entries[0].CreatedAt != entries[1].CreatedAt {
		t.Error("entries have different timestamps - should be created atomically")
	}
}
