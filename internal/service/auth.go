package service

import (
	"context"
	"go-musthave-diploma-tpl/internal/repository/postgres"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string, logger *zap.Logger) (string, error)
	GetUserByLogin(ctx context.Context, login string, logger *zap.Logger) (*postgres.User, error)
}

type AuthService struct {
	userRepo UserRepository
	secret   string
	logger   *zap.Logger
}

func NewAuthService(userRepo UserRepository, secret string, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		secret:   secret,
		logger:   logger,
	}
}

func checkLogin(login, password string, logger *zap.Logger) (bool, error) {
	if login == "" || password == "" {
		logger.Error("login or password is empty")
		return false, nil
	}
	if login == password {
		logger.Error("No secure", zap.String("login", login))
		return false, nil
	}
	if len(password) < 8 {
		logger.Error("password is too short, need 8 symbol", zap.String("login", login))
	}
	return true, nil
}

func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	accept, err := checkLogin(login, password, s.logger)
	if err != nil {
		s.logger.Error("login or password is wrong", zap.String("login", login), zap.Error(err))
		return "", err
	}
	if accept {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password", zap.Error(err))
			return "", err
		}

		userID, err := s.userRepo.CreateUser(ctx, login, string(hash), s.logger)
		if err != nil {
			s.logger.Error("failed to create user", zap.Error(err))
			return "", err
		}

		token, err := GenerateToken(userID, s.secret)
		if err != nil {
			s.logger.Error("failed to generate token", zap.Error(err))
			return "", err
		}
		return token, nil
	}
	return "", nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	accept, err := checkLogin(login, password, s.logger)
	if err != nil {
		s.logger.Error("login or password is wrong", zap.String("login", login), zap.Error(err))
		return "", err
	}
	if accept {
		user, err := s.userRepo.GetUserByLogin(ctx, login, s.logger)
		if err != nil {
			s.logger.Error("failed to get user by login", zap.String("login", login), zap.Error(err))
			return "", err
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			s.logger.Error("password is wrong", zap.String("login", login), zap.Error(err))
			return "", err
		}
		token, err := GenerateToken(user.ID, s.secret)
		if err != nil {
			s.logger.Error("failed to generate token", zap.Error(err))
			return "", err
		}
		return token, nil

	}
	return "", nil
}
