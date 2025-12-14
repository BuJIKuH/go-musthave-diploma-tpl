package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go-musthave-diploma-tpl/internal/middleware"
	"io"
	"net/http"

	"go-musthave-diploma-tpl/internal/repository/postgres"

	"go.uber.org/zap"
)

type OrdersServicer interface {
	UploadOrder(ctx context.Context, userID, number string) error
	ListOrders(ctx context.Context, userID string) ([]postgres.Order, error)
}

type OrdersHandler struct {
	ordersService OrdersServicer
	logger        *zap.Logger
}

func NewOrdersHandler(ordersService OrdersServicer, logger *zap.Logger) *OrdersHandler {
	return &OrdersHandler{
		ordersService: ordersService,
		logger:        logger,
	}
}

func (h *OrdersHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "order number required", http.StatusBadRequest)
		return
	}

	number := string(body)

	err = h.ordersService.UploadOrder(r.Context(), userID, number)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrOrderExists):
			http.Error(w, "order already exists", http.StatusConflict)
		default:
			h.logger.Error("upload order error", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.ordersService.ListOrders(r.Context(), userID)
	if err != nil {
		h.logger.Error("list orders error", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
