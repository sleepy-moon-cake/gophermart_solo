package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

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