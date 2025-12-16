package model

import (
	"github.com/shopspring/decimal"
)

// SPU (Standard Product Unit) represents a product aggregation.
type SPU struct {
	Base
	Name        string `gorm:"not null;type:varchar(100)" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	CategoryID  uint64 `gorm:"index;not null" json:"category_id"`
	SKUs        []SKU  `gorm:"foreignKey:SPUID" json:"skus"`
}

// SKU (Stock Keeping Unit) represents a specific product variant.
type SKU struct {
	Base
	SPUID      uint64          `gorm:"index;not null" json:"spu_id"`
	Attributes JSONB           `gorm:"type:jsonb" json:"attributes"` // Dynamic attributes (Color, Size)
	Price      decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"price"`
	Stock      int             `gorm:"not null;check:stock >= 0" json:"stock"`
}
