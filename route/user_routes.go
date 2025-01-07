package route

import (
	"gin_iventory/controller"
	"gin_iventory/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup) {
	api.POST("/login", middleware.LoginHandler)

	api.GET("/user", controller.GetAllUserHandler)
	api.POST("/user", controller.CreateUserHandler)

	api.GET("/admin", controller.GetAllAdminHandler)
	api.POST("/admin", controller.CreateAdminHandler)

	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/logout", middleware.LogoutHandler)

		api.GET("/user/:id", controller.GetUserHandler)
		api.PUT("/user/:id", controller.UpdateUserHandler)
		api.DELETE("/user/:id", controller.DeleteUserHandler)

		api.GET("/admin/:id", controller.GetAdminHandler)
		api.PUT("/admin/:id", controller.UpdateAdminHandler)
		api.DELETE("/admin/:id", controller.DeleteAdminHandler)
	}
}
