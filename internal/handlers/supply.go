package handlers

import (
	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// var DB = database.DB // если у тебя так подключено
func CreateSupplyRequest(c *gin.Context) {
	db := database.GetDB()

	fmt.Println(">>> CreateSupplyRequest called")

	var input struct {
		ItemID   string `json:"item_id"`
		Quantity int    `json:"quantity"`
		Reason   string `json:"reason"`
		UserID   string `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req := models.SupplyRequest{
		ID:          uuid.New().String(),
		ItemID:      input.ItemID,
		RequestedBy: input.UserID,
		Quantity:    input.Quantity,
		Reason:      input.Reason,
		Status:      "created",
	}

	if err := db.Create(&req).Error; err != nil {
		fmt.Println("DB ERROR:", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(">>> Created OK")

	c.JSON(200, gin.H{
		"success": true,
		"request": req,
	})
}

func ApproveByEngineer(c *gin.Context) {

	id := c.Param("id")

	var req models.SupplyRequest

	db := database.GetDB()
	if err := db.First(&req, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	req.Status = "approved_by_engineer"

	db.Save(&req)

	c.JSON(200, gin.H{"success": true, "status": req.Status})
}

func AssignProcurement(c *gin.Context) {

	id := c.Param("id")

	var input struct {
		AssignedTo string `json:"assigned_to"`
	}

	c.ShouldBindJSON(&input)

	task := models.ProcurementTask{
		ID:         uuid.New().String(),
		RequestID:  id,
		AssignedTo: input.AssignedTo,
		Status:     "assigned",
	}

	db := database.GetDB()

	db.Create(&task)

	db.Model(&models.SupplyRequest{}).
		Where("id = ?", id).
		Update("status", "assigned_to_procurement")

	c.JSON(200, gin.H{"success": true})
}

func SelectSupplier(c *gin.Context) {

	id := c.Param("id")

	var input struct {
		SupplierID string  `json:"supplier_id"`
		Price      float64 `json:"price"`
	}

	c.ShouldBindJSON(&input)

	db := database.GetDB()
	db.Model(&models.ProcurementTask{}).
		Where("request_id = ?", id).
		Updates(map[string]interface{}{
			"supplier_id": input.SupplierID,
			"price":       input.Price,
			"status":      "supplier_selected",
		})

	db.Model(&models.SupplyRequest{}).
		Where("id = ?", id).
		Update("status", "supplier_selected")

	c.JSON(200, gin.H{"success": true})
}

func ApproveByManager(c *gin.Context) {
	db := database.GetDB()
	id := c.Param("id")

	db.Model(&models.SupplyRequest{}).
		Where("id = ?", id).
		Update("status", "approved_by_manager")

	c.JSON(200, gin.H{"success": true})
}

func ApproveByCommercial(c *gin.Context) {
	db := database.GetDB()
	id := c.Param("id")

	db.Model(&models.SupplyRequest{}).
		Where("id = ?", id).
		Update("status", "approved_by_commercial")

	c.JSON(200, gin.H{"success": true})
}

func ReceiveSupply(c *gin.Context) {
	db := database.GetDB()
	id := c.Param("id")

	var req models.SupplyRequest

	if err := db.First(&req, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "request not found"})
		return
	}

	// увеличиваем количество товара
	db.Model(&models.Item{}).
		Where("id = ?", req.ItemID).
		Update("quantity", gorm.Expr("quantity + ?", req.Quantity))

	// меняем статус
	db.Model(&models.SupplyRequest{}).
		Where("id = ?", id).
		Update("status", "received")

	c.JSON(200, gin.H{
		"success": true,
		"message": "Товар принят на склад",
	})
}

// POST /api/supply/:id/reject-commercial
func RejectByCommercial(c *gin.Context) {
	db := database.GetDB()
	id := c.Param("id")

	// Сбрасываем статус на самый первый этап
	db.Model(&models.SupplyRequest{}).Where("id = ?", id).Update("status", "created")

	c.JSON(200, gin.H{"success": true, "message": "Заявка возвращена на уточнение"})
}
func GetSupplyRequests(c *gin.Context) {
	db := database.GetDB()
	var requests []models.SupplyRequest

	// Загружаем все заявки на снабжение
	if err := db.Order("created_at DESC").Find(&requests).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    requests,
	})
}
