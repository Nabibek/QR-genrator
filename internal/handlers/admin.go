package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

// ============================================================================
// ITEMS
// ============================================================================

// CreateItemRequest - поля для создания товара
type CreateItemRequest struct {
	Name           string `json:"name" binding:"required"`
	SKU            string `json:"sku" binding:"required"`
	Description    string `json:"description"`
	Quantity       int    `json:"quantity"`
	Unit           string `json:"unit"` // шт, кг, м, л
	Category       string `json:"category"`
	PartNumber     string `json:"part_number"`
	BatchNumber    string `json:"batch_number"`
	BatchQuantity  int    `json:"batch_quantity"`
	BatchArrivedAt string `json:"batch_arrived_at"` // ISO8601
	LocationID     string `json:"location_id"`
}

// AdminCreateItem POST /api/admin/item
func AdminCreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := database.GetDB()

	// Парсим время приезда партии
	var arrivedAt *time.Time
	if req.BatchArrivedAt != "" {
		t, err := time.Parse(time.RFC3339, req.BatchArrivedAt)
		if err != nil {
			// Пробуем формат datetime-local из HTML
			t, err = time.Parse("2006-01-02T15:04", req.BatchArrivedAt)
		}
		if err == nil {
			arrivedAt = &t
		}
	}

	item := models.Item{
		ID:             "item_" + uuid.New().String()[:8],
		Name:           req.Name,
		SKU:            req.SKU,
		Description:    req.Description,
		Quantity:       req.Quantity,
		Unit:           req.Unit,
		Category:       req.Category,
		PartNumber:     req.PartNumber,
		BatchNumber:    req.BatchNumber,
		BatchQuantity:  req.BatchQuantity,
		BatchArrivedAt: arrivedAt,
		LocationID:     req.LocationID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Генерируем QR сразу при создании
	qrPath := fmt.Sprintf("qrcodes/item_%s.png", item.ID)
	_ = qrcode.WriteFile("ITEM:"+item.ID, qrcode.High, 256, qrPath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"item":    item,
		"qr_url":  "/qrcodes/item_" + item.ID + ".png",
	})
}

// AdminGetItems GET /api/admin/items
func AdminGetItems(c *gin.Context) {
	db := database.GetDB()

	search := c.Query("search")
	category := c.Query("category")

	query := db.Preload("Location")

	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(part_number) LIKE ?", like, like, like)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var items []models.Item
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "items": items})
}

// AdminUpdateItem PUT /api/admin/item/:id
func AdminUpdateItem(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()

	var item models.Item
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Товар не найден"})
		return
	}

	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	item.Name = req.Name
	item.SKU = req.SKU
	item.Description = req.Description
	item.Quantity = req.Quantity
	item.Unit = req.Unit
	item.Category = req.Category
	item.PartNumber = req.PartNumber
	item.BatchNumber = req.BatchNumber
	item.BatchQuantity = req.BatchQuantity
	item.LocationID = req.LocationID
	item.UpdatedAt = time.Now()

	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "item": item})
}

