package models

import "time"

// WorkOrder — заявка механика на получение деталей
type WorkOrder struct {
	ID              string          `gorm:"primaryKey" json:"id"`
	MechanicID      string          `gorm:"index" json:"mechanic_id"`
	Mechanic        *User           `gorm:"foreignKey:MechanicID;references:ID" json:"mechanic,omitempty"`
	Equipment       string          `json:"equipment"`        // название техники
	EquipmentNumber string          `json:"equipment_number"` // гос/инв номер
	WorkType        string          `json:"work_type"`        // ремонт, ТО, замена...
	Priority        string          `json:"priority"`         // normal, urgent
	Description     string          `json:"description"`
	Status          string          `json:"status"` // draft, pending, collecting, ready, issued
	Items           []WorkOrderItem `gorm:"foreignKey:WorkOrderID" json:"items,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// WorkOrderItem — строка заявки (одна деталь)
type WorkOrderItem struct {
	ID            int64  `gorm:"primaryKey" json:"id"`
	WorkOrderID   string `gorm:"index" json:"work_order_id"`
	ItemID        string `json:"item_id"` // если нашли в каталоге
	Name          string `json:"name"`    // название (ручной ввод или из каталога)
	PartNumber    string `json:"part_number"`
	Unit          string `json:"unit"`
	Quantity      int    `json:"quantity"`
	Justification string `json:"justification"` // обоснование
	PhotoURL      string `json:"photo_url"`     // фото детали
	Status        string `json:"status"`        // pending, collected, not_found
}

func (WorkOrderItem) TableName() string { return "work_order_items" }
