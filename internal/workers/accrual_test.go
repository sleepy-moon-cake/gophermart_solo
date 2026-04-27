package workers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/stretchr/testify/assert"
)

// Простой мок репозитория
type mockRepo struct {
	updateCalled bool
}

func (m *mockRepo) GetUnhandledOrders(ctx context.Context) ([]models.Order, error) {
	return nil, nil
}

func (m *mockRepo) UpdateStatusOrder(ctx context.Context, number string, accrual int, status string) error {
	m.updateCalled = true
	return nil
}

func TestHandleOrder_Simple(t *testing.T) {
	// 1. Создаем фейковый сервер, который имитирует систему начислений
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.AccrualRecord{
			Order:   "79927398713",
			Status:  "PROCESSED",
			Accrual: 500.5,
		})
	}))
	defer srv.Close()

	// 2. Инициализируем воркер с нашим мок-репозиторием
	repo := &mockRepo{}
	worker := CreateAccrualWorker(repo, srv.URL)

	// 3. Тестируем обработку
	err := worker.handleOrder(context.Background(), models.Order{Number: "79927398713"})

	// 4. Проверки
	assert.NoError(t, err)
	assert.True(t, repo.updateCalled, "Репозиторий должен был быть вызван")
}

func TestHandleOrder_429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "10")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	repo := &mockRepo{}
	worker := CreateAccrualWorker(repo, srv.URL)

	err := worker.handleOrder(context.Background(), models.Order{Number: "123"})

	assert.NoError(t, err)
	worker.mu.RLock()
	assert.True(t, worker.retryTime.After(time.Now()), "retryTime должен обновиться")
	worker.mu.RUnlock()
}
