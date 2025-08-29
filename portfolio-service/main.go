package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"portfolio-service/database"
	"portfolio-service/grpc-client"
	"portfolio-service/handlers"
	"portfolio-service/services"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:3000"
		}

		origins := strings.Split(allowedOrigins, ",")
		origin := r.Header.Get("Origin")

		allowedOrigin := ""
		for _, o := range origins {
			// Fixed: use 'o' instead of 'allowedOrigin'
			if strings.TrimSpace(o) == origin {
				allowedOrigin = origin
				break
			}
		}

		// Fixed: set header when allowedOrigin is NOT empty
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Connect to Market Data Service via gRPC
	marketClient, err := grpcclient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to Market Data Service: %v", err)
	}
	defer marketClient.Close()

	// Create portfolio service
	portfolioService := services.NewPortfolioService(db, marketClient)

	// Create handlers
	h := handlers.NewHandlers(portfolioService)

	// Setup HTTP routes with CORS enabled for all endpoints
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	mux.HandleFunc("/portfolio/", h.GetPortfolioHandler)
	mux.HandleFunc("/buy", h.BuyStockHandler)
	mux.HandleFunc("/sell", h.SellStockHandler)
	mux.HandleFunc("/transactions/", h.GetTransactionsHandler)

	// Create HTTP server with CORS middleware applied to all routes
	server := &http.Server{
		Addr:    ":8003",
		Handler: enableCORS(mux), // Apply CORS to all routes
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("Portfolio Service starting...")
		fmt.Println("Server running on http://localhost:8003")
		fmt.Println("Available endpoints:")
		fmt.Println("- GET  /health")
		fmt.Println("- GET  /portfolio/ (requires JWT token)")
		fmt.Println("- POST /buy")
		fmt.Println("- POST /sell")
		fmt.Println("- GET  /transactions/ (requires JWT token)")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down Portfolio Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Portfolio Service stopped")
}
