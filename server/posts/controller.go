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

// @Success  200  {array}  model
// @Tags     posts
// @param    s       query  string    false  "{'$and': [ {'title': { '$cont':'cul' } } ]}"
// @param    fields  query  string    false  "fields to select eg: name,age"
// @param    page    query  int       false  "page of pagination"
// @param    limit   query  int       false  "limit of pagination"
// @param    join    query  string    false  "join relations eg: category, parent"
// @param    filter  query  []string  false  "filters eg: name||$eq||ad price||$gte||200"
// @param    sort    query  []string  false  "filters eg: created_at,desc title,asc"
// @Router   /posts [get]
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

// @Success  200  {object}  model
// @Tags     posts
// @param    id    path  string  true  "uuid of item"
// @Router   /posts/{id} [get]
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

// @Success  201  {object}  model
// @Tags     posts
// @param    {object}  body  model  true  "item to create"
// @Router   /posts [post]
func (c *Controller) create(ctx *gin.Context) {
	var item model
	if err := ctx.ShouldBind(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := c.service.Create(&item)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": item})
}

// @Success  200  {string}  string  "ok"
// @Tags     posts
// @param    id  path  string  true  "uuid of item"
// @Router   /posts/{id} [delete]
func (c *Controller) delete(ctx *gin.Context) {
	var item common.ById
	if err := ctx.ShouldBindUri(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindUri(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ValidateErrors(err)})
		return
	}
	id, err := uuid.FromBytes([]byte(item.ID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = c.service.Delete(&model{ID: id})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// @Success  200  {string}  string  "ok"
// @Tags     posts
// @param    id  path  string  true  "uuid of item"
// @param    item  body  model   true  "update body"
// @Router   /posts/{id} [put]
func (c *Controller) update(ctx *gin.Context) {
	var item model
	var byId common.ById
	if err := ctx.ShouldBind(&item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindUri(&byId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ValidateErrors(err)})
		return
	}
	id, err := uuid.Parse(byId.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindUri(&byId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = c.service.Update(&model{ID: id}, &item)
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
