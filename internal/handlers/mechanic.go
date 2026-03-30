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
	orderID := fmt.Sprintf("WO-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:4])
	order := models.WorkOrder{
		ID:              orderID,
		MechanicID:      req.MechanicID,
		Equipment:       req.Equipment,
		EquipmentNumber: req.EquipmentNumber,
		WorkType:        req.WorkType,
		Priority:        req.Priority,
		Description:     req.Description,
		Status:          "pending",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	db.Create(&order)

	for i, it := range req.Items {

		orderItem := models.WorkOrderItem{
			WorkOrderID: order.ID,
			ItemID:      it.ItemID,
			Name:        it.Name,
			Quantity:    it.Quantity,
			Status:      "pending",
		}
		db.Create(&orderItem)

		// 1. Проверяем реальное наличие на складе
		var item models.Item
		err := db.First(&item, "id = ?", it.ItemID).Error

		// 2. Считаем сколько ВСЕГО заявок в таблице снабжения, чтобы чередовать их
		var totalSupplyReqs int64
		db.Model(&models.SupplyRequest{}).Count(&totalSupplyReqs)

		// ЛОГИКА ОТПРАВКИ В СНАБЖЕНИЕ:
		// - Если товара нет в базе вообще (err != nil)
		// - ИЛИ если товара меньше чем нужно (item.Quantity < it.Quantity)
		// - ИЛИ (для демо) если общее кол-во заявок четное
		// sendToSupply := (err != nil) || (item.Quantity < it.Quantity) || (totalSupplyReqs%2 == 0)
		sendToSupply := (err != nil) || true

		if sendToSupply {
			// Генерируем ID для заявки на снабжение
			sID := fmt.Sprintf("REQ-%d%d", time.Now().Unix()%10000, i)

			supplyReq := models.SupplyRequest{
				ID:          sID,
				ItemID:      it.ItemID,
				ItemName:    it.Name, // Сохраняем имя товара
				RequestedBy: req.MechanicID,
				Quantity:    it.Quantity,
				Reason:      fmt.Sprintf("Заявка %s: %s (Техника: %s)", order.ID, it.Justification, req.Equipment),
				Status:      "created",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := db.Create(&supplyReq).Error; err != nil {
				fmt.Println("Ошибка создания SupplyRequest:", err)
			}

			// Обновляем статус позиции в заказе механика
			db.Model(&orderItem).Update("status", "awaiting_supply")
			fmt.Printf(">>> Товар %s отправлен в СНАБЖЕНИЕ (ID: %s)\n", it.Name, sID)
		} else {
			db.Model(&orderItem).Update("status", "in_stock")
			fmt.Printf(">>> Товар %s есть НА СКЛАДЕ\n", it.Name)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"order_id": order.ID,
		"message":  "Заявка обработана",
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
