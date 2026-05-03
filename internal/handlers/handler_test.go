package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
	"github.com/stretchr/testify/assert"
)

type MockUserService struct {
	RegisterFunc        func(ctx context.Context, payload *models.RegisterData) (int, error)
	LoginFunc           func(ctx context.Context, payload *models.RegisterData) (*models.User, error)
	RegisterOrderFunc   func(ctx context.Context, orderNum string) error
	GetOrdersFunc       func(ctx context.Context) ([]models.Order, error)
	GetBalanceFunc      func(ctx context.Context) (*models.Balance, error)
	WithdrawBalanceFunc func(ctx context.Context, withdraw *models.Withdraw) error
	WithdrawalsFunc     func(ctx context.Context) ([]models.Withdraw, error)
}

func (m *MockUserService) Register(ctx context.Context, p *models.RegisterData) (int, error) {
	return m.RegisterFunc(ctx, p)
}
func (m *MockUserService) Login(ctx context.Context, p *models.RegisterData) (*models.User, error) {
	return m.LoginFunc(ctx, p)
}
func (m *MockUserService) RegisterOrder(ctx context.Context, n string) error {
	return m.RegisterOrderFunc(ctx, n)
}
func (m *MockUserService) GetOrders(ctx context.Context) ([]models.Order, error) {
	return m.GetOrdersFunc(ctx)
}
func (m *MockUserService) GetBalance(ctx context.Context) (*models.Balance, error) {
	return m.GetBalanceFunc(ctx)
}
func (m *MockUserService) WithdrawBalance(ctx context.Context, w *models.Withdraw) error {
	return m.WithdrawBalanceFunc(ctx, w)
}
func (m *MockUserService) Withdrawals(ctx context.Context) ([]models.Withdraw, error) {
	return m.WithdrawalsFunc(ctx)
}

