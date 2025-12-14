package service

import (
	"context"
	"errors"
	"go-musthave-diploma-tpl/internal/repository/postgres"

	"go.uber.org/zap"
)

type OrdersServicer interface {
	CreateOrder(ctx context.Context, userID, number string, logger *zap.Logger) error
	GetOrderByUser(ctx context.Context, userID string, logger *zap.Logger) ([]postgres.Order, error)
}

type OrdersService struct {
	orderRepo OrdersServicer
	logger    *zap.Logger
}

func NewOrdersService(logger *zap.Logger, orderRepo OrdersServicer) *OrdersService {
	return &OrdersService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (s *OrdersService) UploadOrder(ctx context.Context, userID, number string) error {
	if number == "" {
		return errors.New("order number required")
	}

	if err := s.orderRepo.CreateOrder(ctx, userID, number, s.logger); err != nil {
		if errors.Is(err, postgres.ErrOrderExists) {
			return postgres.ErrOrderExists
		}
		return err
	}

	return nil
}

func (s *OrdersService) ListOrders(ctx context.Context, userID string) ([]postgres.Order, error) {
	return s.orderRepo.GetOrderByUser(ctx, userID, s.logger)
}
