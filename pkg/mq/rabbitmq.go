package mq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultReconnectDelay   = 1 * time.Second
	maxReconnectDelay       = 30 * time.Second
	confirmationChannelSize = 1000 // Buffer for async confirmations
)

// RabbitMQ defines the interface for message queue operations.
type RabbitMQ interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	Consume(queue string, handler func(ctx context.Context, body []byte) error) error
	Close() error
}

type consumerConfig struct {
	queue   string
	handler func(ctx context.Context, body []byte) error
}

type rabbitMQ struct {
	url    string
	logger *slog.Logger

	mu      sync.RWMutex
	conn    *amqp.Connection
	channel *amqp.Channel

	isConnected bool
	notifyClose chan *amqp.Error

	// Async Confirmation Handling
	notifyConfirm chan amqp.Confirmation

	// Consumer Recovery
	consumers []consumerConfig

	reconnectDly time.Duration
}

// NewRabbitMQ creates a new RabbitMQ client with automatic reconnection and async publisher confirms.
func NewRabbitMQ(url string, logger *slog.Logger) (RabbitMQ, error) {
	mq := &rabbitMQ{
		url:          url,
		logger:       logger,
		reconnectDly: defaultReconnectDelay,
		consumers:    make([]consumerConfig, 0),
	}

	if err := mq.connect(); err != nil {
		return nil, err
	}

	go mq.reconnectLoop()

	return mq, nil
}

func (r *rabbitMQ) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Enable Publisher Confirms
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	r.conn = conn
	r.channel = ch
	r.notifyClose = make(chan *amqp.Error, 1)
	r.conn.NotifyClose(r.notifyClose)

	// Critical: Create NotifyPublish channel ONCE per channel creation
	// Buffer it to avoid blocking the AMQP library logic
	r.notifyConfirm = make(chan amqp.Confirmation, confirmationChannelSize)
	r.channel.NotifyPublish(r.notifyConfirm)

	// Start background goroutine to handle confirmations
	go r.handleConfirmations(r.notifyConfirm)

	r.isConnected = true
	r.logger.Info("Connected to RabbitMQ")

	return nil
}

// handleConfirmations processes async acks/nacks from the broker.
func (r *rabbitMQ) handleConfirmations(confirms <-chan amqp.Confirmation) {
	for c := range confirms {
		if c.Ack {
			// Message successfully delivered
			// In a full implementation, you might track DeliveryTags to resolve specific promises
			// r.logger.Debug("Message confirmed", "tag", c.DeliveryTag)
		} else {
			// Message failed
			r.logger.Error("Message failed to publish (Nack)", "tag", c.DeliveryTag)
			// TODO: Metric: rabbitmq_published_failed_total.Inc()
		}
	}
}

// reconnectLoop handles automatic reconnection and consumer recovery.
func (r *rabbitMQ) reconnectLoop() {
	for {
		err := <-r.notifyClose
		if err == nil {
			return // Connection closed gracefully via Close()
		}

		r.logger.Error("RabbitMQ connection lost, reconnecting...", "error", err)

		r.mu.Lock()
		r.isConnected = false
		r.mu.Unlock()

		for {
			time.Sleep(r.reconnectDly)
			if err := r.connect(); err == nil {
				r.logger.Info("RabbitMQ reconnected")
				r.reconnectDly = defaultReconnectDelay

				// Recover Consumers
				r.recoverConsumers()
				break
			}

			// Exponential Backoff
			if r.reconnectDly < maxReconnectDelay {
				r.reconnectDly *= 2
			}
			r.logger.Info("Retrying RabbitMQ connection...", "delay", r.reconnectDly)
		}
	}
}

// recoverConsumers re-registers all consumers after a reconnection.
func (r *rabbitMQ) recoverConsumers() {
	r.mu.RLock()
	// Copy consumers to avoid holding lock during registration calls (though registration needs lock too, be careful)
	// Actually Consume() acquires lock. So we should NOT hold lock here if calling r.Consume() directly.
	// But r.consumers is protected by r.mu.
	consumers := make([]consumerConfig, len(r.consumers))
	copy(consumers, r.consumers)
	r.mu.RUnlock()

	for _, cfg := range consumers {
		r.logger.Info("Recovering consumer", "queue", cfg.queue)
		// We call internal consume or helper that doesn't append to list again?
		// No, Consume() appends to list. If we call it again, we duplicate.
		// We need a separate internal method `startConsumer` that takes channel and config.

		// Actually, simpler: consume just appends to registry. `recoverConsumers` just starts them on the NEW channel.
		// But wait, the standard `Consume` function does two things:
		// 1. Adds to registry.
		// 2. Starts consuming.
		// So I need to split these or be careful.

		// Strategy:
		// Call internalStartConsumer for each.
		if err := r.internalStartConsumer(cfg.queue, cfg.handler); err != nil {
			r.logger.Error("Failed to recover consumer", "queue", cfg.queue, "error", err)
		}
	}
}

// internalStartConsumer registers the consumer on the current channel.
// It assumes r.mu is NOT held (it acquires it).
func (r *rabbitMQ) internalStartConsumer(queue string, handler func(ctx context.Context, body []byte) error) error {
	r.mu.RLock()
	if !r.isConnected {
		r.mu.RUnlock()
		return errors.New("rabbitmq not connected")
	}
	ch := r.channel
	r.mu.RUnlock()

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set Qos: %w", err)
	}

	msgs, err := ch.Consume(
		queue,
		"",    // consumer
		false, // auto-ack: FALSE
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			ctx := context.Background()
			if err := handler(ctx, d.Body); err != nil {
				r.logger.Error("Failed to process message", "queue", queue, "error", err)
				d.Nack(false, false)
			} else {
				d.Ack(false)
			}
		}
		r.logger.Info("Consumer stopped (channel closed)", "queue", queue)
	}()

	return nil
}

// Publish sends a persistent message asynchronously.
// It does NOT wait for confirmation, ensuring high throughput.
func (r *rabbitMQ) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.isConnected {
		return errors.New("rabbitmq not connected")
	}

	// Publish is non-blocking regarding network I/O wait for Ack.
	// It writes to the socket buffer.
	err := r.channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume registers a consumer and adds it to the registry for recovery.
func (r *rabbitMQ) Consume(queue string, handler func(ctx context.Context, body []byte) error) error {
	r.mu.Lock()
	r.consumers = append(r.consumers, consumerConfig{
		queue:   queue,
		handler: handler,
	})
	r.mu.Unlock()

	return r.internalStartConsumer(queue, handler)
}

func (r *rabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil && !r.conn.IsClosed() {
		return r.conn.Close()
	}
	return nil
}

func (r *rabbitMQ) GetChannel() *amqp.Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel
}
