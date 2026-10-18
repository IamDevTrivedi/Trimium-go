package root

type Service interface {
	GetStatus() (map[string]any, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) GetStatus() (map[string]any, error) {
	return nil, nil
}
