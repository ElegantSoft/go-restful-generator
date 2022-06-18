package posts

import (
	"fmt"
	"github.com/ElegantSoft/go-crud-starter/common"
	"github.com/ElegantSoft/go-crud-starter/crud"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"math"
	"net/http"
)

type Controller struct {
	service *Service
}

func (c *Controller) findAll(ctx *gin.Context) {
	var api crud.GetAllRequest
	if api.Limit == 0 {
		api.Limit = 20
	}
	if err := ctx.ShouldBindQuery(&api); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var result []model
	var totalRows int64
	err := c.service.Find(api, &result, &totalRows)
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
}

func (c *Controller) findOne(ctx *gin.Context) {
	var api crud.GetAllRequest
	var item common.ById
	if err := ctx.ShouldBindQuery(&api); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindUri(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ValidateErrors(err)})
		return
	}

	api.Filter = append(api.Filter, fmt.Sprintf("id||$eq||%s", item.ID))

	var result model

	err := c.service.FindOne(api, &result)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, result)
}

func (s *Controller) Create(ctx *gin.Context) {
	var item model
	if err := ctx.ShouldBind(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.service.Create(&item)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": item})
}

func (s *Controller) Delete(ctx *gin.Context) {
	var item common.ById
	if err := ctx.ShouldBindUri(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.FromBytes([]byte(item.ID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.service.Delete(&model{ID: id})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (s *Controller) Update(ctx *gin.Context) {
	var item model
	var byId common.ById
	if err := ctx.ShouldBind(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := uuid.FromBytes([]byte(byId.ID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindUri(&byId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.service.Update(&model{ID: id}, &item)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}
