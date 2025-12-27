package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/simonkvalheim/hm9-banking/internal/auth"
	"github.com/simonkvalheim/hm9-banking/internal/model"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes sets up the auth routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.RefreshToken)
	r.Post("/auth/logout", h.Logout)
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	customer, err := h.authService.Register(r.Context(), req)
	if err != nil {
		switch err {
		case model.ErrEmailAlreadyExists:
			writeError(w, http.StatusConflict, err.Error())
		case model.ErrInvalidEmail, model.ErrPasswordTooShort,
			model.ErrPasswordTooWeak, model.ErrFirstNameRequired,
			model.ErrLastNameRequired:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	// Return customer without sensitive data (password_hash already excluded via json:"-")
	writeJSON(w, http.StatusCreated, customer)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokens, err := h.authService.Login(r.Context(), req)
	if err != nil {
		switch err {
		case model.ErrInvalidCredentials:
			writeError(w, http.StatusUnauthorized, "Invalid email or password")
		case model.ErrAccountLocked:
			writeError(w, http.StatusForbidden, "Account is temporarily locked")
		case model.ErrAccountSuspended:
			writeError(w, http.StatusForbidden, "Account is suspended")
		default:
			writeError(w, http.StatusInternalServerError, "Login failed")
		}
		return
	}

	// Set refresh token as HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HttpOnly: true,                    // JavaScript cannot access
		Secure:   false,                   // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode, // CSRF protection
		MaxAge:   7 * 24 * 60 * 60,        // 7 days in seconds
	})

	// Return access token in response body
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": tokens.AccessToken,
		"expires_at":   tokens.ExpiresAt,
		"token_type":   "Bearer",
	})
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "No refresh token")
		return
	}

	tokens, err := h.authService.RefreshTokens(r.Context(), cookie.Value)
	if err != nil {
		// Clear invalid cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // Delete cookie
		})
		writeError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Update refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": tokens.AccessToken,
		"expires_at":   tokens.ExpiresAt,
		"token_type":   "Bearer",
	})
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete cookie
	})

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
