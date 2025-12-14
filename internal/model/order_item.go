package model

import (
	"gorm.io/gorm"
)

// OrderItem represents a single item within an order.
type OrderItem struct {
	gorm.Model
	OrderID  uint    `gorm:"not null;index" json:"order_id"` // Foreign key to Order
	SKUID    uint    `gorm:"not null;index" json:"sku_id"`   // Foreign key to SKU
	Quantity int     `gorm:"not null" json:"quantity"`
	Price    float64 `gorm:"not null;type:numeric(10,2)" json:"price"` // Price at the time of order
}
