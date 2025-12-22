package handler

import (
	"testing"

	"github.com/simonkvalheim/hm9-banking/internal/model"
)

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		wantErr bool
	}{
		{
			name:    "valid positive amount",
			amount:  "100.00",
			wantErr: false,
		},
		{
			name:    "valid small amount",
			amount:  "0.01",
			wantErr: false,
		},
		{
			name:    "valid integer amount",
			amount:  "100",
			wantErr: false,
		},
		{
			name:    "valid large amount",
			amount:  "999999.99",
			wantErr: false,
		},
		{
			name:    "zero amount",
			amount:  "0",
			wantErr: true,
		},
		{
			name:    "negative amount",
			amount:  "-100.00",
			wantErr: true,
		},
		{
			name:    "empty amount",
			amount:  "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			amount:  "   ",
			wantErr: true,
		},
		{
			name:    "invalid format - letters",
			amount:  "abc",
			wantErr: true,
		},
		{
			name:    "invalid format - mixed",
			amount:  "100abc",
			wantErr: true,
		},
		{
			name:    "amount with spaces - trimmed",
			amount:  " 100.00 ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAmount(%q) error = %v, wantErr %v", tt.amount, err, tt.wantErr)
			}
			if tt.wantErr && err != nil && err != model.ErrInvalidAmount {
				t.Errorf("validateAmount(%q) error = %v, want ErrInvalidAmount", tt.amount, err)
			}
		})
	}
}
