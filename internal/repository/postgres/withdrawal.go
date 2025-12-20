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
	ErrNotEnoughFunds = errors.New("not enough funds")
	ErrInvalidOrder   = errors.New("invalid order number")
)

type Withdrawal struct {
	ID          string
	UserID      string
	OrderNumber string
	Sum         decimal.Decimal
	ProcessedAt time.Time
}

type WithdrawalRepository struct {
	db *sql.DB
}

func NewWithdrawalRepository(db *sql.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

func (r *WithdrawalRepository) Create(
	ctx context.Context,
	userID, orderNumber string,
	sum decimal.Decimal,
	logger *zap.Logger,
) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, userID, orderNumber, sum)
	if err != nil {
		if strings.Contains(err.Error(), "withdrawals_user_id_order_number_key") {
			logger.Warn("withdrawal order already exists", zap.String("order", orderNumber))
			return ErrInvalidOrder
		}
		logger.Error("failed to create withdrawal", zap.Error(err))
		return err
	}

	logger.Info("withdrawal created",
		zap.String("order", orderNumber),
		zap.String("sum", sum.String()),
	)
	return nil
}

func (r *WithdrawalRepository) ListByUser(
	ctx context.Context,
	userID string,
	logger *zap.Logger,
) ([]Withdrawal, error) {
	query := `
		SELECT id, user_id, order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("failed to query withdrawals", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var res []Withdrawal
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.OrderNumber,
			&w.Sum,
			&w.ProcessedAt,
		); err != nil {
			logger.Error("failed to scan withdrawal", zap.Error(err))
			return nil, err
		}
		res = append(res, w)
	}
	if err := rows.Err(); err != nil {
		logger.Error("rows iteration error", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (r *WithdrawalRepository) GetTotals(
	ctx context.Context,
	userID string,
	logger *zap.Logger,
) (accrued, withdrawn decimal.Decimal, err error) {
	query := `
		SELECT
			COALESCE(SUM(o.accrual), 0),
			COALESCE((
				SELECT SUM(w.sum)
				FROM withdrawals w
				WHERE w.user_id = $1
			), 0)
		FROM orders o
		WHERE o.user_id = $1
		  AND o.status = 'PROCESSED'
	`

	err = r.db.QueryRowContext(ctx, query, userID).Scan(&accrued, &withdrawn)
	if err != nil {
		logger.Error("failed to calculate totals", zap.Error(err))
		return decimal.Zero, decimal.Zero, err
	}

	return accrued, withdrawn, nil
}
