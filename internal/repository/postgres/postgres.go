package postgres

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewStorage(db *pgxpool.Pool) *Storage {
	return &Storage{
		db: db,
	}
}

func NewPool(ctx context.Context, cfg config.ServiceConfig) (*pgxpool.Pool, error) {
	dsn, err := cfg.Postgres.GetDSN()
	if err != nil {
		return nil, e.WrapIfErr("failed to get postgres dsn", err)
	}
	poolCfg, err := newPoolConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, e.WrapIfErr("failed to create new pool with config", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, e.WrapIfErr("failed to ping database", err)
	}

	slog.Info("Pool created")
	return pool, nil
}

func newPoolConfig(dsn string) (*pgxpool.Config, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, e.WrapIfErr("failed to parse pgxpool config", err)
	}

	poolCfg.MaxConns = int32(runtime.NumCPU() * 2)
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.MinConns = 2

	return poolCfg, nil
}

func (s *Storage) Close() {
	s.db.Close()
}
