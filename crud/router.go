package crud

import (
	"encoding/json"
	"github.com/ElegantSoft/go-crud-starter/db"
	"github.com/ElegantSoft/go-crud-starter/db/models"
	"github.com/gin-gonic/gin"
	"log"
	"math"
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

		var result []models.Post
		tx := db.DB.Model(models.Post{})
		if api.Page > 0 {
			tx.Limit(int(api.Limit)).Offset(int((api.Page - 1) * api.Limit))
		}
		err, transaction := searchToQuery(s, tx.Model(&models.Post{}))
		if err != nil {
			log.Printf("err -> %+v", err)
		}
		transaction.Scan(&result)

		log.Printf("api params -> %+v", api)
		log.Printf("search is -> %+v", s)
		var data interface{}
		var totalRows int64
		db.DB.Model(&models.Post{}).Count(&totalRows)
		if api.Page > 0 {
			data = map[string]interface{}{
				"data":       result,
				"total":      totalRows,
				"totalPages": int(math.Ceil(float64(totalRows) / float64(api.Limit))),
			}
		} else {
			data = result
		}
		ctx.JSON(200, gin.H{"api": api, "data": data})
	})
}
