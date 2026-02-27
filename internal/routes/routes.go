package routes

import (
	"QR-GENERATOR/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "🚀 Warehouse API is running",
		})
	})

	// Публичные маршруты
	api := router.Group("/api")
	{
		api.POST("/login", handlers.Login)
		api.GET("/me", handlers.CurrentUser)
		api.GET("/item/:id", handlers.GetItem)
		api.GET("/item/:id/history", handlers.GetItemHistory)
		api.POST("/move", handlers.MoveItem)
	}

	// Админ маршруты
	admin := router.Group("/api/admin")
	{
		// Товары
		admin.POST("/item", handlers.AdminCreateItem)
		admin.GET("/items", handlers.AdminGetItems)
		admin.PUT("/item/:id", handlers.AdminUpdateItem)
		admin.DELETE("/item/:id", handlers.AdminDeleteItem)
		admin.GET("/item/:id/qr", handlers.AdminGetItemQR)
		admin.POST("/item/:id/photo", handlers.AdminUploadInvoicePhoto)

		// Локации
		admin.GET("/locations", handlers.AdminGetLocations)
		admin.POST("/location", handlers.AdminCreateLocation)
		admin.GET("/location/:id/qr", handlers.AdminGetLocationQR)

		// Категории
		admin.GET("/categories", handlers.AdminGetCategories)
	}

	// Статические файлы
	router.StaticFile("/", "./static/index.html")
	router.StaticFile("/admin", "./static/admin.html")
	router.Static("/css", "./static/css")
	router.Static("/js", "./static/js")
	router.Static("/qrcodes", "./qrcodes")
	router.Static("/invoices", "./static/invoices")

	// 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Маршрут не найден",
		})
	})
}
