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

	"github.com/golang-jwt/jwt/v5"
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

// extractUserIDFromToken extracts user ID from JWT token
func (h *Handlers) extractUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", fmt.Errorf("bearer token missing")
	}

	// Parse token without verification (since we trust the auth service)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"]
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
	}

	// Convert to string
	switch v := userID.(type) {
	case float64:
		return strconv.Itoa(int(v)), nil
	case int:
		return strconv.Itoa(v), nil
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("invalid user_id type in token")
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

	// Extract user_id from JWT token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Authentication failed: %v", err),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	fmt.Printf("Getting portfolio for user: %s\n", userID)

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
	fmt.Printf("✅ Portfolio sent for user %s\n", userID)
}

// BuyStockHandler processes stock purchases
// BuyStockHandler processes stock purchases
func (h *Handlers) BuyStockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user_id from JWT token (SECURITY FIX)
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Authentication failed: %v", err),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	var buyReq models.BuyRequest
	err = json.NewDecoder(r.Body).Decode(&buyReq)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  "Invalid JSON request",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process buy order using userID from JWT token
	err = h.portfolioService.BuyStock(context.Background(), userID, buyReq.Symbol, buyReq.Shares)
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
	fmt.Printf("Buy order: User %s bought %d shares of %s\n", userID, buyReq.Shares, buyReq.Symbol)
}

// SellStockHandler processes stock sales
func (h *Handlers) SellStockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user_id from JWT token (SECURITY FIX)
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Authentication failed: %v", err),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	var sellReq models.SellRequest
	err = json.NewDecoder(r.Body).Decode(&sellReq)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  "Invalid JSON request",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Process sell order using userID from JWT token
	err = h.portfolioService.SellStock(context.Background(), userID, sellReq.Symbol, sellReq.Shares)
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
	fmt.Printf("💸 Sell order: User %s sold %d shares of %s\n", userID, sellReq.Shares, sellReq.Symbol)
}

// GetTransactionsHandler retrieves transaction history
func (h *Handlers) GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user_id from JWT token
	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		response := models.APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("Authentication failed: %v", err),
		}
		w.WriteHeader(http.StatusUnauthorized)
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
	fmt.Printf("Transaction history requested for user %s\n", userID)
}
