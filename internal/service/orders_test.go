package service_test

import (
	"context"
	"errors"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Мок репозитория
type mockOrdersRepo struct {
	CreateFunc func(ctx context.Context, userID, number string, logger *zap.Logger) error
	GetFunc    func(ctx context.Context, userID string, logger *zap.Logger) ([]postgres.Order, error)
}

func (m *mockOrdersRepo) CreateOrder(ctx context.Context, userID, number string, logger *zap.Logger) error {
	return m.CreateFunc(ctx, userID, number, logger)
}

func (m *mockOrdersRepo) GetOrderByUser(ctx context.Context, userID string, logger *zap.Logger) ([]postgres.Order, error) {
	return m.GetFunc(ctx, userID, logger)
}

func TestOrdersService_UploadOrder(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			CreateFunc: func(ctx context.Context, userID, number string, logger *zap.Logger) error {
				return nil
			},
		}
		svc := service.NewOrdersService(logger, mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.NoError(t, err)
	})

	t.Run("empty_number", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{}
		svc := service.NewOrdersService(logger, mockRepo)

		err := svc.UploadOrder(ctx, "user1", "")
		require.Error(t, err)
		require.Equal(t, "order number required", err.Error())
	})

	t.Run("order_exists", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			CreateFunc: func(ctx context.Context, userID, number string, logger *zap.Logger) error {
				return postgres.ErrOrderExists
			},
		}
		svc := service.NewOrdersService(logger, mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.ErrorIs(t, err, postgres.ErrOrderExists)
	})

	t.Run("other_error", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			CreateFunc: func(ctx context.Context, userID, number string, logger *zap.Logger) error {
				return errors.New("some error")
			},
		}
		svc := service.NewOrdersService(logger, mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.Error(t, err)
		require.Equal(t, "some error", err.Error())
	})
}

func TestOrdersService_ListOrders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	t.Run("orders_exist", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetFunc: func(ctx context.Context, userID string, logger *zap.Logger) ([]postgres.Order, error) {
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
		svc := service.NewOrdersService(logger, mockRepo)

		orders, err := svc.ListOrders(ctx, "user1")
		require.NoError(t, err)
		require.Len(t, orders, 1)
		require.Equal(t, "12345678903", orders[0].Number)
	})

	t.Run("no_orders", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetFunc: func(ctx context.Context, userID string, logger *zap.Logger) ([]postgres.Order, error) {
				return []postgres.Order{}, nil
			},
		}
		svc := service.NewOrdersService(logger, mockRepo)

		orders, err := svc.ListOrders(ctx, "user1")
		require.NoError(t, err)
		require.Len(t, orders, 0)
	})

	t.Run("error_from_repo", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetFunc: func(ctx context.Context, userID string, logger *zap.Logger) ([]postgres.Order, error) {
				return nil, errors.New("db error")
			},
		}
		svc := service.NewOrdersService(logger, mockRepo)

		orders, err := svc.ListOrders(ctx, "user1")
		require.Error(t, err)
		require.Nil(t, orders)
	})
}
