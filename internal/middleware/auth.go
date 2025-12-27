package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/simonkvalheim/hm9-banking/internal/auth"
)

// ContextKey is the type for context keys to avoid collisions
type ContextKey string

const (
	// CustomerIDKey is the context key for the authenticated customer ID
	CustomerIDKey ContextKey = "customer_id"
	// CustomerEmailKey is the context key for the authenticated customer email
	CustomerEmailKey ContextKey = "customer_email"
)

// AuthMiddleware validates JWT tokens and adds customer info to context
type AuthMiddleware struct {
	authService *auth.Service
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth is middleware that requires a valid access token
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "Missing authorization header")
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeUnauthorized(w, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			writeUnauthorized(w, "Invalid or expired token")
			return
		}

		// Ensure it's an access token (not a refresh token)
		if claims.TokenType != "access" {
			writeUnauthorized(w, "Invalid token type")
			return
		}

		// Add customer info to request context
		ctx := context.WithValue(r.Context(), CustomerIDKey, claims.CustomerID)
		ctx = context.WithValue(ctx, CustomerEmailKey, claims.Email)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCustomerID extracts the customer ID from the request context
// Returns uuid.Nil if not authenticated (shouldn't happen if RequireAuth was used)
func GetCustomerID(ctx context.Context) uuid.UUID {
	id, ok := ctx.Value(CustomerIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}

// GetCustomerEmail extracts the customer email from the request context
func GetCustomerEmail(ctx context.Context) string {
	email, ok := ctx.Value(CustomerEmailKey).(string)
	if !ok {
		return ""
	}
	return email
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error": "` + message + `"}`))
}
