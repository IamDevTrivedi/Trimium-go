package health

import (
	"context"
	"time"
)

const pingTimeout = 2 * time.Second

type StatusResponse struct {
	Status string `json:"status"`
}

type Service interface {
	GetAppStatus(ctx context.Context) (*StatusResponse, error)
	GetRedisStatus(ctx context.Context) (*StatusResponse, error)
	GetMongoStatus(ctx context.Context) (*StatusResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAppStatus(_ context.Context) (*StatusResponse, error) {
	return &StatusResponse{Status: "ok"}, nil
}

func (s *service) GetRedisStatus(ctx context.Context) (*StatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := s.repo.PingRedis(ctx); err != nil {
		return nil, err
	}
	return &StatusResponse{Status: "ok"}, nil
}

func (s *service) GetMongoStatus(ctx context.Context) (*StatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := s.repo.PingMongo(ctx); err != nil {
		return nil, err
	}
	return &StatusResponse{Status: "ok"}, nil
}
