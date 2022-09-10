package seed

import (
	"github.com/ElegantSoft/go-crud-starter/db"
	"github.com/ElegantSoft/go-crud-starter/db/models"
	"github.com/bxcodec/faker/v3"
	"log"
)

func SeedPosts() {
	for i := 0; i < 20; i++ {
		cat := models.Category{
			Name: faker.Word(),
		}
		log.Printf("will create cat %v %v", i, cat)
		err := db.DB.Create(&cat).Error
		if err != nil {
			log.Printf("error while seed category %+v", err)
		}
		for l := 0; l < 50; l++ {
			post := models.Post{
				Title:       faker.Word(),
				Description: faker.Paragraph(),
				CategoryID:  cat.ID,
				Price:       uint32(l + 1*i + 1),
			}
			err := db.DB.Create(&post).Error
			if err != nil {
				log.Printf("error while seed post %+v", err)
			}
		}
	}
}
