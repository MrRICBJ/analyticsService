package service

import (
	"analitycsService/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const keyID = "X-Tantum-Authorization"

type service struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
	repo   repository.Repo
}

type Service interface {
	SaveRowData(ctx context.Context, headers http.Header, bodyBytes []byte, time time.Time) error
}

func NewService(
	db *pgxpool.Pool,
	logger *logrus.Logger,
	repo repository.Repo,
) Service {
	return &service{
		db:     db,
		logger: logger,
		repo:   repo,
	}
}

type data struct {
	Headers   http.Header `json:"headers"`
	BodyBytes []byte      `json:"body"`
}

func (s *service) SaveRowData(ctx context.Context, headers http.Header, bodyBytes []byte, time time.Time) error {
	jsonData, err := json.Marshal(data{Headers: headers, BodyBytes: bodyBytes})
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to marshal data to JSON: %s", err.Error()))
		return err
	}
	return s.repo.SaveRowData(ctx, s.db, time, headers.Get(keyID), jsonData)
}
