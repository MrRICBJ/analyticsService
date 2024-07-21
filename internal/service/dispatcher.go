package service

import (
	"analitycsService/internal/model"
	"analitycsService/internal/repository"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"time"
)

type Dispatcher struct {
	workers  []Worker
	db       *pgxpool.Pool
	queue    chan model.Task
	repo     repository.Repo
	logger   *logrus.Logger
	interval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

func NewDispatcher(ctx context.Context, numWorkers int, db *pgxpool.Pool, repo repository.Repo, logger *logrus.Logger,
	internal time.Duration) *Dispatcher {
	ctx, ctxCancel := context.WithCancel(ctx)
	dispatcher := &Dispatcher{
		workers:  make([]Worker, numWorkers),
		db:       db,
		queue:    make(chan model.Task, numWorkers*2),
		repo:     repo,
		logger:   logger,
		interval: internal,

		ctx:    ctx,
		cancel: ctxCancel,
	}
	for i := 0; i < numWorkers; i++ {
		dispatcher.workers[i] = NewWorker(ctx, i, db, dispatcher.queue, repo, logger)
	}
	return dispatcher
}

func (d *Dispatcher) Run() {
	for _, worker := range d.workers {
		worker.Start()
	}

	go func() {
		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()
		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				tasks, err := d.repo.ProcessRawData(d.ctx, d.db)
				if err != nil {
					d.logger.Error(fmt.Sprintf("error processing raw data: %s", err.Error()))
					continue
				}
				for _, task := range tasks {
					d.queue <- task
				}
			}
		}
	}()
}

func (d *Dispatcher) Stop() {
	d.cancel()
}
