package auth

import "github.com/hibiken/asynq"

type Service interface {
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
