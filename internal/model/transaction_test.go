package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateTransferRequest_Validate(t *testing.T) {
	validFromID := uuid.New()
	validToID := uuid.New()

	tests := []struct {
		name    string
		request CreateTransferRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: CreateTransferRequest{
				FromAccountID: validFromID,
				ToAccountID:   validToID,
				Amount:        "100.00",
				Currency:      "NOK",
			},
			wantErr: nil,
		},
		{
			name: "missing from account",
			request: CreateTransferRequest{
				FromAccountID: uuid.Nil,
				ToAccountID:   validToID,
				Amount:        "100.00",
				Currency:      "NOK",
			},
			wantErr: ErrInvalidFromAccount,
		},
		{
			name: "missing to account",
			request: CreateTransferRequest{
				FromAccountID: validFromID,
				ToAccountID:   uuid.Nil,
				Amount:        "100.00",
				Currency:      "NOK",
			},
			wantErr: ErrInvalidToAccount,
		},
		{
			name: "same source and destination",
			request: CreateTransferRequest{
				FromAccountID: validFromID,
				ToAccountID:   validFromID, // Same as from
				Amount:        "100.00",
				Currency:      "NOK",
			},
			wantErr: ErrSameAccount,
		},
		{
			name: "empty amount",
			request: CreateTransferRequest{
				FromAccountID: validFromID,
				ToAccountID:   validToID,
				Amount:        "",
				Currency:      "NOK",
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "invalid currency - too short",
			request: CreateTransferRequest{
				FromAccountID: validFromID,
				ToAccountID:   validToID,
				Amount:        "100.00",
				Currency:      "NO",
			},
			wantErr: ErrInvalidCurrency,
		},
		{
			name: "invalid currency - too long",
			request: CreateTransferRequest{
				FromAccountID: validFromID,
				ToAccountID:   validToID,
				Amount:        "100.00",
				Currency:      "NOKK",
			},
			wantErr: ErrInvalidCurrency,
		},
		{
			name: "valid request with reference",
			request: CreateTransferRequest{
				FromAccountID: validFromID,
				ToAccountID:   validToID,
				Amount:        "50.00",
				Currency:      "EUR",
				Reference:     "Rent payment",
			},
			wantErr: nil,
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
