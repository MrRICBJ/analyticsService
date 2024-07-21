package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

type Config struct {
	DSN string `envconfig:"POSTGRES_DSN" required:"true"`
}

func New(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	dsn, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
