package root

import "context"

type Service interface {
	GetAppInfo(ctx context.Context) (*AppInfo, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAppInfo(_ context.Context) (*AppInfo, error) {
	return &AppInfo{
		Name:    "Trimium API",
		Version: "1.0.0",
		Status:  "ok",
	}, nil
}
