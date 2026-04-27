package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/configs"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/database"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/handlers"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/middlewares"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/repositories"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/services"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/workers"
)

func main() {
	err := run()

	if err != nil {
		slog.Error("failed to run app")
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := configs.GetConfig()

	db, err := database.DataBase(ctx, database.DbConfig{DSN: config.DatabaseSoruceName})

	if err := database.RunMigration(database.DbConfig{DSN: config.DatabaseSoruceName}); err != nil {
		slog.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	if err != nil {
		slog.Error("failed to init database", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	userRepository := repositories.NewUserRepository(db)

	userService := services.NewUserService(userRepository)

	orderRepository := repositories.NewOrderRepository(db)
	accrualWorker := workers.CreateAccrualWorker(orderRepository, config.AccrualSystemAddress)

	// Start worker
	go func() {
		slog.Info("accrual worker started")
		accrualWorker.Run(ctx)
	}()

	router := handlers.CreateRouter(userService, config.SecretKey,
		middlewares.AuthMiddleware(middlewares.SetSecretKey(config.SecretKey)),
		middlewares.LoggerMiddleware,
	)

	srv := http.Server{
		Addr:    config.ServerAddress,
		Handler: router,
	}

	// Start Server
	go func() {
		slog.Info("server starting", "addr", config.ServerAddress)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
		}
	}()

	<-ctx.Done()

	ctxWithTime, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(ctxWithTime); err != nil {
		slog.Error("forced shutdown", "error", err)
	}

	slog.Info("Stop server")

	return nil
}
