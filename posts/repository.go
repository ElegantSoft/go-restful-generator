package posts

import (
	"github.com/ElegantSoft/go-restful-generator/crud"
	"github.com/ElegantSoft/go-restful-generator/db"
	"github.com/ElegantSoft/go-restful-generator/db/models"
)

type model = models.Post

type Repository struct {
	crud.Repository[model]
}

func InitRepository() *Repository {
	return &Repository{
		Repository: crud.Repository[model]{
			DB:    db.DB,
			Model: &model{},
		},
	}
}
