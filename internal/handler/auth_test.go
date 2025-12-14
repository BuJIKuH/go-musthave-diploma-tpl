package handler

import (
	"bytes"
	"context"
	"errors"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

/*
================ MOCK AUTH SERVICE =================
*/

type mockAuthService struct {
	registerFn func(ctx context.Context, login, password string) (string, error)
	loginFn    func(ctx context.Context, login, password string) (string, error)
}

func (m *mockAuthService) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFn(ctx, login, password)
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFn(ctx, login, password)
}

/*
================ TESTS =================
*/

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		mockService    *mockAuthService
		wantStatusCode int
		wantHeader     string
	}{
		{
			name:    "#1 success",
			reqBody: `{"login":"user","password":"pass1234"}`,
			mockService: &mockAuthService{
				registerFn: func(ctx context.Context, login, password string) (string, error) {
					return "token123", nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantHeader:     "Bearer token123",
		},
		{
			name:    "#2 invalid JSON",
			reqBody: `invalid-json`,
			mockService: &mockAuthService{
				registerFn: func(ctx context.Context, login, password string) (string, error) {
					return "", nil
				},
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:    "#3 login exists",
			reqBody: `{"login":"user","password":"pass1234"}`,
			mockService: &mockAuthService{
				registerFn: func(ctx context.Context, login, password string) (string, error) {
					return "", postgres.ErrUserExists
				},
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:    "#4 internal error",
			reqBody: `{"login":"user","password":"pass1234"}`,
			mockService: &mockAuthService{
				registerFn: func(ctx context.Context, login, password string) (string, error) {
					return "", errors.New("some error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			handler := NewAuthHandler(tt.mockService, logger)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.reqBody))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, res.Header.Get("Authorization"))
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		mockService    *mockAuthService
		wantStatusCode int
		wantHeader     string
	}{
		{
			name:    "#1 success",
			reqBody: `{"login":"user","password":"pass1234"}`,
			mockService: &mockAuthService{
				loginFn: func(ctx context.Context, login, password string) (string, error) {
					return "token123", nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantHeader:     "Bearer token123",
		},
		{
			name:    "#2 invalid JSON",
			reqBody: `invalid-json`,
			mockService: &mockAuthService{
				loginFn: func(ctx context.Context, login, password string) (string, error) {
					return "", nil
				},
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:    "#3 invalid credentials",
			reqBody: `{"login":"user","password":"pass1234"}`,
			mockService: &mockAuthService{
				loginFn: func(ctx context.Context, login, password string) (string, error) {
					return "", errors.New("invalid credentials")
				},
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:    "#4 internal error",
			reqBody: `{"login":"user","password":"pass1234"}`,
			mockService: &mockAuthService{
				loginFn: func(ctx context.Context, login, password string) (string, error) {
					return "", errors.New("some error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			handler := NewAuthHandler(tt.mockService, logger)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.reqBody))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			if tt.wantHeader != "" {
				assert.Equal(t, tt.wantHeader, res.Header.Get("Authorization"))
			}
		})
	}
}
