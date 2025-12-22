package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/simonkvalheim/hm9-banking/internal/processor"
	"github.com/simonkvalheim/hm9-banking/internal/queue"
)

func main() {
	// Load configuration
	cfg := loadConfig()

	// Connect to database
	db, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Initialize processor and worker
	transferProcessor := processor.NewTransferProcessor(db)
	worker := queue.NewWorker(redisClient, transferProcessor)

	// Create context that cancels on shutdown signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutdown signal received, stopping worker...")
		cancel()
		worker.Stop()
	}()

	// Start the worker
	log.Println("Starting transaction worker...")
	worker.Start(ctx)

	log.Println("Worker stopped")
}

// Config holds all configuration for the worker
type Config struct {
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
}

// loadConfig reads configuration from environment variables
func loadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fjord:fjordpass@localhost:5432/fjorddb?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	return Config{
		DatabaseURL:   dbURL,
		RedisURL:      redisURL,
		RedisPassword: redisPassword,
	}
}

// connectDB creates a connection pool to PostgreSQL
func connectDB(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
