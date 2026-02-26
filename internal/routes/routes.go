package routes

import (
	"QR-GENERATOR/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes - регистрирует все API маршруты
func SetupRoutes(router *gin.Engine) {
	// Сначала регистрируем API маршруты, потом статические файлы
	// (порядок важен!)

	// Health check (для проверки доступности сервера)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "🚀 Warehouse API is running",
		})
	})

	// Группа API маршрутов с префиксом /api
	api := router.Group("/api")
	{
		// Аутентификация
		api.POST("/login", handlers.Login)
		api.GET("/me", handlers.CurrentUser)

		// Товары
		api.GET("/item/:id", handlers.GetItem)
		api.GET("/item/:id/history", handlers.GetItemHistory)

		// Перемещения
		api.POST("/move", handlers.MoveItem)
	}

	// Статические файлы (HTML, CSS, JS) - регистрируем ПОСЛЕ API
	// router.Static("/", "./static") - перехватит все запросы
	// вместо этого раздаём файлы селективно
	router.StaticFile("/", "./static/index.html")
	router.Static("/css", "./static/css")
	router.Static("/js", "./static/js")
	router.Static("/qrcodes", "./qrcodes")

	// 404 обработчик
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Маршрут не найден",
		})
	})
}
