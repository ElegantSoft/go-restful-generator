package crud

import (
	"fmt"
	"math"
	"net/http"

	"github.com/ElegantSoft/go-crud-starter/db"
	"github.com/ElegantSoft/go-crud-starter/db/models"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(routerGroup *gin.RouterGroup) {
	repo := NewRepository[models.Post](db.DB, &models.Post{})
	s := NewService[models.Post](repo)

	routerGroup.GET("", func(ctx *gin.Context) {
		var api GetAllRequest
		if err := ctx.ShouldBindQuery(&api); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var result []models.Post
		var totalRows int64
		err := s.Find(api, &result, &totalRows)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var data interface{}
		if api.Page > 0 {
			data = map[string]interface{}{
				"data":       result,
				"total":      totalRows,
				"totalPages": int(math.Ceil(float64(totalRows) / float64(api.Limit))),
			}
		} else {
			data = result
		}
		ctx.JSON(200, data)
	})

	routerGroup.GET(":id", func(ctx *gin.Context) {
		var api GetAllRequest
		var item ById
		if err := ctx.ShouldBindQuery(&api); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := ctx.ShouldBindUri(&item); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		api.Filter = append(api.Filter, fmt.Sprintf("id||$eq||%s", item.ID))

		var result models.Post

		err := s.FindOne(api, &result)
		if err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, result)
	})
}
