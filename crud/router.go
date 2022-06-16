package crud

import (
	"encoding/json"
	"github.com/ElegantSoft/go-crud-starter/db"
	"github.com/ElegantSoft/go-crud-starter/db/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strings"
)

func RegisterRoutes(routerGroup *gin.RouterGroup) {
	routerGroup.GET("", func(ctx *gin.Context) {
		var api GetAll
		var s map[string]interface{}

		if err := ctx.ShouldBindQuery(&api); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if len(api.Fields) > 0 {
			api.Fields = strings.Split(api.Fields[0], ",")
		}

		if len(api.S) > 0 {
			err := json.Unmarshal([]byte(api.S), &s)
			if err != nil {
				log.Printf("error while convert search %v", err)
			}
		}

		txString := db.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
			var result []models.Post

			err, transaction := searchToQuery(s, tx.Model(&models.Post{}))
			if err != nil {
				log.Printf("err -> %+v", err)
			}
			return transaction.Scan(&result)
		})
		log.Printf("tx -> %+v", txString)

		log.Printf("api params -> %+v", api)
		log.Printf("search is -> %+v", s)
		ctx.JSON(200, gin.H{"data": api})
	})
}
