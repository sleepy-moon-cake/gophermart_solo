package services

import (
	"context"
	"testing"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type MockRepository struct {
	RegisterFunc        func(ctx context.Context, p *models.RegisterParams) (int, error)
	GetUserByLoginFunc  func(ctx context.Context, login string) (*models.User, error)
	RegisterOrderFunc   func(ctx context.Context, userID int, orderNum string) error
	GetUserOrdersFunc   func(ctx context.Context, userID int) ([]models.Order, error)
	GetUserBalanceFunc  func(ctx context.Context, userID int) (*models.Balance, error)
	WithdrawBalanceFunc func(ctx context.Context, userID int, w *models.Withdraw) error
	WithdrawalsFunc     func(ctx context.Context, userID int) ([]models.Withdraw, error)
}

func (m *MockRepository) Register(ctx context.Context, p *models.RegisterParams) (int, error) {
	return m.RegisterFunc(ctx, p)
}
func (m *MockRepository) GetUserByLogin(ctx context.Context, l string) (*models.User, error) {
	return m.GetUserByLoginFunc(ctx, l)
}
func (m *MockRepository) RegisterOrder(ctx context.Context, id int, n string) error {
	return m.RegisterOrderFunc(ctx, id, n)
}
func (m *MockRepository) GetUserOrders(ctx context.Context, id int) ([]models.Order, error) {
	return m.GetUserOrdersFunc(ctx, id)
}
func (m *MockRepository) GetUserBalance(ctx context.Context, id int) (*models.Balance, error) {
	return m.GetUserBalanceFunc(ctx, id)
}
func (m *MockRepository) WithdrawBalance(ctx context.Context, id int, w *models.Withdraw) error {
	return m.WithdrawBalanceFunc(ctx, id, w)
}
func (m *MockRepository) Withdrawals(ctx context.Context, id int) ([]models.Withdraw, error) {
	return m.WithdrawalsFunc(ctx, id)
}

func TestUserService_Register(t *testing.T) {
	mockRepo := &MockRepository{
		RegisterFunc: func(ctx context.Context, p *models.RegisterParams) (int, error) {
			err := bcrypt.CompareHashAndPassword([]byte(p.HashPassword), []byte("password123"))
			if err != nil {
				return 0, err
			}
			return 1, nil
		},
	}
	svc := NewUserService(mockRepo)

	t.Run("Success registration", func(t *testing.T) {
		id, err := svc.Register(context.Background(), &models.RegisterData{Login: "test", Password: "password123"})
		assert.NoError(t, err)
		assert.Equal(t, 1, id)
	})
}

func TestUserService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_pass"), bcrypt.DefaultCost)

	mockRepo := &MockRepository{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			if login == "found" {
				return &models.User{ID: 1, Login: "found", Password: string(hashedPassword)}, nil
			}
			return nil, shared.ErrNotFound
		},
	}
	svc := NewUserService(mockRepo)

	t.Run("Successful Login", func(t *testing.T) {
		user, err := svc.Login(context.Background(), &models.RegisterData{Login: "found", Password: "correct_pass"})
		assert.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		_, err := svc.Login(context.Background(), &models.RegisterData{Login: "found", Password: "wrong_pass"})
		assert.ErrorIs(t, err, shared.ErrNotMatchPassword)
	})
}

func TestUserService_RegisterOrder(t *testing.T) {
	mockRepo := &MockRepository{
		RegisterOrderFunc: func(ctx context.Context, userID int, orderNum string) error {
			return nil
		},
	}
	svc := NewUserService(mockRepo)

	t.Run("Unauthorized - No UserID in context", func(t *testing.T) {
		err := svc.RegisterOrder(context.Background(), "12345")
		assert.ErrorIs(t, err, shared.ErrUnauthorized)
	})

	t.Run("Success with UserID in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), shared.UserID, 1)
		err := svc.RegisterOrder(ctx, "12345")
		assert.NoError(t, err)
	})
}

func TestUserService_GetBalance(t *testing.T) {
	expectedBalance := &models.Balance{Current: 100, Withdrawn: 50}
	mockRepo := &MockRepository{
		GetUserBalanceFunc: func(ctx context.Context, userID int) (*models.Balance, error) {
			return expectedBalance, nil
		},
	}
	svc := NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), shared.UserID, 1)
	balance, err := svc.GetBalance(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
}

func TestUserService_GetOrders_Success(t *testing.T) {
	mockRepo := &MockRepository{
		GetUserOrdersFunc: func(ctx context.Context, userID int) ([]models.Order, error) {
			return []models.Order{{Number: "123"}}, nil
		},
	}
	svc := NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), shared.UserID, 1)
	orders, err := svc.GetOrders(ctx)

	assert.NoError(t, err)
	assert.Len(t, orders, 1)
}

func TestUserService_Withdrawals_Success(t *testing.T) {
	mockRepo := &MockRepository{
		WithdrawalsFunc: func(ctx context.Context, userID int) ([]models.Withdraw, error) {
			return []models.Withdraw{{OrderNumber: "123", Sum: 100}}, nil
		},
	}
	svc := NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), shared.UserID, 1)
	res, err := svc.Withdrawals(ctx)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestUserService_WithdrawBalance_Success(t *testing.T) {
	mockRepo := &MockRepository{
		WithdrawBalanceFunc: func(ctx context.Context, userID int, w *models.Withdraw) error {
			return nil
		},
	}
	svc := NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), shared.UserID, 1)
	err := svc.WithdrawBalance(ctx, &models.Withdraw{OrderNumber: "79927398713", Sum: 100})

	assert.NoError(t, err)
}
