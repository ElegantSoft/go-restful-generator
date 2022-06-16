package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	server := gin.Default()

	server.GET("", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "ok 5"})
	})

	err := server.Run()
	if err != nil {
		log.Printf("error while starting server %+v", err)
	}
}
