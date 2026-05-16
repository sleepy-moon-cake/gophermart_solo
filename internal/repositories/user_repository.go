package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

type UserRepository struct {
	db *sql.DB
}

func (r *UserRepository) Register(ctx context.Context, userCred *models.RegisterParams) (int, error) {
	var lastInsertId int

	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id",
		userCred.Login,
		userCred.HashPassword,
	).Scan(&lastInsertId)

	if err == nil {
		return lastInsertId, nil
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return 0, fmt.Errorf("register user: %w", shared.ErrWriteConflict)
	}

	return 0, fmt.Errorf("register user: %w", err)
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowContext(ctx,
		"SELECT id,login,password FROM users where login = $1",
		login).Scan(&user.ID, &user.Login, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get password hash:%w", shared.ErrNotFound)
		}

		return nil, fmt.Errorf("get password hash:%w", err)
	}

	return &user, nil
}

func (r *UserRepository) RegisterOrder(ctx context.Context, userId int, orderNumber string) error {
	var ownerId int

	err := r.db.QueryRowContext(ctx,
		"INSERT INTO orders (owner_id, number) VALUES ($1, $2) RETURNING owner_id",
		userId,
		orderNumber,
	).Scan(&ownerId)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {

			selectErr := r.db.QueryRowContext(ctx,
				"SELECT owner_id FROM orders WHERE number = $1",
				orderNumber).Scan(&ownerId)

			if selectErr != nil {
				return fmt.Errorf("register order: select: %w", selectErr)
			}

			if ownerId == userId {
				return fmt.Errorf("register order: %w", shared.ErrAlreadyExists)
			}

			return fmt.Errorf("register order: %w", shared.ErrWriteConflict)
		}

		return fmt.Errorf("register order: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserOrders(ctx context.Context, userId int) ([]models.Order, error) {
	var orders = make([]models.Order, 0)

	rows, err := r.db.QueryContext(ctx,
		"SELECT number, status, accrual, uploaded_at FROM orders WHERE owner_id = $1 ORDER BY uploaded_at ASC",
		userId)

	if err != nil {
		return nil, fmt.Errorf("get user orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("get user orders scan: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get user orders rows has error: %w", err)
	}

	return orders, nil
}

func (r *UserRepository) GetUserBalance(ctx context.Context, userId int) (*models.Balance, error) {
	var balance models.Balance

	err := r.db.QueryRowContext(ctx,
		"SELECT current, withdrawn FROM balance WHERE owner_id = $1", userId,
	).Scan(&balance.Current, &balance.Withdrawn)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.Balance{Current: 0, Withdrawn: 0}, nil
		}

		return nil, err
	}

	return &balance, nil
}

func (r *UserRepository) WithdrawBalance(ctx context.Context, userID int, withdraw *models.Withdraw) error {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("withdrawBalance: %w", err)
	}
	defer tx.Rollback()

	resB, err := tx.ExecContext(ctx,
		"UPDATE balance SET current = current - $1, withdrawn = withdrawn + $1 WHERE owner_id = $2 AND current >= $1",
		withdraw.Sum, userID,
	)
	if err != nil {
		return fmt.Errorf("withdrawBalance update: %w", err)
	}

	if value, _ := resB.RowsAffected(); value == 0 {
		return fmt.Errorf("withdrawBalance: withdrawn: %w", shared.ErrNoAffectedRows)
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO withdrawals (order_number, sum, owner_id) VALUES ($1, $2, $3)",
		withdraw.OrderNumber, withdraw.Sum, userID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return fmt.Errorf("withdrawBalance insert conflict: %w", shared.ErrWriteConflict)
		}
		return fmt.Errorf("withdrawBalance insert: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("withdrawBalance commit: %w", err)
	}

	return nil
}

func (r *UserRepository) Withdrawals(ctx context.Context, userID int) ([]models.Withdraw, error) {
	query := `
		SELECT order_number, sum, processed_at 
		FROM withdrawals 
		WHERE owner_id = $1 
		ORDER BY processed_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("Withdrawals, select: %w", err)
	}
	defer rows.Close()

	records := make([]models.Withdraw, 0)

	for rows.Next() {
		var record models.Withdraw

		if err := rows.Scan(&record.OrderNumber, &record.Sum, &record.ProcessedAt); err != nil {
			return nil, fmt.Errorf("Withdrawals, scan: %w", err)
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Withdrawals, rows err: %w", err)
	}

	return records, nil
}
