package main

import (
	"github.com/ElegantSoft/go-crud-starter/crud"
	"github.com/ElegantSoft/go-crud-starter/db"
	"github.com/ElegantSoft/go-crud-starter/db/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	server := gin.New()
	server.Use(gin.Recovery())
	if os.Getenv("GIN_MODE") == "debug" {
		server.Use(gin.Logger())
	}
	gin.SetMode(os.Getenv("GIN_MODE"))
	if err := db.Open(os.Getenv("DB_URL")); err != nil {
		log.Fatal(err)
	}
	log.Println("server started")

	server.GET("", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "ok 5"})
	})

	// migrations
	db.AddUUIDExtension()

	if err := db.DB.AutoMigrate(
		models.Category{},
		models.Post{},
	); err != nil {
		log.Fatal(err)
	}

	crudGroup := server.Group("crud")

	crud.RegisterRoutes(crudGroup)

	//seed.SeedPosts()

	err := server.Run()
	if err != nil {
		log.Printf("error while starting server %+v", err)
	}
}
