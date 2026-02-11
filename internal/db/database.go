// internal/db/database.go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jd7008911/aogeri-api/internal/config"
)

type Database struct {
	Pool    *pgxpool.Pool
	Queries *Queries
}

func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = 25
	poolCfg.MinConns = 5
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	queries := New(pool)

	return &Database{
		Pool:    pool,
		Queries: queries,
	}, nil
}

func (db *Database) Close() {
	db.Pool.Close()
}
