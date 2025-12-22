package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// QueueName is the Redis list key for pending transactions
	QueueName = "transactions:pending"
)

// TransactionMessage is the message published to the queue
type TransactionMessage struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Type          string    `json:"type"`
	PublishedAt   time.Time `json:"published_at"`
}

// Publisher handles publishing messages to Redis
type Publisher struct {
	client *redis.Client
}

// NewPublisher creates a new Publisher
func NewPublisher(client *redis.Client) *Publisher {
	return &Publisher{client: client}
}

// PublishTransaction publishes a transaction to the processing queue
func (p *Publisher) PublishTransaction(ctx context.Context, transactionID uuid.UUID, txType string) error {
	msg := TransactionMessage{
		TransactionID: transactionID,
		Type:          txType,
		PublishedAt:   time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Use RPUSH to add to the end of the list (FIFO queue)
	if err := p.client.RPush(ctx, QueueName, data).Err(); err != nil {
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	return nil
}

// QueueLength returns the current number of messages in the queue
func (p *Publisher) QueueLength(ctx context.Context) (int64, error) {
	return p.client.LLen(ctx, QueueName).Result()
}
