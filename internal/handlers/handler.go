package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/utils"
)

// * `POST /api/user/register` — регистрация пользователя;
// * `POST /api/user/login` — аутентификация пользователя;
// * `POST /api/user/orders` — загрузка пользователем номера заказа для расчёта;
// * `GET /api/user/orders` — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// * `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности пользователя;
// * `POST /api/user/balance/withdraw` — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
// * `GET /api/user/withdrawals` — получение информации о выводе средств с накопительного счёта пользователем.

type UserService interface {
	Register(ctx context.Context, payload *models.RegisterData) (int, error)
	Login(ctx context.Context, payload *models.RegisterData) (*models.User, error)
	RegisterOrder(context.Context, string) error
	// GetOrders() error
	// GetBalance() error
	// GetWithdrawals() error
}

// POST /api/user/orders

func CreateRouter(service UserService,
	secretKey string,
	authWM func(http.Handler) http.Handler,
	loggerWM func(http.Handler) http.Handler,
) http.Handler {
	router := chi.NewRouter()

	h := UserHandler{service: service, secretKey: secretKey}

	router.Use(loggerWM)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.Register)

		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(authWM)
			r.Post("/orders", h.RegisterOrder)
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
		slog.Error("failed to build jwt", "error", err)
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

		if errors.Is(err, shared.ErrNotMatchPassword) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := utils.BuildJwtToken(user.ID, h.secretKey)

	if err != nil {
		slog.Error("failed to build jwt", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) RegisterOrder(w http.ResponseWriter, r *http.Request) {
	if value := r.Header.Get("Content-Type"); value != "text/plain" {
		slog.Error("failed to register order", "content-type", value)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	body, errRead := io.ReadAll(r.Body)
	if errRead != nil || len(body) == 0 {
		slog.Error("empty or unreadable body", "error", errRead)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var orderNumber string = strings.TrimSpace(string(body))

	if !utils.IsLuhnValid(orderNumber) {
		// 422
		slog.Error("order number is not valid")
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	err := h.service.RegisterOrder(r.Context(), orderNumber)

	if err != nil {
		if errors.Is(err, shared.ErrAlreadyExists) {
			// 200
			slog.Error("order already registered by current user", "error", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, shared.ErrWriteConflict) {
			// 409
			slog.Error("order already registered by another user", "error", err)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		// 500
		slog.Error("failed to register ordet", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// 202
	w.WriteHeader(http.StatusAccepted)
}
