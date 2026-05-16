package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
)

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db}
}

type OrderRepository struct {
	db *sql.DB
}

func (r *OrderRepository) GetUnhandledOrders(ctx context.Context) ([]models.Order, error) {
	limit := 50

	row, err := r.db.QueryContext(ctx,
		`SELECT number, status, owner_id, uploaded_at 
		FROM orders 
		WHERE status IN ($1, $2)
		LIMIT $3`,
		models.OrderStatusNew, models.OrderStatusProcessing, limit)

	if err != nil {
		return nil, fmt.Errorf("GetUnhandledOrders: query: %w", err)
	}

	defer row.Close()

	var records = make([]models.Order, 0, limit)

	for row.Next() {

		var record models.Order

		if err := row.Scan(&record.Number, &record.Status, &record.OwnerID, &record.UploadedAt); err != nil {
			return nil, fmt.Errorf("GetUnhandledOrders: scan: %w", err)
		}

		records = append(records, record)
	}

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("GetUnhandledOrders: row.err: %w", err)
	}

	return records, nil
}

func (r *OrderRepository) UpdateStatusOrder(ctx context.Context, orderNumber string, accrual int, status string) error {
	if status != models.OrderStatusProcessed {
		_, err := r.db.ExecContext(
			ctx,
			"UPDATE orders SET status = $1 WHERE number = $2",
			status, orderNumber,
		)
		if err != nil {
			return fmt.Errorf("UpdateStatusOrder: %w", err)
		}
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("UpdateProcessedOrder:BeginTx: %w", err)
	}

	defer tx.Rollback()

	var userID int

	if err := tx.QueryRowContext(ctx, "SELECT owner_id FROM orders WHERE number = $1", orderNumber).Scan(&userID); err != nil {
		return fmt.Errorf("UpdateProcessedOrder: select user: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		models.OrderStatusProcessed, accrual, orderNumber,
	)

	if err != nil {
		return fmt.Errorf("UpdateProcessedOrder:update status: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO balance (owner_id, current) VALUES ($1, $2) 
		ON CONFLICT (owner_id)
		DO UPDATE SET current = balance.current + $2`, userID, accrual)

	if err != nil {
		return fmt.Errorf("UpdateProcessedOrder:update balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
