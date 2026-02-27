package routes

import (
	"QR-GENERATOR/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "🚀 Warehouse API is running"})
	})

	// Публичные
	api := router.Group("/api")
	{
		api.POST("/login", handlers.Login)
		api.GET("/me", handlers.CurrentUser)
		api.GET("/item/:id", handlers.GetItem)
		api.GET("/item/:id/history", handlers.GetItemHistory)
		api.POST("/move", handlers.MoveItem)
	}

	// Админ
	admin := router.Group("/api/admin")
	{
		admin.POST("/item", handlers.AdminCreateItem)
		admin.GET("/items", handlers.AdminGetItems)
		admin.PUT("/item/:id", handlers.AdminUpdateItem)
		admin.DELETE("/item/:id", handlers.AdminDeleteItem)
		admin.GET("/item/:id/qr", handlers.AdminGetItemQR)
		admin.POST("/item/:id/photo", handlers.AdminUploadInvoicePhoto)
		admin.GET("/locations", handlers.AdminGetLocations)
		admin.POST("/location", handlers.AdminCreateLocation)
		admin.GET("/location/:id/qr", handlers.AdminGetLocationQR)
		admin.GET("/categories", handlers.AdminGetCategories)
	}

	// Механик
	mechanic := router.Group("/api/mechanic")
	{
		mechanic.POST("/order", handlers.CreateWorkOrder)
		mechanic.GET("/orders", handlers.GetMyOrders)
		mechanic.GET("/order/:id", handlers.GetWorkOrder)
		mechanic.PUT("/order/:id/status", handlers.UpdateOrderStatus)
		mechanic.POST("/order/:id/qr", handlers.GenerateOrderQR)
		mechanic.POST("/order/:id/issue", handlers.IssueOrder)
	}

	// Статика
	router.StaticFile("/", "./static/index.html")
	router.StaticFile("/admin", "./static/admin.html")
	router.StaticFile("/mechanic", "./static/mechanic.html")
	router.Static("/css", "./static/css")
	router.Static("/js", "./static/js")
	router.Static("/qrcodes", "./qrcodes")
	router.Static("/invoices", "./static/invoices")

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Маршрут не найден"})
	})
}
