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

	"github.com/simonkvalheim/hm9-banking/internal/handler"
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

	// Initialize repositories
	accountRepo := repository.NewAccountRepository(db)

	// Initialize handlers
	accountHandler := handler.NewAccountHandler(accountRepo)

	// Set up router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)    // Logs each request
	r.Use(middleware.Recoverer) // Recovers from panics gracefully

	// Health check (no auth needed)
	r.Get("/health", healthHandler(db))

	// API routes
	r.Route("/v1", func(r chi.Router) {
		accountHandler.RegisterRoutes(r)
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
	Port        string
	DatabaseURL string
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

	return Config{
		Port:        port,
		DatabaseURL: dbURL,
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