package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/simonkvalheim/hm9-banking/internal/auth"
	"github.com/simonkvalheim/hm9-banking/internal/handler"
	appMiddleware "github.com/simonkvalheim/hm9-banking/internal/middleware"
	"github.com/simonkvalheim/hm9-banking/internal/processor"
	"github.com/simonkvalheim/hm9-banking/internal/queue"
	"github.com/simonkvalheim/hm9-banking/internal/repository"
)

func main() {
	// Load configuration from environment
	cfg := loadConfig()

	// Connect to database
	db, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Defer Redis cleanup (will be set if async mode enabled)
	var redisCleanup func()
	defer func() {
		if redisCleanup != nil {
			redisCleanup()
		}
	}()

	// Initialize repositories
	accountRepo := repository.NewAccountRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	customerRepo := repository.NewCustomerRepository(db)

	// Initialize auth service
	authConfig := auth.DefaultConfig(cfg.JWTSecret)
	authService := auth.NewService(authConfig, customerRepo)

	// Initialize processor
	transferProcessor := processor.NewTransferProcessor(db)

	// Initialize queue publisher if async mode is enabled
	var publisher *queue.Publisher
	if cfg.AsyncMode {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisURL,
			Password: cfg.RedisPassword,
			DB:       0,
		})
		redisCleanup = func() { redisClient.Close() }

		// Test Redis connection
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
		log.Println("Connected to Redis (async mode enabled)")
		publisher = queue.NewPublisher(redisClient)
	} else {
		log.Println("Running in sync mode (set ASYNC_MODE=true for async processing)")
	}

	// Initialize handlers
	accountHandler := handler.NewAccountHandler(accountRepo)
	transferHandler := handler.NewTransferHandler(txRepo, accountRepo, transferProcessor, publisher)
	authHandler := handler.NewAuthHandler(authService)

	// Initialize auth middleware
	authMiddleware := appMiddleware.NewAuthMiddleware(authService)

	// Set up router
	r := chi.NewRouter()

	// Middleware
	r.Use(appMiddleware.CORS(appMiddleware.DefaultCORSConfig())) // CORS for frontend
	r.Use(middleware.Logger)                                     // Logs each request
	r.Use(middleware.Recoverer)                                  // Recovers from panics gracefully

	// Health check (no auth needed)
	r.Get("/health", healthHandler(db))

	// Auth routes (public - no auth required)
	authHandler.RegisterRoutes(r)

	// Protected API routes (require authentication)
	r.Route("/v1", func(r chi.Router) {
		// Apply auth middleware to all /v1 routes
		r.Use(authMiddleware.RequireAuth)

		accountHandler.RegisterRoutes(r)
		transferHandler.RegisterRoutes(r)
	})

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Graceful shutdown setup
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// Config holds all configuration for the application
type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
	AsyncMode     bool   // If true, use Redis queue for async processing
	JWTSecret     string // Secret for signing JWT tokens
}

// loadConfig reads configuration from environment variables
func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default for local development
		dbURL = "postgres://fjord:fjordpass@localhost:5432/fjorddb?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Enable async mode if ASYNC_MODE=true
	asyncMode := os.Getenv("ASYNC_MODE") == "true"

	// JWT secret for token signing
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Default for local development - CHANGE IN PRODUCTION!
		jwtSecret = "dev-secret-change-in-production-use-openssl-rand-base64-32"
		log.Println("WARNING: Using default JWT_SECRET for development. Set JWT_SECRET environment variable in production!")
	}

	return Config{
		Port:          port,
		DatabaseURL:   dbURL,
		RedisURL:      redisURL,
		RedisPassword: redisPassword,
		AsyncMode:     asyncMode,
		JWTSecret:     jwtSecret,
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

	// Verify connection works
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

// healthHandler returns a handler that checks database connectivity
func healthHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Check database connection
		if err := db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "unhealthy", "database": "disconnected"}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "database": "connected"}`)
	}
}
