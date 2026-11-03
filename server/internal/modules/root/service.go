package root

import (
	"context"

	"github.com/hibiken/asynq"
)

type Service interface {
	GetAppInfo(ctx context.Context) (*AppInfo, error)
}

type service struct {
	repo  Repository
	queue *asynq.Client
}

func NewService(repo Repository, queue *asynq.Client) Service {
	return &service{
		repo:  repo,
		queue: queue,
	}
}

func (s *service) GetAppInfo(_ context.Context) (*AppInfo, error) {
	return &AppInfo{
		Name:    "Trimium API",
		Version: "1.0.0",
		Status:  "ok",
	}, nil
}
