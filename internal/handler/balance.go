package handler

import (
	"encoding/json"
	"go-musthave-diploma-tpl/internal/middleware"
	"net/http"

	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type BalanceHandler struct {
	service *service.BalanceService
	logger  *zap.Logger
}

func NewBalanceHandler(s *service.BalanceService, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{service: s, logger: logger}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	current, withdrawn, err := h.service.GetBalance(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := map[string]decimal.Decimal{
		"current":   current,
		"withdrawn": withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Order string          `json:"order"`
		Sum   decimal.Decimal `json:"sum"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.service.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case err == postgres.ErrNotEnoughFunds:
		w.WriteHeader(http.StatusPaymentRequired)
	case err == postgres.ErrInvalidOrder:
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *BalanceHandler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	list, err := h.service.ListWithdrawals(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
