package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() (*pgxpool.Pool, error) {

	connection := os.Getenv("DATABASE_URL")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("Ошибка создания пула: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Ошибка связи с БД: %w", err)
	}

	return pool, nil
}
