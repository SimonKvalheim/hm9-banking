package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/simonkvalheim/hm9-banking/internal/model"
	"github.com/simonkvalheim/hm9-banking/internal/repository"
)

// Config holds authentication configuration
type Config struct {
	JWTSecret          []byte        // Secret key for signing tokens
	AccessTokenExpiry  time.Duration // How long access tokens are valid
	RefreshTokenExpiry time.Duration // How long refresh tokens are valid
	MaxFailedAttempts  int           // Lock account after this many failures
	LockDuration       time.Duration // How long to lock account
}

// DefaultConfig returns sensible defaults
func DefaultConfig(jwtSecret string) Config {
	return Config{
		JWTSecret:          []byte(jwtSecret),
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		MaxFailedAttempts:  5,
		LockDuration:       15 * time.Minute,
	}
}

// Claims represents the JWT payload
type Claims struct {
	jwt.RegisteredClaims
	CustomerID uuid.UUID `json:"customer_id"`
	Email      string    `json:"email"`
	TokenType  string    `json:"token_type"` // "access" or "refresh"
}

// Service handles authentication operations
type Service struct {
	config       Config
	customerRepo *repository.CustomerRepository
}

// NewService creates a new auth service
func NewService(config Config, customerRepo *repository.CustomerRepository) *Service {
	return &Service{
		config:       config,
		customerRepo: customerRepo,
	}
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"` // Access token expiry
}

// Register creates a new customer account
func (s *Service) Register(ctx context.Context, req model.CreateCustomerRequest) (*model.Customer, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if email already exists
	existing, err := s.customerRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, model.ErrEmailAlreadyExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create customer
	customer := &model.Customer{
		ID:                uuid.New(),
		Email:             req.Email,
		PasswordHash:      string(hash),
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Status:            model.CustomerStatusActive,
		PreferredLanguage: "en",
		Timezone:          "UTC",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return s.customerRepo.Create(ctx, customer)
}

// Login authenticates a customer and returns tokens
func (s *Service) Login(ctx context.Context, req model.LoginRequest) (*TokenPair, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find customer
	customer, err := s.customerRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal whether email exists
		return nil, model.ErrInvalidCredentials
	}

	// Check if account can login
	if !customer.CanLogin() {
		if customer.IsLocked() {
			return nil, model.ErrAccountLocked
		}
		return nil, model.ErrAccountSuspended
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password))
	if err != nil {
		// Wrong password - increment failed attempts
		s.handleFailedLogin(ctx, customer)
		return nil, model.ErrInvalidCredentials
	}

	// Success - reset failed attempts and update last login
	s.customerRepo.ResetFailedAttempts(ctx, customer.ID)
	s.customerRepo.UpdateLastLogin(ctx, customer.ID)

	// Generate tokens
	return s.generateTokenPair(customer)
}

// RefreshTokens generates new tokens using a valid refresh token
func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Parse and validate the refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Ensure it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	// Fetch customer to ensure they still exist and are active
	customer, err := s.customerRepo.GetByID(ctx, claims.CustomerID)
	if err != nil {
		return nil, err
	}

	if customer.Status != model.CustomerStatusActive {
		return nil, model.ErrAccountSuspended
	}

	// Generate new token pair
	return s.generateTokenPair(customer)
}

// ValidateToken parses and validates a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.config.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// generateTokenPair creates access and refresh tokens for a customer
func (s *Service) generateTokenPair(customer *model.Customer) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.config.AccessTokenExpiry)
	refreshExpiry := now.Add(s.config.RefreshTokenExpiry)

	// Access token claims
	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   customer.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			Issuer:    "fjord-bank",
		},
		CustomerID: customer.ID,
		Email:      customer.Email,
		TokenType:  "access",
	}

	// Refresh token claims
	refreshClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   customer.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			Issuer:    "fjord-bank",
		},
		CustomerID: customer.ID,
		Email:      customer.Email,
		TokenType:  "refresh",
	}

	// Sign tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessSigned, err := accessToken.SignedString(s.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSigned, err := refreshToken.SignedString(s.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessSigned,
		RefreshToken: refreshSigned,
		ExpiresAt:    accessExpiry,
	}, nil
}

// handleFailedLogin increments failed attempts and locks if necessary
func (s *Service) handleFailedLogin(ctx context.Context, customer *model.Customer) {
	attempts, _ := s.customerRepo.IncrementFailedAttempts(ctx, customer.ID)

	if attempts >= s.config.MaxFailedAttempts {
		lockUntil := time.Now().Add(s.config.LockDuration)
		s.customerRepo.LockAccount(ctx, customer.ID, lockUntil)
	}
}

// HashPassword is a utility for hashing passwords (useful for tests/seeding)
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
