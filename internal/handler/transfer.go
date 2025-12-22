package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/simonkvalheim/hm9-banking/internal/model"
	"github.com/simonkvalheim/hm9-banking/internal/repository"
)

// TransferHandler handles HTTP requests for transfers
type TransferHandler struct {
	txRepo      *repository.TransactionRepository
	accountRepo *repository.AccountRepository
}

// NewTransferHandler creates a new TransferHandler
func NewTransferHandler(txRepo *repository.TransactionRepository, accountRepo *repository.AccountRepository) *TransferHandler {
	return &TransferHandler{
		txRepo:      txRepo,
		accountRepo: accountRepo,
	}
}

// RegisterRoutes sets up the transfer routes on the given router
func (h *TransferHandler) RegisterRoutes(r chi.Router) {
	r.Post("/transfers", h.CreateTransfer)
	r.Get("/transactions/{id}", h.GetTransaction)
}

// CreateTransfer handles POST /transfers
// Idempotency-Key header is required for safe retries
func (h *TransferHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	// Extract idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	// Check for existing transaction with this idempotency key
	existingTx, err := h.txRepo.GetByIdempotencyKey(r.Context(), idempotencyKey)
	if err == nil && existingTx != nil {
		// Transaction already exists - return existing result (idempotent behavior)
		writeJSON(w, http.StatusAccepted, model.TransferResponse{
			TransactionID: existingTx.ID,
			Status:        existingTx.Status,
			CreatedAt:     existingTx.InitiatedAt,
		})
		return
	}
	if err != nil && !errors.Is(err, model.ErrTransactionNotFound) {
		writeError(w, http.StatusInternalServerError, "Failed to check idempotency")
		return
	}

	// Parse request body
	var req model.CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request fields
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate amount is a positive number
	if err := validateAmount(req.Amount); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate source account exists and is active
	fromAccount, err := h.accountRepo.GetByID(r.Context(), req.FromAccountID)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			writeError(w, http.StatusBadRequest, "Source account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to validate source account")
		return
	}
	if fromAccount.Status != model.AccountStatusActive {
		writeError(w, http.StatusBadRequest, "Source account is not active")
		return
	}

	// Validate destination account exists and is active
	toAccount, err := h.accountRepo.GetByID(r.Context(), req.ToAccountID)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			writeError(w, http.StatusBadRequest, "Destination account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to validate destination account")
		return
	}
	if toAccount.Status != model.AccountStatusActive {
		writeError(w, http.StatusBadRequest, "Destination account is not active")
		return
	}

	// Validate currencies match
	if fromAccount.Currency != toAccount.Currency {
		writeError(w, http.StatusBadRequest, "Currency mismatch between accounts")
		return
	}
	if req.Currency != fromAccount.Currency {
		writeError(w, http.StatusBadRequest, "Request currency does not match account currency")
		return
	}

	// Create the transaction
	now := time.Now()
	txID := uuid.New()

	tx := model.Transaction{
		ID:             txID,
		IdempotencyKey: idempotencyKey,
		Type:           model.TransactionTypeTransfer,
		Status:         model.TransactionStatusPending,
		Reference:      req.Reference,
		InitiatedAt:    now,
		Amount:         req.Amount,
		Currency:       req.Currency,
		FromAccountID:  &req.FromAccountID,
		ToAccountID:    &req.ToAccountID,
	}

	parties := []model.TransactionParty{
		{
			ID:            uuid.New(),
			TransactionID: txID,
			AccountID:     req.FromAccountID,
			Role:          "source",
		},
		{
			ID:            uuid.New(),
			TransactionID: txID,
			AccountID:     req.ToAccountID,
			Role:          "destination",
		},
	}

	createdTx, err := h.txRepo.Create(r.Context(), tx, parties)
	if err != nil {
		if errors.Is(err, model.ErrTransactionExists) {
			// Race condition: another request created it first
			// Fetch and return the existing transaction
			existingTx, fetchErr := h.txRepo.GetByIdempotencyKey(r.Context(), idempotencyKey)
			if fetchErr != nil {
				writeError(w, http.StatusInternalServerError, "Failed to create transfer")
				return
			}
			writeJSON(w, http.StatusAccepted, model.TransferResponse{
				TransactionID: existingTx.ID,
				Status:        existingTx.Status,
				CreatedAt:     existingTx.InitiatedAt,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to create transfer")
		return
	}

	// Return 202 Accepted - transaction created but processing is async
	writeJSON(w, http.StatusAccepted, model.TransferResponse{
		TransactionID: createdTx.ID,
		Status:        createdTx.Status,
		CreatedAt:     createdTx.InitiatedAt,
	})
}

// GetTransaction handles GET /transactions/{id}
func (h *TransferHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid transaction ID format")
		return
	}

	tx, err := h.txRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrTransactionNotFound) {
			writeError(w, http.StatusNotFound, "Transaction not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get transaction")
		return
	}

	// Build response using the new columns directly
	detail := model.TransactionDetail{
		Transaction:   *tx,
		FromAccountID: tx.FromAccountID,
		ToAccountID:   tx.ToAccountID,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
	}

	writeJSON(w, http.StatusOK, detail)
}

// validateAmount checks if the amount is a valid positive decimal
func validateAmount(amount string) error {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return model.ErrInvalidAmount
	}

	// Parse as float to validate format
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return model.ErrInvalidAmount
	}

	if val <= 0 {
		return model.ErrInvalidAmount
	}

	return nil
}
