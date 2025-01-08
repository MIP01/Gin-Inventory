package route

import (
	"gin_iventory/controller"
	"gin_iventory/middleware"

	"github.com/gin-gonic/gin"
)

func SetupItemRoutes(api *gin.RouterGroup) {
	api.GET("/item", controller.GetAllItemHandler)
	api.GET("/item/:id", controller.GetItemHandler)

	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/item", controller.CreateItemHandler)
		api.PUT("/item/:id", controller.UpdateItemHandler)
		api.DELETE("/item/:id", controller.DeleteItemHandler)

		api.GET("/chart", controller.GetTransactionsHandler)
		api.POST("/chart", controller.CreateTransactionHandler)
		api.PUT("/chart/:id", controller.UpdateTransactionHandler)
		api.DELETE("/chart/:id", controller.DeleteTransactionHandler)
	}
}
