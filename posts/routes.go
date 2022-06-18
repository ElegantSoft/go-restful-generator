package posts

import "github.com/gin-gonic/gin"

func RegisterRoutes(routerGroup *gin.RouterGroup) {
	service := InitService()
	controller := NewController(service)

	routerGroup.GET("", controller.findAll)
	routerGroup.GET(":id", controller.findOne)
	//routerGroup.POST( "/", controller.Create)
	//routerGroup.DELETE( "/:id", controller.Delete)
	//routerGroup.PUT( "/:id", controller.Update)
}
