package posts

import (
	"github.com/ElegantSoft/go-crud-starter/crud"
	"github.com/ElegantSoft/go-crud-starter/db/models"
)

type Service struct {
	crud.Service[models.Post]
	repo *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		Service: *crud.NewService[models.Post](repository),
		repo:    repository,
	}
}

func InitService() *Service {
	return &Service{
		repo: InitRepository(),
	}
}
