package model

import (
	"gorm.io/gorm"
)

// Order represents a customer's order.
type Order struct {
	gorm.Model
	UserID      uint    `gorm:"not null;index" json:"user_id"` // Foreign key to User
	OrderNumber string  `gorm:"uniqueIndex;not null;type:varchar(50)" json:"order_number"`
	TotalAmount float64 `gorm:"not null;type:numeric(10,2)" json:"total_amount"`
	Status      string  `gorm:"not null;type:varchar(50);default:'pending'" json:"status"` // e.g., "pending", "paid", "shipped", "completed", "cancelled"
}
