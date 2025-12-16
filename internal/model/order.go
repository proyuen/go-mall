package model

import (
	"github.com/shopspring/decimal"
)

type Order struct {
	Base
	UserID      uint64          `gorm:"index;not null" json:"user_id"`
	OrderNumber string          `gorm:"uniqueIndex;not null;type:varchar(64)" json:"order_number"`
	TotalAmount decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"total_amount"`
	Status      string          `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Items       []OrderItem     `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	Base
	OrderID       uint64          `gorm:"index;not null" json:"order_id"`
	SKUID         uint64          `gorm:"index;not null" json:"sku_id"`
	SnapshotName  string          `gorm:"not null;type:varchar(255)" json:"snapshot_name"`
	SnapshotImage string          `gorm:"type:varchar(255)" json:"snapshot_image"`
	Price         decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"price"` // Price at the time of order
	Quantity      int             `gorm:"not null;check:quantity > 0" json:"quantity"`
}
