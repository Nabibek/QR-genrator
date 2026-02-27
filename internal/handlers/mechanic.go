package handlers

import (
	"fmt"
	"net/http"
	"time"

	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateWorkOrderRequest struct {
	Equipment       string               `json:"equipment" binding:"required"`
	EquipmentNumber string               `json:"equipment_number" binding:"required"`
	WorkType        string               `json:"work_type" binding:"required"`
	Priority        string               `json:"priority"`
	Description     string               `json:"description"`
	MechanicID      string               `json:"mechanic_id"`
	Items           []WorkOrderItemInput `json:"items" binding:"required,min=1"`
}

type WorkOrderItemInput struct {
	ItemID        string `json:"item_id"`
	Name          string `json:"name" binding:"required"`
	PartNumber    string `json:"part_number"`
	Unit          string `json:"unit"`
	Quantity      int    `json:"quantity" binding:"required,min=1"`
	Justification string `json:"justification"`
}

// CreateWorkOrder POST /api/mechanic/order
func CreateWorkOrder(c *gin.Context) {
	var req CreateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := database.GetDB()

	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	// Генерируем читаемый ID: WO-20240226-XXXX
	orderID := fmt.Sprintf("WO-%s-%s",
		time.Now().Format("20060102"),
		uuid.New().String()[:4])

	order := models.WorkOrder{
		ID:              orderID,
		MechanicID:      req.MechanicID,
		Equipment:       req.Equipment,
		EquipmentNumber: req.EquipmentNumber,
		WorkType:        req.WorkType,
		Priority:        priority,
		Description:     req.Description,
		Status:          "pending",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Сохраняем позиции
	for _, it := range req.Items {
		unit := it.Unit
		if unit == "" {
			unit = "шт"
		}
		item := models.WorkOrderItem{
			WorkOrderID:   order.ID,
			ItemID:        it.ItemID,
			Name:          it.Name,
			PartNumber:    it.PartNumber,
			Unit:          unit,
			Quantity:      it.Quantity,
			Justification: it.Justification,
			Status:        "pending",
		}
		db.Create(&item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"order_id": order.ID,
		"message":  "Заявка успешно создана",
	})
}

// GetMyOrders GET /api/mechanic/orders
func GetMyOrders(c *gin.Context) {
	mechanicID := c.Query("mechanic_id")
	if mechanicID == "" {
		// Берём из заголовка Authorization (упрощённо)
		mechanicID = c.GetHeader("X-User-ID")
	}

	db := database.GetDB()

	var orders []models.WorkOrder
	query := db.Preload("Items").Order("created_at DESC")

	if mechanicID != "" {
		query = query.Where("mechanic_id = ?", mechanicID)
	}

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Добавляем items_count для карточек
	type OrderWithCount struct {
		models.WorkOrder
		ItemsCount int `json:"items_count"`
	}

	result := make([]map[string]interface{}, 0, len(orders))
	for _, o := range orders {
		m := map[string]interface{}{
			"id":               o.ID,
			"mechanic_id":      o.MechanicID,
			"equipment":        o.Equipment,
			"equipment_number": o.EquipmentNumber,
			"work_type":        o.WorkType,
			"priority":         o.Priority,
			"description":      o.Description,
			"status":           o.Status,
			"items":            o.Items,
			"items_count":      len(o.Items),
			"created_at":       o.CreatedAt,
			"updated_at":       o.UpdatedAt,
		}
		result = append(result, m)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "orders": result})
}

// GetWorkOrder GET /api/mechanic/order/:id
func GetWorkOrder(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()

	var order models.WorkOrder
	if err := db.Preload("Items").Preload("Mechanic").First(&order, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Заявка не найдена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "order": order})
}
