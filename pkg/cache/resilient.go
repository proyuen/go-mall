package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

// resilientCache is a decorator for Cache that adds Circuit Breaker and Retry logic.
type resilientCache struct {
	next    Cache
	breaker *gobreaker.CircuitBreaker
}

// NewResilientCache creates a new resilient cache wrapper.
func NewResilientCache(next Cache) Cache {
	st := gobreaker.Settings{
		Name:    "redis-cache",
		Timeout: 30 * time.Second, // Duration to stay in Open state
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip if >= 10 requests and >= 50% failure rate
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.5
		},
	}

	return &resilientCache{
		next:    next,
		breaker: gobreaker.NewCircuitBreaker(st),
	}
}

// executeWithRetry wraps the operation with Retry logic and executes it via Circuit Breaker.
func (c *resilientCache) executeWithRetry(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	// Retry wrapper
	retryOp := func() (interface{}, error) {
		var lastErr error
		for i := 0; i < 3; i++ { // Max 3 attempts
			// Check context cancellation before attempt
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			res, err := operation()
			if err == nil {
				return res, nil
			}

			// Do not retry on permanent errors (logic errors) or context errors
			// Ideally, we'd distinguish "retryable" errors. For now, we assume redis errors are mostly temporary (network/timeout).
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}

			lastErr = err
			// Simple exponential backoff: 10ms, 20ms, 40ms
			// Respect context during sleep
			sleepDuration := time.Duration(10<<i) * time.Millisecond
			select {
			case <-time.After(sleepDuration):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
	}

	// Execute via Circuit Breaker
	return c.breaker.Execute(retryOp)
}

// Get retrieves a value from the cache with resilience.
func (c *resilientCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.executeWithRetry(ctx, func() (interface{}, error) {
		return c.next.Get(ctx, key)
	})
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// Set stores a value in the cache with resilience.
func (c *resilientCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	_, err := c.executeWithRetry(ctx, func() (interface{}, error) {
		return nil, c.next.Set(ctx, key, value, expiration)
	})
	return err
}

// SetNX stores a value in the cache only if it does not exist with resilience.
func (c *resilientCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	res, err := c.executeWithRetry(ctx, func() (interface{}, error) {
		return c.next.SetNX(ctx, key, value, expiration)
	})
	if err != nil {
		return false, err
	}
	return res.(bool), nil
}

// Del deletes keys from the cache with resilience.
func (c *resilientCache) Del(ctx context.Context, keys ...string) error {
	_, err := c.executeWithRetry(ctx, func() (interface{}, error) {
		return nil, c.next.Del(ctx, keys...)
	})
	return err
}

// MGet retrieves multiple values with resilience.
func (c *resilientCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	val, err := c.executeWithRetry(ctx, func() (interface{}, error) {
		return c.next.MGet(ctx, keys...)
	})
	if err != nil {
		return nil, err
	}
	return val.([]interface{}), nil
}

// Close closes the underlying cache.
func (c *resilientCache) Close() error {
	return c.next.Close()
}