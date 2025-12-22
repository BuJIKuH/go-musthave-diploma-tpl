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

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// --- Мок для BalanceService ---
type mockBalanceService struct {
	GetBalanceFunc      func(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error)
	WithdrawFunc        func(ctx context.Context, userID, order string, sum decimal.Decimal) error
	ListWithdrawalsFunc func(ctx context.Context, userID string) ([]postgres.Withdrawal, error)
}

func (m *mockBalanceService) GetBalance(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error) {
	return m.GetBalanceFunc(ctx, userID)
}

func (m *mockBalanceService) Withdraw(ctx context.Context, userID, order string, sum decimal.Decimal) error {
	return m.WithdrawFunc(ctx, userID, order, sum)
}

func (m *mockBalanceService) ListWithdrawals(ctx context.Context, userID string) ([]postgres.Withdrawal, error) {
	return m.ListWithdrawalsFunc(ctx, userID)
}

// --- Тест GetBalance ---
func TestBalanceHandler_GetBalance(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockBalanceService{
		GetBalanceFunc: func(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error) {
			withdrawn := decimal.NewFromInt(200)
			accrued := decimal.NewFromInt(1200)
			current := accrued.Sub(withdrawn) // 1200-200=1000
			return current, withdrawn, nil
		},
	}

	h := handler.NewBalanceHandler(mockSvc, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, "user1"))
	rr := httptest.NewRecorder()
	h.GetBalance(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]float32
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["current"] != 1000 || resp["withdrawn"] != 200 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

// --- Тест Withdraw ---
func TestBalanceHandler_Withdraw(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockBalanceService{
		WithdrawFunc: func(ctx context.Context, userID, order string, sum decimal.Decimal) error {
			if sum.Equal(decimal.NewFromInt(0)) {
				return postgres.ErrInvalidOrder
			}
			return nil
		},
	}

	h := handler.NewBalanceHandler(mockSvc, logger)

	tests := []struct {
		name           string
		userID         string
		body           string
		wantStatusCode int
	}{
		{"success", "user1", `{"order":"123","sum":100}`, http.StatusOK},
		{"invalid order", "user1", `{"order":"123","sum":0}`, http.StatusUnprocessableEntity},
		{"no user", "", `{"order":"123","sum":100}`, http.StatusUnauthorized},
		{"bad body", "user1", `{`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(tt.body))
			if tt.userID != "" {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, tt.userID))
			}
			rr := httptest.NewRecorder()
			h.Withdraw(rr, req)
			if rr.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatusCode)
			}
		})
	}
}

// --- Тест ListWithdrawals ---
func TestBalanceHandler_ListWithdrawals(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	mockSvc := &mockBalanceService{
		ListWithdrawalsFunc: func(ctx context.Context, userID string) ([]postgres.Withdrawal, error) {
			if userID == "empty" {
				return []postgres.Withdrawal{}, nil
			}
			return []postgres.Withdrawal{
				{
					OrderNumber: "123",
					Sum:         "100.50",
					ProcessedAt: time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewBalanceHandler(mockSvc, logger)

	tests := []struct {
		name           string
		userID         string
		wantStatusCode int
		wantBodyCount  int
	}{
		{"success", "user1", http.StatusOK, 1},
		{"no withdrawals", "empty", http.StatusNoContent, 0},
		{"unauthorized", "", http.StatusUnauthorized, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance/withdrawals", nil)
			if tt.userID != "" {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserCtxKey, tt.userID))
			}
			rr := httptest.NewRecorder()
			h.ListWithdrawals(rr, req)
			if rr.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatusCode)
			}
			if tt.wantBodyCount > 0 {
				var resp []handler.WithdrawalResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if len(resp) != tt.wantBodyCount {
					t.Errorf("got %d items, want %d", len(resp), tt.wantBodyCount)
				}
			}
		})
	}
}
