package models

import "time"

// Portfolio represents a user's complete portfolio
type Portfolio struct {
	UserID     int       `json:"user_id"`
	TotalValue float64   `json:"total_value"`
	Cash       float64   `json:"cash"`
	Holdings   []Holding `json:"holdings"`
}

// Holding represents a stock position in user's portfolio
type Holding struct {
	Symbol       string  `json:"symbol"`
	Shares       int     `json:"shares"`
	AvgPrice     float64 `json:"avg_price"`
	CurrentPrice float64 `json:"current_price"`
	TotalValue   float64 `json:"total_value"`
	GainLoss     float64 `json:"gain_loss"`
}

// Transaction represents a buy/sell transaction
type Transaction struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	Symbol          string    `json:"symbol"`
	Shares          int       `json:"shares"`
	Price           float64   `json:"price"`
	TransactionType string    `json:"transaction_type"` // "BUY" or "SELL"
	TotalAmount     float64   `json:"total_amount"`
	Timestamp       time.Time `json:"timestamp"`
}

// BuyRequest represents a stock purchase request
type BuyRequest struct {
	Symbol string `json:"symbol"`
	Shares int    `json:"shares"`
	UserID int    `json:"user_id"`
}

// SellRequest represents a stock sale request
type SellRequest struct {
	Symbol string `json:"symbol"`
	Shares int    `json:"shares"`
	UserID int    `json:"user_id"`
}

// APIResponse represents standard API response format
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// StockPrice represents current stock price data
type StockPrice struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	ChangePercent float64 `json:"change_percent"`
}
