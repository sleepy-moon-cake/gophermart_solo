package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
	"golang.org/x/crypto/bcrypt"
)


type Repository interface {
	Register(context.Context, *models.RegisterParams) (int,error)
	GetUserByLogin(context.Context, string) (*models.User,error)
}

type UserService struct {
	repository Repository
}

func NewUserService(repository Repository) *UserService{
	return &UserService{ repository}
}

func (s *UserService) Register(ctx context.Context ,payload *models.RegisterData) (int,error){
	hash,err:=bcrypt.GenerateFromPassword([]byte(payload.Password),bcrypt.DefaultCost)

	if err != nil {
		return 0,fmt.Errorf("register: %w",err)
	}

	userAuthInfo:=&models.RegisterParams{Login: payload.Login, HashPassword: string(hash)}

	userID,err:= s.repository.Register(ctx, userAuthInfo)
	if err != nil {
		return 0, fmt.Errorf("register: %w",err)
	}

	return  userID,nil
}

func (s *UserService) Login(ctx context.Context ,payload *models.RegisterData) (*models.User,error){
	user,err:= s.repository.GetUserByLogin(ctx,payload.Login)

	if err != nil {
		return nil,fmt.Errorf("login:hash: %w",err)
	}

	if err:= bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(payload.Password)); err !=nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil,fmt.Errorf("login:compare: %w",shared.ErrNotMatchPassword)
		}

		return nil,fmt.Errorf("login:compare: %w",err)
	}

	return user, nil
}

// type UserService interface {
// 	Register(ctx context.Context, payload *models.RegisterData) error
// 	Login(ctx context.Context, payload *models.RegisterData) error
// 	GetOrders() error
// 	GetBalance() error
// 	GetWithdrawals() error
// }
