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

	"portfolio-service/database"
	"portfolio-service/grpc-client"
	"portfolio-service/handlers"
	"portfolio-service/services"
)

func main() {
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Connect to Market Data Service via gRPC
	marketClient, err := grpcclient.Connect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Market Data Service: %v", err)
	}
	defer marketClient.Close()

	// Create portfolio service
	portfolioService := services.NewPortfolioService(db, marketClient)

	// Create handlers
	h := handlers.NewHandlers(portfolioService)

	// Setup HTTP routes
	http.HandleFunc("/health", h.HealthHandler)
	http.HandleFunc("/portfolio/", h.GetPortfolioHandler)
	http.HandleFunc("/buy", h.BuyStockHandler)
	http.HandleFunc("/sell", h.SellStockHandler)
	http.HandleFunc("/transactions/", h.GetTransactionsHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8003",
		Handler: nil, // Use default ServeMux
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("🚀 Portfolio Service starting...")
		fmt.Println("Server running on http://localhost:8003")
		fmt.Println("Available endpoints:")
		fmt.Println("- GET  /health")
		fmt.Println("- GET  /portfolio/{user_id}")
		fmt.Println("- POST /buy")
		fmt.Println("- POST /sell")
		fmt.Println("- GET  /transactions/{user_id}")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down Portfolio Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	fmt.Println("✅ Portfolio Service stopped")
}
