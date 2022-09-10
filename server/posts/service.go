package posts

import (
	"github.com/ElegantSoft/go-crud-starter/server/crud"
)

type Service struct {
	crud.Service[model]
	repo *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		Service: *crud.NewService[model](repository),
		repo:    repository,
	}
}

func InitService() *Service {
	return &Service{
		repo:    InitRepository(),
		Service: *crud.NewService[model](InitRepository()),
	}
}
