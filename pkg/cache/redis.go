package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/proyuen/go-mall/pkg/config"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

//go:generate mockgen -source=$GOFILE -destination=../../internal/mocks/cache_mock.go -package=mocks
// Cache defines a focused interface for caching operations.
type Cache interface {
	// Get retrieves a value from the cache.
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in the cache with a given expiration.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Del deletes one or more keys from the cache.
	Del(ctx context.Context, keys ...string) error

	// MGet retrieves multiple values from the cache.
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)

	// Close closes the Redis client.
	Close() error
}

type redisCache struct {
	client *redis.Client
	prefix string
	group  singleflight.Group
}

// NewRedisClient initializes a new Redis client.
func NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error) {
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 100
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     poolSize,
		MinIdleConns: poolSize / 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return client, nil
}

// NewRedisCache creates a new Redis cache wrapper using an existing client.
func NewRedisCache(client *redis.Client, keyPrefix string) Cache {
	return &redisCache{
		client: client,
		prefix: keyPrefix,
	}
}

// buildKey uses strings.Builder for minimized allocation key generation.
func (r *redisCache) buildKey(key string) string {
	if r.prefix == "" {
		return key
	}
	var b strings.Builder
	b.Grow(len(r.prefix) + 1 + len(key))
	b.WriteString(r.prefix)
	b.WriteByte(':')
	b.WriteString(key)
	return b.String()
}

// buildKeys constructs multiple keys efficiently.
func (r *redisCache) buildKeys(keys []string) []string {
	if r.prefix == "" {
		return keys
	}
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = r.buildKey(k)
	}
	return prefixed
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	builtKey := r.buildKey(key)

	// Use singleflight to prevent cache stampede
	val, err, _ := r.group.Do(builtKey, func() (interface{}, error) {
		res, err := r.client.Get(ctx, builtKey).Result()
		if err == redis.Nil {
			return "", nil
		}
		if err != nil {
			return "", fmt.Errorf("failed to get key '%s' from Redis: %w", key, err)
		}
		return res, nil
	})

	if err != nil {
		return "", err
	}
	return val.(string), nil
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, r.buildKey(key), value, expiration).Err()
}

func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, r.buildKeys(keys)...).Err()
}

func (r *redisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	return r.client.MGet(ctx, r.buildKeys(keys)...).Result()
}

func (r *redisCache) Close() error {
	return r.client.Close()
}