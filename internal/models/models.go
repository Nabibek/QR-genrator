package models

import (
	"time"

	"gorm.io/gorm"
)

// Location represents a warehouse location/shelf/bin
type Location struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"uniqueIndex" json:"code"`
	Description string    `json:"description"`
	Row         string    `json:"row"`
	Section     string    `json:"section"`
	Shelf       string    `json:"shelf"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Item represents an inventory item
type Item struct {
	ID             string         `gorm:"primaryKey" json:"id"`
	Name           string         `json:"name"`
	SKU            string         `gorm:"uniqueIndex" json:"sku"`
	Description    string         `json:"description"`
	Quantity       int            `json:"quantity"`
	Unit           string         `json:"unit"`     // шт, кг, м, л
	Category       string         `json:"category"` // категория/тип
	PartNumber     string         `json:"part_number"`
	BatchNumber    string         `json:"batch_number"`
	BatchQuantity  int            `json:"batch_quantity"`   // количество привезённого
	BatchArrivedAt *time.Time     `json:"batch_arrived_at"` // время приезда партии
	InvoicePhoto   string         `json:"invoice_photo"`    // путь к фото накладной
	LocationID     string         `gorm:"index" json:"location_id"`
	Location       *Location      `gorm:"foreignKey:LocationID;references:ID" json:"location,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// User represents a warehouse operator/admin
type User struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex" json:"username"`
	Email        string         `gorm:"uniqueIndex" json:"email"`
	PasswordHash string         `json:"-"`
	Role         string         `json:"role"` // admin, operator
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// ItemMovement represents the audit log of item movements
type ItemMovement struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	ItemID         string    `gorm:"index" json:"item_id"`
	Item           *Item     `gorm:"foreignKey:ItemID;references:ID" json:"item,omitempty"`
	FromLocationID string    `json:"from_location_id"`
	FromLocation   *Location `gorm:"foreignKey:FromLocationID;references:ID" json:"from_location,omitempty"`
	ToLocationID   string    `gorm:"index" json:"to_location_id"`
	ToLocation     *Location `gorm:"foreignKey:ToLocationID;references:ID" json:"to_location,omitempty"`
	UserID         string    `gorm:"index" json:"user_id"`
	User           *User     `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Notes          string    `json:"notes"`
	MovedAt        time.Time `json:"moved_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func (ItemMovement) TableName() string {
	return "item_movements"
}
