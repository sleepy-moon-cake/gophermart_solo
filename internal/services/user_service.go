package services

import (
	"context"
	"fmt"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"golang.org/x/crypto/bcrypt"
)


type Repository interface {
	Register(context.Context, *models.RegisterParams) error
	GetHashedPasswordByLogin(context.Context, string) (string,error)
}

type UserService struct {
	repository Repository
}

func NewUserService(repository Repository) *UserService{
	return &UserService{ repository}
}

func (s *UserService) Register(ctx context.Context ,payload *models.RegisterData) error{
	hash,err:=bcrypt.GenerateFromPassword([]byte(payload.Password),bcrypt.DefaultCost)

	if err != nil {
		return fmt.Errorf("register: %w",err)
	}

	userAuthInfo:=&models.RegisterParams{Login: payload.Login, HashPassword: string(hash)}

	return  s.repository.Register(ctx,userAuthInfo)
}

func (s *UserService) Login(ctx context.Context ,payload *models.RegisterData) error{
	hashPassword,err:= s.repository.GetHashedPasswordByLogin(ctx,payload.Login)

	if err != nil {
		return fmt.Errorf("login:hash: %w",err)
	}

	if err:= bcrypt.CompareHashAndPassword([]byte(hashPassword),[]byte(payload.Password)); err !=nil {
		return fmt.Errorf("login:compare: %w",err)
	}

	return nil
}

// type UserService interface {
// 	Register(ctx context.Context, payload *models.RegisterData) error
// 	Login(ctx context.Context, payload *models.RegisterData) error
// 	GetOrders() error
// 	GetBalance() error
// 	GetWithdrawals() error
// }
