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
	if !isValidOrderNumber(number) {
		return postgres.ErrInvalidOrder
	}

	err := s.orderRepo.CreateOrder(ctx, userID, number, s.logger)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderExists) {
			userOrders, getErr := s.orderRepo.GetOrderByUser(ctx, userID, s.logger)
			if getErr == nil {
				for _, o := range userOrders {
					if o.Number == number {
						return nil
					}
				}
			}
			return postgres.ErrOrderExists
		}
		return err
	}

	return nil
}

func (s *OrdersService) ListOrders(ctx context.Context, userID string) ([]postgres.Order, error) {
	return s.orderRepo.GetOrderByUser(ctx, userID, s.logger)
}

func isValidOrderNumber(number string) bool {
	if len(number) < 10 || len(number) > 19 {
		return false
	}

	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}
