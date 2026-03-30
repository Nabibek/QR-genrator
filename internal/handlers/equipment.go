package handlers

import (
	"net/http"
	"strings"
	"time"

	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateEquipmentRequest struct {
	Name          string `json:"name" binding:"required"`
	Type          string `json:"type"`
	LicensePlate  string `json:"license_plate" binding:"required"`
	Year          int    `json:"year"`
	PurchasedAt   string `json:"purchased_at"`   // ISO date "2023-05-12"
	WarrantyUntil string `json:"warranty_until"` // ISO date "2026-05-12"
	UnderWarranty bool   `json:"under_warranty"`
}

// AdminCreateEquipment POST /api/admin/equipment
func AdminCreateEquipment(c *gin.Context) {
	var req CreateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := database.GetDB()

	var purchasedAt, warrantyUntil *time.Time
	if req.PurchasedAt != "" {
		t, err := time.Parse("2006-01-02", req.PurchasedAt)
		if err == nil {
			purchasedAt = &t
		}
	}
	if req.WarrantyUntil != "" {
		t, err := time.Parse("2006-01-02", req.WarrantyUntil)
		if err == nil {
			warrantyUntil = &t
			// Авто-определяем: на гарантии если дата не прошла
			req.UnderWarranty = t.After(time.Now())
		}
	}

	eq := models.Equipment{
		ID:            "eq_" + uuid.New().String()[:8],
		Name:          req.Name,
		Type:          req.Type,
		LicensePlate:  req.LicensePlate,
		Year:          req.Year,
		PurchasedAt:   purchasedAt,
		WarrantyUntil: warrantyUntil,
		UnderWarranty: req.UnderWarranty,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.Create(&eq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "equipment": eq})
}

// AdminGetEquipment GET /api/admin/equipment  — поиск + список
func AdminGetEquipment(c *gin.Context) {
	db := database.GetDB()
	search := c.Query("search")
	eqType := c.Query("type")

	query := db.Order("name ASC")
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(license_plate) LIKE ? OR LOWER(type) LIKE ?",
			like, like, like,
		)
	}
	if eqType != "" {
		query = query.Where("type = ?", eqType)
	}

	var list []models.Equipment
	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "equipment": list})
}

// AdminGetEquipmentTypes GET /api/admin/equipment/types — уникальные типы
func AdminGetEquipmentTypes(c *gin.Context) {
	db := database.GetDB()
	var types []string
	db.Model(&models.Equipment{}).Distinct("type").Where("type != ''").Pluck("type", &types)
	c.JSON(http.StatusOK, gin.H{"success": true, "types": types})
}

// AdminUpdateEquipment PUT /api/admin/equipment/:id
func AdminUpdateEquipment(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()

	var eq models.Equipment
	if err := db.First(&eq, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Техника не найдена"})
		return
	}

	var req CreateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	eq.Name = req.Name
	eq.Type = req.Type
	eq.LicensePlate = req.LicensePlate
	eq.Year = req.Year
	eq.UnderWarranty = req.UnderWarranty
	eq.UpdatedAt = time.Now()

	if req.PurchasedAt != "" {
		t, err := time.Parse("2006-01-02", req.PurchasedAt)
		if err == nil {
			eq.PurchasedAt = &t
		}
	}
	if req.WarrantyUntil != "" {
		t, err := time.Parse("2006-01-02", req.WarrantyUntil)
		if err == nil {
			eq.WarrantyUntil = &t
			eq.UnderWarranty = t.After(time.Now())
		}
	}

	db.Save(&eq)
	c.JSON(http.StatusOK, gin.H{"success": true, "equipment": eq})
}

// AdminDeleteEquipment DELETE /api/admin/equipment/:id
func AdminDeleteEquipment(c *gin.Context) {
	id := c.Param("id")
	db := database.GetDB()
	db.Delete(&models.Equipment{}, "id = ?", id)
	c.JSON(http.StatusOK, gin.H{"success": true})
}
