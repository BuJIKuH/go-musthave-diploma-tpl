package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"go-musthave-diploma-tpl/internal/handler"
	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

// --- Мок для OrdersServicer ---
type mockOrdersService struct {
	UploadFunc func(ctx context.Context, userID, number string) error
	ListFunc   func(ctx context.Context, userID string) ([]postgres.Order, error)
}

func (m *mockOrdersService) UploadOrder(ctx context.Context, userID, number string) error {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, userID, number)
	}
	return nil
}

func (m *mockOrdersService) ListOrders(ctx context.Context, userID string) ([]postgres.Order, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, userID)
	}
	return nil, nil
}

// --- Тест UploadOrder ---
func TestOrdersHandler_UploadOrder(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockSvc := &mockOrdersService{
		UploadFunc: func(ctx context.Context, userID, number string) error {
			if number == "exists" {
				return postgres.ErrOrderExists
			}
			return nil
		},
	}

	h := handler.NewOrdersHandler(mockSvc, logger)

	tests := []struct {
		name           string
		userID         string
		body           string
		wantStatusCode int
	}{
		{"success", "user1", "12345678903", http.StatusAccepted},
		{"order exists", "user1", "exists", http.StatusConflict},
		{"no user", "", "12345678903", http.StatusUnauthorized},
		{"empty body", "user1", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(tt.body))
			if tt.userID != "" {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, tt.userID))
			}
			rr := httptest.NewRecorder()
			h.UploadOrder(rr, req)
			if rr.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

// --- Тест ListOrders ---
func TestOrdersHandler_ListOrders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockSvc := &mockOrdersService{
		ListFunc: func(ctx context.Context, userID string) ([]postgres.Order, error) {
			if userID == "empty" {
				return []postgres.Order{}, nil
			}
			return []postgres.Order{
				{
					ID:         "1",
					Number:     "12345678903",
					UserID:     userID,
					Status:     "NEW",
					UploadedAt: time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewOrdersHandler(mockSvc, logger)

	tests := []struct {
		name           string
		userID         string
		wantStatusCode int
		wantBodyCount  int
	}{
		{"success", "user1", http.StatusOK, 1},
		{"no orders", "empty", http.StatusNoContent, 0},
		{"unauthorized", "", http.StatusUnauthorized, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if tt.userID != "" {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, tt.userID))
			}
			rr := httptest.NewRecorder()
			h.ListOrders(rr, req)
			if rr.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatusCode)
			}
			if tt.wantBodyCount > 0 {
				var orders []postgres.Order
				if err := json.NewDecoder(rr.Body).Decode(&orders); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if len(orders) != tt.wantBodyCount {
					t.Errorf("got %d orders, want %d", len(orders), tt.wantBodyCount)
				}
			}
		})
	}
}
