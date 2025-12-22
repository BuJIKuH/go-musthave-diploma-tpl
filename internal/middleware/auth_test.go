package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-musthave-diploma-tpl/internal/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "testsecret"
	logger := zaptest.NewLogger(t)

	// создаём тестовый хендлер, который пишет userID в ответ
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r)
		if !ok {
			http.Error(w, "no user ID in context", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, userID)
	})

	validToken, _ := service.GenerateToken("user-123", secret)

	tests := []struct {
		name           string
		authHeader     string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "#1 valid token",
			authHeader:     "Bearer " + validToken,
			wantStatusCode: http.StatusOK,
			wantBody:       "user-123",
		},
		{
			name:           "#2 missing header",
			authHeader:     "",
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "missing auth header\n",
		},
		{
			name:           "#3 invalid header format",
			authHeader:     "Token " + validToken,
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "invalid auth header\n",
		},
		{
			name:           "#4 invalid token",
			authHeader:     "Bearer invalidtoken",
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "invalid token\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			middleware := AuthMiddleware(secret, logger)
			middleware(nextHandler).ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)

			body := w.Body.String()
			assert.Equal(t, tt.wantBody, body)
		})
	}
}

func TestGetUserID(t *testing.T) {
	// создаём request с контекстом
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, UserCtxKey, "user-123")
	req = req.WithContext(ctx)

	userID, ok := GetUserID(req)
	assert.True(t, ok)
	assert.Equal(t, "user-123", userID)

	// тест без значения
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	userID2, ok2 := GetUserID(req2)
	assert.False(t, ok2)
	assert.Equal(t, "", userID2)
}
