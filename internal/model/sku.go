package model

import (
	"gorm.io/gorm"
)

// SKU (Stock Keeping Unit) represents a specific product with unique attributes (e.g., color, size).
type SKU struct {
	gorm.Model
	SPUID      uint    `gorm:"not null;index" json:"spu_id"` // Foreign key to SPU
	Attributes string  `gorm:"type:jsonb" json:"attributes"` // JSON string for attributes like {"color": "red", "size": "M"}
	Price      float64 `gorm:"not null;type:numeric(10,2)" json:"price"`
	Stock      int     `gorm:"not null;default:0" json:"stock"`
	Image      string  `gorm:"type:varchar(255)" json:"image_url"`
}
