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
		"SELECT number, status, uploaded_at FROM orders WHERE owner_id = $1 ORDER BY uploaded_at ASC",
		userId)

	if err != nil {
		return nil, fmt.Errorf("get user orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.Number, &order.Status, &order.UploadedAt)
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
