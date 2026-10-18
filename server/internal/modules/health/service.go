package health

type Service interface {
	GetStatus() (map[string]any, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) GetStatus() (map[string]any, error) {
	return map[string]any{
		"status": "ok",
	}, nil
}
