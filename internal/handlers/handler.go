package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/models"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
)

// * `POST /api/user/register` — регистрация пользователя;
// * `POST /api/user/login` — аутентификация пользователя;
// * `POST /api/user/orders` — загрузка пользователем номера заказа для расчёта;
// * `GET /api/user/orders` — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// * `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности пользователя;
// * `POST /api/user/balance/withdraw` — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
// * `GET /api/user/withdrawals` — получение информации о выводе средств с накопительного счёта пользователем.

type UserService interface {
	Register(ctx context.Context, payload *models.RegisterData) error
	Login(ctx context.Context, payload *models.RegisterData) error
	// GetOrders() error
	// GetBalance() error
	// GetWithdrawals() error
}

func CreateRouter (service UserService, 
	authWM func(http.Handler) http.Handler,
	loggerWM func(http.Handler) http.Handler,
	) http.Handler{
	router :=chi.NewRouter()
	
	h:=UserHandler{service: service}

	router.Use(loggerWM)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.Register)

		r.Post("/login",h.Login)
		
		r.Group(func(r chi.Router) {
			r.Use(authWM)
		})
	})

	return router
}

type UserHandler struct {
	service UserService
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request){
	var payload models.RegisterRequest

	if err:=json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("register: decode","error",err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return;
	}

	if err:= h.service.Register(r.Context(), &payload); err !=nil {
		slog.Error("register:","error",err)

		if errors.Is(err, shared.ErrUserConflict) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return;
	}

	w.WriteHeader(http.StatusOK)
} 

func (h *UserHandler) Login (w http.ResponseWriter, r *http.Request) {
	var payload models.LoginRequest

	if err:=json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("login: decode","error",err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return;
	}

	if err:= h.service.Login(r.Context(),&payload); err !=nil {
		slog.Error("login:","error",err)

		if errors.Is(err, shared.ErrNotMatchPassword){
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return;
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return;
	}
	
	w.WriteHeader(http.StatusOK)
}
