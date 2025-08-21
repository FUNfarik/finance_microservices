package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-service/models"
	"portfolio-service/services"
	"strconv"
	"strings"
)

type Handlers struct {
	portfolioService *services.PortfolioService
}

// NewHandlers creates a new handlers instance
func NewHandlers(portfolioService *services.PortfolioService) *Handlers {
	return &Handlers{
		portfolioService: portfolioService,
	}
}

// HealthHandler returns service health status
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.APIResponse{
		Status:  "success",
		Message: "Portfolio Service is healthy",
		Data: map[string]interface{}{
			"service": "portfolio-service",
			"port":    "8003",
			"version": "1.0.0",
		},
	}
	json.NewEncoder(w).Encode(response)
}

// GetPortfolioHandler retrieves user's portfolio
func (h *Handlers) GetPortfolioHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user_id from URL path
	path := strings.TrimPrefix(r.URL.Path, "/portfolio/")
	userID, err := strconv.Atoi(path)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  "Invalid user ID",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get portfolio from service
	portfolio, err := h.portfolioService.GetPortfolio(context.Background(), userID)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Failed to get portfolio: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.APIResponse{
		Status:  "success",
		Message: "Portfolio retrieved successfully",
		Data:    portfolio,
	}
	json.NewEncoder(w).Encode(response)
	fmt.Printf("📊 Portfolio requested for user %d\n", userID)
}

// BuyStockHandler processes stock purchases
func (h *Handlers) BuyStockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var buyReq models.BuyRequest
	err := json.NewDecoder(r.Body).Decode(&buyReq)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  "Invalid JSON request",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process buy order
	err = h.portfolioService.BuyStock(context.Background(), buyReq.UserID, buyReq.Symbol, buyReq.Shares)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Failed to buy stock: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Successfully bought %d shares of %s", buyReq.Shares, buyReq.Symbol),
		Data: map[string]interface{}{
			"symbol": buyReq.Symbol,
			"shares": buyReq.Shares,
			"action": "BUY",
		},
	}
	json.NewEncoder(w).Encode(response)
	fmt.Printf("Buy order: User %d bought %d shares of %s\n", buyReq.UserID, buyReq.Shares, buyReq.Symbol)
}

// SellStockHandler processes stock sales
func (h *Handlers) SellStockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var sellReq models.SellRequest
	err := json.NewDecoder(r.Body).Decode(&sellReq)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  "Invalid JSON request",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process sell order
	err = h.portfolioService.SellStock(context.Background(), sellReq.UserID, sellReq.Symbol, sellReq.Shares)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Failed to sell stock: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.APIResponse{
		Status:  "success",
		Message: fmt.Sprintf("Successfully sold %d shares of %s", sellReq.Shares, sellReq.Symbol),
		Data: map[string]interface{}{
			"symbol": sellReq.Symbol,
			"shares": sellReq.Shares,
			"action": "SELL",
		},
	}
	json.NewEncoder(w).Encode(response)
	fmt.Printf("💸 Sell order: User %d sold %d shares of %s\n", sellReq.UserID, sellReq.Shares, sellReq.Symbol)
}

// GetTransactionsHandler retrieves transaction history
func (h *Handlers) GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user_id from URL path
	path := strings.TrimPrefix(r.URL.Path, "/transactions/")
	userID, err := strconv.Atoi(path)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  "Invalid user ID",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get transactions from service
	transactions, err := h.portfolioService.GetTransactions(context.Background(), userID)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Failed to get transactions: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.APIResponse{
		Status:  "success",
		Message: "Transactions retrieved successfully",
		Data:    transactions,
	}
	json.NewEncoder(w).Encode(response)
	fmt.Printf("Transaction history requested for user %d\n", userID)
}
