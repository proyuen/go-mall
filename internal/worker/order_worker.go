package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/proyuen/go-mall/internal/service"
	"github.com/proyuen/go-mall/pkg/cache"
	"github.com/proyuen/go-mall/pkg/mq"
)

// OrderMessage represents the payload for order creation events.
type OrderMessage struct {
	OrderID  uint64 `json:"order_id"`
	SKUID    uint64 `json:"sku_id"`
	Quantity int    `json:"quantity"`
}

// OrderWorker handles asynchronous tasks related to orders.
type OrderWorker struct {
	mq       mq.RabbitMQ
	invSvc   *service.InventoryService
	orderSvc service.OrderService
	cache    cache.Cache
	logger   *slog.Logger
}

// NewOrderWorker creates a new OrderWorker.
func NewOrderWorker(mq mq.RabbitMQ, invSvc *service.InventoryService, orderSvc service.OrderService, cache cache.Cache, logger *slog.Logger) *OrderWorker {
	return &OrderWorker{
		mq:       mq,
		invSvc:   invSvc,
		orderSvc: orderSvc,
		cache:    cache,
		logger:   logger,
	}
}

// Start begins consuming messages from the queue.
func (w *OrderWorker) Start() error {
	w.logger.Info("Starting OrderWorker...")
	return w.mq.Consume("orders.created", w.handleOrderCreated)
}

func (w *OrderWorker) handleOrderCreated(ctx context.Context, body []byte) error {
	var msg OrderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		w.logger.Error("Poison Pill: Failed to unmarshal order message", "error", err, "body", string(body))
		return nil // Ack to drop bad message
	}

	logger := w.logger.With("order_id", msg.OrderID)

	// Idempotency Check using Atomic SetNX
	idempotencyKey := fmt.Sprintf("processed:order:%d", msg.OrderID)
	acquired, err := w.cache.SetNX(ctx, idempotencyKey, "1", 24*time.Hour)
	if err != nil {
		logger.Error("Transient: Failed to check idempotency key", "error", err)
		return err // Retry
	}
	if !acquired {
		logger.Info("Duplicate ignored: Order already processed")
		return nil // Ack
	}

	logger.Info("Processing order event", "sku_id", msg.SKUID)

	// 1. Deduct Stock safely
	skuStr := fmt.Sprintf("%d", msg.SKUID)
	if err := w.invSvc.DeductStock(ctx, skuStr, msg.Quantity); err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			logger.Error("Terminal: Insufficient stock", "error", err)
			// Execute Compensation: Fail Order (TODO: Implement order failure logic)
			// We do NOT delete the idempotency key here. This prevents retrying a terminal error.
			return nil // Ack
		}

		logger.Error("Transient: Failed to deduct stock", "error", err)
		// System error (e.g. DB timeout) -> Delete idempotency key to allow retry
		if delErr := w.cache.Del(ctx, idempotencyKey); delErr != nil {
			logger.Error("Failed to rollback idempotency key", "error", delErr)
		}
		return err // Retry
	}

	logger.Info("Order processed successfully")
	return nil
}
