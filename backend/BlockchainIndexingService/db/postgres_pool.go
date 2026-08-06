package db

import (
	"BlockchainIndexingService/utils"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := utils.GetEnv("DB_CONNECTION_STR", "postgres://admin:admin@localhost:5433/database?sslmode=disable")

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %v", err)
	}

	config.MaxConns = 25
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Postgres: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging Postgres: %v", err)
	}

	return pool, nil
}
