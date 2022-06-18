package posts

import (
	"github.com/ElegantSoft/go-crud-starter/crud"
	"github.com/ElegantSoft/go-crud-starter/db"
	"github.com/ElegantSoft/go-crud-starter/db/models"
)

type Repository struct {
	crud.Repository[models.Post]
}

func InitRepository() *Repository {
	return &Repository{
		Repository: crud.Repository[models.Post]{
			DB:    db.DB,
			Model: &models.Post{},
		},
	}
}
