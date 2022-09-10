package posts

import (
	"github.com/ElegantSoft/go-crud-starter/server/crud"
	"github.com/ElegantSoft/go-crud-starter/server/db"
	"github.com/ElegantSoft/go-crud-starter/server/db/models"
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
