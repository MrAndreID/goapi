package version

import (
	"context"
)

type Service struct {
	Name    string
	Version string
}

func NewService(name, version string) *Service {
	return &Service{
		Name:    name,
		Version: version,
	}
}

type InterfaceService interface {
	Read(context.Context) Info
}

func (s *Service) Read(ctx context.Context) Info {
	return Info{
		Name:    s.Name,
		Version: s.Version,
	}
}
