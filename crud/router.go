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

var qtb = &queryToDBConverter{}

func RegisterRoutes(routerGroup *gin.RouterGroup) {
	routerGroup.GET("", func(ctx *gin.Context) {
		var api GetAll
		var s map[string]interface{}

		if err := ctx.ShouldBindQuery(&api); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if len(api.S) > 0 {
			err := json.Unmarshal([]byte(api.S), &s)
			if err != nil {
				log.Printf("error while convert search %v", err)
			}
		}

		var result []models.Post
		tx := db.DB.Model(models.Post{})

		if len(api.Fields) > 0 {
			fields := strings.Split(api.Fields, ",")
			tx.Select(fields)
		}
		if len(api.Join) > 0 {
			qtb.relationsMapper(api.Join, tx)
		}
		if api.Page > 0 {
			tx.Limit(int(api.Limit)).Offset(int((api.Page - 1) * api.Limit))
		}

		if len(api.Filter) > 0 {
			qtb.filterMapper(api.Filter, tx)
		}

		err := qtb.searchMapper(s, tx)
		if err != nil {
			log.Printf("err -> %+v", err)
		}
		tx.Find(&result)

		log.Printf("api params -> %+v", api)
		log.Printf("search is -> %+v", s)
		var data interface{}
		var totalRows int64
		tx.Count(&totalRows)
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
