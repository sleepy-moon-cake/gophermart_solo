package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/utils"
)

type UserService interface {
	Register(ctx context.Context, payload *models.RegisterData) (int, error)
	Login(ctx context.Context, payload *models.RegisterData) (*models.User, error)
	RegisterOrder(context.Context, string) error
	GetOrders(context.Context) ([]models.Order, error)
	GetBalance(context.Context) (*models.Balance, error)
	WithdrawBalance(context.Context, *models.Withdraw) error
	Withdrawals(context.Context) ([]models.Withdraw, error)
}

func CreateRouter(service UserService,
	secretKey string,
	authWM func(http.Handler) http.Handler,
	loggerWM func(http.Handler) http.Handler,
	compressWM func(http.Handler) http.Handler,
) http.Handler {
	router := chi.NewRouter()

	h := UserHandler{service: service, secretKey: secretKey}

	router.Use(compressWM)

	router.Use(loggerWM)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.Register)

		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(authWM)
			r.Post("/orders", h.RegisterOrder)
			r.Get("/orders", h.GetOrders)
			r.Get("/balance", h.GetBalance)
			r.Post("/balance/withdraw", h.WithdrawBalance)
		})
	})

	return router
}

type UserHandler struct {
	service   UserService
	secretKey string
	expiredAt time.Duration
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var payload models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("register: decode", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if payload.Login == "" || payload.Password == "" {
		slog.Error("register: empty payload params")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := h.service.Register(r.Context(), &payload)

	if err != nil {
		slog.Error("register:", "error", err)

		if errors.Is(err, shared.ErrWriteConflict) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := utils.BuildJwtToken(userID, h.secretKey)

	if err != nil {
		slog.Error("register: failed to build jwt", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("login: decode", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(r.Context(), &payload)

	if err != nil {
		slog.Error("login:", "error", err)

		if errors.Is(err, shared.ErrNotMatchPassword) || errors.Is(err, shared.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := utils.BuildJwtToken(user.ID, h.secretKey)

	if err != nil {
		slog.Error("login: failed to build jwt", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) RegisterOrder(w http.ResponseWriter, r *http.Request) {
	if value := r.Header.Get("Content-Type"); value != "text/plain" {
		slog.Error("registerOrder: failed to register order", "content-type", value)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, errRead := io.ReadAll(r.Body)
	if errRead != nil || len(body) == 0 {
		slog.Error("registerOrder: empty or unreadable body", "error", errRead)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var orderNumber string = strings.TrimSpace(string(body))

	if !utils.IsLuhnValid(orderNumber) {
		// 422
		slog.Error("registerOrder: order number is not valid")
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	err := h.service.RegisterOrder(r.Context(), orderNumber)

	if err != nil {
		if errors.Is(err, shared.ErrAlreadyExists) {
			// 200
			slog.Error("registerOrder: order already registered by current user", "error", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, shared.ErrWriteConflict) {
			// 409
			slog.Error("registerOrder: order already registered by another user", "error", err)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		// 500
		slog.Error("registerOrder: failed to register order", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// 202
	w.WriteHeader(http.StatusAccepted)
}

func (h *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.GetOrders(r.Context())

	if err != nil {
		slog.Error("getOrders: failed to get orders", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		slog.Info("get orders: empty")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var body = make([]models.OrderResponse, 0, len(orders))

	for _, v := range orders {
		res := models.OrderResponse{
			Number:     v.Number,
			Status:     v.Status,
			UploadedAt: v.UploadedAt,
		}

		if v.Status == models.OrderStatusProcessed {
			accrualFloat64 := float64(v.Accrual) / 100
			res.Accrual = &accrualFloat64
		}

		body = append(body, res)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("getOrders: failed to Encode orders", "error", err)
	}
}

func (h *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := h.service.GetBalance(r.Context())

	if err != nil {
		slog.Error("getBalance: failed to get balance", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	balancefloat := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   float64(balance.Current) / 100,
		Withdrawn: float64(balance.Withdrawn) / 100,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(balancefloat); err != nil {
		slog.Error("getBalance: failed to encode balance", "error", err)
		return
	}
}

func (h *UserHandler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	var request models.WithdrawRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// 400
		slog.Error("withdrawBalance: failed to decode")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if request.Sum <= 0 {
		// 400
		slog.Error("withdrawBalance: invalid sum")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !utils.IsLuhnValid(request.OrderNumber) {
		// 422
		slog.Error("withdrawBalance: order number is not valid")
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	intSum := int(math.Round(float64(request.Sum) * 100))

	err := h.service.WithdrawBalance(r.Context(), &models.Withdraw{
		OrderNumber: request.OrderNumber,
		Sum:         intSum,
	})

	if err != nil {
		if errors.Is(err, shared.ErrNoAffectedRows) {
			// 402
			slog.Error("withdrawBalance", "error", err)
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
			return
		}

		if errors.Is(err, shared.ErrWriteConflict) {
			// 409
			slog.Error("withdrawBalance", "error", err)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		// 500
		slog.Error("withdrawBalance", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	withdraws, err := h.service.Withdrawals(r.Context())

	if err != nil {
		slog.Error("Withdrawals", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdraws) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	body := make([]models.WithdrawWithCent, 0, len(withdraws))

	for _, v := range withdraws {
		body = append(body, models.WithdrawWithCent{
			OrderNumber: v.OrderNumber,
			Sum:         float64(v.Sum) / 100,
			ProcessedAt: v.ProcessedAt,
		})
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("Withdrawals, encode", "error", err)
	}
}

