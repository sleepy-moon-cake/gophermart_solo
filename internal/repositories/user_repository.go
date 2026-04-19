package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
)

var ErrUserConflict = errors.New("user already exist")

var ErrNotFound = errors.New("user not found")

func NewUserRepository(db *sql.DB) *UserRepository{
	return &UserRepository{db}
}

type UserRepository struct{
	db *sql.DB
}

func (r *UserRepository) Register(ctx context.Context,userCred *models.RegisterParams) error {
	_,err:=r.db.ExecContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2)",
		userCred.Login, 
		userCred.HashPassword,
	)

	if err == nil {
		return nil
	}
	
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return fmt.Errorf("register user: %w",ErrUserConflict)
	}

	return  fmt.Errorf("register user: %w",err)
}

func (r *UserRepository) GetHashedPasswordByLogin(ctx context.Context, login string) (string, error){
	var hash string

	err:=r.db.QueryRowContext(ctx,
		"SELECT password FROM users where login = $1",
		login).Scan(&hash)

	if err !=nil {
		if errors.Is(err, sql.ErrNoRows){
			return "",fmt.Errorf("get password hash:%w", ErrNotFound)
		}

		return "",fmt.Errorf("get password hash:%w", err)
	}
	
	return hash, nil
}
