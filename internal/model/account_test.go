package model

import (
	"testing"
)

func TestCreateAccountRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateAccountRequest
		wantErr error
	}{
		{
			name: "valid checking account",
			request: CreateAccountRequest{
				AccountType: AccountTypeChecking,
				Currency:    "NOK",
			},
			wantErr: nil,
		},
		{
			name: "valid savings account",
			request: CreateAccountRequest{
				AccountType: AccountTypeSavings,
				Currency:    "EUR",
			},
			wantErr: nil,
		},
		{
			name: "valid loan account",
			request: CreateAccountRequest{
				AccountType: AccountTypeLoan,
				Currency:    "USD",
			},
			wantErr: nil,
		},
		{
			name: "invalid account type",
			request: CreateAccountRequest{
				AccountType: "invalid",
				Currency:    "NOK",
			},
			wantErr: ErrInvalidAccountType,
		},
		{
			name: "empty account type",
			request: CreateAccountRequest{
				AccountType: "",
				Currency:    "NOK",
			},
			wantErr: ErrInvalidAccountType,
		},
		{
			name: "currency too short",
			request: CreateAccountRequest{
				AccountType: AccountTypeChecking,
				Currency:    "NO",
			},
			wantErr: ErrInvalidCurrency,
		},
		{
			name: "currency too long",
			request: CreateAccountRequest{
				AccountType: AccountTypeChecking,
				Currency:    "NOKK",
			},
			wantErr: ErrInvalidCurrency,
		},
		{
			name: "empty currency",
			request: CreateAccountRequest{
				AccountType: AccountTypeChecking,
				Currency:    "",
			},
			wantErr: ErrInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccountType_Constants(t *testing.T) {
	// Verify constants have expected values
	if AccountTypeChecking != "checking" {
		t.Errorf("AccountTypeChecking = %v, want checking", AccountTypeChecking)
	}
	if AccountTypeSavings != "savings" {
		t.Errorf("AccountTypeSavings = %v, want savings", AccountTypeSavings)
	}
	if AccountTypeLoan != "loan" {
		t.Errorf("AccountTypeLoan = %v, want loan", AccountTypeLoan)
	}
}

func TestAccountStatus_Constants(t *testing.T) {
	// Verify constants have expected values
	if AccountStatusActive != "active" {
		t.Errorf("AccountStatusActive = %v, want active", AccountStatusActive)
	}
	if AccountStatusFrozen != "frozen" {
		t.Errorf("AccountStatusFrozen = %v, want frozen", AccountStatusFrozen)
	}
	if AccountStatusClosed != "closed" {
		t.Errorf("AccountStatusClosed = %v, want closed", AccountStatusClosed)
	}
}
