package root

import (
	"context"

	"server/internal/queue"
)

const (
	welcomeFromEmail = "hello@example.com"
	welcomeSubject   = "Welcome to Trimium!"
	welcomeHTMLBody  = "<h1>Welcome!</h1><p>Thank you for joining Trimium.</p>"
)

type Service interface {
	GetAppInfo(ctx context.Context) (*AppInfo, error)
}

type service struct {
	repo  Repository
	queue queue.Queue
}

type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

func NewService(repo Repository, queue queue.Queue) Service {
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
