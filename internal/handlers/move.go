package handlers

import (
	"log"
	"net/http"
	"time"

	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"

	"github.com/gin-gonic/gin"
)

// MoveRequest - запрос на перемещение товара
type MoveRequest struct {
	ItemID       string `json:"item_id" binding:"required"`
	ToLocationID string `json:"to_location_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	Notes        string `json:"notes"`
}

// MoveResponse - ответ при успешном перемещении
type MoveResponse struct {
	Success  bool                 `json:"success"`
	Message  string               `json:"message"`
	Movement *models.ItemMovement `json:"movement,omitempty"`
}

// MoveItem - обработчик POST /api/move
// Перемещает товар из текущей локации в новую и записывает в историю
func MoveItem(c *gin.Context) {
	var req MoveRequest

	// Парси JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Невалидные данные: " + err.Error(),
		})
		return
	}

	db := database.GetDB()

	// Проверяем, существует ли товар
	var item models.Item
	if err := db.First(&item, "id = ?", req.ItemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Товар не найден",
		})
		return
	}

	// Проверяем, существует ли целевая локация
	var targetLoc models.Location
	if err := db.First(&targetLoc, "id = ?", req.ToLocationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Целевая локация не найдена",
		})
		return
	}

	// Проверяем, существует ли пользователь
	var user models.User
	if err := db.First(&user, "id = ?", req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Пользователь не найден",
		})
		return
	}

	// Сохраняем исходную локацию
	fromLocationID := item.LocationID

	// Обновляем lokacию товара
	if err := db.Model(&item).Update("location_id", req.ToLocationID).Error; err != nil {
		log.Printf("Ошибка при обновлении локации товара: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Ошибка при обновлении локации товара",
		})
		return
	}

	// Создаём запись в истории перемещений
	movement := models.ItemMovement{
		ItemID:         req.ItemID,
		FromLocationID: fromLocationID,
		ToLocationID:   req.ToLocationID,
		UserID:         req.UserID,
		Notes:          req.Notes,
		MovedAt:        time.Now(),
	}

	if err := db.Create(&movement).Error; err != nil {
		log.Printf("Ошибка при создании записи в истории: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Ошибка при записи в историю",
		})
		return
	}

	log.Printf("✓ Товар %s перемещён из %s в %s (оператор: %s)",
		req.ItemID, fromLocationID, req.ToLocationID, req.UserID)

	c.JSON(http.StatusOK, MoveResponse{
		Success:  true,
		Message:  "Товар успешно перемещён",
		Movement: &movement,
	})
}
