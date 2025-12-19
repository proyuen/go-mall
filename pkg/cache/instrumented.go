package cache

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Prometheus Metrics
var (
	redisRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_request_duration_seconds",
			Help:    "Duration of Redis requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	redisErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_errors_total",
			Help: "Total number of Redis errors",
		},
		[]string{"operation"},
	)

	redisCacheResults = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_cache_results_total",
			Help: "Total number of cache hits and misses",
		},
		[]string{"type"}, // hit, miss
	)
)

func init() {
	// Register metrics with Prometheus default registry
	// Ensure idempotent registration or handle errors in production if registered multiple times
	prometheus.MustRegister(redisRequestDuration)
	prometheus.MustRegister(redisErrors)
	prometheus.MustRegister(redisCacheResults)
}

// InstrumentedCache is a decorator for Cache that adds observability.
type instrumentedCache struct {
	next   Cache
	tracer trace.Tracer
}

// NewInstrumentedCache creates a new instrumented cache wrapper.
func NewInstrumentedCache(next Cache) Cache {
	return &instrumentedCache{
		next:   next,
		tracer: otel.Tracer("pkg/cache"),
	}
}

// Helper to record metrics and trace
func (c *instrumentedCache) observe(ctx context.Context, operation string, err error, start time.Time) {
	duration := time.Since(start).Seconds()
	redisRequestDuration.WithLabelValues(operation).Observe(duration)

	span := trace.SpanFromContext(ctx)
	if err != nil {
		redisErrors.WithLabelValues(operation).Inc()
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}
}

// Get retrieves a value from the cache.
func (c *instrumentedCache) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	ctx, span := c.tracer.Start(ctx, "redis.Get", trace.WithAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", "GET"),
		attribute.String("db.statement", key), // Be careful with PII in production
	))
	defer span.End()

	val, err := c.next.Get(ctx, key)
	c.observe(ctx, "get", err, start)

	// Record hit/miss only if no error
	if err == nil {
		if val == "" {
			redisCacheResults.WithLabelValues("miss").Inc()
		} else {
			redisCacheResults.WithLabelValues("hit").Inc()
		}
	}

	return val, err
}

// Set stores a value in the cache.
func (c *instrumentedCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	ctx, span := c.tracer.Start(ctx, "redis.Set", trace.WithAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", "SET"),
		attribute.String("db.statement", key),
	))
	defer span.End()

	err := c.next.Set(ctx, key, value, expiration)
	c.observe(ctx, "set", err, start)
	return err
}

// Del deletes keys from the cache.
func (c *instrumentedCache) Del(ctx context.Context, keys ...string) error {
	start := time.Now()
	// Join keys for tracing, but truncate if too long to avoid huge spans
	joinedKeys := strings.Join(keys, ",")
	if len(joinedKeys) > 100 {
		joinedKeys = joinedKeys[:100] + "..."
	}

	ctx, span := c.tracer.Start(ctx, "redis.Del", trace.WithAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", "DEL"),
		attribute.String("db.statement", joinedKeys),
	))
	defer span.End()

	err := c.next.Del(ctx, keys...)
	c.observe(ctx, "del", err, start)
	return err
}

// MGet retrieves multiple values.
func (c *instrumentedCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	start := time.Now()
	joinedKeys := strings.Join(keys, ",")
	if len(joinedKeys) > 100 {
		joinedKeys = joinedKeys[:100] + "..."
	}

	ctx, span := c.tracer.Start(ctx, "redis.MGet", trace.WithAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", "MGET"),
		attribute.String("db.statement", joinedKeys),
	))
	defer span.End()

	vals, err := c.next.MGet(ctx, keys...)
	c.observe(ctx, "mget", err, start)
	return vals, err
}

// Close closes the underlying cache.
func (c *instrumentedCache) Close() error {
	return c.next.Close()
}
