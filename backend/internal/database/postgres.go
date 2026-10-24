package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Dimonchester/ttk-room-map/backend/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbName,
		os.Getenv("DB_SSLMODE"),
	)

	logger.Info(logger.CatDB, "подключение к PostgreSQL (%s:%s/%s)...", host, port, dbName)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(pingCtx, dsn)
	if err != nil {
		logger.Error(logger.CatDB, "ошибка создания пула соединений БД: %v", err)
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		logger.Error(logger.CatDB, "БД %s:%s не отвечает на ping: %v", host, port, err)
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info(logger.CatDB, "успешное подключение к PostgreSQL (%s:%s/%s)", host, port, dbName)
	return pool, nil
}
