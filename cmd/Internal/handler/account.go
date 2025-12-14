package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/simonkvalheim/hm9-banking/cmd/internal/model"
	"github.com/simonkvalheim/hm9-banking/cmd/internal/repository"
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
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	account, err := h.repo.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	writeJSON(w, http.StatusCreated, account)
}

// List handles GET /accounts
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.repo.List(r.Context(), 100)
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
func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	writeJSON(w, http.StatusOK, account)
}

// GetBalance handles GET /accounts/{id}/balance
func (h *AccountHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid account ID format")
		return
	}

	balance, err := h.repo.GetBalance(r.Context(), id)
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