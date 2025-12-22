package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/simonkvalheim/hm9-banking/internal/processor"
)

// Worker consumes messages from the queue and processes transactions
type Worker struct {
	client    *redis.Client
	processor *processor.TransferProcessor
	stopCh    chan struct{}
}

// NewWorker creates a new Worker
func NewWorker(client *redis.Client, proc *processor.TransferProcessor) *Worker {
	return &Worker{
		client:    client,
		processor: proc,
		stopCh:    make(chan struct{}),
	}
}

// Start begins consuming messages from the queue
// This runs in a loop until Stop() is called
func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker started, listening for transactions...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopping due to context cancellation")
			return
		case <-w.stopCh:
			log.Println("Worker stopping due to stop signal")
			return
		default:
			// Use BLPOP for blocking pop with timeout
			// This waits up to 5 seconds for a message, then loops to check for stop signal
			result, err := w.client.BLPop(ctx, 5*time.Second, QueueName).Result()
			if err != nil {
				if err == redis.Nil {
					// Timeout, no message available - continue loop
					continue
				}
				if ctx.Err() != nil {
					// Context cancelled
					return
				}
				log.Printf("Error reading from queue: %v", err)
				time.Sleep(1 * time.Second) // Brief pause before retry
				continue
			}

			// result[0] is the queue name, result[1] is the message
			if len(result) < 2 {
				continue
			}

			w.processMessage(ctx, result[1])
		}
	}
}

// Stop signals the worker to stop processing
func (w *Worker) Stop() {
	close(w.stopCh)
}

// processMessage handles a single message from the queue
func (w *Worker) processMessage(ctx context.Context, data string) {
	var msg TransactionMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	log.Printf("Processing transaction %s (type: %s)", msg.TransactionID, msg.Type)

	result, err := w.processor.Process(ctx, msg.TransactionID)
	if err != nil {
		log.Printf("Failed to process transaction %s: %v", msg.TransactionID, err)
		// In production, you might want to:
		// - Retry with exponential backoff
		// - Move to a dead-letter queue after max retries
		// - Send alerts
		return
	}

	if result.Success {
		log.Printf("Transaction %s completed successfully", msg.TransactionID)
	} else {
		log.Printf("Transaction %s failed: %s", msg.TransactionID, result.ErrorMessage)
	}
}

// ProcessOne processes a single message synchronously (useful for testing)
func (w *Worker) ProcessOne(ctx context.Context) error {
	result, err := w.client.LPop(ctx, QueueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil // No message available
		}
		return err
	}

	w.processMessage(ctx, result)
	return nil
}
