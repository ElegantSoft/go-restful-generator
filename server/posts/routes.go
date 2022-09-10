package posts

import "github.com/gin-gonic/gin"

func RegisterRoutes(routerGroup *gin.RouterGroup) {
	service := InitService()
	controller := NewController(service)

	routerGroup.GET("", controller.findAll)
	routerGroup.GET(":id", controller.findOne)
	routerGroup.POST("", controller.create)
	routerGroup.DELETE(":id", controller.delete)
	routerGroup.PUT(":id", controller.update)
}
