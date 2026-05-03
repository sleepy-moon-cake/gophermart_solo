package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
)

type Repository interface {
	GetUnhandledOrders(ctx context.Context) ([]models.Order, error)
	UpdateStatusOrder(ctx context.Context, orderNumber string, accrual int, status string) error
}

type AccrualWorker struct {
	r           Repository
	accrualAddr string
	client      *http.Client
	mu          sync.RWMutex
	retryTime   time.Time
}

func CreateAccrualWorker(r Repository, addr string) *AccrualWorker {
	var worker = AccrualWorker{r: r, client: &http.Client{Timeout: time.Second * 10}, accrualAddr: addr}

	return &worker
}

func (w *AccrualWorker) сreateWorkers(ctx context.Context, source chan []models.Order) {
	channels := make(chan models.Order)

	workerNumbers := 10

	var wg sync.WaitGroup

	for range workerNumbers {
		wg.Add(1)
		go func() {
			for order := range channels {
				w.mu.RLock()
				retryTime := w.retryTime
				pause := time.Now().Before(retryTime)
				w.mu.RUnlock()

				if pause {
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Until(retryTime)):
					}
				}

				if err := w.handleOrder(ctx, order); err != nil {
					slog.Error("worker failed to handle order",
						"order", order.Number,
						"error", err,
					)
				}
			}
			wg.Done()
		}()
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case orders, ok := <-source:
			if !ok {
				break loop
			}

			for _, order := range orders {
				channels <- order
			}
		}
	}

	close(channels)

	wg.Wait()
}

func (w *AccrualWorker) Run(ctx context.Context) {
	channel := make(chan []models.Order)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.сreateWorkers(ctx, channel)
	}()

	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

loop:
	for {
		select {
		case <-ctx.Done():
			close(channel)
			wg.Wait()
			return
		case <-ticker.C:
			w.mu.RLock()
			pause := time.Now().Before(w.retryTime)
			w.mu.RUnlock()

			slog.Info("Check unhandled orders")

			if pause {
				continue loop
			}

			orders, err := w.r.GetUnhandledOrders(ctx)

			if err != nil {
				slog.Error("worker error", "error", err)
				continue loop
			}

			if len(orders) > 0 {
				select {
				case channel <- orders:
				case <-ctx.Done():
				}
			}

			slog.Info("Check unhandled orders", slog.Int("Orders length", len(orders)))
		}
	}

}

func (w *AccrualWorker) handleOrder(ctx context.Context, order models.Order) error {
	fullURL, err := url.JoinPath(w.accrualAddr, "/api/orders/", order.Number)

	if err != nil {
		slog.Error("handleOrder:JoinPath", "error", err)
		return fmt.Errorf("handleOrder:JoinPath: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)

	if err != nil {
		slog.Error("handleOrder:NewRequestWithContext", "error", err)
		return fmt.Errorf("handleOrder:NewRequestWithContext: %w", err)
	}

	response, err := w.client.Do(req)

	if err != nil {
		slog.Error("handleOrder: client.Do", "error", err)
		return fmt.Errorf("handleOrder: client.Do: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests {
		delay := response.Header.Get("Retry-After")
		timeDelay, err := strconv.Atoi(delay)

		if err != nil {
			slog.Error("handleOrder:read Retry-After", "error", err)
			return fmt.Errorf("handleOrder:read Retry-After: %w", err)
		}

		w.mu.Lock()
		w.retryTime = time.Now().Add(time.Duration(timeDelay) * time.Second)
		w.mu.Unlock()

		slog.Info("Lock request to accrual service", slog.Int("TIME", timeDelay))

		return nil
	}

	if response.StatusCode == http.StatusOK {
		var accrual models.AccrualRecord

		if err := json.NewDecoder(response.Body).Decode(&accrual); err != nil {
			slog.Error("handleOrder: decode", "error", err)
			return fmt.Errorf("handleOrder: decode: %w", err)
		}

		accrualInt := int(math.Round(accrual.Accrual * 100))

		if err := w.r.UpdateStatusOrder(ctx, accrual.Order, accrualInt, accrual.Status); err != nil {
			slog.Error("failed to update db", "error", err)
			return err
		}

		slog.Info("Order handled",
			slog.String("Number", order.Number),
			slog.String("Status", accrual.Status),
			slog.Int("accrualInt", accrualInt),
		)

		return nil
	}

	if response.StatusCode == http.StatusNoContent {
		if err := w.r.UpdateStatusOrder(ctx, order.Number, 0, models.OrderStatusInvalid); err != nil {
			slog.Error("failed to update db", "error", err)
			return err
		}

		slog.Info("Order handled",
			slog.String("Number", order.Number),
			slog.String("Status", models.OrderStatusInvalid),
		)
		return nil
	}

	return nil
}