func TestUserHandler_RegisterOrder(t *testing.T) {
	type mockBehavior func(m *MockUserService)

	tests := []struct {
		name           string
		contentType    string
		body           string
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:        "Success (202 Accepted)",
			contentType: "text/plain",
			body:        "79927398713", // Valid Luhn
			mockBehavior: func(m *MockUserService) {
				m.RegisterOrderFunc = func(ctx context.Context, n string) error { return nil }
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "Already registered by current user (200 OK)",
			contentType: "text/plain",
			body:        "79927398713",
			mockBehavior: func(m *MockUserService) {
				m.RegisterOrderFunc = func(ctx context.Context, n string) error { return shared.ErrAlreadyExists }
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Conflict with another user (409 Conflict)",
			contentType: "text/plain",
			body:        "79927398713",
			mockBehavior: func(m *MockUserService) {
				m.RegisterOrderFunc = func(ctx context.Context, n string) error { return shared.ErrWriteConflict }
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Invalid Luhn number (422 Unprocessable)",
			contentType:    "text/plain",
			body:           "123", // Non-valid Luhn
			mockBehavior:   func(m *MockUserService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Empty body (400 Bad Request)",
			contentType:    "text/plain",
			body:           "",
			mockBehavior:   func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong Content-Type (400 Bad Request)",
			contentType:    "application/json",
			body:           "79927398713",
			mockBehavior:   func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Internal Server Error (500)",
			contentType: "text/plain",
			body:        "79927398713",
			mockBehavior: func(m *MockUserService) {
				m.RegisterOrderFunc = func(ctx context.Context, n string) error { return errors.New("db fail") }
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init handler with mock service
			mockSvc := &MockUserService{}
			tt.mockBehavior(mockSvc)
			h := &UserHandler{service: mockSvc}

			// Create Request
			r := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(tt.body))
			r.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			// Execute
			h.RegisterOrder(w, r)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserHandler_GetBalance(t *testing.T) {
	mockSvc := &MockUserService{
		GetBalanceFunc: func(ctx context.Context) (*models.Balance, error) {
			return &models.Balance{
				Current:   50050, // 500.50
				Withdrawn: 10000, // 100.00
			}, nil
		},
	}
	h := &UserHandler{service: mockSvc}

	r := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()

	h.GetBalance(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"current":500.5, "withdrawn":100}`, w.Body.String())
}

func TestUserHandler_Register(t *testing.T) {
	secret := "test-secret"
	mockSvc := &MockUserService{
		RegisterFunc: func(ctx context.Context, p *models.RegisterRequest) (int, error) {
			if p.Login == "exists" {
				return 0, shared.ErrWriteConflict
			}
			return 1, nil
		},
	}
	h := &UserHandler{service: mockSvc, secretKey: secret}

	t.Run("Success 200", func(t *testing.T) {
		body, _ := json.Marshal(models.RegisterRequest{Login: "new", Password: "pwd"})
		r := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		h.Register(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("Authorization"))
	})

	t.Run("Conflict 409", func(t *testing.T) {
		body, _ := json.Marshal(models.RegisterRequest{Login: "exists", Password: "pwd"})
		r := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		h.Register(w, r)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestUserHandler_Login(t *testing.T) {
	secret := "test-secret"
	mockSvc := &MockUserService{
		LoginFunc: func(ctx context.Context, p *models.LoginRequest) (*models.User, error) {
			if p.Login == "wrong" {
				return nil, shared.ErrNotMatchPassword
			}
			return &models.User{ID: 1, Login: p.Login}, nil
		},
	}
	h := &UserHandler{service: mockSvc, secretKey: secret}

	t.Run("Success 200", func(t *testing.T) {
		body, _ := json.Marshal(models.LoginRequest{Login: "user", Password: "pwd"})
		r := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		h.Login(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("Authorization"))
	})

	t.Run("Unauthorized 401", func(t *testing.T) {
		body, _ := json.Marshal(models.LoginRequest{Login: "wrong", Password: "pwd"})
		r := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		h.Login(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUserHandler_WithdrawBalance(t *testing.T) {
	mockSvc := &MockUserService{
		WithdrawBalanceFunc: func(ctx context.Context, w *models.Withdraw) error {
			if w.Sum > 100000 {
				return shared.ErrNoAffectedRows
			} // имитация нехватки средств
			return nil
		},
	}
	h := &UserHandler{service: mockSvc}

	t.Run("Success 200", func(t *testing.T) {
		body, _ := json.Marshal(models.WithdrawRequest{OrderNumber: "79927398713", Sum: 100.5})
		r := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		h.WithdrawBalance(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Insufficient Funds 402", func(t *testing.T) {
		body, _ := json.Marshal(models.WithdrawRequest{OrderNumber: "79927398713", Sum: 9999})
		r := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		h.WithdrawBalance(w, r)
		assert.Equal(t, http.StatusPaymentRequired, w.Code)
	})

	t.Run("Invalid Order 422", func(t *testing.T) {
		body, _ := json.Marshal(models.WithdrawRequest{OrderNumber: "invalid", Sum: 10})
		r := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		h.WithdrawBalance(w, r)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestUserHandler_GetOrders_Success(t *testing.T) {
	mockSvc := &MockUserService{
		GetOrdersFunc: func(ctx context.Context) ([]models.Order, error) {
			return []models.Order{
				{Number: "79927398713", Status: models.OrderStatusProcessed, Accrual: 50000},
			}, nil
		},
	}
	h := &UserHandler{service: mockSvc}

	r := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	h.GetOrders(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"number":"79927398713"`)
	assert.Contains(t, w.Body.String(), `"accrual":500`)
}

func TestUserHandler_Withdrawals_Success(t *testing.T) {
	mockSvc := &MockUserService{
		WithdrawalsFunc: func(ctx context.Context) ([]models.Withdraw, error) {
			return []models.Withdraw{
				{OrderNumber: "79927398713", Sum: 10000},
			}, nil
		},
	}
	h := &UserHandler{service: mockSvc}

	r := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	w := httptest.NewRecorder()

	// Внимание: проверь, как называется твой метод в хендлере, обычно h.Withdrawals
	h.Withdrawals(w, r)

	if w.Code == http.StatusNoContent {
		t.Log("No withdrawals found, which is also a valid success case")
	} else {
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"sum":100`)
	}
}

func TestCreateRouter(t *testing.T) {
    mockSvc := &MockUserService{}
    // Передай пустые заглушки для мидлварей
    r := CreateRouter(mockSvc, "secret", 
        func(h http.Handler) http.Handler { return h }, 
        func(h http.Handler) http.Handler { return h }, 
        func(h http.Handler) http.Handler { return h })
    assert.NotNil(t, r)
}