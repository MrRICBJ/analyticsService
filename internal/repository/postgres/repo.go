package postgres

import (
	"analitycsService/internal/model"
	"analitycsService/internal/repository"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type repo struct{}

func NewRepo() repository.Repo {
	return &repo{}
}

const saveData = `
	insert into analytics (
	    	time,
            user_id,
            data
    )
	values ($1, $2, $3)`

func (r *repo) SaveData(ctx context.Context, tx pgx.Tx, time time.Time, userID string, data []byte) error {
	_, err := tx.Exec(ctx, saveData, time, userID, data)
	if err != nil {
		return err
	}
	return nil
}

const saveRowData = `
	insert into raw_analytics (
	    	time,
            user_id,
            data
    )
	values ($1, $2, $3)`

func (r *repo) SaveRowData(ctx context.Context, db *pgxpool.Pool, time time.Time, userID string, data []byte) error {
	_, err := db.Exec(ctx, saveRowData, time, userID, data)
	if err != nil {
		return err
	}
	return nil
}

const updateProcessedData = `
	update raw_analytics 
	set processed = $1 
	where id = $2
`

func (r *repo) UpdateProcessedData(ctx context.Context, db *pgxpool.Pool, taskID int, processed bool) error {
	_, err := db.Exec(ctx, updateProcessedData, processed, taskID)
	if err != nil {
		return err
	}
	return nil
}

const selectAndMarkProcessed = `
	with selected as (
		select id, time, user_id, data
		from raw_analytics 
		where processed = false
		limit 10
	)
	update raw_analytics
	set processed = true
	from selected
	where raw_analytics.id = selected.id
	returning selected.id, selected.time, selected.user_id, selected.data
`

func (r *repo) ProcessRawData(ctx context.Context, db *pgxpool.Pool) ([]model.Task, error) {
	rows, err := db.Query(ctx, selectAndMarkProcessed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.ID, &task.Time, &task.UserID, &task.Data); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

const deleteRawData = `
	delete from raw_analytics
	where id = $1
`

func (r *repo) DeleteTask(ctx context.Context, tx pgx.Tx, id int) error {
	_, err := tx.Exec(ctx, deleteRawData, id)
	if err != nil {
		return err
	}

	return nil
}
