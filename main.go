package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/signal"
	"syscall"

	"analitycsService/internal/app"
	"github.com/go-playground/validator"
)

var serviceName = "analyticsService"

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	logger := logrus.New()
	appConfig, err := app.NewConfigFromEnv()
	if err != nil {
		logger.Fatal(fmt.Sprintf("can't read config: %s", err.Error()))
	}

	appConfig.ServiceName = serviceName
	validate := validator.New()
	err = validate.Struct(appConfig)
	if err != nil {
		logger.Fatal(fmt.Sprintf("app config validation failed: %s", err.Error()))
	}

	application, err := app.New(
		ctx,
		appConfig,
	)
	if err != nil {
		logger.Fatal(fmt.Sprintf("application could not been initialized: %s", err.Error()))
	}

	if err = application.Run(); err != nil {
		logger.Fatal(fmt.Sprintf("application terminated abnormally: %s", err.Error()))
	}
}
