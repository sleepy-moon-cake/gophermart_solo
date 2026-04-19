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

func NewUserRepository(db *sql.DB) *UserRepository{
	return &UserRepository{db}
}

type UserRepository struct{
	db *sql.DB
}

func (r *UserRepository) Register(ctx context.Context,userCred *models.RegisterParams) (int,error) {
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
		return 0, fmt.Errorf("register user: %w", shared.ErrUserConflict)
	}

	return 0, fmt.Errorf("register user: %w", err)
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error){
	var user models.User

	err:=r.db.QueryRowContext(ctx,
		"SELECT id,login,password FROM users where login = $1",
		login).Scan(&user.ID, &user.Login, &user.Password)

	if err !=nil {
		if errors.Is(err, sql.ErrNoRows){
			return nil,fmt.Errorf("get password hash:%w", shared.ErrNotFound)
		}

		return nil,fmt.Errorf("get password hash:%w", err)
	}
	
	return &user, nil
}
