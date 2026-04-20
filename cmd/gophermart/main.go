package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/configs"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/database"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/handlers"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/middlewares"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/repositories"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/services"
)

func main() {
	err := run()

	if err != nil {
		slog.Error("failed to run app")
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	config := configs.GetConfig()

	db, err := database.DataBase(ctx, database.DbConfig{DSN: config.DatabaseSoruceName})

	if err := database.RunMigration(database.DbConfig{DSN: config.DatabaseSoruceName}); err != nil {
		slog.Error("faied to migrate db", "error", err)
		os.Exit(1)
	}

	if err != nil {
		slog.Error("failed to init database", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	userRepository := repositories.NewUserRepository(db)

	userService := services.NewUserService(userRepository)

	router := handlers.CreateRouter(userService, config.SecretKey,
		middlewares.AuthMiddleware(middlewares.SetSecretKey(config.SecretKey)),
		middlewares.LoggerMiddleware,
	)

	slog.Info("server starting", "addr", config.ServerAddress)

	if err := http.ListenAndServe(config.ServerAddress, router); err != nil {
		slog.Error("server failed")
		return err
	}

	return nil
}
