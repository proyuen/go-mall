package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/proyuen/go-mall/internal/model"
	"github.com/proyuen/go-mall/internal/repository"
	"github.com/proyuen/go-mall/pkg/database"
	"github.com/proyuen/go-mall/pkg/utils"
	"github.com/shopspring/decimal" // Import decimal package
)

// OrderCreateReq defines the request structure for creating a new order.
type OrderCreateReq struct {
	UserID uint64         `json:"user_id,string"` // Changed to uint64
	Items  []OrderItemReq `json:"items"`
}

type OrderItemReq struct {
	SKUID    uint64 `json:"sku_id,string"` // Changed to uint64
	Quantity int    `json:"quantity"`
}

// OrderCreateResp defines the response structure after creating an order.
type OrderCreateResp struct {
	OrderID     uint64          `json:"order_id,string"` // Changed to uint64
	OrderNumber string          `json:"order_number"`
	TotalAmount decimal.Decimal `json:"total_amount"`    // Changed to decimal.Decimal
}

//go:generate mockgen -source=$GOFILE -destination=../mocks/order_service_mock.go -package=mocks
// OrderService defines the interface for order business logic.
type OrderService interface {
	CreateOrder(ctx context.Context, req *OrderCreateReq) (*OrderCreateResp, error)
}

type orderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	txManager   database.TransactionManager
}

// NewOrderService creates a new OrderService instance.
func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, txManager database.TransactionManager) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		txManager:   txManager,
	}
}

// CreateOrder handles order creation logic: stock validation/deduction and order saving.
func (s *orderService) CreateOrder(ctx context.Context, req *OrderCreateReq) (*OrderCreateResp, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("order items cannot be empty")
	}

	// 1. Prepare data
	totalAmount := decimal.Zero // Changed to decimal.Decimal
	var orderItems []model.OrderItem
	
	// Generate a unique order number
	orderNumber := fmt.Sprintf("%d%s", time.Now().UnixNano(), utils.RandomString(6))

	// 2. Iterate items to check price and prepare order items
	for _, itemReq := range req.Items {
		// Fetch SKU details for price
		sku, err := s.productRepo.GetSKUByID(ctx, itemReq.SKUID)
		if err != nil {
			if errors.Is(err, repository.ErrSKUNotFound) {
				return nil, fmt.Errorf("SKU %d not found", itemReq.SKUID)
			}
			return nil, fmt.Errorf("failed to get SKU %d: %w", itemReq.SKUID, err)
		}

		if itemReq.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for SKU %d", itemReq.SKUID)
		}

		// Initial stock check
		if sku.Stock < itemReq.Quantity {
			return nil, fmt.Errorf("not enough stock for SKU %d", itemReq.SKUID)
		}

		// Calculate item total using decimal
		itemTotal := sku.Price.Mul(decimal.NewFromInt(int64(itemReq.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)

		orderItems = append(orderItems, model.OrderItem{
			SKUID:    itemReq.SKUID,
			Quantity: itemReq.Quantity,
			Price:    sku.Price, // Use SKU's price at the time of order
		})
	}

	// 3. Create Order Model
	order := &model.Order{
		UserID:      req.UserID,
		OrderNumber: orderNumber,
		TotalAmount: totalAmount,
		Status:      "pending",
	}

	// 4. Execute Transaction: Deduct Stock AND Create Order atomically
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// a. Deduct Stock
		for _, item := range orderItems {
			// Deduct stock (Quantity * -1) using transaction context
			if err := s.productRepo.UpdateSKUStock(txCtx, item.SKUID, -item.Quantity); err != nil {
				return fmt.Errorf("failed to deduct stock for SKU %d: %w", item.SKUID, err)
			}
		}

		// b. Create Order using transaction context
		if err := s.orderRepo.CreateOrder(txCtx, order, orderItems); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &OrderCreateResp{
		OrderID:     order.ID,
		OrderNumber: order.OrderNumber,
		TotalAmount: totalAmount,
	}, nil
}
