package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var (
	ErrOrderExists         = errors.New("order already exists")
	ErrOrderUploadedByUser = errors.New("order already uploaded by this user")
)

type OrderRepository struct {
	db *sql.DB
}

type Order struct {
	ID         string              `db:"id"`
	Number     string              `db:"number"`
	UserID     string              `db:"user_id"`
	Status     string              `db:"status"`
	Accrual    decimal.NullDecimal `db:"accrual"`
	UploadedAt time.Time           `db:"uploaded_at"`
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, userID, number string, logger *zap.Logger) error {
	query := `INSERT INTO orders (user_id, number) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, query, userID, number)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			logger.Warn("order already exists", zap.String("number", number))
			return ErrOrderExists
		}
		logger.Error("failed to insert order", zap.Error(err))
		return err
	}
	logger.Info("order created", zap.String("number", number))
	return nil
}

func (r *OrderRepository) GetOrderByUser(ctx context.Context, userID string, logger *zap.Logger) ([]Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("failed to query order", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			logger.Error("failed to scan order", zap.Error(err))
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		logger.Error("rows iteration error", zap.Error(err))
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) GetOrdersForProcessing(ctx context.Context, limit int, logger *zap.Logger) ([]Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		ORDER BY uploaded_at
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		logger.Error("failed to get orders for processing", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		logger.Error("rows iteration error", zap.Error(err))
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *decimal.Decimal, logger *zap.Logger) error {
	var dbAccrual decimal.NullDecimal
	if accrual != nil {
		dbAccrual = decimal.NullDecimal{
			Decimal: *accrual,
			Valid:   true,
		}
	} else {
		dbAccrual = decimal.NullDecimal{
			Valid: false,
		}
	}

	query := `
		UPDATE orders
		SET status = $2, accrual = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, orderID, status, dbAccrual)
	if err != nil {
		logger.Error("failed to update order", zap.Error(err))
		return err
	}
	return nil
}
