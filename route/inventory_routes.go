package route

import (
	"gin_iventory/controller"
	"gin_iventory/middleware"

	"github.com/gin-gonic/gin"
)

func SetupItemRoutes(api *gin.RouterGroup) {
	api.GET("/item", controller.GetAllItemHandler)
	api.GET("/item/:item_id", controller.GetItemHandler)

	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.POST("/item", controller.CreateItemHandler)
		auth.PUT("/item/:item_id", controller.UpdateItemHandler)
		auth.DELETE("/item/:item_id", controller.DeleteItemHandler)

		auth.GET("/chart", controller.GetTransactionsHandler)
		auth.POST("/chart", controller.CreateTransactionHandler)
		auth.PUT("/chart/:chart_id", controller.UpdateTransactionHandler)
		auth.DELETE("/chart/:chart_id", controller.DeleteTransactionHandler)

		auth.GET("/detail", controller.GetAllDetailHandler)
		auth.GET("/detail/:detail_id", controller.GetDetailHandler)
		auth.POST("/detail", controller.CreateDetailHandler)
		auth.PUT("/detail/:detail_id", controller.UpdateDetailHandler)
		auth.DELETE("/detail/:detail_id", controller.DeleteDetailHandler)
	}
}
