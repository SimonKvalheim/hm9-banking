package model

import (
	"time"
	"unicode"

	"github.com/google/uuid"
)

// CustomerStatus represents the current status of a customer
type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusSuspended CustomerStatus = "suspended"
	CustomerStatusClosed    CustomerStatus = "closed"
)

// Customer represents a bank customer
type Customer struct {
	ID           uuid.UUID      `json:"id"`
	Email        string         `json:"email"`
	PasswordHash string         `json:"-"` // Never serialize password hash!
	FirstName    string         `json:"first_name"`
	LastName     string         `json:"last_name"`

	// Optional contact
	Phone         *string `json:"phone,omitempty"`
	PhoneVerified bool    `json:"phone_verified"`
	EmailVerified bool    `json:"email_verified"`

	// Optional address
	AddressLine1 *string `json:"address_line1,omitempty"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         *string `json:"city,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty"`
	Country      *string `json:"country,omitempty"`

	// Optional identity
	DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
	NationalIDNumber *string    `json:"-"` // Sensitive - don't serialize

	// Status & security
	Status              CustomerStatus `json:"status"`
	FailedLoginAttempts int            `json:"-"` // Don't expose
	LockedUntil         *time.Time     `json:"-"` // Don't expose
	LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
	PasswordChangedAt   *time.Time     `json:"-"`

	// Preferences
	PreferredLanguage string `json:"preferred_language"`
	Timezone          string `json:"timezone"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsLocked returns true if the customer account is currently locked
func (c *Customer) IsLocked() bool {
	if c.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*c.LockedUntil)
}

// CanLogin returns true if the customer is allowed to attempt login
func (c *Customer) CanLogin() bool {
	return c.Status == CustomerStatusActive && !c.IsLocked()
}

// CreateCustomerRequest is the payload for registration
type CreateCustomerRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Validate checks if the registration request is valid
func (r CreateCustomerRequest) Validate() error {
	if r.Email == "" || !isValidEmail(r.Email) {
		return ErrInvalidEmail
	}
	if len(r.Password) < 8 {
		return ErrPasswordTooShort
	}
	if !isStrongPassword(r.Password) {
		return ErrPasswordTooWeak
	}
	if r.FirstName == "" {
		return ErrFirstNameRequired
	}
	if r.LastName == "" {
		return ErrLastNameRequired
	}
	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Basic check - contains @ and at least one dot after @
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex < 1 || atIndex >= len(email)-1 {
		return false
	}
	afterAt := email[atIndex+1:]
	for _, c := range afterAt {
		if c == '.' {
			return true
		}
	}
	return false
}

// isStrongPassword checks password complexity
// Requires: 8+ chars, at least one uppercase, one lowercase, one digit
func isStrongPassword(password string) bool {
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

// LoginRequest is the payload for authentication
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate checks if the login request has required fields
func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return ErrInvalidEmail
	}
	if r.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}
