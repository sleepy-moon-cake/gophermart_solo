package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DbConfig struct{
	DSN string
}


func DataBase(ctx context.Context,config DbConfig) (*sql.DB,error) {
	db,err:=sql.Open("pgx", config.DSN)

	if err !=nil {
		return nil, fmt.Errorf("open db: %w",err)
	}

	pingCtx, cancel:= context.WithTimeout(ctx, 5 * time.Second);
	defer cancel()

	if err:=db.PingContext(pingCtx); err !=nil {
		return nil, fmt.Errorf("configration ping error: %w", err)
	}

	slog.Info("Database connected")

	return  db, nil
}

func RunMigration(config DbConfig) error{
	slog.Info("Run migration")

	m,err:=migrate.New("file://migrations",config.DSN)

	if err !=nil {
		return fmt.Errorf("run migration:new: %w",err)
	}

	if err:=m.Up(); err !=nil{
		if errors.Is(err, migrate.ErrNoChange) {
			return nil 
		}

		return fmt.Errorf("run migration:up: %w",err)
	}

	slog.Info("Migration success")

	return nil
}