// AdminDeleteItem DELETE /api/admin/item/:id
func AdminDeleteItem(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()

	if err := db.Delete(&models.Item{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AdminGetItemQR GET /api/admin/item/:id/qr — генерирует и отдаёт QR PNG
func AdminGetItemQR(c *gin.Context) {
	id := c.Param("id")

	qrPath := fmt.Sprintf("qrcodes/item_%s.png", id)

	// Генерируем если не существует
	if _, err := os.Stat(qrPath); os.IsNotExist(err) {
		if err := qrcode.WriteFile("ITEM:"+id, qrcode.High, 256, qrPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Ошибка генерации QR"})
			return
		}
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=qr_%s.png", id))
	c.File(qrPath)
}

// AdminUploadInvoicePhoto POST /api/admin/item/:id/photo
func AdminUploadInvoicePhoto(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()

	var item models.Item
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Товар не найден"})
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Файл не найден"})
		return
	}

	// Сохраняем в static/invoices/
	dir := "static/invoices"
	os.MkdirAll(dir, 0755)

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("invoice_%s%s", id, ext)
	savePath := filepath.Join(dir, filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Ошибка сохранения файла"})
		return
	}

	item.InvoicePhoto = "/invoices/" + filename
	item.UpdatedAt = time.Now()
	db.Save(&item)

	c.JSON(http.StatusOK, gin.H{"success": true, "photo_url": item.InvoicePhoto})
}

// ============================================================================
// LOCATIONS
// ============================================================================

type CreateLocationRequest struct {
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Row         string `json:"row"`
	Section     string `json:"section"`
	Shelf       string `json:"shelf"`
}

// AdminGetLocations GET /api/admin/locations
func AdminGetLocations(c *gin.Context) {
	db := database.GetDB()
	var locations []models.Location
	db.Order("code ASC").Find(&locations)
	c.JSON(http.StatusOK, gin.H{"success": true, "locations": locations})
}

// AdminCreateLocation POST /api/admin/location
func AdminCreateLocation(c *gin.Context) {
	var req CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := database.GetDB()

	loc := models.Location{
		ID:          "loc_" + uuid.New().String()[:8],
		Code:        req.Code,
		Description: req.Description,
		Row:         req.Row,
		Section:     req.Section,
		Shelf:       req.Shelf,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(&loc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Генерируем QR для локации
	qrPath := fmt.Sprintf("qrcodes/loc_%s.png", loc.ID)
	_ = qrcode.WriteFile("LOC:"+loc.ID, qrcode.High, 256, qrPath)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"location": loc,
		"qr_url":   "/qrcodes/loc_" + loc.ID + ".png",
	})
}

// AdminGetLocationQR GET /api/admin/location/:id/qr
func AdminGetLocationQR(c *gin.Context) {
	id := c.Param("id")
	qrPath := fmt.Sprintf("qrcodes/loc_%s.png", id)

	if _, err := os.Stat(qrPath); os.IsNotExist(err) {
		if err := qrcode.WriteFile("LOC:"+id, qrcode.High, 256, qrPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Ошибка генерации QR"})
			return
		}
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=qr_loc_%s.png", id))
	c.File(qrPath)
}

// ============================================================================
// CATEGORIES
// ============================================================================

// AdminGetCategories GET /api/admin/categories — уникальные категории из БД
func AdminGetCategories(c *gin.Context) {
	db := database.GetDB()
	var categories []string
	db.Model(&models.Item{}).Distinct("category").Where("category != ''").Pluck("category", &categories)
	c.JSON(http.StatusOK, gin.H{"success": true, "categories": categories})
}

// handlers/mechanic.go — добавить эти два метода

// UpdateOrderStatus PUT /api/mechanic/order/:id/status
func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	c.ShouldBindJSON(&req)
	db := database.GetDB()
	db.Model(&models.WorkOrder{}).Where("id = ?", id).Update("status", req.Status)
	c.JSON(200, gin.H{"success": true})
}

// GenerateOrderQR POST /api/mechanic/order/:id/qr
func GenerateOrderQR(c *gin.Context) {
	id := c.Param("id")
	qrPath := fmt.Sprintf("qrcodes/order_%s.png", id)
	qrcode.WriteFile("WO:"+id, qrcode.High, 256, qrPath)
	c.JSON(200, gin.H{"success": true, "qr_url": "/qrcodes/order_" + id + ".png"})
}

// IssueOrder POST /api/mechanic/order/:id/issue — списывает остатки
func IssueOrder(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	var order models.WorkOrder
	db.Preload("Items").First(&order, "id = ?", id)
	for _, item := range order.Items {
		if item.ItemID != "" {
			db.Model(&models.Item{}).Where("id = ?", item.ItemID).
				UpdateColumn("quantity", gorm.Expr("quantity - ?", item.Quantity))
		}
	}
	db.Model(&models.WorkOrder{}).Where("id = ?", id).Update("status", "issued")
	c.JSON(200, gin.H{"success": true})
}
