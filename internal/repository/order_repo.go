package repository

import (
	"context"
	"fmt"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/pkg/database"
	"gorm.io/gorm"
)

//go:generate mockgen -source=$GOFILE -destination=../mocks/order_repo_mock.go -package=mocks
// OrderRepository defines the interface for order data operations.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order, items []model.OrderItem) error
}

// orderRepository implements OrderRepository using GORM.
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository instance.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CreateOrder saves a new Order and its associated OrderItems in a single transaction.
func (r *orderRepository) CreateOrder(ctx context.Context, order *model.Order, items []model.OrderItem) error {
	db := database.GetDBFromContext(ctx, r.db)
	
	// Create the order
	if err := db.Create(order).Error; err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Set OrderID for each item and create them
	for i := range items {
		items[i].OrderID = order.ID
		if err := db.Create(&items[i]).Error; err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}
	return nil
}