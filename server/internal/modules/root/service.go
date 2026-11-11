package root

import (
	"context"
	"errors"

	"server/internal/queue"
	"server/internal/tasks"
)

const (
	welcomeFromEmail = "hello@example.com"
	welcomeSubject   = "Welcome to Trimium!"
	welcomeHTMLBody  = "<h1>Welcome!</h1><p>Thank you for joining Trimium.</p>"
)

type Service interface {
	GetAppInfo(ctx context.Context) (*AppInfo, error)
	SendWelcomeEmail(ctx context.Context, to []string) error
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

func (s *service) SendWelcomeEmail(ctx context.Context, to []string) error {
	if len(to) == 0 {
		return errors.New("root: welcome email requires at least one recipient")
	}

	return s.queue.SendEmail(ctx, tasks.EmailPayload{
		FromEmail: welcomeFromEmail,
		ToEmail:   to,
		Subject:   welcomeSubject,
		HTMLBody:  welcomeHTMLBody,
	})
}
