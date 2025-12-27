package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/simonkvalheim/hm9-banking/internal/middleware"
	"github.com/simonkvalheim/hm9-banking/internal/model"
	"github.com/simonkvalheim/hm9-banking/internal/repository"
)

// AccountHandler handles HTTP requests for accounts
type AccountHandler struct {
	repo *repository.AccountRepository
}

// NewAccountHandler creates a new AccountHandler
func NewAccountHandler(repo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

// RegisterRoutes sets up the account routes on the given router
func (h *AccountHandler) RegisterRoutes(r chi.Router) {
	r.Route("/accounts", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.GetByID)
		r.Get("/{id}/balance", h.GetBalance)
	})
}

// Create handles POST /accounts
// Associates new account with authenticated customer
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	customerID := middleware.GetCustomerID(r.Context())
	if customerID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req model.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create account linked to this customer
	account, err := h.repo.CreateForCustomer(r.Context(), req, customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	writeJSON(w, http.StatusCreated, account)
}

// List handles GET /accounts
// Returns only accounts belonging to the authenticated customer
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	customerID := middleware.GetCustomerID(r.Context())
	if customerID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Only get accounts for this customer
	accounts, err := h.repo.GetByCustomerID(r.Context(), customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list accounts")
		return
	}

	// Return empty array instead of null if no accounts
	if accounts == nil {
		accounts = []model.Account{}
	}

	writeJSON(w, http.StatusOK, accounts)
}

// GetByID handles GET /accounts/{id}
// Verifies the account belongs to the authenticated customer
func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	customerID := middleware.GetCustomerID(r.Context())
	if customerID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid account ID format")
		return
	}

	account, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, "Account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get account")
		return
	}

	// Authorization check: account must belong to authenticated customer
	if account.CustomerID == nil || *account.CustomerID != customerID {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	writeJSON(w, http.StatusOK, account)
}

// GetBalance handles GET /accounts/{id}/balance
// Optional query parameter: as_of (ISO 8601 timestamp) for point-in-time balance
// Verifies the account belongs to the authenticated customer
func (h *AccountHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	customerID := middleware.GetCustomerID(r.Context())
	if customerID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid account ID format")
		return
	}

	// Verify account belongs to customer before getting balance
	account, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, "Account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get account")
		return
	}

	// Authorization check
	if account.CustomerID == nil || *account.CustomerID != customerID {
		writeError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Parse optional as_of query parameter
	var asOf *time.Time
	asOfParam := r.URL.Query().Get("as_of")
	if asOfParam != "" {
		parsed, err := time.Parse(time.RFC3339, asOfParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid as_of format: use ISO 8601 (e.g., 2024-12-13T10:00:00Z)")
			return
		}
		asOf = &parsed
	}

	balance, err := h.repo.GetBalanceAtTime(r.Context(), id, asOf)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, "Account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get balance")
		return
	}

	writeJSON(w, http.StatusOK, balance)
}

// Helper functions for HTTP responses

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}