package root

type Service interface {
	GetStatus() map[string]any
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) GetStatus() map[string]any {
	return map[string]any{
		"message": "Welcome to Trimium API",
		"status":  "ok",
	}
}
