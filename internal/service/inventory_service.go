package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/proyuen/go-mall/pkg/cache"
	"github.com/redis/go-redis/v9"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type InventoryService struct {
	cache       cache.Cache
	redisClient *redis.Client
}

func NewInventoryService(c cache.Cache, redisClient *redis.Client) *InventoryService {
	return &InventoryService{
		cache:       c,
		redisClient: redisClient,
	}
}

// DeductStock safely deducts stock for a given SKU using a distributed lock.
// It follows the pattern: Lock -> Get -> Check -> Update -> Unlock.
func (s *InventoryService) DeductStock(ctx context.Context, sku string, quantity int) error {
	lockKey := fmt.Sprintf("lock:sku:%s", sku)
	stockKey := fmt.Sprintf("stock:sku:%s", sku)

	// 1. Acquire Lock
	// We use the raw redis client to create the lock instance.
	lock := cache.NewRedisLock(s.redisClient, lockKey)

	// Create a context with timeout for acquiring the lock to prevent indefinite waiting
	lockCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Attempt to acquire lock with a 10s TTL (Watchdog will extend this if needed)
	acquired, err := lock.Lock(lockCtx, 10*time.Second)
	if err != nil {
		// Could be context timeout or redis error
		return fmt.Errorf("failed to acquire lock for sku %s: %w", sku, err)
	}
	if !acquired {
		return fmt.Errorf("failed to acquire lock for sku %s: timeout", sku)
	}

	// 2. Defer Unlock
	defer func() {
		// Use a detached context for unlock to ensure it runs even if the request context is canceled
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer unlockCancel()
		
		if err := lock.Unlock(unlockCtx); err != nil {
			// In production, log this error
			fmt.Printf("CRITICAL: failed to unlock %s: %v\n", lockKey, err)
		}
	}()

	// 3. Get Stock
	val, err := s.cache.Get(ctx, stockKey)
	if err != nil {
		return fmt.Errorf("failed to get stock from cache: %w", err)
	}

	currentStock := 0
	if val != "" {
		currentStock, err = strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("data corruption: invalid stock value '%s' for sku %s", val, sku)
		}
	}

	// 4. Business Rule Check
	if currentStock < quantity {
		return ErrInsufficientStock
	}

	// 5. Update Stock
	newStock := currentStock - quantity
	// Write back to cache (Simulating DB update)
	// Using 0 expiration (or keep existing) if supported, but Cache.Set requires duration.
	// We'll use 24 hours to keep it persistent-like.
	if err := s.cache.Set(ctx, stockKey, newStock, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}
