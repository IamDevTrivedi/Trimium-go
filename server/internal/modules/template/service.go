package template

import "server/internal/queue"

type Service interface {
}

type service struct {
	repo  Repository
	queue queue.Queue
}

func NewService(repo Repository, queue queue.Queue) Service {
	return &service{
		repo:  repo,
		queue: queue,
	}
}
