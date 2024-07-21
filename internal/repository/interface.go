package repository

import (
	"analitycsService/internal/model"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Repo interface {
	SaveRowData(ctx context.Context, db *pgxpool.Pool, time time.Time, userID string, data []byte) error
	UpdateProcessedData(ctx context.Context, db *pgxpool.Pool, userID int, processed bool) error
	ProcessRawData(ctx context.Context, db *pgxpool.Pool) ([]model.Task, error)
	SaveData(ctx context.Context, tx pgx.Tx, time time.Time, userID string, data []byte) error
	DeleteTask(ctx context.Context, tx pgx.Tx, id int) error
}
