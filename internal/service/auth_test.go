package service

import (
	"context"
	"errors"
	"testing"

	"go-musthave-diploma-tpl/internal/repository/postgres"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/crypto/bcrypt"
)

/*
================ MOCK =================
*/

type mockUserRepo struct {
	createUserFn     func(ctx context.Context, login, hash string) (string, error)
	getUserByLoginFn func(ctx context.Context, login string) (*postgres.User, error)
}

func (m *mockUserRepo) CreateUser(
	ctx context.Context,
	login string,
	passwordHash string,
	logger *zap.Logger,
) (string, error) {
	return m.createUserFn(ctx, login, passwordHash)
}

func (m *mockUserRepo) GetUserByLogin(
	ctx context.Context,
	login string,
	logger *zap.Logger,
) (*postgres.User, error) {
	return m.getUserByLoginFn(ctx, login)
}

/*
================ TESTS =================
*/

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name      string
		login     string
		password  string
		repo      *mockUserRepo
		wantErr   bool
		wantToken bool
	}{
		{
			name:     "#1 success register",
			login:    "test",
			password: "strongpassword",
			repo: &mockUserRepo{
				createUserFn: func(ctx context.Context, login, hash string) (string, error) {
					return "user-id-1", nil
				},
			},
			wantErr:   false,
			wantToken: true,
		},
		{
			name:     "#2 repo error",
			login:    "test",
			password: "strongpassword",
			repo: &mockUserRepo{
				createUserFn: func(ctx context.Context, login, hash string) (string, error) {
					return "", errors.New("db error")
				},
			},
			wantErr:   true,
			wantToken: false,
		},
		{
			name:      "#3 empty login",
			login:     "",
			password:  "strongpassword",
			repo:      &mockUserRepo{},
			wantErr:   false,
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			service := NewAuthService(tt.repo, "secret", logger)

			token, err := service.Register(context.Background(), tt.login, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantToken {
				assert.NotEmpty(t, token)
			} else {
				assert.Empty(t, token)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	password := "strongpassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		login     string
		password  string
		repo      *mockUserRepo
		wantErr   bool
		wantToken bool
	}{
		{
			name:     "#1 success login",
			login:    "test",
			password: password,
			repo: &mockUserRepo{
				getUserByLoginFn: func(ctx context.Context, login string) (*postgres.User, error) {
					return &postgres.User{
						ID:           "user-id-1",
						PasswordHash: string(hash),
					}, nil
				},
			},
			wantErr:   false,
			wantToken: true,
		},
		{
			name:     "#2 wrong password",
			login:    "test",
			password: "wrongpassword",
			repo: &mockUserRepo{
				getUserByLoginFn: func(ctx context.Context, login string) (*postgres.User, error) {
					return &postgres.User{
						ID:           "user-id-1",
						PasswordHash: string(hash),
					}, nil
				},
			},
			wantErr:   true,
			wantToken: false,
		},
		{
			name:     "#3 repo error",
			login:    "test",
			password: password,
			repo: &mockUserRepo{
				getUserByLoginFn: func(ctx context.Context, login string) (*postgres.User, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr:   true,
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			service := NewAuthService(tt.repo, "secret", logger)

			token, err := service.Login(context.Background(), tt.login, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantToken {
				assert.NotEmpty(t, token)
			} else {
				assert.Empty(t, token)
			}
		})
	}
}
