package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrLockNotHeld = errors.New("lock not held")
)

const (
	lockScript = `
		return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
	`
	unlockScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	renewScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
)

type RedisLock struct {
	client    *redis.Client
	key       string
	id        string
	stopWatch chan struct{}
}

// NewRedisLock creates a new distributed lock instance.
func NewRedisLock(client *redis.Client, key string) *RedisLock {
	return &RedisLock{
		client:    client,
		key:       key,
		id:        uuid.New().String(),
		stopWatch: make(chan struct{}),
	}
}

// Lock attempts to acquire the lock with a blocking wait.
// It tries to acquire the lock in a loop, sleeping for a short interval between attempts,
// until the lock is acquired or the context is cancelled/timed out.
// ttl is the expiration time for the lock.
// Returns true if lock is acquired, false if context is cancelled, or an error if Redis fails.
func (l *RedisLock) Lock(ctx context.Context, ttl time.Duration) (bool, error) {
	retryInterval := 50 * time.Millisecond // Recommended interval for retries

	for {
		// Attempt to acquire the lock
		resp, err := l.client.Eval(ctx, lockScript, []string{l.key}, l.id, ttl.Milliseconds()).Result()

		if err != nil && err != redis.Nil { // General Redis error
			return false, fmt.Errorf("redis error during lock attempt: %w", err)
		}

		if resp == "OK" { // Lock acquired successfully
			go l.watchdog(ttl)
			return true, nil
		}

		// Lock not acquired (err == redis.Nil or resp != "OK"). Wait and retry.
		select {
		case <-ctx.Done():
			return false, ctx.Err() // Context cancelled or timed out
		case <-time.After(retryInterval):
			// Sleep for retryInterval before next attempt
			continue
		}
	}
}

// Unlock releases the lock.
func (l *RedisLock) Unlock(ctx context.Context) error {
	// Signal watchdog to stop
	// Use a non-blocking send to ensure we don't block if watchdog is not ready or channel is full
	select {
	case l.stopWatch <- struct{}{}:
	default:
	}

	resp, err := l.client.Eval(ctx, unlockScript, []string{l.key}, l.id).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if count, ok := resp.(int64); ok && count == 1 {
		return nil
	}
	// If count is 0, it means we lost the lock or didn't hold it, or it expired before unlock.
	return ErrLockNotHeld
}

// watchdog extends the lock TTL periodically.
// It runs in a separate goroutine and renews the lock until it's stopped or loses the lock.
func (l *RedisLock) watchdog(ttl time.Duration) {
	// Renew every 1/3 of TTL. Should be less than half to avoid race conditions.
	renewInterval := ttl / 3
	if renewInterval <= 0 { // Ensure positive interval
		renewInterval = 1 * time.Second
	}
	ticker := time.NewTicker(renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopWatch:
			return // Unlock called, stop watchdog
		case <-ticker.C:
			// Use a new background context for renewal to not be tied to the original Lock() call's context
			// and give it a short timeout.
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			resp, err := l.client.Eval(ctx, renewScript, []string{l.key}, l.id, ttl.Milliseconds()).Result()
			cancel() // Release context resources

			// If renewal failed or we are no longer the owner (resp == 0), stop watchdog
			if err != nil || (resp != nil && resp.(int64) == 0) {
				return // Lock lost, failed to renew, or Redis error.
			}
		}
	}
}