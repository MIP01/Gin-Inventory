package route

import (
	"Gin-Inventory/controller"
	"Gin-Inventory/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup) {
	api.POST("/login", middleware.LoginHandler)

	api.GET("/user", controller.GetAllUserHandler)
	api.POST("/user", controller.CreateUserHandler)

	api.GET("/admin", controller.GetAllAdminHandler)
	api.POST("/admin", controller.CreateAdminHandler)

	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/logout", middleware.LogoutHandler)

		auth.GET("/user/:id", controller.GetUserHandler)
		auth.PUT("/user/:id", controller.UpdateUserHandler)
		auth.DELETE("/user/:id", controller.DeleteUserHandler)

		auth.GET("/admin/:id", controller.GetAdminHandler)
		auth.PUT("/admin/:id", controller.UpdateAdminHandler)
		auth.DELETE("/admin/:id", controller.DeleteAdminHandler)
	}
}
