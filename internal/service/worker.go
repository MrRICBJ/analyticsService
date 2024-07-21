package service

import (
	"analitycsService/internal/model"
	"analitycsService/internal/repository"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	id     int
	db     *pgxpool.Pool
	jobs   <-chan model.Task
	repo   repository.Repo
	logger *logrus.Logger
	ctx    context.Context
}

func NewWorker(ctx context.Context, id int, db *pgxpool.Pool, tasks <-chan model.Task, repo repository.Repo,
	logger *logrus.Logger) Worker {
	return Worker{
		id:     id,
		db:     db,
		jobs:   tasks,
		repo:   repo,
		ctx:    ctx,
		logger: logger,
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case <-w.ctx.Done():
				w.logger.Info(fmt.Sprintf("worker shutting down worker %d", w.id))
				return
			case task := <-w.jobs:
				w.logger.Info(fmt.Sprintf("worker %d processing task.Id: %d", w.id, task.ID))

				if err := w.ProcessTask(task); err != nil {
					// вот тут мы можем не изменить processed на false
					if err = w.repo.UpdateProcessedData(w.ctx, w.db, task.ID, false); err != nil {
						w.logger.Error(fmt.Sprintf("failed to update processed data: %s", err.Error()))
						continue
					}
				}

				w.logger.Info(fmt.Sprintf("worker %d finished processing task.id: %d", w.id, task.ID))
			}
		}
	}()
}

func (w *Worker) ProcessTask(task model.Task) error {
	tx, err := w.db.Begin(w.ctx)
	if err != nil {
		w.logger.Error(fmt.Sprintf("worker %d; task.ID %d; failed to begin transaction", w.id, task.ID))
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(w.ctx)
			w.logger.Error(fmt.Sprintf("worker %d; task.ID %d; transaction rolled back due to error", w.id, task.ID))
			return
		}
		err = tx.Commit(w.ctx)
		if err != nil {
			w.logger.Error(fmt.Sprintf("worker %d; task.ID %d; failed to commit transaction", w.id, task.ID))
		}
	}()

	if err := w.repo.SaveData(w.ctx, tx, task.Time, task.UserID, task.Data); err != nil {
		w.logger.Error(fmt.Sprintf("failed to save data: %s", err.Error()))
		return err
	}

	if err := w.repo.DeleteTask(w.ctx, tx, task.ID); err != nil {
		w.logger.Error(fmt.Sprintf("failed to update processed data: %s", err.Error()))
		return err
	}
	return nil
}
