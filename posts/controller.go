package posts

//
//import (
//	"github.com/gin-gonic/gin"
//	"net/http"
//)
//
//type Controller struct {
//	service *Service
//}
//
//func (s *Controller) Find(ctx *gin.Context) {
//	var item ById
//	if err := ctx.ShouldBindUri(&item); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	err, found := s.service.Find(item.ID)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
//		return
//	}
//	ctx.JSON(http.StatusOK, gin.H{"data": &found})
//}
//
//func (s *Controller) Create(ctx *gin.Context) {
//	var item Interest
//	if err := ctx.ShouldBind(&item); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	err, found := s.service.Create(&item)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
//		return
//	}
//	ctx.JSON(http.StatusOK, gin.H{"data": &found})
//}
//
//func (s *Controller) Delete(ctx *gin.Context) {
//	var item ById
//	if err := ctx.ShouldBindUri(&item); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	err := s.service.Delete(item.ID)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
//		return
//	}
//	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
//}
//
//func (s *Controller) Update(ctx *gin.Context) {
//	var item Interest
//	var byId ById
//	if err := ctx.ShouldBind(&item); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	if err := ctx.ShouldBindUri(&byId); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//	err := s.service.Update(&item, byId.ID)
//	if err != nil {
//		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
//		return
//	}
//	ctx.JSON(http.StatusOK, gin.H{"message": "updated"})
//}
//
//
//
//
//func NewController(service *Service) *Controller {
//	return &Controller{
//		service: service,
//	}
//}
