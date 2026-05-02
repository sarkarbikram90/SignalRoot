package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/signalroot/signalroot/internal/config"
)

// DB wraps the pgxpool for database operations.
type DB struct {
	Pool   *pgxpool.Pool
	Logger *zap.Logger
}

// New creates a new database connection pool.
func New(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*DB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.DatabaseMaxConnections)

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL",
		zap.String("host", poolCfg.ConnConfig.Host),
		zap.String("database", poolCfg.ConnConfig.Database),
		zap.Int32("max_conns", poolCfg.MaxConns),
	)

	return &DB{Pool: pool, Logger: logger}, nil
}

// Close shuts down the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// Health checks database connectivity.
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
