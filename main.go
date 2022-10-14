package main

import (
	"log"
	"os"

	"github.com/ElegantSoft/go-restful-generator/crud"
	"github.com/ElegantSoft/go-restful-generator/db"
	"github.com/ElegantSoft/go-restful-generator/db/models"
	_ "github.com/ElegantSoft/go-restful-generator/docs"
	"github.com/ElegantSoft/go-restful-generator/posts"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	server := gin.New()
	server.Use(gin.Recovery())

	config := cors.DefaultConfig()
	// config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowAllOrigins = true
	config.AddAllowHeaders("Authorization")

	server.Use(cors.New(config))

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
	postsGroup := server.Group("posts")

	crud.RegisterRoutes(crudGroup)
	posts.RegisterRoutes(postsGroup)

	server.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// seed.SeedPosts()

	err := server.Run()
	if err != nil {
		log.Printf("error while starting server %+v", err)
	}
}
