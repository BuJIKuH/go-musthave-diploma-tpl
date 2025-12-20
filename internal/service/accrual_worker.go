package service

import (
	"context"
	"errors"
	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"time"

	"go.uber.org/zap"
)

type AccrualWorker struct {
	OrderRepo *postgres.OrderRepository
	Client    *accrual.Client
	Logger    *zap.Logger
}

func NewAccrualWorker(orderRepo *postgres.OrderRepository, client *accrual.Client, logger *zap.Logger) *AccrualWorker {
	return &AccrualWorker{
		OrderRepo: orderRepo,
		Client:    client,
		Logger:    logger,
	}
}

func NewAccrualClient(cfg *config.Config) *accrual.Client {
	return accrual.NewClient(cfg.AccrualSystemAddress)
}

func (w *AccrualWorker) Process(_ context.Context) {
	dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	orders, err := w.OrderRepo.GetOrdersForProcessing(dbCtx, 10, w.Logger)
	if err != nil {
		w.Logger.Error("failed to get orders for processing", zap.Error(err))
		return
	}

	for _, order := range orders {
		resp, err := w.Client.GetOrder(dbCtx, order.Number)
		if err != nil {
			if errors.Is(err, accrual.ErrTooManyRequests) {
				w.Logger.Warn("accrual rate limit")
				return
			}
			continue
		}

		switch resp.Status {
		case accrual.StatusInvalid:
			_ = w.OrderRepo.UpdateOrderStatus(dbCtx, order.ID, "INVALID", nil, w.Logger)
		case accrual.StatusProcessing:
			_ = w.OrderRepo.UpdateOrderStatus(dbCtx, order.ID, "PROCESSING", nil, w.Logger)
		case accrual.StatusProcessed:
			dec := resp.Accrual
			_ = w.OrderRepo.UpdateOrderStatus(dbCtx, order.ID, "PROCESSED", &dec, w.Logger)
		}
	}
}
