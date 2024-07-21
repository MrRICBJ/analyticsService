package app

import (
	"analitycsService/internal/api/v1"
	repository2 "analitycsService/internal/repository"
	"analitycsService/internal/service"
	"analitycsService/pkg/postgres"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	postgres_migrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm/v2"
	"net/http"
	"sync"
	"time"

	repository "analitycsService/internal/repository/postgres"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultRouteGroup = "/"
	internal          = 1 * time.Second
)

type Application interface {
	RegisterBackgroundJob(backgroundJob func() error)
	RegisterStopHandler(stopHandler func())

	Run() error
}

func New(ctx context.Context, config *Config) (Application, error) {
	appInstance := &application{
		config: config,
		ctx:    ctx,
	}
	var err error

	appInstance.logger = logrus.New()
	appInstance.logger.SetLevel(convertLogLevel(config.LogLevel))
	if err = appInstance.initDB(ctx, config.Database); err != nil {
		return nil, fmt.Errorf("initDB: %s", err.Error())
	}
	appInstance.repo = repository.NewRepo()
	appInstance.service = service.NewService(appInstance.db, appInstance.logger, appInstance.repo)

	appInstance.initHttpServer(appInstance.ctx, api.New(appInstance.logger, appInstance.service))
	dispatcher := service.NewDispatcher(ctx, config.NumWorkers, appInstance.db, appInstance.repo, appInstance.logger, internal)
	dispatcher.Run()
	appInstance.RegisterStopHandler(func() {
		dispatcher.Stop()
	})
	return appInstance, nil
}

func (a *application) initDB(ctx context.Context, cfg postgres.Config) error {
	var err error
	a.db, err = postgres.New(ctx, cfg)
	if err != nil {
		return err
	}

	a.RegisterStopHandler(func() {
		a.db.Close()
	})

	db := stdlib.OpenDBFromPool(a.db)

	driver, err := postgres_migrate.WithInstance(db, &postgres_migrate.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations/.",
		"postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (a *application) initHttpServer(ctx context.Context, handler api.Api) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	privateGroup := engine.Group(defaultRouteGroup)
	handler.RegisterPrivateHandlers(privateGroup)

	server := &http.Server{
		Handler:           engine,
		Addr:              a.config.HTTP.ListenAddr,
		IdleTimeout:       a.config.HTTP.KeepaliveTime + a.config.HTTP.KeepaliveTimeout,
		ReadHeaderTimeout: a.config.HTTP.ReadHeaderTimeout,
	}
	a.RegisterStopHandler(func() { _ = server.Shutdown(ctx) })

	a.RegisterBackgroundJob(func() error {
		a.logger.Info(fmt.Sprintf("starting HTTP server on addr %s", a.config.HTTP.ListenAddr))
		return server.ListenAndServe()
	})
}

type application struct {
	config         *Config
	ctx            context.Context
	db             *pgxpool.Pool
	service        service.Service
	repo           repository2.Repo
	logger         *logrus.Logger
	backgroundJobs []func() error
	stopHandlers   []func()
	tracer         *apm.Tracer
}

func (a *application) Run() error {
	defer a.stop()
	errors := a.startBackgroundJobs()

	select {
	case <-a.ctx.Done():
		return nil
	case err := <-errors:
		return err
	}
}

func (a *application) RegisterBackgroundJob(backgroundJob func() error) {
	a.backgroundJobs = append(a.backgroundJobs, backgroundJob)
}

func (a *application) RegisterStopHandler(stopHandler func()) {
	a.stopHandlers = append(a.stopHandlers, stopHandler)
}

func (a *application) startBackgroundJobs() chan error {
	errors := make(chan error)

	for _, job := range a.backgroundJobs {
		_job := job // to prevent variable override during loop iterations
		go func() {
			errors <- _job()
		}()
	}

	return errors
}

func (a *application) stop() {
	var wg sync.WaitGroup
	wg.Add(len(a.stopHandlers))
	for _, handler := range a.stopHandlers {
		_handler := handler // to prevent variable override during loop iterations
		go func() {
			defer wg.Done()
			_handler()
		}()
	}
	wg.Wait()
}
