package app

import (
	"analitycsService/pkg/postgres"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	DotEnvFilename = ".env"
)

type Config struct {
	LogLevel    string           `envconfig:"LOG_LEVEL" default:"debug"`
	HTTP        HTTPServerConfig `envconfig:"HTTP"`
	Database    postgres.Config  `envconfig:"DB"`
	NumWorkers  int              `envconfig:"NUM_WORKERS" default:"1"`
	ServiceName string
}

type HTTPServerConfig struct {
	ListenAddr        string        `envconfig:"API_LISTEN_ADDR" default:":8080"`
	KeepaliveTime     time.Duration `envconfig:"KEEPALIVE_TIME" default:"60s"`
	KeepaliveTimeout  time.Duration `envconfig:"KEEPALIVE_TIMEOUT" default:"5s"`
	ReadHeaderTimeout time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"5s"`
}

func NewConfigFromEnv() (*Config, error) {
	cfg := &Config{}
	var err error
	_ = godotenv.Load(DotEnvFilename)
	err = envconfig.Process("", cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable process .env")
	}

	return cfg, nil
}

func convertLogLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}
