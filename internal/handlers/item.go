package handlers

import (
	"net/http"

	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ItemResponse - ответ с информацией о товаре
type ItemResponse struct {
	Success bool         `json:"success"`
	Item    *models.Item `json:"item,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// GetItem - обработчик GET /api/item/:id
// Возвращает информацию о товаре с его текущей локацией
func GetItem(c *gin.Context) {
	itemID := c.Param("id")

	if itemID == "" {
		c.JSON(http.StatusBadRequest, ItemResponse{
			Success: false,
			Error:   "ID товара не указан",
		})
		return
	}

	db := database.GetDB()
	var item models.Item

	// Получаем товар с информацией о локации (joinом)
	if err := db.Preload("Location").First(&item, "id = ?", itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ItemResponse{
				Success: false,
				Error:   "Товар не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ItemResponse{
				Success: false,
				Error:   "Ошибка при получении товара",
			})
		}
		return
	}

	c.JSON(http.StatusOK, ItemResponse{
		Success: true,
		Item:    &item,
	})
}

// GetItemHistory - обработчик GET /api/item/:id/history
// Возвращает историю всех перемещений товара
type ItemHistoryResponse struct {
	Success   bool                  `json:"success"`
	ItemID    string                `json:"item_id"`
	Movements []models.ItemMovement `json:"movements,omitempty"`
	Total     int64                 `json:"total"`
	Error     string                `json:"error,omitempty"`
}

func GetItemHistory(c *gin.Context) {
	itemID := c.Param("id")

	if itemID == "" {
		c.JSON(http.StatusBadRequest, ItemHistoryResponse{
			Success: false,
			Error:   "ID товара не указан",
		})
		return
	}

	db := database.GetDB()

	// Проверяем, существует ли товар
	var item models.Item
	if err := db.First(&item, "id = ?", itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ItemHistoryResponse{
				Success: false,
				Error:   "Товар не найден",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ItemHistoryResponse{
				Success: false,
				Error:   "Ошибка при получении товара",
			})
		}
		return
	}

	var movements []models.ItemMovement
	var count int64

	// Получаем историю перемещений (сортируем по дате, новые первыми)
	if err := db.Where("item_id = ?", itemID).
		Preload("FromLocation").
		Preload("ToLocation").
		Preload("User").
		Order("moved_at DESC").
		Find(&movements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ItemHistoryResponse{
			Success: false,
			Error:   "Ошибка при получении истории",
		})
		return
	}

	count = int64(len(movements))

	c.JSON(http.StatusOK, ItemHistoryResponse{
		Success:   true,
		ItemID:    itemID,
		Movements: movements,
		Total:     count,
	})
}
